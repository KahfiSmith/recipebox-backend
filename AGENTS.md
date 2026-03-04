# Repository Guidelines (Backend Go - RecipeBox)

Dokumen ini adalah panduan implementasi untuk agent/contributor di repo ini.
Tujuan: konsisten, cepat, dan tidak over-engineering.

## Source of Truth (Baca Dulu)
- API summary: `docs/api.md`
- Arsitektur ringkas: `docs/architecture.md`
- Route registry: `internal/routes/*.go`
- Controller contract (Swagger annotations): `internal/controller/*.go`
- Model & repository contract: `internal/models/*.go`, `internal/repository/*.go`
- Migration SQL: `migrations/*.sql`

Jika ada konflik, ikuti behavior code yang sedang berjalan + route aktif di router.

## Non-Negotiables (Hard Rules)
- Gunakan alur: `routes -> controller -> service -> repository -> models`.
- Jangan lompat layer (controller tidak query DB langsung).
- Response API konsisten:
  - success: `{"data": ...}` atau `{"message": "..."}`
  - error: `{"error": "..."}`
- Status code konsisten:
  - `400` input/validasi
  - `401` unauthorized
  - `404` not found
  - `409` conflict
  - `500` internal error
- Jangan ubah kontrak endpoint existing tanpa menyebut dampak backward compatibility.
- Untuk ubah schema DB: tambah migration baru, jangan edit migration lama yang sudah dipakai environment lain.
- Jangan commit secret/credential ke git.

## Engineering Principles (Principal Engineer Mindset)
- **Correctness over speed**: lebih baik perubahan kecil tapi benar daripada cepat tapi regress.
- **Design for maintainability**: kode harus mudah dibaca engineer lain dalam 3-6 bulan ke depan.
- **Stability first**: optimasi/perubahan besar hanya jika ada kebutuhan nyata (data, error, atau scale).
- **Explicit trade-off**: jika memilih solusi cepat, tulis batasan dan risiko secara jelas.
- **Backward compatibility by default**: endpoint/payload lama dijaga kecuali ada keputusan breaking yang disetujui.
- **Observability-aware changes**: setiap perubahan behavior penting harus tetap bisa diverifikasi dari log, status code, atau test.
- **Secure by default**: validasi input, batasi surface error, dan hindari leak data sensitif.
- **Simplicity is a feature**: hindari abstraksi tambahan jika belum memberi nilai jelas.

## Core Design Principles (DRY, SOLID, KISS)
- **DRY (Don't Repeat Yourself)**:
  - Hindari duplikasi logic lintas controller/service/repository.
  - Ekstrak helper hanya jika dipakai berulang (>=2-3 tempat) dan benar-benar meningkatkan kejelasan.
- **SOLID (practical use)**:
  - Single Responsibility: tiap layer fokus tugasnya sendiri.
  - Open/Closed: perluas behavior via method/service baru, minim ubah flow lama.
  - Liskov + Interface Segregation: kontrak interface repository harus kecil, jelas, dan tidak memaksa implementasi tidak relevan.
  - Dependency Inversion: service bergantung pada interface repository, bukan implementasi konkret.
- **KISS (Keep It Simple, Stupid)**:
  - Pilih solusi paling sederhana yang memenuhi kebutuhan saat ini.
  - Hindari pattern/abstraction berlebihan untuk kasus kecil.
  - Jika ragu antara desain kompleks vs sederhana, pilih yang sederhana dan mudah diuji.

## Architecture Notes (Simple but Strict)

### Layer Responsibilities
- `routes`: registrasi path + method + middleware.
- `controller`: HTTP boundary, parse request, mapping error -> status code.
- `service`: validasi input + business rules.
- `repository`: data access (GORM/SQL), tidak tahu HTTP.
- `models/dto`: shape data persistence dan transport.

### Dependency Direction
`controller -> service -> repository -> models`

## API Contract Discipline
- Jika endpoint berubah, sinkronkan:
  1. route
  2. handler/controller
  3. dto request/response
  4. docs/api.md
  5. swagger annotation
- Semua endpoint auth-protected harus tetap pakai middleware auth pada route group.

## Swagger & Docs Rules
- Header swagger ada di `cmd/api/main.go`.
- Anotasi endpoint ada di controller (`// @Summary`, `@Param`, `@Success`, `@Router`, dst).
- Setelah ubah anotasi/endpoint, regenerate:
  - `bash scripts/swagger-generate.sh`
- Jika `openapi.yaml` dipakai sebagai canonical source di repo, pastikan file itu tetap ada dan tidak tertinggal update.

## Migration Rules
- Gunakan file migration SQL di `migrations/` untuk schema change.
- Jalankan verifikasi migration:
  - `bash scripts/migrate-up.sh`
  - `bash scripts/migrate-status.sh`
  - `bash scripts/migrate-down.sh 1` (bila perlu rollback test)

## Testing Rules
Minimal sebelum handoff:
- `go test ./...`
- uji manual endpoint yang diubah (curl/Postman)
- jika endpoint berubah, cek docs yang ter-generate

Jika local environment tidak punya Go/tooling, agent wajib jujur dan tulis command verifikasi untuk user.

## Implementation Patterns

### Pattern A: Tambah CRUD sederhana
1. Tambah route method (`POST/PUT/DELETE` sesuai kebutuhan)
2. Tambah handler controller
3. Tambah method service + validasi minimum
4. Tambah method repository + query by `user_id` bila resource scoped user
5. Tambah/adjust test yang terdampak
6. Update docs

### Pattern B: Update behavior endpoint existing
- Hindari breaking change payload lama.
- Tambah field optional jika butuh evolusi response.
- Pertahankan message/error style existing.

### Pattern C: Error Handling
- Gunakan error domain (`not found`, `validation`) di service/repository.
- Mapping ke HTTP hanya di controller.

## Anti Over-Engineering Rules
- Jangan pecah modul/abstraction baru untuk perubahan kecil.
- Jangan tambah dependency eksternal tanpa alasan jelas.
- Jangan refactor area yang tidak diminta user.
- Prioritaskan perbaikan yang small, clear, dan testable.

## Output Contract (untuk Agent)
Saat memberi hasil ke user, urutan wajib:
1. Solusi singkat
2. File yang diubah
3. Cara verifikasi
4. Risiko/hal belum diverifikasi

## Quick Command Reference
```bash
# tests
go test ./...

# swagger
bash scripts/swagger-generate.sh

# migration
bash scripts/migrate-up.sh
bash scripts/migrate-status.sh
bash scripts/migrate-down.sh 1
```
