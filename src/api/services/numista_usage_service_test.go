package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	for range 5 {
		usage.Record(models.NumistaRequestPrices, true)
	}
	usage.Record(models.NumistaRequestSearch, false)
	usage.Record(models.NumistaRequestType, false)
	usage.Record("issuers", false)
	if got := usage.Usage(); got.Month != "2026-10" || got.Requests != 8 || got.Limit != 10 || got.PauseAt != 9 || got.BackgroundPaused {
		t.Fatalf("usage = %+v", got)
	}
	// Each request is counted by what it asked for, and apart when made in the background.
	if got := usage.Usage(); got.Prices != 5 || got.Searches != 1 || got.Types != 1 || got.Issues != 0 || got.Other != 1 || got.Background != 5 {
		t.Fatalf("by kind = %+v", got)
	}
	usage.Record(models.NumistaRequestIssues, false)
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
	kinds    []string
}

func (u *fixedNumistaUsage) Record(kind string, background bool) {
	u.recorded++
	if background {
		kind += " (background)"
	}
	u.kinds = append(u.kinds, kind)
}
func (u *fixedNumistaUsage) BackgroundPaused() bool   { return u.paused }
func (u *fixedNumistaUsage) Refused(time.Duration)    {}
func (u *fixedNumistaUsage) RefusingUntil() time.Time { return time.Time{} }

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
	if strings.Join(usage.kinds, ", ") != "type, type (background)" {
		t.Fatalf("kinds = %v", usage.kinds)
	}
}

func TestNumistaRequestKind(t *testing.T) {
	for path, want := range map[string]string{
		"/types?q=2+euro&lang=pt":                      models.NumistaRequestSearch,
		"/types/2933?lang=pt":                          models.NumistaRequestType,
		"/types/2933":                                  models.NumistaRequestType,
		"/types/2933/issues":                           models.NumistaRequestIssues,
		"/types/2933/issues/14184/prices?currency=EUR": models.NumistaRequestPrices,
		"/issuers?lang=pt":                             models.NumistaRequestOther,
		"/types/2933/issues/14184":                     models.NumistaRequestOther,
	} {
		if got := numistaRequestKind(path); got != want {
			t.Errorf("numistaRequestKind(%q) = %q, want %q", path, got, want)
		}
	}
}

// Once Numista refuses a request for its quota, no request is sent (nor
// counted) for as long as it said, by hand or in the background; then they
// go again.
func TestHTTPNumistaClient_WaitsAfterAQuotaRefusal(t *testing.T) {
	usage, _ := newNumistaUsageTest(t)
	now := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	usage.now = func() time.Time { return now }
	requests, refuse := 0, true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		if refuse {
			w.Header().Set("Retry-After", "600")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"id": 1, "title": "2 Euros"}`))
	}))
	t.Cleanup(server.Close)
	client, err := NewHTTPNumistaClient(NumistaClientConfig{BaseURL: server.URL, APIKey: func() string { return "key" }})
	if err != nil {
		t.Fatal(err)
	}
	client.WithUsage(usage)

	var numistaErr *NumistaError
	if _, err := client.TypeDetail(context.Background(), 1, "en"); !errors.As(err, &numistaErr) || numistaErr.Kind != NumistaErrorQuotaLimited || requests != 1 {
		t.Fatalf("refused: err %v, requests %d", err, requests)
	}
	if got := usage.Usage(); got.Requests != 1 || got.PausedUntil == nil || !got.PausedUntil.Equal(now.Add(10*time.Minute)) {
		t.Fatalf("after the refusal = %+v", got)
	}
	refuse = false
	now = now.Add(9 * time.Minute)
	if _, err := client.TypeDetail(context.Background(), 1, "en"); !errors.As(err, &numistaErr) || numistaErr.Kind != NumistaErrorQuotaLimited || numistaErr.RetryAfterSeconds == nil || requests != 1 {
		t.Fatalf("while waiting: err %v, requests %d", err, requests)
	}
	if got := usage.Usage(); got.Requests != 1 {
		t.Fatalf("a request not sent is not counted: %+v", got)
	}
	now = now.Add(2 * time.Minute)
	if _, err := client.TypeDetail(context.Background(), 1, "en"); err != nil || requests != 2 || usage.Usage().PausedUntil != nil {
		t.Fatalf("after the wait: err %v, requests %d", err, requests)
	}

	// Without a Retry-After, the wait is a quarter of an hour; never more than a day.
	usage.Refused(0)
	if got := usage.RefusingUntil(); !got.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("default wait until %v", got)
	}
	usage.Refused(72 * time.Hour)
	if got := usage.RefusingUntil(); !got.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("longest wait until %v", got)
	}
}
