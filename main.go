// main.go
package main

import (
	"log"
	"os"
	"strings"

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

	// Redis + Kafka
	config.ConnectRedis()
	config.InitKafka()
	services.InitPriceProducer()
	services.StartPriceConsumer() // fan-out price ticks

	// Start Binance price streams (multiple symbols)
	syms := os.Getenv("SYMBOLS")
	if syms == "" {
		syms = "btcusdt,ethusdt"
	}
	for _, s := range strings.Split(syms, ",") {
		symbol := strings.TrimSpace(s)
		if symbol == "" {
			continue
		}
		priceChan := make(chan float64)
		go services.ListenPriceStream(symbol, priceChan)
	}

	// Auto Migrate
	config.DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Wallet{},
		&models.WalletTransaction{},
		&models.Trade{},
	)

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
