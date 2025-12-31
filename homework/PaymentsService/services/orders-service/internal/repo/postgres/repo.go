package postgres

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres/db"
)

type Repo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	slog.Default().With("service", "orders-service", "component", "repo").Info("repository initialized")
	return &Repo{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Repo) Pool() *pgxpool.Pool {
	slog.Default().With("service", "orders-service", "component", "repo").Info("repository pool accessed")
	return r.pool
}

func (r *Repo) Q() *db.Queries {
	slog.Default().With("service", "orders-service", "component", "repo").Info("repository queries accessed")
	return r.q
}

func (r *Repo) WithTx(ctx context.Context, fn func(tx pgx.Tx, q *db.Queries) error) (err error) {
	start := time.Now()
	logger := slog.Default().With("service", "orders-service", "component", "repo")
	logger.Info("transaction start")
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		logger.Error("transaction begin failed", "err", err)
		return err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				logger.Error("transaction rollback failed", "err", rbErr)
			}
			logger.Error("transaction failed", "err", err, "duration", time.Since(start))
		} else {
			logger.Info("transaction completed", "duration", time.Since(start))
		}
	}()

	qtx := db.New(tx)
	if err = fn(tx, qtx); err != nil {
		logger.Error("transaction function failed", "err", err)
		return err
	}
	return tx.Commit(ctx)
}
