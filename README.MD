# RecipeBox Backend API (Go)

Backend REST API untuk RecipeBox, aplikasi web untuk mengelola resep, mencari resep berdasarkan bahan, membuat shopping list otomatis, dan menyusun meal plan.

Backend ini dirancang API-first agar mudah diintegrasikan dengan frontend terpisah.

## Gambaran Umum

RecipeBox Backend menyediakan API untuk:

- CRUD resep masakan
- Pencarian resep berdasarkan bahan yang tersedia
- Generate shopping list dari resep terpilih
- Manajemen meal plan (kalender harian atau mingguan)
- Autentikasi user (register, login, refresh token, logout, profile)

## Tech Stack

- Language: Go
- HTTP Router: chi
- Database: PostgreSQL
- ORM/Query: GORM
- Migration: golang-migrate
- Config: environment variables (.env)
- API Style: REST JSON

## Struktur Proyek

```text
.
├── cmd/
│   └── api/                 # entrypoint aplikasi
│       └── main.go
├── .env.example             # contoh konfigurasi
├── internal/
│   ├── config/                # load env dan config
│   │   └── config.go
│   ├── db/                    # database connection
│   │   └── postgres.go
│   ├── server/                # bootstrap router + server
│   │   ├── router.go
│   │   └── server.go
│   ├── controller/            # HTTP handler
│   │   └── auth_controller.go
│   ├── routes/                # route registration
│   │   ├── routes.go
│   │   └── auth_routes.go
│   ├── dto/                   # request/response payload
│   │   └── auth_dto.go
│   ├── entity/                # domain/entity model
│   │   └── auth.go
│   ├── repository/            # data access
│   │   ├── auth_repo.go
│   │   └── auth_repo_gorm.go
│   ├── service/               # business logic
│   │   └── auth_service.go
│   ├── middleware/            # middleware lintas route
│   │   └── jwt.go
│   └── utils/                 # helper bersama
│       └── response.go
├── migrations/              # database migrations
├── scripts/                 # helper scripts
├── go.mod
└── README.MD
```

## Alur Layering

- `controller`: HTTP request/response handling.
- `service`: business logic.
- `repository`: akses data (GORM/Postgres).
- `entity`/`dto`: model domain dan payload API.

## Menjalankan Lokal (draft)

1. Salin `.env.example` menjadi `.env`
2. Jalankan migration database
3. Jalankan server:

```bash
go run ./cmd/api
```

## Endpoint Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/verify-email/request`
- `POST /api/v1/auth/verify-email/confirm`
- `POST /api/v1/auth/password/forgot`
- `POST /api/v1/auth/password/reset`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me` (Bearer access token)

Catatan auth:
- Endpoint sensitif (`login`, `refresh`, `verify-email/request`, `password/forgot`) sudah memakai rate limit per-IP (`AUTH_RATE_LIMIT_PER_MINUTE`).
- Email verifikasi dan reset password dikirim via SMTP (`SMTP_*`).
- Saat `AUTH_DEBUG_EXPOSE_TOKENS=true`, token one-time masih dikembalikan di response (mode lokal/debug).
- Saat `AUTH_DEBUG_EXPOSE_TOKENS=false`, `SMTP_HOST` wajib terisi.

Contoh base URL untuk frontend:

- `http://localhost:8080/api/v1`

## Menjalankan Redis (Docker Compose)

```bash
docker compose up -d redis
docker compose ps
```

Untuk berhenti:

```bash
docker compose down
```

## Migration Helpers

Gunakan helper script berikut:

```bash
./scripts/migrate-up.sh
./scripts/migrate-down.sh 1
./scripts/migrate-status.sh
```

Script membaca `DATABASE_URL` dari `.env`.
