package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/analisys-functions"
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/handler/helpers"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/embedding"
)

const (
	closestVectorLimit         = 10
	plagiarismPercentThreshold = 50.0
	matchSimilarityThreshold   = 60.0
)

type submissionMatchAggregate struct {
	matchedChunks int
	totalScore    float64
}

// (POST /analyze)
func (h *Handler) AnalyzeFile(w http.ResponseWriter, r *http.Request) {
	createdAt := time.Now()

	var analisysRequest api.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&analisysRequest); err != nil {
		log.Printf("Invalid request body: %v", err)
		if err := helpers.WriteError(w, http.StatusBadRequest, api.VALIDATIONERROR, "Invalid request body"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	if strings.TrimSpace(analisysRequest.WorkId) == "" || strings.TrimSpace(analisysRequest.SubmissionId) == "" {
		if err := helpers.WriteError(w, http.StatusBadRequest, api.VALIDATIONERROR, "workId and submissionId are required"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	fileBody, err := helpers.DownloadFile(r.Context(), h.fileStoringClient, analisysRequest.FileId, w)
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		return
	}

	var file analisys_functions.File
	file.Data = fileBody

	file.Name, file.Size, file.ContentType, err = helpers.GetFileInfo(r.Context(), h.fileStoringClient, analisysRequest.FileId, w)
	if err != nil {
		log.Printf("Failed to get file info: %v", err)
		return
	}

	chunks, err := analisys_functions.SplitFileIntoChunks(file)
	if err != nil {
		log.Printf("Failed to split file into chunks: %v", err)
		return
	}

	resp, err := h.embeddingClient.EmbedWithResponse(r.Context(), embedding.EmbedRequest{
		Chunks: chunks,
	})

	if err != nil || resp.StatusCode() != http.StatusOK {
		log.Printf("Failed to get embeddings: %v", err)
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to get embeddings"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	if resp.JSON500 != nil {
		log.Printf("Embedding service error: %v", resp.JSON500.Message)
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Embedding service error"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	if resp.JSON400 != nil {
		log.Printf("Embedding service bad request: %v", resp.JSON400.Message)
		if err := helpers.WriteError(w, http.StatusBadRequest, api.VALIDATIONERROR, "Embedding service bad request"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	if resp.JSON200 == nil || len(resp.JSON200.Embeddings) == 0 {
		log.Printf("Embedding service returned empty embeddings")
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Embedding service returned empty embeddings"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	matchedBySubmission := make(map[string]*submissionMatchAggregate)
	totalScore := 0.0
	for _, emb := range resp.JSON200.Embeddings {
		vector := make([]float32, len(emb.Embedding))
		for i, v := range emb.Embedding {
			vector[i] = float32(v)
		}

		matches, err := helpers.FindClosestFilesByWorkID(r.Context(), h.qdrantClient, analisysRequest.WorkId, vector, closestVectorLimit)
		if err != nil {
			log.Printf("Failed to search in qdrant: %v", err)
			if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to search in qdrant"); err != nil {
				log.Printf("Failed to write error response: %v", err)
			}
			return
		}

		bestMatchSubmission := ""
		bestScore := 0.0
		for _, match := range matches {
			submissionID, ok := payloadString(match.Payload, "submissionId")
			if ok && submissionID == analisysRequest.SubmissionId {
				continue
			}
			score := float64(match.Score)
			if score > bestScore {
				bestScore = score
				bestMatchSubmission = submissionID
			}
		}
		if bestMatchSubmission != "" {
			aggregate := matchedBySubmission[bestMatchSubmission]
			if aggregate == nil {
				aggregate = &submissionMatchAggregate{}
				matchedBySubmission[bestMatchSubmission] = aggregate
			}
			aggregate.matchedChunks++
			aggregate.totalScore += bestScore
		}
		totalScore += bestScore

		payload := map[string]interface{}{
			"submissionId": analisysRequest.SubmissionId,
			"chunkId":      emb.ChunkId,
			"chunkIndex":   emb.ChunkIndex,
			"fileId":       analisysRequest.FileId.String(),
		}
		if err := helpers.AddVectorForWorkID(r.Context(), h.qdrantClient, analisysRequest.WorkId, emb.ChunkId, vector, payload); err != nil {
			log.Printf("Failed to add vector to qdrant: %v", err)
			if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to store embeddings"); err != nil {
				log.Printf("Failed to write error response: %v", err)
			}
			return
		}
	}

	similarityPercent := 0.0
	if len(resp.JSON200.Embeddings) > 0 {
		similarityPercent = (totalScore / float64(len(resp.JSON200.Embeddings))) * 100
	}
	plagiarismDetected := similarityPercent >= plagiarismPercentThreshold
	log.Printf("analysis completed: workId=%s submissionId=%s similarity=%.2f%% plagiarism=%t", analisysRequest.WorkId, analisysRequest.SubmissionId, similarityPercent, plagiarismDetected)

	matchedSubmissions := make([]api.MatchedSubmission, 0, len(matchedBySubmission))
	for submissionID, aggregate := range matchedBySubmission {
		if aggregate.matchedChunks == 0 {
			continue
		}
		averageSimilarity := (aggregate.totalScore / float64(aggregate.matchedChunks)) * 100
		if averageSimilarity < matchSimilarityThreshold {
			continue
		}
		matchedSubmissions = append(matchedSubmissions, api.MatchedSubmission{
			SubmissionId:      submissionID,
			MatchedChunks:     aggregate.matchedChunks,
			SimilarityPercent: averageSimilarity,
		})
	}
	sort.Slice(matchedSubmissions, func(i, j int) bool {
		return matchedSubmissions[i].SimilarityPercent > matchedSubmissions[j].SimilarityPercent
	})

	completedAt := time.Now()
	report := api.ReportResponse{
		SubmissionId:       analisysRequest.SubmissionId,
		WorkId:             analisysRequest.WorkId,
		Status:             api.ReportResponseStatusDONE,
		PlagiarismDetected: plagiarismDetected,
		SimilarityPercent:  float32(similarityPercent),
		CreatedAt:          createdAt,
		CompletedAt:        &completedAt,
		MatchedSubmissions: &matchedSubmissions,
	}
	if err := h.reportStore.SaveReport(r.Context(), report, matchedSubmissions); err != nil {
		log.Printf("Failed to save report: %v", err)
		if err := helpers.WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to store report"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	msg := "Analysis queued. Use GET /reports/{submissionId} to check status."
	if err := json.NewEncoder(w).Encode(api.AnalyzeResponse{
		QueuedAt:     createdAt,
		WorkId:       analisysRequest.WorkId,
		Status:       api.AnalyzeResponseStatusQUEUED,
		SubmissionId: analisysRequest.SubmissionId,
		Message:      &msg,
	}); err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

}

func payloadString(payload map[string]interface{}, key string) (string, bool) {
	if payload == nil {
		return "", false
	}
	value, ok := payload[key]
	if !ok {
		return "", false
	}
	asString, ok := value.(string)
	return asString, ok
}
