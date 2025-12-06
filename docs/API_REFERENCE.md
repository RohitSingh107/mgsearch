# API Reference

Complete documentation of all backend API endpoints, their use cases, request/response formats, and examples.

## Table of Contents

1. [Health Check](#health-check)
2. [Authentication Endpoints](#authentication-endpoints)
3. [Store Management](#store-management)
4. [Session Storage](#session-storage)
5. [Storefront Search](#storefront-search)
6. [Webhooks](#webhooks)
7. [Legacy Search Endpoints](#legacy-search-endpoints)

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

## Authentication Endpoints

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

**Request Fields:**
- `shop` (string, required) - Shopify shop domain (must end with `.myshopify.com`)
- `redirect_uri` (string, optional) - OAuth redirect URI. If not provided, defaults to `{SHOPIFY_APP_URL}/auth/callback`. **Important:** The redirect_uri is used exactly as provided without modification.

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/auth/shopify/begin \
  -H "Content-Type: application/json" \
  -d '{
    "shop": "acme-store.myshopify.com",
    "redirect_uri": "https://deluxe-compilation-haven-destiny.trycloudflare.com/auth/callback"
  }'
```

**Request Example (without redirect_uri - uses default):**
```bash
curl -X POST http://localhost:8080/api/auth/shopify/begin \
  -H "Content-Type: application/json" \
  -d '{
    "shop": "acme-store.myshopify.com"
  }'
```

**Response:**
```json
{
  "authUrl": "https://acme-store.myshopify.com/admin/oauth/authorize?client_id=...&scope=...&redirect_uri=...&state=...",
  "state": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response Fields:**
- `authUrl` (string) - Full Shopify OAuth authorization URL to redirect merchant to
- `state` (string) - JWT state token for CSRF protection (valid for 15 minutes)

**Status Codes:**
- `200 OK` - OAuth URL generated successfully
- `400 Bad Request` - Invalid shop domain (must end with `.myshopify.com`)
- `500 Internal Server Error` - Failed to generate state token or build auth URL

**Use in Remix:**
```typescript
// Get the current app URL (e.g., from environment or request)
const appUrl = process.env.SHOPIFY_APP_URL || "https://your-app.ngrok.io";
const redirectUri = `${appUrl}/auth/callback`; // Remix route

const response = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/begin`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ 
    shop: "acme-store.myshopify.com",
    redirect_uri: redirectUri // Pass the Remix callback route
  }),
});
const { authUrl } = await response.json();
window.location.href = authUrl; // Redirect merchant to Shopify
```

**Important Notes:**
- The `redirect_uri` must match exactly what's configured in your Shopify Partner Dashboard
- For Remix apps, this is typically `/auth/callback` (not `/api/auth/shopify/callback`)
- The backend uses the `redirect_uri` exactly as provided, without modification

---

### `GET /api/auth/shopify/callback`

**Use Case:** Handle OAuth callback from Shopify after merchant authorizes the app. Exchanges authorization code for access token, stores encrypted token, creates store record, generates API keys, and returns session JWT.

**Authentication:** None required (validated via HMAC signature from Shopify)

**Query Parameters:**
- `code` (string, required) - Authorization code from Shopify
- `shop` (string, required) - Shop domain
- `state` (string, required) - State token from `/begin` endpoint
- `hmac` (string, required) - HMAC signature for verification
- `timestamp` (string, required) - Request timestamp
- `host` (string, optional) - Shopify admin host

**Optional Headers:**
- `X-Meilisearch-Url` (string) - Override Meilisearch URL for this store (defaults to `MEILISEARCH_URL` env var)
- `X-Meilisearch-Api-Key` (string) - Override Meilisearch API key for this store (defaults to `MEILISEARCH_API_KEY` env var)

**Request Example:**
```bash
curl "http://localhost:8080/api/auth/shopify/callback?code=abc123&shop=acme-store.myshopify.com&state=xyz&hmac=...&timestamp=1234567890"
```

**Response:**
```json
{
  "store": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "shop_domain": "acme-store.myshopify.com",
    "shop_name": "acme-store",
    "product_index_uid": "acme_store_all_products",
    "meilisearch_index_uid": "acme_store_all_products",
    "meilisearch_document_type": "product",
    "plan_level": "free",
    "status": "active",
    "sync_state": {
      "status": "pending_initial_sync"
    },
    "installed_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "installation successful"
}
```

**Response Fields:**
- `store` (object) - Public store information (sensitive fields like API keys excluded)
- `token` (string) - JWT session token for authenticated requests (valid for 24 hours)
- `message` (string) - Success message

**Status Codes:**
- `200 OK` - Installation successful
- `400 Bad Request` - Missing required parameters
- `401 Unauthorized` - Invalid HMAC or state token

**Note:** This endpoint is for the traditional OAuth flow where Shopify redirects directly to the backend. For frontend-handled OAuth, use `/api/auth/shopify/install` instead.

---

### `POST /api/auth/shopify/exchange`

**Use Case:** Optional helper endpoint to exchange OAuth authorization code for access token. Frontend can use this if they want the backend to handle the token exchange, or they can exchange it directly with Shopify.

**Authentication:** None required

**Request Body:**
```json
{
  "shop": "acme-store.myshopify.com",
  "code": "authorization_code_from_shopify"
}
```

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/auth/shopify/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "shop": "acme-store.myshopify.com",
    "code": "abc123def456"
  }'
```

**Response:**
```json
{
  "access_token": "shpat_abc123def456...",
  "scope": "read_products,write_products,read_product_listings"
}
```

**Response Fields:**
- `access_token` (string) - Shopify OAuth access token
- `scope` (string) - Granted OAuth scopes

**Status Codes:**
- `200 OK` - Token exchange successful
- `400 Bad Request` - Invalid shop domain or missing code
- `500 Internal Server Error` - Token exchange failed

**Use in Remix (Optional):**
```typescript
// After receiving code from Shopify OAuth callback
const response = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/exchange`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ shop, code }),
});
const { access_token } = await response.json();
// Then send to /api/auth/shopify/install
```

---

### `POST /api/auth/shopify/install`

**Use Case:** Store OAuth data after frontend completes OAuth flow. Frontend handles OAuth directly with Shopify, then sends the access token and shop data to this endpoint for storage.

**Authentication:** None required

**Request Body:**
```json
{
  "shop": "acme-store.myshopify.com",
  "access_token": "shpat_abc123def456...",
  "shop_name": "Acme Store",
  "meilisearch_url": "https://your-meilisearch.com",
  "meilisearch_api_key": "your-meilisearch-key"
}
```

**Request Fields:**
- `shop` (string, required) - Shopify shop domain (must end with `.myshopify.com`)
- `access_token` (string, required) - OAuth access token from Shopify (already exchanged from code)
- `shop_name` (string, optional) - Display name for the shop (defaults to shop domain)
- `meilisearch_url` (string, optional) - Meilisearch URL for this store (defaults to `MEILISEARCH_URL` env var or `X-Meilisearch-Url` header)
- `meilisearch_api_key` (string, optional) - Meilisearch API key for this store (defaults to `MEILISEARCH_API_KEY` env var or `X-Meilisearch-Api-Key` header)

**Optional Headers:**
- `X-Meilisearch-Url` (string) - Override Meilisearch URL for this store
- `X-Meilisearch-Api-Key` (string) - Override Meilisearch API key for this store

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/auth/shopify/install \
  -H "Content-Type: application/json" \
  -d '{
    "shop": "acme-store.myshopify.com",
    "access_token": "shpat_abc123def456...",
    "shop_name": "Acme Store"
  }'
```

