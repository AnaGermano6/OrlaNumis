package services

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// commonsImagesOnly limits a Commons search to images, not the scanned
// documents and PDFs a coin's words also find.
const commonsImagesOnly = " filetype:bitmap"

// commonsPhotoQueries are the Wikimedia Commons searches for a coin, best
// first. Within the categories of the country's coins ("Coins of Austria" and
// all below it), with the denomination and year and then without the year;
// then all of Commons in English, as it names coin photos ("200 lire San
// Marino 1995 coin"), with the year and then without it (a type looks the
// same from year to year); then the text typed, as it is and simplified.
// Each asks for images only.
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
	// A year before Christ is written differently on Commons: left out.
	year := ""
	if coin.Year > 0 {
		year = strconv.Itoa(coin.Year)
	}
	if denomination != "" && country != "" {
		kind := "coin"
		category := `deepcat:"Coins of ` + commonsCountry(country) + `"`
		if coin.Banknote {
			kind = "banknote"
			category = `deepcat:"Banknotes of ` + commonsCountry(country) + `"`
		}
		if year != "" {
			add(category, denomination, year)
		}
		add(category, denomination)
		if year != "" {
			add(denomination, country, year, kind)
		}
		add(denomination, country, kind)
	}
	query := strings.Join(strings.Fields(coin.Query), " ")
	add(query)
	add(simplifyPhotoQuery(query))
	return queries
}

// commonsCountries are the countries Commons names its categories after
// otherwise than in plain English ("Coins of the United Kingdom").
var commonsCountries = map[string]string{
	"united kingdom":                   "the United Kingdom",
	"united states":                    "the United States",
	"netherlands":                      "the Netherlands",
	"czechia":                          "the Czech Republic",
	"czech republic":                   "the Czech Republic",
	"philippines":                      "the Philippines",
	"bahamas":                          "the Bahamas",
	"gambia":                           "the Gambia",
	"maldives":                         "the Maldives",
	"comoros":                          "the Comoros",
	"dominican republic":               "the Dominican Republic",
	"central african republic":         "the Central African Republic",
	"united arab emirates":             "the United Arab Emirates",
	"marshall islands":                 "the Marshall Islands",
	"solomon islands":                  "the Solomon Islands",
	"cook islands":                     "the Cook Islands",
	"faroe islands":                    "the Faroe Islands",
	"democratic republic of the congo": "the Democratic Republic of the Congo",
	"türkiye":                          "Turkey",
	"vatican city":                     "Vatican City",
}

// commonsCountry is a country as Commons' category names write it.
func commonsCountry(country string) string {
	if name, ok := commonsCountries[strings.ToLower(strings.TrimSpace(country))]; ok {
		return name
	}
	return strings.TrimSpace(country)
}

var (
	// A photo taken with a coin beside something, to show its size.
	commonsScaleImage = regexp.MustCompile(`(?i)to indicate scale`)
	// What shows a photo is of a coin or a banknote (matched without accents).
	commonsCoinWords = regexp.MustCompile(`\b(coins?|banknotes?|numismatics?|munzen?|moedas?|monedas?|pieces?|notas?)\b`)
)

// commonsRelevance tells whether a Commons photo shows this coin, from its
// title and categories, and how well: a photo of a coin of the coin's
// country, of its denomination or year; the denomination counts more (a type
// looks the same from year to year), both most. A photo that only
// shows a coin for scale is never kept. Without the coin's country and
// denomination (a text typed) every photo is kept as found.
func commonsRelevance(coin CoinPhotoQuery, photo models.CoinPhotoCandidate) (keep bool, score int) {
	text := photo.Title + " | " + photo.Categories
	if commonsScaleImage.MatchString(text) {
		return false, 0
	}
	denomination := englishDenomination(coin.Denomination)
	country := strings.TrimSpace(coin.Country)
	if denomination == "" || country == "" {
		return true, 0
	}
	folded := foldASCII(strings.NewReplacer("_", " ", "-", " ").Replace(text))
	countryHit := strings.Contains(folded, foldASCII(commonsCountryWord(country)))
	yearHit := coin.Year > 0 && regexp.MustCompile(`\b`+strconv.Itoa(coin.Year)+`\b`).MatchString(folded)
	valueHit := denominationPattern(denomination).MatchString(folded)
	if !commonsCoinWords.MatchString(folded) || !countryHit || !(yearHit || valueHit) {
		return false, 0
	}
	if valueHit {
		score += 4
	}
	if yearHit {
		score += 2
	}
	return true, score
}

// commonsCountryWord is the part of a country's name its coins' titles and
// categories share ("Austria" in "Austrian 5 euro"), without "the".
func commonsCountryWord(country string) string {
	return strings.TrimPrefix(commonsCountry(country), "the ")
}

// denominationPattern finds a denomination in a title or category:
// "2.5 escudos" as "2.5 escudo", "2,5 escudos", "2½ escudos"; "5 euro" as
// "5 euro", "5-euro", "5 €"; "5 cents" as "5 cent", "5c".
func denominationPattern(denomination string) *regexp.Regexp {
	number, unit, _ := strings.Cut(denomination, " ")
	numbers := []string{regexp.QuoteMeta(number)}
	if whole, half := strings.CutSuffix(number, ".5"); half {
		numbers = append(numbers, regexp.QuoteMeta(whole)+`,5`, regexp.QuoteMeta(whole)+`½`)
	}
	units := []string{}
	for _, word := range strings.Fields(foldASCII(unit)) {
		if len(word) > 4 {
			word = word[:4]
		}
		units = append(units, regexp.QuoteMeta(word))
	}
	switch {
	case strings.Contains(unit, "euro"):
		units = append(units, "€", "eur")
	case strings.Contains(unit, "cent"):
		units = append(units, `c\b`, "ct")
	case strings.Contains(unit, "pence") || strings.Contains(unit, "penny"):
		units = append(units, `p\b`)
	}
	return regexp.MustCompile(`(^|[^\d.,])(` + strings.Join(numbers, "|") + `)\s*(` + strings.Join(units, "|") + `)`)
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
