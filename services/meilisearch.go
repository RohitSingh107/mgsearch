package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"mgsearch/config"
	"mgsearch/models"
	"strings"

	meilisearch "github.com/meilisearch/meilisearch-go"
)

type MeilisearchService struct {
	client meilisearch.ServiceManager
}

// NewMeilisearchService creates a new Meilisearch service instance backed by the official SDK
func NewMeilisearchService(cfg *config.Config) *MeilisearchService {
	client := meilisearch.New(
		cfg.MeilisearchURL,
		meilisearch.WithAPIKey(cfg.MeilisearchAPIKey),
	)

	return &MeilisearchService{
		client: client,
	}
}

// Search performs a search request to Meilisearch
// indexName: the name of the index to search (e.g., "test_index")
// request: the search request body (can contain any Meilisearch parameters)
func (s *MeilisearchService) Search(indexName string, request *models.SearchRequest) (*models.SearchResponse, error) {
	searchRequest, err := toSDKSearchRequest(request)
	if err != nil {
		return nil, err
	}

	index := s.client.Index(indexName)
	searchResponse, err := index.Search("", searchRequest)
	if err != nil {
		return nil, fmt.Errorf("meilisearch search failed: %w", err)
	}

	// Convert SDK response back into a flexible map for handlers
	raw, err := json.Marshal(searchResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search response: %w", err)
	}

	var response models.SearchResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search response: %w", err)
	}

	return &response, nil
}

// IndexDocument indexes a single document into the specified Meilisearch index.
// The document is wrapped in an array to comply with Meilisearch's bulk indexing API.
func (s *MeilisearchService) IndexDocument(indexName string, document models.Document) (*models.IndexDocumentResponse, error) {
	index := s.client.Index(indexName)
	taskInfo, err := index.AddDocuments([]models.Document{document}, nil)
	if err != nil {
		return nil, fmt.Errorf("meilisearch indexing failed: %w", err)
	}

	raw, err := json.Marshal(taskInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal indexing response: %w", err)
	}

	var response models.IndexDocumentResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal indexing response: %w", err)
	}

	return &response, nil
}

// DeleteDocument removes a single document by identifier.
func (s *MeilisearchService) DeleteDocument(indexName, documentID string) error {
	if indexName == "" || documentID == "" {
		return fmt.Errorf("index name and document id are required")
	}
	index := s.client.Index(indexName)
	_, err := index.DeleteDocument(documentID)
	return err
}

// EnsureIndex creates the index if it does not already exist.
func (s *MeilisearchService) EnsureIndex(indexUID string) error {
	if indexUID == "" {
		return fmt.Errorf("index uid is required")
	}

	_, err := s.client.GetIndex(indexUID)
	if err == nil {
		return nil
	}

	var meiliErr *meilisearch.Error
	if errors.As(err, &meiliErr) {
		if meiliErr.MeilisearchApiError.Code != "index_not_found" {
			return err
		}
	} else {
		return err
	}

	_, err = s.client.CreateIndex(&meilisearch.IndexConfig{
		Uid: indexUID,
	})
	return err
}

func toSDKSearchRequest(request *models.SearchRequest) (*meilisearch.SearchRequest, error) {
	if request == nil {
		return nil, errors.New("search request cannot be nil")
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	var searchRequest meilisearch.SearchRequest
	if err := json.Unmarshal(payload, &searchRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search request: %w", err)
	}

	if searchRequest.Query == "" {
		// The SDK expects `q` but allow `query` for convenience
		if raw, ok := (*request)["query"]; ok {
			if queryString, isString := raw.(string); isString && strings.TrimSpace(queryString) != "" {
				searchRequest.Query = queryString
			}
		}
	}

	return &searchRequest, nil
}
