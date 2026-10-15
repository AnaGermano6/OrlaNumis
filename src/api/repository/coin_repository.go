package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrStorageSlotOccupied = errors.New("storage slot is occupied")

// CoinListFilters holds optional filters for listing coins.
type CoinListFilters struct {
	// Kind keeps the coins or the banknotes; every item when "".
	Kind     string
	Category string
	Era      string
	Country  string
	// Denomination keeps the coins of one face value, as written ("2 euro").
	Denomination string
	// Countries, when not nil, keeps only coins of these countries (none
	// when empty), e.g. the collection's extinct countries.
	Countries []string
	// IDs, when not nil, keeps only these coins (none when empty), e.g. the
	// ones still legal tender.
	IDs []uint
	// Repeated keeps the coins held more than once (a quantity above one).
	Repeated     bool
	Presentation string
	Rarity       string
	Search       string
	Wishlist     *bool
	Sold         *bool
	TagID        *uint
	SetID        *uint
	SortField    string
	SortOrder    string
	Seed         *int // for SortField == "random"; deterministic per-seed shuffle
	Page         int
	Limit        int
}

// OwnedCoinFilters are lightweight scoped filters used by collection chat tools.
type OwnedCoinFilters struct {
	Category      string
	Material      string
	Era           string
	Ruler         string
	Wishlist      *bool
	Sold          *bool
	Search        string
	MissingFields []string
}

// CategoryCount holds a category name and its coin count.
type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

// MaterialCount holds a material name and its coin count.
type MaterialCount struct {
	Material string `json:"material"`
	Count    int64  `json:"count"`
}

// GradeCount holds a grade and its coin count.
type GradeCount struct {
	Grade string `json:"grade"`
	Count int64  `json:"count"`
}

// EraCount holds an era and its coin count.
type EraCount struct {
	Era   string `json:"era"`
	Count int64  `json:"count"`
}

// RulerCount holds a ruler name and its coin count.
type RulerCount struct {
	Ruler string `json:"ruler"`
	Count int64  `json:"count"`
}

// PriceRange holds a price range label and its coin count.
type PriceRange struct {
	Range string `json:"range"`
	Count int64  `json:"count"`
}

// ValueSummary holds aggregate value metrics.
type ValueSummary struct {
	TotalPurchasePrice float64 `json:"totalPurchasePrice"`
	TotalCurrentValue  float64 `json:"totalCurrentValue"`
	AvgPurchasePrice   float64 `json:"avgPurchasePrice"`
	AvgCurrentValue    float64 `json:"avgCurrentValue"`
}

// SoldSummary holds aggregate sold coin metrics.
type SoldSummary struct {
	TotalSoldPrice    float64 `json:"totalSoldPrice"`
	TotalPurchaseCost float64 `json:"totalPurchaseCost"`
}

// CollectionStats holds all aggregate statistics for a user's collection.
type CollectionStats struct {
	TotalCoins    int64           `json:"totalCoins"`
	TotalWishlist int64           `json:"totalWishlist"`
	TotalSold     int64           `json:"totalSold"`
	ByCategory    []CategoryCount `json:"byCategory"`
	ByMaterial    []MaterialCount `json:"byMaterial"`
	ByGrade       []GradeCount    `json:"byGrade"`
	ByEra         []EraCount      `json:"byEra"`
	ByRuler       []RulerCount    `json:"byRuler"`
	ByPriceRange  []PriceRange    `json:"byPriceRange"`
	Values        ValueSummary    `json:"values"`
	SoldValues    SoldSummary     `json:"soldValues"`
}

