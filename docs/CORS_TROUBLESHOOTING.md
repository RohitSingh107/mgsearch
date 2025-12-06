# CORS Troubleshooting Guide

## Quick Fix Checklist

1. ✅ **Restart your Go server** after adding CORS middleware
2. ✅ **Clear browser cache** or use incognito mode
3. ✅ **Check the Origin header** in browser DevTools Network tab
4. ✅ **Verify preflight (OPTIONS) requests** are returning 200/204

## Testing CORS

### 1. Test Preflight Request

```bash
curl -X OPTIONS 'https://your-ngrok-url.ngrok-free.dev/api/v1/search' \
  -H 'Origin: https://mg-store-207095.myshopify.com' \
  -H 'Access-Control-Request-Method: POST' \
  -H 'Access-Control-Request-Headers: Content-Type,X-Storefront-Key' \
  -v
```

**Expected Response:**
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://mg-store-207095.myshopify.com
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, X-Storefront-Key, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 43200
```

### 2. Test Actual Request

```bash
curl -X POST 'https://your-ngrok-url.ngrok-free.dev/api/v1/search' \
  -H 'Origin: https://mg-store-207095.myshopify.com' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-key' \
  -d '{"q": "shows", "limit": 10}' \
  -v
```

**Expected Response:**
```
HTTP/1.1 200 OK
Access-Control-Allow-Origin: https://mg-store-207095.myshopify.com
Access-Control-Allow-Credentials: true
Content-Type: application/json
...
```

## Common Issues

### Issue: Still getting CORS errors after restart

**Solution:**
1. Make sure you **restarted the server** after adding CORS middleware
2. Check server logs to see if CORS middleware is being applied
3. Verify the middleware is added **before** route handlers in `main.go`

### Issue: Preflight returns 404

**Solution:**
- The handler now explicitly handles OPTIONS requests
- Make sure the route is registered: `v1.GET("/search", ...)` and `v1.POST("/search", ...)`

### Issue: "Access-Control-Allow-Origin" header missing

**Solution:**
- CORS middleware should add this automatically
- Handler also adds it explicitly as fallback
- Check browser DevTools → Network → Headers tab

### Issue: Credentials not working

**Solution:**
- `AllowCredentials: true` is set in CORS config
- Frontend must include `credentials: 'include'` in fetch options
- Cannot use `Access-Control-Allow-Origin: *` with credentials

## Debug Steps

### 1. Check Server Logs

Add logging to see what origins are being checked:

```go
// In middleware/cors_middleware.go
AllowOriginFunc: func(origin string) bool {
    log.Printf("CORS check for origin: %s", origin)
    // ... rest of logic
}
```

### 2. Check Browser DevTools

1. Open DevTools → Network tab
2. Filter by "search"
3. Click on the failed request
4. Check:
   - **Request Headers**: Look for `Origin` header
   - **Response Headers**: Look for `Access-Control-Allow-Origin`
   - **Status**: Should be 200/204, not 404/500

### 3. Verify Origin Format

The origin from Shopify will be:
```
https://mg-store-207095.myshopify.com
```

The middleware checks for `.myshopify.com` in the origin string, which should match.

## Force Allow All Origins (Development Only)

If you're still having issues, temporarily allow all origins:

```go
AllowOriginFunc: func(origin string) bool {
    return true // Allow all origins (development only!)
},
```

**⚠️ WARNING**: Only use this for development. Never in production!

## Production Configuration

For production, restrict to specific stores:

```go
AllowOriginFunc: func(origin string) bool {
    allowedStores := []string{
        "https://mg-store-207095.myshopify.com",
        "https://another-store.myshopify.com",
    }
    for _, allowed := range allowedStores {
        if origin == allowed {
            return true
        }
    }
    return false
},
```

## Still Not Working?

1. **Check ngrok headers**: Some ngrok versions add extra headers that might interfere
2. **Check reverse proxy**: If behind nginx/cloudflare, they might strip CORS headers
3. **Check browser console**: Look for specific CORS error messages
4. **Test with Postman/curl**: Bypass browser to verify server is working

