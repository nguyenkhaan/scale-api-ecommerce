# Kế hoạch triển khai SaleScale

**Trạng thái:** Đề xuất để hai thành viên duyệt. Chưa đánh dấu hoàn thành checklist nào. Nguồn: [PRD](../product/prd.md), [spec](../spec.md), [truy vết](../technical/traceability.md), [kiểm thử](../technical/testing.md).

## Cách thực hiện

Đi lần lượt theo phase, nhưng trong mỗi phase FE và BE thực hiện song song sau khi thống nhất payload OpenAPI. FE có thể dùng dữ liệu mẫu đúng schema trước khi API sẵn sàng. Tại checkpoint, bỏ dữ liệu mẫu và chạy luồng thật. Trọng tâm thời gian và test dành cho BE checkout, idempotency, payment, expiry và benchmark. Chia mỗi Step thành task một phiên, khoảng 1–5 file; task phải nêu FR/INV/TEST, điều kiện đạt và lệnh/chứng cứ xác minh.

| Phase | Kết quả kiểm được | Phụ thuộc |
|---|---|---|
| [00 — Nền tảng](phases/phase-00-foundation.md) | FE/BE chạy, PostgreSQL/migration/test lab sẵn sàng | Spec được duyệt |
| [01 — Tài khoản](phases/phase-01-identity.md) | Đăng ký/login/logout và phân quyền qua nhiều instance | 00 |
| [02 — Catalog và tồn](phases/phase-02-catalog-inventory.md) | Quản trị tạo hàng, khách xem hàng, tồn có ledger | 01 |
| [03 — Flash sale và giỏ](phases/phase-03-sale-cart.md) | Cấu hình sale, giỏ nhiều SKU, preview cơ bản | 02 |
| [04 — Checkout](phases/phase-04-checkout.md) | Đơn nguyên tử, idempotency, lỗi theo dòng | 03 |
| [05 — Payment và expiry](phases/phase-05-payment-lifecycle.md) | Luồng mua đầu cuối và TEST-01–25 | 04 |
| [06 — P1 sản phẩm](phases/phase-06-p1-product.md) | Thông báo, đối soát, phần quản trị còn lại | 05 |
| [07 — Đo và tối ưu](phases/phase-07-performance.md) | E0–E4, Redis, rate limit, 1–n instance | 05; 06 cho E1 |
| [08 — Triển khai và bàn giao](phases/phase-08-delivery.md) | Compose/AKS demo, tài liệu tái lập, nghiệm thu | 06–07 |

## Cổng SDD và quyết định còn mở

1. **SPECIFY:** hai thành viên duyệt PRD/spec/contract P0; giải quyết TECH-01 trước bootstrap, TECH-02 trước simulator. `OPEN-02` chốt trước nghiệm thu payment. P1 contract thiếu phải chi tiết hóa trước task P1.
2. **PLAN:** duyệt thứ tự phase, ranh giới FE/BE, tài nguyên lab và checkpoint này trước khi tạo task triển khai. Kế hoạch không tự phê duyệt đặc tả.
3. **TASKS:** tách từng Step thành task vừa một phiên, có mã yêu cầu, phụ thuộc, file dự kiến, acceptance và verification; duyệt task trước khi code.
4. **IMPLEMENT:** test và oracle dữ liệu sau mỗi lát cắt; không dùng benchmark thay bằng chứng đúng nghiệp vụ.

`OPEN-01` chốt tải/cấu hình sau E0; `OPEN-03` chốt mục tiêu thông báo sau E0/E1. Mục tiêu DEC-14/15 không tự đổi. Rủi ro chính: khóa SKU nóng, deadlock, callback cạnh tranh expiry, pool tăng theo replica, cache miss dồn DB; mỗi phase tương ứng có test/fault/metric.

Hoàn tất dự án khi TEST-01–29 và hành trình trình duyệt đạt hoặc P1 được giảm phạm vi bằng quyết định đã duyệt; INV-01–12 không vi phạm; E0–E4 và giới hạn kết luận được báo rõ; hướng dẫn dựng lại môi trường chạy được.
