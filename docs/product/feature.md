# Chức năng BE cần có

Danh mục nhanh từ [PRD](prd.md), [FR](function-requirement.md); chi tiết ở [OpenAPI](../technical/api/openapi.yaml).

| Nhóm | Chức năng | Phase |
|---|---|---|
| Nền tảng | Go/Gin API, cấu hình, migration, seed, PostgreSQL, health/log | 00 |
| Auth | Đăng ký, login/logout, session PostgreSQL, CSRF, role/owner guard | 01 |
| Catalog | Danh mục, sản phẩm/SKU, tìm/lọc/phân trang, snapshot lịch sử | 02 |
| Inventory | Điều chỉnh có ledger; đọc on_hand/reserved/available | 02 |
| Flash sale | Nháp, lên lịch, chống chồng, dừng, quota, giới hạn mua | 03 |
| Cart/preview | Cập nhật giỏ, chọn dòng, kiểm giá, nhận người nhận | 03–04 |
| Checkout | Một transaction giữ đủ nhiều SKU, lỗi theo dòng, order snapshot | 04 |
| Idempotency | Hash/key theo user, replay, conflict, lookup sau mất phản hồi | 04 |
| Payment | Một session/đơn, simulator, xác thực/lưu callback, đối soát | 05–06 |
| Order lifecycle | Xem/hủy, expiry worker, kết thúc reservation đúng một lần | 05 |
| Notification P1 | Outbox, RabbitMQ, inbox, retry và DLQ | 06 |
| Tối ưu/vận hành | Metrics, Redis đọc công khai, rate limit, k6, nhiều instance, AKS | 07–08 |

FR-001–FR-049 và TEST-01–TEST-29 trong PRD là nguồn kiểm tra đầy đủ. Các contract P1 còn thiếu phải được chốt trước task P1.
