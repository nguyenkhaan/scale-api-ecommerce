# ADR-007: Dùng GORM/pgx và Goose cho PostgreSQL

## Trạng thái

Accepted

## Ngày

2026-10-24

## Bối cảnh

Phase 00 cần chốt TECH-01 để kết nối Neon PostgreSQL, quản lý schema có version và dùng middleware HTTP nhất quán. Dự án đã có Gin, GORM và PostgreSQL driver dựa trên pgx; Goose bổ sung migration SQL có version và lệnh vận hành nhất quán.

## Quyết định

- Gin tiếp tục là HTTP router và middleware stack.
- GORM dùng cho ORM/repository; PostgreSQL driver của GORM dùng pgx bên dưới.
- Goose quản lý migration SQL có version. Migration được embed vào binary để mọi môi trường chạy cùng một nội dung.
- Migration chỉ tiến lên trong luồng triển khai thông thường. Reset chỉ dành cho test database và cần đồng thời `ENV=test` cùng `ALLOW_DATABASE_RESET=true`.
- Seed admin dùng câu lệnh có tham số, chỉ chạy ở development/test khi `ALLOW_ADMIN_SEED=true`, và chỉ lưu password hash.

## Phương án đã cân nhắc

### GORM AutoMigrate

Đơn giản nhưng không cung cấp lịch sử SQL rõ ràng hoặc rollback có kiểm soát. Không phù hợp yêu cầu migration chạy lặp và tái lập lab.

### Chỉ dùng Goose CLI ngoài ứng dụng

Ít mã Go hơn nhưng yêu cầu mỗi môi trường cài đúng CLI. Binary migration riêng với SQL embed tạo cùng contract cho Neon development, test và môi trường triển khai.

### Viết migration runner riêng

Giảm dependency nhưng phải tự xử lý version, transaction, trạng thái và lỗi. Không chọn vì lặp lại chức năng đã ổn định của Goose.

## Hệ quả

- Schema thay đổi phải đi qua migration SQL mới; không sửa migration đã áp dụng ở môi trường chia sẻ.
- Repository vẫn dùng GORM và có thể lấy `*sql.DB` khi transaction hoặc công cụ migration cần API chuẩn.
- Goose trở thành dependency build của các command migration, nhưng không chạy tự động trong API process.
- Phase 01 có thể mở rộng bảng `users` bằng migration mới mà không đổi lịch sử Phase 00.
