package services

import (
	"regexp"
	"strings"
)

// The national side of the euro coins differs by country and, in some
// countries, by series (a new monarch or pope, a redesign). The ECB shows one
// image per series on each country's page.
const ecbNationalSidesPage = "https://www.ecb.europa.eu/euro/coins/html/%s.en.html"

type ecbSeries struct {
	// from is the first year of the series; 0 when the country has one.
	from int
	path string
}

type ecbNationalPage struct {
	code  string
	sides map[string][]ecbSeries
	// until is the last year the ECB has a design for (0: no limit).
	until int
	// ambiguous is a year two series were struck in.
	ambiguous int
}

// ecbNationalSides lists the ECB's images of each country's national sides.
// Estonia, Slovenia and Monaco are missing: their ECB pages are unavailable.
var ecbNationalSides = map[string]ecbNationalPage{
	"Andorra": {code: "ad", sides: map[string][]ecbSeries{
		"2euro":  {{0, "ad/ad_2euro.jpg"}},
		"1euro":  {{0, "ad/Andorra_1euro.jpg"}},
		"50cent": {{0, "ad/Andorra_50cent.jpg"}},
		"20cent": {{0, "ad/Andorra_20cent.jpg"}},
		"10cent": {{0, "ad/Andorra_10cent.jpg"}},
		"5cent":  {{0, "ad/Andorra_5cent.jpg"}},
		"2cent":  {{0, "ad/Andorra_2cent.jpg"}},
		"1cent":  {{0, "ad/Andorra_1cent.jpg"}},
	}},
	"Austria": {code: "at", sides: map[string][]ecbSeries{
		"2euro":  {{0, "at/Austria_2Euro.jpg"}},
		"1euro":  {{0, "at/Austria_1Euro.jpg"}},
		"50cent": {{0, "at/Austria_50cent.jpg"}},
		"20cent": {{0, "at/Austria_20cent.jpg"}},
		"10cent": {{0, "at/Austria_10cent.jpg"}},
		"5cent":  {{0, "at/Austria_5cent.jpg"}},
		"2cent":  {{0, "at/Austria_2cent.jpg"}},
		"1cent":  {{0, "at/Austria_1cent.jpg"}},
	}},
	"Belgium": {code: "be", sides: map[string][]ecbSeries{
		"2euro":  {{1999, "be/Belgium_2euroAl.jpg"}, {2008, "be/Belgium_2euro_2008.jpg"}, {2014, "be/Belgium_2euro_2014ph.jpg"}},
		"1euro":  {{1999, "be/Belgium_1999_1euro.gif"}, {2008, "be/Belgium_1euro_2011.jpg"}, {2014, "be/Belgium_1euro_2014.jpg"}},
		"50cent": {{1999, "be/Belgium_50cent_2002.jpg"}, {2008, "be/Belgium_50cent_2008.jpg"}, {2014, "be/Belgium_50cent_2014.jpg"}},
		"20cent": {{1999, "be/Belgium_20cent_2002.jpg"}, {2008, "be/Belgium_20cent_2009.jpg"}, {2014, "be/Belgium_20cent_2014.jpg"}},
		"10cent": {{1999, "be/Belgium_10cent_1999.jpg"}, {2008, "be/Belgium_10cent_2008.jpg"}, {2014, "be/Belgium_2014_10cent.jpg"}},
		"5cent":  {{1999, "be/Belgium_5cent_1999.jpg"}, {2008, "be/Belgium_5cent_2011.jpg"}, {2014, "be/Belgium_5cent_2014.jpg"}},
		"2cent":  {{1999, "be/Belgium_2cent_2000.jpg"}, {2008, "be/Belgium_2cent_2013.jpg"}, {2014, "be/Belgium_2cent_2014.jpg"}},
		"1cent":  {{1999, "be/Belgium_1cent_2000.jpg"}, {2008, "be/Belgium_1cent_2013.jpg"}, {2014, "be/Belgium_1cent_2014.jpg"}},
	}},
	"Bulgaria": {code: "bg", sides: map[string][]ecbSeries{
		"2euro":  {{0, "bg/Bulgaria_2euro.jpg"}},
		"1euro":  {{0, "bg/Bulgaria_1euro.jpg"}},
		"50cent": {{0, "bg/Bulgaria_50cent.jpg"}},
		"20cent": {{0, "bg/Bulgaria_20cent.jpg"}},
		"10cent": {{0, "bg/Bulgaria_10cent.jpg"}},
		"5cent":  {{0, "bg/Bulgaria_5cent.jpg"}},
		"2cent":  {{0, "bg/Bulgaria_2cent.jpg"}},
		"1cent":  {{0, "bg/Bulgaria_1cent.jpg"}},
	}},
	"Croatia": {code: "hr", sides: map[string][]ecbSeries{
		"2euro":  {{0, "hr/Croatia_2euro.jpg"}},
		"1euro":  {{0, "hr/Croatia_1euro.jpg"}},
		"50cent": {{0, "hr/Croatia_50cent.jpg"}},
		"20cent": {{0, "hr/Croatia_20cent.jpg"}},
		"10cent": {{0, "hr/Croatia_10cent.jpg"}},
		"5cent":  {{0, "hr/Croatia_5cent.jpg"}},
		"2cent":  {{0, "hr/Croatia_2cent.jpg"}},
		"1cent":  {{0, "hr/Croatia_1cent.jpg"}},
	}},
	"Cyprus": {code: "cy", sides: map[string][]ecbSeries{
		"2euro":  {{0, "cy/Cyprus_2Euro.jpg"}},
		"1euro":  {{0, "cy/Cyprus_1euro.jpg"}},
		"50cent": {{0, "cy/Cyprus_50cent.jpg"}},
		"20cent": {{0, "cy/Cyprus_20cent.jpg"}},
		"10cent": {{0, "cy/Cyprus_10cent.jpg"}},
		"5cent":  {{0, "cy/Cyprus_5cent.jpg"}},
		"2cent":  {{0, "cy/Cyprus_2cent.jpg"}},
		"1cent":  {{0, "cy/Cyprus_1cent.jpg"}},
	}},
	"Finland": {code: "fi", sides: map[string][]ecbSeries{
		"2euro":  {{0, "fi/Finland_2euro.jpg"}},
		"1euro":  {{0, "fi/Finland_1euro.jpg"}},
		"50cent": {{0, "fi/Finland_50cent.jpg"}},
		"20cent": {{0, "fi/Finland_20cent.jpg"}},
		"10cent": {{0, "fi/Finland_10cent.jpg"}},
		"5cent":  {{0, "fi/Finland_5cent.jpg"}},
		"2cent":  {{0, "fi/Finland_2cent.jpg"}},
		"1cent":  {{0, "fi/Finland_1cent.jpg"}},
	}},
	"France": {code: "fr", sides: map[string][]ecbSeries{
		"2euro":  {{1999, "fr/France_2euro.jpg"}, {2022, "fr/France_2euro_2022.jpg"}},
		"1euro":  {{1999, "fr/France_1Euro.jpg"}, {2022, "fr/France_1euro_2022.jpg"}},
		"50cent": {{1999, "fr/France_50cent.jpg"}, {2024, "fr/France_50cent_2024.jpg"}},
		"20cent": {{1999, "fr/France_20cent.jpg"}, {2024, "fr/France_20cent_2024.jpg"}},
		"10cent": {{1999, "fr/France_10cent.jpg"}, {2024, "fr/France_10cent_2024.jpg"}},
		"5cent":  {{0, "fr/France_5cent.jpg"}},
		"2cent":  {{0, "fr/France_2cent.jpg"}},
		"1cent":  {{0, "fr/France_1cent.jpg"}},
	}},
	"Germany": {code: "de", sides: map[string][]ecbSeries{
		"2euro":  {{0, "de/Germany_2euro.jpg"}},
		"1euro":  {{0, "de/Germany_1euro.jpg"}},
		"50cent": {{0, "de/Germany_50cent.jpg"}},
		"20cent": {{0, "de/Germany_20cent.jpg"}},
		"10cent": {{0, "de/Germany_10cent.jpg"}},
		"5cent":  {{0, "de/Germany_5cent.jpg"}},
		"2cent":  {{0, "de/Germany_2cent.jpg"}},
		"1cent":  {{0, "de/Germany_1cent.jpg"}},
	}},
	"Greece": {code: "gr", sides: map[string][]ecbSeries{
		"2euro":  {{0, "gr/Greece_2euro.jpg"}},
		"1euro":  {{0, "gr/Greece_1euro.jpg"}},
		"50cent": {{0, "gr/Greece_50cent.jpg"}},
		"20cent": {{0, "gr/Greece_20cent.jpg"}},
		"10cent": {{0, "gr/Greece_10cent.jpg"}},
		"5cent":  {{0, "gr/Greece_5cent.jpg"}},
		"2cent":  {{0, "gr/Greece_2cent.jpg"}},
		"1cent":  {{0, "gr/Greece_1cent.jpg"}},
	}},
	"Ireland": {code: "ie", sides: map[string][]ecbSeries{
		"2euro":  {{0, "ie/Ireland_2euro.jpg"}},
		"1euro":  {{0, "ie/Ireland_1euro.jpg"}},
		"50cent": {{0, "ie/Ireland_50cent.jpg"}},
		"20cent": {{0, "ie/Ireland_20cent.jpg"}},
		"10cent": {{0, "ie/Ireland_10cent.jpg"}},
		"5cent":  {{0, "ie/Ireland_5cent.jpg"}},
		"2cent":  {{0, "ie/Ireland_2cent.jpg"}},
		"1cent":  {{0, "ie/Ireland_1cent.jpg"}},
	}},
	"Italy": {code: "it", sides: map[string][]ecbSeries{
		"2euro":  {{0, "it/Italy_2euro.jpg"}},
		"1euro":  {{0, "it/Italy_1euro.jpg"}},
		"50cent": {{0, "it/Italy_50cent.jpg"}},
		"20cent": {{0, "it/Italy_20cent.jpg"}},
		"10cent": {{0, "it/Italy_10cent.jpg"}},
		"5cent":  {{0, "it/Italy_5cent.jpg"}},
		"2cent":  {{0, "it/Italy_2cent.jpg"}},
		"1cent":  {{0, "it/Italy_1cent.jpg"}},
	}},
	"Latvia": {code: "lv", sides: map[string][]ecbSeries{
		"2euro":  {{0, "lv/Latvia_2euro.jpg"}},
		"1euro":  {{0, "lv/Latvia_1euro.jpg"}},
		"50cent": {{0, "lv/Latvia_50cent.jpg"}},
		"20cent": {{0, "lv/Latvia_20cent.jpg"}},
		"10cent": {{0, "lv/Latvia_10cent.jpg"}},
		"5cent":  {{0, "lv/Latvia_5cent.jpg"}},
		"2cent":  {{0, "lv/Latvia_2cent.jpg"}},
		"1cent":  {{0, "lv/Latvia_1cent.jpg"}},
	}},
	"Lithuania": {code: "lt", sides: map[string][]ecbSeries{
		"2euro":  {{0, "lt/Lithuania_2euro_2015.jpg"}},
		"1euro":  {{0, "lt/Lithuania_1euro_2015.jpg"}},
		"50cent": {{0, "lt/Lithuania_50cent_2014.jpg"}},
		"20cent": {{0, "lt/Lithuania_20cent_2015.jpg"}},
		"10cent": {{0, "lt/Lithuania_10cent_2015.jpg"}},
		"5cent":  {{0, "lt/Lithuania_5cent_2015.jpg"}},
		"2cent":  {{0, "lt/Lithuania_2cent_2015.jpg"}},
		"1cent":  {{0, "lt/Lithuania_1cent_2015.jpg"}},
	}},
	"Luxembourg": {code: "lu", until: 2025, sides: map[string][]ecbSeries{
		"2euro":  {{0, "lu/Luxembourg_2euro.jpg"}},
		"1euro":  {{0, "lu/Luxembourg_1Euro.jpg"}},
		"50cent": {{0, "lu/Luxembourg_50cent.jpg"}},
		"20cent": {{0, "lu/Luxembourg_20cent.jpg"}},
		"10cent": {{0, "lu/Luxembourg_10cent.jpg"}},
		"5cent":  {{0, "lu/Luxembourg_5cent.jpg"}},
		"2cent":  {{0, "lu/Luxembourg_2cent.jpg"}},
		"1cent":  {{0, "lu/Luxembourg_1cent.jpg"}},
	}},
	"Malta": {code: "mt", sides: map[string][]ecbSeries{
		"2euro":  {{0, "mt/Malta_2euro.jpg"}},
		"1euro":  {{0, "mt/Malta_1Euro.jpg"}},
		"50cent": {{0, "mt/Malta_50cent.jpg"}},
		"20cent": {{0, "mt/Malta_20cent.jpg"}},
		"10cent": {{0, "mt/Malta_10cent.jpg"}},
		"5cent":  {{0, "mt/Malta_5cent.jpg"}},
		"2cent":  {{0, "mt/Malta_2cent.jpg"}},
		"1cent":  {{0, "mt/Malta_1cent.jpg"}},
	}},
	"Netherlands": {code: "nl", sides: map[string][]ecbSeries{
		"2euro":  {{1999, "nl/Netherlands_2euro_2001.jpg"}, {2014, "nl/Netherlands_2Euro.jpg"}},
		"1euro":  {{1999, "nl/Netherlands_1euro_2000.jpg"}, {2014, "nl/Netherlands_1euro_2014.jpg"}},
		"50cent": {{1999, "nl/Netherlands_50cents_2000.jpg"}, {2014, "nl/Netherlands_50cent.jpg"}},
		"20cent": {{1999, "nl/Netherlands_20cent_2000.jpg"}, {2014, "nl/Netherlands_20cent.jpg"}},
		"10cent": {{1999, "nl/Netherlands_10cent_2000.jpg"}, {2014, "nl/Netherlands_10cent.jpg"}},
		"5cent":  {{1999, "nl/Netherlands_5cent_2000.jpg"}, {2014, "nl/Netherlands_5cent_2014.jpg"}},
		"2cent":  {{1999, "nl/Netherlands_2cent_2000.jpg"}, {2014, "nl/Netherlands_2cent_2014.jpg"}},
		"1cent":  {{1999, "nl/Netherlands_1cent_2000.jpg"}, {2014, "nl/Netherlands_1cent_2014.jpg"}},
	}},
	"Portugal": {code: "pt", sides: map[string][]ecbSeries{
		"2euro":  {{0, "pt/Portugal_2euro.jpg"}},
		"1euro":  {{0, "pt/Portugal_1Euro.jpg"}},
		"50cent": {{0, "pt/Portugal_50cent.jpg"}},
		"20cent": {{0, "pt/Portugal_20cent.jpg"}},
		"10cent": {{0, "pt/Portugal_10cent.jpg"}},
		"5cent":  {{0, "pt/Portugal_5cent.jpg"}},
		"2cent":  {{0, "pt/Portugal_2cent.jpg"}},
		"1cent":  {{0, "pt/Portugal_1cent.jpg"}},
	}},
	"San Marino": {code: "sm", sides: map[string][]ecbSeries{
		"2euro":  {{2008, "sm/SanMarino_2euro_2010.jpg"}, {2017, "sm/SanMarino_2euro_2017.jpg"}},
		"1euro":  {{2008, "sm/SanMarino_1euro_2010.jpg"}, {2017, "sm/SanMarino_1euro_2017.jpg"}},
		"50cent": {{2008, "sm/SanMarino_50cent.jpg"}, {2017, "sm/SanMarino_50cent_2017.jpg"}},
		"20cent": {{2008, "sm/SanMarino_20cent.jpg"}, {2017, "sm/SanMarino_20cent_2017.jpg"}},
		"10cent": {{2008, "sm/SanMarino_10cent.jpg"}, {2017, "sm/SanMarino_10cent_2017.jpg"}},
		"5cent":  {{2008, "sm/SanMarino_5cent.jpg"}, {2017, "sm/SanMarino_5cent_2017.jpg"}},
		"2cent":  {{2008, "sm/SanMarino_2cent.jpg"}, {2017, "sm/SanMarino_2cent_2017.jpg"}},
		"1cent":  {{2008, "sm/SanMarino_1cent.jpg"}, {2017, "sm/SanMarino_1cent_2017.jpg"}},
	}},
	"Slovakia": {code: "sk", sides: map[string][]ecbSeries{
		"2euro":  {{0, "sk/Slovakia_2Euro.jpg"}},
		"1euro":  {{0, "sk/Slovakia_1Euro.jpg"}},
		"50cent": {{0, "sk/Slovakia_50cent.jpg"}},
		"20cent": {{0, "sk/Slovakia_20cent.jpg"}},
		"10cent": {{0, "sk/Slovakia_10cent.jpg"}},
		"5cent":  {{0, "sk/Slovakia_5cent.jpg"}},
		"2cent":  {{0, "sk/Slovakia_2cent.jpg"}},
		"1cent":  {{0, "sk/Slovakia_1cent.jpg"}},
	}},
	"Spain": {code: "es", sides: map[string][]ecbSeries{
		"2euro":  {{1999, "es/Spain_2Euro_2003.jpg"}, {2010, "es/Spain_2euro_2014.jpg"}, {2015, "es/Spain_2euro_2015.jpg"}},
		"1euro":  {{1999, "es/Spain_1Euro_1999.jpg"}, {2010, "es/Spain_1euro_2014.jpg"}, {2015, "es/Spain_1Euro_2015.jpg"}},
		"50cent": {{1999, "es/Spain_50cent_1999.jpg"}, {2010, "es/Spain_50cent_2014.jpg"}},
		"20cent": {{1999, "es/Spain_20cent_2001.jpg"}, {2010, "es/Spain_20cent.jpg"}},
		"10cent": {{1999, "es/Spain_10cent_2003.jpg"}, {2010, "es/Spain_10cent.jpg"}},
		"5cent":  {{1999, "es/Spain_5cent_1999.jpg"}, {2010, "es/Spain_5cent.jpg"}},
		"2cent":  {{1999, "es/Spain_2cent_2002.jpg"}, {2010, "es/Spain_2cent.jpg"}},
		"1cent":  {{1999, "es/Spain_1cent_1999.jpg"}, {2010, "es/Spain_1cent.jpg"}},
	}},
	"Vatican City": {code: "va", until: 2024, ambiguous: 2005, sides: map[string][]ecbSeries{
		"2euro":  {{2002, "va/Vaticano_2euro_2002.jpg"}, {2005, "va/Vaticano_2euro_SedeVacante.jpg"}, {2006, "va/Vaticano_2euro_2013.jpg"}, {2014, "va/Vaticano_2euro_2016.jpg"}, {2017, "va/Vatican_2euro_2017.jpg"}},
		"1euro":  {{2002, "va/Vatican_1euro_2002.jpg"}, {2005, "va/Vatican_1euro_SedeVacante.jpg"}, {2006, "va/Vatican_1euro_2013.jpg"}, {2014, "va/Vatican_1euro_2016.jpg"}, {2017, "va/Vatican_1euro_2017.jpg"}},
		"50cent": {{2002, "va/Vatican_50cent_2002.jpg"}, {2005, "va/Vatican_50cent_SedeVacante.jpg"}, {2006, "va/Vatican_50cent_2013.jpg"}, {2014, "va/Vatican_50cent_2016.jpg"}, {2017, "va/Vatican_50cent_2017.jpg"}},
		"20cent": {{2002, "va/Vatican_20cent_2002.jpg"}, {2005, "va/Vatican_20cent_SedeVacante.jpg"}, {2006, "va/Vatican_20cent_2013.jpg"}, {2014, "va/Vatican_20cent_2016.jpg"}, {2017, "va/Vatican_20cent_2017.jpg"}},
		"10cent": {{2002, "va/Vatican_10cent_2002.jpg"}, {2005, "va/Vatican_10cent_SedeVacante.jpg"}, {2006, "va/Vatican_10cent_2013.jpg"}, {2014, "va/Vatican_10cent_2016.jpg"}, {2017, "va/Vatican_10cent_2017.jpg"}},
		"5cent":  {{2002, "va/Vatican_5cent_2002.jpg"}, {2005, "va/Vatican_5cent_SedeVacante.jpg"}, {2006, "va/Vatican_5cent_2013.jpg"}, {2014, "va/Vatican_5cent_2016.jpg"}, {2017, "va/Vatican_5cent_2017.jpg"}},
		"2cent":  {{2002, "va/Vatican_2cent_2002.jpg"}, {2005, "va/Vatican_2cent_SedeVacante.jpg"}, {2006, "va/Vatican_2cent_2013.jpg"}, {2014, "va/Vatican_2cent_2016.jpg"}, {2017, "va/Vatican_2cent_2017.jpg"}},
		"1cent":  {{2002, "va/Vatican_1cent_2002.jpg"}, {2005, "va/Vatican_1cent_SedeVacante.jpg"}, {2006, "va/Vatican_1cent_2013.jpg"}, {2014, "va/Vatican_1cent_2016.jpg"}, {2017, "va/Vatican_1cent_2017.jpg"}},
	}},
}

