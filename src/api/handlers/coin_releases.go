package handlers

import (
	"net/http"
	"strconv"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CoinReleasesResponse lists new and upcoming coins, by source.
type CoinReleasesResponse struct {
	Groups []services.CoinReleaseGroup `json:"groups"`
}

// CoinReleasesHandler serves new and upcoming coins, marking those the user has.
type CoinReleasesHandler struct {
	releases *services.CoinReleasesService
	coins    *repository.CoinRepository
	photos   *services.CoinPhotoSearchService
}

// NewCoinReleasesHandler creates the handler; photos finds which releases
// the user's coins are, and may be nil.
func NewCoinReleasesHandler(releases *services.CoinReleasesService, coins *repository.CoinRepository, photos *services.CoinPhotoSearchService) *CoinReleasesHandler {
	return &CoinReleasesHandler{releases: releases, coins: coins, photos: photos}
}

// Releases godoc
//
//	@Summary		List new and upcoming coins
//	@Description	The ECB's €2 commemorative coins of the latest year it lists, and the collector coins the mints of Portugal (INCM), France (Monnaie de Paris) and Austria (Münze Österreich) sell, of the last twelve months and to come: upcoming first, then newest first. Each names its country and says whether a coin of the collection is it.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	CoinReleasesResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-releases [get]
func (h *CoinReleasesHandler) Releases(c *gin.Context) {
	groups := h.releases.Releases(c.Request.Context())
	years := map[int]bool{}
	for _, group := range groups {
		for _, release := range group.Releases {
			if year, err := strconv.Atoi(release.From[:min(4, len(release.From))]); err == nil {
				years[year] = true
			}
		}
	}
	list := make([]int, 0, len(years))
	for year := range years {
		list = append(list, year)
	}
	coins, err := h.coins.CollectionCoinsOfYears(c.GetUint("userId"), models.CoinKindCoin, list)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list the coin releases"})
		return
	}
	queries := make([]services.CoinPhotoQuery, 0, len(coins))
	for i := range coins {
		query := services.CoinPhotoQueryFor(&coins[i])
		// Its Numista type as cached: no request to Numista for this list.
		query.CachedTypesOnly = true
		queries = append(queries, query)
	}
	h.releases.MarkOwned(c.Request.Context(), groups, queries, h.photos)
	wishlist, err := h.coins.WishlistCoinsWithReferences(c.GetUint("userId"), models.CoinKindCoin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list the coin releases"})
		return
	}
	h.releases.MarkWishlist(c.Request.Context(), groups, wishlistQueries(wishlist, years), h.photos)
	c.JSON(http.StatusOK, CoinReleasesResponse{Groups: groups})
}

// wishlistQueries describes the wishlist's coins (of these years, when
// given) to find them among the releases, from the cache only.
func wishlistQueries(coins []models.Coin, years map[int]bool) []services.CoinPhotoQuery {
	queries := []services.CoinPhotoQuery{}
	for i := range coins {
		if years != nil && (coins[i].Year == nil || !years[*coins[i].Year]) {
			continue
		}
		query := services.CoinPhotoQueryFor(&coins[i])
		query.CachedTypesOnly = true
		queries = append(queries, query)
	}
	return queries
}
