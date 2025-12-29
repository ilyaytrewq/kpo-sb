package kafka

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"

	eventsv1 "github.com/ilyaytrewq/payments-service/gen/go/events/v1"

	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres"
)

type PaymentResultConsumer struct {
	repo   *postgres.Repo
	reader *kafka.Reader
}

func NewPaymentResultConsumer(repo *postgres.Repo, r *kafka.Reader) *PaymentResultConsumer {
	slog.Default().With("service", "orders-service", "component", "kafka").Info("payment result consumer initialized")
	return &PaymentResultConsumer{repo: repo, reader: r}
}

func (c *PaymentResultConsumer) Run(ctx context.Context) error {
	logger := slog.Default().With("service", "orders-service", "component", "kafka")
	logger.Info("payment result consumer run start")
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("payment result consumer context done")
				return nil
			}
			logger.Error("payment result fetch failed", "err", err)
			return err
		}

		if err := c.handleMessage(ctx, m); err != nil {
			logger.Error("payment result handle error", "err", err, "offset", m.Offset)
			// offset НЕ коммитим => Kafka доставит снова
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			logger.Error("payment result commit failed", "err", err, "offset", m.Offset)
			return err
		}
		logger.Info("payment result message committed", "offset", m.Offset)
	}
}

func (c *PaymentResultConsumer) handleMessage(ctx context.Context, m kafka.Message) error {
	logger := slog.Default().With("service", "orders-service", "component", "kafka")
	logger.Info("payment result handle message start", "offset", m.Offset)
	var ev eventsv1.PaymentResult
	if err := proto.Unmarshal(m.Value, &ev); err != nil {
		// плохое сообщение лучше “проглотить” и закоммитить, иначе будет бесконечный цикл
		logger.Error("payment result unmarshal failed", "err", err, "offset", m.Offset)
		return nil
	}

	msgID, err := uuid.Parse(ev.GetEventId())
	if err != nil {
		logger.Error("payment result invalid event id", "err", err, "event_id", ev.GetEventId())
		return nil
	}

	orderID, err := uuid.Parse(ev.GetOrderId())
	if err != nil {
		logger.Error("payment result invalid order id", "err", err, "order_id", ev.GetOrderId())
		return nil
	}

	newStatus := "CANCELLED"
	if ev.GetStatus() == eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_SUCCESS {
		newStatus = "FINISHED"
	}

	err = c.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertInboxCheck(ctx, pgtype.UUID{
			Bytes: msgID,
			Valid: true,
		})
		if err != nil {
			logger.Error("payment result inbox insert failed", "err", err, "event_id", ev.GetEventId())
			return err
		}
		if inserted == 0 {
			logger.Info("payment result already processed", "event_id", ev.GetEventId())
			return nil
		}

		if err := q.UpdateOrderStatusIfNew(ctx, db.UpdateOrderStatusIfNewParams{
			OrderID: pgtype.UUID{
				Bytes: orderID,
				Valid: true,
			},
			Status: newStatus,
		}); err != nil {
			logger.Error("payment result update order failed", "err", err, "order_id", ev.GetOrderId(), "status", newStatus)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("payment result handle message failed", "err", err, "order_id", ev.GetOrderId())
		return err
	}
	logger.Info("payment result handle message completed", "order_id", ev.GetOrderId(), "status", newStatus)
	return nil
}
