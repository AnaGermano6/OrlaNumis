package repository

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
)

func TestSetRepository_GetCoinsMatchingCriteria_SimpleFiltersAndKindOfCoin(t *testing.T) {
	db := setupTestDB(t)
	year := func(value int) *int { return &value }
	coins := []models.Coin{
		{Name: "Gago Coutinho", UserID: 1, Country: "Portugal", Denomination: "2 euro", Year: year(2022)},
		{Name: "Euro 2004", UserID: 1, Country: "Portugal", Denomination: "1 euro", Year: year(2004)},
		{Name: "Faróis", UserID: 1, Country: "Portugal", Denomination: "5 euro", Year: year(2025), Presentation: "proof_set"},
		{Name: "Marianne", UserID: 1, Country: "France", Denomination: "2 euro", Year: year(2022)},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}
	// The app's classifier (here: names) says which coins are of a kind.
	asked := []string{}
	repo := NewSetRepository(db).WithCoinClassifier(func(userID uint, kind string) ([]uint, error) {
		asked = append(asked, kind)
		switch kind {
		case "commemorative":
			return []uint{coins[0].ID, coins[3].ID}, nil
		case "collector":
			return []uint{coins[2].ID}, nil
		}
		return nil, nil
	})
	match := func(rules ...interface{}) []string {
		t.Helper()
		got, err := repo.GetCoinsMatchingCriteria(1, map[string]interface{}{"operator": "and", "rules": rules})
		if err != nil {
			t.Fatalf("GetCoinsMatchingCriteria failed: %v", err)
		}
		return coinNames(got)
	}
	rule := func(field, op string, value interface{}) map[string]interface{} {
		return map[string]interface{}{"field": field, "op": op, "value": value}
	}

	if names := match(rule("country", "eq", "Portugal"), rule("year", "between", []interface{}{float64(2004), float64(2022)})); !reflect.DeepEqual(names, []string{"Euro 2004", "Gago Coutinho"}) {
		t.Fatalf("Portugal 2004–2022 = %v", names)
	}
	if names := match(rule("denomination", "eq", "2 euro"), rule("classification", "eq", "commemorative")); !reflect.DeepEqual(names, []string{"Gago Coutinho", "Marianne"}) {
		t.Fatalf("2 € commemoratives = %v", names)
	}
	if names := match(rule("presentation", "eq", "proof_set"), rule("classification", "eq", "collector")); !reflect.DeepEqual(names, []string{"Faróis"}) {
		t.Fatalf("proof collector coins = %v", names)
	}
	// A kind no coin is of matches nothing.
	if names := match(rule("classification", "eq", "circulation")); len(names) != 0 {
		t.Fatalf("circulation coins = %v", names)
	}
	if !reflect.DeepEqual(asked, []string{"commemorative", "collector", "circulation"}) {
		t.Fatalf("asked = %v", asked)
	}
}

func TestSetRepository_GetCoinsInSet_UsesManualSortOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "Emperors", SetType: models.CoinSetTypeStandard}
	coins := []models.Coin{
		{Name: "Trajan", UserID: 1},
		{Name: "Augustus", UserID: 1},
		{Name: "Hadrian", UserID: 1},
	}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}

	addedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	memberships := []models.CoinSetMembership{
		{SetID: set.ID, CoinID: coins[0].ID, AddedAt: addedAt, SortOrder: 2},
		{SetID: set.ID, CoinID: coins[1].ID, AddedAt: addedAt, SortOrder: 0},
		{SetID: set.ID, CoinID: coins[2].ID, AddedAt: addedAt, SortOrder: 1},
	}
	if err := db.Create(&memberships).Error; err != nil {
		t.Fatalf("create memberships: %v", err)
	}

	got, err := repo.GetCoinsInSet(set.ID, 1)
	if err != nil {
		t.Fatalf("GetCoinsInSet failed: %v", err)
	}
	names := coinNames(got)
	want := []string{"Augustus", "Hadrian", "Trajan"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("expected order %v, got %v", want, names)
	}
}

