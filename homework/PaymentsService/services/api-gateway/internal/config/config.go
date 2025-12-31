package config

import "os"

type Config struct {
	HTTPAddr         string
	BasePath         string
	OrdersGRPCAddr   string
	PaymentsGRPCAddr string
}

func MustLoad() Config {
	return Config{
		HTTPAddr:         getenv("GATEWAY_HTTP_ADDR", ":8080"),
		BasePath:         getenv("GATEWAY_BASE_PATH", "/api/v1"),
		OrdersGRPCAddr:   getenv("ORDERS_GRPC_ADDR", "orders-service:9001"),
		PaymentsGRPCAddr: getenv("PAYMENTS_GRPC_ADDR", "payments-service:9002"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
