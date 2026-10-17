package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var setHandlerDBCounter uint64

func setupSetHandlerRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupSetHandlerTestDB(t)
	setRepo := repository.NewSetRepository(db)
	setService := services.NewSetService(setRepo)
	handler := NewSetHandler(setRepo, setService)

	r := gin.New()
	protected := r.Group("/api")
	protected.Use(coinTestAuthMiddleware())
	protected.PUT("/sets/:id/coins/order", handler.ReorderCoins)
	protected.PUT("/sets/:id", handler.Update)
	return r, db
}

func setupSetHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:set_handler_%d_%d?mode=memory&cache=shared",
		time.Now().UnixNano(), atomic.AddUint64(&setHandlerDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.User{}, &models.Coin{}, &models.CoinImage{},
		&models.CoinSet{}, &models.CoinSetMembership{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestSetHandler_ReorderCoins_SavesManualOrder(t *testing.T) {
	router, db := setupSetHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	set, coins := createSetWithCoins(t, db, 1, models.CoinSetTypeStandard, []string{"Trajan", "Augustus", "Hadrian"})

	body := map[string][]uint{"coinIds": []uint{coins[1].ID, coins[2].ID, coins[0].ID}}
	w := performSetOrderRequest(t, router, set.ID, 1, body)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	got, err := repository.NewSetRepository(db).GetCoinsInSet(set.ID, 1)
	if err != nil {
		t.Fatalf("GetCoinsInSet failed: %v", err)
	}
	want := []string{"Augustus", "Hadrian", "Trajan"}
	if names := handlerCoinNames(got); !equalStrings(names, want) {
		t.Fatalf("expected order %v, got %v", want, names)
	}
}

func TestSetHandler_ReorderCoins_RejectsNonMemberAndPreservesOrder(t *testing.T) {
	router, db := setupSetHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	set, coins := createSetWithCoins(t, db, 1, models.CoinSetTypeStandard, []string{"Augustus", "Trajan"})
	nonMember := models.Coin{Name: "Nero", UserID: 1}
	if err := db.Create(&nonMember).Error; err != nil {
		t.Fatalf("create non-member: %v", err)
	}

	body := map[string][]uint{"coinIds": []uint{coins[1].ID, nonMember.ID}}
	w := performSetOrderRequest(t, router, set.ID, 1, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	got, err := repository.NewSetRepository(db).GetCoinsInSet(set.ID, 1)
	if err != nil {
		t.Fatalf("GetCoinsInSet failed: %v", err)
	}
	want := []string{"Augustus", "Trajan"}
	if names := handlerCoinNames(got); !equalStrings(names, want) {
		t.Fatalf("order changed after rejected request: want %v, got %v", want, names)
	}
}

func TestSetHandler_ReorderCoins_RejectsSmartSet(t *testing.T) {
	router, db := setupSetHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	set := models.CoinSet{UserID: 1, Name: "Smart Romans", SetType: models.CoinSetTypeSmart}
	coin := models.Coin{Name: "Augustus", UserID: 1}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("create coin: %v", err)
	}

	body := map[string][]uint{"coinIds": []uint{coin.ID}}
	w := performSetOrderRequest(t, router, set.ID, 1, body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func createSetWithCoins(t *testing.T, db *gorm.DB, userID uint, setType models.CoinSetType, names []string) (models.CoinSet, []models.Coin) {
	t.Helper()
	set := models.CoinSet{UserID: userID, Name: "Set", SetType: setType}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}

	coins := make([]models.Coin, 0, len(names))
	for _, name := range names {
		coins = append(coins, models.Coin{Name: name, UserID: userID})
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}

	addedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	memberships := make([]models.CoinSetMembership, 0, len(coins))
	for i, coin := range coins {
		memberships = append(memberships, models.CoinSetMembership{
			SetID:     set.ID,
			CoinID:    coin.ID,
			AddedAt:   addedAt,
			SortOrder: i,
		})
	}
	if err := db.Create(&memberships).Error; err != nil {
		t.Fatalf("create memberships: %v", err)
	}

	return set, coins
}

func performSetOrderRequest(t *testing.T, router *gin.Engine, setID, userID uint, body map[string][]uint) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/sets/"+strconvUint(setID)+"/coins/order", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(userID))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func strconvUint(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

func handlerCoinNames(coins []models.Coin) []string {
	names := make([]string, 0, len(coins))
	for _, coin := range coins {
		names = append(names, coin.Name)
	}
	return names
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestSetHandler_GetChecklist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSetHandlerTestDB(t)
	if err := db.AutoMigrate(&models.CoinReference{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	createTestUser(t, db, 1, "owner")
	setRepo := repository.NewSetRepository(db)
	checklist := services.NewSetChecklistService(services.NewCoinReleasesService(nil), repository.NewCoinRepository(db), nil)
	handler := NewSetHandler(setRepo, services.NewSetService(setRepo).WithChecklist(checklist))
	r := gin.New()
	protected := r.Group("/api")
	protected.Use(coinTestAuthMiddleware())
	protected.GET("/sets/:id/checklist", handler.GetChecklist)

	year := 2002
	if err := db.Create(&models.Coin{UserID: 1, Name: "1 euro", Country: "Portugal", Denomination: "1 euro", Year: &year}).Error; err != nil {
		t.Fatalf("create coin: %v", err)
	}
	series := models.JSONObject{"source": "challenge", "challenge": "annual-series", "country": "PT", "year": 2002}
	set := models.CoinSet{UserID: 1, Name: "Portugal 2002", SetType: models.CoinSetTypeChecklist, Checklist: &series}
	standard := models.CoinSet{UserID: 1, Name: "Mine", SetType: models.CoinSetTypeStandard}
	if err := db.Create(&[]*models.CoinSet{&set, &standard}).Error; err != nil {
		t.Fatalf("create sets: %v", err)
	}
	get := func(path string, userID uint) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/api/sets/"+path+"/checklist", nil)
		req.Header.Set("Authorization", authHeader(userID))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	w := get(strconvUint(set.ID), 1)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var body SetChecklistResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	owned := []string{}
	for _, slot := range body.Slots {
		if slot.Owned {
			owned = append(owned, slot.Key)
		}
	}
	if len(body.Slots) != 8 || !body.Complete || !equalStrings(owned, []string{"annual:1euro"}) {
		t.Fatalf("checklist = %+v", body)
	}
	for path, want := range map[string]int{strconvUint(standard.ID): http.StatusBadRequest, "nine": http.StatusBadRequest, "999": http.StatusNotFound} {
		if w := get(path, 1); w.Code != want {
			t.Fatalf("GET %s: expected %d, got %d: %s", path, want, w.Code, w.Body.String())
		}
	}
	// Another user's set is not found.
	createTestUser(t, db, 2, "other")
	if w := get(strconvUint(set.ID), 2); w.Code != http.StatusNotFound {
		t.Fatalf("other user: expected 404, got %d", w.Code)
	}
}
