package handlers

import (
	"net/http"
	"encoding/json"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) GetFileInfo(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID, params api.GetFileInfoParams) {
	info, err := h.service.Head(r.Context(), h.service.Config.Bucket, fileId.String())
	if err != nil {
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to get file info")
		return
	}
	/*
	type Info struct {
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
}
	*/
	response := api.FileInfoResponse{
		FileId:      fileId,
		ContentType: &info.ContentType,
		SizeBytes:        &info.Size,
		StoredAt:  info.StoredAt,
		OriginalFileName: &info.FileName,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode response")
		return
	}
}