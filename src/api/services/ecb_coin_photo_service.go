package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/briandenicola/ancient-coins-api/models"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// The ECB publishes an official image of every €2 commemorative coin, one page
// per year, with the issuing country and the coin's feature in English.
const (
	ecbBaseURL               = "https://www.ecb.europa.eu"
	ecbCommemorativePagePath = "/euro/coins/comm/html/comm_%d.en.html"
	ecbPageLimit             = 3 << 20
	ecbPageTimeout           = 15 * time.Second
	// Past years do not change; the current year's page grows as coins are issued.
	ecbPastYearCacheTTL    = 90 * 24 * time.Hour
	ecbCurrentYearCacheTTL = 7 * 24 * time.Hour
	// The first €2 commemorative coins were issued in 2004.
	ecbFirstCommemorativeYear = 2004
)

// EcbCoinPhoto is a €2 commemorative coin on the ECB site.
type EcbCoinPhoto struct {
	Year     int    `json:"year"`
	Country  string `json:"country"`
	Feature  string `json:"feature"`
	ImageURL string `json:"imageUrl"`
	// IssueDate is when the coin is issued, as the ECB writes it ("June
	// 2025", "Fourth quarter of 2025"); Volume how many are struck.
	IssueDate string `json:"issueDate,omitempty"`
	Volume    int64  `json:"volume,omitempty"`
}

// ECBCoinsClient fetches ECB coin pages.
type ECBCoinsClient interface {
	CommemorativePage(ctx context.Context, year int) ([]byte, error)
}

// HTTPECBCoinsClient reads the ECB website.
type HTTPECBCoinsClient struct {
	client  *http.Client
	baseURL string
}

// NewHTTPECBCoinsClient creates the client; a nil client gets a default one.
func NewHTTPECBCoinsClient(client *http.Client) *HTTPECBCoinsClient {
	if client == nil {
		client = &http.Client{Timeout: ecbPageTimeout}
	}
	return &HTTPECBCoinsClient{client: client, baseURL: ecbBaseURL}
}

// CommemorativePage returns the HTML of a year's €2 commemorative coins page.
func (c *HTTPECBCoinsClient) CommemorativePage(ctx context.Context, year int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+fmt.Sprintf(ecbCommemorativePagePath, year), nil)
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
		return nil, fmt.Errorf("ecb page %d: status %d", year, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, ecbPageLimit))
}

// EcbCoinPhotoService finds the ECB image of a €2 commemorative coin.
type EcbCoinPhotoService struct {
	client ECBCoinsClient
	cache  *NumistaResponseCache
	now    func() time.Time
}

// NewEcbCoinPhotoService creates the service. Parsed pages are kept in the
// shared response cache.
func NewEcbCoinPhotoService(client ECBCoinsClient, cache *NumistaResponseCache) *EcbCoinPhotoService {
	return &EcbCoinPhotoService{client: client, cache: cache, now: time.Now}
}

// Commemoratives lists a year's €2 commemorative coins with their images.
func (s *EcbCoinPhotoService) Commemoratives(ctx context.Context, year int) ([]EcbCoinPhoto, error) {
	if year < ecbFirstCommemorativeYear || year > s.now().Year() {
		return nil, nil
	}
	ttl := ecbPastYearCacheTTL
	if year == s.now().Year() {
		ttl = ecbCurrentYearCacheTTL
	}
	body, _, _, err := s.cache.Get(ctx, "ecb:comm:v3:"+strconv.Itoa(year), ttl, func(ctx context.Context) ([]byte, error) {
		page, err := s.client.CommemorativePage(ctx, year)
		if err != nil {
			return nil, err
		}
		return json.Marshal(ParseEcbCommemoratives(page, year))
	})
	if err != nil {
		return nil, err
	}
	var photos []EcbCoinPhoto
	if err := json.Unmarshal(body, &photos); err != nil {
		return nil, err
	}
	return photos, nil
}

