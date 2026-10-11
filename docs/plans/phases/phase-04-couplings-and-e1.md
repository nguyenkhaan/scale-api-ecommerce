# Phase 04 — RabbitMQ, outbox và tối ưu các kết nối: E1

**Thời gian:** Tuần 7–8  
**Phụ thuộc:** Phase 03  
**Kết quả cần đạt:** Chuyển hộp thông báo đơn từ ghi trực tiếp sang xử lý qua RabbitMQ, đồng thời dùng số liệu để chọn 3–5 đường gọi đáng tối ưu.

## Nghiên cứu cần làm

- [ ] **R04.1 — Nghiên cứu RabbitMQ Work Queue.** Đọc [tutorial Go Work Queues](https://www.rabbitmq.com/tutorials/tutorial-two-go) để hiểu publisher, consumer, ACK và cách chia message cho nhiều worker. Áp dụng một queue bền vững và một loại notification consumer; chỉ tăng số instance consumer khi số liệu cho thấy queue bị tồn.
- [ ] **R04.2 — Nghiên cứu publisher confirm và ACK.** Đọc [RabbitMQ confirms](https://www.rabbitmq.com/docs/confirms) và [Reliable Publishing với Go](https://www.rabbitmq.com/tutorials/tutorial-seven-go). Outbox publisher chỉ đánh dấu event đã gửi sau publisher confirm; notification consumer chỉ ACK sau khi transaction ghi notification đã commit.
- [ ] **R04.3 — Nghiên cứu transactional outbox.** Đọc [AWS Prescriptive Guidance về transactional outbox](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html), rồi dùng PostgreSQL transaction cùng tài liệu RabbitMQ reliability để thiết kế hàng outbox: cùng transaction cập nhật đơn và lưu event; outbox publisher gửi event sau commit. Điều này tránh mất thông báo nếu DB commit thành công nhưng broker đang dừng.
- [ ] **R04.4 — Đọc các pool hiện có.** Nghiên cứu Go net/http Transport/httptrace, pgxpool, go-redis và amqp091-go đúng phiên bản trong repo. Mỗi thư viện giữ kết nối khác nhau, vì vậy đo từng pool riêng thay vì sao chép một con số chung.
- [ ] **R04.5 — Vẽ đường xử lý checkout/notification.** Dùng Go trace, httptrace, log PostgreSQL và RabbitMQ metrics để phân thời gian API, DB, payment simulator, outbox, broker và notification consumer. Bản đồ giúp xác định lời gọi nào nằm trên đường request và lời gọi nào có thể chạy nền.

## Thực hiện từ 3 đến 5 phép thử kết nối

- [ ] **C04.1 — API ↔ PostgreSQL qua pgxpool.** Thử MaxConns/MinConns và thời gian chờ lấy kết nối với checkout. Cộng giới hạn của tất cả API, expiry worker, outbox publisher và notification worker để DB không bị mở quá nhiều kết nối.
- [ ] **C04.2 — API ↔ payment simulator qua HTTP.** Dùng một http.Client/http.Transport chung; thử riêng MaxIdleConns, MaxIdleConnsPerHost, MaxConnsPerHost, Dialer.Timeout, TLSHandshakeTimeout, ResponseHeaderTimeout và Client.Timeout. Dùng httptrace đo thời gian chờ/reuse khi mở phiên thanh toán, tránh pool quá nhỏ hoặc giữ request vô hạn.
- [ ] **C04.3 — API ↔ Redis qua go-redis.** Dùng một go-redis client dùng chung để đọc catalog; căn cứ đúng phiên bản thư viện mà chỉnh PoolSize, MinIdleConns và PoolTimeout. Đo phần thời gian Redis và ảnh hưởng lên DB trước khi bật cache ở Phase 06.
- [ ] **C04.4 — Outbox publisher ↔ PostgreSQL/RabbitMQ.** Dùng PostgreSQL outbox để lưu `event_id`, `order_id`, `event_type` và `occurred_at` cùng transaction đổi trạng thái; không đưa thông tin người nhận vào message. Outbox publisher đọc event chưa gửi, publish message bền vững qua amqp091-go và chờ publisher confirm trước khi đánh dấu đã gửi. API, payment callback và expiry worker chỉ ghi outbox trong transaction, không publish trực tiếp lên broker.
- [ ] **C04.5 — Notification worker ↔ PostgreSQL/RabbitMQ.** Dùng amqp091-go manual ACK và pgxpool để ghi notification theo event ID unique rồi mới ACK. Bắt đầu với prefetch nhỏ; tăng consumer/prefetch chỉ khi queue tồn mà DB còn dư khả năng ghi.
- [ ] **C04.6 — NGINX ↔ API, đo tiếp ở Phase 05.** Dùng upstream keepalive/access log của NGINX để đo request chia tới mỗi API; giữ phép thử này riêng để không nhầm hiệu quả pool với hiệu quả scale API.

## Xây và đo feature RabbitMQ

- [ ] **C04.7 — Tạo topology cho hộp thông báo.** Dùng Docker Compose chạy RabbitMQ và Go package `github.com/rabbitmq/amqp091-go`; tạo durable exchange `order-status`, durable queue `order-status-events` và binding tương ứng. Chỉ một loại notification worker consume queue này; queue không chứa yêu cầu checkout hoặc quyết định nghiệp vụ đơn hàng.
- [ ] **C04.8 — Chống gửi/ghi trùng.** Dùng event_id trong outbox, unique constraint PostgreSQL ở bảng notifications, publisher confirms và consumer ACK thủ công. Nếu publisher chết sau publish hoặc consumer chết trước ACK, event có thể lặp nhưng khách chỉ nhận một notification.
- [ ] **C04.9 — Xử lý message lỗi.** Dùng RabbitMQ retry có giới hạn và dead-letter queue `order-status-events.dlq`; đo số retry và dead-letter để message lỗi không lặp vô hạn và vẫn có chỗ cho người vận hành tra cứu.
- [ ] **C04.10 — So E0 với E1.** Dùng cùng k6 workload, CPU/RAM, dữ liệu và số API để so notification ghi trực tiếp ở E0 với RabbitMQ/outbox ở E1. Ghi p95/p99 checkout, đơn hợp lệ, số query DB, oldest outbox age, queue ready/unacked và thời gian thông báo tới khách.
- [ ] **C04.11 — Tối ưu từng pool sau khi có số liệu.** Giữ workload E0 và đổi một thông số mỗi lần. Chỉ giữ cấu hình cải thiện thời gian chờ/đơn hợp lệ mà không làm PostgreSQL, Redis, RabbitMQ hoặc payment simulator quá tải.

## Quy tắc chọn thông số

- Dùng một http.Client/Transport chung; luôn đọc hết và đóng response body để Go tái sử dụng kết nối.
- Thử riêng MaxIdleConns, MaxIdleConnsPerHost và MaxConnsPerHost; chỉ tăng giới hạn khi trace cho thấy đang chờ pool.
- Đặt timeout riêng cho dial, TLS, response header và toàn bộ HTTP request; tổng timeout không vượt ngân sách checkout.
- Với RabbitMQ, ghi consumer count, prefetch, unacked và ready messages; tăng prefetch từng bước vì quá nhiều message đang xử lý sẽ chiếm RAM và ghi dồn vào DB.
- Khi RabbitMQ tắt, checkout phải tiếp tục dựa trên PostgreSQL; outbox được gửi sau khi broker phục hồi.
- Chỉ outbox publisher kết nối RabbitMQ để publish; không tạo publisher riêng cho từng event nếu chúng dùng chung một outbox và cùng chính sách giao nhận.
- Chạy mỗi cấu hình ít nhất ba lần với cùng k6 workload; không kết luận từ một lần chạy tốt nhất.

## Tiêu chí hoàn thành

- [ ] Có ít nhất ba thử nghiệm kết nối, mỗi thử nghiệm đổi một thông số chính.
- [ ] E1 notification hoạt động qua RabbitMQ; consumer restart không mất event hoặc tạo notification trùng.
- [ ] Checkout không chờ RabbitMQ và không thay đổi trạng thái đơn chỉ vì queue lỗi.
- [ ] RabbitMQ dừng không làm mất event: event còn trạng thái chờ trong PostgreSQL và được publish lại sau khi broker phục hồi.
- [ ] Đã ghi thông số, thời gian chờ, p95/p99, lỗi, tài nguyên, outbox backlog và queue depth.
- [ ] Đã xác định 3–5 kết nối đáng tối ưu và giải thích bằng số đo.

**Đưa vào báo cáo:** So sánh notification đồng bộ/bất đồng bộ, outbox flow, ACK/confirm, pool đã thử và ảnh hưởng tới checkout/DB.
