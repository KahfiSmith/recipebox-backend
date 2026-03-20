# Architecture

## Overview

RecipeBox backend is a layered Go HTTP API built around Chi, GORM, PostgreSQL, and Redis.
The current system is organized into two main functional areas:

- Auth and account lifecycle:
  - registration
  - login
  - access token validation
  - rotating refresh token flow
  - logout and token revocation
  - email verification
  - password reset
- RecipeBox dashboard domain:
  - dashboard overview
  - recipes CRUD
  - meal plans CRUD
  - shopping items CRUD

The architecture favors explicit dependency wiring, clear layer boundaries, and stable API behavior.
Persistent business data lives in PostgreSQL.
Short-lived runtime state and cache live in Redis.

## Architectural Goals

- Keep request handling predictable through strict layering.
- Keep durable state in PostgreSQL and volatile state in Redis.
- Keep auth/session invalidation fast without making Redis the source of truth for user data.
- Keep API contracts stable and explicit.
- Keep implementation changes small and reversible.

## High-Level Component Map

- `cmd/api`
  - process entrypoint
  - configuration load
  - database connection
  - server start and graceful shutdown
- `internal/config`
  - environment loading and validation
- `internal/server`
  - dependency wiring
  - router construction
  - HTTP server lifecycle
- `internal/routes`
  - endpoint registration and route grouping
- `internal/controller`
  - request parsing
  - HTTP status mapping
  - response envelope writing
- `internal/service`
  - business logic
  - orchestration across repository and Redis-backed state/cache abstractions
- `internal/repository`
  - PostgreSQL persistence logic
- `internal/store/redis`
  - Redis implementations for auth-state and dashboard cache
- `internal/db`
  - PostgreSQL connection and startup schema checks/migration helpers
- `internal/models`
  - persistence/domain structures
- `internal/dto`
  - request/response transport payloads
- `internal/middleware`
  - auth JWT validation
  - auth rate limiting
  - CORS
  - real IP handling
- `internal/notification`
  - SMTP email delivery
- `internal/utils`
  - JSON response helpers
  - request IP helpers

## Dependency Direction

Required dependency flow:

`routes -> controller -> service -> repository -> models`

Additional allowed supporting directions:

- `server -> config/controller/service/repository/store/middleware`
- `middleware.AuthJWT -> service.AuthService`
- `service -> store abstractions`
- `internal/store/redis -> service interfaces`
- `controller -> dto/models/utils/middleware`

Prohibited patterns:

- controller querying PostgreSQL directly
- controller using Redis directly
- repository returning HTTP responses or status codes
- controller containing business rules that belong in service
- service depending on concrete HTTP request/response types

## Runtime Components

- API server:
  - Go `net/http` with Chi router
- Database:
  - PostgreSQL via GORM
- Redis:
  - `github.com/redis/go-redis/v9`
  - used for auth runtime state, rate limiting, and dashboard caching
- Authentication:
  - HS256 JWT access tokens
  - rotating refresh tokens
- Email delivery:
  - SMTP sender for verification and password reset flows

## Bootstrap and Startup Flow

Application startup currently works as follows:

1. `cmd/api/main.go` loads validated config from environment.
2. PostgreSQL connection is opened via `internal/db.OpenPostgres`.
3. `server.NewServer` runs startup schema checks and model synchronization.
4. Redis client is created and pinged.
5. Repositories are created on top of GORM.
6. Services are created and then configured with Redis-backed stores.
7. Optional SMTP sender is configured for auth email flows.
8. Controllers are created from services.
9. Router and HTTP server are constructed.
10. Server starts listening and shuts down gracefully on `SIGINT` or `SIGTERM`.

## Layer Responsibilities

### `routes`

- Register endpoint paths and methods.
- Group protected vs public routes.
- Apply middleware at route-group level.

### `controller`

- Decode request body, path params, and query params.
- Read authenticated user context from middleware.
- Call service methods.
- Map domain errors to HTTP status codes.
- Write JSON responses using the shared envelope format.

