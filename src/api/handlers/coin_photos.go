package handlers

import (
	"net/http"
	"strings"

	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CoinPhotoHandler finds real photos of coins in public sources.
type CoinPhotoHandler struct {
	svc *services.CoinPhotoSearchService
}

// NewCoinPhotoHandler creates a new CoinPhotoHandler.
func NewCoinPhotoHandler(svc *services.CoinPhotoSearchService) *CoinPhotoHandler {
	return &CoinPhotoHandler{svc: svc}
}

// Search finds real photos of a coin.
//
//	@Summary		Search real coin photos
//	@Description	Returns real photos of a coin with their author and license: the Numista catalogue photos of numistaTypeId, then Wikimedia Commons files matching q. Commons searches are cached for a week. Import a chosen photo with POST /coins/{id}/images/import.
//	@Tags			Coins
//	@Produce		json
//	@Param			q				query		string	false	"Search text, e.g. title and year (required unless numistaTypeId is set)"
//	@Param			numistaTypeId	query		int		false	"Numista type whose catalogue photos to include"
//	@Param			lang			query		string	false	"Language of catalogue data"	default(en)
//	@Success		200				{object}	models.CoinPhotoSearchResult
//	@Failure		400				{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-photos/search [get]
func (h *CoinPhotoHandler) Search(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if len([]rune(query)) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q must be at most 200 characters"})
		return
	}
	typeID, ok := numistaQueryID(c, "numistaTypeId")
	if !ok {
		return
	}
	lang, ok := numistaQueryLang(c)
	if !ok {
		return
	}
	if query == "" && typeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q or numistaTypeId is required"})
		return
	}
	c.JSON(http.StatusOK, h.svc.Search(c.Request.Context(), query, typeID, lang))
}
