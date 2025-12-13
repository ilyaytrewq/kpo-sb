package Handler

import (
	"encoding/json"
	"log"
	"github.com/google/uuid"
	"net/http"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
)

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.NewString()
	}

	log.Printf("[Info] GetSubmission called, request_id: %s, submission_id: %s", reqID, submissionId)

	resp, err := h.fileAnalysisClient.GetReportWithResponse(r.Context(), submissionId)
	if err != nil {
		log.Printf("[Error] file-analysis request failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusBadGateway, api.INTERNALERROR, "File analysis service is unavailable"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if resp.StatusCode() == http.StatusOK {
		if resp.JSON200 != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(resp.JSON200); err != nil {
				log.Printf("[Error] encode response failed, request_id: %s, err: %v", reqID, err)
				if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode response"); err != nil {
					log.Printf("[Error] Failed write error response: %v", err)
				}
			}
			return
		}

		body := resp.Body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		n, err := w.Write(body)
		if err != nil {
			log.Printf("[Error] write response failed, request_id: %s, err: %v", reqID, err)
		} else if n != len(body) {
			log.Printf("[Error] write response incomplete, request_id: %s, written: %d, expected: %d", reqID, n, len(body))
		}
		return
	}

	switch resp.StatusCode() {
	case http.StatusNotFound:
		if resp.JSON404 != nil {
			if err := writeError(w, http.StatusNotFound, api.FORBIDDEN, resp.JSON404.Message) ; err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		if err := writeError(w, http.StatusNotFound, api.SUBMISSIONNOTFOUND, "Report not found") ; err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return

	case http.StatusBadRequest:
		if resp.JSON404 != nil {
			if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, resp.JSON404.Message); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "Invalid request"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return

	default:
		log.Printf("[Error] file-analysis returned non-OK status: %s, status: %d, body: %s", reqID, resp.StatusCode(), string(resp.Body))
		if err := writeError(w, http.StatusBadGateway, api.INTERNALERROR, "Unexpected response from file analysis"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
}
