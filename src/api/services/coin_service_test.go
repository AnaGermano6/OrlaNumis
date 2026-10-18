package services

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var testDBCounter uint64

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbName := fmt.Sprintf("file:coin_service_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), atomic.AddUint64(&testDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.User{}, &models.Coin{}, &models.CoinImage{}, &models.CoinReference{}, &models.CatalogRegistry{}, &models.AppSetting{},
		&models.MintLocation{},
		&models.CoinJournal{},
		&models.CoinValueHistory{},
		&models.AvailabilityResult{},
		&models.CoinSet{}, &models.CoinSetMembership{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func newTestCoinService(db *gorm.DB) *CoinService {
	repo := repository.NewCoinRepository(db)
	return NewCoinService(repo)
}

func newTestCoinServiceWithCatalogRegistry(db *gorm.DB) *CoinService {
	repo := repository.NewCoinRepository(db)
	catalogRepo := repository.NewCatalogRegistryRepository(db)
	return NewCoinService(repo).WithCatalogRegistrySupport(catalogRepo)
}

func newTestCoinServiceWithSettings(db *gorm.DB) *CoinService {
	repo := repository.NewCoinRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	return NewCoinService(repo).WithSettingsSupport(NewSettingsService(settingsRepo))
}

func newTestCoinServiceWithReferences(db *gorm.DB) *CoinService {
	repo := repository.NewCoinRepository(db)
	catalogRepo := repository.NewCatalogRegistryRepository(db)
	refRepo := repository.NewCoinReferenceRepository(db)
	refSvc := NewCoinReferenceService(refRepo, catalogRepo)
	return NewCoinService(repo).WithReferenceSupport(refRepo, refSvc)
}

func ptrFloat(v float64) *float64 { return &v }

func TestCreateCoin_Success(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{
		Name:          "Denarius of Augustus",
		Category:      models.CategoryModern,
		Material:      models.MaterialSilver,
		PurchasePrice: ptrFloat(250.0),
		CurrentValue:  ptrFloat(300.0),
		UserID:        1,
	}

	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("CreateCoin failed: %v", err)
	}

	if coin.ID == 0 {
		t.Fatal("expected coin ID to be set after create")
	}

	// Verify coin exists in DB
	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found in DB: %v", err)
	}
	if found.Name != "Denarius of Augustus" {
		t.Errorf("expected name 'Denarius of Augustus', got %q", found.Name)
	}
}

func TestCreateCoin_RejectsSameOwnerWishlistReferenceURLDuplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	sourceURL := "https://dealer.example/coin/123"

	first := &models.Coin{
		Name: "First listing", Category: models.CategoryModern, Material: models.MaterialSilver,
		ReferenceURL: sourceURL, IsWishlist: true, UserID: 1,
	}
	if err := svc.CreateCoin(first); err != nil {
		t.Fatal(err)
	}
	duplicate := &models.Coin{
		Name: "Duplicate listing", Category: models.CategoryModern, Material: models.MaterialSilver,
		ReferenceURL: sourceURL, IsWishlist: true, UserID: 1,
	}
	err := svc.CreateCoin(duplicate)
	if !errors.Is(err, ErrWishlistReferenceDuplicate) {
		t.Fatalf("duplicate error=%v, want ErrWishlistReferenceDuplicate", err)
	}
	var withCoin *WishlistDuplicateError
	if !errors.As(err, &withCoin) || withCoin.CoinID != first.ID {
		t.Fatalf("duplicate error=%#v, want the first coin %d", err, first.ID)
	}

	for _, allowed := range []*models.Coin{
		{Name: "Other owner", Category: models.CategoryModern, ReferenceURL: sourceURL, IsWishlist: true, UserID: 2},
		{Name: "URL-less wishlist", Category: models.CategoryModern, IsWishlist: true, UserID: 1},
		{Name: "Collection coin", Category: models.CategoryModern, ReferenceURL: sourceURL, UserID: 1},
	} {
		if err := svc.CreateCoin(allowed); err != nil {
			t.Fatalf("ordinary create was changed for %q: %v", allowed.Name, err)
		}
	}
}

