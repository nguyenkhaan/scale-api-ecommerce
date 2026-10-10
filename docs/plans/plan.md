# Kế hoạch triển khai SaleScale

## 1. Mục đích của kế hoạch

Tài liệu này là checklist triển khai toàn bộ đồ án SaleScale trong 12 tuần với nhóm 2 sinh viên. Kế hoạch bắt đầu từ việc chốt yêu cầu và dựng database, sau đó hoàn thiện từng tính năng theo chiều dọc: database → backend → frontend → kiểm thử → tích hợp.

Mục tiêu cuối cùng là có:

- Backend Go + Gin chạy đúng nghiệp vụ flash sale.
- Giao diện demo tối thiểu cho khách hàng và quản trị viên.
- PostgreSQL làm nguồn dữ liệu chính; Redis chỉ hỗ trợ cache đọc.
- Thanh toán mô phỏng, worker hết hạn đơn và xử lý callback an toàn.
- Cấu hình chạy 1 hoặc nhiều API instance sau NGINX.
- Bộ test nghiệp vụ, test đồng thời và benchmark k6 có thể chạy lại.
- Báo cáo so sánh trước/sau tối ưu, kèm dữ liệu gốc và giới hạn của kết quả.

## 2. Tài liệu nguồn

Kế hoạch phải bám các tài liệu sau:

- `docs/prd.md`: phạm vi, quy tắc nghiệp vụ, bất biến, hiệu năng và nghiệm thu.
- `docs/function-requirement.md`: FR-001 đến FR-046.
- `docs/user-story.md`: US-01 đến US-20.
- `docs/architecture.md`: kiến trúc mục tiêu và các ràng buộc kỹ thuật.
- `docs/spec/product-spec.md`: các quyết định sản phẩm đã xác nhận.

Nếu các tài liệu mâu thuẫn, nhóm phải dừng tính năng liên quan, ghi lại điểm mâu thuẫn và chốt với giảng viên/chủ sản phẩm trước khi viết tiếp.

## 3. Nguyên tắc triển khai

### 3.1. Làm theo tính năng, không làm theo tầng

Không xây toàn bộ database rồi mới làm toàn bộ API và cuối cùng mới làm giao diện. Sau phần nền tảng tối thiểu, mỗi tính năng phải được hoàn thiện theo thứ tự:

1. Chốt hành vi và hợp đồng API của tính năng.
2. Thêm migration và dữ liệu cần thiết cho tính năng đó.
3. Làm backend và test backend.
4. Làm frontend song song bằng hợp đồng API hoặc dữ liệu giả.
5. Kết nối frontend với backend.
6. Chạy test tự động và kiểm tra bằng trình duyệt.
7. Chỉ chuyển sang tính năng tiếp theo khi checkpoint đạt.

### 3.2. Backend và frontend chạy song song

Với mỗi tính năng:

- Thành viên BE chốt request, response, mã lỗi và migration trước.
- Thành viên FE dùng hợp đồng API đã chốt để làm màn hình và trạng thái giao diện ngay, không chờ BE hoàn thành toàn bộ.
- Khi API sẵn sàng, FE bỏ dữ liệu giả và chạy kiểm tra tích hợp.
- Nếu hợp đồng API thay đổi, phải cập nhật tài liệu trước rồi mới sửa hai phía.

### 3.3. Ưu tiên tính đúng đắn trước hiệu năng

- Không benchmark một phiên bản còn vi phạm tồn kho, quota, idempotency hoặc thanh toán.
- Không bỏ kiểm tra nghiệp vụ để làm số liệu nhanh hơn.
- Không dùng số request bị từ chối nhanh để tuyên bố throughput đơn hàng tăng.
- PostgreSQL quyết định đơn, tồn, quota và idempotency; Redis không được là nguồn quyết định duy nhất.

### 3.4. Giữ phạm vi vừa sức đồ án

- Dùng một ứng dụng backend chia module; không tách microservice.
- Worker hết hạn có thể là một chế độ chạy khác của cùng mã nguồn backend.
- Không thêm Kubernetes, sharding, thanh toán thật, vận chuyển thật hoặc marketplace.
- Chỉ thêm thư viện khi thư viện chuẩn hoặc công cụ đang có không giải quyết được việc cần làm.

## 4. Các quyết định triển khai

Các giá trị dưới đây đã được chọn làm cơ sở triển khai và nghiệm thu. Mọi thay đổi phải cập nhật PRD, đặc tả sản phẩm, tiêu chí nghiệm thu và test liên quan.

| Nội dung | Quyết định |
|---|---|
| Thời gian giữ hàng | 5 phút, lưu hạn cụ thể trên đơn |
| Quota flash sale | Dùng chung tồn vật lý; quota chỉ là trần ưu đãi |
| Giới hạn mua | Theo tài khoản + chương trình + SKU; tính cả đang giữ và đã mua |
| Giỏ thường và flash sale | Cho phép cùng giỏ, mỗi dòng ghi rõ nguồn giá |
| Chọn dòng checkout | Cho chọn một tập dòng; không tự lấy toàn bộ giỏ |
| Giỏ sau khi đặt | Giữ nguyên khi chờ; chỉ xóa an toàn sau khi xác nhận |
| Thanh toán | Chỉ dùng bộ mô phỏng bắt buộc |
| Phiên thanh toán | Một phiên cho một đơn; retry trả phiên cũ |
| Tiền | VND số nguyên; không phí vận chuyển, không tách thuế |
| Người nhận | Tên, số điện thoại, địa chỉ văn bản |
| Giới hạn giỏ | Tối đa 20 SKU; số lượng theo cấu hình SKU/chương trình |
| Xác thực | Session cookie an toàn; session lưu ở PostgreSQL để dùng được nhiều instance |
| FE | React + Vite |
| Mục tiêu hiệu năng | p95 checkout ≤ 500 ms, p99 ≤ 1.000 ms tại tải tham chiếu; lỗi hệ thống < 1% |
| Chậm giải phóng | 95% reservation được giải phóng trong 30 giây sau hạn tại tải tham chiếu |

## 5. Kiến trúc mục tiêu tối thiểu

```text
Trình duyệt khách/quản trị
          |
        NGINX
          |
   +------+------+
   |             |
Go/Gin API 1  Go/Gin API 2
   |             |
   +------+------+
          |
     PostgreSQL  <--- Worker hết hạn
          |
        Redis (cache catalog, không quyết định checkout)

Bộ mô phỏng thanh toán ---> callback qua NGINX ---> API
k6 -----------------------> NGINX/API
```

Các module backend dự kiến:

- `auth`: đăng ký, đăng nhập, đăng xuất, session và phân quyền.
- `store`: thông tin cửa hàng.
- `catalog`: danh mục, sản phẩm và SKU.
- `inventory`: tồn kho, điều chỉnh và reservation.
- `sale`: chương trình flash sale, quota và giới hạn mua.
- `cart`: giỏ và xem trước giá.
- `order`: checkout, idempotency, đơn và vòng đời đơn.
- `payment`: phiên thanh toán, sự kiện và đối soát.
- `worker`: quét và giải phóng đơn hết hạn.
- `observability`: log, metrics và request ID.

## 6. Cấu trúc thư mục mục tiêu

Đây là cấu trúc dự kiến, có thể điều chỉnh nhẹ khi chốt kiến trúc:

```text
app/
  backend/
    cmd/api/                 # Điểm chạy API
    cmd/worker/              # Điểm chạy worker hết hạn
    internal/                # Module nghiệp vụ
    migrations/              # Migration SQL theo từng tính năng
    tests/                   # Test tích hợp và test đồng thời
  frontend/
    src/
      api/                   # Hàm gọi API
      features/              # Màn hình/component theo tính năng
      routes/                # Route khách và quản trị
      shared/                # Thành phần dùng chung tối thiểu
tests/
  load/                      # Kịch bản k6
  evidence/                  # Kết quả kiểm chứng, không chứa bí mật
docs/
  plans/
  api/
  reports/
```