// DistributionCell holds a single cell in the era × category heat map.
type DistributionCell struct {
	Era      string `json:"era"`
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// CoinRepository encapsulates all coin-related database operations.
type CoinRepository struct {
	db *gorm.DB
}

// NewCoinRepository creates a new CoinRepository.
func NewCoinRepository(db *gorm.DB) *CoinRepository {
	return &CoinRepository{db: db}
}

// WithTx returns a shallow copy of the repository that uses tx for all queries.
func (r *CoinRepository) WithTx(tx *gorm.DB) *CoinRepository {
	return &CoinRepository{db: tx}
}

// DB exposes the underlying *gorm.DB so callers can create transactions.
func (r *CoinRepository) DB() *gorm.DB {
	return r.db
}

func omitCoinRelationships(db *gorm.DB) *gorm.DB {
	return db.Omit("Tags", "Sets")
}

func applyOwnedFilterConditions(query *gorm.DB, filters OwnedCoinFilters) *gorm.DB {
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.Material != "" {
		query = query.Where("material = ?", filters.Material)
	}
	if filters.Era != "" {
		query = query.Where("era = ?", filters.Era)
	}
	if filters.Ruler != "" {
		query = query.Where("LOWER(ruler) LIKE LOWER(?)", "%"+filters.Ruler+"%")
	}
	if filters.Wishlist != nil {
		query = query.Where("is_wishlist = ?", *filters.Wishlist)
	}
	if filters.Sold != nil {
		query = query.Where("is_sold = ?", *filters.Sold)
	}
	if filters.Search != "" {
		term := "%" + strings.TrimSpace(filters.Search) + "%"
		query = query.Where(
			"name LIKE ? OR denomination LIKE ? OR ruler LIKE ? OR notes LIKE ?",
			term, term, term, term,
		)
	}
	missingConditions := make([]string, 0, len(filters.MissingFields))
	for _, field := range filters.MissingFields {
		if condition, ok := ownedMissingFieldConditions[field]; ok {
			missingConditions = append(missingConditions, "("+condition+")")
		}
	}
	if len(missingConditions) > 0 {
		query = query.Where(strings.Join(missingConditions, " OR "))
	}
	return query
}

var ownedMissingFieldConditions = map[string]string{
	"denomination":    "denomination = ''",
	"ruler":           "ruler = ''",
	"era":             "era = ''",
	"mint":            "mint = ''",
	"material":        "material = '' OR material = 'Other'",
	"weightGrams":     "weight_grams IS NULL OR weight_grams <= 0",
	"diameterMm":      "diameter_mm IS NULL OR diameter_mm <= 0",
	"grade":           "grade = ''",
	"purchasePrice":   "purchase_price IS NULL",
	"currentValue":    "current_value IS NULL",
	"purchaseDate":    "purchase_date IS NULL",
	"storageLocation": "storage_location_id IS NULL",
	"notes":           "notes = ''",
	"referenceUrl":    "reference_url = ''",
	"referenceText":   "reference_text = ''",
}

var allowedSortFields = map[string]string{
	"created_at":    "created_at",
	"updated_at":    "updated_at",
	"current_value": "current_value",
	"purchase_date": "purchase_date",
	"name":          "name",
	"quantity":      "quantity",
}

var searchFields = []string{
	"name", "denomination", "ruler", "era", "mint",
	"obverse_inscription", "reverse_inscription", "notes", "rarity_rating", "country",
	// Banknotes: a serial number, its series and signatories.
	"serial_number", "series", "signatures",
}

// List returns a paginated, filtered list of coins for a user.
func (r *CoinRepository) List(userID uint, filters CoinListFilters) ([]models.Coin, int64, error) {
	query := r.db.Scopes(OwnedBy(userID), OfKind(filters.Kind)).Preload("Images").Preload("Tags").Preload("Sets").Preload("References").Preload("StorageLocation").Preload("MintLocation")

	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.IDs != nil {
		if len(filters.IDs) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("coins.id IN ?", filters.IDs)
		}
	}
	if filters.Countries != nil {
		if len(filters.Countries) == 0 {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("TRIM(country) IN ?", filters.Countries)
		}
	}
	if filters.Country != "" {
		query = query.Where("country = ?", filters.Country)
	}
	if denomination := strings.TrimSpace(filters.Denomination); denomination != "" {
		query = query.Where("TRIM(denomination) = ?", denomination)
	}
	if filters.Presentation != "" {
		query = query.Where("presentation = ?", filters.Presentation)
	}
	if filters.Repeated {
		query = query.Where("quantity > 1")
	}
	if filters.Rarity != "" {
		query = query.Where("rarity = ?", filters.Rarity)
	}
	if filters.Era != "" {
		query = query.Where("era = ?", filters.Era)
	}
	if filters.Wishlist != nil {
		query = query.Where("is_wishlist = ?", *filters.Wishlist)
	}
	if filters.Sold != nil {
		query = query.Where("is_sold = ?", *filters.Sold)
	}
	if filters.TagID != nil {
		query = query.Where("id IN (SELECT coin_id FROM coin_tags WHERE tag_id = ?)", *filters.TagID)
	}
	if filters.SetID != nil {
		query = query.Where("id IN (SELECT coin_id FROM coin_set_memberships WHERE set_id = ?)", *filters.SetID)
	}
	if filters.Search != "" {
		term := "%" + filters.Search + "%"
		sub := r.db.Where("name LIKE ?", term)
		for _, f := range searchFields[1:] {
			sub = sub.Or(f+" LIKE ?", term)
		}
		query = query.Where(sub)
	}

	var total int64
	if err := query.Model(&models.Coin{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	limit := filters.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}
	offset := (page - 1) * limit

	var coins []models.Coin

	// Seeded deterministic random sort: same seed produces the same order across pages.
	// Hash mixes id with the seed and Knuth's golden-ratio multiplier (2654435761) so
	// the modulo wraps for every row, producing a true permutation. Without the large
	// multiplier, `id*seed + seed` is monotonic in id and degenerates to insertion
	// order. abs() protects against any 64-bit overflow.
	// Seed is bound via %d after being validated by strconv.Atoi in the handler — safe
	// from SQL injection. Note: gorm.Expr placeholders inside Order() are silently
	// dropped by this GORM build, so the int is formatted directly into the SQL.
	if filters.SortField == "random" && filters.Seed != nil {
		seed := *filters.Seed
		orderExpr := fmt.Sprintf("abs((id * 2654435761) + (id * %d) + %d) %% 1000000", seed, seed)
		if err := query.Order(orderExpr).Offset(offset).Limit(limit).Find(&coins).Error; err != nil {
			return nil, 0, err
		}
		return coins, total, nil
	}

	col, ok := allowedSortFields[filters.SortField]
	if !ok {
		col = "updated_at"
	}
	order := filters.SortOrder
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	if err := query.
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: col},
			Desc:   order != "asc",
		}).
		Offset(offset).
		Limit(limit).
		Find(&coins).Error; err != nil {
		return nil, 0, err
	}
	return coins, total, nil
}

// FindByID returns a single coin owned by the user, with images preloaded.
func (r *CoinRepository) FindByID(id uint, userID uint) (*models.Coin, error) {
	var coin models.Coin
	err := r.db.Scopes(OwnedByID(id, userID)).Preload("Images").Preload("Tags").Preload("Sets").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(&coin).Error
	if err != nil {
		return nil, err
	}
	return &coin, nil
}

// ListMatchedImperialFigures returns the user's active-collection coins
// (non-wishlist, non-sold) that have been matched to a RomanImperialFigure,
// with images preloaded, for F028's emperor tracker.
func (r *CoinRepository) ListMatchedImperialFigures(userID uint) ([]models.Coin, error) {
	var coins []models.Coin
	err := r.db.Scopes(ActiveRomanCollection(userID)).
		Where("roman_imperial_figure_id IS NOT NULL").
		Preload("Images").
		Order("id ASC").
		Find(&coins).Error
	return coins, err
}

// FindOwnedByNameCandidates returns up to limit owned coins with name matches.
func (r *CoinRepository) FindOwnedByNameCandidates(userID uint, name string, limit int) ([]models.Coin, error) {
	if limit < 1 {
		limit = 5
	}
	var coins []models.Coin
	err := r.db.Model(&models.Coin{}).
		Scopes(OwnedBy(userID)).
		Where("LOWER(name) LIKE LOWER(?)", "%"+strings.TrimSpace(name)+"%").
		Order("name ASC").
		Limit(limit).
		Find(&coins).Error
	return coins, err
}

