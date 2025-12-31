package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewBalanceCacheNilClient(t *testing.T) {
	if got := NewBalanceCache(nil, time.Second); got != nil {
		t.Fatal("NewBalanceCache(nil) should return nil")
	}
}

func TestBalanceCacheNilReceiver(t *testing.T) {
	var c *BalanceCache
	if got, err := c.Get(context.Background(), "user-1"); err != nil || got != nil {
		t.Fatalf("BalanceCache.Get(nil) = (%v, %v), want (nil, nil)", got, err)
	}
	if err := c.Set(context.Background(), Balance{UserID: "user-1", Balance: 10}); err != nil {
		t.Fatalf("BalanceCache.Set(nil) error: %v", err)
	}
}

func TestBalanceCacheKey(t *testing.T) {
	if got := key("user-123"); got != "payments:balance:user-123" {
		t.Fatalf("key() = %q, want %q", got, "payments:balance:user-123")
	}
}