Không tạo abstraction hoặc thư mục chỉ để “dùng sau”. Chỉ tạo khi tính năng đầu tiên thật sự cần.

## 7. Definition of Done dùng cho mọi bước

Một bước chỉ được đánh dấu hoàn thành khi:

- [ ] Hành vi bám đúng BR, INV, FR, US và TEST liên quan.
- [ ] Migration có đường chạy rõ ràng và có thể dựng database mới từ đầu.
- [ ] Backend kiểm tra dữ liệu đầu vào, quyền truy cập và lỗi nghiệp vụ.
- [ ] Không log mật khẩu, token, secret hoặc dữ liệu riêng tư không cần thiết.
- [ ] Test nhỏ nhất cần thiết đã được thêm và chạy thành công.
- [ ] Nếu có UI, đã kiểm tra trạng thái tải, thành công, lỗi và thao tác lặp bằng trình duyệt.
- [ ] Tài liệu API được cập nhật cùng thay đổi.
- [ ] Lệnh build/test của BE và FE chạy thành công.
- [ ] Không làm hỏng tính năng đã hoàn thành ở giai đoạn trước.

---

# 8. Kế hoạch theo giai đoạn

## Giai đoạn 0 — Hoàn thiện đặc tả và cách làm chung (Tuần 1)

### Mục tiêu

Hoàn thiện hợp đồng kỹ thuật đủ rõ để hai thành viên làm song song mà không tự đoán khác nhau. Các quyết định sản phẩm ở Mục 4 đã được ghi trong `docs/spec/product-spec.md`.

### Bước 0.1 — Viết tài liệu kiến trúc tối thiểu

- [ ] Hoàn thiện `docs/architecture.md` với sơ đồ thành phần và luồng request.
- [ ] Ghi rõ cách tổ chức backend theo module, không tách microservice.
- [ ] Ghi rõ PostgreSQL là nguồn dữ liệu có thẩm quyền.
- [ ] Ghi rõ Redis chỉ cache dữ liệu catalog công khai.
- [ ] Mô tả worker hết hạn chạy độc lập nhưng dùng chung code và database.
- [ ] Mô tả cách chạy 1 và 2 API instance qua NGINX.

**Kết quả cần có:** Hai thành viên hiểu cùng một kiến trúc và giới hạn trách nhiệm của từng thành phần.

### Bước 0.2 — Chốt quy ước API trước khi BE/FE tách việc

- [x] Chọn prefix và version API: thống nhất chọn `/api` cho toàn bộ hệ thống (xem chi tiết tại `docs/api/conventions.md`).
- [x] Chốt cấu trúc response thành công và response lỗi có `success`, `data`, `error`, `meta` đồng nhất.
- [x] Mọi lỗi có `request_id`, mã lỗi máy đọc được (`code`), thông báo ngắn và chi tiết theo dòng khi cần (`details`).
- [x] Chốt phân trang (`page`, `limit`), lọc, tìm kiếm và định dạng thời gian UTC (`ISO 8601`).
- [x] Chốt cách gửi idempotency key ở checkout (`Idempotency-Key` header dạng UUID).
- [x] Đặc tả cách FE gửi session cookie (`salescale_session`) và cách backend bảo vệ CSRF.
- [x] Viết hợp đồng API đầu tiên cho auth, catalog và health check (`docs/api/conventions.md`).

**Phân công song song:** BE viết request/response; FE rà lại hợp đồng từ góc nhìn màn hình và trạng thái lỗi.

### Bước 0.3 — Chốt công cụ và lệnh chuẩn

- [x] Dùng React + Vite; chốt package manager là `npm` và phiên bản Node v20/v22 LTS.
- [x] Chốt thư viện truy cập PostgreSQL (`pgx/v5`), công cụ migration (`golang-migrate/migrate/v4`) và test (`testify`).
- [x] Chốt thư viện tối thiểu cho Gin (`gin-gonic/gin`), session/password (`bcrypt`), Redis (`go-redis/v9`) và metrics.
- [x] Ghi lệnh chạy dev, build, format, lint, unit test, integration test và benchmark vào README (hỗ trợ cả Go CLI và NPM scripts).
- [x] Mọi máy thành viên phải chạy được cùng lệnh, không phụ thuộc cấu hình bí mật cá nhân (dùng `.env.example`).

### Checkpoint Giai đoạn 0

- [x] `product-spec.md` và `architecture.md` phản ánh đúng quyết định và kiến trúc triển khai.
- [x] Hợp đồng API nền tảng được BE và FE cùng xác nhận (`docs/api/conventions.md`).
- [x] Các lệnh dự án và quy tắc branch/commit đã rõ.

---

## Giai đoạn 1 — Môi trường chạy và nền tảng database (Tuần 1–2)

### Mục tiêu

Dựng được môi trường local có PostgreSQL, Redis, backend và frontend; migration chạy được từ database trống.

### Bước 1.1 — Hoàn thiện Docker Compose cho môi trường local

- [ ] Thêm PostgreSQL với volume, health check, user/database lấy từ biến môi trường.
- [ ] Thêm Redis với health check.
- [ ] Thêm API, frontend và worker ở mức scaffold; NGINX có thể thêm ở cuối bước này hoặc trước bài test nhiều instance.
- [ ] Không commit password thật; chỉ cập nhật `.env.example`.
- [ ] Thêm lệnh dựng mới, dừng và reset riêng môi trường test/benchmark.
- [ ] Kiểm tra restart API không làm mất dữ liệu PostgreSQL.

### Bước 1.2 — Tạo cơ chế migration

- [ ] Tạo thư mục migration và bảng theo dõi phiên bản migration.
- [ ] Viết migration đầu tiên cho extension hoặc kiểu dữ liệu dùng chung nếu thật sự cần.
- [ ] Có lệnh chạy migration lên database mới.
- [ ] Có cách rollback migration trong giai đoạn phát triển.
- [ ] Không dùng thao tác reset phá dữ liệu ngoài môi trường local/test.

### Bước 1.3 — Tạo khung backend có thể chạy

- [ ] Thay `Hello world` bằng entrypoint API tối thiểu.
- [ ] Đọc cấu hình từ biến môi trường và dừng sớm nếu thiếu cấu hình bắt buộc.
- [ ] Tạo connection pool PostgreSQL với giới hạn rõ ràng.
- [ ] Kết nối Redis nhưng không làm health của nghiệp vụ ghi phụ thuộc Redis.
- [ ] Thêm endpoint liveness và readiness.
- [ ] Thêm shutdown có kiểm soát để đóng HTTP server và connection pool.

### Bước 1.4 — Tạo khung frontend song song

- [ ] Khởi tạo ứng dụng FE theo công cụ đã chốt.
- [ ] Thêm router, layout khách và layout quản trị tối thiểu.
- [ ] Tạo lớp gọi API dùng chung và đọc base URL từ biến môi trường.
- [ ] Tạo trang kiểm tra kết nối health API.
- [ ] Có màn hình lỗi kết nối rõ ràng, không để trang trắng.

### Bước 1.5 — Kiểm tra tích hợp nền tảng

- [ ] Từ máy mới, làm theo README để dựng toàn bộ môi trường.
- [ ] Backend kết nối PostgreSQL thành công. PostgresSQL được sử dụng trên Neondb hoặc Suppabase
- [ ] FE gọi được health API qua cấu hình local.
- [ ] Dừng Redis vẫn không làm sai trạng thái PostgreSQL; readiness phản ánh đúng phụ thuộc bắt buộc.
- [ ] Migration chạy thành công trên database trống.

### Checkpoint Giai đoạn 1

- [ ] Docker Compose chạy ổn định.
- [ ] BE và FE build/test được độc lập.
- [ ] Database có thể dựng lại từ đầu.
- [ ] Chưa có dữ liệu bí mật trong Git.

---

## Giai đoạn 2 — Tài khoản, phân quyền và thông tin cửa hàng (Tuần 2–3)

**Phạm vi:** FR-001 đến FR-004; US-01.

