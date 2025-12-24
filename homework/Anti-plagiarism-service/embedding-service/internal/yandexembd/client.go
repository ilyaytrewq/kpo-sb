package yandexembd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	BaseURL  string
	APIKey   string
	FolderID string
	HTTP     *http.Client
}

func NewClient() (*Client, error) {
	APIKey := os.Getenv("API_KEY")
	if APIKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable is not set")
	}

	FolderID := os.Getenv("FOLDER_ID")
	if FolderID == "" {
		return nil, fmt.Errorf("FOLDER_ID environment variable is not set")
	}

	BaseURL := os.Getenv("URL")
	if BaseURL == "" {
		return nil, fmt.Errorf("URL environment variable is not set")
	}

	log.Printf("Yandex Embeddings Client initialized with FolderID: %s, BaseURL: %s", FolderID, BaseURL)

	return &Client{
		BaseURL:  BaseURL,
		APIKey:   APIKey,
		FolderID: FolderID,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}
