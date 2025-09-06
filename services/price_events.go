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
var tickerWriter *kafka.Writer
var tickerTopic string

func InitPriceProducer() {
	priceTopic = os.Getenv("KAFKA_TOPIC_PRICE_FEED")
	tickerTopic = os.Getenv("KAFKA_TOPIC_TICKER_STATS")
	priceWriter = config.NewWriter(priceTopic)
	tickerWriter = config.NewWriter(tickerTopic)
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

func PublishTickerStats(symbol string, change float64, changePct float64) error {
	msg := events.TickerStats{
		Symbol:    symbol,
		Change:    change,
		ChangePct: changePct,
		Timestamp: time.Now().UnixMilli(),
	}
	b, _ := json.Marshal(msg)
	return tickerWriter.WriteMessages(config.KafkaCtx, kafka.Message{Value: b})
}
