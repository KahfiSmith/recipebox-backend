# API Reference

Primary generated API artifacts:
- `docs/swagger.yaml`
- `docs/swagger.json`

Postman import:
- Import `docs/swagger.json`
- Import `docs/recipebox.postman_collection.json` for a curated collection grouped into `Auth`, `Dashboard`, `Recipes`, `Meal Plans`, and `Shopping Items`

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

## Pagination
- List endpoints support query params:
  - `limit` (default `20`, max `100`)
  - `offset` (default `0`)
- Applies to:
  - `GET /api/v1/recipes`
  - `GET /api/v1/meal-plans`
  - `GET /api/v1/shopping-items`

## Auth Behavior Notes
- `POST /api/v1/auth/register` creates the user and immediately triggers the email-verification flow for the registered email address. If debug token exposure is enabled, the register response also includes `data.emailVerification` so development flows do not block on SMTP delivery.
- Verification emails include the one-time token/code in the email body. Password-reset emails include an 8-digit numeric reset code in the email body. If `FRONTEND_BASE_URL` is configured, the email also includes the frontend action link.
- `POST /api/v1/auth/login` and `POST /api/v1/auth/refresh` set refresh token in HTTP-only cookie (`refresh_token`), and do not expose refresh token in response body.
- `POST /api/v1/auth/refresh` accepts refresh token from cookie first, and falls back to request body (`refreshToken`).
- `POST /api/v1/auth/logout` revokes refresh token and clears cookie; when a bearer access token is provided, access session is revoked/blacklisted.
- Access tokens are validated via JWT middleware and checked against Redis-backed auth state.
- Protected dashboard endpoints and `GET /api/v1/auth/me` require `Authorization: Bearer <access_token>`.

## Dashboard Performance Notes
- `GET /api/v1/dashboard` uses short-lived server-side Redis cache per user.
- Recipe/meal-plan/shopping write endpoints invalidate cached dashboard overview for that user.
