# Phase 5: Integration & Polish

## Objective

Complete the MVP with end-to-end testing, production deployment configuration, performance optimization, documentation, and final polish.

## Duration Estimate

4 development days

## Prerequisites

- Phase 3 completed (WebSocket working)
- Phase 4 completed (Frontend functional)
- All components individually tested
- Backend and frontend can communicate

---

## Tasks

### Task 5.1: End-to-End Test Suite

**Description**: Implement comprehensive E2E tests using Playwright for critical user journeys.

**TDD Approach**:
```typescript
// e2e/auth.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('user can register a new account', async ({ page }) => {
    await page.goto('/register');
    
    await page.fill('[name="username"]', 'newuser');
    await page.fill('[name="email"]', 'newuser@example.com');
    await page.fill('[name="password"]', 'SecurePass123!');
    await page.fill('[name="confirmPassword"]', 'SecurePass123!');
    await page.fill('[name="displayName"]', 'New User');
    
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL('/');
    await expect(page.locator('[data-testid="user-menu"]')).toContainText('New User');
  });

  test('user can login with valid credentials', async ({ page }) => {
    // Assume user exists from seed data
    await page.goto('/login');
    
    await page.fill('[name="email"]', 'alice@example.com');
    await page.fill('[name="password"]', 'password123');
    
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL('/');
    await expect(page.locator('[data-testid="chat-sidebar"]')).toBeVisible();
  });

  test('user sees error on invalid login', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('[name="email"]', 'alice@example.com');
    await page.fill('[name="password"]', 'wrongpassword');
    
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[role="alert"]')).toContainText('Invalid');
  });

  test('user can logout', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('[name="email"]', 'alice@example.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL('/');
    
    // Logout
    await page.click('[data-testid="user-menu"]');
    await page.click('text=Logout');
    
    await expect(page).toHaveURL('/login');
  });
});

// e2e/chat.spec.ts
test.describe('Chat', () => {
  test.beforeEach(async ({ page }) => {
    // Login as alice
    await page.goto('/login');
    await page.fill('[name="email"]', 'alice@example.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await expect(page).toHaveURL('/');
  });

  test('user can view chat list', async ({ page }) => {
    await expect(page.locator('[data-testid="chat-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="chat-item"]').first()).toBeVisible();
  });

  test('user can select a chat', async ({ page }) => {
    await page.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    
    await expect(page.locator('[data-testid="chat-header"]')).toContainText('Tech Discussion');
    await expect(page.locator('[data-testid="message-list"]')).toBeVisible();
  });

  test('user can send a message', async ({ page }) => {
    await page.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    
    await page.fill('[data-testid="message-input"]', 'Hello from E2E test!');
    await page.click('[data-testid="send-button"]');
    
    await expect(page.locator('[data-testid="message-bubble"]').last())
      .toContainText('Hello from E2E test!');
  });

  test('user receives message in real-time', async ({ page, browser }) => {
    // Open second browser context as Bob
    const bobContext = await browser.newContext();
    const bobPage = await bobContext.newPage();
    
    await bobPage.goto('/login');
    await bobPage.fill('[name="email"]', 'bob@example.com');
    await bobPage.fill('[name="password"]', 'password123');
    await bobPage.click('button[type="submit"]');
    
    // Both users open same chat
    await page.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    await bobPage.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    
    // Alice sends message
    await page.fill('[data-testid="message-input"]', 'Hello Bob!');
    await page.click('[data-testid="send-button"]');
    
    // Bob receives it
    await expect(bobPage.locator('[data-testid="message-bubble"]').last())
      .toContainText('Hello Bob!');
    
    await bobContext.close();
  });

  test('user sees typing indicator', async ({ page, browser }) => {
    const bobContext = await browser.newContext();
    const bobPage = await bobContext.newPage();
    
    await bobPage.goto('/login');
    await bobPage.fill('[name="email"]', 'bob@example.com');
    await bobPage.fill('[name="password"]', 'password123');
    await bobPage.click('button[type="submit"]');
    
    await page.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    await bobPage.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
    
    // Bob starts typing
    await bobPage.fill('[data-testid="message-input"]', 'I am typing...');
    
    // Alice sees typing indicator
    await expect(page.locator('[data-testid="typing-indicator"]'))
      .toContainText('Bob is typing');
    
    await bobContext.close();
  });
});

// e2e/markdown.spec.ts
test.describe('Markdown & Code', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('[name="email"]', 'alice@example.com');
    await page.fill('[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.click('[data-testid="chat-item"]:has-text("Tech Discussion")');
  });

  test('code blocks render with syntax highlighting', async ({ page }) => {
    const codeMessage = '```javascript\nconst x = 42;\nconsole.log(x);\n```';
    
    await page.fill('[data-testid="message-input"]', codeMessage);
    await page.click('[data-testid="send-button"]');
    
    const codeBlock = page.locator('[data-testid="message-bubble"]').last().locator('pre code');
    await expect(codeBlock).toBeVisible();
    await expect(codeBlock).toHaveClass(/language-javascript/);
  });

  test('mermaid diagrams render as SVG', async ({ page }) => {
    const mermaidMessage = '```mermaid\ngraph TD\n  A[Start] --> B[End]\n```';
    
    await page.fill('[data-testid="message-input"]', mermaidMessage);
    await page.click('[data-testid="send-button"]');
    
    await expect(page.locator('[data-testid="message-bubble"]').last().locator('svg')).toBeVisible();
  });
});
```

