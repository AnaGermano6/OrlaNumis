package repository

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var coinRepositoryTestDBCounter uint64

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:coin_repository_%d_%d?mode=memory&cache=shared&_pragma=busy_timeout(5000)", time.Now().UnixNano(), atomic.AddUint64(&coinRepositoryTestDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{}, &models.StorageLocation{}, &models.Coin{}, &models.CoinImage{}, &models.CoinReference{},
		&models.CoinJournal{},
		&models.CoinValueHistory{}, &models.CoinComment{},
		&models.AvailabilityResult{},
		&models.ValuationRun{}, &models.ValuationResult{},
		&models.Tag{}, &models.CoinTag{},
		&models.CoinSet{}, &models.CoinSetMembership{}, &models.CoinSetTarget{}, &models.SmartCriteriaTemplate{},
		&models.QuickCaptureDraft{}, &models.QuickCaptureDraftImage{}, &models.QuickCaptureDraftReference{}, &models.DraftLifecycleEvent{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func addStorageSlotUniqueIndex(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`CREATE UNIQUE INDEX idx_coins_storage_location_slot_unique
		ON coins(storage_location_id, storage_slot)
		WHERE storage_slot IS NOT NULL`).Error; err != nil {
		t.Fatalf("create storage slot unique index: %v", err)
	}
}

func TestCoinRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	coin := &models.Coin{
		Name:     "Test Denarius",
		Category: models.CategoryRoman,
		Material: models.MaterialSilver,
		UserID:   1,
	}

	if err := repo.Create(coin); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if coin.ID == 0 {
		t.Fatal("expected coin ID to be set")
	}

	found, err := repo.FindByID(coin.ID, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "Test Denarius" {
		t.Errorf("expected name 'Test Denarius', got %q", found.Name)
	}
	if found.Category != models.CategoryRoman {
		t.Errorf("expected category Roman, got %q", found.Category)
	}
}

func TestCoinRepository_FindByID_WrongUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	coin := &models.Coin{Name: "Private Coin", Category: models.CategoryGreek, UserID: 1}
	repo.Create(coin)

	_, err := repo.FindByID(coin.ID, 999)
	if err == nil {
		t.Fatal("expected error when fetching coin with wrong user ID")
	}
}

func TestCoinRepository_FindWishlistByReferenceURLIsOwnerScoped(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)
	sourceURL := "https://dealer.example/coin/123"
	for _, coin := range []*models.Coin{
		{Name: "Owner one", Category: models.CategoryRoman, ReferenceURL: sourceURL, IsWishlist: true, UserID: 1},
		{Name: "Owner two", Category: models.CategoryRoman, ReferenceURL: sourceURL, IsWishlist: true, UserID: 2},
		{Name: "Collection coin", Category: models.CategoryRoman, ReferenceURL: sourceURL, UserID: 3},
	} {
		if err := repo.Create(coin); err != nil {
			t.Fatal(err)
		}
	}

	found, err := repo.FindWishlistByReferenceURL(2, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	if found.UserID != 2 || !found.IsWishlist {
		t.Fatalf("owner-scoped result=%#v", found)
	}
	if _, err := repo.FindWishlistByReferenceURL(3, sourceURL); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("collection coin matched wishlist lookup: %v", err)
	}
}

func TestCoinRepository_WithTx(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	err := db.Transaction(func(tx *gorm.DB) error {
		txRepo := repo.WithTx(tx)
		coin := &models.Coin{
			Name:     "TX Coin",
			Category: models.CategoryRoman,
			UserID:   1,
		}
		if err := txRepo.Create(coin); err != nil {
			return err
		}

		// Should be visible within the transaction
		found, err := txRepo.FindByID(coin.ID, 1)
		if err != nil {
			return err
		}
		if found.Name != "TX Coin" {
			t.Errorf("expected 'TX Coin', got %q", found.Name)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}

	// Should also be visible after commit
	var count int64
	db.Model(&models.Coin{}).Where("name = ?", "TX Coin").Count(&count)
	if count != 1 {
		t.Error("expected coin to persist after transaction commit")
	}
}

func TestCoinRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	coin := &models.Coin{Name: "To Delete", Category: models.CategoryRoman, UserID: 1}
	repo.Create(coin)

	// Add related data
	db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "img.jpg"})
	db.Create(&models.CoinJournal{CoinID: coin.ID, UserID: 1, Entry: "test"})

	rows, err := repo.Delete(coin.ID, 1)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	var coinCount, imgCount, journalCount int64
	db.Model(&models.Coin{}).Where("id = ?", coin.ID).Count(&coinCount)
	db.Model(&models.CoinImage{}).Where("coin_id = ?", coin.ID).Count(&imgCount)
	db.Model(&models.CoinJournal{}).Where("coin_id = ?", coin.ID).Count(&journalCount)

	if coinCount != 0 {
		t.Error("coin should be deleted")
	}
	if imgCount != 0 {
		t.Error("coin image should be deleted")
	}
	if journalCount != 0 {
		t.Error("journal entry should be deleted")
	}
}

