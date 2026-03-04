#!/usr/bin/env bash
set -euo pipefail

if ! command -v swag >/dev/null 2>&1; then
  echo "swag CLI not found. Install with:"
  echo "  go install github.com/swaggo/swag/cmd/swag@latest"
  exit 1
fi

swag init \
  -g cmd/api/main.go \
  -o docs \
  --parseDependency \
  --parseInternal

echo "Swagger docs generated in docs/"
