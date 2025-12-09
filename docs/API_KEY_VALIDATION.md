# API Key Validation and Security

This document explains how API key validation works and the security measures in place.

## Overview

API keys provide programmatic access to client-specific endpoints. Each API key is bound to a specific client and can only be used to access that client's resources.

## API Key Flow

### 1. API Key Generation
```
User (JWT authenticated) → Generates API key for Client A
   ↓
API key is stored with SHA-256 hash
   ↓
Raw key shown once: "a1b2c3d4e5f6..."
   ↓
Key is associated with Client A's ID
```

### 2. API Key Usage
```
Request to: /api/v1/clients/acme-corp/products/search
   ↓
Authorization: Bearer <api_key>
   ↓
Middleware validates API key
   ↓
Finds Client that owns the key
   ↓
Verifies "acme-corp" matches Client.Name
   ↓
✅ Allowed or ❌ 403 Forbidden
```

## Validation Rules

### Rule 1: Valid API Key Hash
```go
// API key is hashed with SHA-256
apiKeyHash := sha256(apiKey)

// Lookup in database
client := FindByAPIKey(apiKeyHash)
```

**If fails:** `401 Unauthorized - "invalid API key"`

### Rule 2: API Key is Active
```go
if !apiKey.IsActive {
    return 401 Unauthorized
}
```

**If fails:** `401 Unauthorized - "invalid API key"`

### Rule 3: API Key Not Expired
```go
if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(now) {
    return 401 Unauthorized - "API key has expired"
}
```

**If fails:** `401 Unauthorized - "API key has expired"`

### Rule 4: Client Name Matches (NEW!)
```go
clientNameInURL := c.Param("client_name")
if clientNameInURL != "" && clientNameInURL != client.Name {
    return 403 Forbidden - "API key does not belong to this client"
}
```

**If fails:** `403 Forbidden - "API key does not belong to this client"`

## Security Scenarios

### ✅ Scenario 1: Correct Usage
```bash
# Client "acme-corp" has API key "abc123..."

curl -X POST http://localhost:8080/api/v1/clients/acme-corp/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "laptop"}'

# Result: 200 OK ✅
# The client name matches, request proceeds
```

### ❌ Scenario 2: Wrong Client Name
```bash
# Client "acme-corp" has API key "abc123..."
# But someone tries to use it for "other-company"

curl -X POST http://localhost:8080/api/v1/clients/other-company/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "laptop"}'

# Result: 403 Forbidden ❌
# {
#   "error": "API key does not belong to this client",
#   "code": "FORBIDDEN"
# }
```

### ❌ Scenario 3: Random Client Name
```bash
# Client "acme-corp" has API key "abc123..."
# Someone uses random/non-existent client name

curl -X POST http://localhost:8080/api/v1/clients/random-xyz/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "laptop"}'

# Result: 403 Forbidden ❌
# The API key belongs to "acme-corp", not "random-xyz"
```

### ❌ Scenario 4: Invalid API Key
```bash
curl -X POST http://localhost:8080/api/v1/clients/acme-corp/products/search \
  -H "Authorization: Bearer invalid-key" \
  -d '{"q": "laptop"}'

# Result: 401 Unauthorized ❌
# {
#   "error": "invalid API key",
#   "code": "UNAUTHORIZED"
# }
```

## Why This Matters

### Without Client Name Validation (Before Fix)
```
Attacker has API key for "their-client"
   ↓
Makes request to: /api/v1/clients/victim-client/products/search
   ↓
API key is valid ✅
   ↓
Request proceeds ❌ (Wrong!)
   ↓
Attacker accesses victim's data 🚨
```

### With Client Name Validation (After Fix)
```
Attacker has API key for "their-client"
   ↓
Makes request to: /api/v1/clients/victim-client/products/search
   ↓
API key is valid ✅
   ↓
Client name check: "victim-client" != "their-client" ❌
   ↓
403 Forbidden - Request blocked ✅
   ↓
Attacker cannot access victim's data 🔒
```

## Implementation Details

### Middleware Code
```go
// After validating API key and finding client
clientNameParam := c.Param("client_name")
if clientNameParam != "" && clientNameParam != client.Name {
    c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
        "error": "API key does not belong to this client",
        "code":  "FORBIDDEN",
    })
    return
}
```

