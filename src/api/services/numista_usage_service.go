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
}

// NewNumistaUsageService creates the service; settings and logger may be nil.
func NewNumistaUsageService(repo *repository.NumistaUsageRepository, settings *SettingsService, logger *Logger) *NumistaUsageService {
	return &NumistaUsageService{repo: repo, settings: settings, logger: logger, now: time.Now}
}

// Record counts a request made now.
func (s *NumistaUsageService) Record() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	if err := s.repo.Add(now.Format("2006-01"), 1, now); err != nil && s.logger != nil {
		s.logger.Warn("numista", "Could not count a Numista request: %v", err)
	}
}

// Usage is this month's count against the quota.
func (s *NumistaUsageService) Usage() models.NumistaUsageSummary {
	month := s.now().UTC().Format("2006-01")
	requests, err := s.repo.Requests(month)
	if err != nil && s.logger != nil {
		s.logger.Warn("numista", "Could not read the Numista request count: %v", err)
	}
	limit := s.limit()
	pauseAt := limit * numistaBackgroundShare / 100
	return models.NumistaUsageSummary{Month: month, Requests: requests, Limit: limit, PauseAt: pauseAt, BackgroundPaused: requests >= pauseAt}
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
