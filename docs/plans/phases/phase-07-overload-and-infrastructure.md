# Phase 07 — Quá tải, phục hồi và tăng tài nguyên: E4/I1

**Thời gian:** Tuần 10–11  
**Phụ thuộc:** Phase 05; dùng Phase 06 nếu cache đã hoàn thành  
**Kết quả cần đạt:** Khi tải cao hoặc một dịch vụ lỗi, checkout vẫn giữ đúng dữ liệu, notification được gửi bù và chỉ tăng tài nguyên đang thật sự giới hạn hệ thống.

## Nghiên cứu cần làm

- [ ] **R07.1 — Chọn điểm rate limit.** Đọc NGINX limit_req và Go context để chọn giới hạn cho login, catalog, checkout và callback. Rate limit ngăn request quá mức chiếm DB/API nhưng phải từ chối trước khi giữ hàng.
- [ ] **R07.2 — Thiết kế retry an toàn.** Đọc NGINX proxy retry, Go net/http và RabbitMQ ACK/retry. Dùng idempotency key cho checkout; với notification, cho phép message giao lại nhưng consumer không tạo notification trùng.
- [ ] **R07.3 — Tìm bottleneck bằng Prometheus/Grafana.** Theo dõi CPU/RAM, DB pool/lock, Redis, NGINX, RabbitMQ ready/unacked, outbox age, notification worker và máy k6. Các số đo này cho biết nên tăng API, DB, broker hay consumer.
- [ ] **R07.4 — Lập phương án tăng hạ tầng.** Dùng Docker Compose limits hoặc cấu hình máy lab lập các mức CPU/RAM/disk có thể thử riêng cho API, PostgreSQL, Redis, RabbitMQ và k6. Kế hoạch tách nâng máy khỏi thêm tiến trình API.
- [ ] **R07.5 — Chốt tài nguyên cloud mục tiêu.** Đọc tài liệu [Azure Standard Load Balancer trong AKS](https://learn.microsoft.com/en-us/azure/aks/load-balancer-standard), [NGINX Gateway Fabric](https://docs.nginx.com/nginx-gateway-fabric), [Kubernetes StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) và [RabbitMQ queue durability](https://www.rabbitmq.com/docs/queues). Chốt manifest tối thiểu, version, request/limit, volume và secret trước khi triển khai; cloud target là môi trường demo riêng, không thay baseline Docker Compose.

## Việc triển khai và đo

- [ ] **C07.1 — Giới hạn request ở NGINX.** Dùng limit_req cho login, catalog, checkout và callback; trả HTTP 429 khi vượt ngưỡng. Dùng k6 kiểm tra request bị từ chối không tạo order/reservation.
- [ ] **C07.2 — Đặt timeout theo feature.** Dùng Go context cho PostgreSQL/HTTP/outbox publisher và NGINX timeout cho request; tổng timeout checkout giữ trong ngân sách. Timeout khi publish RabbitMQ chỉ giữ event ở trạng thái chờ, không được đổi trạng thái đơn thành thất bại.
- [ ] **C07.3 — Chạy tải quá mức và RabbitMQ chậm.** Dùng k6 LOAD-09 và LOAD-10 vượt rồi hạ tải; đo order thành công, hết hàng, 429, timeout, DB wait, outbox age, queue depth và thời gian hệ thống hồi phục.
- [ ] **C07.4 — Thử RabbitMQ outage.** Dùng Docker Compose dừng RabbitMQ sau khi PostgreSQL đã lưu order/outbox. Checkout phải trả trạng thái đơn chính xác; khi broker chạy lại, publisher gửi bù và notification xuất hiện.
- [ ] **C07.5 — Thử consumer restart/redelivery.** Dừng notification worker sau khi RabbitMQ giao message và chạy lại; dùng ACK thủ công, event ID unique PostgreSQL và dead-letter queue để xác nhận notification không mất hoặc nhân đôi.
- [ ] **C07.6 — Thử Redis/API/expiry worker lỗi.** Dừng từng service bằng Docker Compose; kiểm tra catalog fallback có giới hạn, NGINX bỏ API chết, và worker hết hạn tiếp tục trả hàng sau restart.
- [ ] **C07.7 — Tăng một tài nguyên mỗi lần.** Dựa trên Prometheus metrics từ E0–E3, tăng riêng CPU/RAM/disk của thành phần đang nghẽn. Giữ code, dữ liệu và k6 không đổi để chứng minh phần cứng đó cải thiện đúng feature.
- [ ] **C07.8 — So sánh API scaling với hạ tầng.** Đặt kết quả 1/2/4 API trong cùng CPU/RAM cạnh kết quả tăng CPU/RAM hoặc capacity PostgreSQL/RabbitMQ. Báo riêng để phân biệt cách chia tải với thêm tài nguyên.
- [ ] **C07.9 — Chốt kiến trúc cuối.** Dùng kết quả để chọn số API, pool, NGINX policy, Redis, RabbitMQ consumer/prefetch và hạ tầng. Chỉ giữ thành phần giúp tăng đơn hợp lệ, giảm độ trễ hoặc bảo đảm notification; không thêm queue cho checkout hay microservice nếu số liệu không yêu cầu.
- [ ] **C07.10 — Triển khai workload ứng dụng lên AKS.** Dùng cùng image đã đo ở local để tạo Backend Deployment hai replica cùng ClusterIP Service; expiry worker, outbox publisher và notification worker dùng Deployment riêng. Khai báo readiness/liveness probe, request/limit và tham chiếu ConfigMap/Secret; không đặt worker hoặc RabbitMQ chung pod backend.
- [ ] **C07.11 — Dựng đường vào AKS.** Cài NGINX Gateway Fabric, tạo Gateway/HTTPRoute và expose data plane bằng Service `LoadBalancer` để AKS cấp Azure Standard Load Balancer L4. Smoke test phải chứng minh request L7 đi qua NGINX tới cả hai backend replica và việc dừng một pod không làm mất toàn bộ API.
- [ ] **C07.12 — Dựng RabbitMQ cho cloud demo.** Chạy RabbitMQ một node bằng StatefulSet, PVC và ClusterIP Service trong AKS; credential lấy từ Secret và không mở AMQP/management UI ra Internet. Đây là cấu hình demo có lưu dữ liệu khi pod restart, không phải broker HA; outbox publisher và notification worker kết nối qua service nội bộ.
- [ ] **C07.13 — Kết nối dịch vụ dữ liệu và smoke test phục hồi.** Kết nối Azure Database for PostgreSQL, Upstash Redis và Cloudflare R2 bằng secret theo môi trường. Chạy riêng các ca tạo đơn, nhận notification, dừng RabbitMQ rồi khôi phục để xác nhận outbox gửi bù; không trộn kết quả này với benchmark Docker Compose E0–E4.

## Tiêu chí hoàn thành

- [ ] TEST-22/23 và TEST-26..29 kiểm tra Redis, rate limit, RabbitMQ outage, redelivery và DLQ.
- [ ] Có biểu đồ quá tải/hồi phục, tách lỗi nghiệp vụ, 429, lỗi hệ thống và notification backlog.
- [ ] Mỗi phép tăng tài nguyên chỉ đổi một yếu tố; có số liệu trước/sau.
- [ ] Chỉ ra bottleneck trước/sau và giới hạn của máy lab.
- [ ] Cloud target trên AKS chạy được luồng demo lõi qua Azure Standard Load Balancer và NGINX Gateway Fabric; hai backend replica cùng các worker là workload tách biệt.
- [ ] RabbitMQ cloud chỉ dùng endpoint nội bộ; broker restart không làm checkout thất bại và event chờ được gửi bù.
- [ ] Manifest, version, request/limit, secret placeholder, dung lượng PVC và chi phí thử nghiệm được ghi lại; không cam kết HA từ cấu hình một node.

**Đưa vào báo cáo:** Cấu hình rate limit/timeout, các sự cố đã thử, E4/I1 và lý do chọn hạ tầng cuối.
