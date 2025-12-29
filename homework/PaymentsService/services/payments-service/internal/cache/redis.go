package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type BalanceCache struct {
	client *redis.Client
	ttl    time.Duration
}

type Balance struct {
	UserID  string `json:"user_id"`
	Balance int64  `json:"balance"`
}

func NewBalanceCache(client *redis.Client, ttl time.Duration) *BalanceCache {
	if client == nil {
		slog.Default().With("service", "payments-service", "component", "cache").Info("balance cache disabled")
		return nil
	}
	slog.Default().With("service", "payments-service", "component", "cache").Info("balance cache initialized", "ttl", ttl.String())
	return &BalanceCache{client: client, ttl: ttl}
}

func (c *BalanceCache) Get(ctx context.Context, userID string) (*Balance, error) {
	start := time.Now()
	logger := slog.Default().With("service", "payments-service", "component", "cache")
	if c == nil {
		logger.Info("balance cache get skipped (nil cache)", "user_id", userID)
		return nil, nil
	}
	val, err := c.client.Get(ctx, key(userID)).Result()
	if err == redis.Nil {
		logger.Info("balance cache miss", "user_id", userID, "duration", time.Since(start))
		return nil, nil
	}
	if err != nil {
		logger.Error("balance cache get failed", "user_id", userID, "err", err, "duration", time.Since(start))
		return nil, err
	}
	var cached Balance
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		logger.Error("balance cache unmarshal failed", "user_id", userID, "err", err, "duration", time.Since(start))
		return nil, err
	}
	logger.Info("balance cache hit", "user_id", userID, "duration", time.Since(start))
	return &cached, nil
}

func (c *BalanceCache) Set(ctx context.Context, balance Balance) error {
	start := time.Now()
	logger := slog.Default().With("service", "payments-service", "component", "cache")
	if c == nil {
		logger.Info("balance cache set skipped (nil cache)", "user_id", balance.UserID)
		return nil
	}
	data, err := json.Marshal(balance)
	if err != nil {
		logger.Error("balance cache marshal failed", "user_id", balance.UserID, "err", err, "duration", time.Since(start))
		return err
	}
	if err := c.client.Set(ctx, key(balance.UserID), data, c.ttl).Err(); err != nil {
		logger.Error("balance cache set failed", "user_id", balance.UserID, "err", err, "duration", time.Since(start))
		return err
	}
	logger.Info("balance cache set", "user_id", balance.UserID, "duration", time.Since(start))
	return nil
}

func key(userID string) string {
	slog.Default().With("service", "payments-service", "component", "cache").Info("balance cache key generated", "user_id", userID)
	return "payments:balance:" + userID
}