### Bước 2.1 — Chốt hợp đồng tính năng

- [ ] Chốt API đăng ký, đăng nhập, đăng xuất, lấy thông tin người dùng hiện tại.
- [ ] Chốt API xem/sửa thông tin cửa hàng.
- [ ] Chốt role `CUSTOMER` và `ADMIN`; không cho client tự chọn `ADMIN`.
- [ ] Chốt thời gian hết hạn session và hành vi khi session hết hạn.
- [ ] Chốt cách tạo tài khoản admin ban đầu bằng lệnh hoặc seed được bảo vệ.

### Bước 2.2 — BE: migration và nghiệp vụ tài khoản

- [ ] Tạo bảng người dùng với email chuẩn hóa, unique, password hash, role và trạng thái.
- [ ] Tạo bảng session PostgreSQL dùng chung cho mọi API instance.
- [ ] Tạo bảng cấu hình cửa hàng một bản ghi cho mỗi deployment.
- [ ] Băm mật khẩu bằng cơ chế phù hợp; không lưu hoặc trả password rõ.
- [ ] Thêm giới hạn thử đăng nhập và không tiết lộ email có tồn tại qua lỗi không cần thiết.
- [ ] Tạo các API endpoint cần thiết cho quá trình phần quyền: login, register. Ngoài ra có thể bổ sung (không nằm trong nghiệp vụ) một số API cần thiết nữa cho authentication: change-email, change-password, refresh. Các API này cần phải thực hiện đúng nghiệp vụ. Khi thay đổi các thông tin quan trọng (email, password) cần phải có bước verify để xác minh hành động của người dùng có hợp lệ không.
- [ ] Middleware kiểm tra đăng nhập, role và quyền sở hữu server-side.
- [ ] Ghi audit cho thay đổi quản trị quan trọng.

### Bước 2.3 — FE: luồng tài khoản và cửa hàng chạy song song

- [ ] Làm trang đăng ký, đăng nhập và đăng xuất.
- [ ] Hiển thị lỗi validation và sai thông tin đăng nhập rõ ràng.
- [ ] Chặn route quản trị với người không có quyền; backend vẫn là nơi quyết định cuối cùng.
- [ ] Làm trang quản trị thông tin cửa hàng.
- [ ] Làm phần hiển thị tên/logo/liên hệ công khai.
- [ ] Xử lý session hết hạn bằng cách đưa người dùng về trang đăng nhập mà không lộ dữ liệu cũ.

### Bước 2.4 — Test và tích hợp

- [ ] Test email trùng, email không hợp lệ và mật khẩu sai.
- [ ] Test logout/session hết hạn không truy cập API riêng tư.
- [ ] Test khách không gọi được API admin.
- [ ] Test khách A không đọc dữ liệu riêng của khách B.
- [ ] Kiểm tra luồng đăng ký → đăng nhập → xem trang cá nhân bằng trình duyệt.

### Checkpoint Giai đoạn 2

- [ ] US-01 đạt đầy đủ.
- [ ] Admin được tạo bằng cách có kiểm soát.
- [ ] Quyền khách/admin hoạt động ở cả UI và API.

---

## Giai đoạn 3 — Catalog, SKU và tồn kho (Tuần 3–4)

**Phạm vi:** FR-005 đến FR-013; US-02, US-03, US-06; TEST-21.

### Bước 3.1 — Chốt dữ liệu và API catalog

- [ ] Chốt trường của danh mục, sản phẩm, SKU và thuộc tính biến thể.
- [ ] Chốt quy tắc sản phẩm đơn giản vẫn có một SKU.
- [ ] Đặc tả giá bằng VND số nguyên lớn hơn 0; không có phí vận chuyển hoặc thuế tách riêng.
- [ ] Chốt API public cho danh sách/chi tiết, phân trang, lọc danh mục và tìm tên.
- [ ] Chốt API admin tạo/sửa/ngừng hoạt động và điều chỉnh tồn.

### Bước 3.2 — BE: migration catalog và inventory

- [ ] Tạo bảng category, product, SKU và quan hệ tương ứng.
- [ ] Đặt unique constraint cho mã SKU.
- [ ] Dùng trạng thái ngừng hoạt động thay vì xóa dữ liệu đã được tham chiếu.
- [ ] Tạo dữ liệu tồn theo SKU với `on_hand` và `reserved`.
- [ ] Tạo bảng lịch sử điều chỉnh gồm số trước/sau, lượng thay đổi, lý do, người làm và thời điểm.
- [ ] Thêm constraint/check phù hợp để ngăn số âm và dữ liệu sai cơ bản.

### Bước 3.3 — BE: API quản trị và API công khai

- [ ] Làm CRUD cần thiết cho danh mục, sản phẩm và SKU.
- [ ] Làm API nhập/điều chỉnh tồn trong transaction.
- [ ] Từ chối giảm `on_hand` xuống thấp hơn `reserved`.
- [ ] Làm API xem `on_hand`, `reserved` và `available` cho admin.
- [ ] API public chỉ trả dữ liệu đang hoạt động và dữ liệu tồn phù hợp để hiển thị; không coi số hiển thị là cam kết giữ hàng.
- [ ] Khi sửa tên/giá hiện tại, không làm thay đổi snapshot đơn trong tương lai.

### Bước 3.4 — FE: catalog và inventory chạy song song

- [ ] Làm trang khách xem danh sách, phân trang, lọc danh mục và tìm kiếm đơn giản.
- [ ] Làm trang chi tiết sản phẩm, chọn đúng SKU/biến thể.
- [ ] Làm màn hình admin quản lý danh mục, sản phẩm và SKU.
- [ ] Làm màn hình nhập/điều chỉnh tồn có lý do bắt buộc.
- [ ] Hiển thị riêng số hiện còn, đang giữ và khả dụng cho admin.
- [ ] Không cho UI xóa cứng hàng đã bán; dùng thao tác ngừng hoạt động.

### Bước 3.5 — Test và tích hợp

- [ ] Test SKU trùng, giá sai, dữ liệu bắt buộc thiếu và trạng thái ngừng bán.
- [ ] Test phân trang/lọc/tìm kiếm.
- [ ] Test điều chỉnh tồn âm hoặc dưới lượng giữ bị từ chối.
- [ ] Test lịch sử điều chỉnh có đủ tác nhân, lý do và số trước/sau.
- [ ] Kiểm tra các màn hình khách/admin bằng trình duyệt.

### Checkpoint Giai đoạn 3

- [ ] Admin tạo được catalog và nhập tồn.
- [ ] Khách xem đúng sản phẩm/SKU đang bán.
- [ ] Số liệu tồn tuân thủ INV-01, INV-02 và INV-12 ở phạm vi hiện có.

---

## Giai đoạn 4 — Chương trình flash sale (Tuần 4)

**Phạm vi:** FR-014 đến FR-018; US-04, US-05; TEST-20.

### Bước 4.1 — Chốt hợp đồng và trạng thái chương trình

- [ ] Chốt trạng thái `DRAFT`, `SCHEDULED`, `RUNNING`, `ENDED`, `STOPPED`.
- [ ] Chốt cách xác định trạng thái theo giờ database/server, không tin đồng hồ client.
- [ ] Chốt quy tắc một SKU không có hai khoảng nhận đơn chồng nhau.
- [ ] Đặc tả dữ liệu từng SKU: giá sale, quota và giới hạn mỗi tài khoản.
- [ ] Chốt hành vi hủy lịch, sửa nháp và dừng chương trình đang chạy.

### Bước 4.2 — BE: migration và API flash sale

- [ ] Tạo bảng campaign và campaign item.
- [ ] Lưu quota tổng, quota đang giữ và quota đã bán theo SKU/chương trình.
- [ ] Kiểm tra kết thúc sau bắt đầu, số lượng nguyên dương và giá sale thấp hơn giá thường lúc cấu hình.
- [ ] Chặn SKU trùng trong chương trình và khoảng chạy chồng nhau.
- [ ] Khóa giá/quota/giới hạn sau khi lên lịch.
- [ ] Dừng chương trình chỉ chặn đơn mới, không phá reservation cũ.
- [ ] API public trả mốc thời gian và nguồn giá; backend luôn kiểm tra lại khi checkout.

