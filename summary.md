# RecipeBox Backend (Go) - Ringkasan

Dokumen ini merangkum kondisi proyek saat ini dan arah pengembangan yang
direncanakan. Tujuannya sebagai "quick read" tanpa harus membuka banyak file.

## Gambaran Singkat

RecipeBox adalah backend REST API untuk aplikasi manajemen resep, pencarian
resep berdasarkan bahan, pembuatan shopping list otomatis, dan penyusunan
meal plan (harian/mingguan). Backend dirancang API-first dan terpisah dari
frontend.

## Status Saat Ini (Scaffold)

- Entry point `cmd/api/main.go` hanya melakukan log bootstrap.
- Struktur layer telah disiapkan di `internal/` (handler/service/repo) untuk
  domain recipes, mealplan, shopping, tetapi belum ada implementasi.
- Transport HTTP dan infrastruktur (server, router, middleware, db) masih
  placeholder.
- OpenAPI `api/openapi.yaml` ada, tetapi `paths` masih kosong.
- Belum ada dependency di `go.mod`, migrasi dan scripts masih kosong.

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

- Implementasi wiring server (config loader, db, router, middleware).
- Definisi kontrak API pada OpenAPI (CRUD recipes, search, shopping list,
  meal plan).
- Implementasi layer handler/service/repo untuk setiap domain.
- Penambahan migrasi schema database.
- Penambahan dependency (router, pg driver, migrate).

## Catatan

Dokumen ini bersifat ringkas dan dapat diperbarui seiring implementasi
berjalan.
