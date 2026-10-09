package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

const (
	defaultCommonsAPIURL  = "https://commons.wikimedia.org/w/api.php"
	commonsSearchCacheTTL = 7 * 24 * time.Hour
	commonsSearchLimit    = 12
	commonsResponseLimit  = 2 << 20
	commonsTimeout        = 15 * time.Second
	// Wikimedia asks clients to identify themselves with a way to learn more;
	// anonymous or browser-like agents are rate-limited much sooner.
	CommonsUserAgent = "Aurearia/1.0 (https://github.com/briandenicola/Aurearia; self-hosted coin collection app)"
)

// CommonsClient searches Wikimedia Commons files.
type CommonsClient interface {
	SearchFiles(ctx context.Context, query string, limit int) ([]byte, error)
}

// HTTPCommonsClient calls the Wikimedia Commons API.
type HTTPCommonsClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPCommonsClient creates a client; a nil client uses a default with a timeout.
func NewHTTPCommonsClient(client *http.Client) *HTTPCommonsClient {
	if client == nil {
		client = &http.Client{Timeout: commonsTimeout}
	}
	return &HTTPCommonsClient{client: client, baseURL: defaultCommonsAPIURL}
}

// SearchFiles returns the raw JSON of a file search with image details.
func (c *HTTPCommonsClient) SearchFiles(ctx context.Context, query string, limit int) ([]byte, error) {
	values := url.Values{
		"action":              {"query"},
		"format":              {"json"},
		"formatversion":       {"2"},
		"generator":           {"search"},
		"gsrnamespace":        {"6"},
		"gsrlimit":            {strconv.Itoa(limit)},
		"gsrsearch":           {query},
		"prop":                {"imageinfo"},
		"iiprop":              {"url|mime|extmetadata"},
		"iiurlwidth":          {"400"},
		"iiextmetadatafilter": {"Artist|LicenseShortName"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", CommonsUserAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("commons search returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, commonsResponseLimit+1))
	if err != nil {
		return nil, err
	}
	if len(body) > commonsResponseLimit {
		return nil, errors.New("commons response exceeded limit")
	}
	return body, nil
}

// CoinPhotoSearchService finds real photos of a coin type: the Numista
// catalogue photos of a linked type, then matching Wikimedia Commons files.
type CoinPhotoSearchService struct {
	catalog *NumistaCatalogService
	commons CommonsClient
	cache   *NumistaResponseCache
}

// NewCoinPhotoSearchService creates the service. catalog may be nil.
func NewCoinPhotoSearchService(catalog *NumistaCatalogService, commons CommonsClient, cache *NumistaResponseCache) *CoinPhotoSearchService {
	return &CoinPhotoSearchService{catalog: catalog, commons: commons, cache: cache}
}

// Search returns photos for a Numista type (when numistaTypeID > 0) and for
// the text query. Commons searches are cached for a week.
func (s *CoinPhotoSearchService) Search(ctx context.Context, query string, numistaTypeID int, lang string) models.CoinPhotoSearchResult {
	query = strings.Join(strings.Fields(query), " ")
	result := models.CoinPhotoSearchResult{Status: models.NumistaDataStatusOK, Query: query, Photos: []models.CoinPhotoCandidate{}}
	seen := map[string]bool{}
	add := func(photo models.CoinPhotoCandidate) {
		if photo.ImageURL == "" || seen[photo.ImageURL] {
			return
		}
		seen[photo.ImageURL] = true
		result.Photos = append(result.Photos, photo)
	}

	if numistaTypeID > 0 && s.catalog != nil {
		pageURL, _ := models.CanonicalNumistaURL(numistaTypeID)
		for _, photo := range s.catalog.Photos(ctx, numistaTypeID, lang) {
			imageURL := photo.Picture
			if imageURL == "" {
				imageURL = photo.Thumbnail
			}
			add(models.CoinPhotoCandidate{
				Source:    models.CoinPhotoSourceNumista,
				Side:      photo.Side,
				Thumbnail: photo.Thumbnail,
				ImageURL:  imageURL,
				PageURL:   pageURL,
				Author:    photo.Copyright,
				License:   photo.License,
			})
		}
	}

	if query == "" {
		return result
	}
	photos, err := s.searchCommons(ctx, query)
	if err == nil && len(photos) == 0 {
		if simpler := simplifyPhotoQuery(query); simpler != query {
			photos, err = s.searchCommons(ctx, simpler)
		}
	}
	if err != nil && len(result.Photos) == 0 {
		result.Status = models.NumistaDataStatusUnavailable
		return result
	}
	for _, photo := range photos {
		add(photo)
	}
	return result
}

