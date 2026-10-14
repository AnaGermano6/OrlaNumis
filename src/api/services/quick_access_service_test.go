package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var quickAccessTestDBCounter uint64

func newQuickAccessTestService(t *testing.T) (*QuickAccessService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:quick_access_service_%d_%d?mode=memory&cache=shared",
		time.Now().UnixNano(), atomic.AddUint64(&quickAccessTestDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(
		&models.Coin{}, &models.CoinImage{}, &models.CoinSet{},
		&models.QuickAccessPin{},
		&models.CoinJournal{}, &models.CoinValueHistory{}, &models.CoinComment{},
		&models.AvailabilityRun{}, &models.AvailabilityResult{}, &models.CoinTag{},
		&models.CoinSetMembership{}, &models.CoinSetChecklistKey{}, &models.CoinReference{}, &models.StorageLocation{},
		&models.MintLocation{},
	); err != nil {
		t.Fatal(err)
	}

	return NewQuickAccessService(repository.NewQuickAccessRepository(db)), db
}

func TestQuickAccessLifecycleCleanupAndPreservation(t *testing.T) {
	quickAccess, db := newQuickAccessTestService(t)
	userID := uint(41)
	coinRepo := repository.NewCoinRepository(db)
	coinSvc := NewCoinService(coinRepo, nil).WithQuickAccessSupport(quickAccess)

	wishlist := models.Coin{UserID: userID, Name: "Wishlist", IsWishlist: true}
	if err := db.Create(&wishlist).Error; err != nil {
		t.Fatal(err)
	}
	pinned, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, wishlist.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := coinSvc.PurchaseCoin(&wishlist, userID); err != nil {
		t.Fatal(err)
	}
	list, err := quickAccess.List(userID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Items) != 1 || list.Items[0].Coin == nil || list.Items[0].Coin.Classification != "owned" || !list.Items[0].PinnedAt.Equal(pinned.PinnedAt) {
		t.Fatalf("purchase did not preserve/reclassify pin: %#v", list.Items)
	}
	genericWishlist := models.Coin{UserID: userID, Name: "Generic Wishlist", IsWishlist: true}
	if err := db.Create(&genericWishlist).Error; err != nil {
		t.Fatal(err)
	}
	genericPin, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, genericWishlist.ID)
	if err != nil {
		t.Fatal(err)
	}
	genericUpdate := models.Coin{IsWishlist: false}
	if err := coinSvc.UpdateCoinWithFields(&genericWishlist, &genericUpdate, []string{"IsWishlist"}, userID, "manual", false); err != nil {
		t.Fatal(err)
	}
	var genericPreserved models.QuickAccessPin
	if err := db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, models.QuickAccessTargetCoin, genericWishlist.ID).First(&genericPreserved).Error; err != nil {
		t.Fatal(err)
	}
	if !genericPreserved.PinnedAt.Equal(genericPin.PinnedAt) {
		t.Fatal("generic wishlist purchase changed pin time")
	}
	updates := models.Coin{IsSold: true}
	if err := coinSvc.UpdateCoinWithFields(&wishlist, &updates, []string{"IsSold"}, userID, "manual", false); err != nil {
		t.Fatal(err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, wishlist.ID, 0)
	updates.IsSold = false
	if err := coinSvc.UpdateCoinWithFields(&wishlist, &updates, []string{"IsSold"}, userID, "manual", false); err != nil {
		t.Fatal(err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, wishlist.ID, 0)

	deleteCoin := models.Coin{UserID: userID, Name: "Delete Coin"}
	if err := db.Create(&deleteCoin).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, deleteCoin.ID); err != nil {
		t.Fatal(err)
	}
	if rows, err := coinSvc.DeleteCoin(deleteCoin.ID, userID); err != nil || rows != 1 {
		t.Fatalf("delete coin rows=%d err=%v", rows, err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, deleteCoin.ID, 0)

	sellCoin := models.Coin{UserID: userID, Name: "Dedicated Sell"}
	if err := db.Create(&sellCoin).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, sellCoin.ID); err != nil {
		t.Fatal(err)
	}
	if err := coinSvc.SellCoin(&sellCoin, map[string]interface{}{"is_sold": true}, userID); err != nil {
		t.Fatal(err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, sellCoin.ID, 0)

	bulkSold := models.Coin{UserID: userID, Name: "Bulk Sold"}
	bulkDeleted := models.Coin{UserID: userID, Name: "Bulk Deleted"}
	if err := db.Create(&bulkSold).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&bulkDeleted).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, bulkSold.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, bulkDeleted.ID); err != nil {
		t.Fatal(err)
	}
	if affected, err := coinSvc.BulkMarkSold([]uint{bulkSold.ID}, userID); err != nil || affected != 1 {
		t.Fatalf("bulk sell affected=%d err=%v", affected, err)
	}
	if affected, err := coinSvc.BulkDeleteCoins([]uint{bulkDeleted.ID}, userID); err != nil || affected != 1 {
		t.Fatalf("bulk delete affected=%d err=%v", affected, err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, bulkSold.ID, 0)
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, bulkDeleted.ID, 0)

	set := models.CoinSet{UserID: userID, Name: "Delete Me"}
	if err := db.Create(&set).Error; err != nil {
		t.Fatal(err)
	}
	membershipCoin := models.Coin{UserID: userID, Name: "Set Member"}
	if err := db.Create(&membershipCoin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.CoinSetMembership{SetID: set.ID, CoinID: membershipCoin.ID, AddedAt: time.Now().UTC()}).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoinSet, set.ID); err != nil {
		t.Fatal(err)
	}
	setSvc := NewSetService(repository.NewSetRepository(db), repository.NewTagRepository(db)).WithQuickAccessSupport(quickAccess)
	if err := setSvc.DeleteSet(set.ID, userID); err != nil {
		t.Fatal(err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoinSet, set.ID, 0)
	var membershipCount int64
	if err := db.Model(&models.CoinSetMembership{}).Where("set_id = ?", set.ID).Count(&membershipCount).Error; err != nil {
		t.Fatal(err)
	}
	if membershipCount != 0 {
		t.Fatalf("set delete left %d memberships", membershipCount)
	}

	legacySet := models.CoinSet{UserID: userID, Name: "Legacy Toggle"}
	if err := db.Create(&legacySet).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := setSvc.UpdateSet(legacySet.ID, userID, map[string]interface{}{"pinned": true}); err != nil {
		t.Fatal(err)
	}
	var legacyPin models.QuickAccessPin
	if err := db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, models.QuickAccessTargetCoinSet, legacySet.ID).First(&legacyPin).Error; err != nil {
		t.Fatal(err)
	}
	var mirrored models.CoinSet
	if err := db.First(&mirrored, legacySet.ID).Error; err != nil {
		t.Fatal(err)
	}
	if mirrored.PinnedAt == nil || !mirrored.PinnedAt.Equal(legacyPin.PinnedAt) {
		t.Fatalf("legacy endpoint mirror=%v pin=%s", mirrored.PinnedAt, legacyPin.PinnedAt)
	}
	if _, err := setSvc.UpdateSet(legacySet.ID, userID, map[string]interface{}{"pinned": true}); err != nil {
		t.Fatal(err)
	}
	var repeated models.QuickAccessPin
	if err := db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, models.QuickAccessTargetCoinSet, legacySet.ID).First(&repeated).Error; err != nil {
		t.Fatal(err)
	}
	if repeated.ID != legacyPin.ID || !repeated.PinnedAt.Equal(legacyPin.PinnedAt) {
		t.Fatalf("legacy repeated pin changed identity/time: before=%#v after=%#v", legacyPin, repeated)
	}
	if _, err := setSvc.UpdateSet(legacySet.ID, userID, map[string]interface{}{"pinned": false}); err != nil {
		t.Fatal(err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoinSet, legacySet.ID, 0)
	time.Sleep(20 * time.Millisecond)
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoinSet, legacySet.ID); err != nil {
		t.Fatal(err)
	}
	var repinned models.QuickAccessPin
	if err := db.Where("user_id = ? AND target_type = ? AND target_id = ?", userID, models.QuickAccessTargetCoinSet, legacySet.ID).First(&repinned).Error; err != nil {
		t.Fatal(err)
	}
	if repinned.ID == legacyPin.ID || !repinned.PinnedAt.After(legacyPin.PinnedAt) {
		t.Fatalf("unpin/re-pin did not create new identity/time: before=%#v after=%#v", legacyPin, repinned)
	}

	rollbackCoin := models.Coin{UserID: userID, Name: "Rollback Coin"}
	if err := db.Create(&rollbackCoin).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoin, rollbackCoin.ID); err != nil {
		t.Fatal(err)
	}
	installQuickAccessDeleteFailure(t, db)
	rollbackUpdate := models.Coin{IsSold: true}
	if err := coinSvc.UpdateCoinWithFields(&rollbackCoin, &rollbackUpdate, []string{"IsSold"}, userID, "manual", false); err == nil {
		t.Fatal("expected forced pin cleanup failure")
	}
	removeQuickAccessDeleteFailure(t, db)
	var reloadedRollbackCoin models.Coin
	if err := db.First(&reloadedRollbackCoin, rollbackCoin.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reloadedRollbackCoin.IsSold {
		t.Fatal("coin sale was not rolled back after cleanup failure")
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoin, rollbackCoin.ID, 1)

	rollbackSet := models.CoinSet{UserID: userID, Name: "Rollback Set"}
	if err := db.Create(&rollbackSet).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := quickAccess.Pin(userID, models.QuickAccessTargetCoinSet, rollbackSet.ID); err != nil {
		t.Fatal(err)
	}
	installQuickAccessDeleteFailure(t, db)
	if err := setSvc.DeleteSet(rollbackSet.ID, userID); err == nil {
		t.Fatal("expected set cleanup failure")
	}
	removeQuickAccessDeleteFailure(t, db)
	if err := db.First(&models.CoinSet{}, rollbackSet.ID).Error; err != nil {
		t.Fatalf("set delete was not rolled back: %v", err)
	}
	assertQuickAccessPinCount(t, db, userID, models.QuickAccessTargetCoinSet, rollbackSet.ID, 1)
}

