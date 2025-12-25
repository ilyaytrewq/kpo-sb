package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"services/orders/internal/app"
	"services/orders/internal/config"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg); err != nil {
		log.Println("orders service stopped with error:", err)
		os.Exit(1)
	}
}