**Response:**
```json
{
  "store": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "shop_domain": "acme-store.myshopify.com",
    "shop_name": "Acme Store",
    "product_index_uid": "acme_store_all_products",
    "meilisearch_index_uid": "acme_store_all_products",
    "meilisearch_document_type": "product",
    "plan_level": "free",
    "status": "active",
    "sync_state": {
      "status": "pending_initial_sync"
    },
    "installed_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "installation successful"
}
```

**Response Fields:**
- `store` (object) - Public store information (sensitive fields excluded)
- `token` (string) - JWT session token for authenticated requests (valid for 24 hours)
- `message` (string) - Success message

**Status Codes:**
- `200 OK` - Installation successful
- `400 Bad Request` - Invalid shop domain or missing access_token
- `500 Internal Server Error` - Storage failed

**Use in Remix (Frontend OAuth Flow):**
```typescript
// In your /auth/callback route after receiving OAuth code from Shopify
export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");
  const shop = url.searchParams.get("shop");
  
  if (!code || !shop) {
    return json({ error: "Missing OAuth parameters" }, { status: 400 });
  }

  // Option 1: Exchange code yourself
  const tokenResponse = await fetch(`https://${shop}/admin/oauth/access_token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      client_id: process.env.SHOPIFY_API_KEY,
      client_secret: process.env.SHOPIFY_API_SECRET,
      code,
    }),
  });
  const { access_token } = await tokenResponse.json();

  // Option 2: Or use backend exchange endpoint
  // const exchangeResponse = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/exchange`, {
  //   method: "POST",
  //   headers: { "Content-Type": "application/json" },
  //   body: JSON.stringify({ shop, code }),
  // });
  // const { access_token } = await exchangeResponse.json();

  // Send to backend for storage
  const installResponse = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/install`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      shop,
      access_token,
      shop_name: "My Store", // Optional
    }),
  });

  const { store, token } = await installResponse.json();
  
  // Store session token in cookie/session
  // Redirect to dashboard
  return redirect("/dashboard");
}
```