// SQLite enforces the foreign keys of images and catalogue references (the app
// turns them on), so a coin that has them must still delete, alone or in bulk.
func TestCoinRepository_DeleteCoinWithReferencesAndImagesUnderForeignKeys(t *testing.T) {
	db := setupTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// One connection, so the pragma applies to every statement below.
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		t.Fatal(err)
	}
	repo := NewCoinRepository(db)
	db.Create(&models.User{Username: "fk-owner"})
	newCoin := func(name string) *models.Coin {
		coin := &models.Coin{Name: name, Category: models.CategoryModern, UserID: 1}
		if err := repo.Create(coin); err != nil {
			t.Fatalf("create: %v", err)
		}
		db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "KM", Number: "84"})
		db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: name + ".jpg"})
		return coin
	}

	single := newCoin("single")
	if rows, err := repo.Delete(single.ID, 1); err != nil || rows != 1 {
		t.Fatalf("Delete: rows=%d err=%v", rows, err)
	}
	// Another user's coin is left alone, references included.
	other := newCoin("other")
	if rows, err := repo.Delete(other.ID, 2); err != nil || rows != 0 {
		t.Fatalf("Delete by non-owner: rows=%d err=%v", rows, err)
	}
	first, second := newCoin("first"), newCoin("second")
	if rows, err := repo.BulkDelete([]uint{first.ID, second.ID}, 1); err != nil || rows != 2 {
		t.Fatalf("BulkDelete: rows=%d err=%v", rows, err)
	}

	var coins, refs, images int64
	db.Model(&models.Coin{}).Count(&coins)
	db.Model(&models.CoinReference{}).Count(&refs)
	db.Model(&models.CoinImage{}).Count(&images)
	if coins != 1 || refs != 1 || images != 1 {
		t.Fatalf("left coins=%d references=%d images=%d, want only the other user's coin", coins, refs, images)
	}
}

func TestCoinRepository_CoinExists(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	coin := &models.Coin{Name: "Exists Test", Category: models.CategoryRoman, UserID: 1}
	repo.Create(coin)

	exists, err := repo.CoinExists(coin.ID, 1)
	if err != nil {
		t.Fatalf("CoinExists failed: %v", err)
	}
	if !exists {
		t.Error("expected coin to exist")
	}

	exists, err = repo.CoinExists(coin.ID, 999)
	if err != nil {
		t.Fatalf("CoinExists failed: %v", err)
	}
	if exists {
		t.Error("expected coin to not exist for wrong user")
	}
}

func TestCoinRepository_Scopes_OwnedBy(t *testing.T) {
	db := setupTestDB(t)

	// Create coins for two users
	db.Create(&models.Coin{Name: "User1 Coin A", Category: models.CategoryRoman, UserID: 1})
	db.Create(&models.Coin{Name: "User1 Coin B", Category: models.CategoryGreek, UserID: 1})
	db.Create(&models.Coin{Name: "User2 Coin", Category: models.CategoryRoman, UserID: 2})

	var coins []models.Coin
	db.Scopes(OwnedBy(1)).Find(&coins)
	if len(coins) != 2 {
		t.Errorf("expected 2 coins for user 1, got %d", len(coins))
	}

	db.Scopes(OwnedBy(2)).Find(&coins)
	if len(coins) != 1 {
		t.Errorf("expected 1 coin for user 2, got %d", len(coins))
	}
}

func TestCoinRepository_Scopes_ActiveCollection(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.Coin{Name: "Active", Category: models.CategoryRoman, UserID: 1, IsWishlist: false, IsSold: false})
	db.Create(&models.Coin{Name: "Wishlist", Category: models.CategoryRoman, UserID: 1, IsWishlist: true, IsSold: false})
	db.Create(&models.Coin{Name: "Sold", Category: models.CategoryRoman, UserID: 1, IsWishlist: false, IsSold: true})

	var coins []models.Coin
	db.Scopes(ActiveCollection(1)).Find(&coins)
	if len(coins) != 1 {
		t.Fatalf("expected 1 active coin, got %d", len(coins))
	}
	if coins[0].Name != "Active" {
		t.Errorf("expected 'Active', got %q", coins[0].Name)
	}
}

func TestCoinRepository_ListMatchedImperialFigures(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	figureID := uint(42)
	otherFigureID := uint(7)
	matched := models.Coin{Name: "Matched", Category: models.CategoryRoman, UserID: 1, RomanImperialFigureID: &figureID}
	if err := db.Create(&matched).Error; err != nil {
		t.Fatalf("seed matched coin: %v", err)
	}
	db.Create(&models.CoinImage{CoinID: matched.ID, FilePath: "/uploads/matched.jpg"})
	db.Create(&models.Coin{Name: "Unmatched", Category: models.CategoryRoman, UserID: 1})
	db.Create(&models.Coin{Name: "Greek Matched", Category: models.CategoryGreek, UserID: 1, RomanImperialFigureID: &figureID})
	db.Create(&models.Coin{Name: "Matched Wishlist", Category: models.CategoryRoman, UserID: 1, RomanImperialFigureID: &otherFigureID, IsWishlist: true})
	db.Create(&models.Coin{Name: "Matched Sold", Category: models.CategoryRoman, UserID: 1, RomanImperialFigureID: &otherFigureID, IsSold: true})
	db.Create(&models.Coin{Name: "Other User Matched", Category: models.CategoryRoman, UserID: 2, RomanImperialFigureID: &figureID})

	coins, err := repo.ListMatchedImperialFigures(1)
	if err != nil {
		t.Fatalf("ListMatchedImperialFigures failed: %v", err)
	}
	if len(coins) != 1 {
		t.Fatalf("expected 1 matched active coin, got %d: %+v", len(coins), coins)
	}
	if coins[0].Name != "Matched" {
		t.Errorf("expected 'Matched', got %q", coins[0].Name)
	}
	if len(coins[0].Images) != 1 || coins[0].Images[0].FilePath != "/uploads/matched.jpg" {
		t.Errorf("expected preloaded image, got %+v", coins[0].Images)
	}
}

