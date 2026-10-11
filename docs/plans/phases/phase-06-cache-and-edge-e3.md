# Phase 06 — Redis, CDN và cache an toàn: E3

**Thời gian:** Tuần 9–10  
**Phụ thuộc:** Phase 05  
**Kết quả cần đạt:** Tăng tốc GET catalog mà không cache nhầm dữ liệu riêng tư hoặc dùng dữ liệu cũ để quyết định checkout.

## Nghiên cứu cần làm

- [ ] **R06.1 — Nghiên cứu cache-aside.** Đọc [Redis cache-aside](https://redis.io/docs/latest/develop/use-cases/cache-aside/) để hiểu Go/Gin đọc Redis trước, đọc PostgreSQL khi cache miss và lưu bản sao cho request catalog tiếp theo. Cách này nhắm vào feature đọc sản phẩm, không đổi feature checkout.
- [ ] **R06.2 — Nghiên cứu HTTP cache/CDN.** Đọc [RFC 9111](https://www.rfc-editor.org/rfc/rfc9111) và [Cloudflare cache behavior](https://developers.cloudflare.com/cache/concepts/default-cache-behavior/). Gắn Cache-Control theo route để chỉ lưu nội dung công khai và không dùng chung cart/order/session/notification giữa các khách.
- [ ] **R06.3 — Ràng buộc TTL theo giờ sale.** Dùng campaign trong PostgreSQL để đặt TTL không vượt giờ sale mở/đóng hoặc bị dừng. Điều này ngăn catalog báo giá sale đã hết hiệu lực.
- [ ] **R06.4 — Ghi giới hạn phép thử edge.** So tài liệu Cloudflare với NGINX proxy_cache. Một NGINX cache ở lab chỉ thử hành vi lưu nội dung, không chứng minh lợi ích về vị trí máy chủ toàn cầu.

## Việc triển khai và đo

- [ ] **C06.1 — Chọn route được cache.** Dùng Go/Gin đặt Cache-Control cho GET danh sách/chi tiết sản phẩm công khai và no-store cho login, session, cart, preview, order, payment, notification inbox và admin. Điều này bảo vệ dữ liệu riêng tư.
- [ ] **C06.2 — Thêm cache catalog bằng go-redis.** Dùng cache-aside cho danh sách/chi tiết sản phẩm; đưa filter, page và SKU vào Redis key để các truy vấn khác nhau không trả cùng một nội dung.
- [ ] **C06.3 — Làm mới cache khi admin sửa hàng.** Dùng API Go/Gin xóa Redis key liên quan khi sản phẩm hoặc campaign đổi và đặt TTL theo giờ sale. TTL giúp dữ liệu cũ tự hết hạn nếu invalidation bị bỏ lỡ.
- [ ] **C06.4 — Giới hạn kết nối Redis.** Dùng một go-redis client dùng chung, đặt PoolSize/PoolTimeout theo phiên bản thư viện và đo chờ pool. Việc này tránh mỗi request GET mở kết nối mới khi Redis chậm.
- [ ] **C06.5 — Đo cache stampede.** Dùng k6 gửi nhiều GET cùng key khi Redis rỗng/hết TTL và đếm query PostgreSQL. Chỉ thêm Go singleflight nếu số đo cho thấy nhiều request cùng đọc lại một sản phẩm gây dồn tải DB.
- [ ] **C06.6 — Thử cache tại NGINX.** Nếu còn thời gian, bật NGINX proxy_cache cho GET catalog công khai; đặt cache key/TTL rõ và bypass cookie, Authorization, Set-Cookie. Điều này kiểm tra một phần hành vi edge mà không làm cache notification cá nhân.
- [ ] **C06.7 — Đo E2/E3 hỗn hợp.** Dùng k6 chạy GET catalog cùng checkout khi Redis lạnh, ấm, vừa invalidated, tắt và hết TTL đồng thời. So p95/p99, DB load, đơn hợp lệ và cache hit/miss để biết cache giúp đọc nhanh có làm checkout bị ảnh hưởng không.

## Tiêu chí hoàn thành

- [ ] TEST-22 xác nhận checkout luôn đọc giá/tồn từ PostgreSQL, không từ Redis/CDN.
- [ ] Thử admin cập nhật giá/tên/campaign, Redis lỗi và hai tài khoản có dữ liệu riêng.
- [ ] Notification inbox, session, cart và order có no-store hoặc không được đưa vào public cache.
- [ ] Ghi hit/miss, DB query, p95/p99, đơn hợp lệ và tối đa dữ liệu cũ có thể hiển thị.

**Đưa vào báo cáo:** Route/header được cache, Redis/NGINX config, kết quả E2/E3 và giới hạn của CDN giả lập.
