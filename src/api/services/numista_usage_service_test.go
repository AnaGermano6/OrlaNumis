package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newNumistaUsageTest(t *testing.T) (*NumistaUsageService, *SettingsService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.NumistaUsage{}, &models.AppSetting{}); err != nil {
		t.Fatal(err)
	}
	settings := NewSettingsService(repository.NewSettingsRepository(db))
	return NewNumistaUsageService(repository.NewNumistaUsageRepository(db), settings, nil), settings
}

// Requests are counted by month against the quota; above 90% of it, the
// background ones wait for the next month.
func TestNumistaUsage_CountsByMonthAndPausesTheBackground(t *testing.T) {
	usage, settings := newNumistaUsageTest(t)
	now := time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)
	usage.now = func() time.Time { return now }
	if err := settings.SetSetting(SettingNumistaMonthlyLimit, "10"); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		usage.Record()
	}
	if got := usage.Usage(); got.Month != "2026-10" || got.Requests != 8 || got.Limit != 10 || got.PauseAt != 9 || got.BackgroundPaused {
		t.Fatalf("usage = %+v", got)
	}
	usage.Record()
	if !usage.BackgroundPaused() {
		t.Fatal("at 90% of the quota the background waits")
	}
	// A new month starts from nothing.
	now = now.Add(2 * time.Hour)
	if got := usage.Usage(); got.Month != "2026-11" || got.Requests != 0 || got.BackgroundPaused {
		t.Fatalf("next month = %+v", got)
	}
	// An unreadable limit is the free plan's.
	if err := settings.SetSetting(SettingNumistaMonthlyLimit, "lots"); err != nil {
		t.Fatal(err)
	}
	if got := usage.Usage(); got.Limit != 2000 || got.PauseAt != 1800 {
		t.Fatalf("default limit = %+v", got)
	}
}

type fixedNumistaUsage struct {
	paused   bool
	recorded int
}

func (u *fixedNumistaUsage) Record()                { u.recorded++ }
func (u *fixedNumistaUsage) BackgroundPaused() bool { return u.paused }

// The client counts every request it makes; a background one is not made
// while the background waits, and fails as Numista's quota would.
func TestHTTPNumistaClient_CountsRequestsAndHoldsTheBackground(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = w.Write([]byte(`{"id": 1, "title": "2 Euros"}`))
	}))
	t.Cleanup(server.Close)
	client, err := NewHTTPNumistaClient(NumistaClientConfig{BaseURL: server.URL, APIKey: func() string { return "key" }})
	if err != nil {
		t.Fatal(err)
	}
	usage := &fixedNumistaUsage{paused: true}
	client.WithUsage(usage)

	if _, err := client.TypeDetail(context.Background(), 1, "en"); err != nil || usage.recorded != 1 || requests != 1 {
		t.Fatalf("a lookup by hand: err %v, recorded %d, requests %d", err, usage.recorded, requests)
	}
	_, err = client.TypeDetail(WithNumistaBackground(context.Background()), 1, "en")
	var numistaErr *NumistaError
	if !errors.As(err, &numistaErr) || numistaErr.Kind != NumistaErrorQuotaLimited || usage.recorded != 1 || requests != 1 {
		t.Fatalf("a background request while paused: err %v, recorded %d, requests %d", err, usage.recorded, requests)
	}
	usage.paused = false
	if _, err := client.TypeDetail(WithNumistaBackground(context.Background()), 1, "en"); err != nil || usage.recorded != 2 || requests != 2 {
		t.Fatalf("a background request: err %v, recorded %d, requests %d", err, usage.recorded, requests)
	}
}
