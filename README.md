# MGSearch – Shopify-native search backend

MGSearch is a Go microservice that onboards Shopify merchants, syncs products into Meilisearch, and exposes both admin and storefront search APIs. It ships with a reproducible Nix-based developer environment that provisions MongoDB, Redis, and Meilisearch locally.

## Quick start (Nix)

1. **Install MongoDB** (required, not included in Nix due to build requirements):

   - **Arch/CachyOS**: `sudo pacman -S mongodb`
   - **Ubuntu/Debian**: `sudo apt install mongodb`
   - **macOS**: `brew install mongodb-community`
   - **Or download**: https://www.mongodb.com/try/download/community

2. **Enter the dev shell (installs Go, Redis, Meilisearch, etc.):**

   ```bash
   nix --extra-experimental-features 'nix-command flakes' develop
   ```

   The shell exports sane defaults such as:

   - `DATABASE_URL=mongodb://localhost:27017/mgsearch`
   - `REDIS_URL=redis://127.0.0.1:6381/0`
   - `MEILISEARCH_URL=http://127.0.0.1:7701`
   - `MEILISEARCH_API_KEY=dev-master-key`

3. **Start local services (MongoDB & Redis only):**

   ```bash
   just dev-up
   ```

   - MongoDB listens on `27017`
   - Redis listens on `6381`

   Stop them with `just dev-down` and inspect status via `just dev-status`.

3. **Configure your Meilisearch Cloud host**

   Set `MEILISEARCH_URL` / `MEILISEARCH_API_KEY` in `.env` to point at your managed Meilisearch instance (we no longer run Meilisearch locally).

4. **Run the API server:**

   ```bash
   go run main.go
   ```

   Database schema migrations run automatically on boot.

4. **Run tests/linting as needed:**

   ```bash
   just fmt
   just lint
   just test
   ```

### Without Nix

Install Go 1.23+, MongoDB 7+, and Redis 7+ manually, then point `MEILISEARCH_URL` / `MEILISEARCH_API_KEY` at your hosted Meilisearch deployment before running the service.

## Code Structure

This project follows a clean, layered architecture pattern that separates concerns and makes the codebase maintainable and testable.

### Directory Overview

```
mgsearch/
├── main.go                 # Application entry point, server setup, route registration
├── config/                 # Configuration management
│   └── config.go          # Loads environment variables and provides Config struct
├── models/                # Domain models (data structures)
│   ├── store.go           # Store model (Shopify merchant data)
│   ├── session.go         # Session model (OAuth session data)
│   └── search.go          # Search request/response models
├── repositories/          # Data access layer (database operations)
│   ├── store_repository.go    # MongoDB operations for stores
│   └── session_repository.go  # MongoDB operations for sessions
├── services/              # Business logic and external service integrations
│   ├── meilisearch.go     # Meilisearch API client wrapper
│   └── shopify.go         # Shopify API client wrapper
├── handlers/              # HTTP request handlers (API endpoints)
│   ├── auth.go            # OAuth authentication endpoints
│   ├── store.go           # Store management endpoints
│   ├── session.go         # Session management endpoints
│   ├── storefront.go      # Public storefront search endpoint
│   ├── search.go          # Legacy search endpoints
│   ├── webhook.go         # Shopify webhook handler
│   ├── settings.go        # Meilisearch settings management
│   └── tasks.go           # Meilisearch task status
├── middleware/            # HTTP middleware (authentication, CORS, etc.)
│   ├── auth_middleware.go     # JWT token validation
│   ├── api_key_middleware.go  # API key validation
│   └── cors_middleware.go     # CORS configuration
├── pkg/                    # Reusable packages (shared utilities)
│   ├── auth/              # Authentication utilities (JWT, state tokens)
│   ├── database/          # Database connection and migrations
│   └── security/          # Encryption/decryption utilities
├── testhelpers/           # Test utilities and helpers
│   ├── test_setup.go      # Test database setup
│   └── router_setup.go    # Test router configuration
└── scripts/               # Utility scripts
    ├── dev/               # Development scripts
    └── *.sh               # Helper shell scripts
```

