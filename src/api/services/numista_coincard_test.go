package services

import "testing"

func TestIssuedInCoincard(t *testing.T) {
	for comments, want := range map[string]bool{
		"(en) <p>Issued in a coincard</p>":                                 true,
		"<p>Mintage: 25 000</p><p>(en) Issued in a Coincard</p>":           true,
		"<p>Emitido num cartão de moedas.</p>":                             true,
		"Moeda no interior do cartão de apresentação oficial. Fotografias": true,
		"Mais 10 000 UN em cartão de moedas<br />":                         false,
		"<p>Also issued in a coincard for the anniversary.</p>":            false,
		"Cartão BU":                      false,
		"Cartão de prova":                false,
		"Cartão de moedas: Prova: 1 500": false,
		"A moeda foi emitida em carteiras e com qualidade BNC num blister.": false,
		"Cartão da moeda © papilio (CC BY-NC-SA )":                          false,
		"Cartão oficial com o conjunto completo, incluindo 1 moeda":         false,
		"": false,
	} {
		if got := IssuedInCoincard(comments); got != want {
			t.Errorf("IssuedInCoincard(%q) = %v, want %v", comments, got, want)
		}
	}
}
