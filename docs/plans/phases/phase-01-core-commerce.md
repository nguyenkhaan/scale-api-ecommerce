# Phase 01 — Xây chức năng bán hàng lõi

**Thời gian:** Tuần 1–4  
**Phụ thuộc:** Phase 00  
**Kết quả cần đạt:** Quản trị chuẩn bị hàng và flash sale; khách đăng nhập, xem sản phẩm, tạo giỏ và xem trước đơn.

## Nghiên cứu cần làm

- [ ] **R01.1 — Thiết kế dữ liệu bán hàng.** Dùng PostgreSQL và BR/INV trong PRD để vẽ bảng tài khoản, session, sản phẩm, SKU, tồn, chương trình sale và giỏ. Bản thiết kế giúp xác định dữ liệu nào phải bền vững để giá và tồn không bị quyết định bởi trình duyệt.
- [ ] **R01.2 — Thiết kế đăng nhập qua nhiều API instance.** Đọc middleware Go/Gin và cookie HTTP; lưu session trong PostgreSQL, dùng HttpOnly/Secure/SameSite và CSRF token. Khi NGINX chuyển một khách sang API khác, khách vẫn đăng nhập và request thay đổi dữ liệu vẫn được bảo vệ.
- [ ] **R01.3 — Chọn nguồn cho giá và tồn.** Dùng PostgreSQL làm nguồn dữ liệu chính; catalog và preview chỉ để khách xem tại thời điểm đọc. Checkout sẽ đọc lại PostgreSQL để không bán theo giá hoặc tồn cũ.
- [ ] **R01.4 — Chốt hợp đồng API/giao diện.** Dùng Go/Gin và React/Vite để thống nhất endpoint, quyền, dữ liệu vào/ra và mã lỗi trước khi chia việc. Hai người có thể làm backend và giao diện song song mà ít phải sửa lúc ghép.

## Việc triển khai

- [ ] **C01.1 — Tạo cấu trúc DB cho nghiệp vụ.** Dùng golang-migrate tạo bảng PostgreSQL và pgx để Go đọc/ghi; migration giúp hai máy phát triển có cùng cấu trúc cho tài khoản, hàng hóa, tồn và sale.
- [ ] **C01.2 — Làm tài khoản và quyền.** Dùng Go/Gin, PostgreSQL session, golang.org/x/crypto/bcrypt, middleware và CSRF token để triển khai đăng ký/đăng nhập/đăng xuất. Việc này cho phép nhiều API cùng kiểm tra danh tính mà không cho khách xem giỏ người khác hoặc gọi chức năng admin.
- [ ] **C01.3 — Làm catalog và quản lý SKU.** Dùng PostgreSQL cho dữ liệu sản phẩm/SKU và Go/Gin cho endpoint quản trị/catalog; ghi lịch sử điều chỉnh tồn để truy ra ai đổi kho, đồng thời từ chối SKU trùng hoặc tồn âm.
- [ ] **C01.4 — Làm chương trình flash sale.** Dùng Go/Gin và PostgreSQL để lưu giờ mở/đóng, giá ưu đãi, quota và giới hạn mua theo SKU. Kiểm tra cấu hình trước khi chạy để khách không đặt theo giá hết hạn hoặc chương trình chồng nhau.
- [ ] **C01.5 — Làm giỏ và preview.** Dùng PostgreSQL lưu giỏ theo tài khoản và Go/Gin tính tổng VND nguyên từ giá DB. Giỏ chỉ ghi ý định mua; bước này không giữ hàng và không sửa quota.
- [ ] **C01.6 — Làm giao diện demo tối thiểu.** Dùng React/Vite cho đăng nhập, danh sách hàng, trang admin, giỏ và preview; dùng hợp đồng API của R01.4 để nối giao diện với Go/Gin mà không đặt logic giá/tồn ở frontend.
- [ ] **C01.7 — Tạo dữ liệu mẫu an toàn.** Dùng SQL hoặc lệnh Go chỉ cho PostgreSQL test để nạp tài khoản, SKU, tồn và campaign giống nhau. Dữ liệu cố định giúp k6 so sánh các phiên bản trong cùng điều kiện và tránh reset DB thật.

## Tiêu chí hoàn thành

- [ ] Khách đăng nhập và không đọc được giỏ/tài khoản khác; admin không thể được cấp quyền từ request khách.
- [ ] Admin tạo sản phẩm, SKU, tồn và chương trình; thao tác sai bị Go/Gin từ chối.
- [ ] Khách xem giá, tạo giỏ và preview; các thao tác đọc không giữ hàng hoặc trừ quota.
- [ ] Tổng tiền là VND nguyên; migration chạy được trên PostgreSQL trống.
- [ ] Có thể dùng trình duyệt hoàn thành luồng khách/admin cơ bản.

**Đưa vào báo cáo:** Sơ đồ dữ liệu PostgreSQL, phân quyền Go/Gin và lý do frontend không quyết định giá/tồn.
