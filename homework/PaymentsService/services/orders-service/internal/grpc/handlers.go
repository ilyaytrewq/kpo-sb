package grpc

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ilyaytrewq/payments-service/order-service/internal/cache"
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
	repo  *postgres.Repo
	cache *cache.OrderCache
}

var logger = slog.Default().With("service", "orders-service", "component", "grpc")

func NewHandlers(repo *postgres.Repo, cache *cache.OrderCache) *Handlers {
	logger.Info("handlers initialized")
	return &Handlers{repo: repo, cache: cache}
}

func (h *Handlers) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (resp *ordersv1.CreateOrderResponse, err error) {
	start := time.Now()
	logger.Info("create order start", "user_id", req.GetUserId(), "amount", req.GetAmount(), "has_idempotency_key", req.GetIdempotencyKey() != "")
	defer func() {
		if err != nil {
			logger.Error("create order failed", "err", err, "duration", time.Since(start))
			return
		}
		orderID := ""
		if resp != nil && resp.Order != nil {
			orderID = resp.Order.OrderId
		}
		logger.Info("create order completed", "order_id", orderID, "duration", time.Since(start))
	}()

	if req.GetUserId() == "" {
		err = status.Error(codes.InvalidArgument, "user_id is required")
		logger.Error("create order validation failed", "err", err)
		return nil, err
	}
	if req.GetAmount() <= 0 {
		err = status.Error(codes.InvalidArgument, "amount must be > 0")
		logger.Error("create order validation failed", "err", err)
		return nil, err
	}
	if req.GetDescription() == "" {
		err = status.Error(codes.InvalidArgument, "description is required")
		logger.Error("create order validation failed", "err", err)
		return nil, err
	}

	err = h.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		idemKey := req.GetIdempotencyKey()
		var (
			orderID     string
			userID      string
			amount      int64
			description string
			statusText  string
			createdAt   time.Time
		)

		if idemKey == "" {
			row, err := q.CreateOrder(ctx, db.CreateOrderParams{
				UserID:      req.GetUserId(),
				Amount:      req.GetAmount(),
				Description: req.GetDescription(),
			})
			if err != nil {
				logger.Error("failed to create order", "err", err)
				return err
			}
			orderID = row.OrderID.String()
			userID = row.UserID
			amount = row.Amount
			description = row.Description
			statusText = row.Status
			createdAt = row.CreatedAt.Time
		} else {
			row, err := q.CreateOrderIdempotent(ctx, db.CreateOrderIdempotentParams{
				UserID:      req.GetUserId(),
				Amount:      req.GetAmount(),
				Description: req.GetDescription(),
				IdempotencyKey: pgtype.Text{
					String: idemKey,
					Valid:  true,
				},
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					existing, err := q.GetOrderByIdempotency(ctx, db.GetOrderByIdempotencyParams{
						UserID: req.GetUserId(),
						IdempotencyKey: pgtype.Text{
							String: idemKey,
							Valid:  true,
						},
					})
					if err != nil {
						logger.Error("failed to load order by idempotency key", "err", err)
						return err
					}
					if existing.Amount != req.GetAmount() || existing.Description != req.GetDescription() {
						err = status.Error(codes.FailedPrecondition, "idempotency key reuse with different parameters")
						logger.Error("idempotency key reuse with different parameters", "err", err)
						return err
					}
					resp = &ordersv1.CreateOrderResponse{
						Order: &ordersv1.Order{
							OrderId:     existing.OrderID.String(),
							UserId:      existing.UserID,
							Amount:      existing.Amount,
							Description: existing.Description,
							Status:      mapOrderStatus(existing.Status),
							CreatedAt:   timestamppb.New(existing.CreatedAt.Time),
						},
					}
					return nil
				}
				logger.Error("failed to create order with idempotency key", "err", err)
				return err
			}
			orderID = row.OrderID.String()
			userID = row.UserID
			amount = row.Amount
			description = row.Description
			statusText = row.Status
			createdAt = row.CreatedAt.Time
		}

		ev := &eventsv1.PaymentRequested{
			EventId:    uuid.NewString(),
			OccurredAt: timestamppb.Now(),
			OrderId:    orderID,
			UserId:     req.GetUserId(),
			Amount:     req.GetAmount(),
		}

		payload, err := proto.Marshal(ev)
		if err != nil {
			err = status.Error(codes.Internal, "failed to marshal event")
			logger.Error("failed to marshal payment requested event", "err", err)
			return err
		}

		_, err = q.InsertOutbox(ctx, db.InsertOutboxParams{
			Topic:    "payments.payment_requested.v1",
			KafkaKey: orderID,
			Payload:  payload,
		})
		if err != nil {
			logger.Error("failed to insert outbox event", "err", err)
			return err
		}

		resp = &ordersv1.CreateOrderResponse{
			Order: &ordersv1.Order{
				OrderId:     orderID,
				UserId:      userID,
				Amount:      amount,
				Description: description,
				Status:      mapOrderStatus(statusText),
				CreatedAt:   timestamppb.New(createdAt),
			},
		}
		return nil
	})

	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = st.Err()
			return nil, err
		}
		err = status.Error(codes.Internal, "failed to create order")
		return nil, err
	}
	return resp, nil
}

