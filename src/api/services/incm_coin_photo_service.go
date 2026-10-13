package services

import (
	"context"

	"encoding/json"
	"fmt"
	"github.com/briandenicola/ancient-coins-api/models"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// The INCM (Imprensa Nacional-Casa da Moeda), Portugal's mint, sells its
// collector coins in an online shop whose public product feed carries the
// official photos of each coin. It covers the recent Portuguese collector
// coins, whose Numista photos the app cannot load.
const (
	incmShopURL         = "https://loja.incm.pt"
	incmProductsPath    = "%s/products.json?limit=250&page=%d"
	incmPageLimit       = 8 << 20
	incmPageTimeout     = 20 * time.Second
	incmMaxPages        = 30
	incmCoinsCacheKey   = "incm:coins:v4"
	incmCoinsCacheTTL   = 7 * 24 * time.Hour
	incmMemoryLifetime  = time.Hour
	incmRetryDelay      = 10 * time.Minute
	incmThumbnailWidth  = "400"
	incmImagePathPrefix = "/s/files/1/0525/1910/8762/"
)

// IncmImageHost and IncmImagePath locate the INCM shop's photos on Shopify's CDN.
const (
	IncmImageHost = "cdn.shopify.com"
	IncmImagePath = incmImagePathPrefix
)

// INCMShopClient reads the INCM shop's product feed, in Portuguese (lang "")
// or English ("en").
type INCMShopClient interface {
	ProductsPage(ctx context.Context, lang string, page int) ([]byte, error)
}

// HTTPINCMShopClient reads the shop over HTTP.
type HTTPINCMShopClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPINCMShopClient creates the client; a nil client gets a default one.
func NewHTTPINCMShopClient(client *http.Client) *HTTPINCMShopClient {
	if client == nil {
		client = &http.Client{Timeout: incmPageTimeout}
	}
	return &HTTPINCMShopClient{client: client, baseURL: incmShopURL}
}

// ProductsPage returns one page (up to 250 products) of the shop's feed.
func (c *HTTPINCMShopClient) ProductsPage(ctx context.Context, lang string, page int) ([]byte, error) {
	prefix := ""
	if lang != "" {
		prefix = "/" + lang
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fmt.Sprintf(incmProductsPath, prefix, page), nil)
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
		return nil, fmt.Errorf("incm products page %d: status %d", page, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, incmPageLimit))
}

// IncmCoinPhotoService finds a coin's photos in the INCM shop.
type IncmCoinPhotoService struct {
	client INCMShopClient
	cache  *NumistaResponseCache
	now    func() time.Time

	mu       sync.Mutex
	coins    []mintCoin
	loadedAt time.Time
}

// NewIncmCoinPhotoService creates the service; the shop's coins are kept in
// the shared response cache for a week.
func NewIncmCoinPhotoService(client INCMShopClient, cache *NumistaResponseCache) *IncmCoinPhotoService {
	return &IncmCoinPhotoService{client: client, cache: cache, now: time.Now}
}

// Coins returns the shop's coins, reading the feed at most once a week.
func (s *IncmCoinPhotoService) Coins(ctx context.Context) ([]mintCoin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.coins != nil && s.now().Sub(s.loadedAt) < incmMemoryLifetime {
		return s.coins, nil
	}
	body, _, _, err := s.cache.Get(ctx, incmCoinsCacheKey, incmCoinsCacheTTL, func(ctx context.Context) ([]byte, error) {
		coins, err := s.readShop(ctx, "")
		if err != nil {
			return nil, err
		}
		// English titles are a bonus: without them only Portuguese ones match.
		if english, err := s.readShop(ctx, "en"); err == nil {
			titles := map[string]string{}
			for _, coin := range english {
				titles[coin.PageURL] = coin.Title
			}
			for i := range coins {
				if title := titles[coins[i].PageURL]; title != "" && title != coins[i].Title {
					coins[i].OtherTitles = append(coins[i].OtherTitles, title)
				}
			}
		}
		return json.Marshal(coins)
	})
	var coins []mintCoin
	if err == nil {
		err = json.Unmarshal(body, &coins)
	}
	if err != nil {
		// Try the shop again in a few minutes, not on every search.
		s.coins, s.loadedAt = []mintCoin{}, s.now().Add(incmRetryDelay-incmMemoryLifetime)
		return nil, err
	}
	s.coins, s.loadedAt = coins, s.now()
	return coins, nil
}

