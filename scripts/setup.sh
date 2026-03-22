#!/usr/bin/env bash
set -euo pipefail

echo "Setting up ByteRoom development environment..."

# Check prerequisites
command -v go    >/dev/null 2>&1 || { echo "ERROR: Go is required but not installed."; exit 1; }
command -v node  >/dev/null 2>&1 || { echo "ERROR: Node.js is required but not installed."; exit 1; }
command -v docker>/dev/null 2>&1 || { echo "ERROR: Docker is required but not installed."; exit 1; }
command -v npm   >/dev/null 2>&1 || { echo "ERROR: npm is required but not installed."; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Copy env file if not present
if [ ! -f "$ROOT_DIR/backend/.env" ]; then
  cp "$ROOT_DIR/backend/.env.example" "$ROOT_DIR/backend/.env"
  echo "Created backend/.env from .env.example — please update JWT_SECRET"
fi

# Start database
echo "Starting PostgreSQL..."
docker compose -f "$ROOT_DIR/docker-compose.yml" up -d postgres

# Wait for database readiness
echo "Waiting for database to be healthy..."
until docker compose -f "$ROOT_DIR/docker-compose.yml" exec -T postgres pg_isready -U byteroom; do
  sleep 1
done

# Run migrations
echo "Running database migrations..."
cd "$ROOT_DIR/backend"
if command -v migrate >/dev/null 2>&1; then
  make migrate-up
else
  echo "WARNING: golang-migrate not found. Install it with:"
  echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
fi

# Install backend dependencies
echo "Downloading Go modules..."
cd "$ROOT_DIR/backend"
go mod download

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd "$ROOT_DIR/frontend"
npm install

echo ""
echo "Setup complete! Run scripts/dev.sh to start development servers."
