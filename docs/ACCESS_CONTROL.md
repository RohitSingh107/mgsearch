# Access Control Documentation

This document describes the access control mechanisms implemented in the MGSearch authentication system.

## Overview

The system implements **role-based access control** where:
- Users can belong to multiple clients
- Clients can have multiple users
- Users can only access clients they are members of

## Access Control Rules

### User-Client Relationship

When a client is created:
1. The creating user is automatically added to the client's `user_ids` array
2. The client ID is automatically added to the user's `client_ids` array
3. This creates a bidirectional many-to-many relationship

### Protected Endpoints

All client-specific endpoints require that the authenticated user belongs to the client:

#### ✅ Protected Client Operations
- `GET /api/v1/auth/clients/:client_id` - Get client details
- `POST /api/v1/auth/clients/:client_id/api-keys` - Generate API key
- `DELETE /api/v1/auth/clients/:client_id/api-keys/:key_id` - Revoke API key

#### ✅ Automatic Access Control
Each protected endpoint:
1. Verifies the user is authenticated (via JWT token)
2. Extracts the user ID from the JWT claims
3. Fetches the client from the database
4. Checks if the user's ID is in the client's `user_ids` array
5. Returns 403 Forbidden if access is denied

## Implementation Details

### Access Verification Function

All client endpoints use a centralized `verifyClientAccess()` helper function:

```go
func (h *UserAuthHandler) verifyClientAccess(
    c *gin.Context, 
    clientID, 
    userID primitive.ObjectID
) (*models.Client, error)
```

