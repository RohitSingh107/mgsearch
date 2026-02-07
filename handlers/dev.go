package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"mgsearch/config"

	"github.com/gin-gonic/gin"
)

type DevHandler struct {
	cfg *config.Config
}

func NewDevHandler(cfg *config.Config) *DevHandler {
	return &DevHandler{cfg: cfg}
}

func (h *DevHandler) ProxyQdrant(c *gin.Context) {
	if h.cfg.QdrantURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "QDRANT_URL is not configured"})
		return
	}

	qdrantURL := h.cfg.QdrantURL
	// Ensure scheme is present
	if !strings.HasPrefix(qdrantURL, "http://") && !strings.HasPrefix(qdrantURL, "https://") {
		qdrantURL = "http://" + qdrantURL
	}

	remote, err := url.Parse(qdrantURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Qdrant URL configuration"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("path")

		// Add API Key
		if h.cfg.QdrantAPIKey != "" {
			req.Header.Set("api-key", h.cfg.QdrantAPIKey)
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *DevHandler) ProxyMeilisearch(c *gin.Context) {
	if h.cfg.MeilisearchURL == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MEILISEARCH_URL is not configured"})
		return
	}

	meilisearchURL := h.cfg.MeilisearchURL
	if !strings.HasPrefix(meilisearchURL, "http://") && !strings.HasPrefix(meilisearchURL, "https://") {
		meilisearchURL = "http://" + meilisearchURL
	}

	remote, err := url.Parse(meilisearchURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Meilisearch URL configuration"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	proxy.Director = func(req *http.Request) {
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = c.Param("path")

		// Add API Key
		if h.cfg.MeilisearchAPIKey != "" {
			req.Header.Set("Authorization", "Bearer "+h.cfg.MeilisearchAPIKey)
		}
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
