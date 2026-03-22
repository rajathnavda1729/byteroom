# ByteRoom: Developer onboarding

This guide orients new contributors to the repository: what the system is for, where to start reading code, how requests flow through the stack, and where to go for authoritative deep dives on the harder topics. It complements the [README](../README.md) (quick start, API table), [HLD](./HLD.md) (architecture intent), [LLD](./LLD.md) (schemas and contracts), and [local testing guide](./local-testing-guide.md).

---

## Why ByteRoom exists (significance)

ByteRoom is a **real-time chat product aimed at technical teams**. Compared to a generic chat app, the important differences are:

- **Rich technical content** — Markdown, syntax highlighting, Mermaid diagrams, and Excalidraw-style embeds are first-class, not afterthoughts.
- **Safety by design** — User-supplied markdown and HTML are dangerous by default (XSS). The backend sanitizes persisted content; the frontend must still treat untrusted strings carefully when rendering.
- **Durability and perceived latency** — Messages are stored in PostgreSQL; WebSockets carry live updates. Idempotent message IDs support safe retries.

The **layered backend** (HTTP handlers → domain services → repositories → PostgreSQL, plus a WebSocket hub) keeps business rules testable and infrastructure swappable. The **SPA frontend** talks REST for CRUD and history, and WebSocket for live events—matching how many production chat products are structured at modest scale.

---

## Where to start (first sessions)

Use this order so each document and directory builds on the last.

| Step | What to do | Why |
|------|------------|-----|
| 1 | Run the app per [README](../README.md) Quick Start | Confirms toolchain (Go, Node, Docker/Postgres). |
| 2 | Skim [HLD](./HLD.md) sections 1–4 | Context, components, and message flow at a glance. |
| 3 | Open `backend/cmd/server/main.go` | Single place where DB, services, hub, router, and shutdown are wired. |
| 4 | Read `backend/internal/api/router.go` | All HTTP routes and global middleware in one file. |
| 5 | Open `frontend/src/App.tsx` then `frontend/src/pages/ChatPage.tsx` | Routing and main chat UI entry. |
| 6 | Run tests per [local-testing-guide](./local-testing-guide.md) | Validates your environment and shows how quality is enforced. |
| 7 | Follow [CONTRIBUTING](../CONTRIBUTING.md) for PRs | Commit style, lint, and review expectations. |

**If you are fixing a bug:** check [issue-tracker](./issue-tracker.md) for similar past fixes. **If you are implementing a feature:** cross-check contracts in [LLD](./LLD.md) and the phase notes under [execution-phases](./execution-phases/README.md).

---

## Code navigation

### Repository layout

| Path | Role |
|------|------|
| `backend/cmd/server/` | Process entrypoint: config, DB, dependency wiring, `http.Server`, graceful shutdown. |
| `backend/internal/api/` | HTTP router (`router.go`), handlers, shared response helpers. |
| `backend/internal/api/middleware/` | JWT auth, CORS, logging, security headers, request IDs. |
| `backend/internal/domain/` | Domain entities and services (`user`, `chat`, `message`). Depends on repository interfaces, not Postgres. |
| `backend/internal/infrastructure/postgres/` | SQL repositories, migrations context (see LLD for schema). |
| `backend/internal/infrastructure/sanitizer/` | Markdown/HTML sanitization before persist. |
| `backend/internal/infrastructure/s3/` | Optional pre-signed upload URLs. |
| `backend/internal/websocket/` | Hub, client model, event router, message and typing handlers. |
| `frontend/src/pages/` | Top-level routes (login, register, chat shell). |
| `frontend/src/components/` | UI: `auth/`, `chat/` (messages, markdown, mermaid, uploads). |
| `frontend/src/services/` | REST client (`api.ts`). |
| `frontend/src/hooks/` | WebSocket hook (`useWebSocket.ts`). |
| `frontend/src/stores/` | Zustand stores (e.g. auth). |
| `frontend/e2e/` | Playwright specs. |
| `docs/` | HLD, LLD, testing, phases, this onboarding doc. |

### Tracing a user action

**Send a message over WebSocket**

1. Client connects: `GET /ws` with token — `backend/internal/api/handler/ws_handler.go`.
2. Frames are dispatched by event name — `backend/internal/websocket/event_router.go`.
3. `message.send` handling persists via domain — `backend/internal/websocket/message_handler.go` → `backend/internal/domain/message/service.go`.
4. Content is sanitized — `backend/internal/infrastructure/sanitizer/` (wired in `main.go` via a small adapter).
5. Broadcasts go through the hub — `backend/internal/websocket/hub.go`.

**Load history or use REST**

- Follow the route in `router.go` to the matching file under `backend/internal/api/handler/`.
- Handlers call domain services; services use `postgres` repositories.

**Frontend parity**

- API calls: `frontend/src/services/api.ts`.
- Live updates: `frontend/src/hooks/useWebSocket.ts` and chat components under `frontend/src/components/chat/`.

### Search and test tips

- **Find a route:** search for the path string (e.g. `/api/chats`) in `router.go`, then open the handler.
- **Find event types:** search `Event` constants in `backend/internal/websocket/`.
- **Backend tests:** live next to packages (`*_test.go`); integration-style DB tests under `infrastructure/postgres/`.
- **Frontend:** Vitest for units; Playwright under `frontend/e2e/` for full flows.

---

## Complex topics and reliable deep dives

These are the areas most likely to slow you down without background. Sources below are **canonical or widely trusted** (specs, official docs, OWASP)—use them when you need theory, not just this repo’s implementation.

### WebSockets (protocol and browser API)

