package services

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

// The Monnaie de Paris, France's mint, lists every product of its shop, with
// its photos, in a public sitemap (one per language). Coin products are named
// "<topic>-<value>eur-<metal>-coin-<quality>-yeardate-<year>". The first photo
// shows the coin's design; the other side is the photo named "revers", or the
// one taken right after it (checked to really show a coin, as it can also be
// the packaging). Each product's page says when the coin goes on sale and
// how many are struck: read in the background for this year's coins and
// later ones, one page at a time.
const (
	mdpShopURL         = "https://www.monnaiedeparis.fr"
	mdpSitemapPath     = "/media/sitemap/sitemap_mdp_%s.xml"
	mdpSitemapLimit    = 32 << 20
	mdpImageLimit      = 8 << 20
	mdpTimeout         = 30 * time.Second
	mdpCoinsCacheKey   = "mdp:coins:v3"
	mdpCoinsCacheTTL   = 7 * 24 * time.Hour
	mdpProductLimit    = 8 << 20
	mdpProductCacheKey = "mdp:product:v1:"
	mdpProductCacheTTL = 14 * 24 * time.Hour
	// A page that could not be read is tried again the next day.
	mdpProductRetryTTL = 24 * time.Hour
	// The pause between two product pages read in the background.
	mdpProductDelay     = 2 * time.Second
	mdpShapeCacheKey    = "mdp:shape:v1:"
	mdpShapeCacheTTL    = 365 * 24 * time.Hour
	mdpMemoryLifetime   = time.Hour
	mdpRetryDelay       = 10 * time.Minute
	mdpThumbnailOptions = "?optimize=medium&fit=bounds&height=400&width=400"
)

// MdpImagePath is where the Monnaie de Paris serves its product photos.
const (
	MdpImageHost = "www.monnaiedeparis.fr"
	MdpImagePath = "/media/catalog/product/"
)

// MDPShopClient reads the Monnaie de Paris sitemaps, product pages and photos.
type MDPShopClient interface {
	Sitemap(ctx context.Context, lang string) ([]byte, error)
	Product(ctx context.Context, url string) ([]byte, error)
	Image(ctx context.Context, url string) ([]byte, error)
}

// HTTPMDPShopClient reads them over HTTP.
type HTTPMDPShopClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPMDPShopClient creates the client; a nil client gets a default one.
func NewHTTPMDPShopClient(client *http.Client) *HTTPMDPShopClient {
	if client == nil {
		client = &http.Client{Timeout: mdpTimeout}
	}
	return &HTTPMDPShopClient{client: client, baseURL: mdpShopURL}
}

// Sitemap returns the shop's sitemap in "fr" or "en".
func (c *HTTPMDPShopClient) Sitemap(ctx context.Context, lang string) ([]byte, error) {
	return c.get(ctx, c.baseURL+fmt.Sprintf(mdpSitemapPath, lang), mdpSitemapLimit)
}

// Product returns one of the shop's product pages.
func (c *HTTPMDPShopClient) Product(ctx context.Context, url string) ([]byte, error) {
	if !strings.HasPrefix(url, c.baseURL+"/en/") && !strings.HasPrefix(url, c.baseURL+"/fr/") {
		return nil, fmt.Errorf("not a Monnaie de Paris product: %s", url)
	}
	return c.get(ctx, url, mdpProductLimit)
}

// Image returns one of the shop's photos.
func (c *HTTPMDPShopClient) Image(ctx context.Context, url string) ([]byte, error) {
	if !strings.HasPrefix(url, c.baseURL+MdpImagePath) {
		return nil, fmt.Errorf("not a Monnaie de Paris photo: %s", url)
	}
	return c.get(ctx, url, mdpImageLimit)
}

func (c *HTTPMDPShopClient) get(ctx context.Context, url string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", CommonsUserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("monnaie de paris %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, limit))
}

// MdpCoinPhotoService finds a French coin's photos in the Monnaie de Paris shop.
type MdpCoinPhotoService struct {
	client MDPShopClient
	cache  *NumistaResponseCache
	now    func() time.Time
	// productDelay is the pause between two product pages read in the background.
	productDelay time.Duration

	mu       sync.Mutex
	coins    []mintCoin
	loadedAt time.Time
	// reading is true while product pages are read in the background; reads
	// waits for that.
	reading bool
	reads   sync.WaitGroup
}

