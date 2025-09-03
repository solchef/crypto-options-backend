package models

import (
	"time"
)

// Wallet represents a user's wallet
type WalletTransaction struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	WalletID  uint      `gorm:"not null" json:"wallet_id"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"` // deposit, withdraw, trade, bonus, etc.
	Reference string    `json:"reference"`
	CreatedAt time.Time `json:"created_at"`
	Wallet    Wallet    `gorm:"foreignKey:WalletID"`
}
