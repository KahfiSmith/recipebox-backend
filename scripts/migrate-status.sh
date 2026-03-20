set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="$ROOT_DIR/.env"
MIGRATIONS_DIR="$ROOT_DIR/migrations"

if [[ -f "$ENV_FILE" ]]; then
  set -a; source "$ENV_FILE"; set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL is not set. Populate .env first." >&2
  exit 1
fi

if command -v migrate >/dev/null 2>&1; then
  migrate -path "$MIGRATIONS_DIR" -database "$DATABASE_URL" version || true
fi

if command -v psql >/dev/null 2>&1; then
  echo "Current tables:"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "\\dt"
  exit 0
fi

echo "psql was not found. Cannot check table status." >&2
exit 1
