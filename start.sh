#!/usr/bin/env bash
# System startup with PM2 process manager.
# Usage:
#   ./start.sh                 # start backend + frontend (dev mode)
#   ./start.sh --prod          # production mode (build binary + next build)
#   ./start.sh --with-postgres # also start PostgreSQL via Homebrew
#   ./start.sh --ingest        # auto-ingest blueprint + model data after start
#   ./start.sh stop            # stop all services
#   ./start.sh restart         # restart all services
#   ./start.sh logs            # tail PM2 logs
#   ./start.sh status          # show PM2 process list

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
CONFIG="$ROOT/ecosystem.config.cjs"
SYSCFG="$ROOT/system.cfg.json"

# Parse arguments
ACTION="start"
WITH_POSTGRES=0
PROD_MODE=0
AUTO_INGEST=0
for arg in "$@"; do
  case "$arg" in
    stop|restart|logs|status) ACTION="$arg" ;;
    --with-postgres) WITH_POSTGRES=1 ;;
    --prod) PROD_MODE=1 ;;
    --ingest) AUTO_INGEST=1 ;;
    -h|--help)
      echo "Usage: $0 [start|stop|restart|logs|status] [--prod] [--with-postgres] [--ingest]"
      echo "  start            Start backend + frontend via PM2 (default: dev mode)"
      echo "  stop             Stop all PM2 services"
      echo "  restart          Restart all PM2 services"
      echo "  logs             Tail PM2 logs"
      echo "  status           Show PM2 process list"
      echo "  --prod           Production mode: compile Go binary + next build"
      echo "  --with-postgres  Also start PostgreSQL via Homebrew"
      echo "  --ingest         Auto-ingest blueprint + model data after start"
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

# Read config from system.cfg.json
BE_PORT=8889; FE_PORT=8089; BE_URL="http://localhost:8889"; FE_URL="http://localhost:8089"
if command -v python3 >/dev/null 2>&1 && [[ -f "$SYSCFG" ]]; then
  eval "$(python3 -c "
import json
cfg = json.load(open('$SYSCFG'))
be, fe = cfg['backend'], cfg['frontend']
print(f'BE_PORT={be[\"port\"]}')
print(f'FE_PORT={fe[\"port\"]}')
print(f'BE_URL={be[\"url\"]}')
print(f'FE_URL={fe[\"url\"]}')
")"
fi

# PostgreSQL (optional)
if [[ "$WITH_POSTGRES" -eq 1 ]]; then
  if command -v systemctl >/dev/null 2>&1; then
    # Linux (EC2, Ubuntu, Amazon Linux, etc.)
    sudo systemctl start postgresql 2>/dev/null \
      || sudo systemctl start postgresql-15 2>/dev/null \
      || sudo systemctl start postgresql-14 2>/dev/null \
      || { echo "Could not start PostgreSQL via systemctl. Start it manually."; exit 1; }
    sleep 2
  elif command -v brew >/dev/null 2>&1; then
    # macOS via Homebrew
    brew services start postgresql@15 2>/dev/null || brew services start postgresql@14 2>/dev/null || {
      echo "Could not start PostgreSQL via Homebrew. Start it manually."
      exit 1
    }
    sleep 2
  else
    echo "No supported service manager found. Start PostgreSQL manually."
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

# Cross-platform sed in-place (macOS uses -i '', GNU/Linux uses -i)
sedi() { if [[ "$OSTYPE" == darwin* ]]; then sed -i '' "$@"; else sed -i "$@"; fi; }

# Sync config from system.cfg.json into .env files
sedi "s|^SERVER_PORT=.*|SERVER_PORT=$BE_PORT|" "$BACKEND/.env"
grep -q "^CORS_ORIGIN=" "$BACKEND/.env" \
  && sedi "s|^CORS_ORIGIN=.*|CORS_ORIGIN=$FE_URL|" "$BACKEND/.env" \
  || echo "CORS_ORIGIN=$FE_URL" >> "$BACKEND/.env"
sedi "s|^NEXT_PUBLIC_API_URL=.*|NEXT_PUBLIC_API_URL=$BE_URL|" "$FRONTEND/.env.local"
sedi "s|^AUTH_URL=.*|AUTH_URL=$FE_URL|" "$FRONTEND/.env.local"
echo "Config synced from system.cfg.json (backend:$BE_PORT, frontend:$FE_PORT)"

# Detect package manager (pnpm > npm)
if command -v pnpm >/dev/null 2>&1; then
  PKG="pnpm"
else
  PKG="npm"
fi

# Install frontend deps if needed
if [[ ! -d "$FRONTEND/node_modules" ]]; then
  (cd "$FRONTEND" && $PKG install)
fi

# Production build
if [[ "$PROD_MODE" -eq 1 ]]; then
  echo "Building backend binary..."
  (cd "$BACKEND" && go build -o server ./cmd/server)
  echo "Building frontend..."
  (cd "$FRONTEND" && $PKG run build)
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
echo "Backend:  $BE_URL"
echo "Frontend: $FE_URL"
echo "Tracer:   $FE_URL/tracer"

# Auto-ingest on first run
if [[ "$AUTO_INGEST" -eq 1 ]]; then
  echo ""
  echo "Waiting for backend to start..."
  sleep 5
  echo "Ingesting blueprints..."
  curl -s -X POST "$BE_URL/api/blueprints/ingest" | python3 -m json.tool
  echo "Ingesting models..."
  curl -s -X POST "$BE_URL/api/models/ingest" | python3 -m json.tool
  echo "Data ingestion complete."
fi
