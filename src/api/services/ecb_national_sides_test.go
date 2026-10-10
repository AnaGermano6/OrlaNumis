package services

import "testing"

func TestEcbNationalSideImage(t *testing.T) {
	const img = "https://www.ecb.europa.eu/euro/coins/common/shared/img/"
	cases := []struct {
		denomination string
		year         int
		country      string
		want         string
		page         string
	}{
		{"2 euro", 2010, "Portugal", img + "pt/Portugal_2euro.jpg", "https://www.ecb.europa.eu/euro/coins/html/pt.en.html"},
		{"10 cêntimos", 2002, "Portugal", img + "pt/Portugal_10cent.jpg", "https://www.ecb.europa.eu/euro/coins/html/pt.en.html"},
		// One image per series: Beatrix, then Willem-Alexander.
		{"1 euro", 2013, "Países Baixos", img + "nl/Netherlands_1euro_2000.jpg", "https://www.ecb.europa.eu/euro/coins/html/nl.en.html"},
		{"1 euro", 2014, "Netherlands", img + "nl/Netherlands_1euro_2014.jpg", "https://www.ecb.europa.eu/euro/coins/html/nl.en.html"},
		{"2 euro", 2009, "Bélgica", img + "be/Belgium_2euro_2008.jpg", "https://www.ecb.europa.eu/euro/coins/html/be.en.html"},
		{"1 euro", 2000, "Belgium", img + "be/Belgium_1999_1euro.gif", "https://www.ecb.europa.eu/euro/coins/html/be.en.html"},
		{"50 cent", 2024, "França", img + "fr/France_50cent_2024.jpg", "https://www.ecb.europa.eu/euro/coins/html/fr.en.html"},
		{"5 cent", 2024, "França", img + "fr/France_5cent.jpg", "https://www.ecb.europa.eu/euro/coins/html/fr.en.html"},
		{"2 euro", 2016, "Espanha", img + "es/Spain_2euro_2015.jpg", "https://www.ecb.europa.eu/euro/coins/html/es.en.html"},
		{"20 cent", 2016, "Espanha", img + "es/Spain_20cent.jpg", "https://www.ecb.europa.eu/euro/coins/html/es.en.html"},
		// 2010 is Pope Benedict XVI's series.
		{"1 euro", 2010, "Vaticano", img + "va/Vatican_1euro_2013.jpg", ""},
		// No design on the ECB site: before San Marino's 2008 series, the
		// Vatican's two 2005 series, its coins after 2024, Monaco.
		{"1 euro", 2004, "San Marino", "", ""},
		{"1 euro", 2005, "Vaticano", "", ""},
		{"1 euro", 2025, "Vaticano", "", ""},
		{"2 euro", 2010, "Mónaco", "", ""},
		// Not a regular euro coin.
		{"5 euro", 2021, "Portugal", "", ""},
		{"100 escudos", 1999, "Portugal", "", ""},
	}
	for _, c := range cases {
		got, page := EcbNationalSideImage(c.denomination, c.year, c.country)
		if got != c.want || (c.page != "" && page != c.page) {
			t.Errorf("EcbNationalSideImage(%q, %d, %q) = %q, %q; want %q", c.denomination, c.year, c.country, got, page, c.want)
		}
	}
}

func TestRegularEuroType(t *testing.T) {
	year := func(y int) *int { return &y }
	cases := []struct {
		title    string
		min, max *int
		want     bool
	}{
		{"2 Euros (2nd map)", year(2008), year(2026), true},
		{"1 Euro Cent", year(2002), year(2026), true},
		{"2 Euro Cents", year(2026), year(2026), true},
		{"10 Euro Cents (Simone Veil)", year(2024), year(2026), true},
		{"2 Euros (1st map)", year(2002), year(2002), true},
		{"2 Euros (Thuringia)", year(2022), year(2022), false},
		{"2 Euros", nil, nil, false},
	}
	for _, c := range cases {
		if got := RegularEuroType(c.title, c.min, c.max); got != c.want {
			t.Errorf("RegularEuroType(%q) = %v, want %v", c.title, got, c.want)
		}
	}
}