- **What ByteRoom uses them for:** duplex events (messages, typing, ping/pong) after an HTTP upgrade.
- **Deep dive:** [RFC 6455 — The WebSocket Protocol](https://www.rfc-editor.org/rfc/rfc6455) (normative wire behavior); [MDN — WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket) (browser-side usage).

### Go HTTP server, routing, and graceful shutdown

- **What matters here:** `net/http` mux patterns in Go 1.22+, middleware chaining, closing listeners and long-lived goroutines on shutdown.
- **Deep dive:** [Go `net/http` package documentation](https://pkg.go.dev/net/http); [Go documentation — Common mistakes](https://go.dev/doc/articles/error_handling.html) for error handling patterns.

### `gorilla/websocket` (Go library)

- **What ByteRoom uses it for:** upgrading connections and framing messages in the hub.
- **Deep dive:** [gorilla/websocket on pkg.go.dev](https://pkg.go.dev/github.com/gorilla/websocket) and the project’s examples on GitHub.

### JWT and authentication

- **What ByteRoom uses them for:** stateless auth for REST; WebSocket auth via query parameter on upgrade (see handler code—treat tokens as secrets in transit).
- **Deep dive:** [RFC 7519 — JSON Web Token (JWT)](https://www.rfc-editor.org/rfc/rfc7519); library docs for [golang-jwt/jwt](https://pkg.go.dev/github.com/golang-jwt/jwt/v5).

### XSS, HTML sanitization, and safe Markdown

- **Why it matters:** Chat content can contain HTML-like payloads; storing or rendering them unsafely enables session theft and defacement.
- **Deep dive:** [OWASP — Cross Site Scripting (XSS)](https://owasp.org/www-community/attacks/xss/); [OWASP — DOM based XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/DOM_based_XSS_Prevention_Cheat_Sheet.html). For the Go sanitizer used in backend: [bluemonday](https://github.com/microcosm-cc/bluemonday) and its policy documentation.

### PostgreSQL, transactions, and schema design

- **What matters here:** message ordering indexes, foreign keys, and idempotent inserts keyed by client message IDs (see LLD).
- **Deep dive:** [PostgreSQL documentation — Tutorial](https://www.postgresql.org/docs/current/tutorial.html); [PostgreSQL documentation — Indexes](https://www.postgresql.org/docs/current/indexes.html).

### React, hooks, and client state

- **What ByteRoom uses:** function components, hooks, Zustand for auth, Vite as the dev/build tool.
- **Deep dive:** [React — Learn](https://react.dev/learn) (especially “Thinking in React” and hooks rules); [Zustand documentation](https://docs.pmnd.rs/zustand/getting-started/introduction).

### Content Security Policy and related browser defenses

- **Why it matters:** Defense in depth alongside sanitization; our middleware sets security headers—understand what they do before changing them.
- **Deep dive:** [MDN — Content Security Policy (CSP)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP); [OWASP — Content Security Policy Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html).

### Pre-signed S3 uploads

- **What ByteRoom uses them for:** browser-direct uploads without passing file bytes through the API server.
- **Deep dive:** [AWS documentation — Presigned URLs](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html) (or your S3-compatible provider’s equivalent).

---

## Glossary

| Term | Meaning in ByteRoom |
|------|---------------------|
| **ACK / ack** | Server acknowledgment that a message was accepted or deduplicated (see WebSocket event names in code and README). |
| **Bluemonday** | Go HTML sanitizer library used when persisting markdown/content (via policy in `internal/infrastructure/sanitizer`). |
| **Client-generated message ID** | UUID supplied by the client so retries do not create duplicates; stored as `messages.id` per LLD. |
| **Domain layer** | `internal/domain/*` — business rules and service APIs depending on interfaces, not concrete DB code. |
| **Event router** | WebSocket component that maps `event` fields in frames to handlers (`internal/websocket/event_router.go`). |
| **GFM** | GitHub Flavored Markdown — tables, strikethrough, etc., where enabled in the renderer. |
| **Hub** | In-memory registry of active WebSocket clients and broadcast logic (`internal/websocket/hub.go`). |
| **Idempotency** | Sending the same logical message twice with the same ID does not create two rows; see message service behavior. |
| **JWT** | JSON Web Token used for authenticated REST and WebSocket establishment. |
| **LLD** | Low-level design doc: schema, endpoints, and implementation-oriented detail (`docs/LLD.md`). |
| **Middleware** | HTTP wrappers for auth, CORS, logging, security headers, request IDs (`internal/api/middleware`). |
| **Mermaid** | Text-based diagram syntax rendered in the chat UI (lazy-loaded on the frontend). |
| **Repository (pattern)** | Structs that implement persistence for domain entities (e.g. `MessageRepository` in Postgres). |
| **Sanitizer** | Component that strips or escapes dangerous HTML/markdown before storage or display policy enforcement. |
| **WebSocket frame** | JSON envelope with an `event` name and payload fields exchanged over `/ws`. |
| **Zustand** | Lightweight global state library used for authentication state on the frontend. |

---

## Related documents

- [README](../README.md) — setup, API summary, architecture one-liner  
- [CONTRIBUTING](../CONTRIBUTING.md) — workflow and standards  
- [local-testing-guide](./local-testing-guide.md) — tests and manual checks  
- [HLD](./HLD.md) / [LLD](./LLD.md) — design intent and contracts  
- [execution-phases](./execution-phases/README.md) — phased delivery narrative  

Welcome to the codebase—when in doubt, start from `main.go` and `router.go`, then follow the call chain into the domain package that matches your feature.
