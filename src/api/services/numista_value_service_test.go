package services

import (
	"context"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"gorm.io/gorm"
)

// fakeNumistaUsage counts the fake client's requests as the month's usage.
type fakeNumistaUsage struct {
	client  *fakeNumistaPriceClient
	base    int
	limit   int
	pauseAt int
}

func (f *fakeNumistaUsage) Usage() models.NumistaUsageSummary {
	requests := f.base + f.client.issueCalls + f.client.priceCalls
	return models.NumistaUsageSummary{Month: "2026-10", Requests: requests, Limit: f.limit, PauseAt: f.pauseAt, BackgroundPaused: requests >= f.pauseAt}
}

type numistaValueFixture struct {
	db     *gorm.DB
	svc    *NumistaValueService
	client *fakeNumistaPriceClient
	usage  *fakeNumistaUsage
	repo   *repository.NumistaValueRepository
}

func newNumistaValueFixture(t *testing.T) numistaValueFixture {
	t.Helper()
	db := setupTestDB(t)
	if err := db.AutoMigrate(&models.NumistaValueRun{}, &models.NumistaValueCheck{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	market, client, _, _ := newMarketValueFixture()
	usage := &fakeNumistaUsage{client: client, base: 1000, limit: 2000, pauseAt: 1800}
	repo := repository.NewNumistaValueRepository(db)
	svc := NewNumistaValueService(repo, market, usage, NewSettingsService(repository.NewSettingsRepository(db)), nil)
	svc.now = func() time.Time { return time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC) }
	return numistaValueFixture{db: db, svc: svc, client: client, usage: usage, repo: repo}
}

func (f numistaValueFixture) coin(t *testing.T, coin models.Coin, numistaType string) *models.Coin {
	t.Helper()
	if err := f.db.Create(&coin).Error; err != nil {
		t.Fatalf("create %s: %v", coin.Name, err)
	}
	if numistaType != "" {
		if err := f.db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: numistaType}).Error; err != nil {
			t.Fatalf("reference: %v", err)
		}
	}
	return &coin
}

func (f numistaValueFixture) load(t *testing.T, id uint) models.Coin {
	t.Helper()
	var coin models.Coin
	if err := f.db.First(&coin, id).Error; err != nil {
		t.Fatalf("load %d: %v", id, err)
	}
	return coin
}

func valuePtr(v float64) *float64 { return &v }

func TestNumistaGradeForCoinGrade(t *testing.T) {
	cases := map[string]string{
		"UNC": "unc", "BNC": "unc", "MS 63": "unc", "FDC": "unc", "Soberba": "unc",
		"AU": "au", "About Unc": "au", "XF": "xf", "Ch XF": "xf", "Bela": "xf",
		"VF": "vf", "MBC": "vf", "F": "f", "BC": "f", "VG": "vg", "G": "g", "REG": "g",
		"PRF": "", "Proof": "", "": "", "  ": "",
	}
	for grade, want := range cases {
		got, ok := NumistaGradeForCoinGrade(grade)
		if got != want || ok != (want != "") {
			t.Errorf("NumistaGradeForCoinGrade(%q) = %q, %v; want %q", grade, got, ok, want)
		}
	}
}

func TestNumistaPriceForGrade(t *testing.T) {
	prices := []models.NumistaGradePrice{{Grade: "vg", Price: 1}, {Grade: "vf", Price: 2}, {Grade: "unc", Price: 5}}
	cases := map[string]float64{"vf": 2, "xf": 2, "au": 2, "unc": 5, "f": 1, "vg": 1, "g": 1}
	for grade, want := range cases {
		if got, ok := numistaPriceForGrade(prices, grade); !ok || got != want {
			t.Errorf("price for %s = %v, %v; want %v", grade, got, ok, want)
		}
	}
	if _, ok := numistaPriceForGrade(nil, "vf"); ok {
		t.Error("no prices must give no price")
	}
}

func TestNumistaValueService_DailyBudgetSpreadsWhatIsLeft(t *testing.T) {
	f := newNumistaValueFixture(t)
	day := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC) // 12 days left, the 20th included
	usage := func(requests int) models.NumistaUsageSummary {
		return models.NumistaUsageSummary{Requests: requests, Limit: 2000, PauseAt: 1800, BackgroundPaused: requests >= 1800}
	}
	if got := f.svc.dailyBudget(usage(1000), day); got != 40 {
		t.Fatalf("plenty left: budget %d, want the cap 40", got)
	}
	if got := f.svc.dailyBudget(usage(1700), day); got != 8 {
		t.Fatalf("100 left over 12 days: budget %d, want 8", got)
	}
	if got := f.svc.dailyBudget(usage(1795), day); got != 1 {
		t.Fatalf("a few left: budget %d, want 1", got)
	}
	if got := f.svc.dailyBudget(usage(1800), day); got != 0 {
		t.Fatalf("paused: budget %d, want 0", got)
	}
	if got := f.svc.dailyBudget(usage(1790), time.Date(2026, 10, 31, 23, 0, 0, 0, time.UTC)); got != 10 {
		t.Fatalf("last day: budget %d, want all 10 left", got)
	}
	if err := f.svc.settings.SetSetting(SettingNumistaValuesDailyRequests, "5"); err != nil {
		t.Fatal(err)
	}
	if got := f.svc.dailyBudget(usage(0), day); got != 5 {
		t.Fatalf("a lower cap: budget %d, want 5", got)
	}
	if quotaLevel(usage(1499)) != models.NumistaQuotaOK || quotaLevel(usage(1500)) != models.NumistaQuotaLow || quotaLevel(usage(1800)) != models.NumistaQuotaPaused {
		t.Fatal("quota levels: ok below 75%, low from 75%, paused from the pause")
	}
}

