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
//	@Description	Returns real photos of a coin with their author and license: the ECB's image of a €2 commemorative coin (when year is given and the coin is one; match=exact when the ECB lists exactly this coin), the Numista catalogue photos of numistaTypeId, then Wikimedia Commons images: searched in English from the denomination, country and year (as Commons names coin photos, e.g. "200 lire San Marino 1995 coin"), then for q. Commons searches are cached for a week, ECB pages for longer. Import a chosen photo with POST /coins/{id}/images/import.
//	@Tags			Coins
//	@Produce		json
//	@Param			q				query		string	false	"Search text, e.g. title and year (required unless numistaTypeId, or year and country, are set)"
//	@Param			numistaTypeId	query		int		false	"Numista type whose catalogue photos to include"
//	@Param			lang			query		string	false	"Language of catalogue data"	default(en)
//	@Param			year			query		int		false	"Coin year: a €2 commemorative's ECB image is looked up, and Commons is searched with it"
//	@Param			country			query		string	false	"Coin country (English or Portuguese name; English finds more on Commons)"
//	@Param			denomination	query		string	false	"Coin denomination, e.g. 2 euro or 200 liras"
//	@Param			subject			query		string	false	"Coin subject, for a coin not linked to Numista"
//	@Param			kind			query		string	false	"What the photos are of"	Enums(coin, banknote)	default(coin)
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
	year, ok := optionalIntQuery(c, "year", -5000, 3000)
	if !ok {
		return
	}
	limited := func(name string) (string, bool) {
		value := strings.TrimSpace(c.Query(name))
		if len([]rune(value)) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be at most 200 characters"})
			return "", false
		}
		return value, true
	}
	country, ok := limited("country")
	if !ok {
		return
	}
	denomination, ok := limited("denomination")
	if !ok {
		return
	}
	subject, ok := limited("subject")
	if !ok {
		return
	}
	kind := strings.TrimSpace(c.DefaultQuery("kind", "coin"))
	if kind != "coin" && kind != "banknote" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be coin or banknote"})
		return
	}
	// The ECB lookup needs only the year and country.
	if query == "" && typeID == 0 && (year == 0 || country == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "q, numistaTypeId or year and country are required"})
		return
	}
	c.JSON(http.StatusOK, h.svc.SearchCoin(c.Request.Context(), services.CoinPhotoQuery{
		Query: query, NumistaTypeID: typeID, Lang: lang, Year: year, Country: country, Denomination: denomination, Subject: subject,
		Banknote: kind == "banknote",
	}))
}
