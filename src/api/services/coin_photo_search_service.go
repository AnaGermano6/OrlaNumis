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
	ecb     *EcbCoinPhotoService
	mints   []MintPhotoSource
}

// WithECB adds the ECB's images of €2 commemorative coins.
func (s *CoinPhotoSearchService) WithECB(ecb *EcbCoinPhotoService) *CoinPhotoSearchService {
	s.ecb = ecb
	return s
}

// WithMint adds a national mint's photos of the collector coins it sells.
func (s *CoinPhotoSearchService) WithMint(mint MintPhotoSource) *CoinPhotoSearchService {
	s.mints = append(s.mints, mint)
	return s
}

// CoinPhotoQuery describes the coin whose photos to find.
type CoinPhotoQuery struct {
	Query         string
	NumistaTypeID int
	Lang          string
	// Year, Country, Denomination and Subject identify a €2 commemorative
	// coin on the ECB site; without a year the ECB is not searched.
	Year         int
	Country      string
	Denomination string
	Subject      string
	// Material helps tell a coin's silver version from its base-metal one.
	Material string
}

// NewCoinPhotoSearchService creates the service. catalog may be nil.
func NewCoinPhotoSearchService(catalog *NumistaCatalogService, commons CommonsClient, cache *NumistaResponseCache) *CoinPhotoSearchService {
	return &CoinPhotoSearchService{catalog: catalog, commons: commons, cache: cache}
}

// Search returns photos for a Numista type (when numistaTypeID > 0) and for
// the text query. Commons searches are cached for a week.
func (s *CoinPhotoSearchService) Search(ctx context.Context, query string, numistaTypeID int, lang string) models.CoinPhotoSearchResult {
	return s.SearchCoin(ctx, CoinPhotoQuery{Query: query, NumistaTypeID: numistaTypeID, Lang: lang})
}

