# Quick Start Guide - Authentication

This guide will help you get started with the authentication system in under 5 minutes.

## Prerequisites

- MongoDB running and accessible
- Environment variables configured (see below)
- Go 1.23+ installed

## Environment Setup

Create a `.env` file or set these environment variables:

```bash
# Required
DATABASE_URL=mongodb://localhost:27017/mgsearch
JWT_SIGNING_KEY=your-secret-signing-key-change-this-in-production
MEILISEARCH_URL=http://localhost:7700
MEILISEARCH_API_KEY=your-meilisearch-api-key

# Optional
PORT=8080
```

## Start the Server

```bash
cd /home/rohit/mydata/code/git_repos/mgsearch
go run main.go
```

## Basic Workflow

### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register/user \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "securepassword123",
    "first_name": "Alice",
    "last_name": "Smith"
  }'
```

**Response:**
```json
{
  "message": "user registered successfully",
  "user": { ... },
  "token": "eyJhbGc..."
}
```

Save the `token` value - you'll need it for authenticated requests.

### 2. Login (Alternative to Registration)

If you already have an account:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "alice@example.com",
    "password": "securepassword123"
  }'
```

### 3. Create a Client (Organization)

```bash
curl -X POST http://localhost:8080/api/v1/auth/register/client \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{
    "name": "my-company",
    "description": "My Company Search Service"
  }'
```

**Response:**
```json
{
  "message": "client registered successfully",
  "client": {
    "id": "675458a3e3b0a...",
    ...
  }
}
```

Save the `client.id` value.

### 4. Generate an API Key

```bash
curl -X POST http://localhost:8080/api/v1/auth/clients/CLIENT_ID_HERE/api-keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{
    "name": "Production API Key",
    "permissions": ["read", "write"]
  }'
```

**Response:**
```json
{
  "message": "API key generated successfully",
  "api_key": "a1b2c3d4e5f6...",
  "key_id": "675458b3e3b0a...",
  "prefix": "a1b2c3d4",
  "warning": "Save this API key now. You won't be able to see it again."
}
```

⚠️ **IMPORTANT**: Save the `api_key` value immediately! You can't retrieve it later.

### 5. Use Your API Key

Now you can use the API key to access protected resources:

#### Search Example
```bash
curl -X POST http://localhost:8080/api/v1/clients/my-company/products/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY_HERE" \
  -d '{
    "q": "laptop",
    "limit": 10
  }'
```

#### Index Document Example
```bash
curl -X POST http://localhost:8080/api/v1/clients/my-company/products/documents \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY_HERE" \
  -d '{
    "id": "prod-123",
    "title": "Gaming Laptop",
    "price": 1299.99,
    "category": "electronics"
  }'
```

#### Update Settings Example
```bash
curl -X PATCH http://localhost:8080/api/v1/clients/my-company/products/settings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY_HERE" \
  -d '{
    "searchableAttributes": ["title", "category"],
    "rankingRules": ["words", "typo", "proximity"]
  }'
```

## Common Operations

### Get Your User Profile
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

### Update Your Profile
```bash
curl -X PUT http://localhost:8080/api/v1/auth/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE" \
  -d '{
    "first_name": "Alicia",
    "last_name": "Johnson"
  }'
```

### List Your Clients
```bash
curl http://localhost:8080/api/v1/auth/clients \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

### Get Client Details (including API keys)
```bash
curl http://localhost:8080/api/v1/auth/clients/CLIENT_ID_HERE \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

### Revoke an API Key
```bash
curl -X DELETE http://localhost:8080/api/v1/auth/clients/CLIENT_ID_HERE/api-keys/KEY_ID_HERE \
  -H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

## Authentication Methods

### JWT Token (for user operations)
Use the token you receive from login/registration:
```bash
-H "Authorization: Bearer YOUR_JWT_TOKEN_HERE"
```

### API Key (for client operations)
Use the API key you generated:
```bash
-H "Authorization: Bearer YOUR_API_KEY_HERE"
```

Or alternatively:
```bash
-H "X-API-Key: YOUR_API_KEY_HERE"
```

## Troubleshooting

### 401 Unauthorized
- Check that you're including the correct token/API key
- Verify the token hasn't expired (JWT tokens expire after 24 hours)
- Check that the API key is active and hasn't been revoked

### 403 Forbidden
- You're trying to access a client you don't have permission for
- Make sure you're using the correct client ID that belongs to your user

### 404 Not Found
- Check the endpoint URL
- Verify the client ID or resource ID exists

### 409 Conflict
- Email already registered (use login instead)
- Client name already exists (use a different name)

## Pro Tips

1. **Store JWT tokens securely** - Never commit them to version control
2. **Rotate API keys regularly** - Generate new keys and revoke old ones periodically
3. **Use descriptive names for API keys** - e.g., "Production Server", "Development", "CI/CD Pipeline"
4. **Set expiration dates** - Add `"expires_at": "2026-12-31T23:59:59Z"` when generating keys
5. **Monitor last_used_at** - Check when API keys were last used to identify unused keys

## Next Steps

- Read the full [API Documentation](AUTH_API.md)
- Review the [Implementation Summary](IMPLEMENTATION_SUMMARY.md)
- Set up proper environment variables for production
- Implement client-side token refresh logic
- Add API key rotation to your security practices

## Support

For issues or questions:
1. Check the logs: `go run main.go` will show detailed error messages
2. Verify your environment variables are set correctly
3. Ensure MongoDB is running and accessible
4. Check the [Auth API Documentation](AUTH_API.md) for detailed endpoint information

Happy coding! 🚀

