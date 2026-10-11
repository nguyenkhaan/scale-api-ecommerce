# ADR-002: Dùng Azure Kubernetes Service cho môi trường cloud mục tiêu

## Trạng thái

Accepted

## Ngày

2026-10-11

## Thay thế

ADR này thay thế [ADR-001](001-use-eks-for-cloud-target.md).

## Bối cảnh

SaleScale cần chạy hai hoặc nhiều Go/Gin backend instance dùng chung PostgreSQL trong một Kubernetes được quản lý. Chủ dự án đã thay quyết định cloud target từ Amazon EKS sang Azure Kubernetes Service (AKS). Docker Compose vẫn cần được giữ làm môi trường local và baseline benchmark cho nhóm hai người.

Sơ đồ tổng quan chỉ cần thể hiện các cụm hạ tầng chính. RabbitMQ, workers, order events, payment simulator, observability và k6 sẽ được mô tả trong các sơ đồ luồng chi tiết riêng.

## Quyết định

- Dùng Azure Kubernetes Service (AKS) cho môi trường cloud mục tiêu.
- Dùng Azure Standard Load Balancer làm cổng L4 vào cluster.
- Dùng NGINX Gateway Fabric trong AKS để reverse proxy, rate limit và chia request L7 cho hai backend instance Go/Gin.
- Dùng Azure Database for PostgreSQL Flexible Server làm primary database.
- Giữ Upstash Redis cho cache catalog và Cloudflare R2 cho object storage.
- Chạy expiry worker, outbox publisher và notification worker bằng các Deployment riêng với backend. Cloud demo chạy RabbitMQ một node bằng StatefulSet, PVC và ClusterIP Service trong AKS; broker không nằm trong pod backend và không được mở public.
- Giữ Docker Compose làm môi trường local và baseline; chưa đưa HPA, multi-region hoặc database HA vào MVP.

## Phương án đã cân nhắc

### Amazon EKS

- Ưu: Kubernetes được quản lý, phù hợp với kiến trúc đã phác thảo trước đó.
- Nhược: không còn phù hợp với lựa chọn cloud mới của chủ dự án.
- Kết luận: bị thay thế bởi AKS.

### Azure Application Gateway đi thẳng tới backend

- Ưu: tích hợp sâu với Azure và xử lý L7.
- Nhược: trùng vai trò với NGINX, trong khi NGINX load balancing và rate limiting là phần cần nghiên cứu của đồ án.
- Kết luận: chưa dùng trong MVP.

## Hệ quả

- Manifest/IaC cloud phải dùng tài nguyên AKS, Azure Load Balancer và Azure Database for PostgreSQL thay cho AWS.
- Azure Load Balancer chỉ đưa traffic L4 vào cluster; NGINX vẫn là bộ cân bằng tải L7 của ứng dụng.
- Hai backend instance phải stateless và cùng dùng PostgreSQL/Redis; không giữ session hay idempotency trong memory của một pod.
- RabbitMQ một node chỉ phục vụ demo và kiểm chứng outage/recovery, không chứng minh broker HA. Outbox trong PostgreSQL giữ event khi broker dừng.
- Kết quả benchmark cloud phải báo riêng với baseline Docker Compose.
- Việc dùng AKS không tự động đồng nghĩa với autoscaling, HA production hoặc cải thiện throughput.

## Nguồn

- [Azure Kubernetes Service](https://learn.microsoft.com/en-us/azure/aks/what-is-aks)
- [Azure Standard Load Balancer trong AKS](https://learn.microsoft.com/en-us/azure/aks/load-balancer-standard)
- [Azure Database for PostgreSQL Flexible Server](https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/overview)
