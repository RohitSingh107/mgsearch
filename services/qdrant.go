package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mgsearch/config"
	"net/http"
	"time"
)

type QdrantService struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewQdrantService(cfg *config.Config) *QdrantService {
	return &QdrantService{
		baseURL: cfg.QdrantURL,
		apiKey:  cfg.QdrantAPIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ProxyQuery forwards a raw JSON body to the Qdrant query endpoint
func (s *QdrantService) ProxyQuery(collectionName string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/collections/%s/points/query", s.baseURL, collectionName)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("api-key", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Keep Recommend for legacy/internal use if needed, or remove if fully deprecated.
// Using simplified structs for Recommend to match current Qdrant API if we kept it.
type RecommendRequest struct {
	Query struct {
		Recommend struct {
			Positive []interface{} `json:"positive"`
			Negative []interface{} `json:"negative"`
		} `json:"recommend"`
	} `json:"query"`
	Limit int `json:"limit,omitempty"`
}

type QdrantPoint struct {
	ID      interface{}            `json:"id"`
	Score   float64                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
}

type QdrantResponse struct {
	Result struct {
		Points []QdrantPoint `json:"points"`
	} `json:"result"`
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

func (s *QdrantService) Recommend(collectionName string, positiveIDs []interface{}, limit int) (*QdrantResponse, error) {
	url := fmt.Sprintf("%s/collections/%s/points/query", s.baseURL, collectionName)

	if limit <= 0 {
		limit = 10 // Default limit
	}

	reqBody := RecommendRequest{}
	reqBody.Query.Recommend.Positive = positiveIDs
	reqBody.Query.Recommend.Negative = []interface{}{}
	reqBody.Limit = limit

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("api-key", s.apiKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("qdrant error (status %d): %s", resp.StatusCode, string(body))
	}

	var qdrantResp QdrantResponse
	if err := json.Unmarshal(body, &qdrantResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &qdrantResp, nil
}
