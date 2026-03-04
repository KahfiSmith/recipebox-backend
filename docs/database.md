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
