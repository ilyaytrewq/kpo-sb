package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/embedding-service/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/embedding-service/internal/yandexembd"
)

const (
	maxBytes = 50_000
)

type handler struct {
	model string
	client *yandexembd.Client
}

func NewHandler() (*handler, error) {
	client, err := yandexembd.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create yandex embedding client: %w", err)
	}
	model := fmt.Sprintf("emb://%s/text-search-doc/latest", client.FolderID)
	return &handler{model: model, client: client}, nil
}


func (h *handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
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
		return fmt.Errorf("failed to write error response: %w", err)
	}
	return nil
}