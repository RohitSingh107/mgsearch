# Postman Collection Guide

This guide explains how to import and use the Postman collection for testing MGSearch Authentication APIs.

## Import the Collection

### Method 1: Import File
1. Open Postman
2. Click **Import** button (top left)
3. Select the file: `postman_collection.json`
4. Click **Import**

### Method 2: Drag and Drop
1. Open Postman
2. Drag the `postman_collection.json` file into the Postman window
3. The collection will be automatically imported

## Collection Structure

The collection is organized into 5 folders:

### 📁 Authentication
- **Register User** - Create a new user account
- **Login** - Login with email/password
- **Get Current User** - Get authenticated user's profile
- **Update User Profile** - Update user information

### 📁 Client Management
- **Register Client** - Create a new client/organization
- **Get User's Clients** - List all clients for the user
- **Get Client Details** - Get specific client with API keys

### 📁 API Key Management
- **Generate API Key** - Create an API key
- **Generate API Key with Expiration** - Create a temporary API key
- **Revoke API Key** - Deactivate an API key

### 📁 API Key Protected Endpoints
- **Search Documents (API Key)** - Search using Authorization header
- **Search Documents (X-API-Key)** - Search using X-API-Key header
- **Index Document** - Add documents to the index
- **Update Index Settings** - Configure Meilisearch settings
- **Get Task Status** - Check Meilisearch task status

### 📁 Public Endpoints
- **Storefront Search (GET)** - Public search via GET
- **Storefront Search (POST)** - Public search via POST
- **Health Check** - Verify server is running

## Collection Variables

The collection uses variables to automatically store and reuse values:

