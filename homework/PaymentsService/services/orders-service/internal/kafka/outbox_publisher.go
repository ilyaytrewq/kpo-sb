package kafka

import (
	"context"
	"log"
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
	return &OutboxPublisher{repo: repo, w: w, interval: interval, batch: batch}
}

func (p *OutboxPublisher) Run(ctx context.Context) error {
	t := time.NewTicker(p.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := p.publishOnce(ctx); err != nil {
				log.Println("outbox publish error:", err)
			}
		}
	}
}

func (p *OutboxPublisher) publishOnce(ctx context.Context) error {
	return p.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		rows, err := q.LockUnsentOutbox(ctx, int32(p.batch))
		if err != nil {
			return err
		}
		if len(rows) == 0 {
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
				continue
			}

			if err := q.MarkOutboxSent(ctx, r.ID); err != nil {
				return err
			}
		}

		return nil
	})
}
