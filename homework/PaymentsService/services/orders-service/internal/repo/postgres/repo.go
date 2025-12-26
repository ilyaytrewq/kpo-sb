package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/ilyaytrewq/payments-service/order-service/internal/repo/postgres/db"
)

type Repo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		pool: pool,
		q:    db.New(pool),
	}
}

func (r *Repo) Pool() *pgxpool.Pool { return r.pool }
func (r *Repo) Q() *db.Queries      { return r.q }

func (r *Repo) WithTx(ctx context.Context, fn func(tx pgx.Tx, q *db.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	qtx := db.New(tx)
	if err := fn(tx, qtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
