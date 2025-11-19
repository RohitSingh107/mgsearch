package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mgsearch/config"
	"mgsearch/models"
	"net/http"
)

type MeilisearchService struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewMeilisearchService creates a new Meilisearch service instance
func NewMeilisearchService(cfg *config.Config) *MeilisearchService {
	return &MeilisearchService{
		baseURL:    cfg.MeilisearchURL,
		apiKey:     cfg.MeilisearchAPIKey,
		httpClient: &http.Client{},
	}
}

// Search performs a search request to Meilisearch
// indexName: the name of the index to search (e.g., "test_index")
// request: the search request body (can contain any Meilisearch parameters)
func (s *MeilisearchService) Search(indexName string, request *models.SearchRequest) (*models.SearchResponse, error) {
	// Construct the Meilisearch search endpoint
	url := fmt.Sprintf("%s/indexes/%s/search", s.baseURL, indexName)

	// Marshal request body to JSON (preserves any structure with filters, sort, nested objects, etc.)
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meilisearch error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var searchResponse models.SearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &searchResponse, nil
}
