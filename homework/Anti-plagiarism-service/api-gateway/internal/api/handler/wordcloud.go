package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode"

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

	removeStopwords := false
	if params.RemoveStopwords != nil {
		removeStopwords = *params.RemoveStopwords
	}

	language := "en"
	if params.Language != nil && *params.Language != "" {
		language = *params.Language
	}

	minWordLength := 3
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
		Format          string  `json:"format"`
		Width           int     `json:"width"`
		Height          int     `json:"height"`
		BackgroundColor string  `json:"backgroundColor,omitempty"`
		UseWordList     bool    `json:"useWordList"`
		Words           [][]any `json:"words"`
		Scale           string  `json:"scale,omitempty"`
		FontScale       int     `json:"fontScale,omitempty"`
	}

	payload := qcReq{
		Format:          format,
		Width:           width,
		Height:          height,
		BackgroundColor: "white",
		UseWordList:     true,
		Words:           buildWordList(txt),
		Scale:           "linear",
		FontScale:       18,
	}

	log.Printf("Params: %+v", payload)

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

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = "wordcloud service returned non-image response"
		}
		writeJSONError(w, http.StatusBadGateway, msg)
		return
	}

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

func buildWordList(text string) [][]any {
	m := make(map[string]int)
	var cur []rune

	flush := func() {
		if len(cur) == 0 {
			return
		}
		w := strings.ToLower(string(cur))
		cur = cur[:0]
		if len([]rune(w)) < 2 {
			return
		}
		m[w]++
	}

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur = append(cur, r)
		} else {
			flush()
		}
	}
	flush()

	type kv struct {
		w string
		c int
	}
	arr := make([]kv, 0, len(m))
	for w, c := range m {
		arr = append(arr, kv{w: w, c: c})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].c != arr[j].c {
			return arr[i].c > arr[j].c
		}
		return arr[i].w < arr[j].w
	})

	if len(arr) > 200 {
		arr = arr[:200]
	}

	out := make([][]any, 0, len(arr))
	for _, x := range arr {
		weight := x.c
		if weight == 1 {
			weight = 3
		}
		out = append(out, []any{x.w, weight})
	}
	return out
}