func (h *Handlers) ListOrders(ctx context.Context, req *ordersv1.ListOrdersRequest) (resp *ordersv1.ListOrdersResponse, err error) {
	start := time.Now()
	logger.Info("list orders start", "user_id", req.GetUserId(), "limit", req.GetLimit(), "page_token", req.GetPageToken() != "")
	defer func() {
		if err != nil {
			logger.Error("list orders failed", "err", err, "duration", time.Since(start))
			return
		}
		count := 0
		if resp != nil {
			count = len(resp.Orders)
		}
		logger.Info("list orders completed", "orders_count", count, "duration", time.Since(start))
	}()

	if req.GetUserId() == "" {
		err = status.Error(codes.InvalidArgument, "user_id is required")
		logger.Error("list orders validation failed", "err", err)
		return nil, err
	}

	limit := int32(50)
	if req.GetLimit() > 0 {
		limit = req.GetLimit()
	}
	offset := int32(0)
	if req.GetPageToken() != "" {
		n, err := decodeOffset(req.GetPageToken())
		if err != nil {
			err = status.Error(codes.InvalidArgument, "invalid page_token")
			logger.Error("list orders invalid page token", "err", err)
			return nil, err
		}
		offset = n
	}

	rows, err := h.repo.Q().ListOrders(ctx, db.ListOrdersParams{
		UserID: req.GetUserId(),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		err = status.Error(codes.Internal, "failed to list orders")
		logger.Error("list orders query failed", "err", err)
		return nil, err
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

	resp = &ordersv1.ListOrdersResponse{
		Orders:        out,
		NextPageToken: nextToken,
	}
	return resp, nil
}

func (h *Handlers) GetOrder(ctx context.Context, req *ordersv1.GetOrderRequest) (resp *ordersv1.GetOrderResponse, err error) {
	start := time.Now()
	logger.Info("get order start", "user_id", req.GetUserId(), "order_id", req.GetOrderId())
	defer func() {
		if err != nil {
			logger.Error("get order failed", "err", err, "duration", time.Since(start))
			return
		}
		logger.Info("get order completed", "duration", time.Since(start))
	}()

	if req.GetUserId() == "" || req.GetOrderId() == "" {
		err = status.Error(codes.InvalidArgument, "user_id and order_id are required")
		logger.Error("get order validation failed", "err", err)
		return nil, err
	}

	oid, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		err = status.Error(codes.InvalidArgument, "invalid order_id")
		logger.Error("get order invalid order id", "err", err)
		return nil, err
	}

	if cached, err := h.cache.Get(ctx, req.GetOrderId()); err == nil && cached != nil {
		logger.Info("get order cache hit", "order_id", req.GetOrderId())
		if cached.UserID == req.GetUserId() {
			resp = &ordersv1.GetOrderResponse{
				Order: &ordersv1.Order{
					OrderId:     cached.OrderID,
					UserId:      cached.UserID,
					Amount:      cached.Amount,
					Description: cached.Description,
					Status:      mapOrderStatus(cached.Status),
					CreatedAt:   timestamppb.New(cached.CreatedAt),
				},
			}
			return resp, nil
		}
	}
	logger.Info("get order cache miss", "order_id", req.GetOrderId())

	r, err := h.repo.Q().GetOrder(ctx, db.GetOrderParams{
		OrderID: pgtype.UUID{
			Bytes: oid,
			Valid: true,
		},
		UserID: req.GetUserId(),
	})
	if err != nil {
		err = status.Error(codes.NotFound, "order not found")
		logger.Error("get order query failed", "err", err)
		return nil, err
	}

	if h.cache != nil {
		if err := h.cache.Set(ctx, cache.Order{
			OrderID:     r.OrderID.String(),
			UserID:      r.UserID,
			Amount:      r.Amount,
			Description: r.Description,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Time,
		}); err != nil {
			logger.Error("failed to set order cache", "err", err, "order_id", r.OrderID.String())
		}
	}

	resp = &ordersv1.GetOrderResponse{
		Order: &ordersv1.Order{
			OrderId:     r.OrderID.String(),
			UserId:      r.UserID,
			Amount:      r.Amount,
			Description: r.Description,
			Status:      mapOrderStatus(r.Status),
			CreatedAt:   timestamppb.New(r.CreatedAt.Time),
		},
	}
	return resp, nil
}

func mapOrderStatus(s string) ordersv1.OrderStatus {
	logger.Info("map order status", "status", s)
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
	start := time.Now()
	logger.Info("encode offset start", "offset", n)
	encoded := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(int(n))))
	logger.Info("encode offset completed", "duration", time.Since(start))
	return encoded
}

func decodeOffset(s string) (int32, error) {
	start := time.Now()
	logger.Info("decode offset start", "has_value", s != "")
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		logger.Error("decode offset failed", "err", err, "duration", time.Since(start))
		return 0, err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		logger.Error("decode offset failed", "err", err, "duration", time.Since(start))
		return 0, err
	}
	logger.Info("decode offset completed", "offset", n, "duration", time.Since(start))
	return int32(n), nil
}
