# ByteRoom: High-Level Design (HLD)

## 1. Executive Summary

ByteRoom is a real-time chat platform optimized for technical discussions. This document outlines the high-level architecture for Phase 1 (MVP targeting 100 DAU), with considerations for future scale to 1M DAU.

### Design Goals

| Goal | Target | Rationale |
|------|--------|-----------|
| Latency | < 500ms end-to-end | Real-time conversation feel |
| Durability | Zero message loss | Trust and reliability |
| Availability | 99.9% uptime | Always-on communication |
| Security | XSS-safe content | Code snippets are attack vectors |

## 2. System Context

```mermaid
C4Context
    title ByteRoom System Context Diagram

    Person(user, "CS Geek", "Discusses DSA, system design, and code")
    
    System(byteroom, "ByteRoom", "Real-time chat for technical discussions")
    
    System_Ext(s3, "AWS S3", "Media storage")
    System_Ext(auth, "Auth Provider", "User authentication")
    
    Rel(user, byteroom, "Sends messages, views diagrams", "HTTPS, WSS")
    Rel(byteroom, s3, "Stores/retrieves media", "HTTPS")
    Rel(byteroom, auth, "Validates users", "HTTPS")
```

## 3. High-Level Architecture

### 3.1 Phase 1 Architecture (Monolithic)

```mermaid
graph TB
    subgraph "Client Layer"
        Web[React SPA]
        Mobile[Future Mobile App]
    end

    subgraph "API Gateway Layer"
        LB[Load Balancer / Nginx]
    end

    subgraph "Application Layer"
        Mono[Go Monolith]
        
        subgraph "Mono Components"
            REST[REST API Handler]
            WSH[WebSocket Hub]
            MsgSvc[Message Service]
            ChatSvc[Chat Service]
            UserSvc[User Service]
            AuthMW[Auth Middleware]
            SanMW[Sanitizer]
        end
    end

    subgraph "Data Layer"
        PG[(PostgreSQL)]
        S3[(AWS S3)]
    end

    Web -->|HTTPS| LB
    Mobile -->|HTTPS| LB
    LB -->|HTTP/WS| Mono
    
    REST --> AuthMW
    WSH --> AuthMW
    AuthMW --> MsgSvc
    AuthMW --> ChatSvc
    AuthMW --> UserSvc
    MsgSvc --> SanMW
    MsgSvc --> PG
    ChatSvc --> PG
    UserSvc --> PG
    REST --> S3
```

### 3.2 Component Responsibilities

| Component | Responsibility |
|-----------|---------------|
| **Load Balancer** | SSL termination, request routing, health checks |
| **REST API Handler** | HTTP endpoints for history, uploads, auth |
| **WebSocket Hub** | Connection management, message routing, broadcasts |
| **Message Service** | Message validation, persistence, idempotency |
| **Chat Service** | Room management, member operations |
| **User Service** | User CRUD, profile management |
| **Auth Middleware** | JWT validation, session management |
| **Sanitizer** | XSS prevention using bluemonday |

## 4. Core Workflows

### 4.1 Message Delivery Flow

```mermaid
sequenceDiagram
    autonumber
    participant Sender as Sender (Client)
    participant Hub as WebSocket Hub
    participant Svc as Message Service
    participant San as Sanitizer
    participant DB as PostgreSQL
    participant Recipients as Recipients (Clients)

    Sender->>Hub: WS: message.send
    Hub->>Svc: Process(message)
    
    alt Message ID exists (retry)
        Svc->>DB: Check message_id
        DB-->>Svc: Exists
        Svc-->>Hub: Return existing ACK
    else New message
        Svc->>San: Sanitize content
        San-->>Svc: Clean content
        Svc->>DB: INSERT message
        DB-->>Svc: Commit success
    end
    
    Hub-->>Sender: WS: message.ack (status: true)
    
    par Broadcast to online recipients
        Hub->>Recipients: WS: message.new
    end
```

