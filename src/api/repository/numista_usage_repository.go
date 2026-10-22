package repository

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NumistaUsageRepository keeps the count of Numista requests by month.
type NumistaUsageRepository struct {
	db *gorm.DB
}

func NewNumistaUsageRepository(db *gorm.DB) *NumistaUsageRepository {
	return &NumistaUsageRepository{db: db}
}

// Add counts one more request in month, of the given kind
// (models.NumistaRequest…), and among the background ones when it was made
// there.
func (r *NumistaUsageRepository) Add(month, kind string, background bool, at time.Time) error {
	usage := models.NumistaUsage{Month: month, Requests: 1, UpdatedAt: at}
	var column string
	switch kind {
	case models.NumistaRequestSearch:
		usage.Searches, column = 1, "searches"
	case models.NumistaRequestType:
		usage.Types, column = 1, "types"
	case models.NumistaRequestIssues:
		usage.Issues, column = 1, "issues"
	case models.NumistaRequestPrices:
		usage.Prices, column = 1, "prices"
	default:
		usage.Other, column = 1, "other"
	}
	if background {
		usage.Background = 1
	}
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "month"}},
		DoUpdates: clause.Assignments(map[string]any{
			"requests":   gorm.Expr("requests + 1"),
			column:       gorm.Expr(column + " + 1"),
			"background": gorm.Expr("background + ?", usage.Background),
			"updated_at": at,
		}),
	}).Create(&usage).Error
}

// Month is the month's count, all zero when no request was made.
func (r *NumistaUsageRepository) Month(month string) (models.NumistaUsage, error) {
	usage := models.NumistaUsage{Month: month}
	err := r.db.Where("month = ?", month).Limit(1).Find(&usage).Error
	return usage, err
}
