# Spec: SaleScale

**Trạng thái:** Draft để nhóm rà soát
**Giai đoạn SDD:** SPECIFY
**Cập nhật:** 18/10/2026

## Giả định đang dùng

1. PRD là nguồn sự thật cho nghiệp vụ; tài liệu này chỉ quy định cách triển khai và liên kết tới PRD.
2. Sản phẩm là web application cho một cửa hàng trên mỗi bản triển khai; frontend là React + Vite, backend là Go + Gin.
3. API là HTTP JSON theo hợp đồng OpenAPI; xác thực bằng opaque session cookie, session lưu trong PostgreSQL.
4. PostgreSQL là nguồn có thẩm quyền cho session, giá, tồn kho, quota, đơn, idempotency và thanh toán. Redis chỉ tăng tốc đọc công khai.
5. P0 được đặc tả đủ để tạo migration, handler và test. P1 cùng hạ tầng sau P0 chỉ chốt boundary và contract ở giai đoạn này.
6. Repository mới có chương trình Go tối thiểu; các lệnh dành cho frontend, migration, lint, integration test và benchmark phải được bổ sung khi công cụ tương ứng được chọn.

Mọi giả định làm thay đổi `BR`, `FR`, `INV` hoặc tiêu chí nghiệm thu phải được đưa vào PRD trước khi triển khai.

## Mục tiêu

Xây dựng backend flash sale cho một doanh nghiệp, cho phép khách mua nhiều SKU trong một đơn và giữ toàn bộ hàng một cách nguyên tử dưới tải đồng thời. Hệ thống phải xử lý retry, callback thanh toán lặp/trễ và worker hết hạn mà không bán vượt tồn, tạo đơn trùng hoặc kết thúc một reservation nhiều lần.

Người dùng trực tiếp của đặc tả là hai thành viên dự án và coding agent. Thành công khi họ có thể chuyển từ mã yêu cầu sang schema, endpoint và test tương ứng mà không phải tự diễn giải lại nghiệp vụ.

## Phạm vi đặc tả hiện tại

### Chi tiết đến mức triển khai

- Đăng ký, đăng nhập, đăng xuất, session và phân quyền khách/quản trị.
- Catalog, product, SKU, tồn kho và lịch sử điều chỉnh.
- Flash sale, quota và giới hạn mua.
- Giỏ hàng, preview, xác nhận giá và người nhận.
- Checkout nhiều SKU nguyên tử, idempotency và tra cứu kết quả chưa rõ.
- Order, reservation, payment simulator, callback, hủy và expiry worker.
- Logging/metrics tối thiểu để truy vết request, order và payment event.

### Chỉ chốt boundary và contract

- Transactional outbox, RabbitMQ, notification inbox và dead-letter handling.
- Redis catalog cache và bảo vệ cache stampede.
- Rate limiting/overload protection.
- Benchmark k6 và so sánh E0–E4.
- Docker Compose hoàn chỉnh và triển khai AKS.
- Frontend demo React + Vite ngoài các payload cần cho API contract.

## Tech stack

| Lớp | Quyết định | Trạng thái |
|---|---|---|
| Backend | Go `1.27.1`, Gin | Go đã khai báo; Gin chưa được thêm dependency |
| Frontend | React + Vite | Đã chốt sản phẩm; chưa bootstrap |
| Primary data | PostgreSQL | Đã chốt; schema P0 ở `technical/data-model.md` |
| Cache | Redis cache-aside cho catalog công khai | Boundary P1 |
| Message broker | RabbitMQ cho notification qua outbox | Boundary P1 |
| Edge/gateway | Cloudflare; Azure Load Balancer; NGINX Gateway Fabric | Cloud target |
| Local lab | Docker Compose | File hiện là placeholder |
| Observability | Prometheus, Grafana, structured logs | Contract trước, cấu hình sau |
| Load test | k6 | Kịch bản trong PRD; script sau baseline |

