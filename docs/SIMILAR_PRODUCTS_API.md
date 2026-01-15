# Similar Products API

The Similar Products API provides product recommendations using Qdrant's vector similarity search. This endpoint allows storefronts to display "Similar Products" or "You may also like" sections based on vector embeddings.

## Overview

- **Endpoint**: `/api/v1/similar`
- **Methods**: `GET`, `POST`
- **Authentication**: X-Storefront-Key header (same as Search API)
- **CORS**: Fully supported for cross-origin requests

## Configuration

Add the following environment variables to your `.env` file:

```bash
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=your-api-key-here
```

For Qdrant Cloud, use your cluster URL:

```bash
QDRANT_URL=https://your-cluster.qdrant.tech
QDRANT_API_KEY=your-cloud-api-key
```

## Store Setup

Each store needs a Qdrant collection configured. The collection name is determined by:

1. `qdrant_collection_name` field (if explicitly set)
2. Falls back to `product_index_uid`
3. Falls back to `shop_domain`

To set the collection name when creating/updating a store:

```json
{
  "shop_domain": "mystore.myshopify.com",
  "qdrant_collection_name": "mystore_products"
}
```

## API Reference

### GET Request

Fetch similar products using query parameters.

**Example:**

```bash
curl -H "X-Storefront-Key: <your-key>" \
  "http://localhost:8080/api/v1/similar?id=24&limit=5"
```

**Parameters:**

- `id` (required): Product ID (string or integer)
- `limit` (optional): Number of recommendations to return (default: 10)

### POST Request

Fetch similar products using JSON body.

**Example:**

```bash
curl -X POST \
  -H "X-Storefront-Key: <your-key>" \
  -H "Content-Type: application/json" \
  -d '{"id": 24, "limit": 5}' \
  "http://localhost:8080/api/v1/similar"
```

**Request Body:**

```json
{
  "id": 24,
  "limit": 10
}
```

**Fields:**

- `id` (required): Product ID (can be string or number)
- `limit` (optional): Maximum number of results (default: 10)

## Response Format

The API returns the Qdrant response directly:

```json
{
  "result": [
    {
      "id": 25,
      "score": 0.95,
      "payload": {
        "title": "Similar Product",
        "price": 29.99,
        "image_url": "https://...",
        ...
      }
    },
    {
      "id": 26,
      "score": 0.92,
      "payload": {
        "title": "Another Similar Product",
        "price": 34.99,
        ...
      }
    }
  ],
  "status": "ok",
  "time": 0.003
}
```

**Response Fields:**

- `result`: Array of similar products
  - `id`: Product ID in Qdrant
  - `score`: Similarity score (0-1, higher is more similar)
  - `payload`: Product metadata stored in Qdrant
- `status`: Request status ("ok" or error)
- `time`: Query execution time in seconds

## Error Responses

### Missing Storefront Key

```json
{
  "error": "missing storefront key"
}
```

**Status Code**: 401

### Invalid Storefront Key

```json
{
  "error": "invalid storefront key"
}
```

**Status Code**: 401

### Missing Product ID

```json
{
  "error": "missing product id"
}
```

**Status Code**: 400

### Collection Not Configured

```json
{
  "error": "qdrant collection not configured for store"
}
```

**Status Code**: 500

### Qdrant Service Error

```json
{
  "error": "failed to fetch similar products",
  "details": "qdrant error (status 404): collection not found"
}
```

**Status Code**: 500

## Integration Examples

### JavaScript/TypeScript

```typescript
const getSimilarProducts = async (productId: number, limit: number = 10) => {
  const response = await fetch(
    `https://api.yourdomain.com/api/v1/similar?id=${productId}&limit=${limit}`,
    {
      headers: {
        'X-Storefront-Key': 'your-storefront-key'
      }
    }
  );
  
  if (!response.ok) {
    throw new Error('Failed to fetch similar products');
  }
  
  return response.json();
};
```

### React Component

```tsx
import { useEffect, useState } from 'react';

interface SimilarProduct {
  id: number;
  score: number;
  payload: {
    title: string;
    price: number;
    image_url: string;
  };
}

