package models

import "time"

// NumistaUsage counts the requests made to Numista's API in a month
// ("2026-10"), against the free plan's monthly quota.
type NumistaUsage struct {
	Month     string    `gorm:"primaryKey;size:7" json:"month"`
	Requests  int       `gorm:"not null;default:0" json:"requests"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NumistaUsageSummary is the month's count against the quota: above PauseAt,
// the requests nobody waits for (the catalogue, market values) are not made.
type NumistaUsageSummary struct {
	Month            string `json:"month" example:"2026-10"`
	Requests         int    `json:"requests" example:"1234"`
	Limit            int    `json:"limit" example:"2000"`
	PauseAt          int    `json:"pauseAt" example:"1800"`
	BackgroundPaused bool   `json:"backgroundPaused"`
}
