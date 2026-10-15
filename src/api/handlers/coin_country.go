package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CoinCountryHandler lists coin countries and fills them from Numista.
type CoinCountryHandler struct {
	coins *repository.CoinRepository
	svc   *services.CoinCountryService
}

// NewCoinCountryHandler creates a new CoinCountryHandler.
func NewCoinCountryHandler(coins *repository.CoinRepository, svc *services.CoinCountryService) *CoinCountryHandler {
	return &CoinCountryHandler{coins: coins, svc: svc}
}

// Countries lists the collection's countries with coin counts.
//
//	@Summary		List collection countries
//	@Description	Returns each country of the user's collection coins (excluding wishlist and sold) with how many coins it has, alphabetically.
//	@Tags			Coins
//	@Produce		json
//	@Param			kind	query		string	false	"Only coins or only banknotes (every item when omitted)"	Enums(coin, banknote)
//	@Success		200		{array}		repository.CountryCount
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/countries [get]
func (h *CoinCountryHandler) Countries(c *gin.Context) {
	kind, ok := coinKindQuery(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'coin' or 'banknote'"})
		return
	}
	counts, err := h.coins.CountryCounts(c.GetUint("userId"), kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list countries"})
		return
	}
	if counts == nil {
		counts = []repository.CountryCount{}
	}
	c.JSON(http.StatusOK, counts)
}

// Denominations lists the collection's face values with coin counts.
//
//	@Summary		List collection face values
//	@Description	Returns each face value of the user's collection coins (excluding wishlist and sold), as written, with how many coins have it, the most common first.
//	@Tags			Coins
//	@Produce		json
//	@Param			kind	query		string	false	"Only coins or only banknotes (every item when omitted)"	Enums(coin, banknote)
//	@Success		200		{array}		repository.DenominationCount
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/denominations [get]
func (h *CoinCountryHandler) Denominations(c *gin.Context) {
	kind, ok := coinKindQuery(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be 'coin' or 'banknote'"})
		return
	}
	counts, err := h.coins.DenominationCounts(c.GetUint("userId"), kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list face values"})
		return
	}
	if counts == nil {
		counts = []repository.DenominationCount{}
	}
	c.JSON(http.StatusOK, counts)
}

// FillFromNumista sets missing countries from the coins' Numista types.
//
//	@Summary		Fill coin countries from Numista
//	@Description	For the user's coins that have no country but are linked to a Numista type, sets the country from the type's issuer. Handles up to 25 coins per call (issuer details are cached for 30 days); call again while remaining > 0. Stops early when Numista is not configured or the quota is spent (see status).
//	@Tags			Coins
//	@Produce		json
//	@Param			coinId	query		int		false	"Only this coin"
//	@Param			lang	query		string	false	"Language of country names"	default(en)
//	@Success		200		{object}	models.CoinCountryFillResult
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/country-from-numista [post]
func (h *CoinCountryHandler) FillFromNumista(c *gin.Context) {
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}
	var coinID uint64
	if raw := strings.TrimSpace(c.Query("coinId")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 32)
		if err != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "coinId must be a positive integer"})
			return
		}
		coinID = parsed
	}
	// The gallery fills countries on its own when grouped by country: it
	// waits when the month's Numista quota runs low.
	c.JSON(http.StatusOK, h.svc.FillFromNumista(services.WithNumistaBackground(c.Request.Context()), c.GetUint("userId"), uint(coinID), lang))
}
