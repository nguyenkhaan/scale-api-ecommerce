# Phase 00 — Chốt phạm vi và dựng lab đo

**Thời gian:** Tuần 1
**Kết quả cần đạt:** Nhóm hiểu SaleScale sẽ chạy qua những thành phần nào, có môi trường để kiểm chứng checkout và feature hộp thông báo.

## Nghiên cứu và quyết định

- [ ] **R00.1 — Vẽ hai mức sơ đồ kiến trúc.** Sơ đồ tổng quan cloud chỉ thể hiện Cloudflare, frontend, Azure Standard Load Balancer, NGINX Gateway Fabric, hai backend instance trong AKS, PostgreSQL, Redis và R2. Các sơ đồ chi tiết riêng mới thể hiện RabbitMQ, workers, order events, payment simulator, observability và k6; PostgreSQL vẫn là nguồn quyết định đơn/tồn.
- [ ] **R00.2 — Tách hai cách scale.** Đọc giới hạn CPU/RAM của Docker Compose rồi lập hai giả thuyết: chạy nhiều API trong cùng ngân sách tài nguyên có tác dụng gì, và cấp thêm CPU/RAM cho thành phần đang nghẽn có tác dụng gì. Tách hai phép đo để báo cáo không nhầm scale phần mềm với nâng máy.
- [ ] **R00.3 — Lập bản đồ kết nối.** Dựa vào Go net/http, pgxpool, go-redis, RabbitMQ AMQP và NGINX upstream để ghi đường gọi giữa API, DB, cache, simulator, publisher và consumer. Với từng đường gọi, ghi feature sử dụng nó cùng thời gian chờ, giới hạn và lỗi cần đo.
- [ ] **R00.4 — Đọc tài liệu gốc.** Nghiên cứu [NGINX chia tải](https://nginx.org/en/docs/http/load_balancing.html), [NGINX upstream](https://nginx.org/en/docs/http/ngx_http_upstream_module.html), [Go HTTP Transport](https://pkg.go.dev/net/http#Transport), [PostgreSQL connections](https://www.postgresql.org/docs/current/runtime-config-connection.html), [HTTP Cache RFC 9111](https://www.rfc-editor.org/rfc/rfc9111), [Redis cache-aside](https://redis.io/docs/latest/develop/use-cases/cache-aside/) và [RabbitMQ Work Queues cho Go](https://www.rabbitmq.com/tutorials/tutorial-two-go). Ghi ý nào áp dụng cho catalog, checkout hoặc hộp thông báo để phần research trả lời đúng bài toán SaleScale.
- [ ] **R00.5 — Chọn workload nghiên cứu.** Dùng k6 để xác định LOAD-01..10: giỏ nhiều SKU, một SKU nóng, thiếu hàng, đọc catalog cùng checkout, retry, callback trễ, Redis lỗi, RabbitMQ/consumer chậm và tải vượt mức. Các kịch bản này dùng chung trong từng lần so sánh để kết luận có căn cứ.
- [ ] **R00.6 — Chốt ranh giới triển khai AKS.** Dựa vào [ADR-002](../../decisions/002-use-aks-for-cloud-target.md), lập bảng ánh xạ: Azure Standard Load Balancer L4 → NGINX Gateway Fabric L7 → Backend Deployment hai replica; expiry worker, outbox publisher và notification worker là các Deployment riêng; RabbitMQ là broker riêng, không chạy trong container hoặc pod backend; Azure Database for PostgreSQL, Upstash Redis và Cloudflare R2 nằm ngoài cluster. Bảng này là nguồn chuẩn cho manifest cloud, độc lập với sơ đồ giải thích component.

## Việc dựng lab

- [ ] **C00.1 — Ghi nhận điểm xuất phát.** Đọc README, Dockerfile, Docker Compose và mã Go hiện tại; ghi phần nào chạy được để nhóm biết cần dựng tiếp API, DB, gateway hay giao diện.
- [ ] **C00.2 — Dựng dịch vụ bằng Docker Compose.** Chạy PostgreSQL, Redis, RabbitMQ 4.x, NGINX và một Go/Gin API; RabbitMQ dùng cho notification event, Redis để sau mới cache catalog. Thêm volume và healthcheck để dữ liệu thử nghiệm không mất khi restart và để các service chờ đúng thứ tự khởi động.
- [ ] **C00.3 — Thêm health endpoint bằng Go/Gin.** Tạo /health/live để xác nhận tiến trình API đang chạy và /health/ready để xác nhận PostgreSQL sẵn sàng. Không phụ thuộc RabbitMQ trong ready check vì checkout vẫn được phép chạy khi broker ngừng; outbox sẽ giữ thông báo để gửi sau.
- [ ] **C00.4 — Cho NGINX chuyển tiếp API.** Cấu hình NGINX làm reverse proxy tới một API và chuyển request ID tới Go/Gin. Đây là đường request sẽ dùng để đo login, catalog và checkout ở các phase sau.
- [ ] **C00.5 — Tạo mẫu ghi kết quả.** Dùng Git, k6 và Docker Compose để ghi commit, phiên bản công cụ, CPU/RAM, dữ liệu seed, cấu hình broker/queue và lệnh chạy. Nhờ vậy nhóm có thể tái lập cả mốc E0 không queue và E1 có RabbitMQ.

## Tiêu chí hoàn thành

- [ ] Có sơ đồ cơ bản/nâng cao, trong đó thể hiện rõ công nghệ nào phục vụ feature nào.
- [ ] Dựng được PostgreSQL, Redis, RabbitMQ, NGINX và API từ repo theo hướng dẫn.
- [ ] Gọi được health endpoint qua NGINX; RabbitMQ có thể tắt mà health/checkout API vẫn sẵn sàng nếu PostgreSQL hoạt động.
- [ ] Có LOAD-01..10, mẫu lưu cấu hình và câu hỏi nghiên cứu ban đầu.
- [ ] Có bảng ánh xạ Docker Compose → AKS và chỉ rõ API, từng worker, RabbitMQ cùng các dịch vụ dữ liệu chạy ở đâu.

**Đưa vào báo cáo:** Bài toán flash sale, lý do chọn từng công nghệ, vai trò RabbitMQ trong hộp thông báo và cách phân biệt scale API với tăng phần cứng.
