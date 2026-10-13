package handlers

import (
	"errors"
	"net/http"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CollectionCurrencyHandler lets an admin switch the currency the collection's
// prices and values are kept in.
type CollectionCurrencyHandler struct {
	svc *services.CollectionCurrencyService
}

// NewCollectionCurrencyHandler creates the handler.
func NewCollectionCurrencyHandler(svc *services.CollectionCurrencyService) *CollectionCurrencyHandler {
	return &CollectionCurrencyHandler{svc: svc}
}

// CollectionCurrencyRequest selects the new collection currency.
type CollectionCurrencyRequest struct {
	// Currency is EUR or a currency of the ECB reference rates (USD, GBP, CHF…).
	Currency string `json:"currency" binding:"required" example:"EUR"`
}

// Change switches the collection currency and converts stored amounts.
//
//	@Summary		Change the collection currency
//	@Description	Switches the currency coin prices and values are kept in and every page shows (the euro or a currency of the ECB daily reference rates: USD, GBP, CHF, JPY…) and converts every stored price, value, value history and snapshot with the ECB reference rates, in one transaction. Switching to the current currency changes nothing. Admin only.
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			body	body		CollectionCurrencyRequest	true	"New currency"
//	@Success		200		{object}	services.CollectionCurrencyChange
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Failure		502		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/collection-currency [post]
func (h *CollectionCurrencyHandler) Change(c *gin.Context) {
	var req CollectionCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	change, err := h.svc.Change(c.Request.Context(), req.Currency)
	if errors.Is(err, services.ErrUnsupportedCollectionCurrency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, services.ErrExchangeRateUnavailable) {
		respondError(c, http.StatusBadGateway, "The ECB exchange rate is unavailable; the currency was not changed", err)
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not convert the collection values; the currency was not changed", err)
		return
	}
	c.JSON(http.StatusOK, change)
}
