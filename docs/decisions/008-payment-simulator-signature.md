# ADR-008: Ký callback của payment simulator bằng HMAC-SHA256 có timestamp

## Trạng thái

Accepted

## Ngày

2026-11-22

## Bối cảnh

TECH-02 yêu cầu chốt cách xác thực callback từ bộ mô phỏng thanh toán trước khi xây simulator. BR-40 chỉ tin kết quả từ nguồn thử nghiệm được xác thực; TEST-17 kiểm sai chữ ký, sai phiên, sai số tiền/tiền tệ. Callback đi qua HTTP (sau này qua NGINX tới nhiều API instance), nên xác thực phải là stateless và không dựa vào session trình duyệt.

## Quyết định

- Simulator gửi header `X-Simulator-Signature: t=<unix seconds>,v1=<hex>` với `v1 = HMAC-SHA256(secret, "<t>." + raw body)`. Secret lấy từ biến môi trường `PAYMENT_SIMULATOR_SECRET`, không hard-code, không log.
- Backend tính lại HMAC trên đúng byte của body, so sánh thời gian hằng và từ chối khi lệch giờ quá 300 giây. Chống phát lại trong cửa sổ đó dựa vào `event_id` duy nhất.
- Lần gửi sai chữ ký vẫn được lưu bền vững với `source_verified = false`, `result = 'rejected'` để truy vết, nhưng không bao giờ chiếm chỗ `event_id`: dedup chỉ áp dụng cho sự kiện đã xác thực (unique index một phần), nên sự kiện thật cùng ID đến sau vẫn được xử lý.
- Điều khiển simulator (`/payment-simulator/*`) chỉ bật ở dev/test và cần `X-Operator-Key` hoặc session quản trị; khách không có quyền này (PRD §3.2).
- Secret rỗng thì không mount callback và simulator (log cảnh báo), để máy dev chưa cấu hình vẫn chạy được API. Ngoài dev/test, secret ngắn hơn 32 byte hoặc trùng giá trị mẫu trong `.env.example` làm API dừng khi khởi động.
- Sự kiện chưa xác thực được lưu với `payment_session_id = NULL`, để không trộn vào lịch sử của phiên thật. Sự kiện đã xác thực trùng `event_id` nhưng khác nội dung (băm trên `event_id`, phiên, loại, số tiền, tiền tệ) được coi là mâu thuẫn: trả `conflict` và mở case đối soát.

## Phương án đã cân nhắc

### HMAC không có timestamp

Đơn giản hơn nhưng một callback hợp lệ bị bắt được có thể gửi lại vô thời hạn; chỉ còn `event_id` chặn lặp. Không chọn.

### Chữ ký bất đối xứng (Ed25519)

Không cần chia sẻ secret với API nhưng thêm quản lý khóa cho một simulator nội bộ. Không cần ở MVP.

### Mutual TLS / IP allowlist

Phụ thuộc hạ tầng triển khai và khó kiểm thử ở lab. Không chọn.

## Hệ quả

- API và simulator phải dùng chung `PAYMENT_SIMULATOR_SECRET`; đổi secret cần đổi cả hai.
- Đồng hồ máy gửi và máy nhận phải lệch dưới 5 phút.
- Bảng `payment_events` lưu cả lần gửi chưa xác thực; khi có rate limit (Phase 07) cần áp cho endpoint callback để tránh làm đầy bảng.