**Returns:**
- `(client, nil)` - Access granted, returns the client
- `(nil, nil)` - Access denied (user not in client's user list)
- `(nil, error)` - Client not found or database error

### Response Codes

| Scenario | HTTP Status | Response |
|----------|-------------|----------|
| User not authenticated | 401 | `{"error": "user not authenticated"}` |
| Invalid client ID format | 400 | `{"error": "invalid client ID"}` |
| Client doesn't exist | 404 | `{"error": "client not found"}` |
| User not member of client | 403 | `{"error": "access denied to this client"}` |
| Access granted | 200/201 | Success response |

## Security Considerations

### 1. Separation of Concerns
- **User Authentication**: Handled by JWT middleware
- **Client Authorization**: Handled by `verifyClientAccess()` in handlers
- **API Key Authentication**: Handled by API key middleware (for programmatic access)

### 2. Defense in Depth
Multiple layers of security:

**For JWT Protected Endpoints:**
1. JWT token validation (middleware)
2. User ID extraction from verified token
3. Database lookup to verify current client membership
4. Check against client's user_ids array

**For API Key Protected Endpoints:**
1. API key validation (middleware)
2. Client lookup by API key hash
3. API key expiration check
4. **Client name verification** - URL client_name must match the client that owns the API key

### 3. No ID Enumeration
- Invalid client IDs return 404 (not found)
- Valid client IDs without access return 403 (forbidden)
- This prevents attackers from discovering which client IDs exist

### 4. Real-time Verification
Access is verified on every request:
- No caching of permissions
- Changes to user-client relationships take effect immediately
- Removing a user from a client instantly revokes their access

### 5. API Key Scope Validation
API keys are bound to specific clients:
- Each API key belongs to exactly one client
- The client name in the URL must match the client that owns the API key
- Prevents using an API key to access other clients' resources
- Returns 403 Forbidden if client name doesn't match

## Usage Examples

### Example 1: Authorized Access

```bash
# User alice@example.com creates a client
curl -X POST http://localhost:8080/api/v1/auth/register/client \
  -H "Authorization: Bearer <alice_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "acme-corp"}'

# Response: 201 Created
# alice is now a member of acme-corp

# Alice can generate API keys for her client
curl -X POST http://localhost:8080/api/v1/auth/clients/<client_id>/api-keys \
  -H "Authorization: Bearer <alice_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Production Key"}'

# Response: 201 Created ✅
```

### Example 2: Unauthorized Access Attempt

```bash
# User bob@example.com tries to access Alice's client
curl -X GET http://localhost:8080/api/v1/auth/clients/<alice_client_id> \
  -H "Authorization: Bearer <bob_jwt_token>"

# Response: 403 Forbidden ❌
{
  "error": "access denied to this client"
}
```

### Example 3: List User's Clients

```bash
# Alice lists her accessible clients
curl -X GET http://localhost:8080/api/v1/auth/clients \
  -H "Authorization: Bearer <alice_jwt_token>"

# Response: 200 OK
{
  "clients": [
    {
      "id": "...",
      "name": "acme-corp",
      "user_ids": ["alice_user_id"],
      ...
    }
  ]
}
```

## Multi-Tenant Architecture

The access control system supports multi-tenancy:

### Scenario: Multiple Clients per User

```
User: alice@example.com
├── Client A: acme-corp
├── Client B: widgets-inc
└── Client C: tech-solutions

Alice can:
✅ Manage all three clients
✅ Generate API keys for any of them
✅ View all three in GET /api/v1/auth/clients
```

### Scenario: Multiple Users per Client

```
Client: acme-corp
├── User A: alice@example.com (creator)
├── User B: bob@example.com (added later)
└── User C: charlie@example.com (added later)

All users can:
✅ Generate API keys for acme-corp
✅ View acme-corp details
✅ Revoke API keys for acme-corp
```

## Future Enhancements

### Planned Features

1. **Role-Based Permissions**
   - Admin: Can add/remove users, manage settings
   - Member: Can generate API keys
   - Viewer: Read-only access

2. **User Invitation System**
   - Allow existing members to invite new users
   - Email verification for invitations
   - Pending invitations management

3. **Audit Logging**
   - Track all client access attempts
   - Log API key generation and revocation
   - Monitor user additions/removals

4. **Resource-Level Permissions**
   - Per-index permissions
   - Read vs. Write access control
   - API key scope restrictions

5. **Team Management**
   - Organize users into teams
   - Assign team-level permissions
   - Hierarchical access control

## Troubleshooting

### Issue: User can't access their own client

**Possible causes:**
1. User was removed from the client's `user_ids` array
2. Database inconsistency (user has client_id but client doesn't have user_id)
3. JWT token contains old user ID after account recreation

**Solution:**
```bash
# Check user's client_ids
GET /api/v1/auth/me

# Check client's user_ids
GET /api/v1/auth/clients/:client_id

# Re-add user to client if needed (manual database operation)
db.clients.updateOne(
  { _id: ObjectId("client_id") },
  { $addToSet: { user_ids: ObjectId("user_id") } }
)

db.users.updateOne(
  { _id: ObjectId("user_id") },
  { $addToSet: { client_ids: ObjectId("client_id") } }
)
```

### Issue: 403 Forbidden on valid client

**Debug steps:**
1. Verify JWT token is valid and not expired
2. Check that the user ID in the token matches the database
3. Verify the client ID in the URL is correct
4. Confirm the user's ID is in the client's `user_ids` array

### Issue: Access works intermittently

**Possible causes:**
1. Multiple JWT tokens with different user IDs
2. Database replication lag
3. Caching issues (though we don't cache permissions)

**Solution:**
- Always use the latest JWT token
- Verify database connection is stable
- Check server logs for detailed error messages

## Testing Access Control

Use the Postman collection to test access control:

```
1. Register User A
2. Register User B (different email)
3. User A creates Client X
4. User A can access Client X ✅
5. User B tries to access Client X ❌ (403 Forbidden)
6. User A lists clients → sees Client X ✅
7. User B lists clients → doesn't see Client X ✅
```

## Summary

✅ **Implemented:**
- User-client many-to-many relationships
- JWT-based user authentication
- Client membership verification
- Automatic access control on all client endpoints
- Proper error codes (401, 403, 404)
- Real-time permission checking

🔒 **Security Features:**
- No permission caching
- Defense in depth
- No ID enumeration
- Centralized access verification

📋 **Best Practices:**
- Always verify JWT token first
- Check client access for every operation
- Return appropriate HTTP status codes
- Log access attempts for audit trails

