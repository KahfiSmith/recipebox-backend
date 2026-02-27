# Postman Auth Testing Guide

Panduan ini untuk testing endpoint auth RecipeBox Backend lewat Postman.

## 1) Prasyarat

- Server jalan di `http://localhost:8080`
- Database dan migrasi sudah siap
- `.env` minimal sudah benar untuk:
  - `JWT_SECRET`
  - `DATABASE_URL`
  - `AUTH_DEBUG_EXPOSE_TOKENS`
  - `AUTH_RATE_LIMIT_PER_MINUTE`

Base URL endpoint auth: `http://localhost:8080/api/v1/auth`

## 2) Import OpenAPI ke Postman

1. Buka Postman
2. Klik `Import`
3. Pilih file `openapi.yaml` dari root repo
4. Generate collection baru

Jika `openapi.yaml` berubah, lakukan import ulang agar request/response terbaru ikut ter-update.

## 3) Setup Environment Postman

Buat environment dengan variable:

- `baseUrl` = `http://localhost:8080`
- `email` = email test (contoh `test1@example.com`)
- `password` = password test (contoh `Password123!`)
- `newPassword` = password baru untuk reset (contoh `NewPassword123!`)
- `accessToken` = kosong dulu
- `verifyToken` = kosong dulu
- `resetToken` = kosong dulu
- `refreshToken` = opsional (jika tidak pakai cookie)

## 3.1) Auto Variable (Tanpa Copy Paste Token)

Tambahkan script di Postman agar token otomatis tersimpan ke environment.

### Login -> simpan access token

Request: `POST /api/v1/auth/login`  
Tab: `Tests`

```javascript
const res = pm.response.json();
const accessToken = res?.data?.tokens?.accessToken;
if (accessToken) {
  pm.environment.set("accessToken", accessToken);
}
```

### Verify Email Request -> simpan verify token (debug mode)

Request: `POST /api/v1/auth/verify-email/request`  
Tab: `Tests`

```javascript
const res = pm.response.json();
const verifyToken = res?.data?.token;
if (verifyToken) {
  pm.environment.set("verifyToken", verifyToken);
}
```

### Forgot Password -> simpan reset token (debug mode)

Request: `POST /api/v1/auth/password/forgot`  
Tab: `Tests`

```javascript
const res = pm.response.json();
const resetToken = res?.data?.token;
if (resetToken) {
  pm.environment.set("resetToken", resetToken);
}
```

### Gunakan access token otomatis untuk endpoint protected

- Atur Authorization di level collection:
  - Type: `Bearer Token`
  - Token: `{{accessToken}}`
- Untuk request seperti `/auth/me`, pilih `Inherit auth from parent`.

## 4) Urutan Test yang Disarankan

### Penting Sebelum Mulai

- `POST /register` tidak mengembalikan access token atau verify token.
- Access token pertama hanya didapat setelah `POST /login`.
- Verify token didapat dari `POST /verify-email/request`:
  - dari response `data.token` jika `AUTH_DEBUG_EXPOSE_TOKENS=true`
  - dari email inbox jika `AUTH_DEBUG_EXPOSE_TOKENS=false`

### A. Register

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/register`
- Body:

```json
{
  "name": "Postman User",
  "email": "{{email}}",
  "password": "{{password}}"
}
```

Expected: `201 Created`

Catatan:
- Endpoint ini hanya membuat akun user.
- Jika langsung login sebelum verifikasi, server akan balas `403 email not verified`.

### B. Request Email Verification

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/verify-email/request`
- Body:

```json
{
  "email": "{{email}}"
}
```

Expected: `200 OK`

Catatan:
- Jika `AUTH_DEBUG_EXPOSE_TOKENS=true`, response biasanya berisi `data.token`.
- Simpan ke variable `verifyToken`.

### C. Confirm Email Verification

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/verify-email/confirm`
- Body:

```json
{
  "token": "{{verifyToken}}"
}
```

Expected: `200 OK`

### D. Login

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/login`
- Body:

```json
{
  "email": "{{email}}",
  "password": "{{password}}"
}
```

Expected: `200 OK`

Yang perlu disimpan:
- `data.tokens.accessToken` -> simpan ke `accessToken`
- Cookie `refresh_token` (otomatis tersimpan jika cookie jar Postman aktif)

### E. Get Profile (Me)

- Method: `GET`
- URL: `{{baseUrl}}/api/v1/auth/me`
- Header:
  - `Authorization: Bearer {{accessToken}}`

Expected: `200 OK`

### F. Refresh Token

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/refresh`

Opsi 1 (disarankan): tanpa body, pakai cookie `refresh_token` dari login.  
Opsi 2 (fallback): kirim body:

```json
{
  "refreshToken": "{{refreshToken}}"
}
```

Expected: `200 OK` dan dapat `accessToken` baru.

### G. Logout

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/logout`

Opsi 1 (disarankan): pakai cookie `refresh_token`.  
Opsi 2 (fallback): body `refreshToken`.

Expected: `200 OK`

## 5) Flow Reset Password (Opsional)

### A. Request Forgot Password

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/password/forgot`
- Body:

```json
{
  "email": "{{email}}"
}
```

Jika debug expose aktif, simpan `data.token` ke `resetToken`.

### B. Reset Password

- Method: `POST`
- URL: `{{baseUrl}}/api/v1/auth/password/reset`
- Body:

```json
{
  "token": "{{resetToken}}",
  "newPassword": "{{newPassword}}"
}
```

Expected: `200 OK`

Setelah itu, login lagi pakai password baru.

## 6) Error Umum Saat Testing

- `401 invalid credentials`: email/password salah
- `403 email not verified`: akun belum verifikasi email
- `400 invalid request body`: format JSON/body tidak sesuai
- `429 too many requests`: kena rate-limit endpoint sensitif

Endpoint dengan rate-limit:
- `POST /login`
- `POST /verify-email/request`
- `POST /password/forgot`
- `POST /refresh`
