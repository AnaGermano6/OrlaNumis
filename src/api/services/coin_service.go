package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"gorm.io/gorm"
)

var (
	ErrCoinInvalidEra             = errors.New("era is not supported")
	ErrCoinInvalidCategory        = errors.New("category is not supported")
	ErrCoinInvalidPresentation    = errors.New("presentation is not supported")
	ErrCoinInvalidRarity          = errors.New("rarity is not supported")
	ErrCoinInvalidCirculation     = errors.New("circulation is not supported")
	ErrCoinInvalidKind            = errors.New("kind must be coin or banknote")
	ErrCoinInvalidNoteVariant     = errors.New("note variant must be replacement or specimen")
	ErrCoinInvalidQuantity        = errors.New("quantity must be between 1 and 9999")
	ErrWishlistReferenceDuplicate = errors.New("a wishlist coin already uses this source URL")
)

// WishlistDuplicateError is ErrWishlistReferenceDuplicate with the wishlist
// coin that already uses the source URL.
type WishlistDuplicateError struct {
	CoinID uint
}

func (e *WishlistDuplicateError) Error() string { return ErrWishlistReferenceDuplicate.Error() }

// Is makes errors.Is(err, ErrWishlistReferenceDuplicate) hold.
func (e *WishlistDuplicateError) Is(target error) bool {
	return target == ErrWishlistReferenceDuplicate
}

var builtInCoinEras = map[models.Era]struct{}{
	models.EraAncient:  {},
	models.EraMedieval: {},
	models.EraModern:   {},
}

var builtInCoinCategories = map[models.Category]struct{}{
	models.CategoryModern: {},
	models.CategoryOther:  {},
}

// BuiltInCoinCategoryValues returns the built-in category baseline as a
// plain string slice, for callers (e.g. AI collection search) that need to
// merge it with the admin-defined CoinCategories list.
func BuiltInCoinCategoryValues() []string {
	values := make([]string, 0, len(builtInCoinCategories))
	for k := range builtInCoinCategories {
		values = append(values, string(k))
	}
	return values
}

// BuiltInCoinEraValues is the Era equivalent of BuiltInCoinCategoryValues.
func BuiltInCoinEraValues() []string {
	values := make([]string, 0, len(builtInCoinEras))
	for k := range builtInCoinEras {
		values = append(values, string(k))
	}
	return values
}

// CoinService handles coin business logic and orchestrates repository calls.
type CoinService struct {
	repo                *repository.CoinRepository
	refRepo             *repository.CoinReferenceRepository
	refSvc              *CoinReferenceService
	mintLocationRepo    *repository.MintLocationRepository
	catalogRegistryRepo *repository.CatalogRegistryRepository
	settingsSvc         *SettingsService
}

type preparedCoinCreator struct {
	service                       *CoinService
	coin                          *models.Coin
	allowAcknowledgedURLDuplicate bool
}

func (c preparedCoinCreator) CreateCoinInTransaction(tx *gorm.DB) (uint, error) {
	if err := c.service.createPreparedCoinInTx(tx, c.coin, c.allowAcknowledgedURLDuplicate); err != nil {
		return 0, err
	}
	return c.coin.ID, nil
}

// NewCoinService creates a new CoinService.
func NewCoinService(repo *repository.CoinRepository) *CoinService {
	return &CoinService{repo: repo}
}

// WithReferenceSupport enables structured reference orchestration during coin create/update workflows.
func (s *CoinService) WithReferenceSupport(
	refRepo *repository.CoinReferenceRepository,
	refSvc *CoinReferenceService,
) *CoinService {
	s.refRepo = refRepo
	s.refSvc = refSvc
	return s
}

// WithMintLocationSupport enables mint-location visibility validation during coin create/update workflows.
func (s *CoinService) WithMintLocationSupport(mintLocationRepo *repository.MintLocationRepository) *CoinService {
	s.mintLocationRepo = mintLocationRepo
	return s
}

// WithCatalogRegistrySupport enables data-driven coin era validation.
func (s *CoinService) WithCatalogRegistrySupport(catalogRegistryRepo *repository.CatalogRegistryRepository) *CoinService {
	s.catalogRegistryRepo = catalogRegistryRepo
	return s
}

// WithSettingsSupport enables validation against admin-configured coin properties.
func (s *CoinService) WithSettingsSupport(settingsSvc *SettingsService) *CoinService {
	s.settingsSvc = settingsSvc
	return s
}

func (s *CoinService) PreparedCoinCreator(coin *models.Coin) repository.TransactionalCoinCreator {
	return preparedCoinCreator{service: s, coin: coin}
}

