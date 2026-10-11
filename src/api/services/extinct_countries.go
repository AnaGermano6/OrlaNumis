package services

import (
	"regexp"
	"strings"
	"unicode"
)

// extinctCountryNames are countries that no longer exist (dissolved, split,
// absorbed, or colonies and empires that ended), as Numista, uCoin and people
// write them in Portuguese and English, folded (lower case, no accents or
// punctuation). A country that only changed its name (Ceylon, Burma) is not
// extinct, nor one that carried on as the same state (West Germany).
var extinctCountryNames = map[string]bool{}

func init() {
	for _, name := range []string{
		// Dissolved or split in the 20th century.
		"União Soviética", "União das Repúblicas Socialistas Soviéticas", "URSS", "Soviet Union", "USSR", "Union of Soviet Socialist Republics",
		"Jugoslávia", "Iugoslávia", "Reino da Jugoslávia", "Yugoslavia", "Kingdom of Yugoslavia",
		"Sérvia e Montenegro", "Serbia and Montenegro",
		"Checoslováquia", "Tchecoslováquia", "Czechoslovakia",
		"Alemanha Oriental", "República Democrática Alemã", "RDA", "East Germany", "German Democratic Republic", "GDR", "DDR",
		"Antilhas Holandesas", "Antilhas Neerlandesas", "Netherlands Antilles",
		"República Árabe Unida", "United Arab Republic",
		"Iémen do Sul", "Iêmen do Sul", "South Yemen", "Iémen do Norte", "Iêmen do Norte", "North Yemen",
		"Vietname do Sul", "Vietnã do Sul", "South Vietnam",
		"Zaire", "Rodésia", "Rhodesia", "Rodésia do Sul", "Southern Rhodesia",
		"Rodésia e Niassalândia", "Federação da Rodésia e Niassalândia", "Rhodesia and Nyasaland",
		"Biafra", "Katanga", "Manchukuo", "Tannu Tuva",
		// Empires and kingdoms.
		"Império Otomano", "Ottoman Empire", "Império Russo", "Russian Empire",
		"Áustria-Hungria", "Império Austro-Húngaro", "Austria-Hungary", "Austro-Hungarian Empire", "Império Austríaco", "Austrian Empire",
		"Império Alemão", "German Empire", "Prússia", "Prussia", "Baviera", "Bavaria", "Saxónia", "Saxônia", "Saxony", "Württemberg", "Hanôver", "Hanover",
		"Estados Papais", "Estados Pontifícios", "Papal States",
		"Reino das Duas Sicílias", "Duas Sicílias", "Two Sicilies", "Kingdom of the Two Sicilies",
		"Reino da Sardenha", "Kingdom of Sardinia",
		"Reino do Havai", "Reino do Havaí", "Kingdom of Hawaii", "Havai", "Havaí", "Hawaii",
		"Terra Nova", "Newfoundland",
		"Transvaal", "República Sul-Africana (ZAR)", "Estado Livre de Orange", "Orange Free State",
		"Cidade Livre de Danzig", "Danzig", "Free City of Danzig", "Sarre", "Saar", "Saarland", "Protetorado do Sarre",
		// Colonies and colonial federations that ended.
		"Indochina Francesa", "French Indochina",
		"África Ocidental Francesa", "French West Africa", "África Equatorial Francesa", "French Equatorial Africa",
		"África Oriental Britânica", "British East Africa", "África Ocidental Britânica", "British West Africa",
		"Índias Orientais Holandesas", "Índias Orientais Neerlandesas", "Netherlands East Indies", "Dutch East Indies",
		"Índia Britânica", "British India", "Índia Portuguesa", "Portuguese India",
		"Estabelecimentos do Estreito", "Straits Settlements",
	} {
		extinctCountryNames[foldCountry(name)] = true
	}
}

var countryParentheses = regexp.MustCompile(`\([^)]*\)`)

func foldCountry(name string) string {
	return strings.Join(strings.FieldsFunc(foldASCII(name), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }), " ")
}

// IsExtinctCountry tells whether a coin's country no longer exists. Numista's
// forms count too: "Jugoslávia, República Federal Socialista da" by the name
// before the comma, "União Soviética (URSS)" with or without the brackets.
func IsExtinctCountry(country string) bool {
	country = strings.TrimSpace(country)
	if country == "" {
		return false
	}
	candidates := []string{country, countryParentheses.ReplaceAllString(country, " ")}
	if head, _, found := strings.Cut(country, ","); found {
		candidates = append(candidates, head, countryParentheses.ReplaceAllString(head, " "))
	}
	for _, match := range countryParentheses.FindAllString(country, -1) {
		candidates = append(candidates, strings.Trim(match, "()"))
	}
	for _, candidate := range candidates {
		if extinctCountryNames[foldCountry(candidate)] {
			return true
		}
	}
	return false
}
