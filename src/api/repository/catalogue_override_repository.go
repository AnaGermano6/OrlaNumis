package repository

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CatalogueOverrideRepository keeps the coins each user hid in the
// catalogue, or said are the same as another.
type CatalogueOverrideRepository struct {
	db *gorm.DB
}

func NewCatalogueOverrideRepository(db *gorm.DB) *CatalogueOverrideRepository {
	return &CatalogueOverrideRepository{db: db}
}

// List is the user's choices, oldest first.
func (r *CatalogueOverrideRepository) List(userID uint) ([]models.CatalogueOverride, error) {
	overrides := []models.CatalogueOverride{}
	err := r.db.Where("user_id = ?", userID).Order("created_at, id").Find(&overrides).Error
	return overrides, err
}

// Set hides a coin (sameAs empty) or makes it the same as another, in place
// of what the user had chosen for it.
func (r *CatalogueOverrideRepository) Set(userID uint, coinKey, sameAs string, at time.Time) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "coin_key"}},
		DoUpdates: clause.Assignments(map[string]any{"same_as": sameAs, "created_at": at}),
	}).Create(&models.CatalogueOverride{UserID: userID, CoinKey: coinKey, SameAs: sameAs, CreatedAt: at}).Error
}

// Delete forgets the user's choices for the coins with these keys.
func (r *CatalogueOverrideRepository) Delete(userID uint, coinKeys []string) error {
	if len(coinKeys) == 0 {
		return nil
	}
	return r.db.Where("user_id = ? AND coin_key IN ?", userID, coinKeys).Delete(&models.CatalogueOverride{}).Error
}
