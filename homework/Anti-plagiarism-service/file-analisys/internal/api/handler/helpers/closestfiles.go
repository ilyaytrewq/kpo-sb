package helpers

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"

	qdrantclient "github.com/ilyaytrewq/kpo-sb/anti-plagiarism-service/file-analisys/internal/qdrant"
)

const (
	searchLimitFactor  = 5
	maxSearchLimitSize = 500
	maxClosestResults  = 10
	maxCollectionName  = 50
)

func FindClosestFilesByWorkID(
	ctx context.Context,
	qdrantClient *qdrantclient.Client,
	workID string,
	vector []float32,
	limit int,
) ([]qdrantclient.SearchResult, error) {
	if qdrantClient == nil {
		return nil, errors.New("qdrant client is nil")
	}
	if workID == "" {
		return nil, errors.New("workID is empty")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}
	if limit > maxClosestResults {
		limit = maxClosestResults
	}

	collectionName, err := collectionNameForWorkID(workID)
	if err != nil {
		return nil, err
	}
	if err := qdrantClient.EnsureCollectionNamed(ctx, collectionName); err != nil {
		return nil, fmt.Errorf("ensure qdrant collection: %w", err)
	}

	searchLimit := limit * searchLimitFactor
	if searchLimit > maxSearchLimitSize {
		searchLimit = maxSearchLimitSize
	}
	if searchLimit < limit {
		searchLimit = limit
	}

	results, err := qdrantClient.SearchInCollection(ctx, collectionName, vector, searchLimit)
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func AddVectorForWorkID(
	ctx context.Context,
	qdrantClient *qdrantclient.Client,
	workID string,
	id string,
	vector []float32,
	payload map[string]interface{},
) error {
	if qdrantClient == nil {
		return errors.New("qdrant client is nil")
	}
	if workID == "" {
		return errors.New("workID is empty")
	}

	collectionName, err := collectionNameForWorkID(workID)
	if err != nil {
		return err
	}
	if err := qdrantClient.EnsureCollectionNamed(ctx, collectionName); err != nil {
		return fmt.Errorf("ensure qdrant collection: %w", err)
	}

	if err := qdrantClient.AddVectorToCollection(ctx, collectionName, id, vector, payload); err != nil {
		return fmt.Errorf("qdrant add vector: %w", err)
	}
	return nil
}

func collectionNameForWorkID(workID string) (string, error) {
	trimmed := strings.TrimSpace(workID)
	if trimmed == "" {
		return "", errors.New("workID is empty")
	}

	sanitized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '-':
			return r
		default:
			return '_'
		}
	}, trimmed)
	if len(sanitized) > maxCollectionName {
		sanitized = sanitized[:maxCollectionName]
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(trimmed))
	suffix := fmt.Sprintf("%08x", hasher.Sum32())

	return fmt.Sprintf("work_%s_%s", sanitized, suffix), nil
}
