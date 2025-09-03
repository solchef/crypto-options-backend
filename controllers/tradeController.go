package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/models"
	"github.com/solchef/crypto-options-backend/services"

	"github.com/gin-gonic/gin"
)

// PlaceTrade godoc
// @Summary Place a trade
// @Description Place a new trade with immediate debit from wallet
// @Tags trade
// @Accept json
// @Produce json
// @Param trade body object{wallet_id=uint,asset=string,amount=number,direction=string,duration=int} true "Trade request"
// @Success 200 {object} models.Trade
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /trades/place [post]
func PlaceTrade(c *gin.Context) {
	var req struct {
		WalletID  uint    `json:"wallet_id"`
		Asset     string  `json:"asset"`
		Amount    float64 `json:"amount"`
		Direction string  `json:"direction"` // "UP" or "DOWN"
		Duration  int     `json:"duration"`  // seconds
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userID := c.GetUint("userID")

	// Get wallet
	var wallet models.Wallet
	if err := config.DB.First(&wallet, "id = ? AND user_id = ?", req.WalletID, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	// Check balance
	if wallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	// Get current price from Binance
	price, err := services.GetCurrentPriceByAPI(req.Asset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch price"})
		return
	}

	// Immediate debit
	wallet.Balance -= req.Amount
	config.DB.Save(&wallet)

	// Save trade
	trade := models.Trade{
		UserID:     userID,
		WalletID:   wallet.ID,
		Asset:      req.Asset,
		Amount:     req.Amount,
		Direction:  req.Direction,
		EntryPrice: price,
		Duration:   req.Duration,
		CreatedAt:  time.Now(),
		ExpiredAt:  time.Now().Add(time.Duration(req.Duration) * time.Second),
		Status:     "OPEN",
	}
	config.DB.Create(&trade)

	config.DB.Create(&models.WalletTransaction{
		WalletID:  trade.WalletID,
		Amount:    -req.Amount,
		Type:      "trade",
		Reference: fmt.Sprintf("Trade #%d", trade.ID),
		CreatedAt: time.Now(),
	})

	// 🔹 Launch goroutine to settle after expiry
	go services.SettleTrade(trade.ID)

	c.JSON(http.StatusOK, trade)
}

// GetOpenTrades godoc
// @Summary Get open trades
// @Description Retrieve all open trades for the logged-in user
// @Tags trade
// @Produce json
// @Success 200 {array} models.Trade
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /trades/open [get]
func GetOpenTrades(c *gin.Context) {
	var trades []models.Trade
	config.DB.Where("user_id = ? AND status = ?", c.GetUint("userID"), "OPEN").Find(&trades)
	c.JSON(http.StatusOK, trades)
}

// GetTradeHistory godoc
// @Summary Get trade history
// @Description Retrieve all trades (open and closed) for the logged-in user
// @Tags trade
// @Produce json
// @Success 200 {array} models.Trade
// @Failure 401 {object} map[string]string
// @Security ApiKeyAuth
// @Router /trades/history [get]
func GetTradeHistory(c *gin.Context) {
	var trades []models.Trade
	config.DB.Where("user_id = ?", c.GetUint("userID")).Find(&trades)
	c.JSON(http.StatusOK, trades)
}

// CloseTrade godoc
// @Summary Close a trade
// @Description Manually close a trade
// @Tags trade
// @Accept json
// @Produce json
// @Param trade body object{trade_id=uint} true "Trade ID to close"
// @Success 200 {object} models.Trade
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /trades/close [post]
func CloseTrade(c *gin.Context) {
	var input struct {
		TradeID uint `json:"trade_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var trade models.Trade
	if err := config.DB.First(&trade, input.TradeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trade not found"})
		return
	}

	trade.Status = "closed"
	config.DB.Save(&trade)
	c.JSON(http.StatusOK, trade)
}
