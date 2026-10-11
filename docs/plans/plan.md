# Lộ trình đồ án SaleScale

**Thời lượng:** 12 tuần, 2 sinh viên
**Cách dùng:** File này mô tả kiến trúc và thứ tự làm việc. Mỗi phase có một file nhiệm vụ riêng. Trong mỗi bước, công nghệ được gắn với chức năng backend mà nó phục vụ và kết quả cần đưa vào báo cáo.

- [Phase 00 — Phạm vi và môi trường đo](phases/phase-00-scope-and-lab.md)
- [Phase 01 — Luồng bán hàng lõi](phases/phase-01-core-commerce.md)
- [Phase 02 — Checkout, trạng thái đơn và hộp thông báo](phases/phase-02-checkout-correctness.md)
- [Phase 03 — Theo dõi và đo mốc ban đầu E0](phases/phase-03-observability-and-e0.md)
- [Phase 04 — RabbitMQ, các kết nối và mốc E1](phases/phase-04-couplings-and-e1.md)
- [Phase 05 — NGINX và mở rộng API E2](phases/phase-05-nginx-and-e2.md)
- [Phase 06 — Redis, CDN và cache E3](phases/phase-06-cache-and-edge-e3.md)
- [Phase 07 — Quá tải và tăng tài nguyên E4/I1](phases/phase-07-overload-and-infrastructure.md)
- [Phase 08 — Báo cáo và bảo vệ](phases/phase-08-report-and-defense.md)

## Mục tiêu

Xây backend flash sale đúng nghiệp vụ trước, rồi đo từng thay đổi để biết cách nào tăng số đơn hợp lệ hoặc giảm thời gian phản hồi. Mỗi phase kết hợp đọc tài liệu công nghệ với thực nghiệm trên SaleScale; kết quả có thể dùng để giải thích lựa chọn kiến trúc khi báo cáo.

Dùng Go/Gin cho API; React/Vite cho giao diện demo; PostgreSQL làm nguồn dữ liệu chính cho tài khoản, giá, tồn, đơn và trạng thái thanh toán; Redis làm bản sao đọc nhanh cho danh sách sản phẩm công khai; NGINX làm cổng nhận request, chia tải và giới hạn request; k6 tạo tải thử. RabbitMQ phục vụ một feature backend cụ thể: chuyển sự kiện trạng thái đơn sang worker để ghi vào hộp thông báo khách hàng mà không bắt checkout chờ việc này. Payment simulator tạo kết quả thanh toán thử nghiệm; worker Go xử lý đơn hết hạn và gửi outbox event.

RabbitMQ không quyết định đơn có được tạo, thanh toán, hủy hay hết hạn. Trạng thái đơn và sự kiện cần gửi được lưu trong PostgreSQL; nếu RabbitMQ dừng, checkout vẫn trả kết quả DB và thông báo sẽ tới sau khi worker gửi lại event.

Yêu cầu theo dõi trong tài liệu hiện tại: FR-001..049, US-01..21, TEST-01..29 và INV-01..12. Mục tiêu ở mức tải tham chiếu là checkout p95 ≤ 500 ms, p99 ≤ 1.000 ms, lỗi hệ thống < 1%, và 95% hàng giữ được trả trong 30 giây sau hạn. Chốt tải tham chiếu sau E0; không đổi mục tiêu chỉ để làm đẹp kết quả.

Repo hiện mới có chương trình Go “Hello world”; Dockerfile, Docker Compose, NGINX và frontend chưa được hoàn thiện. Vì vậy phase đầu dựng môi trường chạy lại được trước khi đo.

## Sơ đồ E0 — Một API, thông báo ghi trực tiếp vào PostgreSQL

    Trình duyệt hoặc k6
             |
           NGINX
             |
        Go/Gin API (1)
          /        \
         v          v
    PostgreSQL    Redis (chưa dùng)
       ^   |
       |   +-- notification inbox được ghi cùng thay đổi trạng thái đơn
       |
    Worker hết hạn

    Payment simulator --HTTP--> API/NGINX

E0 có hộp thông báo để so sánh cùng một feature về sau, nhưng chưa dùng RabbitMQ. API ghi thông báo vào PostgreSQL cùng giao dịch cập nhật trạng thái đơn. Đây là mốc đo trực tiếp, một API, cache Redis tắt.

