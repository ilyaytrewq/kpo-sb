package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/oapi-codegen/runtime/types"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/analisys-functions"
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/handler/helpers"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/embedding"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/filequeue"
)

const (
	closestVectorLimit         = 30
	plagiarismPercentThreshold = 50.0
	matchSimilarityThreshold   = 60.0
)

type submissionMatchAggregate struct {
	matchedChunks int
	totalScore    float64
}

// (POST /analyze)
func (h *Handler) AnalyzeFile(w http.ResponseWriter, r *http.Request) {
	queuedAt := time.Now().UTC()
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

	if h.queue == nil {
		if err := helpers.WriteError(w, http.StatusServiceUnavailable, api.SERVICEUNAVAILABLE, "Processing queue is unavailable"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	if err := h.saveQueuedReport(r.Context(), analisysRequest, queuedAt); err != nil {
		log.Printf("Failed to save queued report: %v", err)
	}

	if err := h.queue.Enqueue(filequeue.Job{
		ID: analisysRequest.SubmissionId,
		Run: func(ctx context.Context) {
			h.processAnalyze(ctx, analisysRequest, queuedAt)
		},
	}); err != nil {
		log.Printf("Failed to enqueue analysis: %v", err)
		_ = h.saveErrorReport(r.Context(), analisysRequest, queuedAt, "Analysis queue is unavailable")
		if err := helpers.WriteError(w, http.StatusServiceUnavailable, api.SERVICEUNAVAILABLE, "Analysis queue is unavailable"); err != nil {
			log.Printf("Failed to write error response: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	msg := "Analysis queued. Use GET /reports/{submissionId} to check status."
	if err := json.NewEncoder(w).Encode(api.AnalyzeResponse{
		QueuedAt:     queuedAt,
		WorkId:       analisysRequest.WorkId,
		Status:       api.AnalyzeResponseStatusQUEUED,
		SubmissionId: analisysRequest.SubmissionId,
		Message:      &msg,
	}); err != nil {
		log.Printf("Failed to write response: %v", err)
		return
	}

}

func (h *Handler) processAnalyze(ctx context.Context, req api.AnalyzeRequest, createdAt time.Time) {
	if err := h.saveProcessingReport(ctx, req, createdAt); err != nil {
		log.Printf("Failed to save processing report: %v", err)
	}

	fileBody, err := h.downloadFile(ctx, req.FileId)
	if err != nil {
		h.logAndSaveError(ctx, req, createdAt, "Failed to download file", err)
		return
	}

	fileName, fileSize, fileContentType, err := h.getFileInfo(ctx, req.FileId)
	if err != nil {
		h.logAndSaveError(ctx, req, createdAt, "Failed to get file info", err)
		return
	}

	file := analisys_functions.File{
		Name:        fileName,
		Size:        fileSize,
		ContentType: fileContentType,
		Data:        fileBody,
	}

	chunks, err := analisys_functions.SplitFileIntoChunks(file)
	if err != nil {
		h.logAndSaveError(ctx, req, createdAt, "Failed to split file into chunks", err)
		return
	}

	resp, err := h.embeddingClient.EmbedWithResponse(ctx, embedding.EmbedRequest{
		Chunks: chunks,
	})
	if err != nil || resp.StatusCode() != http.StatusOK {
		h.logAndSaveError(ctx, req, createdAt, "Failed to get embeddings", err)
		return
	}
	if resp.JSON500 != nil {
		h.logAndSaveError(ctx, req, createdAt, "Embedding service error", errors.New(resp.JSON500.Message))
		return
	}
	if resp.JSON400 != nil {
		h.logAndSaveError(ctx, req, createdAt, "Embedding service bad request", errors.New(resp.JSON400.Message))
		return
	}
	if resp.JSON200 == nil || len(resp.JSON200.Embeddings) == 0 {
		h.logAndSaveError(ctx, req, createdAt, "Embedding service returned empty embeddings", nil)
		return
	}

	matchedBySubmission := make(map[string]*submissionMatchAggregate)
	totalScore := 0.0
	for _, emb := range resp.JSON200.Embeddings {
		vector := make([]float32, len(emb.Embedding))
		for i, v := range emb.Embedding {
			vector[i] = float32(v)
		}

		matches, err := helpers.FindClosestFilesByWorkID(ctx, h.qdrantClient, req.WorkId, vector, closestVectorLimit)
		if err != nil {
			h.logAndSaveError(ctx, req, createdAt, "Failed to search in qdrant", err)
			return
		}

		bestMatchSubmission := ""
		bestScore := 0.0
		for _, match := range matches {
			submissionID, ok := payloadString(match.Payload, "submissionId")
			if ok && submissionID == req.SubmissionId {
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
			"submissionId": req.SubmissionId,
			"chunkId":      emb.ChunkId,
			"chunkIndex":   emb.ChunkIndex,
			"fileId":       req.FileId.String(),
		}
		if err := helpers.AddVectorForWorkID(ctx, h.qdrantClient, req.WorkId, emb.ChunkId, vector, payload); err != nil {
			h.logAndSaveError(ctx, req, createdAt, "Failed to store embeddings", err)
			return
		}
	}

	similarityPercent := 0.0
	if len(resp.JSON200.Embeddings) > 0 {
		similarityPercent = (totalScore / float64(len(resp.JSON200.Embeddings))) * 100
	}
	plagiarismDetected := similarityPercent >= plagiarismPercentThreshold
	log.Printf("analysis completed: workId=%s submissionId=%s similarity=%.2f%% plagiarism=%t", req.WorkId, req.SubmissionId, similarityPercent, plagiarismDetected)

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

	completedAt := time.Now().UTC()
	report := api.ReportResponse{
		SubmissionId:       req.SubmissionId,
		WorkId:             req.WorkId,
		Status:             api.ReportResponseStatusDONE,
		PlagiarismDetected: plagiarismDetected,
		SimilarityPercent:  float32(similarityPercent),
		CreatedAt:          createdAt,
		CompletedAt:        &completedAt,
		MatchedSubmissions: &matchedSubmissions,
	}
	if err := h.reportStore.SaveReport(ctx, report, matchedSubmissions); err != nil {
		log.Printf("Failed to save report: %v", err)
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

func (h *Handler) downloadFile(ctx context.Context, fileId types.UUID) ([]byte, error) {
	resp, err := h.fileStoringClient.DownloadFileWithResponse(ctx, fileId, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		if resp.JSON404 != nil {
			return nil, fmt.Errorf("file not found: %s", resp.JSON404.Message)
		}
		if resp.JSON500 != nil {
			return nil, fmt.Errorf("file storing error: %s", resp.JSON500.Message)
		}
		return nil, fmt.Errorf("unexpected file-storing response: %s", resp.Status())
	}
	return resp.Body, nil
}

func (h *Handler) getFileInfo(ctx context.Context, fileId types.UUID) (*string, int64, *string, error) {
	resp, err := h.fileStoringClient.GetFileInfoWithResponse(ctx, fileId, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		if resp.JSON404 != nil {
			return nil, 0, nil, fmt.Errorf("file not found: %s", resp.JSON404.Message)
		}
		if resp.JSON500 != nil {
			return nil, 0, nil, fmt.Errorf("file storing error: %s", resp.JSON500.Message)
		}
		return nil, 0, nil, fmt.Errorf("unexpected file-storing response: %s", resp.Status())
	}
	if resp.JSON200 == nil {
		return nil, 0, nil, fmt.Errorf("empty file info response")
	}
	return resp.JSON200.OriginalFileName, *resp.JSON200.SizeBytes, resp.JSON200.ContentType, nil
}

func (h *Handler) saveQueuedReport(ctx context.Context, req api.AnalyzeRequest, createdAt time.Time) error {
	report := api.ReportResponse{
		SubmissionId:       req.SubmissionId,
		WorkId:             req.WorkId,
		Status:             api.ReportResponseStatusQUEUED,
		PlagiarismDetected: false,
		SimilarityPercent:  0,
		CreatedAt:          createdAt,
	}
	return h.reportStore.SaveReport(ctx, report, nil)
}

func (h *Handler) saveProcessingReport(ctx context.Context, req api.AnalyzeRequest, createdAt time.Time) error {
	report := api.ReportResponse{
		SubmissionId:       req.SubmissionId,
		WorkId:             req.WorkId,
		Status:             api.ReportResponseStatusPROCESSING,
		PlagiarismDetected: false,
		SimilarityPercent:  0,
		CreatedAt:          createdAt,
	}
	return h.reportStore.SaveReport(ctx, report, nil)
}

func (h *Handler) saveErrorReport(ctx context.Context, req api.AnalyzeRequest, createdAt time.Time, msg string) error {
	completedAt := time.Now().UTC()
	report := api.ReportResponse{
		SubmissionId:       req.SubmissionId,
		WorkId:             req.WorkId,
		Status:             api.ReportResponseStatusERROR,
		PlagiarismDetected: false,
		SimilarityPercent:  0,
		CreatedAt:          createdAt,
		CompletedAt:        &completedAt,
		ErrorMessage:       &msg,
	}
	return h.reportStore.SaveReport(ctx, report, nil)
}

func (h *Handler) logAndSaveError(ctx context.Context, req api.AnalyzeRequest, createdAt time.Time, msg string, err error) {
	if err != nil {
		log.Printf("%s: %v", msg, err)
	} else {
		log.Printf("%s", msg)
	}
	if saveErr := h.saveErrorReport(ctx, req, createdAt, msg); saveErr != nil {
		log.Printf("Failed to save error report: %v", saveErr)
	}
}
