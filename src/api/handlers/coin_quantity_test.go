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

// A coin or note is one unless the collector says how many; the number can
// be raised later, is never below one, and a copy starts again at one. The
// summary counts every coin held and its value with the quantities, and the
// different coins apart.
func TestCoinQuantity_DefaultsToOneAndCanBeRaised(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")

	create := func(body map[string]any) models.Coin {
		t.Helper()
		w := coinRequest(router, http.MethodPost, "/api/coins", body)
		var created models.Coin
		if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || w.Code != http.StatusCreated {
			t.Fatalf("create %v: %d %s", body, w.Code, w.Body.String())
		}
		return created
	}
	stored := func(id uint) int {
		t.Helper()
		var coin models.Coin
		if err := db.First(&coin, id).Error; err != nil {
			t.Fatal(err)
		}
		return coin.Quantity
	}

	coin := create(map[string]any{"name": "5 cêntimos 1984", "category": "Modern", "material": "Bronze", "country": "África do Sul"})
	if coin.Quantity != 1 || stored(coin.ID) != 1 {
		t.Fatalf("new coin quantity = %d (stored %d), want 1", coin.Quantity, stored(coin.ID))
	}
	note := create(map[string]any{"name": "1000 escudos", "kind": "banknote", "category": "Modern", "material": "Paper", "quantity": 3, "currentValue": 1.5})
	if note.Quantity != 3 || stored(note.ID) != 3 {
		t.Fatalf("new note quantity = %d (stored %d), want 3", note.Quantity, stored(note.ID))
	}
	if w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "x", "category": "Modern", "material": "Other", "quantity": -2}); w.Code != http.StatusBadRequest {
		t.Fatalf("negative quantity on create: %d %s", w.Code, w.Body.String())
	}

	path := fmt.Sprintf("/api/coins/%d", coin.ID)
	if w := coinRequest(router, http.MethodPut, path, map[string]any{"quantity": 4}); w.Code != http.StatusOK {
		t.Fatalf("raise quantity: %d %s", w.Code, w.Body.String())
	}
	if got := stored(coin.ID); got != 4 {
		t.Fatalf("raised quantity = %d, want 4", got)
	}
	// An edit that does not send the quantity leaves it as it was.
	if w := coinRequest(router, http.MethodPut, path, map[string]any{"grade": "VF"}); w.Code != http.StatusOK {
		t.Fatalf("edit without quantity: %d %s", w.Code, w.Body.String())
	}
	if got := stored(coin.ID); got != 4 {
		t.Fatalf("quantity after another edit = %d, want 4", got)
	}
	for _, bad := range []int{0, 10000} {
		if w := coinRequest(router, http.MethodPut, path, map[string]any{"quantity": bad}); w.Code != http.StatusBadRequest {
			t.Fatalf("quantity %d: %d %s", bad, w.Code, w.Body.String())
		}
	}
	if got := stored(coin.ID); got != 4 {
		t.Fatalf("quantity after rejected edits = %d, want 4", got)
	}

	w := coinRequest(router, http.MethodPost, fmt.Sprintf("/api/coins/%d/duplicate", coin.ID), nil)
	var copied models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &copied); err != nil || w.Code != http.StatusCreated {
		t.Fatalf("duplicate: %d %s", w.Code, w.Body.String())
	}
	if copied.Quantity != 1 {
		t.Fatalf("copy quantity = %d, want 1", copied.Quantity)
	}

	// The summary counts every coin held, and its value, with the quantities;
	// and each different coin once.
	repo := repository.NewCoinRepository(db)
	summary := NewCollectionSummaryHandler(repo).WithSettings(services.NewSettingsService(repository.NewSettingsRepository(db)))
	group := router.Group("/api/quantity-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/summary", summary.Summary)
	for query, want := range map[string]struct {
		coins, unique int64
		value         float64
	}{"": {8, 3, 4.5}, "?kind=coin": {5, 2, 0}, "?kind=banknote": {3, 1, 4.5}} {
		var got struct {
			Coins       int64   `json:"coins"`
			UniqueCoins int64   `json:"uniqueCoins"`
			TotalValue  float64 `json:"totalValue"`
		}
		w := coinRequest(router, http.MethodGet, "/api/quantity-test/summary"+query, nil)
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Coins != want.coins || got.UniqueCoins != want.unique || got.TotalValue != want.value {
			t.Fatalf("summary%s = %+v, want %+v (%s)", query, got, want, w.Body.String())
		}
	}
}
