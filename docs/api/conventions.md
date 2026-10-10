# Quy ước API — SaleScale

Tài liệu này định nghĩa chuẩn mực giao tiếp giữa Client (Frontend React + Vite / Tester / k6) và Server (Backend Go + Gin). Mọi API endpoint phát triển trong hệ thống SaleScale đều phải tuân thủ nghiêm ngặt các quy ước này.

---

## 1. Prefix và Định tuyến (Routing)

- **Quyết định Prefix:** Sử dụng tiền tố `/api` thống nhất cho toàn bộ hệ thống API backend (không phân version `/api/v1` theo quyết định triển khai).
- **Quy tắc đặt tên URL:**
  - Sử dụng danh từ số nhiều (kebab-case nếu từ ghép) cho resources:
    - `/api/products`
    - `/api/flash-sales`
    - `/api/cart/items`
    - `/api/orders`
  - Các hành động đặc thù hoặc RPC-like được đặt ở đuôi bằng động từ:
    - `POST /api/orders/checkout`
    - `POST /api/auth/login`
    - `POST /api/auth/logout`
  - Endpoint kiểm tra hệ thống:
    - `GET /api/health`
    - `GET /api/ping`

---

## 2. Tiêu chuẩn Request

### 2.1. HTTP Headers
- `Content-Type: application/json` (cho các request có body: POST, PUT, PATCH).
- `Accept: application/json`.
- `X-Request-ID`: Chuỗi UUID định danh duy nhất của request. Client có thể tự sinh gửi lên; nếu client không gửi, Middleware phía Backend sẽ tự động sinh UUID v4 và gán vào context xử lý cũng như trả về trong header của Response.
- `Idempotency-Key`: Bắt buộc với các request tạo đơn hoặc thanh toán (`POST /api/orders/checkout`), định dạng UUID v4, nhằm ngăn chặn việc submit trùng lặp khi mạng chập chờn.
- `Cookie`: Xác thực phiên làm việc (Session) thông qua HTTP cookie bảo mật `salescale_session` (`HttpOnly`, `SameSite=Lax` hoặc `Strict`, `Secure` khi chạy production).

### 2.2. Request Body
- Định dạng: JSON.
- Đặt tên trường (field naming): `snake_case` (ví dụ: `first_name`, `product_id`, `quantity`, `idempotency_key`).
- Tiền tệ: Số nguyên VND (`int64`), không chứa phần thập phân (ví dụ: `150000` biểu thị 150.000 VNĐ).
- Thời gian: Chuỗi ISO 8601 UTC dạng `YYYY-MM-DDTHH:mm:ssZ`.

### 2.3. Query Parameters (Pagination, Sort, Filter)
- Phân trang chuẩn (01-indexed):
  - `page`: Số trang hiện tại, kiểu số nguyên >= 1 (mặc định: `1`).
  - `limit`: Số phần tử trên một trang, kiểu số nguyên (mặc định: `20`, tối đa: `100`).
- Sắp xếp:
  - `sort_by`: Tên trường cần sắp xếp (ví dụ: `created_at`, `price`).
  - `order`: Hướng sắp xếp, nhận giá trị `asc` hoặc `desc` (mặc định: `desc`).
- Lọc (Filter):
  - Dùng query param trực tiếp: `status=PENDING`, `keyword=iphone`, `category_id=5`.

---

## 3. Tiêu chuẩn Response

Mọi phản hồi từ backend SaleScale đều có cấu trúc JSON đồng nhất với trường nhận biết `success: true | false` và phần `meta` luôn chứa `request_id` cùng `timestamp`.

### 3.1. Success Response (2xx)

#### Cấu trúc dữ liệu đơn lẻ hoặc thao tác thành công:
```json
{
  "success": true,
  "data": {
    "id": "ord_102938475",
    "status": "PENDING",
    "total_amount": 350000,
    "expires_at": "2026-10-10T14:55:00Z"
  },
  "meta": {
    "request_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "timestamp": "2026-10-10T14:50:00Z"
  }
}
```

