package services

import (
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

// The Münze Österreich, Austria's mint, keeps a catalogue of every euro
// collector coin it has issued since 2002, one page per face value: each
// entry gives the coin's year, German name, series and a picture of it. The
// coins still sold have a page in its shop (listed in its sitemap), named
// "<value>-euro-<metal>muenze-<name>-<finish>", with a photo of each side:
// "…_VS…" is the value side (Numista's obverse), "…_RS…" the design. The
// catalogue's picture often shows both sides, or several versions, at once
// (cut off in a coin's frame), so it is never used: coins the shop no longer
// sells have no photo, but stay in the list so their words still tell the
// other designs apart. The coins of this year and the last that it sells
// get the photos of their shop page in the list, for the new releases.
const (
	moeShopURL          = "https://www.muenzeoesterreich.at"
	moeCataloguePath    = "/sammeln/fuer-sammler/muenzkatalog/%s-euro-muenzen"
	moeSitemapPath      = "/sitemap.xml"
	moePageLimit        = 8 << 20
	moeTimeout          = 30 * time.Second
	moeCoinsCacheKey    = "moe:coins:v3"
	moeCoinsCacheTTL    = 7 * 24 * time.Hour
	moeProductCacheKey  = "moe:product:v1:"
	moeProductCacheTTL  = 30 * 24 * time.Hour
	moeMemoryLifetime   = time.Hour
	moeRetryDelay       = 10 * time.Minute
	moeCatalogueTimeout = 2 * time.Minute
)

// MoeImageHost and MoeImagePath locate the Münze Österreich's photos.
const (
	MoeImageHost = "www.muenzeoesterreich.at"
	MoeImagePath = "/var/storage/images/"
)

// moeCatalogueValues are the face values of the catalogue's pages.
var moeCatalogueValues = []string{"3", "5", "10", "20", "25", "50", "100"}

// MOEShopClient reads pages of the Münze Österreich's site.
type MOEShopClient interface {
	// Page returns the page at path ("/sitemap.xml", "/produkte/…").
	Page(ctx context.Context, path string) ([]byte, error)
}

// HTTPMOEShopClient reads them over HTTP.
type HTTPMOEShopClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPMOEShopClient creates the client; a nil client gets a default one.
func NewHTTPMOEShopClient(client *http.Client) *HTTPMOEShopClient {
	if client == nil {
		client = &http.Client{Timeout: moeTimeout}
	}
	return &HTTPMOEShopClient{client: client, baseURL: moeShopURL}
}

// Page returns one of the site's pages.
func (c *HTTPMOEShopClient) Page(ctx context.Context, path string) ([]byte, error) {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return nil, fmt.Errorf("not a Münze Österreich page: %s", path)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
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
		return nil, fmt.Errorf("münze österreich %s: status %d", path, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, moePageLimit))
}

// MoeCoinPhotoService finds an Austrian coin's photos at the Münze Österreich.
type MoeCoinPhotoService struct {
	client MOEShopClient
	cache  *NumistaResponseCache
	now    func() time.Time

	mu       sync.Mutex
	coins    []mintCoin
	loadedAt time.Time
}

// NewMoeCoinPhotoService creates the service; the catalogue is kept in the
// shared response cache for a week, each product page's photos for a month.
func NewMoeCoinPhotoService(client MOEShopClient, cache *NumistaResponseCache) *MoeCoinPhotoService {
	return &MoeCoinPhotoService{client: client, cache: cache, now: time.Now}
}

// Country, Source and Credit describe the Münze Österreich as a MintPhotoSource.
func (s *MoeCoinPhotoService) Country() string { return "Austria" }
func (s *MoeCoinPhotoService) Source() string  { return models.CoinPhotoSourceMOE }
func (s *MoeCoinPhotoService) Credit() string  { return "Münze Österreich" }

// Thumbnail asks the site for its smaller copy of a photo.
func (s *MoeCoinPhotoService) Thumbnail(image string) string {
	return strings.NewReplacer("/_aliases/product_full/", "/_aliases/product_preview/", "/_aliases/large/", "/_aliases/small/").Replace(image)
}

