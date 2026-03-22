# Phase 1: Foundation & Project Setup

## Objective

Establish the project infrastructure, development environment, and CI/CD pipeline. This phase creates the scaffolding for test-driven development.

## Duration Estimate

3 development days

## Prerequisites

- Go 1.22+ installed
- Node.js 20+ installed
- Docker & Docker Compose installed
- PostgreSQL client tools
- Git configured

---

## Tasks

### Task 1.1: Initialize Backend Project Structure

**Description**: Create Go module with proper directory structure following clean architecture.

**Subtasks**:
- [ ] Initialize Go module (`go mod init github.com/byteroom/backend`)
- [ ] Create directory structure:
  ```
  backend/
  ├── cmd/server/main.go
  ├── internal/
  │   ├── api/handler/
  │   ├── api/middleware/
  │   ├── websocket/
  │   ├── domain/message/
  │   ├── domain/chat/
  │   ├── domain/user/
  │   ├── infrastructure/postgres/
  │   └── config/
  ├── migrations/
  ├── Makefile
  └── .env.example
  ```
- [ ] Add essential dependencies to `go.mod`
- [ ] Create `Makefile` with common commands

**Test Requirements**:
```bash
# Verify project compiles
go build ./...
```

**Exit Criteria**:
- [ ] `go build ./...` succeeds with no errors
- [ ] All directories created as specified
- [ ] `go.mod` contains required dependencies

---

### Task 1.2: Initialize Frontend Project Structure

**Description**: Create React + TypeScript + Vite project with Tailwind CSS.

**Subtasks**:
- [ ] Create Vite project: `npm create vite@latest frontend -- --template react-ts`
- [ ] Install and configure Tailwind CSS
- [ ] Install testing libraries (Vitest, Testing Library)
- [ ] Create directory structure:
  ```
  frontend/
  ├── src/
  │   ├── components/ui/
  │   ├── components/chat/
  │   ├── hooks/
  │   ├── stores/
  │   ├── services/
  │   ├── types/
  │   └── utils/
  ├── vitest.config.ts
  └── tailwind.config.js
  ```
- [ ] Configure path aliases (`@/`)
- [ ] Add base TypeScript types

**Test Requirements**:
```bash
# Verify build succeeds
npm run build

# Verify tests run
npm run test
```

**Exit Criteria**:
- [ ] `npm run build` produces dist folder
- [ ] `npm run test` runs (even with 0 tests)
- [ ] Tailwind classes work in sample component
- [ ] TypeScript strict mode enabled

---

### Task 1.3: Docker Development Environment

**Description**: Create Docker Compose setup for local development with PostgreSQL.

**Subtasks**:
- [ ] Create `docker-compose.yml` with services:
  - PostgreSQL 15
  - (Optional) pgAdmin for database management
- [ ] Create `docker-compose.test.yml` for integration tests
- [ ] Add healthchecks for database readiness
- [ ] Create initialization scripts for test database

**Files to Create**:

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: byteroom
      POSTGRES_PASSWORD: byteroom_dev
      POSTGRES_DB: byteroom
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U byteroom"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

**Test Requirements**:
```bash
# Verify containers start
docker-compose up -d
docker-compose ps  # Should show healthy

# Verify database connection
psql -h localhost -U byteroom -d byteroom -c "SELECT 1"
```

**Exit Criteria**:
- [ ] `docker-compose up -d` starts all services
- [ ] PostgreSQL is accessible on localhost:5432
- [ ] Database connection from Go app succeeds

---

### Task 1.4: Database Migration Setup

**Description**: Set up database migration tool and create initial schema.

**Subtasks**:
- [ ] Install golang-migrate: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- [ ] Create initial migration files:
  - `000001_create_users_table.up.sql`
  - `000001_create_users_table.down.sql`
  - `000002_create_chats_table.up.sql`
  - `000002_create_chats_table.down.sql`
  - `000003_create_messages_table.up.sql`
  - `000003_create_messages_table.down.sql`
- [ ] Add migration commands to Makefile
- [ ] Create seed data script for development

