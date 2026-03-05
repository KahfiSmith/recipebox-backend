# Architecture

## Layering

- `controller`: HTTP handlers, input parsing, and response writing.
- `service`: business logic and orchestration.
- `repository`: data access (GORM + PostgreSQL).
- `store/redis`: Redis-backed short-lived auth state and counters.
- `models`: domain models + table mapping.
- `dto`: request/response payload contracts.

## Runtime Components

- API server: Go + Chi router.
- Database: PostgreSQL.
- Cache and auth-state store: Redis (`github.com/redis/go-redis/v9`).
- Auth: JWT access token + rotating refresh token.
- Email notification: SMTP sender.

## Redis Responsibilities

- Auth rate limiting (`rl:auth:*`) for sensitive auth endpoints.
- Access-token session state (active session keys by JWT `jti`).
- Access-token blacklist state for immediate invalidation on logout/revoke.
- OTP/token short-lived state for email verification and password reset flows.
- Refresh-token active state mirror for fast validation/revocation checks.

PostgreSQL remains the durable source of truth for users and persisted auth records.

## Request Flow (Simplified)

1. Request enters route + middleware.
2. Controller performs initial payload validation.
3. Service executes business logic.
4. Service coordinates repository (PostgreSQL) and/or Redis store interactions as needed.
5. Controller returns JSON response.

## Migration Strategy

- SQL migrations (`migrations/`) remain the source of versioned schema changes.
- GORM AutoMigrate is used to sync basic structure at startup.
- Complex changes (rename/drop/backfill) must still be handled explicitly via SQL migrations.
