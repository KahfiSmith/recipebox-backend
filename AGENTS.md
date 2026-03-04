# Repository Guidelines (Backend Go - RecipeBox)

This document is the implementation guide for agents/contributors in this repository.
Goal: consistent, fast delivery with minimal over-engineering.

## Source of Truth (Read First)
- Architecture: `docs/architecture.md`
- API reference: `docs/api.md`
- Database notes: `docs/database.md`
- Repository rules: `docs/rules.md`
- Coding standards: `docs/coding-standards.md`
- Implementation patterns: `docs/patterns.md`

If conflicts exist, follow active router behavior and the running code path.

## Non-Negotiables (Hard Rules)
- Follow this flow: `routes -> controller -> service -> repository -> models`.
- Do not skip layers (controllers must not query DB directly).
- Keep API response format consistent:
  - success: `{"data": ...}` or `{"message": "..."}`
  - error: `{"error": "..."}`
- Keep HTTP status mapping consistent:
  - `400` input/validation
  - `401` unauthorized
  - `404` not found
  - `409` conflict
  - `500` internal error
- Do not change existing endpoint contracts without stating backward-compatibility impact.
- For DB schema changes: add new migrations; do not edit old applied migrations.
- Never commit secrets/credentials.

## Engineering Principles
- Correctness over speed.
- Maintainability over cleverness.
- Backward compatibility by default.
- Simplicity is a feature.

## Core Design Principles (DRY, SOLID, KISS)
- **DRY**: avoid repeated logic; extract only when it improves clarity.
- **SOLID**: keep responsibilities clear and interfaces small.
- **KISS**: prefer the simplest design that solves the current problem.

## Implementation Patterns
- Add/update endpoint: sync route, controller, service, repository, dto, and docs.
- Keep domain errors in service/repository; map to HTTP in controller.
- For user-scoped resources, enforce `user_id` filtering in repository queries.

## Swagger Rules
- Swagger header: `cmd/api/main.go`
- Endpoint annotations: `internal/controller/*.go`
- Regenerate when endpoint/annotation changes:
  - `bash scripts/swagger-generate.sh`

## Testing Rules
Minimum before handoff:
- `go test ./...`
- manual validation of changed endpoints via curl/Postman

If tooling is unavailable locally, state it explicitly and provide exact commands for the user.

## Output Contract (Agent)
Always report in this order:
1. Short solution summary
2. Files changed
3. Verification steps
4. Risks/limitations not yet verified

## Quick Command Reference
```bash
# tests
go test ./...

# swagger
bash scripts/swagger-generate.sh

# migration
bash scripts/migrate-up.sh
bash scripts/migrate-status.sh
bash scripts/migrate-down.sh 1
```