### Bước 4.3 — FE: quản trị và hiển thị sale chạy song song

- [ ] Làm form tạo/sửa chương trình nháp.
- [ ] Cho chọn SKU, nhập giá sale, quota và giới hạn mua.
- [ ] Hiển thị lỗi thời gian, giá, quota và chồng chương trình theo đúng dòng.
- [ ] Làm thao tác lên lịch, hủy lịch và dừng nhận đơn.
- [ ] Hiển thị giá sale, thời gian và đếm ngược cho khách; ghi rõ đếm ngược chỉ để hiển thị.
- [ ] Hiển thị thống kê lượng đang giữ, đã bán và bị từ chối cho admin ở mức P1.

### Bước 4.4 — Test và tích hợp

- [ ] Test chương trình chưa mở, đúng thời gian mở, đúng thời điểm kết thúc và đã dừng.
- [ ] Test không sửa trực tiếp campaign item sau khi lên lịch.
- [ ] Test cùng SKU không có thời gian chạy chồng nhau.
- [ ] Test dừng/kết thúc không giải phóng reservation hợp lệ của đơn cũ.
- [ ] Kiểm tra quản trị tạo chương trình và khách xem chương trình bằng trình duyệt.

### Checkpoint Giai đoạn 4

- [ ] US-04 và US-05 đạt.
- [ ] Catalog hiển thị đúng nguồn giá theo thời gian.
- [ ] Chưa dùng Redis để quyết định chương trình có được checkout hay không.

---

## Giai đoạn 5 — Giỏ hàng và xem trước đơn (Tuần 4–5)

**Phạm vi:** FR-019 đến FR-023; US-07, US-08.

### Bước 5.1 — Đặc tả hợp đồng giỏ và preview

- [ ] Ghi giới hạn tối đa 20 dòng SKU và giới hạn số lượng mỗi dòng theo cấu hình SKU/chương trình vào hợp đồng API.
- [ ] Nhận đúng tập dòng khách chọn để checkout, không tự lấy toàn bộ giỏ.
- [ ] Chốt `cart_version` hoặc dữ liệu tương đương để tránh xóa nhầm giỏ đã sửa.
- [ ] Chốt response preview gồm SKU, số lượng, đơn giá, nguồn giá, campaign, tiền tệ và tổng.
- [ ] Yêu cầu tên, số điện thoại và địa chỉ văn bản của người nhận cùng validation tương ứng.

### Bước 5.2 — BE: migration và API giỏ

- [ ] Tạo bảng cart và cart item theo chủ sở hữu.
- [ ] Đặt unique cho một SKU trong một giỏ; thêm cùng SKU phải gộp số lượng.
- [ ] Làm API thêm, sửa số lượng, xóa và xem giỏ.
- [ ] Kiểm tra quyền sở hữu ở mọi thao tác.
- [ ] Không cập nhật tồn hoặc quota khi sửa giỏ.
- [ ] Không dùng cache công khai cho giỏ.

### Bước 5.3 — BE: API xem trước

- [ ] Nhận tập dòng khách chọn và dữ liệu người nhận dự kiến.
- [ ] Đọc giá hiện hành từ PostgreSQL và tính tổng bằng số nguyên.
- [ ] Ghi rõ mỗi dòng dùng giá thường hay flash sale.
- [ ] Trả cảnh báo hàng ngừng bán, sale không hoạt động hoặc dữ liệu cần xác nhận lại.
- [ ] Không tạo order, reservation hoặc quyền mua trong preview.

### Bước 5.4 — FE: giỏ và preview chạy song song

- [ ] Làm thêm SKU từ trang sản phẩm vào giỏ.
- [ ] Làm sửa số lượng, xóa dòng và chọn dòng cần mua.
- [ ] Hiển thị giới hạn dòng/số lượng ngay trên UI nhưng vẫn dựa vào backend để xác nhận.
- [ ] Làm form người nhận và màn hình preview tổng tiền.
- [ ] Hiển thị nguồn giá từng dòng và cảnh báo “giỏ chưa giữ hàng”.
- [ ] Khi giá thay đổi, buộc người dùng xem và xác nhận lại.

### Bước 5.5 — Test và tích hợp

- [ ] Test SKU trùng được gộp.
- [ ] Test giỏ rỗng, số lượng sai, quá giới hạn và truy cập giỏ người khác.
- [ ] Test preview không thay đổi `reserved`, tồn hoặc quota.
- [ ] Test tổng tiền và nguồn giá đúng.
- [ ] Kiểm tra luồng chọn sản phẩm → giỏ → preview bằng trình duyệt.

### Checkpoint Giai đoạn 5

- [ ] US-07 và US-08 đạt.
- [ ] Giỏ nhiều SKU hoạt động nhưng chưa tạo reservation.
- [ ] FE đã sẵn sàng gọi API checkout theo hợp đồng ở giai đoạn sau.

---

## Giai đoạn 6 — Checkout nguyên tử và idempotency (Tuần 5–6)

**Phạm vi:** FR-024 đến FR-029; US-09, US-10, US-11; TEST-01 đến TEST-10.

### Bước 6.1 — Chốt transaction checkout trước khi viết code

- [ ] Liệt kê toàn bộ dữ liệu phải kiểm tra: chủ giỏ, SKU hoạt động, số lượng, giá xác nhận, campaign, quota, giới hạn mua, người nhận và tiền tệ.
- [ ] Quy định mọi checkout nhiều SKU khóa theo cùng một thứ tự SKU để giảm deadlock.
- [ ] Chốt điểm bắt đầu/kết thúc transaction; không gọi thanh toán trong transaction.
- [ ] Chốt cách ghi idempotency request, hash nội dung chuẩn hóa và kết quả.
- [ ] Chốt lỗi ưu tiên: dữ liệu/hàng hóa → chương trình/giá → giới hạn mua → thiếu số lượng.

### Bước 6.2 — BE: migration đơn, reservation và idempotency

- [ ] Tạo bảng order với owner, trạng thái, tổng, tiền tệ, hạn thanh toán và snapshot người nhận.
- [ ] Tạo order item chứa snapshot SKU, tên/thuộc tính, số lượng, đơn giá và nguồn giá.
- [ ] Tạo reservation theo order item, SKU, số lượng, hạn và campaign nếu có.
- [ ] Tạo bảng/bằng chứng idempotency gắn với tài khoản, loại thao tác, key, hash request và kết quả.
- [ ] Thêm unique constraint để một tài khoản + key chỉ có một ý định checkout.
- [ ] Thêm index phục vụ tra cứu đơn, idempotency và đơn tới hạn; chỉ thêm index có truy vấn cụ thể.

### Bước 6.3 — BE: triển khai checkout nguyên tử

- [ ] Chuẩn hóa và gộp SKU trước khi transaction.
- [ ] Trong transaction, đọc/khóa SKU theo thứ tự nhất quán.
- [ ] Kiểm tra lại giá từ dữ liệu có thẩm quyền; trả `PRICE_CHANGED` và giá mới nếu khác xác nhận.
- [ ] Kiểm tra đồng thời tồn vật lý, quota sale và giới hạn mua.
- [ ] Nếu bất kỳ dòng nào sai, rollback toàn bộ; không để order chờ hoặc reservation đã commit.
- [ ] Nếu hợp lệ, tăng `reserved`/`quota_reserved`, tạo đủ order item và reservation, rồi tạo order `PENDING_PAYMENT`.
- [ ] Lưu hạn thanh toán sau 5 phút trên đơn; retry không gia hạn.
- [ ] Trả lỗi theo dòng nhưng không che lỗi database/hạ tầng thành hết hàng.

### Bước 6.4 — BE: xử lý retry và kết quả chưa rõ

