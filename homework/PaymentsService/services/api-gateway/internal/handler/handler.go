package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
	paymentsv1 "github.com/ilyaytrewq/payments-service/gen/go/payments/v1"
	gateway "github.com/ilyaytrewq/payments-service/gen/openapi/gateway"
)

const requestTimeout = 5 * time.Second

type Handler struct {
	orders   ordersv1.OrdersServiceClient
	payments paymentsv1.PaymentsServiceClient
}

var logger = slog.Default().With("service", "api-gateway", "component", "handler")

func New(orders ordersv1.OrdersServiceClient, payments paymentsv1.PaymentsServiceClient) *Handler {
	logger.Info("handler initialized")
	return &Handler{orders: orders, payments: payments}
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request, params gateway.ListOrdersParams) {
	start := time.Now()
	userID, _ := resolveUserID(params.XUserId)
	logger.Info("list orders start", "user_id", userID)

	req := &ordersv1.ListOrdersRequest{UserId: userID}
	if params.Limit != nil {
		req.Limit = int32(*params.Limit)
	}
	if params.PageToken != nil {
		req.PageToken = string(*params.PageToken)
	}

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.orders.ListOrders(ctx, req)
	if err != nil {
		logger.Error("list orders grpc failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	out := make([]gateway.Order, 0, len(resp.GetOrders()))
	for _, order := range resp.GetOrders() {
		if mapped := mapOrder(order); mapped != nil {
			out = append(out, *mapped)
		}
	}

	writeJSON(w, http.StatusOK, gateway.ListOrdersResponse{
		UserId: userID,
		Orders: out,
	})
	logger.Info("list orders completed", "user_id", userID, "orders_count", len(out), "duration", time.Since(start))
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request, params gateway.CreateOrderParams) {
	start := time.Now()
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)
	logger.Info("create order start", "user_id", userID, "has_idempotency_key", idempotencyKey != "")

	var body gateway.CreateOrderRequest
	if err := decodeJSON(r, &body); err != nil {
		logger.Error("create order decode failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeError(w, userID, http.StatusBadRequest, err.Error())
		return
	}
	if body.Amount <= 0 || strings.TrimSpace(body.Description) == "" {
		logger.Error("create order validation failed", "user_id", userID, "amount", body.Amount, "duration", time.Since(start))
		writeError(w, userID, http.StatusBadRequest, "amount must be > 0 and description is required")
		return
	}

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.orders.CreateOrder(ctx, &ordersv1.CreateOrderRequest{
		UserId:         userID,
		Amount:         body.Amount,
		Description:    body.Description,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		logger.Error("create order grpc failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	mapped := mapOrder(resp.GetOrder())
	if mapped == nil {
		logger.Error("create order mapping failed", "user_id", userID, "duration", time.Since(start))
		writeError(w, userID, http.StatusInternalServerError, "empty order response")
		return
	}

	writeJSON(w, http.StatusCreated, gateway.CreateOrderResponse{
		UserId: userID,
		Order:  *mapped,
	})
	logger.Info("create order completed", "user_id", userID, "order_id", mapped.OrderId, "duration", time.Since(start))
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request, orderId gateway.OrderIdPath, params gateway.GetOrderParams) {
	start := time.Now()
	userID, _ := resolveUserID(params.XUserId)
	logger.Info("get order start", "user_id", userID, "order_id", orderId)

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.orders.GetOrder(ctx, &ordersv1.GetOrderRequest{
		UserId:  userID,
		OrderId: string(orderId),
	})
	if err != nil {
		logger.Error("get order grpc failed", "err", err, "user_id", userID, "order_id", orderId, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	mapped := mapOrder(resp.GetOrder())
	if mapped == nil {
		logger.Error("get order mapping failed", "user_id", userID, "order_id", orderId, "duration", time.Since(start))
		writeError(w, userID, http.StatusInternalServerError, "empty order response")
		return
	}

	writeJSON(w, http.StatusOK, gateway.GetOrderResponse{
		UserId: userID,
		Order:  *mapped,
	})
	logger.Info("get order completed", "user_id", userID, "order_id", mapped.OrderId, "duration", time.Since(start))
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request, params gateway.CreateAccountParams) {
	start := time.Now()
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)
	logger.Info("create account start", "user_id", userID, "has_idempotency_key", idempotencyKey != "")

	if err := decodeOptionalJSON(r); err != nil {
		logger.Error("create account decode failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeError(w, userID, http.StatusBadRequest, err.Error())
		return
	}

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.payments.CreateAccount(ctx, &paymentsv1.CreateAccountRequest{
		UserId:         userID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		logger.Error("create account grpc failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusCreated, gateway.CreateAccountResponse{
		UserId:  userID,
		Balance: resp.GetAccount().GetBalance(),
	})
	logger.Info("create account completed", "user_id", userID, "duration", time.Since(start))
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request, params gateway.GetBalanceParams) {
	start := time.Now()
	userID := string(params.XUserId)
	if strings.TrimSpace(userID) == "" {
		logger.Error("get balance validation failed", "duration", time.Since(start))
		writeError(w, "", http.StatusBadRequest, "X-User-Id header is required")
		return
	}
	logger.Info("get balance start", "user_id", userID)

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.payments.GetBalance(ctx, &paymentsv1.GetBalanceRequest{UserId: userID})
	if err != nil {
		logger.Error("get balance grpc failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusOK, gateway.GetBalanceResponse{
		UserId:  userID,
		Balance: resp.GetBalance(),
	})
	logger.Info("get balance completed", "user_id", userID, "duration", time.Since(start))
}

func (h *Handler) TopUpAccount(w http.ResponseWriter, r *http.Request, params gateway.TopUpAccountParams) {
	start := time.Now()
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)
	logger.Info("top up start", "user_id", userID, "has_idempotency_key", idempotencyKey != "")

	var body gateway.TopUpAccountRequest
	if err := decodeJSON(r, &body); err != nil {
		logger.Error("top up decode failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeError(w, userID, http.StatusBadRequest, err.Error())
		return
	}
	if body.Amount <= 0 {
		logger.Error("top up validation failed", "user_id", userID, "amount", body.Amount, "duration", time.Since(start))
		writeError(w, userID, http.StatusBadRequest, "amount must be > 0")
		return
	}

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.payments.TopUp(ctx, &paymentsv1.TopUpRequest{
		UserId:         userID,
		Amount:         body.Amount,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		logger.Error("top up grpc failed", "err", err, "user_id", userID, "duration", time.Since(start))
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusOK, gateway.TopUpAccountResponse{
		UserId:  userID,
		Balance: resp.GetAccount().GetBalance(),
	})
	logger.Info("top up completed", "user_id", userID, "duration", time.Since(start))
}

func mapOrder(order *ordersv1.Order) *gateway.Order {
	logger.Info("map order start", "has_order", order != nil)
	if order == nil {
		logger.Error("map order failed (nil order)")
		return nil
	}

	var createdAt *time.Time
	if order.GetCreatedAt() != nil {
		t := order.GetCreatedAt().AsTime()
		createdAt = &t
	}

	mapped := &gateway.Order{
		OrderId:     order.GetOrderId(),
		UserId:      order.GetUserId(),
		Amount:      order.GetAmount(),
		Description: order.GetDescription(),
		Status:      mapOrderStatus(order.GetStatus()),
		CreatedAt:   createdAt,
	}
	logger.Info("map order completed", "order_id", mapped.OrderId)
	return mapped
}

func mapOrderStatus(status ordersv1.OrderStatus) gateway.OrderStatus {
	logger.Info("map order status", "status", status.String())
	switch status {
	case ordersv1.OrderStatus_ORDER_STATUS_FINISHED:
		return gateway.OrderStatus("FINISHED")
	case ordersv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return gateway.OrderStatus("CANCELLED")
	case ordersv1.OrderStatus_ORDER_STATUS_NEW:
		return gateway.OrderStatus("NEW")
	default:
		return gateway.OrderStatus("NEW")
	}
}

func resolveUserID(header *gateway.UserIdHeader) (string, bool) {
	logger.Info("resolve user id start", "header_present", header != nil)
	if header != nil && strings.TrimSpace(string(*header)) != "" {
		return string(*header), false
	}
	newID := uuid.NewString()
	logger.Info("generated user id", "user_id", newID)
	return newID, true
}

func getHeader(header *gateway.IdempotencyKeyHeader) string {
	if header == nil {
		return ""
	}
	logger.Info("idempotency key header resolved", "has_value", strings.TrimSpace(string(*header)) != "")
	return string(*header)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	logger.Info("write json response", "status", status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, userID string, statusCode int, message string) {
	logger.Error("write error response", "user_id", userID, "status", statusCode, "message", message)
	resp := gateway.ErrorResponse{Error: message}
	if userID != "" {
		resp.UserId = &userID
	}
	writeJSON(w, statusCode, resp)
}

func WriteBadRequest(w http.ResponseWriter, userID string, err error) {
	message := "bad request"
	if err != nil {
		message = err.Error()
	}
	logger.Error("write bad request", "user_id", userID, "message", message)
	writeError(w, userID, http.StatusBadRequest, message)
}

func writeGRPCError(w http.ResponseWriter, userID string, err error) {
	st, ok := status.FromError(err)
	if !ok {
		logger.Error("write grpc error failed to parse status", "user_id", userID, "err", err)
		writeError(w, userID, http.StatusInternalServerError, "internal error")
		return
	}
	logger.Error("write grpc error", "user_id", userID, "grpc_code", st.Code().String(), "message", st.Message())
	writeError(w, userID, grpcCodeToStatus(st.Code()), st.Message())
}

func grpcCodeToStatus(code codes.Code) int {
	logger.Info("grpc code to status", "grpc_code", code.String())
	switch code {
	case codes.InvalidArgument, codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func decodeJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		logger.Error("decode json failed (empty body)")
		return fmt.Errorf("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		logger.Error("decode json failed", "err", err)
		return err
	}
	logger.Info("decode json completed")
	return nil
}

func decodeOptionalJSON(r *http.Request) error {
	if r.Body == nil || r.ContentLength == 0 {
		logger.Info("decode optional json skipped (empty body)")
		return nil
	}
	var payload map[string]interface{}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		logger.Error("decode optional json failed", "err", err)
		return err
	}
	if len(payload) > 0 {
		logger.Error("decode optional json failed (non-empty body)")
		return fmt.Errorf("request body must be empty")
	}
	logger.Info("decode optional json completed")
	return nil
}

func withTimeout(r *http.Request) (context.Context, func()) {
	logger.Info("with timeout", "timeout", requestTimeout.String())
	return context.WithTimeout(r.Context(), requestTimeout)
}

var _ gateway.ServerInterface = (*Handler)(nil)
