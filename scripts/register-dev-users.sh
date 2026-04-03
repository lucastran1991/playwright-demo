#!/usr/bin/env bash
# Register multiple local dev accounts via POST /api/auth/register.
# Requires backend on BASE (default http://127.0.0.1:8080). Safe to re-run (409 = already exists).
set -euo pipefail

BASE="${BASE:-http://127.0.0.1:8080}"
PASS="${DEV_PASSWORD:-password123}"

register_one() {
  local name="$1" email="$2"
  local code body
  body=$(curl -sS -w '\n%{http_code}' -X POST "${BASE}/api/auth/register" \
    -H "Content-Type: application/json" \
    -d "$(printf '{"name":"%s","email":"%s","password":"%s"}' "$name" "$email" "$PASS")")
  code=$(echo "$body" | tail -n1)
  case "$code" in
    201) echo "OK  created  $email" ;;
    409) echo "SKIP exists  $email" ;;
    *)   echo "FAIL $code  $email — $(echo "$body" | sed '$d' | head -c 200)"; return 1 ;;
  esac
}

echo "Registering dev users against ${BASE} ..."
for i in $(seq 1 30); do
  if curl -sf "${BASE}/health" >/dev/null 2>&1; then break; fi
  sleep 0.5
done
if ! curl -sf "${BASE}/health" >/dev/null 2>&1; then
  echo "Backend not reachable at ${BASE}. Start it: cd backend && go run ./cmd/server"
  exit 1
fi

register_one "Admin" "admin@test.com"
register_one "Developer" "developer@test.com"
register_one "Viewer" "viewer@test.com"
register_one "QA User" "qa@test.com"

echo ""
echo "Default password for all: ${PASS}"
echo "Log in at http://localhost:3000 with any email above."