### `service`

- Validate business inputs.
- Apply business rules and auth/session rules.
- Coordinate repository operations and Redis-backed state/cache operations.
- Keep HTTP-specific concerns out of the business layer.

### `repository`

- Encapsulate PostgreSQL queries and writes.
- Scope user-owned data by `user_id`.
- Return domain errors such as not-found or conflict conditions.

### `store/redis`

- Implement service-facing interfaces for:
  - auth session state
  - token blacklist state
  - OTP/token one-time use state
  - refresh token active-state mirror
  - dashboard response caching

### `models`

- Define persistent entities and table mappings.

### `dto`

- Define request/response payload shapes exposed by the HTTP layer and Swagger.

## HTTP Surface

Current endpoint groups:

- System:
  - `GET /healthz`
- Auth:
  - `/api/v1/auth/*`
- Dashboard and RecipeBox resources:
  - `/api/v1/dashboard`
  - `/api/v1/recipes`
  - `/api/v1/meal-plans`
  - `/api/v1/shopping-items`

Protected routes use `AuthJWT` middleware.
Public auth endpoints selectively use auth-specific rate limiting.

## Request Lifecycle

Typical protected request flow:

1. Request enters the Chi router.
2. Global middleware runs:
  - request ID
  - CORS
  - real IP resolution
  - panic recovery
  - request logging
  - request timeout
3. Route-specific middleware runs:
  - auth JWT validation for protected endpoints
  - auth rate limit for sensitive public auth endpoints
4. Controller parses input and extracts route/query/body data.
5. Controller calls service.
6. Service validates business rules and orchestrates repository plus Redis-backed store interactions.
7. Repository persists or reads durable data from PostgreSQL.
8. Service returns domain data or domain errors.
9. Controller maps the result to the API response envelope.

## Cross-Cutting API Conventions

Response envelopes:

- success: `{"data": ...}` or `{"message": "..."}`
- error: `{"error": "..."}`

Common status mapping:

- `400` invalid input or validation failure
- `401` unauthorized or invalid token
- `404` resource not found
- `409` conflict
- `500` internal server error

## Authentication and Session Architecture

### Auth Flow Summary

- Registration creates a user in PostgreSQL.
- Login validates credentials and email verification status.
- Successful login issues:
  - JWT access token for bearer auth
  - refresh token stored in HTTP-only cookie
- Refresh rotates the refresh token and issues a new access token.
- Logout revokes refresh token state and invalidates access-token session state.

### Access Tokens

- Signed with HS256 using the configured JWT secret.
- Include standard registered claims:
  - issuer
  - audience
  - subject
  - expiration
  - token ID (`jti`)
- Middleware validates bearer token format.
- `AuthService.ParseAccessToken` validates:
  - JWT signature and claims
  - Redis blacklist state
  - Redis active access-session state

### Refresh Tokens

- Refresh tokens are opaque random strings.
- Only token hashes are persisted.
- PostgreSQL is the durable source of truth.
- Redis stores a fast active-state mirror for revocation and rotation checks.
- Refresh token reuse or metadata mismatch can trigger rejection and broader revocation behavior.

### Email Verification and Password Reset

- Verification and reset use one-time tokens.
- Token hashes are persisted in PostgreSQL.
- Redis stores short-lived one-time-use markers for fast consume/reject behavior.
- Email delivery is handled through SMTP when configured.
- In non-production-style debug mode, token exposure can be enabled for development flow support.

### Auth Rate Limiting

Sensitive auth endpoints use Redis-backed rate limiting keyed by client IP plus route path:

- register
- login
- verify-email request
- password forgot
- refresh

## Dashboard and RecipeBox Architecture

The dashboard area is centered around `DashboardService` and `RecipeBoxRepository`.

Covered resources:

- recipes
- meal plans
- shopping items

### Dashboard Read Model

`GET /api/v1/dashboard` assembles:

