# Phase 00 — Thiết lập nền tảng

**Phụ thuộc:** SPECIFY và PLAN được duyệt; chốt TECH-01. **Đầu ra:** bộ khung chạy được. Mỗi Step tách thành task 1–5 file khi vào TASKS.

## Step 1 — Dựng cấu trúc dự án

- [ ] BE: tạo `cmd/api`, `cmd/expiry-worker`, `internal/`, cấu hình môi trường và health endpoint; giữ Go module chạy được.
- [ ] FE: tạo `app/frontend` React + Vite, router, layout cơ bản, API client theo OpenAPI và trang thử kết nối.
- [ ] Kiểm: `go build ./...`, `go test ./...`, frontend build/test theo manifest mới; ghi lệnh thật vào spec.

## Step 2 — PostgreSQL và môi trường local

- [ ] BE: chọn migration tool/driver theo TECH-01, kết nối PostgreSQL, chạy migration smoke test và chuẩn bị seed admin được bảo vệ cho phase 01.
- [ ] FE: cấu hình base URL/proxy, xử lý lỗi mạng và trạng thái tải chung; dùng fixture cùng schema API.
- [ ] Kiểm: Compose khởi động API + PostgreSQL, migration chạy lặp an toàn trên DB thử nghiệm, FE gọi health.

## Step 3 — Cổng chất lượng đầu tiên

- [ ] BE: thiết lập test DB, formatter/static checks và log `request_id`; không log secret.
- [ ] FE: thiết lập build, lint/test và kiểm tra màn hình nhỏ/bàn phím cơ bản.
- [ ] Checkpoint: tài liệu lệnh/seed/reset được cập nhật; hai nhánh tích hợp trên môi trường local. `FR-046`, `NFR-11` có bằng chứng ban đầu.
