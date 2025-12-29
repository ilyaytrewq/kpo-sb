package app

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
	"github.com/ilyaytrewq/payments-service/order-service/internal/cache"
	"github.com/ilyaytrewq/payments-service/order-service/internal/config"
	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpcsvc "github.com/ilyaytrewq/payments-service/order-service/internal/grpc"
	kafkasvc "github.com/ilyaytrewq/payments-service/order-service/internal/kafka"
)

func Run(ctx context.Context, cfg config.Config) error {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "app")
	logger.Info("orders service starting", "grpc_addr", cfg.GRPCAddr, "redis_addr", cfg.RedisAddr != "", "kafka_brokers", len(cfg.KafkaBrokers))

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create db pool", "err", err)
		return err
	}
	defer pool.Close()

	repo := postgres.NewRepo(pool)

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers...),
		Topic:        cfg.TopicPaymentRequested,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 50 * time.Millisecond,
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("failed to close kafka writer", "err", err)
		}
	}()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		Topic:          cfg.TopicPaymentResult,
		GroupID:        cfg.ConsumerGroupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: 0,
	})
	defer reader.Close()

	outbox := kafkasvc.NewOutboxPublisher(repo, writer, cfg.OutboxPollInterval, cfg.OutboxBatchSize)
	consumer := kafkasvc.NewPaymentResultConsumer(repo, reader)

	var cacheClient *redis.Client
	if cfg.RedisAddr != "" {
		cacheClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		defer func() {
			if err := cacheClient.Close(); err != nil {
				logger.Error("failed to close redis client", "err", err)
			}
		}()
	}
	orderCache := cache.NewOrderCache(cacheClient, cfg.CacheTTL)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcUnaryLogger()))
	ordersv1.RegisterOrdersServiceServer(grpcServer, grpcsvc.NewHandlers(repo, orderCache))
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		logger.Error("failed to listen on grpc address", "err", err, "grpc_addr", cfg.GRPCAddr)
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("grpc listening", "grpc_addr", cfg.GRPCAddr)
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("grpc shutting down")
		grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error {
		err := outbox.Run(ctx)
		if err != nil {
			logger.Error("outbox publisher stopped with error", "err", err)
		}
		return err
	})
	g.Go(func() error {
		err := consumer.Run(ctx)
		if err != nil {
			logger.Error("payment result consumer stopped with error", "err", err)
		}
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Error("orders service stopped with error", "err", err, "duration", time.Since(start))
	} else {
		logger.Info("orders service stopped", "duration", time.Since(start))
	}
	return err
}
