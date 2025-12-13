package Handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	service "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/service"
)

func (h *Handler) CreateWork(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateWork handler called, body: %v", r.Body)

	var work api.WorkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&work); err != nil {
		log.Printf("[Error] Failed parse body: %v", err)
		if err := writeError(w, http.StatusBadRequest,  api.VALIDATIONERROR, fmt.Sprintf("Failed parse body with error: %v", err)); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	workResponse, err := h.worksService.CreateWork(r.Context(), work.Name, &work.Description)
	if err != nil {
		log.Printf("[Error] Failed create work: %v", err)
		switch err {
		case service.ErrWorkAlreadyExists:
			if err := writeError(w, http.StatusConflict, api.WORKALREADYEXISTS, "Work already exists"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
		default:
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(workResponse); err != nil {
		log.Printf("[Error] Failed encode response: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
}
