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
	coins *repository.CoinRepository
	rates *services.ExchangeRateService
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

// CollectionSummaryResponse is the collection summary, its value in euros and
// its rarest, oldest and most valuable coins.
type CollectionSummaryResponse struct {
	repository.CollectionSummary
	Highlights repository.CollectionHighlights `json:"highlights"`
	// TotalValueEUR is TotalValue converted with ExchangeRate; null when no rate is available.
	TotalValueEUR *float64               `json:"totalValueEur"`
	ExchangeRate  *services.ExchangeRate `json:"exchangeRate,omitempty"`
}

// Summary returns the number of coins, coin types, countries and coincards in
// the collection, its total value in euros and its notable coins.
//
//	@Summary		Get collection summary
//	@Description	Counts the user's collection coins (excluding wishlist and sold), their distinct types (coins linked to the same Numista type count once; other coins once per distinct name), their distinct countries, how many coins have no country, and how many coins are in a coincard. Also sums the coins' current values (stored in US dollars) and converts the total to euros with the ECB daily reference rate (cached; null when unavailable), and names the rarest (fixed rarity scale; ties go to the more valuable), oldest (lowest year) and most valuable (current value) coins.
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
	response := CollectionSummaryResponse{CollectionSummary: summary, Highlights: highlights}
	if h.rates != nil {
		if rate, err := h.rates.USDPerEUR(c.Request.Context()); err == nil {
			total := math.Round(summary.TotalValue/rate.USDPerEUR*100) / 100
			response.TotalValueEUR = &total
			response.ExchangeRate = &rate
		}
	}
	c.JSON(http.StatusOK, response)
}
