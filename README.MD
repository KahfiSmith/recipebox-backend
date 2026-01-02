# RecipeBox Backend API (Go)

Backend REST API untuk RecipeBox, aplikasi web untuk mengelola resep, mencari resep berdasarkan bahan, membuat shopping list otomatis, dan menyusun meal plan.

Backend ini dirancang API-first agar mudah diintegrasikan dengan frontend terpisah.

Fase awal backend tanpa autentikasi. Auth dan multi-user akan ditambahkan pada tahap berikutnya.

## Gambaran Umum

RecipeBox Backend menyediakan API untuk:

- CRUD resep masakan
- Pencarian resep berdasarkan bahan yang tersedia
- Generate shopping list dari resep terpilih
- Manajemen meal plan (kalender harian atau mingguan)

## Tech Stack

- Language: Go
- HTTP Router: chi
- Database: PostgreSQL
- ORM/Query: database/sql + sqlc (opsional)
- Migration: golang-migrate
- Config: environment variables (.env)
- API Style: REST JSON
- API Spec: OpenAPI (api/openapi.yaml)

## Struktur Proyek

```text
.
├── api/                     # OpenAPI spec
│   └── openapi.yaml
├── cmd/
│   └── api/                 # entrypoint aplikasi
│       └── main.go
├── configs/                 # contoh konfigurasi
│   └── .env.example
├── internal/
│   ├── app/                 # wiring server dan dependency
│   │   └── server.go
│   ├── config/              # load env dan config
│   │   └── config.go
│   ├── db/                  # database connection
│   │   └── postgres.go
│   ├── transport/
│   │   └── http/            # router, middleware, response
│   │       ├── middleware.go
│   │       ├── response.go
│   │       └── router.go
│   ├── mealplan/            # fitur meal plan
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repo.go
│   │   └── service.go
│   ├── recipes/             # fitur resep
│   │   ├── handler.go
│   │   ├── model.go
│   │   ├── repo.go
│   │   └── service.go
│   └── shopping/            # fitur shopping list
│       ├── handler.go
│       ├── model.go
│       ├── repo.go
│       └── service.go
├── migrations/              # database migrations
├── scripts/                 # helper scripts
├── go.mod
└── README.MD
```

## Alur Layering

- handler: HTTP request/response
- service: business logic
- repo: akses database

## Menjalankan Lokal (draft)

1. Salin `configs/.env.example` menjadi `.env`
2. Jalankan migration database
3. Jalankan server:

```bash
go run ./cmd/api
```
