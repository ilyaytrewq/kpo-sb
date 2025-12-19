package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

func (h *Handler) CreateWork(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateWork handler called, body: %v", r.Body)

	var work api.WorkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&work); err != nil {
		log.Printf("[Error] Failed parse body: %v", err)
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, fmt.Sprintf("Failed parse body with error: %v", err)); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if strings.TrimSpace(work.WorkId) == "" || strings.TrimSpace(work.Name) == "" || strings.TrimSpace(work.Description) == "" {
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "workId, name and description are required"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	created, err := h.store.CreateWork(r.Context(), work.WorkId, work.Name, work.Description)
	if err != nil {
		log.Printf("[Error] Failed create work: %v", err)
		switch err {
		case store.ErrWorkAlreadyExists:
			if err := writeError(w, http.StatusConflict, api.WORKALREADYEXISTS, "Work already exists"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
		default:
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(api.WorkCreateResponse{
		WorkId:      created.WorkID,
		Name:        created.Name,
		Description: created.Description,
		CreatedAt:   created.CreatedAt,
	}); err != nil {
		log.Printf("[Error] Failed encode response: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
}
