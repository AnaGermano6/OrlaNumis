package services

import (
	"context"
	"sync"
	"time"

	"github.com/briandenicola/ancient-coins-api/repository"
)

// NumistaValueScheduler runs the batch of coin values from Numista's prices
// once a day, at the configured time, while it is on.
type NumistaValueScheduler struct {
	svc      *NumistaValueService
	repo     *repository.NumistaValueRepository
	settings *SettingsService
	logger   *Logger
	stopCh   chan struct{}
	once     sync.Once
}

// NewNumistaValueScheduler creates the scheduler.
func NewNumistaValueScheduler(svc *NumistaValueService, repo *repository.NumistaValueRepository, settings *SettingsService, logger *Logger) *NumistaValueScheduler {
	return &NumistaValueScheduler{svc: svc, repo: repo, settings: settings, logger: logger, stopCh: make(chan struct{})}
}

// Start runs the daily loop. Call from a goroutine.
func (s *NumistaValueScheduler) Start() {
	// A batch the app stopped in the middle of is over.
	if err := s.repo.AbandonRunning(time.Now()); err != nil {
		s.logger.Warn("numista-values", "Could not close unfinished batches: %v", err)
	}
	scheduleLoop{
		category:     "numista-values",
		name:         "Numista values batch",
		logger:       s.logger,
		stopCh:       s.stopCh,
		initialDelay: 90 * time.Second,
		next:         s.nextRun,
		run:          s.runCycle,
	}.Run()
}

// Stop ends the loop. Safe to call more than once.
func (s *NumistaValueScheduler) Stop() {
	s.once.Do(func() { close(s.stopCh) })
}

// RunNow makes one batch now, on or off.
func (s *NumistaValueScheduler) RunNow() error {
	_, err := s.svc.Run(context.Background(), "manual")
	return err
}

// GetStatus returns the scheduler's runtime status.
func (s *NumistaValueScheduler) GetStatus() SchedulerStatus {
	return SchedulerStatus{
		Name:      "numista-values",
		Enabled:   s.svc.Enabled(),
		IsRunning: s.svc.IsRunning(),
		NextRunIn: s.timeUntilNextRun(),
	}
}

func (s *NumistaValueScheduler) timeUntilNextRun() time.Duration {
	now := time.Now()
	return max(s.nextRun(now, false).Sub(now), 0)
}

// nextRun is the start time on the day after the last scheduled batch, in the
// schedule's zone (so restarts do not reset the schedule).
func (s *NumistaValueScheduler) nextRun(now time.Time, catchUp bool) time.Time {
	var last *time.Time
	if run := s.repo.LastScheduledRun(); run != nil {
		last = run.CompletedAt
	}
	hour, minute := parseStartTime(s.settings.GetSetting(SettingNumistaValuesStartTime), 5, 0)
	schedule := dailySchedule{
		Hour:     hour,
		Minute:   minute,
		Interval: 24 * time.Hour,
		Location: scheduleLocation(s.settings, s.logger),
	}
	return schedule.next(now, last, catchUp)
}

func (s *NumistaValueScheduler) runCycle() {
	if !s.svc.Enabled() {
		s.logger.Debug("numista-values", "Numista values batch is off, skipping")
		return
	}
	if _, err := s.svc.Run(context.Background(), "scheduled"); err != nil {
		s.logger.Warn("numista-values", "Numista values batch did not run: %v", err)
	}
}