### Architecture Layers

#### 1. **Models** (`models/`)
Domain models define the core data structures used throughout the application. They use BSON tags for MongoDB serialization and JSON tags for API responses.

**Key Models:**
- **`Store`**: Represents a Shopify merchant/store with encrypted tokens, API keys, Meilisearch configuration, and sync state
- **`Session`**: Represents a Shopify OAuth session with user information and access tokens
- **`SearchRequest`/`SearchResponse`**: Flexible JSON structures for Meilisearch search operations

**Example:**
```go
type Store struct {
    ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    ShopDomain           string            `json:"shop_domain" bson:"shop_domain"`
    EncryptedAccessToken []byte            `json:"-" bson:"encrypted_access_token"`
    // ... more fields
}
```

#### 2. **Repositories** (`repositories/`)
The repository pattern abstracts database operations, providing a clean interface for data access. All MongoDB-specific logic is contained here.

**Responsibilities:**
- CRUD operations (Create, Read, Update, Delete)
- Query building and execution
- Data transformation between models and database documents

**Key Repositories:**
- **`StoreRepository`**: Manages store data (create/update, find by domain/ID/API key, update sync state)
- **`SessionRepository`**: Manages session data (store, load, delete, find by shop, cleanup expired)

**Example:**
```go
type StoreRepository struct {
    collection *mongo.Collection
}

func (r *StoreRepository) GetByShopDomain(ctx context.Context, domain string) (*models.Store, error) {
    var store models.Store
    err := r.collection.FindOne(ctx, bson.M{"shop_domain": domain}).Decode(&store)
    // ... error handling
    return &store, nil
}
```

#### 3. **Services** (`services/`)
Services encapsulate business logic and external API integrations. They handle complex operations that may involve multiple steps or external services.

**Key Services:**
- **`MeilisearchService`**: Wraps Meilisearch SDK for search, indexing, settings, and task management
- **`ShopifyService`**: Handles Shopify API calls (OAuth, webhook verification, product fetching)

**Example:**
```go
type MeilisearchService struct {
    client     meilisearch.ServiceManager
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func (s *MeilisearchService) Search(indexName string, request *models.SearchRequest) (*models.SearchResponse, error) {
    // Business logic for search operations
}
```

#### 4. **Handlers** (`handlers/`)
Handlers are HTTP request processors that handle routing, request validation, response formatting, and orchestration between repositories and services.

**Request Flow:**
1. Receive HTTP request
2. Validate input (query params, body, headers)
3. Call repository/service methods
4. Format and return HTTP response

**Key Handlers:**
- **`AuthHandler`**: OAuth flow (begin, callback, exchange, install)
- **`StoreHandler`**: Store information endpoints (current store, sync status)
- **`SessionHandler`**: Session CRUD operations
- **`StorefrontHandler`**: Public search API with storefront key authentication
- **`SearchHandler`**: Legacy search endpoints
- **`WebhookHandler`**: Shopify webhook processing (products create/update/delete)
- **`SettingsHandler`**: Meilisearch index settings management
- **`TasksHandler`**: Meilisearch task status queries

**Example:**
```go
type StoreHandler struct {
    repo *repositories.StoreRepository
}

func (h *StoreHandler) GetCurrentStore(c *gin.Context) {
    storeID, _ := middleware.GetStoreID(c)
    store, err := h.repo.GetByID(c.Request.Context(), storeID)
    // ... error handling
    c.JSON(http.StatusOK, store.ToPublicView())
}
```

#### 5. **Middleware** (`middleware/`)
Middleware functions intercept HTTP requests to add cross-cutting concerns like authentication, CORS, and request validation.

**Key Middleware:**
- **`AuthMiddleware`**: Validates JWT tokens and extracts store information
- **`APIKeyMiddleware`**: Validates API keys for session endpoints
- **`CORSMiddleware`**: Handles CORS headers for storefront requests