func TestCoinRepository_QuickCaptureDraftsExcludedAndPromotedCoinAppearsOnce(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)
	userID := uint(1)

	active := models.Coin{Name: "Active normal coin", Category: models.CategoryRoman, Material: models.MaterialSilver, UserID: userID, IsWishlist: false, IsSold: false}
	wishlist := models.Coin{Name: "Wishlist coin", Category: models.CategoryGreek, Material: models.MaterialGold, UserID: userID, IsWishlist: true, IsSold: false}
	sold := models.Coin{Name: "Sold coin", Category: models.CategoryRoman, Material: models.MaterialBronze, UserID: userID, IsWishlist: false, IsSold: true}
	promoted := models.Coin{Name: "Promoted Quick Capture coin", Category: models.CategoryRoman, Material: models.MaterialSilver, UserID: userID, IsWishlist: false, IsSold: false}
	for _, coin := range []*models.Coin{&active, &wishlist, &sold, &promoted} {
		if err := db.Create(coin).Error; err != nil {
			t.Fatalf("seed coin: %v", err)
		}
	}
	promotedCoinID := promoted.ID
	drafts := []models.QuickCaptureDraft{
		{UserID: userID, WorkingTitle: "Sparse active draft", Status: models.QuickCaptureDraftStatusActive},
		{UserID: userID, WorkingTitle: "Promoted draft", Status: models.QuickCaptureDraftStatusPromoted, PromotedCoinID: &promotedCoinID},
	}
	for i := range drafts {
		if err := db.Create(&drafts[i]).Error; err != nil {
			t.Fatalf("seed draft: %v", err)
		}
	}

	activeFilter := false
	coins, total, err := repo.List(userID, CoinListFilters{Wishlist: &activeFilter, Sold: &activeFilter, Page: 1, Limit: 50, SortField: "name", SortOrder: "asc"})
	if err != nil {
		t.Fatalf("list active coins: %v", err)
	}
	if total != 2 || len(coins) != 2 {
		t.Fatalf("expected only active normal/promoted coins, total=%d len=%d coins=%+v", total, len(coins), coins)
	}
	names := map[string]int{}
	for _, coin := range coins {
		names[coin.Name]++
	}
	if names["Sparse active draft"] != 0 || names["Promoted draft"] != 0 {
		t.Fatalf("quick-capture draft rows leaked into normal coin list: %v", names)
	}
	if names["Promoted Quick Capture coin"] != 1 {
		t.Fatalf("promoted normal coin should appear exactly once, got names=%v", names)
	}
}

func TestCoinRepository_Scopes_PublicCoins(t *testing.T) {
	db := setupTestDB(t)

	db.Create(&models.Coin{Name: "Public", Category: models.CategoryRoman, UserID: 1, IsPrivate: false})
	db.Create(&models.Coin{Name: "Private", Category: models.CategoryRoman, UserID: 1, IsPrivate: true})
	db.Create(&models.Coin{Name: "Wishlist", Category: models.CategoryRoman, UserID: 1, IsWishlist: true})

	var coins []models.Coin
	db.Scopes(PublicCoins(1)).Find(&coins)
	if len(coins) != 1 {
		t.Fatalf("expected 1 public coin, got %d", len(coins))
	}
	if coins[0].Name != "Public" {
		t.Errorf("expected 'Public', got %q", coins[0].Name)
	}
}

func ptrFloat(v float64) *float64 { return &v }

func ptrTime(v time.Time) *time.Time { return &v }

func assertFloatNear(t *testing.T, got float64, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("expected %.4f, got %.4f", want, got)
	}
}

func TestCoinRepository_List_RandomSort(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)

	// Insert 10 coins for the same user.
	for i := 1; i <= 10; i++ {
		if err := repo.Create(&models.Coin{
			Name:     "Coin",
			Category: models.CategoryRoman,
			UserID:   1,
		}); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	listIDs := func(filters CoinListFilters) []uint {
		coins, _, err := repo.List(1, filters)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		ids := make([]uint, len(coins))
		for i, c := range coins {
			ids[i] = c.ID
		}
		return ids
	}

	// Baseline: created_at desc (newest first) — should be 10, 9, 8, ... 1.
	desc := listIDs(CoinListFilters{SortField: "created_at", SortOrder: "desc", Page: 1, Limit: 50})

	// Same seed twice must yield the same order (deterministic).
	seed := 12345
	a := listIDs(CoinListFilters{SortField: "random", Seed: &seed, Page: 1, Limit: 50})
	b := listIDs(CoinListFilters{SortField: "random", Seed: &seed, Page: 1, Limit: 50})
	if len(a) != 10 || len(b) != 10 {
		t.Fatalf("expected 10 coins, got %d and %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("random sort not deterministic for same seed at index %d: %v vs %v", i, a, b)
		}
	}

	// Random ordering must NOT equal the natural insertion / created_at order.
	differs := false
	for i := range a {
		if a[i] != desc[i] {
			differs = true
			break
		}
	}
	if !differs {
		t.Fatalf("random sort produced the same order as created_at desc; the seed has no effect: %v", a)
	}

	// A different seed should produce a different ordering than the first seed.
	seed2 := 99999
	c := listIDs(CoinListFilters{SortField: "random", Seed: &seed2, Page: 1, Limit: 50})
	differs = false
	for i := range a {
		if a[i] != c[i] {
			differs = true
			break
		}
	}
	if !differs {
		t.Fatalf("different seeds produced identical ordering; seed has no effect: %v", a)
	}
}