**Subtasks**:
- [ ] Set up Playwright with configuration
- [ ] Create test fixtures and seed data
- [ ] Write authentication E2E tests
- [ ] Write chat flow E2E tests
- [ ] Write real-time messaging E2E tests
- [ ] Write markdown/code rendering tests
- [ ] Add visual regression tests
- [ ] Configure CI to run E2E tests

**Exit Criteria**:
- [ ] All E2E tests pass locally
- [ ] E2E tests run in CI pipeline
- [ ] Critical user journeys covered
- [ ] Multi-user scenarios tested

---

### Task 5.2: Performance Testing & Optimization

**Description**: Benchmark and optimize performance for target metrics.

**TDD Approach**:
```go
// internal/websocket/hub_benchmark_test.go
func BenchmarkHub_BroadcastToChat(b *testing.B) {
    hub := NewHub()
    go hub.Run()
    
    // Add 100 clients to a chat
    for i := 0; i < 100; i++ {
        client := &Client{
            userID:  fmt.Sprintf("user-%d", i),
            chatIDs: []string{"chat-1"},
            send:    make(chan []byte, 256),
        }
        hub.Register(client)
    }
    time.Sleep(50 * time.Millisecond)
    
    msg := []byte(`{"event":"message.new","data":{"content":"test"}}`)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        hub.BroadcastToChat("chat-1", msg, nil)
    }
}

func BenchmarkHub_ConcurrentBroadcasts(b *testing.B) {
    hub := NewHub()
    go hub.Run()
    
    // 10 chats, 50 clients each
    for chat := 0; chat < 10; chat++ {
        chatID := fmt.Sprintf("chat-%d", chat)
        for i := 0; i < 50; i++ {
            client := &Client{
                userID:  fmt.Sprintf("user-%d-%d", chat, i),
                chatIDs: []string{chatID},
                send:    make(chan []byte, 256),
            }
            hub.Register(client)
        }
    }
    time.Sleep(100 * time.Millisecond)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        msg := []byte(`{"event":"message.new","data":{}}`)
        i := 0
        for pb.Next() {
            chatID := fmt.Sprintf("chat-%d", i%10)
            hub.BroadcastToChat(chatID, msg, nil)
            i++
        }
    })
}

// internal/domain/message/repository_benchmark_test.go
func BenchmarkMessageRepository_Save(b *testing.B) {
    db := setupBenchDB(b)
    repo := postgres.NewMessageRepository(db)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        msg := &Message{
            ID:          uuid.New().String(),
            ChatID:      "bench-chat",
            SenderID:    "bench-user",
            ContentType: "markdown",
            Content:     "Benchmark message content",
        }
        repo.Save(ctx, msg)
    }
}

func BenchmarkMessageRepository_FindByChatID(b *testing.B) {
    db := setupBenchDB(b)
    repo := postgres.NewMessageRepository(db)
    ctx := context.Background()
    
    // Seed 10000 messages
    for i := 0; i < 10000; i++ {
        repo.Save(ctx, &Message{
            ID:      fmt.Sprintf("msg-%d", i),
            ChatID:  "bench-chat",
            Content: "Test content",
        })
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        repo.FindByChatID(ctx, "bench-chat", 50, time.Now())
    }
}
```