### 4.2 User Connection Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant LB as Load Balancer
    participant Hub as WebSocket Hub
    participant Auth as Auth Middleware
    participant DB as PostgreSQL

    Client->>LB: WSS Upgrade Request + JWT
    LB->>Hub: Forward WS Request
    Hub->>Auth: Validate JWT
    
    alt Valid Token
        Auth-->>Hub: User ID
        Hub->>Hub: Register connection
        Hub-->>Client: WS Connected
        Hub->>DB: Update user online status
    else Invalid Token
        Auth-->>Hub: Error
        Hub-->>Client: WS Close (401)
    end

    loop Heartbeat
        Hub->>Client: WS: ping
        Client-->>Hub: WS: pong
    end
```

### 4.3 Media Upload Flow

```mermaid
sequenceDiagram
    autonumber
    participant Client
    participant API as REST API
    participant S3
    participant Hub as WebSocket Hub
    participant DB as PostgreSQL

    Client->>API: POST /api/upload/request {filename, mime_type}
    API->>S3: Generate pre-signed URL
    S3-->>API: Pre-signed URL (15 min expiry)
    API-->>Client: {upload_url, file_key}
    
    Client->>S3: PUT file binary
    S3-->>Client: 200 OK
    
    Client->>Hub: WS: message.send {content_type: "image", content: file_key}
    Hub->>DB: INSERT message with media reference
    Hub-->>Client: WS: message.ack
