# Development Tools

## Proxy Endpoints

The application includes proxy endpoints to help developers interact directly with the backend search engines (Qdrant and Meilisearch) without needing to manually configure connections or manage credentials on their local machine. These endpoints use the server's configured credentials.

**Base URL:** `/api/dev/proxy`

### Qdrant Proxy

Proxies requests to the configured Qdrant instance.

**Endpoint:** `ANY /api/dev/proxy/qdrant/*path`

**Example:**

```bash
# Query points in a collection
curl --location 'http://localhost:8080/api/dev/proxy/qdrant/collections/my_collection/points/query' \
--header 'Content-Type: application/json' \
--data '{
  "with_payload": true,
  "query": {
    "recommend": {
      "positive": [12345],
      "negative": []
    }
  }
}'
```

### Meilisearch Proxy

Proxies requests to the configured Meilisearch instance.

**Endpoint:** `ANY /api/dev/proxy/meilisearch/*path`

**Example:**

```bash
# Search in an index
curl --location 'http://localhost:8080/api/dev/proxy/meilisearch/indexes/my_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "search term"
}'

# Get keys
curl --location 'http://localhost:8080/api/dev/proxy/meilisearch/keys'
```

## Setup

These endpoints are automatically available when the server is running. They rely on the following environment variables:

- `QDRANT_URL` and `QDRANT_API_KEY` for Qdrant proxy.
- `MEILISEARCH_URL` and `MEILISEARCH_API_KEY` for Meilisearch proxy.
