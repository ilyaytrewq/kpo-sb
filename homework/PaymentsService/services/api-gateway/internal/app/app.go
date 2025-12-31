package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
	paymentsv1 "github.com/ilyaytrewq/payments-service/gen/go/payments/v1"
	gateway "github.com/ilyaytrewq/payments-service/gen/openapi/gateway"

	"github.com/ilyaytrewq/payments-service/api-gateway/internal/config"
	"github.com/ilyaytrewq/payments-service/api-gateway/internal/handler"
)

func Run(ctx context.Context, cfg config.Config) error {
	start := time.Now()
	logger := slog.Default().With("service", "api-gateway", "component", "app")
	logger.Info("api gateway starting", "http_addr", cfg.HTTPAddr, "base_path", cfg.BasePath)
	ordersConn, err := grpc.DialContext(ctx, cfg.OrdersGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to dial orders grpc", "err", err, "addr", cfg.OrdersGRPCAddr)
		return err
	}
	defer ordersConn.Close()

	paymentsConn, err := grpc.DialContext(ctx, cfg.PaymentsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to dial payments grpc", "err", err, "addr", cfg.PaymentsGRPCAddr)
		return err
	}
	defer paymentsConn.Close()

	apiHandler := handler.New(
		ordersv1.NewOrdersServiceClient(ordersConn),
		paymentsv1.NewPaymentsServiceClient(paymentsConn),
	)

	router := chi.NewRouter()
	router.Use(requestLogger)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:8088",
			"http://127.0.0.1:8088",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://158.160.219.201:8088",
		},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-User-Id",
			"Idempotency-Key",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, cfg.BasePath) {
				if strings.TrimSpace(r.Header.Get("Idempotency-Key")) == "" {
					userID := r.Header.Get("X-User-Id")
					logger.Error("missing idempotency key", "path", r.URL.Path, "user_id", userID)
					handler.WriteBadRequest(w, userID, errors.New("idempotency key is required"))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	router.Get("/health", healthHandler)
	router.Get("/heath", healthHandler)

	// Важно: preflight OPTIONS должен матчиться роутером, иначе будет 404 и "Failed to fetch"
	router.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	gateway.HandlerWithOptions(apiHandler, gateway.ChiServerOptions{
		BaseURL:    cfg.BasePath,
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			userID := r.Header.Get("X-User-Id")
			logger.Error("gateway handler error", "err", err, "path", r.URL.Path, "user_id", userID)
			handler.WriteBadRequest(w, userID, err)
		},
	})

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("gateway listening", "http_addr", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error("gateway shutdown failed", "err", err, "duration", time.Since(start))
			return err
		}
		logger.Info("gateway shutdown completed", "duration", time.Since(start))
		return nil
	case err := <-errCh:
		if err != nil {
			logger.Error("gateway stopped with error", "err", err, "duration", time.Since(start))
		} else {
			logger.Info("gateway stopped", "duration", time.Since(start))
		}
		return err
	}
}
