package services

import (
	"slices"

	"github.com/briandenicola/ancient-coins-api/models"
)

// catalogueID is what identifies a coin of the catalogue to its users.
func catalogueID(coin CoinRelease) string {
	if coin.ID != "" {
		return coin.ID
	}
	if len(coin.Keys) > 0 {
		return coin.Keys[0]
	}
	return ""
}

// FindCatalogueCoin finds the coin of the catalogue one of whose keys is id.
func FindCatalogueCoin(coins []CoinRelease, id string) (CoinRelease, bool) {
	for _, coin := range coins {
		if slices.Contains(coin.Keys, id) {
			return coin, true
		}
	}
	return CoinRelease{}, false
}

// ApplyCatalogueOverrides applies a user's choices to coins of the catalogue.
// A coin said to be the same as another is listed once, as that one, which
// takes its keys (a coin of the collection identified as either is then
// it); a hidden coin is left out. Both are returned apart, a joined one with
// SameAs the id of the coin it is listed as. A choice about a coin not among
// coins (another country's, or one gone from the catalogue) changes nothing.
func ApplyCatalogueOverrides(coins []CoinRelease, overrides []models.CatalogueOverride) (shown, apart []CoinRelease) {
	apart = []CoinRelease{}
	if len(overrides) == 0 {
		return coins, apart
	}
	coins = append([]CoinRelease{}, coins...)
	at := map[string]int{}
	for i := range coins {
		coins[i].ID = catalogueID(coins[i])
		for _, key := range coins[i].Keys {
			if _, ok := at[key]; !ok {
				at[key] = i
			}
		}
	}
	// The coins joined together, each group listed as its last target.
	parent := make([]int, len(coins))
	for i := range parent {
		parent[i] = i
	}
	root := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	hidden := map[int]bool{}
	for _, override := range overrides {
		i, ok := at[override.CoinKey]
		if !ok {
			continue
		}
		if override.SameAs == "" {
			hidden[i] = true
			continue
		}
		j, ok := at[override.SameAs]
		if !ok {
			continue
		}
		if ri, rj := root(i), root(j); ri != rj {
			parent[ri] = rj
		}
	}
	for i := range coins {
		if r := root(i); r != i {
			// A new slice: the catalogue's coins share theirs.
			coins[r].Keys = slices.Concat(coins[r].Keys, coins[i].Keys)
		}
	}
	shown = make([]CoinRelease, 0, len(coins))
	for i, coin := range coins {
		switch r := root(i); {
		case r != i:
			coin.SameAs = coins[r].ID
			apart = append(apart, coin)
		case hidden[i]:
			apart = append(apart, coin)
		default:
			shown = append(shown, coin)
		}
	}
	return shown, apart
}
