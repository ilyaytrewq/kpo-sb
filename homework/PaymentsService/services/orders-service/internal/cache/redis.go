package cache

import (
	"context"
	"encoding/json"
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
		return nil
	}
	return &OrderCache{client: client, ttl: ttl}
}

func (c *OrderCache) Get(ctx context.Context, orderID string) (*Order, error) {
	if c == nil {
		return nil, nil
	}
	val, err := c.client.Get(ctx, key(orderID)).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var cached Order
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

func (c *OrderCache) Set(ctx context.Context, order Order) error {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key(order.OrderID), data, c.ttl).Err()
}

func key(orderID string) string {
	return "orders:order:" + orderID
}
