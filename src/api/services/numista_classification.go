package services

import (
	"sort"
	"strings"
)

// Numista sorts coin types into kinds ("Moedas circulantes comuns",
// "Circulating commemorative coins"…), written in the language the type was
// asked in. The usual ones get a key of their own so both languages count
// together; any other kind is keyed by its label.
const (
	ClassificationStandard                 = "standard"
	ClassificationCirculatingCommemorative = "circulating-commemorative"
	ClassificationNonCirculating           = "non-circulating"
)

var classificationOrder = map[string]int{
	ClassificationStandard:                 0,
	ClassificationCirculatingCommemorative: 1,
	ClassificationNonCirculating:           2,
}

// NumistaClassificationKey returns the key of a Numista kind of coin, or ""
// when there is no label.
func NumistaClassificationKey(label string) string {
	label = strings.Join(strings.Fields(label), " ")
	folded := strings.ReplaceAll(foldASCII(label), "-", " ")
	switch {
	case label == "":
		return ""
	case strings.Contains(folded, "nao circulante"), strings.Contains(folded, "non circulat"), strings.Contains(folded, "no circulante"):
		return ClassificationNonCirculating
	case strings.Contains(folded, "comemorativ"), strings.Contains(folded, "commemorativ"), strings.Contains(folded, "conmemorativ"):
		return ClassificationCirculatingCommemorative
	case strings.Contains(folded, "circulantes comuns"), strings.Contains(folded, "standard circulation"), strings.Contains(folded, "circulation courante"):
		return ClassificationStandard
	}
	return label
}

// CoinClassification is one kind of coin and the collection's coins of it.
type CoinClassification struct {
	Key string
	// Label is Numista's name for it, as first found.
	Label   string
	CoinIDs []uint
}

// SortedClassifications lists the kinds: the usual ones first, then the
// others by name.
func (c CoinCirculation) SortedClassifications() []CoinClassification {
	list := make([]CoinClassification, 0, len(c.Classifications))
	for _, group := range c.Classifications {
		list = append(list, *group)
	}
	sort.Slice(list, func(i, j int) bool {
		oi, iKnown := classificationOrder[list[i].Key]
		oj, jKnown := classificationOrder[list[j].Key]
		switch {
		case iKnown && jKnown:
			return oi < oj
		case iKnown != jKnown:
			return iKnown
		}
		return list[i].Label < list[j].Label
	})
	return list
}

// ClassificationCoins returns the coins of one kind; an unknown kind has none.
func (c CoinCirculation) ClassificationCoins(key string) []uint {
	if group := c.Classifications[key]; group != nil {
		return group.CoinIDs
	}
	return []uint{}
}
