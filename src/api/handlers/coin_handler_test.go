package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/briandenicola/ancient-coins-api/services"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const coinTestJWTSecret = "coin-handler-test-secret"

var coinHandlerDBCounter atomic.Uint64

func setupCoinHandlerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:coin_handler_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), coinHandlerDBCounter.Add(1))), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.User{}, &models.Coin{}, &models.CoinImage{}, &models.CoinReference{}, &models.CatalogRegistry{}, &models.AppSetting{},
		&models.CoinJournal{},
		&models.CoinValueHistory{},
		&models.AvailabilityResult{},
		&models.ValuationRun{}, &models.ValuationResult{},
		&models.CoinSet{}, &models.CoinSetMembership{},

		&models.QuickCaptureDraft{}, &models.QuickCaptureDraftImage{}, &models.QuickCaptureDraftReference{}, &models.DraftLifecycleEvent{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func makeCoinTestJWT(userID uint) string {
	claims := jwt.MapClaims{
		"userId":   float64(userID),
		"username": "testuser",
		"role":     "user",
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(coinTestJWTSecret))
	return signed
}

// authMiddleware is a simplified version for testing that extracts userId from JWT
func coinTestAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenString := authHeader[len("Bearer "):]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(coinTestJWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		c.Set("userId", uint(claims["userId"].(float64)))
		c.Set("userRole", claims["role"])
		c.Next()
	}
}

func setupCoinHandlerRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupCoinHandlerTestDB(t)
	coinRepo := repository.NewCoinRepository(db)
	catalogRegistryRepo := repository.NewCatalogRegistryRepository(db)
	coinReferenceRepo := repository.NewCoinReferenceRepository(db)
	coinReferenceSvc := services.NewCoinReferenceService(coinReferenceRepo, catalogRegistryRepo)
	settingsRepo := repository.NewSettingsRepository(db)
	settingsSvc := services.NewSettingsService(settingsRepo)
	coinSvc := services.NewCoinService(coinRepo).
		WithReferenceSupport(coinReferenceRepo, coinReferenceSvc).
		WithCatalogRegistrySupport(catalogRegistryRepo).
		WithSettingsSupport(settingsSvc)
	handler := NewCoinHandler(coinRepo, coinSvc, services.NewLogger(100)).WithSettingsSupport(settingsSvc)
	coinReferenceHandler := NewCoinReferenceHandler(coinReferenceRepo, coinReferenceSvc, nil)

	r := gin.New()
	protected := r.Group("/api")
	protected.Use(coinTestAuthMiddleware())
	protected.GET("/coins", handler.List)
	protected.GET("/coins/:id", handler.Get)
	protected.POST("/coins", handler.Create)
	protected.PUT("/coins/:id", handler.Update)
	protected.POST("/coins/:id/references", coinReferenceHandler.Create)
	protected.POST("/coins/:id/duplicate", handler.Duplicate)
	protected.DELETE("/coins/:id", handler.Delete)
	protected.POST("/coins/match-category-era", handler.MatchCategoryEra)
	protected.GET("/coin-options", handler.GetCoinOptions)

	return r, db
}

func TestCoinReferenceHandler_CreatePersistsCanonicalNumistaReference(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "numista-reference-owner")

	if err := db.Create(&models.CatalogRegistry{
		Catalog:     "Numista",
		DisplayName: "Numista",
		Era:         models.EraModern,
	}).Error; err != nil {
		t.Fatalf("seed Numista catalog: %v", err)
	}
	coin := models.Coin{
		Name: "Selected Numista coin", Category: models.CategoryModern,
		Material: models.MaterialSilver, UserID: 1,
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("seed coin: %v", err)
	}

	body := bytes.NewBufferString(`{"catalog":"numista","number":"12345","uri":"https://en.numista.com/catalogue/pieces12345.html"}`)
	request := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/coins/%d/references", coin.ID), body)
	request.Header.Set("Authorization", authHeader(1))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var persisted models.CoinReference
	if err := db.Where("coin_id = ?", coin.ID).First(&persisted).Error; err != nil {
		t.Fatalf("load persisted reference: %v", err)
	}
	if persisted.Catalog != "Numista" || persisted.Number != "12345" ||
		persisted.URI != "https://en.numista.com/catalogue/pieces12345.html" {
		t.Fatalf("persisted reference changed: %+v", persisted)
	}
}

func TestCoinHandler_Update_PromotedQuickCaptureCoinUsesExistingEditContract(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "promoted-editor")

	coin := models.Coin{
		Name:       "Quick Capture promoted coin",
		Category:   models.CategoryModern,
		Material:   models.MaterialSilver,
		Era:        models.EraAncient,
		UserID:     1,
		Notes:      "Promoted from Quick Capture",
		IsWishlist: false,
	}

	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("create promoted coin: %v", err)
	}
	promotedCoinID := coin.ID
	promotedAt := time.Now()
	if err := db.Create(&models.QuickCaptureDraft{
		UserID:         1,
		WorkingTitle:   "Quick Capture promoted coin",
		Status:         models.QuickCaptureDraftStatusPromoted,
		PromotedCoinID: &promotedCoinID,
		PromotedAt:     &promotedAt,
	}).Error; err != nil {
		t.Fatalf("create linked promoted draft: %v", err)
	}

	payload := map[string]interface{}{
		"name":     "Edited promoted coin",
		"notes":    "Existing edit flow works after promotion",
		"era":      string(models.EraAncient),
		"category": string(models.CategoryOther),
		"material": string(models.MaterialBronze),
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected promoted coin edit to use normal 200 contract, got %d: %s", w.Code, w.Body.String())
	}
	var resp models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode edited coin: %v", err)
	}
	if resp.Name != "Edited promoted coin" || resp.Notes != "Existing edit flow works after promotion" {
		t.Fatalf("promoted coin edit did not persist normal fields: %+v", resp)
	}
	if resp.IsWishlist {
		t.Fatalf("Quick Capture v1 promoted coin should remain active collection coin, got wishlist=%v", resp.IsWishlist)
	}
}

