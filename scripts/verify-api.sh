#!/usr/bin/env bash
# Run from repo root. Starts Postgres (brew), backend, hits APIs, stops backend.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
BACKEND="$ROOT/backend"
PORT="${SERVER_PORT:-8080}"
BASE="http://127.0.0.1:${PORT}"

if lsof -ti tcp:"$PORT" >/dev/null 2>&1; then
  echo "Port $PORT in use — stopping stale process so this script can bind."
  lsof -ti tcp:"$PORT" | xargs kill 2>/dev/null || true
  sleep 1
fi

need_file() {
  local f="$1"
  if [[ ! -f "$f" ]]; then
    echo "Missing $f — copy from example and configure."
    exit 1
  fi
}

need_file "$BACKEND/.env"
need_file "$ROOT/frontend/.env.local"

# Sample CSV domains live under testdata (default ./blueprint/Node & Edge is often missing in dev).
export BLUEPRINT_DIR="$BACKEND/testdata/blueprint"

if command -v brew >/dev/null 2>&1; then
  brew services start postgresql@15 2>/dev/null || brew services start postgresql@14 2>/dev/null || true
  sleep 2
fi
createdb app_dev 2>/dev/null || true

echo "=== Starting Go server ==="
(cd "$BACKEND" && go run ./cmd/server) &
SRV_PID=$!
cleanup() { kill "$SRV_PID" 2>/dev/null || true; }
trap cleanup EXIT

for i in $(seq 1 40); do
  if curl -sf "$BASE/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.5
done

echo "=== GET /health ==="
curl -sS "$BASE/health" | head -c 500
echo ""

echo "=== GET /api/blueprints/types (public) ==="
curl -sS "$BASE/api/blueprints/types" | head -c 800
echo ""

echo "=== Register (idempotent may 409) ==="
REG=$(curl -sS -w '\n%{http_code}' -X POST "$BASE/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"name":"ApiVerify","email":"api-verify@test.com","password":"password123"}')
HTTP=$(echo "$REG" | tail -n1)
BODY=$(echo "$REG" | sed '$d')
echo "HTTP $HTTP"
echo "$BODY" | head -c 400
echo ""

extract_token() {
  python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('access_token') or (d.get('data') or {}).get('access_token',''))" 2>/dev/null || true
}

TOKEN=""
if echo "$BODY" | grep -q 'access_token'; then
  TOKEN=$(echo "$BODY" | extract_token)
fi

if [[ -z "$TOKEN" ]]; then
  echo "=== Login ==="
  LOG=$(curl -sS -w '\n%{http_code}' -X POST "$BASE/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"api-verify@test.com","password":"password123"}')
  LHTTP=$(echo "$LOG" | tail -n1)
  LBODY=$(echo "$LOG" | sed '$d')
  echo "HTTP $LHTTP"
  TOKEN=$(echo "$LBODY" | extract_token)
fi

if [[ -n "$TOKEN" ]]; then
  echo "=== GET /api/auth/me ==="
  curl -sS -w '\nHTTP %{http_code}\n' "$BASE/api/auth/me" \
    -H "Authorization: Bearer $TOKEN" | head -c 500
  echo ""
  echo "=== POST /api/blueprints/ingest ==="
  curl -sS -w '\nHTTP %{http_code}\n' -X POST "$BASE/api/blueprints/ingest" \
    -H "Authorization: Bearer $TOKEN" | head -c 1200
  echo ""
  echo "=== GET /api/blueprints/types (after ingest) ==="
  curl -sS "$BASE/api/blueprints/types" | head -c 800
  echo ""
  echo "=== GET /api/blueprints/tree/testdomain (testdata domain) ==="
  curl -sS -w '\nHTTP %{http_code}\n' "$BASE/api/blueprints/tree/testdomain" | head -c 1200
  echo ""
else
  echo "No access token; skip ingest."
fi

echo "=== Done ==="
