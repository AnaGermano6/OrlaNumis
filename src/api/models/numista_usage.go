package models

import "time"

// Kinds of Numista request, as the monthly count tells them apart.
const (
	NumistaRequestSearch = "search"
	NumistaRequestType   = "type"
	NumistaRequestIssues = "issues"
	NumistaRequestPrices = "prices"
	NumistaRequestOther  = "other"
)

// NumistaUsage counts the requests made to Numista's API in a month
// ("2026-10"), against the free plan's monthly quota: in all, by what they
// asked for, and those made in the background. The kinds were first counted
// in 2026-10, so that month's add up to less than its requests.
type NumistaUsage struct {
	Month      string    `gorm:"primaryKey;size:7" json:"month"`
	Requests   int       `gorm:"not null;default:0" json:"requests"`
	Searches   int       `gorm:"not null;default:0" json:"searches"`
	Types      int       `gorm:"not null;default:0" json:"types"`
	Issues     int       `gorm:"not null;default:0" json:"issues"`
	Prices     int       `gorm:"not null;default:0" json:"prices"`
	Other      int       `gorm:"not null;default:0" json:"other"`
	Background int       `gorm:"not null;default:0" json:"background"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// NumistaUsageSummary is the month's count against the quota: above PauseAt,
// the requests nobody waits for (the catalogue, market values) are not made.
type NumistaUsageSummary struct {
	Month            string `json:"month" example:"2026-10"`
	Requests         int    `json:"requests" example:"1234"`
	Limit            int    `json:"limit" example:"2000"`
	PauseAt          int    `json:"pauseAt" example:"1800"`
	BackgroundPaused bool   `json:"backgroundPaused"`
	// The requests by what they asked for: searches, types, a type's issues,
	// an issue's prices, the rest (the list of issuers); and how many were
	// made in the background.
	Searches   int `json:"searches" example:"120"`
	Types      int `json:"types" example:"80"`
	Issues     int `json:"issues" example:"60"`
	Prices     int `json:"prices" example:"300"`
	Other      int `json:"other" example:"1"`
	Background int `json:"background" example:"250"`
	// PausedUntil is set while Numista refuses requests for its quota: until
	// then none is sent.
	PausedUntil *time.Time `json:"pausedUntil,omitempty"`
}