func (s *CoinService) PreparedCoinCreatorWithAcknowledgedURLDuplicate(coin *models.Coin) repository.TransactionalCoinCreator {
	return preparedCoinCreator{service: s, coin: coin, allowAcknowledgedURLDuplicate: true}
}

// CreateCoin creates a coin and records a value snapshot in a single transaction.
func (s *CoinService) CreateCoin(coin *models.Coin) error {
	if err := s.prepareCoinForCreate(coin); err != nil {
		return err
	}
	return s.createPreparedCoinInTx(nil, coin, false)
}

// CreateCoinInTx creates a coin using the same validation path as CreateCoin.
// When tx is provided, the write and value snapshot are part of that caller-owned transaction.
func (s *CoinService) CreateCoinInTx(tx *gorm.DB, coin *models.Coin) error {
	if err := s.prepareCoinForCreate(coin); err != nil {
		return err
	}
	return s.createPreparedCoinInTx(tx, coin, false)
}

func (s *CoinService) prepareCoinForCreate(coin *models.Coin) error {
	if err := s.validateMintLocation(coin.MintLocationID, coin.UserID); err != nil {
		return err
	}
	coin.Era = models.Era(strings.TrimSpace(string(coin.Era)))
	if err := s.validateCoinEra(coin.Era); err != nil {
		return err
	}
	coin.Category = models.Category(strings.TrimSpace(string(coin.Category)))
	if err := s.validateCoinCategory(coin.Category); err != nil {
		return err
	}
	coin.ReferenceURL = strings.TrimSpace(coin.ReferenceURL)
	if err := normalizeCoinAttributes(coin); err != nil {
		return err
	}
	// A year typed by the user wins; otherwise it is read from the date.
	if coin.Year == nil {
		coin.Year = models.DeriveCoinYear(coin.DateRange)
	}
	// A new coin or note is one unless told otherwise.
	if coin.Quantity == 0 {
		coin.Quantity = 1
	}
	if !models.ValidCoinQuantity(coin.Quantity) {
		return ErrCoinInvalidQuantity
	}
	return nil
}

// normalizeCoinAttributes tidies the country and year and checks presentation, rarity and circulation.
func normalizeCoinAttributes(coin *models.Coin) error {
	// There is no year 0 between 1 BC and AD 1; treat it as "not given".
	if coin.Year != nil && *coin.Year == 0 {
		coin.Year = nil
	}
	coin.Country = models.NormalizeCoinCountry(coin.Country)
	coin.MonetarySystem = strings.Join(strings.Fields(coin.MonetarySystem), " ")
	coin.Presentation = strings.TrimSpace(coin.Presentation)
	if !models.ValidCoinPresentation(coin.Presentation) {
		return ErrCoinInvalidPresentation
	}
	coin.Rarity = strings.TrimSpace(coin.Rarity)
	if !models.ValidCoinRarity(coin.Rarity) {
		return ErrCoinInvalidRarity
	}
	coin.Circulation = strings.TrimSpace(coin.Circulation)
	if !models.ValidCoinCirculation(coin.Circulation) {
		return ErrCoinInvalidCirculation
	}
	if coin.Kind = strings.TrimSpace(coin.Kind); coin.Kind == "" {
		coin.Kind = models.CoinKindCoin
	}
	if !models.ValidCoinKind(coin.Kind) {
		return ErrCoinInvalidKind
	}
	coin.NoteVariant = strings.TrimSpace(coin.NoteVariant)
	if !models.ValidNoteVariant(coin.NoteVariant) {
		return ErrCoinInvalidNoteVariant
	}
	for _, field := range []*string{&coin.Series, &coin.SerialNumber, &coin.Signatures, &coin.Printer, &coin.Watermark, &coin.CertifiedGrade, &coin.CertificationNumber} {
		*field = strings.TrimSpace(*field)
	}
	return nil
}

func (s *CoinService) createPreparedCoinInTx(tx *gorm.DB, coin *models.Coin, allowAcknowledgedURLDuplicate bool) error {
	create := func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		if coin.IsWishlist && coin.ReferenceURL != "" && !allowAcknowledgedURLDuplicate {
			if existing, err := txRepo.FindWishlistByReferenceURL(coin.UserID, coin.ReferenceURL); err == nil {
				return &WishlistDuplicateError{CoinID: existing.ID}
			} else if !repository.IsRecordNotFound(err) {
				return err
			}
		}
		pendingReferences := coin.References
		coin.References = nil
		if err := txRepo.Create(coin); err != nil {
			return err
		}

		if s.refRepo != nil && s.refSvc != nil && pendingReferences != nil {
			normalized, err := s.refSvc.NormalizeAndValidate(pendingReferences)
			if err != nil {
				return err
			}

			for i := range normalized {
				normalized[i].CoinID = coin.ID
			}

			txRefRepo := s.refRepo.WithTx(tx)
			if err := txRefRepo.CreateBatch(normalized); err != nil {
				return err
			}

			coin.References = normalized
		}

		return nil
	}
	if tx != nil {
		return create(tx)
	}
	return s.repo.DB().Transaction(create)
}