func TestCoinRequestsPreserveCoinDateRange(t *testing.T) {
	created := (CoinCreateRequest{
		Name: "Dated coin", DateRange: "138–161 CE",
	}).toCoin(1)
	if created.DateRange != "138–161 CE" {
		t.Fatalf("created date range = %q", created.DateRange)
	}

	dateRange := "c. 330–335"
	updated, fields := (CoinUpdateRequest{DateRange: &dateRange}).toCoin(
		&models.Coin{ID: 2, UserID: 1, DateRange: "330 CE"},
		map[string]bool{},
	)
	if updated.DateRange != dateRange || !slices.Contains(fields, "DateRange") {
		t.Fatalf("updated date range=%q fields=%v", updated.DateRange, fields)
	}
}

func createTestUser(t *testing.T, db *gorm.DB, id uint, username string) {
	t.Helper()
	user := models.User{ID: id, Username: username, PasswordHash: "hash", Email: username + "@test.com"}
	db.Create(&user)
}

func authHeader(userID uint) string {
	return "Bearer " + makeCoinTestJWT(userID)
}

// --- List ---

func TestCoinHandler_List_Authenticated(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "listuser")

	// Create a coin for user 1
	coin := models.Coin{Name: "Test Denarius", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	db.Create(&coin)

	req := httptest.NewRequest(http.MethodGet, "/api/coins", nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	coins, ok := resp["coins"].([]interface{})
	if !ok {
		t.Fatal("expected coins array in response")
	}
	if len(coins) != 1 {
		t.Errorf("expected 1 coin, got %d", len(coins))
	}
}

func TestCoinHandler_List_OnlyOwnCoins(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "user1")
	createTestUser(t, db, 2, "user2")

	db.Create(&models.Coin{Name: "User1 Coin", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1})
	db.Create(&models.Coin{Name: "User2 Coin", Category: models.CategoryOther, Material: models.MaterialGold, UserID: 2})

	req := httptest.NewRequest(http.MethodGet, "/api/coins", nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	coins := resp["coins"].([]interface{})
	if len(coins) != 1 {
		t.Errorf("expected 1 coin (own), got %d", len(coins))
	}
}

// --- Get ---

func TestCoinHandler_Get_OwnCoin(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "getuser")

	coin := models.Coin{Name: "My Aureus", Category: models.CategoryModern, Material: models.MaterialGold, UserID: 1}
	db.Create(&coin)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/coins/%d", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCoinHandler_Get_OtherUserCoin_NotFound(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	createTestUser(t, db, 2, "intruder")

	coin := models.Coin{Name: "Private Coin", Category: models.CategoryOther, Material: models.MaterialSilver, UserID: 1}
	db.Create(&coin)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/coins/%d", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(2)) // Different user
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for other user's coin, got %d", w.Code)
	}
}

// --- Create ---

func TestCoinHandler_Create_Success(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "creator")

	coinData := map[string]interface{}{
		"name":     "New Tetradrachm",
		"category": "Other",
		"material": "Silver",
	}
	body, _ := json.Marshal(coinData)

	req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "New Tetradrachm" {
		t.Errorf("expected name 'New Tetradrachm', got %v", resp["name"])
	}
	if resp["userId"] != float64(1) {
		t.Errorf("expected userId 1, got %v", resp["userId"])
	}
}

func TestCoinHandler_Create_ReturnsConflictForDuplicateWishlistReferenceURL(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "wishlist-duplicate")
	body, _ := json.Marshal(map[string]interface{}{
		"name":         "Dealer denarius",
		"category":     "Modern",
		"material":     "Silver",
		"referenceUrl": "https://dealer.example/coin/123",
		"isWishlist":   true,
	})

	for attempt, wantStatus := range []int{http.StatusCreated, http.StatusConflict} {
		req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader(1))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		if recorder.Code != wantStatus {
			t.Fatalf("attempt %d status=%d, want %d: %s", attempt+1, recorder.Code, wantStatus, recorder.Body.String())
		}
		if wantStatus == http.StatusConflict && !strings.Contains(recorder.Body.String(), `"code":"wishlist_duplicate"`) {
			t.Fatalf("duplicate response=%s", recorder.Body.String())
		}
	}
}

func TestCoinHandler_Create_RejectsOverlongCategory(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "creator")

	coinData := map[string]interface{}{
		"name":     "Overlong Category Coin",
		"category": strings.Repeat("x", 65),
		"material": "Silver",
	}
	body, _ := json.Marshal(coinData)

	req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an over-length category, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCoinHandler_Create_RejectsUnsupportedCategory(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "creator")

	coinData := map[string]interface{}{
		"name":     "Unsupported Category Coin",
		"category": "Not A Real Category",
		"material": "Silver",
	}
	body, _ := json.Marshal(coinData)

	req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for an unsupported category, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCoinHandler_Create_InvalidPayload(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "creator")

	req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

// --- MatchCategoryEra ---

func TestCoinHandler_MatchCategoryEra_ExactBuiltIn(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "matcher")

	body, _ := json.Marshal(map[string]string{"type": "category", "value": "modern"})
	req := httptest.NewRequest(http.MethodPost, "/api/coins/match-category-era", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp MatchCategoryEraResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Matched || resp.Match != "Modern" {
		t.Fatalf("expected matched=true match=Modern, got %+v", resp)
	}
}

func TestCoinHandler_MatchCategoryEra_AdminDefinedNearMiss(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "matcher")
	settingsRepo := repository.NewSettingsRepository(db)
	if err := settingsRepo.Upsert(services.SettingCoinCategories, "Modern\nOther\nCeltic"); err != nil {
		t.Fatalf("seed coin categories setting: %v", err)
	}

	body, _ := json.Marshal(map[string]string{"type": "category", "value": "celtic coins"})
	req := httptest.NewRequest(http.MethodPost, "/api/coins/match-category-era", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp MatchCategoryEraResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !resp.Matched || resp.Match != "Celtic" {
		t.Fatalf("expected matched=true match=Celtic, got %+v", resp)
	}
}