**Makefile Commands**:
```makefile
migrate-up:
	migrate -path migrations -database "postgres://byteroom:byteroom_dev@localhost:5432/byteroom?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://byteroom:byteroom_dev@localhost:5432/byteroom?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)
```

**Test Requirements**:
```bash
# Verify migrations run
make migrate-up
make migrate-down
make migrate-up  # Should be idempotent
```

**Exit Criteria**:
- [ ] All migration files created
- [ ] `make migrate-up` creates all tables
- [ ] `make migrate-down` drops all tables cleanly
- [ ] Migrations are idempotent

---

### Task 1.5: CI/CD Pipeline Setup

**Description**: Configure GitHub Actions for automated testing and linting.

**Subtasks**:
- [ ] Create `.github/workflows/test.yml`
- [ ] Configure Go tests with coverage
- [ ] Configure Frontend tests with coverage
- [ ] Add linting (golangci-lint, ESLint)
- [ ] Add pre-commit hooks

**GitHub Actions Workflow**:

```yaml
# .github/workflows/test.yml
name: Test Suite

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
      - name: Run tests
        run: go test ./... -coverprofile=coverage.out
        working-directory: backend
      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 80%"
            exit 1
          fi
        working-directory: backend

  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
        working-directory: frontend
      - run: npm run test:coverage
        working-directory: frontend
```

**Test Requirements**:
```bash
# Verify workflow syntax
act -n  # dry run with act tool

# Verify hooks work
git commit --allow-empty -m "test: verify hooks"
```

**Exit Criteria**:
- [ ] GitHub Actions workflow file created
- [ ] CI runs on push/PR to main
- [ ] Tests execute in CI environment
- [ ] Coverage thresholds enforced

---

### Task 1.6: Development Scripts & Documentation

**Description**: Create developer onboarding scripts and documentation.

**Subtasks**:
- [ ] Create `scripts/setup.sh` for first-time setup
- [ ] Create `scripts/dev.sh` to start development environment
- [ ] Update README with setup instructions
- [ ] Create CONTRIBUTING.md with guidelines
- [ ] Add .env.example files

**Setup Script**:
```bash
#!/bin/bash
# scripts/setup.sh

echo "🚀 Setting up ByteRoom development environment..."

# Check prerequisites
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed."; exit 1; }
command -v node >/dev/null 2>&1 || { echo "Node.js is required but not installed."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed."; exit 1; }

# Start database
docker-compose up -d postgres

# Wait for database
echo "Waiting for database..."
until docker-compose exec -T postgres pg_isready -U byteroom; do
  sleep 1
done

# Run migrations
cd backend && make migrate-up && cd ..

# Install backend dependencies
cd backend && go mod download && cd ..

# Install frontend dependencies
cd frontend && npm install && cd ..

echo "✅ Setup complete! Run 'scripts/dev.sh' to start development."
```

**Test Requirements**:
```bash
# Verify setup script runs
./scripts/setup.sh

# Verify dev script starts services
./scripts/dev.sh
```

**Exit Criteria**:
- [ ] New developer can set up project with single script
- [ ] README contains clear setup instructions
- [ ] All environment variables documented

---

## Phase 1 Exit Criteria Summary

### Automated Verification

```bash
# Run all verification commands
./scripts/verify-phase-1.sh
```

| Check | Command | Expected Result |
|-------|---------|-----------------|
| Backend compiles | `cd backend && go build ./...` | Exit 0 |
| Frontend builds | `cd frontend && npm run build` | Exit 0 |
| Docker starts | `docker-compose up -d && docker-compose ps` | All healthy |
| Migrations run | `make migrate-up` | Exit 0 |
| Tests execute | `go test ./... && npm test` | Exit 0 |

### Manual Verification

- [ ] Project structure matches specification
- [ ] CI/CD pipeline triggers on PR
- [ ] Developer can onboard using README

### Deliverables

1. ✅ Backend Go project with clean architecture structure
2. ✅ Frontend React/TypeScript project with Vite
3. ✅ Docker Compose development environment
4. ✅ Database migrations for all entities
5. ✅ GitHub Actions CI/CD pipeline
6. ✅ Developer setup documentation

---

## Next Phase

Upon completion, proceed to [Phase 2: Backend Core](./phase-2-backend-core.md)
