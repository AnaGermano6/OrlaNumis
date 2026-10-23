package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
		{Name: "2 euro", Country: "Portugal", UserID: 1, Quantity: 3},
		{Name: "1 escudo", Country: "Portugal", UserID: 1},
		{Name: "10 dinar", Country: "Jugoslávia", UserID: 1},
		{Name: "Hand-made", Country: "Portugal", UserID: 1, Quantity: 2},
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
	list := NewCoinHandler(coinRepo, services.NewCoinService(coinRepo), services.NewLogger(10)).WithCirculation(circulation)
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
	// the dinar and the hand-made coin. Each coin counts with its quantity:
	// three 2 euro coins, two hand-made ones.
	if got.CurrentCoins != 3 || got.CirculationUnknown != 2 || got.CommemorativeCoins != 3 || got.CommemorativeUnknown != 3 {
		t.Fatalf("summary = %+v", got)
	}
	// Numista's kinds of coin, the usual ones in their order, each with the
	// label it was cached with, and as many coins as the gallery lists.
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
	// The two hand-made coins are now the current and commemorative ones.
	if got.CurrentCoins != 2 || got.CirculationUnknown != 0 || got.CommemorativeCoins != 2 || got.CommemorativeUnknown != 1 {
		t.Fatalf("summary with circulations set by hand = %+v", got)
	}
	if current := names("/api/circulation-test?circulation=current&commemorative=true"); len(current) != 1 || current[0] != "Hand-made" {
		t.Fatalf("current commemorative = %v", current)
	}

	// The same groups in one answer, for the map.
	groupsRoute := router.Group("/api/status-groups-test")
	groupsRoute.Use(coinTestAuthMiddleware())
	groupsRoute.GET("", list.StatusGroups)
	w = coinRequest(router, http.MethodGet, "/api/status-groups-test", nil)
	var groups CoinStatusGroups
	if err := json.Unmarshal(w.Body.Bytes(), &groups); err != nil || w.Code != http.StatusOK {
		t.Fatalf("status groups: %d %s", w.Code, w.Body.String())
	}
	// Hand-made is current and commemorative; the 2 euro (set by hand), the
	// escudo and the dinar are demonetized; Numista says nothing of the dinar's type.
	if len(groups.Current) != 1 || groups.Current[0] != coins[3].ID || len(groups.Demonetized) != 3 ||
		len(groups.Commemorative) != 1 || groups.Commemorative[0] != coins[3].ID || len(groups.Regular) != 2 {
		t.Fatalf("groups = %+v", groups)
	}
}

