package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupAvailServiceDB creates an in-memory SQLite DB with required tables.
func setupAvailServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Coin{},
		&models.CoinImage{},
		&models.AvailabilityRun{},
		&models.AvailabilityResult{},
		&models.AppSetting{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// TestCheckURL_200KeywordDetection verifies that the keyword detector works for exact patterns.
func TestCheckURL_200KeywordDetection(t *testing.T) {
	tests := []struct {
		name           string
		htmlBody       string
		expectedStatus string
	}{
		{
			name:           "Exact >sold< pattern",
			htmlBody:       `<div><button>sold</button></div>`,
			expectedStatus: "unavailable",
		},
		{
			name:           "Status: sold text",
			htmlBody:       `<p>Status: Sold</p>`,
			expectedStatus: "unavailable",
		},
		{
			name:           "Add to cart available",
			htmlBody:       `<button id="addToCart">Add to Cart</button>`,
			expectedStatus: "available",
		},
		{
			name:           "Generic purchase text remains ambiguous",
			htmlBody:       `<div>See purchase history for this coin.</div>`,
			expectedStatus: "unknown",
		},
		{
			name:           "Ambiguous page",
			htmlBody:       `<div>Contact us for details</div>`,
			expectedStatus: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.htmlBody))
			}))
			defer server.Close()

			svc := &AvailabilityService{logger: NewLogger(100)}
			result, err := svc.CheckURL(server.URL)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %s, got %s (reason: %s)", tt.expectedStatus, result.Status, result.Reason)
			}
		})
	}
}

// TestCheckURL_VCoinsSoldBannerBug is the regression test for the specific bug:
// VCoins sold pages have standalone "Sold" text that doesn't always match
// the >sold< keyword pattern and falls through to "unknown" status.
func TestCheckURL_VCoinsSoldBannerBug(t *testing.T) {
	tests := []struct {
		name           string
		htmlBody       string
		expectedStatus string
		description    string
	}{
		{
			name: "VCoins exact HTML structure",
			htmlBody: `<!DOCTYPE html>
<html><head><title>Coin</title></head>
<body>
<div class="item-status">Sold</div>
<div class="follow-buttons">
  <button>Follow Store</button>
  <button>Add to Watch List</button>
</div>
</body></html>`,
			expectedStatus: "unavailable",
			description:    "VCoins page with standalone 'Sold' in styled div",
		},
		{
			name: "Sold with surrounding whitespace",
			htmlBody: `<html><body>
<div class="status">
  Sold
</div>
</body></html>`,
			expectedStatus: "unavailable",
			description:    "Sold text with newlines/whitespace",
		},
		{
			name: "Sold in span",
			htmlBody: `<html><body>
<span class="badge">Sold</span>
</body></html>`,
			expectedStatus: "unavailable",
			description:    "Sold in span element",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.htmlBody))
			}))
			defer server.Close()

			svc := &AvailabilityService{logger: NewLogger(100)}
			result, err := svc.CheckURL(server.URL)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// REGRESSION: These pages should be classified as unavailable, not unknown
			if result.Status != tt.expectedStatus {
				t.Errorf("REGRESSION: %s - expected status %s, got %s (reason: %s)",
					tt.description, tt.expectedStatus, result.Status, result.Reason)
			}
		})
	}
}

// TestCheckURL_404ReturnsUnavailable verifies that 404 responses are immediately classified as unavailable.
func TestCheckURL_404ReturnsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	svc := &AvailabilityService{logger: NewLogger(100)}
	result, err := svc.CheckURL(server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Status != "unavailable" {
		t.Errorf("expected status unavailable for 404, got %s", result.Status)
	}
	if result.HttpStatus == nil || *result.HttpStatus != 404 {
		t.Errorf("expected HttpStatus 404, got %v", result.HttpStatus)
	}
}

