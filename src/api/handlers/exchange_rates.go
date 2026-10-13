package handlers

import (
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// ExchangeRatesHandler serves the ECB daily euro reference rates.
type ExchangeRatesHandler struct {
	rates *services.ExchangeRateService
}

// NewExchangeRatesHandler creates the handler.
func NewExchangeRatesHandler(rates *services.ExchangeRateService) *ExchangeRatesHandler {
	return &ExchangeRatesHandler{rates: rates}
}

// ExchangeRatesResponse is a day's ECB rates and the currencies a collection
// can be kept in.
type ExchangeRatesResponse struct {
	services.ExchangeRates
	// Currencies are the collection currencies (the euro first).
	Currencies []string `json:"currencies"`
}

// Rates returns the latest ECB euro reference rates.
//
//	@Summary		Get exchange rates
//	@Description	Returns the ECB daily euro reference rates (how much of each currency one euro buys; cached for a few hours, the last stored rates when the ECB cannot be reached) and the currencies a collection can be kept in, to convert amounts between them.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	ExchangeRatesResponse
//	@Failure		503	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/exchange-rates [get]
func (h *ExchangeRatesHandler) Rates(c *gin.Context) {
	rates, err := h.rates.Rates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Exchange rates are unavailable"})
		return
	}
	c.JSON(http.StatusOK, ExchangeRatesResponse{ExchangeRates: rates, Currencies: services.CollectionCurrencies})
}
