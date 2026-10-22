package models

import "testing"

func TestDeriveCoinYear(t *testing.T) {
	cases := []struct {
		in   string
		want *int
	}{
		{"1968", intPtr(1968)},
		{"27 BC – AD 14", intPtr(-27)},
		{"27 BC - AD 14", intPtr(-27)},
		{"c. 300 a.C.", intPtr(-300)},
		{"44 a.C.–14 d.C.", intPtr(-44)},
		{"c. 46-44 BC", intPtr(-46)},
		{"46–44 a.C.", intPtr(-46)},
		{"100 BC – AD 50", intPtr(-100)},
		{"AD 14 – 37", intPtr(14)},
		{"14 AD - 37 AD", intPtr(14)},
		{"AD 312", intPtr(312)},
		{"1953–1974", intPtr(1953)},
		{"253-268", intPtr(253)},
		{"1st century BC", nil},
		{"séc. XV", nil},
		{"Imperial", nil},
		{"", nil},
	}
	for _, tc := range cases {
		got := DeriveCoinYear(tc.in)
		if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
			t.Errorf("DeriveCoinYear(%q) = %v, want %v", tc.in, deref(got), deref(tc.want))
		}
	}
}

func TestCoinPresentationAndRarityValidation(t *testing.T) {
	for _, value := range []string{"", "coincard", "slab", "loose"} {
		if !ValidCoinPresentation(value) {
			t.Errorf("presentation %q should be valid", value)
		}
	}
	for _, value := range []string{"Coincard", "box", " coincard"} {
		if ValidCoinPresentation(value) {
			t.Errorf("presentation %q should be invalid", value)
		}
	}
	if !ValidCoinRarity("very_rare") || ValidCoinRarity("R2") {
		t.Fatal("rarity validation")
	}
	if got := NormalizeCoinCountry("  São   Tomé  "); got != "São Tomé" {
		t.Fatalf("NormalizeCoinCountry = %q", got)
	}
}

func intPtr(v int) *int { return &v }

func deref(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func TestCoinCirculationFacts(t *testing.T) {
	for value, want := range map[string]CoinCirculationSetting{
		"current":                       {},
		"current_commemorative":         {Commemorative: true},
		"demonetized":                   {Demonetized: true},
		"demonetized_commemorative":     {Demonetized: true, Commemorative: true},
		"non_circulating":               {NonCirculating: true},
		"non_circulating_commemorative": {NonCirculating: true, Commemorative: true},
	} {
		got, ok := CoinCirculationFacts(value)
		if !ok || got != want || !ValidCoinCirculation(value) {
			t.Errorf("CoinCirculationFacts(%q) = %+v, %v", value, got, ok)
		}
	}
	if _, ok := CoinCirculationFacts(""); ok || !ValidCoinCirculation("") {
		t.Error("an empty circulation is not set by hand, and is valid")
	}
	if _, ok := CoinCirculationFacts("legal"); ok || ValidCoinCirculation("legal") {
		t.Error("an unknown circulation is not valid")
	}
}

func TestMonetarySystemFromNumista(t *testing.T) {
	cases := []struct{ name, fullName, want string }{
		{"Escudo", "Escudo (1911-2001)", "Escudo (1911-2001)"},
		{"Real", "Real (decimalizado, 1835-1910)", "Real (decimalizado, 1835-1910)"},
		{"Real", "Real (1994-hoje)", "Real (1994-hoje)"},
		{"Euro", "Euro (2002-hoje)", "Euro"},
		{"Euro", "Euro (2023-hoje)", "Euro"},
		{" euro ", "", "Euro"},
		{"Peseta", "", "Peseta"},
		{"Libra  esterlina", " Libra esterlina  (decimalizada, 1971-hoje) ", "Libra esterlina (decimalizada, 1971-hoje)"},
		{"", "", ""},
	}
	for _, tc := range cases {
		if got := MonetarySystemFromNumista(tc.name, tc.fullName); got != tc.want {
			t.Errorf("MonetarySystemFromNumista(%q, %q) = %q, want %q", tc.name, tc.fullName, got, tc.want)
		}
	}
}