// Coins returns the catalogue's coins, reading the site at most once a week.
func (s *MoeCoinPhotoService) Coins(ctx context.Context) ([]mintCoin, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.coins != nil && s.now().Sub(s.loadedAt) < moeMemoryLifetime {
		return s.coins, nil
	}
	body, _, _, err := s.cache.Get(ctx, moeCoinsCacheKey, moeCoinsCacheTTL, func(ctx context.Context) ([]byte, error) {
		ctx, cancel := context.WithTimeout(ctx, moeCatalogueTimeout)
		defer cancel()
		// Without the shop's products only the catalogue's pictures are known.
		sitemap, _ := s.client.Page(ctx, moeSitemapPath)
		products := ParseMoeSitemap(sitemap)
		coins := []mintCoin{}
		for _, value := range moeCatalogueValues {
			page, err := s.client.Page(ctx, fmt.Sprintf(moeCataloguePath, value))
			if err != nil {
				return nil, err
			}
			coins = append(coins, ParseMoeCatalogue(value, page, products)...)
		}
		return json.Marshal(coins)
	})
	var coins []mintCoin
	if err == nil {
		err = json.Unmarshal(body, &coins)
	}
	if err != nil {
		// Try the site again in a few minutes, not on every search.
		s.coins, s.loadedAt = []mintCoin{}, s.now().Add(moeRetryDelay-moeMemoryLifetime)
		return nil, err
	}
	coins = s.withProductPhotos(ctx, coins)
	s.coins, s.loadedAt = coins, s.now()
	return coins, nil
}

// ReleasePicture shows an Austrian coin by its design (the shop's "RS"):
// its value side is often the same for a whole series, its photo even
// shared between the shop's pages (the €3 coins' is the Troll's).
func (s *MoeCoinPhotoService) ReleasePicture(coin mintCoin) string {
	if coin.Reverse != "" {
		return coin.Reverse
	}
	return coin.Obverse
}

// withProductPhotos gives the coins of this year and the last that the shop
// sells the photos of each side from their page (each page kept a month),
// so that the new releases show them. A page that cannot be read leaves its
// coin without a photo, until the list is read again.
func (s *MoeCoinPhotoService) withProductPhotos(ctx context.Context, coins []mintCoin) []mintCoin {
	since := s.now().Year() - 1
	for i := range coins {
		coin := &coins[i]
		if coin.Obverse != "" || coin.Year < since || !strings.HasPrefix(coin.PageURL, moeShopURL+"/produkte/") {
			continue
		}
		if photo := s.Confirm(ctx, &MintPhoto{Title: coin.Title, PageURL: coin.PageURL}); photo != nil && photo.Obverse != "" {
			coin.Obverse, coin.Reverse = photo.Obverse, photo.Reverse
		}
	}
	return coins
}

