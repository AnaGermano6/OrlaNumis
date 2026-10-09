package repository

import (
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestNumistaCacheRepository_SaveReplacesEntryForKey(t *testing.T) {
	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.NumistaCacheEntry{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := NewNumistaCacheRepository(db)

	missing, err := repo.Find("prices:1:2:EUR")
	if err != nil || missing != nil {
		t.Fatalf("Find on empty cache = %+v, %v; want nil, nil", missing, err)
	}

	first := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	second := first.Add(48 * time.Hour)
	if err := repo.Save(&models.NumistaCacheEntry{Key: "prices:1:2:EUR", Payload: `{"prices":[]}`, FetchedAt: first}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if err := repo.Save(&models.NumistaCacheEntry{Key: "prices:1:2:EUR", Payload: `{"prices":[{"grade":"f","price":1}]}`, FetchedAt: second}); err != nil {
		t.Fatalf("second save: %v", err)
	}

	got, err := repo.Find("prices:1:2:EUR")
	if err != nil || got == nil {
		t.Fatalf("Find = %+v, %v", got, err)
	}
	if got.Payload != `{"prices":[{"grade":"f","price":1}]}` || !got.FetchedAt.Equal(second) {
		t.Fatalf("entry was not replaced: %+v", got)
	}
	var count int64
	db.Model(&models.NumistaCacheEntry{}).Count(&count)
	if count != 1 {
		t.Fatalf("expected one row, got %d", count)
	}
}
