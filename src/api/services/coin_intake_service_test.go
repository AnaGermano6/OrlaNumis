package services

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var coinIntakeTestDBCounter uint64

func setupCoinIntakeTest(t *testing.T) (*CoinIntakeService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:coin_intake_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), atomic.AddUint64(&coinIntakeTestDBCounter, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.User{}, &models.Coin{}, &models.CoinImage{},
		&models.CoinReference{}, &models.CoinJournal{},
		&models.CoinIntakeDraft{},
	); err != nil {
		t.Fatal(err)
	}
	coinRepo := repository.NewCoinRepository(db)
	coinService := NewCoinService(coinRepo)
	service := NewCoinIntakeService(repository.NewCoinIntakeDraftRepository(db), coinRepo, nil, nil).
		WithCoinService(coinService)
	return service, db
}

func seedCoinIntakeDraft(t *testing.T, db *gorm.DB, userID uint, payload map[string]interface{}) models.CoinIntakeDraft {
	t.Helper()
	encoded, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	draft := models.CoinIntakeDraft{
		UserID:            userID,
		DraftPayload:      string(encoded),
		ConfidenceSummary: `{}`,
		Evidence:          `[]`,
		UnresolvedFields:  `[]`,
		Status:            models.CoinIntakeDraftStatusDrafted,
		ExpiresAt:         time.Now().Add(time.Hour),
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatal(err)
	}
	return draft
}

func TestCoinIntakeService_CommitDraftPersistsMultiWordOverrides(t *testing.T) {
	service, db := setupCoinIntakeTest(t)
	draft := seedCoinIntakeDraft(t, db, 1, map[string]interface{}{
		"name":     "Draft Coin",
		"category": "Modern",
		"material": "Copper",
	})

	result, err := service.CommitDraft(1, IntakeCommitRequest{
		DraftID: draft.ID,
		Confirm: true,
		Overrides: map[string]interface{}{
			"name":               "Roman Imperial As of Vespasian",
			"weightGrams":        10.56,
			"diameterMm":         10.0,
			"obverseInscription": "IMP CAESAR VESPASIAN",
			"reverseInscription": "S C",
			"obverseDescription": "Laureate head of Vespasian right",
			"reverseDescription": "Eagle standing facing on globe",
			"purchasePrice":      120.0,
			"currentValue":       135.0,
			"purchaseDate":       "2026-09-17T00:00:00Z",
			"purchaseLocation":   "VCoins",
		},
	})
	if err != nil {
		t.Fatalf("CommitDraft error: %v", err)
	}

	var coin models.Coin
	if err := db.First(&coin, result.CoinID).Error; err != nil {
		t.Fatal(err)
	}
	if coin.WeightGrams == nil || *coin.WeightGrams != 10.56 {
		t.Fatalf("WeightGrams = %v, want 10.56", coin.WeightGrams)
	}
	if coin.DiameterMm == nil || *coin.DiameterMm != 10 {
		t.Fatalf("DiameterMm = %v, want 10", coin.DiameterMm)
	}
	if coin.ObverseInscription != "IMP CAESAR VESPASIAN" || coin.ReverseInscription != "S C" {
		t.Fatalf("inscriptions = %q / %q", coin.ObverseInscription, coin.ReverseInscription)
	}
	if coin.ObverseDescription != "Laureate head of Vespasian right" || coin.ReverseDescription != "Eagle standing facing on globe" {
		t.Fatalf("descriptions = %q / %q", coin.ObverseDescription, coin.ReverseDescription)
	}
	if coin.PurchasePrice == nil || *coin.PurchasePrice != 120 {
		t.Fatalf("PurchasePrice = %v, want 120", coin.PurchasePrice)
	}
	if coin.CurrentValue == nil || *coin.CurrentValue != 135 {
		t.Fatalf("CurrentValue = %v, want 135", coin.CurrentValue)
	}
	if coin.PurchaseDate == nil || coin.PurchaseDate.UTC().Format(time.DateOnly) != "2026-09-17" {
		t.Fatalf("PurchaseDate = %v, want 2026-09-17", coin.PurchaseDate)
	}
	if coin.PurchaseLocation != "VCoins" {
		t.Fatalf("PurchaseLocation = %q, want VCoins", coin.PurchaseLocation)
	}
}
