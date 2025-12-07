# Authentication Types and API Usage

This document explains all authentication mechanisms used in MGSearch and which APIs require authentication.

## Authentication Types Overview

MGSearch uses **4 different authentication mechanisms** depending on the endpoint:

1. **JWT Session Tokens** - For admin/dashboard endpoints
2. **Storefront API Keys** - For public storefront search
3. **Optional API Keys** - For session management endpoints
4. **HMAC Signature Verification** - For Shopify webhooks

---

## 1. JWT Session Tokens (Bearer Token)

### Type
**JWT (JSON Web Token)** using HS256 algorithm

### How It Works
- Tokens are generated after successful OAuth installation
- Contains `store_id` and `shop_domain` claims
- Signed with `JWT_SIGNING_KEY` (32-byte hex string)
- Valid for 24 hours by default
- Uses `Authorization: Bearer <token>` header format

### Token Structure
```json
{
  "store_id": "507f1f77bcf86cd799439011",
  "shop": "example-store.myshopify.com",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Implementation
- **Middleware**: `middleware.AuthMiddleware.RequireStoreSession()`
- **Package**: `pkg/auth/session.go`
- **Algorithm**: HS256 (HMAC-SHA256)

### APIs Using JWT Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/stores/current` | GET | Get authenticated store information |
| `/api/stores/sync-status` | GET | Get store sync status |

### Example Request
```bash
curl -X GET http://localhost:8080/api/stores/current \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### How to Get a JWT Token
1. Complete OAuth installation flow (`POST /api/auth/shopify/install`)
2. Response includes a `token` field with the JWT
3. Use this token in subsequent requests

---

## 2. Storefront API Keys (X-Storefront-Key Header)

### Type
**API Key** passed via custom header

### How It Works
- Each store has a unique public API key (`api_key_public`)
- Key is generated during store installation
- Used for public storefront search (no user authentication required)
- Validates that requests come from authorized storefronts
- Uses `X-Storefront-Key: <key>` header format

### Implementation
- **Handler**: `handlers.StorefrontHandler.Search()`
- **Validation**: Looks up store by `api_key_public` in database
- **No middleware**: Validation happens directly in handler

### APIs Using Storefront Key Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/search` | GET | Public storefront search (query params) |
| `/api/v1/search` | POST | Public storefront search (JSON body) |

### Example Request
```bash
curl -X GET "http://localhost:8080/api/v1/search?q=shoes&limit=10" \
  -H "X-Storefront-Key: abc123def456..."
```

### How to Get Storefront Key
1. Install store via OAuth (`POST /api/auth/shopify/install`)
2. Response includes `store.api_key_public`
3. Or query store info with JWT token: `GET /api/stores/current`
4. Response includes `api_key_public` field

---

## 3. Optional API Keys (Bearer Token)

### Type
**Static API Key** (optional, configurable)

### How It Works
- Configured via `SESSION_API_KEY` environment variable
- If set, all session endpoints require this key
- If not set, session endpoints are publicly accessible
- Uses `Authorization: Bearer <api-key>` header format
- Simple string comparison (not JWT)

### Implementation
- **Middleware**: `middleware.OptionalAPIKeyMiddleware()`
- **Config**: `SESSION_API_KEY` environment variable
- **Behavior**: 
  - If `SESSION_API_KEY` is empty → No authentication required
  - If `SESSION_API_KEY` is set → Requires exact match

### APIs Using Optional API Key Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/sessions` | POST | Store Shopify OAuth session |
| `/api/sessions/:id` | GET | Load session by ID |
| `/api/sessions/:id` | DELETE | Delete session by ID |
| `/api/sessions/batch` | DELETE | Delete multiple sessions |
| `/api/sessions/shop/:shop` | GET | Find sessions by shop domain |

### Example Request (when SESSION_API_KEY is set)
```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Authorization: Bearer your-session-api-key" \
  -H "Content-Type: application/json" \
  -d '{"id": "session-123", "shop": "example.myshopify.com", ...}'
```

### Configuration
```bash
# In .env file
SESSION_API_KEY=your-secret-api-key-here
```

**Note**: If `SESSION_API_KEY` is not set, these endpoints are **publicly accessible** (no authentication required).

---

## 4. HMAC Signature Verification (Shopify Webhooks)

### Type
**HMAC-SHA256** signature verification

### How It Works
- Shopify sends webhooks with `X-Shopify-Hmac-Sha256` header
- Signature is HMAC-SHA256 of request body using webhook secret
- Verifies webhook authenticity and integrity
- Uses `X-Shopify-Hmac-Sha256: <signature>` and `X-Shopify-Shop-Domain: <domain>` headers

### Implementation
- **Handler**: `handlers.WebhookHandler.HandleShopifyWebhook()`
- **Service**: `services.ShopifyService.VerifyWebhookSignature()`
- **Algorithm**: HMAC-SHA256
- **Secret**: `SHOPIFY_WEBHOOK_SECRET` or store-specific `webhook_secret`