func TestCreateCoin_ConcurrentWishlistReferenceURLCreatesAtMostOne(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	const attempts = 8
	start := make(chan struct{})
	errs := make(chan error, attempts)

	for i := 0; i < attempts; i++ {
		go func(index int) {
			<-start
			errs <- svc.CreateCoin(&models.Coin{
				Name: fmt.Sprintf("Concurrent listing %d", index), Category: models.CategoryModern,
				ReferenceURL: "https://dealer.example/coin/concurrent", IsWishlist: true, UserID: 1,
			})
		}(i)
	}
	close(start)
	for i := 0; i < attempts; i++ {
		<-errs
	}

	var count int64
	if err := db.Model(&models.Coin{}).
		Where("user_id = ? AND is_wishlist = ? AND reference_url = ?", 1, true, "https://dealer.example/coin/concurrent").
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("concurrent wishlist creates persisted %d rows, want 1", count)
	}
}

func TestDuplicateCoin_CopiesFieldsAssociationsAndExcludesMediaAndCardState(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	purchasePrice := 250.0
	currentValue := 375.0
	weight := 3.42
	diameter := 19.5
	now := time.Now().UTC().Truncate(time.Second)
	source := models.Coin{
		Name:                  "Augustus Denarius",
		Category:              models.CategoryModern,
		Denomination:          "Denarius",
		Ruler:                 "Augustus",
		Era:                   models.EraAncient,
		Mint:                  "Lugdunum",
		Material:              models.MaterialSilver,
		WeightGrams:           &weight,
		DiameterMm:            &diameter,
		Grade:                 "VF",
		ObverseInscription:    "CAESAR AVGVSTVS",
		ReverseInscription:    "DIVI F",
		ObverseDescription:    "Laureate head right",
		ReverseDescription:    "Bull butting right",
		RarityRating:          "RIC I 167a",
		PurchasePrice:         &purchasePrice,
		CurrentValue:          &currentValue,
		CurrentValueUpdatedAt: &now,
		PurchaseDate:          &now,
		PurchaseLocation:      "Dealer",
		Notes:                 "Important provenance",
		AIAnalysis:            "Combined analysis",
		ObverseAnalysis:       "Obverse analysis",
		ReverseAnalysis:       "Reverse analysis",
		ReferenceURL:          "https://example.test/ref",
		ReferenceText:         "Reference text",
		ListingStatus:         "available",
		ListingCheckedAt:      &now,
		ListingCheckReason:    "checked",
		UserID:                1,
	}
	if err := db.Create(&source).Error; err != nil {
		t.Fatalf("failed to seed source coin: %v", err)
	}
	if err := db.Create(&models.CoinImage{CoinID: source.ID, FilePath: "coins/source-obverse.jpg", ImageType: models.ImageTypeObverse, IsPrimary: true}).Error; err != nil {
		t.Fatalf("failed to seed image: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: source.ID, Catalog: "RIC", Volume: "I", Number: "167a", URI: "https://example.test/ric"}).Error; err != nil {
		t.Fatalf("failed to seed reference: %v", err)
	}
	set := models.CoinSet{UserID: 1, Name: "Twelve Caesars", SetType: models.CoinSetTypeOpen}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("failed to seed set: %v", err)
	}
	if err := db.Create(&models.CoinSetMembership{CoinID: source.ID, SetID: set.ID, AddedAt: now, SortOrder: 7, Notes: "set note"}).Error; err != nil {
		t.Fatalf("failed to seed set membership: %v", err)
	}

	duplicate, err := svc.DuplicateCoin(source.ID, 1)
	if err != nil {
		t.Fatalf("DuplicateCoin failed: %v", err)
	}

	if duplicate.ID == source.ID {
		t.Fatal("expected duplicate to receive a new ID")
	}
	if duplicate.Name != "Augustus Denarius (duplicate)" {
		t.Fatalf("unexpected duplicate name: %q", duplicate.Name)
	}
	if duplicate.UserID != source.UserID || duplicate.Category != source.Category || duplicate.Material != source.Material || duplicate.Era != source.Era {
		t.Fatalf("duplicate did not preserve owner/category/material/era: %#v", duplicate)
	}
	if duplicate.AIAnalysis != source.AIAnalysis || duplicate.ObverseAnalysis != source.ObverseAnalysis || duplicate.ReverseAnalysis != source.ReverseAnalysis {
		t.Fatalf("duplicate did not preserve analysis fields")
	}
	if duplicate.CurrentValue == nil || *duplicate.CurrentValue != currentValue || duplicate.PurchasePrice == nil || *duplicate.PurchasePrice != purchasePrice {
		t.Fatalf("duplicate did not preserve value fields")
	}
	if len(duplicate.Images) != 0 {
		t.Fatalf("expected images to be excluded, got %#v", duplicate.Images)
	}
	if len(duplicate.References) != 1 || duplicate.References[0].CoinID != duplicate.ID || duplicate.References[0].Catalog != "RIC" {
		t.Fatalf("expected copied reference for duplicate, got %#v", duplicate.References)
	}
	if len(duplicate.Sets) != 1 || duplicate.Sets[0].ID != set.ID {
		t.Fatalf("expected copied set association, got %#v", duplicate.Sets)
	}
	var copiedMembership models.CoinSetMembership
	if err := db.Where("coin_id = ? AND set_id = ?", duplicate.ID, set.ID).First(&copiedMembership).Error; err != nil {
		t.Fatalf("expected copied set membership: %v", err)
	}
	if copiedMembership.SortOrder != 7 || copiedMembership.Notes != "set note" || !copiedMembership.AddedAt.Equal(now) {
		t.Fatalf("membership metadata was not copied correctly: %#v", copiedMembership)
	}
	var imageCount int64
	if err := db.Model(&models.CoinImage{}).Where("coin_id = ?", duplicate.ID).Count(&imageCount).Error; err != nil {
		t.Fatalf("failed to count duplicate images: %v", err)
	}
	if imageCount != 0 {
		t.Fatalf("expected media to be excluded, got images=%d", imageCount)
	}
}

