package services

import (
	"context"
	"errors"
	"testing"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestCreateCoin_NormalizesCountryAndDerivesYear(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	coin := &models.Coin{
		Name: "50 Escudos", Category: models.CategoryModern, Material: models.MaterialSilver, UserID: 1,
		DateRange: "1968", Country: "  Portugal ", Presentation: "coincard", Rarity: "rare",
	}

	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("CreateCoin: %v", err)
	}

	var found models.Coin
	db.First(&found, coin.ID)
	if found.Country != "Portugal" || found.Year == nil || *found.Year != 1968 || found.Presentation != "coincard" || found.Rarity != "rare" {
		t.Fatalf("stored %+v (year %v)", found, found.Year)
	}
}

func TestCreateCoin_RejectsUnknownPresentationAndRarity(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	err := svc.CreateCoin(&models.Coin{Name: "A", Category: models.CategoryModern, UserID: 1, Presentation: "shoebox"})
	if !errors.Is(err, ErrCoinInvalidPresentation) {
		t.Fatalf("presentation: %v", err)
	}
	err = svc.CreateCoin(&models.Coin{Name: "B", Category: models.CategoryModern, UserID: 1, Rarity: "R2"})
	if !errors.Is(err, ErrCoinInvalidRarity) {
		t.Fatalf("rarity: %v", err)
	}
}

func TestCreateCoin_KeepsACirculationSetByHandAndRejectsUnknownOnes(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)

	coin := &models.Coin{Name: "2 Euros", Category: models.CategoryModern, UserID: 1, Circulation: " current_commemorative "}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("CreateCoin: %v", err)
	}
	var found models.Coin
	db.First(&found, coin.ID)
	if found.Circulation != "current_commemorative" {
		t.Fatalf("circulation = %q", found.Circulation)
	}

	err := svc.CreateCoin(&models.Coin{Name: "C", Category: models.CategoryModern, UserID: 1, Circulation: "legal"})
	if !errors.Is(err, ErrCoinInvalidCirculation) {
		t.Fatalf("circulation: %v", err)
	}
}

func TestUpdateCoin_RecomputesYearWhenTheDateChanges(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	coin := &models.Coin{Name: "Denarius", Category: models.CategoryRoman, Material: models.MaterialSilver, UserID: 1, DateRange: "1968"}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("CreateCoin: %v", err)
	}

	if err := svc.UpdateCoinWithFields(coin, &models.Coin{DateRange: "27 BC - AD 14"}, []string{"DateRange"}, 1, "manual", false); err != nil {
		t.Fatalf("UpdateCoinWithFields: %v", err)
	}
	var found models.Coin
	db.First(&found, coin.ID)
	if found.Year == nil || *found.Year != -27 {
		t.Fatalf("year = %v", found.Year)
	}

	if err := svc.UpdateCoinWithFields(coin, &models.Coin{Presentation: "box"}, []string{"Presentation"}, 1, "manual", false); !errors.Is(err, ErrCoinInvalidPresentation) {
		t.Fatalf("invalid presentation update: %v", err)
	}
}

func TestCoinList_FiltersByCountryPresentationAndRarity(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	for _, coin := range []*models.Coin{
		{Name: "A", Category: models.CategoryModern, UserID: 1, Country: "Portugal", Presentation: "coincard", Rarity: "common"},
		{Name: "B", Category: models.CategoryModern, UserID: 1, Country: "Portugal", Presentation: "loose", Rarity: "rare"},
		{Name: "C", Category: models.CategoryModern, UserID: 1, Country: "Spain", Presentation: "coincard", Rarity: "rare"},
	} {
		if err := svc.CreateCoin(coin); err != nil {
			t.Fatalf("CreateCoin: %v", err)
		}
	}
	repo := repository.NewCoinRepository(db)
	names := func(filters repository.CoinListFilters) []string {
		filters.Page, filters.Limit, filters.SortField, filters.SortOrder = 1, 50, "name", "asc"
		coins, _, err := repo.List(1, filters)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		out := []string{}
		for _, coin := range coins {
			out = append(out, coin.Name)
		}
		return out
	}

	if got := names(repository.CoinListFilters{Presentation: "coincard"}); len(got) != 2 || got[0] != "A" || got[1] != "C" {
		t.Fatalf("coincard: %v", got)
	}
	if got := names(repository.CoinListFilters{Country: "Portugal", Rarity: "rare"}); len(got) != 1 || got[0] != "B" {
		t.Fatalf("Portugal + rare: %v", got)
	}
	counts, err := repo.CountryCounts(1)
	if err != nil || len(counts) != 2 || counts[0] != (repository.CountryCount{Country: "Portugal", Count: 2}) || counts[1].Country != "Spain" {
		t.Fatalf("country counts = %+v (%v)", counts, err)
	}
}

