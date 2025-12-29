package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyaytrewq/payments-service/api-gateway/internal/app"
	"github.com/ilyaytrewq/payments-service/api-gateway/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "api-gateway")
	slog.SetDefault(logger)

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("api gateway stopped with error", "err", err)
		os.Exit(1)
	}
}