#### Cấu trúc danh sách có phân trang:
```json
{
  "success": true,
  "data": [
    {
      "id": "prod_1",
      "name": "Bàn phím cơ SaleScale",
      "price": 450000
    }
  ],
  "meta": {
    "request_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "timestamp": "2026-10-10T14:50:00Z",
    "pagination": {
      "page": 1,
      "limit": 20,
      "total_items": 45,
      "total_pages": 3
    }
  }
}
```

### 3.2. Error Response (4xx, 5xx)

```json
{
  "success": false,
  "error": {
    "code": "OUT_OF_STOCK",
    "message": "Sản phẩm đã hết hàng hoặc tồn kho không đủ số lượng yêu cầu.",
    "details": [
      {
        "field": "items[0].quantity",
        "issue": "Requested 5, but available stock is only 2"
      }
    ]
  },
  "meta": {
    "request_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
    "timestamp": "2026-10-10T14:50:00Z"
  }
}
```

---

## 4. Xử lý Exception và Danh mục Mã Lỗi (Error Codes)

### 4.1. Kiến trúc Exception trong Backend Go
- Go không dùng cơ chế try-catch truyền thống mà sử dụng cấu trúc lỗi tùy biến `AppError`:
```go
type AppError struct {
    StatusCode int           `json:"-"`
    Code       string        `json:"code"`
    Message    string        `json:"message"`
    Details    interface{}   `json:"details,omitempty"`
    Err        error         `json:"-"` // Internal error for logging
}
```
- **Recovery Middleware**: Tự động bắt mọi `panic`, ghi log stack trace đầy đủ (loại bỏ thông tin nhạy cảm), và tự động chuyển hóa thành response HTTP 500 với code `INTERNAL_SERVER_ERROR`.
- Không bao giờ trả lỗi thuần túy dạng raw string hoặc để lộ thông tin cấu trúc database / server ra ngoài client.

### 4.2. Từ điển Mã Lỗi (Error Codes)

| Mã lỗi (`code`) | HTTP Status | Ý nghĩa |
|---|---|---|
| `BAD_REQUEST` | 400 | Request không hợp lệ (sai cú pháp JSON, thiếu tham số bắt buộc). |
| `VALIDATION_FAILED` | 422 | Dữ liệu đầu vào không thỏa mãn quy tắc validation (chi tiết nằm trong `details`). |
| `UNAUTHORIZED` | 401 | Chưa đăng nhập hoặc session đã hết hạn / không hợp lệ. |
| `FORBIDDEN` | 403 | Không có quyền truy cập tài nguyên (ví dụ: Customer cố truy cập API Admin). |
| `NOT_FOUND` | 404 | Tài nguyên không tồn tại. |
| `METHOD_NOT_ALLOWED` | 405 | HTTP Method không được hỗ trợ trên endpoint này. |
| `CONFLICT` | 409 | Xung đột trạng thái dữ liệu (ví dụ: email đã được đăng ký). |
| `IDEMPOTENCY_CONFLICT` | 409 | Idempotency key đã được xử lý hoặc đang được xử lý bởi request khác. |
| `OUT_OF_STOCK` | 400 / 409 | Tồn kho thực tế không đủ để thực hiện giữ hàng (reservation). |
| `FLASH_SALE_LIMIT_EXCEEDED`| 400 | Người dùng đã vượt quá giới hạn mua cho phép của chương trình Flash Sale. |
| `FLASH_SALE_NOT_ACTIVE` | 400 | Chương trình Flash Sale chưa bắt đầu hoặc đã kết thúc. |
| `RESERVATION_EXPIRED` | 410 | Đơn hàng giữ chỗ đã hết hạn 5 phút và đã bị giải phóng. |
| `PAYMENT_FAILED` | 400 | Giao dịch thanh toán thất bại hoặc bị từ chối từ bộ mô phỏng. |
| `RATE_LIMITED` | 429 | Gửi quá nhiều request trong thời gian ngắn. |
| `INTERNAL_SERVER_ERROR` | 500 | Lỗi nội bộ hệ thống chưa được kiểm soát. |

