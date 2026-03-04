# Architecture

## Layering

- `controller`: HTTP handler, parse input/output response.
- `service`: business logic dan orchestration.
- `repository`: data access (GORM + Postgres).
- `models`: model domain + mapping tabel.
- `dto`: contract payload request/response.

## Runtime Components

- API server: Go + Chi router.
- Database: PostgreSQL.
- Cache/rate limit store: Redis.
- Auth: JWT access token + rotating refresh token.
- Email notification: SMTP sender.

## Request Flow (Simplified)

1. Request masuk ke route + middleware.
2. Controller validasi payload awal.
3. Service menjalankan business logic.
4. Repository baca/tulis data ke DB.
5. Controller kirim response JSON.

## Migration Strategy

- Migration SQL (`migrations/`) tetap jadi sumber perubahan schema versi-per-versi.
- GORM AutoMigrate dipakai untuk sinkronisasi struktur dasar saat startup.
- Perubahan kompleks (rename/drop/backfill) tetap ditangani explicit via migration SQL.
