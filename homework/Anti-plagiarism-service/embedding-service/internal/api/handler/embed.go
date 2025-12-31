package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"unicode/utf8"

	api "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/embedding-service/internal/api/generated"
)

func (h *handler) Embed(w http.ResponseWriter, r *http.Request) {
	var embedRequest api.EmbedRequest
	if err := json.NewDecoder(r.Body).Decode(&embedRequest); err != nil {
		log.Printf("Failed to decode embed request: %v", err)
		return
	}

	minDim := 1 << 30
	maxDim := 0
	var embedResponse api.EmbedResult
	for _, chunk := range embedRequest.Chunks {
		var embedding api.ChunkEmbedding
		embedding.ChunkId = chunk.ChunkId
		if len(chunk.Text) > maxBytes {
			log.Printf("Chunk text exceeds maximum allowed size: %d bytes", len(chunk.Text))
			writeError(w, http.StatusBadRequest, api.BADREQUEST, "Chunk text exceeds maximum allowed size")
			return
		}
		if len(chunk.Text) == 0 {
			log.Printf("Chunk text is empty")
			writeError(w, http.StatusBadRequest, api.BADREQUEST, "Chunk text is empty")
			return
		}
		if !utf8.ValidString(chunk.Text) {
			log.Printf("Chunk text is not valid UTF-8: %s", chunk.Text)
			writeError(w, http.StatusBadRequest, api.BADREQUEST, "Chunk text is not valid UTF-8")
			return
		}
		embeddings, err := h.client.EmbedText(r.Context(), h.model, chunk.Text)
		if err != nil {
			log.Printf("Failed to get embedding from Yandex API: %v", err)
			writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to get embedding from external service")
			return
		}
		embedding.Dimension = len(embeddings)
		if embedding.Dimension < minDim {
			minDim = embedding.Dimension
		}
		if embedding.Dimension > maxDim {
			maxDim = embedding.Dimension
		}
		embedding.Embedding = embeddings
		embedResponse.Embeddings = append(embedResponse.Embeddings, embedding)
	}
	if minDim != maxDim {
		log.Printf("Inconsistent embedding dimensions: min=%d max=%d", minDim, maxDim)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Inconsistent embedding dimensions")
		return
	}
	embedResponse.Dimension = maxDim
	embedResponse.Model = h.model

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(embedResponse); err != nil {
		log.Printf("Failed to encode embed response: %v", err)
		writeError(w, http.StatusInternalServerError, api.INTERNALERROR, "Failed to encode embed response")
		return
	}
}
