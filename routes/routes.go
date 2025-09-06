package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/solchef/crypto-options-backend/controllers"
	"github.com/solchef/crypto-options-backend/middleware"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")

	//markets
	api.GET("/market/history", controllers.GetPriceHistory)
	// Public routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/refresh", controllers.Refresh)
		auth.POST("/logout", controllers.Logout)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/users", controllers.GetUsers)
		protected.POST("/users", controllers.CreateUser)
		protected.GET("/profile", controllers.Profile)

		// Trades
		protected.POST("/trades/place", controllers.PlaceTrade)
		protected.GET("/trades/open", controllers.GetOpenTrades)
		protected.GET("/trades/history", controllers.GetTradeHistory)
		protected.POST("/trades/close", controllers.CloseTrade)

		//wallets
		protected.GET("/wallets", controllers.GetWallets)
		protected.POST("/wallets/deposit", controllers.Deposit)
		protected.POST("/wallets/withdraw", controllers.Withdraw)
		protected.GET("/wallets/transactions", controllers.GetWalletTransactions)
	}
}
