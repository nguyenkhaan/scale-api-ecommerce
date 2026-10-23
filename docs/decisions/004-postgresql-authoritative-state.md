# ADR-004: PostgreSQL giữ trạng thái nghiệp vụ có thẩm quyền

## Trạng thái

Accepted

## Ngày

2026-10-18

## Bối cảnh

SaleScale chạy nhiều API/worker instance và phải bảo toàn inventory, sale quota, purchase limit, idempotency, order và payment khi request cạnh tranh. Trạng thái trong memory không được chia sẻ giữa instance; Redis cache và RabbitMQ có thể tạm ngừng hoặc giao message lặp.

## Quyết định

- PostgreSQL primary là nguồn sự thật cho session, catalog hiện hành, inventory, flash sale counters, cart, idempotency, order/reservation, payment event, reconciliation, outbox và notification.
- Các thay đổi liên quan nhiều counter/bảng được commit trong một PostgreSQL transaction với row lock và constraint.
- Redis chỉ cache dữ liệu đọc công khai và không tham gia quyết định checkout.
- RabbitMQ chỉ vận chuyển event đã được ghi vào outbox; message không quyết định order state.
- API/worker không dùng mutex hoặc map in-memory làm cơ chế đúng đắn xuyên instance.

## Phương án đã cân nhắc

### Redis atomic counter/lock làm nguồn tồn

Có thể nhanh trong một số workload nhưng tạo dual source với order/payment và làm recovery/đối soát phức tạp hơn phạm vi đồ án. Không chọn cho MVP.

### Checkout hoàn toàn qua queue

Giảm tranh chấp tại API nhưng thay đổi trải nghiệm và mô hình đồng bộ đã chốt, đồng thời làm idempotency/feedback phức tạp. Không chọn.

### Mutex theo SKU trong process

Không hoạt động khi có nhiều API pod và mất state khi restart. Loại bỏ.

## Hệ quả

- Khóa SKU nóng có thể là giới hạn throughput thực; đó là kết quả cần đo, không được che bằng state khác.
- Pool và transaction duration phải được kiểm soát trên tổng số instance/worker.
- Integration test cần PostgreSQL thật để kiểm row lock, constraint và deadlock.
- Redis/RabbitMQ outage không được phá tính đúng đắn order; có thể làm đọc hoặc notification chậm.

