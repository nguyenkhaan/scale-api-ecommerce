# SaleScale

SaleScale là backend flash sale cho một cửa hàng doanh nghiệp, tập trung vào checkout nhiều SKU đúng dưới tải đồng thời. Dự án nghiên cứu transaction, idempotency, reservation, payment event, expiry worker và khả năng mở rộng nhiều API instance mà vẫn bảo toàn tồn kho.

Repository có đặc tả draft và kế hoạch triển khai đề xuất theo SDD. Mã nguồn mới là chương trình Go tối thiểu; Docker Compose, frontend và các service chưa được bootstrap.

## Đọc trước khi phát triển

- [PRD](docs/product/prd.md) — nguồn sự thật nghiệp vụ.
- [Mục lục SDD](docs/CONTENT.md) — thứ tự đọc và cây tài liệu.
- [Kế hoạch](docs/plans/plan.md) — các phase FE/BE và checkpoint.
- [OpenAPI](docs/technical/api/openapi.yaml) — hợp đồng HTTP P0/P1.
- [ADRs](docs/decisions/) — quyết định kiến trúc và lý do.

## Chạy trạng thái hiện tại

Yêu cầu Go `1.27.1` theo `app/backend/go.mod`.

```bash
cd app/backend
go run .
```

Các kiểm tra hiện có:

```bash
cd app/backend
go build ./...
go test ./...
go vet ./...
```

## Cấu trúc hiện tại

```text
app/backend/        chương trình Go tối thiểu
docs/               nguồn nghiệp vụ và đặc tả SDD
docker-compose.yaml placeholder, chưa dùng được
Dockerfile          placeholder, chưa dùng được
```

## Trạng thái SDD

`SPECIFY → PLAN → TASKS → IMPLEMENT`

Đặc tả và kế hoạch cần được hai thành viên duyệt trước khi tách task và triển khai tính năng.
