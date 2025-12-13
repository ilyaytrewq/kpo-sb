package Handler

import (
	"net/http"
	"log"
	"encoding/json"
)

func (h *Handler) GetWorkStats(w http.ResponseWriter, r *http.Request, workId string) {
	log.Printf("GetWorkStats handler called for workId: %s", workId)

	statsResponse, err := h.fileAnalysisClient.GetWorkReportsWithResponse(r.Context(), workId)
	if err != nil {
		log.Printf("[Error] Failed to get work stats: %v", err)
		http.Error(w, "Failed to get work stats", http.StatusInternalServerError)
		return
	}

	if statsResponse.StatusCode() != http.StatusOK {
		log.Printf("[Error] File analysis service returned non-OK status: %d", statsResponse.StatusCode())
		http.Error(w, "Failed to get work stats", statsResponse.StatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statsResponse.JSON200); err != nil {
		log.Printf("[Error] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
