# MGSearch – Shopify-native search backend

MGSearch is a Go microservice that onboards Shopify merchants, syncs products into Meilisearch, and exposes both admin and storefront search APIs. It ships with a reproducible Nix-based developer environment that provisions Postgres, Redis, and Meilisearch locally.

## Quick start (Nix)

1. **Enter the dev shell (installs Go, Postgres, Redis, Meilisearch, etc.):**

   ```bash
   nix --extra-experimental-features 'nix-command flakes' develop
   ```

   The shell exports sane defaults such as:

   - `DATABASE_URL=postgres://mgsearch:mgsearch@localhost:5544/mgsearch?sslmode=disable`
   - `REDIS_URL=redis://127.0.0.1:6381/0`
   - `MEILISEARCH_URL=http://127.0.0.1:7701`
   - `MEILISEARCH_API_KEY=dev-master-key`

2. **Start local services (Postgres & Redis only):**

   ```bash
   just dev-up
   ```

   - Postgres listens on `5544`
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

Install Go 1.23+, Postgres 16+, and Redis 7+ manually, then point `MEILISEARCH_URL` / `MEILISEARCH_API_KEY` at your hosted Meilisearch deployment before running the service.

## Configuration

The service reads environment variables directly or from a `.env` file. Important keys:

| Variable | Description |
| --- | --- |
| `MEILISEARCH_URL` | Meilisearch host (default comes from dev shell) |
| `MEILISEARCH_API_KEY` | Admin key for Meilisearch |
| `DATABASE_URL` | Postgres connection string |
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

