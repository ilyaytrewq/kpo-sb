package handler

import (
	"io"
	"log"
	"net/http"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-storing/internal/api/generated"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID, params api.DownloadFileParams) {
	log.Printf("Downloading file with ID: %s", fileId.String())
	fileReader, contentType, err := h.service.Download(r.Context(), h.service.Bucket(), fileId.String())
	if err != nil {
		log.Printf("Failed to download file: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to download file")
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	bytes, err := io.ReadAll(fileReader)
	if err != nil {
		log.Printf("Failed to read file: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to read file")
		return
	}
	if n, err := w.Write(bytes); err != nil {
		log.Printf("Failed to write file to response: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to write file to response")
		return
	} else if n != len(bytes) {
		log.Printf("Incomplete write: wrote %d of %d bytes", n, len(bytes))
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to write complete file to response")
		return
	}

}