Không khóa phiên bản dependency chưa xuất hiện trong `go.mod` hoặc manifest frontend. Khi thêm dependency, ghi phiên bản thực tế vào manifest và cập nhật tài liệu nếu nó thay đổi kiến trúc.

## Lệnh

### Lệnh chạy được ở trạng thái repository hiện tại

```bash
cd app/backend
go run .
go build ./...
go test ./...
go vet ./...
gofmt -w .
```

### Hợp đồng lệnh phải có sau bootstrap

```bash
docker compose up --build
docker compose down
cd app/backend && go test ./... -race -cover
cd app/backend && golangci-lint run
cd app/frontend && npm run dev
cd app/frontend && npm run build
cd app/frontend && npm test
k6 run tests/load/<scenario>.js
```

Các lệnh ở nhóm thứ hai là interface vận hành cần đạt, chưa phải bằng chứng rằng cấu hình hiện tại đã hỗ trợ.

## Cấu trúc dự án mục tiêu

```text
app/
  backend/
    cmd/api/              API process
    cmd/expiry-worker/    worker giải phóng đơn hết hạn
    internal/
      auth/               controller, service và DTO xác thực
      catalog/            controller, service và DTO catalog
      inventory/          controller, service và DTO tồn kho
      flashsale/          controller, service và DTO flash sale
      cart/               controller, service và DTO giỏ hàng
      order/              controller, service và DTO đơn/checkout
      payment/            controller, service và DTO thanh toán
      model/              model ORM dùng chung
      httpserver/         router, middleware và HTTP response chung
      bootstrap/          khởi tạo và gắn dependency thủ công
      config/             cấu hình ứng dụng
    migrations/           PostgreSQL migrations
  frontend/               React + Vite demo
docs/                     PRD, spec, architecture, contracts và ADR
deploy/
  compose/                local/baseline
  aks/                    cloud target
tests/
  integration/            test PostgreSQL/Redis/RabbitMQ theo boundary
  contract/               kiểm tra OpenAPI và event contract
  load/                   k6 workloads và seed/reset
```

Đây là layout mục tiêu, không phải kế hoạch triển khai. Mỗi thay đổi cấu trúc lớn cần ADR hoặc cập nhật spec trước.

## Code style

- Package và biến dùng tên ngắn nhưng thể hiện nghiệp vụ; tránh tên chung như `utils`, `manager`, `data`.
- Mỗi module API có `controller.go`, `service.go` và `dto.go`; model ORM nằm trong `internal/model/`.
- Handler chỉ chuyển đổi HTTP; transaction và quy tắc nghiệp vụ nằm ở service.
- Service dùng ORM đã chọn trực tiếp; chỉ tạo abstraction truy cập dữ liệu khi một use case thực sự cần nó.
- Truyền `context.Context` làm tham số đầu tiên cho I/O; không cất context trong struct.
- Tiền dùng số nguyên VND (`int64` trong Go, `BIGINT` trong PostgreSQL).
- Thời gian lưu UTC dưới dạng `timestamptz`; quyết định hết hạn dùng giờ database trong transaction.
- Lỗi nghiệp vụ có mã ổn định trong OpenAPI; không buộc client phân tích câu chữ.
- Không dùng mutex cục bộ để bảo vệ tồn, quota hoặc idempotency giữa nhiều instance.

```go
type PlaceOrderCommand struct {
	UserID         uuid.UUID
	IdempotencyKey string
	CartItemIDs    []uuid.UUID
	ExpectedTotal  int64
	Recipient      Recipient
}

func (service *OrderService) PlaceOrder(
	ctx context.Context,
	command PlaceOrderCommand,
) (OrderResult, error) {
	return service.orders.PlaceAtomically(ctx, command)
}
```

## Chiến lược kiểm thử

[testing.md](technical/testing.md) là nguồn chi tiết. Tối thiểu:

