package models

import "time"

// CatalogueOverride is a user's choice about a coin of the catalogue: hidden
// (SameAs empty), or the same coin as another of the catalogue, which it is
// then listed as. CoinKey and SameAs are catalogue coin ids.
type CatalogueOverride struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_catalogue_override_user_coin" json:"-"`
	CoinKey   string    `gorm:"size:600;not null;uniqueIndex:idx_catalogue_override_user_coin" json:"id"`
	SameAs    string    `gorm:"size:600;not null;default:''" json:"sameAs,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