// UpdateCoin applies updates to an existing coin. If the current value changed
// and the source is not "estimate", it records a value history entry and a
// journal entry. A value snapshot is always recorded afterward.
func (s *CoinService) UpdateCoin(existing *models.Coin, updates *models.Coin, userID uint, source string) error {
	return s.updateCoin(existing, updates, nil, userID, source)
}

// UpdateCoinWithFields applies a presence-aware update. Only fields named in
// updateFields are persisted, allowing explicit zero values to be saved while
// omitted request fields preserve existing values.
func (s *CoinService) UpdateCoinWithFields(existing *models.Coin, updates *models.Coin, updateFields []string, userID uint, source string) error {
	return s.updateCoin(existing, updates, updateFields, userID, source)
}

func (s *CoinService) updateCoin(existing *models.Coin, updates *models.Coin, updateFields []string, userID uint, source string) error {
	oldValue := existing.CurrentValue
	if updateFields == nil || containsString(updateFields, "MintLocationID") {
		if err := s.validateMintLocation(updates.MintLocationID, userID); err != nil {
			return err
		}
	}
	eraProvided := updateFields == nil || containsString(updateFields, "Era")
	if eraProvided {
		updates.Era = models.Era(strings.TrimSpace(string(updates.Era)))
	}
	existingEra := models.Era(strings.TrimSpace(string(existing.Era)))
	if eraProvided && updates.Era != existingEra {
		if err := s.validateCoinEra(updates.Era); err != nil {
			return err
		}
	}
	if err := normalizeCoinAttributes(updates); err != nil {
		return err
	}
	// Without named fields, a zero quantity was not given and is not written.
	quantityGiven := containsString(updateFields, "Quantity") || (updateFields == nil && updates.Quantity != 0)
	if quantityGiven && !models.ValidCoinQuantity(updates.Quantity) {
		return ErrCoinInvalidQuantity
	}
	yearProvided := updateFields == nil || containsString(updateFields, "Year")
	dateProvided := updateFields == nil || containsString(updateFields, "DateRange")
	switch {
	case yearProvided && updates.Year == nil:
		// An empty year means "read it from the date".
		dateRange := existing.DateRange
		if dateProvided {
			dateRange = updates.DateRange
		}
		updates.Year = models.DeriveCoinYear(dateRange)
	case !yearProvided && dateProvided:
		// The date changed without a year: keep the year in step with it.
		updates.Year = models.DeriveCoinYear(updates.DateRange)
		updateFields = append(updateFields, "Year")
	}
	categoryProvided := updateFields == nil || containsString(updateFields, "Category")
	if categoryProvided {
		updates.Category = models.Category(strings.TrimSpace(string(updates.Category)))
	}
	existingCategory := models.Category(strings.TrimSpace(string(existing.Category)))
	if categoryProvided && updates.Category != existingCategory {
		if err := s.validateCoinCategory(updates.Category); err != nil {
			return err
		}
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if updateFields == nil || len(updateFields) > 0 {
			if err := txRepo.Update(existing, updates, updateFields...); err != nil {
				return err
			}
		}
		if s.refRepo != nil && s.refSvc != nil && updates.References != nil {
			normalized, err := s.refSvc.NormalizeAndValidate(updates.References)
			if err != nil {
				return err
			}

			txRefRepo := s.refRepo.WithTx(tx)
			if err := txRefRepo.ReplaceForCoin(existing.ID, userID, normalized); err != nil {
				return err
			}

			existing.References = normalized
		}

		currentValueProvided := updateFields == nil || containsString(updateFields, "CurrentValue")
		if currentValueProvided && updates.CurrentValue != nil {
			newVal := *updates.CurrentValue
			oldVal := 0.0
			if oldValue != nil {
				oldVal = *oldValue
			}
			if newVal != oldVal {
				// Update CurrentValueUpdatedAt whenever CurrentValue changes
				now := time.Now()
				existing.CurrentValueUpdatedAt = &now
				if err := txRepo.UpdateField(existing, "current_value_updated_at", now); err != nil {
					return err
				}

				if source == "estimate" {
					// Applied on-demand AI estimate: persist to value history, no journal.
					if err := txRepo.RecordValueHistory(&models.CoinValueHistory{
						CoinID:     existing.ID,
						UserID:     userID,
						Value:      newVal,
						Confidence: "",
						Source:     models.ValueHistorySourceAIEstimate,
						RecordedAt: now,
					}); err != nil {
						return err
					}
				} else {
					if err := txRepo.RecordValueHistory(&models.CoinValueHistory{
						CoinID:     existing.ID,
						UserID:     userID,
						Value:      newVal,
						Confidence: "manual",
						Source:     models.ValueHistorySourceManual,
						RecordedAt: now,
					}); err != nil {
						return err
					}
					if err := txRepo.CreateJournalEntry(&models.CoinJournal{
						CoinID: existing.ID,
						UserID: userID,
						Entry:  fmt.Sprintf("Current value updated manually: $%.2f", newVal),
					}); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

// DeleteCoin deletes a coin and records a value snapshot if rows were affected.
// Returns the number of rows affected.
func (s *CoinService) DeleteCoin(id, userID uint) (int64, error) {
	var rows int64
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		rows, err = txRepo.DeleteInTx(id, userID)
		return err
	})
	return rows, err
}

// DuplicateCoin clones an owned coin without carrying over image/card-related rows.
func (s *CoinService) DuplicateCoin(id, userID uint) (*models.Coin, error) {
	var duplicate *models.Coin
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		created, err := txRepo.Duplicate(id, userID)
		if err != nil {
			return err
		}
		duplicate = created
		return nil
	})
	return duplicate, err
}

// PurchaseCoin marks a wishlist coin as purchased and records a value snapshot.
// The coin's purchase fields (price, date, location) should be set on the model
// before calling this method.
func (s *CoinService) PurchaseCoin(coin *models.Coin, userID uint) error {
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		updates := map[string]interface{}{
			"is_wishlist":       false,
			"purchase_price":    coin.PurchasePrice,
			"purchase_date":     coin.PurchaseDate,
			"purchase_location": coin.PurchaseLocation,
		}
		return txRepo.UpdateFields(coin, updates)
	})
}

