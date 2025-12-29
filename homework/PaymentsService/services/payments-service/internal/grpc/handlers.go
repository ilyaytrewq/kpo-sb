package grpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

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

var logger = slog.Default().With("service", "payments-service", "component", "grpc")

func NewHandlers(repo *postgres.Repo, cache *cache.BalanceCache) *Handlers {
	logger.Info("handlers initialized")
	return &Handlers{repo: repo, cache: cache}
}

func (h *Handlers) CreateAccount(ctx context.Context, req *paymentsv1.CreateAccountRequest) (resp *paymentsv1.CreateAccountResponse, err error) {
	start := time.Now()
	logger.Info("create account start", "user_id", req.GetUserId(), "has_idempotency_key", req.GetIdempotencyKey() != "")
	defer func() {
		if err != nil {
			logger.Error("create account failed", "err", err, "duration", time.Since(start))
			return
		}
		logger.Info("create account completed", "duration", time.Since(start))
	}()

	userID := req.GetUserId()
	if userID == "" {
		err = status.Error(codes.InvalidArgument, "user_id is required")
		logger.Error("create account validation failed", "err", err)
		return nil, err
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
				err = status.Error(codes.AlreadyExists, "account already exists")
				logger.Error("create account conflict", "err", err)
				return nil, err
			}
			err = status.Error(codes.Internal, "failed to create account")
			logger.Error("create account failed", "err", err)
			return nil, err
		}
		accountUserID = account.UserID
		accountBalance = account.Balance
	} else {
		account, err := h.repo.Q().CreateAccountIdempotent(ctx, userID)
		if err != nil {
			err = status.Error(codes.Internal, "failed to create account")
			logger.Error("create account idempotent failed", "err", err)
			return nil, err
		}
		accountUserID = account.UserID
		accountBalance = account.Balance
	}

	if h.cache != nil {
		if err := h.cache.Set(ctx, cache.Balance{
			UserID:  accountUserID,
			Balance: accountBalance,
		}); err != nil {
			logger.Error("cache set failed", "err", err, "user_id", accountUserID)
		}
	}

	resp = &paymentsv1.CreateAccountResponse{
		Account: &paymentsv1.Account{
			UserId:  accountUserID,
			Balance: accountBalance,
		},
	}
	return resp, nil
}

