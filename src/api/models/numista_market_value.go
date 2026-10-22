package models

import "time"

// NumistaCacheEntry stores a raw Numista API response that changes rarely (a
// type's issues, an issue's estimated prices) so repeat views and restarts do
// not spend the monthly free-plan quota. Keys look like "issues:2933" or
// "prices:2933:14184:EUR".
type NumistaCacheEntry struct {
	Key       string    `gorm:"primaryKey;type:varchar(128)" json:"key"`
	Payload   string    `gorm:"type:text;not null" json:"-"`
	FetchedAt time.Time `gorm:"not null" json:"fetchedAt"`
}

// Statuses of Numista-backed lookups (market value, catalogue). Everything
// except "ok" carries no data; the UI explains the reason instead.
const (
	NumistaDataStatusOK           = "ok"
	NumistaDataStatusNotLinked    = "not_linked"
	NumistaDataStatusNoIssues     = "no_issues"
	NumistaDataStatusNoPrices     = "no_prices"
	NumistaDataStatusUnconfigured = "unconfigured"
	NumistaDataStatusUnauthorized = "unauthorized"
	NumistaDataStatusQuotaLimited = "quota_limited"
	NumistaDataStatusUnavailable  = "unavailable"
)

// NumistaIssueSummary is one issue (year / mint mark variant) of a Numista type.
type NumistaIssueSummary struct {
	ID int `json:"id"`
	// Year is the Gregorian year for dated issues (negative for BC).
	Year       *int   `json:"year,omitempty"`
	MinYear    *int   `json:"minYear,omitempty"`
	MaxYear    *int   `json:"maxYear,omitempty"`
	MintLetter string `json:"mintLetter,omitempty"`
	Mintage    *int64 `json:"mintage,omitempty"`
	Comment    string `json:"comment,omitempty"`
	// A banknote's issues are told apart by the date printed on the note
	// (Month and Day with Year), its signatories, the issue's own Pick number
	// ("176a") and whether its notes are replacement or specimen notes.
	Month      int                     `json:"month,omitempty" example:"9"`
	Day        int                     `json:"day,omitempty" example:"13"`
	Signatures []NumistaIssueSignature `json:"signatures,omitempty"`
	Pick       string                  `json:"pick,omitempty" example:"176a"`
	Variant    string                  `json:"variant,omitempty" enums:"replacement,specimen"`
}

// NumistaIssueSignature is one of the people who signed a banknote issue.
type NumistaIssueSignature struct {
	Name  string `json:"name" example:"José da Silva Lopes"`
	Title string `json:"title,omitempty" example:"Governor"`
}

// NumistaGradePrice is Numista's estimated price for one grade (g, vg, f, vf,
// xf, au, unc).
type NumistaGradePrice struct {
	Grade string  `json:"grade"`
	Price float64 `json:"price"`
}

// CoinMarketValue is Numista's estimated market value per grade for the issue
// of a coin's linked Numista type.
type CoinMarketValue struct {
	Status          string                `json:"status"`
	NumistaTypeID   int                   `json:"numistaTypeId,omitempty"`
	NumistaURL      string                `json:"numistaUrl,omitempty"`
	Issues          []NumistaIssueSummary `json:"issues"`
	SelectedIssueID int                   `json:"selectedIssueId,omitempty"`
	Currency        string                `json:"currency"`
	Prices          []NumistaGradePrice   `json:"prices"`
	// FetchedAt is when Numista last returned these prices.
	FetchedAt *time.Time `json:"fetchedAt,omitempty"`
	// Stale is true when a refresh failed and stored values are being shown.
	Stale bool `json:"stale"`
	// CatalogPhotos are Numista's reference photos of the type, with credit.
	CatalogPhotos []NumistaCatalogPhoto `json:"catalogPhotos"`
}
