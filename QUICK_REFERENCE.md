# MGSearch - Quick Reference Guide

A condensed reference for developers working with MGSearch.

## System Overview

**MGSearch** = Dual-purpose search microservice (SaaS + Shopify)
- **Tech Stack:** Go + Gin + MongoDB + Meilisearch + Qdrant
- **Architecture:** Clean layered (Handler → Service → Repository → Database)
- **Authentication:** JWT + API Keys + Shopify OAuth

---

## Data Models Quick Reference

### Core Entities

| Entity | Primary Key | Purpose | Key Fields |
|--------|-------------|---------|------------|
| **User** | `_id` (ObjectID) | SaaS platform users | email, password_hash, client_ids[] |
| **Client** | `_id` (ObjectID) | SaaS tenants/customers | name, api_keys[], user_ids[] |
| **Index** | `_id` (ObjectID) | Search index metadata | client_id, name, uid |
| **APIKey** | `_id` (ObjectID) | Client authentication | key (hashed), permissions[], expires_at |
| **Store** | `_id` (ObjectID) | Shopify merchants | shop_domain, encrypted_token, api_key_public |
| **Session** | `_id` (string) | Shopify OAuth sessions | shop, access_token (encrypted) |

### Relationships

```
User ↔ Client (Many-to-Many via client_ids/user_ids)
Client → Index (One-to-Many)
Client → APIKey (One-to-Many, embedded)
Store → Session (One-to-Many, related by shop_domain)
```

---

## API Endpoints Cheat Sheet

### Authentication & Users

```bash
# Register new user
POST /api/v1/auth/register/user
Body: {"email", "password", "first_name", "last_name"}
Response: {user, token}

# Login
POST /api/v1/auth/login
Body: {"email", "password"}
Response: {user, token}

# Get current user
GET /api/v1/auth/me
Auth: JWT
Response: {user}

# Create client
POST /api/v1/auth/register/client
Auth: JWT
Body: {"name", "description"}
Response: {client}

# Generate API key
POST /api/v1/clients/:client_id/api-keys
Auth: JWT
Body: {"name", "permissions", "expires_at"}
Response: {api_key} (⚠️ shown once!)
```

### Client Management

```bash
# List clients
GET /api/v1/clients
Auth: JWT
Response: {clients: [...]}

# Get client details
GET /api/v1/clients/:client_id
Auth: JWT
Response: {client}

# Revoke API key
DELETE /api/v1/clients/:client_id/api-keys/:key_id
Auth: JWT
Response: {message: "revoked"}
```

### Index Management

```bash
# Create index
POST /api/v1/clients/:client_id/indexes
Auth: JWT
Body: {"name": "products", "primary_key": "id"}
Response: {index, task}

# List indexes
GET /api/v1/clients/:client_id/indexes
Auth: JWT
Response: [index1, index2, ...]

# Index document
POST /api/v1/clients/:client_id/indexes/:index_name/documents
Auth: JWT
Body: {document object}
Response: {taskUid, status}

# Update settings
PATCH /api/v1/clients/:client_id/indexes/:index_name/settings
Auth: JWT
Body: {searchableAttributes, filterableAttributes, ...}
Response: {taskUid}
```

### Search Operations

```bash
# Search (SaaS)
POST /api/v1/clients/:client_id/indexes/:index_name/search
Auth: API Key
Body: {"q": "query", "filter": "...", "limit": 20}
Response: {hits: [...], query: "...", ...}

# Get task status
GET /api/v1/clients/:client_id/tasks/:task_id
Auth: API Key
Response: {status: "succeeded", ...}

# Search (Storefront)
GET/POST /api/v1/search
Header: X-Storefront-Key
Body/Query: {"q": "shoes", "limit": 20}
Response: {hits: [...]}

# Similar products
GET/POST /api/v1/similar
Header: X-Storefront-Key
Body/Query: {"id": 123456, "limit": 10}
Response: {result: [...]}
```

### Shopify OAuth

```bash
# Begin OAuth
POST /api/auth/shopify/begin
Body: {"shop": "store.myshopify.com"}
Response: {authUrl, state}

# OAuth callback (handled by Shopify)
GET /api/auth/shopify/callback?code=...&state=...&hmac=...
Response: {store, token}

# Install store (manual)
POST /api/auth/shopify/install
Body: {"shop", "access_token", "shop_name"}
Response: {store, token}
```

### Shopify Store Management

```bash
# Get current store
GET /api/stores/current
Auth: Shopify Session JWT
Response: {store public view}

# Get sync status
GET /api/stores/sync-status
Auth: Shopify Session JWT
Response: {sync_state, index_uid, ...}
```

### Session Management

```bash
# Store session
POST /api/sessions
Auth: Optional (SESSION_API_KEY)
Body: {session object from Shopify}
Response: {success: true}
Side Effect: Auto-creates Store

# Load session
GET /api/sessions/:id
Auth: Optional
Response: {session}

# Delete session
DELETE /api/sessions/:id
Auth: Optional
Response: 204 No Content

# Find sessions by shop
GET /api/sessions/shop/:shop_domain
Auth: Optional
Response: [session1, session2, ...]
```

