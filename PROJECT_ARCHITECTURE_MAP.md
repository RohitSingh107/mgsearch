# MGSearch - Complete Project Architecture Map

## Table of Contents
1. [Project Overview](#project-overview)
2. [Architecture Diagram](#architecture-diagram)
3. [Data Models & Relationships](#data-models--relationships)
4. [API Endpoints Map](#api-endpoints-map)
5. [Authentication Flows](#authentication-flows)
6. [Service Integration Map](#service-integration-map)
7. [Use Cases & Data Flows](#use-cases--data-flows)
8. [Technology Stack](#technology-stack)

---

## Project Overview

**MGSearch** is a dual-purpose search microservice built with Go that provides:
1. **Shopify App Backend**: Search infrastructure for Shopify merchants
2. **SaaS Search Platform**: Multi-tenant search service for general clients

### Core Capabilities
- Full-text search via Meilisearch
- Vector-based similarity search via Qdrant
- Multi-tenant architecture
- Secure authentication (JWT + API Keys)
- Real-time webhook processing
- Shopify OAuth integration

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │
│  │  Shopify Admin   │  │  Shopify         │  │  SaaS Dashboard  │  │
│  │  Dashboard       │  │  Storefront      │  │  (Web/Mobile)    │  │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘  │
│           │                     │                     │             │
│           │ (JWT)               │ (Storefront Key)    │ (JWT/API)   │
│           │                     │                     │             │
└───────────┼─────────────────────┼─────────────────────┼─────────────┘
            │                     │                     │
            ▼                     ▼                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      MIDDLEWARE LAYER                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │ JWT          │  │ API Key      │  │ Auth         │              │
│  │ Middleware   │  │ Middleware   │  │ Middleware   │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
│                                                                       │
│  ┌──────────────────────────────────────────────────┐              │
│  │            CORS Middleware                        │              │
│  └──────────────────────────────────────────────────┘              │
│                                                                       │
└───────────────────────────────────┬─────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        HANDLER LAYER (API Routes)                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │ User Auth       │  │ Search          │  │ Storefront      │    │
│  │ Handler         │  │ Handler         │  │ Handler         │    │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘    │
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │ Shopify Auth    │  │ Store           │  │ Session         │    │
│  │ Handler         │  │ Handler         │  │ Handler         │    │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘    │
│                                                                       │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐    │
│  │ Index           │  │ Settings        │  │ Tasks           │    │
│  │ Handler         │  │ Handler         │  │ Handler         │    │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘    │
│                                                                       │
│  ┌─────────────────┐                                                 │
│  │ Webhook         │                                                 │
│  │ Handler         │                                                 │
│  └─────────────────┘                                                 │
│                                                                       │
└───────────────────────────────────┬─────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     REPOSITORY LAYER (Data Access)                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │ User         │  │ Client       │  │ Store        │              │
│  │ Repository   │  │ Repository   │  │ Repository   │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
│                                                                       │
│  ┌──────────────┐  ┌──────────────┐                                 │
│  │ Session      │  │ Index        │                                 │
│  │ Repository   │  │ Repository   │                                 │
│  └──────────────┘  └──────────────┘                                 │
│                                                                       │
└───────────────────────────────────┬─────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    EXTERNAL SERVICES LAYER                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌────────────────────┐  ┌────────────────────┐                     │
│  │  Meilisearch       │  │  Qdrant            │                     │
│  │  Service           │  │  Service           │                     │
│  │  (Search)          │  │  (Vector Search)   │                     │
│  └────────────────────┘  └────────────────────┘                     │
│                                                                       │
│  ┌────────────────────┐                                              │
│  │  Shopify           │                                              │
│  │  Service           │                                              │
│  │  (OAuth/API)       │                                              │
│  └────────────────────┘                                              │
│                                                                       │
└───────────────────────────────────┬─────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      DATA STORAGE LAYER                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ┌────────────────────────────────────────────────────┐             │
│  │              MongoDB Database                       │             │
│  │                                                     │             │
│  │  • users          • clients      • stores          │             │
│  │  • sessions       • indexes                        │             │
│  └────────────────────────────────────────────────────┘             │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Data Models & Relationships

### Entity Relationship Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                        DATA MODEL RELATIONSHIPS                      │
└────────────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │     User        │
                    ├─────────────────┤
                    │ • _id (ObjID)   │
                    │ • email         │
                    │ • password_hash │
                    │ • first_name    │
                    │ • last_name     │
                    │ • client_ids[]  │◄────────┐
                    │ • is_active     │         │
                    │ • created_at    │         │
                    │ • updated_at    │         │
                    └─────────────────┘         │
                                                │ Many-to-Many
                    ┌─────────────────┐         │
                    │    Client       │         │
                    ├─────────────────┤         │
                    │ • _id (ObjID)   │─────────┘
                    │ • name          │
                    │ • description   │
                    │ • user_ids[]    │◄────────┐
                    │ • api_keys[]    │◄────────┤
                    │ • is_active     │         │
                    │ • created_at    │         │
                    │ • updated_at    │         │
                    └────────┬────────┘         │
                             │                  │
                             │ One-to-Many      │ One-to-Many
                             │                  │
                             ▼                  │
                    ┌─────────────────┐         │
                    │     Index       │         │
                    ├─────────────────┤         │
                    │ • _id (ObjID)   │         │
                    │ • client_id     │─────────┘
                    │ • name          │
                    │ • uid           │
                    │ • primary_key   │
                    │ • created_at    │
                    │ • updated_at    │
                    └─────────────────┘


                    ┌─────────────────────┐
                    │   APIKey (Embedded) │
                    ├─────────────────────┤
                    │ • _id (ObjID)       │
                    │ • key (hashed)      │
                    │ • name              │
                    │ • key_prefix        │
                    │ • permissions[]     │
                    │ • is_active         │
                    │ • last_used_at      │
                    │ • created_at        │
                    │ • expires_at        │
                    └─────────────────────┘


┌──────────────────────────── SHOPIFY MODELS ────────────────────────────┐
│                                                                          │
│   ┌─────────────────────┐              ┌─────────────────────┐         │
│   │      Store          │              │      Session        │         │
│   ├─────────────────────┤              ├─────────────────────┤         │
│   │ • _id (ObjID)       │              │ • _id (string)      │         │
│   │ • shop_domain       │◄──────┐      │ • shop              │         │
│   │ • shop_name         │       │      │ • state             │         │
│   │ • encrypted_token   │       │      │ • is_online         │         │
│   │ • api_key_public    │       │      │ • scope             │         │
│   │ • api_key_private   │       │      │ • expires           │         │
│   │ • product_index_uid │       │      │ • access_token      │         │
│   │ • meilisearch_*     │       │      │ • user_id           │         │
│   │ • qdrant_collection │       │      │ • created_at        │         │
│   │ • plan_level        │       │      │ • updated_at        │         │
│   │ • status            │       │      └─────────────────────┘         │
│   │ • webhook_secret    │       │                                      │
│   │ • installed_at      │       │                                      │
│   │ • sync_state        │       │                                      │
│   └─────────────────────┘       │                                      │
│                                 │ Related by shop_domain               │
│                                 │                                      │
└─────────────────────────────────┴──────────────────────────────────────┘
```

### Model Details

#### 1. **User** (SaaS Platform Users)
- Represents dashboard/admin users
- Can own multiple clients
- Authenticated via JWT tokens
- Password stored as bcrypt hash

#### 2. **Client** (SaaS Tenants)
- Multi-tenant entity
- Contains multiple API keys
- Can have multiple indexes
- Associated with multiple users

#### 3. **Index** (Meilisearch Indexes)
- Belongs to a single client
- Maps to actual Meilisearch index
- UID format: `{client_name}__{index_name}`

#### 4. **APIKey** (Embedded in Client)
- SHA-256 hashed for security
- Can have expiration date
- Tracks last usage
- Scoped to specific client

#### 5. **Store** (Shopify Merchants)
- Represents installed Shopify app
- Stores encrypted Shopify access token
- Has dedicated Meilisearch index
- Configured with Qdrant collection

#### 6. **Session** (Shopify OAuth Sessions)
- Managed by Remix frontend
- Stores Shopify OAuth data
- Auto-creates Store on storage

---

## API Endpoints Map

### Endpoint Overview by Category

```
┌─────────────────────────────────────────────────────────────────┐
│                    API ENDPOINTS STRUCTURE                       │
└─────────────────────────────────────────────────────────────────┘

PUBLIC ENDPOINTS (No Auth)
├── GET  /ping                           → Health check

SHOPIFY STOREFRONT (Storefront Key Required)
├── GET  /api/v1/search                  → Search products
├── POST /api/v1/search                  → Search with filters
├── GET  /api/v1/similar                 → Similar products
└── POST /api/v1/similar                 → Similar products

SAAS USER AUTHENTICATION (JWT)
├── POST /api/v1/auth/register/user      → Register new user
├── POST /api/v1/auth/login              → User login
├── GET  /api/v1/auth/me                 → Get current user
└── PUT  /api/v1/auth/user               → Update user

SAAS CLIENT MANAGEMENT (JWT Required)
├── POST /api/v1/auth/register/client           → Create new client
├── GET  /api/v1/clients                        → List user's clients
├── GET  /api/v1/clients/:id                    → Get client details
├── POST /api/v1/clients/:id/api-keys           → Generate API key
├── DELETE /api/v1/clients/:id/api-keys/:key_id → Revoke API key
├── POST /api/v1/clients/:id/indexes            → Create index
├── GET  /api/v1/clients/:id/indexes            → List indexes
├── POST /api/v1/clients/:id/indexes/:index/documents → Index document
└── PATCH /api/v1/clients/:id/indexes/:index/settings → Update settings

SAAS SEARCH API (API Key Required)
├── POST /api/v1/clients/:id/indexes/:index/search → Perform search
└── GET  /api/v1/clients/:id/tasks/:task_id        → Get task status

SHOPIFY OAUTH (Legacy)
├── POST /api/auth/shopify/begin          → Start OAuth flow
├── GET  /api/auth/shopify/callback       → OAuth callback
├── POST /api/auth/shopify/exchange       → Exchange code for token
└── POST /api/auth/shopify/install        → Install store

SHOPIFY STORE MANAGEMENT (Session JWT)
├── GET  /api/stores/current              → Get current store
└── GET  /api/stores/sync-status          → Get sync status

SHOPIFY SESSION MANAGEMENT (Optional API Key)
├── POST   /api/sessions                  → Store/update session
├── GET    /api/sessions/:id              → Load session
├── DELETE /api/sessions/:id              → Delete session
├── DELETE /api/sessions/batch            → Delete multiple sessions
└── GET    /api/sessions/shop/:shop       → Get sessions by shop

SHOPIFY WEBHOOKS (HMAC Verified)
└── POST /webhooks/shopify/:topic/:subtopic → Handle webhook
```

### Detailed Endpoint Specifications

#### **SaaS Platform Endpoints**

##### 1. User Authentication

```
POST /api/v1/auth/register/user
├─ Purpose: Register new user account
├─ Auth: None
├─ Body:
│  {
│    "email": "user@example.com",
│    "password": "securepassword123",
│    "first_name": "John",
│    "last_name": "Doe"
│  }
└─ Response: User object + JWT token

POST /api/v1/auth/login
├─ Purpose: Authenticate user
├─ Auth: None
├─ Body:
│  {
│    "email": "user@example.com",
│    "password": "securepassword123"
│  }
└─ Response: User object + JWT token

GET /api/v1/auth/me
├─ Purpose: Get current user info
├─ Auth: JWT (Bearer token)
└─ Response: User object

PUT /api/v1/auth/user
├─ Purpose: Update user profile
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "first_name": "Jane",
│    "last_name": "Smith"
│  }
└─ Response: Updated user object
```

##### 2. Client Management

```
POST /api/v1/auth/register/client
├─ Purpose: Create new client (tenant)
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "name": "my-app",
│    "description": "My Application"
│  }
└─ Response: Client object

GET /api/v1/clients
├─ Purpose: List all user's clients
├─ Auth: JWT (Bearer token)
└─ Response: Array of client objects

POST /api/v1/clients/:client_id/api-keys
├─ Purpose: Generate new API key
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "name": "Production Key",
│    "permissions": ["search", "index"],
│    "expires_at": "2025-12-31T23:59:59Z"  // optional
│  }
└─ Response: API key (shown only once)

DELETE /api/v1/clients/:client_id/api-keys/:key_id
├─ Purpose: Revoke API key
├─ Auth: JWT (Bearer token)
└─ Response: Success message
```

##### 3. Index Management

```
POST /api/v1/clients/:client_id/indexes
├─ Purpose: Create new search index
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "name": "movies",
│    "primary_key": "id"  // optional
│  }
└─ Response: Index object + Meilisearch task

GET /api/v1/clients/:client_id/indexes
├─ Purpose: List client's indexes
├─ Auth: JWT (Bearer token)
└─ Response: Array of index objects
```

##### 4. Document Management

```
POST /api/v1/clients/:client_id/indexes/:index_name/documents
├─ Purpose: Index a document
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "id": "123",
│    "title": "Inception",
│    "genre": "Sci-Fi",
│    "year": 2010
│  }
└─ Response: Meilisearch task object
```

##### 5. Search Operations

```
POST /api/v1/clients/:client_id/indexes/:index_name/search
├─ Purpose: Perform search
├─ Auth: API Key (Bearer token)
├─ Body:
│  {
│    "q": "inception",
│    "filter": "genre = 'Sci-Fi'",
│    "sort": ["year:desc"],
│    "limit": 20,
│    "offset": 0
│  }
└─ Response: Meilisearch search results
```

##### 6. Index Settings

```
PATCH /api/v1/clients/:client_id/indexes/:index_name/settings
├─ Purpose: Update index settings
├─ Auth: JWT (Bearer token)
├─ Body:
│  {
│    "searchableAttributes": ["title", "description"],
│    "filterableAttributes": ["genre", "year"],
│    "sortableAttributes": ["year", "rating"],
│    "rankingRules": [
│      "words",
│      "typo",
│      "proximity",
│      "attribute",
│      "sort",
│      "exactness"
│    ]
│  }
└─ Response: Meilisearch task object
```

##### 7. Task Status

```
GET /api/v1/clients/:client_id/tasks/:task_id
├─ Purpose: Get task status
├─ Auth: API Key (Bearer token)
└─ Response: Meilisearch task details
```

#### **Shopify Platform Endpoints**

##### 1. Storefront Search (Public with Storefront Key)

```
GET/POST /api/v1/search
├─ Purpose: Search products in storefront
├─ Auth: X-Storefront-Key header
├─ Query Params (GET):
│  • q: Search query
│  • limit: Results limit (default: 20)
│  • offset: Pagination offset
│  • sort: Sort parameters
│  • filters: JSON filter string
└─ Body (POST):
   {
     "q": "shoes",
     "filter": "price < 100",
     "limit": 20
   }
└─ Response: Meilisearch results

GET/POST /api/v1/similar
├─ Purpose: Get similar products (vector search)
├─ Auth: X-Storefront-Key header
├─ Query Params (GET):
│  • id: Product ID
│  • limit: Results limit (default: 10)
└─ Body (POST):
   {
     "id": 123456789,
     "limit": 10
   }
└─ Response: Qdrant recommendations
```

##### 2. OAuth Flow

```
POST /api/auth/shopify/begin
├─ Purpose: Initiate Shopify OAuth
├─ Auth: None
├─ Body:
│  {
│    "shop": "mystore.myshopify.com",
│    "redirect_uri": "https://app.example.com/callback"
│  }
└─ Response: Authorization URL + state token

GET /api/auth/shopify/callback
├─ Purpose: Handle OAuth callback
├─ Auth: HMAC signature validation
├─ Query Params:
│  • shop: Shop domain
│  • code: Authorization code
│  • state: State token
│  • hmac: HMAC signature
└─ Response: Store object + session JWT

POST /api/auth/shopify/install
├─ Purpose: Install store with access token
├─ Auth: None
├─ Body:
│  {
│    "shop": "mystore.myshopify.com",
│    "access_token": "shpat_...",
│    "shop_name": "My Store"
│  }
└─ Response: Store object + session JWT
```

##### 3. Store Management

```
GET /api/stores/current
├─ Purpose: Get current store details
├─ Auth: Shopify Session JWT
└─ Response: Store public view

GET /api/stores/sync-status
├─ Purpose: Get product sync status
├─ Auth: Shopify Session JWT
└─ Response: Sync state information
```

##### 4. Session Management (Remix Backend)

```
POST /api/sessions
├─ Purpose: Store/update Shopify session
├─ Auth: Optional (SESSION_API_KEY)
├─ Body: Shopify session object
└─ Response: Success message
└─ Side Effect: Auto-creates Store record

GET /api/sessions/:id
├─ Purpose: Load session by ID
├─ Auth: Optional (SESSION_API_KEY)
└─ Response: Session object (decrypted)

DELETE /api/sessions/:id
├─ Purpose: Delete session
├─ Auth: Optional (SESSION_API_KEY)
└─ Response: 204 No Content

GET /api/sessions/shop/:shop
├─ Purpose: Get all sessions for shop
├─ Auth: Optional (SESSION_API_KEY)
└─ Response: Array of sessions
```

##### 5. Webhooks

```
POST /webhooks/shopify/:topic/:subtopic
├─ Purpose: Handle Shopify webhooks
├─ Auth: HMAC signature verification
├─ Supported Events:
│  • products/create → Index new product
│  • products/update → Update product in index
│  • products/delete → Remove product from index
└─ Response: Processing status
```

---

## Authentication Flows

### Authentication Methods Overview

```
┌────────────────────────────────────────────────────────────────┐
│                   AUTHENTICATION METHODS                        │
└────────────────────────────────────────────────────────────────┘

1. JWT Token (SaaS Platform Users)
   ├─ Used For: Dashboard access, admin operations
   ├─ Header: Authorization: Bearer <jwt_token>
   ├─ Contains: user_id, email, client_id (optional)
   └─ Expiry: 24 hours

2. Client API Key (SaaS Platform Applications)
   ├─ Used For: Programmatic search access
   ├─ Header: Authorization: Bearer <api_key> or X-API-Key
   ├─ Hashed: SHA-256
   ├─ Scoped: To specific client
   └─ Expiry: Optional

3. Shopify Session JWT (Shopify Admin)
   ├─ Used For: Shopify app admin endpoints
   ├─ Header: Authorization: Bearer <session_token>
   ├─ Contains: store_id, shop_domain
   └─ Expiry: 24 hours

4. Storefront Key (Shopify Public Search)
   ├─ Used For: Public product search
   ├─ Header: X-Storefront-Key: <public_key>
   ├─ Scope: Read-only search
   └─ No expiry

5. HMAC Signature (Shopify Webhooks)
   ├─ Used For: Webhook validation
   ├─ Header: X-Shopify-Hmac-Sha256
   ├─ Validates: Request authenticity
   └─ Secret: Shared with Shopify

6. Optional Session API Key (Session Endpoints)
   ├─ Used For: Backend session management
   ├─ Header: Authorization: Bearer <session_api_key>
   └─ Optional: If not set, no auth required
```

### Detailed Authentication Flows

#### Flow 1: SaaS User Registration & Login

```
┌────────┐                     ┌────────────┐                  ┌──────────┐
│ Client │                     │  MGSearch  │                  │ MongoDB  │
└───┬────┘                     └─────┬──────┘                  └────┬─────┘
    │                                │                              │
    │ POST /auth/register/user       │                              │
    │ {email, password, name}        │                              │
    ├───────────────────────────────>│                              │
    │                                │                              │
    │                                │ Hash password (bcrypt)       │
    │                                │ Create user record           │
    │                                ├─────────────────────────────>│
    │                                │                              │
    │                                │<─────────────────────────────┤
    │                                │ User saved                   │
    │                                │                              │
    │                                │ Generate JWT                 │
    │                                │ (user_id, email, 24h exp)    │
    │                                │                              │
    │<───────────────────────────────┤                              │
    │ {user, token}                  │                              │
    │                                │                              │
    │ POST /auth/login               │                              │
    │ {email, password}              │                              │
    ├───────────────────────────────>│                              │
    │                                │                              │
    │                                │ Find user by email           │
    │                                ├─────────────────────────────>│
    │                                │<─────────────────────────────┤
    │                                │                              │
    │                                │ Verify password              │
    │                                │ Generate JWT                 │
    │                                │                              │
    │<───────────────────────────────┤                              │
    │ {user, token}                  │                              │
    │                                │                              │
```

#### Flow 2: Client Creation & API Key Generation

```
┌────────┐                     ┌────────────┐                  ┌──────────┐
│ Client │                     │  MGSearch  │                  │ MongoDB  │
└───┬────┘                     └─────┬──────┘                  └────┬─────┘
    │                                │                              │
    │ POST /auth/register/client     │                              │
    │ Authorization: Bearer <JWT>    │                              │
    │ {name, description}            │                              │
    ├───────────────────────────────>│                              │
    │                                │                              │
    │                                │ Verify JWT                   │
    │                                │ Extract user_id              │
    │                                │                              │
    │                                │ Create client                │
    │                                │ Link to user                 │
    │                                ├─────────────────────────────>│
    │                                │<─────────────────────────────┤
    │                                │                              │
    │<───────────────────────────────┤                              │
    │ {client}                       │                              │
    │                                │                              │
    │ POST /clients/:id/api-keys     │                              │
    │ Authorization: Bearer <JWT>    │                              │
    │ {name, permissions}            │                              │
    ├───────────────────────────────>│                              │
    │                                │                              │
    │                                │ Generate random key (32B)    │
    │                                │ Hash key (SHA-256)           │
    │                                │ Store hashed key             │
    │                                ├─────────────────────────────>│
    │                                │<─────────────────────────────┤
    │                                │                              │
    │<───────────────────────────────┤                              │
    │ {api_key: "raw_key_once"}      │                              │
    │ ⚠️  Save this! Won't show again │                              │
    │                                │                              │
```

#### Flow 3: Shopify OAuth Installation

```
┌────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐  ┌──────────┐
│Shopify │  │  Frontend  │  │  MGSearch  │  │ MongoDB  │  │Meilisrch │
│Merchant│  │   (Remix)  │  │  Backend   │  │          │  │          │
└───┬────┘  └─────┬──────┘  └─────┬──────┘  └────┬─────┘  └────┬─────┘
    │             │               │              │             │
    │ Click       │               │              │             │
    │ "Install"   │               │              │             │
    ├────────────>│               │              │             │
    │             │               │              │             │
    │             │ POST /auth/shopify/begin     │             │
    │             │ {shop, redirect_uri}         │             │
    │             ├──────────────>│              │             │
    │             │               │              │             │
    │             │               │ Generate     │             │
    │             │               │ state token  │             │
    │             │               │ (JWT, 15min) │             │
    │             │               │              │             │
    │             │<──────────────┤              │             │
    │             │ {authUrl, state}             │             │
    │             │               │              │             │
    │<────────────┤               │              │             │
    │ Redirect to │               │              │             │
    │ Shopify     │               │              │             │
    │             │               │              │             │
    │ Grant       │               │              │             │
    │ Permissions │               │              │             │
    │             │               │              │             │
    │ Redirect    │               │              │             │
    │ to callback │               │              │             │
    ├────────────────────────────>│              │             │
    │ ?code=xxx&state=yyy&hmac=zzz│              │             │
    │             │               │              │             │
    │             │               │ Validate HMAC│             │
    │             │               │ Verify state │             │
    │             │               │              │             │
    │             │               │ Exchange code│             │
    │             │               │ for token    │             │
    │             │               │ (Shopify API)│             │
    │             │               │              │             │
    │             │               │ Encrypt token│             │
    │             │               │ Generate keys│             │
    │             │               │              │             │
    │             │               │ Create Store │             │
    │             │               ├─────────────>│             │
    │             │               │<─────────────┤             │
    │             │               │              │             │
    │             │               │ Create Index │             │
    │             │               ├──────────────────────────>│
    │             │               │<──────────────────────────┤
    │             │               │              │             │
    │             │               │ Generate     │             │
    │             │               │ session JWT  │             │
    │             │               │              │             │
    │<────────────────────────────┤              │             │
    │ {store, token}              │              │             │
    │             │               │              │             │
```

#### Flow 4: Storefront Search

```
┌──────────┐         ┌────────────┐         ┌──────────────┐
│ Shopify  │         │  MGSearch  │         │ Meilisearch  │
│Storefront│         │  Backend   │         │              │
└────┬─────┘         └─────┬──────┘         └──────┬───────┘
     │                     │                       │
     │ GET /api/v1/search  │                       │
     │ X-Storefront-Key    │                       │
     │ ?q=shoes&limit=20   │                       │
     ├────────────────────>│                       │
     │                     │                       │
     │                     │ Validate key          │
     │                     │ Find store by key     │
     │                     │                       │
     │                     │ Search in index       │
     │                     ├──────────────────────>│
     │                     │                       │
     │                     │<──────────────────────┤
     │                     │ Search results        │
     │                     │                       │
     │<────────────────────┤                       │
     │ {hits, ...}         │                       │
     │                     │                       │
```

#### Flow 5: Similar Products (Vector Search)

```
┌──────────┐    ┌────────────┐    ┌────────┐    ┌──────────────┐
│ Shopify  │    │  MGSearch  │    │MongoDB │    │    Qdrant    │
│Storefront│    │  Backend   │    │        │    │              │
└────┬─────┘    └─────┬──────┘    └───┬────┘    └──────┬───────┘
     │                │               │                 │
     │ GET /similar?id=123            │                 │
     │ X-Storefront-Key               │                 │
     ├───────────────>│               │                 │
     │                │               │                 │
     │                │ Validate key  │                 │
     │                │ Get store     │                 │
     │                ├──────────────>│                 │
     │                │<──────────────┤                 │
     │                │               │                 │
     │                │ Get collection name             │
     │                │               │                 │
     │                │ Recommend     │                 │
     │                │ (product_id)  │                 │
     │                ├──────────────────────────────> │
     │                │               │                 │
     │                │<──────────────────────────────┤│
     │                │ Similar products               │
     │                │               │                 │
     │<───────────────┤               │                 │
     │ {results}      │               │                 │
     │                │               │                 │
```

#### Flow 6: Webhook Processing

```
┌────────┐         ┌────────────┐         ┌──────────┐         ┌──────────────┐
│Shopify │         │  MGSearch  │         │ MongoDB  │         │ Meilisearch  │
│        │         │  Backend   │         │          │         │              │
└───┬────┘         └─────┬──────┘         └────┬─────┘         └──────┬───────┘
    │                    │                     │                      │
    │ POST /webhooks/    │                     │                      │
    │ shopify/products/  │                     │                      │
    │ update             │                     │                      │
    │ X-Shopify-Hmac-Sha256                    │                      │
    │ X-Shopify-Shop-Domain                    │                      │
    │ {product data}     │                     │                      │
    ├───────────────────>│                     │                      │
    │                    │                     │                      │
    │                    │ Verify HMAC         │                      │
    │                    │ (signature valid?)  │                      │
    │                    │                     │                      │
    │                    │ Get store           │                      │
    │                    ├────────────────────>│                      │
    │                    │<────────────────────┤                      │
    │                    │                     │                      │
    │                    │ Enrich document     │                      │
    │                    │ (add shop, store_id)│                      │
    │                    │                     │                      │
    │                    │ Index/Update document                      │
    │                    ├───────────────────────────────────────────>│
    │                    │                     │                      │
    │                    │<───────────────────────────────────────────┤
    │                    │ Task created        │                      │
    │                    │                     │                      │
    │<───────────────────┤                     │                      │
    │ {status: "processed"}                    │                      │
    │                    │                     │                      │
```

---

## Service Integration Map

### External Service Dependencies

```
┌────────────────────────────────────────────────────────────────────┐
│                    EXTERNAL SERVICE INTEGRATIONS                    │
└────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                         Meilisearch                              │
│                     (Primary Search Engine)                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Operations:                                                     │
│  ├─ Create Index                                                 │
│  ├─ Index Documents (single/bulk)                                │
│  ├─ Search Documents                                             │
│  ├─ Update Settings                                              │
│  ├─ Delete Documents                                             │
│  └─ Get Task Status                                              │
│                                                                  │
│  Features:                                                       │
│  ├─ Full-text search                                             │
│  ├─ Typo tolerance                                               │
│  ├─ Filtering & faceting                                         │
│  ├─ Sorting                                                      │
│  ├─ Highlighting                                                 │
│  └─ Custom ranking rules                                         │
│                                                                  │
│  Configuration:                                                  │
│  ├─ URL: MEILISEARCH_URL                                         │
│  └─ API Key: MEILISEARCH_API_KEY                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                           Qdrant                                 │
│                   (Vector Similarity Search)                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Operations:                                                     │
│  └─ Recommend (similar items by ID)                              │
│                                                                  │
│  Use Cases:                                                      │
│  ├─ Similar product recommendations                              │
│  ├─ "Customers also viewed"                                      │
│  └─ Product discovery                                            │
│                                                                  │
│  Features:                                                       │
│  ├─ Vector-based similarity                                      │
│  ├─ Fast approximate nearest neighbor                            │
│  └─ Positive/negative examples                                   │
│                                                                  │
│  Configuration:                                                  │
│  ├─ URL: QDRANT_URL                                              │
│  └─ API Key: QDRANT_API_KEY                                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                         Shopify                                  │
│                  (E-commerce Platform API)                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Operations:                                                     │
│  ├─ OAuth Flow                                                   │
│  │  ├─ Build install URL                                         │
│  │  ├─ Validate HMAC                                             │
│  │  └─ Exchange code for token                                   │
│  │                                                               │
│  └─ Webhooks                                                     │
│     ├─ Verify signature (HMAC-SHA256)                            │
│     ├─ products/create                                           │
│     ├─ products/update                                           │
│     └─ products/delete                                           │
│                                                                  │
│  Configuration:                                                  │
│  ├─ API Key: SHOPIFY_API_KEY                                     │
│  ├─ API Secret: SHOPIFY_API_SECRET                               │
│  ├─ App URL: SHOPIFY_APP_URL                                     │
│  ├─ Scopes: SHOPIFY_SCOPES                                       │
│  └─ Webhook Secret: SHOPIFY_WEBHOOK_SECRET                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────┐
│                          MongoDB                                 │
│                   (Primary Database)                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Collections:                                                    │
│  ├─ users          → Platform users                              │
│  ├─ clients        → SaaS tenants                                │
│  ├─ indexes        → Search indexes metadata                     │
│  ├─ stores         → Shopify merchants                           │
│  └─ sessions       → Shopify OAuth sessions                      │
│                                                                  │
│  Features:                                                       │
│  ├─ Document-based storage                                       │
│  ├─ Flexible schema                                              │
│  ├─ Indexes for fast queries                                     │
│  └─ Atomic operations                                            │
│                                                                  │
│  Configuration:                                                  │
│  ├─ URL: DATABASE_URL                                            │
│  └─ Max Conns: DATABASE_MAX_CONNS                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Service Communication Patterns

```
┌────────────────────────────────────────────────────────────────┐
│                  SERVICE COMMUNICATION FLOW                      │
└────────────────────────────────────────────────────────────────┘

User Request Flow (SaaS):
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────>│ Handler  │────>│ MongoDB  │     │Meilisrch │
│          │     │ (Search) │     │ (Get     │     │          │
│          │     │          │     │  Client) │     │          │
│          │     │          │<────│          │     │          │
│          │     │          │                      │          │
│          │     │          │─────────────────────>│          │
│          │     │          │     Search Query     │          │
│          │     │          │<─────────────────────│          │
│          │<────│          │     Results          │          │
└──────────┘     └──────────┘                      └──────────┘


Shopify Webhook Flow:
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│ Shopify  │────>│ Webhook  │────>│ MongoDB  │     │Meilisrch │
│          │     │ Handler  │     │ (Get     │     │          │
│          │     │          │     │  Store)  │     │          │
│          │     │          │<────│          │     │          │
│          │     │          │                      │          │
│          │     │          │─────────────────────>│          │
│          │     │          │   Index Document     │          │
│          │     │          │<─────────────────────│          │
│          │<────│          │     Task ID          │          │
└──────────┘     └──────────┘                      └──────────┘


Storefront Search Flow:
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│Storefront│────>│Storefront│────>│ MongoDB  │     │Meilisrch │     │  Qdrant  │
│          │     │ Handler  │     │ (Get     │     │          │     │          │
│          │     │          │     │  Store)  │     │          │     │          │
│          │     │          │<────│          │     │          │     │          │
│          │     │          │                      │          │     │          │
│          │     │          │─────────────────────>│          │     │          │
│          │     │          │  Search (text)       │          │     │          │
│          │     │          │<─────────────────────│          │     │          │
│          │     │          │                                        │          │
│          │     │          │───────────────────────────────────────>│          │
│          │     │          │         Similar (vector)               │          │
│          │     │          │<───────────────────────────────────────│          │
│          │<────│          │                                        │          │
└──────────┘     └──────────┘                                        └──────────┘
```

---

## Use Cases & Data Flows

### Use Case 1: SaaS Client Onboarding

```
┌────────────────────────────────────────────────────────────────┐
│              USE CASE: SaaS Client Onboarding                   │
└────────────────────────────────────────────────────────────────┘

Actors: Developer (User)
Goal: Set up search for their application

Steps:
1. Register User Account
   ├─ POST /api/v1/auth/register/user
   └─ Receive JWT token

2. Create Client (Tenant)
   ├─ POST /api/v1/auth/register/client
   ├─ Authenticated with JWT
   └─ Receive client_id

3. Generate API Key
   ├─ POST /api/v1/clients/:client_id/api-keys
   ├─ Save API key (shown only once!)
   └─ Use for programmatic access

4. Create Search Index
   ├─ POST /api/v1/clients/:client_id/indexes
   ├─ Body: {name: "products", primary_key: "id"}
   └─ Index created in Meilisearch

5. Configure Index Settings
   ├─ PATCH /api/v1/clients/:client_id/indexes/products/settings
   ├─ Body: {searchableAttributes, filterableAttributes, etc}
   └─ Settings applied

6. Index Documents
   ├─ POST /api/v1/clients/:client_id/indexes/products/documents
   ├─ Body: Document object
   └─ Repeat for all documents

7. Integrate Search
   ├─ POST /api/v1/clients/:client_id/indexes/products/search
   ├─ Use API key for authentication
   └─ Application is live!

Data Flow:
User → JWT → Client → API Key → Index → Documents → Search Ready
```

### Use Case 2: Shopify Merchant Installation

```
┌────────────────────────────────────────────────────────────────┐
│         USE CASE: Shopify Merchant Installation                 │
└────────────────────────────────────────────────────────────────┘

Actors: Shopify Merchant
Goal: Enable search on their storefront

Steps:
1. Initiate Installation
   ├─ Merchant clicks "Install App"
   ├─ Frontend: POST /api/auth/shopify/begin
   └─ Redirect to Shopify authorization

2. Grant Permissions
   ├─ Merchant approves scopes
   └─ Shopify redirects back with code

3. Complete Installation
   ├─ Backend: GET /api/auth/shopify/callback
   ├─ Verify HMAC signature
   ├─ Exchange code for access token
   ├─ Create Store record (encrypted token)
   ├─ Generate storefront API key
   └─ Create Meilisearch index

4. Initial Sync (Manual/Scheduled)
   ├─ Fetch products from Shopify API
   ├─ Index products in Meilisearch
   └─ Update sync_state

5. Configure Webhooks
   ├─ Register webhooks with Shopify
   │  ├─ products/create
   │  ├─ products/update
   │  └─ products/delete
   └─ Real-time sync enabled

6. Embed Search Widget
   ├─ Get storefront API key from dashboard
   ├─ Add search widget to theme
   └─ Search is live!

Data Flow:
Merchant → OAuth → Store → Index → Webhooks → Storefront Search
```

### Use Case 3: Storefront Product Search

```
┌────────────────────────────────────────────────────────────────┐
│           USE CASE: Storefront Product Search                   │
└────────────────────────────────────────────────────────────────┘

Actors: Shopify Store Customer
Goal: Find products quickly

Steps:
1. Customer Enters Query
   ├─ Types "red shoes" in search box
   └─ Frontend sends request

2. Search Request
   ├─ POST /api/v1/search
   ├─ Headers: X-Storefront-Key: <public_key>
   ├─ Body: {q: "red shoes", filter: "available = true", limit: 20}
   └─ Backend validates storefront key

3. Retrieve Store Config
   ├─ Find store by public API key
   └─ Get index UID

4. Execute Search
   ├─ Forward query to Meilisearch
   ├─ Search in store's index
   └─ Apply filters, sorting

5. Return Results
   ├─ Meilisearch returns hits
   ├─ Format response
   └─ Send to frontend

6. Display Results
   ├─ Show products with highlights
   └─ Customer finds what they need!

Performance:
├─ CORS enabled for storefront domains
├─ No authentication overhead
└─ Fast response times (<100ms)
```

### Use Case 4: Similar Product Recommendations

```
┌────────────────────────────────────────────────────────────────┐
│       USE CASE: Similar Product Recommendations                 │
└────────────────────────────────────────────────────────────────┘

Actors: Shopify Store Customer
Goal: Discover related products

Steps:
1. Customer Views Product
   ├─ Product page loads
   └─ Frontend requests similar products

2. Request Similar Products
   ├─ GET /api/v1/similar?id=123456789&limit=10
   ├─ Headers: X-Storefront-Key: <public_key>
   └─ Backend validates key

3. Retrieve Store Config
   ├─ Find store by public key
   └─ Get Qdrant collection name

4. Vector Search
   ├─ Send request to Qdrant
   ├─ Use product ID as positive example
   ├─ Find nearest neighbors in vector space
   └─ Return similar products by embedding

5. Return Recommendations
   ├─ Qdrant returns scored results
   ├─ Format with product details
   └─ Send to frontend

6. Display "You May Also Like"
   ├─ Show recommended products
   └─ Increase discoverability!

Benefits:
├─ Vector-based similarity (beyond keywords)
├─ Fast lookups
└─ Personalized experience
```

### Use Case 5: Real-time Product Sync

```
┌────────────────────────────────────────────────────────────────┐
│           USE CASE: Real-time Product Sync                      │
└────────────────────────────────────────────────────────────────┘

Actors: Shopify (System)
Goal: Keep search index in sync

Trigger: Product Updated in Shopify

Steps:
1. Product Changed in Shopify
   ├─ Merchant updates product price
   └─ Shopify triggers webhook

2. Webhook Delivery
   ├─ POST /webhooks/shopify/products/update
   ├─ Headers:
   │  ├─ X-Shopify-Hmac-Sha256: <signature>
   │  └─ X-Shopify-Shop-Domain: store.myshopify.com
   └─ Body: Full product data

3. Verify Signature
   ├─ Calculate HMAC of request body
   ├─ Compare with header signature
   └─ Reject if invalid

4. Find Store
   ├─ Query MongoDB by shop_domain
   └─ Get index UID

5. Enrich Document
   ├─ Add shop_domain
   ├─ Add store_id
   └─ Add document_type

6. Update Index
   ├─ POST to Meilisearch
   ├─ Upsert document
   └─ Return task ID

7. Confirmation
   ├─ Respond to Shopify (200 OK)
   └─ Product now searchable!

Latency:
├─ Webhook → Index: <500ms
└─ Near real-time updates
```

### Use Case 6: Multi-Index Management

```
┌────────────────────────────────────────────────────────────────┐
│          USE CASE: Multi-Index Management                       │
└────────────────────────────────────────────────────────────────┘

Actors: SaaS Developer
Goal: Manage multiple search indexes

Scenario: Movie Streaming Platform

Steps:
1. Create Movies Index
   ├─ POST /api/v1/clients/:id/indexes
   ├─ Body: {name: "movies", primary_key: "id"}
   └─ Index UID: my_app__movies

2. Configure Movies Settings
   ├─ PATCH /api/v1/clients/:id/indexes/movies/settings
   ├─ Body:
   │  {
   │    searchableAttributes: ["title", "description", "actors"],
   │    filterableAttributes: ["genre", "year", "rating"],
   │    sortableAttributes: ["year", "rating"]
   │  }
   └─ Applied

3. Create TV Shows Index
   ├─ POST /api/v1/clients/:id/indexes
   ├─ Body: {name: "tv_shows", primary_key: "id"}
   └─ Index UID: my_app__tv_shows

4. Configure TV Shows Settings
   ├─ Different attributes
   └─ Separate configuration

5. Index Content
   ├─ Movies → my_app__movies
   └─ TV Shows → my_app__tv_shows

6. Search Both Indexes
   ├─ Search movies: POST /clients/:id/indexes/movies/search
   └─ Search shows: POST /clients/:id/indexes/tv_shows/search

Benefits:
├─ Logical separation
├─ Independent configuration
└─ Scalable architecture
```

---

## Technology Stack

### Core Technologies

```
┌────────────────────────────────────────────────────────────────┐
│                       TECHNOLOGY STACK                          │
└────────────────────────────────────────────────────────────────┘

Backend Framework:
├─ Language: Go 1.23+
├─ Web Framework: Gin (HTTP router)
├─ Architecture: Clean layered architecture
└─ Pattern: Repository + Service + Handler

Database:
├─ Primary: MongoDB 7+
├─ Driver: go.mongodb.org/mongo-driver
├─ Schema: Document-based (flexible)
└─ Indexes: Auto-created on collections

Search Engine:
├─ Full-text: Meilisearch
├─ Vector: Qdrant
├─ SDK: meilisearch-go (official)
└─ Protocol: HTTP REST API

Authentication:
├─ JWT: golang-jwt/jwt (v4)
├─ Hashing: bcrypt (passwords), SHA-256 (API keys)
├─ Encryption: AES-256-GCM (tokens, sensitive data)
└─ State Tokens: JWT (short-lived)

HTTP Client:
├─ Standard: net/http
├─ Timeout: Configurable per service
└─ Connection Pooling: Enabled

Security:
├─ CORS: gin-contrib/cors
├─ HMAC: crypto/hmac (Shopify webhooks)
├─ Encryption: crypto/aes
└─ Random: crypto/rand

Development:
├─ Environment: Nix (reproducible dev shells)
├─ Task Runner: Just (Justfile)
├─ Config: godotenv (.env files)
└─ Linting: golangci-lint

Deployment:
├─ Container: Docker
├─ Platform: Elastic Beanstalk (AWS)
└─ Config: Dockerrun.aws.json
```

### Package Dependencies

```
Key Dependencies (from go.mod):

Web Framework:
├─ github.com/gin-gonic/gin v1.11.0
└─ github.com/gin-contrib/cors v1.7.6

Database:
└─ go.mongodb.org/mongo-driver v1.17.6

Search:
└─ github.com/meilisearch/meilisearch-go v0.34.2

Authentication:
├─ github.com/golang-jwt/jwt/v4 v4.5.2
└─ golang.org/x/crypto v0.40.0

Configuration:
└─ github.com/joho/godotenv v1.5.1
```

### Code Organization Principles

```
Architectural Patterns:

1. Layered Architecture
   ├─ Handlers: HTTP request/response
   ├─ Services: Business logic & external APIs
   ├─ Repositories: Data access
   └─ Models: Data structures

2. Dependency Injection
   ├─ Services injected into handlers
   ├─ Repositories injected into handlers
   └─ Config passed to all layers

3. Clean Separation
   ├─ SaaS logic separate from Shopify logic
   ├─ Authentication separate per use case
   └─ Models represent domain entities

4. Middleware Pattern
   ├─ Authentication as middleware
   ├─ CORS as middleware
   └─ Composable request processing

5. Repository Pattern
   ├─ Abstract database operations
   ├─ Single source of truth for queries
   └─ Testable without real database

6. Service Pattern
   ├─ Encapsulate external API calls
   ├─ Handle retries, errors
   └─ Provide clean interface to handlers
```

---

## Summary Statistics

```
┌────────────────────────────────────────────────────────────────┐
│                      PROJECT METRICS                            │
└────────────────────────────────────────────────────────────────┘

Data Models:           6 (User, Client, Index, APIKey, Store, Session)
API Endpoints:         40+
Authentication Types:  6 (JWT, API Key, Shopify JWT, Storefront Key, HMAC, Optional)
Handlers:             10 (UserAuth, Search, Storefront, Auth, Store, etc)
Services:              3 (Meilisearch, Qdrant, Shopify)
Repositories:          5 (User, Client, Index, Store, Session)
Middleware:            4 (JWT, APIKey, Auth, CORS)
External Services:     4 (MongoDB, Meilisearch, Qdrant, Shopify)

Architecture Pattern:  Clean Layered Architecture
Primary Language:      Go 1.23+
Web Framework:         Gin
Database:             MongoDB
Search Engines:        Meilisearch + Qdrant
```

---

## Quick Reference Links

### Key Documentation Files
- `README.md` - Project overview and setup
- `docs/API_REFERENCE.md` - Complete API documentation
- `docs/AUTH_API.md` - Authentication guide
- `docs/SIMILAR_PRODUCTS_API.md` - Vector search guide
- `docs/REMIX_QUICKSTART.md` - Shopify integration

### Important Handlers
- `handlers/user_auth.go` - SaaS user & client management
- `handlers/search.go` - SaaS search operations
- `handlers/storefront.go` - Shopify public search
- `handlers/auth.go` - Shopify OAuth flow
- `handlers/webhook.go` - Shopify webhook processing

### Core Services
- `services/meilisearch.go` - Search operations
- `services/qdrant.go` - Vector similarity
- `services/shopify.go` - Shopify API integration

### Authentication
- `middleware/jwt_middleware.go` - JWT validation
- `middleware/apikey_middleware.go` - API key validation
- `middleware/auth_middleware.go` - Shopify session validation
- `pkg/auth/jwt.go` - JWT generation/parsing

---

**Generated:** 2026-01-15
**Version:** Based on current main branch
