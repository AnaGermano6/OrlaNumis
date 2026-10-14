package models

import "time"

type QuickAccessTargetType string

const (
	QuickAccessTargetCoin    QuickAccessTargetType = "coin"
	QuickAccessTargetCoinSet QuickAccessTargetType = "coin_set"
)

func (t QuickAccessTargetType) Valid() bool {
	switch t {
	case QuickAccessTargetCoin, QuickAccessTargetCoinSet:
		return true
	default:
		return false
	}
}

type QuickAccessPin struct {
	ID         uint                  `gorm:"primaryKey" json:"-"`
	UserID     uint                  `gorm:"not null;uniqueIndex:idx_quick_access_owner_target;index:idx_quick_access_owner_time,priority:1" json:"-"`
	TargetType QuickAccessTargetType `gorm:"type:varchar(24);not null;check:target_type IN ('coin','coin_set');uniqueIndex:idx_quick_access_owner_target;index:idx_quick_access_target,priority:1" json:"type"`
	TargetID   uint                  `gorm:"not null;uniqueIndex:idx_quick_access_owner_target;index:idx_quick_access_target,priority:2" json:"id"`
	PinnedAt   time.Time             `gorm:"not null;index:idx_quick_access_owner_time,priority:2,sort:desc" json:"pinnedAt"`
	CreatedAt  time.Time             `json:"-"`
}