**Example:**
```go
func (m *AuthMiddleware) RequireStoreSession() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract and validate JWT token
        // Set store context for handlers
    }
}
```

#### 6. **Packages** (`pkg/`)
Reusable utility packages that provide common functionality across the application.

**Key Packages:**
- **`pkg/auth`**: JWT token generation/validation, state token management
- **`pkg/database`**: MongoDB client creation, connection management, migrations
- **`pkg/security`**: AES-GCM encryption/decryption for sensitive data (tokens, API keys)

#### 7. **Configuration** (`config/`)
Centralized configuration management that loads environment variables and provides a typed `Config` struct.

**Features:**
- Loads from `.env` file or environment variables
- Environment variables take precedence
- Provides defaults for development

### Data Flow Example

Here's how a typical request flows through the application:

**Example: `GET /api/stores/current`**

1. **Request arrives** → `main.go` routes to `StoreHandler.GetCurrentStore`
2. **Middleware** → `AuthMiddleware.RequireStoreSession()` validates JWT token
3. **Handler** → `StoreHandler.GetCurrentStore()` extracts store ID from context
4. **Repository** → `StoreRepository.GetByID()` queries MongoDB
5. **Model** → `Store.ToPublicView()` converts to public representation
6. **Response** → Handler returns JSON response

```
HTTP Request
    ↓
[Middleware: Auth] → Validates JWT, sets context
    ↓
[Handler: StoreHandler] → Extracts store ID from context
    ↓
[Repository: StoreRepository] → Queries MongoDB
    ↓
[Model: Store] → Converts to public view
    ↓
HTTP Response (JSON)
```

### Testing Structure

Tests are co-located with handlers in `handlers/*_test.go` files. The `testhelpers/` package provides:
- Test database setup and cleanup
- Test configuration
- Test router setup utilities

**Test Pattern:**
```go
func TestStoreHandler_GetCurrentStore(t *testing.T) {
    router, _, token, cleanup := setupStoreTest(t)
    defer cleanup()
    
    // Test cases with table-driven tests
    // ...
}
```

### Key Design Principles

1. **Separation of Concerns**: Each layer has a single responsibility
2. **Dependency Injection**: Handlers receive repositories/services as dependencies
3. **Interface-Based Design**: Services can be easily mocked for testing
4. **Error Handling**: Consistent error responses across all endpoints
5. **Security**: Sensitive data (tokens, keys) encrypted at rest
6. **Testability**: Clean architecture enables comprehensive test coverage

## Authentication

MGSearch uses multiple authentication mechanisms depending on the endpoint:

### Authentication Types

1. **JWT Session Tokens** - For admin/dashboard endpoints
   - Used by: `/api/stores/current`, `/api/stores/sync-status`
   - Header: `Authorization: Bearer <jwt-token>`
   - Generated after OAuth installation, valid for 24 hours

2. **Storefront API Keys** - For public storefront search
   - Used by: `/api/v1/search` (GET/POST)
   - Header: `X-Storefront-Key: <public-key>`
   - Unique per store, generated during installation

3. **Optional API Keys** - For session management endpoints
   - Used by: `/api/sessions/*` (all endpoints)
   - Header: `Authorization: Bearer <api-key>`
   - Only required if `SESSION_API_KEY` environment variable is set

4. **HMAC Signature Verification** - For Shopify webhooks
   - Used by: `/webhooks/shopify/:topic/:subtopic`
   - Headers: `X-Shopify-Hmac-Sha256`, `X-Shopify-Shop-Domain`
   - Verifies webhook authenticity using HMAC-SHA256

**📖 Complete Authentication Guide**: See [`docs/AUTHENTICATION_TYPES.md`](docs/AUTHENTICATION_TYPES.md) for detailed documentation, examples, and troubleshooting.

## Configuration

The service reads environment variables directly or from a `.env` file. Important keys:

