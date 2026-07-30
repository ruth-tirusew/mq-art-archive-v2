#!/usr/bin/env bash
# Apply goose migrations against a temporary database, then drop it.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
: "${DATABASE_URL:?DATABASE_URL is required}"

GO="$(command -v go || true)"
GO_DOCKER_IMAGE="${GO_DOCKER_IMAGE:-golang:1.25-bookworm}"

# Derive admin URL and create a disposable DB name.
BASE_URL="${DATABASE_URL%%\?*}"
QUERY=""
if [[ "$DATABASE_URL" == *"?"* ]]; then
  QUERY="?${DATABASE_URL#*\?}"
fi

# Strip trailing /dbname
DB_HOST_PART="${BASE_URL%/*}"
CHECK_DB="mq_migrate_check_$$"
CHECK_URL="${DB_HOST_PART}/${CHECK_DB}${QUERY}"
ADMIN_URL="${DB_HOST_PART}/postgres${QUERY}"

cleanup() {
  psql "$ADMIN_URL" -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${CHECK_DB};" >/dev/null 2>&1 || true
}
trap cleanup EXIT

psql "$ADMIN_URL" -v ON_ERROR_STOP=1 -c "DROP DATABASE IF EXISTS ${CHECK_DB};" >/dev/null
psql "$ADMIN_URL" -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${CHECK_DB};" >/dev/null

run_goose() {
  if [[ -n "$GO" ]]; then
    (cd "$ROOT/backend/api" && GOTOOLCHAIN=local "$GO" run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$CHECK_URL" up)
  else
    docker run --rm --network host \
      -e "DATABASE_URL=$CHECK_URL" \
      -v "$ROOT/backend/api:/app" -w /app \
      "$GO_DOCKER_IMAGE" \
      sh -c 'go install github.com/pressly/goose/v3/cmd/goose@latest && goose -dir migrations postgres "$DATABASE_URL" up'
  fi
}

run_goose
echo "migrate-check OK (${CHECK_DB})"
