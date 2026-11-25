# Phase 04 — Checkout nhiều SKU nguyên tử (An)

**Phụ thuộc:** 03. **Mục tiêu:** `FR-021–029`, `INV-01–05/09–11`, `TEST-01–10/25`. Đây là lát cắt rủi ro cao nhất.

**Rà soát mã nguồn 25/11/2026 (nhánh `dev`):** `0/9` — **chưa triển khai**. Chỉ có phần nền: migration `00009_create_orders.sql` (bảng `orders`, `order_items`, `inventory_reservations`, `idempotency_records`) và model GORM tương ứng, dùng chung với Phase 05. Chưa có endpoint đặt đơn (`POST /orders`), service giữ hàng hay xử lý idempotency key. Nút "Đặt hàng (Phase 04)" trên giỏ hàng vẫn bị khóa.

## Step 1 — Order và reservation

- [ ] BE: migration order/items/reservations/idempotency; khóa SKU theo thứ tự ổn định; giữ đủ hoặc rollback toàn bộ.
- [ ] FE: form người nhận, xác nhận tổng/nguồn giá; hiển thị lỗi theo SKU và giữ giỏ khi bị từ chối.
- [ ] Kiểm: `TEST-01–04`, `08–10`, oracle INV sau cạnh tranh, không partial order.

## Step 2 — Retry an toàn

- [ ] BE: canonical hash, unique user/operation/key, lưu business rejection và response ổn định, endpoint tra cứu kết quả.
- [ ] FE: giữ key cho một ý định qua timeout/reload; tra cứu hoặc retry cùng key; tạo key mới sau khi sửa lựa chọn.
- [ ] Kiểm: `TEST-05–07`, 2 API instance, mất response sau commit, cùng key khác payload.

## Step 3 — Cổng checkout

- [ ] BE: snapshot giá/người nhận, hạn 5 phút cố định, metric new/replay/rejection/error.
- [ ] FE: trang kết quả hiện mã đơn/hạn và trạng thái chưa rõ; không hiển thị đơn thành công giả.
- [ ] Checkpoint: `TEST-01–10/25`, oracle INV-01–05/09–11 và HTTP contract qua trước payment/benchmark.
