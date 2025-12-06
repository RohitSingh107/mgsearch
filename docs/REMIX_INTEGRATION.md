# Remix Frontend Integration Guide

Complete implementation plan for connecting your Shopify Remix app to the Go backend.

## Architecture Overview

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────────┐
│   Shopify       │         │   Remix Frontend │         │   Go Backend    │
│   Admin/Store   │◄───────►│   (OAuth UI)     │◄───────►│   (API + DB)    │
└─────────────────┘         └──────────────────┘         └─────────────────┘
                                      │                            │
                                      │                            │
                                      ▼                            ▼
                              ┌──────────────────┐         ┌─────────────────┐
                              │   Storefront     │         │  Meilisearch    │
                              │   (Theme)        │         │  (Cloud)        │
                              └──────────────────┘         └─────────────────┘
```

## Integration Flow

### 1. OAuth Installation Flow
- **Remix**: Handles OAuth UI and redirects
- **Go Backend**: Stores encrypted tokens, creates store records, generates API keys
- **Communication**: Remix → Go backend via REST API

### 2. Admin Dashboard
- **Remix**: Renders merchant dashboard UI
- **Go Backend**: Provides store data, sync status, search analytics
- **Authentication**: JWT session tokens from Go backend

### 3. Storefront Search
- **Remix/Theme**: Injects search UI into merchant's store
- **Go Backend**: Public search endpoint with storefront API keys
- **No Remix involvement**: Direct storefront → Go backend

## Implementation Steps

### Step 1: Environment Configuration

#### Remix `.env` (keep existing Shopify CLI config)
```env
# Shopify App Credentials (from Partners dashboard)
SHOPIFY_API_KEY=your_api_key
SHOPIFY_API_SECRET=your_api_secret
SHOPIFY_SCOPES=read_products,write_products,read_product_listings,read_collection_listings,read_inventory,write_webhooks
SHOPIFY_APP_URL=https://your-ngrok-url.ngrok.io

# Go Backend URL
GO_BACKEND_URL=http://localhost:8080
# Or in production:
# GO_BACKEND_URL=https://api.yourdomain.com
```

#### Go Backend `.env` (already configured)
```env
SHOPIFY_API_KEY=your_api_key  # Same as Remix
SHOPIFY_API_SECRET=your_api_secret  # Same as Remix
SHOPIFY_APP_URL=https://your-ngrok-url.ngrok.io  # Same as Remix
```

### Step 2: Create Remix API Routes

#### 2.1 OAuth Installation Handler

**File: `app/routes/auth.install.tsx`**

```typescript
import { json, redirect } from "@remix-run/node";
import { useLoaderData } from "@remix-run/react";
import type { ActionFunctionArgs, LoaderFunctionArgs } from "@remix-run/node";

const GO_BACKEND_URL = process.env.GO_BACKEND_URL || "http://localhost:8080";

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const shop = url.searchParams.get("shop");

  if (!shop) {
    return json({ error: "Shop parameter required" }, { status: 400 });
  }

  // Construct redirect URI for Remix callback route
  const appUrl = new URL(request.url).origin;
  const redirectUri = `${appUrl}/auth/callback`;

  // Call Go backend to initiate OAuth
  try {
    const response = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/begin`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ shop, redirect_uri: redirectUri }),
    });

    if (!response.ok) {
      const error = await response.json();
      return json({ error: error.error || "Failed to start OAuth" }, { status: 500 });
    }

    const data = await response.json();
    return json({ authUrl: data.authUrl, state: data.state });
  } catch (error) {
    return json({ error: "Backend unavailable" }, { status: 503 });
  }
}

export async function action({ request }: ActionFunctionArgs) {
  const formData = await request.formData();
  const shop = formData.get("shop") as string;

  if (!shop) {
    return json({ error: "Shop required" }, { status: 400 });
  }

  // Construct redirect URI for Remix callback route
  const appUrl = new URL(request.url).origin;
  const redirectUri = `${appUrl}/auth/callback`;

  // Redirect to Go backend's auth URL
  const response = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/begin`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ shop, redirect_uri: redirectUri }),
  });

  if (!response.ok) {
    return json({ error: "Failed to start OAuth" }, { status: 500 });
  }

  const { authUrl } = await response.json();
  return redirect(authUrl);
}

