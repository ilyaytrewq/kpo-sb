package helpers

import (
	"context"
	"net/http"
	"fmt"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/clients/filestoring"
	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/api/generated"
)

func GetFileInfo(ctx context.Context, fileStoringClient *filestoring.ClientWithResponses, fileId openapi_types.UUID, w http.ResponseWriter) (*string, int64, *string, error) {

	getFileInfoResponse, err := fileStoringClient.GetFileInfoWithResponse(ctx, fileId, nil)
	if err != nil || getFileInfoResponse.StatusCode() != http.StatusOK {
		WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to get file info")
		return nil, 0, nil, fmt.Errorf("failed to get file info: %v", err)
	}

	if getFileInfoResponse.JSON404 != nil {
		WriteError(w, http.StatusNotFound, api.FILENOTFOUND, "File not found")
		return nil, 0, nil, fmt.Errorf("file not found: %v", getFileInfoResponse.JSON404.Message)
	}

	if getFileInfoResponse.JSON500 != nil {
		WriteError(w, http.StatusInternalServerError, api.INTERNALERROR, "File storing service error")
		return nil, 0, nil, fmt.Errorf("file storing service error: %v", getFileInfoResponse.JSON500.Message)
	}

	return getFileInfoResponse.JSON200.OriginalFileName, *getFileInfoResponse.JSON200.SizeBytes, getFileInfoResponse.JSON200.ContentType, nil
}