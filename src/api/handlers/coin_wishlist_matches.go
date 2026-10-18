package handlers

import (
	"net/http"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/gin-gonic/gin"
)

// WishlistMatch is a wishlist coin that is a collection coin.
type WishlistMatch struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// WishlistMatchesResponse lists the wishlist coins a collection coin is.
type WishlistMatchesResponse struct {
	Coins []WishlistMatch `json:"coins"`
}

// WishlistMatches godoc
//
//	@Summary		Find a collection coin on the wishlist
//	@Description	The wishlist coins that are the collection coin, so that one just added to the collection can leave the wishlist: of the same Numista type or official photo (the ECB's image of a €2 commemorative, a mint's page of a collector coin), or of the same kind and year under the same name. Numista is not asked. A wishlist coin has none.
//	@Tags			Coins
//	@Produce		json
//	@Param			id	path		int	true	"Coin ID"
//	@Success		200	{object}	WishlistMatchesResponse
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		404	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coins/{id}/wishlist-matches [get]
func (h *CoinReleasesHandler) WishlistMatches(c *gin.Context) {
	userID := c.GetUint("userId")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coin ID"})
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
	response := WishlistMatchesResponse{Coins: []WishlistMatch{}}
	if coin.IsWishlist {
		c.JSON(http.StatusOK, response)
		return
	}
	wishlist, err := h.coins.WishlistCoinsWithReferences(userID, coin.Kind)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list the wishlist"})
		return
	}
	for _, match := range h.releases.WishlistMatches(c.Request.Context(), coin, wishlist, h.photos) {
		response.Coins = append(response.Coins, WishlistMatch{ID: match.ID, Name: match.Name})
	}
	c.JSON(http.StatusOK, response)
}
