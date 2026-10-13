package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	owned := map[string]bool{}
	for _, coin := range portugal.Coins {
		owned[coin.Title] = coin.Owned
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
