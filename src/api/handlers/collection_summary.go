package handlers

import (
	"math"
	"net/http"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CollectionSummaryHandler serves the collection's headline numbers.
type CollectionSummaryHandler struct {
	coins    *repository.CoinRepository
	rates    *services.ExchangeRateService
	settings *services.SettingsService
}

// NewCollectionSummaryHandler creates a new CollectionSummaryHandler.
func NewCollectionSummaryHandler(coins *repository.CoinRepository) *CollectionSummaryHandler {
	return &CollectionSummaryHandler{coins: coins}
}

// WithExchangeRates adds the collection value in euros.
func (h *CollectionSummaryHandler) WithExchangeRates(rates *services.ExchangeRateService) *CollectionSummaryHandler {
	h.rates = rates
	return h
}

// WithSettings adds the collection currency, the currency amounts are kept in.
func (h *CollectionSummaryHandler) WithSettings(settings *services.SettingsService) *CollectionSummaryHandler {
	h.settings = settings
	return h
}

// CollectionSummaryResponse is the collection summary, its value in euros and
// its rarest, oldest and most valuable coins.
type CollectionSummaryResponse struct {
	repository.CollectionSummary
	Highlights repository.CollectionHighlights `json:"highlights"`
	// Currency is the collection currency (USD or EUR): TotalValue and the
	// highlights' values are in it.
	Currency string `json:"currency" enums:"USD,EUR"`
	// TotalValueEUR is TotalValue in euros: as is for a euro collection, else
	// converted with ExchangeRate; null when no rate is available.
	TotalValueEUR *float64 `json:"totalValueEur"`
	// ExchangeRate is the latest ECB rate, when available.
	ExchangeRate *services.ExchangeRate `json:"exchangeRate,omitempty"`
}

// Summary returns the number of coins, coin types, countries and coincards in
// the collection, its total value in euros and its notable coins.
//
//	@Summary		Get collection summary
//	@Description	Counts the user's collection coins (excluding wishlist and sold), their distinct types (coins linked to the same Numista type count once; other coins once per distinct name), their distinct countries, how many coins have no country, and how many coins are in a coincard. Also sums the coins' current values (kept in the collection currency, USD or EUR) and gives the total in euros, converted from dollars with the ECB daily reference rate (cached; null when unavailable), and names the rarest (fixed rarity scale; ties go to the more valuable), oldest (lowest year) and most valuable (current value) coins.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	CollectionSummaryResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/summary [get]
func (h *CollectionSummaryHandler) Summary(c *gin.Context) {
	summary, err := h.coins.CollectionSummary(c.GetUint("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to summarise the collection"})
		return
	}
	highlights, err := h.coins.CollectionHighlights(c.GetUint("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to summarise the collection"})
		return
	}
	response := CollectionSummaryResponse{CollectionSummary: summary, Highlights: highlights, Currency: h.settings.CollectionCurrency()}
	if response.Currency == services.CurrencyEUR {
		total := math.Round(summary.TotalValue*100) / 100
		response.TotalValueEUR = &total
	}
	if h.rates != nil {
		if rate, err := h.rates.USDPerEUR(c.Request.Context()); err == nil {
			response.ExchangeRate = &rate
			if response.TotalValueEUR == nil {
				total := math.Round(summary.TotalValue/rate.USDPerEUR*100) / 100
				response.TotalValueEUR = &total
			}
		}
	}
	c.JSON(http.StatusOK, response)
}
