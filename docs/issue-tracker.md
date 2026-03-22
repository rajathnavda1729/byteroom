# ByteRoom issue tracker

This document records **known issues, root causes, and fixes** so regressions and context are easy to find. Add new rows as issues are discovered and resolved.

**How to add an entry**

1. Assign the next **ID** (`BR-NNN`).
2. Fill **Area**, **Summary**, **Symptoms**, **Root cause**, **Fix**, and **Status**.
3. Optionally list **Files / PR** for traceability.

---

## Index (quick lookup)

| ID | Area | Summary | Status |
|----|------|---------|--------|
| BR-001 | DevOps | `migrate: command not found` | Fixed |
| BR-002 | Backend | `JWT_SECRET is required` despite `.env` | Fixed |
| BR-003 | Frontend | Registration / API returns Not Found | Fixed |
| BR-004 | Frontend | `TypeError` on chat list (`res.chats` undefined) | Fixed |
| BR-005 | Frontend | Vite console `EPIPE` / WS proxy noise | Mitigated |
| BR-006 | Backend | `GET /ws` returns 500 | Fixed |
| BR-007 | Chat | Sidebar shows “Chat” / wrong display name | Fixed |
| BR-008 | Real-time | Messages not received by other user | Fixed |
| BR-009 | Auth | `user.id` undefined after login/register | Fixed |
| BR-010 | API | Message history empty in UI (HTTP contract) | Fixed |
| BR-011 | Chat | Duplicate direct chats for same two users | Fixed |
| BR-012 | Real-time | Messages not persisted; UI looked “sent” | Fixed |
| BR-013 | Real-time | Receiver needs reload to see new messages | Fixed |
| BR-014 | Frontend | Vitest picked up Playwright specs | Fixed |
| BR-015 | Backend | Health handler / JWT mocks after refactors | Fixed |

---

## Detailed log

### BR-001 — `migrate: command not found`

| Field | Detail |
|--------|--------|
| **Area** | DevOps / local setup |
| **Symptoms** | Shell cannot find `migrate` after `go install`. |
| **Root cause** | `GOPATH/bin` (or `go env GOPATH`/bin) not on `PATH`. |
| **Fix** | Document in README / `docs/local-testing-guide.md`; `Makefile install-tools`; user adds `export PATH="$PATH:$(go env GOPATH)/bin"`. |
| **Status** | Fixed (docs + tooling) |

---

### BR-002 — `JWT_SECRET is required` with existing `.env`

| Field | Detail |
|--------|--------|
| **Area** | Backend / config |
| **Symptoms** | Server exits on startup even when `backend/.env` exists. |
| **Root cause** | Config read only `os.Getenv`; `.env` was never loaded. |
| **Fix** | `godotenv.Load()` at start of `config.Load()` (dev convenience). Production should still inject env via platform. |
| **Status** | Fixed |

---

### BR-003 — Signup / API “Not Found” from the browser

| Field | Detail |
|--------|--------|
| **Area** | Frontend (Vite) |
| **Symptoms** | `POST /api/auth/register` (and other `/api/*`) hit Vite and 404. |
| **Root cause** | No dev proxy from Vite to the Go server on `:8080`. |
| **Fix** | `vite.config.ts` → `server.proxy` for `/api` and `/ws`. |
| **Status** | Fixed |

---

### BR-004 — Chat page crash: `Cannot read properties of undefined (reading 'map')`

| Field | Detail |
|--------|--------|
| **Area** | Frontend + API contract |
| **Symptoms** | Crash on `/` after login when processing chat list. |
| **Root cause** | Backend returned a raw JSON array; frontend expected `{ chats: [...] }`. |
| **Fix** | Backend wraps list in `{ "chats": [...] }`; frontend uses `(res.chats ?? [])`. |
| **Status** | Fixed |

---

### BR-005 — Vite dev server `EPIPE` / WS proxy errors

| Field | Detail |
|--------|--------|
| **Area** | Frontend (Vite) |
| **Symptoms** | Repeated `Error: write EPIPE` / proxy socket errors after signup or WS reconnect. |
| **Root cause** | HMR WebSocket vs `/ws` proxy overlap; StrictMode double-mount closing sockets quickly. |
| **Fix** | Dedicated HMR port in `vite.config.ts`; proxy `configure` to ignore benign `EPIPE`/`ECONNRESET` in dev. |
| **Status** | Mitigated (dev noise; not an app logic bug) |

---

### BR-006 — WebSocket upgrade returns HTTP 500

| Field | Detail |
|--------|--------|
| **Area** | Backend / middleware |
| **Symptoms** | Logs show `GET /ws 500`; WS never stays up. |
| **Root cause** | `Logger` middleware wrapped `ResponseWriter` without implementing `http.Hijacker`, so Gorilla `Upgrade` failed. |
| **Fix** | `responseRecorder.Hijack()` delegating to the underlying writer. |
| **Status** | Fixed |

---

### BR-007 — Sidebar / header show “Chat” or the logged-in user’s name

| Field | Detail |
|--------|--------|
| **Area** | Backend + Frontend |
| **Symptoms** | Direct chats labeled generically or as self; filter “other member” ineffective. |
| **Root cause** | Chat list query did not load member display fields; `user.id` often missing (see BR-009); name logic used wrong member. |
| **Fix** | `GetMemberDetails` + member DTOs on list/create; `chatDisplayName` / header use “other” member vs `currentUserId`. |
| **Status** | Fixed |

