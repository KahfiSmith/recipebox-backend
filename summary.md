# RecipeBox Backend (Go) - Ringkasan

Dokumen ini merangkum kondisi proyek saat ini dan arah pengembangan yang
direncanakan. Tujuannya sebagai "quick read" tanpa harus membuka banyak file.

## Gambaran Singkat

RecipeBox adalah backend REST API untuk aplikasi manajemen resep, pencarian
resep berdasarkan bahan, pembuatan shopping list otomatis, dan penyusunan
meal plan (harian/mingguan). Backend dirancang API-first dan terpisah dari
frontend.

## Status Saat Ini

- Entry point `cmd/api/main.go` sudah melakukan bootstrap config, koneksi DB,
  start server, dan graceful shutdown.
- Infrastruktur HTTP sudah aktif (`chi` router, middleware, timeout, health
  endpoint).
- Modul auth sudah diimplementasikan:
  - register/login/refresh/logout/me
  - JWT access token
  - refresh token rotation (token disimpan dalam hash)
- OpenAPI sudah memuat endpoint recipes dan auth.
- Migrasi auth tables sudah tersedia di `migrations/`.

## Struktur Proyek Inti

- `cmd/api`:
  entrypoint aplikasi.
- `internal/app`:
  wiring server dan dependency.
- `internal/config`:
  definisi konfigurasi aplikasi.
- `internal/db`:
  koneksi database (Postgres).
- `internal/transport/http`:
  router, middleware, response.
- `internal/recipes`, `internal/mealplan`, `internal/shopping`:
  domain modules.
- `api/openapi.yaml`:
  kontrak API (draft).
- `configs/.env.example`:
  template env vars.

## Konfigurasi Dasar

File contoh `configs/.env.example`:

- `APP_ENV`: environment (development/production).
- `HTTP_ADDR`: address untuk server HTTP.
- `DATABASE_URL`: koneksi Postgres.

## Rencana Pengembangan (High-Level)

- Implementasi CRUD recipes + relasi ingredients.
- Implementasi shopping list generation.
- Implementasi meal plan harian/mingguan.
- Tambah observability (structured logging + metrics).

## Catatan

Dokumen ini bersifat ringkas dan dapat diperbarui seiring implementasi
berjalan.