func TestCoinCirculation_CollectorCoinsAreNotCurrent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatal(err)
	}
	cache := services.NewNumistaResponseCache(repository.NewNumistaCacheRepository(db), nil, nil)
	for key, payload := range map[string]string{
		// A Monnaie de Paris collector coin: legal tender, but not for circulation.
		"type:535663:pt": `{"id": 535663, "title": "10 Euros (Cinderella, silver 333‰)", "type": "Moedas não circulantes", "commemorated_topic": "Cinderela", "demonetization": {"is_demonetized": false}}`,
		"type:100:pt":    `{"id": 100, "title": "2 Euros", "type": "Moedas circulantes comemorativas", "commemorated_topic": "Jogos Olímpicos", "demonetization": {"is_demonetized": false}}`,
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: payload}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coins := []models.Coin{
		{Name: "10 euro 2025 – Cinderela", Country: "França", UserID: 1},
		{Name: "2 euro", Country: "Portugal", UserID: 1},
		{Name: "Proof set coin", Country: "Portugal", UserID: 1, Circulation: "non_circulating_commemorative"},
		{Name: "Hand-made current", Country: "Portugal", UserID: 1, Circulation: "current"},
	}
	for i := range coins {
		coins[i].Category, coins[i].Material = models.CategoryModern, models.MaterialSilver
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	for coin, number := range map[int]string{0: "535663", 1: "100"} {
		if err := db.Create(&models.CoinReference{CoinID: coins[coin].ID, Catalog: "Numista", Number: number}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coinRepo := repository.NewCoinRepository(db)
	circulation := services.NewCoinCirculationService(coinRepo, services.NewNumistaCatalogService(nil, cache))
	handler := NewCoinHandler(coinRepo, services.NewCoinService(coinRepo), services.NewLogger(10)).WithCirculation(circulation)
	group := router.Group("/api/collector-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/statuses", handler.StatusGroups)
	group.GET("/coins/:id/status", handler.Status)
	group.GET("/coins", handler.List)

	w := coinRequest(router, http.MethodGet, "/api/collector-test/statuses", nil)
	var groups CoinStatusGroups
	if err := json.Unmarshal(w.Body.Bytes(), &groups); err != nil || w.Code != http.StatusOK {
		t.Fatalf("statuses: %d %s", w.Code, w.Body.String())
	}
	if fmt.Sprint(groups.Current) != fmt.Sprint([]uint{coins[1].ID, coins[3].ID}) || fmt.Sprint(groups.NonCirculating) != fmt.Sprint([]uint{coins[0].ID, coins[2].ID}) || len(groups.Demonetized) != 0 {
		t.Fatalf("groups = %+v", groups)
	}
	// Collector coins still count as commemorative.
	if fmt.Sprint(groups.Commemorative) != fmt.Sprint([]uint{coins[0].ID, coins[1].ID, coins[2].ID}) {
		t.Fatalf("commemorative = %v", groups.Commemorative)
	}

	w = coinRequest(router, http.MethodGet, fmt.Sprintf("/api/collector-test/coins/%d/status", coins[0].ID), nil)
	var status services.CoinStatus
	if err := json.Unmarshal(w.Body.Bytes(), &status); err != nil || w.Code != http.StatusOK {
		t.Fatalf("status: %d %s", w.Code, w.Body.String())
	}
	if status.Current == nil || !*status.Current || status.NonCirculating == nil || !*status.NonCirculating || status.Commemorative == nil || !*status.Commemorative {
		t.Fatalf("Cinderella status = %+v", status)
	}

	w = coinRequest(router, http.MethodGet, "/api/collector-test/coins?circulation=current&sort=name&order=asc", nil)
	var list struct {
		Coins []models.Coin `json:"coins"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || len(list.Coins) != 2 || list.Coins[0].Name != "2 euro" || list.Coins[1].Name != "Hand-made current" {
		t.Fatalf("current coins = %d %s", w.Code, w.Body.String())
	}
}

// Banknotes are kept apart from coins in the galleries: the list, the
// countries and the summary can each be limited to one kind.
func TestCoinKind_ListCountriesAndSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	repo := repository.NewCoinRepository(db)
	summary := NewCollectionSummaryHandler(repo).WithSettings(services.NewSettingsService(repository.NewSettingsRepository(db)))
	group := router.Group("/api/kind-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/summary", summary.Summary)
	group.GET("/countries", NewCoinCountryHandler(repo, nil).Countries)

	// Created through the API, a coin is a coin unless it says otherwise.
	for _, body := range []map[string]any{
		{"name": "100 escudos", "category": "Modern", "material": "Other", "country": "Portugal"},
		{"name": "1000 escudos", "kind": "banknote", "category": "Modern", "material": "Other", "country": "Portugal", "currentValue": 12},
		{"name": "10 marcos", "kind": "banknote", "category": "Modern", "material": "Other", "country": "Alemanha"},
	} {
		if w := coinRequest(router, http.MethodPost, "/api/coins", body); w.Code != http.StatusCreated {
			t.Fatalf("create %v: %d %s", body, w.Code, w.Body.String())
		}
	}
	if w := coinRequest(router, http.MethodPost, "/api/coins", map[string]any{"name": "Selo", "kind": "stamp", "category": "Modern", "material": "Other"}); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown kind: %d %s", w.Code, w.Body.String())
	}

	names := func(path string) []string {
		w := coinRequest(router, http.MethodGet, path, nil)
		var got struct {
			Coins []models.Coin `json:"coins"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
		list := []string{}
		for _, coin := range got.Coins {
			list = append(list, coin.Kind+":"+coin.Name)
		}
		return list
	}
	if got := names("/api/coins?kind=coin"); len(got) != 1 || got[0] != "coin:100 escudos" {
		t.Fatalf("coins = %v", got)
	}
	if got := names("/api/coins?kind=banknote&sort=name&order=asc"); len(got) != 2 || got[0] != "banknote:10 marcos" || got[1] != "banknote:1000 escudos" {
		t.Fatalf("banknotes = %v", got)
	}
	if got := names("/api/coins"); len(got) != 3 {
		t.Fatalf("every item without a kind: %v", got)
	}
	if w := coinRequest(router, http.MethodGet, "/api/coins?kind=stamp", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown kind filter: %d", w.Code)
	}

	var countries []repository.CountryCount
	w := coinRequest(router, http.MethodGet, "/api/kind-test/countries?kind=banknote", nil)
	if err := json.Unmarshal(w.Body.Bytes(), &countries); err != nil || len(countries) != 2 || countries[0].Country != "Alemanha" || countries[1].Count != 1 {
		t.Fatalf("banknote countries = %+v (%s)", countries, w.Body.String())
	}

	var got struct {
		Coins      int64   `json:"coins"`
		Countries  int64   `json:"countries"`
		TotalValue float64 `json:"totalValue"`
		Highlights struct {
			MostValuable *struct{ Name string } `json:"mostValuable"`
		} `json:"highlights"`
	}
	w = coinRequest(router, http.MethodGet, "/api/kind-test/summary?kind=coin", nil)
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Coins != 1 || got.Countries != 1 || got.TotalValue != 0 || got.Highlights.MostValuable != nil {
		t.Fatalf("coin summary = %+v (%s)", got, w.Body.String())
	}
	w = coinRequest(router, http.MethodGet, "/api/kind-test/summary?kind=banknote", nil)
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got.Coins != 2 || got.Countries != 2 || got.TotalValue != 12 ||
		got.Highlights.MostValuable == nil || got.Highlights.MostValuable.Name != "1000 escudos" {
		t.Fatalf("banknote summary = %+v (%s)", got, w.Body.String())
	}

	// Editing can turn a banknote back into a coin.
	var banknote models.Coin
	db.Where("name = ?", "10 marcos").First(&banknote)
	if w := coinRequest(router, http.MethodPut, fmt.Sprintf("/api/coins/%d", banknote.ID), map[string]any{"kind": "coin"}); w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	if got := names("/api/coins?kind=coin"); len(got) != 2 {
		t.Fatalf("after edit = %v", got)
	}
}

// A banknote keeps its own details, found by its serial number.
func TestBanknoteDetails_SaveEditAndSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	body := map[string]any{
		"name": "1000 escudos 1996", "kind": "banknote", "category": "Modern", "material": "Paper", "country": "Portugal",
		"series": "Chapa 13", "serialNumber": " CH 1234567 ", "signatures": "António de Sousa / Luís Campos e Cunha", "printer": "Valora",
		"widthMm": 160, "heightMm": 75, "watermark": "Pedro Álvares Cabral", "noteVariant": "replacement",
		"certifiedGrade": "PMG 66 EPQ", "certificationNumber": "8076543-012", "numistaIssueId": 542229,
	}
	w := coinRequest(router, http.MethodPost, "/api/coins", body)
	var created models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil || w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	if created.Kind != "banknote" || created.SerialNumber != "CH 1234567" || created.Series != "Chapa 13" || created.Printer != "Valora" ||
		created.WidthMm == nil || *created.WidthMm != 160 || created.HeightMm == nil || *created.HeightMm != 75 ||
		created.NoteVariant != "replacement" || created.CertifiedGrade != "PMG 66 EPQ" || created.CertificationNumber != "8076543-012" || created.Material != "Paper" ||
		created.NumistaIssueID != 542229 {
		t.Fatalf("created = %+v", created)
	}

	w = coinRequest(router, http.MethodGet, "/api/coins?kind=banknote&search=1234567", nil)
	var found struct {
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &found); err != nil || found.Total != 1 {
		t.Fatalf("search by serial number: %s", w.Body.String())
	}

	path := fmt.Sprintf("/api/coins/%d", created.ID)
	if w := coinRequest(router, http.MethodPut, path, map[string]any{"noteVariant": "error"}); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown variant: %d %s", w.Code, w.Body.String())
	}
	if w := coinRequest(router, http.MethodPut, path, map[string]any{"numistaIssueId": -1}); w.Code != http.StatusBadRequest {
		t.Fatalf("negative issue: %d %s", w.Code, w.Body.String())
	}
	if w := coinRequest(router, http.MethodPut, path, map[string]any{"noteVariant": "", "heightMm": nil, "signatures": "Vítor Constâncio", "numistaIssueId": 0}); w.Code != http.StatusOK {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	var saved models.Coin
	db.First(&saved, created.ID)
	if saved.NoteVariant != "" || saved.HeightMm != nil || saved.WidthMm == nil || saved.Signatures != "Vítor Constâncio" || saved.SerialNumber != "CH 1234567" || saved.NumistaIssueID != 0 {
		t.Fatalf("saved = %+v", saved)
	}
}

// A banknote's copy is a banknote of the same issue, without the copied
// note's serial number or certificate.
func TestBanknoteDuplicate_KeepsTheKindAndIssueDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	width := 160.0
	note := models.Coin{
		Name: "2000 escudos", Kind: models.CoinKindBanknote, Category: models.CategoryModern, Material: "Paper", UserID: 1,
		Series: "Chapa 1", Printer: "Valora", WidthMm: &width, Watermark: "Bartolomeu Dias", NoteVariant: "replacement",
		SerialNumber: "AB 123", CertifiedGrade: "PMG 64 EPQ", CertificationNumber: "123-001", NumistaIssueID: 513930,
	}
	if err := db.Create(&note).Error; err != nil {
		t.Fatal(err)
	}
	w := coinRequest(router, http.MethodPost, fmt.Sprintf("/api/coins/%d/duplicate", note.ID), nil)
	var copied models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &copied); err != nil || w.Code != http.StatusCreated {
		t.Fatalf("duplicate: %d %s", w.Code, w.Body.String())
	}
	if copied.Kind != models.CoinKindBanknote || copied.Series != "Chapa 1" || copied.Printer != "Valora" || copied.WidthMm == nil ||
		copied.Watermark != "Bartolomeu Dias" || copied.NoteVariant != "replacement" || copied.CertifiedGrade != "PMG 64 EPQ" ||
		copied.NumistaIssueID != 513930 || copied.SerialNumber != "" || copied.CertificationNumber != "" {
		t.Fatalf("copied = %+v", copied)
	}
}

