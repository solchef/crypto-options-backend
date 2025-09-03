package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solchef/crypto-options-backend/services"
)

// PricePoint is used by swagger to describe a single candle/point in the history response.
type PricePoint struct {
	Time  string  `json:"time" example:"2025-09-02 12:00:00"` // timestamp or formatted time
	Value float64 `json:"value" example:"43210.5"`            // close price
}

// GetPriceHistory godoc
// @Summary      Get 24h price history
// @Description  Returns the past 24h price history for a given symbol
// @Tags         Market
// @Accept       json
// @Produce      json
// @Param        symbol query string true "Trading symbol (e.g. btcusdt)"
// @Success      200 {object} interface{} "Price history data"
// @Failure      400 {object} map[string]string "Bad Request"
// @Failure      500 {object} map[string]string "Internal Server Error"
// @Router       /market/history [get]
func GetPriceHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol query param required"})
		return
	}

	history, err := services.GetPriceHistory(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Try encoding directly (whether it's [][]interface{} or map[string]interface{})
	c.JSON(http.StatusOK, history)
}
