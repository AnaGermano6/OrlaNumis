package services

import (
	"context"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// coinCountryBatch bounds Numista calls per request; issuer details are
// cached for 30 days, so repeat types cost nothing.
const coinCountryBatch = 25

// CoinCountryService fills a coin's country from the issuer of its linked
// Numista type.
type CoinCountryService struct {
	coins   *repository.CoinRepository
	catalog *NumistaCatalogService
}

// NewCoinCountryService creates the service.
func NewCoinCountryService(coins *repository.CoinRepository, catalog *NumistaCatalogService) *CoinCountryService {
	return &CoinCountryService{coins: coins, catalog: catalog}
}

// FillFromNumista sets the country of the user's coins (or only coinID when
// non-zero) that have none but are linked to Numista, at most a batch per
// call. It stops early when Numista is not configured or the quota is spent.
func (s *CoinCountryService) FillFromNumista(ctx context.Context, userID, coinID uint, lang string) models.CoinCountryFillResult {
	result := models.CoinCountryFillResult{Status: models.NumistaDataStatusOK}
	coins, total, err := s.coins.FindMissingCountryWithNumista(userID, coinID, coinCountryBatch)
	if err != nil {
		result.Status = models.NumistaDataStatusUnavailable
		return result
	}
	for i := range coins {
		typeID := NumistaTypeIDForCoin(&coins[i])
		if typeID == 0 {
			result.Failed++
			continue
		}
		detail, err := s.catalog.typeDetail(ctx, typeID, lang)
		if err != nil {
			status := numistaStatusForError(err)
			if status == models.NumistaDataStatusUnconfigured || status == models.NumistaDataStatusUnauthorized || status == models.NumistaDataStatusQuotaLimited {
				result.Status = status
				break
			}
			result.Failed++
			continue
		}
		country := models.NormalizeCoinCountry(detail.Issuer.Name)
		if country == "" {
			result.Failed++
			continue
		}
		if err := s.coins.UpdateCountry(coins[i].ID, country); err != nil {
			result.Failed++
			continue
		}
		result.Updated++
	}
	result.Remaining = max(int(total)-result.Updated, 0)
	return result
}
