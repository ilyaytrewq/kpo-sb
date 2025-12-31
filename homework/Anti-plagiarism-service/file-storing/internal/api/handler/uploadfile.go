package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request, params api.UploadFileParams) {
	var uploadRequest api.UploadFileRequest
	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		writeError(w, http.StatusBadRequest, api.BADREQUEST, "Failed to parse multipart form")
		return
	}
	meta := r.FormValue("metadata")
	if err := json.Unmarshal([]byte(meta), &uploadRequest.Metadata); err != nil {
		log.Printf("Invalid metadata JSON: %v", err)
		writeError(w, http.StatusBadRequest, api.BADREQUEST, "Invalid metadata JSON")
		return
	}
	if strings.TrimSpace(uploadRequest.Metadata.WorkId) == "" {
		writeError(w, http.StatusBadRequest, api.BADREQUEST, "workId is required")
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form data: %v", err)
		writeError(w, http.StatusBadRequest, api.BADREQUEST, "Failed to get file from form data")
		return
	}
	defer file.Close()

	contentType := "application/octet-stream"
	if uploadRequest.Metadata.ContentType != nil && strings.TrimSpace(*uploadRequest.Metadata.ContentType) != "" {
		contentType = *uploadRequest.Metadata.ContentType
	}
	originalFileName := fileHeader.Filename
	if uploadRequest.Metadata.OriginalFileName != nil && strings.TrimSpace(*uploadRequest.Metadata.OriginalFileName) != "" {
		originalFileName = *uploadRequest.Metadata.OriginalFileName
	}

	var reader io.Reader = file
	fileID := uuid.New()
	bucket := h.service.Bucket()
	log.Printf("Uploading file: Bucket=%s, FileID=%s, WorkID=%s, OriginalFileName=%s, ContentType=%s, Size=%d", bucket, fileID.String(), uploadRequest.Metadata.WorkId, originalFileName, contentType, fileHeader.Size)

	if err := h.service.Upload(r.Context(), bucket, fileID.String(), contentType, originalFileName, reader, fileHeader.Size); err != nil {
		log.Printf("Failed to upload file: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to upload file")
		return
	}

	storedAt := time.Now().UTC()
	size := fileHeader.Size
	result := api.UploadFileResult{
		FileId:           fileID,
		StoredAt:         storedAt,
		SizeBytes:        &size,
		ContentType:      &contentType,
		OriginalFileName: &originalFileName,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("Failed to encode upload response: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode response")
		return
	}
}
