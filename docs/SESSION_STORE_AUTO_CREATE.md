# Automatic Store Creation from Sessions

## Overview

When a session is stored via `POST /api/sessions`, the system **automatically creates or updates a store** with all the necessary details. This ensures that stores are always in sync with sessions.

## How It Works

1. **Session is stored** via `POST /api/sessions`
2. **Store is automatically created/updated** using:
   - Shop domain from session
   - Access token from session (decrypted and re-encrypted for store)
   - Default values for other fields

## What Gets Created

When a session is stored, a store is created with:

- **Shop Domain**: From `session.shop`
- **Shop Name**: Auto-generated from shop domain (e.g., "mg store" from "mg-store-207095.myshopify.com")
- **Access Token**: Encrypted Shopify access token from session
- **API Keys**: Auto-generated public and private keys
- **Index UID**: Auto-generated Meilisearch index name
- **Meilisearch Config**: Uses values from environment/config
- **Status**: Set to "active"
- **Plan Level**: Set to "free"
- **Sync State**: Set to "pending_initial_sync"

## Behavior

### New Store (Doesn't Exist)
- Creates a new store with all default values
- Generates new API keys
- Creates Meilisearch index if configured

### Existing Store
- Updates the access token (if changed)
- Updates the `updated_at` timestamp
- Keeps existing API keys and configuration

## Example

### Store a Session

```bash
curl -X POST 'http://localhost:8080/api/sessions' \
  -H 'Content-Type: application/json' \
  -d '{
    "id": "mg-store-207095.myshopify.com_1234567890",
    "shop": "mg-store-207095.myshopify.com",
    "state": "random-state-string",
    "isOnline": false,
    "accessToken": "shpat_abc123...",
    "scope": "read_products,write_products"
  }'
```

**What happens:**
1. ✅ Session is stored in `sessions` table
2. ✅ Store is automatically created/updated in `stores` table
3. ✅ Storefront key is generated and available

### Response

```json
{
  "success": true,
  "message": "Session stored successfully, store created/updated"
}
```

## Store Details Created

After storing a session, you can query the store:

```bash
# Get store by domain
psql $DATABASE_URL -c "SELECT id, shop_domain, api_key_public FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';"
```

You'll get:
- **Store ID (UUID)**: For generating JWT tokens
- **Storefront Key**: For search endpoint (`X-Storefront-Key`)
- **All other store configuration**

## Use Cases

### 1. Remix App OAuth Flow

When the Remix app completes OAuth and stores a session:

```javascript
// Remix stores session
await fetch('http://localhost:8080/api/sessions', {
  method: 'POST',
  body: JSON.stringify(sessionData)
});

// Store is automatically created!
// No need to call /api/auth/shopify/install separately
```

### 2. Direct Session Storage

You can store sessions directly, and stores will be created automatically:

```bash
POST /api/sessions
{
  "id": "...",
  "shop": "mg-store-207095.myshopify.com",
  "accessToken": "shpat_...",
  ...
}
```

## Important Notes

1. **Idempotent**: Storing the same session multiple times is safe
2. **Store Updates**: If store exists, only access token is updated
3. **Error Handling**: If store creation fails, session is still stored (with warning)
4. **Meilisearch Index**: Index is created automatically if Meilisearch is configured
5. **API Keys**: New keys are generated only for new stores

## Configuration Required

For store creation to work fully, ensure these are set in `.env`:

```bash
MEILISEARCH_URL=https://your-meilisearch-url.com
MEILISEARCH_API_KEY=your-key
ENCRYPTION_KEY=32-byte-hex-string
```

If Meilisearch config is missing, store will still be created but index won't be initialized.

## Verification

After storing a session, verify the store was created:

```bash
# Check stores table
psql $DATABASE_URL -c "SELECT shop_domain, api_key_public, status FROM stores;"

# Check sessions table
psql $DATABASE_URL -c "SELECT shop, is_online, created_at FROM sessions;"
```

Both should show the new entries.

## Troubleshooting

### Store Not Created

- Check server logs for errors
- Verify `shop` field is a valid `.myshopify.com` domain
- Check that `accessToken` is provided
- Ensure database connection is working

### Store Created But Missing Fields

- Check that Meilisearch config is set (for index creation)
- Verify encryption key is valid (32-byte hex)
- Check server logs for warnings

