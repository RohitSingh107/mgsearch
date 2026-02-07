# API Reference

Complete documentation of all backend API endpoints, their use cases, request/response formats, and examples.

## Table of Contents

1. [Health Check](#health-check)
2. [SaaS Authentication & Management](#saas-authentication--management)
3. [Client Search & Operations](#client-search--operations)
4. [Shopify Authentication Endpoints](#shopify-authentication-endpoints)
5. [Store Management](#store-management)
6. [Session Storage](#session-storage)
7. [Storefront Search](#storefront-search)
8. [Webhooks](#webhooks)
9. [Development Proxy Endpoints](#development-proxy-endpoints)

---

## Health Check

### `GET /ping`

**Use Case:** Verify the API server is running and responsive.

**Authentication:** None required

**Request:**
```bash
curl http://localhost:8080/ping
```

**Response:**
```json
{
  "message": "pong"
}
```

**Status Codes:**
- `200 OK` - Server is healthy

---

## SaaS Authentication & Management

These endpoints are for the SaaS platform where users manage clients (organizations), API keys, and indexes.

**Authentication:** JWT Bearer Token (obtained via login/register)

### `POST /api/v1/auth/register/user`

Register a new user account.

**Authentication:** None

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

### `POST /api/v1/auth/login`

Login to get a JWT token.

**Authentication:** None

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": { ... }
}
```

### `GET /api/v1/auth/me`

Get current user profile.

### `GET /api/v1/clients`

List all clients the user has access to.

### `GET /api/v1/clients/:client_id`

Get details of a specific client, including API keys.

### `POST /api/v1/clients/:client_id/api-keys`

Generate a new API key for a client.

**Request Body:**
```json
{
  "name": "Production Key",
  "permissions": ["search", "documents"]
}
```

**Response:**
```json
{
  "api_key": "generated-api-key",
  "id": "key-id",
  ...
}
```

### `DELETE /api/v1/clients/:client_id/api-keys/:key_id`

Revoke an API key.

### `POST /api/v1/clients/:client_id/indexes`

Create a new Meilisearch index for the client.

**Request Body:**
```json
{
  "name": "products",
  "primary_key": "id"
}
```

### `GET /api/v1/clients/:client_id/indexes`

List all indexes for a client.

---

## Client Search & Operations

These endpoints are used by your applications (or your client's applications) to interact with the search engine.

**Authentication:** API Key (Bearer Token or `X-API-Key` header)
**Base URL:** `/api/v1/clients/:client_id/indexes/:index_name`

### `POST .../search`

Perform a search query.

**Authentication:** Required - Client API Key

**Path Parameters:**
- `client_id` (string) - The Client ID
- `index_name` (string) - The user-friendly index name (e.g., "products")

**Request Body:**
Any valid Meilisearch search parameters.

```json
{
  "q": "running shoes",
  "filter": "price < 100",
  "limit": 20
}
```

### `POST .../documents`

Index a document.

**Authentication:** Required - Client API Key

**Request Body:**
A single JSON object.

```json
{
  "id": "123",
  "title": "Nike Air Max",
  "price": 99.99
}
```

### `PATCH .../settings`

Update index settings.

**Authentication:** Required - Client API Key

**Request Body:**
Meilisearch settings object.

```json
{
  "searchableAttributes": ["title", "description"],
  "filterableAttributes": ["price", "brand"]
}
```

### `GET /api/v1/clients/:client_id/tasks/:task_id`

Get the status of an asynchronous task (like document indexing or settings update).

---

## Shopify Authentication Endpoints

### `POST /api/auth/shopify/begin`

**Use Case:** Initiate the Shopify OAuth installation flow. Called by Remix frontend when a merchant clicks "Install App".

**Authentication:** None required

**Request Body:**
```json
{
  "shop": "your-store.myshopify.com",
  "redirect_uri": "https://your-app-url.com/auth/callback"
}
```

... (Rest of Shopify Auth documentation remains the same)

---

## Store Management

... (Rest of Store Management documentation remains the same)

---

## Session Storage

... (Rest of Session Storage documentation remains the same)

---

## Storefront Search

### `GET /api/v1/search`

**Use Case:** Public search endpoint for Shopify storefronts.

**Authentication:** Required - Storefront API key in `X-Storefront-Key` header

... (Rest of Storefront Search documentation remains the same)

---

## Webhooks

... (Rest of Webhooks documentation remains the same)

---

## Development Proxy Endpoints

These endpoints act as a proxy to the underlying search engines (Qdrant and Meilisearch) for development and debugging purposes. They forward requests directly to the cloud services using the configured credentials.

**Base URL:** `/api/dev/proxy`

### `ANY /api/dev/proxy/qdrant/*path`

Proxies requests to the Qdrant Cloud instance.

**Example Request:**

```bash
curl --location 'http://localhost:8080/api/dev/proxy/qdrant/collections/my_collection/points/query' \
--header 'Content-Type: application/json' \
--data '{
  "with_payload": true,
  "query": {
    "recommend": {
      "positive": [12345],
      "negative": []
    }
  }
}'
```

### `ANY /api/dev/proxy/meilisearch/*path`

Proxies requests to the Meilisearch Cloud instance.

**Example Request:**

```bash
curl --location 'http://localhost:8080/api/dev/proxy/meilisearch/indexes/my_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "search term"
}'
```