---

### BR-008 — Messages not received on the other side (early)

| Field | Detail |
|--------|--------|
| **Area** | Real-time / Hub |
| **Symptoms** | Sender sees activity; receiver sees nothing until later fixes. |
| **Root cause** | Hub rooms populated only at WS connect; new chats did not add existing connections to rooms; duplicate DMs possible (BR-011). |
| **Fix** | `SubscribeUserToRoom`, `BroadcastToUser`, `chat.new` payload; frontend `addChat` on `chat.new`; dedup DMs (BR-011). |
| **Status** | Fixed (see also BR-012, BR-013) |

---

### BR-009 — `user.id` undefined in auth store

| Field | Detail |
|--------|--------|
| **Area** | Frontend / API contract |
| **Symptoms** | WS/UI logic depending on `user.id` broken; wrong peer labels. |
| **Root cause** | Auth responses use nested `user: { user_id, display_name, ... }`; pages assumed flat `user_id` on response. |
| **Fix** | `LoginPage` / `RegisterPage` map `res.user.user_id` → `id`, etc.; persist user JSON in `localStorage` with token for refresh. |
| **Status** | Fixed |

---

### BR-010 — Message history always empty in UI (HTTP 200)

| Field | Detail |
|--------|--------|
| **Area** | API contract |
| **Symptoms** | “No messages yet” after selecting a chat that has rows in DB. |
| **Root cause** | `GET .../messages` returned a raw array; UI expected `{ messages: [...] }`, so `res.messages` was `undefined`. |
| **Fix** | Backend returns `{ "messages": [...] }`; frontend `res.messages ?? []`. |
| **Status** | Fixed |

---

### BR-011 — Two “direct” threads for the same two users

| Field | Detail |
|--------|--------|
| **Area** | Backend / domain |
| **Symptoms** | Each user sees a different thread; messages only in one room. |
| **Root cause** | `CreateDirect` always inserted a new chat; no lookup for an existing DM between the pair. |
| **Fix** | `Repository.FindDirectBetween` + `CreateDirect` returns existing room when present. |
| **Status** | Fixed (existing duplicate rows need one-time DB cleanup in dev) |

---

### BR-012 — No rows in `messages` while UI showed sent bubbles

| Field | Detail |
|--------|--------|
| **Area** | Frontend + Backend |
| **Symptoms** | DB `messages` empty; optimistic bubbles + delayed “confirm” looked like success. |
| **Root cause** | `useWebSocket.send` no-oped when `readyState !== OPEN`; no HTTP fallback. |
| **Fix** | `POST /api/chats/{id}/messages` (auth) reuses `MessageService.Send`; hub `BroadcastToChatExceptUser` for peers; client sends via `api.post`. |
| **Status** | Fixed |

---

### BR-013 — Receiver must reload to see new messages

| Field | Detail |
|--------|--------|
| **Area** | Hub + Frontend (Vite) |
| **Symptoms** | History correct after reload; live updates missing. |
| **Root cause** | (1) Hub used separate channels; `select` could run **broadcast before register**. (2) WS registered asynchronously before pumps. (3) Vite `/ws` proxy sometimes drops server→client frames. (4) JWT in query not URL-encoded (`+` → space). |
| **Fix** | Single FIFO `hubOp` queue; `RegisterWithAck` before pumps; dev WS to `ws://127.0.0.1:8080/ws` via `resolveWebSocketURL`; `encodeURIComponent(token)`; optional visibility refetch; dev `console.debug` for WS events. |
| **Status** | Fixed |

---

### BR-014 — `Playwright Test did not expect test.describe` under `npm test`

| Field | Detail |
|--------|--------|
| **Area** | Frontend / tooling |
| **Symptoms** | Vitest fails on `e2e/*.spec.ts`. |
| **Root cause** | Vitest discovered Playwright files. |
| **Fix** | `vitest.config.ts` → `exclude: ['**/e2e/**']`. |
| **Status** | Fixed |

---

### BR-015 — Backend tests / vet failures after interface changes

| Field | Detail |
|--------|--------|
| **Area** | Backend / tests |
| **Symptoms** | Mock types missing `Search`, `FindDirectBetween`; `NewChatHandler` arity; message list JSON shape in tests. |
| **Root cause** | Interfaces and handlers evolved; tests not updated. |
| **Fix** | Update mocks and test decoders (`messages` wrapper, `Send` on message service mock, etc.). |
| **Status** | Fixed |

---

## Template (copy for new issues)

```markdown
### BR-NNN — <short title>

| Field | Detail |
|--------|--------|
| **Area** | Backend / Frontend / DevOps / Real-time / … |
| **Symptoms** | What the user or logs saw |
| **Root cause** | Why it happened |
| **Fix** | What we changed (and where) |
| **Status** | Open / In progress / Fixed / Won’t fix |
| **Files / PR** | Optional: paths or link |
```

---

## Related docs

- [Local testing guide](./local-testing-guide.md) — setup, smoke tests, troubleshooting  
- [README](../README.md) — API overview, quick start  
- [CONTRIBUTING](../CONTRIBUTING.md) — workflow and standards  
