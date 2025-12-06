package models

// SearchRequest represents the search request from client
// It can contain any Meilisearch search parameters (q, filter, sort, limit, offset, etc.)
// and supports multi-level nested JSON structures
type SearchRequest map[string]interface{}

// SearchResponse represents the response from Meilisearch
// This will be passed through as-is from Meilisearch
type SearchResponse map[string]interface{}

// Document represents a single Meilisearch document payload
type Document map[string]interface{}

// IndexDocumentResponse represents the asynchronous task response from Meilisearch
type IndexDocumentResponse map[string]interface{}
