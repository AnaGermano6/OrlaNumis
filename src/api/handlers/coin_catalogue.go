package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

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
	// CommemorativeYears are the country's €2 commemoratives year by year,
	// oldest first, with how many of each year the collection has (euro
	// countries only).
	CommemorativeYears []CatalogueYear `json:"commemorativeYears,omitempty"`
}

// CoinCatalogueCountriesResponse lists the catalogue's countries, and year by
// year (oldest first) the euro area's €2 commemoratives, with how many of
// each year the collection has.
type CoinCatalogueCountriesResponse struct {
	Countries          []CatalogueCountry `json:"countries"`
	CommemorativeYears []CatalogueYear    `json:"commemorativeYears"`
	// Circulation is each euro country's circulation coins year by year
	// (the eight of each year, 1 cent to 2 euro), by country code, with how
	// many of them the collection has.
	Circulation []CatalogueCirculation `json:"circulation"`
}

// CatalogueCirculation is a euro country's circulation coins, year by year
// from its first euro coins to this year.
type CatalogueCirculation struct {
	Country string                         `json:"country" example:"PT"`
	Years   []services.EuroCirculationYear `json:"years"`
}

// CatalogueYear is how many coins of one year a part of the catalogue has,
// and how many of them the collection has.
type CatalogueYear struct {
	Year  int `json:"year" example:"2025"`
	Total int `json:"total"`
	Owned int `json:"owned"`
}

// catalogueYears adds a coin to its year's count, by the year it was first issued.
func catalogueYears(years map[int]*CatalogueYear, coin services.CoinRelease) {
	if len(coin.From) < 4 {
		return
	}
	year, err := strconv.Atoi(coin.From[:4])
	if err != nil {
		return
	}
	entry := years[year]
	if entry == nil {
		entry = &CatalogueYear{Year: year}
		years[year] = entry
	}
	entry.Total++
	if coin.Owned {
		entry.Owned++
	}
}

// sortedYears lists the years' counts, oldest first.
func sortedYears(years map[int]*CatalogueYear) []CatalogueYear {
	list := make([]CatalogueYear, 0, len(years))
	for _, entry := range years {
		list = append(list, *entry)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Year < list[j].Year })
	return list
}

// CoinCatalogueResponse is one country's catalogue in one currency, newest first. Complete
// is false while Numista has not answered every search of the catalogue
// (Portugal's escudos, the €5, €10 and €20 coins).
type CoinCatalogueResponse struct {
	Country string                 `json:"country" example:"PT"`
	Coins   []services.CoinRelease `json:"coins"`
	// Hidden are the coins the user hid, and those the user joined to
	// another (sameAs), which is listed in their place.
	Hidden   []services.CoinRelease `json:"hidden"`
	Complete bool                   `json:"complete"`
}

// CatalogueOverrideRequest hides a coin of the catalogue, or joins it to
// another of the same country (sameAs), which is then listed in its place.
type CatalogueOverrideRequest struct {
	ID     string `json:"id" binding:"required,max=600"`
	SameAs string `json:"sameAs,omitempty" binding:"max=600"`
}

// JointIssuesResponse lists the euro area's joint issues, newest first.
type JointIssuesResponse struct {
	Issues []services.JointIssue `json:"issues"`
}

// JointIssues godoc
//
//	@Summary		Joint issues of the euro area
//	@Description	The €2 coins the euro area's countries issued together (the Treaty of Rome in 2007, the EU flag in 2015…), newest first, each with every country's version.
//	@Tags			Coins
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	JointIssuesResponse
//	@Failure		401	{object}	ErrorResponse
//	@Router			/coin-catalogue/joint-issues [get]
func (h *CoinCatalogueHandler) JointIssues(c *gin.Context) {
	c.JSON(http.StatusOK, JointIssuesResponse{Issues: h.releases.JointIssues(c.Request.Context())})
}

// CoinCatalogueHandler serves the catalogue of commemorative and collector
// coins by country, marking those the user has.
type CoinCatalogueHandler struct {
	releases  *services.CoinReleasesService
	coins     *repository.CoinRepository
	photos    *services.CoinPhotoSearchService
	overrides *repository.CatalogueOverrideRepository
}

