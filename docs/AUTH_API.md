# Authentication API Documentation

This document describes the authentication system that supports both JWT and API Key authentication.

## Overview

The authentication system provides:
- **User Management**: User registration, login, and profile management
- **Client Management**: Multi-tenant client (organization) support
- **JWT Authentication**: Token-based authentication for users
- **API Key Authentication**: API key-based authentication for programmatic access
- **Many-to-Many Relationships**: Users can belong to multiple clients, and clients can have multiple users

## Data Models

### User
- `id`: ObjectID
- `email`: Unique email address (used for login)
- `password_hash`: Bcrypt hashed password
- `first_name`: User's first name
- `last_name`: User's last name
- `client_ids`: Array of client IDs the user belongs to
- `is_active`: Boolean flag for soft deletion
- `created_at`: Timestamp
- `updated_at`: Timestamp

### Client
- `id`: ObjectID
- `name`: Unique client name
- `description`: Optional description
- `user_ids`: Array of user IDs who have access
- `api_keys`: Array of API key objects
- `is_active`: Boolean flag for soft deletion
- `created_at`: Timestamp
- `updated_at`: Timestamp

### API Key
- `id`: ObjectID
- `key`: SHA-256 hashed API key
- `name`: Human-readable name for the key
- `key_prefix`: First 8 characters for identification
- `permissions`: Array of permission strings (future use)
- `is_active`: Boolean flag
- `last_used_at`: Last usage timestamp
- `created_at`: Timestamp
- `expires_at`: Optional expiration timestamp

## Authentication Methods

### JWT Authentication
Used for user-facing endpoints. Include the JWT token in the Authorization header:
```
Authorization: Bearer <jwt_token>
```

### API Key Authentication
Used for programmatic access to client resources. Include the API key in either:
- Authorization header: `Authorization: Bearer <api_key>`
- X-API-Key header: `X-API-Key: <api_key>`

## API Endpoints

### User Registration
**POST** `/api/v1/auth/register/user`

Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "message": "user registered successfully",
  "user": {
    "id": "...",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "client_ids": [],
    "is_active": true,
    "created_at": "2025-12-07T10:00:00Z",
    "updated_at": "2025-12-07T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### User Login
**POST** `/api/v1/auth/login`

Login with email and password.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "message": "login successful",
  "user": { ... },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Get Current User
**GET** `/api/v1/auth/me`

Get the current authenticated user's information.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "user": {
    "id": "...",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "client_ids": ["..."],
    "is_active": true,
    "created_at": "2025-12-07T10:00:00Z",
    "updated_at": "2025-12-07T10:00:00Z"
  }
}
```

### Update User
**PUT** `/api/v1/auth/user`

Update the current user's profile.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith"
}
```

**Response:**
```json
{
  "message": "user updated successfully",
  "user": { ... }
}
```

### Register Client
**POST** `/api/v1/auth/register/client`

Create a new client (organization/tenant). The authenticated user becomes associated with this client.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "name": "my-company",
  "description": "My Company's Search Service"
}
```

**Response:**
```json
{
  "message": "client registered successfully",
  "client": {
    "id": "...",
    "name": "my-company",
    "description": "My Company's Search Service",
    "user_ids": ["..."],
    "api_keys": [],
    "is_active": true,
    "created_at": "2025-12-07T10:00:00Z",
    "updated_at": "2025-12-07T10:00:00Z"
  }
}
```

### Get User's Clients
**GET** `/api/v1/auth/clients`

Get all clients the authenticated user has access to.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "clients": [
    {
      "id": "...",
      "name": "my-company",
      "description": "My Company's Search Service",
      "user_ids": ["..."],
      "api_keys": [ ... ],
      "is_active": true,
      "created_at": "2025-12-07T10:00:00Z",
      "updated_at": "2025-12-07T10:00:00Z"
    }
  ]
}
```

### Get Client Details
**GET** `/api/v1/auth/clients/:client_id`

