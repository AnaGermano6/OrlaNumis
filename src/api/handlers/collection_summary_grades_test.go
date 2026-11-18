package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/gin-gonic/gin"
)

func TestCollectionSummaryHandler_ByGrade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	for _, coin := range []models.Coin{
		{Name: "Escudo", Kind: models.CoinKindCoin, Quantity: 3, Grade: "VF"},
		{Name: "Libra", Kind: models.CoinKindCoin, Grade: " VF "},
		{Name: "Euro", Kind: models.CoinKindCoin, Grade: "UNC"},
		{Name: "Sem estado", Kind: models.CoinKindCoin, Quantity: 2},
		{Name: "1000 escudos", Kind: models.CoinKindBanknote, Quantity: 2, Grade: "UNC"},
		{Name: "Desejada", Kind: models.CoinKindCoin, IsWishlist: true, Grade: "XF"},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	handler := NewCollectionSummaryHandler(repository.NewCoinRepository(db))
	group := router.Group("/api/grades-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Summary)

	// Each grade with the pieces held of it, most first; those without one and the wishlist's left out.
	for kind, want := range map[string][]repository.GradeCount{
		"":         {{Grade: "VF", Count: 4}, {Grade: "UNC", Count: 3}},
		"coin":     {{Grade: "VF", Count: 4}, {Grade: "UNC", Count: 1}},
		"banknote": {{Grade: "UNC", Count: 2}},
	} {
		w := coinRequest(router, http.MethodGet, "/api/grades-test?kind="+kind, nil)
		var got struct {
			ByGrade []repository.GradeCount `json:"byGrade"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("summary %q: %d %s", kind, w.Code, w.Body.String())
		}
		if !reflect.DeepEqual(got.ByGrade, want) {
			t.Fatalf("summary %q byGrade = %+v, want %+v", kind, got.ByGrade, want)
		}
	}
}
