#!/usr/bin/env bash
# If GET /api/blueprints/types is empty, POST /api/blueprints/ingest (requires auth).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BASE="${BASE:-http://127.0.0.1:8080}"
EMAIL="${INGEST_LOGIN_EMAIL:-developer@test.com}"
PASS="${INGEST_LOGIN_PASSWORD:-password123}"

count_types() {
  curl -sS "${BASE}/api/blueprints/types" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('data') or []))"
}

wait_health() {
  for _ in $(seq 1 60); do
    curl -sf "${BASE}/health" >/dev/null 2>&1 && return 0
    sleep 0.5
  done
  return 1
}

if ! wait_health; then
  echo "Backend not running at ${BASE}. Start it first (e.g. ./start.sh or go run ./cmd/server from backend with BLUEPRINT_DIR)."
  exit 1
fi

N=$(count_types)
echo "Blueprint types in DB: ${N}"

if [[ "${N}" != "0" ]]; then
  echo "Already ingested; skipping."
  curl -sS "${BASE}/api/blueprints/types" | head -c 600
  echo ""
  exit 0
fi

echo "No blueprint data — logging in and ingesting ..."
LOG=$(curl -sS -X POST "${BASE}/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "$(printf '{"email":"%s","password":"%s"}' "$EMAIL" "$PASS")")
TOKEN=$(echo "$LOG" | python3 -c "import sys,json; print(json.load(sys.stdin).get('access_token',''))" 2>/dev/null || true)
if [[ -z "$TOKEN" ]]; then
  echo "Login failed. Response:"
  echo "$LOG" | head -c 500
  echo ""
  exit 1
fi

ING=$(curl -sS -w "\n%{http_code}" -X POST "${BASE}/api/blueprints/ingest" \
  -H "Authorization: Bearer ${TOKEN}")
CODE=$(echo "$ING" | tail -n1)
BODY=$(echo "$ING" | sed '$d')
echo "Ingest HTTP ${CODE}"
echo "$BODY" | head -c 800
echo ""

if [[ "$CODE" != "200" ]]; then
  exit 1
fi

N2=$(count_types)
echo "Blueprint types after ingest: ${N2}"