func TestDuplicateCoin_RejectsOtherUsersCoin(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	source := models.Coin{Name: "Private Coin", Category: models.CategoryOther, UserID: 1}
	if err := db.Create(&source).Error; err != nil {
		t.Fatalf("failed to seed source coin: %v", err)
	}

	duplicate, err := svc.DuplicateCoin(source.ID, 2)
	if !repository.IsRecordNotFound(err) {
		t.Fatalf("expected not found for another user's coin, duplicate=%#v err=%v", duplicate, err)
	}
	var count int64
	if err := db.Model(&models.Coin{}).Where("name = ?", "Private Coin (duplicate)").Count(&count).Error; err != nil {
		t.Fatalf("failed to count duplicate coins: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no duplicate to be created, got %d", count)
	}
}

func TestUpdateCoin_RecordsValueHistory(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{
		Name:         "Tetradrachm",
		Category:     models.CategoryOther,
		Material:     models.MaterialSilver,
		CurrentValue: ptrFloat(500.0),
		UserID:       1,
	}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	// Update with a new current value (source != "estimate")
	updates := &models.Coin{CurrentValue: ptrFloat(750.0)}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin failed: %v", err)
	}

	// Verify value history was recorded
	var history []models.CoinValueHistory
	db.Where("coin_id = ?", coin.ID).Find(&history)
	if len(history) != 1 {
		t.Fatalf("expected 1 value history entry, got %d", len(history))
	}
	if history[0].Value != 750.0 {
		t.Errorf("expected history value 750, got %f", history[0].Value)
	}
	if history[0].Confidence != "manual" {
		t.Errorf("expected confidence 'manual', got %q", history[0].Confidence)
	}

	// Verify journal entry was created
	if history[0].Source != models.ValueHistorySourceManual {
		t.Errorf("expected source %q, got %q", models.ValueHistorySourceManual, history[0].Source)
	}

	var journals []models.CoinJournal
	db.Where("coin_id = ?", coin.ID).Find(&journals)
	if len(journals) != 1 {
		t.Fatalf("expected 1 journal entry, got %d", len(journals))
	}
}

