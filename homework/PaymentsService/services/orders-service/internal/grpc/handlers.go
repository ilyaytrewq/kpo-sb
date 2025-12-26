package grpc

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/google/uuid"
	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "github.com/ilyaytrewq/payments-service/gen/go/events/v1"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"

	"github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres"
)

type Handlers struct {
	ordersv1.UnimplementedOrdersServiceServer
	repo *postgres.Repo
}

func NewHandlers(repo *postgres.Repo) *Handlers {
	return &Handlers{repo: repo}
}

func (h *Handlers) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be > 0")
	}
	if req.GetDescription() == "" {
		return nil, status.Error(codes.InvalidArgument, "description is required")
	}

	var resp *ordersv1.CreateOrderResponse

	err := h.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		row, err := q.CreateOrder(ctx, db.CreateOrderParams{
			UserID:      req.GetUserId(),
			Amount:      req.GetAmount(),
			Description: req.GetDescription(),
		})
		if err != nil {
			return err
		}

		orderID := row.OrderID.String()

		ev := &eventsv1.PaymentRequested{
			EventId:    uuid.NewString(),
			OccurredAt: timestamppb.Now(),
			OrderId:    orderID,
			UserId:     req.GetUserId(),
			Amount:     req.GetAmount(),
		}

		payload, err := proto.Marshal(ev)
		if err != nil {
			return status.Error(codes.Internal, "failed to marshal event")
		}

		_, err = q.InsertOutbox(ctx, db.InsertOutboxParams{
			Topic:    "payments.payment_requested.v1",
			KafkaKey: orderID,
			Payload:  payload,
		})
		if err != nil {
			return err
		}

		resp = &ordersv1.CreateOrderResponse{
			Order: &ordersv1.Order{
				OrderId:     orderID,
				UserId:      row.UserID,
				Amount:      row.Amount,
				Description: row.Description,
				Status:      ordersv1.OrderStatus_ORDER_STATUS_NEW,
				CreatedAt:   timestamppb.New(row.CreatedAt.Time),
			},
		}
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create order")
	}
	return resp, nil
}

func (h *Handlers) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (*ordersv1.ListOrdersResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int32(50)
	if req.GetLimit() > 0 {
		limit = req.GetLimit()
	}
	offset := int32(0)
	if req.GetPageToken() != "" {
		n, err := decodeOffset(req.GetPageToken())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid page_token")
		}
		offset = n
	}

	rows, err := h.repo.Q().ListOrders(ctx, db.ListOrdersParams{
		UserID: req.GetUserId(),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list orders")
	}

	out := make([]*ordersv1.Order, 0, len(rows))
	for _, r := range rows {
		out = append(out, &ordersv1.Order{
			OrderId:     r.OrderID.String(),
			UserId:      r.UserID,
			Amount:      r.Amount,
			Description: r.Description,
			Status:      mapOrderStatus(r.Status),
			CreatedAt:   timestamppb.New(r.CreatedAt.Time),
		})
	}

	nextToken := ""
	if len(rows) == int(limit) {
		nextToken = encodeOffset(offset + limit)
	}

	return &ordersv1.ListOrdersResponse{
		Orders:        out,
		NextPageToken: nextToken,
	}, nil
}

func (h *Handlers) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (*ordersv1.GetOrderResponse, error) {
	if req.GetUserId() == "" || req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and order_id are required")
	}

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order_id")
	}
	r, err := h.repo.Q().GetOrder(ctx, db.GetOrderParams{
		OrderID: pgtype.UUID{
			Bytes: oid,
			Valid: true,
		},
		UserID: req.GetUserId(),
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, "order not found")
	}

	return &ordersv1.GetOrderResponse{
		Order: &ordersv1.Order{
			OrderId:     r.OrderID.String(),
			UserId:      r.UserID,
			Amount:      r.Amount,
			Description: r.Description,
			Status:      mapOrderStatus(r.Status),
			CreatedAt:   timestamppb.New(r.CreatedAt.Time),
		},
	}, nil
}

func mapOrderStatus(s string) ordersv1.OrderStatus {
	switch s {
	case "NEW":
		return ordersv1.OrderStatus_ORDER_STATUS_NEW
	case "FINISHED":
		return ordersv1.OrderStatus_ORDER_STATUS_FINISHED
	case "CANCELLED":
		return ordersv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return ordersv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func encodeOffset(n int32) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(n))))
}

func decodeOffset(s string) (int32, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return int32(n), nil
}
