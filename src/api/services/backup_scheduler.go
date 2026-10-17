package services

import (
	"sync"
	"time"
)

// backupRetryAfterFailure spaces out the tries of a backup that failed.
const backupRetryAfterFailure = time.Hour

// BackupScheduler makes an automatic full backup every day at the start
// time, and at once when the last one is older than the day's (the app was
// off, or the computer asleep, at that time).
type BackupScheduler struct {
	backups  *BackupService
	settings *SettingsService
	logger   *Logger
	stopCh   chan struct{}
	once     sync.Once

	mu          sync.Mutex
	lastAttempt time.Time
}

func NewBackupScheduler(backups *BackupService, settings *SettingsService, logger *Logger) *BackupScheduler {
	return &BackupScheduler{backups: backups, settings: settings, logger: logger, stopCh: make(chan struct{})}
}

// Start begins the daily loop. Call from a goroutine.
func (s *BackupScheduler) Start() {
	s.logger.Info("backup", "Backup scheduler started")
	scheduleLoop{
		category:     "backup",
		name:         "automatic backup",
		logger:       s.logger,
		stopCh:       s.stopCh,
		initialDelay: 2 * time.Minute,
		next:         func(now time.Time, _ bool) time.Time { return s.nextRun(now) },
		run:          s.runCycle,
	}.Run()
}

// Stop signals the scheduler to shut down. Safe to call more than once.
func (s *BackupScheduler) Stop() {
	s.once.Do(func() { close(s.stopCh) })
}

// RunNow makes an automatic backup now.
func (s *BackupScheduler) RunNow() error {
	_, err := s.backups.Create(BackupKindAutomatic)
	return err
}

func (s *BackupScheduler) enabled() bool {
	return s.settings.GetSetting(SettingBackupEnabled) == "true"
}

// GetStatus returns the scheduler runtime status.
func (s *BackupScheduler) GetStatus() SchedulerStatus {
	return SchedulerStatus{Name: "Automatic Backup", Enabled: s.enabled(), IsRunning: s.backups.Running(), NextRunIn: s.timeUntilNextRun()}
}

func (s *BackupScheduler) timeUntilNextRun() time.Duration {
	now := time.Now()
	return s.nextRun(now).Sub(now)
}

func (s *BackupScheduler) schedule() dailySchedule {
	h, m := parseStartTime(s.settings.GetSetting(SettingBackupStartTime), 3, 0)
	return dailySchedule{Hour: h, Minute: m, Location: scheduleLocation(s.settings, s.logger)}
}

// nextRun is the day's start time after the last automatic backup; now when
// that time has passed (or there was none yet), but not sooner than an hour
// after a try that failed.
func (s *BackupScheduler) nextRun(now time.Time) time.Time {
	schedule := s.schedule()
	if !s.enabled() {
		return schedule.slotAfter(now)
	}
	due := now
	if last := s.backups.LastAutomatic(); last != nil {
		due = schedule.slotAfter(*last)
	}
	if due.After(now) {
		return due
	}
	s.mu.Lock()
	retry := s.lastAttempt.Add(backupRetryAfterFailure)
	s.mu.Unlock()
	if retry.After(now) {
		return retry
	}
	return now
}

func (s *BackupScheduler) runCycle() {
	if !s.enabled() {
		return
	}
	s.mu.Lock()
	s.lastAttempt = time.Now()
	s.mu.Unlock()
	if _, err := s.backups.Create(BackupKindAutomatic); err != nil && err != ErrBackupRunning {
		s.logger.Error("backup", "Automatic backup failed: %v", err)
	}
}
