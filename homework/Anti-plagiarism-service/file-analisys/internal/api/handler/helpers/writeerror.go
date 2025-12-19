package helpers

import (
	"net/http"
	"encoding/json"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)


func WriteError(w http.ResponseWriter, status int, code api.ErrorCode, msg string) error {
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
