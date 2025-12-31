package yandexembd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type embReq struct {
	Model          string `json:"model"`
	Input          string `json:"input"`
	EncodingFormat string `json:"encoding_format,omitempty"`
}

type embResp struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func (c *Client) EmbedText(ctx context.Context, modelURI, text string) ([]float64, error) {
	reqBody := embReq{
		Model:          modelURI,
		Input:          text,
		EncodingFormat: "float",
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL + "/embeddings"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("OpenAI-Project", c.FolderID)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Yandex embeddings API error: status=%d body=%s", resp.StatusCode, string(raw))
		return nil, fmt.Errorf("yandex embeddings: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var out embResp
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if len(out.Data) == 0 {
		return nil, fmt.Errorf("yandex embeddings: empty data")
	}
	return out.Data[0].Embedding, nil
}
