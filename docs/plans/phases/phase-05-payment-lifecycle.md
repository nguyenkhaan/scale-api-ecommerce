# Phase 05 — Thanh toán mô phỏng và vòng đời đơn (Lân)

**Phụ thuộc:** 04; chốt TECH-02 và OPEN-02 trước nghiệm thu. **Mục tiêu:** `FR-030–040`, `INV-06–08/11/12`, `TEST-11–19/25`.

**Trạng thái kiểm tra 22/11/2026:** Step 1–2 triển khai xong trên nhánh `feat/phase-05-payment-lifecycle`, làm song song với Phase 04 (checkout chưa merge): test và e2e tạo đơn chờ thanh toán bằng SQL đúng như checkout để lại. TECH-02 đã chốt (ADR-008); OPEN-02 vẫn mở. Backend 175 test (PostgreSQL), frontend 177 test, e2e qua API 28/28 bước, đã chạy hành trình thanh toán/hủy trên trình duyệt. Sau review đã sửa: deadlock giữa xác nhận đơn và checkout (khóa dòng giỏ trước inventory), xác nhận sau hạn khi chờ khóa (kiểm lại bằng `clock_timestamp()`), phiên thanh toán tạo trong lúc đơn đang bị hủy, worker kẹt sau đơn lỗi.

## Step 1 — Phiên và callback

- [x] BE: một payment session/đơn; simulator tạo success/failure/delay/duplicate; callback xác thực nguồn, phiên, tiền, event ID và lưu bền vững.
  - Bằng chứng: `POST /orders/{id}/payment-session` trả phiên cũ khi gọi lại, không gia hạn; callback ký HMAC-SHA256 có timestamp (ADR-008), lưu event trước khi trả 202; simulator gửi qua HTTP tới callback thật với delay/lặp/mất callback/sai chữ ký/sai số tiền.
- [x] FE: mở lại cùng phiên, theo dõi trạng thái từ order API; điều khiển simulator chỉ hiện ở môi trường test đúng quyền.
  - Bằng chứng: trang `/orders/:id/payment` poll trạng thái đơn mỗi 2 giây; `/admin/payment-simulator` chỉ cho admin, backend chỉ bật simulator ở dev/test.
- [x] Kiểm: `TEST-11/12/17/18`, event lặp/sai/mâu thuẫn không trừ tồn thêm.
  - Bằng chứng: `test/payment_callback_test.go`, `payment_callback_security_test.go`; e2e: gửi thành công 3 lần chỉ trừ tồn 1 lần; sai chữ ký được lưu `rejected` nhưng sự kiện thật cùng ID vẫn xác nhận được.

## Step 2 — Kết thúc reservation

- [x] BE: một đường transition có guard cho confirmed/failed/cancelled/expired; worker quét batch từ PostgreSQL, chạy lại an toàn; case đối soát cho success muộn.
  - Bằng chứng: `order.Finish` là đường duy nhất kết thúc đơn, khóa theo thứ tự của data-model, đơn quá hạn theo giờ database luôn thành `expired`; `cmd/expiry-worker` dùng `FOR UPDATE SKIP LOCKED`, mỗi đơn một transaction ngắn.
- [x] FE: danh sách/chi tiết đơn, hủy hợp lệ, hiển thị expired dù worker chưa cập nhật và trạng thái cần đối soát.
  - Bằng chứng: `/orders`, `/orders/:id` dùng `payable` và `reconciliation_required` từ server.
- [x] Kiểm: `TEST-13–16/19/25`, callback đối đầu hủy/hết hạn, restart và nhiều worker; oracle INV-06–08/12.
  - Bằng chứng: `test/order_lifecycle_test.go` (xác nhận và hủy đồng thời chỉ một bên thắng; chỉ xóa dòng giỏ chưa sửa), `order_api_cancel_test.go`, `order_expiry_test.go` (3 worker, 36 đơn, mỗi đơn hết hạn đúng một lần), `payment_callback_race_test.go`; oracle INV-02/03/08/09/10/11 sau mỗi kịch bản.

## Step 3 — Baseline đúng nghiệp vụ

- [ ] BE: hoàn thiện metrics order/payment/expiry, seed/reset và oracle SQL đầy đủ.
  - Đã có: oracle SQL trong test, worker log số liệu mỗi chu kỳ. Còn lại: metrics Prometheus (cần chốt thêm dependency).
- [ ] FE: chạy hành trình khách từ xem hàng tới kết quả đơn; kiểm trạng thái tải/lỗi.
  - Chờ checkout (Phase 04) để nối giỏ → đặt đơn → thanh toán.
- [ ] Checkpoint: `TEST-01–25` qua trong lab, INV-01–12 không vi phạm; mới được đo E0.

## Việc checkout (Phase 04) cần làm để khớp Phase 05

- Ghi `order_items.cart_item_id` và `cart_item_version` của dòng giỏ đã đặt; thiếu thì thanh toán thành công sẽ giữ nguyên giỏ (TEST-25).
- Giữ thứ tự khóa inventory → flash sale item → usage theo ID tăng dần như data-model, vì `order.Finish` dùng cùng thứ tự.
- Migration `00010` thêm `orders.updated_at`, `orders.terminal_at` (NULL khi còn `pending_payment`): đơn mới tạo không cần ghi hai cột này.
