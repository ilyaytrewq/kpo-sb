package Handler

import (
	"encoding/json"
	"log"
	"net/http"
)

func (h *Handler) ListWorkReports(w http.ResponseWriter, r *http.Request, workId string) {
	log.Printf("ListWorkReports handler called for workId: %s", workId)

	h.fileAnalysisClient.GetWorkReportsWithResponse(r.Context(), workId)

	reportsResponse, err := h.fileAnalysisClient.GetWorkReportsWithResponse(r.Context(), workId)
	if err != nil {
		log.Printf("[Error] Failed to get work reports: %v", err)
		http.Error(w, "Failed to get work reports", http.StatusInternalServerError)
		return
	}

	if reportsResponse.StatusCode() != http.StatusOK {
		log.Printf("[Error] File analysis service returned non-OK status: %d", reportsResponse.StatusCode())
		http.Error(w, "Failed to get work reports", reportsResponse.StatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reportsResponse.JSON200); err != nil {
		log.Printf("[Error] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	
}