func TestUpdateCoin_NormalizesAndReplacesReferences(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithReferences(db)
	if err := db.Create(&models.CatalogRegistry{
		Catalog:        "RIC",
		DisplayName:    "Roman Imperial Coinage",
		Era:            models.EraAncient,
		VolumeRequired: true,
	}).Error; err != nil {
		t.Fatalf("failed to seed catalog registry: %v", err)
	}

	coin := &models.Coin{Name: "Reference Coin", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}
	if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "RIC", Volume: "I", Number: "1"}).Error; err != nil {
		t.Fatalf("failed to seed existing reference: %v", err)
	}

	updates := &models.Coin{
		References: []models.CoinReference{
			{Catalog: " ric ", Volume: " II ", Number: " 12 "},
		},
	}
	if err := svc.UpdateCoinWithFields(coin, updates, []string{}, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoinWithFields failed: %v", err)
	}

	var refs []models.CoinReference
	if err := db.Where("coin_id = ?", coin.ID).Find(&refs).Error; err != nil {
		t.Fatalf("failed to query references: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("expected one replacement reference, got %d", len(refs))
	}
	if refs[0].Catalog != "RIC" || refs[0].Volume != "II" || refs[0].Number != "12" {
		t.Fatalf("expected normalized RIC II 12, got %#v", refs[0])
	}
}

