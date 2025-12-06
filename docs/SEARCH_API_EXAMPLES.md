# Search API - Correct cURL Examples

## Endpoint

**Base URL**: `https://your-ngrok-url.ngrok-free.dev` (or your server URL)

**Endpoint**: `/api/v1/search`

**Required Header**: `X-Storefront-Key: your-storefront-api-key`

**Note**: This endpoint does NOT require a JWT token - only the storefront key is needed.

---

## 1. Simple Search (POST with JSON)

```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10
  }'
```

---

## 2. Search with Filter (POST with JSON)

### Simple Filter (String)
```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10,
    "filter": "price > 100"
  }'
```

### Complex Filter (Array)
```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10,
    "filter": ["price > 100", "category = \"electronics\""]
  }'
```

### Nested Filter (Object)
```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10,
    "filter": {
      "AND": [
        {"price": {"gt": 100}},
        {"category": "electronics"}
      ]
    }
  }'
```

---

## 3. Search with Sort (POST with JSON)

```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10,
    "sort": ["price:asc", "created_at:desc"]
  }'
```

---

## 4. Search with Filter and Sort (POST with JSON)

```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 10,
    "offset": 0,
    "filter": ["price > 100", "in_stock = true"],
    "sort": ["price:asc"]
  }'
```

---

## 5. Advanced Search with All Parameters (POST with JSON)

```bash
curl -X POST 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search' \
  -H 'Content-Type: application/json' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true' \
  --data-raw '{
    "q": "shows",
    "limit": 20,
    "offset": 0,
    "filter": ["price > 50 AND price < 500", "category = \"electronics\""],
    "sort": ["price:asc", "rating:desc"],
    "facets": ["category", "brand"],
    "attributesToRetrieve": ["title", "price", "image", "url"],
    "attributesToHighlight": ["title", "description"]
  }'
```

---

## 6. GET Request (Query Parameters)

You can also use GET with query parameters:

```bash
curl -X GET 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search?q=shows&limit=10' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true'
```

### GET with Filter (URL encoded)
```bash
curl -X GET 'https://brooklyn-cupolated-ambroise.ngrok-free.dev/api/v1/search?q=shows&limit=10&filters=["price%20%3E%20100"]' \
  -H 'X-Storefront-Key: your-storefront-api-key' \
  -H 'ngrok-skip-browser-warning: true'
```

---

## Common Issues

### ❌ Wrong: Missing `/api/v1` prefix
```bash
curl 'https://your-url.ngrok-free.dev/search'  # 404 Not Found
```

### ✅ Correct: Full path
```bash
curl 'https://your-url.ngrok-free.dev/api/v1/search'
```

### ❌ Wrong: Missing `X-Storefront-Key` header
```bash
curl -X POST 'https://your-url.ngrok-free.dev/api/v1/search' \
  --data-raw '{"q": "shows"}'  # 401 Unauthorized
```

### ✅ Correct: Include header
```bash
curl -X POST 'https://your-url.ngrok-free.dev/api/v1/search' \
  -H 'X-Storefront-Key: your-key' \
  --data-raw '{"q": "shows"}'
```

---

## Filter Examples

### Price Range
```json
{
  "q": "laptop",
  "filter": "price >= 500 AND price <= 2000"
}
```

### Multiple Categories
```json
{
  "q": "shoes",
  "filter": ["category = \"men\"", "category = \"women\""]
}
```

### In Stock Only
```json
{
  "q": "phone",
  "filter": "in_stock = true AND inventory_quantity > 0"
}
```

### Complex AND/OR Logic
```json
{
  "q": "electronics",
  "filter": {
    "AND": [
      {"OR": [{"category": "phones"}, {"category": "tablets"}]},
      {"price": {"gte": 100, "lte": 1000}}
    ]
  }
}
```

---

## Response Format

All requests return Meilisearch search results:

```json
{
  "hits": [
    {
      "id": "product-123",
      "title": "Product Name",
      "price": 99.99,
      ...
    }
  ],
  "query": "shows",
  "processingTimeMs": 5,
  "limit": 10,
  "offset": 0,
  "estimatedTotalHits": 42
}
```

