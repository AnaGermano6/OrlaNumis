package services

import "testing"

func TestNumistaClassificationKey(t *testing.T) {
	for label, want := range map[string]string{
		"Moedas circulantes comuns":        ClassificationStandard,
		"Standard circulation coins":       ClassificationStandard,
		"Pièces de circulation courante":   ClassificationStandard,
		"Moedas circulantes comemorativas": ClassificationCirculatingCommemorative,
		"Circulating commemorative coins":  ClassificationCirculatingCommemorative,
		"Moedas não circulantes":           ClassificationNonCirculating,
		"Non-circulating coins":            ClassificationNonCirculating,
		"  Patterns ":                      "Patterns",
		"":                                 "",
	} {
		if got := NumistaClassificationKey(label); got != want {
			t.Errorf("NumistaClassificationKey(%q) = %q, want %q", label, got, want)
		}
	}
}

func TestCoinCirculation_SortedClassifications(t *testing.T) {
	sorted := CoinCirculation{Classifications: map[string]*CoinClassification{
		"Tokens":                     {Key: "Tokens", Label: "Tokens", CoinIDs: []uint{9}},
		ClassificationNonCirculating: {Key: ClassificationNonCirculating, Label: "Moedas não circulantes", CoinIDs: []uint{3}},
		"Patterns":                   {Key: "Patterns", Label: "Patterns", CoinIDs: []uint{8}},
		ClassificationStandard:       {Key: ClassificationStandard, Label: "Moedas circulantes comuns", CoinIDs: []uint{1, 2}},
	}}
	keys := []string{}
	for _, group := range sorted.SortedClassifications() {
		keys = append(keys, group.Key)
	}
	if want := []string{ClassificationStandard, ClassificationNonCirculating, "Patterns", "Tokens"}; len(keys) != len(want) || keys[0] != want[0] || keys[1] != want[1] || keys[2] != want[2] || keys[3] != want[3] {
		t.Fatalf("order = %v, want %v", keys, want)
	}
	if ids := sorted.ClassificationCoins("Medals"); ids == nil || len(ids) != 0 {
		t.Fatalf("an unknown kind has no coins, got %v", ids)
	}
}
