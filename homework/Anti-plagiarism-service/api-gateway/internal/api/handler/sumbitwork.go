package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/google/uuid"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	fileanalysis "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/fileanalysis"
	filestoring "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/clients/filestoring"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/store"
)

func (h *Handler) SubmitWork(w http.ResponseWriter, r *http.Request, workId string) {
	ctx := r.Context()
	reqID := r.Header.Get("X-Request-Id")
	if reqID == "" {
		reqID = uuid.NewString()
	}

	if strings.TrimSpace(workId) == "" {
		if err := writeError(w, http.StatusBadRequest, api.VALIDATIONERROR, "workId is required"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if _, err := h.store.GetWork(ctx, workId); err != nil {
		if err == store.ErrWorkNotFound {
			if err := writeError(w, http.StatusNotFound, api.WORKNOTFOUND, "Work with this workId not found"); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		log.Printf("[Error] Failed to fetch work: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Internal server error"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
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

	uploadedAt := time.Now().UTC()
	meta := map[string]interface{}{
		"workId":           workId,
		"originalFileName": fh.Filename,
		"contentType":      ct,
		"submittedAt":      uploadedAt.Format(time.RFC3339),
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

	submissionID := uuid.NewString()
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
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "File storage is unavailable"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if resp.JSON200 == nil {
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
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, resp.JSON500.Message); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		log.Printf("unexpected file-storing response, request_id: %s, status: %s", reqID, resp.Status())
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Unexpected response from file storage"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	fileID := resp.JSON200.FileId
	analysisResp, err := h.fileAnalysisClient.AnalyzeFileWithResponse(ctx, fileanalysis.AnalyzeRequest{
		FileId:       fileID,
		WorkId:       workId,
		SubmissionId: submissionID,
	})
	if err != nil {
		log.Printf("[Error] file-analysis request failed, request_id: %s, err: %v", reqID, err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "File analysis service is unavailable"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}
	if analysisResp.JSON202 == nil {
		if analysisResp.JSON400 != nil {
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, analysisResp.JSON400.Message); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		if analysisResp.JSON404 != nil {
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, analysisResp.JSON404.Message); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		if analysisResp.JSON500 != nil {
			if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, analysisResp.JSON500.Message); err != nil {
				log.Printf("[Error] Failed write error response: %v", err)
			}
			return
		}
		log.Printf("unexpected file-analysis response, request_id: %s, status: %s", reqID, analysisResp.Status())
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Unexpected response from file analysis"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	if err := h.store.CreateSubmission(ctx, store.Submission{
		SubmissionID: submissionID,
		WorkID:       workId,
		FileID:       fileID,
		UploadedAt:   uploadedAt,
	}); err != nil {
		log.Printf("[Error] Failed to store submission: %v", err)
		if err := writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to store submission"); err != nil {
			log.Printf("[Error] Failed write error response: %v", err)
		}
		return
	}

	msg := "Submission accepted. Plagiarism check is queued. Use GET /works/{workId}/reports or GET /submissions/{submissionId} to check results."
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(api.SubmissionUploadResponse{
		SubmissionId: submissionID,
		WorkId:       workId,
		FileId:       fileID,
		Status:       api.SubmissionUploadResponseStatusQUEUED,
		UploadedAt:   uploadedAt,
		Message:      &msg,
	}); err != nil {
		log.Printf("[Error] Failed to encode response: %v", err)
	}
}
