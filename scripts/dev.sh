#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Ensure database is running
docker compose -f "$ROOT_DIR/docker-compose.yml" up -d postgres

# Load backend env
if [ -f "$ROOT_DIR/backend/.env" ]; then
  set -a && source "$ROOT_DIR/backend/.env" && set +a
fi

echo "Starting ByteRoom development servers..."
echo "  Backend  → http://localhost:8080"
echo "  Frontend → http://localhost:5173"
echo ""
echo "Press Ctrl+C to stop all servers."

# Start backend and frontend in parallel
(cd "$ROOT_DIR/backend" && go run ./cmd/server) &
BACKEND_PID=$!

(cd "$ROOT_DIR/frontend" && npm run dev) &
FRONTEND_PID=$!

cleanup() {
  echo ""
  echo "Stopping servers..."
  kill "$BACKEND_PID" "$FRONTEND_PID" 2>/dev/null || true
}
trap cleanup INT TERM

wait "$BACKEND_PID" "$FRONTEND_PID"