// NewMdpCoinPhotoService creates the service; the shop's coins are kept in
// the shared response cache for a week, what each product page says for two.
func NewMdpCoinPhotoService(client MDPShopClient, cache *NumistaResponseCache) *MdpCoinPhotoService {
	return &MdpCoinPhotoService{client: client, cache: cache, now: time.Now, productDelay: mdpProductDelay}
}

// Country, Source and Credit describe the Monnaie de Paris as a MintPhotoSource.
func (s *MdpCoinPhotoService) Country() string { return "France" }
func (s *MdpCoinPhotoService) Source() string  { return models.CoinPhotoSourceMDP }
func (s *MdpCoinPhotoService) Credit() string  { return "Monnaie de Paris" }

// Thumbnail asks the shop for a smaller copy of a photo.
func (s *MdpCoinPhotoService) Thumbnail(image string) string {
	return image + mdpThumbnailOptions
}

// Coins returns the shop's coins, reading the sitemaps at most once a week.
func (s *MdpCoinPhotoService) Coins(ctx context.Context) ([]mintCoin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.coins != nil && s.now().Sub(s.loadedAt) < mdpMemoryLifetime {
		return s.coins, nil
	}
	body, _, _, err := s.cache.Get(ctx, mdpCoinsCacheKey, mdpCoinsCacheTTL, func(ctx context.Context) ([]byte, error) {
		english, err := s.client.Sitemap(ctx, "en")
		if err != nil {
			return nil, err
		}
		// French names are a bonus: without them only English ones match.
		french, _ := s.client.Sitemap(ctx, "fr")
		coins, err := ParseMdpSitemaps(english, french)
		if err != nil {
			return nil, err
		}
		return json.Marshal(coins)
	})
	var coins []mintCoin
	if err == nil {
		err = json.Unmarshal(body, &coins)
	}
	if err != nil {
		// Try the shop again in a few minutes, not on every search.
		s.coins, s.loadedAt = []mintCoin{}, s.now().Add(mdpRetryDelay-mdpMemoryLifetime)
		return nil, err
	}
	s.coins, s.loadedAt = s.withProductDetails(coins), s.now()
	return s.coins, nil
}

// mdpProductDetails is what a product page says of its coin: the month it
// goes on sale ("2026-12") and how many are struck; empty when unknown.
type mdpProductDetails struct {
	Released string `json:"released,omitempty"`
	Volume   int64  `json:"volume,omitempty"`
}

// withProductDetails adds to this year's coins and later ones what their
// product pages said when read; the pages not read yet, or read long ago,
// are read in the background (the coins are read again once they are).
func (s *MdpCoinPhotoService) withProductDetails(coins []mintCoin) []mintCoin {
	var unread []string
	for i := range coins {
		if coins[i].Year < s.now().Year() {
			continue
		}
		entry := s.cache.find(mdpProductCacheKey + mdpSlug(coins[i].PageURL))
		var details mdpProductDetails
		if entry == nil || json.Unmarshal([]byte(entry.Payload), &details) != nil || s.now().Sub(entry.FetchedAt) >= mdpProductTTL(details) {
			unread = append(unread, coins[i].PageURL)
		}
		coins[i].Released, coins[i].Volume = details.Released, details.Volume
	}
	if len(unread) > 0 && !s.reading {
		s.reading = true
		s.reads.Add(1)
		go s.readProductPages(unread)
	}
	return coins
}

// mdpProductTTL keeps what a page said for two weeks; a page that said
// nothing is read again the next day.
func mdpProductTTL(details mdpProductDetails) time.Duration {
	if details == (mdpProductDetails{}) {
		return mdpProductRetryTTL
	}
	return mdpProductCacheTTL
}