### Development Proxy

```bash
# Qdrant Proxy
ANY /api/dev/proxy/qdrant/*path

# Meilisearch Proxy
ANY /api/dev/proxy/meilisearch/*path
```

---

## Authentication Quick Reference

| Auth Type | Header Format | Used For | Where |
|-----------|---------------|----------|-------|
| **JWT** | `Authorization: Bearer <jwt>` | User operations | SaaS dashboard, admin |
| **API Key** | `Authorization: Bearer <key>` | Search operations | SaaS client apps |
| **Storefront Key** | `X-Storefront-Key: <key>` | Public search | Shopify storefront |
| **Shopify Session JWT** | `Authorization: Bearer <jwt>` | Store management | Shopify admin |
| **Session API Key** | `Authorization: Bearer <key>` | Session endpoints | Remix backend (optional) |
| **HMAC** | `X-Shopify-Hmac-Sha256: <sig>` | Webhook validation | Shopify webhooks |

---

## Index Naming Convention

### SaaS Platform
```
Format: {client_name}__{index_name}
Example: my_app__products, acme_corp__movies
```

### Shopify
```
Format: {shop_slug}_all_products
Example: test_store_all_products
Slug: shop domain without .myshopify.com, hyphens → underscores
```

---

## Environment Variables

### Required
```bash
MEILISEARCH_URL=http://localhost:7701
MEILISEARCH_API_KEY=master-key
DATABASE_URL=mongodb://localhost:27017/mgsearch
SHOPIFY_API_KEY=your_api_key
SHOPIFY_API_SECRET=your_secret
SHOPIFY_APP_URL=https://your-app.com
JWT_SIGNING_KEY=32-byte-hex-string
ENCRYPTION_KEY=32-byte-hex-string
```

### Optional
```bash
PORT=8080
DATABASE_MAX_CONNS=10
SHOPIFY_SCOPES=read_products,write_products,...
SHOPIFY_WEBHOOK_SECRET=webhook_secret
SESSION_API_KEY=optional_session_key
QDRANT_URL=https://qdrant.example.com
QDRANT_API_KEY=qdrant_key
```

---

## Common Operations

### SaaS: Complete Onboarding Flow

```bash
# 1. Register user
curl -X POST http://localhost:8080/api/v1/auth/register/user \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"secure123","first_name":"John","last_name":"Doe"}'
# Save the JWT token

# 2. Create client
curl -X POST http://localhost:8080/api/v1/auth/register/client \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{"name":"my-app","description":"My Application"}'
# Save client_id

# 3. Generate API key
curl -X POST http://localhost:8080/api/v1/clients/<CLIENT_ID>/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{"name":"Production Key"}'
# ⚠️ SAVE API KEY - shown only once!

# 4. Create index
curl -X POST http://localhost:8080/api/v1/clients/<CLIENT_ID>/indexes \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{"name":"products","primary_key":"id"}'

# 5. Configure index
curl -X PATCH http://localhost:8080/api/v1/clients/<CLIENT_ID>/indexes/products/settings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{
    "searchableAttributes": ["title", "description"],
    "filterableAttributes": ["category", "price"],
    "sortableAttributes": ["price", "created_at"]
  }'

# 6. Index a document
curl -X POST http://localhost:8080/api/v1/clients/<CLIENT_ID>/indexes/products/documents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT>" \
  -d '{"id":1,"title":"Product 1","price":99.99,"category":"electronics"}'

# 7. Search
curl -X POST http://localhost:8080/api/v1/clients/<CLIENT_ID>/indexes/products/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <API_KEY>" \
  -d '{"q":"product","filter":"category = electronics","limit":10}'
```

### Shopify: Installation & Search

```bash
# 1. Begin OAuth (from frontend)
curl -X POST http://localhost:8080/api/auth/shopify/begin \
  -H "Content-Type: application/json" \
  -d '{"shop":"test-store.myshopify.com"}'
# Redirect merchant to authUrl

# 2. After callback, store is created with api_key_public
# Get storefront key from dashboard

# 3. Search from storefront
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -H "X-Storefront-Key: <PUBLIC_KEY>" \
  -d '{"q":"shoes","limit":20}'

# 4. Get similar products
curl -X GET "http://localhost:8080/api/v1/similar?id=123456&limit=10" \
  -H "X-Storefront-Key: <PUBLIC_KEY>"
```

---

## Code Architecture Quick Map

```
main.go
├─ Loads config
├─ Connects to MongoDB
├─ Initializes services (Meilisearch, Qdrant, Shopify)
├─ Creates repositories (User, Client, Store, Session, Index)
├─ Creates handlers (UserAuth, Search, Storefront, Auth, etc.)
├─ Sets up middleware (JWT, APIKey, Auth, CORS)
└─ Starts Gin server

Request Flow:
Client → Middleware → Handler → Service/Repository → Database/External API
```

### Handler → Repository → Model Mapping

