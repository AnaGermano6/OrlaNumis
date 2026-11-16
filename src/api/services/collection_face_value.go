package services

import (
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/briandenicola/ancient-coins-api/repository"
)

// escudosPerEuro is the rate the escudo was fixed at when the euro replaced
// it (Council Regulation (EC) No 2866/98): 200.482 escudos to the euro.
const escudosPerEuro = 200.482

// CollectionFaceValue is the face value of a collection's items, in euros.
// Only face values in euros or in escudos (at their fixed rate) count: other
// currencies have no rate to the euro that means anything for old coins.
type CollectionFaceValue struct {
	// EUR sums the known face values, each times its quantity.
	EUR float64
	// Without counts the items held, with their quantities, whose face value
	// is not known in euros.
	Without int64
	// ValuedEUR is the face value of the items that also have a current
	// value, and Value their current value (in the collection currency), to
	// compare the two on the same items.
	ValuedEUR float64
	Value     float64
}

// CollectionFaceValueService works out the face value of a collection's
// items: from their Numista types (as cached when they were looked up), else
// from their denominations. It never asks Numista.
type CollectionFaceValueService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService

	mu    sync.Mutex
	known map[int]NumistaTypeFacts
}

// NewCollectionFaceValueService creates the service.
func NewCollectionFaceValueService(coins *repository.CoinRepository, catalog *NumistaCatalogService) *CollectionFaceValueService {
	return &CollectionFaceValueService{coins: coins, catalog: catalog, known: map[int]NumistaTypeFacts{}}
}

// Summarise sums the face values of the user's collection items (not the
// wishlist's) of a kind ("" for all).
func (s *CollectionFaceValueService) Summarise(userID uint, kind string) (CollectionFaceValue, error) {
	rows, err := s.coins.CollectionFaceValueRows(userID, kind)
	if err != nil {
		return CollectionFaceValue{}, err
	}
	var result CollectionFaceValue
	for _, row := range rows {
		euros, ok := s.faceValueEUR(row)
		if !ok {
			result.Without += row.Quantity
			continue
		}
		result.EUR += euros * float64(row.Quantity)
		if row.CurrentValue != nil {
			result.ValuedEUR += euros * float64(row.Quantity)
			result.Value += *row.CurrentValue * float64(row.Quantity)
		}
	}
	return result, nil
}

// faceValueEUR is one item's face value in euros: its Numista type's, else
// its denomination's in its monetary system (a denomination in euros needs
// none).
func (s *CollectionFaceValueService) faceValueEUR(row repository.CoinFaceValueRow) (float64, bool) {
	amount, system := 0.0, strings.TrimSpace(row.MonetarySystem)
	if facts, ok := s.typeFacts(row.NumistaTypeID); ok && facts.FaceValue > 0 {
		amount = facts.FaceValue
		if facts.MonetarySystem != "" {
			system = facts.MonetarySystem
		}
	} else {
		amount = ParseFaceValue(row.Denomination)
		if system == "" && faceValueInEuros.MatchString(foldASCII(row.Denomination)) {
			system = "Euro"
		}
	}
	perUnit, ok := eurosPerUnit(system)
	if !ok || amount <= 0 {
		return 0, false
	}
	return amount * perUnit, true
}

func (s *CollectionFaceValueService) typeFacts(numistaTypeID string) (NumistaTypeFacts, bool) {
	typeID, err := strconv.Atoi(strings.TrimSpace(numistaTypeID))
	if err != nil || typeID <= 0 || s.catalog == nil {
		return NumistaTypeFacts{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if facts, ok := s.known[typeID]; ok {
		return facts, true
	}
	facts, ok := s.catalog.CachedTypeFacts(typeID)
	if ok {
		s.known[typeID] = facts
	}
	return facts, ok
}

// eurosPerUnit is what one unit of a monetary system is worth in euros: the
// euro, and the escudo that it replaced at a fixed rate.
func eurosPerUnit(monetarySystem string) (float64, bool) {
	switch system := strings.Join(strings.Fields(monetarySystem), " "); {
	case strings.EqualFold(system, "Euro"):
		return 1, true
	case strings.EqualFold(system, "Escudo (1911-2001)"):
		return 1 / escudosPerEuro, true
	}
	return 0, false
}

var (
	// A number with its thousands ("1.000", "1 000"), then its decimals
	// ("2,5"), or escudo cents after the sign ("2$50", "1000$00").
	faceValueNumber = regexp.MustCompile(`(\d{1,3}(?:[. ]\d{3})+|\d+)(?:\$(\d{1,2})|[.,](\d+))?`)
	// Cents of the main unit: cêntimos, centavos, centimes, centesimi, cents.
	faceValueCents   = regexp.MustCompile(`\b(cent|cents|centimos?|centavos?|centimes?|centesimi|centesimo|ct|cts)\b`)
	faceValueInEuros = regexp.MustCompile(`\beuros?\b|€`)
)

// ParseFaceValue reads a denomination's face value in its currency's main
// unit: 2 for "2 Euros", 0.5 for "50 Cêntimos", 2.5 for "2,5 Escudos" or
// "2$50", 0.5 for "½ Euro"; 0 when it says none.
func ParseFaceValue(denomination string) float64 {
	text := foldASCII(denomination)
	value := 0.0
	if match := faceValueNumber.FindStringSubmatch(text); match != nil {
		value, _ = strconv.ParseFloat(strings.NewReplacer(".", "", " ", "").Replace(match[1]), 64)
		switch {
		case match[2] != "":
			cents, _ := strconv.ParseFloat(match[2], 64)
			value += cents / 100
		case match[3] != "":
			part, _ := strconv.ParseFloat("0."+match[3], 64)
			value += part
		}
	}
	switch {
	case strings.Contains(denomination, "½"):
		value += 0.5
	case strings.Contains(denomination, "¼"):
		value += 0.25
	}
	if faceValueCents.MatchString(text) {
		value /= 100
	}
	return value
}
