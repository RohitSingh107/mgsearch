package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a CORS middleware configured for Shopify storefronts
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// Use AllowOriginFunc to dynamically allow Shopify storefronts
		AllowOriginFunc: func(origin string) bool {
			// Allow empty origin (same-origin requests, some tools)
			if origin == "" {
				return true
			}
			
			// Allow Shopify storefronts (*.myshopify.com)
			// Handle both http and https - check if contains .myshopify.com
			if strings.Contains(origin, ".myshopify.com") {
				return true
			}
			
			// Allow localhost for development
			if strings.HasPrefix(origin, "http://localhost") || 
			   strings.HasPrefix(origin, "https://localhost") {
				return true
			}
			
			// Allow ngrok and other tunnel services for development
			if strings.Contains(origin, "ngrok") {
				return true
			}
			
			// For development: allow all origins
			// In production, you may want to be more restrictive
			return true
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Storefront-Key",
			"X-Requested-With",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"X-API-Key",
			"Access-Control-Request-Method",
			"Access-Control-Request-Headers",
			"ngrok-skip-browser-warning", // Required for ngrok tunnels
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