**Important Notes:**
- This endpoint is designed for frontend-handled OAuth flows
- The frontend is responsible for exchanging the OAuth code for an access token
- The backend encrypts and stores the access token securely
- Meilisearch configuration can be provided per-store or uses defaults
- `500 Internal Server Error` - Token exchange failed, encryption failed, or database error

**What Happens Behind the Scenes:**
1. Validates HMAC signature from Shopify
2. Verifies state token matches shop domain
3. Exchanges authorization code for permanent access token via Shopify API
4. Encrypts access token using AES-GCM
5. Generates public/private API keys for storefront search
6. Generates webhook secret
7. Creates/updates store record in database
8. Ensures Meilisearch index exists (`{shop}_all_products`)
9. Returns JWT session token for dashboard access

**Use in Remix:**
```typescript
// Shopify redirects to: /auth/callback?code=...&shop=...&state=...
const url = new URL(request.url);
const callbackUrl = new URL(`${GO_BACKEND_URL}/api/auth/shopify/callback`);
url.searchParams.forEach((value, key) => {
  callbackUrl.searchParams.set(key, value);
});

const response = await fetch(callbackUrl.toString());
const { token, store } = await response.json();

// Store token in cookie and redirect to dashboard
headers.append("Set-Cookie", `mgsearch_session=${token}; HttpOnly; SameSite=Lax`);
return redirect("/app", { headers });
```

---

## Store Management

### `GET /api/stores/current`

**Use Case:** Get current authenticated store information. Used by Remix dashboard to display store details, sync status, and configuration.

**Authentication:** Required - JWT session token in `Authorization: Bearer <token>` header

**Request Example:**
```bash
curl http://localhost:8080/api/stores/current \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "shop_domain": "acme-store.myshopify.com",
  "shop_name": "acme-store",
  "product_index_uid": "acme_store_all_products",
  "meilisearch_index_uid": "acme_store_all_products",
  "meilisearch_document_type": "product",
  "meilisearch_url": "https://your-cloud.meilisearch.com",
  "plan_level": "free",
  "status": "active",
  "sync_state": {
    "status": "pending_initial_sync",
    "last_full_sync": null,
    "products_count": 0
  },
  "installed_at": "2024-01-15T10:30:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Response Fields:**
- `id` (string) - Store UUID
- `shop_domain` (string) - Shopify shop domain
- `shop_name` (string) - Shop name
- `product_index_uid` (string) - Legacy index name (for backward compatibility)
- `meilisearch_index_uid` (string) - Meilisearch index name (format: `{shop}_all_products`)
- `meilisearch_document_type` (string) - Document type (default: `"product"`)
- `meilisearch_url` (string) - Meilisearch host URL for this store
- `plan_level` (string) - Subscription plan (`"free"`, `"pro"`, `"enterprise"`)
- `status` (string) - Store status (`"active"`, `"suspended"`, `"uninstalled"`)
- `sync_state` (object) - Sync status and metadata
  - `status` (string) - Current sync status
  - `last_full_sync` (string, nullable) - ISO timestamp of last full sync
  - `products_count` (number) - Number of products indexed
- `installed_at` (string) - ISO timestamp of installation
- `created_at` (string) - ISO timestamp of record creation
- `updated_at` (string) - ISO timestamp of last update

**Note:** Sensitive fields (encrypted tokens, API keys) are excluded from the response.

**Status Codes:**
- `200 OK` - Store found and returned
- `401 Unauthorized` - Missing or invalid JWT token
- `500 Internal Server Error` - Store not found or database error

**Use in Remix:**
```typescript
const store = await apiClient.getCurrentStore(request);
// Display store info, sync status, etc. in dashboard
```

---

### `GET /api/stores/sync-status`

**Use Case:** Get current sync status for authenticated store. Lightweight endpoint optimized for polling. Use this instead of `/api/stores/current` when you only need sync status updates.

**Authentication:** Required - JWT session token in `Authorization: Bearer <token>` header

**Request Example:**
```bash
curl http://localhost:8080/api/stores/sync-status \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Response:**
```json
{
  "store_id": "550e8400-e29b-41d4-a716-446655440000",
  "shop_domain": "acme-store.myshopify.com",
  "sync_state": {
    "status": "syncing",
    "last_full_sync": "2024-01-15T10:30:00Z",
    "products_count": 1250,
    "products_synced": 850,
    "products_remaining": 400,
    "sync_started_at": "2024-01-15T11:00:00Z",
    "estimated_completion": "2024-01-15T11:15:00Z",
    "errors": []
  },
  "index_uid": "acme_store_all_products",
  "document_type": "product"
}
```

