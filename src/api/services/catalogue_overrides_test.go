package services

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/briandenicola/ancient-coins-api/models"
	"github.com/briandenicola/ancient-coins-api/repository"
)

func overrideTitles(coins []CoinRelease) string {
	titles := []string{}
	for _, coin := range coins {
		title := coin.Title
		if coin.SameAs != "" {
			title += "=" + coin.SameAs
		}
		titles = append(titles, title)
	}
	return strings.Join(titles, ", ")
}

// A hidden coin is left out, a joined one is listed as the other, which
// takes its keys; joins follow one another, and choices about other coins
// change nothing.
func TestApplyCatalogueOverrides(t *testing.T) {
	keys := make([]string, 1, 4)
	keys[0] = "a"
	catalogue := []CoinRelease{
		{Title: "A", Keys: keys},
		{Title: "B", Keys: []string{"b", "b2"}},
		{Title: "C", Keys: []string{"c"}},
		{Title: "D", Keys: []string{"d"}},
		{Title: "E", Keys: []string{"e"}},
	}
	shown, apart := ApplyCatalogueOverrides(catalogue, []models.CatalogueOverride{
		{CoinKey: "b2", SameAs: "a"},
		{CoinKey: "c", SameAs: "b"},
		{CoinKey: "d"},
		{CoinKey: "gone"},
		{CoinKey: "e", SameAs: "gone"},
	})
	if got := overrideTitles(shown); got != "A, E" {
		t.Fatalf("shown = %s", got)
	}
	if got := overrideTitles(apart); got != "B=a, C=a, D" {
		t.Fatalf("apart = %s", got)
	}
	if !slices.Equal(shown[0].Keys, []string{"a", "b", "b2", "c"}) || shown[0].ID != "a" {
		t.Fatalf("A = %+v", shown[0])
	}
	// The catalogue's own coins are as they were.
	if len(catalogue[0].Keys) != 1 || catalogue[1].SameAs != "" || keys[:2][1] != "" {
		t.Fatalf("catalogue changed: %+v", catalogue)
	}

	shown, apart = ApplyCatalogueOverrides(catalogue, nil)
	if len(shown) != 5 || len(apart) != 0 {
		t.Fatalf("without choices: %d shown, %d apart", len(shown), len(apart))
	}
	if coin, ok := FindCatalogueCoin(catalogue, "b2"); !ok || coin.Title != "B" {
		t.Fatalf("FindCatalogueCoin(b2) = %+v %v", coin, ok)
	}
}

// A checklist of the catalogue leaves out the coins its user hid, and lists
// a joined coin as the other, filled by a coin of the collection that is
// either.
func TestChecklistSlots_CatalogueLeavesOutHiddenAndJoinedCoins(t *testing.T) {
	test := newChecklistTest(t)
	if err := test.db.AutoMigrate(&models.CatalogueOverride{}); err != nil {
		t.Fatal(err)
	}
	overrides := repository.NewCatalogueOverrideRepository(test.db)
	test.checklist.WithOverrides(overrides)
	gago := test.gagoCoutinho(t)
	ctx := context.Background()
	definition := ChecklistDefinition{Source: ChecklistSourceCatalogue, Country: "PT"}
	// Erasmus, Gago Coutinho, D. Catarina.
	catalogue := test.checklist.CatalogueSlots(ctx, definition)
	if len(catalogue) != 3 || catalogue[1].Key != gagoCoutinhoImage {
		t.Fatalf("catalogue slots = %+v", catalogue)
	}
	now := time.Now()
	if err := overrides.Set(1, catalogue[2].Key, "", now); err != nil {
		t.Fatal(err)
	}
	if err := overrides.Set(1, gagoCoutinhoImage, catalogue[0].Key, now); err != nil {
		t.Fatal(err)
	}

	slots, err := test.checklist.Slots(ctx, 1, definition)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 1 || slots[0].Key != catalogue[0].Key || !slots[0].Owned || slots[0].CoinID != gago.ID {
		t.Fatalf("slots = %+v", slots)
	}
	// Another user's checklist lists them all; the catalogue's slots too.
	if slots, _ := test.checklist.Slots(ctx, 2, definition); len(slots) != 3 {
		t.Fatalf("another user's slots = %+v", slots)
	}
	if catalogue := test.checklist.CatalogueSlots(ctx, definition); len(catalogue) != 3 {
		t.Fatalf("catalogue slots = %+v", catalogue)
	}

	// Shown again.
	if err := overrides.Delete(1, []string{catalogue[2].Key, gagoCoutinhoImage}); err != nil {
		t.Fatal(err)
	}
	if slots, _ := test.checklist.Slots(ctx, 1, definition); len(slots) != 3 || !slots[1].Owned || slots[0].Owned {
		t.Fatalf("slots after showing again = %+v", slots)
	}
}