func (s *CoinPhotoSearchService) searchCommons(ctx context.Context, query string) ([]models.CoinPhotoCandidate, error) {
	digest := sha256.Sum256([]byte(strings.ToLower(query)))
	body, _, _, err := s.cache.Get(ctx, "commons:"+hex.EncodeToString(digest[:]), commonsSearchCacheTTL, func(ctx context.Context) ([]byte, error) {
		return s.commons.SearchFiles(ctx, query, commonsSearchLimit)
	})
	if err != nil {
		return nil, err
	}
	return parseCommonsPhotos(body)
}

var commonsImageMimes = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}

func parseCommonsPhotos(body []byte) ([]models.CoinPhotoCandidate, error) {
	var response struct {
		Query struct {
			Pages []struct {
				Title     string `json:"title"`
				Index     int    `json:"index"`
				ImageInfo []struct {
					Mime           string `json:"mime"`
					URL            string `json:"url"`
					ThumbURL       string `json:"thumburl"`
					DescriptionURL string `json:"descriptionurl"`
					ExtMetadata    map[string]struct {
						Value json.RawMessage `json:"value"`
					} `json:"extmetadata"`
				} `json:"imageinfo"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	pages := response.Query.Pages
	// Pages come keyed by title; "index" is the search rank.
	for i := 1; i < len(pages); i++ {
		for j := i; j > 0 && pages[j].Index < pages[j-1].Index; j-- {
			pages[j], pages[j-1] = pages[j-1], pages[j]
		}
	}
	photos := []models.CoinPhotoCandidate{}
	for _, page := range pages {
		if len(page.ImageInfo) == 0 {
			continue
		}
		info := page.ImageInfo[0]
		if !commonsImageMimes[info.Mime] || httpsOnly(info.URL) == "" {
			continue
		}
		metadata := func(name string) string {
			var value string
			if field, ok := info.ExtMetadata[name]; ok && json.Unmarshal(field.Value, &value) == nil {
				return plainText(value, 200)
			}
			return ""
		}
		thumbnail := httpsOnly(info.ThumbURL)
		if thumbnail == "" {
			thumbnail = info.URL
		}
		photos = append(photos, models.CoinPhotoCandidate{
			Source:    models.CoinPhotoSourceWikimedia,
			Thumbnail: thumbnail,
			ImageURL:  info.URL,
			PageURL:   httpsURL(info.DescriptionURL),
			Title:     commonsFileTitle(page.Title),
			Author:    metadata("Artist"),
			License:   metadata("LicenseShortName"),
		})
	}
	return photos, nil
}

var (
	htmlTag         = regexp.MustCompile(`<[^>]*>`)
	photoQueryNoise = regexp.MustCompile(`[()\[\]{}"“”'’,;:/–—-]+`)
)

// plainText turns Commons' HTML metadata ("<a href=…>Name</a>") into text.
func plainText(value string, max int) string {
	text := strings.Join(strings.Fields(html.UnescapeString(htmlTag.ReplaceAllString(value, " "))), " ")
	if runes := []rune(text); len(runes) > max {
		text = string(runes[:max-1]) + "…"
	}
	return text
}

func commonsFileTitle(title string) string {
	title = strings.TrimPrefix(title, "File:")
	if dot := strings.LastIndex(title, "."); dot > 0 {
		title = title[:dot]
	}
	return title
}

// httpsURL accepts Commons' protocol-relative links ("//commons…").
func httpsURL(value string) string {
	if strings.HasPrefix(value, "//") {
		value = "https:" + value
	}
	return httpsOnly(value)
}

// simplifyPhotoQuery drops punctuation that makes Commons' full-text search
// miss, e.g. "Semis - Augustus (Gaius Caesar)" → "Semis Augustus Gaius Caesar".
func simplifyPhotoQuery(query string) string {
	return strings.Join(strings.Fields(photoQueryNoise.ReplaceAllString(query, " ")), " ")
}
