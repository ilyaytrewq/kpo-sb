package cache

import (
	"context"
	"encoding/json"
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
		return nil
	}
	return &BalanceCache{client: client, ttl: ttl}
}

func (c *BalanceCache) Get(ctx context.Context, userID string) (*Balance, error) {
	if c == nil {
		return nil, nil
	}
	val, err := c.client.Get(ctx, key(userID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cached Balance
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func (c *BalanceCache) Set(ctx context.Context, balance Balance) error {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(balance)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(balance.UserID), data, c.ttl).Err()
}

func key(userID string) string {
	return "payments:balance:" + userID
}