var (
	ecbBoxSplit    = regexp.MustCompile(`<div class="box"[^>]*>`)
	ecbImage       = regexp.MustCompile(`(?i)<img src="([^"]+\.(?:jpg|jpeg|png))"`)
	ecbCountry     = regexp.MustCompile(`(?s)<h3>(.*?)</h3>`)
	ecbFeature     = regexp.MustCompile(`(?is)<strong>\s*Feature\s*:?(?:\s|&nbsp;)*</strong>\s*:?(.*?)</p>`)
	ecbIssueDate   = regexp.MustCompile(`(?is)Issuing date\s*:?(.*?)</p>`)
	ecbVolume      = regexp.MustCompile(`(?is)Issuing volume\s*:?(.*?)</p>`)
	ecbNumber      = regexp.MustCompile(`\d[\d\s.,]*`)
	ecbTags        = regexp.MustCompile(`<[^>]+>`)
	ecbNonLetters  = regexp.MustCompile(`[^a-z]+`)
	ecbJointIssues = regexp.MustCompile(`(?i)euro area|joint`)
)

// ParseEcbCommemoratives reads a year page: each box holds the coin image(s),
// the issuing country (or "Euro area countries" for a joint issue, whose images
// are named after each country) and the feature. A joint-issue image whose
// file does not name its country is kept with no country, so a coin of that
// joint issue is never taken for another coin of its country.
func ParseEcbCommemoratives(page []byte, year int) []EcbCoinPhoto {
	pageURL := ecbBaseURL + fmt.Sprintf(ecbCommemorativePagePath, year)
	base := pageURL[:strings.LastIndex(pageURL, "/")+1]
	imageURL := func(src string) string {
		switch {
		case strings.HasPrefix(src, "https://"):
			return src
		case strings.HasPrefix(src, "/"):
			return ecbBaseURL + src
		}
		return base + src
	}
	var photos []EcbCoinPhoto
	for _, box := range ecbBoxSplit.Split(string(page), -1)[1:] {
		heading := ecbCountry.FindStringSubmatch(box)
		feature := ecbFeature.FindStringSubmatch(box)
		if heading == nil || feature == nil {
			continue
		}
		country := cleanECBText(heading[1])
		issueDate, volume := "", int64(0)
		if match := ecbIssueDate.FindStringSubmatch(box); match != nil {
			issueDate = strings.TrimSpace(strings.TrimLeft(cleanECBText(match[1]), ": "))
		}
		if match := ecbVolume.FindStringSubmatch(box); match != nil {
			volume = ecbCoinCount(cleanECBText(match[1]))
		}
		joint := ecbJointIssues.MatchString(country) || EcbCountryName(country) == ""
		seen := map[string]bool{}
		for _, image := range ecbImage.FindAllStringSubmatch(box, -1) {
			coinCountry := EcbCountryName(country)
			if joint {
				coinCountry = ecbCountryFromFile(image[1])
			}
			// One image per country: a second view of the same coin adds nothing.
			if (coinCountry == "" && !joint) || (coinCountry != "" && seen[coinCountry]) {
				continue
			}
			seen[coinCountry] = true
			photos = append(photos, EcbCoinPhoto{
				Year:      year,
				Country:   coinCountry,
				Feature:   cleanECBText(feature[1]),
				ImageURL:  imageURL(image[1]),
				IssueDate: issueDate,
				Volume:    volume,
			})
		}
	}
	return photos
}

// ecbCoinCount reads an issuing volume: "20 000 000 coins", "1.5 million
// coins"; 0 when it says no number.
func ecbCoinCount(text string) int64 {
	number := ecbNumber.FindString(text)
	if number == "" {
		return 0
	}
	if strings.Contains(strings.ToLower(text), "million") {
		value, err := strconv.ParseFloat(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(number), " ", ""), ",", "."), 64)
		if err != nil {
			return 0
		}
		return int64(value * 1e6)
	}
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, number)
	count, _ := strconv.ParseInt(digits, 10, 64)
	return count
}

func cleanECBText(value string) string {
	text := html.UnescapeString(ecbTags.ReplaceAllString(value, " "))
	return strings.Join(strings.Fields(strings.ReplaceAll(text, " ", " ")), " ")
}

