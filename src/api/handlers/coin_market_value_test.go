package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type stubNumistaPriceClient struct{ calls int }

func (s *stubNumistaPriceClient) Issues(context.Context, int) ([]byte, error) {
	s.calls++
	return []byte(`[{"id": 7, "gregorian_year": 1968}]`), nil
}

func (s *stubNumistaPriceClient) IssuePrices(context.Context, int, int, string) ([]byte, error) {
	s.calls++
	return []byte(`{"currency":"USD","prices":[{"grade":"vf","price":12.5}]}`), nil
}

func setupMarketValueRouter(t *testing.T) (*gin.Engine, *gorm.DB, *stubNumistaPriceClient) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupCoinHandlerTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatalf("migrate cache: %v", err)
	}
	client := &stubNumistaPriceClient{}
	svc := services.NewNumistaMarketValueService(client, services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil))
	handler := NewCoinMarketValueHandler(repository.NewCoinRepository(db), svc)

	r := gin.New()
	protected := r.Group("/api")
	protected.Use(coinTestAuthMiddleware())
	protected.GET("/coins/:id/market-value", handler.MarketValue)
	return r, db, client
}

func getMarketValue(router *gin.Engine, path string, userID uint) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", authHeader(userID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCoinMarketValueHandler_ReturnsPricesForOwnedLinkedCoin(t *testing.T) {
	router, db, client := setupMarketValueRouter(t)
	createTestUser(t, db, 1, "collector")
	coin := models.Coin{Name: "Escudo", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1, DateRange: "1968"}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("seed coin: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: "2933", URI: "https://en.numista.com/catalogue/pieces2933.html"}).Error; err != nil {
		t.Fatalf("seed reference: %v", err)
	}

	w := getMarketValue(router, "/api/coins/"+itoa(coin.ID)+"/market-value?currency=usd", 1)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp models.CoinMarketValue
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != models.NumistaDataStatusOK || resp.Currency != "USD" || resp.SelectedIssueID != 7 || len(resp.Prices) != 1 || resp.Prices[0].Price != 12.5 {
		t.Fatalf("unexpected response: %+v", resp)
	}

	// A second view is served from the stored copy.
	getMarketValue(router, "/api/coins/"+itoa(coin.ID)+"/market-value?currency=USD", 1)
	if client.calls != 2 {
		t.Fatalf("expected 2 provider calls in total, got %d", client.calls)
	}
}

func TestCoinMarketValueHandler_RejectsOtherUsersCoinsAndBadInput(t *testing.T) {
	router, db, client := setupMarketValueRouter(t)
	createTestUser(t, db, 1, "owner")
	createTestUser(t, db, 2, "other")
	coin := models.Coin{Name: "Denarius", Category: models.CategoryRoman, Material: models.MaterialSilver, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("seed coin: %v", err)
	}
	base := "/api/coins/" + itoa(coin.ID) + "/market-value"

	if w := getMarketValue(router, base, 2); w.Code != http.StatusNotFound {
		t.Fatalf("other user's coin: expected 404, got %d", w.Code)
	}
	for _, query := range []string{"?currency=euro", "?currency=E1R", "?issueId=0", "?issueId=abc"} {
		if w := getMarketValue(router, base+query, 1); w.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", query, w.Code)
		}
	}

	w := getMarketValue(router, base, 1)
	var resp models.CoinMarketValue
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if w.Code != http.StatusOK || resp.Status != models.NumistaDataStatusNotLinked || resp.Currency != "EUR" {
		t.Fatalf("unlinked coin: %d %+v", w.Code, resp)
	}
	if client.calls != 0 {
		t.Fatalf("provider must not be called, got %d calls", client.calls)
	}
}
