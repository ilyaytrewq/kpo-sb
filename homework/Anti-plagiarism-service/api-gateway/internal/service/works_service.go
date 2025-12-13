package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/db/gen"
)

type WorksService struct {
	q *dbgen.Queries
}

func NewWorksService(q *dbgen.Queries) *WorksService {
	return &WorksService{q: q}
}

func (s *WorksService) CreateWork(ctx context.Context, name string, description *string) (dbgen.Work, error) {
	id := pgtype.UUID{
		Bytes: uuid.New(),
		Valid: true,
	}
	desc := pgtype.Text{
		String: *description,
		Valid: true,
	}
	return s.q.CreateWork(ctx, dbgen.CreateWorkParams{
		WorkID:      id,
		Name:       name,
		Description: desc,
	})
}

func (s *WorksService) GetWork(ctx context.Context, workID uuid.UUID) (dbgen.Work, error) {
	id := pgtype.UUID{
		Bytes: workID,
		Valid: true,
	}
	return s.q.GetWork(ctx, id)
}