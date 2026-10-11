# ADR-001: Dùng Amazon EKS cho môi trường cloud mục tiêu

## Trạng thái

Superseded by [ADR-002](002-use-aks-for-cloud-target.md)

## Ngày

2026-10-11

## Bối cảnh

SaleScale cần chứng minh nhiều Go/Gin API instance có thể dùng chung trạng thái PostgreSQL trong bài toán flash sale. Dự án vẫn cần một lab Docker Compose nhỏ, tái lập được cho nhóm hai người trong 12 tuần, nhưng chủ dự án đã chọn Amazon EKS cho bản triển khai cloud mục tiêu.

Community `ingress-nginx` đã ngừng bảo trì từ tháng 3/2026. Kiến trúc vẫn cần NGINX cho reverse proxy, cân bằng tải và rate limit mà không dựa vào controller đã nghỉ hưu.

## Quyết định

- Dùng Amazon EKS với managed node group cho môi trường cloud mục tiêu.
- Dùng AWS Load Balancer Controller để tạo Network Load Balancer trước cluster.
- Dùng NGINX Gateway Fabric và Kubernetes Gateway API thay cho community `ingress-nginx`.
- Chạy Go/Gin API, order worker, notification worker, Prometheus và Grafana trong EKS.
- Dùng RDS PostgreSQL làm primary, Upstash Redis làm cache catalog, RabbitMQ làm broker notification và Cloudflare R2 làm object storage.
- Giữ Docker Compose làm môi trường local và baseline benchmark; không đưa HPA, multi-region, database HA hoặc broker HA vào MVP.

## Phương án đã cân nhắc

### Chỉ dùng Docker Compose

- Ưu: ít chi phí và dễ tái lập nhất.
- Nhược: không đáp ứng quyết định dùng EKS cho môi trường cloud.
- Kết luận: giữ cho lab, không dùng làm cloud target.

### Community ingress-nginx

- Ưu: quen thuộc và có nhiều ví dụ.
- Nhược: đã ngừng release, bugfix và security update từ tháng 3/2026.
- Kết luận: không dùng cho triển khai mới.

### AWS ALB đi thẳng tới API

- Ưu: ít thành phần hơn.
- Nhược: bỏ mất NGINX và phép thử NGINX rate limit/load-balancing đã nằm trong phạm vi đồ án.
- Kết luận: không chọn cho kiến trúc này.

## Hệ quả

- Nhóm phải duy trì manifest/IaC EKS ngoài cấu hình Docker Compose.
- NLB và EKS thêm chi phí, độ phức tạp mạng và biến nhiễu benchmark; số liệu local và cloud phải báo riêng.
- API và worker phải stateless ở cấp pod; session, idempotency và công việc nền không được giữ chỉ trong memory.
- Một gateway replica được dùng trong phép đo MVP để giữ ngưỡng NGINX rate limit xác định. Scale gateway toàn hệ thống cần thiết kế limiter dùng trạng thái chung.
- Việc dùng EKS không tự động đồng nghĩa với autoscaling, HA production hoặc cải thiện throughput.

## Nguồn

- [AWS EKS load balancing](https://docs.aws.amazon.com/eks/latest/best-practices/load-balancing.html)
- [NGINX Gateway Fabric](https://docs.nginx.com/nginx-gateway-fabric)
- [Kubernetes: ingress-nginx retirement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/)