func TestCoinHandler_MatchCategoryEra_NoConfidentMatch(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "matcher")

	body, _ := json.Marshal(map[string]string{"type": "era", "value": "space age"})
	req := httptest.NewRequest(http.MethodPost, "/api/coins/match-category-era", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp MatchCategoryEraResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Matched {
		t.Fatalf("expected no confident match, got %+v", resp)
	}
}

// --- GetCoinOptions ---

func getCoinOptions(t *testing.T, router *gin.Engine) CoinOptionsResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/coin-options", nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp CoinOptionsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func TestCoinHandler_GetCoinOptions_Defaults(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "options-reader")

	resp := getCoinOptions(t, router)

	if !slices.Equal(resp.Categories, []string{"Modern", "Other"}) {
		t.Fatalf("categories = %v", resp.Categories)
	}
	if !slices.Equal(resp.Eras, []string{"ancient", "medieval", "modern"}) {
		t.Fatalf("eras = %v", resp.Eras)
	}
	pt, ok := resp.Labels["pt-PT"]
	if !ok || len(pt.Categories) != 0 || len(pt.Eras) != 0 {
		t.Fatalf("expected empty pt-PT labels, got %+v", resp.Labels)
	}
	if resp.Currency != "USD" {
		t.Fatalf("currency = %q", resp.Currency)
	}
}

func TestCoinHandler_GetCoinOptions_ConfiguredListsAndLabels(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "options-reader")
	settingsRepo := repository.NewSettingsRepository(db)
	seed := map[string]string{
		services.SettingCoinCategories:         "Modern\nCeltic",
		services.SettingCoinEras:               "ancient\nRepublican",
		services.SettingCoinCategoryLabelsPtPT: "Celtic=Celta",
		services.SettingCoinEraLabelsPtPT:      "Republican=Republicana",
		services.SettingCollectionCurrency:     "EUR",
	}
	for key, value := range seed {
		if err := settingsRepo.Upsert(key, value); err != nil {
			t.Fatalf("seed %s: %v", key, err)
		}
	}

	resp := getCoinOptions(t, router)

	if !slices.Equal(resp.Categories, []string{"Modern", "Celtic"}) {
		t.Fatalf("categories = %v", resp.Categories)
	}
	if !slices.Equal(resp.Eras, []string{"ancient", "Republican"}) {
		t.Fatalf("eras = %v", resp.Eras)
	}
	pt := resp.Labels["pt-PT"]
	if pt.Categories["Celtic"] != "Celta" || pt.Eras["Republican"] != "Republicana" {
		t.Fatalf("pt-PT labels = %+v", pt)
	}
	if resp.Currency != "EUR" {
		t.Fatalf("currency = %q", resp.Currency)
	}
}

// --- Duplicate ---

func TestCoinHandler_Duplicate_PreservesOwnedDataAndAssociationsExcludingMedia(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "owner")

	purchasePrice := 125.50
	currentValue := 175.75
	purchaseDate := time.Date(2025, time.March, 14, 0, 0, 0, 0, time.UTC)
	coin := models.Coin{
		Name:               "Augustus Denarius",
		Category:           models.CategoryModern,
		Denomination:       "Denarius",
		Ruler:              "Augustus",
		Era:                models.EraAncient,
		Mint:               "Rome",
		Material:           models.MaterialSilver,
		PurchasePrice:      &purchasePrice,
		CurrentValue:       &currentValue,
		PurchaseDate:       &purchaseDate,
		PurchaseLocation:   "Auction",
		Notes:              "Cabinet tone",
		ReferenceURL:       "https://example.test/reference",
		ReferenceText:      "RIC I 1",
		UserID:             1,
		ObverseInscription: "CAESAR",
		ReverseInscription: "AVGVSTVS",
		ObverseDescription: "Laureate head",
		ReverseDescription: "Temple",
		AIAnalysis:         "Existing analysis",
		ObverseAnalysis:    "Obverse analysis",
		ReverseAnalysis:    "Reverse analysis",
		ListingStatus:      "available",
		ListingCheckReason: "seeded",
		ListingCheckedAt:   &purchaseDate,
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}
	if err := db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "uploads/original-obverse.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
		t.Fatalf("failed to seed image: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "I", Number: "12", URI: "https://example.test/ric"}).Error; err != nil {
		t.Fatalf("failed to seed reference: %v", err)
	}
	setRepo := repository.NewSetRepository(db)
	set := models.CoinSet{UserID: 1, Name: "Roman Core", SetType: models.CoinSetTypeOpen}
	if err := setRepo.Create(&set); err != nil {
		t.Fatalf("failed to seed set: %v", err)
	}
	if err := setRepo.AddCoinToSet(coin.ID, set.ID, 1, "keeper"); err != nil {
		t.Fatalf("failed to attach set: %v", err)
	}
	if err := db.Model(&models.CoinSetMembership{}).Where("coin_id = ? AND set_id = ?", coin.ID, set.ID).Update("sort_order", 7).Error; err != nil {
		t.Fatalf("failed to set membership sort order: %v", err)
	}
	var originalMembership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, set.ID).First(&originalMembership).Error; err != nil {
		t.Fatalf("failed to load original membership: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/coins/%d/duplicate", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var duplicated models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &duplicated); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if duplicated.ID == 0 || duplicated.ID == coin.ID {
		t.Fatalf("expected a new coin id, got %d from original %d", duplicated.ID, coin.ID)
	}
	if duplicated.Name != "Augustus Denarius (duplicate)" {
		t.Fatalf("expected duplicate suffix, got %q", duplicated.Name)
	}
	if duplicated.UserID != 1 || duplicated.Category != coin.Category || duplicated.Material != coin.Material || duplicated.Era != coin.Era {
		t.Fatalf("expected owned scalar data to be preserved, got user=%d category=%q material=%q era=%q", duplicated.UserID, duplicated.Category, duplicated.Material, duplicated.Era)
	}
	if len(duplicated.Images) != 0 {
		t.Fatalf("expected duplicate response to exclude images/media, got %#v", duplicated.Images)
	}
	if len(duplicated.Sets) != 1 || duplicated.Sets[0].ID != set.ID {
		t.Fatalf("expected set association to be preserved, got %#v", duplicated.Sets)
	}
	if len(duplicated.References) != 1 || duplicated.References[0].Catalog != "RIC" || duplicated.References[0].Number != "12" {
		t.Fatalf("expected reference association to be preserved, got %#v", duplicated.References)
	}

	var newImageCount int64
	if err := db.Model(&models.CoinImage{}).Where("coin_id = ?", duplicated.ID).Count(&newImageCount).Error; err != nil {
		t.Fatalf("failed to count duplicated images: %v", err)
	}
	if newImageCount != 0 {
		t.Fatalf("expected no image rows for duplicated coin, got %d", newImageCount)
	}
	var membership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", duplicated.ID, set.ID).First(&membership).Error; err != nil {
		t.Fatalf("expected duplicated set membership: %v", err)
	}
	if membership.Notes != originalMembership.Notes || membership.SortOrder != originalMembership.SortOrder || !membership.AddedAt.Equal(originalMembership.AddedAt) {
		t.Fatalf("expected membership metadata to be preserved, got %#v want %#v", membership, originalMembership)
	}
}