// TestCoinRepository_Update_PreservesSets tests regression where updating a coin
// with Sets relation must NOT corrupt existing memberships by recreating them
// without AddedAt (violating NOT NULL constraint).
func TestCoinRepository_Update_PreservesSets(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)
	setRepo := NewSetRepository(db)

	// Create a coin
	coin := &models.Coin{
		Name:     "Test Aureus",
		Category: models.CategoryRoman,
		Material: models.MaterialGold,
		UserID:   1,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}

	// Create a set
	set := &models.CoinSet{
		UserID:  1,
		Name:    "Roman Gold",
		SetType: models.CoinSetTypeOpen,
	}
	if err := setRepo.Create(set); err != nil {
		t.Fatalf("Create set failed: %v", err)
	}

	// Add coin to set via SetRepository (correct path with AddedAt)
	if err := setRepo.AddCoinToSet(coin.ID, set.ID, 1, ""); err != nil {
		t.Fatalf("AddCoinToSet failed: %v", err)
	}

	// Verify membership was created with AddedAt
	var membership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, set.ID).First(&membership).Error; err != nil {
		t.Fatalf("membership not found: %v", err)
	}
	if membership.AddedAt.IsZero() {
		t.Fatal("membership.AddedAt is zero; should be set by AddCoinToSet")
	}
	originalAddedAt := membership.AddedAt

	// Now update the coin via CoinRepository
	updates := &models.Coin{
		Name: "Updated Aureus",
	}
	if err := coinRepo.Update(coin, updates); err != nil {
		t.Fatalf("Update coin failed: %v", err)
	}

	// Verify coin was updated
	if coin.Name != "Updated Aureus" {
		t.Errorf("expected updated name 'Updated Aureus', got %q", coin.Name)
	}

	// Critical: verify membership still exists with the same AddedAt
	var updatedMembership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, set.ID).First(&updatedMembership).Error; err != nil {
		t.Fatalf("membership disappeared after coin update: %v", err)
	}
	if updatedMembership.AddedAt.IsZero() {
		t.Fatal("membership.AddedAt is zero after update; Omit('Sets') failed")
	}
	if !updatedMembership.AddedAt.Equal(originalAddedAt) {
		t.Errorf("membership.AddedAt changed from %v to %v; should be preserved", originalAddedAt, updatedMembership.AddedAt)
	}
}

// TestCoinRepository_Update_WithSetsField tests that passing coin.Sets in an update
// does NOT modify memberships due to Omit("Sets") in Update method.
func TestCoinRepository_Update_WithSetsField(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)
	setRepo := NewSetRepository(db)

	// Create coin and two sets
	coin := &models.Coin{
		Name:     "Test Solidus",
		Category: models.CategoryByzantine,
		UserID:   1,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}

	set1 := &models.CoinSet{UserID: 1, Name: "Byzantine Core", SetType: models.CoinSetTypeOpen}
	set2 := &models.CoinSet{UserID: 1, Name: "High Grade", SetType: models.CoinSetTypeOpen}
	if err := setRepo.Create(set1); err != nil {
		t.Fatalf("Create set1 failed: %v", err)
	}
	if err := setRepo.Create(set2); err != nil {
		t.Fatalf("Create set2 failed: %v", err)
	}

	// Add coin to set1 only
	if err := setRepo.AddCoinToSet(coin.ID, set1.ID, 1, "initial"); err != nil {
		t.Fatalf("AddCoinToSet failed: %v", err)
	}

	var count int64
	db.Model(&models.CoinSetMembership{}).Where("coin_id = ?", coin.ID).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 membership before update, got %d", count)
	}

	// Attempt to update coin with coin.Sets = [set2] (should be ignored by Omit)
	updates := &models.Coin{
		Name: "Updated Solidus",
		Sets: []models.CoinSet{*set2},
	}
	if err := coinRepo.Update(coin, updates); err != nil {
		t.Fatalf("Update coin failed: %v", err)
	}

	// Verify name was updated but Sets relationship was NOT replaced
	if coin.Name != "Updated Solidus" {
		t.Errorf("expected name 'Updated Solidus', got %q", coin.Name)
	}

	// Should still have exactly 1 membership (set1), not replaced by set2
	db.Model(&models.CoinSetMembership{}).Where("coin_id = ?", coin.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 membership after update (should be ignored), got %d", count)
	}

	var membership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, set1.ID).First(&membership).Error; err != nil {
		t.Fatal("original membership (set1) disappeared; Omit('Sets') failed")
	}
	if membership.AddedAt.IsZero() {
		t.Fatal("membership.AddedAt is zero; should be preserved")
	}

	// Verify set2 was NOT added
	var set2Count int64
	db.Model(&models.CoinSetMembership{}).Where("coin_id = ? AND set_id = ?", coin.ID, set2.ID).Count(&set2Count)
	if set2Count != 0 {
		t.Error("set2 was added despite Omit('Sets'); update should not touch Sets")
	}
}

