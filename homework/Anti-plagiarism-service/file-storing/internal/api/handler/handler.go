package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
	service "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/service"
)

const (
	maxFileSize = 8 << 20
)

type Handler struct {
	service ObjectService
}

type ObjectService interface {
	Upload(ctx context.Context, bucket, key, contentType, originalFileName string, body io.Reader, size int64) error
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, string, error)
	Head(ctx context.Context, bucket, key string) (service.Info, error)
	Bucket() string
}

func NewHandler() (*Handler, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	s3Service, err := service.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return &Handler{service: s3Service}, nil
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func writeError(w http.ResponseWriter, status int, code api.ErrorCode, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(api.ErrorResponse{
		Code:    code,
		Message: msg,
	}); err != nil {
		return err
	}

	return nil
}