func TestCoinCountryService_FillsCountriesFromNumista(t *testing.T) {
	db := setupTestDB(t)
	coinSvc := newTestCoinService(db)
	linked := &models.Coin{Name: "2.50 Escudos", Category: models.CategoryModern, UserID: 1}
	known := &models.Coin{Name: "Known", Category: models.CategoryModern, UserID: 1, Country: "Portugal"}
	unlinked := &models.Coin{Name: "Unlinked", Category: models.CategoryModern, UserID: 1}
	other := &models.Coin{Name: "Other user", Category: models.CategoryModern, UserID: 2}
	for _, coin := range []*models.Coin{linked, known, unlinked, other} {
		if err := coinSvc.CreateCoin(coin); err != nil {
			t.Fatalf("CreateCoin: %v", err)
		}
	}
	for _, coinID := range []uint{linked.ID, known.ID, other.ID} {
		if err := db.Create(&models.CoinReference{CoinID: coinID, Catalog: "Numista", Number: "2933"}).Error; err != nil {
			t.Fatalf("reference: %v", err)
		}
	}
	catalog, _, client, _ := newCatalogFixture()
	svc := NewCoinCountryService(repository.NewCoinRepository(db), catalog)

	got := svc.FillFromNumista(context.Background(), 1, 0, "pt")

	if got.Status != models.NumistaDataStatusOK || got.Updated != 1 || got.Remaining != 0 || got.Failed != 0 {
		t.Fatalf("result = %+v", got)
	}
	country := func(id uint) string {
		var coin models.Coin
		db.First(&coin, id)
		return coin.Country
	}
	if country(linked.ID) != "Angola" || country(known.ID) != "Portugal" || country(unlinked.ID) != "" || country(other.ID) != "" {
		t.Fatalf("countries: linked %q known %q unlinked %q other %q", country(linked.ID), country(known.ID), country(unlinked.ID), country(other.ID))
	}
	if client.lastLanguage != "pt" {
		t.Fatalf("lang = %q", client.lastLanguage)
	}

	// Numista without a key stops the fill and says why.
	if err := db.Model(&models.Coin{}).Where("id = ?", linked.ID).Update("country", "").Error; err != nil {
		t.Fatalf("reset: %v", err)
	}
	unconfiguredCatalog, _, unconfigured, _ := newCatalogFixture()
	unconfigured.detailErr = &NumistaError{Kind: NumistaErrorUnconfigured}
	stopped := NewCoinCountryService(repository.NewCoinRepository(db), unconfiguredCatalog).FillFromNumista(context.Background(), 1, linked.ID, "pt")
	if stopped.Status != models.NumistaDataStatusUnconfigured || stopped.Updated != 0 || stopped.Remaining != 1 {
		t.Fatalf("unconfigured: %+v", stopped)
	}
}

func TestCollectionSummary_CountsCoinsTypesAndCountries(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	coins := []*models.Coin{
		{Name: "50 Escudos 1968 – Portugal", Category: models.CategoryModern, UserID: 1, Country: "Portugal", Presentation: "coincard", CurrentValue: ptrFloat(23.81)},
		{Name: "50 Escudos 1969 – Portugal", Category: models.CategoryModern, UserID: 1, Country: "Portugal", Presentation: "coincard", CurrentValue: ptrFloat(26.19)},
		{Name: "Denarius", Category: models.CategoryRoman, UserID: 1},
		{Name: " denarius ", Category: models.CategoryRoman, UserID: 1},
		{Name: "2.50 Escudos", Category: models.CategoryModern, UserID: 1, Country: "Angola"},
		{Name: "Wanted", Category: models.CategoryModern, UserID: 1, Country: "Spain", IsWishlist: true, Presentation: "coincard", CurrentValue: ptrFloat(1000)},
		{Name: "Sold", Category: models.CategoryModern, UserID: 1, Country: "France", IsSold: true},
		{Name: "Someone else's", Category: models.CategoryModern, UserID: 2, Country: "Italy"},
	}
	for _, coin := range coins {
		if err := svc.CreateCoin(coin); err != nil {
			t.Fatalf("CreateCoin: %v", err)
		}
	}
	// Both 50 Escudos coins are the same Numista type.
	for _, coin := range coins[:2] {
		if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: "4930"}).Error; err != nil {
			t.Fatalf("reference: %v", err)
		}
	}

	got, err := repository.NewCoinRepository(db).CollectionSummary(1)

	want := repository.CollectionSummary{Coins: 5, Types: 3, Countries: 2, WithoutCountry: 2, Coincards: 2, TotalValue: 50, UnvaluedCoins: 3}
	if err != nil || got != want {
		t.Fatalf("summary = %+v (%v), want %+v", got, err, want)
	}
	empty, err := repository.NewCoinRepository(db).CollectionSummary(99)
	if err != nil || empty != (repository.CollectionSummary{}) {
		t.Fatalf("empty collection = %+v (%v)", empty, err)
	}
}

