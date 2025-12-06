# Authentication Summary

## Quick Reference

| Endpoint | Authentication Required | What You Need |
|----------|------------------------|---------------|
| `/api/v1/search` | ✅ **Storefront Key** | `X-Storefront-Key: <public-key>` |
| `/api/stores/current` | ✅ **JWT Token** | `Authorization: Bearer <jwt-token>` |
| `/api/stores/sync-status` | ✅ **JWT Token** | `Authorization: Bearer <jwt-token>` |
| `/api/sessions/*` | ⚠️ **Optional** | `Authorization: Bearer <SESSION_API_KEY>` (if set) |
| `/api/auth/shopify/*` | ❌ **None** | No authentication required |

---

## Search Endpoint (`/api/v1/search`)

**No JWT token needed!** Only requires the storefront key.

### Example Request

```bash
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: abc123def456...' \
  -d '{"q": "shows", "limit": 10}'
```

### How to Get Storefront Key

See [GET_STOREFRONT_KEY.md](./GET_STOREFRONT_KEY.md)

- Get it from store installation response
- Query database: `SELECT api_key_public FROM stores WHERE shop_domain = '...'`
- Get from `/api/stores/current` (requires JWT token)

---

## Admin Endpoints (Require JWT Token)

### `/api/stores/current`
### `/api/stores/sync-status`

These require a **JWT session token** (not the storefront key).

### Example Request

```bash
curl -X GET 'http://localhost:8080/api/stores/current' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

### How to Get JWT Token

See [GET_SESSION_TOKEN.md](./GET_SESSION_TOKEN.md)

- Get it from store installation response
- Generate manually using `scripts/generate-token.go`
- Token expires after 24 hours

---

## Key Differences

### Storefront Key (`X-Storefront-Key`)
- **Used for**: Public search endpoint (`/api/v1/search`)
- **Format**: Simple alphanumeric string (e.g., `abc123def456...`)
- **Doesn't expire**: Valid until store is re-installed
- **Safe for client-side**: Can be used in browser JavaScript
- **No JWT**: Not a JWT token

### JWT Session Token (`Authorization: Bearer`)
- **Used for**: Admin endpoints (`/api/stores/*`)
- **Format**: JWT token (e.g., `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`)
- **Expires**: After 24 hours
- **Keep secure**: Don't expose in client-side code
- **Contains**: Store ID and shop domain

---

## Quick Examples

### Search (Storefront Key Only)

```bash
# No token needed!
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'X-Storefront-Key: your-storefront-key' \
  -H 'Content-Type: application/json' \
  -d '{"q": "shoes", "limit": 10}'
```

### Get Store Info (JWT Token Required)

```bash
# Token required!
curl -X GET 'http://localhost:8080/api/stores/current' \
  -H 'Authorization: Bearer your-jwt-token'
```

---

## Summary

- **Search endpoint**: Only needs `X-Storefront-Key` header (no JWT token)
- **Admin endpoints**: Need `Authorization: Bearer <jwt-token>` header
- **Storefront key**: Get from installation response or database
- **JWT token**: Get from installation response or generate manually

