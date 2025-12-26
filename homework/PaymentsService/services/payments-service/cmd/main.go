package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyaytrewq/payments-service/payments-service/internal/app"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Println("payments service stopped with error:", err)
		os.Exit(1)
	}
}
