package services

import (
	"fmt"
	"time"

	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/models"
	"gorm.io/gorm"
)

// SettleTrade waits until expiry and then resolves the trade
func SettleTrade(tradeID uint) {
	var trade models.Trade
	if err := config.DB.First(&trade, tradeID).Error; err != nil {
		return // trade not found
	}

	// Wait until expiry
	time.Sleep(time.Until(trade.ExpiredAt))

	// Fetch exit price
	exitPrice, err := GetCurrentPriceByAPI(trade.Asset)
	if err != nil {
		return
	}

	var result string
	if (trade.Direction == "UP" && exitPrice > trade.EntryPrice) ||
		(trade.Direction == "DOWN" && exitPrice < trade.EntryPrice) {
		result = "WON"
		payout := trade.Amount * 1.8 // 80% return

		config.DB.Model(&models.Wallet{}).
			Where("id = ?", trade.WalletID).
			Update("balance", gorm.Expr("balance + ?", payout))

		config.DB.Create(&models.WalletTransaction{
			WalletID:  trade.WalletID,
			Amount:    payout,
			Type:      "trade_win",
			Reference: fmt.Sprintf("Trade #%d", trade.ID),
			CreatedAt: time.Now(),
		})
	} else {
		result = "LOST"
	}

	// Update trade status & exit price
	config.DB.Model(&trade).Updates(map[string]interface{}{
		"status":     result,
		"exit_price": exitPrice,
	})

	// TODO: Push WebSocket notification to frontend
	fmt.Printf("Trade %d settled: %s\n", trade.ID, result)
	msg := fmt.Sprintf(`{"trade_id": %d, "status": "%s", "exit_price": %.2f}`, trade.ID, result, exitPrice)
	config.WSHub.SendToUser(trade.UserID, msg)
}
