# Phase 08 — Báo cáo, demo và bảo vệ

**Thời gian:** Tuần 12  
**Phụ thuộc:** Phase 00–07  
**Kết quả cần đạt:** Giảng viên có thể dựng lại SaleScale, theo dõi luồng nghiệp vụ và kiểm tra kết luận tối ưu từ dữ liệu gốc.

## Việc cần làm

- [ ] **C08.1 — Đối chiếu phạm vi.** Dùng PRD, function-requirement.md, user-story.md và product-spec.md để đối chiếu từng FR/US/TEST với endpoint Go/Gin, dữ liệu PostgreSQL và màn hình React/Vite đã triển khai. Bảng trạng thái giúp báo cáo cho thấy feature nào đã hoạt động, còn thiếu hay chưa được kiểm chứng.
- [ ] **C08.2 — Gom điều kiện benchmark.** Dùng k6 và hồ sơ mỗi lần chạy để ghi Git commit, phiên bản công cụ, CPU/RAM, số API, pool, Redis, RabbitMQ, consumer/prefetch, dữ liệu và số lần lặp. Người khác cần những thông tin này để chạy lại phép đo.
- [ ] **C08.3 — Trình bày kết quả nghiệp vụ.** Dùng k6, Prometheus và PostgreSQL để tách request, order mới, order confirmed, replay, hết hàng, 429 và lỗi hệ thống. Kết quả cho biết API tạo đơn hợp lệ thế nào thay vì chỉ đếm request.
- [ ] **C08.4 — So sánh các đường gọi.** Dùng pgxpool, Go httptrace, go-redis, RabbitMQ metrics và NGINX access log lập bảng 3–5 phép thử. Mỗi dòng nêu setting, feature phục vụ, thời gian chờ, lỗi, CPU/RAM và thay đổi checkout/notification.
- [ ] **C08.5 — Vẽ sơ đồ E0 tới bản cuối.** Dùng Mermaid hoặc draw.io thể hiện E0 trực tiếp PostgreSQL, E1 outbox/RabbitMQ, E2 nhiều API, E3 cache và cấu hình cuối. Tách sơ đồ triển khai AKS khỏi sơ đồ luồng RabbitMQ; đánh dấu PostgreSQL là source of truth, còn Redis/message chỉ là dữ liệu dẫn xuất hoặc trung chuyển.
- [ ] **C08.6 — Cập nhật hướng dẫn chạy.** Viết README cho Go/Gin, React/Vite, Docker Compose, PostgreSQL migration/seed, NGINX, Redis, RabbitMQ, worker và k6; bổ sung cách áp manifest AKS, cấu hình secret placeholder và chạy cloud smoke test. Hướng dẫn phải cho phép giảng viên bật/tắt broker và lặp lại benchmark local mà không cần AKS.
- [ ] **C08.7 — Chuẩn bị demo theo feature.** Dùng browser và Docker Compose trình diễn khách đặt giỏ, payment simulator trả kết quả, inbox nhận notification, broker tạm dừng nhưng checkout vẫn chạy, rồi khôi phục RabbitMQ để notification tới sau.
- [ ] **C08.8 — Chuẩn bị phần bảo vệ.** Dùng dashboard Prometheus/Grafana để giải thích vì sao RabbitMQ chỉ phù hợp với notification, cách outbox tránh mất event, cách consumer chống trùng, cách tune pool và vì sao thêm API chưa chắc tăng đơn.
- [ ] **C08.9 — Nêu giới hạn.** Dựa trên máy lab, ghi rõ NGINX proxy_cache không phải CDN toàn cầu, một RabbitMQ node và một NGINX gateway không chứng minh HA, AKS không tự tạo khả năng scale, benchmark không phải SLA và kết quả chỉ đúng với cấu hình đã đo.

## Tự kiểm trước buổi bảo vệ

- [ ] Giải thích được vai trò Go/Gin, PostgreSQL, Redis, RabbitMQ, NGINX, CDN và worker bằng feature SaleScale cụ thể.
- [ ] Có thể chỉ thời gian request ở API/DB/Redis/payment và notification qua outbox/queue/consumer.
- [ ] Có số liệu 1/2/4 API với tổng CPU/RAM không đổi, cùng bảng riêng khi tăng hạ tầng.
- [ ] Chứng minh Redis không quyết định checkout và RabbitMQ không giữ/confirm hàng.
- [ ] Có E0/E1 so sánh notification trực tiếp với notification bất đồng bộ, kèm queue delay/backlog.
- [ ] Nêu được lỗi RabbitMQ nào đã thử, event được gửi bù ra sao và cách tra DLQ.
- [ ] Giải thích được Azure Standard Load Balancer xử lý L4, NGINX Gateway Fabric xử lý L7, còn API/outbox publisher/notification worker là các workload tách biệt trong AKS.
- [ ] Kết luận dựa trên đơn hợp lệ, độ trễ, lỗi và tải; không coi 429/hết hàng là thành tích.
- [ ] Có cấu hình, dữ liệu, scripts và hướng dẫn đủ để chạy lại.

## Tiêu chí hoàn thành

- [ ] TEST-01..29 và INV-01..12 có kết quả trong các phạm vi đã thực hiện.
- [ ] Báo cáo có số liệu gốc, cấu hình, script và Git commit.
- [ ] Kết quả xấu hoặc không cải thiện vẫn được trình bày.
- [ ] Mỗi kết quả hiệu năng được kiểm tra lại tồn, trạng thái đơn và thông báo liên quan.
- [ ] Slide theo thứ tự: bài toán → feature → kiến trúc → cách đo → kết quả → giới hạn → kết luận.
- [ ] Demo chạy theo README trên môi trường sạch.
- [ ] Manifest AKS và cloud smoke test có hướng dẫn tái lập, nhưng số liệu cloud được báo riêng với baseline Docker Compose.

**Bàn giao:** Backend/giao diện demo, báo cáo, slide, hướng dẫn chạy và danh sách giới hạn.
