package app

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"services/orders/internal/config"
	grpcsvc "services/orders/internal/grpc"
	kafkasvc "services/orders/internal/kafka"
	"services/orders/internal/repo/postgres"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
)

func Run(ctx context.Context, cfg config.Config) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := postgres.New(pool)

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.KafkaBrokers...),
		Topic:        cfg.TopicPaymentRequested,
		RequiredAcks: kafka.RequireAll,
		Balancer:     &kafka.Hash{},
		BatchTimeout: 50 * time.Millisecond,
	}
	defer writer.Close()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.KafkaBrokers,
		Topic:       cfg.TopicPaymentResult,
		GroupID:     cfg.ConsumerGroupID,
		MinBytes:    1e3,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	outbox := kafkasvc.NewOutboxPublisher(repo, writer, cfg.OutboxPollInterval, cfg.OutboxBatchSize)
	consumer := kafkasvc.NewPaymentResultConsumer(repo, reader)

	grpcServer := grpc.NewServer()
	ordersv1.RegisterOrdersServiceServer(grpcServer, grpcsvc.NewHandlers(repo))
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