// A country's periods, the most recent first, with their coins' years; the
// gallery filters by one.
func TestCoinPeriods_ListAndFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "collector")
	repo := repository.NewCoinRepository(db)
	group := router.Group("/api/period-test")
	group.Use(coinTestAuthMiddleware())
	group.GET("/periods", NewCoinCountryHandler(repo, nil).Periods)
	year := func(y int) *int { return &y }
	for _, coin := range []models.Coin{
		{Name: "50 Centavos 1945", Country: "Portugal", Year: year(1945), Period: "República Portuguesa 1910–1969"},
		{Name: "100 Escudos 1990", Country: "Portugal", Year: year(1990), Period: "República Portuguesa 1986–2001"},
		{Name: "2 Euro 2002", Country: "Portugal", Year: year(2002), Period: "União Europeia (Euro)"},
		{Name: "2 Euro 2024", Country: "Portugal", Year: year(2024), Period: "União Europeia (Euro)"},
		{Name: "Wanted", Country: "Portugal", Year: year(1930), Period: "República Portuguesa 1910–1969", IsWishlist: true},
		{Name: "Penny", Country: "Reino Unido", Year: year(1971), Period: "Rainha Isabel II"},
		{Name: "Undated", Country: "Portugal"},
	} {
		coin.UserID, coin.Category, coin.Material = 1, models.CategoryModern, models.MaterialOther
		if err := db.Create(&coin).Error; err != nil {
			t.Fatal(err)
		}
	}

	w := coinRequest(router, http.MethodGet, "/api/period-test/periods?country=Portugal&kind=coin", nil)
	var periods []repository.PeriodCount
	if err := json.Unmarshal(w.Body.Bytes(), &periods); err != nil || w.Code != http.StatusOK {
		t.Fatalf("periods: %d %s", w.Code, w.Body.String())
	}
	got := []string{}
	for _, period := range periods {
		got = append(got, fmt.Sprintf("%s %d %d–%d", period.Period, period.Count, *period.FromYear, *period.ToYear))
	}
	want := "União Europeia (Euro) 2 2002–2024|República Portuguesa 1986–2001 1 1990–1990|República Portuguesa 1910–1969 1 1945–1945"
	if strings.Join(got, "|") != want {
		t.Fatalf("periods = %s", strings.Join(got, "|"))
	}
	if w := coinRequest(router, http.MethodGet, "/api/period-test/periods", nil); w.Code != http.StatusBadRequest {
		t.Fatalf("without a country: %d", w.Code)
	}

	w = coinRequest(router, http.MethodGet, "/api/coins?country=Portugal&period=Uni%C3%A3o+Europeia+%28Euro%29&wishlist=false&sort=name&order=asc", nil)
	var list struct {
		Coins []models.Coin `json:"coins"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil || w.Code != http.StatusOK || len(list.Coins) != 2 || list.Coins[0].Name != "2 Euro 2002" {
		t.Fatalf("filtered by period: %d %s", w.Code, w.Body.String())
	}
}
