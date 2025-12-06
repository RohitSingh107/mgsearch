# How to Install a Store and Get Your Tokens

Before you can generate tokens or use the API, you need to **install a store** first. This creates a store record in the database with a UUID.

## Install Store Endpoint

```bash
POST /api/auth/shopify/install
```

### Request

```bash
curl -X POST 'http://localhost:8080/api/auth/shopify/install' \
  -H 'Content-Type: application/json' \
  -d '{
    "shop": "mg-store-207095.myshopify.com",
    "access_token": "shpat_your_shopify_access_token_here",
    "shop_name": "My Store"
  }'
```

### Response

```json
{
  "store": {
    "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",  // <-- UUID store ID
    "shop_domain": "mg-store-207095.myshopify.com",
    "api_key_public": "abc123def456...",  // <-- Storefront key for search
    "shop_name": "My Store",
    ...
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",  // <-- JWT session token
  "message": "installation successful"
}
```

## What You Get

After installation, you'll receive:

1. **Store ID (UUID)** - Use this to generate new tokens later
2. **Storefront Key** (`api_key_public`) - For `/api/v1/search` endpoint
3. **JWT Session Token** - For `/api/stores/current` endpoint (valid 24 hours)

## Using the Tokens

### Storefront Key (for search)

```bash
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'X-Storefront-Key: abc123def456...' \
  -H 'Content-Type: application/json' \
  -d '{"q": "shows", "limit": 10}'
```

### JWT Token (for admin endpoints)

```bash
curl -X GET 'http://localhost:8080/api/stores/current' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

## Generate New Token Later

Once you have the store UUID from installation, you can generate new tokens:

```bash
# Get store UUID from database
psql $DATABASE_URL -c "SELECT id, shop_domain FROM stores;"

# Generate token with the UUID
go run scripts/generate-token.go <store-uuid> mg-store-207095.myshopify.com
```

## Important Notes

- **Store ID is a UUID** - Not a simple number like "1"
- **Install store first** - You can't generate tokens for stores that don't exist
- **UUID format** - Looks like: `a1b2c3d4-e5f6-7890-abcd-ef1234567890`
- **One store per shop** - Installing the same shop again will update the existing store

## Quick Start

1. **Install your store:**
   ```bash
   curl -X POST 'http://localhost:8080/api/auth/shopify/install' \
     -H 'Content-Type: application/json' \
     -d '{
       "shop": "mg-store-207095.myshopify.com",
       "access_token": "your-shopify-token"
     }'
   ```

2. **Save the response:**
   - Store ID (UUID)
   - Storefront key
   - JWT token

3. **Use them:**
   - Storefront key → Search endpoint
   - JWT token → Admin endpoints

