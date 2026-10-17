package repository

import "gorm.io/gorm"

// OwnedBy scopes a query to records belonging to the given user.
func OwnedBy(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	}
}

// OwnedByID scopes a query to a specific record owned by the given user.
func OwnedByID(id uint, userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ? AND user_id = ?", id, userID)
	}
}

// ActiveCollection scopes to the user's coins that are not on the wishlist.
func ActiveCollection(userID uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ? AND is_wishlist = ?", userID, false)
	}
}

// OfKind keeps the coins, or the banknotes (see models.CoinKind…); every
// item when kind is "".
func OfKind(kind string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if kind == "" {
			return db
		}
		return db.Where("kind = ?", kind)
	}
}