Get details of a specific client.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "client": {
    "id": "...",
    "name": "my-company",
    "description": "My Company's Search Service",
    "user_ids": ["..."],
    "api_keys": [
      {
        "id": "...",
        "name": "Production API Key",
        "prefix": "a1b2c3d4",
        "permissions": ["read", "write"],
        "is_active": true,
        "last_used_at": "2025-12-07T10:00:00Z",
        "created_at": "2025-12-07T09:00:00Z",
        "expires_at": null
      }
    ],
    "is_active": true,
    "created_at": "2025-12-07T10:00:00Z",
    "updated_at": "2025-12-07T10:00:00Z"
  }
}
```

### Generate API Key
**POST** `/api/v1/auth/clients/:client_id/api-keys`

Generate a new API key for the client.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "name": "Production API Key",
  "permissions": ["read", "write"],
  "expires_at": "2026-12-07T10:00:00Z"
}
```

**Response:**
```json
{
  "message": "API key generated successfully",
  "api_key": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0",
  "key_id": "...",
  "prefix": "a1b2c3d4",
  "warning": "Save this API key now. You won't be able to see it again."
}
```

**Important:** The raw API key is only shown once during generation. Store it securely.

### Revoke API Key
**DELETE** `/api/v1/auth/clients/:client_id/api-keys/:key_id`

Revoke (deactivate) an API key.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "API key revoked successfully"
}
```

## Using API Keys

Once you have an API key, you can use it to authenticate requests to client-specific endpoints:

### Example: Search with API Key
**POST** `/api/v1/clients/:client_name/:index_name/search`

**Important:** The `:client_name` in the URL **must match** the client that owns the API key. Using an API key with a different client name will return `403 Forbidden`.

**Headers:**
```
Authorization: Bearer <api_key>
```
or
```
X-API-Key: <api_key>
```

**Request Body:**
```json
{
  "q": "search query",
  "limit": 20
}
```

### Example: Index Document with API Key
**POST** `/api/v1/clients/:client_name/:index_name/documents`

**Headers:**
```
Authorization: Bearer <api_key>
```

**Request Body:**
```json
{
  "id": "doc123",
  "title": "Document Title",
  "content": "Document content..."
}
```

### Example: Update Settings with API Key
**PATCH** `/api/v1/clients/:client_name/:index_name/settings`

**Headers:**
```
Authorization: Bearer <api_key>
```

**Request Body:**
```json
{
  "rankingRules": ["words", "typo", "proximity"],
  "searchableAttributes": ["title", "content"]
}
```

### Example: Get Task Status with API Key
**GET** `/api/v1/clients/:client_name/tasks/:task_id`

**Headers:**
```
Authorization: Bearer <api_key>
```

## Workflow Example

### 1. Register a User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register/user \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

Save the returned `token`.

### 2. Create a Client
```bash
curl -X POST http://localhost:8080/api/v1/auth/register/client \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "acme-corp",
    "description": "Acme Corporation Search"
  }'
```

Save the returned `client.id`.

### 3. Generate an API Key
```bash
curl -X POST http://localhost:8080/api/v1/auth/clients/<client_id>/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "Production Key",
    "permissions": ["read", "write"]
  }'
```

Save the returned `api_key`. This is shown only once!

### 4. Use the API Key
```bash
curl -X POST http://localhost:8080/api/v1/clients/acme-corp/products/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <api_key>" \
  -d '{
    "q": "laptop",
    "limit": 10
  }'
```

## Security Notes

1. **Password Security**: Passwords are hashed using bcrypt with a cost factor of 12
2. **JWT Tokens**: Tokens expire after 24 hours by default
3. **API Keys**: 
   - Stored as SHA-256 hashes in the database
   - The raw key is only shown once during generation
   - Can be revoked at any time
   - Support optional expiration dates
   - Track last usage timestamp
4. **Access Control**: 
   - Users can only access clients they are associated with
   - Every client operation verifies user membership
   - API keys are bound to specific clients - client name in URL must match
   - See [Access Control Documentation](ACCESS_CONTROL.md) for details
   - See [API Key Validation](API_KEY_VALIDATION.md) for API key security
5. **Soft Deletes**: Users and clients are soft-deleted (is_active flag) for data integrity

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "error message",
  "code": "ERROR_CODE",
  "details": "additional details (optional)"
}
```

Common HTTP status codes:
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid authentication)
- `403` - Forbidden (insufficient permissions or wrong client name for API key)
- `404` - Not Found
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

## Migration from Legacy Auth

The legacy Shopify-specific authentication endpoints are still available for backward compatibility but are deprecated:
- `/api/auth/shopify/begin`
- `/api/auth/shopify/callback`
- `/api/auth/shopify/exchange`
- `/api/auth/shopify/install`

New applications should use the auth endpoints under `/api/v1` exclusively.