func (s *CoinService) BulkDeleteCoins(coinIDs []uint, userID uint) (int64, error) {
	var affected int64
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		var err error
		affected, err = txRepo.BulkDeleteInTx(coinIDs, userID)
		return err
	})
	return affected, err
}

func (s *CoinService) validateMintLocation(mintLocationID *uint, userID uint) error {
	if mintLocationID == nil || s.mintLocationRepo == nil {
		return nil
	}
	if *mintLocationID == 0 {
		return ErrMintLocationNotFound
	}
	exists, err := s.mintLocationRepo.ExistsVisibleTo(*mintLocationID, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrMintLocationNotFound
	}
	return nil
}

func (s *CoinService) validateCoinEra(era models.Era) error {
	trimmed := strings.TrimSpace(string(era))
	if trimmed == "" {
		return nil
	}
	normalized := models.Era(trimmed)
	if _, ok := builtInCoinEras[normalized]; ok {
		return nil
	}
	if len(trimmed) > 64 {
		return ErrCoinInvalidEra
	}
	if s.settingsSvc != nil && settingListContains(s.settingsSvc.GetSetting(SettingCoinEras), trimmed) {
		return nil
	}
	if s.catalogRegistryRepo == nil {
		return ErrCoinInvalidEra
	}
	exists, err := s.catalogRegistryRepo.EraExists(normalized)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCoinInvalidEra
	}
	return nil
}

// validateCoinCategory mirrors validateCoinEra: built-in categories are
// always accepted regardless of the CoinCategories admin setting (so an
// admin edit to that setting never invalidates them), custom categories are
// accepted if they appear in the setting's admin-defined list. Unlike Era,
// there is no Catalog Registry fallback tier for Category — no such concept
// exists there.
func (s *CoinService) validateCoinCategory(category models.Category) error {
	trimmed := strings.TrimSpace(string(category))
	if trimmed == "" {
		return nil
	}
	normalized := models.Category(trimmed)
	if _, ok := builtInCoinCategories[normalized]; ok {
		return nil
	}
	if len(trimmed) > 64 {
		return ErrCoinInvalidCategory
	}
	if s.settingsSvc != nil && settingListContains(s.settingsSvc.GetSetting(SettingCoinCategories), trimmed) {
		return nil
	}
	return ErrCoinInvalidCategory
}

func settingListContains(value, needle string) bool {
	for _, line := range strings.Split(value, "\n") {
		if strings.EqualFold(strings.TrimSpace(line), needle) {
			return true
		}
	}
	return false
}
