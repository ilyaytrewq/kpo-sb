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

	idemKey := req.GetIdempotencyKey()
	var (
		accountUserID  string
		accountBalance int64
	)
	if idemKey == "" {
		account, err := h.repo.Q().CreateAccount(ctx, userID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, status.Error(codes.AlreadyExists, "account already exists")
			}
			return nil, status.Error(codes.Internal, "failed to create account")
		}
		accountUserID = account.UserID
		accountBalance = account.Balance
	} else {
		account, err := h.repo.Q().CreateAccountIdempotent(ctx, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create account")
		}
		accountUserID = account.UserID
		accountBalance = account.Balance
	}

	if h.cache != nil {
		_ = h.cache.Set(ctx, cache.Balance{
			UserID:  accountUserID,
			Balance: accountBalance,
		})
	}

	return &paymentsv1.CreateAccountResponse{
		Account: &paymentsv1.Account{
			UserId:  accountUserID,
			Balance: accountBalance,
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

	idemKey := req.GetIdempotencyKey()
	if idemKey == "" {
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

	var (
		balance     int64
		updateCache bool
	)
	err := h.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertTopupIdempotency(ctx, db.InsertTopupIdempotencyParams{
			UserID:         userID,
			IdempotencyKey: idemKey,
			Amount:         req.GetAmount(),
		})
		if err != nil {
			return err
		}

		if inserted == 0 {
			existing, err := q.GetTopupIdempotency(ctx, db.GetTopupIdempotencyParams{
				UserID:         userID,
				IdempotencyKey: idemKey,
			})
			if err != nil {
				return err
			}
			if existing.Amount != req.GetAmount() {
				return status.Error(codes.FailedPrecondition, "idempotency key reuse with different parameters")
			}
			balance = existing.BalanceAfter
			return nil
		}

		account, err := q.TopUp(ctx, db.TopUpParams{
			UserID:  userID,
			Balance: req.GetAmount(),
		})
		if err != nil {
			_ = q.DeleteTopupIdempotency(ctx, db.DeleteTopupIdempotencyParams{
				UserID:         userID,
				IdempotencyKey: idemKey,
			})
			if errors.Is(err, pgx.ErrNoRows) {
				return status.Error(codes.NotFound, "account not found")
			}
			return err
		}

		if _, err := q.SetTopupIdempotencyBalance(ctx, db.SetTopupIdempotencyBalanceParams{
			UserID:         userID,
			IdempotencyKey: idemKey,
			BalanceAfter:   account.Balance,
		}); err != nil {
			return err
		}

		balance = account.Balance
		updateCache = true
		return nil
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, st.Err()
		}
		return nil, status.Error(codes.Internal, "failed to top up")
	}

	if updateCache && h.cache != nil {
		_ = h.cache.Set(ctx, cache.Balance{
			UserID:  userID,
			Balance: balance,
		})
	}

	return &paymentsv1.TopUpResponse{
		Account: &paymentsv1.Account{
			UserId:  userID,
			Balance: balance,
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
