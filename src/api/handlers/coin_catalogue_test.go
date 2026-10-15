package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// catalogueECB serves the ECB's 2022 page of €2 commemoratives, and no other year.
type catalogueECB struct{ page []byte }

func (e catalogueECB) CommemorativePage(_ context.Context, year int) ([]byte, error) {
	if year == 2022 {
		return e.page, nil
	}
	return nil, errors.New("ecb page: status 404")
}

// The catalogue lists a country's €2 commemoratives, its coin of a joint
// issue included, marking the one the user has.
func TestCoinCatalogue_CountryAndCountries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile("../services/testdata/ecb_comm_2022.html")
	if err != nil {
		t.Fatal(err)
	}
	year := 2022
	for _, coin := range []models.Coin{
		{Name: "2 euro 2022 – Travessia aérea do Atlântico Sul por Gago Coutinho e Sacadura Cabral", Denomination: "2 euro", Country: "Portugal", Year: &year},
		{Name: "1 euro 2022 – Portugal", Denomination: "1 euro", Country: "Portugal", Year: &year},
		// Wanted, not had.
		{Name: "2 euro 2022 – 35 years of the Erasmus programme", Denomination: "2 euro", Country: "Portugal", Year: &year, IsWishlist: true},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	ecb := services.NewEcbCoinPhotoService(catalogueECB{page: page}, cache)
	handler := NewCoinCatalogueHandler(services.NewCoinReleasesService(ecb), repository.NewCoinRepository(db), services.NewCoinPhotoSearchService(nil, nil, nil).WithECB(ecb))
	group := router.Group("/api/catalogue-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/countries", handler.Countries)
	group.GET("", handler.Country)

	w := coinRequest(router, http.MethodGet, "/api/catalogue-test?country=pt", nil)
	var portugal CoinCatalogueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &portugal); err != nil || w.Code != http.StatusOK || portugal.Country != "PT" {
		t.Fatalf("Portugal: %d %s", w.Code, w.Body.String())
	}
	owned, wishlist := map[string]bool{}, map[string]bool{}
	for _, coin := range portugal.Coins {
		owned[coin.Title], wishlist[coin.Title] = coin.Owned, coin.Wishlist
	}
	if !wishlist["35 years of the Erasmus programme"] || len(wishlist) != 2 || wishlist["The 100th anniversary of the crossing the South Atlantic Ocean by air, achieved in 1922 by Gago Coutinho and Sacadura Cabral"] {
		t.Fatalf("wishlist = %v", wishlist)
	}
	if len(portugal.Coins) != 2 || !owned["The 100th anniversary of the crossing the South Atlantic Ocean by air, achieved in 1922 by Gago Coutinho and Sacadura Cabral"] {
		t.Fatalf("Portugal's coins = %+v", portugal.Coins)
	}
	for title, has := range owned {
		if strings.Contains(title, "Erasmus") && has {
			t.Fatalf("the Erasmus coin is not the user's")
		}
	}

	w = coinRequest(router, http.MethodGet, "/api/catalogue-test/countries", nil)
	var countries CoinCatalogueCountriesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &countries); err != nil || w.Code != http.StatusOK {
		t.Fatalf("countries: %d %s", w.Code, w.Body.String())
	}
	var pt *CatalogueCountry
	for i := range countries.Countries {
		if countries.Countries[i].Country == "PT" {
			pt = &countries.Countries[i]
		}
	}
	if pt == nil || pt.Commemoratives != 2 || pt.OwnedCommemoratives != 1 || pt.Collector != 0 || len(countries.Countries) < 10 {
		t.Fatalf("countries = %+v", countries.Countries)
	}

	if w := coinRequest(router, http.MethodGet, "/api/catalogue-test?country=GB", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("a country outside the catalogue: %d", w.Code)
	}
	if w := coinRequest(router, http.MethodGet, "/api/catalogue-test?country=PT&currency=USD", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("a currency outside the catalogue: %d", w.Code)
	}
	// Portugal in escudos is apart from Portugal in euros; this catalogue has none.
	w = coinRequest(router, http.MethodGet, "/api/catalogue-test?country=PT&currency=pte", nil)
	var escudos CoinCatalogueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &escudos); err != nil || w.Code != http.StatusOK || len(escudos.Coins) != 0 {
		t.Fatalf("Portugal in escudos: %d %s", w.Code, w.Body.String())
	}
}