// NewCoinCatalogueHandler creates the handler; photos finds which coins of
// the catalogue the user's are, and may be nil.
func NewCoinCatalogueHandler(releases *services.CoinReleasesService, coins *repository.CoinRepository, photos *services.CoinPhotoSearchService) *CoinCatalogueHandler {
	return &CoinCatalogueHandler{releases: releases, coins: coins, photos: photos}
}

// WithOverrides lets each user hide coins of the catalogue, or join one to
// another.
func (h *CoinCatalogueHandler) WithOverrides(overrides *repository.CatalogueOverrideRepository) *CoinCatalogueHandler {
	h.overrides = overrides
	return h
}

// userCatalogue applies the user's choices to coins of the catalogue.
func (h *CoinCatalogueHandler) userCatalogue(c *gin.Context, coins []services.CoinRelease) (shown, hidden []services.CoinRelease, ok bool) {
	if h.overrides == nil {
		return coins, []services.CoinRelease{}, true
	}
	overrides, err := h.overrides.List(c.GetUint("userId"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the catalogue"})
		return nil, nil, false
	}
	shown, hidden = services.ApplyCatalogueOverrides(coins, overrides)
	return shown, hidden, true
}

// ownedKeys finds the catalogue keys of the user's coins of a country ("" for
// every country of the catalogue).
func (h *CoinCatalogueHandler) ownedKeys(c *gin.Context, country string) (map[string]uint, bool) {
	coins, err := h.coins.CollectionCoinsWithReferences(c.GetUint("userId"), models.CoinKindCoin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the catalogue"})
		return nil, false
	}
	return h.ownedKeysOf(c, coins, country), true
}

// ownedKeysOf finds the catalogue keys of some of the user's coins, of a
// country ("" for every country of the catalogue).
func (h *CoinCatalogueHandler) ownedKeysOf(c *gin.Context, coins []models.Coin, country string) map[string]uint {
	queries := []services.CoinPhotoQuery{}
	for i := range coins {
		code := services.CatalogueCountryCode(coins[i].Country)
		if code == "" || (country != "" && code != country) {
			continue
		}
		query := services.CoinPhotoQueryFor(&coins[i])
		// Its Numista type as cached: no request to Numista for the catalogue.
		query.CachedTypesOnly = true
		queries = append(queries, query)
	}
	return h.releases.OwnedKeys(c.Request.Context(), queries, h.photos)
}

// Countries godoc
//
//	@Summary		Countries of the coin catalogue
//	@Description	The countries of the catalogue, each in its currency: the euro (every €2 commemorative the ECB has published since 2004, the €5, €10 and €20 coins, and the collector coins the mints of Portugal, France and Austria sell) and Portugal's escudo (its commemorative and collector coins, 1911–2001). Each says how many commemorative and collector coins it lists, and how many of them the user has; a euro country, its €2 commemoratives year by year too. Then, year by year, the euro area's €2 commemoratives, with how many of each year the collection has; and each euro country's circulation coins, the eight of each year since its first euro coins, with how many of them the collection has.
//	@Tags			Coins
//	@Produce		json
//	@Success		200	{object}	CoinCatalogueCountriesResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-catalogue/countries [get]
func (h *CoinCatalogueHandler) Countries(c *gin.Context) {
	catalogue, _, ok := h.userCatalogue(c, h.releases.Catalogue(c.Request.Context()))
	if !ok {
		return
	}
	coins, err := h.coins.CollectionCoinsWithReferences(c.GetUint("userId"), models.CoinKindCoin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the catalogue"})
		return
	}
	services.MarkOwnedReleases(catalogue, h.ownedKeysOf(c, coins, ""))
	byCountry := map[string]*CatalogueCountry{}
	countryYears := map[string]map[int]*CatalogueYear{}
	commemorativeYears := map[int]*CatalogueYear{}
	for _, coin := range catalogue {
		key := coin.Country + "|" + coin.Currency
		country := byCountry[key]
		if country == nil {
			country = &CatalogueCountry{Country: coin.Country, Currency: coin.Currency}
			byCountry[key] = country
			countryYears[key] = map[int]*CatalogueYear{}
		}
		if coin.Currency == "" && coin.Kind == services.CoinReleaseKindCommemorative {
			catalogueYears(commemorativeYears, coin)
			catalogueYears(countryYears[key], coin)
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
	response := CoinCatalogueCountriesResponse{
		Countries:          make([]CatalogueCountry, 0, len(byCountry)),
		CommemorativeYears: sortedYears(commemorativeYears),
		Circulation:        []CatalogueCirculation{},
	}
	for country, years := range services.EuroCirculationYears(coins, time.Now().Year()) {
		response.Circulation = append(response.Circulation, CatalogueCirculation{Country: country, Years: years})
	}
	sort.Slice(response.Circulation, func(i, j int) bool { return response.Circulation[i].Country < response.Circulation[j].Country })
	for key, country := range byCountry {
		if years := countryYears[key]; len(years) > 0 {
			country.CommemorativeYears = sortedYears(years)
		}
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
//	@Description	A country's coins in one currency, newest first: in euros, its €2 commemoratives (every one the ECB has published since 2004), its €5, €10 and €20 coins and the collector coins its mint sells (Portugal, France, Austria); in escudos (currency PTE), Portugal's commemorative and collector coins of 1911–2001. Each says whether a coin of the collection is it. The coins the user hid, or joined to another, are listed apart.
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
	coins, hidden, ok := h.userCatalogue(c, coins)
	if !ok {
		return
	}
	owned, ok := h.ownedKeys(c, country)
	if !ok {
		return
	}
	services.MarkOwnedReleases(coins, owned)
	services.MarkOwnedReleases(hidden, owned)
	wishlist, err := h.coins.WishlistCoinsWithReferences(c.GetUint("userId"), models.CoinKindCoin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read the catalogue"})
		return
	}
	ofCountry := []models.Coin{}
	for _, coin := range wishlist {
		if services.CatalogueCountryCode(coin.Country) == country {
			ofCountry = append(ofCountry, coin)
		}
	}
	services.MarkWishlistReleases(coins, h.releases.OwnedKeys(c.Request.Context(), wishlistQueries(ofCountry, nil), h.photos))
	c.JSON(http.StatusOK, CoinCatalogueResponse{Country: country, Coins: coins, Hidden: hidden, Complete: h.releases.CatalogueComplete()})
}

// SetOverride godoc
//
//	@Summary		Hide a coin of the catalogue, or join it to another
//	@Description	Hides a coin of the catalogue from the user's catalogue and checklist sets; or, with sameAs, says it is the same coin as another of the same country and currency, which is then listed in its place (and is the user's when the user has either). It replaces what the user had chosen for the coin.
//	@Tags			Coins
//	@Accept			json
//	@Param			request	body	CatalogueOverrideRequest	true	"The coin's id, and the other's"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-catalogue/overrides [put]
func (h *CoinCatalogueHandler) SetOverride(c *gin.Context) {
	var request CatalogueOverrideRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	catalogue := h.releases.Catalogue(c.Request.Context())
	coin, ok := services.FindCatalogueCoin(catalogue, request.ID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be a coin of the catalogue"})
		return
	}
	sameAs := ""
	if request.SameAs != "" {
		other, ok := services.FindCatalogueCoin(catalogue, request.SameAs)
		if !ok || other.Country != coin.Country || other.Currency != coin.Currency || other.ID == coin.ID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "sameAs must be another coin of the catalogue, of the same country and currency"})
			return
		}
		sameAs = other.ID
	}
	userID := c.GetUint("userId")
	// One choice per coin, whichever of its ids it was made by.
	if err := h.overrides.Delete(userID, coin.Keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the choice"})
		return
	}
	if err := h.overrides.Set(userID, coin.ID, sameAs, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the choice"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteOverride godoc
//
//	@Summary		Show a coin of the catalogue again
//	@Description	Forgets the user's choice for a coin of the catalogue: hidden, it is listed again; joined to another, it is listed apart again.
//	@Tags			Coins
//	@Param			id	query	string	true	"The coin's id"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/coin-catalogue/overrides [delete]
func (h *CoinCatalogueHandler) DeleteOverride(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}
	keys := []string{id}
	if coin, ok := services.FindCatalogueCoin(h.releases.Catalogue(c.Request.Context()), id); ok {
		keys = coin.Keys
	}
	if err := h.overrides.Delete(c.GetUint("userId"), keys); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the choice"})
		return
	}
	c.Status(http.StatusNoContent)
}