| Variable | Description | Auto-saved |
|----------|-------------|------------|
| `base_url` | Server URL (default: http://localhost:8080) | Manual |
| `jwt_token` | JWT authentication token | ✅ Auto |
| `user_id` | Current user ID | ✅ Auto |
| `client_id` | Current client ID | ✅ Auto |
| `api_key` | Generated API key | ✅ Auto |
| `key_id` | API key ID | ✅ Auto |

### How Auto-save Works
The collection includes test scripts that automatically extract and save important values:
- When you **Register** or **Login**, the JWT token is saved
- When you **Register Client**, the client ID is saved
- When you **Generate API Key**, the API key and key ID are saved

These values are then automatically used in subsequent requests!

## Quick Start Workflow

### 1. Start Your Server
```bash
cd /home/rohit/mydata/code/git_repos/mgsearch
go run main.go
```

### 2. Update Base URL (if needed)
- Go to collection variables (click on the collection → Variables tab)
- Update `base_url` if your server is not on `http://localhost:8080`

### 3. Test the Server
Run: **Health Check** → Should return `{"message": "pong"}`

### 4. Register a User
1. Run: **Authentication → Register User**
2. The response will show your user details and JWT token
3. The JWT token is automatically saved to `{{jwt_token}}`
4. Check the Postman Console to see: "JWT Token saved: ..."

### 5. Create a Client
1. Run: **Client Management → Register Client**
2. Change the client name in the request body if needed
3. The client ID is automatically saved to `{{client_id}}`

### 6. Generate an API Key
1. Run: **API Key Management → Generate API Key**
2. ⚠️ **IMPORTANT**: Copy the API key from the response immediately!
3. The API key is automatically saved to `{{api_key}}`

### 7. Test API Key
1. Update the URL in any "API Key Protected Endpoints" request
   - Replace `my-company` with your actual client name
   - Replace `products` with your index name
2. Run: **Search Documents (API Key)**

## Tips and Tricks

### 1. View Saved Variables
- Click on the collection name
- Go to the **Variables** tab
- You'll see all current values

### 2. Manually Set Variables
If auto-save doesn't work or you want to use existing values:
```
1. Click on the collection → Variables tab
2. Update the "Current value" column
3. Click Save
```

### 3. Check Console Logs
- Open Postman Console (bottom left icon or View → Show Postman Console)
- See detailed logs of what variables were saved

### 4. Use Different Environments
Create Postman environments for different servers:
- **Local**: `http://localhost:8080`
- **Staging**: `https://staging.example.com`
- **Production**: `https://api.example.com`

### 5. Re-login if Token Expires
JWT tokens expire after 24 hours. If you get a 401 error:
1. Run: **Authentication → Login**
2. Your token will be refreshed automatically

### 6. Test Both API Key Methods
The collection includes two ways to send API keys:
- `Authorization: Bearer {{api_key}}` (standard)
- `X-API-Key: {{api_key}}` (alternative)

Both work the same way - use whichever you prefer!

## Common Scenarios

### Scenario 1: New User Testing
```
1. Register User
2. Register Client
3. Generate API Key
4. Search Documents
```

### Scenario 2: Existing User
```
1. Login (instead of Register User)
2. Get User's Clients (to see existing clients)
3. Get Client Details (to see existing API keys)
4. Use existing API key or generate a new one
```

### Scenario 3: Multiple Clients
```
1. Login
2. Register Client (first client)
3. Generate API Key for first client
4. Register Client (second client)  
5. Generate API Key for second client
6. Switch between clients by updating {{client_id}} and {{api_key}}
```

### Scenario 4: API Key Management
```
1. Login
2. Get Client Details (see all API keys)
3. Generate API Key (create new key)
4. Test with new key
5. Revoke API Key (deactivate old key)
```

## Request Body Examples

### Register User
```json
{
  "email": "alice@example.com",
  "password": "securepassword123",
  "first_name": "Alice",
  "last_name": "Johnson"
}
```

### Register Client
```json
{
  "name": "acme-corp",
  "description": "Acme Corporation Search Service"
}
```

### Generate API Key
```json
{
  "name": "Production Key",
  "permissions": ["read", "write"]
}
```

### Generate API Key with Expiration
```json
{
  "name": "Temporary Key",
  "permissions": ["read"],
  "expires_at": "2026-12-31T23:59:59Z"
}
```

### Search Documents
```json
{
  "q": "laptop",
  "limit": 20,
  "offset": 0
}
```

### Index Document
```json
{
  "id": "prod-123",
  "title": "Gaming Laptop",
  "description": "High-performance gaming laptop",
  "price": 1299.99,
  "category": "electronics",
  "tags": ["gaming", "laptop", "electronics"]
}
```

### Update Settings
```json
{
  "searchableAttributes": ["title", "description", "category"],
  "rankingRules": ["words", "typo", "proximity"],
  "filterableAttributes": ["category", "price"]
}
```

## Troubleshooting

### 401 Unauthorized
**Problem**: Token is invalid or expired
**Solution**: 
- Run the **Login** request again
- Check that `{{jwt_token}}` variable is set

### 403 Forbidden
**Problem**: You don't have access to this client OR API key doesn't match client name
**Solution**:
- **For JWT endpoints**: Verify you're using the correct `{{client_id}}` and run **Get User's Clients**
- **For API Key endpoints**: Ensure the client name in the URL matches the client that owns the API key
  - Example: If your API key belongs to "acme-corp", the URL must be `/api/v1/clients/acme-corp/...`
  - You cannot use "acme-corp"'s API key to access `/api/v1/clients/other-company/...`

### 404 Not Found
**Problem**: Resource doesn't exist
**Solution**:
- Check the URL parameters (client_id, key_id, etc.)
- Verify the client name and index name in the URL

### 409 Conflict
**Problem**: Email or client name already exists
**Solution**:
- Use a different email for user registration
- Use a different name for client registration

### Variables Not Saving
**Problem**: Auto-save scripts not running
**Solution**:
- Check Postman Console for error messages
- Manually copy values from response to collection variables
- Ensure the response is successful (201 or 200)

## Advanced Features

### Pre-request Scripts
You can add pre-request scripts to generate dynamic data:

```javascript
// Generate random email
pm.collectionVariables.set("random_email", 
  `user${Date.now()}@example.com`);

// Generate timestamp
pm.collectionVariables.set("timestamp", 
  new Date().toISOString());
```

### Test Scripts
The collection includes test scripts that:
- Validate response status codes
- Extract and save tokens/IDs
- Log important information to console

You can extend these scripts to add more validations!

### Running Collection with Newman
Run the entire collection from command line:

```bash
# Install Newman
npm install -g newman

# Run collection
newman run postman_collection.json \
  --env-var "base_url=http://localhost:8080"
```

## Security Notes

⚠️ **Important Security Considerations:**

1. **Never commit API keys to version control**
2. **Use environment-specific variables** for different servers
3. **Rotate API keys regularly** in production
4. **Store production credentials securely** (use Postman Vault or environment variables)
5. **Don't share Postman collections** with sensitive data embedded

## Testing Access Control

To test that access control works correctly:

### Test Case: Unauthorized Access
```
1. Register User A (alice@example.com)
2. Register User B (bob@example.com)
3. User A creates Client X
4. User A can access Client X ✅
5. Copy Client X's ID
6. Switch JWT token to User B's token
7. Try to access Client X with User B's token
8. Should get 403 Forbidden ❌
```

**How to switch tokens:**
1. After registering User B, copy their JWT token
2. Go to Collection Variables
3. Update `jwt_token` with User B's token
4. Try accessing User A's client
5. You should see: `{"error": "access denied to this client"}`

See [Access Control Documentation](ACCESS_CONTROL.md) for detailed information.

## Support

For more information:
- [API Documentation](AUTH_API.md)
- [Quick Start Guide](QUICK_START.md)
- [Implementation Details](IMPLEMENTATION_SUMMARY.md)
- [Access Control Guide](ACCESS_CONTROL.md)

Happy testing! 🚀