func assertQuickAccessPinCount(t *testing.T, db *gorm.DB, userID uint, targetType models.QuickAccessTargetType, targetID uint, want int64) {
	t.Helper()
	var count int64
	if err := db.Model(&models.QuickAccessPin{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("pin count=%d want %d for %s/%d", count, want, targetType, targetID)
	}
}

func installQuickAccessDeleteFailure(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`CREATE TRIGGER fail_quick_access_cleanup BEFORE DELETE ON quick_access_pins BEGIN SELECT RAISE(ABORT, 'forced cleanup failure'); END`).Error; err != nil {
		t.Fatal(err)
	}
}

func removeQuickAccessDeleteFailure(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(`DROP TRIGGER fail_quick_access_cleanup`).Error; err != nil {
		t.Fatal(err)
	}
}

func TestQuickAccessServicePinListAndIdempotency(t *testing.T) {
	svc, db := newQuickAccessTestService(t)
	userID := uint(11)
	otherUserID := uint(12)
	coin := models.Coin{UserID: userID, Name: "Wishlist Coin", IsWishlist: true}
	set := models.CoinSet{UserID: userID, Name: "Roman Set"}
	foreign := models.Coin{UserID: otherUserID, Name: "Foreign Coin"}
	for _, value := range []interface{}{&coin, &set, &foreign} {
		if err := db.Create(value).Error; err != nil {
			t.Fatal(err)
		}
	}

	first, created, err := svc.Pin(userID, models.QuickAccessTargetCoin, coin.ID)
	if err != nil || !created {
		t.Fatalf("first pin: created=%v err=%v", created, err)
	}
	second, created, err := svc.Pin(userID, models.QuickAccessTargetCoin, coin.ID)
	if err != nil || created {
		t.Fatalf("second pin: created=%v err=%v", created, err)
	}
	if !second.PinnedAt.Equal(first.PinnedAt) {
		t.Fatalf("idempotent pin changed timestamp: %s != %s", second.PinnedAt, first.PinnedAt)
	}
	if _, _, err := svc.Pin(userID, models.QuickAccessTargetCoin, foreign.ID); !errors.Is(err, ErrQuickAccessTargetNotFound) {
		t.Fatalf("foreign pin error = %v", err)
	}

	if _, created, err := svc.Pin(userID, models.QuickAccessTargetCoinSet, set.ID); err != nil || !created {
		t.Fatalf("pin set: created=%v err=%v", created, err)
	}

	queryCount := 0
	if err := db.Callback().Query().Before("gorm:query").Register("feature357_query_count", func(*gorm.DB) {
		queryCount++
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Row().Before("gorm:row").Register("feature357_row_count", func(*gorm.DB) {
		queryCount++
	}); err != nil {
		t.Fatal(err)
	}
	list, err := svc.List(userID)
	if err != nil {
		t.Fatal(err)
	}
	if queryCount != 3 {
		t.Fatalf("list queries=%d want 3 (pins plus one bounded query per target type)", queryCount)
	}
	if len(list.Items) != 2 {
		t.Fatalf("items=%d want 2", len(list.Items))
	}
	for _, item := range list.Items {
		payloads := 0
		for _, present := range []bool{item.Coin != nil, item.CoinSet != nil} {
			if present {
				payloads++
			}
		}
		if payloads != 1 {
			t.Fatalf("item %s/%d has %d payloads", item.Type, item.ID, payloads)
		}
	}
	if list.Items[len(list.Items)-1].Coin == nil || list.Items[len(list.Items)-1].Coin.Classification != "wishlist" {
		t.Fatalf("coin classification not derived: %#v", list.Items[len(list.Items)-1])
	}

	encoded, err := json.Marshal(list.Items[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) == "" {
		t.Fatal("expected DTO JSON")
	}
	for _, privateField := range []string{"userId", "passwordHash", "notes", "description", "sourceUrl"} {
		if strings.Contains(string(encoded), privateField) {
			t.Fatalf("DTO leaked private/full-model field %q: %s", privateField, encoded)
		}
	}

	if err := svc.Unpin(userID, models.QuickAccessTargetCoin, coin.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.Unpin(userID, models.QuickAccessTargetCoin, coin.ID); err != nil {
		t.Fatalf("repeated unpin: %v", err)
	}
}

func TestQuickAccessServiceEligibilityAndSetCap(t *testing.T) {
	svc, db := newQuickAccessTestService(t)
	userID := uint(21)
	if _, err := ParseQuickAccessTargetType("invalid"); !errors.Is(err, ErrInvalidQuickAccessTarget) {
		t.Fatalf("invalid type error=%v", err)
	}
	// Auction lots and calendar events were retired: no longer pinnable.
	for _, retired := range []string{"auction_lot", "calendar_event"} {
		if _, err := ParseQuickAccessTargetType(retired); !errors.Is(err, ErrInvalidQuickAccessTarget) {
			t.Fatalf("%s type error=%v", retired, err)
		}
	}
	sold := models.Coin{UserID: userID, Name: "Sold", IsSold: true}
	if err := db.Create(&sold).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Pin(userID, models.QuickAccessTargetCoin, sold.ID); !errors.Is(err, ErrQuickAccessTargetNotFound) {
		t.Fatalf("sold coin ineligible error=%v", err)
	}

	for i := 0; i < maxQuickAccessCoinSets+1; i++ {
		set := models.CoinSet{UserID: userID, Name: fmt.Sprintf("Set %d", i)}
		if err := db.Create(&set).Error; err != nil {
			t.Fatal(err)
		}
		_, _, err := svc.Pin(userID, models.QuickAccessTargetCoinSet, set.ID)
		if i < maxQuickAccessCoinSets && err != nil {
			t.Fatalf("pin set %d: %v", i, err)
		}
		if i == maxQuickAccessCoinSets && !errors.Is(err, ErrPinLimitReached) {
			t.Fatalf("sixth set error=%v", err)
		}
		if i == 0 {
			var mirrored models.CoinSet
			if err := db.First(&mirrored, set.ID).Error; err != nil {
				t.Fatal(err)
			}
			if mirrored.PinnedAt == nil {
				t.Fatal("set compatibility mirror not populated")
			}
		}
	}
	coin := models.Coin{UserID: userID, Name: "Uncapped Coin"}
	if err := db.Create(&coin).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Pin(userID, models.QuickAccessTargetCoin, coin.ID); err != nil {
		t.Fatalf("non-set pin counted toward set cap: %v", err)
	}
	legacySet := models.CoinSet{UserID: userID, Name: "Legacy Sixth"}
	nextSet := models.CoinSet{UserID: userID, Name: "Blocked Seventh"}
	if err := db.Create(&legacySet).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&nextSet).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.QuickAccessPin{
		UserID: userID, TargetType: models.QuickAccessTargetCoinSet,
		TargetID: legacySet.ID, PinnedAt: time.Now().UTC(),
	}).Error; err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.Pin(userID, models.QuickAccessTargetCoinSet, nextSet.ID); !errors.Is(err, ErrPinLimitReached) {
		t.Fatalf("legacy over-cap additional pin error=%v", err)
	}
}

func TestQuickAccessServiceOmitsStaleWithoutDeleting(t *testing.T) {
	svc, db := newQuickAccessTestService(t)
	pin := models.QuickAccessPin{
		UserID: 31, TargetType: models.QuickAccessTargetCoin, TargetID: 999,
		PinnedAt: time.Now().UTC(),
	}
	if err := db.Create(&pin).Error; err != nil {
		t.Fatal(err)
	}
	list, err := svc.List(31)
	if err != nil {
		t.Fatal(err)
	}
	if list.Items == nil || len(list.Items) != 0 {
		t.Fatalf("items=%#v want empty non-nil slice", list.Items)
	}
	var count int64
	if err := db.Model(&models.QuickAccessPin{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("GET mutated stale pin count=%d", count)
	}
}
