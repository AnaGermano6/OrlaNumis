package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// A coin added to the collection is found on the wishlist under its name and
// year, or by its Numista type; another coin, or another user's, is not.
func TestWishlistMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	createTestUser(t, db, 2, "other")
	handler := NewCoinReleasesHandler(services.NewCoinReleasesService(nil), repository.NewCoinRepository(db), nil)
	group := router.Group("/api/matches-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/:id", handler.WishlistMatches)

	year, other := 2026, 2025
	create := func(coin models.Coin, numista string) uint {
		t.Helper()
		if coin.UserID == 0 {
			coin.UserID = 1
		}
		coin.Category, coin.Material = models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
		if numista != "" {
			if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: numista}).Error; err != nil {
				t.Fatal(err)
			}
		}
		return coin.ID
	}
	gulbenkian := "2 euro 2026 – 70 Anos da Fundação Calouste Gulbenkian"
	wanted := create(models.Coin{Name: gulbenkian, Kind: models.CoinKindCoin, Year: &year, IsWishlist: true}, "")
	byType := create(models.Coin{Name: "2 Euros (Gulbenkian)", Kind: models.CoinKindCoin, Year: &year, IsWishlist: true}, "77")
	create(models.Coin{Name: "2 euro 2025 – 70 Anos da Fundação Calouste Gulbenkian", Kind: models.CoinKindCoin, Year: &other, IsWishlist: true}, "")
	create(models.Coin{Name: gulbenkian, Kind: models.CoinKindCoin, Year: &year, IsWishlist: true, UserID: 2}, "")
	had := create(models.Coin{Name: "2 EURO 2026 –  70 anos da Fundacao Calouste Gulbenkian", Kind: models.CoinKindCoin, Year: &year}, "77")

	w := coinRequest(router, http.MethodGet, fmt.Sprintf("/api/matches-test/%d", had), nil)
	var found WishlistMatchesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil || w.Code != http.StatusOK {
		t.Fatalf("matches: %d %s", w.Code, w.Body.String())
	}
	if len(found.Coins) != 2 || found.Coins[0].ID != wanted || found.Coins[1].ID != byType || found.Coins[0].Name != gulbenkian {
		t.Fatalf("matches = %+v", found.Coins)
	}

	// A wishlist coin is on the wishlist already.
	w = coinRequest(router, http.MethodGet, fmt.Sprintf("/api/matches-test/%d", wanted), nil)
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil || w.Code != http.StatusOK || len(found.Coins) != 0 {
		t.Fatalf("a wishlist coin: %d %s", w.Code, w.Body.String())
	}
	if w := coinRequest(router, http.MethodGet, "/api/matches-test/9999", nil); w.Code != http.StatusNotFound {
		t.Fatalf("an unknown coin: %d", w.Code)
	}
}
