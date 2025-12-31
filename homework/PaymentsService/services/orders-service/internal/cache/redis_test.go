package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewOrderCacheNilClient(t *testing.T) {
	if got := NewOrderCache(nil, time.Second); got != nil {
		t.Fatal("NewOrderCache(nil) should return nil")
	}
}

func TestOrderCacheNilReceiver(t *testing.T) {
	var c *OrderCache
	if got, err := c.Get(context.Background(), "order-1"); err != nil || got != nil {
		t.Fatalf("OrderCache.Get(nil) = (%v, %v), want (nil, nil)", got, err)
	}
	if err := c.Set(context.Background(), Order{OrderID: "order-1"}); err != nil {
		t.Fatalf("OrderCache.Set(nil) error: %v", err)
	}
}

func TestOrderCacheKey(t *testing.T) {
	if got := key("order-123"); got != "orders:order:order-123" {
		t.Fatalf("key() = %q, want %q", got, "orders:order:order-123")
	}
}