export default function Install() {
  const data = useLoaderData<typeof loader>();

  if ("error" in data) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold text-red-600">Error</h1>
        <p>{data.error}</p>
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">Install App</h1>
      <form method="post">
        <input
          type="text"
          name="shop"
          placeholder="your-store.myshopify.com"
          className="border p-2 mr-2"
          required
        />
        <button
          type="submit"
          className="bg-blue-600 text-white px-4 py-2 rounded"
        >
          Install
        </button>
      </form>
    </div>
  );
}
```

#### 2.2 OAuth Callback Handler

**File: `app/routes/auth.callback.tsx`**

```typescript
import { json, redirect } from "@remix-run/node";
import { useLoaderData } from "@remix-run/react";
import type { LoaderFunctionArgs } from "@remix-run/node";

const GO_BACKEND_URL = process.env.GO_BACKEND_URL || "http://localhost:8080";

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  
  // Forward all query params to Go backend
  const callbackUrl = new URL(`${GO_BACKEND_URL}/api/auth/shopify/callback`);
  url.searchParams.forEach((value, key) => {
    callbackUrl.searchParams.set(key, value);
  });

  try {
    const response = await fetch(callbackUrl.toString(), {
      method: "GET",
      headers: { "Accept": "application/json" },
    });

    if (!response.ok) {
      const error = await response.json();
      return json({ 
        error: error.error || "OAuth callback failed",
        details: error.details 
      }, { status: response.status });
    }

    const data = await response.json();
    
    // Store session token in cookie
    const headers = new Headers();
    headers.append(
      "Set-Cookie",
      `mgsearch_session=${data.token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${24 * 60 * 60}`
    );

    // Redirect to dashboard
    return redirect("/app", { headers });
  } catch (error) {
    return json({ 
      error: "Backend unavailable",
      details: error instanceof Error ? error.message : "Unknown error"
    }, { status: 503 });
  }
}

export default function Callback() {
  const data = useLoaderData<typeof loader>();

  if ("error" in data) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold text-red-600">Installation Failed</h1>
        <p>{data.error}</p>
        {data.details && <p className="text-sm text-gray-600">{data.details}</p>}
      </div>
    );
  }

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold">Installing...</h1>
      <p>Please wait while we set up your store.</p>
    </div>
  );
}
```

#### 2.3 Alternative: Frontend-Handled OAuth Flow

**Recommended Approach:** Handle OAuth directly in Remix, then send data to backend for storage.

**File: `app/routes/auth.callback.tsx` (Alternative Implementation)**

```typescript
import { json, redirect } from "@remix-run/node";
import type { LoaderFunctionArgs } from "@remix-run/node";

const GO_BACKEND_URL = process.env.GO_BACKEND_URL || "http://localhost:8080";
const SHOPIFY_API_KEY = process.env.SHOPIFY_API_KEY!;
const SHOPIFY_API_SECRET = process.env.SHOPIFY_API_SECRET!;

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const code = url.searchParams.get("code");
  const shop = url.searchParams.get("shop");
  const hmac = url.searchParams.get("hmac");
  const state = url.searchParams.get("state");

  if (!code || !shop || !hmac) {
    return json({ error: "Missing OAuth parameters" }, { status: 400 });
  }

  // Validate HMAC (optional but recommended)
  // You can use a library like @shopify/shopify-api for this

  try {
    // Exchange code for access token directly with Shopify
    const tokenResponse = await fetch(`https://${shop}/admin/oauth/access_token`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        client_id: SHOPIFY_API_KEY,
        client_secret: SHOPIFY_API_SECRET,
        code,
      }),
    });

    if (!tokenResponse.ok) {
      return json({ error: "Token exchange failed" }, { status: 500 });
    }

    const { access_token } = await tokenResponse.json();

    // Send to backend for storage
    const installResponse = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/install`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        shop,
        access_token,
        shop_name: shop.split(".")[0], // Optional
      }),
    });

    if (!installResponse.ok) {
      const error = await installResponse.json();
      return json({ error: error.error || "Installation failed" }, { status: installResponse.status });
    }

    const { store, token } = await installResponse.json();

    // Store session token in cookie
    const headers = new Headers();
    headers.append(
      "Set-Cookie",
      `mgsearch_session=${token}; Path=/; HttpOnly; SameSite=Lax; Max-Age=${24 * 60 * 60}`
    );

    // Redirect to dashboard
    return redirect("/app", { headers });
  } catch (error) {
    return json({
      error: "Installation failed",
      details: error instanceof Error ? error.message : "Unknown error"
    }, { status: 500 });
  }
}

export default function Callback() {
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold">Installing...</h1>
      <p>Please wait while we set up your store.</p>
    </div>
  );
}
```

