package handler

import (
	"context"
	"time"
	"encoding/json"
	"net/http"

	service "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/service"
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
)

// ServerInterface represents all server handlers.
/*type ServerInterface interface {
	// Upload a document file
	// (POST /files/upload)
	UploadFile(w http.ResponseWriter, r *http.Request, params UploadFileParams)
	// Download a file
	// (GET /files/{fileId})
	DownloadFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID, params DownloadFileParams)
	// Get file metadata
	// (GET /files/{fileId}/info)
	GetFileInfo(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID, params GetFileInfoParams)
	// Health check
	// (GET /health)
	Health(w http.ResponseWriter, r *http.Request)
}
*/

const (
	maxFileSize = 8 << 20 // 64 MB
)

type Handler struct {
	service *service.S3Service
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
