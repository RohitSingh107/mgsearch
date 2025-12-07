package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mgsearch/services"
	"mgsearch/testhelpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupSearchTest(t *testing.T) (*gin.Engine, *services.MeilisearchService) {
	cfg := testhelpers.TestConfig()
	meiliService := services.NewMeilisearchService(cfg)

	searchHandler := NewSearchHandler(meiliService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.POST("/clients/:client_name/:index_name/search", searchHandler.Search)
		v1.POST("/clients/:client_name/:index_name/documents", searchHandler.IndexDocument)
	}

	return router, meiliService
}

func TestSearchHandler_Search(t *testing.T) {
	router, _ := setupSearchTest(t)

	tests := []struct {
		name           string
		clientName     string
		indexName      string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "valid search request",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"q": "test query",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "search with filters and sorting",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"q":      "test query",
				"filter": "genre = action",
				"sort":   []string{"release_date:desc"},
				"limit":  10,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing client name",
			clientName:     "",
			indexName:      "testindex",
			body:           map[string]interface{}{"q": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing index name",
			clientName:     "testclient",
			indexName:      "",
			body:           map[string]interface{}{"q": "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid request body",
			clientName:     "testclient",
			indexName:      "testindex",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			} else {
				bodyBytes = []byte("invalid json")
			}

			url := "/api/v1/clients/" + tt.clientName + "/" + tt.indexName + "/search"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSearchHandler_IndexDocument(t *testing.T) {
	router, _ := setupSearchTest(t)

	tests := []struct {
		name           string
		clientName     string
		indexName      string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "valid document",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"id":    "doc1",
				"title": "Test Document",
				"price": 99.99,
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "empty document",
			clientName:     "testclient",
			indexName:      "testindex",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing client name",
			clientName:     "",
			indexName:      "testindex",
			body:           map[string]interface{}{"id": "doc1"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing index name",
			clientName:     "testclient",
			indexName:      "",
			body:           map[string]interface{}{"id": "doc1"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			clientName:     "testclient",
			indexName:      "testindex",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			} else {
				bodyBytes = []byte("invalid json")
			}

			url := "/api/v1/clients/" + tt.clientName + "/" + tt.indexName + "/documents"
			req := httptest.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