| Variable | Description |
| --- | --- |
| `MEILISEARCH_URL` | Meilisearch host (default comes from dev shell) |
| `MEILISEARCH_API_KEY` | Admin key for Meilisearch |
| `DATABASE_URL` | MongoDB connection string |
| `SHOPIFY_API_KEY` / `SHOPIFY_API_SECRET` | Shopify app credentials |
| `SHOPIFY_APP_URL` | Public URL where Shopify redirects after OAuth |
| `SHOPIFY_SCOPES` | Requested scopes (default defined in config) |
| `JWT_SIGNING_KEY` | 32-byte hex key for signing dashboard sessions |
| `ENCRYPTION_KEY` | 32-byte hex key for encrypting Shopify tokens |

> Tip: copy `env.example` to `.env` and adjust values for local development.

## Remix Frontend Integration

This backend is designed to work with a Remix frontend (created with Shopify CLI). The Remix app handles the OAuth UI flow, while this backend stores encrypted tokens, manages search indexes, and handles webhooks.

### Quick Start

1. **See integration guide**: [`docs/REMIX_INTEGRATION.md`](docs/REMIX_INTEGRATION.md) for complete implementation steps
2. **Quick checklist**: [`docs/REMIX_QUICKSTART.md`](docs/REMIX_QUICKSTART.md) for a step-by-step setup

### Integration Overview

- **Remix Frontend**: Handles OAuth UI, merchant dashboard, theme integration
- **Go Backend**: Stores tokens, syncs products, handles webhooks, provides search APIs
- **Communication**: Remix → Go backend via REST API with JWT session tokens

### Key Integration Points

1. **OAuth Flow**: Remix initiates OAuth, Go backend handles token exchange and storage
2. **Dashboard**: Remix calls `GET /api/stores/current` to display store info
3. **Storefront Search**: Theme JavaScript calls `GET /api/v1/search` with storefront API key

## API Endpoints

**📖 Complete API Documentation:** See [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) for detailed endpoint documentation, request/response formats, examples, and use cases.

### Quick Reference

#### Health Check
```
GET /ping
```

### Search
```
POST /api/v1/clients/:client_name/:index_name/search
```

### Update Settings
```
PATCH /api/v1/clients/:client_name/:index_name/settings
```

### Get Task Details
```
GET /api/v1/clients/:client_name/tasks/:task_id
```

The request body can contain **any valid Meilisearch search parameters** with **multi-level nested JSON structures**. The service forwards the entire request body to Meilisearch as-is.

**Examples:**

Simple search:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame"
}'
```

Search with filters and sorting:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame",
    "filter": "genre = action",
    "sort": ["release_date:desc"],
    "limit": 10,
    "offset": 0
}'
```