// Match returns the catalogue's coin that is clearly the one described, or
// nil; Confirm adds the shop's photos of it.
func (s *MoeCoinPhotoService) Match(ctx context.Context, query MintPhotoQuery) (*MintPhoto, error) {
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

// Confirm takes the photos of each side from the coin's shop page, when it
// has one.
func (s *MoeCoinPhotoService) Confirm(ctx context.Context, photo *MintPhoto) *MintPhoto {
	if photo == nil || !strings.HasPrefix(photo.PageURL, moeShopURL+"/produkte/") {
		return photo
	}
	path := strings.TrimPrefix(photo.PageURL, moeShopURL)
	body, _, _, err := s.cache.Get(ctx, moeProductCacheKey+path, moeProductCacheTTL, func(ctx context.Context) ([]byte, error) {
		page, err := s.client.Page(ctx, path)
		if err != nil {
			return nil, err
		}
		obverse, reverse := ParseMoeProductPhotos(page)
		return json.Marshal([2]string{obverse, reverse})
	})
	var sides [2]string
	if err != nil || json.Unmarshal(body, &sides) != nil || sides[0] == "" {
		return photo
	}
	confirmed := *photo
	confirmed.Obverse, confirmed.Reverse = sides[0], sides[1]
	return &confirmed
}

type moeProduct struct {
	path  string
	metal string
}

var (
	// "5-euro-kupfermuenze-weltmeister-np", "25-euro-silber-niob-muenze-edaphon-hgh";
	// sets ("…-muenzenset-…") are left out.
	moeProductSlug = regexp.MustCompile(`^/produkte/(\d+)-euro-([a-z-]*?)-?muenze-([a-z0-9-]+?)(?:-(?:np|hgh|pp|bu))?$`)
	moeEntry       = regexp.MustCompile(`(?s)<h3>\s*(\d{4})\s+(.*?)\s*</h3>(.*)`)
	moeProductLink = regexp.MustCompile(`href="(?:https://www\.muenzeoesterreich\.at)?(/produkte/[^"#?]+)"`)
	moeImageAlt    = regexp.MustCompile(`(?s)<img [^>]*alt="([^"]*)"`)
	moeTag         = regexp.MustCompile(`<[^>]*>`)
	moeSlugNoise   = regexp.MustCompile(`[^a-z0-9]+`)
	moeProductFull = regexp.MustCompile(`(?:https://www\.muenzeoesterreich\.at)?/var/storage/images/_aliases/product_full/[^"\s,]+`)
)

// ParseMoeSitemap returns the shop's coin products, by face value and name.
func ParseMoeSitemap(sitemap []byte) map[string]map[string][]moeProduct {
	var urls struct {
		Locs []string `xml:"url>loc"`
	}
	products := map[string]map[string][]moeProduct{}
	if len(sitemap) == 0 || xml.Unmarshal(sitemap, &urls) != nil {
		return products
	}
	for _, loc := range urls.Locs {
		path := strings.TrimPrefix(strings.TrimSpace(loc), moeShopURL)
		match := moeProductSlug.FindStringSubmatch(path)
		if match == nil {
			continue
		}
		if products[match[1]] == nil {
			products[match[1]] = map[string][]moeProduct{}
		}
		products[match[1]][match[3]] = append(products[match[1]][match[3]], moeProduct{path: path, metal: moeMetal(match[2])})
	}
	return products
}

var (
	// A product's finish, at the end of its address: "…-pp".
	moeFinish = regexp.MustCompile(`-(pp|hgh|bu|np)$`)
	// "30.000 Polierte Platte (Silber)" in the catalogue's mintages.
	moeVolume = regexp.MustCompile(`(\d{1,3}(?:\.\d{3})*|\d+)\s*(polierte platte|handgehoben|normalpragung)?`)
	// "9. April 2025" (an Austrian January is "Jänner").
	moeDate   = regexp.MustCompile(`(\d{1,2})\.\s*([a-z]+)\s+(\d{4})`)
	moeMonths = map[string]int{"janner": 1, "januar": 1, "feber": 2, "februar": 2, "marz": 3, "april": 4, "mai": 5, "juni": 6,
		"juli": 7, "august": 8, "september": 9, "oktober": 10, "november": 11, "dezember": 12}
	// The finishes of the addresses: proof (Polierte Platte), brilliant
	// uncirculated (handgehoben), and the usual one (Normalprägung).
	moeFinishes = map[string]string{"pp": mintFinishProof, "hgh": mintFinishBNC, "bu": mintFinishBNC, "np": ""}
	// The finishes of the catalogue's mintages, as the addresses name them.
	moeFinishWords = map[string]string{"polierte platte": "pp", "handgehoben": "hgh", "normalpragung": "np"}
)

// moeReleased is the month of an issue date ("9. April 2025" → "2025-04"),
// or "".
func moeReleased(date string) string {
	match := moeDate.FindStringSubmatch(foldASCII(date))
	if match == nil || moeMonths[match[2]] == 0 {
		return ""
	}
	return fmt.Sprintf("%s-%02d", match[3], moeMonths[match[2]])
}

// moeVolumes reads the catalogue's mintages ("30.000 Polierte Platte
// (Silber) 130.000 Normalprägung (Kupfer)"): how many of each finish (by its
// address's name: "pp", "hgh", "np"), and how many in all.
func moeVolumes(text string) (byFinish map[string]int64, total int64) {
	byFinish = map[string]int64{}
	for _, match := range moeVolume.FindAllStringSubmatch(foldASCII(text), -1) {
		volume, _ := strconv.ParseInt(strings.ReplaceAll(match[1], ".", ""), 10, 64)
		if finish := moeFinishWords[match[2]]; finish != "" {
			byFinish[finish] += volume
			if finish == "hgh" {
				byFinish["bu"] += volume
			}
		}
		total += volume
	}
	return byFinish, total
}

func moeMetal(text string) string {
	folded := foldASCII(text)
	switch {
	case strings.Contains(folded, "gold"):
		return "gold"
	case strings.Contains(folded, "silber"):
		return "silver"
	default:
		return "copper"
	}
}

// moeSlug writes a German name the way the shop's addresses do:
// "‘s Glück is a Vogerl" → "s-glueck-is-a-vogerl".
func moeSlug(name string) string {
	name = strings.NewReplacer("ä", "ae", "ö", "oe", "ü", "ue", "Ä", "ae", "Ö", "oe", "Ü", "ue", "ß", "ss", "‘", "", "’", "", "'", "").Replace(name)
	return strings.Trim(moeSlugNoise.ReplaceAllString(foldASCII(name), "-"), "-")
}

// ParseMoeCatalogue reads one face value's catalogue page: a coin for each
// version the shop sells of an entry, or the entry's own picture.
func ParseMoeCatalogue(value string, page []byte, products map[string]map[string][]moeProduct) []mintCoin {
	faceValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	catalogueURL := moeShopURL + fmt.Sprintf(moeCataloguePath, value)
	coins := []mintCoin{}
	for _, item := range strings.Split(string(page), `class="accordion-item"`)[1:] {
		entry := moeEntry.FindStringSubmatch(item)
		if entry == nil {
			continue
		}
		year, _ := strconv.Atoi(entry[1])
		title := strings.Join(strings.Fields(html.UnescapeString(moeTag.ReplaceAllString(entry[2], " "))), " ")
		body := entry[3]
		coin := mintCoin{
			Title:     title,
			DesignKey: mintDesignKey(title),
			Value:     faceValue,
			Year:      year,
			Series:    moeField(body, "Serie"),
			Released:  moeReleased(moeField(body, "Ausgabedatum")),
		}
		volumes, total := moeVolumes(moeField(body, "Auflage"))
		if alt := moeImageAlt.FindStringSubmatch(body); alt != nil {
			coin.Text = html.UnescapeString(alt[1])
		}
		versions := products[value][moeSlug(title)]
		if len(versions) == 0 {
			if link := moeProductLink.FindStringSubmatch(body); link != nil {
				if match := moeProductSlug.FindStringSubmatch(link[1]); match != nil {
					versions = products[value][match[3]]
					if len(versions) == 0 {
						versions = []moeProduct{{path: link[1], metal: moeMetal(match[2])}}
					}
				}
			}
		}
		if len(versions) == 0 {
			// No longer sold: known, but without a photo.
			coin.PageURL, coin.Metal, coin.Volume = catalogueURL, moeMetal(moeField(body, "Legierung")), total
			coins = append(coins, coin)
			continue
		}
		for _, version := range versions {
			sold := coin
			sold.PageURL, sold.Metal = moeShopURL+version.path, version.metal
			finish := moeFinish.FindStringSubmatch(version.path)
			if finish != nil {
				sold.Variant, sold.Volume = mintVariant(version.metal, moeFinishes[finish[1]]), volumes[finish[1]]
			}
			if len(versions) == 1 && sold.Volume == 0 {
				sold.Volume = total
			}
			coins = append(coins, sold)
		}
	}
	return coins
}

// moeField returns the text of a catalogue entry's field ("Serie").
func moeField(body, name string) string {
	match := regexp.MustCompile(`(?s)<strong>` + regexp.QuoteMeta(name) + `</strong></td>\s*<td>(.*?)</td>`).FindStringSubmatch(body)
	if match == nil {
		return ""
	}
	return strings.Join(strings.Fields(html.UnescapeString(moeTag.ReplaceAllString(match[1], " "))), " ")
}

// ParseMoeProductPhotos returns a shop page's photos of the value side
// (obverse) and the design (reverse); "" when it has none.
func ParseMoeProductPhotos(page []byte) (obverse, reverse string) {
	for _, image := range moeProductFull.FindAllString(string(page), -1) {
		image = html.UnescapeString(image)
		if strings.HasPrefix(image, "/") {
			image = moeShopURL + image
		}
		// One side alone, drawn flat: "…_Cu_NP_VS_2D.png", "…_cu_rs.png".
		var side string
		for _, word := range strings.FieldsFunc(strings.ToLower(image[strings.LastIndex(image, "/")+1:]), func(r rune) bool { return r == '_' || r == '-' || r == '.' }) {
			switch {
			case word == "3d" || ((word == "vs" || word == "rs") && side != ""):
				side = "both"
			case word == "vs" || word == "rs":
				side = word
			}
		}
		switch side {
		case "vs":
			if obverse == "" {
				obverse = image
			}
		case "rs":
			if reverse == "" {
				reverse = image
			}
		}
	}
	return obverse, reverse
}
