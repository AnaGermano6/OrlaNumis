package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Catalogue data changes rarely; searches are kept for a day so repeating a
// search while adding several coins costs nothing.
const (
	numistaIssuersCacheTTL = 30 * 24 * time.Hour
	numistaTypeCacheTTL    = 30 * 24 * time.Hour
	numistaSearchCacheTTL  = 24 * time.Hour
	NumistaCatalogPageSize = 20
	// NumistaCatalogMaxPageSize is the most types Numista returns per page.
	NumistaCatalogMaxPageSize = 50
	numistaIssuerMatchLimit   = 20
)

// numistaCatalogueIDs maps reference catalogue codes, as collectors write them
// ("KM# 84"), to Numista's catalogue IDs (from GET /catalogues).
var numistaCatalogueIDs = map[string]int{"KM": 3, "Y": 9}

// NumistaCatalogueID returns Numista's ID for a catalogue code such as "KM".
func NumistaCatalogueID(code string) (int, bool) {
	id, ok := numistaCatalogueIDs[strings.ToUpper(strings.TrimSpace(code))]
	return id, ok
}

// NumistaCatalogClient is the part of the Numista API the catalogue browser uses.
type NumistaCatalogClient interface {
	SearchTypes(ctx context.Context, search NumistaTypeSearch) ([]byte, error)
	TypeDetail(ctx context.Context, typeID int, lang string) ([]byte, error)
	Issuers(ctx context.Context, lang string) ([]byte, error)
	Issues(ctx context.Context, typeID int) ([]byte, error)
}

// NumistaCatalogService searches the Numista catalogue and returns the details
// used to add a coin with everything filled in.
type NumistaCatalogService struct {
	client NumistaCatalogClient
	cache  *NumistaResponseCache

	// ecb adds the ECB's image of a €2 commemorative to search results and types.
	ecb *EcbCoinPhotoService
	// mints add the national mints' photos of the collector coins they sell.
	mints []MintPhotoSource

	// The issuer list is ~1 MB even trimmed; keep the parsed copy per language
	// for as long as the stored payload is unchanged.
	issuersMu   sync.Mutex
	issuersMemo map[string]parsedIssuers
}

type parsedIssuers struct {
	fetchedAt time.Time
	issuers   []models.NumistaCatalogIssuer
	folded    []string
}

// NewNumistaCatalogService creates the service.
func NewNumistaCatalogService(client NumistaCatalogClient, cache *NumistaResponseCache) *NumistaCatalogService {
	return &NumistaCatalogService{client: client, cache: cache, issuersMemo: map[string]parsedIssuers{}}
}

// Issuers returns up to 20 issuers whose name or code matches query, best
// matches first (exact name, then name prefix, word prefix, anywhere).
func (s *NumistaCatalogService) Issuers(ctx context.Context, lang, query string) models.NumistaCatalogIssuersResult {
	result := models.NumistaCatalogIssuersResult{Status: models.NumistaDataStatusOK, Issuers: []models.NumistaCatalogIssuer{}}
	needle := foldForSearch(query)
	if needle == "" {
		return result
	}
	list, err := s.issuerList(ctx, lang)
	if err != nil {
		result.Status = numistaStatusForError(err)
		return result
	}

	type match struct {
		issuer models.NumistaCatalogIssuer
		rank   int
	}
	var matches []match
	for i, folded := range list.folded {
		issuer := list.issuers[i]
		rank := -1
		switch {
		case folded == needle:
			rank = 0
		case strings.HasPrefix(folded, needle):
			rank = 1
		case strings.Contains(" "+folded, " "+needle):
			rank = 2
		case strings.Contains(folded, needle):
			rank = 3
		case strings.Contains(issuer.Code, needle):
			rank = 4
		}
		if rank >= 0 {
			matches = append(matches, match{issuer: issuer, rank: rank})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].rank != matches[j].rank {
			return matches[i].rank < matches[j].rank
		}
		return len(matches[i].issuer.Name) < len(matches[j].issuer.Name)
	})
	for i := 0; i < len(matches) && i < numistaIssuerMatchLimit; i++ {
		result.Issuers = append(result.Issuers, matches[i].issuer)
	}
	return result
}

