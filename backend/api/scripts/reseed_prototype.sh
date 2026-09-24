#!/usr/bin/env bash
# Re-apply prototype seed SQL (migration 00014) against DATABASE_URL.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
MIGRATION="$ROOT/api/migrations/00014_seed_prototype_content.sql"
URL="${DATABASE_URL:-}"

if [[ -z "$URL" ]]; then
  if [[ -f "$ROOT/../.env" ]]; then
    set -a && source "$ROOT/../.env" && set +a
    URL="${DATABASE_URL:-}"
  fi
fi

if [[ -z "$URL" ]]; then
  echo "Set DATABASE_URL (Supabase pooler recommended for IPv4)." >&2
  exit 1
fi

# Prefer pooler host if direct db.*.supabase.co is configured
if [[ "$URL" == *"@db."*".supabase.co"* ]] && [[ -z "${FORCE_DIRECT_DB:-}" ]]; then
  echo "Tip: use the Supabase pooler URL if direct db.* host fails (IPv6)." >&2
fi

awk '/^-- \+goose Up$/,/^-- \+goose Down$/{if(!/^-- \+goose/){print}}' "$MIGRATION" \
  | docker run --rm -i postgres:16-alpine psql "$URL"

echo "Prototype seed SQL applied."
