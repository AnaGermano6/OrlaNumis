package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func summaryRouter(t *testing.T, ecb stubECBClient, currency string) *gin.Engine {
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
	settingsRepo := repository.NewSettingsRepository(db)
	if err := settingsRepo.Upsert(services.SettingCollectionCurrency, currency); err != nil {
		t.Fatalf("seed currency: %v", err)
	}
	handler := NewCollectionSummaryHandler(repository.NewCoinRepository(db)).WithExchangeRates(rates).WithSettings(services.NewSettingsService(settingsRepo))
	group := router.Group("/api/summary-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Summary)
	return router
}

func TestCollectionSummaryHandler_ConvertsTheTotalToEuros(t *testing.T) {
	w := coinRequest(summaryRouter(t, stubECBClient{}, "USD"), http.MethodGet, "/api/summary-test", nil)
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
	w := coinRequest(summaryRouter(t, stubECBClient{err: errors.New("offline")}, "USD"), http.MethodGet, "/api/summary-test", nil)
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	if got["totalValueEur"] != nil || got["exchangeRate"] != nil || got["totalValue"] != 50.0 {
		t.Fatalf("summary without a rate = %v", got)
	}
}

// In a euro collection the values already are euros: no conversion, even
// when the ECB cannot be reached.
func TestCollectionSummaryHandler_EuroCollectionNeedsNoConversion(t *testing.T) {
	w := coinRequest(summaryRouter(t, stubECBClient{err: errors.New("offline")}, "EUR"), http.MethodGet, "/api/summary-test", nil)
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	if got["currency"] != "EUR" || got["totalValue"] != 50.0 || got["totalValueEur"] != 50.0 || got["exchangeRate"] != nil {
		t.Fatalf("euro summary = %v", got)
	}
}

func TestCollectionSummaryHandler_CountsExtinctCountries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	for _, coin := range []models.Coin{
		{Name: "Dinar", Country: "Jugoslávia", UserID: 1},
		{Name: "Dinar 2", Country: "Jugoslávia, República Federal Socialista da", UserID: 1},
		{Name: "Koruna", Country: "Checoslováquia", UserID: 1},
		{Name: "Escudo", Country: "Portugal", UserID: 1},
		// Not in the collection: on the wishlist.
		{Name: "Kurush", Country: "Império Otomano", UserID: 1, IsWishlist: true},
	} {
		coin.Category, coin.Material = models.CategoryModern, models.MaterialSilver
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	handler := NewCollectionSummaryHandler(repository.NewCoinRepository(db)).WithSettings(services.NewSettingsService(repository.NewSettingsRepository(db)))
	group := router.Group("/api/summary-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", handler.Summary)

	w := coinRequest(router, http.MethodGet, "/api/summary-test", nil)
	var got struct {
		Countries        int64 `json:"countries"`
		ExtinctCountries int64 `json:"extinctCountries"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	// Numista's two names for Yugoslavia are two countries, both extinct.
	if got.Countries != 4 || got.ExtinctCountries != 3 {
		t.Fatalf("summary = %+v", got)
	}
}

func TestCoinHandler_ListsCoinsOfExtinctCountries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	createTestUser(t, db, 2, "other")
	for _, coin := range []models.Coin{
		{Name: "Dinar", Country: "Jugoslávia", UserID: 1},
		{Name: "Kurush", Country: "Império Otomano ", UserID: 1},
		{Name: "Escudo", Country: "Portugal", UserID: 1},
		{Name: "Rublo", Country: "União Soviética", UserID: 2},
	} {
		coin.Category, coin.Material = models.CategoryModern, models.MaterialSilver
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}
	list := func(path string) []string {
		w := coinRequest(router, http.MethodGet, path, nil)
		var got struct {
			Coins []models.Coin `json:"coins"`
			Total int64         `json:"total"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		names := []string{}
		for _, coin := range got.Coins {
			names = append(names, coin.Name)
		}
		return names
	}
	if names := list("/api/coins?extinct=true&sort=name&order=asc"); len(names) != 2 || names[0] != "Dinar" || names[1] != "Kurush" {
		t.Fatalf("extinct = %v", names)
	}
	if names := list("/api/coins?extinct=true&search=Escudo"); len(names) != 0 {
		t.Fatalf("filters combine: %v", names)
	}
	db.Where("country <> ?", "Portugal").Delete(&models.Coin{})
	if names := list("/api/coins?extinct=true"); len(names) != 0 {
		t.Fatalf("no extinct country left: %v", names)
	}
}

func TestCollectionSummaryAndList_CurrentCoins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	for key, payload := range map[string]string{
		"type:100:pt": `{"id": 100, "title": "2 Euros", "type": "Moedas circulantes comemorativas", "commemorated_topic": "Jogos Olímpicos", "demonetization": {"is_demonetized": false}}`,
		"type:200:en": `{"id": 200, "title": "1 Escudo", "type": "Standard circulation coins", "demonetization": {"is_demonetized": true, "demonetization_date": "2002-02-28"}}`,
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: payload}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coins := []models.Coin{
		{Name: "2 euro", Country: "Portugal", UserID: 1},
		{Name: "1 escudo", Country: "Portugal", UserID: 1},
		{Name: "10 dinar", Country: "Jugoslávia", UserID: 1},
		{Name: "Hand-made", Country: "Portugal", UserID: 1},
		{Name: "Wished 2 euro", Country: "Portugal", UserID: 1, IsWishlist: true},
	}
	for i := range coins {
		coins[i].Category, coins[i].Material = models.CategoryModern, models.MaterialSilver
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	for coin, number := range map[int]string{0: "100", 1: "200", 4: "100"} {
		if err := db.Create(&models.CoinReference{CoinID: coins[coin].ID, Catalog: "Numista", Number: number}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coinRepo := repository.NewCoinRepository(db)
	circulation := services.NewCoinCirculationService(coinRepo, services.NewNumistaCatalogService(nil, cache))
	summary := NewCollectionSummaryHandler(coinRepo).WithSettings(services.NewSettingsService(repository.NewSettingsRepository(db))).WithCirculation(circulation)
	group := router.Group("/api/summary-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("", summary.Summary)
	list := NewCoinHandler(coinRepo, services.NewCoinService(coinRepo, nil), services.NewLogger(10)).WithCirculation(circulation)
	listGroup := router.Group("/api/circulation-test")
	listGroup.Use(coinTestAuthMiddleware())
	listGroup.GET("", list.List)

	w := coinRequest(router, http.MethodGet, "/api/summary-test", nil)
	var got struct {
		CurrentCoins         int64                      `json:"currentCoins"`
		CirculationUnknown   int64                      `json:"circulationUnknown"`
		CommemorativeCoins   int64                      `json:"commemorativeCoins"`
		CommemorativeUnknown int64                      `json:"commemorativeUnknown"`
		Classifications      []CollectionClassification `json:"classifications"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
		t.Fatalf("summary: %d %s", w.Code, w.Body.String())
	}
	// The wishlist coin does not count; the Yugoslav coin is not current.
	// Only the 2 euro type commemorates something; Numista says nothing of
	// the dinar and the hand-made coin.
	if got.CurrentCoins != 1 || got.CirculationUnknown != 1 || got.CommemorativeCoins != 1 || got.CommemorativeUnknown != 2 {
		t.Fatalf("summary = %+v", got)
	}
	// Numista's kinds of coin, the usual ones in their order, each with the
	// label it was cached with.
	if len(got.Classifications) != 2 || got.Classifications[0] != (CollectionClassification{Key: "standard", Label: "Standard circulation coins", Count: 1}) || got.Classifications[1] != (CollectionClassification{Key: "circulating-commemorative", Label: "Moedas circulantes comemorativas", Count: 1}) {
		t.Fatalf("classifications = %+v", got.Classifications)
	}

	names := func(path string) []string {
		w := coinRequest(router, http.MethodGet, path, nil)
		var res struct {
			Coins []models.Coin `json:"coins"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil || w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		out := []string{}
		for _, coin := range res.Coins {
			out = append(out, coin.Name)
		}
		return out
	}
	if current := names("/api/circulation-test?circulation=current&wishlist=false"); len(current) != 1 || current[0] != "2 euro" {
		t.Fatalf("current = %v", current)
	}
	if demonetized := names("/api/circulation-test?circulation=demonetized&sort=name&order=asc"); len(demonetized) != 2 || demonetized[0] != "1 escudo" || demonetized[1] != "10 dinar" {
		t.Fatalf("demonetized = %v", demonetized)
	}
	if w := coinRequest(router, http.MethodGet, "/api/circulation-test?circulation=maybe", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown circulation: %d", w.Code)
	}
	if commemorative := names("/api/circulation-test?commemorative=true"); len(commemorative) != 1 || commemorative[0] != "2 euro" {
		t.Fatalf("commemorative = %v", commemorative)
	}
	if regular := names("/api/circulation-test?commemorative=false"); len(regular) != 1 || regular[0] != "1 escudo" {
		t.Fatalf("regular = %v", regular)
	}
	if both := names("/api/circulation-test?commemorative=true&circulation=demonetized"); len(both) != 0 {
		t.Fatalf("commemorative and demonetized = %v", both)
	}
	if standard := names("/api/circulation-test?classification=standard"); len(standard) != 1 || standard[0] != "1 escudo" {
		t.Fatalf("standard = %v", standard)
	}
	if none := names("/api/circulation-test?classification=non-circulating"); len(none) != 0 {
		t.Fatalf("non-circulating = %v", none)
	}
	if mixed := names("/api/circulation-test?classification=circulating-commemorative&circulation=current"); len(mixed) != 1 || mixed[0] != "2 euro" {
		t.Fatalf("circulating commemorative and current = %v", mixed)
	}
	if w := coinRequest(router, http.MethodGet, "/api/circulation-test?commemorative=maybe", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown commemorative: %d", w.Code)
	}

	// One coin's status, by the same rules.
	statusGroup := router.Group("/api/status-test")
	statusGroup.Use(coinTestAuthMiddleware())
	statusGroup.GET("/:id", list.Status)
	status := func(coin models.Coin) services.CoinStatus {
		w := coinRequest(router, http.MethodGet, fmt.Sprintf("/api/status-test/%d", coin.ID), nil)
		var got services.CoinStatus
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("status of %s: %d %s", coin.Name, w.Code, w.Body.String())
		}
		return got
	}
	isTrue := func(value *bool) bool { return value != nil && *value }
	isFalse := func(value *bool) bool { return value != nil && !*value }
	if got := status(coins[0]); !isTrue(got.Current) || !isTrue(got.Commemorative) || got.CommemoratedTopic != "Jogos Olímpicos" {
		t.Fatalf("2 euro = %+v", got)
	}
	if got := status(coins[1]); !isFalse(got.Current) || got.DemonetizationDate != "2002-02-28" || !isFalse(got.Commemorative) {
		t.Fatalf("1 escudo = %+v", got)
	}
	// An extinct country's coin is not current; Numista says nothing else of it.
	if got := status(coins[2]); !isFalse(got.Current) || got.Commemorative != nil {
		t.Fatalf("10 dinar = %+v", got)
	}
	if got := status(coins[3]); got.Current != nil || got.Commemorative != nil {
		t.Fatalf("hand-made = %+v", got)
	}
	if w := coinRequest(router, http.MethodGet, "/api/status-test/99999", nil); w.Code != http.StatusNotFound {
		t.Fatalf("missing coin: %d", w.Code)
	}

	// A circulation set by hand wins over Numista: the hand-made coin becomes
	// current and commemorative, the 2 euro coin demonetized and regular.
	if err := db.Model(&coins[3]).Update("circulation", "current_commemorative").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&coins[0]).Update("circulation", "demonetized").Error; err != nil {
		t.Fatal(err)
	}
	coins[0].Circulation, coins[3].Circulation = "demonetized", "current_commemorative"
	if got := status(coins[3]); !isTrue(got.Current) || !isTrue(got.Commemorative) || !got.Manual {
		t.Fatalf("hand-made set by hand = %+v", got)
	}
	if got := status(coins[0]); !isFalse(got.Current) || !isFalse(got.Commemorative) || got.CommemoratedTopic != "" || !got.Manual {
		t.Fatalf("2 euro set by hand = %+v", got)
	}
	w = coinRequest(router, http.MethodGet, "/api/summary-test", nil)
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.CurrentCoins != 1 || got.CirculationUnknown != 0 || got.CommemorativeCoins != 1 || got.CommemorativeUnknown != 1 {
		t.Fatalf("summary with circulations set by hand = %+v", got)
	}
	if current := names("/api/circulation-test?circulation=current&commemorative=true"); len(current) != 1 || current[0] != "Hand-made" {
		t.Fatalf("current commemorative = %v", current)
	}
}