func TestNumistaValueService_UpdatesValuesWithinTheBudget(t *testing.T) {
	f := newNumistaValueFixture(t)
	// Issue 102 (1968): unc 2.75, f 0.5, vf 1 (xf 0 is not a price). Issue 101 (1967): vf 3.
	escudos := f.coin(t, models.Coin{Name: "2.50 Escudos 1968", UserID: 1, Grade: "UNC", DateRange: "1968", CurrentValue: valuePtr(1)}, "2933")
	same := f.coin(t, models.Coin{Name: "Same value", UserID: 1, Grade: "XF", DateRange: "1968", CurrentValue: valuePtr(1)}, "2933")
	worn := f.coin(t, models.Coin{Name: "Worn 1967", UserID: 1, Grade: "G", DateRange: "1967", CurrentValue: valuePtr(0.5)}, "2933")
	noGrade := f.coin(t, models.Coin{Name: "No grade", UserID: 1, DateRange: "1968", CurrentValue: valuePtr(1)}, "2933")
	proof := f.coin(t, models.Coin{Name: "Proof", UserID: 1, Grade: "PRF", DateRange: "1968", CurrentValue: valuePtr(9)}, "2933")
	wanted := f.coin(t, models.Coin{Name: "Wanted", UserID: 1, Grade: "UNC", DateRange: "1968", IsWishlist: true, CurrentValue: valuePtr(7)}, "2933")
	unlinked := f.coin(t, models.Coin{Name: "Unlinked", UserID: 1, Grade: "UNC", CurrentValue: valuePtr(4)}, "")

	status, err := f.svc.Status(0)
	if err != nil || status.Eligible != 3 || status.Due != 3 || status.WithoutGrade != 2 || status.DailyBudget != 40 || status.QuotaLevel != models.NumistaQuotaOK {
		t.Fatalf("status before = %+v (%v)", status, err)
	}

	// A budget of 2 requests: the first coin costs the issues and a price.
	if err := f.svc.settings.SetSetting(SettingNumistaValuesDailyRequests, "2"); err != nil {
		t.Fatal(err)
	}
	first, err := f.svc.Run(context.Background(), "scheduled")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if first.Status != models.NumistaValueRunBudgetSpent || first.Budget != 2 || first.Requests != 2 || first.Checked != 1 || first.Updated != 1 || first.CompletedAt == nil {
		t.Fatalf("first run = %+v", first)
	}
	if got := f.load(t, escudos.ID); got.CurrentValue == nil || *got.CurrentValue != 2.75 || got.CurrentValueUpdatedAt == nil {
		t.Fatalf("escudos value = %+v", got.CurrentValue)
	}
	var history []models.CoinValueHistory
	f.db.Where("coin_id = ?", escudos.ID).Order("recorded_at, id").Find(&history)
	if len(history) != 2 || history[0].Value != 1 || history[0].Source != models.ValueHistorySourceManual ||
		history[1].Value != 2.75 || history[1].Source != models.ValueHistorySourceNumista {
		t.Fatalf("the replaced value stays in the history: %+v", history)
	}

	// The next batch goes on: the 1968 price is cached (free), 1967 costs one request.
	second, err := f.svc.Run(context.Background(), "manual")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if second.Status != models.NumistaValueRunCompleted || second.Requests != 1 || second.Checked != 2 || second.Updated != 1 || second.Unchanged != 1 {
		t.Fatalf("second run = %+v", second)
	}
	if got := f.load(t, same.ID); *got.CurrentValue != 1 || got.CurrentValueUpdatedAt == nil {
		t.Fatalf("XF has no price; VF's (1) is the same value, confirmed: %+v", got)
	}
	if got := f.load(t, worn.ID); *got.CurrentValue != 3 {
		t.Fatalf("G has no price; the nearest higher (VF 3) is used: %v", *got.CurrentValue)
	}
	for _, untouched := range []*models.Coin{noGrade, proof, wanted, unlinked} {
		if got := f.load(t, untouched.ID); *got.CurrentValue != *untouched.CurrentValue {
			t.Fatalf("%s must keep its value: %v", untouched.Name, *got.CurrentValue)
		}
	}

	// Everything was checked: nothing is due for 30 days.
	calls := f.client.issueCalls + f.client.priceCalls
	third, err := f.svc.Run(context.Background(), "scheduled")
	if err != nil || third.Status != models.NumistaValueRunCompleted || third.Checked != 0 || f.client.issueCalls+f.client.priceCalls != calls {
		t.Fatalf("third run = %+v (%v)", third, err)
	}
	status, err = f.svc.Status(5 * time.Hour)
	if err != nil || status.Due != 0 || len(status.Runs) != 3 || status.Runs[0].ID != third.ID {
		t.Fatalf("status after = %+v (%v)", status, err)
	}
	if last := f.repo.LastScheduledRun(); last == nil || last.ID != third.ID {
		t.Fatalf("last scheduled run = %+v", last)
	}
}