## Sơ đồ mục tiêu — Mở rộng phù hợp đồ án

    Trình duyệt --> CDN (chỉ file tĩnh và GET công khai cho phép cache)
                      |
                    NGINX (chia tải, giới hạn request)
                    / | \
                   v  v  v
                API 1 API 2 ... API N
                  |    |       |
                  +----+-------+----> PostgreSQL
                  |    |       |          ^
                  +--> Redis (catalog)   |
                                          |
    Worker hết hạn + outbox publisher ----+
                  |
                  v
              RabbitMQ (order status events)
                  |
                  v
          Notification worker
                  |
                  +---------------------> PostgreSQL (hộp thông báo)

    Payment simulator --HTTP--> API/NGINX
    k6 --> tải; Prometheus --> chỉ số; Grafana --> biểu đồ

CDN chỉ phục vụ file tĩnh và request GET công khai có header cache phù hợp. Nếu lab không có CDN thật, thử NGINX proxy_cache và ghi rõ đây không mô phỏng khoảng cách mạng hay nhiều khu vực. Không cần dịch vụ trả phí. Không đưa Kubernetes, microservice, sharding, database HA hay thanh toán thật vào phạm vi.

## Công nghệ gắn với chức năng backend

| Công nghệ | Chức năng SaleScale | Vì sao cần và điều gì cần kiểm chứng |
|---|---|---|
| Go/Gin | Tài khoản, catalog, giỏ, checkout, thanh toán thử nghiệm, hộp thông báo | Giữ nghiệp vụ trong một backend; đo thời gian từng request và các lời gọi ra ngoài |
| PostgreSQL/pgxpool | Session, sản phẩm, tồn, đơn, trạng thái thanh toán, outbox và notification inbox | Bảo vệ tính đúng; đo transaction, chờ khóa, chờ pool và tổng kết nối khi tăng API |
| Redis/go-redis | Cache GET catalog công khai | Giảm truy vấn đọc DB; xác nhận checkout luôn đọc giá/tồn từ PostgreSQL |
| RabbitMQ/amqp091-go | Chuyển order status event tới notification worker | Tách ghi thông báo khỏi checkout; đo độ trễ giao thông báo, backlog, retry và phục hồi broker |
| NGINX | Reverse proxy, chia tải API và rate limit | Đo 1→N API, cách chia request, keepalive, timeout và hành vi khi một API dừng |
| CDN hoặc NGINX proxy_cache | File tĩnh và GET catalog công khai | Đo khả năng giảm lượt đọc; kiểm tra không lưu giỏ, đơn, tài khoản hay thông báo riêng |
| k6 | Tải đủ hàng, SKU nóng, cache và broker chậm | Phát cùng workload để so sánh các kiến trúc trong điều kiện cố định |
| Prometheus/Grafana | Chỉ số request, DB, Redis, RabbitMQ, worker và hạ tầng | Tìm nơi đang nghẽn trước khi tăng pool, số API hoặc CPU/RAM |

## Các đường gọi cần đo

| Đường gọi | Công nghệ/bộ kết nối | Gắn với feature và số đo cần lấy |
|---|---|---|
| Trình duyệt ↔ NGINX/API | HTTP, cookie, TLS | Đăng nhập, catalog, giỏ, đơn, hộp thông báo; thời gian phản hồi và lỗi |
| NGINX ↔ API | NGINX upstream keepalive | Chia tải checkout/catalog; số request theo API, kết nối dùng lại và timeout |
| API ↔ PostgreSQL | pgxpool | Checkout, session, đơn và notification inbox; thời gian chờ pool, query, khóa và transaction |
| API ↔ Redis | go-redis | GET catalog; thời gian Redis, hit/miss và thời gian fallback DB |
| API ↔ payment simulator | Go net/http Transport | Mở phiên và nhận kết quả thanh toán; thời gian chờ kết nối, timeout và reuse |
| Order worker ↔ PostgreSQL/RabbitMQ | pgxpool, amqp091-go | Trả hàng hết hạn và gửi outbox event; số event chờ, confirm và lỗi publish |
| RabbitMQ ↔ notification worker ↔ PostgreSQL | AMQP consumer, pgxpool | Ghi thông báo cho khách; ACK, prefetch, event trùng, backlog và độ trễ giao |

Các module trong Go gọi nhau trực tiếp trong cùng tiến trình; không tách thành dịch vụ riêng. Payment simulator là lời gọi HTTP bên ngoài. RabbitMQ chỉ được thêm cho feature hộp thông báo, không chuyển quyết định giữ hàng hoặc thanh toán vào queue.

## Cách chọn thông số kết nối

### HTTP tới payment simulator

