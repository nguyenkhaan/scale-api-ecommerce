# Phase 01 — Tài khoản và quyền truy cập (Lân)

**Phụ thuộc:** 00. **Mục tiêu:** `FR-001–003`, `NFR-03–05`, `TEST-24`.

## Step 1 — Đăng ký và đăng nhập

- [ ] BE: migration users/sessions; email chuẩn hóa, password hash, session token chỉ lưu hash, cookie an toàn.
- [ ] FE: form đăng ký/đăng nhập và trạng thái người dùng theo `/v1/auth/*`.
- [ ] Kiểm: input sai/đúng, cookie flags, password không lọt log/response.

## Step 2 — Logout, CSRF, quyền sở hữu 

- [ ] BE: revoke session, middleware role/owner và CSRF cho mutation dùng cookie; admin seed riêng.
- [ ] FE: route guard hiển thị phù hợp, gửi CSRF và logout; không coi route guard là quyền server.
- [ ] Kiểm: `TEST-24`, request qua hai API instance cùng session, logout có hiệu lực cả hai.

## Step 3 — Tích hợp

- [ ] BE: contract test auth error envelope và giới hạn thử login.
- [ ] FE: chạy hành trình đăng ký → login → logout, trạng thái 401/403 và reload trang.
- [ ] Checkpoint: tài khoản khách không tự nhận admin; test auth và build hai phía đạt.
