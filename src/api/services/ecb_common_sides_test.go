package services

import "testing"

func TestEcbCommonSideImage(t *testing.T) {
	const img = "https://www.ecb.europa.eu/euro/coins/common/shared/img/"
	cases := []struct {
		denomination string
		year         int
		country      string
		want         string
	}{
		{"2 euro", 2022, "Portugal", img + "newcommon_2euro_800.jpg"},
		{"2 Euros", 2005, "Alemanha", img + "common_2euro_800.jpg"},
		{"2 euro", 2007, "Alemanha", img + "newcommon_2euro_800.jpg"},
		{"2 euro", 2007, "Portugal", img + "common_2euro_800.jpg"}, // Portugal: new map from 2008
		{"1 euro", 2023, "Croácia", img + "newcommon_1euro_800.jpg"},
		{"50 cêntimos", 2010, "Espanha", img + "newcommon_50cent_800.jpg"},
		{"10 cêntimos de euro", 2003, "França", img + "10common.jpg"},
		{"50 Euro Cents", 2015, "Países Baixos", img + "newcommon_50cent_800.jpg"},
		{"5 cêntimos", 2019, "Itália", img + "common_5cent_800.jpg"},
		{"1 Euro Cent", 2002, "Irlanda", img + "common_1cent_800.jpg"},
		{"2 euro", 1999, "Bélgica", img + "common_2euro_800.jpg"},
		// Not euro coins of the eight denominations.
		{"5 euro", 2022, "Portugal", ""},
		{"2½ euro", 2009, "Portugal", ""},
		{"2 cents", 2005, "Malta", ""},        // Maltese lira cents before 2008
		{"10 centai", 2010, "Lituânia", ""},   // before the euro (2015)
		{"20 centimes", 2000, "França", ""},   // a franc coin
		{"5 cent", 2000, "Países Baixos", ""}, // a guilder coin
		{"50 pence", 2010, "Reino Unido", ""}, // not in the euro area
		{"1 escudo", 1999, "Portugal", ""},
	}
	for _, c := range cases {
		got, page := EcbCommonSideImage(c.denomination, c.year, c.country)
		if got != c.want || (got != "" && page != ecbCommonSidesPage) {
			t.Errorf("EcbCommonSideImage(%q, %d, %q) = %q, want %q", c.denomination, c.year, c.country, got, c.want)
		}
	}
}
