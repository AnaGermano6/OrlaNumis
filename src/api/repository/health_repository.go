package repository

import (
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
)

// HealthRepository encapsulates health scorecard data access.
type HealthRepository struct {
	db *gorm.DB
}

// NewHealthRepository creates a new HealthRepository.
func NewHealthRepository(db *gorm.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

// EligibleCoinRow is a compact row shape for score input queries.
type EligibleCoinRow struct {
	CoinID                uint
	Title                 string
	Category              string
	Kind                  string
	Denomination          string
	Country               string
	MonetarySystem        string
	Mint                  string
	Material              string
	Grade                 string
	ReferenceURL          string
	PurchaseDate          *time.Time
	CurrentValue          *float64
	CurrentValueUpdatedAt *time.Time
	AIAnalysis            string
	ObverseAnalysis       string
	ReverseAnalysis       string
	UpdatedAt             time.Time
	PrimaryImageCount     int64
	ImageCount            int64
}

// ListEligibleCoins returns active collection coins for a user.
func (r *HealthRepository) ListEligibleCoins(userID uint) ([]EligibleCoinRow, error) {
	rows := []EligibleCoinRow{}
	err := r.db.Model(&models.Coin{}).
		Scopes(ActiveCollection(userID)).
		Select(
			"id AS coin_id, name AS title, kind, category, denomination, country, monetary_system, mint, material, grade, reference_url, purchase_date, current_value, current_value_updated_at, ai_analysis, obverse_analysis, reverse_analysis, updated_at, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id AND is_primary = true) AS primary_image_count, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id) AS image_count",
		).
		Order("updated_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ListEligibleCoinsPaged returns active collection coins with optional scope filtering.
func (r *HealthRepository) ListEligibleCoinsPaged(userID uint, page, limit int, scope string) ([]EligibleCoinRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	query := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []EligibleCoinRow{}
	offset := (page - 1) * limit

	selectClause := "id AS coin_id, name AS title, kind, category, denomination, country, monetary_system, mint, material, grade, reference_url, purchase_date, current_value, current_value_updated_at, ai_analysis, obverse_analysis, reverse_analysis, updated_at, " +
		"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id AND is_primary = true) AS primary_image_count, " +
		"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id) AS image_count"

	orderClause := "updated_at DESC, id ASC"
	if scope == "needs_attention" {
		orderClause = "updated_at ASC, id ASC"
	}

	err := query.
		Select(selectClause).
		Order(orderClause).
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// ListAllEligibleCoins returns all active collection coins across all users.
func (r *HealthRepository) ListAllEligibleCoins() ([]EligibleCoinRow, error) {
	rows := []EligibleCoinRow{}
	err := r.db.Model(&models.Coin{}).
		Where("is_wishlist = ?", false).
		Select(
			"id AS coin_id, name AS title, kind, category, denomination, country, monetary_system, mint, material, grade, reference_url, purchase_date, current_value, current_value_updated_at, ai_analysis, obverse_analysis, reverse_analysis, updated_at, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id AND is_primary = true) AS primary_image_count, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id) AS image_count",
		).
		Order("updated_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetSingleEligibleCoin returns a single coin's health data for the requesting user.
func (r *HealthRepository) GetSingleEligibleCoin(coinID, userID uint) (*EligibleCoinRow, error) {
	var row EligibleCoinRow
	err := r.db.Model(&models.Coin{}).
		Scopes(ActiveCollection(userID)).
		Where("id = ?", coinID).
		Select(
			"id AS coin_id, name AS title, kind, category, denomination, country, monetary_system, mint, material, grade, reference_url, purchase_date, current_value, current_value_updated_at, ai_analysis, obverse_analysis, reverse_analysis, updated_at, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id AND is_primary = true) AS primary_image_count, " +
				"(SELECT COUNT(*) FROM coin_images WHERE coin_id = coins.id) AS image_count",
		).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
