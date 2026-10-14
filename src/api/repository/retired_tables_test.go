package repository

import (
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// The tables of retired features (auction tracking, the calendar, shipments,
// notes, value history) are no longer migrated, but an older database still
// has them, foreign keys included. Deleting a coin or a user must still work.
func setupRetiredTablesDB(t *testing.T) *gorm.DB {
	t.Helper()
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
	if err := db.AutoMigrate(
		&models.AgentConversation{}, &models.ApiKey{}, &models.RefreshToken{},
		&models.WebAuthnCredential{}, &models.QuickAccessPin{}, &models.Follow{},
	); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"CREATE TABLE auction_events (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), title TEXT)",
		"CREATE TABLE auction_lots (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), coin_id INTEGER REFERENCES coins(id), event_id INTEGER REFERENCES auction_events(id), title TEXT)",
		"CREATE TABLE price_alerts (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), auction_lot_id INTEGER NOT NULL REFERENCES auction_lots(id))",
		"CREATE TABLE bid_reminders (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), auction_lot_id INTEGER NOT NULL REFERENCES auction_lots(id))",
		"CREATE TABLE shipments (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), coin_id INTEGER NOT NULL REFERENCES coins(id))",
		"CREATE TABLE shipment_events (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), shipment_id INTEGER NOT NULL REFERENCES shipments(id))",
		"CREATE TABLE notes (id INTEGER PRIMARY KEY, user_id INTEGER)",
		"CREATE TABLE value_snapshots (id INTEGER PRIMARY KEY, user_id INTEGER)",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func retiredCount(t *testing.T, db *gorm.DB, sql string, args ...interface{}) int64 {
	t.Helper()
	var count int64
	if err := db.Raw(sql, args...).Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	return count
}

func TestCoinRepository_DeleteLetsGoOfRetiredRows(t *testing.T) {
	db := setupRetiredTablesDB(t)
	repo := NewCoinRepository(db)
	if err := db.Create(&models.User{Username: "collector"}).Error; err != nil {
		t.Fatal(err)
	}
	coins := make([]*models.Coin, 3)
	for i := range coins {
		coins[i] = &models.Coin{Name: "Denarius", Category: models.CategoryRoman, UserID: 1}
		if err := repo.Create(coins[i]); err != nil {
			t.Fatal(err)
		}
	}
	for _, sql := range []string{
		"INSERT INTO auction_lots (id, user_id, coin_id, title) VALUES (1, 1, ?, 'Won lot')",
		"INSERT INTO shipments (id, user_id, coin_id) VALUES (1, 1, ?)",
		"INSERT INTO shipment_events (user_id, shipment_id) VALUES (1, 1)",
	} {
		if err := db.Exec(sql, coins[0].ID).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("INSERT INTO auction_lots (id, user_id, coin_id, title) VALUES (2, 1, ?, 'Other lot')", coins[1].ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO shipments (id, user_id, coin_id) VALUES (2, 1, ?)", coins[2].ID).Error; err != nil {
		t.Fatal(err)
	}

	if rows, err := repo.Delete(coins[0].ID, 1); err != nil || rows != 1 {
		t.Fatalf("Delete: rows=%d err=%v", rows, err)
	}
	if rows, err := repo.BulkDelete([]uint{coins[1].ID, coins[2].ID}, 1); err != nil || rows != 2 {
		t.Fatalf("BulkDelete: rows=%d err=%v", rows, err)
	}

	// A lot only forgets its coin; a shipment goes with it, events included.
	if lots := retiredCount(t, db, "SELECT COUNT(*) FROM auction_lots WHERE coin_id IS NULL"); lots != 2 {
		t.Fatalf("unlinked lots=%d, want both kept", lots)
	}
	if left := retiredCount(t, db, "SELECT (SELECT COUNT(*) FROM shipments) + (SELECT COUNT(*) FROM shipment_events)"); left != 0 {
		t.Fatalf("shipment rows left=%d", left)
	}
}

func TestCoinRepository_DeleteWithoutRetiredTables(t *testing.T) {
	db := setupTestDB(t)
	repo := NewCoinRepository(db)
	coin := &models.Coin{Name: "Sestertius", Category: models.CategoryRoman, UserID: 1}
	if err := repo.Create(coin); err != nil {
		t.Fatal(err)
	}
	if rows, err := repo.Delete(coin.ID, 1); err != nil || rows != 1 {
		t.Fatalf("Delete in a new database: rows=%d err=%v", rows, err)
	}
}

func TestAdminRepository_DeleteUserCascadeRemovesRetiredRows(t *testing.T) {
	db := setupRetiredTablesDB(t)
	for _, name := range []string{"leaving", "staying"} {
		if err := db.Create(&models.User{Username: name, Email: name + "@example.test"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	coin := &models.Coin{Name: "Aureus", Category: models.CategoryRoman, UserID: 1}
	if err := NewCoinRepository(db).Create(coin); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"INSERT INTO auction_events (id, user_id, title) VALUES (1, 1, 'Show')",
		"INSERT INTO auction_lots (id, user_id, coin_id, event_id, title) VALUES (1, 1, ?, 1, 'Lot')",
		"INSERT INTO price_alerts (user_id, auction_lot_id) VALUES (1, 1)",
		"INSERT INTO bid_reminders (user_id, auction_lot_id) VALUES (1, 1)",
		"INSERT INTO shipments (id, user_id, coin_id) VALUES (1, 1, ?)",
		"INSERT INTO shipment_events (user_id, shipment_id) VALUES (1, 1)",
		"INSERT INTO notes (user_id) VALUES (1), (2)",
		"INSERT INTO value_snapshots (user_id) VALUES (1), (2)",
		"INSERT INTO auction_lots (id, user_id, title) VALUES (2, 2, 'Kept lot')",
	} {
		var args []interface{}
		if strings.Contains(sql, "?") {
			args = []interface{}{coin.ID}
		}
		if err := db.Exec(sql, args...).Error; err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	if rows, err := NewAdminRepository(db).DeleteUserCascade(1); err != nil || rows != 1 {
		t.Fatalf("DeleteUserCascade: rows=%d err=%v", rows, err)
	}

	for table, want := range map[string]int64{
		"auction_events": 0, "auction_lots": 1, "price_alerts": 0, "bid_reminders": 0,
		"shipments": 0, "shipment_events": 0, "notes": 1, "value_snapshots": 1, "coins": 0,
	} {
		if got := retiredCount(t, db, "SELECT COUNT(*) FROM "+table); got != want {
			t.Fatalf("%s rows=%d, want %d", table, got, want)
		}
	}
	if kept := retiredCount(t, db, "SELECT COUNT(*) FROM auction_lots WHERE user_id = 2"); kept != 1 {
		t.Fatal("the other user's lot was removed")
	}
}
