package services

import (
	"context"
	"strings"

	"github.com/briandenicola/ancient-coins-api/models"
)

// WishlistMatches returns the wishlist's coins that are the collection coin,
// so that one added to the collection can leave the wishlist: of its Numista
// type or its official photo (the ECB's image of a €2 commemorative, a
// mint's page of a collector coin, see OwnedKeys), or of its kind and year
// under the same name. Only what the app has stored is looked at: Numista is
// not asked.
func (s *CoinReleasesService) WishlistMatches(ctx context.Context, coin *models.Coin, wishlist []models.Coin, photos *CoinPhotoSearchService) []models.Coin {
	matches := []models.Coin{}
	if coin == nil || coin.IsWishlist {
		return matches
	}
	keysOf := func(candidate *models.Coin) map[string]uint {
		query := CoinPhotoQueryFor(candidate)
		query.CachedTypesOnly = true
		return s.OwnedKeys(ctx, []CoinPhotoQuery{query}, photos)
	}
	keys := keysOf(coin)
	for i := range wishlist {
		wanted := &wishlist[i]
		if wanted.ID == coin.ID || !wanted.IsWishlist || wanted.Kind != coin.Kind {
			continue
		}
		if sameCoinName(wanted, coin) || sharesAnyKey(keysOf(wanted), keys) {
			matches = append(matches, *wanted)
		}
	}
	return matches
}

// sameCoinName tells whether two coins of the same year have the same name,
// whatever its case, accents and spacing.
func sameCoinName(a, b *models.Coin) bool {
	if (a.Year == nil) != (b.Year == nil) || a.Year != nil && *a.Year != *b.Year {
		return false
	}
	name := func(coin *models.Coin) string { return strings.Join(strings.Fields(foldASCII(coin.Name)), " ") }
	return name(a) != "" && name(a) == name(b)
}

func sharesAnyKey(a, b map[string]uint) bool {
	for key := range a {
		if _, ok := b[key]; ok {
			return true
		}
	}
	return false
}
