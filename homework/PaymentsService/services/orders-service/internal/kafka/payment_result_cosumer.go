package kafka

import (
	"context"
	"log"

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
	return &PaymentResultConsumer{repo: repo, reader: r}
}

func (c *PaymentResultConsumer) Run(ctx context.Context) error {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := c.handleMessage(ctx, m); err != nil {
			log.Println("payment_result handle error:", err)
			// offset НЕ коммитим => Kafka доставит снова
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			return err
		}
	}
}

func (c *PaymentResultConsumer) handleMessage(ctx context.Context, m kafka.Message) error {
	var ev eventsv1.PaymentResult
	if err := proto.Unmarshal(m.Value, &ev); err != nil {
		// плохое сообщение лучше “проглотить” и закоммитить, иначе будет бесконечный цикл
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

	newStatus := "CANCELLED"
	if ev.GetStatus() == eventsv1.PaymentResultStatus_PAYMENT_RESULT_STATUS_SUCCESS {
		newStatus = "FINISHED"
	}

	return c.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertInboxCheck(ctx, pgtype.UUID{
			Bytes: msgID,
			Valid: true,
		})
		if err != nil {
			return err
		}
		if inserted == 0 {
			return nil
		}

		if err := q.UpdateOrderStatusIfNew(ctx, db.UpdateOrderStatusIfNewParams{
			OrderID: pgtype.UUID{
				Bytes: orderID,
				Valid: true,
			},
			Status: newStatus,
		}); err != nil {
			return err
		}

		return nil
	})

}
