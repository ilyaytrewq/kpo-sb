package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/handler/helpers"
)

// (GET /reports/{submissionId})
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request, submissionId string) {
	submissionId = strings.TrimSpace(submissionId)
	if submissionId == "" {
		if err := helpers.WriteError(w, http.StatusBadRequest, api.VALIDATIONERROR, "submissionId is required"); err != nil {
			return
		}
		return
	}

	report, ok, err := h.reportStore.GetReport(r.Context(), submissionId)
	if err != nil {
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to load report"); err != nil {
			return
		}
		return
	}
	if !ok {
		if err := helpers.WriteError(w, http.StatusNotFound, api.REPORTNOTFOUND, "Report not found for submissionId"); err != nil {
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(report); err != nil {
		return
	}
}
