package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

func (h *Handler) GetSubmission(w http.ResponseWriter, r *http.Request, submissionId string) {
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.NewString()
	}

	log.Printf("[Info] GetSubmission called, request_id: %s, submission_id: %s", reqID, submissionId)

	submission, err := h.store.GetSubmission(r.Context(), submissionId)
	if err != nil {
		if err == store.ErrSubmissionNotFound {
			if err := writeError(w, http.StatusNotFound, api.SUBMISSIONNOTFOUND, "Submission with this submissionId not found"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		log.Printf("[Error] Failed to fetch submission, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	reportResp, err := h.fileAnalysisClient.GetReportWithResponse(r.Context(), submissionId)
	if err != nil {
		log.Printf("[Error] file-analysis request failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "File analysis service is unavailable"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	status := api.SubmissionDetailStatusQUEUED
	var report *api.PlagiarismReport
	if reportResp.JSON200 != nil {
		status = api.SubmissionDetailStatus(reportResp.JSON200.Status)
		report = mapAnalysisReport(reportResp.JSON200)
	} else if reportResp.JSON404 != nil {
		status = api.SubmissionDetailStatusQUEUED
	} else if reportResp.JSON500 != nil {
		log.Printf("[Error] file-analysis error, request_id: %s, err: %s", reqID, reportResp.JSON500.Message)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, reportResp.JSON500.Message); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	} else if reportResp.StatusCode() != http.StatusOK && reportResp.StatusCode() != http.StatusNotFound {
		log.Printf("[Error] file-analysis returned unexpected status, request_id: %s, status: %s", reqID, reportResp.Status())
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Unexpected response from file analysis"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(api.SubmissionDetail{
		SubmissionId: submission.SubmissionID,
		WorkId:       submission.WorkID,
		FileId:       submission.FileID,
		UploadedAt:   submission.UploadedAt,
		Status:       status,
		Report:       report,
	}); err != nil {
		log.Printf("[Error] encode response failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode response"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
	}
}