func TestCoinRepository_Update_PreservesLoadedAssociations(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)
	setRepo := NewSetRepository(db)
	tagRepo := NewTagRepository(db)

	location := models.StorageLocation{UserID: 1, Name: "Cabinet A"}
	if err := db.Create(&location).Error; err != nil {
		t.Fatalf("Create storage location failed: %v", err)
	}
	coin := &models.Coin{
		Name:              "Associated Coin",
		Category:          models.CategoryRoman,
		Material:          models.MaterialSilver,
		UserID:            1,
		StorageLocationID: &location.ID,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}
	if err := db.Create(&models.CoinImage{CoinID: coin.ID, FilePath: "coins/original.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
		t.Fatalf("Create image failed: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "I", Number: "1"}).Error; err != nil {
		t.Fatalf("Create reference failed: %v", err)
	}
	tag := &models.Tag{UserID: 1, Name: "Favorites", Color: "#c9a84c"}
	if err := tagRepo.Create(tag); err != nil {
		t.Fatalf("Create tag failed: %v", err)
	}
	if err := tagRepo.AttachToCoin(coin.ID, tag.ID, 1); err != nil {
		t.Fatalf("Attach tag failed: %v", err)
	}
	set := &models.CoinSet{UserID: 1, Name: "Roman Core", SetType: models.CoinSetTypeOpen}
	if err := setRepo.Create(set); err != nil {
		t.Fatalf("Create set failed: %v", err)
	}
	if err := setRepo.AddCoinToSet(coin.ID, set.ID, 1, "original"); err != nil {
		t.Fatalf("AddCoinToSet failed: %v", err)
	}

	loaded, err := coinRepo.FindByID(coin.ID, 1)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if len(loaded.Images) != 1 || len(loaded.References) != 1 || len(loaded.Tags) != 1 || len(loaded.Sets) != 1 || loaded.StorageLocation == nil {
		t.Fatalf("expected loaded associations before update, got images=%d refs=%d tags=%d sets=%d storage=%v", len(loaded.Images), len(loaded.References), len(loaded.Tags), len(loaded.Sets), loaded.StorageLocation)
	}

	incomingTag := models.Tag{ID: tag.ID + 100, UserID: 1, Name: "Incoming"}
	incomingSet := models.CoinSet{ID: set.ID + 100, UserID: 1, Name: "Incoming Set", SetType: models.CoinSetTypeOpen}
	updates := &models.Coin{
		Name: "Updated Associated Coin",
		Images: []models.CoinImage{
			{ID: loaded.Images[0].ID + 100, CoinID: loaded.ID, FilePath: "coins/incoming.jpg", ImageType: models.ImageTypeReverse},
		},
		References: []models.CoinReference{
			{ID: loaded.References[0].ID + 100, CoinID: loaded.ID, Catalog: "RIC", Volume: "II", Number: "2"},
		},
		Tags:            []models.Tag{incomingTag},
		Sets:            []models.CoinSet{incomingSet},
		StorageLocation: &models.StorageLocation{ID: location.ID + 100, UserID: 1, Name: "Incoming Location"},
	}
	if err := coinRepo.Update(loaded, updates); err != nil {
		t.Fatalf("Update coin failed: %v", err)
	}

	var found models.Coin
	if err := db.Preload("Images").Preload("References").Preload("Tags").Preload("Sets").Preload("StorageLocation").First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found after update: %v", err)
	}
	if found.Name != "Updated Associated Coin" {
		t.Fatalf("expected name update, got %q", found.Name)
	}
	if len(found.Images) != 1 || found.Images[0].FilePath != "coins/original.jpg" {
		t.Fatalf("expected original image to remain, got %#v", found.Images)
	}
	if len(found.References) != 1 || found.References[0].Number != "1" {
		t.Fatalf("expected original reference to remain, got %#v", found.References)
	}
	if len(found.Tags) != 1 || found.Tags[0].ID != tag.ID {
		t.Fatalf("expected original tag to remain, got %#v", found.Tags)
	}
	if len(found.Sets) != 1 || found.Sets[0].ID != set.ID {
		t.Fatalf("expected original set to remain, got %#v", found.Sets)
	}
	if found.StorageLocationID == nil || *found.StorageLocationID != location.ID {
		t.Fatalf("expected original storage location %d to remain, got %v", location.ID, found.StorageLocationID)
	}
	var incomingImageCount int64
	if err := db.Model(&models.CoinImage{}).Where("file_path = ?", "coins/incoming.jpg").Count(&incomingImageCount).Error; err != nil {
		t.Fatalf("failed to count incoming image: %v", err)
	}
	if incomingImageCount != 0 {
		t.Fatal("incoming image association was persisted by scalar update")
	}
}

