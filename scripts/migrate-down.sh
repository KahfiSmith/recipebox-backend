#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"
MIGRATIONS_DIR="$ROOT_DIR/migrations"
STEPS="${1:-1}"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a; source "$ENV_FILE"; set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL belum diset. Isi .env dulu." >&2
  exit 1
fi

if ! [[ "$STEPS" =~ ^[0-9]+$ ]] || [[ "$STEPS" -lt 1 ]]; then
  echo "Argumen steps harus integer >= 1" >&2
  exit 1
fi

if command -v migrate >/dev/null 2>&1; then
  migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" down "$STEPS"
  exit 0
fi

if command -v psql >/dev/null 2>&1; then
  shopt -s nullglob
  files=("$MIGRATIONS_DIR"/*.down.sql)
  if [[ ${#files[@]} -eq 0 ]]; then
    echo "Tidak ada file migration .down.sql" >&2
    exit 1
  fi

  IFS=$'\n' files=($(printf '%s\n' "${files[@]}" | sort -r))
  unset IFS

  count=0
  for file in "${files[@]}"; do
    if [[ "$count" -ge "$STEPS" ]]; then
      break
    fi
    echo "Reverting $file"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
    count=$((count + 1))
  done

  if [[ "$count" -lt "$STEPS" ]]; then
    echo "Warning: requested $STEPS step(s), but only reverted $count file(s)." >&2
  fi
  exit 0
fi

echo "Tool migrate/psql tidak ditemukan. Install salah satu dulu." >&2
exit 1
