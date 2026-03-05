# Repository Rules

This file defines enforceable repository-level rules for backend changes.

## 1) Scope and Change Discipline
- Keep changes task-focused and minimal.
- Avoid unrelated refactors in the same change set.
- If behavior changes, update docs and impacted tests in the same change.

## 2) Layer Boundaries (Strict)
- Required flow: `routes -> controller -> service -> repository -> models`.
- `controller` must not access DB directly.
- `repository` must not contain HTTP concerns (`http.Request`, status codes, response envelopes).
- `service` must not depend on concrete DB adapters directly; use repository interfaces.
- Redis short-lived auth state must be accessed from service/middleware via dedicated store abstractions (`internal/store/redis`), not directly from controllers.

## 3) API Contract Rules
- Success envelope:
  - `{"data": ...}` or `{"message": "..."}`
- Error envelope:
  - `{"error": "..."}`
- HTTP status mapping:
  - `400`: invalid input / validation
  - `401`: unauthorized
  - `404`: resource not found
  - `409`: conflict
  - `500`: internal/unexpected

## 4) Backward Compatibility
- Existing request/response contracts are stable by default.
- Breaking changes require explicit approval.
- Prefer additive changes (optional fields) over contract replacement.

## 5) Authentication and Authorization
- Protected endpoints must stay inside auth middleware groups.
- User-scoped resources must enforce `user_id` filtering at repository query level.
- Do not leak auth internals in error responses.
- Refresh tokens must remain persisted in PostgreSQL and may be mirrored in Redis for active-state checks.

## 6) Database and Migrations
- Schema changes must use SQL migrations under `migrations/`.
- Do not modify old migrations already used in shared environments.
- Complex schema changes (rename/drop/backfill) must be explicit SQL, not implicit magic.

## 7) Swagger and Docs Sync
- API annotation source: controllers.
- After endpoint/annotation changes, regenerate swagger artifacts:
  - `bash scripts/swagger-generate.sh`
- Keep `docs/api.md` endpoint list aligned with current routes.

## 8) Security Rules
- No secrets in git.
- Validate and sanitize input before persistence.
- Do not return sensitive internals (token values, stack traces, SQL errors) in API responses.

## 9) Testing and Verification Gates
Minimum before handoff:
- `go test ./...`
- Manual check of changed endpoints via curl/Postman.
- Docs/swagger updated when API changed.

If local tooling is unavailable, clearly state that and provide exact commands for the user.

## 10) Handoff Format
Always report in this order:
1. Solution summary
2. Files changed
3. Verification steps
4. Risks/limitations
