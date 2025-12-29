package app

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func grpcUnaryLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		logger := slog.Default().With("service", "orders-service", "component", "grpc")
		if err != nil {
			logger.Error("grpc request failed", "method", info.FullMethod, "code", code.String(), "duration", time.Since(start), "err", err)
		} else {
			logger.Info("grpc request completed", "method", info.FullMethod, "code", code.String(), "duration", time.Since(start))
		}
		return resp, err
	}
}
