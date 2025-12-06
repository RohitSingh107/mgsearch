# CORS Configuration

## Overview

CORS (Cross-Origin Resource Sharing) has been configured to allow requests from Shopify storefronts to the Go search service.

## Configuration

The CORS middleware is automatically applied to all routes and allows:

### Allowed Origins
- **Shopify Storefronts**: `*.myshopify.com` (e.g., `mg-store-207095.myshopify.com`)
- **Localhost**: `http://localhost:*` and `https://localhost:*` (for development)
- **Tunnel Services**: ngrok and similar services (for development)

### Allowed Methods
- GET
- POST
- PUT
- PATCH
- DELETE
- HEAD
- OPTIONS (for preflight requests)

### Allowed Headers
- `Origin`
- `Content-Type`
- `Content-Length`
- `Authorization`
- `X-Storefront-Key` (required for search endpoints)
- `X-API-Key`
- `Accept`
- `Accept-Encoding`
- `Accept-Language`
- `X-Requested-With`

### Credentials
- `AllowCredentials: true` - Allows cookies and authentication headers

## How It Works

1. **Preflight Requests**: When a browser makes a cross-origin request, it first sends an OPTIONS request (preflight). The CORS middleware responds with appropriate headers.

2. **Actual Requests**: After preflight is approved, the actual request (GET/POST) is made with CORS headers included.

## Testing CORS

### Test with cURL

```bash
# Test preflight (OPTIONS request)
curl -X OPTIONS 'https://your-ngrok-url.ngrok-free.dev/api/v1/search' \
  -H 'Origin: https://mg-store-207095.myshopify.com' \
  -H 'Access-Control-Request-Method: POST' \
  -H 'Access-Control-Request-Headers: Content-Type,X-Storefront-Key' \
  -v
```

Expected response headers:
```
Access-Control-Allow-Origin: https://mg-store-207095.myshopify.com
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, X-Storefront-Key, ...
Access-Control-Max-Age: 43200
```

### Test Actual Request

```bash
curl -X POST 'https://your-ngrok-url.ngrok-free.dev/api/v1/search' \
  -H 'Origin: https://mg-store-207095.myshopify.com' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-key' \
  -H 'Access-Control-Request-Method: POST' \
  -d '{"q": "shows", "limit": 10}' \
  -v
```

## Common CORS Issues

### Issue: "No 'Access-Control-Allow-Origin' header"

**Solution**: CORS middleware is now configured. Restart your Go server.

### Issue: Preflight request fails

**Solution**: The middleware handles OPTIONS requests automatically. Make sure the middleware is applied before route handlers.

### Issue: Credentials not sent

**Solution**: `AllowCredentials: true` is set. Make sure your frontend includes `credentials: 'include'` in fetch options.

## Frontend Configuration

When making requests from JavaScript in the Shopify theme:

```javascript
fetch('https://your-ngrok-url.ngrok-free.dev/api/v1/search', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Storefront-Key': 'your-storefront-key'
  },
  credentials: 'include', // Include cookies if needed
  body: JSON.stringify({
    q: 'shows',
    limit: 10
  })
})
```

## Production Considerations

For production, you may want to restrict allowed origins:

```go
AllowOriginFunc: func(origin string) bool {
    // Only allow specific Shopify stores
    allowedStores := []string{
        "mg-store-207095.myshopify.com",
        "another-store.myshopify.com",
    }
    for _, store := range allowedStores {
        if origin == "https://"+store {
            return true
        }
    }
    return false
},
```

## Verification

After restarting your server, check the Network tab in browser DevTools:

1. Look for OPTIONS requests (preflight) - should return 200 OK
2. Check response headers for `Access-Control-Allow-Origin`
3. Actual requests should complete without CORS errors

