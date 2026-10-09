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
