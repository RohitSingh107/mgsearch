# Session Storage API - Implementation

This document describes the implemented Session Storage API endpoints for the Go service.

## Overview

The Session Storage API provides REST endpoints for storing and managing Shopify OAuth session data. All endpoints are located under `/api/sessions`.

**Base URL**: Configurable via `PORT` environment variable (default: `http://localhost:8080`)

**Authentication**: Optional Bearer token via `Authorization` header (configured via `SESSION_API_KEY`)

---

## Endpoints

### 1. Store Session

**POST** `/api/sessions`

Creates or updates a session (upsert behavior).

**Request Body:**
```json
{
  "id": "mgstore-9986.myshopify.com_1234567890",
  "shop": "mgstore-9986.myshopify.com",
  "state": "random-state-string-abc123",
  "isOnline": false,
  "scope": "read_products,write_products",
  "expires": "2025-12-26T19:30:00Z",
  "accessToken": "shpat_abc123...",
  "userId": null,
  "createdAt": "2025-01-26T19:00:00Z",
  "updatedAt": "2025-01-26T19:05:00Z"
}
```

**Response:**
- `200 OK` - Success with message
- `400 Bad Request` - Invalid request body or missing required fields
- `401 Unauthorized` - Missing or invalid API key (if `SESSION_API_KEY` is set)
- `500 Internal Server Error` - Server error

**Notes:**
- `accessToken` is **automatically encrypted** before storage using AES-256-GCM
- If session with same `id` exists, it will be updated
- `updatedAt` is automatically set to current time

---

### 2. Load Session

**GET** `/api/sessions/{sessionId}`

Retrieves a session by ID.

**Response:**
- `200 OK` - Session object with decrypted `accessToken`
- `404 Not Found` - Session doesn't exist
- `401 Unauthorized` - Missing or invalid API key (if `SESSION_API_KEY` is set)
- `500 Internal Server Error` - Server error

**Notes:**
- `accessToken` is **automatically decrypted** before returning
- URL encode the `sessionId` parameter

---

### 3. Delete Session

**DELETE** `/api/sessions/{sessionId}`

Deletes a session by ID (idempotent).

**Response:**
- `204 No Content` - Success
- `401 Unauthorized` - Missing or invalid API key (if `SESSION_API_KEY` is set)
- `500 Internal Server Error` - Server error

**Notes:**
- Idempotent: Deleting a non-existent session returns success

---

### 4. Delete Multiple Sessions

**DELETE** `/api/sessions/batch`

Deletes multiple sessions at once.

**Request Body:**
```json
{
  "ids": [
    "mgstore-9986.myshopify.com_1234567890",
    "mgstore-9986.myshopify.com_9876543210"
  ]
}
```

**Response:**
- `200 OK` - Success with count
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Missing or invalid API key (if `SESSION_API_KEY` is set)
- `500 Internal Server Error` - Server error

**Notes:**
- Idempotent: Non-existent IDs are ignored

---

### 5. Find Sessions by Shop

**GET** `/api/sessions/shop/{shop}`

Retrieves all sessions for a specific shop.

**Response:**
- `200 OK` - Array of session objects (empty array if none found)
- `401 Unauthorized` - Missing or invalid API key (if `SESSION_API_KEY` is set)
- `500 Internal Server Error` - Server error

**Notes:**
- Always returns an array, even if empty (`[]`)
- All `accessToken` fields are decrypted before returning
- URL encode the `shop` parameter

---

## Authentication

Authentication is **optional** and controlled by the `SESSION_API_KEY` environment variable:

- **If `SESSION_API_KEY` is empty**: All requests are allowed (no authentication required)
- **If `SESSION_API_KEY` is set**: All requests must include `Authorization: Bearer {SESSION_API_KEY}` header

**Example:**
```bash
curl -X GET http://localhost:8080/api/sessions/test-id \
  -H "Authorization: Bearer your-api-key-here"
```

---

## Security

### Access Token Encryption

All `accessToken` values are automatically encrypted/decrypted:

- **Storage**: Encrypted using AES-256-GCM before saving to database
- **Retrieval**: Decrypted using AES-256-GCM before returning in API responses
- **Key**: Configured via `ENCRYPTION_KEY` environment variable (32-byte hex string)

### Encryption Key Generation

Generate a secure encryption key:
```bash
openssl rand -hex 32
```

Set it in your `.env` file:
```bash
ENCRYPTION_KEY=your-generated-hex-key-here
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message describing what went wrong",
  "code": "ERROR_CODE"
}
```

### Error Codes

- `VALIDATION_ERROR` - Invalid request body or missing required fields
- `UNAUTHORIZED` - Missing or invalid authentication token
- `INTERNAL_ERROR` - Unexpected server error

---

## Example cURL Commands

### Store Session
```bash
curl -X POST http://localhost:8080/api/sessions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "id": "test-shop.myshopify.com_123",
    "shop": "test-shop.myshopify.com",
    "state": "test-state",
    "isOnline": false,
    "accessToken": "shpat_test_token"
  }'
```

### Load Session
```bash
curl -X GET http://localhost:8080/api/sessions/test-shop.myshopify.com_123 \
  -H "Authorization: Bearer your-api-key"
```

### Find Sessions by Shop
```bash
curl -X GET http://localhost:8080/api/sessions/shop/test-shop.myshopify.com \
  -H "Authorization: Bearer your-api-key"
```

### Delete Session
```bash
curl -X DELETE http://localhost:8080/api/sessions/test-shop.myshopify.com_123 \
  -H "Authorization: Bearer your-api-key"
```

### Delete Multiple Sessions
```bash
curl -X DELETE http://localhost:8080/api/sessions/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "ids": ["test-shop.myshopify.com_123", "test-shop.myshopify.com_456"]
  }'
```

---

## Environment Variables

```bash
# Required
ENCRYPTION_KEY=32-byte-hex-string  # For encrypting access tokens

# Optional
SESSION_API_KEY=your-api-key-here  # If set, requires Bearer token auth
PORT=8080                           # Server port (default: 8080)
```

---

## Database Schema

The sessions table structure:

```sql
CREATE TABLE sessions (
  id VARCHAR(255) PRIMARY KEY,
  shop VARCHAR(255) NOT NULL,
  state VARCHAR(255) NOT NULL,
  is_online BOOLEAN NOT NULL DEFAULT false,
  scope TEXT,
  expires TIMESTAMP,
  access_token TEXT NOT NULL,  -- Encrypted
  user_id BIGINT,
  first_name VARCHAR(255),
  last_name VARCHAR(255),
  email VARCHAR(255),
  account_owner BOOLEAN DEFAULT false,
  locale VARCHAR(10),
  collaborator BOOLEAN,
  email_verified BOOLEAN,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_shop (shop),
  INDEX idx_expires (expires)
);
```

---

## Implementation Notes

1. **Encryption**: Access tokens are encrypted using AES-256-GCM and stored as hex-encoded strings
2. **Upsert**: `StoreSession` creates or updates based on `id`
3. **Idempotency**: Delete operations are idempotent (safe to retry)
4. **Optional Auth**: Authentication is only enforced if `SESSION_API_KEY` is configured
5. **Backward Compatibility**: If decryption fails, the handler assumes plaintext (for migration scenarios)

---

## Testing

See the main API requirements document for comprehensive test cases. All endpoints have been implemented according to the specification.

