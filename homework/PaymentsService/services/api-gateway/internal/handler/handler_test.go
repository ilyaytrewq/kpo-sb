package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	ordersv1 "github.com/ilyaytrewq/payments-service/gen/go/orders/v1"
	gateway "github.com/ilyaytrewq/payments-service/gen/openapi/gateway"
)

func TestMapOrderStatus(t *testing.T) {
	tests := []struct {
		name   string
		status ordersv1.OrderStatus
		want   gateway.OrderStatus
	}{
		{"finished", ordersv1.OrderStatus_ORDER_STATUS_FINISHED, gateway.OrderStatus("FINISHED")},
		{"cancelled", ordersv1.OrderStatus_ORDER_STATUS_CANCELLED, gateway.OrderStatus("CANCELLED")},
		{"new", ordersv1.OrderStatus_ORDER_STATUS_NEW, gateway.OrderStatus("NEW")},
		{"unknown", ordersv1.OrderStatus_ORDER_STATUS_UNSPECIFIED, gateway.OrderStatus("NEW")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapOrderStatus(tt.status); got != tt.want {
				t.Fatalf("mapOrderStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapOrder(t *testing.T) {
	now := time.Now().UTC()
	order := &ordersv1.Order{
		OrderId:     "o-1",
		UserId:      "u-1",
		Amount:      100,
		Description: "test",
		Status:      ordersv1.OrderStatus_ORDER_STATUS_FINISHED,
		CreatedAt:   timestamppb.New(now),
	}

	mapped := mapOrder(order)
	if mapped == nil {
		t.Fatal("mapOrder() returned nil")
	}
	if mapped.OrderId != "o-1" || mapped.UserId != "u-1" || mapped.Amount != 100 || mapped.Description != "test" {
		t.Fatalf("mapOrder() unexpected fields: %+v", mapped)
	}
	if mapped.Status != gateway.OrderStatus("FINISHED") {
		t.Fatalf("mapOrder() status = %q, want %q", mapped.Status, "FINISHED")
	}
	if mapped.CreatedAt == nil || !mapped.CreatedAt.Equal(now) {
		t.Fatalf("mapOrder() created_at = %v, want %v", mapped.CreatedAt, now)
	}
}

func TestMapOrderNil(t *testing.T) {
	if mapOrder(nil) != nil {
		t.Fatal("mapOrder(nil) should return nil")
	}
}

func TestResolveUserID(t *testing.T) {
	h := gateway.UserIdHeader("user-1")
	got, generated := resolveUserID(&h)
	if generated {
		t.Fatal("resolveUserID() generated id for provided header")
	}
	if got != "user-1" {
		t.Fatalf("resolveUserID() = %q, want %q", got, "user-1")
	}
}

func TestResolveUserIDGenerated(t *testing.T) {
	got, generated := resolveUserID(nil)
	if !generated {
		t.Fatal("resolveUserID() did not generate id for missing header")
	}
	if got == "" {
		t.Fatal("resolveUserID() generated empty id")
	}
	if _, err := uuid.Parse(got); err != nil {
		t.Fatalf("resolveUserID() generated invalid uuid: %v", err)
	}
}

func TestGetHeader(t *testing.T) {
	if got := getHeader(nil); got != "" {
		t.Fatalf("getHeader(nil) = %q, want empty string", got)
	}
	v := gateway.IdempotencyKeyHeader("key-1")
	if got := getHeader(&v); got != "key-1" {
		t.Fatalf("getHeader() = %q, want %q", got, "key-1")
	}
}

func TestGrpcCodeToStatus(t *testing.T) {
	tests := []struct {
		code codes.Code
		want int
	}{
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.FailedPrecondition, http.StatusBadRequest},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.Unauthenticated, http.StatusUnauthorized},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.Unavailable, http.StatusServiceUnavailable},
		{codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{codes.Internal, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		if got := grpcCodeToStatus(tt.code); got != tt.want {
			t.Fatalf("grpcCodeToStatus(%s) = %d, want %d", tt.code.String(), got, tt.want)
		}
	}
}

func TestDecodeJSON(t *testing.T) {
	t.Run("missing body", func(t *testing.T) {
		req := &http.Request{Body: nil}
		var payload map[string]interface{}
		if err := decodeJSON(req, &payload); err == nil {
			t.Fatal("decodeJSON() expected error for nil body")
		}
	})

	t.Run("valid body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"amount":10,"description":"ok"}`))
		var payload struct {
			Amount      int    `json:"amount"`
			Description string `json:"description"`
		}
		if err := decodeJSON(req, &payload); err != nil {
			t.Fatalf("decodeJSON() error: %v", err)
		}
		if payload.Amount != 10 || payload.Description != "ok" {
			t.Fatalf("decodeJSON() unexpected payload: %+v", payload)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"amount":10,"extra":1}`))
		var payload struct {
			Amount int `json:"amount"`
		}
		if err := decodeJSON(req, &payload); err == nil {
			t.Fatal("decodeJSON() expected error for unknown field")
		}
	})
}

func TestDecodeOptionalJSON(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		if err := decodeOptionalJSON(req); err != nil {
			t.Fatalf("decodeOptionalJSON() error: %v", err)
		}
	})

	t.Run("empty object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		if err := decodeOptionalJSON(req); err != nil {
			t.Fatalf("decodeOptionalJSON() error: %v", err)
		}
	})

	t.Run("non-empty object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"a":1}`))
		if err := decodeOptionalJSON(req); err == nil {
			t.Fatal("decodeOptionalJSON() expected error for non-empty body")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{`))
		if err := decodeOptionalJSON(req); err == nil {
			t.Fatal("decodeOptionalJSON() expected error for invalid json")
		}
	})
}
