package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"

	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres"
)

type OutboxPublisher struct {
	repo     *postgres.Repo
	w        *kafka.Writer
	interval time.Duration
	batch    int
}

func NewOutboxPublisher(repo *postgres.Repo, w *kafka.Writer, interval time.Duration, batch int) *OutboxPublisher {
	slog.Default().With("service", "orders-service", "component", "kafka").Info("outbox publisher initialized", "interval", interval.String(), "batch", batch)
	return &OutboxPublisher{repo: repo, w: w, interval: interval, batch: batch}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "kafka")
	logger.Info("outbox publisher run start", "interval", p.interval.String(), "batch", p.batch)
	t := time.NewTicker(p.interval)
	defer t.Stop()
	defer func() {
		logger.Info("outbox publisher stopped", "duration", time.Since(start))
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Info("outbox publisher context done")
			return nil
		case <-t.C:
			if err := p.publishOnce(ctx); err != nil {
				logger.Error("outbox publish error", "err", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishOnce(ctx context.Context) error {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "kafka")
	logger.Info("outbox publish cycle start")
	return p.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		rows, err := q.LockUnsentOutbox(ctx, int32(p.batch))
		if err != nil {
			logger.Error("failed to lock unsent outbox rows", "err", err)
			return err
		}
		if len(rows) == 0 {
			logger.Info("outbox publish cycle empty", "duration", time.Since(start))
			return nil
		}

		for _, r := range rows {
			msg := kafka.Message{
				Key:   []byte(r.KafkaKey),
				Value: r.Payload,
			}

			if err := p.w.WriteMessages(ctx, msg); err != nil {
				_ = q.MarkOutboxAttemptFailed(ctx, db.MarkOutboxAttemptFailedParams{
					ID: r.ID,
					LastError: pgtype.Text{
						String: err.Error(),
						Valid:  true,
					},
				})
				logger.Error("failed to publish outbox message", "err", err, "outbox_id", r.ID, "kafka_key", r.KafkaKey)
				continue
			}

			if err := q.MarkOutboxSent(ctx, r.ID); err != nil {
				logger.Error("failed to mark outbox as sent", "err", err, "outbox_id", r.ID)
				return err
			}
			logger.Info("outbox message published", "outbox_id", r.ID, "kafka_key", r.KafkaKey)
		}

		logger.Info("outbox publish cycle completed", "count", len(rows), "duration", time.Since(start))
		return nil
	})
}
