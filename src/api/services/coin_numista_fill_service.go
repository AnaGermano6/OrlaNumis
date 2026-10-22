package services

import (
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// CoinNumistaFillService completes coin records from the Numista types the
// app has already cached: the monetary system (of coins and banknotes), and,
// for coins, the material (from the composition), the weight, the diameter,
// and a coin card as presentation when Numista says the type is issued in
// one. It only fills what is empty (a material of Other counts as empty) and
// never asks Numista, so it costs none of the monthly quota.
// Materials that are set but differ from the composition are listed apart,
// as corrections, and change only when asked for.
type CoinNumistaFillService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService
}

// NewCoinNumistaFillService creates the service.
func NewCoinNumistaFillService(coins *repository.CoinRepository, catalog *NumistaCatalogService) *CoinNumistaFillService {
	return &CoinNumistaFillService{coins: coins, catalog: catalog}
}

// Preview tells what a fill would change, without changing anything.
func (s *CoinNumistaFillService) Preview(userID uint) (models.CoinNumistaFillResult, error) {
	result, _, err := s.plan(userID)
	return result, err
}

// Apply fills the coins, with the corrections when withCorrections is
// true, and tells what it changed.
func (s *CoinNumistaFillService) Apply(userID uint, withCorrections bool) (models.CoinNumistaFillResult, error) {
	result, plan, err := s.plan(userID)
	if err != nil {
		return result, err
	}
	fields := plan.fills
	if withCorrections {
		for coinID, material := range plan.corrections {
			if fields[coinID] == nil {
				fields[coinID] = map[string]any{}
			}
			fields[coinID]["material"] = material
		}
		result.Corrected = len(plan.corrections)
	}
	if err := s.coins.UpdateCoinFields(userID, fields); err != nil {
		return result, err
	}
	result.Applied = true
	return result, nil
}

type numistaFillPlan struct {
	fills       map[uint]map[string]any
	corrections map[uint]string
}

func (s *CoinNumistaFillService) plan(userID uint) (models.CoinNumistaFillResult, numistaFillPlan, error) {
	result := models.CoinNumistaFillResult{
		ByMaterial:  map[string]int{},
		Changes:     []models.CoinNumistaFillChange{},
		Corrections: []models.CoinNumistaFillChange{},
	}
	plan := numistaFillPlan{fills: map[uint]map[string]any{}, corrections: map[uint]string{}}
	coins, err := s.coins.FindNumistaLinkedCoins(userID)
	if err != nil {
		return result, plan, err
	}
	for i := range coins {
		coin := &coins[i]
		typeID := NumistaTypeIDForCoin(coin)
		if typeID == 0 {
			continue
		}
		detail := s.catalog.cachedTypeDetail(typeID)
		if detail == nil {
			result.NotCached++
			continue
		}
		change := models.CoinNumistaFillChange{CoinID: coin.ID, Name: coin.Name, IsWishlist: coin.IsWishlist}
		columns := map[string]any{}
		if strings.TrimSpace(coin.MonetarySystem) == "" && detail.MonetarySystem != "" {
			change.MonetarySystem = detail.MonetarySystem
			columns["monetary_system"] = detail.MonetarySystem
			result.MonetarySystems++
		}
		if coin.Kind == models.CoinKindBanknote {
			// A banknote's material, size and presentation are not a coin's.
			if len(columns) > 0 {
				plan.fills[coin.ID] = columns
				result.Changes = append(result.Changes, change)
			}
			continue
		}
		material := MaterialFromComposition(detail.Composition)
		switch {
		case coin.Material == "" || coin.Material == models.MaterialOther:
			if material != models.MaterialOther {
				change.Material, change.Composition = material, detail.Composition
				columns["material"] = string(material)
				result.Materials++
				result.ByMaterial[string(material)]++
			} else if detail.Composition != "" {
				result.Unmatched++
			}
		case material != models.MaterialOther && material != coin.Material:
			result.Corrections = append(result.Corrections, models.CoinNumistaFillChange{
				CoinID: coin.ID, Name: coin.Name, IsWishlist: coin.IsWishlist,
				Material: material, PreviousMaterial: coin.Material, Composition: detail.Composition,
			})
			plan.corrections[coin.ID] = string(material)
		}
		if coin.WeightGrams == nil && detail.WeightGrams != nil {
			change.WeightGrams = detail.WeightGrams
			columns["weight_grams"] = *detail.WeightGrams
			result.Weights++
		}
		if coin.DiameterMm == nil && detail.DiameterMm != nil {
			change.DiameterMm = detail.DiameterMm
			columns["diameter_mm"] = *detail.DiameterMm
			result.Diameters++
		}
		if strings.TrimSpace(coin.Presentation) == "" && detail.IssuedInCoincard {
			change.Presentation = "coincard"
			columns["presentation"] = "coincard"
			result.Presentations++
		}
		if len(columns) > 0 {
			plan.fills[coin.ID] = columns
			result.Changes = append(result.Changes, change)
		}
	}
	result.Coins = len(result.Changes)
	return result, plan, nil
}
