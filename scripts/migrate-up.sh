#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"
MIGRATIONS_DIR="$ROOT_DIR/migrations"

if [[ -f "$ENV_FILE" ]]; then
  # shellcheck disable=SC1090
  set -a; source "$ENV_FILE"; set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL belum diset. Isi .env dulu." >&2
  exit 1
fi

if [[ ! -d "$MIGRATIONS_DIR" ]]; then
  echo "Folder migrations tidak ditemukan: $MIGRATIONS_DIR" >&2
  exit 1
fi

if command -v migrate >/dev/null 2>&1; then
  if [[ $# -gt 0 ]]; then
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up "$1"
  else
    migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" up
  fi
  exit 0
fi

if command -v psql >/dev/null 2>&1; then
  shopt -s nullglob
  files=("$MIGRATIONS_DIR"/*.up.sql)
  if [[ ${#files[@]} -eq 0 ]]; then
    echo "Tidak ada file migration .up.sql" >&2
    exit 1
  fi

  IFS=$'\n' files=($(printf '%s\n' "${files[@]}" | sort))
  unset IFS

  for file in "${files[@]}"; do
    echo "Applying $file"
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
  done
  exit 0
fi

echo "Tool migrate/psql tidak ditemukan. Install salah satu dulu." >&2
exit 1
