package services

import (
	"context"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

// Settings of the daily batch of coin values from Numista's prices: on or
// off, the time it starts, and the most Numista requests it makes in a day.
const (
	SettingNumistaValuesEnabled       = "NumistaValuesEnabled"
	SettingNumistaValuesStartTime     = "NumistaValuesStartTime"
	SettingNumistaValuesDailyRequests = "NumistaValuesDailyRequests"
)

const (
	// numistaValueRecheck is how long a coin's value stands before its price
	// is looked up again: Numista's prices are estimates that move slowly.
	numistaValueRecheck = 30 * 24 * time.Hour
	// numistaQuotaLowShare is the share of the monthly quota (in percent)
	// from which the quota counts as running low.
	numistaQuotaLowShare = 75
	// defaultNumistaValuesDailyRequests caps a day's batch.
	defaultNumistaValuesDailyRequests = 40
	// numistaValueMaxFailures in a row stop a batch: Numista is not answering.
	numistaValueMaxFailures = 3
	numistaValueRunsShown   = 10
)

// ErrNumistaValuesRunning is returned when a batch is already running.
var ErrNumistaValuesRunning = errors.New("a batch of Numista values is already running")

// numistaUsageReader tells how many Numista requests this month has made.
type numistaUsageReader interface {
	Usage() models.NumistaUsageSummary
}

// NumistaValueService keeps the collection's current values in step with
// Numista's estimated prices, a few coins a day: each batch makes at most the
// day's share of what is left of the month's quota (and never more than the
// daily cap), stops once it is spent, and waits for the next month once the
// background share of the quota is (see NumistaUsageService). A coin's price
// is that of its grade, so coins without a grade Numista prices wait for one.
// The new value replaces the current one, which stays in the value history.
type NumistaValueService struct {
	repo     *repository.NumistaValueRepository
	market   *NumistaMarketValueService
	usage    numistaUsageReader
	settings *SettingsService
	logger   *Logger
	now      func() time.Time
	running  atomic.Bool
}

// NewNumistaValueService creates the service; logger may be nil.
func NewNumistaValueService(repo *repository.NumistaValueRepository, market *NumistaMarketValueService, usage numistaUsageReader, settings *SettingsService, logger *Logger) *NumistaValueService {
	return &NumistaValueService{repo: repo, market: market, usage: usage, settings: settings, logger: logger, now: time.Now}
}

// IsRunning tells whether a batch is running.
func (s *NumistaValueService) IsRunning() bool {
	return s.running.Load()
}

// Enabled tells whether the daily batch is on.
func (s *NumistaValueService) Enabled() bool {
	return s.settings.GetSetting(SettingNumistaValuesEnabled) == "true"
}

type numistaValueItem struct {
	coin   models.Coin
	typeID int
	grade  string
	// checkedAt is when the coin was last checked; zero when never.
	checkedAt time.Time
}

type numistaValuePlan struct {
	queue        []numistaValueItem
	eligible     int
	withoutGrade int
}

// plan lists the coins due a check, never checked first, then the longest
// unchecked.
func (s *NumistaValueService) plan(now time.Time) (numistaValuePlan, error) {
	var plan numistaValuePlan
	coins, err := s.repo.LinkedCoins()
	if err != nil {
		return plan, err
	}
	checks, err := s.repo.Checks()
	if err != nil {
		return plan, err
	}
	for i := range coins {
		typeID := NumistaTypeIDForCoin(&coins[i])
		if typeID == 0 {
			continue
		}
		grade, ok := NumistaGradeForCoinGrade(coins[i].Grade)
		if !ok {
			plan.withoutGrade++
			continue
		}
		plan.eligible++
		item := numistaValueItem{coin: coins[i], typeID: typeID, grade: grade}
		if check, seen := checks[coins[i].ID]; seen {
			if now.Sub(check.CheckedAt) < numistaValueRecheck {
				continue
			}
			item.checkedAt = check.CheckedAt
		}
		plan.queue = append(plan.queue, item)
	}
	sort.SliceStable(plan.queue, func(i, j int) bool {
		return plan.queue[i].checkedAt.Before(plan.queue[j].checkedAt)
	})
	return plan, nil
}

// dailyBudget is how many requests a batch may make today: what is left of
// the background share of the month's quota, spread over the days left in the
// month (UTC, as Numista counts), and at most the daily cap.
func (s *NumistaValueService) dailyBudget(usage models.NumistaUsageSummary, now time.Time) int {
	left := usage.PauseAt - usage.Requests
	if usage.BackgroundPaused || left <= 0 {
		return 0
	}
	day := now.UTC()
	firstOfNext := time.Date(day.Year(), day.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	daysLeft := int(firstOfNext.Sub(time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)).Hours() / 24)
	share := max(1, left/max(1, daysLeft))
	return min(share, s.dailyCap())
}

func (s *NumistaValueService) dailyCap() int {
	limit, err := strconv.Atoi(strings.TrimSpace(s.settings.GetSetting(SettingNumistaValuesDailyRequests)))
	if err != nil || limit <= 0 {
		return defaultNumistaValuesDailyRequests
	}
	return limit
}

// quotaLevel tells whether the month's quota is fine, running low or spent
// for background requests.
func quotaLevel(usage models.NumistaUsageSummary) string {
	switch {
	case usage.BackgroundPaused:
		return models.NumistaQuotaPaused
	case usage.Limit > 0 && usage.Requests*100 >= usage.Limit*numistaQuotaLowShare:
		return models.NumistaQuotaLow
	default:
		return models.NumistaQuotaOK
	}
}

// Run makes one batch: it looks up the price of the coins due, in order,
// until they are all checked or the day's budget is spent, and saves what it
// did. trigger is "scheduled" or "manual".
func (s *NumistaValueService) Run(ctx context.Context, trigger string) (models.NumistaValueRun, error) {
	if !s.running.CompareAndSwap(false, true) {
		return models.NumistaValueRun{}, ErrNumistaValuesRunning
	}
	defer s.running.Store(false)

	usage := s.usage.Usage()
	run := models.NumistaValueRun{
		TriggerType: trigger,
		Status:      models.NumistaValueRunRunning,
		StartedAt:   s.now(),
		Budget:      s.dailyBudget(usage, s.now()),
	}
	if err := s.repo.CreateRun(&run); err != nil {
		return run, err
	}
	s.process(ctx, &run, usage)
	run.Requests = max(0, s.usage.Usage().Requests-usage.Requests)
	finished := s.now()
	run.CompletedAt = &finished
	if err := s.repo.SaveRun(&run); err != nil {
		return run, err
	}
	s.log(run)
	return run, nil
}

func (s *NumistaValueService) process(ctx context.Context, run *models.NumistaValueRun, start models.NumistaUsageSummary) {
	if run.Budget == 0 {
		run.Status = models.NumistaValueRunPaused
		return
	}
	plan, err := s.plan(s.now())
	if err != nil {
		run.Status = models.NumistaValueRunFailed
		s.warn("Could not list the coins to value: %v", err)
		return
	}
	currency := s.settings.CollectionCurrency()
	failures := 0
	for _, item := range plan.queue {
		if ctx.Err() != nil {
			run.Status = models.NumistaValueRunFailed
			return
		}
		usage := s.usage.Usage()
		if usage.BackgroundPaused {
			run.Status = models.NumistaValueRunPaused
			return
		}
		if usage.Requests-start.Requests >= run.Budget {
			run.Status = models.NumistaValueRunBudgetSpent
			return
		}
		value := s.market.ForType(WithNumistaBackground(ctx), item.typeID, item.coin.NumistaIssueID, item.coin.DateRange, currency, "")
		switch value.Status {
		case models.NumistaDataStatusOK:
			if value.Stale {
				// Numista did not answer; stored prices are not a new value.
				run.Failed++
				failures++
				break
			}
			failures = 0
			s.apply(run, item, value.Prices)
		case models.NumistaDataStatusNoIssues, models.NumistaDataStatusNoPrices:
			failures = 0
			s.record(run, item.coin.ID, models.NumistaValueCheckNoPrice)
		case models.NumistaDataStatusQuotaLimited:
			run.Status = models.NumistaValueRunPaused
			return
		case models.NumistaDataStatusUnconfigured, models.NumistaDataStatusUnauthorized:
			run.Status = models.NumistaValueRunUnconfigured
			return
		default:
			run.Failed++
			failures++
		}
		if failures >= numistaValueMaxFailures {
			run.Status = models.NumistaValueRunFailed
			return
		}
	}
	run.Status = models.NumistaValueRunCompleted
}

// apply sets a coin's value from Numista's price for its grade.
func (s *NumistaValueService) apply(run *models.NumistaValueRun, item numistaValueItem, prices []models.NumistaGradePrice) {
	price, ok := numistaPriceForGrade(prices, item.grade)
	if !ok {
		s.record(run, item.coin.ID, models.NumistaValueCheckNoPrice)
		return
	}
	price = math.Round(price*100) / 100
	now := s.now()
	if item.coin.CurrentValue != nil && math.Abs(*item.coin.CurrentValue-price) < 0.005 {
		if err := s.repo.ConfirmValue(item.coin.ID, now); err != nil {
			run.Failed++
			s.warn("Could not confirm the value of coin %d: %v", item.coin.ID, err)
			return
		}
		s.record(run, item.coin.ID, models.NumistaValueCheckUnchanged)
		return
	}
	if err := s.repo.SetValue(&item.coin, price, now); err != nil {
		run.Failed++
		s.warn("Could not save the value of coin %d: %v", item.coin.ID, err)
		return
	}
	s.record(run, item.coin.ID, models.NumistaValueCheckUpdated)
}

func (s *NumistaValueService) record(run *models.NumistaValueRun, coinID uint, status string) {
	if err := s.repo.SaveCheck(models.NumistaValueCheck{CoinID: coinID, CheckedAt: s.now(), Status: status}); err != nil {
		s.warn("Could not record the check of coin %d: %v", coinID, err)
	}
	run.Checked++
	switch status {
	case models.NumistaValueCheckUpdated:
		run.Updated++
	case models.NumistaValueCheckUnchanged:
		run.Unchanged++
	case models.NumistaValueCheckNoPrice:
		run.NoPrice++
	}
}

func (s *NumistaValueService) log(run models.NumistaValueRun) {
	if s.logger == nil {
		return
	}
	s.logger.Info("numista-values", "%s batch %s: %d checked, %d updated, %d unchanged, %d without a price, %d failed, %d of %d requests",
		run.TriggerType, run.Status, run.Checked, run.Updated, run.Unchanged, run.NoPrice, run.Failed, run.Requests, run.Budget)
	usage := s.usage.Usage()
	switch quotaLevel(usage) {
	case models.NumistaQuotaPaused:
		s.logger.Warn("numista-values", "Numista quota spent for background requests (%d of %d): values wait for next month", usage.Requests, usage.Limit)
	case models.NumistaQuotaLow:
		s.logger.Warn("numista-values", "Numista quota running low: %d of %d requests; batches pause at %d", usage.Requests, usage.Limit, usage.PauseAt)
	}
}

func (s *NumistaValueService) warn(format string, args ...any) {
	if s.logger != nil {
		s.logger.Warn("numista-values", format, args...)
	}
}

// Status tells how the daily update stands. nextRunIn is the scheduler's.
func (s *NumistaValueService) Status(nextRunIn time.Duration) (models.NumistaValueStatus, error) {
	now := s.now()
	usage := s.usage.Usage()
	month := now.UTC()
	status := models.NumistaValueStatus{
		Enabled:     s.Enabled(),
		IsRunning:   s.IsRunning(),
		Usage:       usage,
		QuotaLevel:  quotaLevel(usage),
		ResumesOn:   time.Date(month.Year(), month.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"),
		DailyBudget: s.dailyBudget(usage, now),
		Currency:    s.settings.CollectionCurrency(),
		Runs:        []models.NumistaValueRun{},
	}
	if status.Enabled {
		status.NextRunIn = nextRunIn.Round(time.Second).String()
	}
	plan, err := s.plan(now)
	if err != nil {
		return status, err
	}
	status.Eligible, status.Due, status.WithoutGrade = plan.eligible, len(plan.queue), plan.withoutGrade
	runs, err := s.repo.Runs(numistaValueRunsShown)
	if err != nil {
		return status, err
	}
	status.Runs = runs
	return status, nil
}