// ListOwnedByFilters returns owner-scoped coins filtered for collection chat queries.
func (r *CoinRepository) ListOwnedByFilters(userID uint, filters OwnedCoinFilters, limit int) ([]models.Coin, error) {
	if limit < 1 {
		limit = 5
	}
	var coins []models.Coin
	query := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID))
	query = applyOwnedFilterConditions(query, filters)
	err := query.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "updated_at"}, Desc: true}).
		Limit(limit).
		Find(&coins).Error
	return coins, err
}

// CountOwnedByFilters counts owner-scoped coins matching collection chat filters.
func (r *CoinRepository) CountOwnedByFilters(userID uint, filters OwnedCoinFilters) (int64, error) {
	var count int64
	query := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID))
	query = applyOwnedFilterConditions(query, filters)
	err := query.Count(&count).Error
	return count, err
}

// TopOwnedByCurrentValue returns the top N active collection coins by current value.
func (r *CoinRepository) TopOwnedByCurrentValue(userID uint, limit int) ([]models.Coin, error) {
	if limit < 1 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}
	var coins []models.Coin
	err := r.db.Model(&models.Coin{}).
		Scopes(ActiveCollection(userID)).
		Where("current_value IS NOT NULL").
		Order(clause.OrderByColumn{Column: clause.Column{Name: "current_value"}, Desc: true}).
		Limit(limit).
		Find(&coins).Error
	return coins, err
}

// Create inserts a new coin and returns it with images preloaded.
func (r *CoinRepository) Create(coin *models.Coin) error {
	if err := r.db.Omit("References").Create(coin).Error; err != nil {
		if isStorageSlotConflict(err, coin.StorageSlot) {
			return ErrStorageSlotOccupied
		}
		return err
	}

	return r.db.Preload("Images").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(coin, coin.ID).Error
}

func isStorageSlotConflict(err error, storageSlot *int) bool {
	if err == nil || storageSlot == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "idx_coins_storage_location_slot_unique") ||
		(strings.Contains(message, "unique constraint failed") &&
			strings.Contains(message, "coins.storage_location_id") &&
			strings.Contains(message, "coins.storage_slot"))
}

// Duplicate creates an owner-scoped copy of a coin without copying media or card state rows.
func (r *CoinRepository) Duplicate(id uint, userID uint) (*models.Coin, error) {
	var source models.Coin
	if err := r.db.Scopes(OwnedByID(id, userID)).
		Preload("References").
		Preload("StorageLocation").Preload("MintLocation").
		First(&source).Error; err != nil {
		return nil, err
	}

	duplicate := models.Coin{
		Name: source.Name + " (duplicate)",
		// A banknote's copy is a banknote of the same issue; its serial number
		// and certificate are the copied note's own, so they are not copied.
		Kind: source.Kind,
		// The copy is one new item, whatever number the original holds.
		Quantity:              1,
		Series:                source.Series,
		Signatures:            source.Signatures,
		Printer:               source.Printer,
		WidthMm:               source.WidthMm,
		HeightMm:              source.HeightMm,
		Watermark:             source.Watermark,
		NoteVariant:           source.NoteVariant,
		CertifiedGrade:        source.CertifiedGrade,
		Category:              source.Category,
		Denomination:          source.Denomination,
		Ruler:                 source.Ruler,
		Era:                   source.Era,
		Mint:                  source.Mint,
		Material:              source.Material,
		WeightGrams:           source.WeightGrams,
		DiameterMm:            source.DiameterMm,
		Grade:                 source.Grade,
		ObverseInscription:    source.ObverseInscription,
		ReverseInscription:    source.ReverseInscription,
		ObverseDescription:    source.ObverseDescription,
		ReverseDescription:    source.ReverseDescription,
		RarityRating:          source.RarityRating,
		Year:                  source.Year,
		Country:               source.Country,
		Presentation:          source.Presentation,
		Rarity:                source.Rarity,
		Circulation:           source.Circulation,
		PurchasePrice:         source.PurchasePrice,
		CurrentValue:          source.CurrentValue,
		CurrentValueUpdatedAt: source.CurrentValueUpdatedAt,
		PurchaseDate:          source.PurchaseDate,
		PurchaseLocation:      source.PurchaseLocation,
		VendorSKU:             source.VendorSKU,
		VendorInvoice:         source.VendorInvoice,
		Notes:                 source.Notes,
		AIAnalysis:            source.AIAnalysis,
		ObverseAnalysis:       source.ObverseAnalysis,
		ReverseAnalysis:       source.ReverseAnalysis,
		ReferenceURL:          source.ReferenceURL,
		ReferenceText:         source.ReferenceText,
		IsWishlist:            source.IsWishlist,
		IsSold:                source.IsSold,
		SoldPrice:             source.SoldPrice,
		SoldDate:              source.SoldDate,
		SoldTo:                source.SoldTo,
		ListingStatus:         source.ListingStatus,
		ListingCheckedAt:      source.ListingCheckedAt,
		ListingCheckReason:    source.ListingCheckReason,
		StorageLocationID:     source.StorageLocationID,
		StorageSlot:           source.StorageSlot,
		IsPrivate:             source.IsPrivate,
		UserID:                source.UserID,
	}
	if source.StorageLocation != nil && source.StorageLocation.Type == models.StorageLocationTypeTray {
		duplicate.StorageLocationID = nil
		duplicate.StorageSlot = nil
	}
	if err := r.db.Omit("Images", "References", "Tags", "Sets", "StorageLocation", "User").Create(&duplicate).Error; err != nil {
		return nil, err
	}

	if len(source.References) > 0 {
		references := make([]models.CoinReference, 0, len(source.References))
		for _, ref := range source.References {
			references = append(references, models.CoinReference{
				CoinID:  duplicate.ID,
				Catalog: ref.Catalog,
				Volume:  ref.Volume,
				Number:  ref.Number,
				URI:     ref.URI,
			})
		}
		if err := r.db.Create(&references).Error; err != nil {
			return nil, err
		}
	}

	var tagIDs []uint
	if err := r.db.Table("coin_tags").
		Select("coin_tags.tag_id").
		Joins("JOIN tags ON tags.id = coin_tags.tag_id").
		Where("coin_tags.coin_id = ? AND tags.user_id = ?", source.ID, userID).
		Scan(&tagIDs).Error; err != nil {
		return nil, err
	}
	if len(tagIDs) > 0 {
		coinTags := make([]models.CoinTag, 0, len(tagIDs))
		for _, tagID := range tagIDs {
			coinTags = append(coinTags, models.CoinTag{CoinID: duplicate.ID, TagID: tagID})
		}
		if err := r.db.Create(&coinTags).Error; err != nil {
			return nil, err
		}
	}

	var memberships []models.CoinSetMembership
	if err := r.db.Table("coin_set_memberships").
		Select("coin_set_memberships.*").
		Joins("JOIN coin_sets ON coin_sets.id = coin_set_memberships.set_id").
		Where("coin_set_memberships.coin_id = ? AND coin_sets.user_id = ?", source.ID, userID).
		Scan(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) > 0 {
		copiedMemberships := make([]models.CoinSetMembership, 0, len(memberships))
		for _, membership := range memberships {
			copiedMemberships = append(copiedMemberships, models.CoinSetMembership{
				SetID:     membership.SetID,
				CoinID:    duplicate.ID,
				AddedAt:   membership.AddedAt,
				SortOrder: membership.SortOrder,
				Notes:     membership.Notes,
			})
		}
		if err := r.db.Create(&copiedMemberships).Error; err != nil {
			return nil, err
		}
	}

	if err := r.db.Preload("Images").Preload("References").Preload("Tags").Preload("Sets").Preload("StorageLocation").Preload("MintLocation").First(&duplicate, duplicate.ID).Error; err != nil {
		return nil, err
	}
	return &duplicate, nil
}

