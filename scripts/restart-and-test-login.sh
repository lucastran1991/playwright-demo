#!/usr/bin/env bash
# Free ports 8080/3000, start backend + frontend, verify login (API + NextAuth providers).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
EMAIL="${TEST_EMAIL:-developer@test.com}"
PASS="${TEST_PASSWORD:-password123}"

for p in 8080 3000; do
  ids=$(lsof -ti tcp:"$p" 2>/dev/null || true)
  if [[ -n "${ids}" ]]; then
    kill $ids 2>/dev/null || true
  fi
done
sleep 1

export BLUEPRINT_DIR="$BACKEND/testdata/blueprint"
(cd "$BACKEND" && go run ./cmd/server) &
BACK_PID=$!
FRONT_PID=""
die() {
  kill "$BACK_PID" 2>/dev/null || true
  [[ -n "$FRONT_PID" ]] && kill "$FRONT_PID" 2>/dev/null || true
  exit 1
}
trap 'kill "$BACK_PID" 2>/dev/null; [[ -n "$FRONT_PID" ]] && kill "$FRONT_PID" 2>/dev/null; exit 130' INT TERM

for _ in $(seq 1 60); do
  curl -sf http://127.0.0.1:8080/health >/dev/null 2>&1 && break
  sleep 0.5
done
curl -sf http://127.0.0.1:8080/health >/dev/null || { echo "Backend failed to start"; die; }
echo "Backend OK http://127.0.0.1:8080"

(cd "$FRONTEND" && pnpm dev) &
FRONT_PID=$!
for _ in $(seq 1 90); do
  curl -sf http://127.0.0.1:3000/ >/dev/null 2>&1 && break
  sleep 0.5
done
curl -sf http://127.0.0.1:3000/ >/dev/null || { echo "Frontend failed to start"; die; }
echo "Frontend OK http://127.0.0.1:3000"

echo ""
echo "=== POST /api/auth/login (Go API) ==="
RESP=$(curl -sS -w "\n%{http_code}" -X POST http://127.0.0.1:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d "$(printf '{"email":"%s","password":"%s"}' "$EMAIL" "$PASS")")
CODE=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
echo "HTTP $CODE"
if [[ "$CODE" != "200" ]]; then
  echo "$BODY"
  die
fi
python3 -c "import json,sys; d=json.loads(sys.stdin.read()); assert d.get('access_token'); print('access_token: present'); print('user:', d.get('user',{}).get('email'))" <<<"$BODY"

echo ""
echo "=== NextAuth providers (Next.js) ==="
curl -sS http://127.0.0.1:3000/api/auth/providers | head -c 400
echo ""

echo ""
echo "Login test passed for $EMAIL"
echo "Servers running: backend PID $BACK_PID, frontend PID $FRONT_PID"
echo "Stop: kill $BACK_PID $FRONT_PID"
trap - INT TERM
