package services

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// MetalTotal is how much of one precious metal a collection holds, each
// coin times its quantity: Fine sums the pure metal of the coins whose
// fineness is known; Gross is the weight of all of them, alloy included.
type MetalTotal struct {
	FineGrams  float64 `json:"fineGrams"`
	GrossGrams float64 `json:"grossGrams"`
	Pieces     int64   `json:"pieces"`
	// UnknownFineness counts the pieces whose fineness nobody says: their
	// weight is in Gross, not in Fine.
	UnknownFineness int64 `json:"unknownFineness"`
}

// PreciousMetals is the gold and silver of a collection.
type PreciousMetals struct {
	Gold   MetalTotal `json:"gold"`
	Silver MetalTotal `json:"silver"`
}

// PreciousMetalsService works out the gold and silver in a collection from
// its coins' material and weight, and how fine the metal is from their
// Numista types' composition ("Prata 925‰", "Silver (.900)", "Bilhão
// (prata 333‰)"), as cached when they were looked up. It never asks Numista.
// A plated coin is not of the metal it is plated with.
type PreciousMetalsService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService
}

// NewPreciousMetalsService creates the service.
func NewPreciousMetalsService(coins *repository.CoinRepository, catalog *NumistaCatalogService) *PreciousMetalsService {
	return &PreciousMetalsService{coins: coins, catalog: catalog}
}

var (
	platedComposition = regexp.MustCompile(`(?i)banhad|plated|plaqu|chapead|folhead|clad`)
	goldComposition   = regexp.MustCompile(`(?i)\b(gold|ouro|or)\b`)
	silverComposition = regexp.MustCompile(`(?i)\b(silver|prata|argent|billon|bilh[aã]o)\b`)
	// 925‰, 916,66 ‰
	perMille = regexp.MustCompile(`(\d{2,4}(?:[.,]\d+)?)\s*‰`)
	// (.900), 0.925, (0,999)
	decimalFineness = regexp.MustCompile(`(?:^|[\s(])0?[.,](\d{2,4})\b`)
)

// Fineness reads how fine the precious metal of a composition is (0.925 for
// "Prata 925‰"); ok is false when it does not say.
func Fineness(composition string) (float64, bool) {
	if match := perMille.FindStringSubmatch(composition); match != nil {
		if value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", "."), 64); err == nil && value > 0 && value <= 1000 {
			return value / 1000, true
		}
	}
	if match := decimalFineness.FindStringSubmatch(composition); match != nil {
		if value, err := strconv.ParseFloat("0."+match[1], 64); err == nil && value > 0 {
			return value, true
		}
	}
	return 0, false
}

// Summarise sums the gold and silver of the user's collection coins (not
// the wishlist's).
func (s *PreciousMetalsService) Summarise(userID uint) (PreciousMetals, error) {
	rows, err := s.coins.CollectionMetalRows(userID)
	if err != nil {
		return PreciousMetals{}, err
	}
	var result PreciousMetals
	for _, row := range rows {
		composition := s.composition(row.NumistaTypeID)
		if platedComposition.MatchString(composition) {
			continue
		}
		total := &result.Silver
		switch {
		case goldComposition.MatchString(composition):
			total = &result.Gold
		case silverComposition.MatchString(composition):
		case row.Material == string(models.MaterialGold):
			total = &result.Gold
		}
		weight := row.WeightGrams * float64(row.Quantity)
		total.GrossGrams += weight
		total.Pieces += row.Quantity
		if fineness, ok := Fineness(composition); ok {
			total.FineGrams += weight * fineness
		} else {
			total.UnknownFineness += row.Quantity
		}
	}
	return result, nil
}

// composition is what the coin's Numista type is made of, as cached; empty
// when it is not linked or not cached.
func (s *PreciousMetalsService) composition(numistaTypeID string) string {
	typeID, err := strconv.Atoi(strings.TrimSpace(numistaTypeID))
	if err != nil || typeID <= 0 || s.catalog == nil {
		return ""
	}
	if detail := s.catalog.cachedTypeDetail(typeID); detail != nil {
		return detail.Composition
	}
	return ""
}