func TestCoinHandler_Duplicate_RejectsNonOwnedCoin(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	createTestUser(t, db, 2, "intruder")

	coin := models.Coin{Name: "Private Coin", Category: models.CategoryOther, Material: models.MaterialSilver, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/coins/%d/duplicate", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(2))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for non-owned coin, got %d: %s", w.Code, w.Body.String())
	}
	var duplicateCount int64
	if err := db.Model(&models.Coin{}).Where("name = ? AND user_id = ?", "Private Coin (duplicate)", uint(2)).Count(&duplicateCount).Error; err != nil {
		t.Fatalf("failed to count duplicates: %v", err)
	}
	if duplicateCount != 0 {
		t.Fatalf("expected no duplicate for non-owner, got %d", duplicateCount)
	}
}

// --- Update ---

func TestCoinHandler_Update_OwnCoin(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	coin := models.Coin{Name: "Old Name", Category: models.CategoryModern, Material: models.MaterialBronze, UserID: 1}
	db.Create(&coin)

	updates := map[string]interface{}{
		"name":     "Updated Name",
		"category": "Modern",
		"material": "Bronze",
	}

	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCoinHandler_Update_OneFieldPreservesAssociationsAndReadOnlyFields(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")
	createTestUser(t, db, 2, "other")

	currentValue := 250.0
	valuationTime := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Second)
	coin := models.Coin{
		Name:                  "Original Denarius",
		Category:              models.CategoryModern,
		Material:              models.MaterialSilver,
		Era:                   models.EraAncient,
		UserID:                1,
		CurrentValue:          &currentValue,
		CurrentValueUpdatedAt: &valuationTime,
		AIAnalysis:            "existing analysis",
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}
	originalID := coin.ID
	originalCreatedAt := coin.CreatedAt

	if err := db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "coins/original.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
		t.Fatalf("failed to seed image: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "II", Number: "12"}).Error; err != nil {
		t.Fatalf("failed to seed reference: %v", err)
	}
	set := models.CoinSet{UserID: 1, Name: "Roman Core", SetType: models.CoinSetTypeOpen}
	if err := repository.NewSetRepository(db).Create(&set); err != nil {
		t.Fatalf("failed to seed set: %v", err)
	}
	if err := repository.NewSetRepository(db).AddCoinToSet(coin.ID, set.ID, 1, "keeper"); err != nil {
		t.Fatalf("failed to attach set: %v", err)
	}

	updates := map[string]interface{}{
		"id":                    originalID + 100,
		"userId":                2,
		"name":                  "Renamed Denarius",
		"createdAt":             time.Now().Add(-24 * time.Hour),
		"aiAnalysis":            "incoming analysis",
		"currentValueUpdatedAt": time.Now(),
		"images": []map[string]interface{}{
			{"id": 999, "coinId": coin.ID, "filePath": "coins/replacement.jpg", "imageType": "reverse"},
		},
		"tags": []map[string]interface{}{
			{"id": 999, "name": "Incoming"},
		},
		"sets": []map[string]interface{}{
			{"id": set.ID + 100, "name": "Incoming Set"},
		},
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.Preload("Images").Preload("References").Preload("Sets").First(&found, originalID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if found.ID != originalID {
		t.Fatalf("expected read-only id %d to be preserved, got %d", originalID, found.ID)
	}
	if found.UserID != 1 {
		t.Fatalf("expected read-only userId 1 to be preserved, got %d", found.UserID)
	}
	if !found.CreatedAt.Equal(originalCreatedAt) {
		t.Fatalf("expected read-only createdAt to be preserved, got %v want %v", found.CreatedAt, originalCreatedAt)
	}
	if found.Name != "Renamed Denarius" {
		t.Fatalf("expected name update, got %q", found.Name)
	}
	if found.Category != models.CategoryModern || found.Material != models.MaterialSilver || found.Era != models.EraAncient {
		t.Fatalf("unexpected scalar sibling mutation: category=%q material=%q era=%q", found.Category, found.Material, found.Era)
	}
	if found.AIAnalysis != "existing analysis" {
		t.Fatalf("expected read-only aiAnalysis to be preserved, got %q", found.AIAnalysis)
	}
	if found.CurrentValueUpdatedAt == nil || !found.CurrentValueUpdatedAt.Equal(valuationTime) {
		t.Fatalf("expected currentValueUpdatedAt to be preserved, got %v want %v", found.CurrentValueUpdatedAt, valuationTime)
	}
	if len(found.Images) != 1 || found.Images[0].FilePath != "coins/original.jpg" {
		t.Fatalf("expected original image association to remain, got %#v", found.Images)
	}
	if len(found.References) != 1 || found.References[0].Number != "12" {
		t.Fatalf("expected original reference association to remain, got %#v", found.References)
	}
	if len(found.Sets) != 1 || found.Sets[0].ID != set.ID {
		t.Fatalf("expected original set association to remain, got %#v", found.Sets)
	}

	var valueHistoryCount int64
	if err := db.Model(&models.CoinValueHistory{}).Where("coin_id = ?", coin.ID).Count(&valueHistoryCount).Error; err != nil {
		t.Fatalf("failed to count value history: %v", err)
	}
	if valueHistoryCount != 0 {
		t.Fatalf("expected no manual value history for name-only edit, got %d", valueHistoryCount)
	}
}

func TestCoinHandler_Update_IgnoresUnknownReadOnlyAndBroadRelationshipFields(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")
	createTestUser(t, db, 2, "other")

	coin := models.Coin{Name: "Typed Contract Coin", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}
	if err := db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "coins/original.jpg", ImageType: models.ImageTypeObverse}).Error; err != nil {
		t.Fatalf("failed to seed image: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Number: "1"}).Error; err != nil {
		t.Fatalf("failed to seed reference: %v", err)
	}
	set := models.CoinSet{UserID: 1, Name: "Original Set", SetType: models.CoinSetTypeOpen}
	if err := repository.NewSetRepository(db).Create(&set); err != nil {
		t.Fatalf("failed to seed set: %v", err)
	}
	if err := repository.NewSetRepository(db).AddCoinToSet(coin.ID, set.ID, 1, "original"); err != nil {
		t.Fatalf("failed to attach set: %v", err)
	}

	updates := map[string]interface{}{
		"name":       "Typed Contract Coin Updated",
		"unknownKey": "ignored",
		"userId":     2,
		"storageLocation": map[string]interface{}{
			"id":   99,
			"name": "Injected Location",
		},
		"images": []map[string]interface{}{
			{"filePath": "coins/injected.jpg", "imageType": "reverse"},
		},
		"tags": []map[string]interface{}{
			{"id": 999, "name": "Injected Tag"},
		},
		"sets": []map[string]interface{}{
			{"id": set.ID + 100, "name": "Injected Set"},
		},
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with ignored unknown/broad fields, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.Preload("Images").Preload("References").Preload("Sets").First(&found, coin.ID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if found.Name != "Typed Contract Coin Updated" {
		t.Fatalf("expected allowlisted name update, got %q", found.Name)
	}
	if found.UserID != 1 {
		t.Fatalf("expected read-only userId to remain 1, got %d", found.UserID)
	}
	if len(found.Images) != 1 || found.Images[0].FilePath != "coins/original.jpg" {
		t.Fatalf("expected image relationship payload to be ignored, got %#v", found.Images)
	}
	if len(found.References) != 1 || found.References[0].Number != "1" {
		t.Fatalf("expected omitted references to be preserved, got %#v", found.References)
	}
	if len(found.Sets) != 1 || found.Sets[0].ID != set.ID {
		t.Fatalf("expected set relationship payload to be ignored, got %#v", found.Sets)
	}
}

func TestCoinHandler_Update_WithSetsPayloadPreservesMemberships(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	setRepo := repository.NewSetRepository(db)
	coin := models.Coin{Name: "Old Name", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	originalSet := models.CoinSet{UserID: 1, Name: "Original Set", SetType: models.CoinSetTypeOpen}
	incomingSet := models.CoinSet{UserID: 1, Name: "Incoming Set", SetType: models.CoinSetTypeOpen}
	if err := setRepo.Create(&originalSet); err != nil {
		t.Fatalf("failed to seed original set: %v", err)
	}
	if err := setRepo.Create(&incomingSet); err != nil {
		t.Fatalf("failed to seed incoming set: %v", err)
	}
	if err := setRepo.AddCoinToSet(coin.ID, originalSet.ID, 1, ""); err != nil {
		t.Fatalf("failed to seed membership: %v", err)
	}

	var originalMembership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, originalSet.ID).First(&originalMembership).Error; err != nil {
		t.Fatalf("failed to find seeded membership: %v", err)
	}
	if originalMembership.AddedAt.IsZero() {
		t.Fatal("seeded membership should have AddedAt")
	}

	updates := map[string]interface{}{
		"name":     "Updated Name",
		"category": "Modern",
		"material": "Silver",
		"sets": []map[string]interface{}{
			{
				"id":      incomingSet.ID,
				"userId":  1,
				"name":    incomingSet.Name,
				"setType": string(models.CoinSetTypeOpen),
			},
		},
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updatedCoin models.Coin
	if err := db.First(&updatedCoin, coin.ID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if updatedCoin.Name != "Updated Name" {
		t.Fatalf("expected updated coin name, got %q", updatedCoin.Name)
	}

	var memberships []models.CoinSetMembership
	if err := db.Where("coin_id = ?", coin.ID).Order("set_id ASC").Find(&memberships).Error; err != nil {
		t.Fatalf("failed to query memberships: %v", err)
	}
	if len(memberships) != 1 {
		t.Fatalf("expected update to preserve exactly 1 membership, got %d", len(memberships))
	}
	if memberships[0].SetID != originalSet.ID {
		t.Fatalf("expected original set membership to remain, got set ID %d", memberships[0].SetID)
	}
	if memberships[0].AddedAt.IsZero() {
		t.Fatal("membership AddedAt should remain populated after coin update")
	}
	if !memberships[0].AddedAt.Equal(originalMembership.AddedAt) {
		t.Fatalf("membership AddedAt changed from %v to %v", originalMembership.AddedAt, memberships[0].AddedAt)
	}
}

func TestCoinHandler_Update_PersistsExplicitFalseAndEmptyStringClears(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	coin := models.Coin{
		Name:               "Clearable Coin",
		Category:           models.CategoryModern,
		Material:           models.MaterialSilver,
		UserID:             1,
		Notes:              "clear me",
		ReferenceURL:       "https://example.test/ref",
		ReferenceText:      "clear reference text",
		PurchaseLocation:   "Old dealer",
		IsWishlist:         true,
		ObverseDescription: "clear obverse",
		ReverseDescription: "clear reverse",
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	updates := map[string]interface{}{
		"notes":              "",
		"referenceUrl":       "",
		"referenceText":      "",
		"purchaseLocation":   "",
		"obverseDescription": "",
		"reverseDescription": "",
		"isWishlist":         false,
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if found.Notes != "" || found.ReferenceURL != "" || found.ReferenceText != "" ||
		found.PurchaseLocation != "" ||
		found.ObverseDescription != "" || found.ReverseDescription != "" {
		t.Fatalf("expected explicit empty strings to persist, got notes=%q refURL=%q refText=%q purchaseLocation=%q obv=%q rev=%q",
			found.Notes, found.ReferenceURL, found.ReferenceText, found.PurchaseLocation, found.ObverseDescription, found.ReverseDescription)
	}
	if found.IsWishlist {
		t.Fatal("expected the explicit false boolean to persist, got wishlist=true")
	}
	if found.Category != models.CategoryModern || found.Material != models.MaterialSilver || found.Name != "Clearable Coin" {
		t.Fatalf("omitted sibling fields changed unexpectedly: name=%q category=%q material=%q", found.Name, found.Category, found.Material)
	}
}

func TestCoinHandler_Update_CurrentValueCreatesManualHistoryAndSnapshot(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	currentValue := 100.0
	coin := models.Coin{
		Name:         "Value Coin",
		Category:     models.CategoryOther,
		Material:     models.MaterialSilver,
		UserID:       1,
		CurrentValue: &currentValue,
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{"currentValue": 125.0})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if found.CurrentValue == nil || *found.CurrentValue != 125.0 {
		t.Fatalf("expected current value 125, got %v", found.CurrentValue)
	}
	if found.CurrentValueUpdatedAt == nil {
		t.Fatal("expected manual current-value edit to set CurrentValueUpdatedAt")
	}

	var history []models.CoinValueHistory
	if err := db.Where("coin_id = ?", coin.ID).Find(&history).Error; err != nil {
		t.Fatalf("failed to query value history: %v", err)
	}
	if len(history) != 1 || history[0].Value != 125.0 || history[0].Confidence != "manual" {
		t.Fatalf("expected one manual value-history row for 125, got %#v", history)
	}

	var journalCount int64
	if err := db.Model(&models.CoinJournal{}).Where("coin_id = ?", coin.ID).Count(&journalCount).Error; err != nil {
		t.Fatalf("failed to count journal entries: %v", err)
	}
	if journalCount != 1 {
		t.Fatalf("expected one journal entry, got %d", journalCount)
	}
}

func TestCoinHandler_Update_ClearsNullableScalarsWhenExplicitNullAndPreservesWhenOmitted(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	purchasePrice := 125.0
	currentValue := 175.0
	weight := 3.5
	diameter := 18.0
	purchaseDate := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	coin := models.Coin{
		Name:          "Nullable Coin",
		Category:      models.CategoryModern,
		Material:      models.MaterialSilver,
		UserID:        1,
		PurchasePrice: &purchasePrice,
		CurrentValue:  &currentValue,
		WeightGrams:   &weight,
		DiameterMm:    &diameter,
		PurchaseDate:  &purchaseDate,
	}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	updates := map[string]interface{}{
		"purchasePrice": nil,
		"currentValue":  nil,
		"purchaseDate":  nil,
		"weightGrams":   nil,
		"diameterMm":    nil,
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("updated coin not found: %v", err)
	}
	if found.PurchasePrice != nil || found.CurrentValue != nil || found.PurchaseDate != nil ||
		found.WeightGrams != nil || found.DiameterMm != nil {
		t.Fatalf("expected explicit nulls to clear nullable scalars, got purchase=%v current=%v purchaseDate=%v weight=%v diameter=%v",
			found.PurchasePrice, found.CurrentValue, found.PurchaseDate, found.WeightGrams, found.DiameterMm)
	}
	if found.Name != "Nullable Coin" || found.Category != models.CategoryModern || found.Material != models.MaterialSilver {
		t.Fatalf("omitted non-nullable fields changed unexpectedly: name=%q category=%q material=%q", found.Name, found.Category, found.Material)
	}

	preserved := models.Coin{
		Name:          "Preserved Nullable Coin",
		Category:      models.CategoryModern,
		Material:      models.MaterialSilver,
		UserID:        1,
		PurchasePrice: &purchasePrice,
		CurrentValue:  &currentValue,
		WeightGrams:   &weight,
		DiameterMm:    &diameter,
		PurchaseDate:  &purchaseDate,
	}
	if err := db.Create(&preserved).Error; err != nil {
		t.Fatalf("failed to seed preserved coin: %v", err)
	}
	nameOnlyBody, _ := json.Marshal(map[string]interface{}{"name": "Renamed Preserved Nullable Coin"})
	nameOnlyReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", preserved.ID), bytes.NewReader(nameOnlyBody))
	nameOnlyReq.Header.Set("Content-Type", "application/json")
	nameOnlyReq.Header.Set("Authorization", authHeader(1))
	nameOnlyRecorder := httptest.NewRecorder()

	router.ServeHTTP(nameOnlyRecorder, nameOnlyReq)

	if nameOnlyRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for name-only update, got %d: %s", nameOnlyRecorder.Code, nameOnlyRecorder.Body.String())
	}
	var preservedFound models.Coin
	if err := db.First(&preservedFound, preserved.ID).Error; err != nil {
		t.Fatalf("preserved coin not found: %v", err)
	}
	if preservedFound.PurchasePrice == nil || *preservedFound.PurchasePrice != purchasePrice ||
		preservedFound.CurrentValue == nil || *preservedFound.CurrentValue != currentValue ||
		preservedFound.PurchaseDate == nil || !preservedFound.PurchaseDate.Equal(purchaseDate) ||
		preservedFound.WeightGrams == nil || *preservedFound.WeightGrams != weight ||
		preservedFound.DiameterMm == nil || *preservedFound.DiameterMm != diameter {
		t.Fatalf("expected omitted nullable scalars to be preserved, got purchase=%v current=%v purchaseDate=%v weight=%v diameter=%v",
			preservedFound.PurchasePrice, preservedFound.CurrentValue, preservedFound.PurchaseDate, preservedFound.WeightGrams, preservedFound.DiameterMm)
	}
}

func TestCoinHandler_Update_ReplacesStructuredReferences(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	if err := db.Create(&models.CatalogRegistry{
		Catalog:        "RIC",
		DisplayName:    "Roman Imperial Coinage",
		Era:            models.EraAncient,
		VolumeRequired: true,
	}).Error; err != nil {
		t.Fatalf("failed to seed catalog registry: %v", err)
	}
	coin := models.Coin{Name: "Reference Coin", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "I", Number: "1"}).Error; err != nil {
		t.Fatalf("failed to seed reference: %v", err)
	}

	updates := map[string]interface{}{
		"name":     "Reference Coin",
		"category": "Modern",
		"material": "Silver",
		"references": []map[string]interface{}{
			{
				"catalog": " ric ",
				"volume":  " II ",
				"number":  " 12 ",
			},
		},
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var refs []models.CoinReference
	if err := db.Where("coin_id = ?", coin.ID).Find(&refs).Error; err != nil {
		t.Fatalf("failed to query references: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected exactly 1 replacement reference, got %d", len(refs))
	}
	if refs[0].Catalog != "RIC" || refs[0].Volume != "II" || refs[0].Number != "12" {
		t.Fatalf("expected normalized replacement reference RIC II 12, got %#v", refs[0])
	}
}

func TestCoinHandler_Update_CustomRegistryEraAccepted(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	if err := db.Create(&models.CatalogRegistry{
		Catalog:     "PROV",
		DisplayName: "Provincial References",
		Era:         models.Era("provincial"),
	}).Error; err != nil {
		t.Fatalf("failed to seed catalog registry: %v", err)
	}

	coin := models.Coin{Name: "Old Era", Category: models.CategoryModern, Material: models.MaterialBronze, UserID: 1, Era: models.EraAncient}
	db.Create(&coin)

	updates := map[string]interface{}{
		"name":     "Updated Era",
		"category": "Modern",
		"material": "Bronze",
		"era":      "provincial",
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for custom registry era, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Era != models.Era("provincial") {
		t.Fatalf("expected era provincial, got %q", found.Era)
	}
}

func TestCoinHandler_Update_AdminConfiguredEraAccepted(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	if err := repository.NewSettingsRepository(db).Upsert(services.SettingCoinEras, "Republican Rome\nRoman Empire\n500-480 BC"); err != nil {
		t.Fatalf("failed to seed coin eras setting: %v", err)
	}
	coin := models.Coin{Name: "Old Era", Category: models.CategoryModern, Material: models.MaterialBronze, UserID: 1, Era: models.EraAncient}
	db.Create(&coin)

	updates := map[string]interface{}{
		"name":     "Updated Era",
		"category": "Modern",
		"material": "Bronze",
		"era":      "Roman Empire",
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin-configured era, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Era != models.Era("Roman Empire") {
		t.Fatalf("expected era Roman Empire, got %q", found.Era)
	}
}

func TestCoinHandler_Update_PreservesUnchangedLegacyEra(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "updater")

	coin := models.Coin{Name: "Legacy Era", Category: models.CategoryModern, Material: models.MaterialBronze, UserID: 1, Era: models.Era("Imperial")}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatalf("failed to seed coin: %v", err)
	}

	updates := map[string]interface{}{
		"name":     "Updated Legacy Era",
		"category": "Modern",
		"material": "Bronze",
		"era":      "Imperial",
	}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for unchanged legacy era, got %d: %s", w.Code, w.Body.String())
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Name != "Updated Legacy Era" {
		t.Fatalf("expected updated name, got %q", found.Name)
	}
	if found.Era != models.Era("Imperial") {
		t.Fatalf("expected legacy era to be preserved, got %q", found.Era)
	}
}

func TestCoinHandler_Update_OtherUserCoin_NotFound(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	createTestUser(t, db, 2, "attacker")

	coin := models.Coin{Name: "Owner Coin", Category: models.CategoryOther, Material: models.MaterialGold, UserID: 1}
	db.Create(&coin)

	updates := map[string]interface{}{"name": "Stolen Name"}
	body, _ := json.Marshal(updates)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/coins/%d", coin.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(2))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for other user's coin, got %d", w.Code)
	}
}

// --- Delete ---

func TestCoinHandler_Delete_OwnCoin(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "deleter")

	coin := models.Coin{Name: "To Delete", Category: models.CategoryModern, Material: models.MaterialBronze, UserID: 1}
	db.Create(&coin)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/coins/%d", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify coin is gone
	var count int64
	db.Model(&models.Coin{}).Where("id = ?", coin.ID).Count(&count)
	if count != 0 {
		t.Error("expected coin to be deleted from DB")
	}
}

func TestCoinHandler_Delete_OtherUserCoin_NotFound(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "owner")
	createTestUser(t, db, 2, "attacker")

	coin := models.Coin{Name: "Protected Coin", Category: models.CategoryOther, Material: models.MaterialSilver, UserID: 1}
	db.Create(&coin)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/coins/%d", coin.ID), nil)
	req.Header.Set("Authorization", authHeader(2))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// Verify coin still exists
	var count int64
	db.Model(&models.Coin{}).Where("id = ?", coin.ID).Count(&count)
	if count != 1 {
		t.Error("expected coin to still exist in DB")
	}
}

// --- Unauthenticated ---

func TestCoinHandler_Unauthenticated(t *testing.T) {
	router, _ := setupCoinHandlerRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/coins", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", w.Code)
	}
}

