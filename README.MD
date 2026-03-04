# RecipeBox Backend API (Go)

Backend REST API untuk RecipeBox.

## Dokumentasi

Semua dokumentasi teknis dipusatkan di folder `docs` agar tidak duplikatif:

```text
docs/
├── swagger.yaml
├── swagger.json
├── openapi.yaml (canonical: /openapi.yaml)
├── postman_collection.json
├── api.md
└── architecture.md
```

Catatan:
- Sumber kontrak API utama adalah `/openapi.yaml`.
- `docs/swagger.yaml` dan `docs/swagger.json` disediakan sebagai alias kompatibilitas.

## Swagger (Swaggo)

Install CLI:

- `go install github.com/swaggo/swag/cmd/swag@latest`

Generate dokumen dari anotasi code:

- `./scripts/swagger-generate.sh`

Hasil generate akan berada di folder `docs/` dan bisa di-import ke Postman.

## Status Singkat

- Stack: Go + Chi + GORM + PostgreSQL + Redis
- Layering: controller, service, repository, models, dto

## Menjalankan Lokal

1. Copy `.env.example` ke `.env`
2. Jalankan Redis:
   - `docker compose up -d redis`
3. Jalankan migration SQL (opsional sesuai kebutuhan environment):
   - `./scripts/migrate-up.sh`
4. Jalankan server:
   - `go run ./cmd/api`

## Struktur Inti

- `cmd/api`: entrypoint
- `internal/controller`: HTTP handlers
- `internal/service`: business logic
- `internal/repository`: data access
- `internal/models`: data model
- `internal/dto`: request/response payload
- `internal/db`: database bootstrap + automigrate helper