// Update applies a scalar patch to an existing coin and reloads read associations.
// When selectFields are supplied, only those fields are persisted, including
// explicit zero values; relationship changes stay on their dedicated paths.
func (r *CoinRepository) Update(existing *models.Coin, updates *models.Coin, selectFields ...string) error {
	// Relationship changes are managed through dedicated tag/set methods.
	// Coin sets require explicit membership writes because AddedAt is NOT NULL.
	query := omitCoinRelationships(r.db.Model(existing))
	if len(selectFields) > 0 {
		query = query.Select(selectFields)
	}
	if err := query.Updates(updates).Error; err != nil {
		return err
	}
	return r.db.Preload("Images").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(existing, existing.ID).Error
}

// UpdateField updates one explicit column without syncing loaded associations.
func (r *CoinRepository) UpdateField(coin *models.Coin, field string, value interface{}) error {
	if err := omitCoinRelationships(r.db.Model(coin)).Update(field, value).Error; err != nil {
		return err
	}
	return r.db.Preload("Images").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(coin, coin.ID).Error
}

// UpdateFields updates multiple explicit columns, including zero/nil values, without syncing loaded associations.
func (r *CoinRepository) UpdateFields(coin *models.Coin, updates map[string]interface{}) error {
	if err := omitCoinRelationships(r.db.Model(coin)).Updates(updates).Error; err != nil {
		return err
	}
	return r.db.Preload("Images").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(coin, coin.ID).Error
}

// UpdateStorageLocationID updates only the storage-location foreign key, including clearing it.
func (r *CoinRepository) UpdateStorageLocationID(coin *models.Coin, storageLocationID *uint) error {
	return r.UpdateStorageAssignment(coin, storageLocationID, nil)
}

// UpdateStorageAssignment atomically persists location and one-based slot.
func (r *CoinRepository) UpdateStorageAssignment(coin *models.Coin, storageLocationID *uint, storageSlot *int) error {
	query := r.db.Model(&models.Coin{}).Where("id = ? AND user_id = ?", coin.ID, coin.UserID)
	err := query.Updates(map[string]interface{}{
		"storage_location_id": storageLocationID,
		"storage_slot":        storageSlot,
	}).Error
	if err != nil {
		if isStorageSlotConflict(err, storageSlot) {
			return ErrStorageSlotOccupied
		}
		return err
	}
	if err := r.db.Preload("Images").Preload("References").Preload("StorageLocation").Preload("MintLocation").First(coin, coin.ID).Error; err != nil {
		return err
	}
	if storageLocationID == nil {
		coin.StorageLocation = nil
	}
	coin.StorageLocationID = storageLocationID
	coin.StorageSlot = storageSlot
	return nil
}

// Delete removes a coin and all associated data (images, journals, value
// history) in a single transaction. Returns rows affected for the coin delete.
func (r *CoinRepository) Delete(id uint, userID uint) (int64, error) {
	var rowsAffected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		rowsAffected, err = r.WithTx(tx).DeleteInTx(id, userID)
		return err
	})
	return rowsAffected, err
}

