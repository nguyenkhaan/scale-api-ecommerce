# SaleScale — Hệ Thống Flash Sale Phân Tán

Hệ thống bán hàng thương mại điện tử chuyên biệt cho kịch bản **Flash Sale** với lưu lượng truy cập cao, đảm bảo tính đúng đắn tuyệt đối về tồn kho (physical stock), giới hạn mua (purchase quota), tính bất biến (idempotency) và giải phóng hàng giữ chỗ (reservation timeout).

---

## 1. Công Nghệ và Nền Tảng

| Thành phần | Công nghệ chính | Ghi chú |
|---|---|---|
| **Backend API** | **Go** + **Gin Framework** | High performance, low latency, memory safe |
| **Frontend** | **React** + **Vite** | Single Page Application, responsive, UX tối ưu cho Flash Sale |
| **Database** | **PostgreSQL** (16/17, Neon / Supabase / Local) | Nguồn dữ liệu có thẩm quyền cao nhất (Single Source of Truth) |
| **Cache Layer** | **Redis** | Chỉ lưu cache đọc catalog công khai, không quyết định checkout |
| **Reverse Proxy** | **NGINX** | Phân phối tải 1..n instance API, SSL termination |
| **Load Testing** | **k6** | Kiểm thử tải, đo p95/p99 latency và throughput |

---

## 2. Quy Ước API

Mọi endpoint dùng chung tiền tố `/api` và một cấu trúc response thống nhất (`success`, `data`, `error`, `meta`). Chi tiết về request, response và danh mục mã lỗi nằm ở [`docs/api/conventions.md`](docs/api/conventions.md).

---

## 3. Cài Đặt và Khởi Chạy

Dự án cung cấp hai cơ chế điều khiển lệnh:
1. **Qua npm scripts** (thống nhất trong monorepo).
2. **Qua các lệnh native Go CLI**.

### 3.1. Cấu hình biến môi trường
Tạo file cấu hình `.env` từ file mẫu:
```bash
cp app/backend/.env.example app/backend/.env
```

Backend tự nạp file `app/backend/.env` khi khởi động. Các biến môi trường cơ bản:
```env
PORT=8080
ENV=development
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
```

### 3.2. Bảng Lệnh Tiêu Chuẩn

#### Backend (`app/backend` hoặc qua root)
| Tác vụ | Lệnh NPM (`app/backend`) | Lệnh Go tương ứng | Mô tả |
|---|---|---|---|
| **Chạy Development** | `npm run dev` | `go run ./cmd/api` | Khởi động HTTP server Backend trên cổng `8080` |
| **Biên dịch Build** | `npm run build` | `go build -o bin/ ./cmd/api` | Biên dịch file thực thi backend vào `bin/` |
| **Kiểm tra Lint** | `npm run lint` | `go vet ./...` | Kiểm tra cú pháp và phát hiện lỗi tĩnh |
| **Định dạng Format** | `npm run format` | `go fmt ./...` | Định dạng toàn bộ mã nguồn Go |
| **Toàn bộ Test** | `npm test` | `go test ./...` | Chạy tất cả bộ kiểm thử |
| **Unit Test** | `npm run test:unit` | `go test -v -short ./...` | Chạy nhanh, bỏ qua các bộ kiểm thử tích hợp |
| **Integration Test** | `npm run test:integration` | `go test -v -run Integration ./...` | Chỉ chạy bộ kiểm thử tích hợp đi qua router và middleware |

#### Frontend (`app/frontend` hoặc qua root)
| Tác vụ | Lệnh NPM (`app/frontend`) | Lệnh từ thư mục gốc | Mô tả |
|---|---|---|---|
| **Cài đặt thư viện** | `npm install` | `npm --prefix app/frontend install` | Cài đặt React 19, Vite, Tailwind CSS v4 |
| **Chạy Development** | `npm run dev` | `npm run dev:frontend` | Khởi chạy Vite dev server tại `http://localhost:5173` |
| **Biên dịch Build** | `npm run build` | `npm run build:frontend` | Build bundle tối ưu vào `dist/` |
| **Xem trước Build** | `npm run preview` | `npm run preview:frontend` | Xem trước production build |

*(Ghi chú: Tại thư mục gốc repository, bạn có thể chạy `npm run build` để build đồng thời cả Backend và Frontend).*

---

## 4. Kiểm Tra Hoạt Động Của Hệ Thống

### 4.1. Khởi chạy đồng thời
1. Khởi động Backend: `npm run dev:backend` (hoặc `cd app/backend && npm run dev`) -> Lắng nghe tại `http://localhost:8080`.
2. Khởi động Frontend: `npm run dev:frontend` (hoặc `cd app/frontend && npm run dev`) -> Mở trình duyệt tại `http://localhost:5173`.
3. Vite Dev Server đã cấu hình proxy tự động chuyển tiếp toàn bộ request `/api/*` về `http://localhost:8080/api/*`.
4. Mở `http://localhost:5173` trên trình duyệt để kiểm tra trạng thái kết nối tự động tới Backend.

### 4.2. Kiểm tra Liveness & Ping từ CLI
```bash
curl http://localhost:8080/api/ping
```
Kết quả:
```json
{
  "success": true,
  "data": {
    "message": "pong"
  },
  "meta": {
    "request_id": "...",
    "timestamp": "2026-10-10T14:55:00Z"
  }
}
```

### 4.3. Kiểm tra Health Check
```bash
curl http://localhost:8080/api/health
```
Kết quả:
```json
{
  "success": true,
  "data": {
    "environment": "development",
    "service": "salescale-backend",
    "status": "UP",
    "uptime": "15s"
  },
  "meta": {
    "request_id": "...",
    "timestamp": "2026-10-10T14:55:15Z"
  }
}
```

### 4.4. Kiểm tra Phản hồi Lỗi Chuẩn (404 Not Found)
```bash
curl http://localhost:8080/api/route-khong-ton-tai
```
Kết quả:
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Không tìm thấy tài nguyên /api/route-khong-ton-tai yêu cầu"
  },
  "meta": {
    "request_id": "...",
    "timestamp": "2026-10-10T14:55:30Z"
  }
}
```

---

## 5. Cấu Trúc Thư Mục Backend

```text
app/backend/
├── cmd/
│   └── api/
│       └── main.go           # Điểm khởi động HTTP Server & Graceful Shutdown
├── internal/
│   ├── config/               # Đọc và quản lý cấu hình môi trường (.env)
│   ├── errors/               # Khái niệm AppError & danh mục Error Codes
│   ├── handler/              # Controller xử lý logic (/api/health, /api/ping)
│   ├── middleware/           # RequestID, Structured Logger, Panic Recovery, CORS
│   ├── response/             # Chuẩn hóa JSON Response & Metadata
│   └── router/               # Định tuyến Router với tiền tố /api
├── .env.example              # Mẫu biến môi trường
├── go.mod                    # Quản lý Golang modules
└── package.json              # NPM scripts wrapper cho các tác vụ chuẩn
```

