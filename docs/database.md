# Database

## Engine
- PostgreSQL

## Migration Source
- SQL files under `migrations/`

## Core Tables
- `users`
- `refresh_tokens`
- `email_verification_tokens`
- `password_reset_tokens`
- `recipes`
- `meal_plans`
- `shopping_items`

## Migration Commands
```bash
bash scripts/migrate-up.sh
bash scripts/migrate-status.sh
bash scripts/migrate-down.sh 1
```

## Notes
- Add new migration files for schema changes.
- Do not modify old migrations that are already applied in shared environments.

## Redis Auth-State (Non-SQL)
- Redis is used for short-lived auth/runtime state and counters via `github.com/redis/go-redis/v9`.
- Current Redis key purposes:
  - `rl:auth:*` auth rate limiting counters.
  - `auth:sess:access:*` active access-token sessions (`jti`).
  - `auth:blacklist:access:*` blacklisted access tokens.
  - `auth:otp:*` OTP/token state for verification/reset flows.
  - `auth:refresh:*` active refresh-token state mirror.
- PostgreSQL remains the durable source of truth for users and persisted auth tables.
