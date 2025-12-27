package app

import (
	"context"
	"log"
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
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
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
			log.Printf("failed to close writer: %v", err)
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
			log.Printf("failed to close reader: %v", err)
		}
	}()

	outbox := kafkasvc.NewOutboxPublisher(repo, writer, cfg.OutboxPollInterval, cfg.OutboxBatchSize)
	consumer := kafkasvc.NewPaymentRequestedConsumer(repo, reader, cfg.TopicPaymentResult)

	var cacheClient *redis.Client
	if cfg.RedisAddr != "" {
		cacheClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
		defer func() {
			if err := cacheClient.Close(); err != nil {
				log.Printf("failed to close redis client: %v", err)
			}
		}()
	}
	balanceCache := cache.NewBalanceCache(cacheClient, cfg.CacheTTL)

	grpcServer := grpc.NewServer()
	paymentsv1.RegisterPaymentsServiceServer(grpcServer, grpcsvc.NewHandlers(repo, balanceCache))
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Println("grpc listening on", cfg.GRPCAddr)
		return grpcServer.Serve(lis)
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Println("shutting down grpc...")
		grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error { return outbox.Run(ctx) })
	g.Go(func() error { return consumer.Run(ctx) })

	return g.Wait()
}
