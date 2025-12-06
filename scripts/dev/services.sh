#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEV_DIR="$ROOT_DIR/.dev"
PGDATA="$DEV_DIR/postgres"
PGSOCKET_DIR="$DEV_DIR/pgsocket"
PGPORT="${PGPORT:-5544}"
REDIS_PORT="${REDIS_PORT:-6381}"
REDIS_PID_FILE="$DEV_DIR/redis.pid"

log() {
  echo "[dev-services] $*"
}

ensure_dirs() {
  mkdir -p "$DEV_DIR"
}

start_postgres() {
  ensure_dirs
  mkdir -p "$PGSOCKET_DIR"
  local current_major
  current_major="$(pg_ctl --version | awk '{print $NF}' | cut -d. -f1)"

  if [ -d "$PGDATA" ] && [ -f "$PGDATA/PG_VERSION" ]; then
    local data_major
    data_major="$(cat "$PGDATA/PG_VERSION" | cut -d. -f1)"
    if [ "$data_major" != "$current_major" ]; then
      log "Detected postgres major version change ($data_major -> $current_major); reinitializing data directory"
      if pg_ctl -D "$PGDATA" status >/dev/null 2>&1; then
        pg_ctl -D "$PGDATA" -m fast stop >/dev/null 2>&1 || true
      fi
      rm -rf "$PGDATA"
    fi
  fi

  if [ ! -d "$PGDATA" ]; then
    log "Initializing postgres data directory"
    initdb --no-locale --encoding=UTF8 -D "$PGDATA" >/dev/null
    cat <<EOF >>"$PGDATA/pg_hba.conf"
host all all 127.0.0.1/32 trust
host all all ::1/128 trust
EOF
  fi

  if pg_ctl -D "$PGDATA" status >/dev/null 2>&1; then
    log "Postgres already running"
  else
    log "Starting postgres on port $PGPORT"
    pg_ctl -D "$PGDATA" -o "-p $PGPORT -k $PGSOCKET_DIR" -l "$DEV_DIR/postgres.log" -w start >/dev/null
    sleep 1
  fi

  log "Ensuring mgsearch database & role exist"
  # Use current user as superuser (initdb creates superuser with current username)
  PG_SUPERUSER="${PGUSER:-${USER:-postgres}}"
  psql --quiet -h 127.0.0.1 -p "$PGPORT" -U "$PG_SUPERUSER" -d postgres <<'SQL' >/dev/null 2>&1 || true
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'mgsearch') THEN
    CREATE ROLE mgsearch LOGIN PASSWORD 'mgsearch';
  END IF;
END $$;
SQL

  psql --quiet -h 127.0.0.1 -p "$PGPORT" -U "$PG_SUPERUSER" -d postgres <<'SQL' >/dev/null 2>&1 || true
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'mgsearch') THEN
    CREATE DATABASE mgsearch OWNER mgsearch;
  END IF;
END $$;
SQL
}

stop_postgres() {
  if [ -d "$PGDATA" ] && pg_ctl -D "$PGDATA" status >/dev/null 2>&1; then
    log "Stopping postgres"
    pg_ctl -D "$PGDATA" stop -m fast >/dev/null
  fi
}

status_postgres() {
  if [ -d "$PGDATA" ] && pg_ctl -D "$PGDATA" status >/dev/null 2>&1; then
    log "Postgres running (port $PGPORT)"
  else
    log "Postgres not running"
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
    start_postgres
    start_redis
    log "Postgres and Redis started"
    ;;
  down)
    stop_redis
    stop_postgres
    log "Postgres and Redis stopped"
    ;;
  status)
    status_postgres
    status_redis
    ;;
  *)
    echo "Usage: $0 {up|down|status}"
    exit 1
    ;;
esac

