# Coding Standards

This document defines coding conventions for maintainable and predictable backend code.

## 1) General Style
- Prefer readability over cleverness.
- Keep functions focused and short where practical.
- Use explicit names that describe intent (`CreateRecipe`, `UpdateMealPlan`, `DeleteShoppingItem`).
- Keep comments meaningful; avoid obvious comments.

## 2) Package and Layer Responsibilities
- `routes`: route + middleware wiring only.
- `controller`: parse request, call service, map errors to HTTP response.
- `service`: validation and business logic.
- `repository`: persistence and query logic only.
- `store/redis`: short-lived auth/session state and counters (rate-limit, blacklist, OTP, active token state).
- `dto/models`: transport and persistence structures.

## 3) Error Handling
- Return domain/business errors from service/repository.
- Map errors to HTTP status in controller only.
- Wrap infrastructure errors with context in repository/service (`fmt.Errorf("context: %w", err)`).
- Keep client-facing error messages clear and safe.

## 4) Validation
- Validate required fields and value bounds in controller/service.
- Reject malformed IDs and invalid payloads early.
- Keep validation rules consistent across create/update endpoints.

## 5) Data Access Standards
- Repository methods should be deterministic and scoped.
- For user-owned entities, query by both `id` and `user_id`.
- On update/delete, return not-found when `RowsAffected == 0`.
- Avoid leaking DB-specific errors to transport layer.
- Keep Redis access behind store implementations/interfaces; avoid direct Redis client calls in controller/service business code.

## 6) API Consistency
- Use shared envelope format:
  - success: `{"data": ...}` or `{"message": "..."}`
  - error: `{"error": "..."}`
- Preserve status code conventions (`400/401/404/409/500`).

## 7) Naming Conventions
- Methods: verb + resource (`ListRecipes`, `CreateRecipe`, `UpdateRecipe`, `DeleteRecipe`).
- DTOs: transport intent (`UpsertRecipeRequest`, `RecipeEnvelope`).
- Interfaces: concise and behavior-focused (`RecipeBoxRepository`).

## 8) Testing Standards
- Add/update tests for behavioral changes.
- Prioritize service + controller tests for API behavior.
- Cover at least:
  - success path
  - validation failure
  - not-found path
  - internal error mapping

## 9) Documentation Standards
- Update `docs/api.md` when endpoints change.
- Keep Swagger annotations in sync with handlers and DTOs.
- Regenerate swagger artifacts after annotation changes.

## 10) Simplicity and Refactoring
- Avoid over-engineering for small feature requests.
- Introduce new abstractions only when there is repeated, proven need.
- Refactor safely and incrementally; keep behavior stable unless explicitly changed.