### APIs Using HMAC Signature Verification

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/webhooks/shopify/:topic/:subtopic` | POST | Receive Shopify webhooks |

### Example Request (from Shopify)
```http
POST /webhooks/shopify/products/create HTTP/1.1
Host: your-app.com
X-Shopify-Shop-Domain: example-store.myshopify.com
X-Shopify-Hmac-Sha256: abc123def456...
Content-Type: application/json

{"id": 12345, "title": "Product Name", ...}
```

### Verification Process
1. Extract `X-Shopify-Hmac-Sha256` header
2. Extract `X-Shopify-Shop-Domain` header
3. Read request body
4. Compute HMAC-SHA256(body, webhook_secret)
5. Compare computed signature with header value
6. If match → webhook is authentic

---

## No Authentication Required

These endpoints are publicly accessible:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ping` | GET | Health check |
| `/` | GET | Root endpoint |
| `/api/auth/shopify/begin` | POST | Start OAuth flow |
| `/api/auth/shopify/callback` | GET | OAuth callback |
| `/api/auth/shopify/exchange` | POST | Exchange OAuth code |
| `/api/auth/shopify/install` | POST | Install store (requires headers for Meilisearch) |
| `/api/v1/clients/:client/:index/search` | POST | Legacy search (no auth) |
| `/api/v1/clients/:client/:index/documents` | POST | Legacy indexing (no auth) |
| `/api/v1/clients/:client/:index/settings` | PATCH | Legacy settings (no auth) |
| `/api/v1/clients/:client/tasks/:task_id` | GET | Legacy tasks (no auth) |

**Note**: The legacy endpoints (`/api/v1/clients/*`) don't require authentication but may be deprecated in favor of authenticated endpoints.

---

## Authentication Summary Table

| Endpoint | Auth Type | Header Format | Required? |
|----------|-----------|---------------|-----------|
| `/api/stores/current` | JWT Token | `Authorization: Bearer <jwt>` | ✅ Yes |
| `/api/stores/sync-status` | JWT Token | `Authorization: Bearer <jwt>` | ✅ Yes |
| `/api/v1/search` | Storefront Key | `X-Storefront-Key: <key>` | ✅ Yes |
| `/api/sessions/*` | Optional API Key | `Authorization: Bearer <key>` | ⚠️ If `SESSION_API_KEY` set |
| `/webhooks/shopify/*` | HMAC Signature | `X-Shopify-Hmac-Sha256: <sig>` | ✅ Yes |
| `/ping`, `/api/auth/*` | None | - | ❌ No |

---

## Security Best Practices

1. **JWT Tokens**:
   - Use HTTPS in production
   - Tokens expire after 24 hours
   - Store `JWT_SIGNING_KEY` securely (32-byte hex)

2. **Storefront Keys**:
   - Keys are public but unique per store
   - Can be rotated by reinstalling store
   - Use HTTPS to prevent key interception

3. **Session API Keys**:
   - Set `SESSION_API_KEY` in production
   - Use strong, random keys
   - Rotate periodically

4. **Webhook Signatures**:
   - Always verify HMAC signatures
   - Use store-specific webhook secrets
   - Never trust webhooks without verification

---

## Getting Authentication Credentials

### JWT Token
```bash
# After OAuth installation
POST /api/auth/shopify/install
# Response: { "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." }
```

### Storefront Key
```bash
# Get from store info
GET /api/stores/current
Authorization: Bearer <jwt-token>
# Response: { "api_key_public": "abc123..." }
```

### Session API Key
```bash
# Set in environment
export SESSION_API_KEY=your-secret-key
```

### Webhook Secret
```bash
# Set in environment or store-specific
export SHOPIFY_WEBHOOK_SECRET=your-webhook-secret
```

---

## Troubleshooting

### "missing authorization header"
- **JWT endpoints**: Include `Authorization: Bearer <token>` header
- **Session endpoints**: Include `Authorization: Bearer <api-key>` if `SESSION_API_KEY` is set

### "invalid token"
- JWT token may be expired (24 hour TTL)
- JWT signing key mismatch
- Token format incorrect

### "missing storefront key"
- Include `X-Storefront-Key: <key>` header
- Verify key is correct for the store

### "invalid webhook signature"
- Webhook secret mismatch
- Request body was modified
- Signature header missing or incorrect

---

## Code References

- **JWT Implementation**: `pkg/auth/session.go`
- **JWT Middleware**: `middleware/auth_middleware.go`
- **Storefront Key Handler**: `handlers/storefront.go` (line 42-52)
- **Optional API Key Middleware**: `middleware/api_key_middleware.go`
- **Webhook Verification**: `services/shopify.go` → `VerifyWebhookSignature()`