func (r *CoinRepository) DeleteInTx(id uint, userID uint) (int64, error) {
	var owned int64
	if err := r.db.Model(&models.Coin{}).Scopes(OwnedByID(id, userID)).Count(&owned).Error; err != nil {
		return 0, err
	}
	if owned == 0 {
		return 0, nil
	}
	// Rows pointing at the coin go first: SQLite enforces the foreign keys of
	// images and catalogue references, so deleting the coin before them fails.
	for _, model := range []interface{}{&models.CoinImage{}, &models.CoinReference{}, &models.CoinJournal{}, &models.CoinValueHistory{}, &models.CoinComment{}, &models.AvailabilityResult{}} {
		if err := r.db.Where("coin_id = ?", id).Delete(model).Error; err != nil {
			return 0, err
		}
	}
	result := r.db.Scopes(OwnedByID(id, userID)).Delete(&models.Coin{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// BulkDelete removes multiple coins and all associated data in a single transaction.
func (r *CoinRepository) BulkDelete(coinIDs []uint, userID uint) (int64, error) {
	var rowsAffected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var err error
		rowsAffected, err = r.WithTx(tx).BulkDeleteInTx(coinIDs, userID)
		return err
	})
	return rowsAffected, err
}

func (r *CoinRepository) BulkDeleteInTx(coinIDs []uint, userID uint) (int64, error) {
	var ownedIDs []uint
	if err := r.db.Model(&models.Coin{}).Where("id IN ? AND user_id = ?", coinIDs, userID).Pluck("id", &ownedIDs).Error; err != nil {
		return 0, err
	}
	if len(ownedIDs) == 0 {
		return 0, nil
	}
	// Dependent rows first, as in DeleteInTx.
	for _, model := range []interface{}{&models.CoinImage{}, &models.CoinReference{}, &models.CoinJournal{}, &models.CoinValueHistory{}, &models.CoinComment{}, &models.AvailabilityResult{}, &models.CoinTag{}} {
		if err := r.db.Where("coin_id IN ?", ownedIDs).Delete(model).Error; err != nil {
			return 0, err
		}
	}
	result := r.db.Where("id IN ? AND user_id = ?", ownedIDs, userID).Delete(&models.Coin{})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// BulkMarkSold marks multiple coins as sold in a single transaction.
func (r *CoinRepository) BulkMarkSold(coinIDs []uint, userID uint) (int64, error) {
	result := r.db.Model(&models.Coin{}).
		Where("id IN ? AND user_id = ? AND is_sold = ?", coinIDs, userID, false).
		Updates(map[string]interface{}{
			"is_sold":   true,
			"sold_date": time.Now(),
		})
	return result.RowsAffected, result.Error
}

func (r *CoinRepository) OwnedActiveIDs(coinIDs []uint, userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.Coin{}).
		Where("id IN ? AND user_id = ? AND is_sold = ?", coinIDs, userID, false).
		Pluck("id", &ids).Error
	return ids, err
}

// BulkAssignLocation assigns a storage location to multiple coins. A nil storageLocationID clears the location.
func (r *CoinRepository) BulkAssignLocation(coinIDs []uint, storageLocationID *uint, userID uint) (int64, error) {
	result := r.db.Model(&models.Coin{}).
		Where("id IN ? AND user_id = ?", coinIDs, userID).
		Updates(map[string]interface{}{"storage_location_id": storageLocationID, "storage_slot": nil})
	return result.RowsAffected, result.Error
}

// GetByIDs returns coins matching the given IDs for the given user.
func (r *CoinRepository) GetByIDs(coinIDs []uint, userID uint) ([]models.Coin, error) {
	var coins []models.Coin
	err := r.db.Where("id IN ? AND user_id = ?", coinIDs, userID).
		Preload("Images").Preload("Tags").Preload("References").Preload("StorageLocation").Preload("MintLocation").Find(&coins).Error
	return coins, err
}

// GetStats returns aggregate collection statistics for a user.
func (r *CoinRepository) GetStats(userID uint) (*CollectionStats, error) {
	stats := &CollectionStats{}
	active := ActiveCollection(userID)

	r.db.Model(&models.Coin{}).Scopes(active).Count(&stats.TotalCoins)
	r.db.Model(&models.Coin{}).Where("user_id = ? AND is_wishlist = ?", userID, true).Count(&stats.TotalWishlist)
	r.db.Model(&models.Coin{}).Where("user_id = ? AND is_sold = ?", userID, true).Count(&stats.TotalSold)

	r.db.Model(&models.Coin{}).
		Select("category, count(*) as count").
		Scopes(active).Group("category").
		Scan(&stats.ByCategory)

	r.db.Model(&models.Coin{}).
		Select("material, count(*) as count").
		Scopes(active).Group("material").
		Scan(&stats.ByMaterial)

	r.db.Model(&models.Coin{}).
		Select("grade, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND grade != ''", userID, false, false).
		Group("grade").Order("count DESC").
		Scan(&stats.ByGrade)

	r.db.Model(&models.Coin{}).
		Select("era, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND era != ''", userID, false, false).
		Group("era").Order("count DESC").
		Scan(&stats.ByEra)

	r.db.Model(&models.Coin{}).
		Select("ruler, count(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND ruler != ''", userID, false, false).
		Group("ruler").Order("count DESC").Limit(10).
		Scan(&stats.ByRuler)

	r.db.Model(&models.Coin{}).
		Select(`CASE
			WHEN purchase_price < 50 THEN 'Under $50'
			WHEN purchase_price >= 50 AND purchase_price < 200 THEN '$50 - $200'
			WHEN purchase_price >= 200 AND purchase_price < 500 THEN '$200 - $500'
			WHEN purchase_price >= 500 AND purchase_price < 1000 THEN '$500 - $1K'
			ELSE '$1K+'
		END as `+"`range`"+`, count(*) as count`).
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND purchase_price IS NOT NULL", userID, false, false).
		Group("`range`").
		Scan(&stats.ByPriceRange)

	r.db.Model(&models.Coin{}).
		Select("COALESCE(SUM(purchase_price), 0) as total_purchase_price, COALESCE(SUM(current_value), 0) as total_current_value, COALESCE(AVG(purchase_price), 0) as avg_purchase_price, COALESCE(AVG(current_value), 0) as avg_current_value").
		Scopes(active).
		Scan(&stats.Values)

	r.db.Model(&models.Coin{}).
		Select("COALESCE(SUM(sold_price), 0) as total_sold_price, COALESCE(SUM(purchase_price), 0) as total_purchase_cost").
		Where("user_id = ? AND is_sold = ?", userID, true).
		Scan(&stats.SoldValues)

	return stats, nil
}

// GetDistribution returns era × category cross-tabulation counts for the heat map.
func (r *CoinRepository) GetDistribution(userID uint) ([]DistributionCell, error) {
	var cells []DistributionCell
	err := r.db.Model(&models.Coin{}).
		Select("era, category, COUNT(*) as count").
		Where("user_id = ? AND is_wishlist = ? AND is_sold = ? AND era != '' AND category != ''", userID, false, false).
		Group("era, category").
		Order("era, category").
		Scan(&cells).Error
	return cells, err
}

// validSuggestionColumns is the allowlist of columns permitted in Suggestions queries.
var validSuggestionColumns = map[string]bool{
	"name":              true,
	"denomination":      true,
	"ruler":             true,
	"era":               true,
	"purchase_location": true,
}

// Suggestions returns distinct values for an autocomplete field.
func (r *CoinRepository) Suggestions(userID uint, column string, q string) ([]string, error) {
	if !validSuggestionColumns[column] {
		return nil, fmt.Errorf("invalid suggestion column: %s", column)
	}

	var values []string
	query := r.db.Model(&models.Coin{}).
		Where("user_id = ? AND "+column+" != ''", userID).
		Distinct(column).
		Order(column)

	if q != "" {
		query = query.Where(column+" LIKE ?", "%"+q+"%")
	}

	err := query.Limit(20).Pluck(column, &values).Error
	return values, err
}

// GetCoinValueHistory returns value tracking entries for a specific coin.
func (r *CoinRepository) GetCoinValueHistory(coinID uint, userID uint) ([]models.CoinValueHistory, error) {
	var entries []models.CoinValueHistory
	err := r.db.Where("coin_id = ? AND user_id = ?", coinID, userID).
		Order("recorded_at ASC").
		Find(&entries).Error
	return entries, err
}

// GetWishlistWithURLs returns wishlist coins with a non-empty ReferenceURL for a user.
func (r *CoinRepository) GetWishlistWithURLs(userID uint) ([]models.Coin, error) {
	var coins []models.Coin
	err := r.db.Scopes(OwnedBy(userID)).
		Where("is_wishlist = ? AND reference_url != ''", true).
		Preload("Images").
		Find(&coins).Error
	return coins, err
}

// GetAllWishlistWithURLs returns all users' wishlist coins with a non-empty ReferenceURL.
func (r *CoinRepository) GetAllWishlistWithURLs() ([]models.Coin, error) {
	var coins []models.Coin
	err := r.db.Where("is_wishlist = ? AND reference_url != ''", true).
		Preload("Images").
		Find(&coins).Error
	return coins, err
}

// GetAllWishlistUserIDs returns the distinct user IDs of every owner with at least one
// wishlist coin, including owners whose wishlist coins have no reference URL (zero URLs).
// Used to fan out one child AvailabilityRun per owner for an admin/scheduled cycle, so an
// owner with zero checkable URLs still gets a terminal (zero-coin) run (FR-006).
func (r *CoinRepository) GetAllWishlistUserIDs() ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&models.Coin{}).
		Where("is_wishlist = ?", true).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error
	return userIDs, err
}