func (h *Handlers) TopUp(ctx context.Context, req *paymentsv1.TopUpRequest) (resp *paymentsv1.TopUpResponse, err error) {
	start := time.Now()
	logger.Info("top up start", "user_id", req.GetUserId(), "amount", req.GetAmount(), "has_idempotency_key", req.GetIdempotencyKey() != "")
	defer func() {
		if err != nil {
			logger.Error("top up failed", "err", err, "duration", time.Since(start))
			return
		}
		logger.Info("top up completed", "duration", time.Since(start))
	}()

	userID := req.GetUserId()
	if userID == "" {
		err = status.Error(codes.InvalidArgument, "user_id is required")
		logger.Error("top up validation failed", "err", err)
		return nil, err
	}
	if req.GetAmount() <= 0 {
		err = status.Error(codes.InvalidArgument, "amount must be > 0")
		logger.Error("top up validation failed", "err", err)
		return nil, err
	}

	idemKey := req.GetIdempotencyKey()
	if idemKey == "" {
		account, err := h.repo.Q().TopUp(ctx, db.TopUpParams{
			UserID:  userID,
			Balance: req.GetAmount(),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				err = status.Error(codes.NotFound, "account not found")
				logger.Error("top up account not found", "err", err)
				return nil, err
			}
			err = status.Error(codes.Internal, "failed to top up")
			logger.Error("top up failed", "err", err)
			return nil, err
		}

		if h.cache != nil {
			if err := h.cache.Set(ctx, cache.Balance{
				UserID:  account.UserID,
				Balance: account.Balance,
			}); err != nil {
				logger.Error("cache set failed", "err", err, "user_id", account.UserID)
			}
		}

		resp = &paymentsv1.TopUpResponse{
			Account: &paymentsv1.Account{
				UserId:  account.UserID,
				Balance: account.Balance,
			},
		}
		return resp, nil
	}

	var (
		balance     int64
		updateCache bool
	)
	err = h.repo.WithTx(ctx, func(_ pgx.Tx, q *db.Queries) error {
		inserted, err := q.InsertTopupIdempotency(ctx, db.InsertTopupIdempotencyParams{
			UserID:         userID,
			IdempotencyKey: idemKey,
			Amount:         req.GetAmount(),
		})
		if err != nil {
			logger.Error("insert topup idempotency failed", "err", err)
			return err
		}

		if inserted == 0 {
			existing, err := q.GetTopupIdempotency(ctx, db.GetTopupIdempotencyParams{
				UserID:         userID,
				IdempotencyKey: idemKey,
			})
			if err != nil {
				logger.Error("get topup idempotency failed", "err", err)
				return err
			}
			if existing.Amount != req.GetAmount() {
				err = status.Error(codes.FailedPrecondition, "idempotency key reuse with different parameters")
				logger.Error("idempotency key reuse with different parameters", "err", err)
				return err
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
				err = status.Error(codes.NotFound, "account not found")
				logger.Error("top up account not found", "err", err)
				return err
			}
			logger.Error("top up failed", "err", err)
			return err
		}

		if _, err := q.SetTopupIdempotencyBalance(ctx, db.SetTopupIdempotencyBalanceParams{
			UserID:         userID,
			IdempotencyKey: idemKey,
			BalanceAfter:   account.Balance,
		}); err != nil {
			logger.Error("set topup idempotency balance failed", "err", err)
			return err
		}

		balance = account.Balance
		updateCache = true
		return nil
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			err = st.Err()
			return nil, err
		}
		err = status.Error(codes.Internal, "failed to top up")
		return nil, err
	}

	if updateCache && h.cache != nil {
		if err := h.cache.Set(ctx, cache.Balance{
			UserID:  userID,
			Balance: balance,
		}); err != nil {
			logger.Error("cache set failed", "err", err, "user_id", userID)
		}
	}

	resp = &paymentsv1.TopUpResponse{
		Account: &paymentsv1.Account{
			UserId:  userID,
			Balance: balance,
		},
	}
	return resp, nil
}

func (h *Handlers) GetBalance(ctx context.Context, req *paymentsv1.GetBalanceRequest) (resp *paymentsv1.GetBalanceResponse, err error) {
	start := time.Now()
	logger.Info("get balance start", "user_id", req.GetUserId())
	defer func() {
		if err != nil {
			logger.Error("get balance failed", "err", err, "duration", time.Since(start))
			return
		}
		logger.Info("get balance completed", "duration", time.Since(start))
	}()

	userID := req.GetUserId()
	if userID == "" {
		err = status.Error(codes.InvalidArgument, "user_id is required")
		logger.Error("get balance validation failed", "err", err)
		return nil, err
	}

	if cached, err := h.cache.Get(ctx, userID); err == nil && cached != nil {
		logger.Info("get balance cache hit", "user_id", userID)
		resp = &paymentsv1.GetBalanceResponse{
			Balance: cached.Balance,
		}
		return resp, nil
	}
	logger.Info("get balance cache miss", "user_id", userID)

	balance, err := h.repo.Q().GetBalance(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = status.Error(codes.NotFound, "account not found")
			logger.Error("get balance account not found", "err", err)
			return nil, err
		}
		err = status.Error(codes.Internal, "failed to get balance")
		logger.Error("get balance failed", "err", err)
		return nil, err
	}

	if h.cache != nil {
		if err := h.cache.Set(ctx, cache.Balance{
			UserID:  userID,
			Balance: balance,
		}); err != nil {
			logger.Error("cache set failed", "err", err, "user_id", userID)
		}
	}

	resp = &paymentsv1.GetBalanceResponse{
		Balance: balance,
	}
	return resp, nil
}
