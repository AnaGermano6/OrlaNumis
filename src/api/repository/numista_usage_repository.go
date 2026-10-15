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

// Add counts n more requests in month.
func (r *NumistaUsageRepository) Add(month string, n int, at time.Time) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "month"}},
		DoUpdates: clause.Assignments(map[string]any{"requests": gorm.Expr("requests + ?", n), "updated_at": at}),
	}).Create(&models.NumistaUsage{Month: month, Requests: n, UpdatedAt: at}).Error
}

// Requests is the month's count, 0 when none was made.
func (r *NumistaUsageRepository) Requests(month string) (int, error) {
	var usage models.NumistaUsage
	err := r.db.Where("month = ?", month).Limit(1).Find(&usage).Error
	return usage.Requests, err
}