// UpdateListingStatus updates only the listing-check fields on a coin.
func (r *CoinRepository) UpdateListingStatus(coinID uint, status, reason string, checkedAt time.Time) error {
	return r.db.Model(&models.Coin{}).Where("id = ?", coinID).
		Updates(map[string]interface{}{
			"listing_status":       status,
			"listing_checked_at":   checkedAt,
			"listing_check_reason": reason,
		}).Error
}

// CoinExists checks if a coin exists for the given user.
func (r *CoinRepository) CoinExists(coinID uint, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Coin{}).Scopes(OwnedByID(coinID, userID)).Count(&count).Error
	return count > 0, err
}

// FindWishlistByReferenceURL returns an owned wishlist coin matching the exact source URL.
func (r *CoinRepository) FindWishlistByReferenceURL(userID uint, referenceURL string) (*models.Coin, error) {
	var coin models.Coin
	err := r.db.Scopes(OwnedBy(userID)).
		Where("is_wishlist = ? AND reference_url = ?", true, referenceURL).
		First(&coin).Error
	if err != nil {
		return nil, err
	}
	return &coin, nil
}

// FindWishlistBySourceAlertCandidateID returns an owned wishlist coin converted from a candidate.
func (r *CoinRepository) FindWishlistBySourceAlertCandidateID(userID uint, candidateID uint) (*models.Coin, error) {
	var coin models.Coin
	err := r.db.Scopes(OwnedBy(userID)).
		Where("is_wishlist = ? AND source_alert_candidate_id = ?", true, candidateID).
		First(&coin).Error
	if err != nil {
		return nil, err
	}
	return &coin, nil
}

// RecordValueHistory creates a coin value history entry.
func (r *CoinRepository) RecordValueHistory(entry *models.CoinValueHistory) error {
	return r.db.Create(entry).Error
}

// CreateJournalEntry creates a journal entry for a coin.
func (r *CoinRepository) CreateJournalEntry(entry *models.CoinJournal) error {
	return r.db.Create(entry).Error
}

// FindMissingCountryWithNumista returns up to limit of the user's coins (or
// just coinID when non-zero) that have no country but are linked to a Numista
// type, plus how many such coins there are in total.
func (r *CoinRepository) FindMissingCountryWithNumista(userID, coinID uint, limit int) ([]models.Coin, int64, error) {
	query := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID)).
		Where("(country IS NULL OR country = '')").
		Where("id IN (SELECT coin_id FROM coin_references WHERE LOWER(catalog) = 'numista')")
	if coinID != 0 {
		query = query.Where("id = ?", coinID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var coins []models.Coin
	err := query.Preload("References").Order("id").Limit(limit).Find(&coins).Error
	return coins, total, err
}

// UpdateCountry sets one coin's country.
func (r *CoinRepository) UpdateCountry(coinID uint, country string) error {
	return r.db.Model(&models.Coin{}).Where("id = ?", coinID).Update("country", country).Error
}

// CountryCount is how many coins the user has from one country.
type CountryCount struct {
	Country string `json:"country"`
	Count   int64  `json:"count"`
}

// CountryCounts lists the countries of the user's collection coins (not
// wishlist or sold) of a kind ("" for all) with how many each has,
// alphabetically.
func (r *CoinRepository) CountryCounts(userID uint, kind string) ([]CountryCount, error) {
	var counts []CountryCount
	err := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID), OfKind(kind)).
		Select("country, COUNT(*) AS count").
		Where("country IS NOT NULL AND country <> '' AND is_wishlist = ? AND is_sold = ?", false, false).
		Group("country").Order("country").Scan(&counts).Error
	return counts, err
}

// withoutImages keeps the user's coins (not banknotes, not sold) with no image at all.
func withoutImages(userID uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(OwnedBy(userID), OfKind(models.CoinKindCoin)).
			Where("is_sold = ? AND NOT EXISTS (SELECT 1 FROM coin_images WHERE coin_images.coin_id = coins.id)", false)
	}
}

