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

The service accepts all valid Meilisearch search parameters as documented in the [official Meilisearch API reference](https://www.meilisearch.com/docs/reference/api/search#search-parameters). All parameters are passed through to Meilisearch as-is.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | String | `""` | Query string |
| `offset` | Integer | `0` | Number of documents to skip |
| `limit` | Integer | `20` | Maximum number of documents returned |
| `hitsPerPage` | Integer | `1` | Maximum number of documents returned for a page |
| `page` | Integer | `1` | Request a specific page of results |
| `filter` | String or Array of strings | `null` | Filter queries by an attribute's value |
| `facets` | Array of strings | `null` | Display the count of matches per facet |
| `distinct` | String | `null` | Restrict search to documents with unique values of specified attribute |
| `attributesToRetrieve` | Array of strings | `["*"]` | Attributes to display in the returned documents |
| `attributesToCrop` | Array of strings | `null` | Attributes whose values have to be cropped |
| `cropLength` | Integer | `10` | Maximum length of cropped value in words |
| `cropMarker` | String | `"…"` | String marking crop boundaries |
| `attributesToHighlight` | Array of strings | `null` | Highlight matching terms contained in an attribute |
| `highlightPreTag` | String | `"<em>"` | String inserted at the start of a highlighted term |
| `highlightPostTag` | String | `"</em>"` | String inserted at the end of a highlighted term |
| `showMatchesPosition` | Boolean | `false` | Return matching terms location |
| `sort` | Array of strings | `null` | Sort search results by an attribute's value |
| `matchingStrategy` | String | `last` | Strategy used to match query terms within documents |
| `showRankingScore` | Boolean | `false` | Display the global ranking score of a document |
| `showRankingScoreDetails` | Boolean | `false` | Adds a detailed global ranking score field |
| `rankingScoreThreshold` | Number | `null` | Excludes results with low ranking scores |
| `attributesToSearchOn` | Array of strings | `["*"]` | Restrict search to the specified attributes |
| `hybrid` | Object | `null` | Return results based on query keywords and meaning (requires `embedder` and `semanticRatio`) |
| `vector` | Array of numbers | `null` | Search using a custom query vector |
| `retrieveVectors` | Boolean | `false` | Return document and query vector data |
| `locales` | Array of strings | `null` | Explicitly specify languages used in a query |
| `media` | Object | `null` | Perform AI-powered search queries with multimodal content (experimental) |

**Examples with advanced parameters:**

Hybrid search (semantic + keyword):
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "kitchen utensils",
    "hybrid": {
        "semanticRatio": 0.9,
        "embedder": "EMBEDDER_NAME"
    }
}'
```

Vector search:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "vector": [0.1, 0.2, 0.3],
    "hybrid": {
        "embedder": "EMBEDDER_NAME"
    }
}'
```

Search with locales:
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/test_index/search' \
--header 'Content-Type: application/json' \
--data '{
    "q": "進撃の巨人",
    "locales": ["jpn"]
}'
```

For complete parameter documentation, see the [Meilisearch Search API Reference](https://www.meilisearch.com/docs/reference/api/search#search-parameters).

## Configuration

- `MEILISEARCH_URL`: Your Meilisearch instance URL (required)
- `MEILISEARCH_API_KEY`: Your Meilisearch API key (required)
- `PORT`: Server port (default: 8080)

