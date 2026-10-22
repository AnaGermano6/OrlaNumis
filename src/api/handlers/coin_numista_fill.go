package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CoinNumistaFillHandler completes coin records from cached Numista types.
type CoinNumistaFillHandler struct {
	svc *services.CoinNumistaFillService
}

// NewCoinNumistaFillHandler creates a new CoinNumistaFillHandler.
func NewCoinNumistaFillHandler(svc *services.CoinNumistaFillService) *CoinNumistaFillHandler {
	return &CoinNumistaFillHandler{svc: svc}
}

// Preview tells what a fill from Numista would change.
//
//	@Summary		Preview a fill of coin records from Numista
//	@Description	For the user's coins and banknotes (collection and wishlist) linked to a Numista type whose details the app has cached, lists what a fill would set: the monetary system from the type's currency when none is set; and, for coins, the material from Numista's composition when the coin's is Other or empty, the weight and the diameter when missing, and a coin card as presentation when none is set and Numista says the type is issued in one. Also lists, as corrections, the coins whose material is set but differs from the composition. Changes nothing and makes no Numista request.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	models.CoinNumistaFillResult
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/numista-fill [get]
func (h *CoinNumistaFillHandler) Preview(c *gin.Context) {
	result, err := h.svc.Preview(c.GetUint("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the coins"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// Apply fills coin records from Numista.
//
//	@Summary		Fill coin records from Numista
//	@Description	Saves what the preview lists (see GET /coins/numista-fill): only empty fields are filled, all coins or none, from the cached Numista types; no Numista request is made. With corrections=true, the listed material corrections are saved too. Returns what was changed.
//	@Tags			Coins
//	@Produce		json
//	@Param			corrections	query		bool	false	"Also replace the materials that differ from Numista's composition"
//	@Success		200			{object}	models.CoinNumistaFillResult
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/numista-fill [post]
func (h *CoinNumistaFillHandler) Apply(c *gin.Context) {
	withCorrections := false
	if raw := strings.TrimSpace(c.Query("corrections")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "corrections must be true or false"})
			return
		}
		withCorrections = parsed
	}
	result, err := h.svc.Apply(c.GetUint("userId"), withCorrections)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fill the coins"})
		return
	}
	c.JSON(http.StatusOK, result)
}