- [ ] Cùng account + key + nội dung trả kết quả đã ghi nhận.
- [ ] Cùng key nhưng nội dung khác trả `IDEMPOTENCY_CONFLICT`.
- [ ] Hai request cùng key chạy đồng thời chỉ tạo tối đa một order.
- [ ] Nếu request đầu đang xử lý, trả kết quả cũ hoặc `REQUEST_IN_PROGRESS` với hướng retry rõ.
- [ ] Có API tra cứu checkout theo key thuộc đúng người dùng.
- [ ] Lỗi hạ tầng trước commit không được lưu thành từ chối nghiệp vụ cuối cùng.

### Bước 6.5 — FE: checkout chạy song song

- [ ] Tạo idempotency key mới cho mỗi ý định checkout đã được người dùng xác nhận.
- [ ] Giữ nguyên key khi retry do timeout/mất phản hồi.
- [ ] Tạo key mới nếu người dùng sửa giỏ, giá xác nhận hoặc người nhận.
- [ ] Khóa nút gửi khi request đang chạy nhưng không coi đây là cơ chế chống trùng chính.
- [ ] Hiển thị lỗi theo SKU cho thiếu hàng, quota, giới hạn mua và giá thay đổi.
- [ ] Không tự bỏ dòng lỗi, không tự đổi sang giá thường và không tự tạo đơn một phần.
- [ ] Với kết quả chưa rõ, hiển thị đang kiểm tra và gọi tra cứu/retry cùng key.

### Bước 6.6 — Test checkout bắt buộc

- [ ] TEST-01: nhiều SKU đủ hàng tạo một order đủ reservation.
- [ ] TEST-02: SKU sau thiếu không để lại reservation ở SKU trước.
- [ ] TEST-03: nhiều khách tranh hàng cuối không bán/giữ vượt tồn.
- [ ] TEST-04: giỏ giao nhau, thứ tự SKU khác nhau không giữ một phần; retry deadlock hữu hạn.
- [ ] TEST-05: cùng key qua hai instance tạo tối đa một order.
- [ ] TEST-06: cùng key, nội dung khác bị xung đột.
- [ ] TEST-07: mất response sau commit vẫn tìm được order cũ.
- [ ] TEST-08: giá thay đổi bắt xác nhận lại.
- [ ] TEST-09: hết quota nhưng còn tồn vẫn từ chối toàn bộ giá sale.
- [ ] TEST-10: nhiều key cùng tài khoản không vượt giới hạn mua.
- [ ] Sau mỗi test, kiểm tra trực tiếp order, reservation, inventory và quota trong database.

### Checkpoint Giai đoạn 6

- [ ] INV-01 đến INV-05, INV-09 đến INV-12 không bị vi phạm.
- [ ] Luồng giỏ → preview → checkout chạy hoàn chỉnh bằng trình duyệt.
- [ ] Checkout chưa phụ thuộc Redis hoặc bộ nhớ cục bộ của một API instance.
- [ ] Đây là mốc đầu tiên có thể tạo baseline nghiệp vụ, chưa phải baseline hiệu năng cuối.

---

## Giai đoạn 7 — Thanh toán mô phỏng, hủy và hết hạn (Tuần 6–7)

**Phạm vi:** FR-030 đến FR-041; US-12 đến US-17; TEST-11 đến TEST-20, TEST-25.

### Bước 7.1 — Đặc tả hợp đồng thanh toán và vòng đời

- [ ] Đặc tả một payment session cho một order; retry trả phiên cũ.
- [ ] Chốt trạng thái payment `PENDING`, `SUCCEEDED`, `FAILED`, `UNKNOWN`.
- [ ] Chốt trạng thái order và các chuyển trạng thái hợp lệ.
- [ ] Chốt chữ ký/xác thực callback của bộ mô phỏng.
- [ ] Chốt loại reconciliation: đến muộn, sai số tiền/tiền tệ/phiên và sự kiện mâu thuẫn.
- [ ] Chốt rằng reconciliation chỉ ghi chú/đánh dấu, không hồi sinh đơn hoặc sửa tồn.

### Bước 7.2 — BE: migration payment và reconciliation

- [ ] Tạo payment session gắn duy nhất với order.
- [ ] Tạo payment event có event ID, transaction ID, payload cần kiểm tra, thời điểm nhận và kết quả xử lý.
- [ ] Tạo unique constraint chống xử lý lặp event/transaction theo quy tắc đã chốt.
- [ ] Tạo reconciliation record với loại lỗi, bằng chứng, trạng thái rà soát và ghi chú.
- [ ] Bảo đảm sự kiện được lưu bền vững trước khi trả xác nhận đã nhận.

### Bước 7.3 — BE: mở phiên và bộ mô phỏng

- [ ] Chỉ chủ đơn còn `PENDING_PAYMENT`, chưa hết hạn và còn đủ reservation được mở phiên.
- [ ] Retry mở phiên trả phiên cũ, không tạo phiên thứ hai và không gia hạn order.
- [ ] Phiên dùng đúng order ID, tổng tiền và tiền tệ snapshot.
- [ ] Bộ mô phỏng tạo được success, failure, delay, timeout, duplicate và out-of-order event.
- [ ] Endpoint điều khiển mô phỏng chỉ dành cho môi trường test/admin được bảo vệ.

### Bước 7.4 — BE: xử lý callback đúng một lần

- [ ] Xác minh nguồn, order, session/transaction, amount, currency và event ID.
- [ ] Success chỉ xác nhận order còn hợp lệ trước hạn và còn reservation.
- [ ] Trong cùng transaction, chuyển toàn bộ reservation thành bán, giảm `reserved` và `on_hand`, chuyển `quota_reserved` sang `quota_sold`, rồi đặt order `CONFIRMED`.
- [ ] Failure đã xác minh đóng order và giải phóng toàn bộ reservation đúng một lần.
- [ ] Timeout/unknown không tự đóng order hoặc giải phóng hàng.
- [ ] Callback lặp không trừ tồn hoặc trả hàng thêm.
- [ ] Failure sau success không đảo order; success muộn không hồi sinh order đã kết thúc.
- [ ] Sai amount/session/currency tạo reconciliation và không xác nhận order.

### Bước 7.5 — BE: hủy đơn và worker hết hạn

- [ ] Khách chỉ hủy order của mình khi còn chờ và chưa hết hạn.
- [ ] Hủy giải phóng toàn bộ reservation/quota đúng một lần.
- [ ] Worker lấy lô hữu hạn các order đến hạn và xử lý theo transaction ngắn.
- [ ] Nhiều worker có thể chạy mà không giải phóng cùng reservation hai lần.
- [ ] Worker restart tiếp tục xử lý việc còn lại từ dữ liệu bền vững.
- [ ] API không cho thanh toán order đã quá hạn dù worker chưa đổi trạng thái lưu trữ.
- [ ] Cạnh tranh callback/hủy/hết hạn chỉ có một kết quả thắng.

### Bước 7.6 — FE: thanh toán, trạng thái đơn và đối soát chạy song song

- [ ] Làm màn hình order chờ thanh toán với đồng hồ đếm ngược chỉ để hiển thị.
- [ ] Làm thao tác mở/lấy lại payment session.
- [ ] Làm màn hình điều khiển payment simulator cho môi trường demo được phép.
- [ ] Poll hoặc refresh trạng thái chính thức từ backend; không tin trang redirect.
- [ ] Làm danh sách/chi tiết order của khách có phân trang.
- [ ] Làm nút hủy chỉ khi backend cho phép.
- [ ] Làm màn hình admin tra cứu order và reconciliation.
- [ ] Không cung cấp nút ép success, sửa giá order hoặc sửa `reserved`.

### Bước 7.7 — Test vòng đời bắt buộc

- [ ] TEST-11 đến TEST-19 cho callback lặp, failure, timeout, cancel, expiry, cạnh tranh và worker restart.
- [ ] TEST-20 cho campaign kết thúc/dừng trong khi order còn hạn.
- [ ] TEST-25 cho việc sửa giỏ trong khi order cũ đang thanh toán.
- [ ] Test payment amount/currency/session sai.
- [ ] Sau mỗi ca, đối chiếu order, reservation, inventory, quota, payment event và reconciliation.
- [ ] Kiểm tra luồng thành công, thất bại, hết hạn và callback muộn bằng trình duyệt.

