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

func setupSettingsTest(t *testing.T) *gin.Engine {
	cfg := testhelpers.TestConfig()
	meiliService := services.NewMeilisearchService(cfg)

	settingsHandler := NewSettingsHandler(meiliService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	{
		v1.PATCH("/clients/:client_name/:index_name/settings", settingsHandler.UpdateSettings)
	}

	return router
}

func TestSettingsHandler_UpdateSettings(t *testing.T) {
	router := setupSettingsTest(t)

	tests := []struct {
		name           string
		clientName     string
		indexName      string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "valid settings update",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"rankingRules": []string{"words", "typo", "proximity"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "update with multiple settings",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"rankingRules":        []string{"words", "typo"},
				"searchableAttributes": []string{"title", "description"},
				"displayedAttributes":  []string{"title", "price"},
				"stopWords":            []string{"the", "a", "an"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "update with nested settings",
			clientName: "testclient",
			indexName:  "testindex",
			body: map[string]interface{}{
				"typoTolerance": map[string]interface{}{
					"minWordSizeForTypos": map[string]interface{}{
						"oneTypo":  8,
						"twoTypos": 10,
					},
					"disableOnAttributes": []string{"title"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing client name",
			clientName:     "",
			indexName:      "testindex",
			body:           map[string]interface{}{"rankingRules": []string{}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing index name",
			clientName:     "testclient",
			indexName:      "",
			body:           map[string]interface{}{"rankingRules": []string{}},
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

			url := "/api/v1/clients/" + tt.clientName + "/" + tt.indexName + "/settings"
			req := httptest.NewRequest("PATCH", url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

