# How to Get Your Session Token (JWT)

The `/api/stores/current` endpoint requires a **JWT session token** (not a simple API key). This token is generated when you install a store.

## What is a Session Token?

- **JWT (JSON Web Token)** - A signed token that contains store information
- **Valid for 24 hours** - Tokens expire after 24 hours
- **Generated automatically** - Created when you install a store via the auth endpoints

---

## Method 1: Get Token from Store Installation (Recommended)

When you install a store, the response includes a session token:

### Install Store Endpoint

```bash
POST /api/auth/shopify/install
```

**Request:**
```json
{
  "shop": "mg-store-207095.myshopify.com",
  "access_token": "shpat_abc123...",
  "shop_name": "My Store"
}
```

**Response:**
```json
{
  "store": {
    "id": "store-uuid",
    "shop_domain": "mg-store-207095.myshopify.com",
    "api_key_public": "abc123...",
    ...
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",  // <-- This is your session token
  "message": "installation successful"
}
```

**Save this token** - you'll need it for authenticated requests.

---

## Method 2: Generate Token Programmatically

If you need to generate a token manually (for testing), you can create a simple Go script:

### Create Token Generator Script

```go
// scripts/generate-token.go
package main

import (
	"fmt"
	"os"
	"time"
	
	"mgsearch/pkg/auth"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run generate-token.go <store-id> <shop-domain>")
		os.Exit(1)
	}
	
	storeID := os.Args[1]
	shop := os.Args[2]
	signingKey := os.Getenv("JWT_SIGNING_KEY")
	
	if signingKey == "" {
		fmt.Println("Error: JWT_SIGNING_KEY environment variable not set")
		os.Exit(1)
	}
	
	token, err := auth.GenerateSessionToken(storeID, shop, []byte(signingKey), 24*time.Hour)
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println(token)
}
```

**Usage:**
```bash
export JWT_SIGNING_KEY=$(grep JWT_SIGNING_KEY .env | cut -d '=' -f2)
go run scripts/generate-token.go <store-id> <shop-domain>
```

---

## Method 3: Query Database for Store ID

First, get your store ID from the database:

```sql
SELECT id, shop_domain FROM stores WHERE shop_domain = 'mg-store-207095.myshopify.com';
```

Then use the store ID to generate a token (see Method 2).

---

## Using the Token

Once you have the token, use it in the `Authorization` header:

```bash
curl -X GET 'http://localhost:8080/api/stores/current' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
```

### Example Response

```json
{
  "id": "store-uuid",
  "shop_domain": "mg-store-207095.myshopify.com",
  "shop_name": "My Store",
  "api_key_public": "abc123...",
  "plan_level": "free",
  "status": "active",
  ...
}
```

---

## Token Expiration

- **Default TTL**: 24 hours
- **After expiration**: You'll get `401 Unauthorized` with "invalid token"
- **To get a new token**: Re-install the store or generate a new one

---

## Quick Test

### 1. Install Store (Get Token)

```bash
curl -X POST 'http://localhost:8080/api/auth/shopify/install' \
  -H 'Content-Type: application/json' \
  -d '{
    "shop": "mg-store-207095.myshopify.com",
    "access_token": "your-shopify-access-token"
  }'
```

**Save the `token` from the response.**

### 2. Use Token to Get Store Info

```bash
curl -X GET 'http://localhost:8080/api/stores/current' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_TOKEN_HERE'
```

---

## Token Structure

The JWT token contains:
- `store_id` - Your store's unique ID
- `shop` - Shop domain (e.g., "mg-store-207095.myshopify.com")
- `exp` - Expiration timestamp
- `iat` - Issued at timestamp

You can decode the token at [jwt.io](https://jwt.io) to see its contents (without the signature).

---

## Troubleshooting

### Error: "missing authorization header"

**Solution**: Make sure you include the `Authorization: Bearer <token>` header.

### Error: "invalid token"

**Possible causes:**
1. Token has expired (24 hours)
2. Token was signed with a different `JWT_SIGNING_KEY`
3. Token is malformed

**Solution**: Generate a new token or re-install the store.

### Error: "unauthorized"

**Solution**: The token doesn't contain valid store information. Check that:
- Store exists in database
- Token was generated with correct store ID
- `JWT_SIGNING_KEY` matches between token generation and validation

---

## Important Notes

1. **Keep tokens secure** - Don't commit them to version control
2. **Tokens expire** - Plan to refresh them before 24 hours
3. **One token per store** - Each store installation generates a new token
4. **JWT_SIGNING_KEY** - Must match between token generation and validation