Dùng một http.Client/http.Transport chung và httptrace để đo kết nối rảnh, bận, được dùng lại, thời gian chờ và thời gian nhận response. Thử riêng MaxIdleConns, MaxIdleConnsPerHost, MaxConnsPerHost, Dialer.Timeout, TLSHandshakeTimeout, ResponseHeaderTimeout và Client.Timeout. Chỉ tăng giới hạn khi request chờ pool và dịch vụ phía sau còn sức xử lý. Đọc hết và đóng response body để Go tái sử dụng kết nối. Dùng cùng kịch bản k6, chạy mỗi cấu hình ít nhất ba lần.

### PostgreSQL, Redis và RabbitMQ

Với pgxpool, ghi MaxConns/MinConns và thời gian chờ lấy kết nối; cộng giới hạn của mọi API, worker và công cụ vận hành trước khi so với khả năng DB. Với go-redis, dùng tài liệu đúng phiên bản đang cài để chọn PoolSize, MinIdleConns và PoolTimeout cho GET catalog. Với RabbitMQ, dùng amqp091-go; bắt đầu bằng một consumer/prefetch nhỏ cho notification worker, sau đó chỉ tăng khi queue ready tăng mà DB còn khả năng ghi. Không dùng một bộ tham số cho mọi dịch vụ.

### NGINX upstream

Dùng upstream keepalive để tái sử dụng kết nối rảnh tới API; đo access log và số request từng API. Keepalive không phải giới hạn tổng request đang xử lý. So sánh cấu hình trên cùng workload và giữ timeout phù hợp với ngân sách checkout.

## Các mốc đo

| Mốc | Thay đổi | Câu hỏi thực nghiệm |
|---|---|---|
| E0 | Một API, PostgreSQL notification inbox, Redis/RabbitMQ chưa dùng | Mốc đúng nghiệp vụ và chi phí của cách ghi notification trực tiếp |
| E1 | RabbitMQ/outbox cho notification; sau đó thử riêng các pool/query | Checkout có giảm việc đồng bộ không và thông báo đến trễ bao lâu? |
| E2 | 1/2/4 API sau NGINX | Thêm API có tăng đơn hợp lệ khi tổng CPU/RAM API giữ nguyên không? |
| E3 | Redis và thử cache tại edge | Giảm tải đọc DB bao nhiêu mà không làm sai dữ liệu checkout? |
| E4 | Rate limit, outage RabbitMQ/Redis/API và phục hồi | Hệ thống từ chối và hồi phục thế nào; đơn, outbox, notification có còn đúng không? |
| I1 | Tăng một tài nguyên hạ tầng mỗi lượt | CPU, RAM, disk hay broker/DB mới là giới hạn? |
| F | Chạy cấu hình cuối đã chọn | Có thể tái lập kết quả và giải thích trade-off không? |

**Thêm API, giữ nguyên hạ tầng:** giữ tổng CPU/RAM dành cho API như E1 rồi chạy 1/2/4 tiến trình. **Tăng hạ tầng:** chỉ tăng một tài nguyên sau khi số liệu chỉ ra nút thắt; báo riêng để phân biệt lợi ích từ cách chia tải với lợi ích từ máy mạnh hơn.

## Lịch 12 tuần

| Tuần | Phase | Bàn giao |
|---|---|---|
| 1 | 00 | Sơ đồ, câu hỏi nghiên cứu và lab Docker Compose |
| 1–4 | 01 | Tài khoản, hàng hóa, sale, giỏ và giao diện demo |
| 4–6 | 02 | Checkout đúng, thanh toán thử và hộp thông báo E0 |
| 6–7 | 03 | Prometheus/Grafana, workload và baseline E0 |
| 7–8 | 04 | RabbitMQ/outbox, các pool và kết quả E1 |
| 8–9 | 05 | NGINX, nhiều API và kết quả E2 |
| 9–10 | 06 | Redis/CDN và kết quả E3 |
| 10–11 | 07 | Quá tải, lỗi dịch vụ, queue recovery và tăng tài nguyên |
| 12 | 08 | Báo cáo, demo, hướng dẫn tái lập và bảo vệ |

Đồ án hoàn thành khi P0 chạy đúng; TEST-01..29 có bằng chứng; INV-01..12 không vi phạm trong các kịch bản đã chạy; có E0/E1/E2 và kết quả E3/E4/I1 hoặc ghi rõ giới hạn. Báo cáo phải tách request, đơn hợp lệ, lỗi nghiệp vụ, lỗi hệ thống, 429, backlog RabbitMQ và độ trễ thông báo. Không kết luận hệ thống mở rộng tuyến tính, sẵn sàng production hay đạt SLA nếu chưa kiểm chứng ngoài cấu hình lab.
