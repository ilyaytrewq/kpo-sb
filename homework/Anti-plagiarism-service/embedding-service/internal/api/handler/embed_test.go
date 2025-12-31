package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/embedding-service/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/embedding-service/internal/yandexembd"
)

type embedRequest struct {
	Input string `json:"input"`
}

type embedRoundTripper struct {
	embedFn func(input string) []float64
}

func (rt *embedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !strings.HasSuffix(req.URL.Path, "/embeddings") {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(bytes.NewBufferString("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	var payload embedRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString("bad request")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	embedding := rt.embedFn(payload.Input)
	if embedding == nil {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("internal error")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}
	resp := map[string]interface{}{
		"data": []map[string]interface{}{
			{
				"embedding": embedding,
				"index":     0,
			},
		},
	}
	buf, _ := json.Marshal(resp)
	header := make(http.Header)
	header.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(buf)),
		Header:     header,
		Request:    req,
	}, nil
}

func newTestHandler(embedFn func(input string) []float64) *handler {
	client := &yandexembd.Client{
		BaseURL:  "http://yandex.test",
		APIKey:   "test",
		FolderID: "folder",
		HTTP:     &http.Client{Transport: &embedRoundTripper{embedFn: embedFn}},
	}
	return &handler{model: "test-model", client: client}
}

func TestEmbed_EmptyChunk(t *testing.T) {
	h := newTestHandler(func(input string) []float64 { return []float64{0.1, 0.2} })
	reqBody, _ := json.Marshal(api.EmbedRequest{
		Chunks: []api.TextChunk{
			{ChunkId: "c1", Text: "", ChunkIndex: 0},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/embed", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h.Embed(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var resp api.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if resp.Code != api.BADREQUEST {
		t.Fatalf("expected code %s, got %s", api.BADREQUEST, resp.Code)
	}
}

func TestEmbed_TooLarge(t *testing.T) {
	h := newTestHandler(func(input string) []float64 { return []float64{0.1, 0.2} })
	tooLarge := strings.Repeat("a", maxBytes+1)
	reqBody, _ := json.Marshal(api.EmbedRequest{
		Chunks: []api.TextChunk{
			{ChunkId: "c1", Text: tooLarge, ChunkIndex: 0},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/embed", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h.Embed(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestEmbed_Success(t *testing.T) {
	h := newTestHandler(func(input string) []float64 { return []float64{0.1, 0.2, 0.3} })
	reqBody, _ := json.Marshal(api.EmbedRequest{
		Chunks: []api.TextChunk{
			{ChunkId: "c1", Text: "hello", ChunkIndex: 0},
			{ChunkId: "c2", Text: "world", ChunkIndex: 1},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/embed", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h.Embed(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	var resp api.EmbedResult
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Dimension != 3 {
		t.Fatalf("expected dimension 3, got %d", resp.Dimension)
	}
	if resp.Model != "test-model" {
		t.Fatalf("unexpected model: %s", resp.Model)
	}
	if len(resp.Embeddings) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(resp.Embeddings))
	}
}

func TestEmbed_InconsistentDimensions(t *testing.T) {
	h := newTestHandler(func(input string) []float64 {
		if strings.Contains(input, "short") {
			return []float64{0.1, 0.2}
		}
		return []float64{0.1, 0.2, 0.3}
	})
	reqBody, _ := json.Marshal(api.EmbedRequest{
		Chunks: []api.TextChunk{
			{ChunkId: "c1", Text: "short", ChunkIndex: 0},
			{ChunkId: "c2", Text: "long", ChunkIndex: 1},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/embed", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	h.Embed(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
