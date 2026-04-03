#!/usr/bin/env bash
# Local dev: PostgreSQL (optional), Go API, Next.js app.
# Usage:
#   ./start.sh                 # start backend + frontend (assumes Postgres already running)
#   ./start.sh --with-postgres # also: brew services start postgresql@15 + createdb app_dev

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"

WITH_POSTGRES=0
for arg in "$@"; do
  case "$arg" in
    --with-postgres) WITH_POSTGRES=1 ;;
    -h|--help)
      echo "Usage: $0 [--with-postgres]"
      echo "  --with-postgres  Try: brew services start postgresql@15 (or @14) and createdb app_dev"
      exit 0
      ;;
  esac
done

if [[ "$WITH_POSTGRES" -eq 1 ]]; then
  if command -v brew >/dev/null 2>&1; then
    brew services start postgresql@15 2>/dev/null || brew services start postgresql@14 2>/dev/null || {
      echo "Could not start PostgreSQL via Homebrew. Start it manually, then re-run without --with-postgres."
      exit 1
    }
    sleep 2
  else
    echo "Homebrew not found; start PostgreSQL yourself, then run without --with-postgres."
    exit 1
  fi
  createdb app_dev 2>/dev/null || true
fi

if [[ ! -f "$BACKEND/.env" ]]; then
  echo "Missing backend/.env — copy from .env.example and set DB_*, JWT_SECRET:"
  echo "  cp $BACKEND/.env.example $BACKEND/.env"
  exit 1
fi

if [[ ! -f "$FRONTEND/.env.local" ]]; then
  echo "Missing frontend/.env.local — copy from .env.example:"
  echo "  cp $FRONTEND/.env.example $FRONTEND/.env.local"
  exit 1
fi

if [[ ! -d "$FRONTEND/node_modules" ]]; then
  (cd "$FRONTEND" && pnpm install)
fi

PIDS=()
cleanup() {
  for pid in "${PIDS[@]:-}"; do
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT INT TERM

echo "Starting backend http://localhost:8080 ..."
# Repo ships sample CSVs under backend/testdata/blueprint (override in backend/.env if needed).
export BLUEPRINT_DIR="$ROOT/blueprint/Node & Edge"
(cd "$BACKEND" && go run ./cmd/server) &
PIDS+=($!)

echo "Starting frontend http://localhost:3000 ..."
(cd "$FRONTEND" && pnpm dev) &
PIDS+=($!)

echo ""
echo "Servers running. Ctrl+C stops both."
echo ""
echo "--- Ingest blueprint (after register/login) ---"
echo "Register (one user):"
echo '  curl -sS -X POST http://localhost:8080/api/auth/register \'
echo '    -H "Content-Type: application/json" \'
echo '    -d '"'"'{"name":"Admin","email":"admin@test.com","password":"password123"}'"'"
echo ""
echo "Or register several dev accounts: bash scripts/register-dev-users.sh"
echo ""
echo "Ingest (replace ACCESS_TOKEN):"
echo '  curl -sS -X POST http://localhost:8080/api/blueprints/ingest \'
echo '    -H "Authorization: Bearer ACCESS_TOKEN"'
echo ""
echo "Verify:"
echo "  curl -sS http://localhost:8080/api/blueprints/types"
echo "  curl -sS http://localhost:8080/api/blueprints/tree/cooling-system"
echo ""
echo "If types are empty, ingest after login:"
echo "  bash scripts/ensure-blueprint-ingested.sh"
echo ""

# Block until a child exits (macOS bash 3.2 has no wait -n)
wait