// Euro area issuers: the ECB's English name, then other names a coin's
// country may be written with (Portuguese, Numista issuer codes).
var ecbCountries = [][]string{
	{"Andorra", "andorre"},
	{"Austria", "austria", "autriche"},
	{"Belgium", "belgica", "belgique"},
	{"Bulgaria", "bulgaria", "bulgarie"},
	{"Croatia", "croacia", "croatie"},
	{"Cyprus", "chipre", "chypre"},
	{"Estonia", "estonie"},
	{"Finland", "finlandia", "finlande"},
	{"France", "franca"},
	{"Germany", "alemanha", "allemagne"},
	{"Greece", "grecia", "grece"},
	{"Ireland", "irlanda", "irlande"},
	{"Italy", "italia", "italie"},
	{"Latvia", "letonia", "lettonie"},
	{"Lithuania", "lituania", "lituanie"},
	{"Luxembourg", "luxemburgo"},
	{"Malta", "malte"},
	{"Monaco", "monaco"},
	{"Netherlands", "paises baixos", "holanda", "pays bas", "nederland"},
	{"Portugal"},
	{"San Marino", "sao marinho", "saint marin"},
	{"Slovakia", "eslovaquia", "slovaquie"},
	{"Slovenia", "eslovenia", "slovenie"},
	{"Spain", "espanha", "espagne"},
	{"Vatican City", "vatican city state", "vaticano", "cidade do vaticano", "estado da cidade do vaticano", "vatican", "santa se", "holy see"},
}

func foldASCII(value string) string {
	folded, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), value)
	if err != nil {
		folded = value
	}
	return strings.TrimSpace(strings.ToLower(folded))
}

// EcbCountryName returns the ECB name of a euro area country written in
// English or Portuguese, or as a Numista issuer code; "" for other countries.
// A Numista issuer name such as "Alemanha, República Federal da" counts by its
// first part.
func EcbCountryName(name string) string {
	wanted := foldASCII(name)
	wanted = strings.TrimPrefix(wanted, "the ")
	if head, _, found := strings.Cut(wanted, ","); found {
		wanted = strings.TrimSpace(head)
	}
	wanted = strings.TrimSpace(ecbNonLetters.ReplaceAllString(wanted, " "))
	for _, names := range ecbCountries {
		for _, candidate := range names {
			if strings.TrimSpace(ecbNonLetters.ReplaceAllString(foldASCII(candidate), " ")) == wanted {
				return names[0]
			}
		}
	}
	return ""
}

// ecbCountryFromFile names the country of a joint-issue image ("Portugal.jpg",
// "Austria1.jpg", "joint_comm_2012_San_Marino.jpg", or in the country's own
// language, "joint_comm_2015_Nederland.jpg"); "" when the file name does not say.
func ecbCountryFromFile(path string) string {
	file := strings.ToLower(path[strings.LastIndex(path, "/")+1:])
	if dot := strings.LastIndex(file, "."); dot >= 0 {
		file = file[:dot]
	}
	letters := ecbNonLetters.ReplaceAllString(foldASCII(file), "")
	found, longest := "", 0
	for _, names := range ecbCountries {
		for _, name := range names {
			key := ecbNonLetters.ReplaceAllString(foldASCII(strings.ToLower(name)), "")
			if strings.Contains(letters, key) && len(key) > longest {
				found, longest = names[0], len(key)
			}
		}
	}
	return found
}

// EcbPhotoQuery describes a coin to find on the ECB pages.
type EcbPhotoQuery struct {
	Year    int
	Country string
	// Title is the coin's Numista title (English), e.g. "2 Euros (Rotary)".
	Title string
	// Subject is the coin's own subject, in any language.
	Subject string
}

