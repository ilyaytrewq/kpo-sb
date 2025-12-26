package kafka

import (
	"context"
	"log"

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
	return &PaymentRequestedConsumer{repo: repo, reader: r, resultTopic: resultTopic}
}

func (c *PaymentRequestedConsumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := c.handleMessage(ctx, m); err != nil {
			log.Println("payment_requested handle error:", err)
			// offset НЕ коммитим => Kafka доставит снова
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

func (c *PaymentRequestedConsumer) handleMessage(ctx context.Context, m kafka.Message) error {
	var ev eventsv1.PaymentRequested
	if err := proto.Unmarshal(m.Value, &ev); err != nil {
		// плохое сообщение лучше “проглотить” и закоммитить
		return nil
	}

	msgID, err := uuid.Parse(ev.GetEventId())
	if err != nil {
		return nil
	}

	orderID, err := uuid.Parse(ev.GetOrderId())
	if err != nil {
		return nil
	}

	if ev.GetUserId() == "" || ev.GetAmount() <= 0 {
		return nil
	}

	return c.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertInboxCheck(ctx, db.InsertInboxCheckParams{
			MessageID: pgtype.UUID{Bytes: msgID, Valid: true},
			OrderID:   pgtype.UUID{Bytes: orderID, Valid: true},
		})
		if err != nil {
			return err
		}
		if inserted == 0 {
			return nil
		}

		res, err := q.TryDeductOnce(ctx, db.TryDeductOnceParams{
			OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
			UserID:  ev.GetUserId(),
			Balance: ev.GetAmount(),
		})
		if err != nil {
			return err
		}

		status := eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_FAIL_INTERNAL
		reason := ""
		if res.OpInserted == 1 {
			status = eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_SUCCESS
		} else {
			exists, err := q.AccountExists(ctx, ev.GetUserId())
			if err != nil {
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
			return err
		}

		if _, err := q.InsertOutbox(ctx, db.InsertOutboxParams{
			Topic:    c.resultTopic,
			KafkaKey: orderID.String(),
			Payload:  payload,
		}); err != nil {
			return err
		}

		return nil
	})
}
