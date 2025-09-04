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
	topic := os.Getenv("KAFKA_TOPIC_PRICE_FEED")
	if topic == "" {
		topic = "price-feed"
	}

	reader := config.NewReader(topic, "price-consumers-1")

	go func() {
		defer reader.Close()
		for {
			m, err := reader.ReadMessage(context.Background())
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

			// broadcast to existing WS (keeps your current UI happy)
			//BroadcastPrice(tick.Symbol, tick.Price)
		}
	}()
}
