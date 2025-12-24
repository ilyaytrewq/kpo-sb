package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/generated"
	"github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/api-gateway/internal/api/handler/text_extract"
)

func (h *Handler) BuildWordCloud(w http.ResponseWriter, r *http.Request, params api.BuildWordCloudParams) {
	format := "png"
	if params.Format != nil && *params.Format != "" {
		format = strings.ToLower(string(*params.Format))
	}
	if format != "png" && format != "svg" {
		writeJSONError(w, http.StatusBadRequest, "invalid format (allowed: png, svg)")
		return
	}

	width := 1000
	if params.Width != nil && *params.Width > 0 {
		width = *params.Width
	}
	height := 1000
	if params.Height != nil && *params.Height > 0 {
		height = *params.Height
	}

	removeStopwords := true
	if params.RemoveStopwords != nil {
		removeStopwords = *params.RemoveStopwords
	}

	language := "ru"
	if params.Language != nil && *params.Language != "" {
		language = *params.Language
	}

	minWordLength := 4
	if params.MinWordLength != nil && *params.MinWordLength > 0 {
		minWordLength = *params.MinWordLength
	}

	const maxUploadBytes int64 = 20 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	_, fh, err := r.FormFile("file")
	if err != nil || fh == nil {
		writeJSONError(w, http.StatusBadRequest, "missing file field")
		return
	}

	txt, err := text_extract.ExtractTextFromMultipart(r.Context(), fh, text_extract.ExtractOptions{MaxBytes: maxUploadBytes})
	if err != nil {
		switch {
		case errors.Is(err, text_extract.ErrTooLarge):
			writeJSONError(w, http.StatusBadRequest, "file is too large")
		case errors.Is(err, text_extract.ErrUnsupportedFormat):
			writeJSONError(w, http.StatusUnprocessableEntity, "unsupported file type")
		case errors.Is(err, text_extract.ErrEmptyText):
			writeJSONError(w, http.StatusUnprocessableEntity, "extracted text is empty")
		default:
			writeJSONError(w, http.StatusUnprocessableEntity, "cannot extract text")
		}
		return
	}

	type qcReq struct {
		Format          string `json:"format"`
		Width           int    `json:"width"`
		Height          int    `json:"height"`
		RemoveStopwords bool   `json:"removeStopwords"`
		Language        string `json:"language"`
		MinWordLength   int    `json:"minWordLength"`
		Text            string `json:"text"`
	}

	payload := qcReq{
		Format:          format,
		Width:           width,
		Height:          height,
		RemoveStopwords: removeStopwords,
		Language:        language,
		MinWordLength:   minWordLength,
		Text:            txt,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://quickchart.io/wordcloud", bytes.NewReader(b))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, "wordcloud service unavailable")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = "wordcloud service error"
		}
		writeJSONError(w, http.StatusBadGateway, msg)
		return
	}

	if format == "png" {
		w.Header().Set("Content-Type", "image/png")
	} else {
		w.Header().Set("Content-Type", "image/svg+xml")
	}
	w.Header().Set("Cache-Control", "public, max-age=300")

	_, _ = io.Copy(w, resp.Body)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
