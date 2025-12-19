package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	fileanalysis "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/fileanalysis"
	filestoring "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/filestoring"
	config "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/config"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/db"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

const (
	maxLength = 64 << 20
)

/*
// ServerInterface represents all server handler.
type ServerInterface interface {
	// Get submission details
	// (GET /submissions/{submissionId})
	GetSubmission(w http.ResponseWriter, r *http.Request, submissionId string)
	// Create a work
	// (POST /works)
	CreateWork(w http.ResponseWriter, r *http.Request)
	// Get analytics by workId
	// (GET /works/{workId}/reports)
	ListWorkReports(w http.ResponseWriter, r *http.Request, workId string)
	// Get statistics by workId
	// (GET /works/{workId}/stats)
	GetWorkStats(w http.ResponseWriter, r *http.Request, workId string)
	// Submit work for review
	// (POST /works/{workId}/submissions)
	SubmitWork(w http.ResponseWriter, r *http.Request, workId string)
}
*/

type Handler struct {
	store              *store.Store
	fileStoringClient  *filestoring.ClientWithResponses
	fileAnalysisClient *fileanalysis.ClientWithResponses
}

func NewHandler() (*Handler, error) {
	ctx := context.Background()
	httpClient := &http.Client{
		Timeout: 3 * time.Second,
	}

	fileStoringConfig, err := config.LoadFileStoringConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load file storing config: %w", err)
	}

	fileStoringClient, err := filestoring.NewClientWithResponses(fileStoringConfig.Url, filestoring.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create file storing client: %w", err)
	}

	fileAnalysisConfig, err := config.LoadFileAnalysisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load file analysis config: %w", err)
	}

	fileAnalysisClient, err := fileanalysis.NewClientWithResponses(fileAnalysisConfig.Url, fileanalysis.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create file analysis client: %w", err)
	}

	dsn, err := config.LoadGatewayDSNFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load gateway db config: %w", err)
	}

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway db: %w", err)
	}

	gatewayStore := store.New(pool)
	if err := gatewayStore.EnsureSchema(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ensure gateway schema: %w", err)
	}

	return &Handler{
		store:              gatewayStore,
		fileStoringClient:  fileStoringClient,
		fileAnalysisClient: fileAnalysisClient,
	}, nil
}

func writeError(w http.ResponseWriter, status int, code api.ErrorCode, msg string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(api.ErrorResponse{
		Code:    code,
		Message: msg,
	}); err != nil {
		return err
	}

	return nil
}
