# Troubleshooting: "Authentication failed. Check GOLANG_API_KEY"

## Issue

You're getting the error: **"Authentication failed. Check GOLANG_API_KEY environment variable"**

This error is **NOT** coming from the Go service codebase. It's likely from:

1. **ngrok authentication** (if using ngrok tunnel)
2. **Reverse proxy** (nginx, cloudflare, etc.)
3. **Another service** in your stack
4. **Load balancer** authentication

## Solutions

### If Using ngrok

ngrok doesn't use `GOLANG_API_KEY` by default. However, if you've set up ngrok authentication:

1. **Check ngrok config**:
   ```bash
   ngrok config check
   ```

2. **Check if you have ngrok auth token set**:
   ```bash
   ngrok config add-authtoken <your-token>
   ```

3. **If using ngrok with basic auth**, you might need to:
   ```bash
   ngrok http 8080 --basic-auth="username:password"
   ```

### If Using a Reverse Proxy

If you have nginx, cloudflare, or another proxy in front:

1. **Check proxy configuration** for API key requirements
2. **Look for** `GOLANG_API_KEY` in proxy config files
3. **Check** if proxy requires authentication headers

### If This is From Another Service

If you have multiple services:

1. **Check** which service is actually returning this error
2. **Look at** the full error response headers
3. **Check** browser DevTools → Network → Response headers

## Quick Check

### 1. Test Direct Connection (Bypass Proxy)

Test directly to your Go service (not through ngrok):

```bash
curl http://localhost:8080/ping
```

If this works, the issue is with ngrok/proxy, not the Go service.

### 2. Check ngrok Status

```bash
# Check ngrok dashboard
curl http://127.0.0.1:4040/api/tunnels
```

### 3. Check Environment Variables

```bash
# Check if GOLANG_API_KEY is set (it shouldn't be needed for Go service)
echo $GOLANG_API_KEY

# Check Go service environment
cat .env | grep -i key
```

## The Go Service Doesn't Use GOLANG_API_KEY

The Go service uses:
- `SESSION_API_KEY` - Optional, for session endpoints only
- `X-Storefront-Key` - Required for `/api/v1/search` (this is the storefront key, not an env var)
- JWT tokens - For authenticated endpoints

**The Go service does NOT check for `GOLANG_API_KEY`.**

## Next Steps

1. **Identify the source**: Check where the error is coming from
   - Browser DevTools → Network → Response
   - Server logs
   - ngrok logs

2. **If from ngrok**: Check ngrok configuration
3. **If from proxy**: Check proxy configuration
4. **If from another service**: Check that service's configuration

## Common Scenarios

### Scenario 1: ngrok Free Plan Limits

ngrok free plan might require authentication. Check your ngrok account.

### Scenario 2: Cloudflare/Proxy

If behind Cloudflare or another proxy, they might require API keys.

### Scenario 3: Multiple Services

If you have multiple services, one of them might be checking for `GOLANG_API_KEY`.

## Still Not Working?

1. **Check the full error response** in browser DevTools
2. **Check server logs** for any authentication middleware
3. **Test with curl** to see the exact error:
   ```bash
   curl -v https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search \
     -H 'X-Storefront-Key: your-key' \
     -H 'Content-Type: application/json' \
     -d '{"q":"test"}'
   ```

