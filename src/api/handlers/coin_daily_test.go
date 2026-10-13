package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

// Coins held more than once are listed apart, most copies first, and the
// summary counts them and the copies left to trade, for coins and banknotes.
func TestRepeatedCoins_ListAndSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	for _, coin := range []models.Coin{
		{Name: "1 escudo 1962", Quantity: 3},
		{Name: "2 euro 2024", Quantity: 1},
		{Name: "50 centavos 1965", Quantity: 2},
		{Name: "1000 escudos", Kind: models.CoinKindBanknote, Quantity: 2},
		{Name: "Wanted", Quantity: 4, IsWishlist: true},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}

	w := coinRequest(router, http.MethodGet, "/api/coins?kind=coin&wishlist=false&repeated=true&sort=quantity&order=desc", nil)
	var list struct {
		Coins []models.Coin `json:"coins"`
		Total int64         `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || w.Code != http.StatusOK || list.Total != 2 {
		t.Fatalf("repeated coins: %d %s", w.Code, w.Body.String())
	}
	if list.Coins[0].Name != "1 escudo 1962" || list.Coins[1].Name != "50 centavos 1965" {
		t.Fatalf("repeated coins = %s, %s", list.Coins[0].Name, list.Coins[1].Name)
	}

	repo := repository.NewCoinRepository(db)
	summary := NewCollectionSummaryHandler(repo).WithSettings(services.NewSettingsService(repository.NewSettingsRepository(db)))
	group := router.Group("/api/repeated-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/summary", summary.Summary)
	for query, want := range map[string][2]int64{"": {3, 4}, "?kind=coin": {2, 3}, "?kind=banknote": {1, 1}} {
		var got struct {
			RepeatedCoins int64 `json:"repeatedCoins"`
			SpareCoins    int64 `json:"spareCoins"`
		}
		w := coinRequest(router, http.MethodGet, "/api/repeated-test/summary"+query, nil)
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.RepeatedCoins != want[0] || got.SpareCoins != want[1] {
			t.Fatalf("summary%s = %+v, want %v (%s)", query, got, want, w.Body.String())
		}
	}
}

// The coin of the day is one with a photo, the same all day; a banknote of
// the day comes from the banknotes, with or without a photo.
func TestDailyCoin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	createTestUser(t, db, 2, "other")
	coins := []models.Coin{
		{Name: "Without photo"},
		{Name: "Photographed A"},
		{Name: "Photographed B"},
		{Name: "Sold with photo", IsSold: true},
		{Name: "1000 escudos", Kind: models.CoinKindBanknote},
	}
	for i := range coins {
		coins[i].UserID, coins[i].Category, coins[i].Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, i := range []int{1, 2, 3} {
		if err := db.Create(&models.CoinImage{CoinID: coins[i].ID, FilePath: "uploads/x.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
			t.Fatal(err)
		}
	}
	group := router.Group("/api/daily-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", NewDailyCoinHandler(repository.NewCoinRepository(db)).Daily)

	daily := func(query string) (int, DailyCoinResponse) {
		t.Helper()
		w := coinRequest(router, http.MethodGet, "/api/daily-test"+query, nil)
		var got DailyCoinResponse
		_ = json.Unmarshal(w.Body.Bytes(), &got)
		return w.Code, got
	}
	seen := map[string]bool{}
	for _, date := range []string{"2026-10-13", "2026-10-14", "2026-10-15", "2026-10-16", "2026-10-17", "2026-10-18"} {
		code, first := daily("?date=" + date)
		_, again := daily("?kind=coin&date=" + date)
		if code != http.StatusOK || first.Coin == nil || first.Date != date || again.Coin == nil || again.Coin.ID != first.Coin.ID {
			t.Fatalf("%s: %d %+v, then %+v", date, code, first, again)
		}
		if len(first.Coin.Images) == 0 || first.Coin.IsSold {
			t.Fatalf("%s: %q has no photo or is sold", date, first.Coin.Name)
		}
		seen[first.Coin.Name] = true
	}
	if len(seen) != 2 {
		t.Fatalf("over six days = %v", seen)
	}
	if _, note := daily("?kind=banknote&date=2026-10-13"); note.Coin == nil || note.Coin.Name != "1000 escudos" {
		t.Fatalf("banknote of the day = %+v", note)
	}
	if code, _ := daily("?date=13/10/2026"); code != http.StatusBadRequest {
		t.Fatalf("bad date: %d", code)
	}
}
