package models

import (
	"time"
)

// Wallet represents a user's wallet
type Wallet struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"` // ✅ indexed, not unique
	Currency  string    `gorm:"not null" json:"currency"`
	Balance   float64   `gorm:"default:0" json:"balance"` // Main trading balance
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      User      `gorm:"foreignKey:UserID"`
}