// --- Investment breakdown ---

// TestCoinHandler_ActiveCollectionCountInvariant verifies that the canonical
// "active collection" count (owned AND NOT wishlist) is identical
// across three query paths:
//   - /coins?wishlist=false total
//   - /stats totalCoins
//   - internal collection_summary tool totalCoins
//
// This locks the predicate contract: no divergence is permitted.
func TestCoinHandler_ActiveCollectionCountInvariant(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "countuser")

	coinRepo := repository.NewCoinRepository(db)
	collectionSvc := services.NewCollectionToolsService(coinRepo, nil)

	coins := []models.Coin{
		{Name: "Active 1", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1, IsWishlist: false},
		{Name: "Active 2", Category: models.CategoryOther, Material: models.MaterialGold, UserID: 1, IsWishlist: false},
		{Name: "Active 3", Category: models.CategoryOther, Material: models.MaterialBronze, UserID: 1, IsWishlist: false},
		{Name: "Wishlist 1", Category: models.CategoryModern, Material: models.MaterialGold, UserID: 1, IsWishlist: true},
		{Name: "Wishlist 2", Category: models.CategoryOther, Material: models.MaterialSilver, UserID: 1, IsWishlist: true},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("failed to seed mixed coin collection: %v", err)
	}

	// Path 1: /coins?wishlist=false total
	req := httptest.NewRequest(http.MethodGet, "/api/coins?wishlist=false", nil)
	req.Header.Set("Authorization", authHeader(1))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 from /coins, got %d: %s", w.Code, w.Body.String())
	}
	var listResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to decode /coins response: %v", err)
	}
	listTotal := int64(listResp["total"].(float64))
	if listTotal != 3 {
		t.Errorf("/coins?wishlist=false: expected total=3, got %d", listTotal)
	}

	// Path 2: /stats totalCoins
	setupStatsRouter := func(t *testing.T, db *gorm.DB) *gin.Engine {
		gin.SetMode(gin.TestMode)
		coinRepo := repository.NewCoinRepository(db)
		catalogRegistryRepo := repository.NewCatalogRegistryRepository(db)
		coinReferenceRepo := repository.NewCoinReferenceRepository(db)
		coinReferenceSvc := services.NewCoinReferenceService(coinReferenceRepo, catalogRegistryRepo)
		coinSvc := services.NewCoinService(coinRepo).
			WithReferenceSupport(coinReferenceRepo, coinReferenceSvc).
			WithCatalogRegistrySupport(catalogRegistryRepo)
		handler := NewCoinHandler(coinRepo, coinSvc, services.NewLogger(100))
		r := gin.New()
		protected := r.Group("/api")
		protected.Use(coinTestAuthMiddleware())
		protected.GET("/stats", handler.Stats)
		return r
	}
	statsRouter := setupStatsRouter(t, db)
	req2 := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	req2.Header.Set("Authorization", authHeader(1))
	w2 := httptest.NewRecorder()
	statsRouter.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 from /stats, got %d: %s", w2.Code, w2.Body.String())
	}
	var statsResp map[string]interface{}
	if err := json.Unmarshal(w2.Body.Bytes(), &statsResp); err != nil {
		t.Fatalf("failed to decode /stats response: %v", err)
	}
	statsTotalCoins := int64(statsResp["totalCoins"].(float64))
	if statsTotalCoins != 3 {
		t.Errorf("/stats: expected totalCoins=3, got %d", statsTotalCoins)
	}

	// Path 3: collection_summary tool totalCoins
	summary, err := collectionSvc.CollectionSummary(1)
	if err != nil {
		t.Fatalf("CollectionSummary error: %v", err)
	}
	if summary.TotalCoins != 3 {
		t.Errorf("CollectionSummary: expected totalCoins=3, got %d", summary.TotalCoins)
	}

	// Invariant: all three paths must return the same count
	if listTotal != statsTotalCoins || statsTotalCoins != summary.TotalCoins {
		t.Errorf("INVARIANT VIOLATION: /coins total=%d, /stats totalCoins=%d, collection_summary totalCoins=%d (expected all=3)",
			listTotal, statsTotalCoins, summary.TotalCoins)
	}

	// Verify the wishlist counts are correct
	statsWishlist := int64(statsResp["totalWishlist"].(float64))
	if statsWishlist != 2 {
		t.Errorf("expected totalWishlist=2, got %d", statsWishlist)
	}
	if summary.TotalWishlist != 2 {
		t.Errorf("collection_summary: expected totalWishlist=2, got %d", summary.TotalWishlist)
	}
}

