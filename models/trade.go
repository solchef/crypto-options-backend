package models

import "time"

type Trade struct {
	ID         uint    `gorm:"primaryKey"`
	UserID     uint    `gorm:"not null"`
	WalletID   uint    `gorm:"not null"`
	Asset      string  `gorm:"not null"` // e.g. BTCUSDT
	Amount     float64 `gorm:"not null"`
	Direction  string  `gorm:"not null"` // "UP" or "DOWN"
	EntryPrice float64 `gorm:"not null"`
	ExitPrice  float64
	Duration   int    `gorm:"not null"`       // in seconds
	Status     string `gorm:"default:'OPEN'"` // OPEN / WON / LOST
	CreatedAt  time.Time
	ExpiredAt  time.Time
}