func TestCoinRepository_Update_WithSelectedFieldsPersistsExplicitZeroValues(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)

	purchasePrice := 125.0
	currentValue := 175.0
	weight := 3.5
	diameter := 18.0
	coin := &models.Coin{
		Name:             "Zero Value Coin",
		Category:         models.CategoryRoman,
		Material:         models.MaterialSilver,
		UserID:           1,
		Notes:            "clear me",
		ReferenceURL:     "https://example.test/ref",
		ReferenceText:    "clear reference text",
		PurchaseLocation: "Old dealer",
		IsPrivate:        true,
		IsWishlist:       true,
		IsSold:           true,
		PurchasePrice:    &purchasePrice,
		CurrentValue:     &currentValue,
		WeightGrams:      &weight,
		DiameterMm:       &diameter,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}

	zero := 0.0
	updates := &models.Coin{
		Notes:            "",
		ReferenceURL:     "",
		ReferenceText:    "",
		PurchaseLocation: "",
		IsPrivate:        false,
		IsWishlist:       false,
		IsSold:           false,
		PurchasePrice:    &zero,
		CurrentValue:     &zero,
		WeightGrams:      &zero,
		DiameterMm:       &zero,
	}
	if err := coinRepo.Update(
		coin,
		updates,
		"Notes",
		"ReferenceURL",
		"ReferenceText",
		"PurchaseLocation",
		"IsPrivate",
		"IsWishlist",
		"IsSold",
		"PurchasePrice",
		"CurrentValue",
		"WeightGrams",
		"DiameterMm",
	); err != nil {
		t.Fatalf("Update coin failed: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found after update: %v", err)
	}
	if found.Notes != "" || found.ReferenceURL != "" || found.ReferenceText != "" || found.PurchaseLocation != "" {
		t.Fatalf("expected empty string clears to persist, got notes=%q refURL=%q refText=%q purchaseLocation=%q",
			found.Notes, found.ReferenceURL, found.ReferenceText, found.PurchaseLocation)
	}
	if found.IsPrivate || found.IsWishlist || found.IsSold {
		t.Fatalf("expected false booleans to persist, got private=%v wishlist=%v sold=%v", found.IsPrivate, found.IsWishlist, found.IsSold)
	}
	if found.PurchasePrice == nil || *found.PurchasePrice != 0 ||
		found.CurrentValue == nil || *found.CurrentValue != 0 ||
		found.WeightGrams == nil || *found.WeightGrams != 0 ||
		found.DiameterMm == nil || *found.DiameterMm != 0 {
		t.Fatalf("expected explicit numeric zeros to persist, got purchase=%v current=%v weight=%v diameter=%v",
			found.PurchasePrice, found.CurrentValue, found.WeightGrams, found.DiameterMm)
	}
	if found.Name != "Zero Value Coin" || found.Category != models.CategoryRoman || found.Material != models.MaterialSilver {
		t.Fatalf("omitted fields changed unexpectedly: name=%q category=%q material=%q", found.Name, found.Category, found.Material)
	}
}

func TestCoinRepository_Update_WithSelectedFieldsPersistsNilNullableScalars(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)

	purchasePrice := 125.0
	currentValue := 175.0
	weight := 3.5
	diameter := 18.0
	purchaseDate := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	soldPrice := 150.0
	soldDate := time.Date(2025, time.February, 20, 0, 0, 0, 0, time.UTC)
	coin := &models.Coin{
		Name:          "Nil Nullable Coin",
		Category:      models.CategoryRoman,
		Material:      models.MaterialSilver,
		UserID:        1,
		PurchasePrice: &purchasePrice,
		CurrentValue:  &currentValue,
		WeightGrams:   &weight,
		DiameterMm:    &diameter,
		PurchaseDate:  &purchaseDate,
		SoldPrice:     &soldPrice,
		SoldDate:      &soldDate,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}

	updates := &models.Coin{}
	if err := coinRepo.Update(
		coin,
		updates,
		"PurchasePrice",
		"CurrentValue",
		"PurchaseDate",
		"SoldPrice",
		"SoldDate",
		"WeightGrams",
		"DiameterMm",
	); err != nil {
		t.Fatalf("Update coin failed: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found after update: %v", err)
	}
	if found.PurchasePrice != nil || found.CurrentValue != nil || found.PurchaseDate != nil ||
		found.SoldPrice != nil || found.SoldDate != nil || found.WeightGrams != nil || found.DiameterMm != nil {
		t.Fatalf("expected selected nil nullable scalars to persist, got purchase=%v current=%v purchaseDate=%v sold=%v soldDate=%v weight=%v diameter=%v",
			found.PurchasePrice, found.CurrentValue, found.PurchaseDate, found.SoldPrice, found.SoldDate, found.WeightGrams, found.DiameterMm)
	}
	if found.Name != "Nil Nullable Coin" || found.Category != models.CategoryRoman || found.Material != models.MaterialSilver {
		t.Fatalf("omitted fields changed unexpectedly: name=%q category=%q material=%q", found.Name, found.Category, found.Material)
	}
}