**Subtasks**:
- [ ] Write Go benchmarks for Hub operations
- [ ] Write Go benchmarks for database operations
- [ ] Run load test with k6 or similar
- [ ] Measure message delivery latency
- [ ] Optimize database queries (EXPLAIN ANALYZE)
- [ ] Add database connection pool tuning
- [ ] Implement message list virtualization in frontend
- [ ] Add lazy loading for heavy components (Mermaid, Excalidraw)
- [ ] Measure and optimize frontend bundle size

**Performance Targets**:

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Message latency | < 500ms p95 | Timestamp diff in E2E test |
| Hub broadcast | > 10,000 msg/s | Go benchmark |
| DB query | < 50ms | Query timing |
| Initial page load | < 3s | Lighthouse |
| JS bundle | < 500KB | Build output |

**Exit Criteria**:
- [ ] Message latency < 500ms at p95
- [ ] Hub handles 500 concurrent connections
- [ ] Database queries optimized
- [ ] Frontend bundle < 500KB
- [ ] Lighthouse performance score ≥ 80

---

### Task 5.3: Docker Production Setup

**Description**: Create production-ready Docker configuration with multi-stage builds.

**Files to Create**:

```dockerfile
# backend/Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
```

```dockerfile
# frontend/Dockerfile
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Nginx runtime
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    environment:
      - PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER}
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=${DB_NAME}
      - JWT_SECRET=${JWT_SECRET}
      - S3_BUCKET=${S3_BUCKET}
      - S3_REGION=${S3_REGION}
    depends_on:
      postgres:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - byteroom

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - byteroom

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=${DB_USER}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped
    networks:
      - byteroom

networks:
  byteroom:
    driver: bridge

volumes:
  postgres_data:
```

**Subtasks**:
- [ ] Create backend multi-stage Dockerfile
- [ ] Create frontend Dockerfile with Nginx
- [ ] Create production docker-compose
- [ ] Configure Nginx for WebSocket proxying
- [ ] Add health check endpoints
- [ ] Create environment variable templates
- [ ] Test full stack locally with production config
- [ ] Document deployment process

**Exit Criteria**:
- [ ] Backend Docker image < 50MB
- [ ] Frontend Docker image < 30MB
- [ ] `docker-compose -f docker-compose.prod.yml up` works
- [ ] Health checks pass
- [ ] All services communicate correctly

---

### Task 5.4: CI/CD Pipeline Enhancement

**Description**: Enhance GitHub Actions pipeline with deployment automation.

