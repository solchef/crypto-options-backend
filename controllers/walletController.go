package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/models"
	"gorm.io/gorm"
)

// GetWallets godoc
// @Summary Get user wallets
// @Description Retrieve all wallets for the logged-in user
// @Tags wallet
// @Produce json
// @Success 200 {array} models.Wallet
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /wallets [get]
func GetWallets(c *gin.Context) {
	userID := c.GetUint("userID")

	fmt.Println("userID from context:", userID)

	var wallets []models.Wallet
	if err := config.DB.Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallets not found"})
		return
	}

	c.JSON(http.StatusOK, wallets)
}

// Deposit godoc
// @Summary Deposit funds
// @Description Deposit a certain amount to a wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Param deposit body object{currency=string,amount=number} true "Deposit request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /wallets/deposit [post]
func Deposit(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ? AND currency = ?", userID, req.Currency).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		wallet.Balance += req.Amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		return tx.Create(&models.WalletTransaction{
			WalletID:  wallet.ID,
			Amount:    req.Amount,
			Type:      "deposit",
			Reference: "manual_deposit",
		}).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Deposit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful", "balance": wallet.Balance})
}

// Withdraw godoc
// @Summary Withdraw funds
// @Description Withdraw a certain amount from a wallet
// @Tags wallet
// @Accept json
// @Produce json
// @Param withdraw body object{amount=number} true "Withdraw request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /wallets/withdraw [post]
func Withdraw(c *gin.Context) {
	userID := c.GetUint("userID")

	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	if wallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		return
	}

	config.DB.Transaction(func(tx *gorm.DB) error {
		wallet.Balance -= req.Amount
		if err := tx.Save(&wallet).Error; err != nil {
			return err
		}

		tx.Create(&models.WalletTransaction{
			WalletID:  wallet.ID,
			Amount:    -req.Amount,
			Type:      "withdraw",
			Reference: "manual_withdrawal",
		})

		return nil
	})

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful", "balance": wallet.Balance})
}

// GetWalletTransactions godoc
// @Summary Get wallet transactions
// @Description Retrieve transactions for user's wallets
// @Tags wallet
// @Produce json
// @Param currency query string false "Filter by currency"
// @Success 200 {array} models.WalletTransaction
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /wallets/transactions [get]
func GetWalletTransactions(c *gin.Context) {
	userID := c.GetUint("userID")
	currency := c.Query("currency")

	var walletIDs []uint
	query := config.DB.Model(&models.Wallet{}).Where("user_id = ?", userID)
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}

	if err := query.Pluck("id", &walletIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wallets"})
		return
	}

	if len(walletIDs) == 0 {
		c.JSON(http.StatusOK, []models.WalletTransaction{})
		return
	}

	// Step 2: Get transactions for these wallets
	var transactions []models.WalletTransaction
	if err := config.DB.
		Where("wallet_id IN ?", walletIDs).
		Order("created_at DESC").
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}
