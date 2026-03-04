# Architecture

## Layering

- `controller`: HTTP handlers, input parsing, and response writing.
- `service`: business logic and orchestration.
- `repository`: data access (GORM + PostgreSQL).
- `models`: domain models + table mapping.
- `dto`: request/response payload contracts.

## Runtime Components

- API server: Go + Chi router.
- Database: PostgreSQL.
- Cache/rate-limit store: Redis.
- Auth: JWT access token + rotating refresh token.
- Email notification: SMTP sender.

## Request Flow (Simplified)

1. Request enters route + middleware.
2. Controller performs initial payload validation.
3. Service executes business logic.
4. Repository reads/writes data to DB.
5. Controller returns JSON response.

## Migration Strategy

- SQL migrations (`migrations/`) remain the source of versioned schema changes.
- GORM AutoMigrate is used to sync basic structure at startup.
- Complex changes (rename/drop/backfill) must still be handled explicitly via SQL migrations.