func TestSetRepository_GetCoinsInSet_DefaultSortOrderFallsBackToName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "Emperors", SetType: models.CoinSetTypeStandard}
	coins := []models.Coin{
		{Name: "Trajan", UserID: 1},
		{Name: "Augustus", UserID: 1},
		{Name: "Hadrian", UserID: 1},
	}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}

	addedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	memberships := []models.CoinSetMembership{
		{SetID: set.ID, CoinID: coins[0].ID, AddedAt: addedAt},
		{SetID: set.ID, CoinID: coins[1].ID, AddedAt: addedAt},
		{SetID: set.ID, CoinID: coins[2].ID, AddedAt: addedAt},
	}
	if err := db.Create(&memberships).Error; err != nil {
		t.Fatalf("create memberships: %v", err)
	}

	got, err := repo.GetCoinsInSet(set.ID, 1)
	if err != nil {
		t.Fatalf("GetCoinsInSet failed: %v", err)
	}
	names := coinNames(got)
	want := []string{"Augustus", "Hadrian", "Trajan"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("expected fallback order %v, got %v", want, names)
	}
}

func TestSetRepository_ReorderCoinsInSet_RejectsInvalidMembersWithoutPartialUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "Emperors", SetType: models.CoinSetTypeStandard}
	memberA := models.Coin{Name: "Augustus", UserID: 1}
	memberB := models.Coin{Name: "Trajan", UserID: 1}
	nonMember := models.Coin{Name: "Nero", UserID: 1}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	if err := db.Create(&[]*models.Coin{&memberA, &memberB, &nonMember}).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}

	addedAt := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	memberships := []models.CoinSetMembership{
		{SetID: set.ID, CoinID: memberA.ID, AddedAt: addedAt, SortOrder: 0},
		{SetID: set.ID, CoinID: memberB.ID, AddedAt: addedAt, SortOrder: 1},
	}
	if err := db.Create(&memberships).Error; err != nil {
		t.Fatalf("create memberships: %v", err)
	}

	err := repo.ReorderCoinsInSet(set.ID, 1, []uint{memberB.ID, nonMember.ID})
	if !errors.Is(err, ErrInvalidSetOrder) {
		t.Fatalf("expected ErrInvalidSetOrder, got %v", err)
	}

	got, err := repo.GetCoinsInSet(set.ID, 1)
	if err != nil {
		t.Fatalf("GetCoinsInSet failed: %v", err)
	}
	names := coinNames(got)
	want := []string{"Augustus", "Trajan"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("order changed after rejected reorder: want %v, got %v", want, names)
	}
}

func TestSetRepository_GetAgenticSetCompletionRequiresExplicitAssignments(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "US Silver Quarters", SetType: models.CoinSetTypeAgentic, CreationMode: models.CoinSetCreationModeDynamic}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	coins := []models.Coin{
		{Name: "1940 Washington Quarter A", UserID: 1, Denomination: "Quarter", Material: "Silver"},
		{Name: "1940 Washington Quarter B", UserID: 1, Denomination: "Quarter", Material: "Silver"},
		{Name: "1941 Washington Quarter", UserID: 1, Denomination: "Quarter", Material: "Silver", IsWishlist: true},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}
	year1940 := 1940
	year1941 := 1941
	denomination := "Quarter"
	material := "Silver"
	targets := []models.CoinSetTarget{
		{SetID: set.ID, Label: "1940 US Silver Quarter", Year: &year1940, Denomination: &denomination, Material: &material, SortOrder: 1},
		{SetID: set.ID, Label: "Duplicate 1940 Slot", Year: &year1940, Denomination: &denomination, Material: &material, SortOrder: 2},
		{SetID: set.ID, Label: "1941 US Silver Quarter", Year: &year1941, Denomination: &denomination, Material: &material, SortOrder: 3},
	}
	if err := db.Create(&targets).Error; err != nil {
		t.Fatalf("create targets: %v", err)
	}

	completion, err := repo.GetSetCompletion(set.ID, 1)
	if err != nil {
		t.Fatalf("GetSetCompletion: %v", err)
	}
	if completion["completedTargets"] != 0 {
		t.Fatalf("expected no implicit auto-matching for agentic sets, got %#v", completion)
	}
	missing, ok := completion["missingTargets"].([]models.CoinSetTarget)
	if !ok || len(missing) != 3 {
		t.Fatalf("expected all unassigned targets to be missing, missing=%#v", completion["missingTargets"])
	}
	matches, ok := completion["targetMatches"].([]AgenticTargetMatch)
	if !ok || len(matches) != 3 {
		t.Fatalf("expected target matches in completion, got %#v", completion["targetMatches"])
	}
	if matches[0].Coin != nil || matches[1].Coin != nil || matches[2].Coin != nil {
		t.Fatalf("expected no auto-assigned coins in agentic completion, got %#v", matches)
	}
}

