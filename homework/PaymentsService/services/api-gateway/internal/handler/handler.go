package handler

import (
	"context"
	"encoding/json"
	"fmt"
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

func New(orders ordersv1.OrdersServiceClient, payments paymentsv1.PaymentsServiceClient) *Handler {
	return &Handler{orders: orders, payments: payments}
}

func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request, params gateway.ListOrdersParams) {
	userID, _ := resolveUserID(params.XUserId)

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
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request, params gateway.CreateOrderParams) {
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)

	var body gateway.CreateOrderRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, userID, http.StatusBadRequest, err.Error())
		return
	}
	if body.Amount <= 0 || strings.TrimSpace(body.Description) == "" {
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
		writeGRPCError(w, userID, err)
		return
	}

	mapped := mapOrder(resp.GetOrder())
	if mapped == nil {
		writeError(w, userID, http.StatusInternalServerError, "empty order response")
		return
	}

	writeJSON(w, http.StatusCreated, gateway.CreateOrderResponse{
		UserId: userID,
		Order:  *mapped,
	})
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request, orderId gateway.OrderIdPath, params gateway.GetOrderParams) {
	userID, _ := resolveUserID(params.XUserId)

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.orders.GetOrder(ctx, &ordersv1.GetOrderRequest{
		UserId:  userID,
		OrderId: string(orderId),
	})
	if err != nil {
		writeGRPCError(w, userID, err)
		return
	}

	mapped := mapOrder(resp.GetOrder())
	if mapped == nil {
		writeError(w, userID, http.StatusInternalServerError, "empty order response")
		return
	}

	writeJSON(w, http.StatusOK, gateway.GetOrderResponse{
		UserId: userID,
		Order:  *mapped,
	})
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request, params gateway.CreateAccountParams) {
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)

	if err := decodeOptionalJSON(r); err != nil {
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
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusCreated, gateway.CreateAccountResponse{
		UserId:  userID,
		Balance: resp.GetAccount().GetBalance(),
	})
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request, params gateway.GetBalanceParams) {
	userID := string(params.XUserId)
	if strings.TrimSpace(userID) == "" {
		writeError(w, "", http.StatusBadRequest, "X-User-Id header is required")
		return
	}

	ctx, cancel := withTimeout(r)
	defer cancel()

	resp, err := h.payments.GetBalance(ctx, &paymentsv1.GetBalanceRequest{UserId: userID})
	if err != nil {
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusOK, gateway.GetBalanceResponse{
		UserId:  userID,
		Balance: resp.GetBalance(),
	})
}

func (h *Handler) TopUpAccount(w http.ResponseWriter, r *http.Request, params gateway.TopUpAccountParams) {
	userID, _ := resolveUserID(params.XUserId)
	idempotencyKey := getHeader(params.IdempotencyKey)

	var body gateway.TopUpAccountRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, userID, http.StatusBadRequest, err.Error())
		return
	}
	if body.Amount <= 0 {
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
		writeGRPCError(w, userID, err)
		return
	}

	writeJSON(w, http.StatusOK, gateway.TopUpAccountResponse{
		UserId:  userID,
		Balance: resp.GetAccount().GetBalance(),
	})
}

func mapOrder(order *ordersv1.Order) *gateway.Order {
	if order == nil {
		return nil
	}

	var createdAt *time.Time
	if order.GetCreatedAt() != nil {
		t := order.GetCreatedAt().AsTime()
		createdAt = &t
	}

	return &gateway.Order{
		OrderId:     order.GetOrderId(),
		UserId:      order.GetUserId(),
		Amount:      order.GetAmount(),
		Description: order.GetDescription(),
		Status:      mapOrderStatus(order.GetStatus()),
		CreatedAt:   createdAt,
	}
}

func mapOrderStatus(status ordersv1.OrderStatus) gateway.OrderStatus {
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
	if header != nil && strings.TrimSpace(string(*header)) != "" {
		return string(*header), false
	}
	return uuid.NewString(), true
}

func getHeader(header *gateway.IdempotencyKeyHeader) string {
	if header == nil {
		return ""
	}
	return string(*header)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, userID string, statusCode int, message string) {
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
	writeError(w, userID, http.StatusBadRequest, message)
}

func writeGRPCError(w http.ResponseWriter, userID string, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, userID, http.StatusInternalServerError, "internal error")
		return
	}
	writeError(w, userID, grpcCodeToStatus(st.Code()), st.Message())
}

func grpcCodeToStatus(code codes.Code) int {
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
		return fmt.Errorf("request body is required")
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}

func decodeOptionalJSON(r *http.Request) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	var payload map[string]interface{}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		return err
	}
	if len(payload) > 0 {
		return fmt.Errorf("request body must be empty")
	}
	return nil
}

func withTimeout(r *http.Request) (context.Context, func()) {
	return context.WithTimeout(r.Context(), requestTimeout)
}

var _ gateway.ServerInterface = (*Handler)(nil)