// EcbNationalSideImage returns the ECB's image of the national side of a
// regular euro coin and the page it comes from, or "" when the ECB has none
// for that country, series or year.
func EcbNationalSideImage(denomination string, year int, country string) (string, string) {
	ecbCountry, key, ok := euroCoin(denomination, year, country)
	page, known := ecbNationalSides[ecbCountry]
	if !ok || !known || (page.until > 0 && year > page.until) || year == page.ambiguous {
		return "", ""
	}
	image := ""
	for _, series := range page.sides[key] {
		if year >= series.from {
			image = series.path
		}
	}
	if image == "" {
		return "", ""
	}
	return ecbBaseURL + "/euro/coins/common/shared/img/" + image, strings.Replace(ecbNationalSidesPage, "%s", page.code, 1)
}

var euroSeriesNote = regexp.MustCompile(`(?i)\((1st|2nd|3rd|first|second|third)\s+map\)`)

// RegularEuroType tells a regular euro coin type ("2 Euros (2nd map)",
// "10 Euro Cents (Simone Veil)" struck over several years) from a €2
// commemorative ("2 Euros (Thuringia)", struck in one year).
func RegularEuroType(title string, minYear, maxYear *int) bool {
	if minYear == nil {
		return false
	}
	if !strings.Contains(title, "(") || euroSeriesNote.MatchString(title) {
		return true
	}
	return maxYear != nil && *maxYear > *minYear
}
