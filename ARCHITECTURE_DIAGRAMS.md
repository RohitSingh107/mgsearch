# MGSearch - Visual Architecture Diagrams

This document contains Mermaid diagrams for visualizing the MGSearch architecture. These diagrams can be rendered in GitHub, GitLab, or any Markdown viewer that supports Mermaid.

## Table of Contents
1. [System Architecture](#system-architecture)
2. [Data Model Relationships](#data-model-relationships)
3. [Authentication Flow - SaaS](#authentication-flow---saas)
4. [Authentication Flow - Shopify](#authentication-flow---shopify)
5. [Search Flow - SaaS](#search-flow---saas)
6. [Search Flow - Storefront](#search-flow---storefront)
7. [Webhook Processing Flow](#webhook-processing-flow)
8. [Component Dependencies](#component-dependencies)

---

## System Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        A1[Shopify Admin Dashboard]
        A2[Shopify Storefront]
        A3[SaaS Dashboard]
        A4[API Clients]
    end
    
    subgraph "API Gateway & Middleware"
        B1[CORS Middleware]
        B2[JWT Middleware]
        B3[API Key Middleware]
        B4[Auth Middleware]
    end
    
    subgraph "Handler Layer"
        C1[UserAuth Handler]
        C2[Search Handler]
        C3[Storefront Handler]
        C4[Auth Handler]
        C5[Store Handler]
        C6[Session Handler]
        C7[Webhook Handler]
        C8[Index Handler]
        C9[Settings Handler]
        C10[Tasks Handler]
    end
    
    subgraph "Service Layer"
        D1[Meilisearch Service]
        D2[Qdrant Service]
        D3[Shopify Service]
    end
    
    subgraph "Repository Layer"
        E1[User Repository]
        E2[Client Repository]
        E3[Store Repository]
        E4[Session Repository]
        E5[Index Repository]
    end
    
    subgraph "Data & External Services"
        F1[(MongoDB)]
        F2[Meilisearch Cloud]
        F3[Qdrant Cloud]
        F4[Shopify API]
    end
    
    A1 --> B2 --> C1
    A1 --> B4 --> C5
    A2 --> B1 --> C3
    A3 --> B2 --> C2
    A4 --> B3 --> C2
    
    C1 --> E1 & E2
    C2 --> E2 & D1
    C3 --> E3 & D1 & D2
    C4 --> E3 & D3 & D1
    C5 --> E3
    C6 --> E4 & E3
    C7 --> E3 & D1 & D3
    C8 --> E2 & E5 & D1
    C9 --> E2 & D1
    C10 --> D1
    
    D1 --> F2
    D2 --> F3
    D3 --> F4
    
    E1 & E2 & E3 & E4 & E5 --> F1
    
    style A1 fill:#e1f5ff
    style A2 fill:#e1f5ff
    style A3 fill:#e1f5ff
    style A4 fill:#e1f5ff
    style F1 fill:#ffe1e1
    style F2 fill:#ffe1e1
    style F3 fill:#ffe1e1
    style F4 fill:#ffe1e1
```

---

## Data Model Relationships

```mermaid
erDiagram
    User ||--o{ UserClient : "has many"
    Client ||--o{ UserClient : "has many"
    Client ||--|{ APIKey : "contains"
    Client ||--o{ Index : "owns"
    
    User {
        ObjectId _id PK
        string email UK
        string password_hash
        string first_name
        string last_name
        ObjectId[] client_ids
        bool is_active
        datetime created_at
        datetime updated_at
    }
    
    Client {
        ObjectId _id PK
        string name UK
        string description
        ObjectId[] user_ids
        APIKey[] api_keys
        bool is_active
        datetime created_at
        datetime updated_at
    }
    
    APIKey {
        ObjectId _id PK
        string key "SHA-256 hashed"
        string name
        string key_prefix
        string[] permissions
        bool is_active
        datetime last_used_at
        datetime created_at
        datetime expires_at
    }
    
    Index {
        ObjectId _id PK
        ObjectId client_id FK
        string name
        string uid "client_name__index_name"
        string primary_key
        datetime created_at
        datetime updated_at
    }
    
    Store {
        ObjectId _id PK
        string shop_domain UK
        string shop_name
        bytes encrypted_access_token
        string api_key_public
        string api_key_private
        string product_index_uid
        string meilisearch_index_uid
        string qdrant_collection_name
        string plan_level
        string status
        datetime installed_at
        map sync_state
    }
    
    Session {
        string _id PK
        string shop
        string state
        bool is_online
        string scope
        datetime expires
        string access_token "encrypted"
        int64 user_id
        datetime created_at
    }
    
    Store ||--o{ Session : "related by shop"
    
    UserClient {
        ObjectId user_id FK
        ObjectId client_id FK
    }
```

---

## Authentication Flow - SaaS

```mermaid
sequenceDiagram
    participant U as User
    participant FE as Frontend
    participant API as MGSearch API
    participant DB as MongoDB
    
    Note over U,DB: User Registration
    U->>FE: Enter credentials
    FE->>API: POST /api/v1/auth/register/user
    API->>API: Hash password (bcrypt)
    API->>DB: Save user
    DB-->>API: User created
    API->>API: Generate JWT (24h)
    API-->>FE: {user, token}
    FE-->>U: Registration success
    
    Note over U,DB: User Login
    U->>FE: Enter credentials
    FE->>API: POST /api/v1/auth/login
    API->>DB: Find user by email
    DB-->>API: User data
    API->>API: Verify password
    API->>API: Generate JWT (24h)
    API-->>FE: {user, token}
    FE-->>U: Login success
    
    Note over U,DB: Create Client
    U->>FE: Create new client
    FE->>API: POST /api/v1/auth/register/client<br/>(Authorization: Bearer <JWT>)
    API->>API: Verify JWT
    API->>API: Extract user_id
    API->>DB: Create client
    DB-->>API: Client created
    API->>DB: Link user to client
    API-->>FE: {client}
    FE-->>U: Client created
    
    Note over U,DB: Generate API Key
    U->>FE: Generate API key
    FE->>API: POST /api/v1/clients/:id/api-keys<br/>(Authorization: Bearer <JWT>)
    API->>API: Verify JWT & access
    API->>API: Generate random key (32B)
    API->>API: Hash key (SHA-256)
    API->>DB: Store hashed key
    DB-->>API: Key saved
    API-->>FE: {api_key: "raw_key_once"}
    FE-->>U: ⚠️ Save key now!
```

---

## Authentication Flow - Shopify

```mermaid
sequenceDiagram
    participant M as Merchant
    participant FE as Remix Frontend
    participant API as MGSearch API
    participant Shopify as Shopify
    participant DB as MongoDB
    participant MS as Meilisearch
    
    Note over M,MS: Installation Flow
    M->>FE: Click "Install App"
    FE->>API: POST /api/auth/shopify/begin
    API->>API: Generate state token (JWT, 15min)
    API-->>FE: {authUrl, state}
    FE->>Shopify: Redirect to authUrl
    Shopify->>M: Request permissions
    M->>Shopify: Approve
    
    Shopify->>API: Redirect to callback<br/>?code=xxx&state=yyy&hmac=zzz
    API->>API: Verify HMAC signature
    API->>API: Validate state token
    API->>Shopify: Exchange code for token
    Shopify-->>API: access_token
    
    API->>API: Encrypt access token (AES-256-GCM)
    API->>API: Generate API keys
    API->>DB: Create Store record
    DB-->>API: Store saved
    
    API->>MS: Create index
    MS-->>API: Index created
    
    API->>API: Generate session JWT (24h)
    API-->>FE: {store, token}
    FE-->>M: Installation complete!
```

---

## Search Flow - SaaS

```mermaid
sequenceDiagram
    participant Client as API Client
    participant API as MGSearch API
    participant DB as MongoDB
    participant MS as Meilisearch
    
    Note over Client,MS: Document Indexing
    Client->>API: POST /api/v1/clients/:id/indexes/:index/documents<br/>(Authorization: Bearer <api_key>)
    API->>API: Validate API key (SHA-256)
    API->>DB: Find client by hashed key
    DB-->>API: Client data
    API->>API: Check key expiration
    API->>API: Construct index UID<br/>(client_name__index_name)
    API->>MS: Index document
    MS-->>API: Task ID
    API-->>Client: {taskUid, status}
    
    Note over Client,MS: Search Query
    Client->>API: POST /api/v1/clients/:id/indexes/:index/search<br/>(Authorization: Bearer <api_key>)
    API->>API: Validate API key
    API->>DB: Find client
    DB-->>API: Client data
    API->>API: Construct index UID
    API->>MS: Search query<br/>{q, filter, sort, limit, ...}
    MS-->>API: Search results
    API-->>Client: {hits, query, ...}
    
    Note over Client,MS: Update Settings
    Client->>API: PATCH /api/v1/clients/:id/indexes/:index/settings<br/>(Authorization: Bearer <JWT>)
    API->>API: Verify JWT
    API->>DB: Verify client access
    API->>MS: Update settings
    MS-->>API: Task ID
    API-->>Client: {taskUid}
    
    Note over Client,MS: Check Task Status
    Client->>API: GET /api/v1/clients/:id/tasks/:task_id<br/>(Authorization: Bearer <api_key>)
    API->>API: Validate API key
    API->>MS: Get task details
    MS-->>API: Task status
    API-->>Client: {status, details, ...}
```

---

## Search Flow - Storefront

```mermaid
sequenceDiagram
    participant Customer as Customer
    participant SF as Storefront
    participant API as MGSearch API
    participant DB as MongoDB
    participant MS as Meilisearch
    participant Q as Qdrant
    
    Note over Customer,Q: Text Search
    Customer->>SF: Search "red shoes"
    SF->>API: POST /api/v1/search<br/>(X-Storefront-Key: <public_key>)
    API->>API: CORS validation
    API->>DB: Find store by public key
    DB-->>API: Store data
    API->>API: Get index UID
    API->>MS: Search query
    MS-->>API: Results
    API-->>SF: {hits, ...}
    SF-->>Customer: Display products
    
    Note over Customer,Q: Similar Products (Vector Search)
    Customer->>SF: View product #123456
    SF->>API: GET /api/v1/similar?id=123456&limit=10<br/>(X-Storefront-Key: <public_key>)
    API->>DB: Find store by key
    DB-->>API: Store data
    API->>API: Get Qdrant collection
    API->>Q: Recommend(positive: [123456])
    Q-->>API: Similar products
    API-->>SF: {result, ...}
    SF-->>Customer: "You may also like..."
```

---

## Webhook Processing Flow

```mermaid
sequenceDiagram
    participant Shopify as Shopify
    participant API as MGSearch Webhook Handler
    participant DB as MongoDB
    participant MS as Meilisearch
    
    Note over Shopify,MS: Product Created/Updated
    Shopify->>API: POST /webhooks/shopify/products/update<br/>Headers:<br/>- X-Shopify-Hmac-Sha256<br/>- X-Shopify-Shop-Domain
    API->>API: Verify HMAC signature
    alt Invalid signature
        API-->>Shopify: 401 Unauthorized
    end
    
    API->>DB: Find store by shop_domain
    DB-->>API: Store data
    alt Store not found
        API-->>Shopify: 404 Not Found
    end
    
    API->>API: Enrich document<br/>+ shop_domain<br/>+ store_id<br/>+ document_type
    API->>MS: Index/Update document
    MS-->>API: Task ID
    API-->>Shopify: 200 OK {status: "processed"}
    
    Note over Shopify,MS: Product Deleted
    Shopify->>API: POST /webhooks/shopify/products/delete
    API->>API: Verify HMAC
    API->>DB: Find store
    DB-->>API: Store data
    API->>API: Extract product ID
    API->>MS: Delete document(id)
    MS-->>API: Task ID
    API-->>Shopify: 200 OK
```

---

## Component Dependencies

```mermaid
graph TD
    subgraph "Main Application"
        Main[main.go]
    end
    
    subgraph "Configuration"
        Config[config/config.go]
    end
    
    subgraph "Handlers"
        H1[UserAuthHandler]
        H2[SearchHandler]
        H3[StorefrontHandler]
        H4[AuthHandler]
        H5[StoreHandler]
        H6[SessionHandler]
        H7[WebhookHandler]
        H8[IndexHandler]
        H9[SettingsHandler]
        H10[TasksHandler]
    end
    
    subgraph "Middleware"
        MW1[JWTMiddleware]
        MW2[APIKeyMiddleware]
        MW3[AuthMiddleware]
        MW4[CORSMiddleware]
    end
    
    subgraph "Services"
        S1[MeilisearchService]
        S2[QdrantService]
        S3[ShopifyService]
    end
    
    subgraph "Repositories"
        R1[UserRepository]
        R2[ClientRepository]
        R3[StoreRepository]
        R4[SessionRepository]
        R5[IndexRepository]
    end
    
    subgraph "Models"
        M1[User]
        M2[Client]
        M3[Store]
        M4[Session]
        M5[Index]
        M6[Search]
    end
    
    subgraph "Security"
        Sec1[JWT Utils]
        Sec2[Password Utils]
        Sec3[Encryption Utils]
        Sec4[Key Generation]
    end
    
    Main --> Config
    Main --> MW1 & MW2 & MW3 & MW4
    Main --> H1 & H2 & H3 & H4 & H5 & H6 & H7 & H8 & H9 & H10
    Main --> S1 & S2 & S3
    Main --> R1 & R2 & R3 & R4 & R5
    
    H1 --> R1 & R2
    H1 --> Sec1 & Sec2
    
    H2 --> R2 & S1
    
    H3 --> R3 & S1 & S2
    
    H4 --> R3 & S1 & S3
    H4 --> Sec3 & Sec4 & Sec1
    
    H5 --> R3
    
    H6 --> R3 & R4 & S1
    H6 --> Sec3
    
    H7 --> R3 & S1 & S3
    
    H8 --> R2 & R5 & S1
    
    H9 --> R2 & S1
    
    H10 --> S1
    
    MW1 --> Sec1
    MW2 --> R2
    MW3 --> Sec1
    
    R1 & R2 & R3 & R4 & R5 --> M1 & M2 & M3 & M4 & M5
    
    H2 & H3 & H9 & H10 --> M6
    
    style Main fill:#ffd700
    style Config fill:#98fb98
    style S1 fill:#ffb6c1
    style S2 fill:#ffb6c1
    style S3 fill:#ffb6c1
```

---

## API Endpoint Tree

```mermaid
graph TD
    Root[MGSearch API]
    
    Root --> Public[Public Endpoints]
    Root --> SaaS[SaaS Platform]
    Root --> Shopify[Shopify Platform]
    Root --> Webhooks[Webhooks]
    
    Public --> Ping[GET /ping]
    
    SaaS --> Auth[Authentication]
    SaaS --> Clients[Client Management]
    SaaS --> Search[Search Operations]
    
    Auth --> RegUser[POST /auth/register/user]
    Auth --> Login[POST /auth/login]
    Auth --> Me[GET /auth/me]
    Auth --> UpdateUser[PUT /auth/user]
    Auth --> RegClient[POST /auth/register/client]
    
    Clients --> ListClients[GET /clients]
    Clients --> GetClient[GET /clients/:id]
    Clients --> GenKey[POST /clients/:id/api-keys]
    Clients --> RevokeKey[DELETE /clients/:id/api-keys/:key]
    Clients --> CreateIndex[POST /clients/:id/indexes]
    Clients --> ListIndexes[GET /clients/:id/indexes]
    Clients --> IndexDoc[POST /clients/:id/indexes/:idx/documents]
    Clients --> UpdateSettings[PATCH /clients/:id/indexes/:idx/settings]
    
    Search --> ClientSearch[POST /clients/:id/indexes/:idx/search]
    Search --> TaskStatus[GET /clients/:id/tasks/:task_id]
    
    Shopify --> StorefrontSearch[Storefront Search]
    Shopify --> ShopifyAuth[OAuth Flow]
    Shopify --> StoreManagement[Store Management]
    Shopify --> Sessions[Session Management]
    
    StorefrontSearch --> SearchProducts[GET/POST /search]
    StorefrontSearch --> SimilarProducts[GET/POST /similar]
    
    ShopifyAuth --> BeginAuth[POST /auth/shopify/begin]
    ShopifyAuth --> Callback[GET /auth/shopify/callback]
    ShopifyAuth --> Exchange[POST /auth/shopify/exchange]
    ShopifyAuth --> Install[POST /auth/shopify/install]
    
    StoreManagement --> CurrentStore[GET /stores/current]
    StoreManagement --> SyncStatus[GET /stores/sync-status]
    
    Sessions --> StoreSession[POST /sessions]
    Sessions --> LoadSession[GET /sessions/:id]
    Sessions --> DeleteSession[DELETE /sessions/:id]
    Sessions --> BatchDelete[DELETE /sessions/batch]
    Sessions --> FindByShop[GET /sessions/shop/:shop]
    
    Webhooks --> ProductWebhook[POST /webhooks/shopify/:topic/:subtopic]
    
    style Root fill:#4169e1,color:#fff
    style Public fill:#90ee90
    style SaaS fill:#ffa07a
    style Shopify fill:#87ceeb
    style Webhooks fill:#dda0dd
```

---

## Security & Authentication Matrix

```mermaid
graph LR
    subgraph "Endpoint Types"
        E1[Public<br/>No Auth]
        E2[SaaS User<br/>JWT]
        E3[SaaS Client<br/>API Key]
        E4[Shopify Admin<br/>Session JWT]
        E5[Storefront<br/>Public Key]
        E6[Webhooks<br/>HMAC]
    end
    
    subgraph "Middleware"
        M1[CORSMiddleware]
        M2[JWTMiddleware]
        M3[APIKeyMiddleware]
        M4[AuthMiddleware]
    end
    
    subgraph "Protected Resources"
        R1[User Management]
        R2[Client Management]
        R3[Search API]
        R4[Store Management]
        R5[Product Search]
        R6[Product Sync]
    end
    
    E1 --> M1 --> R5
    E2 --> M2 --> R1 & R2
    E3 --> M3 --> R3
    E4 --> M4 --> R4
    E5 --> M1 --> R5
    E6 -.HMAC.-> R6
    
    style E1 fill:#90ee90
    style E2 fill:#ffd700
    style E3 fill:#ffa07a
    style E4 fill:#87ceeb
    style E5 fill:#98fb98
    style E6 fill:#dda0dd
```

---

## Data Flow - End to End

### SaaS Platform Flow

```mermaid
flowchart TD
    Start([Developer starts]) --> Register[Register User]
    Register --> Login[Login]
    Login --> CreateClient[Create Client/Tenant]
    CreateClient --> GenAPIKey[Generate API Key]
    GenAPIKey --> CreateIdx[Create Search Index]
    CreateIdx --> ConfigIdx[Configure Index Settings]
    ConfigIdx --> IndexDocs[Index Documents]
    IndexDocs --> Integrate[Integrate into App]
    Integrate --> AppSearch[App performs searches]
    AppSearch --> End([Production!])
    
    style Start fill:#90ee90
    style End fill:#90ee90
    style GenAPIKey fill:#ffd700
    style AppSearch fill:#87ceeb
```

### Shopify Platform Flow

```mermaid
flowchart TD
    Start([Merchant installs]) --> OAuth[Complete OAuth]
    OAuth --> StoreCreated[Store Record Created]
    StoreCreated --> IndexCreated[Meilisearch Index Created]
    IndexCreated --> InitialSync[Initial Product Sync]
    InitialSync --> Webhooks[Register Webhooks]
    Webhooks --> Live[App is Live]
    
    Live --> WebhookEvent[Product Updated]
    WebhookEvent --> VerifyHMAC{Verify HMAC}
    VerifyHMAC -->|Valid| UpdateIndex[Update Search Index]
    VerifyHMAC -->|Invalid| Reject[Reject Request]
    UpdateIndex --> SearchUpdated[Search Updated]
    
    Live --> Customer[Customer Searches]
    Customer --> StorefrontAPI[Storefront API Call]
    StorefrontAPI --> ValidateKey{Validate Key}
    ValidateKey -->|Valid| Search[Execute Search]
    ValidateKey -->|Invalid| Unauthorized[401 Unauthorized]
    Search --> Results[Return Results]
    
    style Start fill:#90ee90
    style Live fill:#ffd700
    style Results fill:#87ceeb
    style SearchUpdated fill:#87ceeb
```

---

## Deployment Architecture

```mermaid
graph TB
    subgraph "Internet"
        Clients[API Clients]
        Shopify[Shopify]
    end
    
    subgraph "AWS Elastic Beanstalk"
        LB[Load Balancer]
        subgraph "EC2 Instances"
            App1[MGSearch Instance 1]
            App2[MGSearch Instance 2]
        end
    end
    
    subgraph "Managed Services"
        Mongo[(MongoDB Atlas)]
        Meili[Meilisearch Cloud]
        Q[Qdrant Cloud]
    end
    
    Clients --> LB
    Shopify --> LB
    
    LB --> App1 & App2
    
    App1 & App2 --> Mongo
    App1 & App2 --> Meili
    App1 & App2 --> Q
    App1 & App2 --> Shopify
    
    style LB fill:#ffd700
    style App1 fill:#87ceeb
    style App2 fill:#87ceeb
    style Mongo fill:#90ee90
    style Meili fill:#ffa07a
    style Q fill:#ffb6c1
```

---

**Note:** These diagrams are created using Mermaid syntax and can be rendered in:
- GitHub
- GitLab
- VS Code (with Mermaid extension)
- Markdown preview tools
- Mermaid Live Editor (https://mermaid.live)

**Generated:** 2026-01-15
**Version:** Current main branch
