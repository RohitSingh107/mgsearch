# MGSearch – Shopify-native & SaaS Search Backend

MGSearch is a Go microservice that provides powerful search capabilities for both Shopify merchants (as an app backend) and SaaS clients (multi-tenant search). It ships with a reproducible Nix-based developer environment that provisions MongoDB, Redis, and Meilisearch locally.

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
│   ├── client.go          # SaaS Client/Tenant model
│   ├── index.go           # Meilisearch Index model
│   ├── user.go            # User model
│   ├── store.go           # Shopify Store model
│   ├── session.go         # Shopify Session model
│   └── search.go          # Search request/response models
├── repositories/          # Data access layer (database operations)
│   ├── client_repository.go
│   ├── index_repository.go
│   ├── user_repository.go
│   ├── store_repository.go
│   └── session_repository.go
├── services/              # Business logic and external service integrations
│   ├── meilisearch.go     # Meilisearch API client wrapper
│   ├── qdrant.go          # Qdrant vector database client wrapper
│   └── shopify.go         # Shopify API client wrapper
├── handlers/              # HTTP request handlers (API endpoints)
│   ├── user_auth.go       # SaaS User/Client auth
│   ├── index.go           # Index management
│   ├── search.go          # Search endpoints
│   ├── settings.go        # Settings management
│   ├── tasks.go           # Task management
│   ├── auth.go            # Shopify OAuth endpoints
│   ├── store.go           # Shopify Store management
│   ├── session.go         # Shopify Session management
│   ├── storefront.go      # Shopify Public search
│   └── webhook.go         # Shopify webhook handler
├── middleware/            # HTTP middleware (authentication, CORS, etc.)
│   ├── jwt_middleware.go      # JWT token validation
│   ├── api_key_middleware.go  # API key validation
│   ├── auth_middleware.go     # Shopify session validation
│   └── cors_middleware.go     # CORS configuration
├── pkg/                    # Reusable packages (shared utilities)
│   ├── auth/              # Authentication utilities (JWT, state tokens)
│   ├── database/          # Database connection and migrations
│   └── security/          # Encryption/decryption utilities
├── testhelpers/           # Test utilities and helpers
├── scripts/               # Utility scripts
└── docs/                  # Documentation
```

## Authentication

MGSearch uses multiple authentication mechanisms:

### 1. SaaS Authentication (JWT & API Keys)
- **JWT Tokens**: For user management (login, register, create clients).
- **Client API Keys**: For programmatic access to search indices.
  - Header: `Authorization: Bearer <api-key>`
  - Scope: Scoped to a specific Client ID.

### 2. Shopify Authentication
- **JWT Session Tokens**: For Shopify App admin/dashboard.
- **Storefront API Keys**: For public storefront search.
- **HMAC Signatures**: For Shopify Webhooks.

**📖 Complete Authentication Guide**: See [`docs/AUTH_API.md`](docs/AUTH_API.md) and [`docs/AUTHENTICATION_TYPES.md`](docs/AUTHENTICATION_TYPES.md).

## Configuration

The service reads environment variables directly or from a `.env` file. Important keys:

| Variable | Description |
| --- | --- |
| `MEILISEARCH_URL` | Meilisearch host (default comes from dev shell) |
| `MEILISEARCH_API_KEY` | Admin key for Meilisearch |
| `DATABASE_URL` | MongoDB connection string |
| `QDRANT_URL` | Qdrant host for vector similarity search (optional) |
| `QDRANT_API_KEY` | Qdrant API key (optional) |
| `SHOPIFY_API_KEY` / `SHOPIFY_API_SECRET` | Shopify app credentials |
| `SHOPIFY_APP_URL` | Public URL where Shopify redirects after OAuth |
| `JWT_SIGNING_KEY` | 32-byte hex key for signing dashboard sessions |
| `ENCRYPTION_KEY` | 32-byte hex key for encrypting Shopify tokens |

> Tip: copy `env.example` to `.env` and adjust values for local development.

## API Endpoints

**📖 Complete API Documentation:** See [`docs/API_REFERENCE.md`](docs/API_REFERENCE.md).

### Quick Reference

#### Health Check
```
GET /ping
```

#### Client Search (SaaS)
```
POST /api/v1/clients/:client_id/indexes/:index_name/search
```

#### Storefront Search (Shopify)
```
GET/POST /api/v1/search
```

#### Similar Products (Shopify)
```
GET/POST /api/v1/similar
```

📖 See [`docs/SIMILAR_PRODUCTS_API.md`](docs/SIMILAR_PRODUCTS_API.md) for vector-based product recommendations using Qdrant.

#### Update Settings (SaaS)
```
PATCH /api/v1/clients/:client_id/indexes/:index_name/settings
```

#### Get Task Details (SaaS)
```
GET /api/v1/clients/:client_id/tasks/:task_id
```

The request body can contain **any valid Meilisearch search parameters**. The service forwards the entire request body to Meilisearch as-is.

**Example:**

```bash
curl --location 'http://localhost:8080/api/v1/clients/<client_id>/indexes/products/search' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <api_key>' \
--data '{
    "q": "endgame",
    "filter": "genre = action",
    "limit": 10
}'
```

---

## Next Steps

- Implement search analytics endpoints
- Add webhook retry mechanism
- Add bulk indexing endpoints for initial sync
