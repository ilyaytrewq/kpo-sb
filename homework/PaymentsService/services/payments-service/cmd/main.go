package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyaytrewq/payments-service/payments-service/internal/app"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).
		With("service", "payments-service")
	slog.SetDefault(logger)

	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("payments service stopped with error", "err", err)
		os.Exit(1)
	}
}
