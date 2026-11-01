package services

import (
	"regexp"
	"strconv"
	"strings"
)

// commonsImagesOnly limits a Commons search to images, not the scanned
// documents and PDFs a coin's words also find.
const commonsImagesOnly = " filetype:bitmap"

// commonsPhotoQueries are the Wikimedia Commons searches for a coin, best
// first. Commons names most coin photos in English ("200 lire San Marino 1995
// coin"), so the coin's denomination and country come first in English, with
// its year and then without it (a type looks the same from year to year);
// then the text typed, as it is and simplified. Each asks for images only.
func commonsPhotoQueries(coin CoinPhotoQuery) []string {
	var queries []string
	seen := map[string]bool{}
	add := func(words ...string) {
		query := strings.Join(strings.Fields(strings.Join(words, " ")), " ")
		if query == "" || seen[strings.ToLower(query)] {
			return
		}
		seen[strings.ToLower(query)] = true
		queries = append(queries, query+commonsImagesOnly)
	}
	denomination := englishDenomination(coin.Denomination)
	country := strings.TrimSpace(coin.Country)
	if denomination != "" && country != "" {
		kind := "coin"
		if coin.Banknote {
			kind = "banknote"
		}
		// A year before Christ is written differently on Commons: left out.
		if coin.Year > 0 {
			add(denomination, country, strconv.Itoa(coin.Year), kind)
		}
		add(denomination, country, kind)
	}
	query := strings.Join(strings.Fields(coin.Query), " ")
	add(query)
	add(simplifyPhotoQuery(query))
	return queries
}

// englishUnits are the Portuguese names of currency units that differ in
// English, as Commons titles use them; others (escudos, centavos, pesetas,
// kopiyok…) are the same in both.
var englishUnits = map[string]string{
	"cêntimo de euro":  "euro cent",
	"cêntimos de euro": "euro cent",
	"cêntimo":          "cent",
	"cêntimos":         "cents",
	"euros":            "euro",
	"libra":            "pound",
	"libras":           "pounds",
	"penny novo":       "new penny",
	"pence novos":      "new pence",
	"xelim":            "shilling",
	"xelins":           "shillings",
	"franco":           "franc",
	"francos":          "francs",
	"liras":            "lire",
	"rublo":            "ruble",
	"rublos":           "rubles",
	"rúpia":            "rupee",
	"rúpias":           "rupees",
	"iene":             "yen",
	"ienes":            "yen",
	"florim":           "florin",
	"florins":          "florins",
	"marco":            "mark",
	"marcos":           "marks",
	"dólar":            "dollar",
	"dólares":          "dollars",
	"coroa":            "crown",
	"coroas":           "crowns",
	"piastra":          "piastre",
	"piastras":         "piastres",
	"dinares":          "dinars",
	"leus":             "lei",
	"meticais":         "meticais",
}

var (
	// "10.000 reis" is ten thousand: a dot between thousands.
	denominationThousands = regexp.MustCompile(`(\d)\.(\d{3})\b`)
	denominationNumber    = regexp.MustCompile(`^([0-9][0-9.,]*|[0-9]*[½¼¾])\s+(.+)$`)
)

// englishDenomination writes a denomination as Commons titles do: the unit in
// English when it differs ("200 liras" → "200 lire"), fractions as decimals
// ("7½ euro" → "7.5 euro"), a comma as a dot ("2,5 escudos" → "2.5 escudos").
// A denomination it cannot read comes back as it is.
func englishDenomination(denomination string) string {
	text := strings.Join(strings.Fields(denomination), " ")
	if text == "" {
		return ""
	}
	match := denominationNumber.FindStringSubmatch(text)
	if match == nil {
		return text
	}
	number, unit := match[1], strings.ToLower(match[2])
	number = denominationThousands.ReplaceAllString(number, "$1$2")
	number = strings.ReplaceAll(number, ",", ".")
	for fraction, decimal := range map[string]string{"½": ".5", "¼": ".25", "¾": ".75"} {
		if strings.HasSuffix(number, fraction) {
			whole := strings.TrimSuffix(number, fraction)
			if whole == "" {
				whole = "0"
			}
			number = whole + decimal
		}
	}
	if english, ok := englishUnits[unit]; ok {
		unit = english
	}
	return number + " " + unit
}