func TestSetRepository_AddCoinToSet_AssignsAgenticSlots(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "Byzantine Emperors", SetType: models.CoinSetTypeAgentic, CreationMode: models.CoinSetCreationModeDynamic}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}
	coins := []models.Coin{
		{Name: "Anastasius Solidus", UserID: 1, Denomination: "Solidus", Material: "Gold"},
		{Name: "Justin I Solidus", UserID: 1, Denomination: "Solidus", Material: "Gold"},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}
	targets := []models.CoinSetTarget{
		{SetID: set.ID, Label: "Anastasius I", SortOrder: 1},
		{SetID: set.ID, Label: "Justin I", SortOrder: 2},
	}
	if err := db.Create(&targets).Error; err != nil {
		t.Fatalf("create targets: %v", err)
	}

	if err := repo.AddCoinToSet(coins[0].ID, set.ID, 1, "", targets[0].ID); err != nil {
		t.Fatalf("assign first slot: %v", err)
	}
	if err := repo.AddCoinToSet(coins[0].ID, set.ID, 1, "", targets[1].ID); err != nil {
		t.Fatalf("move coin to second slot: %v", err)
	}
	if err := repo.AddCoinToSet(coins[1].ID, set.ID, 1, "", targets[1].ID); err != nil {
		t.Fatalf("replace second slot coin: %v", err)
	}

	completion, err := repo.GetSetCompletion(set.ID, 1)
	if err != nil {
		t.Fatalf("GetSetCompletion: %v", err)
	}
	if got := completion["completedTargets"].(int); got != 1 {
		t.Fatalf("expected one completed target after replacement, got %d", got)
	}
	matches := completion["targetMatches"].([]AgenticTargetMatch)
	if matches[0].Coin != nil {
		t.Fatalf("first slot should be empty after coin reassignment, got %#v", matches[0].Coin)
	}
	if matches[1].Coin == nil || matches[1].Coin.ID != coins[1].ID {
		t.Fatalf("second slot should contain replacement coin, got %#v", matches[1].Coin)
	}
}

func coinNames(coins []models.Coin) []string {
	names := make([]string, 0, len(coins))
	for _, coin := range coins {
		names = append(names, coin.Name)
	}
	return names
}

func TestSetRepository_GetSetCompletion_GoalUsesCollectionVsWishlistMembers(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSetRepository(db)

	set := models.CoinSet{UserID: 1, Name: "Goal Set", SetType: models.CoinSetTypeGoal}
	if err := db.Create(&set).Error; err != nil {
		t.Fatalf("create set: %v", err)
	}

	coins := []models.Coin{
		{Name: "Owned 1", UserID: 1, IsWishlist: false},
		{Name: "Owned 2", UserID: 1, IsWishlist: false},
		{Name: "Wish 1", UserID: 1, IsWishlist: true},
		{Name: "Wish 2", UserID: 1, IsWishlist: true},
		{Name: "Wish 3", UserID: 1, IsWishlist: true},
		{Name: "Wish 4", UserID: 1, IsWishlist: true},
		{Name: "Wish 5", UserID: 1, IsWishlist: true},
	}
	if err := db.Create(&coins).Error; err != nil {
		t.Fatalf("create coins: %v", err)
	}
	for i := range coins {
		if err := db.Create(&models.CoinSetMembership{
			SetID:     set.ID,
			CoinID:    coins[i].ID,
			AddedAt:   time.Now(),
			SortOrder: i,
		}).Error; err != nil {
			t.Fatalf("create membership: %v", err)
		}
	}

	completion, err := repo.GetSetCompletion(set.ID, 1)
	if err != nil {
		t.Fatalf("GetSetCompletion failed: %v", err)
	}

	if got := completion["totalTargets"].(int); got != 7 {
		t.Fatalf("expected totalTargets=7, got %d", got)
	}
	if got := completion["completedTargets"].(int); got != 2 {
		t.Fatalf("expected completedTargets=2, got %d", got)
	}
	if got := completion["collectionItems"].(int); got != 2 {
		t.Fatalf("expected collectionItems=2, got %d", got)
	}
	if got := completion["wishlistItems"].(int); got != 5 {
		t.Fatalf("expected wishlistItems=5, got %d", got)
	}
	pct := completion["completionPercentage"].(float64)
	if math.Abs(pct-28.5714285714) > 0.001 {
		t.Fatalf("expected completionPercentage≈28.57, got %.6f", pct)
	}
}
