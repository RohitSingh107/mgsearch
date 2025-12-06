# Remix Integration Quick Start Checklist

## Prerequisites
- [ ] Remix app created with Shopify CLI
- [ ] Go backend running and accessible
- [ ] Shopify Partners app credentials

## Step-by-Step Setup

### 1. Configure Environment Variables

**Remix `.env`:**
```env
GO_BACKEND_URL=http://localhost:8080
# Keep existing SHOPIFY_API_KEY, SHOPIFY_API_SECRET, etc.
```

**Go Backend `.env`:**
```env
SHOPIFY_API_KEY=<same as Remix>
SHOPIFY_API_SECRET=<same as Remix>
SHOPIFY_APP_URL=<your ngrok URL or production domain>
```

### 2. Create Remix Routes

- [ ] `app/routes/auth.install.tsx` - OAuth initiation
- [ ] `app/routes/auth.callback.tsx` - OAuth callback handler
- [ ] `app/routes/app._index.tsx` - Dashboard
- [ ] `app/lib/api.client.ts` - API client utility

### 3. Update Shopify App Settings

- [ ] Set App URL in Partners dashboard
- [ ] Set Allowed redirection URLs: `https://your-domain.com/auth/callback`
- [ ] Enable required scopes

### 4. Test Integration

- [ ] Start Go backend: `go run main.go`
- [ ] Start Remix: `npm run dev`
- [ ] Visit `/auth/install` in Remix
- [ ] Complete OAuth flow
- [ ] Verify dashboard loads with store data

### 5. Verify Backend

- [ ] Check Go backend logs for store creation
- [ ] Verify database has new store record
- [ ] Test API endpoint: `curl http://localhost:8080/api/stores/current` (with auth header)

## Common Issues

**OAuth fails:**
- Check `SHOPIFY_APP_URL` matches Partners dashboard
- Verify callback URL is whitelisted

**Session not working:**
- Check cookie settings
- Verify JWT signing key matches

**API calls fail:**
- Check `GO_BACKEND_URL` is correct
- Verify CORS is configured if on different domains

## Next Steps

See `REMIX_INTEGRATION.md` for complete implementation details.