### Protected Endpoints
All these endpoints now verify client name:
- `POST /api/v1/clients/:client_name/:index_name/search`
- `POST /api/v1/clients/:client_name/:index_name/documents`
- `PATCH /api/v1/clients/:client_name/:index_name/settings`
- `GET /api/v1/clients/:client_name/tasks/:task_id`

## Response Codes

| Scenario | HTTP Status | Message |
|----------|-------------|---------|
| Missing API key | 401 | "missing API key" |
| Invalid API key | 401 | "invalid API key" |
| Expired API key | 401 | "API key has expired" |
| Wrong client name | 403 | "API key does not belong to this client" |
| Valid request | 200/201 | Success |

## Testing

### Test Case 1: Valid Access
```bash
# 1. Register user and client
curl -X POST http://localhost:8080/api/v1/auth/register/user ...
curl -X POST http://localhost:8080/api/v1/auth/register/client \
  -d '{"name": "test-client"}'

# 2. Generate API key
curl -X POST http://localhost:8080/api/v1/auth/clients/<id>/api-keys \
  -d '{"name": "Test Key"}'
# Save API key: "abc123..."

# 3. Use API key with CORRECT client name
curl -X POST http://localhost:8080/api/v1/clients/test-client/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "test"}'

# Expected: 200 OK (or 404 if index doesn't exist yet)
```

### Test Case 2: Invalid Client Name
```bash
# Use the same API key but WRONG client name
curl -X POST http://localhost:8080/api/v1/clients/wrong-client/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "test"}'

# Expected: 403 Forbidden
# {
#   "error": "API key does not belong to this client",
#   "code": "FORBIDDEN"
# }
```

### Test Case 3: Random Client Name
```bash
# Use API key with non-existent client name
curl -X POST http://localhost:8080/api/v1/clients/random123/products/search \
  -H "Authorization: Bearer abc123..." \
  -d '{"q": "test"}'

# Expected: 403 Forbidden
```

## Best Practices

### For API Key Users
1. ✅ **Always use the correct client name** in the URL
2. ✅ **Store API keys securely** (environment variables, secrets manager)
3. ✅ **Rotate keys regularly** for production use
4. ✅ **Set expiration dates** for temporary keys
5. ✅ **Monitor key usage** via last_used_at timestamp

### For API Developers
1. ✅ **Never log raw API keys** (only log prefixes)
2. ✅ **Always hash keys** before storage
3. ✅ **Validate client name** on every request
4. ✅ **Return generic errors** to prevent information leakage
5. ✅ **Track failed attempts** for security monitoring

## Troubleshooting

### Error: "API key does not belong to this client"

**Cause:** The client name in the URL doesn't match the client that owns the API key.

**Solution:**
1. Check which client owns the API key:
   ```bash
   curl -X GET http://localhost:8080/api/v1/auth/clients/<client_id> \
     -H "Authorization: Bearer <jwt_token>"
   ```
2. Use the correct client name from the response
3. Update your API calls to use the correct client name

**Example:**
```bash
# Wrong (if key belongs to "acme-corp")
/api/v1/clients/wrong-name/products/search

# Correct
/api/v1/clients/acme-corp/products/search
```

### How to Find Your Client Name

**Method 1: List your clients**
```bash
curl -X GET http://localhost:8080/api/v1/auth/clients \
  -H "Authorization: Bearer <jwt_token>"
```

**Method 2: Get client details**
```bash
curl -X GET http://localhost:8080/api/v1/auth/clients/<client_id> \
  -H "Authorization: Bearer <jwt_token>"
```

The response will show:
```json
{
  "client": {
    "id": "...",
    "name": "acme-corp",  ← Use this in the URL
    "api_keys": [...]
  }
}
```

## Security Audit Checklist

- [x] API keys are hashed with SHA-256
- [x] Raw keys are only shown once during generation
- [x] API keys can be revoked
- [x] API keys support expiration dates
- [x] API key expiration is checked on every request
- [x] Client name in URL is validated against API key owner
- [x] Different error messages for auth vs. authz (401 vs. 403)
- [x] Last usage timestamp is tracked
- [x] Failed attempts can be logged (implement as needed)

## Summary

✅ **API Key Binding:** Each API key is bound to exactly one client
✅ **URL Validation:** Client name in URL must match API key's client
✅ **Multiple Checks:** Valid key + active + not expired + correct client
✅ **Proper Errors:** 401 for invalid keys, 403 for wrong client
✅ **Security:** Prevents cross-client access even with valid keys

This ensures that API keys can only be used to access their own client's resources, preventing unauthorized cross-client data access.