func TestCoinRepository_StorageAssignmentContractsAndLegacyNullClear(t *testing.T) {
	db := setupTestDB(t)
	coinRepo := NewCoinRepository(db)

	location := models.StorageLocation{UserID: 1, Name: "Cabinet A"}
	if err := db.Create(&location).Error; err != nil {
		t.Fatalf("Create storage location failed: %v", err)
	}

	t.Run("UpdateStorageAssignment is atomic and clears standard slots", func(t *testing.T) {
		db := setupTestDB(t)
		addStorageSlotUniqueIndex(t, db)
		repo := NewCoinRepository(db)
		rows, columns := 2, 2
		tray := models.StorageLocation{UserID: 1, Name: "Tray", Type: models.StorageLocationTypeTray, Rows: &rows, Columns: &columns}
		standard := models.StorageLocation{UserID: 1, Name: "Safe", Type: models.StorageLocationTypeStandard}
		if err := db.Create(&[]models.StorageLocation{tray, standard}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Where("name = ?", "Tray").First(&tray).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Where("name = ?", "Safe").First(&standard).Error; err != nil {
			t.Fatal(err)
		}
		slotOne, slotTwo := 1, 2
		moving := models.Coin{UserID: 1, Name: "Moving", StorageLocationID: &tray.ID, StorageSlot: &slotOne}
		blocking := models.Coin{UserID: 1, Name: "Blocking", StorageLocationID: &tray.ID, StorageSlot: &slotTwo}
		if err := db.Create(&[]*models.Coin{&moving, &blocking}).Error; err != nil {
			t.Fatal(err)
		}

		if err := repo.UpdateStorageAssignment(&moving, &tray.ID, &slotTwo); !errors.Is(err, ErrStorageSlotOccupied) {
			t.Fatalf("expected named-index conflict translation, got %v", err)
		}
		var persisted models.Coin
		if err := db.First(&persisted, moving.ID).Error; err != nil {
			t.Fatal(err)
		}
		if persisted.StorageLocationID == nil || *persisted.StorageLocationID != tray.ID || persisted.StorageSlot == nil || *persisted.StorageSlot != slotOne {
			t.Fatalf("failed move changed prior assignment: %#v", persisted)
		}

		if err := repo.UpdateStorageAssignment(&moving, &standard.ID, nil); err != nil {
			t.Fatalf("move to standard: %v", err)
		}
		if moving.StorageLocationID == nil || *moving.StorageLocationID != standard.ID || moving.StorageSlot != nil {
			t.Fatalf("standard move did not atomically clear slot: %#v", moving)
		}
	})

	t.Run("UpdateStorageAssignment concurrent claims have one winner", func(t *testing.T) {
		db := setupTestDB(t)
		addStorageSlotUniqueIndex(t, db)
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatal(err)
		}
		sqlDB.SetMaxOpenConns(4)

		rows, columns, slot := 1, 1, 1
		tray := models.StorageLocation{UserID: 1, Name: "One Well", Type: models.StorageLocationTypeTray, Rows: &rows, Columns: &columns}
		if err := db.Create(&tray).Error; err != nil {
			t.Fatal(err)
		}
		coins := []models.Coin{{UserID: 1, Name: "First"}, {UserID: 1, Name: "Second"}}
		if err := db.Create(&coins).Error; err != nil {
			t.Fatal(err)
		}

		start := make(chan struct{})
		results := make(chan error, len(coins))
		var wg sync.WaitGroup
		for i := range coins {
			wg.Add(1)
			go func(coin *models.Coin) {
				defer wg.Done()
				<-start
				results <- NewCoinRepository(db).UpdateStorageAssignment(coin, &tray.ID, &slot)
			}(&coins[i])
		}
		close(start)
		wg.Wait()
		close(results)

		successes, conflicts := 0, 0
		for result := range results {
			switch {
			case result == nil:
				successes++
			case errors.Is(result, ErrStorageSlotOccupied):
				conflicts++
			default:
				t.Fatalf("unexpected concurrent claim error: %v", result)
			}
		}
		if successes != 1 || conflicts != 1 {
			t.Fatalf("expected one winner and one slot conflict, got successes=%d conflicts=%d", successes, conflicts)
		}
		var occupied int64
		if err := db.Model(&models.Coin{}).Where("storage_location_id = ? AND storage_slot = ?", tray.ID, slot).Count(&occupied).Error; err != nil {
			t.Fatal(err)
		}
		if occupied != 1 {
			t.Fatalf("expected exactly one persisted claimant, got %d", occupied)
		}
	})
	coin := &models.Coin{
		Name:              "Storage Clear Coin",
		Category:          models.CategoryRoman,
		Material:          models.MaterialSilver,
		UserID:            1,
		StorageLocationID: &location.ID,
	}
	if err := coinRepo.Create(coin); err != nil {
		t.Fatalf("Create coin failed: %v", err)
	}

	if err := coinRepo.UpdateStorageLocationID(coin, nil); err != nil {
		t.Fatalf("UpdateStorageLocationID clear failed: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found after clear: %v", err)
	}
	if found.StorageLocationID != nil {
		t.Fatalf("expected storage location NULL, got %v", found.StorageLocationID)
	}
	if coin.StorageLocationID != nil || coin.StorageLocation != nil {
		t.Fatalf("expected reloaded coin to have nil storage fields, got id=%v location=%v", coin.StorageLocationID, coin.StorageLocation)
	}
}

