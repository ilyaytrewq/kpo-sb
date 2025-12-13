package Handler

import (
	"encoding/json"
	"log"
	"net/http"
	"bytes"
	"mime/multipart"
	"io"
	"net/textproto"
	"github.com/google/uuid"
	"time"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"

	filestoring "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/filestoring"

)
func (h *Handler) SubmitWork(w http.ResponseWriter, r *http.Request, workId string) {
	ctx := r.Context()
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.NewString()
	}

	if err := r.ParseMultipartForm(maxLength); err != nil {
		log.Println("[Error] parse multipart failed", "request_id", reqID, "err", err)
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "Expected multipart/form-data"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	f, fh, err := r.FormFile("file")
	if err != nil {
		log.Printf("[Error] missing file part, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "Missing form field 'file'"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	defer func() { _ = f.Close() }()

	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	filePart, err := mw.CreateFormFile("file", fh.Filename)
	if err != nil {
		log.Printf("[Error] create file part failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to build upload request"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	if _, err := io.Copy(filePart, f); err != nil {
		log.Printf("[Error] copy file failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to read uploaded file"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	meta := map[string]interface{}{
		"workId":           workId,
		"originalFileName": fh.Filename,
		"contentType":      ct,
		"submittedAt":      time.Now().UTC().Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)

	metaHeader := textproto.MIMEHeader{}
	metaHeader.Set("Content-Disposition", `form-data; name="metadata"`)
	metaHeader.Set("Content-Type", "application/json")
	metaPart, err := mw.CreatePart(metaHeader)
	if err != nil {
		log.Printf("[Error] create metadata part failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to build upload metadata"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	if _, err := metaPart.Write(metaBytes); err != nil {
		log.Printf("[Error] write metadata failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to build upload metadata"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if err := mw.Close(); err != nil {
		log.Printf("[Error] close multipart writer failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to finalize upload request"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	params := &filestoring.UploadFileParams{
		XRequestId: &reqID,
	}

	resp, err := h.fileStoringClient.UploadFileWithBodyWithResponse(
		ctx,
		params,
		mw.FormDataContentType(),
		bytes.NewReader(buf.Bytes()),
	)
	if err != nil {
		log.Printf("[Error] file-storing request failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusBadGateway, api.INTERNALERROR, "File storage is unavailable"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if resp.JSON200 != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp.JSON200); err != nil {
			log.Printf("[Error] Failed to encode response: %v", err)
		}
		return
	}

	if resp.JSON400 != nil {
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, resp.JSON400.Message); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	if resp.JSON413 != nil {
		if err := writeError(w, http.StatusRequestEntityTooLarge, api.VALIDATIONERROR, resp.JSON413.Message); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	if resp.JSON500 != nil {
		if err := writeError(w, http.StatusBadGateway, api.INTERNALERROR, resp.JSON500.Message); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	log.Printf("unexpected file-storing response, request_id: %s, status: %s", reqID, resp.Status())
	if err := writeError(w, http.StatusBadGateway, api.INTERNALERROR, "Unexpected response from file storage"); err != nil {
		log.Printf("[Error] Failed write error response: %v", err)
	}
}
