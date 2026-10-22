package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

// Issues almost never change and prices are community estimates that move
// slowly, so both are reused for a month: browsing a collection of a few
// hundred items, or valuing it day by day, stays inside Numista's free
// monthly quota.
const (
	numistaIssuesCacheTTL = 30 * 24 * time.Hour
	numistaPricesCacheTTL = 30 * 24 * time.Hour
)

// NumistaPriceClient is the part of the Numista API the market value lookup uses.
type NumistaPriceClient interface {
	Issues(ctx context.Context, typeID int) ([]byte, error)
	IssuePrices(ctx context.Context, typeID, issueID int, currency string) ([]byte, error)
}

// NumistaMarketValueService looks up Numista's estimated price per grade for
// a Numista type, or for a coin linked to one.
type NumistaMarketValueService struct {
	client  NumistaPriceClient
	cache   *NumistaResponseCache
	catalog *NumistaCatalogService
}

// NewNumistaMarketValueService creates the service.
func NewNumistaMarketValueService(client NumistaPriceClient, cache *NumistaResponseCache) *NumistaMarketValueService {
	return &NumistaMarketValueService{client: client, cache: cache}
}

// WithCatalog adds the type's catalogue photos to results.
func (s *NumistaMarketValueService) WithCatalog(catalog *NumistaCatalogService) *NumistaMarketValueService {
	s.catalog = catalog
	return s
}

// numistaGradeOrder is Numista's grade scale from lowest to highest.
var numistaGradeOrder = map[string]int{"g": 0, "vg": 1, "f": 2, "vf": 3, "xf": 4, "au": 5, "unc": 6}

// NumistaTypeIDForCoin returns the Numista type a coin is linked to through a
// "Numista" catalogue reference, or 0 when it has none.
func NumistaTypeIDForCoin(coin *models.Coin) int {
	if coin == nil {
		return 0
	}
	for _, ref := range coin.References {
		if !strings.EqualFold(strings.TrimSpace(ref.Catalog), "Numista") {
			continue
		}
		if id, err := strconv.Atoi(strings.TrimSpace(ref.Number)); err == nil && id > 0 && id <= models.NumistaMaxID {
			return id
		}
	}
	return 0
}

// ForCoin returns the market value for a coin's linked Numista type. issueID
// selects an issue (0 = the coin's own issue, else the one matching its date,
// else the first). Provider problems are reported through the result's Status.
func (s *NumistaMarketValueService) ForCoin(ctx context.Context, coin *models.Coin, issueID int, currency, lang string) models.CoinMarketValue {
	typeID := NumistaTypeIDForCoin(coin)
	if typeID == 0 {
		return emptyMarketValue(currency)
	}
	if issueID == 0 {
		issueID = coin.NumistaIssueID
	}
	return s.ForType(ctx, typeID, issueID, coin.DateRange, currency, lang)
}

// ForType returns the market value for a Numista type. dateHint (e.g. "1968"
// or "44 BC") picks the default issue when issueID is 0. lang selects the
// language of catalogue photo credits; empty skips the photos.
func (s *NumistaMarketValueService) ForType(ctx context.Context, typeID, issueID int, dateHint, currency, lang string) models.CoinMarketValue {
	result := emptyMarketValue(currency)
	result.NumistaTypeID = typeID
	result.NumistaURL, _ = models.CanonicalNumistaURL(typeID)
	if s.catalog != nil && lang != "" {
		result.CatalogPhotos = s.catalog.Photos(ctx, typeID, lang)
	}

	issues, issuesStale, err := s.Issues(ctx, typeID)
	if err != nil {
		result.Status = numistaStatusForError(err)
		return result
	}
	result.Issues = issues
	if len(issues) == 0 {
		result.Status = models.NumistaDataStatusNoIssues
		result.Stale = issuesStale
		return result
	}
	selected := chooseNumistaIssue(issues, issueID, dateHint)
	result.SelectedIssueID = selected

	pricesKey := fmt.Sprintf("prices:%d:%d:%s", typeID, selected, currency)
	pricesBody, fetchedAt, pricesStale, err := s.cache.Get(ctx, pricesKey, numistaPricesCacheTTL, func(ctx context.Context) ([]byte, error) {
		return s.client.IssuePrices(ctx, typeID, selected, currency)
	})
	if err != nil {
		result.Status = numistaStatusForError(err)
		result.Stale = issuesStale
		return result
	}
	prices, err := parseNumistaPrices(pricesBody)
	if err != nil {
		result.Status = models.NumistaDataStatusUnavailable
		return result
	}
	result.Prices = prices
	result.FetchedAt = &fetchedAt
	result.Stale = issuesStale || pricesStale
	if len(prices) == 0 {
		result.Status = models.NumistaDataStatusNoPrices
	} else {
		result.Status = models.NumistaDataStatusOK
	}
	return result
}

