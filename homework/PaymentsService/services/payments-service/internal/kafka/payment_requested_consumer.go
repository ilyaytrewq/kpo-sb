package kafka

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "github.com/ilyaytrewq/payments-service/gen/go/events/v1"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/repo/postgres"
	db "github.com/ilyaytrewq/payments-service/payments-service/internal/repo/postgres/db"
)

type PaymentRequestedConsumer struct {
	repo        *postgres.Repo
	reader      *kafka.Reader
	resultTopic string
}

func NewPaymentRequestedConsumer(repo *postgres.Repo, r *kafka.Reader, resultTopic string) *PaymentRequestedConsumer {
	slog.Default().With("service", "payments-service", "component", "kafka").Info("payment requested consumer initialized", "result_topic", resultTopic)
	return &PaymentRequestedConsumer{repo: repo, reader: r, resultTopic: resultTopic}
}

func (c *PaymentRequestedConsumer) Run(ctx context.Context) error {
	logger := slog.Default().With("service", "payments-service", "component", "kafka")
	logger.Info("payment requested consumer run start")
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Info("payment requested consumer context done")
				return nil
			}
			logger.Error("payment requested fetch failed", "err", err)
			return err
		}

		if err := c.handleMessage(ctx, m); err != nil {
			logger.Error("payment requested handle error", "err", err, "offset", m.Offset)
			// offset НЕ коммитим => Kafka доставит снова
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			logger.Error("payment requested commit failed", "err", err, "offset", m.Offset)
			return err
		}
		logger.Info("payment requested message committed", "offset", m.Offset)
	}
}

func (c *PaymentRequestedConsumer) handleMessage(ctx context.Context, m kafka.Message) error {
	logger := slog.Default().With("service", "payments-service", "component", "kafka")
	logger.Info("payment requested handle message start", "offset", m.Offset)
	var ev eventsv1.PaymentRequested
	if err := proto.Unmarshal(m.Value, &ev); err != nil {
		// плохое сообщение лучше “проглотить” и закоммитить
		logger.Error("payment requested unmarshal failed", "err", err, "offset", m.Offset)
		return nil
	}

	msgID, err := uuid.Parse(ev.GetEventId())
	if err != nil {
		logger.Error("payment requested invalid event id", "err", err, "event_id", ev.GetEventId())
		return nil
	}

	orderID, err := uuid.Parse(ev.GetOrderId())
	if err != nil {
		logger.Error("payment requested invalid order id", "err", err, "order_id", ev.GetOrderId())
		return nil
	}

	if ev.GetUserId() == "" || ev.GetAmount() <= 0 {
		logger.Error("payment requested invalid payload", "user_id", ev.GetUserId(), "amount", ev.GetAmount())
		return nil
	}

	err = c.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertInboxCheck(ctx, db.InsertInboxCheckParams{
			MessageID: pgtype.UUID{Bytes: msgID, Valid: true},
			OrderID:   pgtype.UUID{Bytes: orderID, Valid: true},
		})
		if err != nil {
			logger.Error("payment requested inbox insert failed", "err", err)
			return err
		}
		if inserted == 0 {
			logger.Info("payment requested already processed", "event_id", ev.GetEventId())
			return nil
		}

		res, err := q.TryDeductOnce(ctx, db.TryDeductOnceParams{
			OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
			UserID:  ev.GetUserId(),
			Balance: ev.GetAmount(),
		})
		if err != nil {
			logger.Error("payment requested deduct failed", "err", err, "order_id", ev.GetOrderId())
			return err
		}

		status := eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_FAIL_INTERNAL
		reason := ""
		if res.OpInserted == 1 {
			status = eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_SUCCESS
		} else {
			exists, err := q.AccountExists(ctx, ev.GetUserId())
			if err != nil {
				logger.Error("payment requested account existence check failed", "err", err, "user_id", ev.GetUserId())
				return err
			}
			if !exists {
				status = eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_FAIL_NO_ACCOUNT
				reason = "account not found"
			} else {
				status = eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_FAIL_NOT_ENOUGH_FUNDS
				reason = "not enough funds"
			}
		}

		result := &eventsv1.PaymentResult{
			EventId:    uuid.NewString(),
			OccurredAt: timestamppb.Now(),
			OrderId:    orderID.String(),
			UserId:     ev.GetUserId(),
			Status:     status,
			Reason:     reason,
		}

		payload, err := proto.Marshal(result)
		if err != nil {
			logger.Error("payment result marshal failed", "err", err, "order_id", ev.GetOrderId())
			return err
		}

		if _, err := q.InsertOutbox(ctx, db.InsertOutboxParams{
			Topic:    c.resultTopic,
			KafkaKey: orderID.String(),
			Payload:  payload,
		}); err != nil {
			logger.Error("payment result outbox insert failed", "err", err, "order_id", ev.GetOrderId())
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("payment requested handle message failed", "err", err, "order_id", ev.GetOrderId())
		return err
	}
	logger.Info("payment requested handle message completed", "order_id", ev.GetOrderId())
	return nil
}