### Checkpoint Giai đoạn 7

- [ ] INV-06 đến INV-08 và INV-11 không bị vi phạm.
- [ ] Luồng mua đầu-cuối hoàn chỉnh.
- [ ] Restart API/worker không làm mất order, reservation hoặc event đã lưu.
- [ ] Bộ mô phỏng đủ dùng, chưa tích hợp cổng thanh toán thật.

---

## Giai đoạn 8 — Quan sát hệ thống và hoàn thiện demo (Tuần 7–8)

**Phạm vi:** FR-042 đến FR-046; US-18 đến US-20; TEST-24.

### Bước 8.1 — BE: log và request ID

- [ ] Sinh hoặc nhận request ID ở gateway/API và trả lại trong response.
- [ ] Log endpoint, kết quả, thời gian xử lý, instance và mã lỗi nghiệp vụ.
- [ ] Liên kết được request với order/payment event khi phù hợp.
- [ ] Không log password, session secret, token hoặc payload riêng tư không cần thiết.
- [ ] Không dùng user ID/order ID làm label metrics có số lượng giá trị không giới hạn.

### Bước 8.2 — BE: metrics tối thiểu

- [ ] Đếm request theo endpoint/status/result group.
- [ ] Đo latency checkout, transaction, query và thời gian chờ pool.
- [ ] Đếm order mới, replay idempotency, order confirmed, lỗi nghiệp vụ, lỗi hệ thống và rate limit.
- [ ] Theo dõi worker: order đến hạn, xử lý thành công, lỗi và độ trễ giải phóng.
- [ ] Theo dõi cache hit/miss sau khi cache được bật ở Giai đoạn 11.

### Bước 8.3 — FE: hoàn thiện hành trình demo

- [ ] Kiểm tra toàn bộ hành trình khách: catalog → SKU → cart → preview → checkout → payment → order.
- [ ] Kiểm tra toàn bộ hành trình admin: store → catalog → inventory → campaign → order → reconciliation.
- [ ] Bổ sung loading, empty state, lỗi mạng, lỗi theo dòng và thông báo retry.
- [ ] Bảo đảm tổng tiền, SKU, nguồn giá, hạn và trạng thái được hiển thị rõ.
- [ ] Kiểm tra thao tác bàn phím, label form và focus lỗi ở mức accessibility cơ bản.

### Bước 8.4 — Kiểm tra quyền và rò rỉ dữ liệu

- [ ] Chạy TEST-24 cho khách A truy cập dữ liệu khách B và khách gọi API admin.
- [ ] Kiểm tra cache không chứa cart/order/payment riêng tư.
- [ ] Kiểm tra response lỗi không có SQL, stack trace hoặc secret.
- [ ] Kiểm tra URL/log trên trình duyệt không làm lộ credential.

### Checkpoint Giai đoạn 8

- [ ] US-19 và US-20 demo được bằng trình duyệt.
- [ ] Có log/metrics đủ để phân tích benchmark.
- [ ] Các màn hình P0 hoàn chỉnh; màn hình P1 còn thiếu phải được ghi rõ.

---

## Giai đoạn 9 — Baseline đúng nghiệp vụ E0 (Tuần 8)

### Mục tiêu

Tạo phiên bản baseline đúng nghiệp vụ trước khi tối ưu. Baseline không được cố tình viết chậm hoặc bỏ tính năng.

### Bước 9.1 — Chuẩn bị dữ liệu và công cụ kiểm thử

- [ ] Viết seed tạo tài khoản, category, product, SKU, inventory và campaign xác định trước.
- [ ] Tạo lệnh reset riêng database benchmark về cùng trạng thái đầu.
- [ ] Tạo kịch bản k6 cho LOAD-01 đến LOAD-07 ở mức baseline.
- [ ] Ghi seed, phân bố SKU, số dòng/giỏ, số lượng/dòng và tỷ lệ đọc/ghi.
- [ ] Đặt generator tách khỏi API nếu tài nguyên cho phép và ghi rõ nơi chạy.

### Bước 9.2 — Chạy toàn bộ test đúng đắn

- [ ] Chạy TEST-01 đến TEST-25.
- [ ] Chạy test đồng thời nhiều lần để tìm lỗi không ổn định.
- [ ] Sau mỗi nhóm test, chạy truy vấn kiểm tra INV-01 đến INV-12.
- [ ] Dừng benchmark nếu có bất biến bị vi phạm.
- [ ] Lưu kết quả, commit/version và cấu hình dùng cho lần chạy.

### Bước 9.3 — Đo E0

- [ ] Chạy 1 API instance, chưa bật cache catalog.
- [ ] Warm-up riêng và không tính vào cửa sổ đo.
- [ ] Chạy LOAD-01 với đủ tồn để tìm điểm bão hòa.
- [ ] Chạy LOAD-02 cho SKU nóng và cart 2–5 dòng.
- [ ] Chạy LOAD-03 cho hàng/quota khan hiếm nhưng báo riêng kết quả.
- [ ] Chạy mỗi cấu hình ít nhất 3 lần; lưu median và độ biến thiên.
- [ ] Ghi target load, actual load, dropped iterations, p95/p99, lỗi, CPU/RAM, DB/pool và lock wait.

### Bước 9.4 — Xác định tải tham chiếu

- [ ] Chọn tải mà generator phát đủ và hệ thống còn đo được ổn định.
- [ ] Chốt cấu hình CPU/RAM, DB, pool, thời lượng và dữ liệu.
- [ ] Đánh giá các mục tiêu đã chọn: p95 checkout ≤ 500 ms, p99 ≤ 1.000 ms, lỗi hệ thống < 1% và 95% reservation được giải phóng trong 30 giây sau hạn.
- [ ] Không thay đổi workload giữa các phiên bản so sánh nếu không ghi rõ lý do.

### Checkpoint Giai đoạn 9

- [ ] E0 vượt toàn bộ test nghiệp vụ.
- [ ] Có dữ liệu gốc và lệnh chạy lại.
- [ ] Có tải/cấu hình tham chiếu được ghi thành văn bản.
- [ ] Có danh sách điểm nghẽn dựa trên số liệu, chưa tối ưu theo cảm tính.

---

## Giai đoạn 10 — Tối ưu transaction, query và pool: E1 (Tuần 8–9)

### Bước 10.1 — Phân tích E0

- [ ] Xác định thời gian nằm trong transaction checkout.
- [ ] Xác định thời gian chờ lock, deadlock/retry và SKU nóng.
- [ ] Xác định query chậm, query lặp và kế hoạch thực thi.
- [ ] Xác định thời gian chờ connection pool và tổng connection.
- [ ] Chỉ chọn thay đổi có vấn đề đo được.

### Bước 10.2 — Tối ưu nhỏ, từng thay đổi một

- [ ] Rút ngắn transaction nhưng giữ nguyên kiểm tra nghiệp vụ.
- [ ] Batch query các dòng giỏ/SKU khi phù hợp.
- [ ] Thêm hoặc sửa index dựa trên query plan.
- [ ] Giữ thứ tự khóa nhất quán ở mọi thao tác nhiều SKU liên quan.
- [ ] Cấu hình timeout và retry hữu hạn; không retry bằng idempotency key mới.
- [ ] Điều chỉnh pool theo ngân sách connection tổng, không tăng vô hạn.
- [ ] Sau mỗi thay đổi, chạy test liên quan và đo riêng tác động.

### Bước 10.3 — Đo E1 công bằng

- [ ] Dùng cùng dữ liệu, seed, workload và tài nguyên như E0.
- [ ] Chạy ít nhất 3 lần mỗi cấu hình.
- [ ] So sánh p95/p99, order mới/giây, lock wait, query, pool và lỗi.
- [ ] Kiểm tra lại INV-01 đến INV-12 sau tải.
- [ ] Giữ cả kết quả thay đổi không cải thiện; không bỏ số liệu xấu.

