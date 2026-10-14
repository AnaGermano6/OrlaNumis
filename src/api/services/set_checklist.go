package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// A checklist set lists the coins to collect: a country's coins in the
// catalogue (all of them, or of one face value), or a challenge. The
// collection's coins fill its slots by themselves.
const (
	ChecklistSourceCatalogue = "catalogue"
	ChecklistSourceChallenge = "challenge"

	// One coin of each euro country, whatever it is.
	ChallengeEuroCountries = "euro-countries"
	// A joint issue of the euro area (the Treaty of Rome in 2007): each
	// country's version of the same €2 coin.
	ChallengeJointIssue = "joint-issue"
	// A country's eight coins of a year, from 1 cent to 2 euros.
	ChallengeAnnualSeries = "annual-series"
)

// ChecklistDefinition says what a checklist set lists.
type ChecklistDefinition struct {
	Source string `json:"source" enums:"catalogue,challenge"`
	// Country is an ISO code: the catalogue's country, or the annual series'.
	Country string `json:"country,omitempty"`
	// Currency is "PTE" for Portugal's escudos; the euro otherwise.
	Currency string `json:"currency,omitempty"`
	// FaceValue keeps the catalogue's coins of that face value; all when 0.
	FaceValue float64 `json:"faceValue,omitempty"`
	Challenge string  `json:"challenge,omitempty" enums:"euro-countries,joint-issue,annual-series"`
	// Year is the annual series' or the joint issue's.
	Year int `json:"year,omitempty"`
}

// ParseChecklistDefinition reads and checks a checklist set's definition.
func ParseChecklistDefinition(raw map[string]interface{}) (ChecklistDefinition, error) {
	var definition ChecklistDefinition
	body, err := json.Marshal(raw)
	if err == nil {
		err = json.Unmarshal(body, &definition)
	}
	if err != nil {
		return ChecklistDefinition{}, fmt.Errorf("checklist must describe a catalogue or a challenge")
	}
	definition.Country = strings.ToUpper(strings.TrimSpace(definition.Country))
	definition.Currency = strings.ToUpper(strings.TrimSpace(definition.Currency))
	switch definition.Source {
	case ChecklistSourceCatalogue:
		if CatalogueCountryName(definition.Country) == "" {
			return ChecklistDefinition{}, fmt.Errorf("checklist country must be a country of the catalogue")
		}
		if definition.Currency != "" && (definition.Currency != "PTE" || definition.Country != "PT") {
			return ChecklistDefinition{}, fmt.Errorf("checklist currency must be PTE for Portugal, or the euro")
		}
		if definition.FaceValue < 0 {
			return ChecklistDefinition{}, fmt.Errorf("checklist face value must not be negative")
		}
		definition.Challenge, definition.Year = "", 0
	case ChecklistSourceChallenge:
		switch definition.Challenge {
		case ChallengeEuroCountries:
			definition.Country, definition.Year = "", 0
		case ChallengeJointIssue:
			if definition.Year < ecbFirstCommemorativeYear || definition.Year > time.Now().Year() {
				return ChecklistDefinition{}, fmt.Errorf("the joint issue needs its year")
			}
			definition.Country = ""
		case ChallengeAnnualSeries:
			if CatalogueCountryName(definition.Country) == "" {
				return ChecklistDefinition{}, fmt.Errorf("the annual series needs a euro country")
			}
			if definition.Year < 1999 || definition.Year > time.Now().Year()+1 {
				return ChecklistDefinition{}, fmt.Errorf("the annual series needs a year of the euro")
			}
		default:
			return ChecklistDefinition{}, fmt.Errorf("checklist challenge is not known")
		}
		definition.Currency, definition.FaceValue = "", 0
	default:
		return ChecklistDefinition{}, fmt.Errorf("checklist source must be catalogue or challenge")
	}
	return definition, nil
}

// JSON is the definition as stored with the set.
func (d ChecklistDefinition) JSON() models.JSONObject {
	body, _ := json.Marshal(d)
	object := models.JSONObject{}
	_ = json.Unmarshal(body, &object)
	return object
}

