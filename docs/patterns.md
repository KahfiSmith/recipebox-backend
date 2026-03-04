# Implementation Patterns

## Pattern 1: Add CRUD Endpoint
1. Register route method/path in `internal/routes`.
2. Add controller handler for parsing + HTTP mapping.
3. Add service method for validation + business rules.
4. Add repository method for DB operation.
5. Scope queries by `user_id` for user-owned resources.
6. Add/update tests.
7. Update `docs/api.md` and regenerate Swagger.

## Pattern 2: Update Existing Endpoint
1. Preserve existing request/response contract by default.
2. Add optional fields instead of breaking old fields.
3. Keep status code and envelope conventions unchanged.
4. Verify impacted tests and manual endpoint behavior.

## Pattern 3: Delete Resource Safely
1. Validate path ID in controller.
2. Service executes delete use-case.
3. Repository delete by `id + user_id` when user-scoped.
4. Return `404` when not found.
5. Return `{"message": "..."}` on success.

## Pattern 4: Error Mapping
- Domain/service validation error -> `400`
- Unauthorized context/token -> `401`
- Missing record -> `404`
- Conflict state -> `409`
- Unexpected infra/runtime error -> `500`

## Pattern 5: Docs Sync
When API behavior changes:
1. Update route/controller annotations.
2. Update `docs/api.md` endpoint list.
3. Run `bash scripts/swagger-generate.sh`.
4. Ensure `docs/swagger.yaml`, `docs/swagger.json`, and `docs/docs.go` are refreshed.
