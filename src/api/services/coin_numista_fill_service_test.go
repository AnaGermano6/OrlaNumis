package services

import (
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func TestCoinNumistaFillService_FillsEmptyFieldsFromCachedTypes(t *testing.T) {
	db := setupTestDB(t)
	catalog, _, client, store := newCatalogFixture()
	cached := func(key, payload string) {
		store.entries[key] = models.NumistaCacheEntry{Key: key, Payload: payload, FetchedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	}
	// Cuproníquel, 3.5 g, 20 mm (in Portuguese).
	cached("type:2933:pt", testCatalogTypeDetail)
	// Copper plated steel issued in a coin card, cached in English only.
	cached("type:700:en", `{"id": 700, "title": "1 Cent", "composition": {"text": "Copper plated steel"},
		"weight": 2.3, "size": 16.25, "comments": "<p>Issued in a coincard.</p>"}`)
	// A metal the app has no material for.
	cached("type:800:pt", `{"id": 800, "title": "10 Pfennig", "composition": {"text": "Ferro"}, "weight": 3, "size": 21}`)

	weight := 4.0
	diameter := 21.0
	coins := []*models.Coin{
		{Name: "2.50 Escudos", UserID: 1},
		{Name: "Silver by hand", UserID: 1, Material: models.MaterialSilver, WeightGrams: &weight},
		{Name: "Wanted cent", UserID: 1, IsWishlist: true},
		{Name: "Not cached", UserID: 1},
		{Name: "Banknote", UserID: 1, Kind: models.CoinKindBanknote},
		{Name: "Other user", UserID: 2},
		{Name: "Iron", UserID: 1, WeightGrams: &weight, DiameterMm: &diameter},
		{Name: "Unlinked", UserID: 1},
		{Name: "Loose by hand", UserID: 1, Presentation: "loose", WeightGrams: &weight, DiameterMm: &diameter},
	}
	types := []string{"2933", "2933", "700", "900", "2933", "2933", "800", "", "700"}
	for i, coin := range coins {
		if err := db.Create(coin).Error; err != nil {
			t.Fatalf("create %s: %v", coin.Name, err)
		}
		if types[i] != "" {
			if err := db.Create(&models.CoinReference{CoinID: coin.ID, Catalog: "Numista", Number: types[i]}).Error; err != nil {
				t.Fatalf("reference: %v", err)
			}
		}
	}
	svc := NewCoinNumistaFillService(repository.NewCoinRepository(db), catalog)

	preview, err := svc.Preview(1)
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if preview.Applied || preview.Coins != 5 || preview.Materials != 3 || preview.Weights != 2 || preview.Diameters != 3 || preview.Presentations != 1 || preview.MonetarySystems != 3 {
		t.Fatalf("preview counts = %+v", preview)
	}
	if preview.ByMaterial["Copper-nickel"] != 1 || preview.ByMaterial["Plated steel"] != 2 || preview.Unmatched != 1 || preview.NotCached != 1 {
		t.Fatalf("preview materials = %+v unmatched %d notCached %d", preview.ByMaterial, preview.Unmatched, preview.NotCached)
	}
	byName := map[string]models.CoinNumistaFillChange{}
	for _, change := range preview.Changes {
		byName[change.Name] = change
	}
	if got := byName["2.50 Escudos"]; got.Material != models.MaterialCopperNickel || got.Composition != "Cuproníquel" || *got.WeightGrams != 3.5 || *got.DiameterMm != 20 || got.Presentation != "" || got.MonetarySystem != "Escudo (1958-1977)" {
		t.Fatalf("escudos change = %+v", got)
	}
	if got := byName["Silver by hand"]; got.Material != "" || got.WeightGrams != nil || got.DiameterMm == nil || *got.DiameterMm != 20 {
		t.Fatalf("a material and weight set by hand stay: %+v", got)
	}
	// Its material is not what Numista says: a correction, kept apart.
	if len(preview.Corrections) != 1 || preview.Corrections[0].Name != "Silver by hand" ||
		preview.Corrections[0].PreviousMaterial != models.MaterialSilver || preview.Corrections[0].Material != models.MaterialCopperNickel {
		t.Fatalf("corrections = %+v", preview.Corrections)
	}
	if got := byName["Wanted cent"]; !got.IsWishlist || got.Material != models.MaterialPlatedSteel || got.Presentation != "coincard" {
		t.Fatalf("wishlist cent = %+v", got)
	}
	if got := byName["Loose by hand"]; got.Material != models.MaterialPlatedSteel || got.Presentation != "" || got.WeightGrams != nil {
		t.Fatalf("a presentation set by hand stays: %+v", got)
	}
	// A banknote gets its monetary system, and nothing a coin has.
	if got := byName["Banknote"]; got.MonetarySystem != "Escudo (1958-1977)" || got.Material != "" || got.WeightGrams != nil || got.DiameterMm != nil {
		t.Fatalf("banknote change = %+v", got)
	}
	if _, ok := byName["Iron"]; ok {
		t.Fatalf("nothing to fill on the iron coin: %+v", byName["Iron"])
	}
	var untouched models.Coin
	db.First(&untouched, coins[0].ID)
	if untouched.Material != models.MaterialOther || untouched.WeightGrams != nil {
		t.Fatalf("the preview must change nothing: %+v", untouched)
	}

	applied, err := svc.Apply(1, false)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !applied.Applied || applied.Coins != 5 || applied.Corrected != 0 {
		t.Fatalf("applied = %+v", applied)
	}
	load := func(id uint) models.Coin {
		var coin models.Coin
		if err := db.First(&coin, id).Error; err != nil {
			t.Fatalf("load %d: %v", id, err)
		}
		return coin
	}
	if got := load(coins[0].ID); got.Material != models.MaterialCopperNickel || *got.WeightGrams != 3.5 || *got.DiameterMm != 20 || got.MonetarySystem != "Escudo (1958-1977)" {
		t.Fatalf("escudos after = %+v", got)
	}
	if got := load(coins[1].ID); got.Material != models.MaterialSilver || *got.WeightGrams != 4 || *got.DiameterMm != 20 {
		t.Fatalf("silver after = %+v", got)
	}
	if got := load(coins[2].ID); got.Material != models.MaterialPlatedSteel || got.Presentation != "coincard" || *got.DiameterMm != 16.25 {
		t.Fatalf("cent after = %+v", got)
	}
	for _, i := range []int{3, 4, 5, 6, 7} {
		if got := load(coins[i].ID); got.Material != models.MaterialOther && got.Material != "" {
			t.Fatalf("%s must stay as it was: %+v", coins[i].Name, got)
		}
	}
	if got := load(coins[4].ID); got.MonetarySystem != "Escudo (1958-1977)" || got.WeightGrams != nil {
		t.Fatalf("banknote after = %+v", got)
	}
	if got := load(coins[5].ID); got.WeightGrams != nil || got.MonetarySystem != "" {
		t.Fatalf("another user's coin was changed: %+v", got)
	}

	again, err := svc.Preview(1)
	if err != nil || again.Coins != 0 || len(again.Changes) != 0 || len(again.Corrections) != 1 {
		t.Fatalf("after applying, nothing is left to fill but the correction: %+v (%v)", again, err)
	}
	corrected, err := svc.Apply(1, true)
	if err != nil || corrected.Corrected != 1 || corrected.Coins != 0 {
		t.Fatalf("with corrections: %+v (%v)", corrected, err)
	}
	if got := load(coins[1].ID); got.Material != models.MaterialCopperNickel || *got.WeightGrams != 4 {
		t.Fatalf("silver corrected = %+v", got)
	}
	if client.detailCalls != 0 || client.searchCalls != 0 {
		t.Fatalf("the fill must not ask Numista: %d detail / %d search calls", client.detailCalls, client.searchCalls)
	}
}
