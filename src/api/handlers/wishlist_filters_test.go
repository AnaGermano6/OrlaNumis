package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

func TestWishlistFiltersAndYearFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	createTestUser(t, db, 2, "other")
	year := func(v int) *int { return &v }
	for _, coin := range []models.Coin{
		{UserID: 1, Name: "7½ euro 2024", Country: "Portugal", Year: year(2024), IsWishlist: true},
		{UserID: 1, Name: "2 euro 2024", Country: "France", Year: year(2024), IsWishlist: true},
		{UserID: 1, Name: "2 euro 2012", Country: "Portugal", Year: year(2012), IsWishlist: true},
		{UserID: 1, Name: "Sem país nem ano", IsWishlist: true},
		{UserID: 1, Name: "Na coleção", Country: "Spain", Year: year(1999)},
		{UserID: 2, Name: "De outro", Country: "Italy", Year: year(2001), IsWishlist: true},
	} {
		coin.Kind, coin.Category, coin.Material = models.CoinKindCoin, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	coinRepo := repository.NewCoinRepository(db)
	group := router.Group("/api/wishlist-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/filters", NewCoinCountryHandler(coinRepo, nil).WishlistFilters)
	group.GET("/coins", NewCoinHandler(coinRepo, services.NewCoinService(coinRepo), services.NewLogger(10)).List)

	// The wishlist's countries and years, each with its items; not the collection's, nor another user's.
	w := coinRequest(router, http.MethodGet, "/api/wishlist-test/filters", nil)
	var filters repository.WishlistFilters
	if err := json.Unmarshal(w.Body.Bytes(), &filters); err != nil || w.Code != http.StatusOK {
		t.Fatalf("filters: %d %s", w.Code, w.Body.String())
	}
	want := repository.WishlistFilters{
		Total:     4,
		Countries: []repository.CountryCount{{Country: "France", Count: 1}, {Country: "Portugal", Count: 2}},
		Years:     []repository.YearCount{{Year: 2024, Count: 2}, {Year: 2012, Count: 1}},
	}
	if !reflect.DeepEqual(filters, want) {
		t.Fatalf("filters = %+v, want %+v", filters, want)
	}

	// The list by year, alone or with a country.
	names := func(query string) []string {
		w := coinRequest(router, http.MethodGet, "/api/wishlist-test/coins?wishlist=true&sort=name&order=asc"+query, nil)
		var got struct {
			Coins []models.Coin `json:"coins"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("list %q: %d %s", query, w.Code, w.Body.String())
		}
		list := []string{}
		for _, coin := range got.Coins {
			list = append(list, coin.Name)
		}
		return list
	}
	if got := names("&year=2024"); !reflect.DeepEqual(got, []string{"2 euro 2024", "7½ euro 2024"}) {
		t.Fatalf("year 2024 = %v", got)
	}
	if got := names("&year=2024&country=Portugal"); !reflect.DeepEqual(got, []string{"7½ euro 2024"}) {
		t.Fatalf("2024 in Portugal = %v", got)
	}
	if w := coinRequest(router, http.MethodGet, "/api/wishlist-test/coins?year=dois", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("a year that is not a number: %d", w.Code)
	}
}
