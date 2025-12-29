package app

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/ilyaytrewq/payments-service/payments-service/internal/cache"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/config"
	grpcsvc "github.com/ilyaytrewq/payments-service/payments-service/internal/grpc"
	kafkasvc "github.com/ilyaytrewq/payments-service/payments-service/internal/kafka"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/repo/postgres"

	paymentsv1 "github.com/ilyaytrewq/payments-service/gen/go/payments/v1"
)

func Run(ctx context.Context, cfg config.Config) error {
	start := time.Now()
	logger := slog.Default().With("service", "payments-service", "component", "app")
	logger.Info("payments service starting", "grpc_addr", cfg.GRPCAddr, "redis_addr", cfg.RedisAddr != "", "kafka_brokers", len(cfg.KafkaBrokers))

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create db pool", "err", err)
		return err
	}
	defer pool.Close()

	repo := postgres.NewRepo(pool)

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers...),
		Topic:        cfg.TopicPaymentResult,
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 50 * time.Millisecond,
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("failed to close kafka writer", "err", err)
		}
	}()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.KafkaBrokers,
		Topic:          cfg.TopicPaymentRequested,
		GroupID:        cfg.ConsumerGroupID,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("failed to close kafka reader", "err", err)
		}
	}()

	outbox := kafkasvc.NewOutboxPublisher(repo, writer, cfg.OutboxPollInterval, cfg.OutboxBatchSize)
	consumer := kafkasvc.NewPaymentRequestedConsumer(repo, reader, cfg.TopicPaymentResult)

	var cacheClient *redis.Client
	if cfg.RedisAddr != "" {
		cacheClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		defer func() {
			if err := cacheClient.Close(); err != nil {
				logger.Error("failed to close redis client", "err", err)
			}
		}()
	}
	balanceCache := cache.NewBalanceCache(cacheClient, cfg.CacheTTL)

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcUnaryLogger()))
	paymentsv1.RegisterPaymentsServiceServer(grpcServer, grpcsvc.NewHandlers(repo, balanceCache))
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
			logger.Error("payment requested consumer stopped with error", "err", err)
		}
		return err
	})

	err = g.Wait()
	if err != nil {
		logger.Error("payments service stopped with error", "err", err, "duration", time.Since(start))
	} else {
		logger.Info("payments service stopped", "duration", time.Since(start))
	}
	return err
}
