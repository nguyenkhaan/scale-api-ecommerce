# Phase 05 — Chia tải bằng NGINX và mở rộng API: E2

**Thời gian:** Tuần 8–9  
**Phụ thuộc:** Phase 04  
**Kết quả cần đạt:** Biết nhiều API có tăng số đơn hợp lệ khi tổng tài nguyên API giữ nguyên hay không, đồng thời bảo đảm session và hộp thông báo dùng chung dữ liệu.

## Nghiên cứu cần làm

- [ ] **R05.1 — So sánh thuật toán NGINX.** Đọc [NGINX load balancing](https://nginx.org/en/docs/http/load_balancing.html) để so round-robin với least_conn. Chọn bằng thử nghiệm cách chia request catalog/checkout khi thời gian xử lý giữa các request khác nhau.
- [ ] **R05.2 — Hiểu upstream keepalive.** Đọc [NGINX upstream module](https://nginx.org/en/docs/http/ngx_http_upstream_module.html); keepalive giữ tối đa N kết nối rảnh từ mỗi NGINX worker tới API để dùng lại; keepalive_timeout quyết định giữ kết nối rảnh bao lâu. Đây không phải giới hạn tổng request đang chạy.
- [ ] **R05.3 — Xác nhận dữ liệu dùng chung.** Dùng session/giỏ/đơn PostgreSQL, notification inbox và RabbitMQ outbox để kiểm tra request qua API nào cũng nhìn thấy cùng trạng thái. Nhờ vậy không cần sticky session để giữ đúng nghiệp vụ.
- [ ] **R05.4 — Đặt chính sách retry.** Nghiên cứu NGINX proxy retry và Go HTTP cho POST checkout/callback. Chỉ retry khi có thể chứng minh an toàn bằng idempotency key, vì API có thể đã commit đơn trước lúc phản hồi mất.

## Việc triển khai và đo

- [ ] **C05.1 — Chạy 1, 2 rồi 4 Go/Gin API.** Dùng Docker Compose đưa các API sau cùng một NGINX và dùng chung PostgreSQL/Redis/RabbitMQ. So số đơn hợp lệ mà không đổi nghiệp vụ.
- [ ] **C05.2 — Cấu hình NGINX upstream.** Dùng upstream server list, keepalive N, keepalive_timeout, health endpoint, request ID và timeout để chuyển traffic tới API khỏe. Thử các giá trị keepalive; access log cho biết request đi đâu và có được connection reuse không.
- [ ] **C05.3 — Thử round-robin và least_conn.** Dùng cùng workload k6 và cấu hình API để so phân phối request, checkout p95/p99, số lỗi và đơn hợp lệ. Chọn thuật toán theo kết quả chứ không theo tên gọi.
- [ ] **C05.4 — Giữ connection budget PostgreSQL.** Dùng pgxpool đặt MaxConns mỗi API sao cho tổng API + expiry/outbox worker + notification worker nằm trong giới hạn DB. Việc này tránh kết quả xấu giả do nhân bản quá nhiều connection.
- [ ] **C05.5 — So E1/E2 cùng ngân sách máy.** Dùng Docker Compose CPU/RAM limits để giữ tổng tài nguyên dành cho API bằng nhau khi chạy 1/2/4 tiến trình; giữ DB, RabbitMQ, Redis, máy k6, seed và tải không đổi.
- [ ] **C05.6 — Kiểm tra notification trên mọi API.** Dùng PostgreSQL session/inbox và RabbitMQ consumer để tạo đơn qua API 1 rồi xem notification qua API 2. Điều này xác nhận feature không phụ thuộc bộ nhớ của một tiến trình.
- [ ] **C05.7 — Tắt một API khi đang tải.** Dùng Docker Compose dừng một API và quan sát NGINX, kết quả checkout, log và notification event; bước này đo ảnh hưởng khi một tiến trình lỗi, không phải kiểm chứng HA production.

## Tiêu chí hoàn thành

- [ ] Session, giỏ, idempotency, trạng thái đơn và hộp thông báo hoạt động qua nhiều API.
- [ ] Có số liệu cho 1/2/4 API: đơn hợp lệ/giây, p95/p99, lỗi, CPU/RAM, chờ DB và phân phối NGINX.
- [ ] RabbitMQ worker count giữ nguyên trong bài đo API scaling để không trộn hai thay đổi.
- [ ] So sánh 1/2/4 API trong cùng tổng CPU/RAM; mọi thử nghiệm tăng máy được báo riêng.
- [ ] Chỉ ra giới hạn tăng API ở DB, lock SKU, gateway, queue hay CPU bằng số liệu.

**Đưa vào báo cáo:** Cấu hình NGINX, phép so thuật toán chia tải, kết quả cùng hạ tầng và kiểm chứng trạng thái dùng chung.
