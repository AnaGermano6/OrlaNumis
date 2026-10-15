package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// OfficialPhotoHandler gives a coin being added its official images, so they
// need not be picked by hand: the European Central Bank's images of a euro
// coin's national side (obverse) and common side (reverse), or the photos a
// national mint (Portugal's INCM, the Monnaie de Paris, the Münze
// Österreich) has of a collector coin.
type OfficialPhotoHandler struct {
	coins  *repository.CoinRepository
	images *ImageHandler
	photos *services.CoinPhotoSearchService
	logger *services.Logger
	// One coin at a time, so the same image is never added twice.
	mu sync.Mutex
	// The ECB's images of the euro coins' sides, the same for many coins,
	// downloaded once instead of once per coin.
	commonSides map[string]officialPhotoFile
}

// officialSharedImages is where the ECB keeps the sides every coin of a
// denomination (and country) shares; at most officialSharedLimit are kept.
const (
	officialSharedImages = "https://www.ecb.europa.eu/euro/coins/common/shared/img/"
	officialSharedLimit  = 256
)

type officialPhotoFile struct {
	data []byte
	ext  string
}

// NewOfficialPhotoHandler creates the handler.
func NewOfficialPhotoHandler(coins *repository.CoinRepository, images *ImageHandler, photos *services.CoinPhotoSearchService, logger *services.Logger) *OfficialPhotoHandler {
	return &OfficialPhotoHandler{coins: coins, images: images, photos: photos, logger: logger}
}

// OfficialPhotosRequest names the sides to fill.
type OfficialPhotosRequest struct {
	Sides []string `json:"sides" binding:"required,min=1,max=2,dive,oneof=obverse reverse" example:"obverse,reverse"`
}

// OfficialPhotosResponse lists the images added.
type OfficialPhotosResponse struct {
	Images []models.CoinImage `json:"images"`
}

// Add gives a coin the official images of the sides asked for.
//
//	@Summary		Add a coin's official images
//	@Description	Adds a coin's official images to the sides asked for that have no image yet: the European Central Bank's images of a euro coin's national side (obverse) and common side (reverse), or, for a Portuguese, French or Austrian collector coin, the photos of its mint (INCM, Monnaie de Paris, Münze Österreich). Their credit is appended to the coin's notes. A coin neither has an image of (older collector coins, other currencies) gets none.
//	@Tags			Coins
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Coin ID"
//	@Param			body	body		OfficialPhotosRequest	true	"Sides to fill"
//	@Success		200		{object}	OfficialPhotosResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/official-photos [post]
func (h *OfficialPhotoHandler) Add(c *gin.Context) {
	userID := c.GetUint("userId")
	coinID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	var req OfficialPhotosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sides must list obverse and/or reverse"})
		return
	}
	sides := make([]models.ImageType, 0, len(req.Sides))
	for _, side := range req.Sides {
		sides = append(sides, models.ImageType(side))
	}
	images, found := h.Attach(c.Request.Context(), uint(coinID), userID, sides)
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
		return
	}
	c.JSON(http.StatusOK, OfficialPhotosResponse{Images: images})
}

// Attach adds the coin's official images for the given sides it has no image
// of and returns them; found is false when the user has no such coin.
func (h *OfficialPhotoHandler) Attach(ctx context.Context, coinID, userID uint, sides []models.ImageType) (added []models.CoinImage, found bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	coin, err := h.coins.FindByID(coinID, userID)
	if err != nil {
		return nil, false
	}
	added = []models.CoinImage{}
	if coin.Year == nil {
		return added, true
	}
	obverse, reverse := h.photos.OfficialPhotos(ctx, services.CoinPhotoQueryFor(coin))
	var credits []string
	for _, side := range []models.ImageType{models.ImageTypeObverse, models.ImageTypeReverse} {
		photo := obverse
		if side == models.ImageTypeReverse {
			photo = reverse
		}
		if photo == nil || !containsSide(sides, side) || hasImageOf(coin.Images, side) {
			continue
		}
		// The obverse leads; the reverse only when the coin has nothing else.
		primary := len(coin.Images) == 0 && len(added) == 0
		if image := h.save(ctx, coin, photo.ImageURL, side, primary); image != nil {
			added = append(added, *image)
			credits = append(credits, officialPhotoCredit(photo.Author, photo.PageURL))
		}
	}
	h.recordCredits(coin, credits)
	return added, true
}

func containsSide(sides []models.ImageType, side models.ImageType) bool {
	for _, wanted := range sides {
		if wanted == side {
			return true
		}
	}
	return false
}

func hasImageOf(images []models.CoinImage, side models.ImageType) bool {
	for _, image := range images {
		if image.ImageType == side {
			return true
		}
	}
	return false
}

func officialPhotoCredit(author, page string) string {
	credit := "© " + author
	if page != "" {
		credit += " – " + page
	}
	return credit
}

// recordCredits appends the photos' credits to the coin's notes, once each.
func (h *OfficialPhotoHandler) recordCredits(coin *models.Coin, credits []string) {
	if len(credits) == 0 {
		return
	}
	notes := strings.TrimRight(coin.Notes, "\n")
	for _, credit := range credits {
		if strings.Contains(notes, credit) {
			continue
		}
		if notes != "" {
			notes += "\n"
		}
		notes += credit
	}
	if len(notes) <= maxCoinNotesLength {
		if err := h.images.repo.UpdateCoinNotes(coin.ID, notes); err != nil {
			h.logger.Warn("images", "Could not record photo credit for coin %d: %v", coin.ID, err)
		}
	}
}

// save downloads an image and stores it as one side of the coin.
func (h *OfficialPhotoHandler) save(ctx context.Context, coin *models.Coin, rawURL string, side models.ImageType, primary bool) *models.CoinImage {
	source, ok := importablePhotoURL(rawURL)
	if !ok {
		return nil
	}
	file, cached := h.commonSides[source.String()]
	if !cached {
		data, ext, _, err := h.images.downloadPhoto(ctx, source)
		if err != nil {
			h.logger.Warn("images", "Official photo for coin %d from %s failed: %v", coin.ID, source.Host, err)
			return nil
		}
		file = officialPhotoFile{data: data, ext: ext}
		if strings.HasPrefix(source.String(), officialSharedImages) && len(h.commonSides) < officialSharedLimit {
			if h.commonSides == nil {
				h.commonSides = map[string]officialPhotoFile{}
			}
			h.commonSides[source.String()] = file
		}
	}
	image, err := h.images.svc.UploadImage(coin.ID, coin.UserID, file.data, file.ext, string(side), primary)
	if err != nil {
		h.logger.Error("images", "Saving official photo for coin %d failed: %v", coin.ID, err)
		return nil
	}
	h.logger.Info("images", "Added the official %s image to coin %d", side, coin.ID)
	return image
}
