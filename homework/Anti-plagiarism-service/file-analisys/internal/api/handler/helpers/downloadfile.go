package helpers

import (
	"context"
	"net/http"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/filestoring"
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)

func DownloadFile(ctx context.Context, fileStoringClient *filestoring.ClientWithResponses, fileId openapi_types.UUID, w http.ResponseWriter) ([]byte, error) {
	downloadFileResponse, err := fileStoringClient.DownloadFileWithResponse(ctx, fileId, nil)
	if err != nil || downloadFileResponse.StatusCode() != http.StatusOK {
		WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to download file")
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	if downloadFileResponse.JSON404 != nil {
		WriteError(w, http.StatusNotFound, api.FILENOTFOUND, "File not found")
		return nil, fmt.Errorf("file not found: %v", downloadFileResponse.JSON404.Message)
	}

	if downloadFileResponse.JSON500 != nil {
		WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "File storing service error")
		return nil, fmt.Errorf("file storing service error: %v", downloadFileResponse.JSON500.Message)
	}

	return downloadFileResponse.Body, nil
}


