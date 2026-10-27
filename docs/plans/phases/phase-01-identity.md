# Phase 01 — Tài khoản và quyền truy cập (Lân)

**Phụ thuộc:** 00. **Mục tiêu:** `FR-001–003`, `NFR-03–05`, `TEST-24`.

**Trạng thái kiểm tra 27/10/2026:** Triển khai xong trên nhánh `feat/phase-01-identity`; còn phần quyền sở hữu của `TEST-24`, làm khi có tài nguyên thuộc khách (giỏ, đơn).

## Step 1 — Đăng ký và đăng nhập

- [x] BE: migration users/sessions; email chuẩn hóa, password hash, session token chỉ lưu hash, cookie an toàn.
  - Bằng chứng: migration `00002_create_sessions_and_login_throttles.sql`; bcrypt cost 12; token/CSRF chỉ lưu HMAC-SHA256 có nhãn mục đích; cookie `salescale_session` HttpOnly + SameSite=Lax, Secure ngoài dev/test; session 7 ngày theo giờ database.
- [x] FE: form đăng ký/đăng nhập và trạng thái người dùng theo `/v1/auth/*`.
  - Bằng chứng: `lib/auth-api.ts`, `auth/AuthProvider.tsx`, `pages/LoginPage.tsx`, `pages/RegisterPage.tsx`; khôi phục người dùng bằng `GET /v1/auth/me` khi tải trang.
- [x] Kiểm: input sai/đúng, cookie flags, password không lọt log/response.
  - Bằng chứng: `TestRegisterRejectsInvalidPayloads`, `TestCredentialRoutesRejectNonJSONContentType`, `TestRegisterCreatesCustomerSessionWithSecureCookies`, `TestDisabledUserCannotLogIn`, `TestAuthLogsContainNoSecretsOrEmails` (kiểm cả log request của Gin, log SQL của GORM và log chuẩn).

## Step 2 — Logout, CSRF, quyền sở hữu 

- [x] BE: revoke session, middleware role/owner và CSRF cho mutation dùng cookie; admin seed riêng.
  - Bằng chứng: `RequireSession` bắt buộc `X-CSRF-Token` cho mọi method không an toàn; `RequireRole`; logout revoke trong PostgreSQL. Quyền sở hữu: service lấy user ID từ `auth.CurrentIdentity`, áp dụng khi có giỏ/đơn. Admin seed giữ từ Phase 00 và đăng nhập được qua API.
- [x] FE: route guard hiển thị phù hợp, gửi CSRF và logout; không coi route guard là quyền server.
  - Bằng chứng: `RequireAuth`, `RequireRole` (chỉ hiển thị); CSRF đọc từ cookie `salescale_csrf` nên vẫn logout được sau khi tải lại trang.
- [ ] Kiểm: `TEST-24`, request qua hai API instance cùng session, logout có hiệu lực cả hai.
  - Đạt: `TestSessionIsSharedAcrossInstancesAndLogoutAppliesToAll`, `TestLogoutRequiresCSRFAndRevokesSessionServerSide`; nửa admin route của `TEST-24` qua `TestCustomerCannotUseAdminRoute`.
  - Còn lại: khách A đọc/sửa dữ liệu của khách B — kiểm khi có giỏ/đơn (Phase 03–05).

## Step 3 — Tích hợp

- [x] BE: contract test auth error envelope và giới hạn thử login.
  - Bằng chứng: mọi test lỗi kiểm `application/problem+json`, `code` và `request_id`; `TestLoginIsThrottledAcrossInstancesAfterRepeatedFailures` (5 lần/15 phút theo email, `429` + `Retry-After`) và `TestLoginThrottleHoldsUnderConcurrentGuesses` (20 request song song qua hai instance, tối đa 5 lần được kiểm mật khẩu).
- [x] FE: chạy hành trình đăng ký → login → logout, trạng thái 401/403 và reload trang.
  - Bằng chứng: chạy trên trình duyệt với backend thật + PostgreSQL ngày 27/10/2026: `/account` khi chưa đăng nhập chuyển sang `/login?next=/account`; đăng ký rồi về `/account`; tải lại trang vẫn đăng nhập; khách vào `/admin` thấy 403; logout trả 204; sai mật khẩu hiện "Email hoặc mật khẩu không đúng".
- [x] Checkpoint: tài khoản khách không tự nhận admin; test auth và build hai phía đạt.
  - Bằng chứng: gửi `role` khi đăng ký bị `422`; backend `go build`, `go vet`, `go test` (có PostgreSQL) đạt; frontend typecheck, lint, 54 test và build đạt.

## Thay đổi hợp đồng

OpenAPI: server dùng tiền tố `/api` và cổng 4000; register/login chỉ nhận `application/json`; thêm mã `EMAIL_ALREADY_REGISTERED` (409), login `429 RATE_LIMITED`, logout `403` khi sai CSRF, cookie `salescale_csrf`; mật khẩu đăng ký tối đa 72 byte (giới hạn bcrypt). Data model: thêm bảng `login_throttles`. GORM log câu SQL lỗi/chậm không kèm giá trị tham số.

## Rủi ro đã biết

- Chưa giới hạn tải bcrypt theo IP hoặc toàn instance: đăng ký và đăng nhập với nhiều email khác nhau vẫn tốn CPU (~250 ms/lần ở cost 12) và có thể đẩy tài nguyên khỏi checkout. Xử lý cùng rate limit theo nhóm endpoint ở Phase 07.
- Đăng nhập tài khoản khác trên cùng trình duyệt không revoke session cũ; session đó không còn nằm trong trình duyệt và tự hết hạn sau 7 ngày.
- Chưa có tác vụ dọn `login_throttles` và session hết hạn; thêm vào worker định kỳ.
