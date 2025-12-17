package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
)

/*
 func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
    // 1. Parse the multipart form (don't use json.Decode here!)
    err := r.ParseMultipartForm(32 << 20) // 32MB max memory
    if err != nil {
        // Send BAD_REQUEST
        return
    }

    // 2. Get the metadata part
    metaPart := r.FormValue("metadata")
    var metadata struct {
        WorkId string `json:"workId"`
        // ... other fields
    }

    // 3. ONLY decode the specific metadata string
    if err := json.Unmarshal([]byte(metaPart), &metadata); err != nil {
        // This is where a JSON error should be caught if the JSON itself is bad
        return
    }

    // 4. Get the file part
    file, header, err := r.FormFile("file")
    // ... process file
}
*/

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
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file from form data: %v", err)
		writeError(w, http.StatusBadRequest, api.BADREQUEST, "Failed to get file from form data")
		return
	}
	defer file.Close()

	var reader io.Reader = file
	log.Printf("Uploading file: Bucket=%s, WorkID=%s, OriginalFileName=%s, ContentType=%s, Size=%d", h.service.Config.Bucket, uploadRequest.Metadata.WorkId, *uploadRequest.Metadata.OriginalFileName, *uploadRequest.Metadata.ContentType, fileHeader.Size)

	if err := h.service.Upload(r.Context(), h.service.Config.Bucket, uploadRequest.Metadata.WorkId, *uploadRequest.Metadata.ContentType, *uploadRequest.Metadata.OriginalFileName, reader, fileHeader.Size); err != nil {
		log.Printf("Failed to upload file: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to upload file")
		return
	}
}
