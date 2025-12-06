# Blytz Platform Service Architecture Diagram

## 🏗️ Overall Architecture

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Frontend<br/>Next.js]
        MOBILE[Mobile App<br/>Flutter]
        ADMIN[Admin Panel]
    end

    subgraph "API Gateway (Port 8080)"
        GATEWAY[Gateway Service<br/>Rate Limiting<br/>CORS<br/>Authentication Proxy]
    end

    subgraph "Core Services"
        AUTH[Auth Service<br/>Port 8084<br/>Better Auth + JWT]
        PRODUCT[Product Service<br/>Port 8082<br/>Product Catalog]
        AUCTION[Auction Service<br/>Port 8083<br/>Redis + PostgreSQL]
        ORDER[Order Service<br/>Port 8085<br/>Cart & Orders]
        PAYMENT[Payment Service<br/>Port 8086<br/>Fiuu Gateway]
        CHAT[Chat Service<br/>Port 8088<br/>Real-time Messaging]
        LOGISTICS[Logistics Service<br/>Port 8087<br/>Ninjavan Shipping]
        LIVEKIT[LiveKit Service<br/>WebRTC Streaming]
    end

    subgraph "Data Layer"
        POSTGRES[(PostgreSQL<br/>8 Databases)]
        REDIS[(Redis<br/>Cache & Real-time)]
        FIREBASE[Firebase<br/>Cloud Functions]
    end

    subgraph "External Services"
        LIVEKIT_CLOUD[LiveKit Cloud<br/>Video Streaming]
        FIUU[Fiuu Payment<br/>Payment Gateway]
        NINJAVAN[Ninjavan<br/>Shipping Service]
    end

    %% Client Connections
    WEB --> GATEWAY
    MOBILE --> GATEWAY
    ADMIN --> GATEWAY

    %% Gateway Routing
    GATEWAY --> AUTH
    GATEWAY --> PRODUCT
    GATEWAY --> AUCTION
    GATEWAY --> ORDER
    GATEWAY --> PAYMENT
    GATEWAY --> CHAT
    GATEWAY --> LOGISTICS

    %% Service-to-Service Dependencies
    AUTH -.-> |JWT Validation| AUCTION
    AUTH -.-> |JWT Validation| ORDER
    AUTH -.-> |JWT Validation| PAYMENT
    AUTH -.-> |JWT Validation| CHAT
    AUTH -.-> |JWT Validation| LOGISTICS

    %% Auction Service Special Connections
    AUCTION --> LIVEKIT
    AUCTION --> FIREBASE
    AUCTION --> REDIS

    %% Payment Connections
    PAYMENT --> FIUU

    %% Logistics Connections
    LOGISTICS --> NINJAVAN

    %% LiveKit Connection
    LIVEKIT --> LIVEKIT_CLOUD

    %% Database Connections
    AUTH --> POSTGRES
    PRODUCT --> POSTGRES
    AUCTION --> POSTGRES
    ORDER --> POSTGRES
    PAYMENT --> POSTGRES
    CHAT --> POSTGRES
    LOGISTICS --> POSTGRES

    %% Redis for Auctions
    AUCTION --> REDIS
```

## 🔄 Service Interaction Flow

```mermaid
sequenceDiagram
    participant U as User
    participant W as Web Frontend
    participant G as Gateway
    participant A as Auth Service
    participant P as Product Service
    participant AU as Auction Service
    participant O as Order Service
    participant PA as Payment Service
    participant R as Redis
    participant DB as PostgreSQL

    User->>W: Browse Products
    W->>G: GET /api/v1/products
    G->>P: Proxy to Product Service
    P->>DB: Fetch Products
    DB-->>P: Product Data
    P-->>G: Products Response
    G-->>W: Products
    W-->>User: Display Products

    User->>W: Login
    W->>G: POST /api/v1/auth/login
    G->>A: Proxy to Auth Service
    A->>DB: Validate Credentials
    DB-->>A: User Data
    A-->>G: JWT Token
    G-->>W: Auth Response
    W-->>User: Login Success

    User->>W: Place Bid
    W->>G: POST /api/v1/auctions/{id}/bid (JWT)
    G->>A: Validate JWT
    A-->>G: Valid Token
    G->>AU: Proxy Bid Request
    AU->>R: Check Current Bid (Redis)
    R-->>AU: Current Bid Info
    AU->>R: Update Bid (Atomic)
    AU->>DB: Persist Bid
    AU-->>G: Bid Success
    G-->>W: Bid Response
    W-->>User: Bid Placed

    User->>W: Checkout
    W->>G: POST /api/v1/orders (JWT)
    G->>O: Create Order
    O->>DB: Save Order
    O-->>G: Order Created
    G->>PA: Process Payment
    PA->>PA: Call Fiuu Gateway
    PA-->>G: Payment Success
    G-->>W: Order Complete
    W-->>User: Order Confirmed
