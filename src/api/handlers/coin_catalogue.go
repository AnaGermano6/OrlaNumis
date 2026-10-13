package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// CatalogueCountry is a country of the catalogue, in one currency (the euro,
// or Portugal's escudo): how many commemorative and collector coins it
// lists, and how many of each the user has.
type CatalogueCountry struct {
	Country string `json:"country" example:"PT"`
	// Currency is "PTE" for Portugal's coins in escudos; empty for the euro.
	Currency            string `json:"currency,omitempty" enums:",PTE"`
	Commemoratives      int    `json:"commemoratives"`
	Collector           int    `json:"collector"`
	OwnedCommemoratives int    `json:"ownedCommemoratives"`
	OwnedCollector      int    `json:"ownedCollector"`
}

// CoinCatalogueCountriesResponse lists the catalogue's countries.
type CoinCatalogueCountriesResponse struct {
	Countries []CatalogueCountry `json:"countries"`
}

// CoinCatalogueResponse is one country's catalogue in one currency, newest first. Complete
// is false while Numista has not answered every search of the catalogue
// (Portugal's escudos, the €5, €10 and €20 coins).
type CoinCatalogueResponse struct {
	Country  string                 `json:"country" example:"PT"`
	Coins    []services.CoinRelease `json:"coins"`
	Complete bool                   `json:"complete"`
}

// CoinCatalogueHandler serves the catalogue of commemorative and collector
// coins by country, marking those the user has.
type CoinCatalogueHandler struct {
	releases *services.CoinReleasesService
	coins    *repository.CoinRepository
	photos   *services.CoinPhotoSearchService
}

// NewCoinCatalogueHandler creates the handler; photos finds which coins of
// the catalogue the user's are, and may be nil.
func NewCoinCatalogueHandler(releases *services.CoinReleasesService, coins *repository.CoinRepository, photos *services.CoinPhotoSearchService) *CoinCatalogueHandler {
	return &CoinCatalogueHandler{releases: releases, coins: coins, photos: photos}
}

// ownedKeys finds the catalogue keys of the user's coins of a country ("" for
// every country of the catalogue).
func (h *CoinCatalogueHandler) ownedKeys(c *gin.Context, country string) (map[string]bool, bool) {
	coins, err := h.coins.CollectionCoinsWithReferences(c.GetUint("userId"), models.CoinKindCoin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the catalogue"})
		return nil, false
	}
	queries := []services.CoinPhotoQuery{}
	for i := range coins {
		code := services.CatalogueCountryCode(coins[i].Country)
		if code == "" || (country != "" && code != country) {
			continue
		}
		query := officialPhotoQuery(&coins[i])
		// Its Numista type as cached: no request to Numista for the catalogue.
		query.CachedTypesOnly = true
		queries = append(queries, query)
	}
	return h.releases.OwnedKeys(c.Request.Context(), queries, h.photos), true
}

// Countries godoc
//
//	@Summary		Countries of the coin catalogue
//	@Description	The countries of the catalogue, each in its currency: the euro (every €2 commemorative the ECB has published since 2004, the €5, €10 and €20 coins, and the collector coins the mints of Portugal, France and Austria sell) and Portugal's escudo (its commemorative and collector coins, 1911–2001). Each says how many commemorative and collector coins it lists, and how many of them the user has.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	CoinCatalogueCountriesResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-catalogue/countries [get]
func (h *CoinCatalogueHandler) Countries(c *gin.Context) {
	catalogue := h.releases.Catalogue(c.Request.Context())
	owned, ok := h.ownedKeys(c, "")
	if !ok {
		return
	}
	services.MarkOwnedReleases(catalogue, owned)
	byCountry := map[string]*CatalogueCountry{}
	for _, coin := range catalogue {
		key := coin.Country + "|" + coin.Currency
		country := byCountry[key]
		if country == nil {
			country = &CatalogueCountry{Country: coin.Country, Currency: coin.Currency}
			byCountry[key] = country
		}
		if coin.Kind == services.CoinReleaseKindCommemorative {
			country.Commemoratives++
			if coin.Owned {
				country.OwnedCommemoratives++
			}
		} else {
			country.Collector++
			if coin.Owned {
				country.OwnedCollector++
			}
		}
	}
	response := CoinCatalogueCountriesResponse{Countries: make([]CatalogueCountry, 0, len(byCountry))}
	for _, country := range byCountry {
		response.Countries = append(response.Countries, *country)
	}
	sort.Slice(response.Countries, func(i, j int) bool {
		a, b := response.Countries[i], response.Countries[j]
		return a.Country < b.Country || a.Country == b.Country && a.Currency < b.Currency
	})
	c.JSON(http.StatusOK, response)
}

// Country godoc
//
//	@Summary		A country's coin catalogue
//	@Description	A country's coins in one currency, newest first: in euros, its €2 commemoratives (every one the ECB has published since 2004), its €5, €10 and €20 coins and the collector coins its mint sells (Portugal, France, Austria); in escudos (currency PTE), Portugal's commemorative and collector coins of 1911–2001. Each says whether a coin of the collection is it.
//	@Tags			Coins
//	@Produce		json
//	@Param			country		query		string	true	"The country's ISO code, e.g. PT"
//	@Param			currency	query		string	false	"PTE for Portugal's coins in escudos; the euro coins when omitted"	Enums(PTE)
//	@Success		200		{object}	CoinCatalogueResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-catalogue [get]
func (h *CoinCatalogueHandler) Country(c *gin.Context) {
	country := strings.ToUpper(strings.TrimSpace(c.Query("country")))
	if services.CatalogueCountryName(country) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country must be the ISO code of a country of the catalogue"})
		return
	}
	currency := strings.ToUpper(strings.TrimSpace(c.Query("currency")))
	if currency != "" && currency != "PTE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currency must be PTE, or omitted for the euro"})
		return
	}
	coins := []services.CoinRelease{}
	for _, coin := range h.releases.Catalogue(c.Request.Context()) {
		if coin.Country == country && coin.Currency == currency {
			coins = append(coins, coin)
		}
	}
	owned, ok := h.ownedKeys(c, country)
	if !ok {
		return
	}
	services.MarkOwnedReleases(coins, owned)
	c.JSON(http.StatusOK, CoinCatalogueResponse{Country: country, Coins: coins, Complete: h.releases.CatalogueComplete()})
}
