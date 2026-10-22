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

// MonetarySystemFromNumista names a coin's monetary system from its Numista
// type's currency: its full name, with its dates ("Escudo (1911-2001)", "Real
// (decimalizado, 1835-1910)"), so the réis of Portugal and the real of Brazil
// stay apart; every euro is "Euro", whatever the country (Numista has one
// euro currency per country). Empty when Numista gives no currency.
func MonetarySystemFromNumista(name, fullName string) string {
	name = strings.Join(strings.Fields(name), " ")
	if strings.EqualFold(name, "euro") {
		return "Euro"
	}
	if full := strings.Join(strings.Fields(fullName), " "); full != "" {
		return full
	}
	return name
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

// CoinNumistaFillChange is what a fill from Numista sets on one coin: only
// the fields it fills appear.
type CoinNumistaFillChange struct {
	CoinID     uint   `json:"coinId"`
	Name       string `json:"name"`
	IsWishlist bool   `json:"isWishlist"`
	// Material is set from Composition, Numista's words for the metal.
	Material    Material `json:"material,omitempty"`
	Composition string   `json:"composition,omitempty"`
	// PreviousMaterial is the material a correction replaces.
	PreviousMaterial Material `json:"previousMaterial,omitempty"`
	// MonetarySystem is set from the type's currency (see MonetarySystemFromNumista).
	MonetarySystem string   `json:"monetarySystem,omitempty"`
	WeightGrams    *float64 `json:"weightGrams,omitempty"`
	DiameterMm     *float64 `json:"diameterMm,omitempty"`
	Presentation   string   `json:"presentation,omitempty"`
}

// CoinNumistaFillResult reports a fill of coin records from the Numista
// types the app has cached: what it would change (preview) or changed.
type CoinNumistaFillResult struct {
	// Applied is true once the changes were saved.
	Applied bool `json:"applied"`
	// Coins is how many coins get at least one field.
	Coins         int `json:"coins"`
	Materials     int `json:"materials"`
	Weights       int `json:"weights"`
	Diameters     int `json:"diameters"`
	Presentations int `json:"presentations"`
	// MonetarySystems is how many coins and banknotes get a monetary system.
	MonetarySystems int `json:"monetarySystems"`
	// ByMaterial counts the materials set, by material.
	ByMaterial map[string]int `json:"byMaterial"`
	// Unmatched coins still have no material: Numista's composition fits
	// none of the app's (e.g. plain steel or iron).
	Unmatched int `json:"unmatched"`
	// NotCached coins are linked to a Numista type whose details the app
	// has not cached yet; they are left as they are.
	NotCached int `json:"notCached"`
	// Changes lists every coin the fill changes, by name.
	Changes []CoinNumistaFillChange `json:"changes"`
	// Corrections lists the coins whose material is set but is not what
	// Numista's composition says (e.g. Silver for silver plated copper).
	// They change only when asked for.
	Corrections []CoinNumistaFillChange `json:"corrections"`
	// Corrected is how many corrections were saved.
	Corrected int `json:"corrected"`
}