// TestCheckURL_410ReturnsUnavailable verifies that 410 Gone responses are immediately classified as unavailable.
func TestCheckURL_410ReturnsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer server.Close()

	svc := &AvailabilityService{logger: NewLogger(100)}
	result, err := svc.CheckURL(server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Status != "unavailable" {
		t.Errorf("expected status unavailable for 410, got %s", result.Status)
	}
	if result.HttpStatus == nil || *result.HttpStatus != 410 {
		t.Errorf("expected HttpStatus 410, got %v", result.HttpStatus)
	}
}

// TestCheckURL_500ReturnsUnknown verifies that server errors are classified as unknown.
func TestCheckURL_500ReturnsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := &AvailabilityService{logger: NewLogger(100)}
	result, err := svc.CheckURL(server.URL)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Status != "unknown" {
		t.Errorf("expected status unknown for 500, got %s", result.Status)
	}
	if result.HttpStatus == nil || *result.HttpStatus != 500 {
		t.Errorf("expected HttpStatus 500, got %v", result.HttpStatus)
	}
}

// TestCheckWishlistForUser_ClassifiesSoldSignalsWithoutAgent is the regression
// test for the VCoins sold page bug. It verifies scheduled/manual summary counts
// reflect obvious sold and purchase signals even when agent escalation is absent.
func TestCheckWishlistForUser_ClassifiesSoldSignalsWithoutAgent(t *testing.T) {
	db := setupAvailServiceDB(t)

	// Create test user
	user := models.User{Username: "testuser", Email: "test@example.com"}
	db.Create(&user)

	// Mock servers for two VCoins sold pages
	soldServer1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>
			<div class="sold-banner">Sold</div>
			Valentinian I coin
		</body></html>`))
	}))
	defer soldServer1.Close()

	soldServer2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>
			<div class="status-box green">Sold</div>
			Constantine II coin
		</body></html>`))
	}))
	defer soldServer2.Close()

	// Mock server for an available coin
	availServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>Available for purchase</body></html>`))
	}))
	defer availServer.Close()

	// Create three wishlist coins with URLs
	coin1 := models.Coin{
		UserID:       user.ID,
		Name:         "Valentinian I GLORIA ROMANORVM",
		ReferenceURL: soldServer1.URL,
		IsWishlist:   true,
	}
	coin2 := models.Coin{
		UserID:       user.ID,
		Name:         "Constantine II",
		ReferenceURL: soldServer2.URL,
		IsWishlist:   true,
	}
	coin3 := models.Coin{
		UserID:       user.ID,
		Name:         "Julius Caesar Denarius",
		ReferenceURL: availServer.URL,
		IsWishlist:   true,
	}
	db.Create(&coin1)
	db.Create(&coin2)
	db.Create(&coin3)

	// Set up service with repositories
	coinRepo := repository.NewCoinRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	settingsSvc := NewSettingsService(settingsRepo)

	svc := &AvailabilityService{
		coinRepo:    coinRepo,
		availRepo:   availRepo,
		agentProxy:  nil,
		settingsSvc: settingsSvc,
		logger:      NewLogger(100),
	}

	// Run the check
	run, err := svc.CheckWishlistForUser(user.ID, "manual", nil, nil)

	if err != nil {
		t.Fatalf("CheckWishlistForUser failed: %v", err)
	}

	// Verify the run summary counts
	if run.CoinsChecked != 3 {
		t.Errorf("expected 3 coins checked, got %d", run.CoinsChecked)
	}

	// REGRESSION: This is the bug — unavailable count should be 2, not 0
	if run.Unavailable != 2 {
		t.Errorf("REGRESSION: expected 2 unavailable (sold) coins, got %d", run.Unavailable)
	}

	if run.Available != 1 {
		t.Errorf("expected 1 available coin, got %d", run.Available)
	}

	// REGRESSION: Unknown should be 0 after page-content checks, not 3.
	if run.Unknown != 0 {
		t.Errorf("REGRESSION: expected 0 unknown after page-content checks, got %d", run.Unknown)
	}

	// Verify individual results in the database
	var results []models.AvailabilityResult
	db.Where("run_id = ?", run.ID).Find(&results)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify both sold coins are marked unavailable
	soldCount := 0
	availCount := 0
	agentUsedCount := 0
	for _, r := range results {
		if r.AgentUsed {
			agentUsedCount++
		}
		if r.Status == "unavailable" {
			soldCount++
			// REGRESSION: Sold coins should be unavailable from page-content signals.
			if r.CoinName == "Valentinian I GLORIA ROMANORVM" || r.CoinName == "Constantine II" {
				if r.HttpStatus == nil || *r.HttpStatus != 200 {
					t.Errorf("coin %s: expected HTTP 200, got %v", r.CoinName, r.HttpStatus)
				}
				if r.Reason == "Requires AI analysis to determine availability" {
					t.Errorf("REGRESSION: coin %s: sold signal was not detected, reason still shows %s", r.CoinName, r.Reason)
				}
			}
		}
		if r.Status == "available" {
			availCount++
		}
	}

	if soldCount != 2 {
		t.Errorf("expected 2 results with status unavailable, got %d", soldCount)
	}
	if availCount != 1 {
		t.Errorf("expected 1 result with status available, got %d", availCount)
	}
	if agentUsedCount != 0 {
		t.Errorf("expected 0 results using agent analysis for clear page signals, got %d", agentUsedCount)
	}
}

// TestCheckWishlistForUser_SummaryCountsWithoutAgent verifies that summary counts
// are correct even when agent is not available or fails.
func TestCheckWishlistForUser_SummaryCountsWithoutAgent(t *testing.T) {
	db := setupAvailServiceDB(t)

	user := models.User{Username: "testuser2", Email: "test2@example.com"}
	db.Create(&user)

	// Mock servers: one 404, one 200 (no agent to resolve)
	notFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer notFoundServer.Close()

	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	coin1 := models.Coin{
		UserID:       user.ID,
		Name:         "Not Found Coin",
		ReferenceURL: notFoundServer.URL,
		IsWishlist:   true,
	}
	coin2 := models.Coin{
		UserID:       user.ID,
		Name:         "Ambiguous Coin",
		ReferenceURL: okServer.URL,
		IsWishlist:   true,
	}
	db.Create(&coin1)
	db.Create(&coin2)

	coinRepo := repository.NewCoinRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	settingsSvc := NewSettingsService(settingsRepo)

	// No agent proxy configured
	svc := &AvailabilityService{
		coinRepo:    coinRepo,
		availRepo:   availRepo,
		agentProxy:  nil,
		settingsSvc: settingsSvc,
		logger:      NewLogger(100),
	}

	run, err := svc.CheckWishlistForUser(user.ID, "manual", nil, nil)

	if err != nil {
		t.Fatalf("CheckWishlistForUser failed: %v", err)
	}

	if run.CoinsChecked != 2 {
		t.Errorf("expected 2 coins checked, got %d", run.CoinsChecked)
	}

	// Without agent, the 404 is unavailable, the 200 stays unknown
	if run.Unavailable != 1 {
		t.Errorf("expected 1 unavailable (404), got %d", run.Unavailable)
	}
	if run.Unknown != 1 {
		t.Errorf("expected 1 unknown (no agent), got %d", run.Unknown)
	}
	if run.Available != 0 {
		t.Errorf("expected 0 available, got %d", run.Available)
	}
}

// TestCheckWishlistForUser_ListingStatusUpdate verifies that coin listing status
// is updated after availability check.
func TestCheckWishlistForUser_ListingStatusUpdate(t *testing.T) {
	db := setupAvailServiceDB(t)

	user := models.User{Username: "testuser3", Email: "test3@example.com"}
	db.Create(&user)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	coin := models.Coin{
		UserID:       user.ID,
		Name:         "Test Coin",
		ReferenceURL: server.URL,
		IsWishlist:   true,
	}
	db.Create(&coin)

	coinRepo := repository.NewCoinRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	settingsSvc := NewSettingsService(settingsRepo)

	svc := &AvailabilityService{
		coinRepo:    coinRepo,
		availRepo:   availRepo,
		agentProxy:  nil,
		settingsSvc: settingsSvc,
		logger:      NewLogger(100),
	}

	_, err := svc.CheckWishlistForUser(user.ID, "manual", nil, nil)
	if err != nil {
		t.Fatalf("CheckWishlistForUser failed: %v", err)
	}

	// Reload coin and verify listing status was updated
	var updatedCoin models.Coin
	db.First(&updatedCoin, coin.ID)

	if updatedCoin.ListingStatus != "unavailable" {
		t.Errorf("expected listing status unavailable, got %s", updatedCoin.ListingStatus)
	}
	if updatedCoin.ListingCheckReason == "" {
		t.Error("expected listing check reason to be set")
	}
	if updatedCoin.ListingCheckedAt == nil {
		t.Error("expected listing checked at to be set")
	}
}

// TestCheckWishlistForUser_RateLimiting verifies that rate limiting is applied
// between URL checks (smoke test — checks that delay exists without mocking time).
func TestCheckWishlistForUser_RateLimiting(t *testing.T) {
	db := setupAvailServiceDB(t)

	user := models.User{Username: "testuser4", Email: "test4@example.com"}
	db.Create(&user)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create multiple coins to trigger rate limiting
	for i := 0; i < 3; i++ {
		coin := models.Coin{
			UserID:       user.ID,
			Name:         fmt.Sprintf("Coin %d", i),
			ReferenceURL: server.URL,
			IsWishlist:   true,
		}
		db.Create(&coin)
	}

	coinRepo := repository.NewCoinRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	settingsSvc := NewSettingsService(settingsRepo)

	svc := &AvailabilityService{
		coinRepo:    coinRepo,
		availRepo:   availRepo,
		agentProxy:  nil,
		settingsSvc: settingsSvc,
		logger:      NewLogger(100),
	}

	start := time.Now()
	_, err := svc.CheckWishlistForUser(user.ID, "manual", nil, nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("CheckWishlistForUser failed: %v", err)
	}

	// With 3 coins and 750ms delay between requests, expect at least 1.5s total
	// (2 delays between 3 requests). Allow some tolerance for test execution.
	minExpected := 1400 * time.Millisecond
	if elapsed < minExpected {
		t.Errorf("expected at least %v for rate limiting, took %v", minExpected, elapsed)
	}
}

// --- Feature 353: cycle-aware service tests (T011, T017, T024-T027, T040, T049) ---

// setupAvailServiceFullDB creates an in-memory DB with the full Feature 353 schema (cycles,
// notifications, pushover settings) — used by tests that need cross-repo aggregation, admin
// fan-out, and notification assertions.
func setupAvailServiceFullDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql db: %v", err)
	}
	// Pushover sends fire on a background goroutine; pin the pool to one connection so it
	// can't land on a second, unmigrated ":memory:" instance.
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Coin{},
		&models.CoinImage{},
		&models.AvailabilityCycle{},
		&models.AvailabilityRun{},
		&models.AvailabilityResult{},
		&models.AppSetting{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

// setupFullAvailabilityService builds an AvailabilityService wired exactly like main.go
// (availRepo + availCycleRepo cross-linked).
func setupFullAvailabilityService(t *testing.T, db *gorm.DB) (*AvailabilityService, *repository.AvailabilityCycleRepository, *repository.AvailabilityRepository) {
	t.Helper()
	coinRepo := repository.NewCoinRepository(db)
	availRepo := repository.NewAvailabilityRepository(db)
	availCycleRepo := repository.NewAvailabilityCycleRepository(db)
	availRepo.WithCycleRepo(availCycleRepo)
	settingsSvc := NewSettingsService(repository.NewSettingsRepository(db))
	logger := NewLogger(100)
	svc := NewAvailabilityService(coinRepo, availRepo, nil, settingsSvc, logger).WithCycleRepo(availCycleRepo)
	return svc, availCycleRepo, availRepo
}

// T011: CheckWishlistForUser always creates exactly one child run whose UserID is the
// invoking owner's ID — never 0 (US1 AC1 / FR-002).
func TestCheckWishlistForUser_CreatesChildRunWithNonZeroOwnerUserID(t *testing.T) {
	db := setupAvailServiceFullDB(t)
	svc, _, _ := setupFullAvailabilityService(t, db)

	user := models.User{Username: "owner-t011", Email: "owner-t011@test.com"}
	db.Create(&user)

	run, err := svc.CheckWishlistForUser(user.ID, models.AvailabilityRunTriggerOwner, &user.ID, nil)
	if err != nil {
		t.Fatalf("CheckWishlistForUser: %v", err)
	}
	if run.UserID == 0 {
		t.Fatal("expected child run UserID > 0, got 0")
	}
	if run.UserID != user.ID {
		t.Fatalf("expected child run UserID %d, got %d", user.ID, run.UserID)
	}

	var count int64
	db.Model(&models.AvailabilityRun{}).Where("user_id = ?", user.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected exactly 1 child run for owner, got %d", count)
	}
}

// T017: RunAdminCycle fans out exactly one child run per distinct wishlist owner (never a
// child with UserID 0) and finalizes the parent cycle to "completed" once every child
// succeeds (FR-006/FR-008).
func TestRunAdminCycle_FansOutOneChildPerOwner_AllSucceed(t *testing.T) {
	db := setupAvailServiceFullDB(t)
	svc, availCycleRepo, _ := setupFullAvailabilityService(t, db)

	available := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><button>Add to Cart</button></body></html>`))
	}))
	defer available.Close()

	owner1 := models.User{Username: "admin-cycle-owner1", Email: "aco1@test.com"}
	owner2 := models.User{Username: "admin-cycle-owner2", Email: "aco2@test.com"}
	db.Create(&owner1)
	db.Create(&owner2)
	db.Create(&models.Coin{UserID: owner1.ID, Name: "Coin 1", ReferenceURL: available.URL, IsWishlist: true})
	db.Create(&models.Coin{UserID: owner2.ID, Name: "Coin 2", ReferenceURL: available.URL, IsWishlist: true})

	adminID := uint(999)
	cycle := &models.AvailabilityCycle{
		TriggerType:   models.AvailabilityRunTriggerAdmin,
		TriggerUserID: &adminID,
		Status:        models.AvailabilityCycleStatusRunning,
		StartedAt:     time.Now(),
	}
	if err := db.Create(cycle).Error; err != nil {
		t.Fatalf("seed cycle: %v", err)
	}

	if err := svc.RunAdminCycle(cycle); err != nil {
		t.Fatalf("RunAdminCycle: %v", err)
	}

	var children []models.AvailabilityRun
	db.Where("cycle_id = ?", cycle.ID).Find(&children)
	if len(children) != 2 {
		t.Fatalf("expected 2 child runs (one per owner), got %d", len(children))
	}
	for _, c := range children {
		if c.UserID == 0 {
			t.Fatal("expected every child run to have UserID > 0")
		}
		if c.Status != models.AvailabilityRunStatusCompleted {
			t.Fatalf("expected child completed, got %q", c.Status)
		}
	}

	reloaded, err := availCycleRepo.GetCycleWithChildren(cycle.ID)
	if err != nil {
		t.Fatalf("GetCycleWithChildren: %v", err)
	}
	if reloaded.Status != models.AvailabilityCycleStatusCompleted {
		t.Fatalf("expected parent cycle completed, got %q", reloaded.Status)
	}
}
