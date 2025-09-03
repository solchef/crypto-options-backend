// main.go
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/solchef/crypto-options-backend/docs" // swagger docs
	"github.com/solchef/crypto-options-backend/middleware"

	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/models"
	"github.com/solchef/crypto-options-backend/routes"
	"github.com/solchef/crypto-options-backend/services"
)

// @title Crypto Options API
// @version 1.0
// @description API documentation for Crypto Options platform
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /api
// @schemes http https

// Swagger security definition
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	}

	// Connect DB
	config.ConnectDB()

	// Auto Migrate
	config.DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Wallet{},
		&models.WalletTransaction{},
		&models.Trade{},
	)

	// Start price streams
	// go services.ListenPriceStream("btcusdt")
	// go services.ListenPriceStream("ethusdt")
	// go services.ListenTickerStream("btcusdt")
	// go services.ListenTickerStream("ethusdt")

	// Setup Gin
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())
	routes.RegisterRoutes(r)

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ✅ Add WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		config.ServeWS(c.Writer, c.Request)
	})

	r.GET("/trading", func(c *gin.Context) {
		services.ServeWS(c.Writer, c.Request)
	})

	// Run server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
