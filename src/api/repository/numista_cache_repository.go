package repository

import (
	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NumistaCacheRepository persists cached Numista API responses.
type NumistaCacheRepository struct {
	db *gorm.DB
}

// NewNumistaCacheRepository creates a new NumistaCacheRepository.
func NewNumistaCacheRepository(db *gorm.DB) *NumistaCacheRepository {
	return &NumistaCacheRepository{db: db}
}

// Find returns the cached entry for key, or nil when there is none.
func (r *NumistaCacheRepository) Find(key string) (*models.NumistaCacheEntry, error) {
	var entry models.NumistaCacheEntry
	result := r.db.Where("key = ?", key).Limit(1).Find(&entry)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &entry, nil
}

// Save creates or replaces the entry for entry.Key.
func (r *NumistaCacheRepository) Save(entry *models.NumistaCacheEntry) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"payload", "fetched_at"}),
	}).Create(entry).Error
}
