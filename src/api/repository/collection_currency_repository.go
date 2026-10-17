package repository

import (
	"fmt"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// CollectionCurrencyRepository converts the amounts kept in the collection
// currency when that currency changes.
type CollectionCurrencyRepository struct {
	db *gorm.DB
}

// NewCollectionCurrencyRepository creates the repository.
func NewCollectionCurrencyRepository(db *gorm.DB) *CollectionCurrencyRepository {
	return &CollectionCurrencyRepository{db: db}
}

// Amounts kept in the collection currency, by table and column. Wishlist
// search alerts carry their own currency and are not converted.
var collectionMoneyColumns = []struct {
	model   any
	columns []string
}{
	{&models.Coin{}, []string{"purchase_price", "current_value"}},
	{&models.CoinValueHistory{}, []string{"value"}},
	{&models.QuickCaptureDraft{}, []string{"purchase_price"}},
	{&models.CoinSetValuationSnapshot{}, []string{"total_value", "total_invested", "avg_value_per_coin"}},
	{&models.ValuationResult{}, []string{"previous_value", "estimated_value"}},
}

// ConvertAmounts multiplies every collection amount by factor (rounded to
// cents) and stores value under settingKey, in one transaction. It returns how
// many coins had a price or value.
func (r *CollectionCurrencyRepository) ConvertAmounts(factor float64, settingKey, value string) (int64, error) {
	var coins int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Coin{}).
			Where("purchase_price IS NOT NULL OR current_value IS NOT NULL").
			Count(&coins).Error; err != nil {
			return err
		}
		for _, table := range collectionMoneyColumns {
			for _, column := range table.columns {
				if err := tx.Model(table.model).
					Where(column+" IS NOT NULL").
					UpdateColumn(column, gorm.Expr("ROUND("+column+" * ?, 2)", factor)).Error; err != nil {
					return fmt.Errorf("convert %s: %w", column, err)
				}
			}
		}
		return tx.Save(&models.AppSetting{Key: settingKey, Value: value}).Error
	})
	return coins, err
}
