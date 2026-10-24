# Phase 00 — Thiết lập nền tảng

**Phụ thuộc:** SPECIFY và PLAN được duyệt; chốt TECH-01. **Đầu ra:** bộ khung chạy được. Mỗi Step tách thành task 1–5 file khi vào TASKS.

**Trạng thái kiểm tra 24/10/2026:** Hoàn thành — `9/9` hạng mục đạt phạm vi Phase 00. TECH-01 được ghi nhận tại ADR-007.

## Step 1 — Dựng cấu trúc dự án

- [x] BE: tạo `cmd/api`, `cmd/expiry-worker`, `internal/`, cấu hình môi trường và health endpoint; giữ Go module chạy được.
  - Bằng chứng: `go build ./...`, `go test ./...`, `go vet ./...` và kiểm tra `gofmt` đều đạt.
- [x] FE: tạo `app/frontend` React + Vite, router, layout cơ bản, API client theo OpenAPI và trang thử kết nối.
  - Bằng chứng: `App.tsx`, `HealthPage.tsx`, `health-api.ts` và route `/health`; TypeScript dùng duy nhất `tsconfig.json`.
- [x] Kiểm: `go build ./...`, `go test ./...`, frontend build/test theo manifest mới; ghi lệnh thật vào spec.
  - Bằng chứng: backend build/test/vet đạt; frontend lint, 6 tests, typecheck và production build đạt; lệnh được cập nhật trong `docs/spec.md`.

## Step 2 — Neon PostgreSQL và môi trường ứng dụng local

- [x] BE: chốt migration tool/driver theo TECH-01, kết nối Neon PostgreSQL và chuẩn bị seed admin được bảo vệ cho phase 01.
  - Bằng chứng: ADR-007 chốt GORM/pgx + Goose; có migration nhúng, command `migrate`, command `seed-admin` và guard test.
- [x] FE: cấu hình base URL/proxy, xử lý lỗi mạng và trạng thái tải chung theo schema health API.
  - Bằng chứng: `VITE_API_BASE_URL`, `VITE_API_PROXY_TARGET`, client kiểm tra payload và UI loading/success/error.
- [x] Kiểm: backend và frontend compile; API client/proxy được kiểm tra bằng test tự động, không yêu cầu PostgreSQL local hoặc Docker Compose trong Phase 00.
  - Bằng chứng: toàn bộ backend gates và frontend gates đạt ngày 24/10/2026.

## Step 3 — Cổng chất lượng đầu tiên

- [x] BE: thiết lập unit test, formatter/static checks và log `request_id`; không log secret.
  - Bằng chứng: `gofmt`, `go test`, race test, `go vet`, build đều đạt; request ID có test và logger không ghi query/header/connection string.
- [x] FE: thiết lập build, lint/test và kiểm tra màn hình nhỏ/bàn phím cơ bản.
  - Bằng chứng: responsive layout, native retry button, accessible status markup và 6 frontend tests đều đạt.
- [x] Checkpoint: tài liệu setup/migration/seed/reset/test được cập nhật; frontend và backend có chung health contract. `FR-046` và phần nền tảng của `NFR-11` có bằng chứng ban đầu.
  - Bằng chứng: README và spec có lệnh vận hành hiện tại; phạm vi triển khai hoàn chỉnh của `NFR-11` vẫn thuộc Phase 08.
