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

### Update Settings
```
PATCH /api/v1/clients/:client_name/:index_name/settings
```

### Get Task Details
```
GET /api/v1/clients/:client_name/tasks/:task_id
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

The request body can contain **any valid Meilisearch index settings parameters** with **multi-level nested JSON structures**. The service forwards the entire request body to Meilisearch as-is.

**Example:**
```bash
curl --location --request PATCH 'http://localhost:8080/api/v1/clients/myclient/movies/settings' \
--header 'Content-Type: application/json' \
--data '{
    "rankingRules": [
        "words",
        "typo",
        "proximity",
        "attribute",
        "sort",
        "exactness",
        "release_date:desc",
        "rank:desc"
    ],
    "distinctAttribute": "movie_id",
    "searchableAttributes": [
        "title",
        "overview",
        "genres"
    ],
    "displayedAttributes": [
        "title",
        "overview",
        "genres",
        "release_date"
    ],
    "stopWords": [
        "the",
        "a",
        "an"
    ],
    "sortableAttributes": [
        "title",
        "release_date"
    ],
    "synonyms": {
        "wolverine": ["xmen", "logan"],
        "logan": ["wolverine"]
    },
    "typoTolerance": {
        "minWordSizeForTypos": {
            "oneTypo": 8,
            "twoTypos": 10
        },
        "disableOnAttributes": ["title"]
    },
    "pagination": {
        "maxTotalHits": 5000
    },
    "faceting": {
        "maxValuesPerFacet": 200
    },
    "searchCutoffMs": 150
}'
```

**Supported Settings Parameters:**

The service accepts all valid Meilisearch index settings parameters including:
- `rankingRules` - Array of strings defining the ranking order
- `distinctAttribute` - String specifying the distinct attribute
- `searchableAttributes` - Array of strings defining searchable attributes
- `displayedAttributes` - Array of strings defining displayed attributes
- `stopWords` - Array of strings defining stop words
- `sortableAttributes` - Array of strings defining sortable attributes
- `synonyms` - Object mapping terms to their synonyms
- `typoTolerance` - Object configuring typo tolerance settings
- `pagination` - Object configuring pagination settings
- `faceting` - Object configuring faceting settings
- `searchCutoffMs` - Number specifying search cutoff time in milliseconds
- And any other Meilisearch index settings parameters, including nested JSON structures

For complete settings documentation, see the [Meilisearch Settings API Reference](https://www.meilisearch.com/docs/reference/api/settings).

Retrieves task details from Meilisearch by task UID. This is useful for checking the status of asynchronous operations like settings updates.

**Example:**
```bash
curl --location 'http://localhost:8080/api/v1/clients/myclient/tasks/15' \
--header 'Content-Type: application/json'
```

**Response:**
The response includes task details such as:
- `uid` - Task UID
- `indexUid` - Index UID associated with the task
- `status` - Task status (enqueued, processing, succeeded, failed)
- `type` - Task type (e.g., "settingsUpdate")
- `details` - Task details (varies by task type)
- `error` - Error information if the task failed
- `duration` - Task execution duration
- `enqueuedAt` - When the task was enqueued
- `startedAt` - When the task started processing
- `finishedAt` - When the task finished

**Example Response:**
```json
{
    "uid": 15,
    "batchUid": 13,
    "indexUid": "test_index",
    "status": "succeeded",
    "type": "settingsUpdate",
    "canceledBy": null,
    "details": {
        "displayedAttributes": ["title", "overview", "genres", "release_date"],
        "searchableAttributes": ["title", "overview", "genres"],
        "sortableAttributes": ["release_date", "title"],
        "rankingRules": ["words", "typo", "proximity", "attribute", "sort", "exactness", "release_date:desc", "rank:desc"],
        "stopWords": ["a", "an", "the"],
        "synonyms": {
            "logan": ["wolverine"],
            "wolverine": ["xmen", "logan"]
        },
        "distinctAttribute": "movie_id",
        "typoTolerance": {
            "minWordSizeForTypos": {
                "oneTypo": 8,
                "twoTypos": 10
            },
            "disableOnAttributes": ["title"]
        },
        "faceting": {
            "maxValuesPerFacet": 200
        },
        "pagination": {
            "maxTotalHits": 5000
        },
        "searchCutoffMs": 150
    },
    "error": null,
    "duration": "PT9.752253947S",
    "enqueuedAt": "2025-11-22T16:35:16.14171112Z",
    "startedAt": "2025-11-22T16:35:16.158978866Z",
    "finishedAt": "2025-11-22T16:35:25.911232813Z"
}
```

**Note:** When you update settings, Meilisearch returns a task response with `taskUid`. You can use that `taskUid` with this endpoint to check the status and get the final result of the settings update.

## Configuration

- `MEILISEARCH_URL`: Your Meilisearch instance URL (required)
- `MEILISEARCH_API_KEY`: Your Meilisearch API key (required)
- `PORT`: Server port (default: 8080)