**Files to Create**:

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  backend-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
          POSTGRES_DB: byteroom_test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache-dependency-path: backend/go.sum
      
      - name: Run migrations
        working-directory: backend
        run: |
          go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          migrate -path migrations -database "postgres://test:test@localhost:5432/byteroom_test?sslmode=disable" up
      
      - name: Run tests
        working-directory: backend
        run: go test ./... -v -race -coverprofile=coverage.out
        env:
          DB_HOST: localhost
          DB_USER: test
          DB_PASSWORD: test
          DB_NAME: byteroom_test
          JWT_SECRET: test-secret
      
      - name: Check coverage
        working-directory: backend
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 80%"
            exit 1
          fi
      
      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: backend
          version: latest

  frontend-test:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install dependencies
        working-directory: frontend
        run: npm ci
      
      - name: Run linter
        working-directory: frontend
        run: npm run lint
      
      - name: Run tests
        working-directory: frontend
        run: npm run test:coverage
      
      - name: Check coverage
        working-directory: frontend
        run: |
          # Parse coverage from vitest output
          npm run test:coverage -- --reporter=text | grep "All files" || true
      
      - name: Build
        working-directory: frontend
        run: npm run build

  e2e-test:
    runs-on: ubuntu-latest
    needs: [backend-test, frontend-test]
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install Playwright
        run: npx playwright install --with-deps
      
      - name: Start services
        run: docker-compose -f docker-compose.test.yml up -d
      
      - name: Wait for services
        run: |
          timeout 60 bash -c 'until curl -s http://localhost:8080/health; do sleep 2; done'
      
      - name: Run E2E tests
        run: npx playwright test
      
      - name: Upload test results
        uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-results
          path: playwright-report/

  build-images:
    runs-on: ubuntu-latest
    needs: [e2e-test]
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push backend
        uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: true
          tags: ghcr.io/${{ github.repository }}/backend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push frontend
        uses: docker/build-push-action@v5
        with:
          context: ./frontend
          push: true
          tags: ghcr.io/${{ github.repository }}/frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

**Subtasks**:
- [ ] Update test workflow with coverage gates
- [ ] Add E2E tests to CI pipeline
- [ ] Configure Docker image building
- [ ] Add security scanning (Trivy/Snyk)
- [ ] Configure branch protection rules
- [ ] Add PR template with checklist
- [ ] Set up deployment workflow (optional)

**Exit Criteria**:
- [ ] CI runs on all PRs
- [ ] Coverage gates enforced (80%)
- [ ] E2E tests run in CI
- [ ] Docker images built on main
- [ ] Pipeline completes in < 10 minutes

---

### Task 5.5: Security Hardening

**Description**: Implement security best practices and hardening.

**Subtasks**:
- [ ] Add rate limiting middleware
- [ ] Implement CORS configuration
- [ ] Add security headers (CSP, HSTS, etc.)
- [ ] Audit dependencies for vulnerabilities
- [ ] Implement request validation
- [ ] Add input size limits
- [ ] Review and test XSS sanitization
- [ ] Add SQL injection prevention validation
- [ ] Implement secure cookie settings

**Code to Implement**:

