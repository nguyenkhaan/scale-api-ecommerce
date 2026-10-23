# ADR-005: Dùng opaque session cookie với session lưu PostgreSQL

## Trạng thái

Accepted

## Ngày

2026-10-18

## Bối cảnh

API cần xác thực khách/quản trị qua nhiều instance, hỗ trợ logout/revoke và không phụ thuộc sticky session. Research ban đầu đề cập cả JWT và cookie nhưng PRD đã chốt server-side session trong PostgreSQL.

## Quyết định

- Client nhận opaque session token bằng cookie `HttpOnly`, `Secure` ở HTTPS và `SameSite` phù hợp.
- Database chỉ lưu hash token; session gắn user, expiry và thời điểm revoke.
- Mọi API riêng tư lấy danh tính/role từ session server-side; không nhận role/user ID có thẩm quyền từ client.
- Mutation dùng cookie phải có CSRF protection. Callback payment và operator endpoint dùng cơ chế xác thực riêng, không dùng browser session.
- Logout, disable user hoặc expiry có hiệu lực với mọi API instance qua PostgreSQL.

## Phương án đã cân nhắc

### JWT self-contained

Giảm database lookup nhưng revoke/logout tức thời và quản lý thay đổi role phức tạp hơn, trong khi PostgreSQL đã là dependency bắt buộc. Không chọn cho MVP.

### Session in-memory/sticky session

Đơn giản cho một instance nhưng trái mục tiêu nhiều instance và recovery. Loại bỏ.

### Session trong Redis

Khả thi nhưng làm xác thực phụ thuộc cache P1 và thêm một nguồn state bền vững cần vận hành. Không cần ở MVP.

## Hệ quả

- Auth request có thêm PostgreSQL access; có thể tối ưu sau khi đo nhưng không thay nguồn sự thật âm thầm.
- Cần cleanup/index session hết hạn và rate limit login.
- Cookie/CSRF/CORS phải được kiểm thử ở HTTP/browser level.
- Không còn JWT secret hoặc logic refresh token trong phạm vi hiện tại.

