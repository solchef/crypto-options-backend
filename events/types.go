package events

type PriceTick struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"` // ms
}

type TickerStats struct {
	Symbol    string  `json:"symbol"`
	Change    float64 `json:"change"`     // 24h change
	ChangePct float64 `json:"change_pct"` // 24h change percent
	Timestamp int64   `json:"timestamp"`  // ms
}

// (Optional) future events
type TradePlaced struct {
	TradeID   string  `json:"tradeId"`
	UserID    uint    `json:"userId"`
	Symbol    string  `json:"symbol"`
	Direction string  `json:"direction"` // CALL/PUT
	Stake     float64 `json:"stake"`
	ExpiryMs  int64   `json:"expiryMs"`
}

type Settlement struct {
	TradeID   string  `json:"tradeId"`
	Symbol    string  `json:"symbol"`
	ExitPrice float64 `json:"exitPrice"`
	Status    string  `json:"status"` // WON/LOST
	Payout    float64 `json:"payout"`
	Timestamp int64   `json:"timestamp"`
}