// Match returns the ECB photo of the coin when one coin of that country and
// year fits, flagged as the match; otherwise every €2 commemorative of that
// country and year, so the user can choose.
func (s *EcbCoinPhotoService) Match(ctx context.Context, query EcbPhotoQuery) ([]models.CoinPhotoCandidate, error) {
	country := EcbCountryName(query.Country)
	if country == "" || query.Year == 0 {
		return nil, nil
	}
	all, err := s.Commemoratives(ctx, query.Year)
	if err != nil {
		return nil, err
	}
	// The coin is one of its country's coins of that year, or a joint-issue
	// coin whose image on the page does not say which country it is.
	var candidates, unknownJoint []EcbCoinPhoto
	// The coin's own year is in many features and tells nothing apart.
	year := strconv.Itoa(query.Year)
	score := func(photo EcbCoinPhoto) int {
		return ecbFeatureScore(photo.Feature, query.Title, year) + ecbFeatureScore(photo.Feature, query.Subject, year)
	}
	ownFeatures := map[string]bool{}
	for _, photo := range all {
		switch photo.Country {
		case country:
			candidates = append(candidates, photo)
			ownFeatures[photo.Feature] = true
		case "":
			unknownJoint = append(unknownJoint, photo)
		}
	}
	// A joint issue whose image for this country is known is not a hiding place.
	unknownJointScore := 0
	hasUnknownJoint := false
	for _, photo := range unknownJoint {
		if !ownFeatures[photo.Feature] {
			hasUnknownJoint = true
			unknownJointScore = max(unknownJointScore, score(photo))
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	best, bestScore, tie := -1, 0, false
	for i, photo := range candidates {
		if s := score(photo); s > bestScore {
			best, bestScore, tie = i, s, false
		} else if s == bestScore && s > 0 {
			tie = true
		}
	}
	switch {
	case bestScore > 0 && bestScore <= unknownJointScore:
		// It reads more like the joint issue, whose image is not identified.
		best = -1
	case bestScore == 0 && len(candidates) == 1 && !hasUnknownJoint:
		// The country's only coin that year, and no coin of it hides elsewhere.
		best, tie = 0, false
	}
	pageURL := ecbBaseURL + fmt.Sprintf(ecbCommemorativePagePath, query.Year)
	toCandidate := func(photo EcbCoinPhoto, match bool) models.CoinPhotoCandidate {
		candidate := models.CoinPhotoCandidate{
			Source:    models.CoinPhotoSourceECB,
			Side:      "obverse",
			Thumbnail: photo.ImageURL,
			ImageURL:  photo.ImageURL,
			PageURL:   pageURL,
			Title:     photo.Feature,
			Author:    "European Central Bank",
		}
		if match {
			candidate.Match = models.CoinPhotoMatchExact
		}
		return candidate
	}
	if best >= 0 && !tie {
		return []models.CoinPhotoCandidate{toCandidate(candidates[best], true)}, nil
	}
	result := make([]models.CoinPhotoCandidate, 0, len(candidates))
	for _, photo := range candidates {
		result = append(result, toCandidate(photo, false))
	}
	return result, nil
}

var ecbStopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "euro": true, "euros": true, "years": true, "year": true, "anniversary": true,
	"dos": true, "das": true, "com": true, "por": true, "para": true, "anos": true, "aniversario": true, "sem": true, "marca": true,
}

func ecbWords(text string) []string {
	var words []string
	for _, word := range strings.FieldsFunc(foldASCII(text), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		numeric := strings.IndexFunc(word, func(r rune) bool { return !unicode.IsDigit(r) }) < 0
		if ecbStopWords[word] || (numeric && len(word) < 2) || (!numeric && len(word) < 3) {
			continue
		}
		words = append(words, word)
	}
	return words
}

// ecbFeatureScore counts the words of text found, loosely, in an ECB feature,
// leaving out ignore (the coin's year).
func ecbFeatureScore(feature, text, ignore string) int {
	featureWords := ecbWords(feature)
	score := 0
	for _, word := range ecbWords(text) {
		if word == ignore {
			continue
		}
		for _, candidate := range featureWords {
			if ecbSimilarWords(word, candidate) {
				score++
				break
			}
		}
	}
	return score
}

func ecbSimilarWords(a, b string) bool {
	if a == b {
		return true
	}
	if a[0] >= '0' && a[0] <= '9' || b[0] >= '0' && b[0] <= '9' {
		return false
	}
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	diff := len(a) - len(b)
	if diff < 0 {
		diff = -diff
	}
	return prefix >= 5 && diff <= 3
}
