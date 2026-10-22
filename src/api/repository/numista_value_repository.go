package repository

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NumistaValueRepository keeps the batches that update coin values from
// Numista's prices, and when each coin was last checked.
type NumistaValueRepository struct {
	db *gorm.DB
}

// NewNumistaValueRepository creates the repository.
func NewNumistaValueRepository(db *gorm.DB) *NumistaValueRepository {
	return &NumistaValueRepository{db: db}
}

// LinkedCoins lists every user's collection items (coins and banknotes, not
// the wishlist) that have a Numista catalogue reference, with their
// references, in id order.
func (r *NumistaValueRepository) LinkedCoins() ([]models.Coin, error) {
	var coins []models.Coin
	err := r.db.Where("is_wishlist = ?", false).
		Where("id IN (SELECT coin_id FROM coin_references WHERE LOWER(TRIM(catalog)) = 'numista')").
		Preload("References").Order("id").Find(&coins).Error
	return coins, err
}

// Checks returns when each coin was last checked, by coin.
func (r *NumistaValueRepository) Checks() (map[uint]models.NumistaValueCheck, error) {
	var rows []models.NumistaValueCheck
	if err := r.db.Find(&rows).Error; err != nil {
		return nil, err
	}
	checks := make(map[uint]models.NumistaValueCheck, len(rows))
	for _, row := range rows {
		checks[row.CoinID] = row
	}
	return checks, nil
}

// SaveCheck records a coin's latest check.
func (r *NumistaValueRepository) SaveCheck(check models.NumistaValueCheck) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "coin_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"checked_at", "status"}),
	}).Create(&check).Error
}

// SetValue makes value a coin's current value from Numista, with a value
// history entry. A coin with no history yet first gets one for the value it
// had, so the history keeps what was replaced.
func (r *NumistaValueRepository) SetValue(coin *models.Coin, value float64, at time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var entries int64
		if err := tx.Model(&models.CoinValueHistory{}).Where("coin_id = ?", coin.ID).Count(&entries).Error; err != nil {
			return err
		}
		if entries == 0 && coin.CurrentValue != nil && *coin.CurrentValue > 0 {
			recorded := coin.CreatedAt
			if coin.CurrentValueUpdatedAt != nil {
				recorded = *coin.CurrentValueUpdatedAt
			}
			if !recorded.Before(at) {
				recorded = at.Add(-time.Second)
			}
			if err := tx.Create(&models.CoinValueHistory{
				CoinID: coin.ID, UserID: coin.UserID, Value: *coin.CurrentValue,
				Confidence: "manual", Source: models.ValueHistorySourceManual, RecordedAt: recorded,
			}).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&models.Coin{}).Where("id = ?", coin.ID).
			Updates(map[string]any{"current_value": value, "current_value_updated_at": at}).Error; err != nil {
			return err
		}
		return tx.Create(&models.CoinValueHistory{
			CoinID: coin.ID, UserID: coin.UserID, Value: value,
			Confidence: "numista", Source: models.ValueHistorySourceNumista, RecordedAt: at,
		}).Error
	})
}

// ConfirmValue records that Numista still gives a coin the value it has.
func (r *NumistaValueRepository) ConfirmValue(coinID uint, at time.Time) error {
	return r.db.Model(&models.Coin{}).Where("id = ?", coinID).Update("current_value_updated_at", at).Error
}

// CreateRun stores a new batch.
func (r *NumistaValueRepository) CreateRun(run *models.NumistaValueRun) error {
	return r.db.Create(run).Error
}

// SaveRun stores a batch's progress or outcome.
func (r *NumistaValueRepository) SaveRun(run *models.NumistaValueRun) error {
	return r.db.Save(run).Error
}

// Runs lists the latest batches, newest first.
func (r *NumistaValueRepository) Runs(limit int) ([]models.NumistaValueRun, error) {
	var runs []models.NumistaValueRun
	err := r.db.Order("started_at DESC, id DESC").Limit(limit).Find(&runs).Error
	return runs, err
}

// LastScheduledRun returns the latest finished scheduled batch, or nil.
func (r *NumistaValueRepository) LastScheduledRun() *models.NumistaValueRun {
	var run models.NumistaValueRun
	err := r.db.Where("trigger_type = ? AND completed_at IS NOT NULL", "scheduled").
		Order("completed_at DESC, id DESC").First(&run).Error
	if err != nil {
		return nil
	}
	return &run
}

// AbandonRunning marks batches left running (the app stopped mid-batch) as failed.
func (r *NumistaValueRepository) AbandonRunning(at time.Time) error {
	return r.db.Model(&models.NumistaValueRun{}).Where("status = ?", models.NumistaValueRunRunning).
		Updates(map[string]any{"status": models.NumistaValueRunFailed, "completed_at": at}).Error
}
