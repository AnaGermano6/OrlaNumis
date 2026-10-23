package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// updateReleasesURL is the project's latest published release (drafts
	// and pre-releases are left out by GitHub).
	updateReleasesURL = "https://api.github.com/repos/AnaGermano6/OrlaNumis/releases/latest"
	// UpdateImage is the published image Watchtower is asked to update.
	UpdateImage = "anagermano/orlanumis"
	// updateCheckTTL is how long a check is kept: GitHub allows 60
	// unauthenticated requests an hour.
	updateCheckTTL = 6 * time.Hour
	// updateNotesLimit bounds the release notes kept and shown.
	updateNotesLimit = 64 << 10
)

// Settings of updates: whether to look for a new version (GitHub), and the
// Watchtower that can update the container (its HTTP API address and token).
const (
	SettingUpdateCheckEnabled = "UpdateCheckEnabled"
	SettingWatchtowerURL      = "WatchtowerURL"
	SettingWatchtowerToken    = "WatchtowerToken"
)

// Update statuses.
const (
	UpdateStatusOK          = "ok"
	UpdateStatusDisabled    = "disabled"
	UpdateStatusUnavailable = "unavailable"
)

var (
	// ErrUpdateNotConfigured is returned when Watchtower is not set up.
	ErrUpdateNotConfigured = errors.New("watchtower is not configured")
	// ErrUpdateRefused is returned when Watchtower refuses the update.
	ErrUpdateRefused = errors.New("watchtower refused the update")
)

