package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

var (
	currencyCodePattern = regexp.MustCompile(`^[A-Z]{3}$`)
	languageCodePattern = regexp.MustCompile(`^[a-z]{2}$`)
)

// numistaQueryCurrency reads ?currency= (ISO 4217, default EUR).
func numistaQueryCurrency(c *gin.Context) (string, bool) {
	currency := strings.ToUpper(strings.TrimSpace(c.DefaultQuery("currency", "EUR")))
	if !currencyCodePattern.MatchString(currency) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be a three-letter ISO 4217 code"})
		return "", false
	}
	return currency, true
}

// numistaQueryLang reads ?lang= (two-letter code, default en).
func numistaQueryLang(c *gin.Context) (string, bool) {
	lang := strings.ToLower(strings.TrimSpace(c.DefaultQuery("lang", "en")))
	if !languageCodePattern.MatchString(lang) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lang must be a two-letter language code"})
		return "", false
	}
	return lang, true
}

// numistaQueryID reads an optional positive Numista ID query parameter (0 when absent).
func numistaQueryID(c *gin.Context, name string) (int, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return 0, true
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 || id > models.NumistaMaxID {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be a positive integer"})
		return 0, false
	}
	return id, true
}

// CoinMarketValueHandler serves Numista's estimated market value for a coin.
type CoinMarketValueHandler struct {
	coins *repository.CoinRepository
	svc   *services.NumistaMarketValueService
}

// NewCoinMarketValueHandler creates a new CoinMarketValueHandler.
func NewCoinMarketValueHandler(coins *repository.CoinRepository, svc *services.NumistaMarketValueService) *CoinMarketValueHandler {
	return &CoinMarketValueHandler{coins: coins, svc: svc}
}

// MarketValue returns Numista's estimated price per grade for a coin.
//
//	@Summary		Get a coin's Numista market value
//	@Description	Returns Numista's estimated price per grade for the issue of the coin's linked Numista type (from its "Numista" catalogue reference), plus the type's catalogue photos with credit. Responses are cached for days to stay inside Numista's free plan; when Numista cannot be reached the last stored values are returned with stale=true. Provider problems are reported in status, not as HTTP errors.
//	@Tags			Coins
//	@Produce		json
//	@Param			id			path		int		true	"Coin ID"
//	@Param			issueId		query		int		false	"Numista issue ID; defaults to the issue matching the coin's date, else the first"
//	@Param			currency	query		string	false	"ISO 4217 currency code"	default(EUR)
//	@Param			lang		query		string	false	"Language of catalogue data"	default(en)
//	@Success		200			{object}	models.CoinMarketValue
//	@Failure		400			{object}	ErrorResponse
//	@Failure		404			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/market-value [get]
func (h *CoinMarketValueHandler) MarketValue(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
		return
	}
	issueID, ok := numistaQueryID(c, "issueId")
	if !ok {
		return
	}
	currency, ok := numistaQueryCurrency(c)
	if !ok {
		return
	}
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}

	coin, err := h.coins.FindByID(uint(id), userID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Coin not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coin"})
		return
	}

	// A coin's page asks for its market value on its own: it waits when the
	// month's Numista quota runs low.
	c.JSON(http.StatusOK, h.svc.ForCoin(services.WithNumistaBackground(c.Request.Context()), coin, issueID, currency, lang))
}
