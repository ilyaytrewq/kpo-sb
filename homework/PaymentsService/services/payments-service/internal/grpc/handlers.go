package grpc

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	paymentsv1 "github.com/ilyaytrewq/payments-service/gen/go/payments/v1"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/cache"
	"github.com/ilyaytrewq/payments-service/payments-service/internal/repo/postgres"
	db "github.com/ilyaytrewq/payments-service/payments-service/internal/repo/postgres/db"
)

type Handlers struct {
	paymentsv1.UnimplementedPaymentsServiceServer
	repo  *postgres.Repo
	cache *cache.BalanceCache
}

func NewHandlers(repo *postgres.Repo, cache *cache.BalanceCache) *Handlers {
	return &Handlers{repo: repo, cache: cache}
}

func (h *Handlers) CreateAccount(ctx context.Context, req *paymentsv1.CreateAccountRequest) (*paymentsv1.CreateAccountResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	account, err := h.repo.Q().CreateAccount(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.AlreadyExists, "account already exists")
		}
		return nil, status.Error(codes.Internal, "failed to create account")
	}

	return &paymentsv1.CreateAccountResponse{
		Account: &paymentsv1.Account{
			UserId:  account.UserID,
			Balance: account.Balance,
		},
	}, nil
}

func (h *Handlers) TopUp(ctx context.Context, req *paymentsv1.TopUpRequest) (*paymentsv1.TopUpResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be > 0")
	}

	account, err := h.repo.Q().TopUp(ctx, db.TopUpParams{
		UserID:  userID,
		Balance: req.GetAmount(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to top up")
	}

	if h.cache != nil {
		_ = h.cache.Set(ctx, cache.Balance{
			UserID:  account.UserID,
			Balance: account.Balance,
		})
	}

	return &paymentsv1.TopUpResponse{
		Account: &paymentsv1.Account{
			UserId:  account.UserID,
			Balance: account.Balance,
		},
	}, nil
}

func (h *Handlers) GetBalance(ctx context.Context, req *paymentsv1.GetBalanceRequest) (*paymentsv1.GetBalanceResponse, error) {
	userID := req.GetUserId()
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if cached, err := h.cache.Get(ctx, userID); err == nil && cached != nil {
		return &paymentsv1.GetBalanceResponse{
			Balance: cached.Balance,
		}, nil
	}

	balance, err := h.repo.Q().GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to get balance")
	}

	if h.cache != nil {
		_ = h.cache.Set(ctx, cache.Balance{
			UserID:  userID,
			Balance: balance,
		})
	}

	return &paymentsv1.GetBalanceResponse{
		Balance: balance,
	}, nil
}