**Response Fields:**
- `store_id` (string) - Store UUID
- `shop_domain` (string) - Shopify shop domain
- `sync_state` (object) - Detailed sync status
  - `status` (string) - Current sync status: `"pending_initial_sync"`, `"syncing"`, `"completed"`, `"failed"`, `"paused"`
  - `last_full_sync` (string, nullable) - ISO timestamp of last completed full sync
  - `products_count` (number, optional) - Total products to sync
  - `products_synced` (number, optional) - Number of products already synced
  - `products_remaining` (number, optional) - Number of products remaining
  - `sync_started_at` (string, optional) - ISO timestamp when current sync started
  - `estimated_completion` (string, optional) - Estimated completion time
  - `errors` (array, optional) - Array of error messages if sync failed
- `index_uid` (string) - Meilisearch index name
- `document_type` (string) - Document type (default: `"product"`)

**Status Codes:**
- `200 OK` - Sync status retrieved successfully
- `401 Unauthorized` - Missing or invalid JWT token
- `500 Internal Server Error` - Store not found or database error

**Use in Remix (Polling):**
```typescript
// Poll sync status every 2 seconds during initial sync
const pollSyncStatus = async () => {
  const response = await fetch(`${GO_BACKEND_URL}/api/stores/sync-status`, {
    headers: {
      'Authorization': `Bearer ${sessionToken}`,
    },
  });
  
  const { sync_state } = await response.json();
  
  if (sync_state.status === 'completed') {
    // Show success message
  } else if (sync_state.status === 'failed') {
    // Show error message
  } else {
    // Show progress: sync_state.products_synced / sync_state.products_count
    setTimeout(pollSyncStatus, 2000); // Poll again in 2 seconds
  }
};
```

**When to Use:**
- **Use `/api/stores/sync-status`** when you only need sync status (lighter, faster for polling)
- **Use `/api/stores/current`** when you need full store information (plan, status, all metadata)

---

## Session Storage

These endpoints are used by the Remix frontend to store and manage Shopify OAuth sessions. Sessions are stored in the backend database instead of using Prisma or in-memory storage.

### `POST /api/sessions`

**Use Case:** Store a Shopify OAuth session in the backend database. Called by Remix frontend after successful OAuth authentication.

**Authentication:** None required

**Request Body:**
```json
{
  "id": "session_id",
  "shop": "store.myshopify.com",
  "state": "state_token",
  "isOnline": false,
  "scope": "read_products,write_products",
  "expires": "2024-01-15T10:30:00Z",
  "accessToken": "encrypted_access_token",
  "userId": 123456789,
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "accountOwner": true,
  "locale": "en",
  "collaborator": false,
  "emailVerified": true
}
```

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test_session",
    "shop": "test-store.myshopify.com",
    "state": "test_state",
    "isOnline": false,
    "scope": "read_products",
    "accessToken": "test_token"
  }'