func (s *NumistaCatalogService) issuerList(ctx context.Context, lang string) (parsedIssuers, error) {
	body, fetchedAt, _, err := s.cache.Get(ctx, "issuers:"+lang, numistaIssuersCacheTTL, func(ctx context.Context) ([]byte, error) {
		raw, err := s.client.Issuers(ctx, lang)
		if err != nil {
			return nil, err
		}
		return trimNumistaIssuers(raw)
	})
	if err != nil {
		return parsedIssuers{}, err
	}

	s.issuersMu.Lock()
	defer s.issuersMu.Unlock()
	if memo, ok := s.issuersMemo[lang]; ok && memo.fetchedAt.Equal(fetchedAt) {
		return memo, nil
	}
	var issuers []models.NumistaCatalogIssuer
	if err := json.Unmarshal(body, &issuers); err != nil {
		return parsedIssuers{}, &NumistaError{Kind: NumistaErrorMalformedResponse}
	}
	folded := make([]string, len(issuers))
	for i, issuer := range issuers {
		folded[i] = foldForSearch(issuer.Name)
	}
	memo := parsedIssuers{fetchedAt: fetchedAt, issuers: issuers, folded: folded}
	s.issuersMemo[lang] = memo
	return memo, nil
}

// trimNumistaIssuers keeps only code, name and parent name of each issuer, so
// the stored copy is a fraction of the provider's response. Numista groups
// issuers in sections that often share the issuer's name (section "Portugal"
// > issuer "Portugal"); such a section is dropped so a name appears once.
func trimNumistaIssuers(raw []byte) ([]byte, error) {
	var response struct {
		Issuers []struct {
			Code   string `json:"code"`
			Name   string `json:"name"`
			Parent *struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"parent"`
		} `json:"issuers"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, &NumistaError{Kind: NumistaErrorMalformedResponse}
	}
	sameNameSections := map[string]bool{}
	for _, issuer := range response.Issuers {
		if issuer.Parent != nil && strings.EqualFold(strings.TrimSpace(issuer.Parent.Name), strings.TrimSpace(issuer.Name)) {
			sameNameSections[strings.TrimSpace(issuer.Parent.Code)] = true
		}
	}
	trimmed := make([]models.NumistaCatalogIssuer, 0, len(response.Issuers))
	for _, issuer := range response.Issuers {
		code, name := strings.TrimSpace(issuer.Code), strings.TrimSpace(issuer.Name)
		if code == "" || name == "" || sameNameSections[code] {
			continue
		}
		entry := models.NumistaCatalogIssuer{Code: code, Name: name}
		if issuer.Parent != nil && !strings.EqualFold(strings.TrimSpace(issuer.Parent.Name), name) {
			entry.Parent = strings.TrimSpace(issuer.Parent.Name)
		}
		trimmed = append(trimmed, entry)
	}
	return json.Marshal(trimmed)
}

// Search returns one page of coin types matching the filters.
func (s *NumistaCatalogService) Search(ctx context.Context, search NumistaTypeSearch) models.NumistaCatalogSearchResult {
	search.Page = max(search.Page, 1)
	if search.Count <= 0 {
		search.Count = NumistaCatalogPageSize
	}
	search.Count = min(search.Count, NumistaCatalogMaxPageSize)
	search.Number = strings.TrimSpace(search.Number)
	result := models.NumistaCatalogSearchResult{Status: models.NumistaDataStatusOK, Page: search.Page, Types: []models.NumistaCatalogSearchItem{}}

	key := fmt.Sprintf("%s|%s|%d|%d|%s|%d|%s|%d", search.Lang, search.Issuer, search.Year, search.Page, strings.ToLower(strings.TrimSpace(search.Query)), search.Catalogue, strings.ToLower(search.Number), search.Count)
	// Coin searches keep the key they always had, so their cache stays valid.
	if search.Category == models.CoinKindBanknote {
		key += "|banknote"
	} else {
		search.Category = ""
	}
	digest := sha256.Sum256([]byte(key))
	body, _, _, err := s.cache.Get(ctx, "search:"+hex.EncodeToString(digest[:]), numistaSearchCacheTTL, func(ctx context.Context) ([]byte, error) {
		return s.client.SearchTypes(ctx, search)
	})
	if err != nil {
		result.Status = numistaStatusForError(err)
		return result
	}
	var response struct {
		Count int `json:"count"`
		Types []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Issuer struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"issuer"`
			MinYear          *int   `json:"min_year"`
			MaxYear          *int   `json:"max_year"`
			ObverseThumbnail string `json:"obverse_thumbnail"`
			ReverseThumbnail string `json:"reverse_thumbnail"`
		} `json:"types"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		result.Status = models.NumistaDataStatusUnavailable
		return result
	}
	result.Count = response.Count
	for _, item := range response.Types {
		if item.ID <= 0 || strings.TrimSpace(item.Title) == "" {
			continue
		}
		result.Types = append(result.Types, models.NumistaCatalogSearchItem{
			ID:               item.ID,
			Title:            strings.TrimSpace(item.Title),
			Issuer:           models.NumistaCatalogIssuer{Code: item.Issuer.Code, Name: item.Issuer.Name},
			MinYear:          item.MinYear,
			MaxYear:          item.MaxYear,
			ObverseThumbnail: httpsOnly(item.ObverseThumbnail),
			ReverseThumbnail: httpsOnly(item.ReverseThumbnail),
		})
	}
	// Numista's photos are refused to the app; euro coins show the ECB's, and
	// collector coins their mint's (INCM, Monnaie de Paris, Münze Österreich).
	for i := range result.Types {
		item := &result.Types[i]
		obverse, reverse := s.ecbSides(ctx, item.Title, item.Issuer, item.MinYear, item.MaxYear)
		if obverse == "" {
			// A type looked at before is matched on its full details.
			query, ok := mintQueryForTitle(item.Title, item.MinYear)
			if detail := s.cachedTypeDetail(item.ID); detail != nil {
				query, ok = mintQueryForType(detail)
			}
			if mint, photo := s.mintPhoto(ctx, item.Issuer, query, ok); photo != nil {
				obverse, reverse = mint.Thumbnail(photo.Obverse), ""
				if photo.Reverse != "" {
					reverse = mint.Thumbnail(photo.Reverse)
				}
			}
		}
		if obverse != "" {
			item.ObverseThumbnail = obverse
		}
		if reverse != "" {
			item.ReverseThumbnail = reverse
		}
	}
	return result
}

// Type returns a coin type's catalogue details and issues.
func (s *NumistaCatalogService) Type(ctx context.Context, typeID int, lang string) models.NumistaCatalogTypeResult {
	result := models.NumistaCatalogTypeResult{Status: models.NumistaDataStatusOK, Issues: []models.NumistaIssueSummary{}}
	detail, err := s.typeDetail(ctx, typeID, lang)
	if err != nil {
		result.Status = numistaStatusForError(err)
		return result
	}
	// Numista's photos are refused to the app; euro coins show the ECB's
	// instead, and recent collector coins their mint's.
	obverse, reverse := s.ecbSides(ctx, detail.Title, detail.Issuer, detail.MinYear, detail.MaxYear)
	copyright, thumbnail := "European Central Bank", func(image string) string { return image }
	if obverse == "" {
		query, ok := mintQueryForType(detail)
		if mint, photo := s.mintPhoto(ctx, detail.Issuer, query, ok); photo != nil {
			photo = mint.Confirm(ctx, photo)
			obverse, reverse, copyright, thumbnail = photo.Obverse, photo.Reverse, mint.Credit(), mint.Thumbnail
		}
	}
	if obverse != "" || reverse != "" {
		withECB := *detail
		withECB.Photos = nil
		for _, side := range []struct{ name, image string }{{"obverse", obverse}, {"reverse", reverse}} {
			if side.image != "" {
				withECB.Photos = append(withECB.Photos, models.NumistaCatalogPhoto{Side: side.name, Thumbnail: thumbnail(side.image), Picture: side.image, Copyright: copyright})
				continue
			}
			for _, photo := range detail.Photos {
				if photo.Side == side.name {
					withECB.Photos = append(withECB.Photos, photo)
				}
			}
		}
		detail = &withECB
	}
	result.Type = detail
	// Issues are optional for prefilling; a failure leaves the list empty.
	if issues, _, err := cachedNumistaIssues(ctx, s.cache, s.client.Issues, typeID); err == nil {
		result.Issues = issues
	}
	return result
}

// Photos returns the type's catalogue photos, or none when they cannot be loaded.
func (s *NumistaCatalogService) Photos(ctx context.Context, typeID int, lang string) []models.NumistaCatalogPhoto {
	detail, err := s.typeDetail(ctx, typeID, lang)
	if err != nil || detail == nil {
		return []models.NumistaCatalogPhoto{}
	}
	return detail.Photos
}

// WithMint shows a national mint's photos of the collector coins it sells in
// search results and type details.
func (s *NumistaCatalogService) WithMint(mint MintPhotoSource) *NumistaCatalogService {
	s.mints = append(s.mints, mint)
	return s
}

// WithECB shows the ECB's official images of euro coins in search results and
// type details, where Numista's own photos cannot be loaded.
func (s *NumistaCatalogService) WithECB(ecb *EcbCoinPhotoService) *NumistaCatalogService {
	s.ecb = ecb
	return s
}

// ecbSides returns the ECB's images of a euro coin type: its obverse (a €2
// commemorative's design, or the country's regular one) and its common
// reverse; "" for a side the ECB has no image of.
func (s *NumistaCatalogService) ecbSides(ctx context.Context, title string, issuer models.NumistaCatalogIssuer, minYear, maxYear *int) (string, string) {
	if s.ecb == nil || minYear == nil {
		return "", ""
	}
	country := issuer.Code
	if EcbCountryName(country) == "" {
		country = issuer.Name
	}
	var obverse string
	if RegularEuroType(title, minYear, maxYear) {
		obverse, _ = EcbNationalSideImage(title, *minYear, country)
	} else {
		obverse, _ = s.ecbImage(ctx, title, issuer, minYear, maxYear)
	}
	reverse, _ := EcbCommonSideImage(title, *minYear, country)
	return obverse, reverse
}

// ecbImage returns the ECB image of a single-year €2 type, or "".
func (s *NumistaCatalogService) ecbImage(ctx context.Context, title string, issuer models.NumistaCatalogIssuer, minYear, maxYear *int) (string, string) {
	if s.ecb == nil || minYear == nil || maxYear == nil || *minYear != *maxYear || !twoEuroDenomination.MatchString(title) {
		return "", ""
	}
	country := issuer.Code
	if EcbCountryName(country) == "" {
		country = issuer.Name
	}
	photos, err := s.ecb.Match(ctx, EcbPhotoQuery{Year: *minYear, Country: country, Title: title})
	if err != nil || len(photos) != 1 || photos[0].Match != models.CoinPhotoMatchExact {
		return "", ""
	}
	return photos[0].ImageURL, photos[0].PageURL
}

// knownTypeDetail returns a type from the cache in any language, so looking
// up a coin's ECB image costs no Numista request; only an uncached type is fetched.
func (s *NumistaCatalogService) knownTypeDetail(ctx context.Context, typeID int, lang string) (*models.NumistaCatalogType, error) {
	if detail := s.cachedTypeDetail(typeID, lang); detail != nil {
		return detail, nil
	}
	return s.typeDetail(ctx, typeID, lang)
}

// cachedTypeDetail returns a type from the cache, preferring the given
// languages and then Portuguese and English; nil when it is not there.
func (s *NumistaCatalogService) cachedTypeDetail(typeID int, langs ...string) *models.NumistaCatalogType {
	for _, cachedLang := range append(langs, "pt", "en") {
		if entry := s.cache.find(fmt.Sprintf("type:%d:%s", typeID, cachedLang)); entry != nil {
			if detail, err := parseNumistaCatalogType([]byte(entry.Payload)); err == nil && detail.ID == typeID {
				return detail
			}
		}
	}
	return nil
}

// NumistaTypeFacts is what a cached Numista type says of its coins.
type NumistaTypeFacts struct {
	// Demonetized is nil when Numista does not say; DemonetizationDate is
	// when, if known.
	Demonetized        *bool
	DemonetizationDate string
	// Commemorative types celebrate something: CommemoratedTopic, as Numista
	// names it.
	Commemorative     bool
	CommemoratedTopic string
	// Classification is Numista's kind of coin, as written in the cache.
	Classification string
}

// CachedTypeFacts tells what a type Numista already described (and the app
// cached) says of its coins; ok is false when it is not cached. It never asks
// Numista.
func (s *NumistaCatalogService) CachedTypeFacts(typeID int) (facts NumistaTypeFacts, ok bool) {
	detail := s.cachedTypeDetail(typeID)
	if detail == nil {
		return NumistaTypeFacts{}, false
	}
	return NumistaTypeFacts{
		Demonetized:        detail.Demonetized,
		DemonetizationDate: detail.DemonetizationDate,
		Commemorative:      strings.TrimSpace(detail.CommemoratedTopic) != "",
		CommemoratedTopic:  strings.TrimSpace(detail.CommemoratedTopic),
		Classification:     detail.Classification,
	}, true
}

// mintPhoto returns the photos of a coin by the mint of its issuer, or nil.
func (s *NumistaCatalogService) mintPhoto(ctx context.Context, issuer models.NumistaCatalogIssuer, query MintPhotoQuery, ok bool) (MintPhotoSource, *MintPhoto) {
	mint := mintSourceFor(s.mints, issuer.Code, issuer.Name)
	if mint == nil || !ok {
		return nil, nil
	}
	photo, _ := mint.Match(ctx, query)
	return mint, photo
}

func (s *NumistaCatalogService) typeDetail(ctx context.Context, typeID int, lang string) (*models.NumistaCatalogType, error) {
	body, _, _, err := s.cache.Get(ctx, fmt.Sprintf("type:%d:%s", typeID, lang), numistaTypeCacheTTL, func(ctx context.Context) ([]byte, error) {
		return s.client.TypeDetail(ctx, typeID, lang)
	})
	if err != nil {
		return nil, err
	}
	detail, err := parseNumistaCatalogType(body)
	if err != nil || detail.ID != typeID {
		return nil, &NumistaError{Kind: NumistaErrorMalformedResponse}
	}
	return detail, nil
}

type providerCatalogSide struct {
	Description        string   `json:"description"`
	Lettering          string   `json:"lettering"`
	Designers          []string `json:"designers"`
	Engravers          []string `json:"engravers"`
	Picture            string   `json:"picture"`
	Thumbnail          string   `json:"thumbnail"`
	PictureCopyright   string   `json:"picture_copyright"`
	PictureLicenseName string   `json:"picture_license_name"`
}

type providerCatalogType struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Issuer struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"issuer"`
	MinYear *int `json:"min_year"`
	MaxYear *int `json:"max_year"`
	Value   struct {
		Text         string   `json:"text"`
		NumericValue *float64 `json:"numeric_value"`
		Currency     struct {
			Name string `json:"name"`
		} `json:"currency"`
	} `json:"value"`
	Ruler []struct {
		Name string `json:"name"`
	} `json:"ruler"`
	Composition struct {
		Text string `json:"text"`
	} `json:"composition"`
	Technique struct {
		Text string `json:"text"`
	} `json:"technique"`
	Weight *float64 `json:"weight"`
	Size   *float64 `json:"size"`
	// Size2 is a banknote's height (Size its width).
	Size2     *float64 `json:"size2"`
	Thickness *float64 `json:"thickness"`
	Shape     string   `json:"shape"`
	Mints     []struct {
		Name string `json:"name"`
	} `json:"mints"`
	// Banknotes: who printed them, their watermark and the issuing bank.
	Printers []struct {
		Name string `json:"name"`
	} `json:"printers"`
	Watermark     providerCatalogSide `json:"watermark"`
	IssuingEntity struct {
		Name string `json:"name"`
	} `json:"issuing_entity"`
	Category          string `json:"category"`
	Series            string `json:"series"`
	Type              string `json:"type"`
	Comments          string `json:"comments"`
	CommemoratedTopic string `json:"commemorated_topic"`
	Demonetization    struct {
		IsDemonetized *bool  `json:"is_demonetized"`
		Date          string `json:"demonetization_date"`
	} `json:"demonetization"`
	Obverse    providerCatalogSide `json:"obverse"`
	Reverse    providerCatalogSide `json:"reverse"`
	Edge       providerCatalogSide `json:"edge"`
	References []struct {
		Catalogue struct {
			Code string `json:"code"`
		} `json:"catalogue"`
		Number string `json:"number"`
	} `json:"references"`
}

func parseNumistaCatalogType(body []byte) (*models.NumistaCatalogType, error) {
	var raw providerCatalogType
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	canonical, err := models.CanonicalNumistaURL(raw.ID)
	if err != nil {
		return nil, err
	}
	detail := &models.NumistaCatalogType{
		ID:                 raw.ID,
		URL:                canonical,
		Title:              strings.TrimSpace(raw.Title),
		Issuer:             models.NumistaCatalogIssuer{Code: raw.Issuer.Code, Name: strings.TrimSpace(raw.Issuer.Name)},
		MinYear:            raw.MinYear,
		MaxYear:            raw.MaxYear,
		Denomination:       strings.TrimSpace(raw.Value.Text),
		FaceValue:          positive(raw.Value.NumericValue),
		Currency:           strings.TrimSpace(raw.Value.Currency.Name),
		CommemoratedTopic:  strings.TrimSpace(raw.CommemoratedTopic),
		Classification:     strings.TrimSpace(raw.Type),
		IssuedInCoincard:   IssuedInCoincard(raw.Comments),
		Demonetized:        raw.Demonetization.IsDemonetized,
		DemonetizationDate: strings.TrimSpace(raw.Demonetization.Date),
		Series:             strings.TrimSpace(raw.Series),
		Rulers:             []string{},
		Composition:        strings.TrimSpace(raw.Composition.Text),
		Technique:          strings.TrimSpace(raw.Technique.Text),
		WeightGrams:        positive(raw.Weight),
		DiameterMm:         positive(raw.Size),
		HeightMm:           positive(raw.Size2),
		Printers:           []string{},
		Watermark:          strings.TrimSpace(raw.Watermark.Description),
		IssuingEntity:      strings.TrimSpace(raw.IssuingEntity.Name),
		Category:           strings.TrimSpace(raw.Category),
		ThicknessMm:        positive(raw.Thickness),
		Shape:              strings.TrimSpace(raw.Shape),
		Mints:              []string{},
		Obverse:            catalogSide(raw.Obverse),
		Reverse:            catalogSide(raw.Reverse),
		Edge:               catalogSide(raw.Edge),
		References:         []models.NumistaCatalogReference{},
		Photos:             []models.NumistaCatalogPhoto{},
	}
	for _, ruler := range raw.Ruler {
		if name := strings.TrimSpace(ruler.Name); name != "" {
			detail.Rulers = append(detail.Rulers, name)
		}
	}
	for _, mint := range raw.Mints {
		if name := strings.TrimSpace(mint.Name); name != "" {
			detail.Mints = append(detail.Mints, name)
		}
	}
	for _, printer := range raw.Printers {
		if name := strings.TrimSpace(printer.Name); name != "" {
			detail.Printers = append(detail.Printers, name)
		}
	}
	for _, ref := range raw.References {
		code, number := strings.TrimSpace(ref.Catalogue.Code), strings.TrimSpace(ref.Number)
		if code != "" && number != "" {
			detail.References = append(detail.References, models.NumistaCatalogReference{Catalogue: code, Number: number})
		}
	}
	for side, source := range map[string]providerCatalogSide{"obverse": raw.Obverse, "reverse": raw.Reverse} {
		if photo, ok := catalogPhoto(side, source); ok {
			detail.Photos = append(detail.Photos, photo)
		}
	}
	sort.Slice(detail.Photos, func(i, j int) bool { return detail.Photos[i].Side < detail.Photos[j].Side })
	return detail, nil
}

func catalogSide(side providerCatalogSide) models.NumistaCatalogSide {
	lettering := strings.ReplaceAll(strings.TrimSpace(side.Lettering), "\r\n", "\n")
	return models.NumistaCatalogSide{Description: strings.TrimSpace(side.Description), Lettering: lettering, Designers: trimmedNames(side.Designers), Engravers: trimmedNames(side.Engravers)}
}

func trimmedNames(names []string) []string {
	var trimmed []string
	for _, name := range names {
		if name = strings.TrimSpace(name); name != "" {
			trimmed = append(trimmed, name)
		}
	}
	return trimmed
}

func catalogPhoto(side string, source providerCatalogSide) (models.NumistaCatalogPhoto, bool) {
	thumbnail := httpsOnly(source.Thumbnail)
	if thumbnail == "" {
		return models.NumistaCatalogPhoto{}, false
	}
	return models.NumistaCatalogPhoto{
		Side:      side,
		Thumbnail: thumbnail,
		Picture:   httpsOnly(source.Picture),
		Copyright: strings.TrimSpace(source.PictureCopyright),
		License:   strings.TrimSpace(source.PictureLicenseName),
	}, true
}

// cachedNumistaIssues returns a type's issues through the shared cache.
func cachedNumistaIssues(ctx context.Context, cache *NumistaResponseCache, load func(context.Context, int) ([]byte, error), typeID int) ([]models.NumistaIssueSummary, bool, error) {
	body, _, stale, err := cache.Get(ctx, fmt.Sprintf("issues:%d", typeID), numistaIssuesCacheTTL, func(ctx context.Context) ([]byte, error) {
		return load(ctx, typeID)
	})
	if err != nil {
		return nil, false, err
	}
	issues, err := parseNumistaIssues(body)
	if err != nil {
		return nil, false, &NumistaError{Kind: NumistaErrorMalformedResponse}
	}
	return issues, stale, nil
}

func httpsOnly(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "https://") {
		return value
	}
	return ""
}

func positive(value *float64) *float64 {
	if value == nil || *value <= 0 {
		return nil
	}
	return value
}

var searchFolder = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a", "å", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o", "ø", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "ñ", "n", "ý", "y", "ÿ", "y",
)

// foldForSearch lower-cases and strips common Latin accents so "Sao Tome"
// finds "São Tomé".
func foldForSearch(value string) string {
	return strings.Join(strings.Fields(searchFolder.Replace(strings.ToLower(value))), " ")
}