// readShop reads every page of the shop's feed in one language.
func (s *IncmCoinPhotoService) readShop(ctx context.Context, lang string) ([]mintCoin, error) {
	var coins []mintCoin
	for page := 1; page <= incmMaxPages; page++ {
		body, err := s.client.ProductsPage(ctx, lang, page)
		if err != nil {
			return nil, err
		}
		products, err := ParseIncmProducts(body)
		if err != nil {
			return nil, err
		}
		if products.count == 0 {
			break
		}
		coins = append(coins, products.coins...)
	}
	return coins, nil
}

type incmProductsPage struct {
	// count is the number of products on the page, coins or not.
	count int
	coins []mintCoin
}

type incmProduct struct {
	Title       string   `json:"title"`
	Handle      string   `json:"handle"`
	ProductType string   `json:"product_type"`
	PublishedAt string   `json:"published_at"`
	Tags        []string `json:"tags"`
	BodyHTML    string   `json:"body_html"`
	Images      []struct {
		Src string `json:"src"`
	} `json:"images"`
}

var (
	incmReleaseDate = regexp.MustCompile(`\[DATA_LANCAMENTO\]\s*(\d{1,2})-(\d{4})`)
	incmIssueLimit  = regexp.MustCompile(`\[LIM_EMISSAO\]\s*(\d[\d .]*)`)
	incmTags        = regexp.MustCompile(`<[^>]+>`)
	incmFaceValue   = regexp.MustCompile(`^(\d+(?:[.,]\d+)?)\s*euros?$`)
	// The side a photo shows is the last mark of its file name: "_A", "-1" or
	// "_b" (an upload suffix such as "_f47d4d3d-…" aside). "-0" and unmarked
	// files show both sides or the packaging.
	incmSideMark   = regexp.MustCompile(`(?i)[_-]([a-z]|\d)$`)
	incmUploadUUID = regexp.MustCompile(`(?i)_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	incmPackaging  = regexp.MustCompile(`(?i)embal|emblag|carregamento`)
)

// ParseIncmProducts reads one page of the shop's feed, keeping the coins that
// have a face value and a photo of each side.
func ParseIncmProducts(body []byte) (incmProductsPage, error) {
	var feed struct {
		Products []incmProduct `json:"products"`
	}
	if err := json.Unmarshal(body, &feed); err != nil {
		return incmProductsPage{}, err
	}
	page := incmProductsPage{count: len(feed.Products)}
	for _, product := range feed.Products {
		if coin, ok := incmCoinFromProduct(product); ok {
			page.coins = append(page.coins, coin)
		}
	}
	return page, nil
}

func incmCoinFromProduct(product incmProduct) (mintCoin, bool) {
	if !strings.EqualFold(strings.TrimSpace(product.ProductType), "Moedas") {
		return mintCoin{}, false
	}
	coin := mintCoin{Title: incmSpaces(product.Title)}
	if handle := strings.TrimSpace(product.Handle); handle != "" {
		coin.PageURL = incmShopURL + "/products/" + handle
	}
	for _, tag := range product.Tags {
		name, value, _ := strings.Cut(strings.TrimSpace(tag), " ")
		switch strings.ToUpper(name) {
		case "VALOR":
			if match := incmFaceValue.FindStringSubmatch(foldASCII(strings.TrimPrefix(strings.TrimSpace(value), "FACIAL "))); match != nil {
				coin.Value, _ = strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64)
			}
		case "METAL":
			coin.Metal = strings.TrimSpace(value)
		case "AUTOR":
			coin.Authors = strings.TrimSpace(value)
		case "SÉRIE", "SERIE":
			coin.Series = strings.TrimSpace(value)
		}
	}
	text := html.UnescapeString(incmTags.ReplaceAllString(product.BodyHTML, " "))
	if match := incmReleaseDate.FindStringSubmatch(text); match != nil {
		coin.Year, _ = strconv.Atoi(match[2])
		if month, _ := strconv.Atoi(match[1]); month >= 1 && month <= 12 {
			coin.Released = fmt.Sprintf("%04d-%02d", coin.Year, month)
		}
	} else if len(product.PublishedAt) >= 4 {
		coin.Year, _ = strconv.Atoi(product.PublishedAt[:4])
	}
	if match := incmIssueLimit.FindStringSubmatch(text); match != nil {
		coin.Volume, _ = strconv.ParseInt(strings.NewReplacer(" ", "", ".", "").Replace(strings.TrimSpace(match[1])), 10, 64)
	}
	if description, _, found := strings.Cut(text, "|*|"); found {
		text = description
	}
	coin.Text = incmSpaces(text)
	for _, image := range product.Images {
		src := strings.TrimSpace(image.Src)
		name := src[strings.LastIndex(src, "/")+1:]
		name, _, _ = strings.Cut(name, "?")
		if dot := strings.LastIndex(name, "."); dot > 0 {
			name = name[:dot]
		}
		name = incmUploadUUID.ReplaceAllString(name, "")
		mark := incmSideMark.FindStringSubmatch(name)
		if mark == nil || incmPackaging.MatchString(name) {
			continue
		}
		switch strings.ToLower(mark[1]) {
		case "a", "1":
			if coin.Obverse == "" {
				coin.Obverse = src
			}
		case "b", "2":
			if coin.Reverse == "" {
				coin.Reverse = src
			}
		default:
			// "-0" shows both sides and "-3" a detail, but "_C" and on mean a
			// set of coins, not one coin.
			if !unicode.IsDigit(rune(mark[1][0])) {
				return mintCoin{}, false
			}
		}
	}
	ok := coin.Value > 0 && coin.Year > 0 && coin.Obverse != "" && coin.Reverse != "" && coin.PageURL != ""
	return coin, ok
}

func incmSpaces(value string) string {
	return strings.Join(strings.FieldsFunc(value, unicode.IsSpace), " ")
}

// Match returns the shop's photos of the coin, or nil when no coin of the
// shop is clearly it.
func (s *IncmCoinPhotoService) Match(ctx context.Context, query MintPhotoQuery) (*MintPhoto, error) {
	if query.Year == 0 || query.Value <= 0 {
		return nil, nil
	}
	coins, err := s.Coins(ctx)
	if err != nil {
		return nil, err
	}
	coin := matchMintCoin(coins, query)
	if coin == nil {
		return nil, nil
	}
	return &MintPhoto{Title: coin.Title, Obverse: coin.Obverse, Reverse: coin.Reverse, PageURL: coin.PageURL}, nil
}

// Country, Source and Credit describe the INCM as a MintPhotoSource.
func (s *IncmCoinPhotoService) Country() string { return "Portugal" }
func (s *IncmCoinPhotoService) Source() string  { return models.CoinPhotoSourceINCM }
func (s *IncmCoinPhotoService) Credit() string  { return "INCM" }

// Confirm has nothing to check: the shop names both sides' photos.
func (s *IncmCoinPhotoService) Confirm(_ context.Context, photo *MintPhoto) *MintPhoto { return photo }

// Thumbnail asks Shopify's CDN for a smaller copy of a shop photo.
func (s *IncmCoinPhotoService) Thumbnail(image string) string {
	if strings.Contains(image, "?") {
		return image + "&width=" + incmThumbnailWidth
	}
	return image + "?width=" + incmThumbnailWidth
}