### Bước 10.4 — FE kiểm tra hồi quy song song

- [ ] Chạy lại luồng khách và admin trong khi BE tối ưu.
- [ ] Kiểm tra response và mã lỗi không thay đổi trái hợp đồng.
- [ ] Kiểm tra timeout/`REQUEST_IN_PROGRESS` vẫn có hướng xử lý rõ.
- [ ] Không thêm UI phức tạp nếu không phục vụ demo hoặc nghiệm thu.

### Checkpoint Giai đoạn 10

- [ ] E1 đúng nghiệp vụ như E0.
- [ ] Mỗi thay đổi có bằng chứng giữ/bỏ rõ ràng.
- [ ] Báo cáo phân biệt cải thiện do query, transaction và pool.

---

## Giai đoạn 11 — Nhiều instance, cache và kiểm soát tải: E2–E4 (Tuần 9–10)

### Bước 11.1 — E2: chạy nhiều API instance

- [ ] Cấu hình NGINX phân phối request tới ít nhất 2 API instance.
- [ ] Bảo đảm instance không giữ state nghiệp vụ cần chia sẻ trong bộ nhớ.
- [ ] Tính tổng connection pool của API và worker trước khi chạy.
- [ ] Test retry idempotency chuyển sang instance khác.
- [ ] So sánh 1 và 2 instance ở tải phân tán SKU và SKU nóng.
- [ ] Ghi rõ trường hợp cùng ngân sách CPU/RAM và trường hợp tăng thêm tài nguyên.
- [ ] Không kỳ vọng scaling tuyến tính khi database/lock là điểm nghẽn.

### Bước 11.2 — E3: cache catalog bằng Redis

- [ ] Chỉ cache danh sách/chi tiết catalog công khai.
- [ ] Cache key phải chứa đầy đủ filter, page và tham số ảnh hưởng response.
- [ ] Đặt TTL tôn trọng thời gian kết thúc flash sale.
- [ ] Invalidate/bỏ cache khi admin sửa catalog/campaign liên quan.
- [ ] Khi Redis lỗi, fallback database có giới hạn hoặc trả lỗi tạm thời; không dồn query vô hạn.
- [ ] Checkout luôn đọc lại giá, tồn và quota từ PostgreSQL.
- [ ] Chạy TEST-22 với cache cũ, Redis lỗi và nhiều cache miss cùng lúc.
- [ ] Đo hit/miss, latency đọc, tải DB và tác động đến checkout hỗn hợp.

### Bước 11.3 — E4: rate limit và hành vi quá tải

- [ ] Chọn nhóm endpoint và ngưỡng riêng cho đăng nhập, catalog, checkout, callback và admin.
- [ ] Từ chối trước transaction khi có thể; request bị giới hạn không được giữ hàng.
- [ ] Trả `RATE_LIMITED` và thông tin retry phù hợp.
- [ ] Bảo đảm callback/worker có ngân sách xử lý, không bị lưu lượng catalog làm đói tài nguyên.
- [ ] Ghi rõ giới hạn nằm ở một NGINX hay toàn hệ thống.
- [ ] Chạy TEST-23 và LOAD-09: tăng tải vượt năng lực rồi giảm về bình thường.
- [ ] Báo riêng rate-limited request, lỗi hệ thống, backlog và thời gian hồi phục.

### Bước 11.4 — Kiểm tra sự cố và khôi phục

- [ ] Chạy LOAD-08: restart API, restart worker và Redis lỗi.
- [ ] Xác nhận dữ liệu đã commit không mất.
- [ ] Xác nhận worker tiếp tục xử lý order hết hạn.
- [ ] Xác nhận checkout không dùng cache để quyết định.
- [ ] Xác nhận retry giữ nguyên idempotency key khi kết quả chưa rõ.

### Bước 11.5 — FE kiểm tra nhiều instance và cache song song

- [ ] Kiểm tra người dùng không nhận session khác nhau khi request đổi instance.
- [ ] Kiểm tra catalog cập nhật không hiển thị sai quá giới hạn TTL đã chốt.
- [ ] Hiển thị rate limit, service unavailable và retry rõ ràng.
- [ ] Không tự động retry vô hạn ở trình duyệt.

### Checkpoint Giai đoạn 11

- [ ] Có dữ liệu E2, E3 và E4 hoặc kết luận có căn cứ rằng một thay đổi không có lợi.
- [ ] TEST-22 và TEST-23 đạt.
- [ ] Tất cả bất biến vẫn đạt sau tải và sự cố.
- [ ] Báo cáo chỉ kết luận trong điều kiện đã đo.

---

## Giai đoạn 12 — Hoàn thiện sản phẩm, báo cáo và bàn giao (Tuần 10–12)

### Bước 12.1 — Rà soát chức năng theo ma trận

- [ ] Đối chiếu FR-001 đến FR-046 với tính năng thực tế.
- [ ] Đối chiếu US-01 đến US-20 với bằng chứng test/browser.
- [ ] Đối chiếu TEST-01 đến TEST-25 với kết quả mới nhất.
- [ ] Đối chiếu INV-01 đến INV-12 sau bộ test cuối.
- [ ] Mọi P0 phải hoàn thành; P1 thiếu phải được ghi rõ lý do và ảnh hưởng.

### Bước 12.2 — Rà soát bảo mật và độ bền

- [ ] Kiểm tra password hash, session/cookie, CSRF/CORS và giới hạn đăng nhập.
- [ ] Kiểm tra quyền sở hữu ở mọi API cart/order/payment.
- [ ] Kiểm tra admin API yêu cầu role đúng.
- [ ] Kiểm tra secret chỉ đi qua biến môi trường.
- [ ] Kiểm tra input, payload limit, pagination và timeout.
- [ ] Kiểm tra restart API/worker và migration trên database mới.
- [ ] Kiểm tra log/metrics không lộ dữ liệu nhạy cảm.

### Bước 12.3 — Hoàn thiện tài liệu kỹ thuật

- [ ] README hướng dẫn cài đặt, biến môi trường, migration, seed, chạy BE/FE/worker và Docker.
- [ ] Tài liệu API có endpoint, quyền, request, response, mã lỗi và ví dụ.
- [ ] `architecture.md` phản ánh hệ thống thật, không giữ sơ đồ cũ.
- [ ] Tài liệu giải thích inventory/reservation/quota và order/payment lifecycle.
- [ ] Tài liệu benchmark có cách reset dữ liệu và chạy từng workload.
- [ ] Ghi rõ giới hạn: không production-ready, không payment thật, không HA database và không SLA thương mại.

### Bước 12.4 — Hoàn thiện báo cáo hiệu năng

- [ ] Trình bày E0 đến E4 với cùng điều kiện so sánh hoặc ghi rõ khác biệt.
- [ ] Báo target load, actual load và dropped iterations.
- [ ] Báo riêng checkout request, order mới, replay, order confirmed, lỗi nghiệp vụ, rate limit và lỗi hệ thống.
- [ ] Báo p95/p99 riêng cho success, business rejection và system error.
- [ ] Báo CPU/RAM, DB connection, pool wait, lock wait, query và cache hit/miss.
- [ ] Giải thích giới hạn SKU nóng và lý do nhiều instance không luôn tăng tuyến tính.
- [ ] Đính kèm dữ liệu gốc, commit/version và cấu hình.

### Bước 12.5 — Kiểm tra demo cuối

- [ ] Tạo lại môi trường từ máy sạch theo README.
- [ ] Seed dữ liệu demo.
- [ ] Demo admin tạo catalog, nhập tồn và lên lịch sale.
- [ ] Demo khách đăng ký, thêm nhiều SKU, preview và checkout.
- [ ] Demo thiếu hàng hoặc giá đổi, không tạo order một phần.
- [ ] Demo payment success, failure, expiry và late callback/reconciliation.
- [ ] Demo 1 và 2 API instance cùng dùng dữ liệu chính xác.
- [ ] Chuẩn bị ảnh/video/bằng chứng dự phòng nếu mạng hoặc máy demo gặp lỗi.

