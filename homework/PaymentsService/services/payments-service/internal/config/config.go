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

	ConsumerGroupID string

	OutboxPollInterval time.Duration
	OutboxBatchSize    int

	RedisAddr string
	CacheTTL  time.Duration
}

func MustLoad() Config {
	return Config{
		GRPCAddr: getenv("PAYMENTS_GRPC_ADDR", ":9002"),

		DatabaseURL: getenv("PAYMENTS_DATABASE_URL", "postgres://postgres:postgres@payments-postgres:5432/payments?sslmode=disable"),

		KafkaBrokers: strings.Split(getenv("KAFKA_BROKERS", "broker:9092"), ","),

		TopicPaymentRequested: getenv("KAFKA_TOPIC_PAYMENT_REQUESTED", "payments.payment_requested.v1"),
		TopicPaymentResult:    getenv("KAFKA_TOPIC_PAYMENT_RESULT", "payments.payment_result.v1"),

		ConsumerGroupID: getenv("KAFKA_PAYMENTS_GROUP_ID", "payments-service"),

		OutboxPollInterval: getenvDuration("OUTBOX_POLL_INTERVAL", 500*time.Millisecond),
		OutboxBatchSize:    getenvInt("OUTBOX_BATCH_SIZE", 50),

		RedisAddr: getenv("PAYMENTS_REDIS_ADDR", "redis:6379"),
		CacheTTL:  getenvDuration("PAYMENTS_CACHE_TTL", 30*time.Second),
	}
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
