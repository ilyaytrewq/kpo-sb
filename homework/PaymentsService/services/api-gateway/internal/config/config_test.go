package config

import "testing"

func TestMustLoadDefaults(t *testing.T) {
	t.Setenv("GATEWAY_HTTP_ADDR", "")
	t.Setenv("GATEWAY_BASE_PATH", "")
	t.Setenv("ORDERS_GRPC_ADDR", "")
	t.Setenv("PAYMENTS_GRPC_ADDR", "")

	cfg := MustLoad()
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
	if cfg.BasePath != "/api/v1" {
		t.Fatalf("BasePath = %q, want %q", cfg.BasePath, "/api/v1")
	}
	if cfg.OrdersGRPCAddr != "orders-service:9001" {
		t.Fatalf("OrdersGRPCAddr = %q, want %q", cfg.OrdersGRPCAddr, "orders-service:9001")
	}
	if cfg.PaymentsGRPCAddr != "payments-service:9002" {
		t.Fatalf("PaymentsGRPCAddr = %q, want %q", cfg.PaymentsGRPCAddr, "payments-service:9002")
	}
}

func TestMustLoadOverrides(t *testing.T) {
	t.Setenv("GATEWAY_HTTP_ADDR", ":9000")
	t.Setenv("GATEWAY_BASE_PATH", "/custom")
	t.Setenv("ORDERS_GRPC_ADDR", "orders:9999")
	t.Setenv("PAYMENTS_GRPC_ADDR", "payments:8888")

	cfg := MustLoad()
	if cfg.HTTPAddr != ":9000" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9000")
	}
	if cfg.BasePath != "/custom" {
		t.Fatalf("BasePath = %q, want %q", cfg.BasePath, "/custom")
	}
	if cfg.OrdersGRPCAddr != "orders:9999" {
		t.Fatalf("OrdersGRPCAddr = %q, want %q", cfg.OrdersGRPCAddr, "orders:9999")
	}
	if cfg.PaymentsGRPCAddr != "payments:8888" {
		t.Fatalf("PaymentsGRPCAddr = %q, want %q", cfg.PaymentsGRPCAddr, "payments:8888")
	}
}