export function SimilarProducts({ productId }: { productId: number }) {
  const [products, setProducts] = useState<SimilarProduct[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchSimilar = async () => {
      try {
        const response = await fetch(
          `/api/v1/similar?id=${productId}&limit=6`,
          {
            headers: {
              'X-Storefront-Key': process.env.NEXT_PUBLIC_STOREFRONT_KEY!
            }
          }
        );
        const data = await response.json();
        setProducts(data.result);
      } catch (error) {
        console.error('Failed to fetch similar products:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchSimilar();
  }, [productId]);

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <h2>You May Also Like</h2>
      <div className="product-grid">
        {products.map((product) => (
          <div key={product.id} className="product-card">
            <img src={product.payload.image_url} alt={product.payload.title} />
            <h3>{product.payload.title}</h3>
            <p>${product.payload.price}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
```

### Remix Loader

```typescript
// app/routes/products.$id.tsx
import { json, type LoaderFunctionArgs } from "@remix-run/node";

export async function loader({ params }: LoaderFunctionArgs) {
  const productId = params.id;
  
  // Fetch similar products
  const similarResponse = await fetch(
    `${process.env.API_URL}/api/v1/similar?id=${productId}&limit=8`,
    {
      headers: {
        'X-Storefront-Key': process.env.STOREFRONT_KEY!
      }
    }
  );
  
  const similarProducts = await similarResponse.json();
  
  return json({ similarProducts });
}
```

## Architecture

### Components

1. **Config** (`config/config.go`): Loads Qdrant URL and API key from environment
2. **Service** (`services/qdrant.go`): Handles HTTP communication with Qdrant API
3. **Handler** (`handlers/storefront.go`): Processes requests and validates authentication
4. **Model** (`models/store.go`): Store configuration including Qdrant collection name

### Flow

1. Client sends request with `X-Storefront-Key` header
2. Handler validates the storefront key and retrieves store configuration
3. Handler extracts product ID and limit from request
4. Handler gets Qdrant collection name from store
5. QdrantService calls Qdrant's recommendation API
6. Results are returned to client

## Qdrant Collection Setup

Before using the Similar Products API, you need to:

1. **Create a Collection**: Create a Qdrant collection with vector embeddings
2. **Index Products**: Upload product vectors to the collection
3. **Configure Store**: Set the collection name in your store configuration

### Example: Creating a Collection

```bash
curl -X PUT "http://localhost:6333/collections/mystore_products" \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": {
      "size": 384,
      "distance": "Cosine"
    }
  }'
```

### Example: Uploading Vectors

```python
from qdrant_client import QdrantClient
from qdrant_client.models import PointStruct

client = QdrantClient(url="http://localhost:6333")

points = [
    PointStruct(
        id=product_id,
        vector=product_embedding,
        payload={
            "title": "Product Title",
            "price": 29.99,
            "image_url": "https://..."
        }
    )
    for product_id, product_embedding in products
]

client.upsert(
    collection_name="mystore_products",
    points=points
)
```

## Best Practices

1. **Limit Results**: Use reasonable limits (5-10) for better performance
2. **Cache Results**: Consider caching recommendations for popular products
3. **Error Handling**: Always handle API errors gracefully
4. **Collection Management**: Keep Qdrant collections in sync with product catalog
5. **Vector Quality**: Use high-quality embeddings for better recommendations

## Troubleshooting

### No Results Returned

- Verify the product ID exists in Qdrant
- Check if the collection has sufficient vectors
- Ensure vector embeddings are properly uploaded

### Collection Not Found

- Verify the collection name in store configuration
- Create the collection in Qdrant if it doesn't exist
- Check Qdrant connection URL and API key

### Slow Response Times

- Add indexes to your Qdrant collection
- Reduce the limit parameter
- Consider using Qdrant's HNSW parameters for faster search
- Upgrade Qdrant instance resources

## Security Considerations

- **API Keys**: Never expose Qdrant API keys in client-side code
- **Storefront Keys**: Storefront keys can be safely used in frontend applications
- **Rate Limiting**: Consider implementing rate limiting for this endpoint
- **CORS**: Configure CORS appropriately for your domains

## See Also

- [Search API Examples](./SEARCH_API_EXAMPLES.md)
- [Quick Start Guide](./QUICK_START.md)
- [Qdrant Documentation](https://qdrant.tech/documentation/)
