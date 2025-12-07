#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEV_DIR="$ROOT_DIR/.dev"
MONGODB_DATA="$DEV_DIR/mongodb"
MONGODB_PORT="${MONGODB_PORT:-27017}"
MONGODB_PID_FILE="$DEV_DIR/mongodb.pid"
REDIS_PORT="${REDIS_PORT:-6381}"
REDIS_PID_FILE="$DEV_DIR/redis.pid"

log() {
  echo "[dev-services] $*"
}

ensure_dirs() {
  mkdir -p "$DEV_DIR"
}

start_mongodb() {
  ensure_dirs
  mkdir -p "$MONGODB_DATA"

  if [ -f "$MONGODB_PID_FILE" ] && kill -0 "$(cat "$MONGODB_PID_FILE")" >/dev/null 2>&1; then
    log "MongoDB already running"
    return
  fi

  if ! command -v mongod >/dev/null 2>&1; then
    log "Error: mongod not found. Please install MongoDB:"
    log "  - Arch/CachyOS: sudo pacman -S mongodb"
    log "  - Ubuntu/Debian: sudo apt install mongodb"
    log "  - macOS: brew install mongodb-community"
    exit 1
  fi

  log "Starting MongoDB on port $MONGODB_PORT"
  mongod --dbpath "$MONGODB_DATA" \
    --port "$MONGODB_PORT" \
    --bind_ip 127.0.0.1 \
    --logpath "$DEV_DIR/mongodb.log" \
    --pidfilepath "$MONGODB_PID_FILE" \
    --fork >/dev/null 2>&1

  # Wait for MongoDB to be ready
  local max_attempts=30
  local attempt=0
  while [ $attempt -lt $max_attempts ]; do
    # Try to connect using mongosh or mongo (depending on version)
    if command -v mongosh >/dev/null 2>&1; then
      if mongosh --quiet --eval "db.adminCommand('ping')" "mongodb://127.0.0.1:$MONGODB_PORT" >/dev/null 2>&1; then
        log "MongoDB is ready"
        return
      fi
    elif command -v mongo >/dev/null 2>&1; then
      if mongo --quiet --eval "db.adminCommand('ping')" "mongodb://127.0.0.1:$MONGODB_PORT" >/dev/null 2>&1; then
        log "MongoDB is ready"
        return
      fi
    else
      # If neither mongosh nor mongo is available, just wait a bit and assume it's ready
      sleep 2
      log "MongoDB started (mongosh/mongo not available for health check)"
      return
    fi
    attempt=$((attempt + 1))
    sleep 1
  done

  log "Warning: MongoDB may not be fully ready yet"
}

stop_mongodb() {
  if [ -f "$MONGODB_PID_FILE" ]; then
    if kill -0 "$(cat "$MONGODB_PID_FILE")" >/dev/null 2>&1; then
      log "Stopping MongoDB"
      kill "$(cat "$MONGODB_PID_FILE")" || true
      # Wait for MongoDB to stop
      local max_attempts=10
      local attempt=0
      while [ $attempt -lt $max_attempts ]; do
        if ! kill -0 "$(cat "$MONGODB_PID_FILE")" >/dev/null 2>&1; then
          break
        fi
        attempt=$((attempt + 1))
        sleep 1
      done
    fi
    rm -f "$MONGODB_PID_FILE"
  fi
}

status_mongodb() {
  if [ -f "$MONGODB_PID_FILE" ] && kill -0 "$(cat "$MONGODB_PID_FILE")" >/dev/null 2>&1; then
    log "MongoDB running (port $MONGODB_PORT)"
  else
    log "MongoDB not running"
  fi
}

start_redis() {
  ensure_dirs
  if [ -f "$REDIS_PID_FILE" ] && kill -0 "$(cat "$REDIS_PID_FILE")" >/dev/null 2>&1; then
    log "Redis already running"
    return
  fi
  log "Starting redis on port $REDIS_PORT"
  redis-server --port "$REDIS_PORT" \
    --save "" \
    --appendonly no \
    --dir "$DEV_DIR" \
    --pidfile "$REDIS_PID_FILE" \
    --logfile "$DEV_DIR/redis.log" \
    --daemonize yes
}

stop_redis() {
  if [ -f "$REDIS_PID_FILE" ]; then
    if kill -0 "$(cat "$REDIS_PID_FILE")" >/dev/null 2>&1; then
      log "Stopping redis"
      kill "$(cat "$REDIS_PID_FILE")" || true
    fi
    rm -f "$REDIS_PID_FILE"
  fi
}

status_redis() {
  if [ -f "$REDIS_PID_FILE" ] && kill -0 "$(cat "$REDIS_PID_FILE")" >/dev/null 2>&1; then
    log "Redis running (port $REDIS_PORT)"
  else
    log "Redis not running"
  fi
}

case "${1:-}" in
  up)
    start_mongodb
    start_redis
    log "MongoDB and Redis started"
    ;;
  down)
    stop_redis
    stop_mongodb
    log "MongoDB and Redis stopped"
    ;;
  status)
    status_mongodb
    status_redis
    ;;
  *)
    echo "Usage: $0 {up|down|status}"
    exit 1
    ;;
esac