func TestNumistaValueService_StopsWhenTheQuotaIsSpent(t *testing.T) {
	f := newNumistaValueFixture(t)
	coin := f.coin(t, models.Coin{Name: "2.50 Escudos 1968", UserID: 1, Grade: "UNC", DateRange: "1968", CurrentValue: valuePtr(1)}, "2933")
	f.usage.base = 1800

	run, err := f.svc.Run(context.Background(), "scheduled")
	if err != nil || run.Status != models.NumistaValueRunPaused || run.Budget != 0 || f.client.issueCalls+f.client.priceCalls != 0 {
		t.Fatalf("paused run = %+v (%v), %d calls", run, err, f.client.issueCalls+f.client.priceCalls)
	}
	if got := f.load(t, coin.ID); *got.CurrentValue != 1 {
		t.Fatalf("value changed while paused: %v", *got.CurrentValue)
	}
	status, err := f.svc.Status(0)
	if err != nil || status.QuotaLevel != models.NumistaQuotaPaused || status.ResumesOn != "2026-11-01" || status.DailyBudget != 0 {
		t.Fatalf("status = %+v (%v)", status, err)
	}
}

func TestNumistaValueService_StopsWithoutANumistaKey(t *testing.T) {
	f := newNumistaValueFixture(t)
	f.coin(t, models.Coin{Name: "A", UserID: 1, Grade: "UNC", DateRange: "1968"}, "2933")
	f.coin(t, models.Coin{Name: "B", UserID: 1, Grade: "UNC", DateRange: "1968"}, "2934")
	f.client.issuesErr = &NumistaError{Kind: NumistaErrorUnconfigured}

	run, err := f.svc.Run(context.Background(), "scheduled")
	if err != nil || run.Status != models.NumistaValueRunUnconfigured || f.client.issueCalls != 1 || run.Checked != 0 {
		t.Fatalf("run = %+v (%v), %d issue calls", run, err, f.client.issueCalls)
	}
}

func TestNumistaValueService_OneBatchAtATime(t *testing.T) {
	f := newNumistaValueFixture(t)
	f.svc.running.Store(true)
	if _, err := f.svc.Run(context.Background(), "manual"); err != ErrNumistaValuesRunning {
		t.Fatalf("err = %v", err)
	}
}

func TestNumistaValueScheduler_RunsOnceADayAtTheStartTime(t *testing.T) {
	f := newNumistaValueFixture(t)
	scheduler := NewNumistaValueScheduler(f.svc, f.repo, f.svc.settings, NewLogger(10))
	at := func(day, hour, minute int) time.Time { return time.Date(2026, 10, day, hour, minute, 0, 0, time.Local) }

	if got := scheduler.nextRun(at(20, 3, 0), false); !got.Equal(at(20, 5, 0)) {
		t.Fatalf("first run = %v, want today at 05:00", got)
	}
	completed := at(20, 5, 2)
	if err := f.repo.CreateRun(&models.NumistaValueRun{TriggerType: "scheduled", Status: models.NumistaValueRunBudgetSpent, StartedAt: at(20, 5, 0), CompletedAt: &completed}); err != nil {
		t.Fatal(err)
	}
	if got := scheduler.nextRun(at(20, 9, 0), false); !got.Equal(at(21, 5, 0)) {
		t.Fatalf("after today's batch = %v, want tomorrow at 05:00", got)
	}
	if err := f.svc.settings.SetSetting(SettingNumistaValuesStartTime, "06:30"); err != nil {
		t.Fatal(err)
	}
	if got := scheduler.nextRun(at(20, 9, 0), false); !got.Equal(at(21, 6, 30)) {
		t.Fatalf("with another start time = %v, want tomorrow at 06:30", got)
	}
	if status := scheduler.GetStatus(); status.Name != "numista-values" || status.Enabled {
		t.Fatalf("status = %+v (off by default)", status)
	}
}