// readProductPages reads product pages one at a time, then has the coins
// read again with them.
func (s *MdpCoinPhotoService) readProductPages(pages []string) {
	defer s.reads.Done()
	for i, page := range pages {
		if i > 0 {
			time.Sleep(s.productDelay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), mdpTimeout)
		_, _, _, _ = s.cache.GetFor(ctx, mdpProductCacheKey+mdpSlug(page), func(payload []byte) time.Duration {
			var details mdpProductDetails
			_ = json.Unmarshal(payload, &details)
			return mdpProductTTL(details)
		}, func(ctx context.Context) ([]byte, error) {
			// A page that cannot be read says nothing, till the next day.
			body, _ := s.client.Product(ctx, page)
			return json.Marshal(ParseMdpProduct(body))
		})
		cancel()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reading, s.loadedAt = false, time.Time{}
}

// mdpSlug is a product page's last part, the product's name.
func mdpSlug(page string) string {
	return page[strings.LastIndex(page, "/")+1:]
}

var (
	// The page's own product, among the others it shows.
	mdpProductSKU = regexp.MustCompile(`catalog_product_view_sku_(\d+)`)
	mdpLaunch     = regexp.MustCompile(`^(\d{4})-(\d{2})-\d{2}`)
	// A coin still to come: "Available on the 02/12/2026", "Disponible le 02/12/2026".
	mdpAvailableOn = regexp.MustCompile(`(?:Available on the|Disponible le)\s+\d{2}/(\d{2})/(\d{4})`)
	mdpTableRow    = regexp.MustCompile(`<th class="col label" scope="row">([^<]*)</th>\s*<td class="col data"[^>]*>([\s\S]*?)</td>`)
	mdpNonDigits   = regexp.MustCompile(`\D`)
)

// ParseMdpProduct reads a product page: the month its coin went on sale (the
// product's launch date) or goes on sale ("Available on the 02/12/2026"), and
// its mintage. A coin still to come shows its year as its mintage: unknown.
func ParseMdpProduct(page []byte) mdpProductDetails {
	var details mdpProductDetails
	rows := map[string]string{}
	for _, row := range mdpTableRow.FindAllSubmatch(page, -1) {
		rows[strings.ToLower(strings.TrimSpace(string(row[1])))] = strings.TrimSpace(html.UnescapeString(mdpTag.ReplaceAllString(string(row[2]), " ")))
	}
	product := mdpProductData(page)
	launch, _ := product["date_lancement"].(string)
	if match := mdpLaunch.FindStringSubmatch(launch); match != nil {
		details.Released = match[1] + "-" + match[2]
	} else if match := mdpAvailableOn.FindSubmatch(page); match != nil {
		details.Released = string(match[2]) + "-" + string(match[1])
	}
	mintage, _ := product["tirage"].(string)
	if mintage == "" {
		mintage = rows["mintage"]
		if mintage == "" {
			mintage = rows["tirage"]
		}
	}
	volume, _ := strconv.ParseInt(mdpNonDigits.ReplaceAllString(mintage, ""), 10, 64)
	if year, _ := strconv.Atoi(rows["millésime"]); volume > 0 && int64(year) != volume {
		details.Volume = volume
	}
	return details
}

// mdpProductData is the shop's data of the page's own product (named by its
// code, "ean128"), from the page's scripts; nil when the page has none.
func mdpProductData(page []byte) map[string]any {
	sku := mdpProductSKU.FindSubmatch(page)
	if sku == nil {
		return nil
	}
	at := bytes.Index(page, []byte(`"ean128":"`+string(sku[1])+`"`))
	// The object around it: the nearest opening brace before it that reads
	// as an object naming that code.
	for start := at; at >= 0 && start >= 0 && at-start < 64<<10; start-- {
		if page[start] != '{' {
			continue
		}
		var fields map[string]any
		if json.NewDecoder(bytes.NewReader(page[start:])).Decode(&fields) == nil && fields["ean128"] == string(sku[1]) {
			return fields
		}
	}
	return nil
}

var mdpTag = regexp.MustCompile(`<[^>]*>`)

// Match returns the shop's photos of the coin, or nil when no coin of the
// shop is clearly it. The reverse is only given when the photo is named so;
// Confirm checks the photo taken after the design.
func (s *MdpCoinPhotoService) Match(ctx context.Context, query MintPhotoQuery) (*MintPhoto, error) {
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
	return &MintPhoto{Title: coin.Title, Obverse: coin.Obverse, Reverse: coin.Reverse, PageURL: coin.PageURL, reverseGuess: coin.ReverseGuess}, nil
}

// Confirm adds the photo taken after the design as the reverse when it shows
// a coin (and not its packaging).
func (s *MdpCoinPhotoService) Confirm(ctx context.Context, photo *MintPhoto) *MintPhoto {
	if photo == nil || photo.Reverse != "" || photo.reverseGuess == "" {
		return photo
	}
	confirmed := *photo
	body, _, _, err := s.cache.Get(ctx, mdpShapeCacheKey+photo.reverseGuess, mdpShapeCacheTTL, func(ctx context.Context) ([]byte, error) {
		data, err := s.client.Image(ctx, photo.reverseGuess)
		if err != nil {
			return nil, err
		}
		return json.Marshal(LooksLikeCoin(data))
	})
	var isCoin bool
	if err == nil && json.Unmarshal(body, &isCoin) == nil && isCoin {
		confirmed.Reverse = photo.reverseGuess
	}
	return &confirmed
}

type mdpSitemap struct {
	URLs []struct {
		Loc   string `xml:"loc"`
		Links []struct {
			Hreflang string `xml:"hreflang,attr"`
			Href     string `xml:"href,attr"`
		} `xml:"link"`
		Images []struct {
			Loc   string `xml:"loc"`
			Title string `xml:"title"`
		} `xml:"image"`
	} `xml:"url"`
}

var (
	// "10eur", "50eu", "1-4eur" (¼ €), "5-000eur" (5,000 €).
	mdpFaceValue = regexp.MustCompile(`(?:^|-)(?:(1)-(2|4)|(\d{1,2})-(\d{3})|(\d+))eu(?:r|ros?)?(?:-|$)`)
	mdpYear      = regexp.MustCompile(`(?:yeardate|millesime|year|vintage)-(\d{4})(?:-|$)`)
	// A coin's slug ends with its year (and a number or market code, "-jp");
	// an object holding a coin goes on ("…-yeardate-2017-steel-letter").
	mdpCoinEnd = regexp.MustCompile(`(?:yeardate|millesime|year|vintage)-\d{4}(?:-(?:\d+|[a-z]{2}))?$`)
	mdpPhotoNo = regexp.MustCompile(`MDP_(\d{6,})_`)
	// Several coins in one product: a set, an album, a box.
	mdpSet = regexp.MustCompile(`(?:^|-)(?:set|sets|lot|lots|box|album|albums|coffret|coffrets|pack|collection|coins|monnaies|\d+x\d+(?:eu\w*)?)(?:-|$)`)
	// Slug words that say nothing about the design.
	mdpNoise = regexp.MustCompile(`^(?:\d+eu(?:r|ros?)?|\d+(?:oz|g|kg)|eu|eur|euro|euros|coin|coins|monnaie|monnaies|de|of|the|quality|qualite|proof|be|bu|brilliant|uncirculated|unciculated|circulating|courante|historic|historique|yeardate|millesime|year|vintage|gold|silver|or|argent|common|metal|metal|commun|high|relief|haut|colorized|colorised|colorisee|antique|finish|finition|patinee|shape|kilo|\d+)$`)
)

// ParseMdpSitemaps reads the English sitemap's coin products, with their
// French names from the French one.
func ParseMdpSitemaps(english, french []byte) ([]mintCoin, error) {
	var en mdpSitemap
	if err := xml.Unmarshal(english, &en); err != nil {
		return nil, err
	}
	frenchTitles := map[string]string{}
	var fr mdpSitemap
	if len(french) > 0 && xml.Unmarshal(french, &fr) == nil {
		for _, url := range fr.URLs {
			if len(url.Images) > 0 {
				frenchTitles[url.Loc] = url.Images[0].Title
			}
		}
	}
	coins := []mintCoin{}
	for _, url := range en.URLs {
		if len(url.Images) == 0 {
			continue
		}
		slug := url.Loc[strings.LastIndex(url.Loc, "/")+1:]
		value, year := mdpValue(slug), mdpYearOf(slug)
		if value <= 0 || year == 0 || !strings.Contains(slug, "coin") || !mdpCoinEnd.MatchString(slug) || mdpSet.MatchString(slug) || len(mdpFaceValue.FindAllString(slug, -1)) > 1 {
			continue
		}
		coin := mintCoin{
			Title:     incmSpaces(url.Images[0].Title),
			DesignKey: mdpDesignWords(slug),
			PageURL:   url.Loc,
			Value:     value,
			Metal:     mdpMetal(slug),
			Variant:   mintVariant(mdpMetal(slug), mdpFinish(slug)),
			Year:      year,
			Obverse:   mdpPhotoURL(url.Images[0].Loc),
		}
		coin.OtherTitles = []string{coin.DesignKey}
		isSet := false
		for _, link := range url.Links {
			if link.Hreflang == "fr-FR" {
				frenchSlug := link.Href[strings.LastIndex(link.Href, "/")+1:]
				isSet = mdpSet.MatchString(frenchSlug)
				coin.OtherTitles = append(coin.OtherTitles, mdpDesignWords(frenchSlug))
				if title := frenchTitles[link.Href]; title != "" {
					coin.OtherTitles = append(coin.OtherTitles, incmSpaces(title))
				}
			}
		}
		first := mdpPhotoNo.FindStringSubmatch(coin.Obverse)
		for _, image := range url.Images[1:] {
			photo := mdpPhotoURL(image.Loc)
			name := strings.ToLower(photo[strings.LastIndex(photo, "/")+1:])
			if strings.Contains(name, "revers") {
				coin.Reverse, coin.ReverseGuess = photo, ""
				break
			}
			if next := mdpPhotoNo.FindStringSubmatch(photo); coin.ReverseGuess == "" && first != nil && next != nil {
				a, _ := strconv.Atoi(first[1])
				b, _ := strconv.Atoi(next[1])
				if b == a+1 {
					coin.ReverseGuess = photo
				}
			}
		}
		if coin.Obverse != "" && !isSet {
			coins = append(coins, coin)
		}
	}
	return coins, nil
}

func mdpValue(slug string) float64 {
	match := mdpFaceValue.FindStringSubmatch(slug)
	switch {
	case match == nil:
		return 0
	case match[1] != "":
		denominator, _ := strconv.ParseFloat(match[2], 64)
		return 1 / denominator
	case match[3] != "":
		value, _ := strconv.ParseFloat(match[3]+match[4], 64)
		return value
	default:
		value, _ := strconv.ParseFloat(match[5], 64)
		return value
	}
}

func mdpYearOf(slug string) int {
	matches := mdpYear.FindAllStringSubmatch(slug, -1)
	if len(matches) == 0 {
		return 0
	}
	year, _ := strconv.Atoi(matches[len(matches)-1][1])
	return year
}

// mdpFinish is a product's finish, from its name's quality: proof ("Belle
// Épreuve"), brilliant uncirculated (BNC) or FDC; "" for the usual one.
func mdpFinish(slug string) string {
	switch {
	case strings.Contains(slug, "proof-quality"):
		return mintFinishProof
	case strings.Contains(slug, "-bu-quality") || strings.Contains(slug, "brilliant-uncirculated"):
		return mintFinishBNC
	case strings.Contains(slug, "fdc-quality"):
		return mintFinishFDC
	}
	return ""
}

func mdpMetal(slug string) string {
	switch {
	case strings.Contains(slug, "gold") || strings.Contains(slug, "-or-"):
		return "gold"
	case strings.Contains(slug, "silver") || strings.Contains(slug, "argent"):
		return "silver"
	default:
		return "common metal"
	}
}

// mdpDesignWords keeps the words of a slug that name the design, once each
// ("snoopy-10eur-silver-coin-snoopy-shape…" is "snoopy").
func mdpDesignWords(slug string) string {
	var words []string
	for _, word := range strings.Split(slug, "-") {
		if word != "" && !mdpNoise.MatchString(word) {
			words = append(words, word)
		}
	}
	return strings.Join(uniqueStrings(words), " ")
}

// mdpPhotoURL drops the resizing options from a photo's address.
func mdpPhotoURL(raw string) string {
	photo, _, _ := strings.Cut(strings.TrimSpace(raw), "?")
	if !strings.HasPrefix(photo, mdpShopURL+MdpImagePath) {
		return ""
	}
	return photo
}