// TestCreateCoin_PersistsReferencesForWishlist asserts the rule that
// wishlist coins are no longer stripped of their References relation.
// IsWishlist records acquisition intent only; it no longer participates in
// reference handling. The incoming reference carries an occupied primary key
// (copied from an unrelated existing coin's reference row) to prove that the
// normal NormalizeAndValidate ID-zeroing path (not a wishlist-specific
// bypass) is what makes this safe — the same defence that always protected
// collection-coin creates.
func TestCreateCoin_PersistsReferencesForWishlist(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithReferences(db)
	if err := db.Create(&models.CatalogRegistry{
		Catalog:     "RSC",
		DisplayName: "Roman Silver Coins",
		Era:         models.EraAncient,
	}).Error; err != nil {
		t.Fatalf("failed to seed catalog registry: %v", err)
	}

	existing := models.Coin{Name: "Existing", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := svc.CreateCoin(&existing); err != nil {
		t.Fatalf("seed existing coin: %v", err)
	}
	existingRef := models.CoinReference{CoinID: existing.ID, Catalog: "RSC", Number: "234"}
	if err := db.Create(&existingRef).Error; err != nil {
		t.Fatalf("seed existing reference: %v", err)
	}

	wishlist := &models.Coin{
		Name:       "Gordian III Denarius",
		Category:   models.CategoryModern,
		Material:   models.MaterialSilver,
		UserID:     1,
		IsWishlist: true,
		References: []models.CoinReference{
			{ID: existingRef.ID, Catalog: "RSC", Number: "234"},
		},
	}
	if err := svc.CreateCoin(wishlist); err != nil {
		t.Fatalf("CreateCoin wishlist with references failed: %v", err)
	}

	if len(wishlist.References) != 1 {
		t.Fatalf("expected wishlist response to carry one normalized reference, got %#v", wishlist.References)
	}
	if wishlist.References[0].ID == existingRef.ID {
		t.Fatalf("expected incoming reference ID to be zeroed and reassigned, got reused ID %d", wishlist.References[0].ID)
	}
	if wishlist.References[0].Catalog != "RSC" || wishlist.References[0].Number != "234" {
		t.Fatalf("expected normalized RSC 234, got %#v", wishlist.References[0])
	}

	var refCount int64
	if err := db.Model(&models.CoinReference{}).Where("coin_id = ?", wishlist.ID).Count(&refCount).Error; err != nil {
		t.Fatalf("failed to count wishlist references: %v", err)
	}
	if refCount != 1 {
		t.Fatalf("expected wishlist coin to persist one reference, got %d", refCount)
	}

	// The pre-existing coin's reference must be untouched — the ID collision
	// must not have mutated or reparented it.
	var unchanged models.CoinReference
	if err := db.First(&unchanged, existingRef.ID).Error; err != nil {
		t.Fatalf("existing reference missing after wishlist create: %v", err)
	}
	if unchanged.CoinID != existing.ID {
		t.Errorf("existing reference CoinID mutated to %d, want %d", unchanged.CoinID, existing.ID)
	}
}

// TestCreateCoin_WishlistAgentStylePayload_PersistsNormalizedReferences proves
// the ADR 0013 rule on the realistic agent path: when CreateCoin receives a
// wishlist coin payload whose References slice carries non-zero IDs (as an
// agent might send from source/database data), the resulting DB row persists
// exactly those references, normalized with fresh IDs — the same outcome a
// collection coin would get. This test bypasses ConvertCandidate and calls
// CreateCoin directly to isolate the service-layer path in
// createPreparedCoinInTx now that the wishlist-specific guard is gone
// (see ADR 0013; FR-049 covers the ConvertCandidate untrusted-source boundary
// separately, in wishlist_search_alert_service_test.go).
func TestCreateCoin_WishlistAgentStylePayload_PersistsNormalizedReferences(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithReferences(db)
	if err := db.Create(&models.CatalogRegistry{
		Catalog:     "RIC",
		DisplayName: "Roman Imperial Coinage",
		Era:         models.EraAncient,
	}).Error; err != nil {
		t.Fatalf("seed catalog: %v", err)
	}

	// Seed a real reference row — its ID represents an occupied primary key
	// that the agent might have looked up from an existing collection coin.
	sourceCoin := &models.Coin{Name: "Source", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1}
	if err := svc.CreateCoin(sourceCoin); err != nil {
		t.Fatalf("seed source coin: %v", err)
	}
	sourceRef := models.CoinReference{CoinID: sourceCoin.ID, Catalog: "RIC", Number: "1"}
	if err := db.Create(&sourceRef).Error; err != nil {
		t.Fatalf("seed source ref: %v", err)
	}

	// Agent-style wishlist payload: IsWishlist=true, References carries a
	// non-zero ID copied from the source coin's reference.
	wishlist := &models.Coin{
		Name:       "Domitian RIC 1",
		Category:   models.CategoryModern,
		Era:        models.EraAncient,
		Material:   models.MaterialSilver,
		UserID:     1,
		IsWishlist: true,
		References: []models.CoinReference{
			{ID: sourceRef.ID, Catalog: "RIC", Number: "1"},
		},
	}
	if err := svc.CreateCoin(wishlist); err != nil {
		t.Fatalf("CreateCoin with agent-style wishlist payload: %v", err)
	}
	if !wishlist.IsWishlist {
		t.Error("result coin must be a wishlist coin")
	}

	// ADR 0013: wishlist coins persist normalized references exactly like
	// collection coins do — no invariant strips them.
	var refCount int64
	if err := db.Model(&models.CoinReference{}).Where("coin_id = ?", wishlist.ID).Count(&refCount).Error; err != nil {
		t.Fatalf("count wishlist refs: %v", err)
	}
	if refCount != 1 {
		t.Errorf("expected 1 persisted reference for wishlist coin, got %d", refCount)
	}

	// The source reference must remain untouched — the ID collision must not
	// have mutated or reparented it (this is what layers 1-3 of the original
	// crash defence still guarantee; see ADR 0013 Context).
	var unchanged models.CoinReference
	if err := db.First(&unchanged, sourceRef.ID).Error; err != nil {
		t.Fatalf("source reference missing after wishlist create: %v", err)
	}
	if unchanged.CoinID != sourceCoin.ID {
		t.Errorf("source reference CoinID mutated to %d, want %d", unchanged.CoinID, sourceCoin.ID)
	}
}

func TestUpdateCoin_EstimateWritesHistoryWithAIEstimateSource(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{
		Name:         "Follis",
		Category:     models.CategoryOther,
		Material:     models.MaterialBronze,
		CurrentValue: ptrFloat(100.0),
		UserID:       1,
	}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{CurrentValue: ptrFloat(120.0)}
	if err := svc.UpdateCoin(coin, updates, 1, "estimate"); err != nil {
		t.Fatalf("UpdateCoin failed: %v", err)
	}

	// source="estimate" should record value history with Source="ai_estimate"
	var history []models.CoinValueHistory
	db.Where("coin_id = ?", coin.ID).Find(&history)
	if len(history) != 1 {
		t.Fatalf("expected 1 value history entry for estimate source, got %d", len(history))
	}
	if history[0].Source != models.ValueHistorySourceAIEstimate {
		t.Errorf("expected Source %q, got %q", models.ValueHistorySourceAIEstimate, history[0].Source)
	}

	// Applied estimate must NOT create a journal entry
	var journals []models.CoinJournal
	db.Where("coin_id = ?", coin.ID).Find(&journals)
	if len(journals) != 0 {
		t.Fatalf("expected 0 journal entries for estimate apply, got %d", len(journals))
	}
}

func TestDeleteCoin_RemovesCoinAndImages(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{
		Name:     "Sestertius",
		Category: models.CategoryModern,
		Material: models.MaterialBronze,
		UserID:   1,
	}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	// Add an image for this coin
	img := models.CoinImage{CoinID: coin.ID, FilePath: "test.jpg", ImageType: models.ImageTypeObverse}
	db.Create(&img)

	rows, err := svc.DeleteCoin(coin.ID, 1)
	if err != nil {
		t.Fatalf("DeleteCoin failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// Coin should be gone
	var count int64
	db.Model(&models.Coin{}).Where("id = ?", coin.ID).Count(&count)
	if count != 0 {
		t.Error("expected coin to be deleted")
	}

	// Image should be gone
	db.Model(&models.CoinImage{}).Where("coin_id = ?", coin.ID).Count(&count)
	if count != 0 {
		t.Error("expected coin image to be deleted")
	}
}

func TestDeleteCoin_WrongUser_NoEffect(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{
		Name:     "Stater",
		Category: models.CategoryOther,
		Material: models.MaterialGold,
		UserID:   1,
	}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	// Try deleting with wrong user ID
	rows, err := svc.DeleteCoin(coin.ID, 999)
	if err != nil {
		t.Fatalf("DeleteCoin failed: %v", err)
	}
	if rows != 0 {
		t.Errorf("expected 0 rows affected for wrong user, got %d", rows)
	}

	// Coin should still exist
	var count int64
	db.Model(&models.Coin{}).Where("id = ?", coin.ID).Count(&count)
	if count != 1 {
		t.Error("expected coin to still exist")
	}
}

func TestPurchaseCoin_UpdatesFields(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	now := time.Now()
	coin := &models.Coin{
		Name:       "Aureus Wishlist",
		Category:   models.CategoryModern,
		Material:   models.MaterialGold,
		IsWishlist: true,
		UserID:     1,
	}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	coin.PurchasePrice = ptrFloat(5000.0)
	coin.PurchaseDate = &now
	coin.PurchaseLocation = "Auction House"

	if err := svc.PurchaseCoin(coin, 1); err != nil {
		t.Fatalf("PurchaseCoin failed: %v", err)
	}

	// Reload and verify
	var updated models.Coin
	db.First(&updated, coin.ID)
	if updated.IsWishlist {
		t.Error("expected IsWishlist to be false after purchase")
	}
	if updated.PurchasePrice == nil || *updated.PurchasePrice != 5000.0 {
		t.Error("expected PurchasePrice to be 5000")
	}
	if updated.PurchaseLocation != "Auction House" {
		t.Errorf("expected PurchaseLocation 'Auction House', got %q", updated.PurchaseLocation)
	}
}

func TestUpdateCoin_AcceptsCustomEraFromCatalogRegistry(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithCatalogRegistry(db)
	if err := db.Create(&models.CatalogRegistry{
		Catalog:     "PROV",
		DisplayName: "Provincial References",
		Era:         models.Era("provincial"),
	}).Error; err != nil {
		t.Fatalf("failed to seed catalog registry: %v", err)
	}

	coin := &models.Coin{Name: "Provincial Bronze", UserID: 1, Era: models.EraAncient}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Era: models.Era("provincial")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin should accept registry-defined era: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Era != models.Era("provincial") {
		t.Fatalf("expected era provincial, got %q", found.Era)
	}
}

func TestUpdateCoin_AcceptsAdminConfiguredEra(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithSettings(db)
	settingsRepo := repository.NewSettingsRepository(db)
	if err := settingsRepo.Upsert(SettingCoinEras, "Republican Rome\nRoman Empire\n500-480 BC"); err != nil {
		t.Fatalf("failed to seed coin eras setting: %v", err)
	}

	coin := &models.Coin{Name: "Imperial Bronze", UserID: 1, Era: models.EraAncient}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Era: models.Era("Roman Empire")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin should accept admin-configured era: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Era != models.Era("Roman Empire") {
		t.Fatalf("expected era Roman Empire, got %q", found.Era)
	}
}

func TestUpdateCoin_RejectsUnsupportedEraWhenCatalogRegistryEnabled(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithCatalogRegistry(db)

	coin := &models.Coin{Name: "Test Coin", UserID: 1, Era: models.EraAncient}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Era: models.Era("unsupported-era")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); !errors.Is(err, ErrCoinInvalidEra) {
		t.Fatalf("expected ErrCoinInvalidEra, got %v", err)
	}
}

func TestUpdateCoin_AcceptsAdminConfiguredCategory(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithSettings(db)
	settingsRepo := repository.NewSettingsRepository(db)
	if err := settingsRepo.Upsert(SettingCoinCategories, "Roman\nGreek\nByzantine\nModern\nOther\nCeltic"); err != nil {
		t.Fatalf("failed to seed coin categories setting: %v", err)
	}

	coin := &models.Coin{Name: "Celtic Stater", UserID: 1, Category: models.CategoryOther}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Category: models.Category("Celtic")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin should accept admin-configured category: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Category != models.Category("Celtic") {
		t.Fatalf("expected category Celtic, got %q", found.Category)
	}
}

func TestUpdateCoin_RejectsUnsupportedCategory(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithSettings(db)

	coin := &models.Coin{Name: "Test Coin", UserID: 1, Category: models.CategoryOther}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Category: models.Category("Not A Real Category")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); !errors.Is(err, ErrCoinInvalidCategory) {
		t.Fatalf("expected ErrCoinInvalidCategory, got %v", err)
	}
}

func TestUpdateCoin_AcceptsBuiltInCategoryEvenIfRemovedFromSetting(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithSettings(db)
	settingsRepo := repository.NewSettingsRepository(db)
	// Admin has edited CoinCategories to no longer list "Roman" - Emperor
	// Tracker's dependency on that literal value must still be honored via
	// the built-in baseline, exactly mirroring how a built-in era remains
	// valid even if removed from CoinEras.
	if err := settingsRepo.Upsert(SettingCoinCategories, "Greek\nByzantine\nModern\nOther"); err != nil {
		t.Fatalf("failed to seed coin categories setting: %v", err)
	}

	coin := &models.Coin{Name: "Denarius", UserID: 1, Category: models.CategoryOther}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("setup: CreateCoin failed: %v", err)
	}

	updates := &models.Coin{Category: models.CategoryModern}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin should accept built-in category Roman regardless of setting contents: %v", err)
	}
}

func TestUpdateCoin_PreservesUnchangedLegacyEra(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinServiceWithCatalogRegistry(db)

	coin := &models.Coin{Name: "Legacy Era Coin", UserID: 1, Era: models.Era("Imperial")}
	if err := db.Create(coin).Error; err != nil {
		t.Fatalf("setup: create legacy era coin failed: %v", err)
	}

	updates := &models.Coin{Name: "Updated Legacy Era Coin", Era: models.Era("Imperial")}
	if err := svc.UpdateCoin(coin, updates, 1, "manual"); err != nil {
		t.Fatalf("UpdateCoin should allow unchanged legacy era: %v", err)
	}

	var found models.Coin
	if err := db.First(&found, coin.ID).Error; err != nil {
		t.Fatalf("coin not found: %v", err)
	}
	if found.Name != "Updated Legacy Era Coin" {
		t.Fatalf("expected updated name, got %q", found.Name)
	}
	if found.Era != models.Era("Imperial") {
		t.Fatalf("expected legacy era to be preserved, got %q", found.Era)
	}
}
