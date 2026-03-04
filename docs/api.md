# API Reference

Primary generated API artifacts:
- `docs/swagger.yaml`
- `docs/swagger.json`

Postman import:
- Import `docs/swagger.json`

Generate Swagger from annotations:
- Install CLI: `go install github.com/swaggo/swag/cmd/swag@latest`
- Run: `bash scripts/swagger-generate.sh`

## Base URL
- Local: `http://localhost:8080`
- API prefix: `/api/v1`

## Main Endpoint Groups
- System:
  - `GET /healthz`
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
- Dashboard / Menus:
  - `GET /api/v1/dashboard`
  - `GET /api/v1/recipes`
  - `POST /api/v1/recipes`
  - `PUT /api/v1/recipes/{id}`
  - `DELETE /api/v1/recipes/{id}`
  - `GET /api/v1/meal-plans`
  - `POST /api/v1/meal-plans`
  - `PUT /api/v1/meal-plans/{id}`
  - `DELETE /api/v1/meal-plans/{id}`
  - `GET /api/v1/shopping-items`
  - `POST /api/v1/shopping-items`
  - `PUT /api/v1/shopping-items/{id}`
  - `DELETE /api/v1/shopping-items/{id}`

For full request/response schemas, use generated Swagger files.
