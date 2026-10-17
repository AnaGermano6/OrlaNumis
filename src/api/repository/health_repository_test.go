package repository

import (
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupHealthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.User{},
		&models.Coin{},
		&models.CoinImage{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func seedHealthTestUsers(t *testing.T, db *gorm.DB) (uint, uint) {
	t.Helper()
	user1 := models.User{Username: "collector1", Email: "c1@test.com"}
	user2 := models.User{Username: "collector2", Email: "c2@test.com"}
	db.Create(&user1)
	db.Create(&user2)
	return user1.ID, user2.ID
}

func seedHealthTestCoins(t *testing.T, db *gorm.DB, userID uint, count int, wishlist bool) []models.Coin {
	t.Helper()
	coins := make([]models.Coin, count)
	for i := 0; i < count; i++ {
		coins[i] = models.Coin{
			Name:       "Test Coin",
			Category:   models.CategoryModern,
			Material:   models.MaterialSilver,
			UserID:     userID,
			IsWishlist: wishlist,
		}
		if err := db.Create(&coins[i]).Error; err != nil {
			t.Fatalf("failed to create coin %d: %v", i, err)
		}
	}
	return coins
}

// --- T009: Repository tests for snapshot upsert and 30-day baseline lookup ---

// --- ListEligibleCoins tests ---

func TestHealthRepository_ListEligibleCoins_ActiveOnly(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 2, false) // active
	seedHealthTestCoins(t, db, userID, 1, true)  // wishlist

	rows, err := repo.ListEligibleCoins(userID)
	if err != nil {
		t.Fatalf("ListEligibleCoins failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 active coins, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoins_EmptyForNoActive(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 2, true) // all wishlist

	rows, err := repo.ListEligibleCoins(userID)
	if err != nil {
		t.Fatalf("ListEligibleCoins failed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 coins, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoins_UserScoped(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	user1, user2 := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, user1, 3, false)
	seedHealthTestCoins(t, db, user2, 2, false)

	rows, err := repo.ListEligibleCoins(user1)
	if err != nil {
		t.Fatalf("ListEligibleCoins failed: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 coins for user1, got %d", len(rows))
	}

	rows, err = repo.ListEligibleCoins(user2)
	if err != nil {
		t.Fatalf("ListEligibleCoins failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("expected 2 coins for user2, got %d", len(rows))
	}
}

// --- T027: Pagination and scope filtering tests ---

func TestHealthRepository_ListEligibleCoinsPaged_DefaultPagination(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 30, false)

	rows, total, err := repo.ListEligibleCoinsPaged(userID, 1, 25, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if total != 30 {
		t.Errorf("expected total=30, got %d", total)
	}
	if len(rows) != 25 {
		t.Errorf("expected 25 rows on page 1, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoinsPaged_SecondPage(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 30, false)

	rows, total, err := repo.ListEligibleCoinsPaged(userID, 2, 25, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if total != 30 {
		t.Errorf("expected total=30, got %d", total)
	}
	if len(rows) != 5 {
		t.Errorf("expected 5 rows on page 2, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoinsPaged_InvalidPageDefaultsToOne(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 10, false)

	rows, _, err := repo.ListEligibleCoinsPaged(userID, 0, 10, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if len(rows) != 10 {
		t.Errorf("expected 10 rows when page=0 defaults to 1, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoinsPaged_LimitBounds(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	seedHealthTestCoins(t, db, userID, 50, false)

	// limit < 1 defaults to 25
	rows, _, err := repo.ListEligibleCoinsPaged(userID, 1, 0, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if len(rows) != 25 {
		t.Errorf("expected limit=0 to default to 25, got %d", len(rows))
	}

	// limit > 100 defaults to 25
	rows, _, err = repo.ListEligibleCoinsPaged(userID, 1, 150, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if len(rows) != 25 {
		t.Errorf("expected limit=150 to default to 25, got %d", len(rows))
	}
}

func TestHealthRepository_ListEligibleCoinsPaged_EmptyResult(t *testing.T) {
	db := setupHealthTestDB(t)
	repo := NewHealthRepository(db)
	userID, _ := seedHealthTestUsers(t, db)

	rows, total, err := repo.ListEligibleCoinsPaged(userID, 1, 25, "all")
	if err != nil {
		t.Fatalf("ListEligibleCoinsPaged failed: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total=0, got %d", total)
	}
	if len(rows) != 0 {
		t.Errorf("expected 0 rows, got %d", len(rows))
	}
}

// --- ListUsersWithEligibleCoins tests ---