```

## 5. Data Architecture

### 5.1 Entity Relationship Diagram

```mermaid
erDiagram
    USERS ||--o{ CHAT_MEMBERS : "joins"
    USERS ||--o{ MESSAGES : "sends"
    CHATS ||--o{ CHAT_MEMBERS : "has"
    CHATS ||--o{ MESSAGES : "contains"

    USERS {
        uuid id PK
        string username UK
        string email UK
        string password_hash
        string display_name
        string avatar_url
        timestamp created_at
        timestamp updated_at
    }

    CHATS {
        uuid id PK
        string name
        enum type "direct|group"
        uuid created_by FK
        timestamp created_at
    }

    CHAT_MEMBERS {
        uuid chat_id PK,FK
        uuid user_id PK,FK
        enum role "admin|member"
        timestamp joined_at
    }

    MESSAGES {
        uuid id PK
        uuid chat_id FK
        uuid sender_id FK
        enum content_type "markdown|diagram_state|image"
        text content
        timestamp created_at
    }
```

### 5.2 Data Flow Summary

| Data Type | Storage | Access Pattern |
|-----------|---------|----------------|
| User profiles | PostgreSQL | Read-heavy, cached |
| Chat metadata | PostgreSQL | Read-heavy |
| Messages | PostgreSQL | Append-only, time-ordered reads |
| Images/Diagrams | S3 | Write-once, CDN-served reads |
| WebSocket state | In-memory (Hub) | Ephemeral, per-connection |

## 6. Technology Choices

### 6.1 Backend

| Technology | Purpose | Justification |
|------------|---------|---------------|
| **Go 1.22+** | Application server | Excellent concurrency (goroutines), strong stdlib, easy deployment |
| **gorilla/websocket** | WebSocket handling | Battle-tested, full RFC 6455 compliance |
| **PostgreSQL 15** | Primary database | ACID compliance, JSON support, rich indexing |
| **bluemonday** | HTML sanitization | Configurable whitelist-based sanitizer for XSS prevention |

### 6.2 Frontend

| Technology | Purpose | Justification |
|------------|---------|---------------|
| **React 18** | UI framework | Component model, hooks, concurrent features |
| **TypeScript** | Type safety | Catch errors early, better IDE support |
| **Vite** | Build tool | Fast HMR, optimized production builds |
| **Tailwind CSS** | Styling | Utility-first, great dark mode support |
| **Zustand** | State management | Lightweight, TypeScript-first |
| **react-markdown** | Markdown rendering | Safe rendering, plugin ecosystem |
| **react-syntax-highlighter** | Code highlighting | 180+ language support |
| **mermaid** | Diagram rendering | Industry standard for text-to-diagram |
| **@excalidraw/excalidraw** | Interactive diagrams | Collaborative whiteboard |

### 6.3 Infrastructure

| Technology | Purpose | Justification |
|------------|---------|---------------|
| **AWS S3** | Object storage | Scalable, pre-signed URLs for secure uploads |
| **Nginx** | Reverse proxy | WebSocket support, SSL termination |
| **Docker** | Containerization | Consistent environments |

## 7. Scalability Considerations

### 7.1 Phase 1 Limits (100 DAU)

| Metric | Estimate | Within Single Instance |
|--------|----------|----------------------|
| Concurrent WebSockets | ~50 | ✅ Yes |
| Messages/second | ~10 | ✅ Yes |
| Database connections | ~20 | ✅ Yes |
| Storage (monthly) | ~5 GB | ✅ Yes |

### 7.2 Phase 2 Evolution Path

```mermaid
graph LR
    subgraph "Phase 1"
        M1[Monolith]
    end

    subgraph "Phase 2"
        E1[Edge Server 1]
        E2[Edge Server 2]
        K[Kafka]
        W[Workers]
        R[Redis]
        C[Cassandra]
    end

    M1 -->|Extract WebSocket| E1
    M1 -->|Extract WebSocket| E2
    M1 -->|Add event bus| K
    M1 -->|Add routing| R
    M1 -->|Migrate data| C
```

**Key extraction points:**
1. WebSocket Hub → Stateless Edge Servers
2. Message persistence → Kafka (durability) + Cassandra (storage)
3. User routing → Redis (user_id → server mapping)
4. Search → Elasticsearch

## 8. Security Architecture

### 8.1 Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant DB

    Client->>API: POST /auth/login {email, password}
    API->>DB: Verify credentials
    DB-->>API: User record
    API->>API: Generate JWT (24h expiry)
    API-->>Client: {token, refresh_token}
    
    Note over Client,API: Subsequent requests
    Client->>API: Request + Authorization: Bearer {token}
    API->>API: Validate JWT signature
    API-->>Client: Response
```

### 8.2 Security Measures

| Threat | Mitigation |
|--------|------------|
| XSS via code blocks | Server-side sanitization with bluemonday |
| Message spoofing | JWT-based authentication, server-assigned sender_id |
| Replay attacks | Message ID idempotency check |
| Unauthorized access | Chat membership validation on all operations |
| Data in transit | TLS/HTTPS for all communications |

## 9. Monitoring & Observability

### 9.1 Key Metrics

| Category | Metrics |
|----------|---------|
| **Latency** | Message delivery time (p50, p95, p99) |
| **Throughput** | Messages/second, WebSocket connections |
| **Errors** | Failed deliveries, sanitization rejections |
| **Saturation** | DB connection pool, memory usage |

### 9.2 Logging Strategy

```
Level   | When to use
--------|------------------------------------------
ERROR   | Failed message persistence, auth failures
WARN    | Reconnection attempts, sanitization blocks
INFO    | User connections, message deliveries
DEBUG   | WebSocket frame details (dev only)
```

## 10. Deployment Architecture

```mermaid
graph TB
    subgraph "Production"
        DNS[DNS / CDN]
        LB[Load Balancer]
        App1[Go App Instance]
        PG[(PostgreSQL Primary)]
        S3[(S3 Bucket)]
    end

    subgraph "Development"
        DevApp[Local Go App]
        DevDB[(Local PostgreSQL)]
        LocalS3[LocalStack S3]
    end

    DNS --> LB
    LB --> App1
    App1 --> PG
    App1 --> S3
```

## 11. Future Considerations (Phase 2)

- **Horizontal scaling**: Stateless edge servers with Redis routing
- **Message durability**: Kafka as ingestion buffer before DB writes
- **Search**: Elasticsearch for code snippet and message search
- **Mobile**: React Native or native iOS/Android apps
- **E2E Encryption**: Optional encrypted rooms for sensitive discussions