**Benefits of Frontend-Handled OAuth:**
- ✅ Full control over OAuth flow in Remix
- ✅ No need to proxy OAuth callbacks through backend
- ✅ Simpler redirect handling
- ✅ Can use Shopify's Remix SDK directly

**Option: Use Backend Exchange Endpoint**

If you prefer to let the backend handle token exchange:

```typescript
// Instead of exchanging directly with Shopify:
const exchangeResponse = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/exchange`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ shop, code }),
});

const { access_token } = await exchangeResponse.json();

// Then send to install endpoint
const installResponse = await fetch(`${GO_BACKEND_URL}/api/auth/shopify/install`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ shop, access_token }),
});
```

### Step 3: Create API Client Utility

**File: `app/lib/api.client.ts`**

```typescript
const GO_BACKEND_URL = process.env.GO_BACKEND_URL || "http://localhost:8080";

export interface Store {
  id: string;
  shop_domain: string;
  shop_name: string;
  product_index_uid: string;
  meilisearch_index_uid: string;
  meilisearch_document_type: string;
  plan_level: string;
  status: string;
  sync_state: Record<string, any>;
  installed_at: string;
}

export interface ApiError {
  error: string;
  details?: string;
}

async function getSessionToken(request: Request): Promise<string | null> {
  const cookieHeader = request.headers.get("Cookie");
  if (!cookieHeader) return null;
  
  const cookies = cookieHeader.split(";").reduce((acc, cookie) => {
    const [key, value] = cookie.trim().split("=");
    acc[key] = value;
    return acc;
  }, {} as Record<string, string>);

  return cookies.mgsearch_session || null;
}

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {},
  request?: Request
): Promise<T> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  };

  // Add session token if available
  if (request) {
    const token = await getSessionToken(request);
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }
  }

  const response = await fetch(`${GO_BACKEND_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    const error: ApiError = await response.json();
    throw new Error(error.error || `API error: ${response.statusText}`);
  }

  return response.json();
}

export const apiClient = {
  // Get current store (authenticated)
  async getCurrentStore(request: Request): Promise<Store> {
    return apiRequest<Store>("/api/stores/current", { method: "GET" }, request);
  },

  // Health check
  async health(): Promise<{ message: string }> {
    return apiRequest<{ message: string }>("/ping", { method: "GET" });
  },
};
```

### Step 4: Create Dashboard Route

**File: `app/routes/app._index.tsx`**

```typescript
import { json } from "@remix-run/node";
import { useLoaderData, Link } from "@remix-run/react";
import type { LoaderFunctionArgs } from "@remix-run/node";
import { apiClient, type Store } from "~/lib/api.client";

export async function loader({ request }: LoaderFunctionArgs) {
  try {
    const store = await apiClient.getCurrentStore(request);
    return json({ store });
  } catch (error) {
    // Not authenticated or store not found
    return json({ 
      store: null, 
      error: error instanceof Error ? error.message : "Failed to load store" 
    });
  }
}

export default function Dashboard() {
  const { store, error } = useLoaderData<typeof loader>();

  if (error || !store) {
    return (
      <div className="p-8">
        <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
        <div className="bg-red-50 border border-red-200 rounded p-4">
          <p className="text-red-800">
            {error || "Store not found. Please install the app first."}
          </p>
          <Link 
            to="/auth/install" 
            className="text-blue-600 underline mt-2 inline-block"
          >
            Install App
          </Link>
        </div>
      </div>
    );
  }

  const syncStatus = store.sync_state?.status || "unknown";

  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold mb-6">Dashboard</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div className="bg-white border rounded-lg p-6 shadow">
          <h2 className="text-xl font-semibold mb-4">Store Information</h2>
          <dl className="space-y-2">
            <div>
              <dt className="text-sm text-gray-600">Shop Domain</dt>
              <dd className="font-medium">{store.shop_domain}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600">Shop Name</dt>
              <dd className="font-medium">{store.shop_name || "N/A"}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600">Plan</dt>
              <dd className="font-medium capitalize">{store.plan_level}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600">Status</dt>
              <dd className="font-medium capitalize">{store.status}</dd>
            </div>
          </dl>
        </div>

        <div className="bg-white border rounded-lg p-6 shadow">
          <h2 className="text-xl font-semibold mb-4">Search Index</h2>
          <dl className="space-y-2">
            <div>
              <dt className="text-sm text-gray-600">Index UID</dt>
              <dd className="font-mono text-sm">{store.meilisearch_index_uid}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600">Document Type</dt>
              <dd className="font-medium capitalize">{store.meilisearch_document_type}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-600">Sync Status</dt>
              <dd className="font-medium capitalize">{syncStatus}</dd>
            </div>
          </dl>
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h2 className="text-xl font-semibold mb-2">Next Steps</h2>
        <ul className="list-disc list-inside space-y-2 text-gray-700">
          <li>Wait for initial product sync to complete</li>
          <li>Configure search settings (facets, filters, etc.)</li>
          <li>Install the search UI in your store theme</li>
        </ul>
      </div>
    </div>
  );
}
```

### Step 5: Update Shopify App Configuration

**File: `app/routes/auth.callback.tsx` (update redirect URL)**

Ensure your Shopify app's callback URL in Partners dashboard matches:
```
https://your-ngrok-url.ngrok.io/auth/callback
```

This will redirect to Remix, which then forwards to Go backend.

### Step 6: Add Error Boundary

**File: `app/routes/app._index.tsx` (add error boundary)**

```typescript
import { useRouteError } from "@remix-run/react";

export function ErrorBoundary() {
  const error = useRouteError();
  
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold text-red-600">Error</h1>
      <p>{error instanceof Error ? error.message : "Unknown error"}</p>
    </div>
  );
}
```

### Step 7: Update Root Layout (Optional)

**File: `app/root.tsx`**

Add a navigation bar or ensure session cookies are handled:

```typescript
// In your root component
export function links() {
  return [
    { rel: "stylesheet", href: stylesheet },
  ];
}
```

## Testing the Integration

### 1. Start Go Backend
```bash
cd /path/to/mgsearch
nix develop  # or your dev environment
just dev-up
go run main.go
```

### 2. Start Remix Dev Server
```bash
cd /path/to/remix-app
npm run dev
```

### 3. Test OAuth Flow
1. Visit `http://localhost:3000/auth/install`
2. Enter a test shop domain
3. Complete OAuth flow
4. Should redirect to `/app` dashboard

### 4. Verify Store Creation
- Check Go backend logs for store creation
- Check database: `psql -h localhost -p 5544 -U mgsearch -d mgsearch -c "SELECT shop_domain, status FROM stores;"`

## Production Considerations

### 1. Environment Variables
- Use secure secret management (AWS Secrets Manager, Vault, etc.)
- Never commit `.env` files
- Use different credentials for dev/staging/prod

### 2. CORS Configuration
If Remix and Go backend are on different domains:
- Add CORS middleware to Go backend
- Configure allowed origins

### 3. Session Security
- Use HTTPS in production
- Set `Secure` flag on session cookies
- Consider shorter session TTLs
- Implement session refresh

### 4. Error Handling
- Add retry logic for API calls
- Implement proper error boundaries
- Log errors for monitoring

### 5. Rate Limiting
- Add rate limiting to Go backend endpoints
- Protect against abuse

## Troubleshooting

### Issue: OAuth callback fails
- Check Go backend logs
- Verify `SHOPIFY_APP_URL` matches Partners dashboard
- Ensure callback URL is whitelisted

### Issue: Session not persisting
- Check cookie settings (HttpOnly, SameSite)
- Verify JWT signing key matches
- Check browser console for cookie issues

### Issue: API calls fail
- Verify `GO_BACKEND_URL` is correct
- Check CORS headers
- Verify session token is being sent

## Next Steps

1. **Add Search Configuration UI**: Let merchants configure facets, filters, etc.
2. **Add Sync Status Monitoring**: Real-time updates on product sync progress
3. **Add Analytics Dashboard**: Show search metrics, popular queries, etc.
4. **Theme Integration**: Create theme app extensions for storefront search
5. **Webhook Status**: Show webhook delivery status and retry failed ones