- summary counts
- recipe list
- meal plan list
- shopping item list

The summary is derived in service from the underlying lists, including pending shopping item count.

### Pagination

List endpoints support `limit` and `offset` with service/controller normalization:

- default limit: `20`
- max limit: `100`

### Dashboard Cache

- Dashboard overview payload is cached in Redis per user.
- Current cache key pattern:
  - `dashboard:overview:{userID}`
- Cache TTL is short-lived.
- Write operations on recipes, meal plans, and shopping items invalidate the dashboard cache for that user.

## Persistence Architecture

### PostgreSQL Responsibilities

Durable data currently stored in PostgreSQL:

- users
- refresh tokens
- email verification tokens
- password reset tokens
- recipes
- meal plans
- shopping items

User-owned resource queries must always be scoped by `user_id`.

### Redis Responsibilities

Redis stores short-lived or derived runtime state:

- `rl:auth:*`
  - auth rate-limit counters
- `auth:sess:access:*`
  - active access-token sessions by token ID
- `auth:blacklist:access:*`
  - explicitly revoked access-token IDs
- `auth:refresh:*`
  - active refresh-token hash mirror
- `auth:otp:*`
  - short-lived OTP/token state for verification and reset flows
- `dashboard:overview:*`
  - cached dashboard overview payload

Rule of record:

- PostgreSQL is the durable source of truth.
- Redis is an acceleration and runtime-state layer, not the primary business datastore.

## Schema and Migration Strategy

Startup schema behavior is intentionally mixed but explicit:

- SQL migrations under `migrations/` remain the source of versioned schema changes.
- Startup checks verify required auth tables and required columns exist.
- RecipeBox models also run through `AutoMigrate` at startup.
- Some compatibility adjustments are handled in startup migration code, including column rename/drop cleanup for older structures.

Implication:

- New shared schema changes should still be introduced with new SQL migrations.
- Old applied migrations should not be edited.
- Complex or risky schema evolution should not rely on implicit runtime auto-migration alone.

## Configuration Architecture

Important runtime configuration areas:

- server address
- PostgreSQL DSN
- Redis address/password/database
- JWT secret
- access and refresh TTLs
- bcrypt cost
- graceful shutdown timeout
- frontend base URL
- auth debug token exposure
- trusted proxy CIDRs
- SMTP host, port, credentials, sender identity
- auth rate limit per minute

Config is loaded from environment, with `.env` support for local development, then validated before server startup proceeds.

## Middleware Architecture

Global middleware stack currently includes:

- request ID
- CORS
- real IP extraction using trusted proxies
- panic recovery
- structured request logging
- timeout of 30 seconds per request

Protected route middleware:

- `AuthJWT`

Selective public route middleware:

- `NewAuthRateLimit`

## Operational Behavior

Current server-level behavior:

- PostgreSQL connections are pooled through GORM/sql DB settings.
- Redis connectivity is checked on startup.
- HTTP server has read, write, and idle timeouts configured.
- Graceful shutdown is triggered by `SIGINT` and `SIGTERM`.
- Redis client is closed during shutdown.

## Testing and Documentation Expectations

When architecture-relevant behavior changes:

- update the implementation in the correct layer
- update tests for the changed behavior
- update docs under `docs/` that describe the changed area
- regenerate Swagger artifacts when endpoint annotations change

## Documentation Sync

`docs/architecture.md` must be updated when there are changes to:

- layering or dependency direction
- startup/bootstrap wiring
- runtime components
- auth/session design
- cache design
- request lifecycle
- schema/migration strategy
- subsystem boundaries

Related docs that must also be updated when relevant:

- `docs/api.md`
  - endpoint, contract, auth-flow, or Swagger-impacting changes
- `docs/database.md`
  - schema, migration, index, or storage responsibility changes
- `docs/rules.md`
  - enforceable repository-rule changes
- `docs/patterns.md`
  - standard implementation workflow changes

Documentation changes should ship in the same change set as the implementation.
