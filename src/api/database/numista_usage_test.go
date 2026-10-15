package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// This month's Numista count starts from the answers cached since its first
// day, once: not the mints' pages kept in the same cache, nor older answers.
func TestSeedNumistaUsage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "usage.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}, &models.NumistaUsage{}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	for key, at := range map[string]time.Time{
		"search:abc":                     now.Add(-time.Hour),
		"type:1:pt":                      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		"prices:1:2:EUR":                 now,
		"issuer-euro:v1:autriche:5:en:1": now,
		"incm:coins:v4":                  now,
		"type:2:en":                      time.Date(2026, 9, 30, 23, 0, 0, 0, time.UTC),
	} {
		if err := db.Create(&models.NumistaCacheEntry{Key: key, Payload: "{}", FetchedAt: at}).Error; err != nil {
			t.Fatal(err)
		}
	}
	for run := 1; run <= 2; run++ {
		if err := seedNumistaUsage(db, now); err != nil {
			t.Fatal(err)
		}
	}
	var usage []models.NumistaUsage
	db.Find(&usage)
	if len(usage) != 1 || usage[0].Month != "2026-10" || usage[0].Requests != 4 {
		t.Fatalf("usage = %+v", usage)
	}
}
