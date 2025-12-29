package config

import "testing"

func TestMustLoadDefaults(t *testing.T) {
	t.Setenv("PAYMENTS_GRPC_ADDR", "")
	t.Setenv("PAYMENTS_DATABASE_URL", "")
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_TOPIC_PAYMENT_REQUESTED", "")
	t.Setenv("KAFKA_TOPIC_PAYMENT_RESULT", "")
	t.Setenv("KAFKA_PAYMENTS_GROUP_ID", "")
	t.Setenv("OUTBOX_POLL_INTERVAL", "")
	t.Setenv("OUTBOX_BATCH_SIZE", "")
	t.Setenv("PAYMENTS_REDIS_ADDR", "")
	t.Setenv("PAYMENTS_CACHE_TTL", "")

	cfg := MustLoad()
	if cfg.GRPCAddr != ":9002" {
		t.Fatalf("GRPCAddr = %q, want %q", cfg.GRPCAddr, ":9002")
	}
	if cfg.DatabaseURL != "postgres://postgres:postgres@payments-postgres:5432/payments?sslmode=disable" {
		t.Fatalf("DatabaseURL = %q, want default", cfg.DatabaseURL)
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "broker:9092" {
		t.Fatalf("KafkaBrokers = %v, want [broker:9092]", cfg.KafkaBrokers)
	}
	if cfg.TopicPaymentRequested != "payments.payment_requested.v1" {
		t.Fatalf("TopicPaymentRequested = %q, want %q", cfg.TopicPaymentRequested, "payments.payment_requested.v1")
	}
	if cfg.TopicPaymentResult != "payments.payment_result.v1" {
		t.Fatalf("TopicPaymentResult = %q, want %q", cfg.TopicPaymentResult, "payments.payment_result.v1")
	}
	if cfg.ConsumerGroupID != "payments-service" {
		t.Fatalf("ConsumerGroupID = %q, want %q", cfg.ConsumerGroupID, "payments-service")
	}
	if cfg.OutboxPollInterval.String() != "500ms" {
		t.Fatalf("OutboxPollInterval = %s, want %s", cfg.OutboxPollInterval, "500ms")
	}
	if cfg.OutboxBatchSize != 50 {
		t.Fatalf("OutboxBatchSize = %d, want %d", cfg.OutboxBatchSize, 50)
	}
	if cfg.RedisAddr != "redis:6379" {
		t.Fatalf("RedisAddr = %q, want %q", cfg.RedisAddr, "redis:6379")
	}
	if cfg.CacheTTL.String() != "30s" {
		t.Fatalf("CacheTTL = %s, want %s", cfg.CacheTTL, "30s")
	}
}

func TestMustLoadOverrides(t *testing.T) {
	t.Setenv("PAYMENTS_GRPC_ADDR", ":9200")
	t.Setenv("PAYMENTS_DATABASE_URL", "postgres://x:y@host:2222/db")
	t.Setenv("KAFKA_BROKERS", "a:1,b:2")
	t.Setenv("KAFKA_TOPIC_PAYMENT_REQUESTED", "t.req")
	t.Setenv("KAFKA_TOPIC_PAYMENT_RESULT", "t.res")
	t.Setenv("KAFKA_PAYMENTS_GROUP_ID", "payments-group")
	t.Setenv("OUTBOX_POLL_INTERVAL", "2s")
	t.Setenv("OUTBOX_BATCH_SIZE", "123")
	t.Setenv("PAYMENTS_REDIS_ADDR", "redis:9999")
	t.Setenv("PAYMENTS_CACHE_TTL", "45s")

	cfg := MustLoad()
	if cfg.GRPCAddr != ":9200" {
		t.Fatalf("GRPCAddr = %q, want %q", cfg.GRPCAddr, ":9200")
	}
	if cfg.DatabaseURL != "postgres://x:y@host:2222/db" {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://x:y@host:2222/db")
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[0] != "a:1" || cfg.KafkaBrokers[1] != "b:2" {
		t.Fatalf("KafkaBrokers = %v, want [a:1 b:2]", cfg.KafkaBrokers)
	}
	if cfg.TopicPaymentRequested != "t.req" {
		t.Fatalf("TopicPaymentRequested = %q, want %q", cfg.TopicPaymentRequested, "t.req")
	}
	if cfg.TopicPaymentResult != "t.res" {
		t.Fatalf("TopicPaymentResult = %q, want %q", cfg.TopicPaymentResult, "t.res")
	}
	if cfg.ConsumerGroupID != "payments-group" {
		t.Fatalf("ConsumerGroupID = %q, want %q", cfg.ConsumerGroupID, "payments-group")
	}
	if cfg.OutboxPollInterval.String() != "2s" {
		t.Fatalf("OutboxPollInterval = %s, want %s", cfg.OutboxPollInterval, "2s")
	}
	if cfg.OutboxBatchSize != 123 {
		t.Fatalf("OutboxBatchSize = %d, want %d", cfg.OutboxBatchSize, 123)
	}
	if cfg.RedisAddr != "redis:9999" {
		t.Fatalf("RedisAddr = %q, want %q", cfg.RedisAddr, "redis:9999")
	}
	if cfg.CacheTTL.String() != "45s" {
		t.Fatalf("CacheTTL = %s, want %s", cfg.CacheTTL, "45s")
	}
}

func TestMustLoadInvalidOverridesFallback(t *testing.T) {
	t.Setenv("OUTBOX_POLL_INTERVAL", "bad")
	t.Setenv("OUTBOX_BATCH_SIZE", "nope")
	t.Setenv("PAYMENTS_CACHE_TTL", "bad")

	cfg := MustLoad()
	if cfg.OutboxPollInterval.String() != "500ms" {
		t.Fatalf("OutboxPollInterval = %s, want %s", cfg.OutboxPollInterval, "500ms")
	}
	if cfg.OutboxBatchSize != 50 {
		t.Fatalf("OutboxBatchSize = %d, want %d", cfg.OutboxBatchSize, 50)
	}
	if cfg.CacheTTL.String() != "30s" {
		t.Fatalf("CacheTTL = %s, want %s", cfg.CacheTTL, "30s")
	}
}
