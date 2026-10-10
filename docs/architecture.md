# Kiến trúc mục tiêu — SaleScale

## Thành phần

```text
React + Vite
      |
    NGINX
      |
Go/Gin API (1..n) ---- Redis (chỉ cache catalog công khai)
      |
 PostgreSQL <---- Worker hết hạn
      ^
      |
Bộ mô phỏng thanh toán
```

- PostgreSQL là nguồn dữ liệu có thẩm quyền cho session, đơn, tồn, quota, giới hạn mua và idempotency.
- Redis không quyết định checkout và không lưu giỏ, đơn hoặc thanh toán riêng tư làm nguồn duy nhất.
- Worker dùng chung code nghiệp vụ và PostgreSQL với API, nhưng chạy như tiến trình độc lập.
- NGINX phân phối request cho một hoặc nhiều API instance; không dùng sticky session làm điều kiện đúng đắn.

## Xác thực

- Backend phát session cookie an toàn; frontend gửi cookie theo credential policy của API.
- Session được lưu trong PostgreSQL để logout, hết hạn và chuyển API instance có cùng kết quả.
- Cookie production dùng `HttpOnly`, `Secure` và `SameSite` phù hợp; request thay đổi trạng thái phải có bảo vệ CSRF.
- Mật khẩu được băm; secret chỉ lấy từ cấu hình môi trường và không ghi log.

## Frontend

- Dùng React + Vite trong cấu trúc frontend đã có của repository.
- Frontend chỉ hiển thị trạng thái; backend vẫn quyết định giá, tồn, quota, thời hạn và trạng thái đơn/thanh toán.

## Ràng buộc chính

- Checkout nhiều SKU và việc giữ hàng chạy trong transaction PostgreSQL ngắn, nguyên tử.
- Một đơn giữ hàng 5 phút và có một payment session từ bộ mô phỏng.
- API, callback, hủy đơn và worker cùng kiểm tra trạng thái/thời hạn trước khi kết thúc reservation.
- Không thêm microservice, Kubernetes, sharding, thanh toán thật hoặc tích hợp vận chuyển trong MVP.