// ChecklistSlot is one coin to collect, filled or not.
type ChecklistSlot struct {
	Key string `json:"key"`
	// Title names the coin; a challenge's country slots leave it empty (the
	// country says it).
	Title      string    `json:"title,omitempty"`
	Country    string    `json:"country,omitempty"`
	Year       int       `json:"year,omitempty"`
	FaceValue  float64   `json:"faceValue,omitempty"`
	FaceValues []float64 `json:"faceValues,omitempty"`
	Currency   string    `json:"currency,omitempty"`
	// ImageURL and PageURL show the coin as its source does.
	ImageURL string `json:"imageUrl,omitempty"`
	PageURL  string `json:"pageUrl,omitempty"`
	Owned    bool   `json:"owned"`
	// The collection's coin filling the slot.
	CoinID    uint   `json:"coinId,omitempty"`
	CoinName  string `json:"coinName,omitempty"`
	CoinImage string `json:"coinImage,omitempty"`
	// CoinValue is the current value of the coin filling the slot.
	CoinValue *float64 `json:"coinValue,omitempty"`
}

// ChecklistSummary counts a checklist's slots.
type ChecklistSummary struct {
	Total int `json:"total"`
	Owned int `json:"owned"`
	// TotalValue sums the current values of the collection's coins filling slots.
	TotalValue float64 `json:"totalValue"`
}

// Completion is the share of slots filled, in percent.
func (s ChecklistSummary) Completion() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Owned) / float64(s.Total) * 100
}

// SetChecklistService works out a checklist set's slots.
type SetChecklistService struct {
	releases *CoinReleasesService
	coins    *repository.CoinRepository
	photos   *CoinPhotoSearchService
	now      func() time.Time
}

// NewSetChecklistService creates the service; photos finds which catalogue
// coins the collection's are, and may be nil.
func NewSetChecklistService(releases *CoinReleasesService, coins *repository.CoinRepository, photos *CoinPhotoSearchService) *SetChecklistService {
	return &SetChecklistService{releases: releases, coins: coins, photos: photos, now: time.Now}
}

// Complete tells whether the catalogue behind the slots lists everything.
func (s *SetChecklistService) Complete() bool {
	return s.releases.CatalogueComplete()
}

// Summarize counts slots.
func Summarize(slots []ChecklistSlot) ChecklistSummary {
	summary := ChecklistSummary{Total: len(slots)}
	counted := map[uint]bool{}
	for _, slot := range slots {
		if !slot.Owned {
			continue
		}
		summary.Owned++
		if slot.CoinValue != nil && !counted[slot.CoinID] {
			counted[slot.CoinID] = true
			summary.TotalValue += *slot.CoinValue
		}
	}
	return summary
}

// CatalogueSlots lists a catalogue checklist's coins, newest first, without
// looking at the collection.
func (s *SetChecklistService) CatalogueSlots(ctx context.Context, definition ChecklistDefinition) []ChecklistSlot {
	slots := []ChecklistSlot{}
	if definition.Source != ChecklistSourceCatalogue {
		return slots
	}
	for _, coin := range s.releases.Catalogue(ctx) {
		if coin.Country != definition.Country || coin.Currency != definition.Currency {
			continue
		}
		if definition.FaceValue != 0 && !releaseHasFaceValue(coin, definition.FaceValue) {
			continue
		}
		slots = append(slots, catalogueSlot(coin))
	}
	return slots
}

func catalogueSlot(coin CoinRelease) ChecklistSlot {
	slot := ChecklistSlot{
		Key: coin.Keys[0], Title: coin.Title, Country: coin.Country, FaceValue: coin.FaceValue, FaceValues: coin.FaceValues,
		Currency: coin.Currency, ImageURL: coin.ImageURL, PageURL: coin.PageURL,
	}
	fmt.Sscanf(coin.From, "%d", &slot.Year)
	return slot
}

