package database

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// An older database still has the tables of removed features, with rows and
// foreign keys to users, coins and each other. With foreign keys on (as
// Connect runs), they are dropped, and nothing else is.
func TestDropRetiredTablesDropsOnlyThemWithTheirRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	// One connection, so the pragma applies to every statement below.
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Coin{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.User{Username: "collector", Email: "collector@example.test"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Coin{Name: "Denarius", UserID: 1}).Error; err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		"CREATE TABLE auction_events (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id))",
		"CREATE TABLE auction_lots (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), coin_id INTEGER REFERENCES coins(id), event_id INTEGER REFERENCES auction_events(id))",
		"CREATE TABLE price_alerts (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), auction_lot_id INTEGER NOT NULL REFERENCES auction_lots(id))",
		"CREATE TABLE bid_reminders (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), auction_lot_id INTEGER NOT NULL REFERENCES auction_lots(id))",
		"CREATE TABLE auction_alert_runs (id INTEGER PRIMARY KEY)",
		"CREATE TABLE auction_ending_runs (id INTEGER PRIMARY KEY)",
		"CREATE TABLE auction_watch_bid_digest_runs (id INTEGER PRIMARY KEY)",
		"CREATE TABLE shipments (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), coin_id INTEGER NOT NULL REFERENCES coins(id))",
		"CREATE TABLE shipment_events (id INTEGER PRIMARY KEY, user_id INTEGER NOT NULL REFERENCES users(id), shipment_id INTEGER NOT NULL REFERENCES shipments(id))",
		"CREATE TABLE notes (id INTEGER PRIMARY KEY, user_id INTEGER)",
		"CREATE TABLE value_snapshots (id INTEGER PRIMARY KEY, user_id INTEGER)",
		"INSERT INTO auction_events (id, user_id) VALUES (1, 1)",
		"INSERT INTO auction_lots (id, user_id, coin_id, event_id) VALUES (1, 1, 1, 1)",
		"INSERT INTO price_alerts (user_id, auction_lot_id) VALUES (1, 1)",
		"INSERT INTO bid_reminders (user_id, auction_lot_id) VALUES (1, 1)",
		"INSERT INTO auction_alert_runs DEFAULT VALUES",
		"INSERT INTO shipments (id, user_id, coin_id) VALUES (1, 1, 1)",
		"INSERT INTO shipment_events (user_id, shipment_id) VALUES (1, 1)",
		"INSERT INTO notes (user_id) VALUES (1)",
		"INSERT INTO value_snapshots (user_id) VALUES (1), (1)",
	} {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}

	// A second run (every later start) finds nothing to drop.
	for run := 1; run <= 2; run++ {
		if err := dropRetiredTables(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	for _, table := range retiredTables {
		if db.Migrator().HasTable(table) {
			t.Fatalf("%s was not dropped", table)
		}
	}
	var users, coins int64
	db.Model(&models.User{}).Count(&users)
	db.Model(&models.Coin{}).Count(&coins)
	if users != 1 || coins != 1 {
		t.Fatalf("users=%d coins=%d, want both kept", users, coins)
	}
}

// Every retired table is one no model migrates any more.
func TestRetiredTablesHaveNoModel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fresh.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	stmt := &gorm.Statement{DB: db}
	for name, newModel := range productionModelConstructors {
		if err := stmt.Parse(newModel()); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, table := range retiredTables {
			if stmt.Schema.Table == table {
				t.Fatalf("model %s still uses the retired table %s", name, table)
			}
		}
	}
	if err := stmt.Parse(&models.User{}); err != nil {
		t.Fatal(err)
	}
	for _, column := range retiredUserColumns {
		if stmt.Schema.LookUpField(column) != nil {
			t.Fatalf("the user model still has the retired column %s", column)
		}
	}
}

