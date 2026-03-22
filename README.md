# ByteRoom

Real-time chat platform for technical teams — with first-class support for **code**, **Mermaid diagrams**, and **Excalidraw** system-design whiteboards.

---

## Features

| Feature | Details |
|---------|---------|
| Real-time messaging | WebSocket with heartbeat & auto-reconnect |
| Markdown rendering | GFM tables, code blocks, links, lists |
| Syntax highlighting | 20+ languages via Prism |
| Mermaid diagrams | Lazy-loaded, dark-mode, zoom controls |
| Excalidraw embed | Read-only & editable whiteboard view |
| Image uploads | Pre-signed S3 URLs (optional) |
| Authentication | JWT with automatic refresh |
| Typing indicators | Per-chat with auto-timeout |

---

## Quick Start

### Prerequisites

| Tool | Version |
|------|---------|
| Go | 1.22+ |
| Node.js | 20+ |
| Docker + Compose | any recent |
| PostgreSQL | 15+ (or use Docker) |

### 1. Start the database

```bash
docker-compose up -d postgres
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your settings (JWT_SECRET is required)
```

### 3. Run the backend

```bash
cd backend
go run ./cmd/server
# Server starts on :8080
```

### 4. Run the frontend

```bash
cd frontend
npm install
npm run dev
# App available at http://localhost:5173
```

---

## Running Tests

> For a complete walkthrough — unit tests, integration tests, manual API smoke tests, WebSocket testing, and E2E Playwright tests — see **[docs/local-testing-guide.md](docs/local-testing-guide.md)**.
>
> Resolved bugs and fixes are logged in **[docs/issue-tracker.md](docs/issue-tracker.md)** (symptoms, root cause, fix).

```bash
# All tests
make test

# Backend only (unit + race detector)
make test-backend

# Frontend only
make test-frontend

# E2E tests (requires backend + frontend running)
make test-e2e

# With coverage
make test-coverage
```

---

## API Reference

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/register` | Register new user |
| `POST` | `/api/auth/login` | Login, returns JWT |
| `GET`  | `/api/users/me` | Get current user |

### Chats

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/chats` | List user's chats |
| `POST` | `/api/chats` | Create new chat |
| `GET`  | `/api/chats/{id}` | Get chat details |
| `POST` | `/api/chats/{id}/members` | Add member |
| `DELETE` | `/api/chats/{id}/members/{userId}` | Remove member |

### Messages

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET`  | `/api/chats/{id}/messages` | Get message history |
| `POST` | `/api/chats/{id}/messages` | Send a message (persisted; body: `message_id`, `content`, `content_type`) |

### Uploads

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/upload/request` | Request pre-signed S3 URL |

### WebSocket

Connect to `ws://host/ws?token=<JWT>`.

| Event | Direction | Description |
|-------|-----------|-------------|
| `message.send` | Client→Server | Send a message |
| `message.ack` | Server→Client | Server acknowledged message |
| `message.new` | Server→Client | New message broadcast |
| `message.error` | Server→Client | Send error |
| `typing.start` | Client→Server | User started typing |
| `typing.stop` | Client→Server | User stopped typing |
| `user.typing` | Server→Client | Typing status broadcast |
| `ping` / `pong` | Both | Application heartbeat |

---

## Architecture

```
frontend (React + Vite)
    │
    ├── HTTP  ──►  /api/*     (Go net/http, mux)
    └── WS    ──►  /ws        (gorilla/websocket Hub)
                                │
                            domain layer
                                │
                         PostgreSQL (messages, chats, users)
                                │
                         S3-compatible (image uploads, optional)
```

See [`docs/HLD.md`](docs/HLD.md) for a high-level diagram and [`docs/LLD.md`](docs/LLD.md) for the detailed design. New contributors should read [`docs/dev-onboarding.md`](docs/dev-onboarding.md) for code navigation, deep-dive references, and a glossary.

---

## Production Deployment

```bash
# Copy and fill in production values
cp .env.example .env

# Build and start all services
docker-compose -f docker-compose.prod.yml up -d

# Verify
curl http://localhost/health
```

See `docker-compose.prod.yml` and `.env.example` for all configuration options.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