| Handler | Repositories Used | Models Used |
|---------|-------------------|-------------|
| UserAuthHandler | UserRepository, ClientRepository | User, Client, APIKey |
| SearchHandler | ClientRepository | Client, SearchRequest |
| StorefrontHandler | StoreRepository | Store, SearchRequest |
| AuthHandler | StoreRepository | Store |
| StoreHandler | StoreRepository | Store |
| SessionHandler | SessionRepository, StoreRepository | Session, Store |
| WebhookHandler | StoreRepository | Store |
| IndexHandler | ClientRepository, IndexRepository | Client, Index |
| SettingsHandler | ClientRepository | Client |
| TasksHandler | - | - |

---

## Security Best Practices

### Password Storage
- ✅ Bcrypt hashing (auto salt)
- ❌ Never store plaintext
- ✅ Min 8 characters enforced

### API Keys
- ✅ SHA-256 hashed in database
- ✅ Generate with crypto/rand (32 bytes)
- ✅ Show raw key only once
- ✅ Support expiration
- ✅ Track last usage

### Shopify Tokens
- ✅ AES-256-GCM encryption
- ✅ 32-byte encryption key
- ✅ Store encrypted in database
- ✅ Decrypt only when needed

### JWT Tokens
- ✅ HMAC-SHA256 signing
- ✅ 24-hour expiration
- ✅ Include user_id, email
- ✅ Verify signature on every request

### HMAC Validation
- ✅ Verify all Shopify webhooks
- ✅ Use constant-time comparison
- ✅ Reject invalid signatures immediately

---

## Debugging Tips

### Check MongoDB Collections
```bash
mongosh mongodb://localhost:27017/mgsearch
db.users.find().pretty()
db.clients.find().pretty()
db.stores.find().pretty()
db.sessions.find().pretty()
db.indexes.find().pretty()
```

### Check Meilisearch Indexes
```bash
curl http://localhost:7701/indexes \
  -H "Authorization: Bearer master-key"
```

### Decode JWT Token
```bash
# Use jwt.io or:
echo '<jwt_token>' | cut -d'.' -f2 | base64 -d | jq
```

### Test API Key
```bash
# Hash your API key
echo -n '<api_key>' | openssl dgst -sha256 -hex

# Find in MongoDB
db.clients.find({"api_keys.key": "<hashed_key>"})
```

### Verify HMAC (Webhook)
```bash
# Calculate HMAC
echo -n '<request_body>' | openssl dgst -sha256 -hmac '<shopify_secret>' -binary | base64
```

---

## Common Errors & Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `401 Unauthorized` | Invalid/expired JWT | Refresh token, login again |
| `401 invalid API key` | Wrong key or not hashed correctly | Check key in database |
| `403 Forbidden` | User doesn't have access to client | Verify user_ids in client |
| `404 Client not found` | Invalid client_id | Check client exists |
| `404 Index not found` | Index not created in Meilisearch | Create index first |
| `invalid hmac` | Webhook signature mismatch | Check SHOPIFY_API_SECRET |
| `missing storefront key` | No X-Storefront-Key header | Add header to request |
| `store index not configured` | Store missing index_uid | Complete OAuth flow |
| `CORS error` | Origin not allowed | Check CORS middleware |

---

## Development Commands

```bash
# Start dev environment (MongoDB + Redis)
just dev-up

# Stop dev environment
just dev-down

# Check service status
just dev-status

# Run the API server
go run main.go

# Format code
just fmt

# Run linter
just lint

# Run tests
just test

# Build for production
go build -o mgsearch main.go
```

---

## Testing Checklist

### SaaS Platform
- [ ] User registration works
- [ ] User login returns JWT
- [ ] JWT authentication works
- [ ] Client creation works
- [ ] API key generation works
- [ ] API key authentication works
- [ ] Index creation works
- [ ] Document indexing works
- [ ] Search returns results
- [ ] Settings update works
- [ ] Task status retrieval works

### Shopify Platform
- [ ] OAuth flow completes
- [ ] Store record created
- [ ] Meilisearch index created
- [ ] Storefront key works
- [ ] Product search works
- [ ] Similar products works
- [ ] Webhook signature validates
- [ ] Product create/update indexes
- [ ] Product delete removes from index
- [ ] Session storage works

---

## Performance Tips

1. **Index Configuration**
   - Set `searchableAttributes` to only needed fields
   - Add `filterableAttributes` for all filter fields
   - Use `sortableAttributes` for sort fields

2. **Search Optimization**
   - Use filters to narrow results
   - Limit results with `limit` parameter
   - Use pagination with `offset`

3. **API Key Management**
   - Reuse API keys across requests
   - Don't regenerate on every request
   - Set reasonable expiration dates

4. **Caching**
   - Cache search results on client side
   - Use ETags for conditional requests
   - Cache index settings

---

## Useful Links

- [Meilisearch Docs](https://docs.meilisearch.com)
- [Qdrant Docs](https://qdrant.tech/documentation)
- [Shopify API Docs](https://shopify.dev/docs/api)
- [Gin Framework](https://gin-gonic.com/docs)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current)

---

**Last Updated:** 2026-01-15
**MGSearch Version:** Current main branch
