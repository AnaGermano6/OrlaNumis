package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

func TestStats_DecadesAndPreciousMetals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	for key, payload := range map[string]string{
		"type:100:pt": `{"id": 100, "title": "1000 Escudos", "composition": {"text": "Prata 925‰"}}`,
		"type:200:pt": `{"id": 200, "title": "Libra", "composition": {"text": "Ouro 916,66‰"}}`,
		"type:300:pt": `{"id": 300, "title": "2 Euros", "composition": {"text": "Cobre  banhado em  prata"}}`,
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: payload}).Error; err != nil {
			t.Fatal(err)
		}
	}
	ptr := func(v float64) *float64 { return &v }
	year := func(v int) *int { return &v }
	coins := []models.Coin{
		{Name: "1000 escudos", Material: models.MaterialSilver, WeightGrams: ptr(10), Quantity: 2, Year: year(1995), PurchasePrice: ptr(20), CurrentValue: ptr(30)},
		{Name: "Libra", Material: models.MaterialGold, WeightGrams: ptr(7.98), Year: year(1912), PurchasePrice: ptr(300)},
		{Name: "Plated", Material: models.MaterialSilver, WeightGrams: ptr(5), Year: year(-45), CurrentValue: ptr(5)},
		{Name: "Billon", Material: models.MaterialBillon, WeightGrams: ptr(3), Year: year(1998), PurchasePrice: ptr(1), CurrentValue: ptr(2)},
		{Name: "Wanted", Material: models.MaterialSilver, WeightGrams: ptr(100), Year: year(2001), IsWishlist: true, PurchasePrice: ptr(99)},
	}
	for i := range coins {
		coins[i].UserID, coins[i].Category = 1, models.CategoryModern
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	for coin, number := range map[int]string{0: "100", 1: "200", 2: "300"} {
		if err := db.Create(&models.CoinReference{CoinID: coins[coin].ID, Catalog: "Numista", Number: number}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coinRepo := repository.NewCoinRepository(db)
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	handler := NewCoinHandler(coinRepo, services.NewCoinService(coinRepo), services.NewLogger(10)).
		WithPreciousMetals(services.NewPreciousMetalsService(coinRepo, services.NewNumistaCatalogService(nil, cache)))
	group := router.Group("/api/stats-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Stats)

	w := coinRequest(router, http.MethodGet, "/api/stats-test", nil)
	var got struct {
		ByDecade       []repository.DecadeCount `json:"byDecade"`
		PreciousMetals services.PreciousMetals  `json:"preciousMetals"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("stats: %d %s", w.Code, w.Body.String())
	}
	// The coins by decade, oldest first (before Christ too); not the wishlist's.
	want := []repository.DecadeCount{{Decade: -40, Count: 1}, {Decade: 1910, Count: 1}, {Decade: 1990, Count: 2}}
	if len(got.ByDecade) != len(want) {
		t.Fatalf("decades = %+v", got.ByDecade)
	}
	for i := range want {
		if got.ByDecade[i] != want[i] {
			t.Fatalf("decades = %+v", got.ByDecade)
		}
	}
	// Gold: the sovereign, 916.66‰. Silver: the two escudos at 925‰, and the
	// billon coin, whose fineness nobody says; the plated coin is none.
	near := func(a, b float64) bool { return math.Abs(a-b) < 0.001 }
	gold, silver := got.PreciousMetals.Gold, got.PreciousMetals.Silver
	if !near(gold.FineGrams, 7.98*0.91666) || !near(gold.GrossGrams, 7.98) || gold.Pieces != 1 || gold.UnknownFineness != 0 {
		t.Fatalf("gold = %+v", gold)
	}
	if !near(silver.FineGrams, 18.5) || !near(silver.GrossGrams, 23) || silver.Pieces != 3 || silver.UnknownFineness != 1 {
		t.Fatalf("silver = %+v", silver)
	}
}