// UpdateStatus says which version runs, which is the latest and what it
// brings, and whether the app can start the update itself (Watchtower).
type UpdateStatus struct {
	Current string `json:"current" example:"1.0.0"`
	Latest  string `json:"latest,omitempty" example:"1.1.0"`
	// UpdateAvailable is true when the latest release is newer than the
	// running version.
	UpdateAvailable bool `json:"updateAvailable"`
	// Notes are the release's notes, in Markdown.
	Notes       string     `json:"notes,omitempty"`
	ReleaseURL  string     `json:"releaseUrl,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	CheckedAt   *time.Time `json:"checkedAt,omitempty"`
	Status      string     `json:"status" enums:"ok,disabled,unavailable"`
	// Watchtower is true when Watchtower's address and token are set, so
	// the update can be started from the app.
	Watchtower bool   `json:"watchtower"`
	Image      string `json:"image" example:"anagermano/orlanumis"`
}

// UpdateService tells whether a newer Orla Numis is out, from the project's
// GitHub releases, and starts an update through Watchtower when one is set
// up: a backup first, then Watchtower pulls the new image and recreates the
// container (the app restarts and migrates its data).
type UpdateService struct {
	current  string
	settings *SettingsService
	backups  *BackupService
	client   *http.Client
	logger   *Logger
	now      func() time.Time
	releases string

	mu        sync.Mutex
	last      *UpdateStatus
	lastCheck time.Time
}

// NewUpdateService creates the service; backups and logger may be nil.
func NewUpdateService(current string, settings *SettingsService, backups *BackupService, logger *Logger) *UpdateService {
	return &UpdateService{
		current: strings.TrimSpace(current), settings: settings, backups: backups, logger: logger,
		client: &http.Client{Timeout: 10 * time.Second}, now: time.Now, releases: updateReleasesURL,
	}
}

// Status is the update status, from a check made in the last hours unless
// force asks for a new one.
func (s *UpdateService) Status(ctx context.Context, force bool) UpdateStatus {
	base := UpdateStatus{Current: s.current, Watchtower: s.watchtowerConfigured(), Image: UpdateImage}
	if strings.EqualFold(strings.TrimSpace(s.settings.GetSetting(SettingUpdateCheckEnabled)), "false") {
		base.Status = UpdateStatusDisabled
		return base
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !force && s.last != nil && s.now().Sub(s.lastCheck) < updateCheckTTL {
		status := *s.last
		status.Watchtower = base.Watchtower
		return status
	}
	status, err := s.check(ctx, base)
	if err != nil {
		s.warn("Could not check for a new version: %v", err)
		// The last good check is kept, marked as not refreshed.
		if s.last != nil {
			kept := *s.last
			kept.Status, kept.Watchtower = UpdateStatusUnavailable, base.Watchtower
			return kept
		}
		base.Status = UpdateStatusUnavailable
		return base
	}
	s.last, s.lastCheck = &status, s.now()
	return status
}

type githubRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

func (s *UpdateService) check(ctx context.Context, status UpdateStatus) (UpdateStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.releases, nil)
	if err != nil {
		return status, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "OrlaNumis/"+s.current)
	resp, err := s.client.Do(req)
	if err != nil {
		return status, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return status, fmt.Errorf("GitHub answered %d", resp.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return status, err
	}
	if release.Draft || release.Prerelease || strings.TrimSpace(release.TagName) == "" {
		return status, errors.New("no published release")
	}
	now := s.now()
	status.Status = UpdateStatusOK
	status.CheckedAt = &now
	status.Latest = strings.TrimPrefix(strings.TrimSpace(release.TagName), "v")
	status.UpdateAvailable = NewerVersion(status.Latest, s.current)
	status.ReleaseURL = httpsOnly(release.HTMLURL)
	if notes := strings.TrimSpace(release.Body); len(notes) > updateNotesLimit {
		status.Notes = notes[:updateNotesLimit]
	} else {
		status.Notes = notes
	}
	if !release.PublishedAt.IsZero() {
		published := release.PublishedAt
		status.PublishedAt = &published
	}
	return status, nil
}

// Apply makes a backup, then asks Watchtower to update the app's image; the
// container restarts on the new version moments later.
func (s *UpdateService) Apply(ctx context.Context) error {
	address, token := s.watchtower()
	if address == "" || token == "" {
		return ErrUpdateNotConfigured
	}
	if s.backups != nil {
		if _, err := s.backups.Create(BackupKindPreUpdate); err != nil {
			return fmt.Errorf("backup before updating: %w", err)
		}
	}
	endpoint, err := url.Parse(strings.TrimRight(address, "/") + "/v1/update")
	if err != nil {
		return ErrUpdateNotConfigured
	}
	query := endpoint.Query()
	query.Set("image", UpdateImage)
	query.Set("async", "true")
	endpoint.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("reach watchtower: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusAccepted:
		s.info("Asked Watchtower to update %s", UpdateImage)
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w: the token was not accepted", ErrUpdateRefused)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: an update is already running", ErrUpdateRefused)
	default:
		return fmt.Errorf("%w: it answered %d", ErrUpdateRefused, resp.StatusCode)
	}
}

func (s *UpdateService) watchtower() (address, token string) {
	address = strings.TrimSpace(s.settings.GetSetting(SettingWatchtowerURL))
	token = strings.TrimSpace(s.settings.GetSetting(SettingWatchtowerToken))
	if parsed, err := url.Parse(address); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		address = ""
	}
	return address, token
}

func (s *UpdateService) watchtowerConfigured() bool {
	address, token := s.watchtower()
	return address != "" && token != ""
}

// NewerVersion tells whether version a (1.2.3, v1.2.3) is newer than b. A
// version that is not three numbers is never newer, nor older.
func NewerVersion(a, b string) bool {
	left, ok := parseVersion(a)
	right, okRight := parseVersion(b)
	if !ok || !okRight {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return left[i] > right[i]
		}
	}
	return false
}

func parseVersion(version string) ([3]int, bool) {
	var parts [3]int
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	version, _, _ = strings.Cut(version, "-")
	version, _, _ = strings.Cut(version, "+")
	fields := strings.Split(version, ".")
	if len(fields) != 3 {
		return parts, false
	}
	for i, field := range fields {
		n, err := strconv.Atoi(field)
		if err != nil || n < 0 {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

func (s *UpdateService) warn(format string, args ...any) {
	if s.logger != nil {
		s.logger.Warn("updates", format, args...)
	}
}

func (s *UpdateService) info(format string, args ...any) {
	if s.logger != nil {
		s.logger.Info("updates", format, args...)
	}
}
