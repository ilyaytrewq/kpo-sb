package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

type fakeStore struct {
	createWorkFn        func(ctx context.Context, workID, name, description string) (store.Work, error)
	getWorkFn           func(ctx context.Context, workID string) (store.Work, error)
	createSubmissionFn  func(ctx context.Context, submission store.Submission) error
	getSubmissionFn     func(ctx context.Context, submissionID string) (store.Submission, error)
	listSubmissionsFn   func(ctx context.Context, workID string) ([]store.Submission, error)
}

func (f *fakeStore) CreateWork(ctx context.Context, workID, name, description string) (store.Work, error) {
	return f.createWorkFn(ctx, workID, name, description)
}

func (f *fakeStore) GetWork(ctx context.Context, workID string) (store.Work, error) {
	return f.getWorkFn(ctx, workID)
}

func (f *fakeStore) CreateSubmission(ctx context.Context, submission store.Submission) error {
	return f.createSubmissionFn(ctx, submission)
}

func (f *fakeStore) GetSubmission(ctx context.Context, submissionID string) (store.Submission, error) {
	return f.getSubmissionFn(ctx, submissionID)
}

func (f *fakeStore) ListSubmissionsByWork(ctx context.Context, workID string) ([]store.Submission, error) {
	return f.listSubmissionsFn(ctx, workID)
}

func TestCreateWork_InvalidJSON(t *testing.T) {
	h := &Handler{store: &fakeStore{
		createWorkFn: func(ctx context.Context, workID, name, description string) (store.Work, error) {
			t.Fatal("CreateWork should not be called on invalid JSON")
			return store.Work{}, nil
		},
		getWorkFn:         func(ctx context.Context, workID string) (store.Work, error) { return store.Work{}, nil },
		createSubmissionFn: func(ctx context.Context, submission store.Submission) error { return nil },
		getSubmissionFn:    func(ctx context.Context, submissionID string) (store.Submission, error) { return store.Submission{}, nil },
		listSubmissionsFn:  func(ctx context.Context, workID string) ([]store.Submission, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	h.CreateWork(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != api.VALIDATIONERROR {
		t.Fatalf("expected code %s, got %s", api.VALIDATIONERROR, resp.Code)
	}
}

func TestCreateWork_MissingFields(t *testing.T) {
	h := &Handler{store: &fakeStore{
		createWorkFn: func(ctx context.Context, workID, name, description string) (store.Work, error) {
			t.Fatal("CreateWork should not be called when required fields missing")
			return store.Work{}, nil
		},
		getWorkFn:         func(ctx context.Context, workID string) (store.Work, error) { return store.Work{}, nil },
		createSubmissionFn: func(ctx context.Context, submission store.Submission) error { return nil },
		getSubmissionFn:    func(ctx context.Context, submissionID string) (store.Submission, error) { return store.Submission{}, nil },
		listSubmissionsFn:  func(ctx context.Context, workID string) ([]store.Submission, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewBufferString(`{"workId":"","name":"","description":""}`))
	rec := httptest.NewRecorder()

	h.CreateWork(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateWork_AlreadyExists(t *testing.T) {
	h := &Handler{store: &fakeStore{
		createWorkFn: func(ctx context.Context, workID, name, description string) (store.Work, error) {
			return store.Work{}, store.ErrWorkAlreadyExists
		},
		getWorkFn:         func(ctx context.Context, workID string) (store.Work, error) { return store.Work{}, nil },
		createSubmissionFn: func(ctx context.Context, submission store.Submission) error { return nil },
		getSubmissionFn:    func(ctx context.Context, submissionID string) (store.Submission, error) { return store.Submission{}, nil },
		listSubmissionsFn:  func(ctx context.Context, workID string) ([]store.Submission, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewBufferString(`{"workId":"hw-1","name":"HW","description":"Desc"}`))
	rec := httptest.NewRecorder()

	h.CreateWork(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != api.WORKALREADYEXISTS {
		t.Fatalf("expected code %s, got %s", api.WORKALREADYEXISTS, resp.Code)
	}
}

func TestCreateWork_Success(t *testing.T) {
	now := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	h := &Handler{store: &fakeStore{
		createWorkFn: func(ctx context.Context, workID, name, description string) (store.Work, error) {
			return store.Work{
				WorkID:      workID,
				Name:        name,
				Description: description,
				CreatedAt:   now,
			}, nil
		},
		getWorkFn:         func(ctx context.Context, workID string) (store.Work, error) { return store.Work{}, nil },
		createSubmissionFn: func(ctx context.Context, submission store.Submission) error { return nil },
		getSubmissionFn:    func(ctx context.Context, submissionID string) (store.Submission, error) { return store.Submission{}, nil },
		listSubmissionsFn:  func(ctx context.Context, workID string) ([]store.Submission, error) { return nil, nil },
	}}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewBufferString(`{"workId":"hw-1","name":"HW","description":"Desc"}`))
	rec := httptest.NewRecorder()

	h.CreateWork(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	var resp api.WorkCreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.WorkId != "hw-1" || resp.Name != "HW" || resp.Description != "Desc" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if !resp.CreatedAt.Equal(now) {
		t.Fatalf("expected createdAt %s, got %s", now, resp.CreatedAt)
	}
}
