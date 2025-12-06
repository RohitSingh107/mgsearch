# Why Do We Need the Storefront Key?

## Purpose of Storefront Key

The **Storefront Key** (`X-Storefront-Key`) is required for the search endpoint to:

### 1. **Multi-Tenant Security** 🔒
- **Identifies which store** is making the request
- **Prevents cross-store data access** - Store A can't search Store B's products
- **Isolates data** - Each store only sees their own products

### 2. **Index Routing** 🎯
- **Maps to correct Meilisearch index** - Each store has its own index
- **Prevents wrong index queries** - Ensures search goes to the right store's data
- **Enables per-store configuration** - Different stores can have different search settings

### 3. **Access Control** 🛡️
- **Validates store exists** - Only registered stores can search
- **Checks store status** - Inactive stores are blocked
- **Prevents unauthorized access** - Random requests can't access search

### 4. **Usage Tracking** 📊
- **Tracks which store** is making requests
- **Enables analytics** - See which stores use search most
- **Rate limiting** - Can limit requests per store (future feature)

## How It Works

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Store     │────────▶│  Go Service  │────────▶│ Meilisearch │
│  Frontend   │ Key     │  (Validates  │ Index   │   Index     │
│             │         │   Key)       │         │             │
└─────────────┘         └──────────────┘         └─────────────┘
```

1. **Storefront sends request** with `X-Storefront-Key` header
2. **Go service validates key** - Looks up store in database
3. **Gets store's index UID** - Each store has a unique index
4. **Routes to correct index** - Searches only that store's products

## Example Flow

### Without Key (❌ Fails)
```bash
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'Content-Type: application/json' \
  -d '{"q": "shoes"}'

# Response: 401 {"error": "missing storefront key"}
```

### With Key (✅ Works)
```bash
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: b55fd4fb98715d11a3ab9120ab4caeb6' \
  -d '{"q": "shoes"}'

# Response: 200 { "hits": [...], ... }
```

## Why Not Use Shop Domain?

You might ask: "Why not just use the shop domain?"

**Problems with shop domain:**
- ❌ **Easy to spoof** - Anyone can fake a shop domain
- ❌ **No validation** - Can't verify store exists
- ❌ **No status check** - Can't block inactive stores
- ❌ **Public information** - Shop domains are visible

**Benefits of storefront key:**
- ✅ **Cryptographically random** - Hard to guess
- ✅ **Validated in database** - Must exist to work
- ✅ **Store status checked** - Inactive stores blocked
- ✅ **Private** - Only store owner knows it

## Security Model

```
Store A Key → Store A Index → Store A Products ✅
Store B Key → Store B Index → Store B Products ✅
Store A Key → Store B Index → ❌ Blocked (wrong key)
No Key      → Any Index      → ❌ Blocked (401)
```

## How to Get Your Storefront Key

### Method 1: From Database (After Store Creation)

```bash
psql $DATABASE_URL -c "SELECT shop_domain, api_key_public FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';"
```

### Method 2: From Session Storage Response

When you store a session, the store is automatically created. Query the database to get the key.

### Method 3: From Store Installation

If you installed via `/api/auth/shopify/install`, the response includes:
```json
{
  "store": {
    "api_key_public": "b55fd4fb98715d11a3ab9120ab4caeb6"
  }
}
```

## Current Store Key

For your store `mg-store-207095.myshopify.com`:

**Storefront Key:** `b55fd4fb98715d11a3ab9120ab4caeb6`

Use it like this:

```bash
curl -X POST 'http://localhost:8080/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: b55fd4fb98715d11a3ab9120ab4caeb6' \
  -d '{"q": "shows", "limit": 10}'
```

## Summary

**Why needed:**
- Security (multi-tenant isolation)
- Index routing (correct store data)
- Access control (validate store exists)
- Usage tracking (analytics)

**It's like a password** - Only stores with the correct key can search their own products.