```

**Response:**
```json
{
  "message": "session stored successfully"
}
```

**Required Fields:**
- `id` (string) - Unique session identifier
- `shop` (string) - Shopify shop domain
- `state` (string) - OAuth state token
- `accessToken` (string) - Encrypted access token

**Optional Fields:**
- `isOnline` (boolean) - Whether this is an online session (default: `false`)
- `scope` (string) - OAuth scopes granted
- `expires` (string, ISO timestamp) - Session expiration time
- `userId` (integer) - Shopify user ID (for online sessions)
- `firstName`, `lastName`, `email` (string) - User information
- `accountOwner` (boolean) - Whether user is account owner
- `locale` (string) - User locale
- `collaborator` (boolean) - Whether user is a collaborator
- `emailVerified` (boolean) - Whether email is verified

**Status Codes:**
- `200 OK` - Session stored successfully
- `400 Bad Request` - Invalid session data or missing required fields
- `500 Internal Server Error` - Storage failed

**Note:** If a session with the same ID already exists, it will be updated (upsert behavior).

---

### `GET /api/sessions/{sessionId}`

**Use Case:** Retrieve a session by ID. Called by Remix frontend to load existing sessions.

**Authentication:** None required

**Path Parameters:**
- `sessionId` (string, required) - Session identifier

**Request Example:**
```bash
curl http://localhost:8080/api/sessions/test_session
```

**Response:**
```json
{
  "id": "test_session",
  "shop": "test-store.myshopify.com",
  "state": "test_state",
  "isOnline": false,
  "scope": "read_products,write_products",
  "expires": "2024-01-15T10:30:00Z",
  "accessToken": "encrypted_access_token",
  "userId": 123456789,
  "firstName": "John",
  "lastName": "Doe",
  "email": "john@example.com",
  "accountOwner": true,
  "locale": "en",
  "collaborator": false,
  "emailVerified": true,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": "2024-01-15T10:30:00Z"
}
```

**Status Codes:**
- `200 OK` - Session found and returned
- `404 Not Found` - Session not found
- `500 Internal Server Error` - Database error

---

### `DELETE /api/sessions/{sessionId}`

**Use Case:** Delete a session by ID. Called by Remix frontend when logging out or when a session expires.

**Authentication:** None required

**Path Parameters:**
- `sessionId` (string, required) - Session identifier

**Request Example:**
```bash
curl -X DELETE http://localhost:8080/api/sessions/test_session
```

**Response:**
```json
{
  "message": "session deleted successfully"
}
```

**Status Codes:**
- `200 OK` - Session deleted successfully (even if session didn't exist)
- `400 Bad Request` - Missing session ID
- `500 Internal Server Error` - Deletion failed

**Note:** Returns success even if the session doesn't exist (idempotent operation).

---

### `DELETE /api/sessions/batch`

**Use Case:** Delete multiple sessions at once. Useful for bulk cleanup operations or app uninstall webhooks.

**Authentication:** None required

**Request Body:**
```json
{
  "ids": ["session_id_1", "session_id_2", "session_id_3"]
}
```

**Request Example:**
```bash
curl -X DELETE http://localhost:8080/api/sessions/batch \
  -H "Content-Type: application/json" \
  -d '{
    "ids": ["session_1", "session_2", "session_3"]
  }'
```

**Response:**
```json
{
  "message": "sessions deleted successfully",
  "count": 3
}
```

**Status Codes:**
- `200 OK` - Sessions deleted successfully
- `400 Bad Request` - Invalid request body or empty ids array
- `500 Internal Server Error` - Deletion failed

**Use Cases:**
- App uninstall webhook: Delete all sessions for a shop
- Cleanup expired sessions
- Bulk session management

---

## Storefront Search

### `GET /api/v1/search`

**Use Case:** Public search endpoint for storefront. Called by JavaScript in merchant's theme to provide search functionality to shoppers. Uses storefront API key for authentication.

**Authentication:** Required - Storefront API key in `X-Storefront-Key` header

**Query Parameters:**
- `query` (string, optional) - Search query string (default: `""`)
- `limit` (integer, optional) - Maximum number of results (default: `20`)
- `offset` (integer, optional) - Number of results to skip (default: `0`)
- `sort` (string[], optional) - Sort criteria (e.g., `["price:asc", "title:desc"]`)
- `filters` (string, optional) - JSON-encoded filter expression

**Request Example:**
```bash
curl "http://localhost:8080/api/v1/search?query=shoes&limit=10&offset=0&sort=price:asc" \
  -H "X-Storefront-Key: abc123def456..."
```

**Request with Filters:**
```bash
curl "http://localhost:8080/api/v1/search?query=shoes&filters=%5B%22price%20%3E%3D%20100%22%2C%20%22in_stock%20%3D%20true%22%5D" \
  -H "X-Storefront-Key: abc123def456..."