### Bước 12.6 — Đóng gói bàn giao

- [ ] Mã nguồn build được từ repository.
- [ ] Không có `.env`, secret, dữ liệu cá nhân hoặc file benchmark quá lớn ngoài quy ước.
- [ ] Có migration, seed và lệnh reset môi trường test.
- [ ] Có frontend demo, backend, worker, NGINX, PostgreSQL và Redis config.
- [ ] Có bộ test và hướng dẫn chạy.
- [ ] Có PRD/spec/architecture/API docs và báo cáo cuối.
- [ ] Gắn tag hoặc ghi commit bàn giao để kết quả benchmark truy ngược được.

### Checkpoint hoàn thành dự án

- [ ] Tất cả chức năng P0 và luồng demo bắt buộc hoàn thành.
- [ ] Không phát hiện vi phạm INV-01 đến INV-12 trong bộ test công bố.
- [ ] TEST-01 đến TEST-25 có bằng chứng.
- [ ] Có so sánh ít nhất 1 và 2 API instance.
- [ ] Hệ thống dựng lại được từ hướng dẫn.
- [ ] Báo cáo nêu trung thực giới hạn và các mục chưa hoàn thành.
- [ ] Giảng viên/chủ sản phẩm duyệt kết quả cuối.

---

## 9. Lịch 12 tuần gợi ý

| Tuần | Trọng tâm | Kết quả chính |
|---|---|---|
| 1 | Giai đoạn 0 và bắt đầu Giai đoạn 1 | Hoàn thiện spec, API convention, kiến trúc và Docker/database |
| 2 | Giai đoạn 1–2 | Môi trường hoàn chỉnh, auth và cửa hàng |
| 3 | Giai đoạn 2–3 | Phân quyền, catalog, SKU và inventory |
| 4 | Giai đoạn 3–5 | Catalog hoàn chỉnh, flash sale, bắt đầu cart |
| 5 | Giai đoạn 5–6 | Cart/preview và transaction checkout |
| 6 | Giai đoạn 6–7 | Idempotency, order, payment simulator |
| 7 | Giai đoạn 7–8 | Hủy/hết hạn/reconciliation, hoàn thiện demo |
| 8 | Giai đoạn 8–10 | Observability, E0 và bắt đầu E1 |
| 9 | Giai đoạn 10–11 | E1, nhiều API instance và cache |
| 10 | Giai đoạn 11–12 | Rate limit, sự cố, chạy benchmark lặp |
| 11 | Giai đoạn 12 | Hoàn thiện test, UI, tài liệu và báo cáo |
| 12 | Giai đoạn 12 | Tổng duyệt, demo, sửa lỗi cuối và bàn giao |

## 10. Phân công hai thành viên

### Thành viên A — trọng tâm BE nghiệp vụ

- Auth, order, inventory, reservation, idempotency, payment và worker.
- Migration và test đồng thời cho các bất biến.
- Metrics backend và tối ưu transaction/query/pool.

### Thành viên B — trọng tâm FE và tính năng quản trị

- Catalog, campaign, cart và các màn hình khách/admin.
- API client, xử lý lỗi, kiểm tra trình duyệt và bằng chứng demo.
- Seed data, k6, tổng hợp dữ liệu benchmark và tài liệu sử dụng.

### Phần phải làm chung

- Chốt hợp đồng API trước mỗi tính năng.
- Review transaction liên quan tiền/tồn.
- Chạy test tích hợp và kiểm tra bất biến.
- Chạy benchmark, phân tích kết quả và viết báo cáo.
- Không để chỉ một người hiểu toàn bộ checkout hoặc cách dựng môi trường.

## 11. Ma trận truy vết yêu cầu

| Nhóm yêu cầu | Giai đoạn thực hiện chính | Giai đoạn kiểm chứng cuối |
|---|---|---|
| FR-001–FR-004, US-01 | 2 | 8, 12 |
| FR-005–FR-013, US-02/03/06 | 3 | 9, 12 |
| FR-014–FR-018, US-04/05 | 4 | 9, 12 |
| FR-019–FR-023, US-07/08 | 5 | 8, 12 |
| FR-024–FR-029, US-09/10/11 | 6 | 9–12 |
| FR-030–FR-041, US-12–US-17 | 7 | 9–12 |
| FR-042–FR-046, US-18–US-20 | 8–12 | 12 |
| TEST-01–TEST-10 | 6 | 9, 10, 11, 12 |
| TEST-11–TEST-21, TEST-25 | 7 | 9, 10, 11, 12 |
| TEST-22–TEST-23 | 11 | 12 |
| TEST-24 | 2 và 8 | 12 |
| INV-01–INV-12 | 3–7 | Sau mọi benchmark và ở Giai đoạn 12 |

## 12. Rủi ro và cách xử lý

| Rủi ro | Ảnh hưởng | Cách xử lý thực tế |
|---|---|---|
| Hai người hiểu API khác nhau | FE/BE phải sửa lại nhiều | Chốt hợp đồng API trước mỗi feature; thay đổi phải cập nhật tài liệu trước |
| Checkout nhiều SKU deadlock | p99 cao hoặc request lỗi | Khóa SKU theo thứ tự thống nhất, transaction ngắn, retry hữu hạn và test giỏ giao nhau |
| Bán vượt/quota vượt | Sai nghiệp vụ nghiêm trọng | PostgreSQL transaction + constraint/lock phù hợp; kiểm tra INV sau test tải |
| Retry tạo đơn trùng | Trừ/giữ hàng nhiều lần | Idempotency lưu trong DB dùng chung; test qua hai instance |
| Callback và worker tranh nhau | Vừa bán vừa trả hàng | Guard trạng thái và thời hạn trong cùng transaction; chỉ một chuyển trạng thái thắng |
| Cache làm checkout sai | Giá/tồn/quota không đúng | Cache chỉ dùng cho đọc catalog; checkout luôn xác minh PostgreSQL |
| Tăng instance làm DB hết connection | Hệ thống chậm hoặc lỗi | Tính pool tổng trước test; đo pool wait và không tăng pool tùy tiện |
| Hết hàng quá sớm khi benchmark | Kết quả throughput gây hiểu nhầm | Tách bài đủ tồn và bài khan hiếm; báo riêng giai đoạn sau hết hàng |
| Máy chạy k6 không phát đủ tải | Không biết giới hạn thật | Theo dõi actual load/dropped iterations; giảm tải hoặc tách generator |
| P1 làm trễ phần cốt lõi | Không hoàn thành checkout/báo cáo | Hoàn thành P0 và bất biến trước; cắt P1 phải ghi rõ và xin duyệt |
| Tài liệu lệch code | Không dựng lại hoặc demo sai | Cập nhật tài liệu trong Definition of Done của từng feature |

## 13. Những việc không làm trong kế hoạch này

- Không xây hệ thống đa tenant dùng chung một deployment.
- Không xây marketplace hoặc nhiều người bán.
- Không tích hợp thanh toán tiền thật, hoàn tiền thật hoặc nhiều cổng thanh toán.
- Không xây vận chuyển, nhiều kho, ERP/POS, hóa đơn điện tử hoặc thuế phức tạp.
- Không xây voucher, điểm thưởng, đánh giá, chat hoặc gợi ý cá nhân hóa.
- Không dùng Kubernetes, autoscaling cloud, microservice hoặc sharding.
- Không đưa toàn bộ checkout vào hàng đợi bất đồng bộ.
- Không tuyên bố production-ready hoặc SLA thương mại từ benchmark đồ án.

## 14. Cổng phê duyệt trước khi bắt đầu code

Nhóm chỉ bắt đầu Giai đoạn 1 sau khi:

- [ ] BE và FE thống nhất hợp đồng API nền tảng.
- [ ] Các thay đổi phạm vi mới được ghi vào PRD thay vì trao đổi miệng.
