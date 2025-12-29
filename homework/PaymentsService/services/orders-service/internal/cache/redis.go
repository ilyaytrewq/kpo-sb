package cache

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

type Order struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewOrderCache(client *redis.Client, ttl time.Duration) *OrderCache {
	if client == nil {
		slog.Default().With("service", "orders-service", "component", "cache").Info("order cache disabled")
		return nil
	}
	slog.Default().With("service", "orders-service", "component", "cache").Info("order cache initialized", "ttl", ttl.String())
	return &OrderCache{client: client, ttl: ttl}
}

func (c *OrderCache) Get(ctx context.Context, orderID string) (*Order, error) {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "cache")
	if c == nil {
		logger.Info("order cache get skipped (nil cache)", "order_id", orderID)
		return nil, nil
	}
	val, err := c.client.Get(ctx, key(orderID)).Result()
	if err == redis.Nil {
		logger.Info("order cache miss", "order_id", orderID, "duration", time.Since(start))
		return nil, nil
	}
	if err != nil {
		logger.Error("order cache get failed", "order_id", orderID, "err", err, "duration", time.Since(start))
		return nil, err
	}
	var cached Order
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		logger.Error("order cache unmarshal failed", "order_id", orderID, "err", err, "duration", time.Since(start))
		return nil, err
	}
	logger.Info("order cache hit", "order_id", orderID, "duration", time.Since(start))
	return &cached, nil
}

func (c *OrderCache) Set(ctx context.Context, order Order) error {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "cache")
	if c == nil {
		logger.Info("order cache set skipped (nil cache)", "order_id", order.OrderID)
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		logger.Error("order cache marshal failed", "order_id", order.OrderID, "err", err, "duration", time.Since(start))
		return err
	}
	if err := c.client.Set(ctx, key(order.OrderID), data, c.ttl).Err(); err != nil {
		logger.Error("order cache set failed", "order_id", order.OrderID, "err", err, "duration", time.Since(start))
		return err
	}
	logger.Info("order cache set", "order_id", order.OrderID, "duration", time.Since(start))
	return nil
}

func key(orderID string) string {
	slog.Default().With("service", "orders-service", "component", "cache").Info("order cache key generated", "order_id", orderID)
	return "orders:order:" + orderID
}
