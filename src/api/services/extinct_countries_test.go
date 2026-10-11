package services

import "testing"

func TestIsExtinctCountry(t *testing.T) {
	for country, want := range map[string]bool{
		"Jugoslávia": true,
		"Jugoslávia, República Federal Socialista da": true,
		"União Soviética (URSS)":                      true,
		"URSS":                                        true,
		"Checoslováquia":                              true,
		"Czechoslovakia":                              true,
		"Antilhas Holandesas":                         true,
		"Antilhas Neerlandesas":                       true,
		"Império Otomano":                             true,
		"Indochina Francesa":                          true,
		"  república democrática alemã ":              true,
		"East Germany":                                true,
		// Countries that still exist, or only changed their name.
		"Alemanha":             false,
		"Alemanha (1871-1948)": false,
		"Portugal":             false,
		"Rússia":               false,
		"Ceilão":               false,
		"Sri Lanka":            false,
		"Eslovénia ":           false,
		"":                     false,
	} {
		if got := IsExtinctCountry(country); got != want {
			t.Errorf("IsExtinctCountry(%q) = %v, want %v", country, got, want)
		}
	}
}