// Issues returns a type's issues from the shared cache.
func (s *NumistaMarketValueService) Issues(ctx context.Context, typeID int) ([]models.NumistaIssueSummary, bool, error) {
	return cachedNumistaIssues(ctx, s.cache, s.client.Issues, typeID)
}

func emptyMarketValue(currency string) models.CoinMarketValue {
	return models.CoinMarketValue{
		Status:        models.NumistaDataStatusNotLinked,
		Issues:        []models.NumistaIssueSummary{},
		Currency:      currency,
		Prices:        []models.NumistaGradePrice{},
		CatalogPhotos: []models.NumistaCatalogPhoto{},
	}
}

type providerIssue struct {
	ID            int    `json:"id"`
	Year          *int   `json:"year"`
	GregorianYear *int   `json:"gregorian_year"`
	MinYear       *int   `json:"min_year"`
	MaxYear       *int   `json:"max_year"`
	MintLetter    string `json:"mint_letter"`
	Mintage       *int64 `json:"mintage"`
	Comment       string `json:"comment"`
	// A banknote issue's printed date, signatories and references.
	Month      int `json:"month"`
	Day        int `json:"day"`
	Signatures []struct {
		SignerName string `json:"signer_name"`
		Title      string `json:"title"`
	} `json:"signatures"`
	References []providerReference `json:"references"`
}

func parseNumistaIssues(body []byte) ([]models.NumistaIssueSummary, error) {
	var raw []providerIssue
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	issues := make([]models.NumistaIssueSummary, 0, len(raw))
	for _, item := range raw {
		if item.ID <= 0 {
			continue
		}
		year := item.GregorianYear
		if year == nil {
			year = item.Year
		}
		summary := models.NumistaIssueSummary{
			ID:         item.ID,
			Year:       year,
			MinYear:    item.MinYear,
			MaxYear:    item.MaxYear,
			MintLetter: strings.TrimSpace(item.MintLetter),
			Mintage:    item.Mintage,
			Comment:    strings.TrimSpace(item.Comment),
		}
		if item.Month >= 1 && item.Month <= 12 {
			summary.Month = item.Month
			if item.Day >= 1 && item.Day <= 31 {
				summary.Day = item.Day
			}
		}
		for _, signature := range item.Signatures {
			if name := strings.TrimSpace(signature.SignerName); name != "" {
				summary.Signatures = append(summary.Signatures, models.NumistaIssueSignature{Name: name, Title: strings.TrimSpace(signature.Title)})
			}
		}
		summary.Pick = issuePickNumber(item.References, summary.Comment)
		summary.Variant = issueNoteVariant(summary.Comment, summary.Pick)
		issues = append(issues, summary)
	}
	return issues, nil
}

func parseNumistaPrices(body []byte) ([]models.NumistaGradePrice, error) {
	var raw struct {
		Prices []models.NumistaGradePrice `json:"prices"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	prices := make([]models.NumistaGradePrice, 0, len(raw.Prices))
	for _, price := range raw.Prices {
		grade := strings.ToLower(strings.TrimSpace(price.Grade))
		if grade == "" || price.Price <= 0 {
			continue
		}
		prices = append(prices, models.NumistaGradePrice{Grade: grade, Price: price.Price})
	}
	sort.SliceStable(prices, func(i, j int) bool {
		return gradeRank(prices[i].Grade) < gradeRank(prices[j].Grade)
	})
	return prices, nil
}

func gradeRank(grade string) int {
	if rank, ok := numistaGradeOrder[grade]; ok {
		return rank
	}
	return len(numistaGradeOrder)
}

var (
	coinDateNumber = regexp.MustCompile(`\d{1,4}`)
	coinDateBC     = regexp.MustCompile(`(?i)\b(B\.?\s?C\.?(E\.?)?|a\.\s?C\.?)(\s|$)`)
)

// chooseNumistaIssue picks the requested issue when it belongs to the type,
// else the issue whose year matches a single year in the date hint, else the
// first issue.
func chooseNumistaIssue(issues []models.NumistaIssueSummary, requested int, dateHint string) int {
	if requested > 0 {
		for _, issue := range issues {
			if issue.ID == requested {
				return requested
			}
		}
	}
	if year, ok := singleCoinYear(dateHint); ok {
		for _, issue := range issues {
			if issue.Year != nil && *issue.Year == year {
				return issue.ID
			}
		}
	}
	return issues[0].ID
}

// singleCoinYear reads a coin date such as "1968" or "44 BC" as one Gregorian
// year (negative for BC). Ranges and dates with several numbers are ambiguous
// and return false.
func singleCoinYear(dateRange string) (int, bool) {
	numbers := coinDateNumber.FindAllString(dateRange, -1)
	if len(numbers) != 1 {
		return 0, false
	}
	year, err := strconv.Atoi(numbers[0])
	if err != nil || year == 0 {
		return 0, false
	}
	if coinDateBC.MatchString(dateRange) {
		year = -year
	}
	return year, true
}
