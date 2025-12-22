package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)

type fakeReportStore struct {
	saveReportFn     func(ctx context.Context, report api.ReportResponse, matches []api.MatchedSubmission) error
	getReportFn      func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error)
	listWorkReportsFn func(ctx context.Context, workID string) ([]api.WorkReportItem, error)
}

func (f *fakeReportStore) SaveReport(ctx context.Context, report api.ReportResponse, matches []api.MatchedSubmission) error {
	if f.saveReportFn == nil {
		return nil
	}
	return f.saveReportFn(ctx, report, matches)
}

func (f *fakeReportStore) GetReport(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) {
	return f.getReportFn(ctx, submissionID)
}

func (f *fakeReportStore) ListWorkReports(ctx context.Context, workID string) ([]api.WorkReportItem, error) {
	return f.listWorkReportsFn(ctx, workID)
}

func TestGetReport_EmptyID(t *testing.T) {
	h := &Handler{reportStore: &fakeReportStore{
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) {
			t.Fatal("GetReport should not be called with empty submissionId")
			return api.ReportResponse{}, false, nil
		},
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/reports/", nil)
	rec := httptest.NewRecorder()

	h.GetReport(rec, req, " ")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetReport_NotFound(t *testing.T) {
	h := &Handler{reportStore: &fakeReportStore{
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) {
			return api.ReportResponse{}, false, nil
		},
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/reports/sub-1", nil)
	rec := httptest.NewRecorder()

	h.GetReport(rec, req, "sub-1")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetReport_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := &Handler{reportStore: &fakeReportStore{
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) {
			report := api.ReportResponse{
				SubmissionId:       submissionID,
				WorkId:             "work-1",
				Status:             api.ReportResponseStatusDONE,
				PlagiarismDetected: true,
				SimilarityPercent:  75.5,
				CreatedAt:          now,
			}
			return report, true, nil
		},
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/reports/sub-1", nil)
	rec := httptest.NewRecorder()

	h.GetReport(rec, req, "sub-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp api.ReportResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.SubmissionId != "sub-1" || resp.WorkId != "work-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetWorkReports_EmptyWorkID(t *testing.T) {
	h := &Handler{reportStore: &fakeReportStore{
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) {
			t.Fatal("ListWorkReports should not be called with empty workId")
			return nil, nil
		},
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) { return api.ReportResponse{}, false, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/works//reports", nil)
	rec := httptest.NewRecorder()

	h.GetWorkReports(rec, req, " ")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetWorkReports_NotFound(t *testing.T) {
	h := &Handler{reportStore: &fakeReportStore{
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) {
			return []api.WorkReportItem{}, nil
		},
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) { return api.ReportResponse{}, false, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/works/work-1/reports", nil)
	rec := httptest.NewRecorder()

	h.GetWorkReports(rec, req, "work-1")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestGetWorkReports_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	h := &Handler{reportStore: &fakeReportStore{
		listWorkReportsFn: func(ctx context.Context, workID string) ([]api.WorkReportItem, error) {
			return []api.WorkReportItem{
				{
					SubmissionId:       "sub-1",
					Status:             api.DONE,
					PlagiarismDetected: false,
					SimilarityPercent:  12.3,
					CreatedAt:          now,
				},
			}, nil
		},
		getReportFn: func(ctx context.Context, submissionID string) (api.ReportResponse, bool, error) { return api.ReportResponse{}, false, nil },
	}}

	req := httptest.NewRequest(http.MethodGet, "/works/work-1/reports", nil)
	rec := httptest.NewRecorder()

	h.GetWorkReports(rec, req, "work-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp api.WorkReportsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp) != 1 || resp[0].SubmissionId != "sub-1" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