```

## 🗄️ Database Architecture

```mermaid
erDiagram
    auth_db ||--o{ users : contains
    auth_db ||--o{ sessions : contains
    
    products_db ||--o{ products : contains
    products_db ||--o{ categories : contains
    
    auction_db ||--o{ auctions : contains
    auction_db ||--o{ bids : contains
    auction_db ||--o{ auction_products : contains
    
    orders_db ||--o{ orders : contains
    orders_db ||--o{ order_items : contains
    orders_db ||--o{ carts : contains
    
    payments_db ||--o{ payments : contains
    payments_db ||--o{ transactions : contains
    
    chat_db ||--o{ chat_rooms : contains
    chat_db ||--o{ messages : contains
    
    logistics_db ||--o{ shipments : contains
    logistics_db ||--o{ tracking_events : contains

    users ||--o{ auctions : creates
    users ||--o{ bids : places
    users ||--o{ orders : makes
    users ||--o{ messages : sends
```

## 🚀 API Routing Structure

```
Frontend → Gateway (8080)
         ├── /api/v1/auth/* → Auth Service (8084)
         ├── /api/v1/products/* → Product Service (8082)
         ├── /api/v1/auctions/* → Auction Service (8083)
         ├── /api/v1/orders/* → Order Service (8085)
         ├── /api/v1/payments/* → Payment Service (8086)
         ├── /api/v1/chat/* → Chat Service (8088)
         ├── /api/v1/logistics/* → Logistics Service (8087)
         └── /api/public/* → Mock data & LiveKit tokens
```

## 🔐 Authentication Flow

```mermaid
graph LR
    USER[User] --> LOGIN[Login Request]
    LOGIN --> AUTH[Auth Service]
    AUTH --> DB[(PostgreSQL)]
    AUTH --> JWT[JWT Token]
    JWT --> STORAGE[Store in Redis]
    
    subgraph "Protected API Calls"
        API_CALL[API Request] --> GATEWAY[Gateway]
        GATEWAY --> VALIDATE[Validate JWT]
        VALIDATE --> AUTH_CHECK[Check Auth Service]
        AUTH_CHECK --> SERVICE[Target Service]
    end
    
    JWT --> API_CALL
```

## 🎯 Key Service Dependencies

### Auth Service (Port 8084)
- **Independent**: No dependencies on other services
- **Provides**: JWT validation for all other services
- **Database**: `auth_db`

### Product Service (Port 8082)
- **Independent**: No dependencies on other services
- **Database**: `products_db`

### Auction Service (Port 8083)
- **Depends on**: Auth Service (user validation)
- **External**: Firebase (notifications), LiveKit (streaming)
- **Database**: `auction_db` + Redis (real-time bidding)

### Order Service (Port 8085)
- **Depends on**: Auth Service, Product Service, Auction Service
- **Database**: `orders_db`

### Payment Service (Port 8086)
- **Depends on**: Auth Service, Order Service
- **External**: Fiuu Payment Gateway
- **Database**: `payments_db`

### Chat Service (Port 8088)
- **Depends on**: Auth Service
- **Database**: `chat_db`

### Logistics Service (Port 8087)
- **Depends on**: Auth Service, Order Service
- **External**: Ninjavan Shipping API
- **Database**: `logistics_db`

### LiveKit Service
- **Depends on**: Auth Service
- **External**: LiveKit Cloud
- **Purpose**: WebRTC video streaming for live auctions

## 📊 Service Health Monitoring

All services expose `/health` endpoints that the Gateway can monitor:

```
Gateway Health Check → Individual Service Health Checks
                    ├── Auth Service Health
                    ├── Product Service Health
                    ├── Auction Service Health
                    ├── Order Service Health
                    ├── Payment Service Health
                    ├── Chat Service Health
                    ├── Logistics Service Health
                    └── Database Connectivity (PostgreSQL + Redis)
```

This architecture ensures:
- **Scalability**: Each service can scale independently
- **Reliability**: Failure in one service doesn't cascade
- **Security**: Centralized authentication through Auth Service
- **Performance**: Redis caching for real-time operations
- **Maintainability**: Clear separation of concerns