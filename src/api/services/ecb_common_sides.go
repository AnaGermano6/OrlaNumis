package services

import (
	"regexp"
	"strconv"
	"strings"
)

// The common (European) side of the euro coins is the same in every country.
// The ECB publishes an image of each one; €2 down to 10 cent were redrawn
// with a new map of Europe from 2007 (2008 in some countries).
const ecbCommonSidesPage = "https://www.ecb.europa.eu/euro/coins/common/html/index.en.html"

var ecbCommonSideImages = map[string][2]string{
	// denomination: {first series, new map from 2007}
	"2euro":  {"common_2euro_800.jpg", "newcommon_2euro_800.jpg"},
	"1euro":  {"common_1euro_800.jpg", "newcommon_1euro_800.jpg"},
	"50cent": {"common_50cent_800.jpg", "newcommon_50cent_800.jpg"},
	"20cent": {"common_20cent_800.jpg", "newcommon_20cent_800.jpg"},
	"10cent": {"10common.jpg", "new10common.png"},
	"5cent":  {"common_5cent_800.jpg", "common_5cent_800.jpg"},
	"2cent":  {"common_2cent_800.jpg", "common_2cent_800.jpg"},
	"1cent":  {"common_1cent_800.jpg", "common_1cent_800.jpg"},
}

// First year each country's euro coins are dated, so a coin of its earlier
// currency (Maltese cents, Lithuanian centai) is never taken for a euro coin.
var euroCoinFirstYear = map[string]int{
	"Belgium": 1999, "Finland": 1999, "France": 1999, "Netherlands": 1999, "Spain": 1999,
	"Germany": 2002, "Ireland": 2002, "Italy": 2002, "Luxembourg": 2002, "Austria": 2002, "Portugal": 2002, "Greece": 2002,
	"Monaco": 2001, "San Marino": 2002, "Vatican City": 2002, "Slovenia": 2007, "Cyprus": 2008, "Malta": 2008,
	"Slovakia": 2009, "Estonia": 2011, "Latvia": 2014, "Andorra": 2014, "Lithuania": 2015, "Croatia": 2023, "Bulgaria": 2026,
}

// Countries whose €2 to 10 cent coins show the new map only from 2008.
var euroNewMapFrom2008 = map[string]bool{"Italy": true, "San Marino": true, "Vatican City": true, "Austria": true, "Portugal": true}

var euroDenomination = regexp.MustCompile(`(?i)^\s*(\d+)\s*(euro\s*cents?|euros?|eur|€|cents?|centimos?|centimes?|centesimos?|centavos?|c)(\b|$)`)

// euroCoin identifies a euro circulation coin: its country as the ECB names
// it and its denomination key ("2euro", "10cent"); ok is false for anything
// else, collector coins included (they have no common side).
func euroCoin(denomination string, year int, country string) (ecbCountry, key string, ok bool) {
	ecbCountry = EcbCountryName(country)
	firstYear, isEuroCountry := euroCoinFirstYear[ecbCountry]
	folded := foldASCII(denomination)
	match := euroDenomination.FindStringSubmatch(folded)
	if !isEuroCountry || year < firstYear || match == nil {
		return "", "", false
	}
	// Until 2001 francs, guilders and pesetas were struck too: a cent of
	// those years counts only when it says euro.
	if year < 2002 && !strings.Contains(folded, "euro") && !strings.Contains(folded, "€") {
		return "", "", false
	}
	unit := "cent"
	if strings.HasPrefix(match[2], "eur") || match[2] == "€" {
		unit = "euro"
		// "10 cêntimos de euro" and "2 euro cent" are cents.
		if strings.Contains(match[2], "cent") {
			unit = "cent"
		}
	}
	value, _ := strconv.Atoi(match[1])
	key = strconv.Itoa(value) + unit
	if _, known := ecbCommonSideImages[key]; !known {
		return "", "", false
	}
	return ecbCountry, key, true
}

// EcbCommonSideImage returns the ECB's image of the common side of a euro
// coin and the page it comes from, or "" when the coin is not a euro coin of
// one of the eight denominations.
func EcbCommonSideImage(denomination string, year int, country string) (string, string) {
	ecbCountry, key, ok := euroCoin(denomination, year, country)
	if !ok {
		return "", ""
	}
	images := ecbCommonSideImages[key]
	newMapYear := 2007
	if euroNewMapFrom2008[ecbCountry] {
		newMapYear = 2008
	}
	image := images[0]
	if year >= newMapYear {
		image = images[1]
	}
	return ecbBaseURL + "/euro/coins/common/shared/img/" + image, ecbCommonSidesPage
}