Search with complex nested filters:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame",
    "filter": ["genre = action", "year > 2020"],
    "facets": ["genre", "year"],
    "attributesToRetrieve": ["title", "description"],
    "attributesToHighlight": ["description"]
}'
```

**Supported Parameters:**

The service accepts all valid Meilisearch search parameters as documented in the [official Meilisearch API reference](https://www.meilisearch.com/docs/reference/api/search#search-parameters). All parameters are passed through to Meilisearch as-is.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | String | `""` | Query string |
| `offset` | Integer | `0` | Number of documents to skip |
| `limit` | Integer | `20` | Maximum number of documents returned |
| `hitsPerPage` | Integer | `1` | Maximum number of documents returned for a page |
| `page` | Integer | `1` | Request a specific page of results |
| `filter` | String or Array of strings | `null` | Filter queries by an attribute's value |
| `facets` | Array of strings | `null` | Display the count of matches per facet |
| `distinct` | String | `null` | Restrict search to documents with unique values of specified attribute |
| `attributesToRetrieve` | Array of strings | `["*"]` | Attributes to display in the returned documents |
| `attributesToCrop` | Array of strings | `null` | Attributes whose values have to be cropped |
| `cropLength` | Integer | `10` | Maximum length of cropped value in words |
| `cropMarker` | String | `"…"` | String marking crop boundaries |
| `attributesToHighlight` | Array of strings | `null` | Highlight matching terms contained in an attribute |
| `highlightPreTag` | String | `"<em>"` | String inserted at the start of a highlighted term |
| `highlightPostTag` | String | `"</em>"` | String inserted at the end of a highlighted term |
| `showMatchesPosition` | Boolean | `false` | Return matching terms location |
| `sort` | Array of strings | `null` | Sort search results by an attribute's value |
| `matchingStrategy` | String | `last` | Strategy used to match query terms within documents |
| `showRankingScore` | Boolean | `false` | Display the global ranking score of a document |
| `showRankingScoreDetails` | Boolean | `false` | Adds a detailed global ranking score field |
| `rankingScoreThreshold` | Number | `null` | Excludes results with low ranking scores |
| `attributesToSearchOn` | Array of strings | `["*"]` | Restrict search to the specified attributes |
| `hybrid` | Object | `null` | Return results based on query keywords and meaning (requires `embedder` and `semanticRatio`) |
| `vector` | Array of numbers | `null` | Search using a custom query vector |
| `retrieveVectors` | Boolean | `false` | Return document and query vector data |
| `locales` | Array of strings | `null` | Explicitly specify languages used in a query |
| `media` | Object | `null` | Perform AI-powered search queries with multimodal content (experimental) |

**Examples with advanced parameters:**

Hybrid search (semantic + keyword):
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "kitchen utensils",
    "hybrid": {
        "semanticRatio": 0.9,
        "embedder": "EMBEDDER_NAME"
    }
}'
```

Vector search:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "vector": [0.1, 0.2, 0.3],
    "hybrid": {
        "embedder": "EMBEDDER_NAME"
    }
}'
```

Search with locales:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "進撃の巨人",
    "locales": ["jpn"]
}'
```

