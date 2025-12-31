package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/handler/helpers"
)

// (GET /works/{workId}/reports)
func (h *Handler) GetWorkReports(w http.ResponseWriter, r *http.Request, workId string) {
	workId = strings.TrimSpace(workId)
	if workId == "" {
		if err := helpers.WriteError(w, http.StatusBadRequest, api.VALIDATIONERROR, "workId is required"); err != nil {
			return
		}
		return
	}

	items, err := h.reportStore.ListWorkReports(r.Context(), workId)
	if err != nil {
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to load reports"); err != nil {
			return
		}
		return
	}
	if len(items) == 0 {
		if err := helpers.WriteError(w, http.StatusNotFound, api.WORKNOTFOUND, "Work not found"); err != nil {
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(api.WorkReportsResponse(items)); err != nil {
		return
	}
}
