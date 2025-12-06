package models

// SearchRequest represents the search request from client
// It can contain any Meilisearch search parameters (q, filter, sort, limit, offset, etc.)
// and supports multi-level nested JSON structures
type SearchRequest map[string]interface{}

// SearchResponse represents the response from Meilisearch
// This will be passed through as-is from Meilisearch
type SearchResponse map[string]interface{}

// SettingsRequest represents the settings update request from client
// It can contain any Meilisearch settings parameters (rankingRules, distinctAttribute,
// searchableAttributes, displayedAttributes, stopWords, sortableAttributes, synonyms,
// typoTolerance, pagination, faceting, searchCutoffMs, etc.) and supports multi-level nested JSON structures
type SettingsRequest map[string]interface{}

// SettingsResponse represents the response from Meilisearch settings update
// This will be passed through as-is from Meilisearch
type SettingsResponse map[string]interface{}

// TaskResponse represents the response from Meilisearch task details
// This will be passed through as-is from Meilisearch
type TaskResponse map[string]interface{}
