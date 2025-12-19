package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Submission struct {
	SubmissionID string
	WorkID       string
	FileID       uuid.UUID
	UploadedAt   time.Time
}

func (s *Store) CreateSubmission(ctx context.Context, submission Submission) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO submissions (submission_id, work_id, file_id, uploaded_at)
		VALUES ($1, $2, $3, $4);
	`, submission.SubmissionID, submission.WorkID, submission.FileID, submission.UploadedAt)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetSubmission(ctx context.Context, submissionID string) (Submission, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT submission_id, work_id, file_id, uploaded_at
		FROM submissions
		WHERE submission_id = $1;
	`, submissionID)

	var submission Submission
	if err := row.Scan(&submission.SubmissionID, &submission.WorkID, &submission.FileID, &submission.UploadedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Submission{}, ErrSubmissionNotFound
		}
		return Submission{}, err
	}
	return submission, nil
}

func (s *Store) ListSubmissionsByWork(ctx context.Context, workID string) ([]Submission, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT submission_id, work_id, file_id, uploaded_at
		FROM submissions
		WHERE work_id = $1
		ORDER BY uploaded_at ASC;
	`, workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []Submission
	for rows.Next() {
		var submission Submission
		if err := rows.Scan(&submission.SubmissionID, &submission.WorkID, &submission.FileID, &submission.UploadedAt); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return submissions, nil
}