// An older database's users still have the integration columns; they are
// dropped, with foreign keys on, and every other column keeps its value.
func TestDropRetiredUserColumnsKeepsTheRest(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Coin{}); err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{
		"numis_bids_username varchar(100)", "numis_bids_password text", "cng_username varchar(100)", "cng_password text",
		"parcel_app_api_key text", "pushover_user_key varchar(100)", "pushover_enabled numeric DEFAULT false",
		"is_public numeric DEFAULT false",
	} {
		if err := db.Exec("ALTER TABLE users ADD COLUMN " + column).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.User{Username: "collector", Email: "collector@example.test", Bio: "Roman bronzes"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE users SET cng_username = 'old', pushover_enabled = true").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Coin{Name: "Follis", UserID: 1}).Error; err != nil {
		t.Fatal(err)
	}

	// A second run (every later start) finds nothing to drop.
	for run := 1; run <= 2; run++ {
		if err := dropRetiredUserColumns(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	for _, column := range retiredUserColumns {
		if db.Migrator().HasColumn("users", column) {
			t.Fatalf("users.%s was not dropped", column)
		}
	}
	var user models.User
	if err := db.First(&user).Error; err != nil {
		t.Fatal(err)
	}
	if user.Username != "collector" || user.Email != "collector@example.test" || user.Bio != "Roman bronzes" {
		t.Fatalf("user=%+v, want its other columns kept", user)
	}
	var coins int64
	db.Model(&models.Coin{}).Where("user_id = ?", user.ID).Count(&coins)
	if coins != 1 {
		t.Fatalf("coins=%d, want the user's coin kept", coins)
	}
}

// An older database's coins still have the storage location (with its slot
// index), the sale and the private mark, and its sets the public mark; they
// are dropped, with foreign keys on, and every other column keeps its value.
// Smart sets lose the rules on those columns.
func TestDropRetiredCoinColumnsAndCriteriaRules(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys=ON").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Coin{}, &models.CoinSet{}); err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{
		"is_sold numeric DEFAULT false", "sold_price real", "sold_date datetime", "sold_to text",
		"storage_location_id integer", "storage_slot integer", "is_private numeric DEFAULT false",
	} {
		if err := db.Exec("ALTER TABLE coins ADD COLUMN " + column).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("ALTER TABLE coin_sets ADD COLUMN is_public numeric DEFAULT false").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX idx_coins_storage_location_slot_unique
		ON coins(storage_location_id, storage_slot) WHERE storage_slot IS NOT NULL`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.User{Username: "collector", Email: "collector@example.test"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Coin{Name: "2 euro 2022", UserID: 1, Notes: "kept"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE coins SET storage_location_id = 3, storage_slot = 2, sold_to = 'old'").Error; err != nil {
		t.Fatal(err)
	}
	criteria := models.JSONObject{"operator": "and", "rules": []interface{}{
		map[string]interface{}{"field": "year", "op": "eq", "value": 1991},
		map[string]interface{}{"field": "isWishlist", "op": "eq", "value": false},
		map[string]interface{}{"field": "isSold", "op": "eq", "value": false},
		map[string]interface{}{"field": "isPrivate", "op": "eq", "value": true},
		map[string]interface{}{"operator": "or", "rules": []interface{}{
			map[string]interface{}{"field": "storageLocationId", "op": "eq", "value": 3},
		}},
	}}
	if err := db.Create(&models.CoinSet{UserID: 1, Name: "1991", SetType: models.CoinSetTypeSmart, SmartCriteria: &criteria}).Error; err != nil {
		t.Fatal(err)
	}

	// A second run (every later start) finds nothing to drop.
	for run := 1; run <= 2; run++ {
		if err := dropRetiredCoinColumns(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if err := dropRetiredSetColumns(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if err := dropRetiredCriteriaRules(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	for _, column := range retiredCoinColumns {
		if db.Migrator().HasColumn("coins", column) {
			t.Fatalf("coins.%s was not dropped", column)
		}
	}
	for _, column := range retiredSetColumns {
		if db.Migrator().HasColumn("coin_sets", column) {
			t.Fatalf("coin_sets.%s was not dropped", column)
		}
	}
	var coin models.Coin
	if err := db.First(&coin).Error; err != nil || coin.Name != "2 euro 2022" || coin.Notes != "kept" {
		t.Fatalf("coin = %+v, %v; want its other columns kept", coin, err)
	}
	var set models.CoinSet
	if err := db.First(&set).Error; err != nil {
		t.Fatal(err)
	}
	rules, _ := (*set.SmartCriteria)["rules"].([]interface{})
	fields := []string{}
	for _, raw := range rules {
		rule, _ := raw.(map[string]interface{})
		field, _ := rule["field"].(string)
		fields = append(fields, field)
	}
	if strings.Join(fields, ",") != "year,isWishlist" {
		t.Fatalf("rules left = %v (%v)", fields, rules)
	}
}

// The ancient categories leave the category list, its labels and the coins;
// the retired features' stored settings are deleted; other values stay.
func TestDropRetiredSettingsAndCategories(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&models.User{}, &models.Coin{}, &models.AppSetting{}); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]string{
		"CoinCategories":         "Roman\nGreek\nByzantine\nModern\nOther\nCeltic",
		"CoinCategoryLabelsPtPT": "Roman=Romana\nCeltic=Celta",
		"CoinOfDayEnabled":       "true",
		"FedExClientSecret":      "old-secret",
		"BackupEnabled":          "true",
	} {
		if err := db.Create(&models.AppSetting{Key: key, Value: value}).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.User{Username: "collector", Email: "collector@example.test"}).Error; err != nil {
		t.Fatal(err)
	}
	for _, category := range []models.Category{"Roman", models.CategoryModern} {
		if err := db.Create(&models.Coin{Name: string(category), Category: category, UserID: 1}).Error; err != nil {
			t.Fatal(err)
		}
	}

	// A second run (every later start) changes nothing more.
	for run := 1; run <= 2; run++ {
		if err := dropRetiredSettings(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if err := dropRetiredCategories(db); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	settings := map[string]string{}
	var rows []models.AppSetting
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		settings[row.Key] = row.Value
	}
	if settings["CoinCategories"] != "Modern\nOther\nCeltic" || settings["CoinCategoryLabelsPtPT"] != "Celtic=Celta" {
		t.Fatalf("categories = %q, labels = %q", settings["CoinCategories"], settings["CoinCategoryLabelsPtPT"])
	}
	for _, key := range []string{"CoinOfDayEnabled", "FedExClientSecret"} {
		if _, ok := settings[key]; ok {
			t.Fatalf("%s was not deleted", key)
		}
	}
	if settings["BackupEnabled"] != "true" {
		t.Fatalf("BackupEnabled = %q, want it kept", settings["BackupEnabled"])
	}
	var categories []string
	if err := db.Model(&models.Coin{}).Order("id").Pluck("category", &categories).Error; err != nil {
		t.Fatal(err)
	}
	if strings.Join(categories, ",") != "Other,Modern" {
		t.Fatalf("coin categories = %v", categories)
	}
}