```

**Response:**
```json
{
  "hits": [
    {
      "id": "123",
      "title": "Running Shoes",
      "price": 99.99,
      "vendor": "Nike",
      "tags": ["sports", "footwear"],
      "_rankingScore": 0.95
    }
  ],
  "query": "shoes",
  "processingTimeMs": 5,
  "limit": 10,
  "offset": 0,
  "estimatedTotalHits": 42,
  "facetDistribution": {
    "vendor": {
      "Nike": 15,
      "Adidas": 12,
      "Puma": 10
    }
  }
}
```

**Response Fields:**
- `hits` (array) - Search results (Meilisearch document format)
- `query` (string) - Original search query
- `processingTimeMs` (integer) - Search execution time in milliseconds
- `limit` (integer) - Results limit used
- `offset` (integer) - Offset used
- `estimatedTotalHits` (integer) - Estimated total matching documents
- `facetDistribution` (object, optional) - Facet counts if requested

**Status Codes:**
- `200 OK` - Search successful
- `401 Unauthorized` - Missing or invalid storefront key
- `500 Internal Server Error` - Search failed or store index not configured

**Use in Storefront JavaScript:**
```javascript
// In merchant's theme JavaScript
const searchInput = document.querySelector('#search-input');
const resultsContainer = document.querySelector('#search-results');

searchInput.addEventListener('input', async (e) => {
  const query = e.target.value;
  if (query.length < 2) return;

  const response = await fetch(
    `https://api.yourdomain.com/api/v1/search?query=${encodeURIComponent(query)}&limit=10`,
    {
      headers: {
        'X-Storefront-Key': '{{ store.api_key_public }}' // Injected by theme
      }
    }
  );

  const data = await response.json();
  displayResults(data.hits);
});
```

**Filter Examples:**
```javascript
// Simple filter
filters: '["price >= 100"]'

// Multiple conditions
filters: '["price >= 100", "in_stock = true", "vendor = Nike"]'

// Complex filter (JSON string)
filters: JSON.stringify([
  ["price >= 100", "price <= 500"],
  "vendor = Nike"
])
```

---

## Webhooks

### `POST /webhooks/shopify/:topic/:subtopic`

**Use Case:** Receive webhooks from Shopify when products are created, updated, or deleted. Automatically syncs changes to Meilisearch index.

**Authentication:** HMAC signature verification (Shopify signs all webhooks)

**Path Parameters:**
- `topic` (string) - Webhook topic (e.g., `"products"`)
- `subtopic` (string) - Webhook subtopic (e.g., `"create"`, `"update"`, `"delete"`)

**Headers:**
- `X-Shopify-Hmac-Sha256` (string, required) - HMAC signature for verification
- `X-Shopify-Shop-Domain` (string, required) - Shop domain that triggered the webhook
- `X-Shopify-Topic` (string, optional) - Webhook topic
- `X-Shopify-Webhook-Id` (string, optional) - Webhook ID

**Request Body:**
Product data in Shopify REST Admin API format (varies by event type).

**Supported Events:**
- `products/create` - New product created
- `products/update` - Product updated
- `products/delete` - Product deleted

**Request Example (Product Create):**
```bash
curl -X POST http://localhost:8080/webhooks/shopify/products/create \
  -H "X-Shopify-Hmac-Sha256: abc123..." \
  -H "X-Shopify-Shop-Domain: acme-store.myshopify.com" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 123456789,
    "title": "New Product",
    "vendor": "Acme Corp",
    "product_type": "Widget",
    "variants": [...],
    "images": [...]
  }'
```

**Response:**
```json
{
  "status": "processed"
}
```

**Status Codes:**
- `200 OK` - Webhook processed successfully
- `400 Bad Request` - Missing required headers
- `401 Unauthorized` - Invalid HMAC signature
- `404 Not Found` - Store not registered
- `500 Internal Server Error` - Failed to update/delete document in Meilisearch

**What Happens Behind the Scenes:**
1. Verifies HMAC signature using Shopify app secret
2. Identifies store by `X-Shopify-Shop-Domain`
3. Loads store's Meilisearch configuration
4. For `products/create` or `products/update`:
   - Transforms product data to search document
   - Adds `shop_domain`, `store_id`, `document_type` metadata
   - Upserts document in Meilisearch index
5. For `products/delete`:
   - Deletes document from Meilisearch index by product ID
6. Returns success response (Shopify expects 200 within 2 seconds)

**Webhook Registration:**
After store installation, register webhooks via Shopify Admin API:
```bash
curl -X POST "https://acme-store.myshopify.com/admin/api/2024-07/webhooks.json" \
  -H "X-Shopify-Access-Token: <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook": {
      "topic": "products/create",
      "address": "https://api.yourdomain.com/webhooks/shopify/products/create",
      "format": "json"
    }
  }'
