package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/embedding"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/filestoring"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/filequeue"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/qdrant"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/reports"
)

/*
// ServerInterface represents all server handler.
type ServerInterface interface {
	// Start plagiarism analysis for a file
	// (POST /analyze)
	AnalyzeFile(w http.ResponseWriter, r *http.Request)
	// Get plagiarism report
	// (GET /reports/{submissionId})
	GetReport(w http.ResponseWriter, r *http.Request, submissionId string)
	// Get all reports for a work
	// (GET /works/{workId}/reports)
	GetWorkReports(w http.ResponseWriter, r *http.Request, workId string)
}

*/

type Handler struct {
	embeddingClient   *embedding.ClientWithResponses
	fileStoringClient *filestoring.ClientWithResponses
	qdrantClient      *qdrant.Client
	reportStore       ReportStore
	queue             *filequeue.Queue
}

func NewHandler() (*Handler, error) {
	embeddingBaseURL := os.Getenv("EMBEDDING_BASE_URL")
	fileStoringBaseURL := os.Getenv("FILE_STORING_BASE_URL")

	if embeddingBaseURL == "" {
		return nil, fmt.Errorf("EMBEDDING_BASE_URL is not set")
	}
	if fileStoringBaseURL == "" {
		return nil, fmt.Errorf("FILE_STORING_BASE_URL is not set")
	}

	embeddingClient, err := embedding.NewClientWithResponses(embeddingBaseURL)
	if err != nil {
		return nil, err
	}
	fileStoringClient, err := filestoring.NewClientWithResponses(fileStoringBaseURL)
	if err != nil {
		return nil, err
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	qdrantClient, err := qdrant.NewClientFromEnv(ctx)
	if err != nil {
		return nil, err
	}

	dsn, err := reports.LoadDSNFromEnv()
	if err != nil {
		return nil, err
	}
	reportStore, err := reports.NewStore(ctx, dsn)
	if err != nil {
		return nil, err
	}
	queue := filequeue.NewQueue(filequeue.LoadConfigFromEnv())

	return &Handler{
		embeddingClient:   embeddingClient,
		fileStoringClient: fileStoringClient,
		qdrantClient:      qdrantClient,
		reportStore:       reportStore,
		queue:             queue,
	}, nil
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
