package handler

import (
	"testing"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	fileanalysis "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/fileanalysis"
)

func TestMapMatchedSubmissions_Nil(t *testing.T) {
	if got := mapMatchedSubmissions(nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestMapAnalysisReport(t *testing.T) {
	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	done := now.Add(time.Minute)
	matches := []fileanalysis.MatchedSubmission{
		{
			SubmissionId:      "sub-2",
			MatchedChunks:     3,
			SimilarityPercent: 55.5,
		},
	}
	report := &fileanalysis.ReportResponse{
		SubmissionId:       "sub-1",
		WorkId:             "work-1",
		Status:             fileanalysis.ReportResponseStatusDONE,
		PlagiarismDetected: true,
		SimilarityPercent:  88.2,
		CreatedAt:          now,
		CompletedAt:        &done,
		MatchedSubmissions: &matches,
	}

	got := mapAnalysisReport(report)
	if got == nil {
		t.Fatal("expected non-nil report")
	}
	if got.ReportId != "rep-sub-1" {
		t.Fatalf("unexpected reportId: %s", got.ReportId)
	}
	if got.Status != api.PlagiarismReportStatusDONE {
		t.Fatalf("unexpected status: %s", got.Status)
	}
	if got.MatchedSubmissions == nil || len(*got.MatchedSubmissions) != 1 {
		t.Fatalf("expected 1 matched submission, got %#v", got.MatchedSubmissions)
	}
	if (*got.MatchedSubmissions)[0].SimilarityPercent != float32(55.5) {
		t.Fatalf("unexpected similarity percent: %v", (*got.MatchedSubmissions)[0].SimilarityPercent)
	}
}

func TestMapWorkReportItem(t *testing.T) {
	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)
	report := fileanalysis.WorkReportItem{
		SubmissionId:       "sub-3",
		Status:             fileanalysis.DONE,
		PlagiarismDetected: false,
		SimilarityPercent:  12.3,
		CreatedAt:          now,
	}
	got := mapWorkReportItem(report)
	if got.ReportId != "rep-sub-3" {
		t.Fatalf("unexpected reportId: %s", got.ReportId)
	}
	if got.Status != api.WorkReportItemStatusDONE {
		t.Fatalf("unexpected status: %s", got.Status)
	}
}
