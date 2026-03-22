.PHONY: help dev-up dev-down test test-backend test-frontend test-e2e test-coverage \
        lint lint-backend lint-frontend build build-backend build-frontend \
        migrate-up migrate-down docker-prod-up docker-prod-down bench clean

# ── Defaults ──────────────────────────────────────────────────────────────────
DOCKER_COMPOSE    ?= docker-compose
BACKEND_DIR       := ./backend
FRONTEND_DIR      := ./frontend
GO_TEST_FLAGS     := -v -race -count=1

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n",$$1,$$2}'

# ── Development ───────────────────────────────────────────────────────────────
dev-up: ## Start development infrastructure (postgres)
	$(DOCKER_COMPOSE) up -d postgres

dev-down: ## Stop development infrastructure
	$(DOCKER_COMPOSE) down

# ── Database ──────────────────────────────────────────────────────────────────
migrate-up: ## Run all pending migrations
	cd $(BACKEND_DIR) && migrate -path migrations \
		-database "$$(go run ./cmd/server -print-dsn 2>/dev/null || echo $$DATABASE_URL)" up

migrate-down: ## Roll back the last migration
	cd $(BACKEND_DIR) && migrate -path migrations \
		-database "$$(go run ./cmd/server -print-dsn 2>/dev/null || echo $$DATABASE_URL)" down 1

# ── Testing ───────────────────────────────────────────────────────────────────
test: test-backend test-frontend ## Run all unit tests

test-backend: ## Run Go unit tests with race detector
	cd $(BACKEND_DIR) && go test $(GO_TEST_FLAGS) ./...

test-frontend: ## Run Vitest unit tests
	cd $(FRONTEND_DIR) && npm test

test-coverage: ## Run all tests with coverage reports
	cd $(BACKEND_DIR) && go test -race -coverprofile=coverage.out ./... && \
		go tool cover -html=coverage.out -o coverage.html
	cd $(FRONTEND_DIR) && npm run test:coverage

test-e2e: ## Run Playwright E2E tests (requires running backend + frontend)
	cd $(FRONTEND_DIR) && npm run test:e2e

test-backend-bench: ## Run Go benchmarks
	cd $(BACKEND_DIR) && go test ./... -run='^$$' -bench=. -benchtime=3s

# ── Linting ───────────────────────────────────────────────────────────────────
lint: lint-backend lint-frontend ## Run all linters

lint-backend: ## Run golangci-lint on backend
	cd $(BACKEND_DIR) && golangci-lint run ./...

lint-frontend: ## Run ESLint on frontend
	cd $(FRONTEND_DIR) && npm run lint

# ── Building ──────────────────────────────────────────────────────────────────
build: build-backend build-frontend ## Build all artifacts

build-backend: ## Build Go binary
	cd $(BACKEND_DIR) && CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server

build-frontend: ## Build frontend bundle
	cd $(FRONTEND_DIR) && npm run build

# ── Docker Production ─────────────────────────────────────────────────────────
docker-prod-build: ## Build production Docker images
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml build

docker-prod-up: ## Start production stack
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml up -d

docker-prod-down: ## Stop production stack
	$(DOCKER_COMPOSE) -f docker-compose.prod.yml down

# ── Utilities ─────────────────────────────────────────────────────────────────
clean: ## Remove build artifacts
	rm -f $(BACKEND_DIR)/bin/server $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/playwright-report

install-tools: ## Install development tools (golangci-lint, migrate, playwright)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	cd $(FRONTEND_DIR) && npx playwright install chromium
