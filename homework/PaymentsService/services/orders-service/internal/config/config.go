package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCAddr string

	DatabaseURL string

	KafkaBrokers []string

	TopicPaymentRequested string
	TopicPaymentResult    string

	OutboxPollInterval time.Duration
	OutboxBatchSize    int

	ConsumerGroupID string
}

func MustLoad() Config {
	cfg := Config{
		GRPCAddr:    getenv("ORDERS_GRPC_ADDR", ":9001"),
		DatabaseURL: getenv("ORDERS_DATABASE_URL", "postgres://postgres:postgres@orders-postgres:5432/orders?sslmode=disable"),

		KafkaBrokers: strings.Split(getenv("KAFKA_BROKERS", "broker:9092"), ","),

		TopicPaymentRequested: getenv("KAFKA_TOPIC_PAYMENT_REQUESTED", "payments.payment_requested.v1"),
		TopicPaymentResult:    getenv("KAFKA_TOPIC_PAYMENT_RESULT", "payments.payment_result.v1"),

		OutboxPollInterval: getenvDuration("OUTBOX_POLL_INTERVAL", 500*time.Millisecond),
		OutboxBatchSize:    getenvInt("OUTBOX_BATCH_SIZE", 50),

		ConsumerGroupID: getenv("KAFKA_ORDERS_GROUP_ID", "orders-service"),
	}
	return cfg
}

func getenv(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

func getenvInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}

func getenvDuration(k string, d time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	dd, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return dd
}
