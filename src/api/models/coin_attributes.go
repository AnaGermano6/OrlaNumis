package models

import (
	"regexp"
	"strconv"
	"strings"
)

// Coin presentations: how a coin is kept or was issued. Empty means not set.
var CoinPresentations = []string{"loose", "coincard", "capsule", "flip", "proof_set", "blister", "slab"}

// Coin rarities, from most to least common. Empty means not set.
var CoinRarities = []string{"common", "uncommon", "scarce", "rare", "very_rare", "extremely_rare"}

// Coin circulations, set by hand: whether the coin is still legal tender
// (current) or not (demonetized), or was issued for collectors rather than
// for circulation (non-circulating); and whether it is commemorative. Empty
// means "as Numista says".
var CoinCirculations = []string{
	"current", "current_commemorative",
	"demonetized", "demonetized_commemorative",
	"non_circulating", "non_circulating_commemorative",
}

// Coin kinds: what a collection item is. The galleries show coins and
// banknotes apart; everything else treats them alike.
const (
	CoinKindCoin     = "coin"
	CoinKindBanknote = "banknote"
)

// NoteVariants are the kinds of banknote that are not an ordinary note: a
// replacement (star) note, printed in place of a faulty one, or a specimen.
// Empty means an ordinary note.
var NoteVariants = []string{"replacement", "specimen"}

// ValidNoteVariant reports whether value is empty or a known note variant.
func ValidNoteVariant(value string) bool {
	return value == "" || containsValue(NoteVariants, value)
}

// Quantity bounds: how many of a coin or note a collector can record.
const (
	CoinQuantityMin = 1
	CoinQuantityMax = 9999
)

// ValidCoinQuantity reports whether value is a quantity a coin can have.
func ValidCoinQuantity(value int) bool {
	return value >= CoinQuantityMin && value <= CoinQuantityMax
}

// ValidCoinKind reports whether value is a known kind.
func ValidCoinKind(value string) bool {
	return value == CoinKindCoin || value == CoinKindBanknote
}

// ValidCoinCirculation reports whether value is empty or a known circulation.
func ValidCoinCirculation(value string) bool {
	return value == "" || containsValue(CoinCirculations, value)
}

// CoinCirculationSetting is what a circulation set by hand says of a coin.
type CoinCirculationSetting struct {
	// Demonetized coins are no longer legal tender.
	Demonetized bool
	// NonCirculating coins were issued for collectors, not for circulation.
	NonCirculating bool
	Commemorative  bool
}

// CoinCirculationFacts reads a circulation set by hand; ok is false when it
// is not set (or unknown).
func CoinCirculationFacts(value string) (setting CoinCirculationSetting, ok bool) {
	base, commemorative := strings.CutSuffix(value, "_commemorative")
	switch base {
	case "current":
		return CoinCirculationSetting{Commemorative: commemorative}, true
	case "demonetized":
		return CoinCirculationSetting{Demonetized: true, Commemorative: commemorative}, true
	case "non_circulating":
		return CoinCirculationSetting{NonCirculating: true, Commemorative: commemorative}, true
	}
	return CoinCirculationSetting{}, false
}

// ValidCoinPresentation reports whether value is empty or a known presentation.
func ValidCoinPresentation(value string) bool {
	return value == "" || containsValue(CoinPresentations, value)
}

// ValidCoinRarity reports whether value is empty or a known rarity.
func ValidCoinRarity(value string) bool {
	return value == "" || containsValue(CoinRarities, value)
}

func containsValue(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

var (
	coinYearNumber  = regexp.MustCompile(`\d{1,4}`)
	coinYearBC      = regexp.MustCompile(`(?i)\b(B\.?\s?C\.?(E\.?)?|a\.\s?C\.?)(\s|$|[-–—,;)])`)
	coinYearAD      = regexp.MustCompile(`(?i)\b(A\.?\s?D\.?|C\.?E\.?|d\.\s?C\.?)(\s|$|[-–—,;)])|\bAD\b`)
	coinYearCentury = regexp.MustCompile(`(?i)century|séc|sec\.|século|seculo|\d(st|nd|rd|th)\b`)
)

// DeriveCoinYear reads the first year of a coin date such as "1968",
// "27 BC – AD 14", "46–44 BC" or "c. 300 a.C." (negative for BC), so coins can
// be sorted and grouped by year. A BC marker applies to the first year unless
// an AD marker comes between them. Centuries and dates without a number give nil.
func DeriveCoinYear(dateRange string) *int {
	if coinYearCentury.MatchString(dateRange) {
		return nil
	}
	loc := coinYearNumber.FindStringIndex(dateRange)
	if loc == nil {
		return nil
	}
	year, err := strconv.Atoi(dateRange[loc[0]:loc[1]])
	if err != nil || year == 0 {
		return nil
	}
	rest := dateRange[loc[1]:]
	if bc := coinYearBC.FindStringIndex(rest); bc != nil && !coinYearAD.MatchString(rest[:bc[0]]) {
		year = -year
	}
	return &year
}

// NormalizeCoinCountry tidies a country name for storage and grouping.
func NormalizeCoinCountry(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

// CoinCountryFillResult reports a fill of missing countries from Numista.
type CoinCountryFillResult struct {
	Status string `json:"status"`
	// Updated coins got a country in this call.
	Updated int `json:"updated"`
	// Remaining linked coins still have no country (call again to continue).
	Remaining int `json:"remaining"`
	// Failed coins could not be resolved (e.g. Numista did not answer).
	Failed int `json:"failed"`
}