- Unit test cho chuẩn hóa request, tính giá/tổng, state transition và mã lỗi.
- Integration test PostgreSQL cho transaction nhiều SKU, constraint, idempotency và worker cạnh tranh.
- HTTP contract test cho OpenAPI, auth, quyền sở hữu và error envelope.
- Fault/concurrency test cho callback lặp/trễ, hủy/hết hạn cạnh tranh và restart worker.
- Load test chỉ chạy trên phiên bản đã qua bất biến P0; số đo tốc độ không thay thế kiểm chứng dữ liệu.

Không đặt coverage phần trăm làm tiêu chí duy nhất. Mọi nhánh thay đổi tiền, tồn, quota, idempotency hoặc trạng thái kết thúc phải có test kiểm tra trực tiếp dữ liệu sau thực thi.

## Boundaries

### Luôn làm

- Liên kết thay đổi với ít nhất một mã `FR` hoặc `NFR` và test tương ứng.
- Cập nhật PRD trước nếu thay đổi nghiệp vụ; cập nhật traceability, contract và test ngay sau đó.
- Dùng transaction PostgreSQL cho thay đổi nhiều bảng liên quan đến tiền/tồn.
- Khóa/tác động nhiều SKU theo thứ tự ổn định.
- Xác thực input, quyền sở hữu và role ở server.
- Chạy formatter, static checks và test liên quan trước khi hoàn thành thay đổi.

### Phải xin ý kiến trước

- Đổi `BR`, `INV`, trạng thái order/reservation hoặc thời hạn giữ hàng.
- Thay đổi public API không tương thích hoặc error code đã công bố.
- Thay đổi schema làm mất dữ liệu, backfill dữ liệu lớn hoặc bỏ constraint bảo vệ bất biến.
- Thêm dependency hạ tầng, đổi PostgreSQL làm nguồn có thẩm quyền hoặc đưa nghiệp vụ checkout vào queue.
- Đổi mục tiêu benchmark, tài nguyên tham chiếu hoặc điều kiện so sánh E0–E4.

### Không được làm

- Commit secret, token, mật khẩu hoặc dữ liệu người dùng thật.
- Tin giá, tổng tiền, trạng thái thanh toán hoặc role do client tự gửi.
- Dùng Redis/cache làm nguồn duy nhất để quyết định checkout.
- Tạo partial order khi một dòng không hợp lệ.
- Retry ý định checkout bằng idempotency key mới khi kết quả cũ chưa rõ.
- Xóa test lỗi hoặc nới bất biến để làm benchmark đẹp hơn.
- Tự hồi sinh đơn đã hết hạn/hủy/thất bại khi callback thành công đến muộn.

## Tiêu chí thành công của phase SPECIFY

- PRD phân định rõ nội dung research nào được chấp nhận, thay thế hoặc giữ làm giả thuyết.
- Domain model mô tả đầy đủ trạng thái order, reservation, payment và flash sale.
- Schema P0 chỉ ra constraint/index và transaction boundary bảo vệ `INV-01` đến `INV-12`.
- OpenAPI mô tả endpoint P0, auth, idempotency, error envelope và boundary P1.
- Test spec nối được `TEST-01` đến `TEST-29` với oracle và cấp kiểm thử.
- Traceability cho phép đi từ `FR` tới tài liệu kỹ thuật và test mà không lặp lại nội dung yêu cầu.
- Hai thành viên dự án duyệt spec trước khi chuyển sang PLAN.

## Câu hỏi còn mở

| Mã | Nội dung | Cổng quyết định |
|---|---|---|
| OPEN-01 | Tải và cấu hình tham chiếu | Sau baseline E0, trước benchmark cuối |
| OPEN-02 | Quy trình thao tác sau khi đối soát thanh toán muộn | Trước nghiệm thu payment P0 |
| OPEN-03 | Mục tiêu thời gian giao notification | Sau E0/E1, trước nghiệm thu P1 |
| TECH-01 | Thư viện migration, PostgreSQL driver/query layer và router middleware cụ thể | Trước PLAN cho bootstrap backend |
| TECH-02 | Cơ chế ký/xác thực callback của payment simulator | Trước triển khai payment simulator |

