package handler

import "log"

type reportStoreCloser interface {
	Close()
}

func (h *Handler) Close() {
	if h == nil {
		return
	}
	if h.queue != nil {
		h.queue.Close()
	}
	if h.qdrantClient != nil {
		if err := h.qdrantClient.Close(); err != nil {
			log.Printf("failed to close qdrant client: %v", err)
		}
	}
	if closer, ok := h.reportStore.(reportStoreCloser); ok {
		closer.Close()
	}
}
