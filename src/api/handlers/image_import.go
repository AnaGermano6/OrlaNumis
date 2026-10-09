package handlers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

const (
	importImageMaxBytes = 20 << 20
	maxCoinNotesLength  = 5000
)

// importImageExtensions maps the content types accepted from photo sources.
var importImageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type importImageRequest struct {
	URL       string `json:"url" binding:"required,max=2000" example:"https://upload.wikimedia.org/wikipedia/commons/5/54/Coin.jpg"`
	ImageType string `json:"imageType" binding:"required" example:"obverse"`
	IsPrimary bool   `json:"isPrimary" example:"true"`
	// Credit is appended to the coin's notes after a successful import.
	Credit string `json:"credit" binding:"max=500" example:"Obverse photo: Jane Doe, CC BY-SA 4.0 (Wikimedia Commons)"`
}

// Wikimedia rate-limits bursts (e.g. a page of thumbnails followed by an
// import) with 429; one paced retry is usually enough.
var (
	photoRetryDefaultDelay = 1500 * time.Millisecond
	photoRetryMaxDelay     = 5 * time.Second
)

func isWikimediaHost(host string) bool {
	host = strings.ToLower(host)
	return host == "upload.wikimedia.org" || host == "thumb.wikimedia.org"
}

// doWithRateLimitRetry sends a GET and, when the server answers 429 or 503,
// waits as asked (Retry-After, capped) and tries once more.
func doWithRateLimitRetry(client *http.Client, req *http.Request) (*http.Response, error) {
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable) {
		return resp, err
	}
	wait := photoRetryDefaultDelay
	if seconds, convErr := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After"))); convErr == nil && seconds > 0 {
		wait = min(time.Duration(seconds)*time.Second, photoRetryMaxDelay)
	}
	resp.Body.Close()
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case <-time.After(wait):
	}
	return client.Do(req)
}

// importablePhotoURL accepts only the public photo sources offered by the
// photo search, so the endpoint cannot be used to fetch arbitrary URLs.
func importablePhotoURL(raw string) (*url.URL, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" {
		return nil, false
	}
	switch strings.ToLower(parsed.Hostname()) {
	case "upload.wikimedia.org", "thumb.wikimedia.org":
		return parsed, strings.HasPrefix(parsed.Path, "/wikipedia/commons/")
	case "en.numista.com":
		return parsed, strings.HasPrefix(parsed.Path, "/catalogue/photos/")
	default:
		return nil, false
	}
}

// ImportRemote saves a real photo from Wikimedia Commons or Numista as one of
// the coin's images.
//
//	@Summary		Import a real coin photo
//	@Description	Downloads a photo from Wikimedia Commons or the Numista catalogue (as returned by GET /coin-photos/search) and saves it as a coin image, with thumbnails. The credit, if given, is appended to the coin's notes. A source that refuses the download (e.g. bot protection) returns 502 with code source_refused.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Coin ID"
//	@Param			body	body		importImageRequest	true	"Photo to import"
//	@Success		201		{object}	models.CoinImage
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Failure		502		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/images/import [post]
func (h *ImageHandler) ImportRemote(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	var req importImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	source, ok := importablePhotoURL(req.URL)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Photos can only be imported from Wikimedia Commons or the Numista catalogue"})
		return
	}
	if _, err := services.NormalizeImageType(req.ImageType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageType must be one of: obverse, reverse, detail, other"})
		return
	}
	coin, err := h.repo.FindCoinByOwner(uint(coinID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}

	data, ext, status, err := h.downloadPhoto(c, source)
	if err != nil {
		h.logger.Warn("images", "Photo import for coin %d from %s failed: %v", coinID, source.Host, err)
		code := "source_failed"
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			code = "source_refused"
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "The photo source did not provide the image", "code": code})
		return
	}

	image, err := h.svc.UploadImage(uint(coinID), userID, data, ext, req.ImageType, req.IsPrimary)
	if err != nil {
		h.logger.Error("images", "Saving imported photo for coin %d failed: %v", coinID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	if credit := strings.TrimSpace(req.Credit); credit != "" {
		notes := strings.TrimRight(coin.Notes, "\n")
		if notes != "" {
			notes += "\n"
		}
		notes += credit
		if len(notes) <= maxCoinNotesLength {
			if err := h.repo.UpdateCoinNotes(coin.ID, notes); err != nil {
				h.logger.Warn("images", "Could not record photo credit for coin %d: %v", coinID, err)
			}
		}
	}

	h.logger.Info("images", "Imported %s photo for coin %d from %s", req.ImageType, coinID, source.Host)
	c.JSON(http.StatusCreated, image)
}

// downloadPhoto fetches an image through the restricted outbound client and
// returns its bytes and file extension, or the HTTP status that refused it.
func (h *ImageHandler) downloadPhoto(c *gin.Context, source *url.URL) ([]byte, string, int, error) {
	if _, err := validateOutboundURL(source.String()); err != nil {
		return nil, "", 0, err
	}
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, source.String(), nil)
	if err != nil {
		return nil, "", 0, err
	}
	req.Header.Set("User-Agent", services.CommonsUserAgent)
	req.Header.Set("Accept", "image/*")
	resp, err := doWithRateLimitRetry(outboundHTTPClientFactory(), req)
	if err != nil {
		return nil, "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", resp.StatusCode, fmt.Errorf("source returned %d", resp.StatusCode)
	}
	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	ext, ok := importImageExtensions[strings.ToLower(mediaType)]
	if !ok {
		return nil, "", resp.StatusCode, fmt.Errorf("unsupported content type %q", mediaType)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, importImageMaxBytes+1))
	if err != nil {
		return nil, "", resp.StatusCode, err
	}
	if len(data) > importImageMaxBytes {
		return nil, "", resp.StatusCode, fmt.Errorf("image exceeds %d bytes", importImageMaxBytes)
	}
	return data, ext, resp.StatusCode, nil
}