// CountWithoutImages counts the user's coins (not banknotes, not sold) with no image.
func (r *CoinRepository) CountWithoutImages(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Coin{}).Scopes(withoutImages(userID)).Count(&count).Error
	return count, err
}

// IDsWithoutImages lists, in order, up to limit of those coins after afterID.
func (r *CoinRepository) IDsWithoutImages(userID, afterID uint, limit int) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.Coin{}).Scopes(withoutImages(userID)).
		Where("coins.id > ?", afterID).Order("coins.id").Limit(limit).Pluck("coins.id", &ids).Error
	return ids, err
}

// DenominationCount is a face value of the collection and how many coins have it.
type DenominationCount struct {
	Denomination string `json:"denomination"`
	Count        int64  `json:"count"`
}

// DenominationCounts lists the face values of the user's collection coins
// (not wishlist or sold) of a kind ("" for every item), the most common first.
func (r *CoinRepository) DenominationCounts(userID uint, kind string) ([]DenominationCount, error) {
	var counts []DenominationCount
	err := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID), OfKind(kind)).
		Select("TRIM(denomination) AS denomination, COUNT(*) AS count").
		Where("TRIM(COALESCE(denomination, '')) <> '' AND is_wishlist = ? AND is_sold = ?", false, false).
		Group("TRIM(denomination)").Order("count DESC, denomination").Scan(&counts).Error
	return counts, err
}

// CoinNumistaType is a collection coin with its country and the Numista type
// it is linked to ("" when none).
type CoinNumistaType struct {
	CoinID        uint
	Country       string
	NumistaTypeID string
	// Circulation is the coin's circulation set by hand, if any.
	Circulation string
}

// CollectionNumistaTypes lists the user's collection coins (not wishlist or
// sold) with their Numista type.
func (r *CoinRepository) CollectionNumistaTypes(userID uint) ([]CoinNumistaType, error) {
	var refs []CoinNumistaType
	err := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID)).
		Select(`coins.id AS coin_id, COALESCE(coins.country, '') AS country, COALESCE(coins.circulation, '') AS circulation,
			COALESCE((SELECT TRIM(r.number) FROM coin_references r
				WHERE r.coin_id = coins.id AND LOWER(r.catalog) = 'numista' ORDER BY r.id LIMIT 1), '') AS numista_type_id`).
		Order("coins.id").Scan(&refs).Error
	return refs, err
}

// DistinctCountries lists the countries of the user's coins (any, wishlist
// and sold included), as written on the coins.
func (r *CoinRepository) DistinctCountries(userID uint) ([]string, error) {
	var countries []string
	err := r.db.Model(&models.Coin{}).Scopes(OwnedBy(userID)).
		Where("TRIM(COALESCE(country, '')) <> ''").
		Distinct().Pluck("TRIM(country)", &countries).Error
	return countries, err
}

// CollectionSummary holds the headline numbers of the user's collection
// (coins not on the wishlist and not sold).
// The counts and the value count every coin held: a coin with a quantity of
// three counts three, and its value three times. UniqueCoins counts each
// different coin once.
type CollectionSummary struct {
	Coins int64 `json:"coins"`
	// UniqueCoins counts each coin once, whatever its quantity.
	UniqueCoins int64 `json:"uniqueCoins"`
	// Types counts distinct coin types: coins linked to the same Numista type
	// count once; other coins count once per distinct name.
	Types          int64 `json:"types"`
	Countries      int64 `json:"countries"`
	WithoutCountry int64 `json:"withoutCountry"`
	// Coincards counts coins whose presentation is "coincard".
	Coincards int64 `json:"coincards"`
	// TotalValue sums the coins' current values times their quantities, in
	// the collection currency.
	TotalValue float64 `json:"totalValue"`
	// UnvaluedCoins have no current value and are left out of TotalValue.
	UnvaluedCoins int64 `json:"unvaluedCoins"`
	// RepeatedCoins are the different coins held more than once; SpareCoins
	// the copies beyond the first, there to trade.
	RepeatedCoins int64 `json:"repeatedCoins"`
	SpareCoins    int64 `json:"spareCoins"`
}

// coinQuantitySQL is how many of a coin are held; rows saved before
// quantities existed count as one.
const coinQuantitySQL = "(CASE WHEN quantity > 0 THEN quantity ELSE 1 END)"

// CollectionQuantities maps the user's collection coins (not wishlist or
// sold) of a kind ("" for all) to how many of each are held.
func (r *CoinRepository) CollectionQuantities(userID uint, kind string) (map[uint]int64, error) {
	var rows []struct {
		ID       uint
		Quantity int64
	}
	err := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID), OfKind(kind)).
		Select("id, " + coinQuantitySQL + " AS quantity").Scan(&rows).Error
	quantities := make(map[uint]int64, len(rows))
	for _, row := range rows {
		quantities[row.ID] = row.Quantity
	}
	return quantities, err
}

// CollectionCoinsOfYears lists the user's collection coins (not wishlist or
// sold) of a kind and of these years, with their catalogue references.
func (r *CoinRepository) CollectionCoinsOfYears(userID uint, kind string, years []int) ([]models.Coin, error) {
	coins := []models.Coin{}
	if len(years) == 0 {
		return coins, nil
	}
	err := r.db.Scopes(ActiveCollection(userID), OfKind(kind)).Preload("References").Where("year IN ?", years).Order("id").Find(&coins).Error
	return coins, err
}

// WishlistCoinsWithReferences lists the user's wishlist coins (not sold) of
// a kind, with their catalogue references.
func (r *CoinRepository) WishlistCoinsWithReferences(userID uint, kind string) ([]models.Coin, error) {
	coins := []models.Coin{}
	err := r.db.Scopes(OwnedBy(userID), OfKind(kind)).Where("is_wishlist = ? AND is_sold = ?", true, false).
		Preload("References").Order("id").Find(&coins).Error
	return coins, err
}

// CollectionCoinsWithReferences lists the user's collection coins (not
// wishlist or sold) of a kind, with their catalogue references and photos.
func (r *CoinRepository) CollectionCoinsWithReferences(userID uint, kind string) ([]models.Coin, error) {
	coins := []models.Coin{}
	err := r.db.Scopes(ActiveCollection(userID), OfKind(kind)).Preload("References").Preload("Images").Order("id").Find(&coins).Error
	return coins, err
}

