package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrWorkAlreadyExists       = errors.New("work already exists")
	ErrWorkNotFound            = errors.New("work not found")
	ErrSubmissionAlreadyExists = errors.New("submission already exists")
	ErrSubmissionNotFound      = errors.New("submission not found")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS works (
			work_id text PRIMARY KEY,
			name text NOT NULL,
			description text NOT NULL,
			created_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		ALTER TABLE IF EXISTS works
		ALTER COLUMN work_id TYPE text USING work_id::text;
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE works
		SET description = ''
		WHERE description IS NULL;
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		ALTER TABLE IF EXISTS works
		ALTER COLUMN description SET NOT NULL;
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS submissions (
			submission_id text PRIMARY KEY,
			work_id text NOT NULL REFERENCES works(work_id) ON DELETE CASCADE,
			file_id uuid NOT NULL,
			uploaded_at timestamptz NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS submissions_work_id_idx ON submissions (work_id);`)
	return err
}
