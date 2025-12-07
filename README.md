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

