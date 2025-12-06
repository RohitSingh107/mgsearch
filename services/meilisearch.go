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

// UpdateSettings performs a PATCH request to update Meilisearch index settings
// indexName: the name of the index to update (e.g., "movies")
// request: the settings update request body (can contain any Meilisearch settings parameters)
func (s *MeilisearchService) UpdateSettings(indexName string, request *models.SettingsRequest) (*models.SettingsResponse, error) {
	// Construct the Meilisearch settings endpoint
	url := fmt.Sprintf("%s/indexes/%s/settings", s.baseURL, indexName)

	// Marshal request body to JSON (preserves any structure with nested objects, arrays, etc.)
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP PATCH request
	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(requestBody))
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

	// Check for HTTP errors (Meilisearch returns 202 Accepted for settings updates)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("meilisearch error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var settingsResponse models.SettingsResponse
	if err := json.Unmarshal(body, &settingsResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &settingsResponse, nil
}

// GetTask retrieves task details from Meilisearch by task UID
// taskUID: the task UID to retrieve (e.g., 15)
func (s *MeilisearchService) GetTask(taskUID string) (*models.TaskResponse, error) {
	// Construct the Meilisearch task endpoint
	url := fmt.Sprintf("%s/tasks/%s", s.baseURL, taskUID)

	// Create HTTP GET request
	req, err := http.NewRequest("GET", url, nil)
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
	var taskResponse models.TaskResponse
	if err := json.Unmarshal(body, &taskResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &taskResponse, nil
}
