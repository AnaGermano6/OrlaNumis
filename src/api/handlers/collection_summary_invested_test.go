package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/gin-gonic/gin"
)

func TestCollectionSummaryHandler_Invested(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	price := func(v float64) *float64 { return &v }
	for _, coin := range []models.Coin{
		{Name: "Escudo", Kind: models.CoinKindCoin, Quantity: 3, PurchasePrice: price(2.5)},
		{Name: "Libra", Kind: models.CoinKindCoin, PurchasePrice: price(300)},
		{Name: "Sem preço", Kind: models.CoinKindCoin, Quantity: 2},
		{Name: "1000 escudos", Kind: models.CoinKindBanknote, Quantity: 2, PurchasePrice: price(10)},
		{Name: "Desejada", Kind: models.CoinKindCoin, IsWishlist: true, PurchasePrice: price(999)},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	handler := NewCollectionSummaryHandler(repository.NewCoinRepository(db))
	group := router.Group("/api/invested-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Summary)

	for kind, want := range map[string]struct {
		invested float64
		pieces   int64
	}{
		// Each price times its quantity; the pieces without a price and the wishlist's left out.
		"":         {invested: 7.5 + 300 + 20, pieces: 6},
		"coin":     {invested: 307.5, pieces: 4},
		"banknote": {invested: 20, pieces: 2},
	} {
		w := coinRequest(router, http.MethodGet, "/api/invested-test?kind="+kind, nil)
		var got struct {
			Invested       float64 `json:"invested"`
			InvestedPieces int64   `json:"investedPieces"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("summary %q: %d %s", kind, w.Code, w.Body.String())
		}
		if got.Invested != want.invested || got.InvestedPieces != want.pieces {
			t.Fatalf("summary %q = %+v, want %+v", kind, got, want)
		}
	}
}