// A user hides a coin of the catalogue, or joins it to another, which is
// then the user's when the user has either; shown again, it is as it was.
func TestCoinCatalogue_HideAndJoin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}, &models.CatalogueOverride{}); err != nil {
		t.Fatal(err)
	}
	page, err := os.ReadFile("../services/testdata/ecb_comm_2022.html")
	if err != nil {
		t.Fatal(err)
	}
	year := 2022
	coin := models.Coin{
		Name: "2 euro 2022 – Travessia aérea do Atlântico Sul por Gago Coutinho e Sacadura Cabral", Denomination: "2 euro", Country: "Portugal", Year: &year,
		UserID: 1, Category: models.CategoryModern, Material: models.MaterialOther,
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatal(err)
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	ecb := services.NewEcbCoinPhotoService(catalogueECB{page: page}, cache)
	handler := NewCoinCatalogueHandler(services.NewCoinReleasesService(ecb), repository.NewCoinRepository(db), services.NewCoinPhotoSearchService(nil, nil, nil).WithECB(ecb)).
		WithOverrides(repository.NewCatalogueOverrideRepository(db))
	group := router.Group("/api/catalogue-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/countries", handler.Countries)
	group.GET("", handler.Country)
	group.PUT("/overrides", handler.SetOverride)
	group.DELETE("/overrides", handler.DeleteOverride)
	portugal := func() CoinCatalogueResponse {
		t.Helper()
		w := coinRequest(router, http.MethodGet, "/api/catalogue-test?country=PT", nil)
		var response CoinCatalogueResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || w.Code != http.StatusOK {
			t.Fatalf("Portugal: %d %s", w.Code, w.Body.String())
		}
		return response
	}
	ptCommemoratives := func() (int, int) {
		t.Helper()
		w := coinRequest(router, http.MethodGet, "/api/catalogue-test/countries", nil)
		var countries CoinCatalogueCountriesResponse
		if err := json.Unmarshal(w.Body.Bytes(), &countries); err != nil || w.Code != http.StatusOK {
			t.Fatalf("countries: %d %s", w.Code, w.Body.String())
		}
		for _, country := range countries.Countries {
			if country.Country == "PT" {
				return country.Commemoratives, country.OwnedCommemoratives
			}
		}
		return 0, 0
	}

	before := portugal()
	if len(before.Coins) != 2 || len(before.Hidden) != 0 {
		t.Fatalf("Portugal = %+v", before)
	}
	var gago, erasmus services.CoinRelease
	for _, coin := range before.Coins {
		if coin.ID == "" {
			t.Fatalf("a coin without an id: %+v", coin)
		}
		if strings.Contains(coin.Title, "Erasmus") {
			erasmus = coin
		} else {
			gago = coin
		}
	}
	if !gago.Owned || erasmus.Owned {
		t.Fatalf("Gago Coutinho = %+v, Erasmus = %+v", gago, erasmus)
	}

	// Hidden.
	if w := coinRequest(router, http.MethodPut, "/api/catalogue-test/overrides", CatalogueOverrideRequest{ID: erasmus.ID}); w.Code != http.StatusNoContent {
		t.Fatalf("hide: %d %s", w.Code, w.Body.String())
	}
	after := portugal()
	if len(after.Coins) != 1 || after.Coins[0].ID != gago.ID || len(after.Hidden) != 1 || after.Hidden[0].ID != erasmus.ID || after.Hidden[0].SameAs != "" {
		t.Fatalf("Portugal with Erasmus hidden = %+v", after)
	}
	if total, owned := ptCommemoratives(); total != 1 || owned != 1 {
		t.Fatalf("PT counts %d, %d owned", total, owned)
	}

	// Gago Coutinho joined to Erasmus, in place of Erasmus hidden: Erasmus is the user's.
	if w := coinRequest(router, http.MethodPut, "/api/catalogue-test/overrides", CatalogueOverrideRequest{ID: gago.ID, SameAs: erasmus.ID}); w.Code != http.StatusNoContent {
		t.Fatalf("join: %d %s", w.Code, w.Body.String())
	}
	if w := coinRequest(router, http.MethodDelete, "/api/catalogue-test/overrides?id="+url.QueryEscape(erasmus.ID), nil); w.Code != http.StatusNoContent {
		t.Fatalf("show Erasmus: %d %s", w.Code, w.Body.String())
	}
	after = portugal()
	if len(after.Coins) != 1 || after.Coins[0].ID != erasmus.ID || !after.Coins[0].Owned || len(after.Hidden) != 1 || after.Hidden[0].SameAs != erasmus.ID {
		t.Fatalf("Portugal with Gago Coutinho joined = %+v", after)
	}
	if total, owned := ptCommemoratives(); total != 1 || owned != 1 {
		t.Fatalf("PT counts %d, %d owned", total, owned)
	}

	// Not a coin of the catalogue, itself, another country's.
	w := coinRequest(router, http.MethodGet, "/api/catalogue-test?country=FR", nil)
	var france CoinCatalogueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &france); err != nil || len(france.Coins) == 0 {
		t.Fatalf("France: %d %s", w.Code, w.Body.String())
	}
	for _, request := range []CatalogueOverrideRequest{
		{ID: "https://example.test/not-a-coin.jpg"},
		{ID: gago.ID, SameAs: gago.ID},
		{ID: gago.ID, SameAs: france.Coins[0].ID},
		{},
	} {
		if w := coinRequest(router, http.MethodPut, "/api/catalogue-test/overrides", request); w.Code != http.StatusBadRequest {
			t.Fatalf("%+v: %d", request, w.Code)
		}
	}

	if w := coinRequest(router, http.MethodDelete, "/api/catalogue-test/overrides?id="+url.QueryEscape(gago.ID), nil); w.Code != http.StatusNoContent {
		t.Fatalf("show Gago Coutinho: %d %s", w.Code, w.Body.String())
	}
	if after := portugal(); len(after.Coins) != 2 || len(after.Hidden) != 0 {
		t.Fatalf("Portugal shown again = %+v", after)
	}
	if w := coinRequest(router, http.MethodDelete, "/api/catalogue-test/overrides", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("no id: %d", w.Code)
	}
}
