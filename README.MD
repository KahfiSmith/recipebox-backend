# RecipeBox Backend API (Go)

REST backend API for RecipeBox.

## Documentation

All technical documentation is centralized in the `docs` folder to avoid duplication:

```text
docs/
├── swagger.yaml
├── swagger.json
├── openapi.yaml (canonical: /openapi.yaml)
├── recipebox.postman_collection.json
├── api.md
└── architecture.md
```

Notes:
- The primary API contract source is `/openapi.yaml`.
- `docs/swagger.yaml` and `docs/swagger.json` are provided as compatibility aliases.

## Swagger (Swaggo)

Install the CLI:

- `go install github.com/swaggo/swag/cmd/swag@latest`

Generate docs from code annotations:

- `./scripts/swagger-generate.sh`

Generated files will be placed in `docs/` and can be imported into Postman.

## Quick Status

- Stack: Go + Chi + GORM + PostgreSQL + Redis
- Layering: controller, service, repository, models, dto

## Run Locally

1. Copy `.env.example` to `.env`
2. Start Redis:
   - `docker compose up -d redis`
3. Run SQL migrations (optional depending on the environment):
   - `./scripts/migrate-up.sh`
4. Start the server:
   - `go run ./cmd/api`

## Core Structure

- `cmd/api`: entrypoint
- `internal/controller`: HTTP handlers
- `internal/service`: business logic
- `internal/repository`: data access
- `internal/models`: data model
- `internal/dto`: request/response payload
- `internal/db`: database bootstrap + automigrate helper
