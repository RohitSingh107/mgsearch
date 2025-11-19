# MGSearch - Meilisearch Proxy Service

A Go-based backend service that acts as a proxy between clients and Meilisearch, forwarding search requests and returning results.

## Project Structure

```
mgsearch/
├── main.go              # Application entry point
├── config/              # Configuration management
│   └── config.go
├── handlers/            # HTTP request handlers
│   └── search.go
├── services/            # Business logic and external service clients
│   └── meilisearch.go
├── models/              # Data models
│   └── search.go
├── go.mod
├── go.sum
└── README.md
```

## Setup

1. **Install dependencies:**
   ```bash
   go mod tidy
   ```

2. **Set environment variables:**
   
   The application automatically loads environment variables from a `.env` file if it exists. Create a `.env` file in the project root:
   ```bash
   # .env file
   MEILISEARCH_URL=https://ms-4a594f30ff0a-34895.sgp.meilisearch.io
   MEILISEARCH_API_KEY=my_secret_master_key
   PORT=8080  # Optional, defaults to 8080
   ```
   
   Alternatively, you can set environment variables directly:
   ```bash
   export MEILISEARCH_URL=https://ms-4a594f30ff0a-34895.sgp.meilisearch.io
   export MEILISEARCH_API_KEY=my_secret_master_key
   export PORT=8080
   ```
   
   **Note:** Environment variables take precedence over `.env` file values.

3. **Run the server:**
   ```bash
   go run main.go
   ```

## API Endpoints

### Health Check
```
GET /ping
```

### Search
```
POST /api/v1/clients/:client_name/:index_name/search
```

The request body can contain **any valid Meilisearch search parameters** with **multi-level nested JSON structures**. The service forwards the entire request body to Meilisearch as-is.

**Examples:**

Simple search:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame"
}'
```

Search with filters and sorting:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame",
    "filter": "genre = action",
    "sort": ["release_date:desc"],
    "limit": 10,
    "offset": 0
}'
```

Search with complex nested filters:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "endgame",
    "filter": ["genre = action", "year > 2020"],
    "facets": ["genre", "year"],
    "attributesToRetrieve": ["title", "description"],
    "attributesToHighlight": ["description"]
}'
```

**Supported Parameters:**
The service accepts any valid Meilisearch search parameters including:
- `q` - Search query
- `filter` - Filter expressions (string or array)
- `sort` - Sort criteria (array)
- `limit` - Number of results
- `offset` - Pagination offset
- `facets` - Facet fields (array)
- `attributesToRetrieve` - Fields to return (array)
- `attributesToCrop` - Fields to crop (array)
- `cropLength` - Crop length
- `attributesToHighlight` - Fields to highlight (array)
- And any other Meilisearch search parameters, including nested JSON structures

## Configuration

- `MEILISEARCH_URL`: Your Meilisearch instance URL (required)
- `MEILISEARCH_API_KEY`: Your Meilisearch API key (required)
- `PORT`: Server port (default: 8080)

