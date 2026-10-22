package models

import "time"

// Statuses of a batch of coin values updated from Numista's prices.
const (
	NumistaValueRunRunning = "running"
	// Completed: every coin due was checked.
	NumistaValueRunCompleted = "completed"
	// BudgetSpent: the day's share of the quota was used; the next run goes on.
	NumistaValueRunBudgetSpent = "budget_spent"
	// Paused: the month's quota is nearly spent; runs resume next month.
	NumistaValueRunPaused = "paused"
	// Unconfigured: Numista has no key, or refused it.
	NumistaValueRunUnconfigured = "unconfigured"
	// Failed: Numista did not answer.
	NumistaValueRunFailed = "failed"
)

// What a batch found for one coin.
const (
	NumistaValueCheckUpdated   = "updated"
	NumistaValueCheckUnchanged = "unchanged"
	// NoPrice: Numista has no price for the coin's issue.
	NumistaValueCheckNoPrice = "no_price"
)

// NumistaValueRun is one batch of the coins' current values updated from
// Numista's estimated prices, within the day's share of the monthly quota.
type NumistaValueRun struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TriggerType string     `gorm:"type:varchar(20);not null" json:"triggerType" enums:"scheduled,manual"`
	Status      string     `gorm:"type:varchar(20);not null;index" json:"status" enums:"running,completed,budget_spent,paused,unconfigured,failed"`
	StartedAt   time.Time  `gorm:"not null;index" json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	// Budget is how many Numista requests the run could make; Requests, how
	// many it made (prices Numista gave in the last days cost none).
	Budget   int `gorm:"not null;default:0" json:"budget"`
	Requests int `gorm:"not null;default:0" json:"requests"`
	// Checked coins had their price looked up: Updated got a new value,
	// Unchanged kept theirs, NoPrice have none on Numista, Failed could not
	// be looked up and are tried again next time.
	Checked   int `gorm:"not null;default:0" json:"checked"`
	Updated   int `gorm:"not null;default:0" json:"updated"`
	Unchanged int `gorm:"not null;default:0" json:"unchanged"`
	NoPrice   int `gorm:"not null;default:0" json:"noPrice"`
	Failed    int `gorm:"not null;default:0" json:"failed"`
}

// NumistaValueCheck is when a coin's price was last looked up in Numista,
// and what was found; it orders the coins of the next batches.
type NumistaValueCheck struct {
	CoinID    uint      `gorm:"primaryKey;autoIncrement:false" json:"coinId"`
	CheckedAt time.Time `gorm:"not null;index" json:"checkedAt"`
	Status    string    `gorm:"type:varchar(20);not null" json:"status" enums:"updated,unchanged,no_price"`
}

// Quota levels of the month's Numista requests, as the value batches see them.
const (
	NumistaQuotaOK = "ok"
	// Low: three quarters of the quota are spent; the batches get smaller.
	NumistaQuotaLow = "low"
	// Paused: the background share is spent; batches wait for next month.
	NumistaQuotaPaused = "paused"
)

// NumistaValueStatus tells how the daily update of coin values from Numista
// stands: the quota, the coins it covers and the last batch.
type NumistaValueStatus struct {
	Enabled   bool `json:"enabled"`
	IsRunning bool `json:"isRunning"`
	// NextRunIn is the time until the next scheduled batch (when enabled).
	NextRunIn string              `json:"nextRunIn" example:"5h0m0s"`
	Usage     NumistaUsageSummary `json:"usage"`
	// QuotaLevel is ok, low (from 75% of the quota) or paused (from PauseAt).
	QuotaLevel string `json:"quotaLevel" enums:"ok,low,paused"`
	// ResumesOn is the first day of next month (UTC), when the quota renews.
	ResumesOn string `json:"resumesOn" example:"2026-11-01"`
	// DailyBudget is how many requests the next batch may make.
	DailyBudget int `json:"dailyBudget"`
	// Eligible coins are linked to Numista and have a grade Numista prices;
	// Due ones have not been checked in the last 30 days.
	Eligible int `json:"eligible"`
	Due      int `json:"due"`
	// WithoutGrade coins are linked to Numista but have no grade Numista
	// prices (none, or a proof); they wait for one.
	WithoutGrade int               `json:"withoutGrade"`
	Currency     string            `json:"currency" example:"EUR"`
	Runs         []NumistaValueRun `json:"runs"`
}
