package services

import (
	"encoding/json"
	"os"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/solchef/crypto-options-backend/config"
	"github.com/solchef/crypto-options-backend/events"
)

var priceWriter *kafka.Writer
var priceTopic string

func InitPriceProducer() {
	priceTopic = os.Getenv("KAFKA_TOPIC_PRICE_FEED")
	if priceTopic == "" {
		priceTopic = "price-feed"
	}
	priceWriter = config.NewWriter(priceTopic)
}

func PublishPriceTick(symbol string, price float64) error {
	msg := events.PriceTick{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now().UnixMilli(),
	}
	b, _ := json.Marshal(msg)
	return priceWriter.WriteMessages(config.KafkaCtx, kafka.Message{Value: b})
}
