package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
	paymentsv1 "github.com/ilyaytrewq/payments-service/gen/go/payments/v1"
	gateway "github.com/ilyaytrewq/payments-service/gen/openapi/gateway"

	"github.com/ilyaytrewq/payments-service/api-gateway/internal/config"
	"github.com/ilyaytrewq/payments-service/api-gateway/internal/handler"
)

func Run(ctx context.Context, cfg config.Config) error {
	ordersConn, err := grpc.DialContext(ctx, cfg.OrdersGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer ordersConn.Close()

	paymentsConn, err := grpc.DialContext(ctx, cfg.PaymentsGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer paymentsConn.Close()

	apiHandler := handler.New(
		ordersv1.NewOrdersServiceClient(ordersConn),
		paymentsv1.NewPaymentsServiceClient(paymentsConn),
	)

	router := chi.NewRouter()

	gateway.HandlerWithOptions(apiHandler, gateway.ChiServerOptions{
		BaseURL:    cfg.BasePath,
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			userID := r.Header.Get("X-User-Id")
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
		log.Printf("gateway listening on %s", cfg.HTTPAddr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