// Slots lists a checklist's slots, each filled by a coin of the collection
// when it has one.
func (s *SetChecklistService) Slots(ctx context.Context, userID uint, definition ChecklistDefinition) ([]ChecklistSlot, error) {
	coins, err := s.coins.CollectionCoinsWithReferences(userID, models.CoinKindCoin)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint]*models.Coin, len(coins))
	for i := range coins {
		byID[coins[i].ID] = &coins[i]
	}
	switch {
	case definition.Source == ChecklistSourceCatalogue:
		return s.catalogueChecklist(ctx, definition, coins, byID), nil
	case definition.Challenge == ChallengeEuroCountries:
		return s.euroCountries(coins), nil
	case definition.Challenge == ChallengeJointIssue:
		return s.jointIssue(ctx, definition.Year, coins, byID), nil
	case definition.Challenge == ChallengeAnnualSeries:
		return annualSeries(definition, coins), nil
	}
	return []ChecklistSlot{}, nil
}

// ownedCatalogueKeys finds the catalogue coins the collection's coins (those
// of a country, or all when "") are.
func (s *SetChecklistService) ownedCatalogueKeys(ctx context.Context, coins []models.Coin, country string) map[string]uint {
	queries := []CoinPhotoQuery{}
	for i := range coins {
		code := CatalogueCountryCode(coins[i].Country)
		if code == "" || country != "" && code != country {
			continue
		}
		query := CoinPhotoQueryFor(&coins[i])
		query.CachedTypesOnly = true
		queries = append(queries, query)
	}
	return s.releases.OwnedKeys(ctx, queries, s.photos)
}

func (s *SetChecklistService) catalogueChecklist(ctx context.Context, definition ChecklistDefinition, coins []models.Coin, byID map[uint]*models.Coin) []ChecklistSlot {
	slots := s.CatalogueSlots(ctx, definition)
	if len(slots) == 0 {
		return slots
	}
	owned := s.ownedCatalogueKeys(ctx, coins, definition.Country)
	keys := map[string][]string{}
	for _, coin := range s.releases.Catalogue(ctx) {
		if coin.Country == definition.Country && coin.Currency == definition.Currency {
			keys[coin.Keys[0]] = coin.Keys
		}
	}
	for i := range slots {
		for _, key := range keys[slots[i].Key] {
			if coinID, ok := owned[key]; ok {
				fill(&slots[i], byID[coinID])
				break
			}
		}
	}
	return slots
}

// fill puts a collection coin in a slot (or marks it filled when the coin
// is not known).
func fill(slot *ChecklistSlot, coin *models.Coin) {
	slot.Owned = true
	if coin == nil {
		return
	}
	slot.CoinID, slot.CoinName, slot.CoinImage, slot.CoinValue = coin.ID, coin.Name, coinPhoto(coin), coin.CurrentValue
}

// coinPhoto is the photo a coin is known by: its obverse, else its first.
func coinPhoto(coin *models.Coin) string {
	for _, image := range coin.Images {
		if image.ImageType == models.ImageTypeObverse {
			return image.FilePath
		}
	}
	for _, image := range coin.Images {
		if image.IsPrimary {
			return image.FilePath
		}
	}
	if len(coin.Images) > 0 {
		return coin.Images[0].FilePath
	}
	return ""
}

