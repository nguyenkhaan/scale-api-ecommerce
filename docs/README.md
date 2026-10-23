# Bộ tài liệu SaleScale

Điểm vào cho người phát triển và coding agent là [CONTENT.md](CONTENT.md). Tài liệu được đọc theo thứ tự dưới đây; không sao chép quy tắc nghiệp vụ sang nhiều nơi nếu có thể liên kết bằng mã định danh.

## Thứ tự đọc

1. [PRD](product/prd.md) — nguồn sự thật cho mục tiêu, phạm vi, `BR`, `FR`, `INV`, `NFR`, `PERF`, `TEST` và quyết định sản phẩm.
2. [Project spec](spec.md) — cách làm việc, stack, lệnh, cấu trúc mã, chuẩn code và ranh giới thay đổi.
3. [Kiến trúc](technical/architecture.md) — ranh giới thành phần, nguồn dữ liệu có thẩm quyền và các luồng hệ thống.
4. [Mô hình miền](technical/domain-model.md) — thuật ngữ, aggregate, trạng thái và điều kiện chuyển trạng thái.
5. [Mô hình dữ liệu](technical/data-model.md) — lược đồ PostgreSQL P0 và cơ chế bảo toàn bất biến.
6. [OpenAPI](technical/api/openapi.yaml) — hợp đồng HTTP cho P0 và boundary P1.
7. [Chiến lược kiểm thử](technical/testing.md) — cấp kiểm thử, oracle dữ liệu và điều kiện qua cổng.
8. [Ma trận truy vết](technical/traceability.md) — nối yêu cầu với domain, dữ liệu, API và test.
9. [Architecture Decision Records](decisions/) — lý do của các quyết định khó hoặc tốn kém nếu đảo ngược.

`product/function-requirement.md` và `product/user-story.md` là các góc nhìn dễ đọc được trích từ PRD. Nếu chúng khác PRD, PRD có thẩm quyền và các góc nhìn phải được cập nhật lại.

`research/` chứa đầu vào nghiên cứu và lịch sử hình thành ý tưởng. Nội dung ở đó không tự động ghi đè quyết định đã chốt trong PRD.

## Thứ tự cập nhật

Khi thay đổi hành vi sản phẩm:

1. Cập nhật PRD và các mã `BR/FR/INV/TEST` bị ảnh hưởng.
2. Cập nhật [traceability.md](technical/traceability.md).
3. Cập nhật domain/schema, OpenAPI và test specification liên quan.
4. Viết ADR mới nếu thay đổi một quyết định kiến trúc; không xóa ADR cũ.
5. Chỉ sau khi đặc tả được duyệt mới lập kế hoạch và task triển khai.

## Trạng thái hiện tại

- Giai đoạn SDD: `SPECIFY`.
- Phạm vi chi tiết: P0 từ xác thực đến checkout, reservation, payment simulator và expiry worker.
- Phạm vi ở mức boundary/contract: RabbitMQ/outbox, notification, cache, rate limit, benchmark và AKS.
- [Kế hoạch triển khai](plans/plan.md) đang ở dạng đề xuất; chưa triển khai tính năng.

