package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type Work struct {
	WorkID      string
	Name        string
	Description string
	CreatedAt   time.Time
}

func (s *Store) CreateWork(ctx context.Context, workID, name, description string) (Work, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO works (work_id, name, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (work_id) DO NOTHING
		RETURNING work_id, name, description, created_at;
	`, workID, name, description)

	var work Work
	if err := row.Scan(&work.WorkID, &work.Name, &work.Description, &work.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Work{}, ErrWorkAlreadyExists
		}
		return Work{}, err
	}
	return work, nil
}

func (s *Store) GetWork(ctx context.Context, workID string) (Work, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT work_id, name, description, created_at
		FROM works
		WHERE work_id = $1;
	`, workID)

	var work Work
	if err := row.Scan(&work.WorkID, &work.Name, &work.Description, &work.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Work{}, ErrWorkNotFound
		}
		return Work{}, err
	}
	return work, nil
}
