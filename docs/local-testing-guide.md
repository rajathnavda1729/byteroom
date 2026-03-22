# Local Testing Guide

This document walks you through every way to run and verify ByteRoom on your local machine — from spinning up the database through to full end-to-end browser tests.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [First-time Setup](#2-first-time-setup)
3. [Starting the Stack](#3-starting-the-stack)
4. [Backend Unit & Integration Tests](#4-backend-unit--integration-tests)
5. [Frontend Unit Tests](#5-frontend-unit-tests)
6. [Manual Smoke Tests (REST API)](#6-manual-smoke-tests-rest-api)
7. [Manual Smoke Tests (WebSocket)](#7-manual-smoke-tests-websocket)
8. [End-to-End (Playwright) Tests](#8-end-to-end-playwright-tests)
9. [Go Benchmarks](#9-go-benchmarks)
10. [Common Troubleshooting](#10-common-troubleshooting)

---

## 1. Prerequisites

| Tool | Min version | Install |
|------|-------------|---------|
| Go | 1.22 | https://go.dev/dl/ |
| Node.js | 20 | https://nodejs.org |
| Docker + Compose | any recent | https://docs.docker.com/get-docker/ |
| `migrate` CLI | v4 | See install steps below |
| `wscat` *(optional)* | any | `npm install -g wscat` |
| `curl` | any | pre-installed on macOS/Linux |

### Installing the `migrate` CLI

`go install` places binaries in `$(go env GOPATH)/bin` (typically `~/go/bin`).  
This directory must be on your `PATH` or the `migrate` command won't be found.

```bash
# 1. Install the binary
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. Add ~/go/bin to your shell PATH (one-time setup)
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc          # reload — or open a new terminal

# 3. Verify
migrate -version
```

> **Already done?** If you've previously run `go install` for any tool and `migrate` is
> still not found, the `PATH` line above is missing from your shell config. The three
> commands above fix it permanently.

### Verify all tools

```bash
go version          # go1.22+
node --version      # v20+
docker --version
migrate -version    # should print a version number
```

---

## 2. First-time Setup

### 2.1 Clone & install dependencies

```bash
git clone <repo-url>
cd byteroom

# Frontend dependencies
cd frontend && npm install && cd ..
```

### 2.2 Create the backend `.env` file

```bash
cd backend
cp ../.env.example .env
```

Edit `backend/.env` and set at minimum:

```dotenv
PORT=8080
ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=byteroom
DB_PASSWORD=byteroom_dev
DB_NAME=byteroom
DB_SSLMODE=disable

JWT_SECRET=dev-secret-min-32-chars-xxxxxxxxx
JWT_EXPIRY_HOURS=24
```

> **Note:** `JWT_SECRET` must be at least 32 characters. Leave S3 variables empty to disable image uploads.

---

## 3. Starting the Stack

### 3.1 Start PostgreSQL (development)

```bash
# From the repo root
docker-compose up -d postgres
```

Wait until healthy:

```bash
docker-compose ps            # postgres should show "healthy"
```

### 3.2 Run database migrations

```bash
cd backend
migrate -path migrations \
  -database "postgres://byteroom:byteroom_dev@localhost:5432/byteroom?sslmode=disable" up
```

Expected output:
```
1/u create_users_table (12ms)
2/u create_chats_table (8ms)
3/u create_messages_table (7ms)
4/u add_chat_member_role (6ms)
```

### 3.3 Start the backend server

The server automatically loads `backend/.env` on startup — no manual `export` needed.

```bash
# Still inside backend/
go run ./cmd/server
```

Expected output:
```
Database connected
ByteRoom server starting on :8080 (env=development)
```

### 3.4 Start the frontend dev server

Open a **new terminal**:

```bash
cd frontend
npm run dev
```

Expected output:
```
  VITE v6.x  ready in 300ms

  ➜  Local:   http://localhost:5173/
```

The app is now accessible at **http://localhost:5173**.

---

## 4. Backend Unit & Integration Tests

### 4.1 Unit tests (no database required)

```bash
cd backend
go test ./... -count=1
```

All packages should report `ok`. The suite takes < 10 s.

### 4.2 Race-detector run

```bash
go test -race ./... -count=1
```

### 4.3 Integration tests (require a live database)

Integration tests are tagged with `//go:build integration` so they are skipped by default. To run them, start the dedicated test database first:

```bash
# From repo root
docker-compose -f docker-compose.test.yml up -d

# Run migrations against the test DB (port 5433)
cd backend
migrate -path migrations \
  -database "postgres://test:test@localhost:5433/byteroom_test?sslmode=disable" up
```

Then run with the `integration` build tag:

```bash
go test -tags integration ./... \
  -count=1 \
  -v \
  DB_HOST=localhost \
  DB_USER=test \
  DB_PASSWORD=test \
  DB_NAME=byteroom_test
```

> The test DB runs on **port 5433** to avoid conflicting with the dev DB on 5432.

### 4.4 Coverage report

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
open coverage.html        # macOS
xdg-open coverage.html   # Linux
```

---

## 5. Frontend Unit Tests

### 5.1 Run all unit tests once

```bash
cd frontend
npm test
```

Expected output:
```
 Test Files  18 passed (18)
      Tests  96 passed (96)
```

### 5.2 Watch mode (re-runs on file save)

```bash
npm run test:watch
```

### 5.3 Coverage report

```bash
npm run test:coverage
```

Opens a coverage summary in the terminal. An HTML report is written to `frontend/coverage/`.

---

## 6. Manual Smoke Tests (REST API)

The backend must be running on port 8080. All examples use `curl`.

### 6.1 Health check

```bash
curl -s http://localhost:8080/health | jq
```

Expected:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2m34s",
  "checks": {
    "database": "healthy",
    "websocket_hub": "healthy: 0 connections"
  }
}
```

### 6.2 Register a user

```bash
curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}' | jq
```

Expected:
```json
{
  "user_id": "...",
  "username": "alice",
  "display_name": "alice",
  "token": "<JWT>"
}
```

### 6.3 Login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"password123"}' \
  | jq -r '.token')

echo "Token: $TOKEN"
```

### 6.4 Get current user

```bash
curl -s http://localhost:8080/api/users/me \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 6.5 Create a group chat

```bash
CHAT_ID=$(curl -s -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Tech Discussion","type":"group"}' \
  | jq -r '.chat_id')

echo "Chat ID: $CHAT_ID"
```

### 6.6 List chats

```bash
curl -s http://localhost:8080/api/chats \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 6.7 Get message history

```bash
curl -s "http://localhost:8080/api/chats/${CHAT_ID}/messages" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 6.8 Verify security headers

```bash
curl -sI http://localhost:8080/health | grep -E "X-Frame|X-Content|X-XSS|Referrer|Content-Security"
```

Expected (one header per line):
```
Content-Security-Policy: default-src 'self'; ...
Referrer-Policy: strict-origin-when-cross-origin
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
```

---

## 7. Manual Smoke Tests (WebSocket)

Requires `wscat` (`npm install -g wscat`) and the `$TOKEN` variable from §6.3.

> **How the SPA connects:** In development, the React app opens the WebSocket at **`ws://127.0.0.1:8080/ws`** (the Go server directly), not via Vite’s `/ws` proxy. That avoids cases where the dev proxy mishandles **server→client** frames (symptom: new messages only appear after a full page reload, while `GET /api/chats/.../messages` still works). In production builds, the client uses the same host as the page (`wss:` / `ws:` from `window.location`). Set **`VITE_WS_URL`** if the API/WebSocket live on another origin.

### 7.1 Connect

```bash
wscat -c "ws://localhost:8080/ws?token=${TOKEN}"
```

You should see `Connected (press CTRL+C to quit)`.

### 7.2 Send a ping

```json
{"event":"ping","data":{}}
```

Expected response:
```json
{"event":"pong","data":{}}
```

### 7.3 Send a chat message

Replace `<CHAT_ID>` with the ID from §6.5:

```json
{"event":"message.send","data":{"chat_id":"<CHAT_ID>","content_type":"markdown","content":"Hello from wscat!"}}
```

Expected response (within the same session):
```json
{"event":"message.ack","data":{"message_id":"...","status":true}}
```

### 7.4 Two-user real-time test

Open **two terminals**, connect as two different users (register a second user `bob` first), both join the same chat, and verify each user receives messages sent by the other.

---

## 8. End-to-End (Playwright) Tests

E2E tests drive a real browser against a running stack. The backend and frontend dev server must both be running (§3.3, §3.4).

### 8.1 Install browser binaries (first time only)

```bash
cd frontend
npx playwright install chromium
```

### 8.2 Run all E2E tests (headless)

```bash
npm run test:e2e
```

### 8.3 Run with interactive UI

```bash
npm run test:e2e:ui
```

This opens the Playwright Test Runner — you can step through tests, inspect the DOM, and see screenshots.

### 8.4 Run a single spec file

```bash
npx playwright test e2e/auth.spec.ts
```

Available spec files:

| File | Covers |
|------|--------|
| `e2e/auth.spec.ts` | Register, login, logout, redirect |
| `e2e/chat.spec.ts` | Chat list, select, send, real-time, typing indicator |
| `e2e/markdown.spec.ts` | Bold/italic, code blocks, links, inline code |

### 8.5 View HTML report after a failure

```bash
npx playwright show-report
```

---

## 9. Go Benchmarks

Benchmarks run against the in-process Hub (no database needed).

```bash
cd backend

# Hub broadcast benchmarks
go test ./internal/websocket/... \
  -run='^$' \
  -bench=BenchmarkHub \
  -benchmem \
  -benchtime=3s

# All benchmarks
go test ./... -run='^$' -bench=. -benchmem -benchtime=2s
```

Sample output:

```
BenchmarkHub_BroadcastToChat-12       12000000    183 ns/op    0 B/op    0 allocs/op
BenchmarkHub_ConcurrentBroadcasts-12   5000000    487 ns/op    0 B/op    0 allocs/op
```

---

## 10. Common Troubleshooting

### `zsh: command not found: migrate`

`~/go/bin` is not on your `PATH`. Fix it permanently:

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc
source ~/.zshrc
migrate -version    # should now work
```

### `migrate: no change` or migrations already applied

```bash
# Check current version
migrate -path backend/migrations \
  -database "postgres://byteroom:byteroom_dev@localhost:5432/byteroom?sslmode=disable" version
```

If needed, force a specific version:

```bash
migrate -path backend/migrations \
  -database "postgres://..." force 4
```

### `connection refused` on port 5432

The postgres container may still be starting:

```bash
docker-compose ps           # check status
docker-compose logs postgres
```

### `invalid or missing JWT_SECRET`

The secret must be at least 32 characters in `backend/.env`.

### Backend fails with `address already in use`

Another process is on port 8080:

```bash
lsof -ti:8080 | xargs kill -9   # macOS/Linux
```

### Frontend dev server fails with `EADDRINUSE`

Port 5173 is in use:

```bash
lsof -ti:5173 | xargs kill -9
```

### E2E tests fail with `page.goto: net::ERR_CONNECTION_REFUSED`

Ensure both the backend (`:8080`) and the frontend dev server (`:5173`) are running before executing Playwright tests.

### Go test cache producing stale results

```bash
cd backend && go clean -testcache
```
