package services

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/events"
)

// StartPriceConsumer reads ticks from Kafka and updates Redis + broadcasts WS.
func StartPriceConsumer() {
	priceTopic := os.Getenv("KAFKA_TOPIC_PRICE_FEED")
	tickerTopic := os.Getenv("KAFKA_TOPIC_TICKER_STATS")

	priceReader := config.NewReader(priceTopic, "price-consumers-1")
	tickerReader := config.NewReader(tickerTopic, "ticker-consumers-1")

	go func() {
		defer priceReader.Close()
		for {
			m, err := priceReader.ReadMessage(context.Background())
			if err != nil {
				log.Println("kafka price consumer error:", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			var tick events.PriceTick
			if err := json.Unmarshal(m.Value, &tick); err != nil {
				continue
			}

			// cache latest price in Redis
			config.Redis.Set(config.Ctx, "price:"+tick.Symbol, tick.Price, 0)

			Broadcast(WSMessage{
				Type: "price_update",
				Data: map[string]interface{}{
					"symbol": tick.Symbol,
					"price":  tick.Price,
				},
				Timestamp: time.Now().UnixMilli(),
			})

		}
	}()

	go func() {
		defer tickerReader.Close()
		for {
			m, err := tickerReader.ReadMessage(context.Background())
			if err != nil {
				log.Println("kafka ticker consumer error:", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			var stats events.TickerStats
			if err := json.Unmarshal(m.Value, &stats); err != nil {
				continue
			}

			// cache latest ticker stats in Redis
			config.Redis.HSet(config.Ctx, "ticker:"+stats.Symbol, map[string]interface{}{
				"change":    stats.Change,
				"changePct": stats.ChangePct,
			})

			// Broadcast(WSMessage{
			// 	Type: "ticker_update",
			// 	Data: map[string]interface{}{
			// 		"symbol":    stats.Symbol,
			// 		"change":    stats.Change,
			// 		"changePct": stats.ChangePct,
			// 	},
			// 	Timestamp: time.Now().UnixMilli(),
			// })
		}
	}()
}
