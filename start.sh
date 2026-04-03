#!/usr/bin/env bash
# Local dev with PM2 process manager.
# Usage:
#   ./start.sh                 # start backend + frontend via PM2
#   ./start.sh --with-postgres # also start PostgreSQL via Homebrew
#   ./start.sh stop            # stop all services
#   ./start.sh restart         # restart all services
#   ./start.sh logs            # tail PM2 logs
#   ./start.sh status          # show PM2 process list

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
CONFIG="$ROOT/ecosystem.config.cjs"

# Parse arguments
ACTION="start"
WITH_POSTGRES=0
for arg in "$@"; do
  case "$arg" in
    stop|restart|logs|status) ACTION="$arg" ;;
    --with-postgres) WITH_POSTGRES=1 ;;
    -h|--help)
      echo "Usage: $0 [start|stop|restart|logs|status] [--with-postgres]"
      echo "  start            Start backend + frontend via PM2 (default)"
      echo "  stop             Stop all PM2 services"
      echo "  restart          Restart all PM2 services"
      echo "  logs             Tail PM2 logs"
      echo "  status           Show PM2 process list"
      echo "  --with-postgres  Also start PostgreSQL via Homebrew"
      exit 0
      ;;
  esac
done

# Handle non-start actions
case "$ACTION" in
  stop)    pm2 stop "$CONFIG" 2>/dev/null; pm2 delete all 2>/dev/null || true; echo "Stopped."; exit 0 ;;
  restart) pm2 flush 2>/dev/null || true; pm2 restart "$CONFIG"; echo "Restarted."; exit 0 ;;
  logs)    pm2 logs; exit 0 ;;
  status)  pm2 list; exit 0 ;;
esac

# --- Start flow ---

# PostgreSQL (optional)
if [[ "$WITH_POSTGRES" -eq 1 ]]; then
  if command -v brew >/dev/null 2>&1; then
    brew services start postgresql@15 2>/dev/null || brew services start postgresql@14 2>/dev/null || {
      echo "Could not start PostgreSQL via Homebrew. Start it manually."
      exit 1
    }
    sleep 2
  else
    echo "Homebrew not found; start PostgreSQL yourself."
    exit 1
  fi
  createdb app_dev 2>/dev/null || true
fi

# Validate env files exist
if [[ ! -f "$BACKEND/.env" ]]; then
  echo "Missing backend/.env — copy from .env.example:"
  echo "  cp $BACKEND/.env.example $BACKEND/.env"
  exit 1
fi

if [[ ! -f "$FRONTEND/.env.local" ]]; then
  echo "Missing frontend/.env.local — copy from .env.example:"
  echo "  cp $FRONTEND/.env.example $FRONTEND/.env.local"
  exit 1
fi

# Sync ports/URLs from system.cfg.json into .env files (single source of truth)
SYSCFG="$ROOT/system.cfg.json"
if command -v python3 >/dev/null 2>&1 && [[ -f "$SYSCFG" ]]; then
  eval "$(python3 -c "
import json
cfg = json.load(open('$SYSCFG'))
be = cfg['backend']
fe = cfg['frontend']
print(f'BE_PORT={be[\"port\"]}')
print(f'FE_PORT={fe[\"port\"]}')
print(f'BE_URL={be[\"url\"]}')
print(f'FE_URL={fe[\"url\"]}')
")"
  # Update backend SERVER_PORT
  sed -i '' "s|^SERVER_PORT=.*|SERVER_PORT=$BE_PORT|" "$BACKEND/.env"
  # Update frontend API URL and AUTH_URL
  sed -i '' "s|^NEXT_PUBLIC_API_URL=.*|NEXT_PUBLIC_API_URL=$BE_URL|" "$FRONTEND/.env.local"
  sed -i '' "s|^AUTH_URL=.*|AUTH_URL=$FE_URL|" "$FRONTEND/.env.local"
  echo "Synced ports from system.cfg.json (backend:$BE_PORT, frontend:$FE_PORT)"
fi

# Install frontend deps if needed
if [[ ! -d "$FRONTEND/node_modules" ]]; then
  (cd "$FRONTEND" && pnpm install)
fi

# Stop existing PM2 processes and flush logs
pm2 delete all 2>/dev/null || true
pm2 flush 2>/dev/null || true

# Start via PM2
pm2 start "$CONFIG"
echo ""
pm2 list
echo ""
echo "Services running via PM2."
echo "  Logs:    ./start.sh logs"
echo "  Stop:    ./start.sh stop"
echo "  Restart: ./start.sh restart"
echo "  Status:  ./start.sh status"
echo ""
echo "Backend:  ${BE_URL:-http://localhost:8889}"
echo "Frontend: ${FE_URL:-http://localhost:8089}"
echo "Tracer:   ${FE_URL:-http://localhost:8089}/tracer"
echo ""
echo "--- Quick setup (first time) ---"
echo "  bash scripts/register-dev-users.sh"
echo "  bash scripts/ensure-blueprint-ingested.sh"
