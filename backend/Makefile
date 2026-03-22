.PHONY: build test test-coverage lint clean migrate-up migrate-down migrate-create run

# Build
build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

# Tests
test:
	go test ./... -short -race

test-integration:
	go test ./... -tags=integration -race

test-coverage:
	go test ./... -short -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1

test-coverage-html:
	go test ./... -short -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Database migrations
DB_URL ?= postgres://byteroom:byteroom_dev@localhost:5432/byteroom?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Clean
clean:
	rm -rf bin/ coverage.out coverage.html
