package services

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewerVersion(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"1.1.0", "1.0.0", true},
		{"v1.0.1", "1.0.0", true},
		{"2.0.0", "1.9.9", true},
		{"1.10.0", "1.9.0", true},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.1.0", false},
		{"1.1.0", "1.1.0-rc.1", false},
		{"1.1.0", "dev", false},
		{"latest", "1.0.0", false},
	} {
		if got := NewerVersion(tc.a, tc.b); got != tc.want {
			t.Errorf("NewerVersion(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// The latest release is asked of GitHub once in a while, its notes kept; a
// failed check keeps the last good one; the check can be turned off.
func TestUpdateService_Status(t *testing.T) {
	settings, _ := newTestSettingsService(t)
	calls, fail := 0, false
	github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if fail {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "OrlaNumis/") {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte(`{"tag_name": "v1.1.0", "html_url": "https://github.com/AnaGermano6/OrlaNumis/releases/tag/v1.1.0",
			"body": "### Collection\n- Periods", "published_at": "2026-11-02T10:00:00Z"}`))
	}))
	t.Cleanup(github.Close)
	now := time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC)
	updates := NewUpdateService("1.0.0", settings, nil, nil)
	updates.releases, updates.now = github.URL, func() time.Time { return now }

	got := updates.Status(context.Background(), false)
	if got.Status != UpdateStatusOK || got.Latest != "1.1.0" || !got.UpdateAvailable || got.Notes != "### Collection\n- Periods" ||
		got.ReleaseURL == "" || got.PublishedAt == nil || got.Watchtower || got.Image != "anagermano/orlanumis" {
		t.Fatalf("status = %+v", got)
	}
	updates.Status(context.Background(), false)
	if calls != 1 {
		t.Fatalf("a second look within hours asked GitHub again: %d calls", calls)
	}
	fail = true
	if got := updates.Status(context.Background(), true); got.Status != UpdateStatusUnavailable || got.Latest != "1.1.0" || calls != 2 {
		t.Fatalf("a failed check keeps the last good one: %+v (%d calls)", got, calls)
	}

	if err := settings.SetSetting(SettingUpdateCheckEnabled, "false"); err != nil {
		t.Fatal(err)
	}
	if got := updates.Status(context.Background(), true); got.Status != UpdateStatusDisabled || got.Latest != "" || calls != 2 {
		t.Fatalf("turned off: %+v (%d calls)", got, calls)
	}
}

// Updating asks Watchtower, with its token, for the app's image only and
// without waiting; it needs Watchtower's address and token.
func TestUpdateService_ApplyAsksWatchtower(t *testing.T) {
	settings, _ := newTestSettingsService(t)
	updates := NewUpdateService("1.0.0", settings, nil, nil)
	if err := updates.Apply(context.Background()); !errors.Is(err, ErrUpdateNotConfigured) {
		t.Fatalf("not configured: %v", err)
	}

	var got *http.Request
	status := http.StatusAccepted
	watchtower := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.WriteHeader(status)
	}))
	t.Cleanup(watchtower.Close)
	for key, value := range map[string]string{SettingWatchtowerURL: watchtower.URL + "/", SettingWatchtowerToken: "secret-token"} {
		if err := settings.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if !updates.Status(context.Background(), false).Watchtower {
		t.Fatal("watchtower is configured")
	}
	if err := updates.Apply(context.Background()); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if got.Method != http.MethodPost || got.URL.Path != "/v1/update" || got.URL.Query().Get("image") != "anagermano/orlanumis" ||
		got.URL.Query().Get("async") != "true" || got.Header.Get("Authorization") != "Bearer secret-token" {
		t.Fatalf("request = %s %s %v", got.Method, got.URL, got.Header)
	}
	status = http.StatusUnauthorized
	if err := updates.Apply(context.Background()); !errors.Is(err, ErrUpdateRefused) {
		t.Fatalf("a refused token: %v", err)
	}
	if err := settings.SetSetting(SettingWatchtowerURL, "ftp://watchtower"); err != nil {
		t.Fatal(err)
	}
	if err := updates.Apply(context.Background()); !errors.Is(err, ErrUpdateNotConfigured) {
		t.Fatalf("an address that is not http: %v", err)
	}
}

// A backup is made before Watchtower is asked, and listed like the others,
// so the data of before the update can be restored.
func TestUpdateService_ApplyBacksUpFirst(t *testing.T) {
	test := newBackupTest(t)
	watchtower := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(watchtower.Close)
	for key, value := range map[string]string{SettingWatchtowerURL: watchtower.URL, SettingWatchtowerToken: "secret-token"} {
		if err := test.settings.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	updates := NewUpdateService("1.0.0", test.settings, test.backups, nil)
	if err := updates.Apply(context.Background()); err != nil {
		t.Fatalf("apply: %v", err)
	}
	backups, err := test.backups.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 || backups[0].Kind != BackupKindPreUpdate {
		t.Fatalf("backups = %+v", backups)
	}
}