// SearchCoin returns the ECB image of a €2 commemorative coin first, then the
// Numista catalogue photos of its type and matching Wikimedia Commons files.
func (s *CoinPhotoSearchService) SearchCoin(ctx context.Context, coin CoinPhotoQuery) models.CoinPhotoSearchResult {
	query, numistaTypeID, lang := strings.Join(strings.Fields(coin.Query), " "), coin.NumistaTypeID, coin.Lang
	result := models.CoinPhotoSearchResult{Status: models.NumistaDataStatusOK, Query: query, Photos: []models.CoinPhotoCandidate{}}
	seen := map[string]bool{}
	add := func(photo models.CoinPhotoCandidate) {
		if photo.ImageURL == "" || seen[photo.ImageURL] {
			return
		}
		seen[photo.ImageURL] = true
		result.Photos = append(result.Photos, photo)
	}

	ecbPhotos := s.ecbPhotos(ctx, coin)
	for _, photo := range ecbPhotos {
		add(photo)
	}
	if len(ecbPhotos) == 0 {
		obverse, reverse := mintCandidates(s.mintPhoto(ctx, coin, s.numistaType(ctx, coin)))
		for _, photo := range []*models.CoinPhotoCandidate{obverse, reverse} {
			if photo != nil {
				add(*photo)
			}
		}
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

var twoEuroDenomination = regexp.MustCompile(`(?i)^\s*2\s*(€|eur|euro|euros)\b`)

// ecbPhotos finds the coin on the ECB's €2 commemorative pages. Only €2
// coins of a single year with a subject qualify, so a regular €2 coin never
// gets a commemorative's image.
func (s *CoinPhotoSearchService) ecbPhotos(ctx context.Context, coin CoinPhotoQuery) []models.CoinPhotoCandidate {
	if s.ecb == nil || coin.Year == 0 {
		return nil
	}
	query := EcbPhotoQuery{Year: coin.Year, Country: coin.Country, Subject: coin.Subject}
	commemorative := strings.TrimSpace(coin.Subject) != ""
	denomination := coin.Denomination
	if coin.NumistaTypeID > 0 && s.catalog != nil {
		if numistaType, err := s.catalog.knownTypeDetail(ctx, coin.NumistaTypeID, coin.Lang); err == nil {
			query.Title = numistaType.Title
			if query.Country == "" {
				query.Country = numistaType.Issuer.Code
			}
			if denomination == "" {
				denomination = numistaType.Title
			}
			commemorative = numistaType.MinYear != nil && numistaType.MaxYear != nil && *numistaType.MinYear == *numistaType.MaxYear
		}
	}
	if !commemorative || !twoEuroDenomination.MatchString(denomination) {
		return nil
	}
	photos, err := s.ecb.Match(ctx, query)
	if err != nil {
		return nil
	}
	return photos
}

// ExactPhoto returns the photo a source identifies as this very coin (the
// ECB's image of a €2 commemorative), or nil.
func (s *CoinPhotoSearchService) ExactPhoto(ctx context.Context, coin CoinPhotoQuery) *models.CoinPhotoCandidate {
	for _, photo := range s.ecbPhotos(ctx, coin) {
		if photo.Match == models.CoinPhotoMatchExact {
			return &photo
		}
	}
	return nil
}

// OfficialPhotos returns the ECB's images of a euro coin: its obverse (the
// €2 commemorative it is, or its country's regular design) and its common
// reverse; nil for a side the ECB has no image of.
func (s *CoinPhotoSearchService) OfficialPhotos(ctx context.Context, coin CoinPhotoQuery) (obverse, reverse *models.CoinPhotoCandidate) {
	if coin.Year == 0 {
		return nil, nil
	}
	country, denomination := coin.Country, coin.Denomination
	// A coin named after its country only, or of a value with no
	// commemoratives, has the regular design.
	regular := strings.TrimSpace(coin.Subject) == "" || !twoEuroDenomination.MatchString(denomination)
	numistaType := s.numistaType(ctx, coin)
	if numistaType != nil {
		regular = RegularEuroType(numistaType.Title, numistaType.MinYear, numistaType.MaxYear)
		if denomination == "" {
			denomination = numistaType.Title
		}
		if country == "" {
			country = numistaType.Issuer.Name
		}
	}
	if regular {
		if image, page := EcbNationalSideImage(denomination, coin.Year, country); image != "" {
			obverse = ecbOfficialPhoto(models.ImageTypeObverse, image, page)
		}
	} else {
		obverse = s.ExactPhoto(ctx, coin)
	}
	if image, page := EcbCommonSideImage(denomination, coin.Year, country); image != "" {
		reverse = ecbOfficialPhoto(models.ImageTypeReverse, image, page)
	}
	// Recent collector coins: their mint's own photos.
	if obverse == nil {
		if mintObverse, mintReverse := mintCandidates(s.mintPhoto(ctx, coin, numistaType)); mintObverse != nil {
			obverse, reverse = mintObverse, mintReverse
		}
	}
	return obverse, reverse
}

// numistaType returns the coin's Numista type, from the cache when it can.
func (s *CoinPhotoSearchService) numistaType(ctx context.Context, coin CoinPhotoQuery) *models.NumistaCatalogType {
	if coin.NumistaTypeID <= 0 || s.catalog == nil {
		return nil
	}
	numistaType, err := s.catalog.knownTypeDetail(ctx, coin.NumistaTypeID, coin.Lang)
	if err != nil {
		return nil
	}
	return numistaType
}

// mintPhoto finds a coin in its mint's catalogue, by its Numista type when
// it has one, else by its own fields.
func (s *CoinPhotoSearchService) mintPhoto(ctx context.Context, coin CoinPhotoQuery, numistaType *models.NumistaCatalogType) (MintPhotoSource, *MintPhoto) {
	query, ok := mintQueryForType(numistaType)
	mint := mintSourceFor(s.mints, coin.Country)
	if numistaType != nil {
		mint = mintSourceFor(s.mints, numistaType.Issuer.Code, numistaType.Issuer.Name)
	} else {
		query = MintPhotoQuery{Year: coin.Year, Value: MintFaceValue(coin.Denomination), Metal: coin.Material, Topic: coin.Subject}
		ok = query.Value > 0
	}
	if mint == nil || !ok {
		return nil, nil
	}
	photo, _ := mint.Match(ctx, query)
	if photo == nil {
		return nil, nil
	}
	return mint, mint.Confirm(ctx, photo)
}

func ecbOfficialPhoto(side models.ImageType, image, page string) *models.CoinPhotoCandidate {
	return &models.CoinPhotoCandidate{
		Source: models.CoinPhotoSourceECB, Side: string(side), Thumbnail: image, ImageURL: image, PageURL: page,
		Author: "European Central Bank", Match: models.CoinPhotoMatchExact,
	}
}