```go
// internal/api/middleware/security.go
func SecurityHeaders() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("X-Content-Type-Options", "nosniff")
            w.Header().Set("X-Frame-Options", "DENY")
            w.Header().Set("X-XSS-Protection", "1; mode=block")
            w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
            w.Header().Set("Content-Security-Policy", 
                "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';")
            next.ServeHTTP(w, r)
        })
    }
}

func RateLimiter(rps int) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(rps), rps*2)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too many requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            for _, allowed := range allowedOrigins {
                if origin == allowed {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                    w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                    w.Header().Set("Access-Control-Max-Age", "86400")
                    break
                }
            }
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**Exit Criteria**:
- [ ] Rate limiting prevents abuse
- [ ] Security headers configured
- [ ] No high/critical vulnerabilities in deps
- [ ] OWASP top 10 addressed
- [ ] Security review completed

---

### Task 5.6: Monitoring & Observability

**Description**: Implement logging, metrics, and health checks.

**Code to Implement**:

```go
// internal/api/handler/health.go
type HealthResponse struct {
    Status    string            `json:"status"`
    Version   string            `json:"version"`
    Uptime    string            `json:"uptime"`
    Checks    map[string]string `json:"checks"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
    checks := make(map[string]string)
    
    // Database check
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    
    if err := h.db.PingContext(ctx); err != nil {
        checks["database"] = "unhealthy: " + err.Error()
    } else {
        checks["database"] = "healthy"
    }
    
    // WebSocket hub check
    checks["websocket_hub"] = fmt.Sprintf("healthy: %d connections", h.hub.ConnectionCount())
    
    status := "healthy"
    statusCode := http.StatusOK
    for _, check := range checks {
        if !strings.HasPrefix(check, "healthy") {
            status = "unhealthy"
            statusCode = http.StatusServiceUnavailable
            break
        }
    }
    
    resp := HealthResponse{
        Status:  status,
        Version: h.version,
        Uptime:  time.Since(h.startTime).String(),
        Checks:  checks,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(resp)
}

// internal/infrastructure/logger/logger.go
type Logger struct {
    logger *slog.Logger
}

func NewLogger(level string, format string) *Logger {
    var handler slog.Handler
    
    opts := &slog.HandlerOptions{
        Level: parseLevel(level),
    }
    
    if format == "json" {
        handler = slog.NewJSONHandler(os.Stdout, opts)
    } else {
        handler = slog.NewTextHandler(os.Stdout, opts)
    }
    
    return &Logger{
        logger: slog.New(handler),
    }
}

func (l *Logger) Info(msg string, args ...any) {
    l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, err error, args ...any) {
    args = append(args, "error", err)
    l.logger.Error(msg, args...)
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
    // Extract request ID, user ID from context
    logger := l.logger
    
    if requestID := ctx.Value(RequestIDKey); requestID != nil {
        logger = logger.With("request_id", requestID)
    }
    
    if userID := ctx.Value(UserIDKey); userID != nil {
        logger = logger.With("user_id", userID)
    }
    
    return &Logger{logger: logger}
}
```

**Subtasks**:
- [ ] Implement structured logging with slog
- [ ] Add request ID middleware
- [ ] Create health check endpoint
- [ ] Add readiness/liveness probes
- [ ] Log key events (auth, messages, errors)
- [ ] Add basic metrics (connections, messages)
- [ ] Create Grafana dashboard (optional)

**Exit Criteria**:
- [ ] Health endpoint returns component status
- [ ] Structured logs in JSON format
- [ ] Request tracing with IDs
- [ ] Key metrics exposed

---

### Task 5.7: Documentation

**Description**: Create comprehensive documentation for developers and users.

**Files to Create**:

```markdown
# README.md (update)

# ByteRoom

Real-time chat platform for technical discussions with native support for code, diagrams, and system design.

## Features

- Real-time messaging with WebSocket
- Markdown with syntax highlighting
- Mermaid diagram rendering
- Excalidraw whiteboard integration
- Image uploads with S3

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Development Setup

```bash
# Clone repository
git clone https://github.com/byteroom/byteroom.git
cd byteroom

# Start infrastructure
docker-compose up -d postgres

# Run backend
cd backend
cp .env.example .env
make migrate-up
go run ./cmd/server

# Run frontend (new terminal)
cd frontend
npm install
npm run dev
```

### Running Tests

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# E2E tests
npm run test:e2e
```

## Architecture

See [docs/HLD.md](docs/HLD.md) for high-level architecture and [docs/LLD.md](docs/LLD.md) for detailed design.

## API Documentation

### REST Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/auth/register | Register new user |
| POST | /api/auth/login | Login user |
| GET | /api/chats | List user's chats |
| POST | /api/chats | Create new chat |
| GET | /api/chats/:id/messages | Get chat history |

### WebSocket Events

| Event | Direction | Description |
|-------|-----------|-------------|
| message.send | Client→Server | Send message |
| message.ack | Server→Client | Message acknowledged |
| message.new | Server→Client | New message received |
| typing.start | Client→Server | User started typing |
| typing.stop | Client→Server | User stopped typing |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT
```

```markdown
# CONTRIBUTING.md

# Contributing to ByteRoom

## Development Workflow

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Write tests first (TDD)
4. Implement feature
5. Ensure tests pass: `make test`
6. Submit pull request

## Code Style

### Go
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `golangci-lint run` before committing

### TypeScript
- Use functional components with hooks
- Follow ESLint configuration
- Run `npm run lint` before committing

## Commit Messages

Format: `<type>(<scope>): <description>`

Types: feat, fix, docs, style, refactor, test, chore

Examples:
- `feat(api): add message pagination`
- `fix(ws): handle reconnection edge case`
- `docs(readme): update setup instructions`

## Pull Request Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No linting errors
- [ ] Passes CI pipeline
```

**Subtasks**:
- [ ] Update README with setup instructions
- [ ] Create CONTRIBUTING.md
- [ ] Document all API endpoints
- [ ] Document WebSocket protocol
- [ ] Add inline code comments where needed
- [ ] Create deployment guide
- [ ] Add troubleshooting section

**Exit Criteria**:
- [ ] New developer can set up project in < 15 minutes
- [ ] API fully documented
- [ ] Architecture docs up to date
- [ ] Deployment process documented

---

### Task 5.8: Final Review & Launch Checklist

**Description**: Comprehensive review and preparation for launch.

**Checklist**:

```markdown
## Pre-Launch Checklist

### Code Quality
- [ ] All tests pass (unit, integration, E2E)
- [ ] Code coverage ≥ 80%
- [ ] No critical linting errors
- [ ] No security vulnerabilities in dependencies
- [ ] Code reviewed by team

### Functionality
- [ ] User registration works
- [ ] User login/logout works
- [ ] Chat creation works
- [ ] Messages send and receive in real-time
- [ ] Typing indicators work
- [ ] Code blocks syntax highlighted
- [ ] Mermaid diagrams render
- [ ] Excalidraw embeds work
- [ ] Image uploads work
- [ ] Mobile responsive

### Performance
- [ ] Message latency < 500ms
- [ ] Page load < 3s
- [ ] Bundle size < 500KB
- [ ] Handles 50 concurrent users

### Security
- [ ] JWT authentication secure
- [ ] XSS prevention tested
- [ ] SQL injection prevented
- [ ] Rate limiting in place
- [ ] HTTPS configured
- [ ] Security headers set

### Operations
- [ ] Health checks implemented
- [ ] Logging configured
- [ ] Error tracking set up
- [ ] Backup strategy defined
- [ ] Monitoring dashboards created

### Documentation
- [ ] README complete
- [ ] API documented
- [ ] Deployment guide written
- [ ] Runbook created
```

**Subtasks**:
- [ ] Run full test suite
- [ ] Perform security audit
- [ ] Load test with target users
- [ ] Review all documentation
- [ ] Test deployment process
- [ ] Create rollback procedure
- [ ] Prepare monitoring alerts
- [ ] Final stakeholder demo

**Exit Criteria**:
- [ ] All checklist items completed
- [ ] Stakeholder approval received
- [ ] Production environment ready
- [ ] Team trained on operations

---

## Phase 5 Exit Criteria Summary

### Automated Verification

```bash
# Full test suite
make test-all

# E2E tests
npm run test:e2e

# Security scan
make security-scan

# Build production images
docker-compose -f docker-compose.prod.yml build
```

### Manual Verification

| Check | Action | Expected Result |
|-------|--------|-----------------|
| Full user journey | Register → Login → Chat → Send | All flows work |
| Real-time | Two users chatting | Messages appear instantly |
| Performance | Load test 50 users | No degradation |
| Mobile | Test on phone | Responsive and functional |

### Quality Gates

- [ ] All automated tests pass
- [ ] E2E tests pass
- [ ] No critical security issues
- [ ] Performance targets met
- [ ] Documentation complete
- [ ] Launch checklist completed

### Deliverables

1. ✅ E2E test suite
2. ✅ Performance benchmarks and optimizations
3. ✅ Production Docker setup
4. ✅ CI/CD pipeline with deployment
5. ✅ Security hardening
6. ✅ Monitoring and observability
7. ✅ Complete documentation
8. ✅ Launch-ready application

---

## MVP Complete

Congratulations! Upon completing Phase 5, ByteRoom MVP is ready for production with:

- Real-time messaging via WebSocket
- Markdown and code block support
- Mermaid diagram rendering
- Excalidraw whiteboard integration
- Image uploads
- 100 DAU capacity
- < 500ms message latency
- 99.9% target availability

### Future Enhancements (Phase 2)

- Horizontal scaling with Redis routing
- Kafka for message durability
- Elasticsearch for search
- Mobile applications
- E2E encryption