// euroCountryCodes are the catalogue's countries, in ISO code order.
func euroCountryCodes() []string {
	codes := make([]string, 0, len(ecbCountryCodes))
	for _, code := range ecbCountryCodes {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	return codes
}

// isEuroCoin tells a coin in euros (a cent to a collector coin) from one of
// the currencies before.
func isEuroCoin(coin *models.Coin) bool {
	if coin.Year == nil {
		return false
	}
	if _, _, ok := euroCoin(coin.Denomination, *coin.Year, coin.Country); ok {
		return true
	}
	return *coin.Year >= 1999 && strings.Contains(foldASCII(coin.Denomination), "euro")
}

// euroCountries: a slot for each euro country, filled by a coin in euros of
// it (one with a photo first).
func (s *SetChecklistService) euroCountries(coins []models.Coin) []ChecklistSlot {
	best := map[string]*models.Coin{}
	for i := range coins {
		coin := &coins[i]
		code := CatalogueCountryCode(coin.Country)
		if code == "" || !isEuroCoin(coin) {
			continue
		}
		if current := best[code]; current == nil || coinPhoto(current) == "" && coinPhoto(coin) != "" {
			best[code] = coin
		}
	}
	year := s.now().Year()
	slots := []ChecklistSlot{}
	for _, code := range euroCountryCodes() {
		slot := ChecklistSlot{Key: "country:" + code, Country: code}
		// The country's €1, as the ECB shows it.
		for _, y := range []int{year, year - 1} {
			if image, page := EcbNationalSideImage("1 euro", y, CatalogueCountryName(code)); image != "" {
				slot.ImageURL, slot.PageURL = image, page
				break
			}
		}
		if coin := best[code]; coin != nil {
			fill(&slot, coin)
			slot.Year = *coin.Year
		}
		slots = append(slots, slot)
	}
	return slots
}

// JointIssue is the euro area's joint issue of a year; nil when it had none.
func (s *SetChecklistService) JointIssue(ctx context.Context, year int) *JointIssue {
	return s.releases.JointIssue(ctx, year)
}

// jointIssue: a slot for each country's version of a joint issue, filled by
// the collection's coin of it.
func (s *SetChecklistService) jointIssue(ctx context.Context, year int, coins []models.Coin, byID map[uint]*models.Coin) []ChecklistSlot {
	issue := s.releases.JointIssue(ctx, year)
	if issue == nil {
		return []ChecklistSlot{}
	}
	// Only the collection's coins of that year can be one of its versions.
	ofYear := []models.Coin{}
	for _, coin := range coins {
		if coin.Year != nil && *coin.Year == year {
			ofYear = append(ofYear, coin)
		}
	}
	owned := s.ownedCatalogueKeys(ctx, ofYear, "")
	slots := make([]ChecklistSlot, 0, len(issue.Coins))
	for _, coin := range issue.Coins {
		slot := ChecklistSlot{
			Key: "joint:" + coin.Country, Title: issue.Title, Country: coin.Country, Year: issue.Year, FaceValue: 2,
			ImageURL: coin.ImageURL, PageURL: issue.PageURL,
		}
		if coinID, ok := owned[coin.ImageURL]; ok {
			fill(&slot, byID[coinID])
		}
		slots = append(slots, slot)
	}
	return slots
}

// annualSeriesCoins are the eight euro coins of a year, as euroCoin names them.
var annualSeriesCoins = []struct {
	key   string
	label string
	value float64
}{
	{"1cent", "1 cent", 0.01}, {"2cent", "2 cent", 0.02}, {"5cent", "5 cent", 0.05}, {"10cent", "10 cent", 0.1},
	{"20cent", "20 cent", 0.2}, {"50cent", "50 cent", 0.5}, {"1euro", "1 euro", 1}, {"2euro", "2 euro", 2},
}

// annualSeries: a country's eight regular coins of a year, each filled by
// a coin of the collection of that value, country and year (not a
// commemorative).
func annualSeries(definition ChecklistDefinition, coins []models.Coin) []ChecklistSlot {
	country := CatalogueCountryName(definition.Country)
	filled := map[string]*models.Coin{}
	for i := range coins {
		coin := &coins[i]
		if coin.Year == nil || *coin.Year != definition.Year || CatalogueCountryCode(coin.Country) != definition.Country {
			continue
		}
		if CoinPhotoQueryFor(coin).Subject != "" {
			continue
		}
		if _, key, ok := euroCoin(coin.Denomination, *coin.Year, coin.Country); ok {
			if current := filled[key]; current == nil || coinPhoto(current) == "" && coinPhoto(coin) != "" {
				filled[key] = coin
			}
		}
	}
	slots := make([]ChecklistSlot, 0, len(annualSeriesCoins))
	for _, value := range annualSeriesCoins {
		slot := ChecklistSlot{Key: "annual:" + value.key, Country: definition.Country, Year: definition.Year, FaceValue: value.value}
		slot.ImageURL, slot.PageURL = EcbNationalSideImage(value.label, definition.Year, country)
		if coin := filled[value.key]; coin != nil {
			fill(&slot, coin)
		}
		slots = append(slots, slot)
	}
	return slots
}