```

---

## Legacy Search Endpoints

These endpoints are maintained for backward compatibility but are not used in the Shopify app flow.

### `POST /api/v1/clients/:client_name/:index_name/search`

**Use Case:** Generic search endpoint that forwards requests directly to Meilisearch. Not tenant-aware.

**Authentication:** None required

**Path Parameters:**
- `client_name` (string) - Client identifier (not used, kept for compatibility)
- `index_name` (string) - Meilisearch index name

**Request Body:**
Any valid Meilisearch search request (flexible JSON structure).

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/v1/clients/myclient/test_index/search \
  -H "Content-Type: application/json" \
  -d '{
    "q": "search query",
    "filter": "genre = action",
    "sort": ["release_date:desc"],
    "limit": 20,
    "offset": 0
  }'
```

**Response:**
Meilisearch search response (same format as `/api/v1/search`).

**Status Codes:**
- `200 OK` - Search successful
- `400 Bad Request` - Invalid request body or missing parameters
- `500 Internal Server Error` - Meilisearch error

---

### `POST /api/v1/clients/:client_name/:index_name/documents`

**Use Case:** Generic document indexing endpoint. Not tenant-aware.

**Authentication:** None required

**Path Parameters:**
- `client_name` (string) - Client identifier (not used)
- `index_name` (string) - Meilisearch index name

**Request Body:**
Single document object (any JSON structure).

**Request Example:**
```bash
curl -X POST http://localhost:8080/api/v1/clients/myclient/test_index/documents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "123",
    "title": "Document Title",
    "content": "Document content..."
  }'
```

**Response:**
```json
{
  "taskUid": 1,
  "indexUid": "test_index",
  "status": "enqueued",
  "type": "documentAddition",
  "enqueuedAt": "2024-01-15T10:30:00Z"
}
```

**Status Codes:**
- `202 Accepted` - Document indexing task enqueued
- `400 Bad Request` - Invalid document or missing parameters
- `500 Internal Server Error` - Meilisearch error

---

## Error Responses

All endpoints return errors in a consistent format:

```json
{
  "error": "Error message",
  "details": "Additional error details (optional)"
}
```

**Common Status Codes:**
- `400 Bad Request` - Invalid request parameters or body
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Backend service unavailable

---

## Rate Limiting

Currently, no rate limiting is implemented. Consider adding:
- Per-store rate limits for storefront search
- Global rate limits for webhook endpoints
- IP-based rate limiting for public endpoints

---

## CORS Configuration

For storefront search to work from merchant themes, ensure CORS is configured:

```go
// Add to main.go if serving from different domain
router.Use(cors.New(cors.Config{
  AllowOrigins: []string{"https://*.myshopify.com"},
  AllowMethods: []string{"GET", "POST", "OPTIONS"},
  AllowHeaders: []string{"X-Storefront-Key", "Content-Type"},
}))
```

---

## Testing Endpoints

### Using curl

```bash
# Health check
curl http://localhost:8080/ping

# Start OAuth
curl -X POST http://localhost:8080/api/auth/shopify/begin \
  -H "Content-Type: application/json" \
  -d '{"shop": "test-store.myshopify.com"}'

# Get current store (requires JWT token)
curl http://localhost:8080/api/stores/current \
  -H "Authorization: Bearer <token>"

# Storefront search (requires storefront key)
curl "http://localhost:8080/api/v1/search?query=test" \
  -H "X-Storefront-Key: <public_key>"
```

### Using Postman/Insomnia

Import the endpoints and configure:
- Environment variables for `GO_BACKEND_URL`, tokens, keys
- Pre-request scripts to set headers dynamically
- Tests to verify responses

---

## Next Steps

- Add pagination helpers for search results
- Implement search analytics endpoints
- Add webhook retry mechanism
- Create admin endpoints for store management
- Add bulk indexing endpoints for initial sync

