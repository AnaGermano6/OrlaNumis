package services

import (
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCoinRecommendationServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Coin{},
		&models.CoinSet{},
		&models.CoinSetMembership{},
		&models.CoinRecommendation{},
		&models.RecommendationFeedback{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestCoinRecommendationService_ListForCoin_SuggestsSet(t *testing.T) {
	db := setupCoinRecommendationServiceDB(t)
	createTestUserRecord(t, db, 1)

	targetCoin := models.Coin{Name: "Julia Domna denarius", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin1 := models.Coin{Name: "Domna 1", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin2 := models.Coin{Name: "Domna 2", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	if err := db.Create(&[]models.Coin{targetCoin, peerCoin1, peerCoin2}).Error; err != nil {
		t.Fatalf("failed to create coins: %v", err)
	}

	var coins []models.Coin
	if err := db.Order("id ASC").Find(&coins).Error; err != nil {
		t.Fatalf("failed to reload coins: %v", err)
	}
	targetCoin = coins[0]
	peerCoin1 = coins[1]
	peerCoin2 = coins[2]

	set := models.CoinSet{UserID: 1, Name: "Women of the Severan Dynasty", SetType: models.CoinSetTypeGoal}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("failed to create set: %v", err)
	}
	if err := db.Create(&[]models.CoinSetMembership{
		{SetID: set.ID, CoinID: peerCoin1.ID},
		{SetID: set.ID, CoinID: peerCoin2.ID},
	}).Error; err != nil {
		t.Fatalf("failed to create set memberships: %v", err)
	}

	svc := NewCoinRecommendationService(
		repository.NewCoinRecommendationRepository(db),
		repository.NewSetRepository(db),
	)

	recommendations, err := svc.ListForCoin(targetCoin.ID, 1)
	if err != nil {
		t.Fatalf("ListForCoin returned error: %v", err)
	}

	foundSet := false
	for _, rec := range recommendations {
		if rec.TargetType == models.RecommendationTargetTypeSet && rec.TargetID == set.ID {
			foundSet = true
		}
	}
	if !foundSet {
		t.Fatalf("expected set recommendation for set %d", set.ID)
	}
}

func TestCoinRecommendationService_AcceptAndRejectFlow(t *testing.T) {
	db := setupCoinRecommendationServiceDB(t)
	createTestUserRecord(t, db, 1)

	targetCoin := models.Coin{Name: "Julia Domna", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin1 := models.Coin{Name: "Domna 1", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin2 := models.Coin{Name: "Domna 2", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	if err := db.Create(&[]models.Coin{targetCoin, peerCoin1, peerCoin2}).Error; err != nil {
		t.Fatalf("failed to create coins: %v", err)
	}
	var coins []models.Coin
	if err := db.Order("id ASC").Find(&coins).Error; err != nil {
		t.Fatalf("failed to reload coins: %v", err)
	}
	targetCoin = coins[0]
	peerCoin1 = coins[1]
	peerCoin2 = coins[2]

	set := models.CoinSet{UserID: 1, Name: "Severans", SetType: models.CoinSetTypeGoal}
	other := models.CoinSet{UserID: 1, Name: "Julia", SetType: models.CoinSetTypeGoal}
	if err := db.Create(&[]*models.CoinSet{&set, &other}).Error; err != nil {
		t.Fatalf("failed to create sets: %v", err)
	}
	if err := db.Create(&[]models.CoinSetMembership{
		{SetID: set.ID, CoinID: peerCoin1.ID},
		{SetID: set.ID, CoinID: peerCoin2.ID},
		{SetID: other.ID, CoinID: peerCoin1.ID},
		{SetID: other.ID, CoinID: peerCoin2.ID},
	}).Error; err != nil {
		t.Fatalf("failed to create set memberships: %v", err)
	}

	svc := NewCoinRecommendationService(
		repository.NewCoinRecommendationRepository(db),
		repository.NewSetRepository(db))
	recommendations, err := svc.ListForCoin(targetCoin.ID, 1)
	if err != nil {
		t.Fatalf("ListForCoin returned error: %v", err)
	}
	if len(recommendations) == 0 {
		t.Fatal("expected at least one recommendation")
	}

	var setRecID uint
	var otherRecID uint
	for _, rec := range recommendations {
		switch rec.TargetID {
		case set.ID:
			setRecID = rec.ID
		case other.ID:
			otherRecID = rec.ID
		}
	}
	if setRecID == 0 || otherRecID == 0 {
		t.Fatalf("expected both set recommendations; set=%d other=%d", setRecID, otherRecID)
	}

	if err := svc.Accept(targetCoin.ID, setRecID, 1); err != nil {
		t.Fatalf("Accept returned error: %v", err)
	}
	var setMembershipCount int64
	if err := db.Model(&models.CoinSetMembership{}).Where("set_id = ? AND coin_id = ?", set.ID, targetCoin.ID).Count(&setMembershipCount).Error; err != nil {
		t.Fatalf("failed counting set memberships: %v", err)
	}
	if setMembershipCount != 1 {
		t.Fatalf("expected accepted recommendation to add set membership, got %d rows", setMembershipCount)
	}

	if err := svc.Reject(targetCoin.ID, otherRecID, 1); err != nil {
		t.Fatalf("Reject returned error: %v", err)
	}
	var rejected models.CoinRecommendation
	if err := db.First(&rejected, otherRecID).Error; err != nil {
		t.Fatalf("failed loading recommendation: %v", err)
	}
	if rejected.Status != models.RecommendationStatusRejected {
		t.Fatalf("expected recommendation status rejected, got %s", rejected.Status)
	}
}

// TestCoinRecommendationService_ListForCoin_FiltersOutWeakMatches verifies that a
// target whose only overlap with a set is category+era (score=0.40, below
// the medium threshold of 0.45) produces no suggestions.
func TestCoinRecommendationService_ListForCoin_FiltersOutWeakMatches(t *testing.T) {
	db := setupCoinRecommendationServiceDB(t)
	createTestUserRecord(t, db, 1)

	targetCoin := models.Coin{Name: "Target", Ruler: "Julia Domna", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin1 := models.Coin{Name: "Peer 1", Ruler: "Other Ruler", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	peerCoin2 := models.Coin{Name: "Peer 2", Ruler: "Other Ruler", Category: models.CategoryModern, Era: models.EraAncient, UserID: 1}
	if err := db.Create(&[]models.Coin{targetCoin, peerCoin1, peerCoin2}).Error; err != nil {
		t.Fatalf("failed to create coins: %v", err)
	}
	var coins []models.Coin
	if err := db.Order("id ASC").Find(&coins).Error; err != nil {
		t.Fatalf("failed to reload coins: %v", err)
	}
	targetCoin = coins[0]
	peerCoin1 = coins[1]
	peerCoin2 = coins[2]

	set := models.CoinSet{UserID: 1, Name: "Broad Roman Set", SetType: models.CoinSetTypeGoal}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("failed to create set: %v", err)
	}
	if err := db.Create(&[]models.CoinSetMembership{
		{SetID: set.ID, CoinID: peerCoin1.ID},
		{SetID: set.ID, CoinID: peerCoin2.ID},
	}).Error; err != nil {
		t.Fatalf("failed to create set memberships: %v", err)
	}

	svc := NewCoinRecommendationService(
		repository.NewCoinRecommendationRepository(db),
		repository.NewSetRepository(db))
	recommendations, err := svc.ListForCoin(targetCoin.ID, 1)
	if err != nil {
		t.Fatalf("ListForCoin returned error: %v", err)
	}
	if len(recommendations) != 0 {
		t.Fatalf("expected no recommendations (score below medium threshold), got %d", len(recommendations))
	}
}

func createTestUserRecord(t *testing.T, db *gorm.DB, userID uint) {
	t.Helper()
	user := models.User{
		ID:           userID,
		Username:     "tester",
		PasswordHash: "hash",
		Email:        "tester@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
}