func TestCoinYear_TypedYearWinsAndEmptyYearFollowsTheDate(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	yearOf := func(id uint) *int {
		var coin models.Coin
		db.First(&coin, id)
		return coin.Year
	}
	typed := -44
	coin := &models.Coin{Name: "Denarius", Category: models.CategoryRoman, UserID: 1, DateRange: "c. 46-44 BC", Year: &typed}
	if err := svc.CreateCoin(coin); err != nil {
		t.Fatalf("CreateCoin: %v", err)
	}
	if got := yearOf(coin.ID); got == nil || *got != -44 {
		t.Fatalf("typed year not kept: %v", got)
	}

	zero := 0
	derived := &models.Coin{Name: "Escudo", Category: models.CategoryModern, UserID: 1, DateRange: "1968", Year: &zero}
	if err := svc.CreateCoin(derived); err != nil {
		t.Fatalf("CreateCoin: %v", err)
	}
	if got := yearOf(derived.ID); got == nil || *got != 1968 {
		t.Fatalf("year 0 must fall back to the date: %v", got)
	}

	// Clearing the year (sent as null) reads it from the existing date again.
	if err := svc.UpdateCoinWithFields(coin, &models.Coin{}, []string{"Year"}, 1, "manual", false); err != nil {
		t.Fatalf("clear year: %v", err)
	}
	if got := yearOf(coin.ID); got == nil || *got != -46 {
		t.Fatalf("cleared year should follow the date: %v", got)
	}

	newYear := 1969
	if err := svc.UpdateCoinWithFields(derived, &models.Coin{DateRange: "1968", Year: &newYear}, []string{"DateRange", "Year"}, 1, "manual", false); err != nil {
		t.Fatalf("set year: %v", err)
	}
	if got := yearOf(derived.ID); got == nil || *got != 1969 {
		t.Fatalf("typed year on update not kept: %v", got)
	}
}

func TestCollectionHighlights_PicksRarestOldestAndMostValuable(t *testing.T) {
	db := setupTestDB(t)
	svc := newTestCoinService(db)
	year := func(v int) *int { return &v }
	coins := map[string]*models.Coin{
		"rare cheap":   {Name: "Rare cheap", Category: models.CategoryModern, UserID: 1, Rarity: "very_rare", CurrentValue: ptrFloat(10), Year: year(1900)},
		"rare dear":    {Name: "Rare dear", Category: models.CategoryModern, UserID: 1, Rarity: "very_rare", CurrentValue: ptrFloat(90), Year: year(1950)},
		"common":       {Name: "Common", Category: models.CategoryModern, UserID: 1, Rarity: "common", CurrentValue: ptrFloat(5), Year: year(1968)},
		"denarius":     {Name: "Denarius", Category: models.CategoryRoman, UserID: 1, Year: year(-44), CurrentValue: ptrFloat(300)},
		"dateless":     {Name: "Dateless gold", Category: models.CategoryModern, UserID: 1, CurrentValue: ptrFloat(1200)},
		"wishlist":     {Name: "Wanted", Category: models.CategoryModern, UserID: 1, Rarity: "extremely_rare", Year: year(-500), CurrentValue: ptrFloat(9000), IsWishlist: true},
		"sold":         {Name: "Sold", Category: models.CategoryModern, UserID: 1, Rarity: "extremely_rare", CurrentValue: ptrFloat(5000), IsSold: true},
		"another user": {Name: "Not mine", Category: models.CategoryModern, UserID: 2, Rarity: "extremely_rare", Year: year(-600), CurrentValue: ptrFloat(8000)},
	}
	for _, coin := range coins {
		if err := svc.CreateCoin(coin); err != nil {
			t.Fatalf("CreateCoin: %v", err)
		}
	}
	if err := db.Create(&models.CoinImage{CoinID: coins["denarius"].ID, FilePath: "coin-4/obverse.jpg", ImageType: models.ImageTypeObverse}).Error; err != nil {
		t.Fatalf("image: %v", err)
	}
	if err := db.Create(&models.CoinImage{CoinID: coins["denarius"].ID, FilePath: "coin-4/reverse.jpg", ImageType: "reverse"}).Error; err != nil {
		t.Fatalf("image: %v", err)
	}

	got, err := repository.NewCoinRepository(db).CollectionHighlights(1)
	if err != nil {
		t.Fatalf("CollectionHighlights: %v", err)
	}
	if got.Rarest == nil || got.Rarest.Name != "Rare dear" {
		t.Fatalf("rarest = %+v (ties go to the more valuable coin)", got.Rarest)
	}
	if got.Oldest == nil || got.Oldest.Name != "Denarius" || got.Oldest.ImagePath != "coin-4/obverse.jpg" {
		t.Fatalf("oldest = %+v", got.Oldest)
	}
	if got.MostValuable == nil || got.MostValuable.Name != "Dateless gold" || *got.MostValuable.CurrentValue != 1200 {
		t.Fatalf("most valuable = %+v", got.MostValuable)
	}

	empty, err := repository.NewCoinRepository(db).CollectionHighlights(99)
	if err != nil || empty.Rarest != nil || empty.Oldest != nil || empty.MostValuable != nil {
		t.Fatalf("empty collection = %+v (%v)", empty, err)
	}
}
