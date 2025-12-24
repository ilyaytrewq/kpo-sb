package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	fileanalysis "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/fileanalysis"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

func (h *Handler) GetWorkStats(w http.ResponseWriter, r *http.Request, workId string) {
	log.Printf("GetWorkStats handler called for workId: %s", workId)

	if strings.TrimSpace(workId) == "" {
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "workId is required"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if _, err := h.store.GetWork(r.Context(), workId); err != nil {
		if err == store.ErrWorkNotFound {
			if err := writeError(w, http.StatusNotFound, api.WORKNOTFOUND, "Work with this workId not found"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		log.Printf("[Error] Failed to fetch work: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	submissions, err := h.store.ListSubmissionsByWork(r.Context(), workId)
	if err != nil {
		log.Printf("[Error] Failed to list submissions: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to fetch submissions"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	reportsResponse, err := h.fileAnalysisClient.GetWorkReportsWithResponse(r.Context(), workId)
	if err != nil {
		log.Printf("[Error] Failed to get work reports: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to get work reports"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	reportMap := make(map[string]fileanalysis.WorkReportItem)
	if reportsResponse.JSON200 != nil {
		for _, report := range *reportsResponse.JSON200 {
			reportMap[report.SubmissionId] = report
		}
	} else if reportsResponse.JSON500 != nil {
		log.Printf("[Error] File analysis service error: %s", reportsResponse.JSON500.Message)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, reportsResponse.JSON500.Message); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	total := len(submissions)
	var lastSubmissionAt *time.Time
	if total > 0 {
		last := submissions[len(submissions)-1].UploadedAt
		lastSubmissionAt = &last
	}

	withPlagiarism := 0
	withoutPlagiarism := 0
	sumSimilarity := 0.0

	if total == 0 && len(reportMap) > 0 {
		for _, report := range reportMap {
			sumSimilarity += float64(report.SimilarityPercent)
			if report.PlagiarismDetected {
				withPlagiarism++
			} else {
				withoutPlagiarism++
			}
		}
		total = withPlagiarism + withoutPlagiarism
	} else {
		for _, submission := range submissions {
			if report, ok := reportMap[submission.SubmissionID]; ok {
				sumSimilarity += float64(report.SimilarityPercent)
				if report.PlagiarismDetected {
					withPlagiarism++
				} else {
					withoutPlagiarism++
				}
			} else {
				withoutPlagiarism++
			}
		}
	}

	averageSimilarity := float32(0)
	if total > 0 {
		averageSimilarity = float32(sumSimilarity / float64(total))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(api.WorkStats{
		WorkId:                       workId,
		TotalSubmissions:             total,
		SubmissionsWithPlagiarism:    withPlagiarism,
		SubmissionsWithoutPlagiarism: withoutPlagiarism,
		AverageSimilarityPercent:     averageSimilarity,
		LastSubmissionAt:             lastSubmissionAt,
	}); err != nil {
		log.Printf("[Error] Failed to encode response: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode response"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
}
