package config

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers []string
}

var Kafka KafkaConfig

func InitKafka() {
	bs := os.Getenv("KAFKA_BROKERS")
	if bs == "" {
		bs = "localhost:9092"
	}
	Kafka = KafkaConfig{Brokers: strings.Split(bs, ",")}
}

func NewWriter(topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(Kafka.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
	}
}

func NewReader(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     Kafka.Brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafka.LastOffset, // start from latest
		MaxWait:     50 * time.Millisecond,
	})
}

var KafkaCtx = context.Background()