// TestCoinHandler_Create_WishlistWithReferencesStoresReferences is the
// primary regression test for the ADR 0013 rule
// (docs/adr/0013-wishlist-coins-may-hold-catalog-references.md): wishlist
// coins are no longer stripped of their References relation.
//
//	POST /api/coins with isWishlist:true and a non-empty references array
//	must return HTTP 201 AND persist those coin_reference rows for the new
//	coin, normalized exactly as a collection coin's references would be.
//	This covers the agent-style create path (the request carries a
//	candidateReferences-style payload directly via the request body's
//	"references" field, as opposed to the WishlistSearchAlertService
//	ConvertCandidate path, which is guarded separately per FR-049).
func TestCoinHandler_Create_WishlistWithReferencesStoresReferences(t *testing.T) {
	router, db := setupCoinHandlerRouter(t)
	createTestUser(t, db, 1, "wishlist-ref-test")

	// Seed the catalog so the reference is not rejected as unknown.
	if err := db.Create(&models.CatalogRegistry{
		Catalog:        "RIC",
		DisplayName:    "Roman Imperial Coinage",
		Era:            models.EraAncient,
		VolumeRequired: true,
	}).Error; err != nil {
		t.Fatalf("seed catalog: %v", err)
	}

	// This is the agent-style payload: a wishlist coin carrying a valid
	// catalog reference (the kind an agent would populate from candidateReferences).
	body := map[string]any{
		"name":       "Domitian Denarius (agent suggestion)",
		"category":   "Modern",
		"material":   "Silver",
		"isWishlist": true,
		"references": []map[string]any{
			{"catalog": "RIC", "volume": "II", "number": "162"},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/coins", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", authHeader(1))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created models.Coin
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !created.IsWishlist {
		t.Error("created coin should be a wishlist coin")
	}

	// ADR 0013: the reference must be persisted, normalized, for the wishlist coin.
	var refs []models.CoinReference
	if err := db.Where("coin_id = ?", created.ID).Find(&refs).Error; err != nil {
		t.Fatalf("query references: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected wishlist coin to persist 1 reference, got %d", len(refs))
	}
	if refs[0].Catalog != "RIC" || refs[0].Volume != "II" || refs[0].Number != "162" {
		t.Fatalf("expected normalized RIC II 162, got %#v", refs[0])
	}
}