func TestCoinRepository_UpdateHelpers_WithLoadedSetsDoNotSyncMemberships(t *testing.T) {
	tests := []struct {
		name   string
		update func(repo *CoinRepository, coin *models.Coin) error
		assert func(t *testing.T, db *gorm.DB, coinID uint)
	}{
		{
			name: "UpdateField",
			update: func(repo *CoinRepository, coin *models.Coin) error {
				return repo.UpdateField(coin, "name", "Updated Field Coin")
			},
			assert: func(t *testing.T, db *gorm.DB, coinID uint) {
				t.Helper()
				var found models.Coin
				if err := db.First(&found, coinID).Error; err != nil {
					t.Fatalf("coin not found: %v", err)
				}
				if found.Name != "Updated Field Coin" {
					t.Fatalf("expected updated name, got %q", found.Name)
				}
			},
		},
		{
			name: "UpdateFields",
			update: func(repo *CoinRepository, coin *models.Coin) error {
				return repo.UpdateFields(coin, map[string]interface{}{"name": "Updated Fields Coin"})
			},
			assert: func(t *testing.T, db *gorm.DB, coinID uint) {
				t.Helper()
				var found models.Coin
				if err := db.First(&found, coinID).Error; err != nil {
					t.Fatalf("coin not found: %v", err)
				}
				if found.Name != "Updated Fields Coin" {
					t.Fatalf("expected updated name, got %q", found.Name)
				}
			},
		},
		{
			name: "UpdateStorageLocationID",
			update: func(repo *CoinRepository, coin *models.Coin) error {
				storageLocationID := uint(5)
				return repo.UpdateStorageLocationID(coin, &storageLocationID)
			},
			assert: func(t *testing.T, db *gorm.DB, coinID uint) {
				t.Helper()
				var found models.Coin
				if err := db.First(&found, coinID).Error; err != nil {
					t.Fatalf("coin not found: %v", err)
				}
				if found.StorageLocationID == nil || *found.StorageLocationID != 5 {
					t.Fatalf("expected storage location 5, got %v", found.StorageLocationID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			coinRepo := NewCoinRepository(db)
			setRepo := NewSetRepository(db)

			coin := &models.Coin{
				Name:     "Loaded Set Coin",
				Category: models.CategoryRoman,
				UserID:   1,
			}
			if err := coinRepo.Create(coin); err != nil {
				t.Fatalf("Create coin failed: %v", err)
			}

			set := &models.CoinSet{UserID: 1, Name: "Loaded Set", SetType: models.CoinSetTypeOpen}
			if err := setRepo.Create(set); err != nil {
				t.Fatalf("Create set failed: %v", err)
			}
			if err := setRepo.AddCoinToSet(coin.ID, set.ID, 1, ""); err != nil {
				t.Fatalf("AddCoinToSet failed: %v", err)
			}

			loaded, err := coinRepo.FindByID(coin.ID, 1)
			if err != nil {
				t.Fatalf("FindByID failed: %v", err)
			}
			if len(loaded.Sets) != 1 {
				t.Fatalf("expected preloaded set, got %d", len(loaded.Sets))
			}

			var originalMembership models.CoinSetMembership
			if err := db.Where("coin_id = ? AND set_id = ?", coin.ID, set.ID).First(&originalMembership).Error; err != nil {
				t.Fatalf("membership not found: %v", err)
			}

			if err := tt.update(coinRepo, loaded); err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			tt.assert(t, db, coin.ID)

			var memberships []models.CoinSetMembership
			if err := db.Where("coin_id = ?", coin.ID).Find(&memberships).Error; err != nil {
				t.Fatalf("failed to query memberships: %v", err)
			}
			if len(memberships) != 1 {
				t.Fatalf("expected exactly 1 membership after %s, got %d", tt.name, len(memberships))
			}
			if memberships[0].SetID != set.ID {
				t.Fatalf("expected original membership to remain, got set ID %d", memberships[0].SetID)
			}
			if memberships[0].AddedAt.IsZero() {
				t.Fatal("membership AddedAt should remain populated")
			}
			if !memberships[0].AddedAt.Equal(originalMembership.AddedAt) {
				t.Fatalf("membership AddedAt changed from %v to %v", originalMembership.AddedAt, memberships[0].AddedAt)
			}
		})
	}
}

func TestCoinRepository_WishlistCoinsWithReferences(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)
	coins := []models.Coin{
		{Name: "Wanted", UserID: 1, IsWishlist: true, References: []models.CoinReference{{Catalog: "Numista", Number: "9002"}}},
		{Name: "Had", UserID: 1},
		{Name: "Sold wish", UserID: 1, IsWishlist: true, IsSold: true},
		{Name: "Wanted note", UserID: 1, IsWishlist: true, Kind: models.CoinKindBanknote},
		{Name: "Someone else's", UserID: 2, IsWishlist: true},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}

	got, err := repo.WishlistCoinsWithReferences(1, models.CoinKindCoin)
	if err != nil {
		t.Fatalf("WishlistCoinsWithReferences failed: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Wanted" || len(got[0].References) != 1 || got[0].References[0].Number != "9002" {
		t.Fatalf("wishlist = %+v", got)
	}
}

// The gallery's face value filter: the face values of the collection with
// their counts, the most common first, and the coins of one of them.
func TestCoinRepository_DenominationCountsAndFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)
	for _, coin := range []models.Coin{
		{Name: "Malta", Denomination: "2 euro", UserID: 1},
		{Name: "Espanha", Denomination: " 2 euro ", UserID: 1},
		{Name: "Escudo", Denomination: "1 escudo", UserID: 1},
		{Name: "Wanted", Denomination: "5 euro", UserID: 1, IsWishlist: true},
		{Name: "Sold", Denomination: "5 euro", UserID: 1, IsSold: true},
		{Name: "No value", UserID: 1},
		{Name: "Another user's", Denomination: "10 euro", UserID: 2},
	} {
		coin.Category = models.CategoryModern
		if err := repo.Create(&coin); err != nil {
			t.Fatal(err)
		}
	}

	counts, err := repo.DenominationCounts(1, "")
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(counts) != "[{2 euro 2} {1 escudo 1}]" {
		t.Fatalf("counts = %v", counts)
	}
	coins, total, err := repo.List(1, CoinListFilters{Denomination: "2 euro", Page: 1, Limit: 50})
	if err != nil || total != 2 || len(coins) != 2 {
		t.Fatalf("2 euro coins = %d of %d, %v", len(coins), total, err)
	}
}
