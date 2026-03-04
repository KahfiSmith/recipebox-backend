# API Reference

Dokumen API utama ada di:
- `/openapi.yaml` (canonical source of truth)

Alias kompatibilitas:
- `/docs/swagger.yaml`
- `/docs/swagger.json`

Postman:
- `/docs/postman_collection.json`

Generate via Swaggo:
- Install: `go install github.com/swaggo/swag/cmd/swag@latest`
- Run: `./scripts/swagger-generate.sh`

## Base URL

- Local: `http://localhost:8080`
- API prefix: `/api/v1`

## Main Endpoint Groups

- System: `GET /healthz`
- Auth:
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/verify-email/request`
  - `POST /api/v1/auth/verify-email/confirm`
  - `POST /api/v1/auth/password/forgot`
  - `POST /api/v1/auth/password/reset`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
  - `GET /api/v1/auth/me`
- Dashboard:
  - `GET /api/v1/dashboard`
  - `GET /api/v1/recipes`
  - `GET /api/v1/meal-plans`
  - `GET /api/v1/shopping-items`

Untuk schema request/response detail, selalu rujuk ke `openapi.yaml`.
