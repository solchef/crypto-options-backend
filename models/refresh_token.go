package models

import "time"

type RefreshToken struct {
	ID        uint   `gorm:"primaryKey" json:"-"`
	UserID    uint   `gorm:"index" json:"-"`
	JTI       string `gorm:"uniqueIndex;size:64"`
	ExpiresAt time.Time
	Revoked   bool `gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
