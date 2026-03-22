#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

PASS=0
FAIL=0

check() {
  local desc="$1"
  shift
  if "$@" >/dev/null 2>&1; then
    echo "  PASS  $desc"
    PASS=$((PASS + 1))
  else
    echo "  FAIL  $desc"
    FAIL=$((FAIL + 1))
  fi
}

echo "=== Phase 1 Verification ==="
echo ""

echo "[ Backend ]"
check "go build ./..." bash -c "cd $ROOT_DIR/backend && go build ./..."
check "go tests pass"  bash -c "cd $ROOT_DIR/backend && JWT_SECRET=verify go test ./... -short"

echo ""
echo "[ Frontend ]"
check "npm run build"  bash -c "cd $ROOT_DIR/frontend && npm run build"
check "npm run test"   bash -c "cd $ROOT_DIR/frontend && npm run test"

echo ""
echo "[ Files ]"
check "docker-compose.yml exists"     test -f "$ROOT_DIR/docker-compose.yml"
check "migrations directory exists"   test -d "$ROOT_DIR/backend/migrations"
check "migration 000001 exists"       test -f "$ROOT_DIR/backend/migrations/000001_create_users_table.up.sql"
check "migration 000002 exists"       test -f "$ROOT_DIR/backend/migrations/000002_create_chats_table.up.sql"
check "migration 000003 exists"       test -f "$ROOT_DIR/backend/migrations/000003_create_messages_table.up.sql"
check "CI workflow exists"            test -f "$ROOT_DIR/.github/workflows/test.yml"
check "backend .env.example exists"   test -f "$ROOT_DIR/backend/.env.example"

echo ""
echo "=== Results: $PASS passed, $FAIL failed ==="
[ "$FAIL" -eq 0 ] || exit 1
