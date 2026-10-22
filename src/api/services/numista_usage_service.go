package services

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

const (
	// defaultNumistaMonthlyLimit is the free plan's monthly quota.
	defaultNumistaMonthlyLimit = 2000
	// numistaBackgroundShare is the share of the quota (in percent) requests
	// made in the background may use; the rest is kept for lookups by hand.
	numistaBackgroundShare = 90
	// After Numista refuses a request for its quota, no request is sent for
	// as long as it says, or numistaRefusalPause when it does not say, and at
	// most numistaMaxRefusalPause.
	numistaRefusalPause    = 15 * time.Minute
	numistaMaxRefusalPause = 24 * time.Hour
)

// NumistaUsageService counts Numista's API requests month by month (UTC) and
// tells when to stop asking in the background.
type NumistaUsageService struct {
	repo     *repository.NumistaUsageRepository
	settings *SettingsService
	logger   *Logger
	now      func() time.Time
	// mu keeps the count of requests made at once in order.
	mu sync.Mutex
	// refusingUntil is when Numista may take requests again after refusing
	// one for its quota; kept in memory, as the refusal is short.
	refusingUntil time.Time
}

// NewNumistaUsageService creates the service; settings and logger may be nil.
func NewNumistaUsageService(repo *repository.NumistaUsageRepository, settings *SettingsService, logger *Logger) *NumistaUsageService {
	return &NumistaUsageService{repo: repo, settings: settings, logger: logger, now: time.Now}
}

// Record counts a request made now, of a kind (models.NumistaRequest…),
// made in the background or by hand.
func (s *NumistaUsageService) Record(kind string, background bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	if err := s.repo.Add(now.Format("2006-01"), kind, background, now); err != nil && s.logger != nil {
		s.logger.Warn("numista", "Could not count a Numista request: %v", err)
	}
}

// Refused notes that Numista refused a request for its quota, and said to
// wait retryAfter (0 when it did not say): until then no request is sent,
// so a page opened again or a batch does not ask in vain.
func (s *NumistaUsageService) Refused(retryAfter time.Duration) {
	if retryAfter <= 0 {
		retryAfter = numistaRefusalPause
	}
	retryAfter = min(retryAfter, numistaMaxRefusalPause)
	s.mu.Lock()
	defer s.mu.Unlock()
	if until := s.now().Add(retryAfter); until.After(s.refusingUntil) {
		s.refusingUntil = until
	}
	if s.logger != nil {
		s.logger.Warn("numista", "Numista refused a request for its quota; none is sent until %s", s.refusingUntil.UTC().Format(time.RFC3339))
	}
}

// RefusingUntil is when requests may be sent again after Numista refused
// one for its quota; the zero time when they may now.
func (s *NumistaUsageService) RefusingUntil() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.now().Before(s.refusingUntil) {
		return time.Time{}
	}
	return s.refusingUntil
}

// Usage is this month's count against the quota.
func (s *NumistaUsageService) Usage() models.NumistaUsageSummary {
	month := s.now().UTC().Format("2006-01")
	usage, err := s.repo.Month(month)
	if err != nil && s.logger != nil {
		s.logger.Warn("numista", "Could not read the Numista request count: %v", err)
	}
	limit := s.limit()
	pauseAt := limit * numistaBackgroundShare / 100
	return models.NumistaUsageSummary{
		Month: month, Requests: usage.Requests, Limit: limit, PauseAt: pauseAt, BackgroundPaused: usage.Requests >= pauseAt,
		Searches: usage.Searches, Types: usage.Types, Issues: usage.Issues, Prices: usage.Prices, Other: usage.Other, Background: usage.Background,
		PausedUntil: s.pausedUntil(),
	}
}

func (s *NumistaUsageService) pausedUntil() *time.Time {
	if until := s.RefusingUntil(); !until.IsZero() {
		until = until.UTC()
		return &until
	}
	return nil
}

// BackgroundPaused tells whether requests made in the background wait for next month.
func (s *NumistaUsageService) BackgroundPaused() bool {
	return s.Usage().BackgroundPaused
}

func (s *NumistaUsageService) limit() int {
	if s.settings == nil {
		return defaultNumistaMonthlyLimit
	}
	limit, err := strconv.Atoi(strings.TrimSpace(s.settings.GetSetting(SettingNumistaMonthlyLimit)))
	if err != nil || limit <= 0 {
		return defaultNumistaMonthlyLimit
	}
	return limit
}
