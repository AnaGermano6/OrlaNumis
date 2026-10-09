package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
)

type stubECBClient struct{ err error }

func (s stubECBClient) DailyRates(context.Context) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []byte(`<Envelope><Cube><Cube time='2026-10-09'><Cube currency='USD' rate='1.25'/></Cube></Cube></Envelope>`), nil
}

func summaryRouter(t *testing.T, ecb stubECBClient) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatalf("migrate cache: %v", err)
	}
	for _, value := range []float64{23.81, 26.19} {
		coin := models.Coin{Name: "Escudo", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1, CurrentValue: &value}
		if err := db.Create(&coin).Error; err != nil {
			t.Fatalf("seed coin: %v", err)
		}
	}
	rates := services.NewExchangeRateService(ecb, services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil))
	handler := NewCollectionSummaryHandler(repository.NewCoinRepository(db)).WithExchangeRates(rates)
	group := router.Group("/api/summary-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Summary)
	return router
}

func TestCollectionSummaryHandler_ConvertsTheTotalToEuros(t *testing.T) {
	w := coinRequest(summaryRouter(t, stubECBClient{}), http.MethodGet, "/api/summary-test", nil)
	var got struct {
		Coins         int64    `json:"coins"`
		TotalValue    float64  `json:"totalValue"`
		TotalValueEUR *float64 `json:"totalValueEur"`
		ExchangeRate  *struct {
			USDPerEUR float64 `json:"usdPerEur"`
			Date      string  `json:"date"`
		} `json:"exchangeRate"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	if got.Coins != 2 || got.TotalValue != 50 || got.TotalValueEUR == nil || *got.TotalValueEUR != 40 || got.ExchangeRate == nil || got.ExchangeRate.Date != "2026-10-09" {
		t.Fatalf("summary = %+v (rate %+v)", got, got.ExchangeRate)
	}
	var withHighlights struct {
		Highlights struct {
			Rarest       *struct{ Name string } `json:"rarest"`
			MostValuable *struct {
				Name         string   `json:"name"`
				CurrentValue *float64 `json:"currentValue"`
			} `json:"mostValuable"`
		} `json:"highlights"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &withHighlights); err != nil || withHighlights.Highlights.Rarest != nil ||
		withHighlights.Highlights.MostValuable == nil || *withHighlights.Highlights.MostValuable.CurrentValue != 26.19 {
		t.Fatalf("highlights = %+v (%v)", withHighlights.Highlights, err)
	}
}

func TestCollectionSummaryHandler_LeavesEurosEmptyWithoutARate(t *testing.T) {
	w := coinRequest(summaryRouter(t, stubECBClient{err: errors.New("offline")}), http.MethodGet, "/api/summary-test", nil)
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	if got["totalValueEur"] != nil || got["exchangeRate"] != nil || got["totalValue"] != 50.0 {
		t.Fatalf("summary without a rate = %v", got)
	}
}
