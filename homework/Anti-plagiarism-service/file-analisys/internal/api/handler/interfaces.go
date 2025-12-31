package handler

import (
	"context"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)

type ReportStore interface {
	SaveReport(ctx context.Context, report api.ReportResponse, matches []api.MatchedSubmission) error
	GetReport(ctx context.Context, submissionID string) (api.ReportResponse, bool, error)
	ListWorkReports(ctx context.Context, workID string) ([]api.WorkReportItem, error)
}
