# Phase 02 — Checkout đúng và hộp thông báo E0

**Thời gian:** Tuần 4–6  
**Phụ thuộc:** Phase 01  
**Kết quả cần đạt:** Đặt giỏ nhiều SKU nguyên tử, xử lý thanh toán/hết hạn an toàn và cho khách xem thông báo về trạng thái đơn.

## Nghiên cứu cần làm

- [ ] **R02.1 — Tìm hiểu transaction và khóa PostgreSQL.** Đọc tài liệu [khóa hàng](https://www.postgresql.org/docs/current/explicit-locking.html), rồi vẽ thao tác đọc, khóa và ghi tồn. Khóa SKU theo cùng thứ tự để hai giỏ giao nhau ít deadlock hơn và không để lại reservation một phần.
- [ ] **R02.2 — Thiết kế idempotency cho checkout.** Dùng unique constraint PostgreSQL để gắn idempotency key với tài khoản và nội dung đơn. Nếu khách mất response rồi retry, API trả kết quả cũ thay vì giữ hàng và tạo đơn thứ hai.
- [ ] **R02.3 — Vẽ trạng thái đơn và thanh toán.** Dùng Mermaid/draw.io mô tả pending, confirmed, failed, cancelled và expired. Sơ đồ giúp Go/Gin, payment callback và worker xử lý đúng khi sự kiện tới trễ hoặc nhiều việc cùng kết thúc một đơn.
- [ ] **R02.4 — Chốt công thức tồn/quota.** Dùng PostgreSQL phân biệt on_hand, reserved, quota_reserved và quota_sold trước khi code. Công thức rõ ràng giúp callback, hủy và worker không cộng/trừ cùng một reservation hai lần.
- [ ] **R02.5 — Tách timeout khỏi lỗi nghiệp vụ.** Dùng Go context và trạng thái UNKNOWN của payment simulator để xác định khi nào tra cứu kết quả. Mất kết nối không được hiểu nhầm thành hết hàng hoặc thanh toán thất bại.
- [ ] **R02.6 — Chốt event của thông báo đơn.** Thống nhất năm loại event: `order.created` (đơn đã tạo và giữ hàng), `order.confirmed`, `order.failed`, `order.cancelled` và `order.expired`. Event chỉ phục vụ hộp thông báo khách; nó không được dùng thay trạng thái chính thức trong PostgreSQL.

## Việc triển khai

- [ ] **C02.1 — Tạo bảng đơn và reservation.** Dùng golang-migrate/PostgreSQL tạo orders, order_items, reservations và idempotency records; lưu giá/tên tại lúc mua để đơn không đổi khi admin cập nhật sản phẩm.
- [ ] **C02.2 — Tạo checkout trong Go/Gin.** Chuẩn hóa SKU rồi dùng một transaction PostgreSQL khóa và kiểm tra giá, tồn, quota, giới hạn mua trước khi ghi đơn. Không gọi payment simulator trong transaction để không giữ khóa khi chờ HTTP.
- [ ] **C02.3 — Giữ toàn bộ giỏ bằng transaction.** Dùng PostgreSQL commit tất cả reservation cùng lúc hoặc rollback toàn bộ khi một dòng thiếu hàng/giá đổi. Nhờ vậy không phát sinh đơn giữ một phần giỏ.
- [ ] **C02.4 — Làm retry an toàn.** Dùng idempotency key, unique constraint và nội dung đơn đã chuẩn hóa trong PostgreSQL; cùng key/cùng nội dung trả đơn cũ, cùng key/nội dung khác bị báo xung đột.
- [ ] **C02.5 — Tạo payment simulator.** Dùng Go/Gin và net/http tạo kết quả thành công, thất bại, delay, timeout và callback lặp. Đây là cách thử luồng thanh toán bất đồng bộ mà không dùng tiền thật.
- [ ] **C02.6 — Ghi nhận callback thanh toán.** Dùng Go/Gin xác thực mã đơn/phiên, số tiền, tiền tệ và event ID rồi cập nhật PostgreSQL trong transaction. Callback lặp chỉ được áp dụng một lần; sự kiện muộn được ghi để đối soát.
- [ ] **C02.7 — Trả hàng hết hạn.** Dùng Go worker, pgxpool và PostgreSQL để quét đơn quá hạn theo lô, khóa an toàn khi có nhiều worker và giải phóng reservation đúng một lần. Nhờ vậy hàng không bị giữ mãi khi khách bỏ thanh toán.
- [ ] **C02.8 — Làm hộp thông báo E0 trong PostgreSQL.** Khi trạng thái đơn đổi, cùng transaction PostgreSQL ghi một notification cho chủ đơn; Go/Gin cung cấp API list/mark-as-read. Ghi trực tiếp tạo đường cơ sở cho feature trước khi chuyển nó sang RabbitMQ ở Phase 04.
- [ ] **C02.9 — Hiển thị notification trong React/Vite.** Thêm mục thông báo vào trang tài khoản/đơn và lấy dữ liệu từ API Go/Gin. Giao diện cho khách kiểm tra kết quả, nhưng trạng thái đơn vẫn lấy từ PostgreSQL.

## Tiêu chí hoàn thành

- [ ] TEST-01..21 và TEST-25 kiểm chứng tồn/đơn trực tiếp trong PostgreSQL.
- [ ] Hai khách mua món cuối, giỏ giao nhau, cùng idempotency key, callback lặp và worker restart không làm sai dữ liệu.
- [ ] Khách xem được notification của mình; tài khoản khác không đọc được.
- [ ] Checkout thành công, thiếu hàng, đổi giá, payment failure và callback muộn có kết quả rõ.
- [ ] E0 notification được ghi đồng bộ với trạng thái đơn; retry không gia hạn reservation 5 phút.

**Đưa vào báo cáo:** Transaction checkout, state machine, idempotency và lý do tạo baseline thông báo trực tiếp trước khi áp dụng message queue.
