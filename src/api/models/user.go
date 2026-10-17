package models

import "time"

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Username           string     `gorm:"uniqueIndex;not null" json:"username"`
	Email              string     `gorm:"uniqueIndex" json:"email"`
	PasswordHash       string     `gorm:"not null" json:"-"`
	Role               UserRole   `gorm:"type:varchar(10);default:'user'" json:"role"`
	AvatarPath         string     `json:"avatarPath"`
	Bio                string     `gorm:"type:text" json:"bio"`
	ZipCode            string     `gorm:"type:varchar(10)" json:"zipCode"`
	PWASwipeNavEnabled bool       `gorm:"column:pwa_swipe_nav_enabled;not null;default:false" json:"pwaSwipeNavEnabled"`
	LockedUntil        *time.Time `gorm:"index" json:"lockedUntil,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}
