# ADR-006: Dùng AKS làm cloud target và Docker Compose làm lab

## Trạng thái

Accepted

## Ngày

2026-10-18

## Bối cảnh

Dự án cần môi trường tái lập cho benchmark và một cloud target để demo nhiều API instance. Research cũ từng ghi không dùng Kubernetes và có bản quyết định EKS; quyết định sản phẩm hiện tại chọn Azure Kubernetes Service. Hai mục tiêu lab và cloud không nên bị trộn khi báo kết quả.

## Quyết định

- Docker Compose là môi trường local/integration/baseline E0–E4.
- AKS là cloud target; Azure Standard Load Balancer đưa traffic L4 tới NGINX Gateway Fabric, rồi NGINX route/load-balance L7 tới Go/Gin API pods.
- Azure Database for PostgreSQL Flexible Server là primary database của cloud target.
- API, expiry worker, outbox publisher và notification worker chạy thành workload riêng.
- RabbitMQ một node chỉ phục vụ demo/lab; chưa tuyên bố broker HA.
- Không đưa HPA, multi-region, database sharding/HA hoặc service mesh vào MVP nếu chưa có thay đổi phạm vi.

## Phương án đã cân nhắc

### Chỉ Docker Compose

Tái lập đơn giản nhưng không đáp ứng cloud target đã chốt. Vẫn giữ làm lab.

### Amazon EKS

Đã từng được nghiên cứu nhưng không còn phù hợp lựa chọn cloud của dự án. Không dùng cho cấu hình mục tiêu.

### Microservices trên Kubernetes

Tăng chi phí giao tiếp/vận hành mà chưa có bằng chứng modular monolith là bottleneck. Không chọn.

## Hệ quả

- Cùng container/application contract phải chạy được ở Compose và AKS.
- Benchmark cloud báo riêng với baseline; AKS không tự chứng minh scaling hiệu quả.
- Ngân sách PostgreSQL connection phải tính trên tổng pod/worker.
- NGINX rate limit cục bộ phải công bố phạm vi khi gateway được nhân bản.
- Manifest/IaC chi tiết được viết trước phase hạ tầng, không thuộc đặc tả P0 hiện tại.