// DailyCoin picks one of the user's collection coins (not wishlist or sold)
// of a kind by seed, among those with a photo when there are any; nil when
// the collection has none. The same seed picks the same coin.
func (r *CoinRepository) DailyCoin(userID uint, kind string, seed uint32) (*models.Coin, error) {
	var ids []uint
	withPhoto := "EXISTS (SELECT 1 FROM coin_images i WHERE i.coin_id = coins.id)"
	if err := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID), OfKind(kind)).Where(withPhoto).Order("id").Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		if err := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID), OfKind(kind)).Order("id").Pluck("id", &ids).Error; err != nil {
			return nil, err
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var coin models.Coin
	if err := r.db.Preload("Images").Preload("References").First(&coin, ids[seed%uint32(len(ids))]).Error; err != nil {
		return nil, err
	}
	return &coin, nil
}

// CollectionSummary returns the collection's coin, type, country and coincard
// counts and its value, for the items of a kind ("" for all).
func (r *CoinRepository) CollectionSummary(userID uint, kind string) (CollectionSummary, error) {
	var summary CollectionSummary
	err := r.db.Model(&models.Coin{}).Scopes(ActiveCollection(userID), OfKind(kind)).Select(strings.ReplaceAll(`
		COALESCE(SUM({q}), 0) AS coins,
		COUNT(*) AS unique_coins,
		COUNT(DISTINCT COALESCE(
			(SELECT 'numista:' || TRIM(r.number) FROM coin_references r
				WHERE r.coin_id = coins.id AND LOWER(r.catalog) = 'numista' ORDER BY r.id LIMIT 1),
			'name:' || LOWER(TRIM(coins.name))
		)) AS types,
		COUNT(DISTINCT CASE WHEN TRIM(COALESCE(country, '')) <> '' THEN TRIM(country) END) AS countries,
		COALESCE(SUM(CASE WHEN TRIM(COALESCE(country, '')) = '' THEN {q} ELSE 0 END), 0) AS without_country,
		COALESCE(SUM(CASE WHEN presentation = 'coincard' THEN {q} ELSE 0 END), 0) AS coincards,
		COALESCE(SUM(current_value * {q}), 0) AS total_value,
		COALESCE(SUM(CASE WHEN current_value IS NULL THEN {q} ELSE 0 END), 0) AS unvalued_coins,
		COALESCE(SUM(CASE WHEN {q} > 1 THEN 1 ELSE 0 END), 0) AS repeated_coins,
		COALESCE(SUM({q} - 1), 0) AS spare_coins`, "{q}", coinQuantitySQL)).
		Scan(&summary).Error
	return summary, err
}

// CoinHighlight is a short description of a notable coin of the collection.
type CoinHighlight struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Country      string   `json:"country"`
	Year         *int     `json:"year"`
	DateRange    string   `json:"dateRange"`
	Rarity       string   `json:"rarity"`
	Grade        string   `json:"grade"`
	CurrentValue *float64 `json:"currentValue"`
	// ImagePath is the primary (or obverse, or first) image, if any.
	ImagePath string `json:"imagePath,omitempty"`
}

// CollectionHighlights names the rarest, oldest and most valuable coins of the
// collection (not wishlist or sold); each is nil when no coin qualifies.
type CollectionHighlights struct {
	Rarest       *CoinHighlight `json:"rarest"`
	Oldest       *CoinHighlight `json:"oldest"`
	MostValuable *CoinHighlight `json:"mostValuable"`
}

// rarityRankSQL orders the fixed rarity scale, highest = rarest.
const rarityRankSQL = `CASE rarity
	WHEN 'extremely_rare' THEN 6 WHEN 'very_rare' THEN 5 WHEN 'rare' THEN 4
	WHEN 'scarce' THEN 3 WHEN 'uncommon' THEN 2 WHEN 'common' THEN 1 ELSE 0 END`

// Ties go to the more valuable coin (coins without a value last), then the older one.
const highlightTieBreakSQL = "current_value IS NULL, current_value DESC, year IS NULL, year ASC, id ASC"

// CollectionHighlights finds the rarest, oldest and most valuable coins of a
// kind ("" for all).
func (r *CoinRepository) CollectionHighlights(userID uint, kind string) (CollectionHighlights, error) {
	pick := func(where, order string) (*CoinHighlight, error) {
		var coins []models.Coin
		err := r.db.Scopes(ActiveCollection(userID), OfKind(kind)).Where(where).Order(order).Limit(1).
			Preload("Images").Find(&coins).Error
		if err != nil || len(coins) == 0 {
			return nil, err
		}
		return newCoinHighlight(&coins[0]), nil
	}
	var highlights CollectionHighlights
	var err error
	if highlights.Rarest, err = pick(rarityRankSQL+" > 0", rarityRankSQL+" DESC, "+highlightTieBreakSQL); err != nil {
		return highlights, err
	}
	if highlights.Oldest, err = pick("year IS NOT NULL", "year ASC, current_value IS NULL, current_value DESC, id ASC"); err != nil {
		return highlights, err
	}
	if highlights.MostValuable, err = pick("current_value IS NOT NULL AND current_value > 0", "current_value DESC, id ASC"); err != nil {
		return highlights, err
	}
	return highlights, nil
}

func newCoinHighlight(coin *models.Coin) *CoinHighlight {
	highlight := &CoinHighlight{
		ID: coin.ID, Name: coin.Name, Country: coin.Country, Year: coin.Year, DateRange: coin.DateRange,
		Rarity: coin.Rarity, Grade: coin.Grade, CurrentValue: coin.CurrentValue,
	}
	var obverse, first string
	for _, image := range coin.Images {
		if image.IsPrimary {
			highlight.ImagePath = image.FilePath
			return highlight
		}
		if obverse == "" && image.ImageType == models.ImageTypeObverse {
			obverse = image.FilePath
		}
		if first == "" {
			first = image.FilePath
		}
	}
	if obverse != "" {
		highlight.ImagePath = obverse
	} else {
		highlight.ImagePath = first
	}
	return highlight
}
