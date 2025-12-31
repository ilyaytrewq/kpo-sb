package reports

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	if dsn == "" {
		return nil, errors.New("analysis database dsn is empty")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if err := ensureSchema(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

func (s *Store) SaveReport(ctx context.Context, report api.ReportResponse, matches []api.MatchedSubmission) error {
	if s == nil || s.pool == nil {
		return errors.New("report store is not initialized")
	}
	if report.SubmissionId == "" {
		return errors.New("submissionId is empty")
	}

	var completedAt interface{}
	if report.CompletedAt != nil {
		completedAt = *report.CompletedAt
	}
	var errorMessage interface{}
	if report.ErrorMessage != nil {
		errorMessage = *report.ErrorMessage
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
		INSERT INTO analysis_reports (
			submission_id,
			work_id,
			status,
			plagiarism_detected,
			similarity_percent,
			created_at,
			completed_at,
			error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (submission_id) DO UPDATE SET
			work_id = EXCLUDED.work_id,
			status = EXCLUDED.status,
			plagiarism_detected = EXCLUDED.plagiarism_detected,
			similarity_percent = EXCLUDED.similarity_percent,
			created_at = EXCLUDED.created_at,
			completed_at = EXCLUDED.completed_at,
			error_message = EXCLUDED.error_message
	`,
		report.SubmissionId,
		report.WorkId,
		string(report.Status),
		report.PlagiarismDetected,
		float64(report.SimilarityPercent),
		report.CreatedAt,
		completedAt,
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("upsert report: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM analysis_matches WHERE submission_id = $1`, report.SubmissionId)
	if err != nil {
		return fmt.Errorf("clear matches: %w", err)
	}

	for _, match := range matches {
		_, err = tx.Exec(ctx, `
			INSERT INTO analysis_matches (
				submission_id,
				matched_submission_id,
				similarity_percent,
				matched_chunks
			) VALUES ($1, $2, $3, $4)
		`,
			report.SubmissionId,
			match.SubmissionId,
			match.SimilarityPercent,
			match.MatchedChunks,
		)
		if err != nil {
			return fmt.Errorf("insert match: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Store) GetReport(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) {
	if s == nil || s.pool == nil {
		return api.ReportResponse{}, false, errors.New("report store is not initialized")
	}
	if submissionID == "" {
		return api.ReportResponse{}, false, errors.New("submissionId is empty")
	}

	var report api.ReportResponse
	var status string
	var similarity float64
	var completedAt pgtype.Timestamptz
	var errorMessage pgtype.Text

	err := s.pool.QueryRow(ctx, `
		SELECT
			submission_id,
			work_id,
			status,
			plagiarism_detected,
			similarity_percent,
			created_at,
			completed_at,
			error_message
		FROM analysis_reports
		WHERE submission_id = $1
	`, submissionID).Scan(
		&report.SubmissionId,
		&report.WorkId,
		&status,
		&report.PlagiarismDetected,
		&similarity,
		&report.CreatedAt,
		&completedAt,
		&errorMessage,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.ReportResponse{}, false, nil
		}
		return api.ReportResponse{}, false, fmt.Errorf("query report: %w", err)
	}

	report.Status = api.ReportResponseStatus(status)
	report.SimilarityPercent = float32(similarity)
	if completedAt.Valid {
		t := completedAt.Time
		report.CompletedAt = &t
	}
	if errorMessage.Valid {
		msg := errorMessage.String
		report.ErrorMessage = &msg
	}

	rows, err := s.pool.Query(ctx, `
		SELECT matched_submission_id, similarity_percent, matched_chunks
		FROM analysis_matches
		WHERE submission_id = $1
		ORDER BY similarity_percent DESC
	`, submissionID)
	if err != nil {
		return api.ReportResponse{}, false, fmt.Errorf("query matches: %w", err)
	}
	defer rows.Close()

	matches := make([]api.MatchedSubmission, 0)
	for rows.Next() {
		var match api.MatchedSubmission
		err = rows.Scan(&match.SubmissionId, &match.SimilarityPercent, &match.MatchedChunks)
		if err != nil {
			return api.ReportResponse{}, false, fmt.Errorf("scan match: %w", err)
		}
		matches = append(matches, match)
	}
	if rows.Err() != nil {
		return api.ReportResponse{}, false, fmt.Errorf("iterate matches: %w", rows.Err())
	}
	report.MatchedSubmissions = &matches

	return report, true, nil
}

func (s *Store) ListWorkReports(ctx context.Context, workID string) ([]api.WorkReportItem, error) {
	if s == nil || s.pool == nil {
		return nil, errors.New("report store is not initialized")
	}
	if workID == "" {
		return nil, errors.New("workId is empty")
	}

	rows, err := s.pool.Query(ctx, `
		SELECT submission_id, status, plagiarism_detected, similarity_percent, created_at, completed_at
		FROM analysis_reports
		WHERE work_id = $1
		ORDER BY created_at
	`, workID)
	if err != nil {
		return nil, fmt.Errorf("query reports: %w", err)
	}
	defer rows.Close()

	reports := make([]api.WorkReportItem, 0)
	for rows.Next() {
		var item api.WorkReportItem
		var status string
		var similarity float64
		var completedAt pgtype.Timestamptz
		if err := rows.Scan(
			&item.SubmissionId,
			&status,
			&item.PlagiarismDetected,
			&similarity,
			&item.CreatedAt,
			&completedAt,
		); err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		item.Status = api.WorkReportItemStatus(status)
		item.SimilarityPercent = float32(similarity)
		if completedAt.Valid {
			t := completedAt.Time
			item.CompletedAt = &t
		}
		reports = append(reports, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate reports: %w", rows.Err())
	}

	return reports, nil
}

func ensureSchema(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("pool is nil")
	}

	statements := []string{
		`
		CREATE TABLE IF NOT EXISTS analysis_reports (
			submission_id TEXT PRIMARY KEY,
			work_id TEXT NOT NULL,
			status TEXT NOT NULL,
			plagiarism_detected BOOLEAN NOT NULL,
			similarity_percent DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			completed_at TIMESTAMPTZ NULL,
			error_message TEXT NULL
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS analysis_matches (
			submission_id TEXT NOT NULL REFERENCES analysis_reports(submission_id) ON DELETE CASCADE,
			matched_submission_id TEXT NOT NULL,
			similarity_percent DOUBLE PRECISION NOT NULL,
			matched_chunks INTEGER NOT NULL,
			PRIMARY KEY (submission_id, matched_submission_id)
		)
		`,
		`CREATE INDEX IF NOT EXISTS analysis_reports_work_id_idx ON analysis_reports(work_id, created_at)`,
	}

	for _, stmt := range statements {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	return nil
}