For complete parameter documentation, see the [Meilisearch Search API Reference](https://www.meilisearch.com/docs/reference/api/search#search-parameters).

The request body can contain **any valid Meilisearch index settings parameters** with **multi-level nested JSON structures**. The service forwards the entire request body to Meilisearch as-is.

**Example:**
```bash
curl --location --request PATCH 'http://localhost:8080/api/v1/clients/myclient/movies/settings' \
--header 'Content-Type: application/json' \
--data '{
    "rankingRules": [
        "words",
        "typo",
        "proximity",
        "attribute",
        "sort",
        "exactness",
        "release_date:desc",
        "rank:desc"
    ],
    "distinctAttribute": "movie_id",
    "searchableAttributes": [
        "title",
        "overview",
        "genres"
    ],
    "displayedAttributes": [
        "title",
        "overview",
        "genres",
        "release_date"
    ],
    "stopWords": [
        "the",
        "a",
        "an"
    ],
    "sortableAttributes": [
        "title",
        "release_date"
    ],
    "synonyms": {
        "wolverine": ["xmen", "logan"],
        "logan": ["wolverine"]
    },
    "typoTolerance": {
        "minWordSizeForTypos": {
            "oneTypo": 8,
            "twoTypos": 10
        },
        "disableOnAttributes": ["title"]
    },
    "pagination": {
        "maxTotalHits": 5000
    },
    "faceting": {
        "maxValuesPerFacet": 200
    },
    "searchCutoffMs": 150
}'
```

**Supported Settings Parameters:**

The service accepts all valid Meilisearch index settings parameters including:
- `rankingRules` - Array of strings defining the ranking order
- `distinctAttribute` - String specifying the distinct attribute
- `searchableAttributes` - Array of strings defining searchable attributes
- `displayedAttributes` - Array of strings defining displayed attributes
- `stopWords` - Array of strings defining stop words
- `sortableAttributes` - Array of strings defining sortable attributes
- `synonyms` - Object mapping terms to their synonyms
- `typoTolerance` - Object configuring typo tolerance settings
- `pagination` - Object configuring pagination settings
- `faceting` - Object configuring faceting settings
- `searchCutoffMs` - Number specifying search cutoff time in milliseconds
- And any other Meilisearch index settings parameters, including nested JSON structures

For complete settings documentation, see the [Meilisearch Settings API Reference](https://www.meilisearch.com/docs/reference/api/settings).

Retrieves task details from Meilisearch by task UID. This is useful for checking the status of asynchronous operations like settings updates.

**Example:**
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/tasks/15' \
--header 'Content-Type: application/json'
```

**Response:**
The response includes task details such as:
- `uid` - Task UID
- `indexUid` - Index UID associated with the task
- `status` - Task status (enqueued, processing, succeeded, failed)
- `type` - Task type (e.g., "settingsUpdate")
- `details` - Task details (varies by task type)
- `error` - Error information if the task failed
- `duration` - Task execution duration
- `enqueuedAt` - When the task was enqueued
- `startedAt` - When the task started processing
- `finishedAt` - When the task finished

**Example Response:**
```json
{
    "uid": 15,
    "batchUid": 13,
    "indexUid": "test_index",
    "status": "succeeded",
    "type": "settingsUpdate",
    "canceledBy": null,
    "details": {
        "displayedAttributes": ["title", "overview", "genres", "release_date"],
        "searchableAttributes": ["title", "overview", "genres"],
        "sortableAttributes": ["release_date", "title"],
        "rankingRules": ["words", "typo", "proximity", "attribute", "sort", "exactness", "release_date:desc", "rank:desc"],
        "stopWords": ["a", "an", "the"],
        "synonyms": {
            "logan": ["wolverine"],
            "wolverine": ["xmen", "logan"]
        },
        "distinctAttribute": "movie_id",
        "typoTolerance": {
            "minWordSizeForTypos": {
                "oneTypo": 8,
                "twoTypos": 10
            },
            "disableOnAttributes": ["title"]
        },
        "faceting": {
            "maxValuesPerFacet": 200
        },
        "pagination": {
            "maxTotalHits": 5000
        },
        "searchCutoffMs": 150
    },
    "error": null,
    "duration": "PT9.752253947S",
    "enqueuedAt": "2025-11-22T16:35:16.14171112Z",
    "startedAt": "2025-11-22T16:35:16.158978866Z",
    "finishedAt": "2025-11-22T16:35:25.911232813Z"
}
```

**Note:** When you update settings, Meilisearch returns a task response with `taskUid`. You can use that `taskUid` with this endpoint to check the status and get the final result of the settings update.

### All Endpoints Summary

| Method | Endpoint | Auth | Use Case |
|--------|----------|------|----------|
| `GET` | `/ping` | None | Health check |
| `POST` | `/api/auth/shopify/begin` | None | Initiate OAuth flow |
| `GET` | `/api/auth/shopify/callback` | HMAC | Handle OAuth callback (backend-handled) |
| `POST` | `/api/auth/shopify/exchange` | None | Exchange OAuth code for token (optional helper) |
| `POST` | `/api/auth/shopify/install` | None | Store OAuth data (frontend-handled OAuth) |
| `GET` | `/api/stores/current` | JWT | Get authenticated store info |
| `GET` | `/api/stores/sync-status` | JWT | Get sync status (optimized for polling) |
| `POST` | `/api/sessions` | None | Store Shopify OAuth session |
| `GET` | `/api/sessions/:id` | None | Load session by ID |
| `DELETE` | `/api/sessions/:id` | None | Delete session by ID |
| `DELETE` | `/api/sessions/batch` | None | Delete multiple sessions |
| `GET` | `/api/v1/search` | Storefront Key | Public storefront search |
| `POST` | `/webhooks/shopify/:topic/:subtopic` | HMAC | Receive Shopify webhooks |
| `POST` | `/api/v1/clients/:client/:index/search` | None | Legacy search (backward compat) |
| `POST` | `/api/v1/clients/:client/:index/documents` | None | Legacy indexing (backward compat) |

**See [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md) for complete documentation with examples.**

