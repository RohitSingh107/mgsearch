# How to Get Your Storefront API Key

The **Storefront API Key** (also called `X-Storefront-Key`) is required to make search requests from your Shopify storefront.

**Important**: This is different from the JWT session token. The search endpoint (`/api/v1/search`) does NOT require a JWT token - only the storefront key.

Here's how to get it:

## Method 1: From API Response (After Store Installation)

When you install a store via `POST /api/auth/shopify/install`, the response includes the storefront key:

```json
{
  "store": {
    "id": "...",
    "shop_domain": "mg-store-207095.myshopify.com",
    "api_key_public": "abc123def456..."  // <-- This is your storefront key
  },
  "token": "...",
  "message": "installation successful"
}
```

**Save this key** - you'll need it for storefront search requests.

---

## Method 2: From Authenticated Store Endpoint

If you're already authenticated, call:

```bash
GET /api/stores/current
Authorization: Bearer <your-session-token>
```

Response includes:
```json
{
  "id": "...",
  "shop_domain": "mg-store-207095.myshopify.com",
  "api_key_public": "abc123def456..."  // <-- Your storefront key
}
```

---


---

## Method 4: Check Installation Response

If you just installed the store, the key was in the installation response. Check your logs or API client response.

---

## Using the Key

Once you have the key, use it in the `X-Storefront-Key` header:

```bash
curl -X POST 'https://your-ngrok-url.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: abc123def456...' \
  -d '{"q": "shows", "limit": 10}'
```

Or in JavaScript:

```javascript
fetch('https://your-ngrok-url.ngrok-free.dev/api/v1/search', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Storefront-Key': 'abc123def456...'  // Your storefront key
  },
  body: JSON.stringify({
    q: 'shows',
    limit: 10
  })
})
```

---

## Important Notes

1. **One key per store**: Each store has a unique storefront key
2. **Public key**: This key is safe to use in client-side JavaScript (it's public)
3. **Don't confuse with private key**: The `api_key_private` is for admin operations only
4. **Key format**: Usually 16-32 character alphanumeric string

---

---



