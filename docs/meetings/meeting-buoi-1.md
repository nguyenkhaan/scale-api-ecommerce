# Hướng dẫn trình bày buổi họp 1 — SaleScale

**Ngày:** 23/10/2026  
**Thời lượng gợi ý:** 60 phút  
**Mục tiêu:** Chốt đặc tả trước khi chia task và bắt đầu triển khai.

## Giả định cho buổi họp

- Đây là buổi họp nội bộ đầu tiên của nhóm 2 người.
- Trọng tâm là duyệt phạm vi, nghiệp vụ, kiến trúc và kế hoạch; chưa review code chi tiết.
- PRD là nguồn nghiệp vụ chính. Tài liệu trong `research/` chỉ dùng để hiểu bối cảnh.

**Tài liệu:** [Quyền tài liệu](../decisions/003-document-authority.md), [Mục lục tài liệu](../CONTENT.md)

## 1. Mở đầu: dự án giải quyết vấn đề gì? — 5 phút

### Cần trình bày

- SaleScale là backend flash sale cho một cửa hàng trên mỗi bản triển khai.
- Bài toán chính: nhiều khách cùng đặt giỏ có nhiều SKU nhưng không được bán vượt tồn, tạo đơn trùng hoặc giữ một phần giỏ.
- Nhóm có 2 người, thời gian dự kiến 12 tuần.
- Kết quả cần có: backend Go/Gin, giao diện React/Vite tối thiểu, kiểm thử nghiệp vụ và báo cáo hiệu năng.

### Cần chốt

- Cả nhóm có cùng hiểu biết về mục tiêu và tiêu chí thành công hay chưa.
- Tính đúng đắn của đơn, tồn kho và thanh toán được ưu tiên trước tốc độ.

**Tài liệu:** [PRD — mục 1–2](../product/prd.md), [Project spec — Mục tiêu](../spec.md), [README dự án](../../README.md)

## 2. Phạm vi MVP — 8 phút

### Cần trình bày

- **P0:** tài khoản, catalog, tồn kho, flash sale, giỏ hàng, checkout nhiều SKU, idempotency, payment simulator và expiry worker.
- **P1 sản phẩm:** hộp thông báo qua outbox/RabbitMQ và phần quản trị bổ sung.
- **Sau P0:** cache Redis, rate limit, benchmark và triển khai cloud.
- Checkout phải đồng bộ và theo nguyên tắc: giữ đủ toàn bộ giỏ hoặc không giữ gì.
- Thanh toán chỉ là mô phỏng; không dùng tiền thật.
- AKS là môi trường cloud mục tiêu; Docker Compose là môi trường local và benchmark.
- Các phần ngoài phạm vi: marketplace, thanh toán thật, nhiều kho, vận chuyển, microservices và tự động scale.

### Cần chốt

- Danh sách P0/P1 có phù hợp với 12 tuần và 2 người không.
- Có chức năng nào cần bỏ, chuyển phase hoặc bổ sung không.
- Không mở rộng phạm vi nếu chưa đánh giá ảnh hưởng đến thời gian và kiểm thử.

**Tài liệu:** [PRD — mục 4](../product/prd.md), [Project spec — Phạm vi](../spec.md), [Danh sách chức năng BE](../product/feature.md), [Danh sách màn hình FE](../product/screen_list.md)

## 3. Luồng nghiệp vụ quan trọng — 10 phút

### Cần trình bày

- Giỏ hàng và bước xem trước không giữ tồn kho.
- Khi checkout, backend kiểm tra lại giá, tồn, quota và giới hạn mua trong một transaction.
- Mỗi ý định checkout dùng một idempotency key; retry cùng ý định không được tạo đơn mới.
- Đơn chờ thanh toán giữ hàng trong 5 phút; retry không gia hạn thời gian.
- Thành công thanh toán, thất bại, hủy và hết hạn cùng cạnh tranh để kết thúc đơn đúng một lần.
- Callback đến muộn không được hồi sinh đơn đã hết hạn hoặc đã hủy; trường hợp này chuyển sang đối soát.

### Cần chốt

- Các trạng thái của flash sale, order, reservation và payment đã đủ rõ chưa.
- Cả nhóm thống nhất cách xử lý mất phản hồi sau checkout và callback thanh toán đến muộn.
- Các bất biến `INV-01` đến `INV-12` là điều kiện bắt buộc, không được nới để tăng tốc.

**Tài liệu:** [PRD — mục 5–8](../product/prd.md), [Mô hình miền](../technical/domain-model.md), [Kiến trúc — luồng checkout](../technical/architecture.md), [User stories](../product/user-story.md)

## 4. Kiến trúc và trách nhiệm từng thành phần — 8 phút

### Cần trình bày

- Backend dùng modular monolith Go/Gin; frontend dùng React/Vite.
- PostgreSQL là nơi quyết định session, giá, tồn kho, quota, đơn, idempotency và thanh toán.
- API và worker không giữ trạng thái nghiệp vụ quan trọng chỉ trong RAM.
- Redis chỉ cache dữ liệu catalog công khai; checkout luôn kiểm tra PostgreSQL.
- RabbitMQ chỉ vận chuyển sự kiện thông báo; broker lỗi không được làm mất đơn đã lưu.
- Expiry worker, outbox publisher và notification worker là các tiến trình riêng.

### Cần chốt

- Chấp nhận PostgreSQL là nguồn dữ liệu có thẩm quyền.
- Chấp nhận modular monolith cho MVP, chưa tách microservices.
- Phân biệt rõ môi trường Compose và AKS khi demo hoặc báo benchmark.

**Tài liệu:** [Kiến trúc](../technical/architecture.md), [ADR-004 — PostgreSQL](../decisions/004-postgresql-authoritative-state.md), [ADR-006 — AKS và Compose](../decisions/006-aks-cloud-target.md), [Project spec — Tech stack](../spec.md)

## 5. Dữ liệu, API và bảo mật — 7 phút

### Cần trình bày

- Schema P0 bao gồm identity/session, catalog/inventory, flash sale, cart, order/reservation/idempotency và payment.
- Các dòng tồn kho và quota được khóa theo thứ tự ổn định để giảm deadlock.
- API đã có hợp đồng cho luồng khách, quản trị, checkout, payment và notification.
- Xác thực dùng opaque session cookie; database chỉ lưu hash của token.
- Mutation dùng cookie cần CSRF protection; callback payment dùng cơ chế xác thực riêng.
- Client không được tự quyết định role, giá, tổng tiền hoặc trạng thái thanh toán.

### Cần chốt

- Duyệt mô hình dữ liệu trước khi tạo migration.
- Duyệt API P0 và cấu trúc lỗi trước khi frontend/backend triển khai song song.
- Chọn công cụ migration, PostgreSQL driver/query layer và middleware cụ thể (`TECH-01`).

**Tài liệu:** [Mô hình dữ liệu](../technical/data-model.md), [OpenAPI](../technical/api/openapi.yaml), [ADR-005 — Session](../decisions/005-postgresql-sessions.md), [Project spec — Câu hỏi mở](../spec.md)

## 6. Kiểm thử và đo hiệu năng — 7 phút

### Cần trình bày

- Test phải kiểm tra dữ liệu trong PostgreSQL, không chỉ kiểm tra HTTP status.
- `TEST-01` đến `TEST-25` là cổng P0; `TEST-26` đến `TEST-29` kiểm tra notification P1.
- Phải kiểm tra các tình huống cạnh tranh: hết hàng, retry, callback lặp, hủy và expiry chạy cùng lúc.
- Chỉ benchmark phiên bản đã qua kiểm thử P0 và không vi phạm `INV-01` đến `INV-12`.
- So sánh E0–E4: baseline, outbox, tối ưu backend/nhiều instance, cache và kiểm soát tải.
- Mục tiêu hiện tại: p95 checkout không quá 500 ms, p99 không quá 1.000 ms và lỗi hệ thống dưới 1% tại tải tham chiếu được công bố.

### Cần chốt

- Thống nhất cổng chất lượng trước merge, trước E0 và trước nghiệm thu.
- Thống nhất cách lưu cấu hình, dữ liệu gốc và kết quả của từng lần benchmark.
- Chưa chốt RPS trước khi có baseline E0.

**Tài liệu:** [Chiến lược kiểm thử](../technical/testing.md), [Ma trận truy vết](../technical/traceability.md), [PRD — mục 14–15](../product/prd.md)

## 7. Kế hoạch 12 tuần và phân công — 7 phút

### Cần trình bày

- Thứ tự thực hiện: nền tảng → tài khoản → catalog/tồn → flash sale/giỏ → checkout → payment/expiry → P1 → hiệu năng → bàn giao.
- FE và BE có thể làm song song sau khi thống nhất payload OpenAPI.
- Gợi ý phân công:
  - Người 1: catalog, flash sale, giỏ hàng và giao diện demo.
  - Người 2: order, inventory, payment và worker.
  - Cả hai: review nghiệp vụ đồng thời, tích hợp và benchmark.
- Mỗi phase phải có checkpoint chạy được và bằng chứng kiểm thử.

### Cần chốt

- Người phụ trách chính và người review cho từng phase.
- Mốc tích hợp chung của FE/BE.
- Phase nào có thể chạy song song và phase nào phải làm tuần tự.

**Tài liệu:** [Kế hoạch triển khai](../plans/plan.md), [Các phase chi tiết](../plans/phases/phase-00-foundation.md), [PRD — mục 16](../product/prd.md)

## 8. Trạng thái hiện tại và bước tiếp theo — 4 phút

### Cần trình bày

- Theo tài liệu, dự án đang ở phase `SPECIFY`; PRD, spec và tài liệu kỹ thuật vẫn là bản draft cần duyệt.
- Kế hoạch đã có nhưng đang ở trạng thái đề xuất; chưa nên coi checklist phase là đã hoàn thành.
- Một số lệnh, Docker Compose, frontend, migration, test harness và benchmark vẫn cần hoàn thiện trong phase triển khai.
- Trước khi báo tiến độ code, đối chiếu lại workspace và chỉ ghi nhận phần đã chạy hoặc đã có test; tài liệu trạng thái có thể chưa phản ánh thay đổi chưa commit.

### Cần chốt

- Buổi họp có duyệt được phase `SPECIFY` hay cần một vòng sửa tài liệu.
- Nếu được duyệt, bước kế tiếp là duyệt `PLAN`, sau đó mới tách task nhỏ để triển khai.
- Chọn ngày checkpoint tiếp theo và đầu ra phải có ở checkpoint đó.

**Tài liệu:** [Trạng thái bộ tài liệu](../README.md), [Project spec — Tiêu chí thành công](../spec.md), [Kế hoạch triển khai](../plans/plan.md)

## 9. Các quyết định cần ghi lại cuối buổi — 4 phút

| Nội dung | Cần quyết định trong buổi 1? | Kết quả |
|---|---|---|
| Duyệt mục tiêu và phạm vi P0/P1 | Có | Chưa ghi |
| Duyệt quy tắc checkout, reservation và payment | Có | Chưa ghi |
| Duyệt kiến trúc và PostgreSQL là nguồn dữ liệu chính | Có | Chưa ghi |
| Duyệt hợp đồng API P0 | Có | Chưa ghi |
| Chọn migration tool, driver/query layer và middleware (`TECH-01`) | Có, trước phase 00 | Chưa ghi |
| Chọn cách ký/xác thực callback simulator (`TECH-02`) | Có thể để trước phase payment | Chưa ghi |
| Chọn tải/cấu hình tham chiếu (`OPEN-01`) | Không; chốt sau E0 | Chưa ghi |
| Chọn cách xử lý nghiệp vụ sau đối soát muộn (`OPEN-02`) | Trước nghiệm thu payment | Chưa ghi |
| Chọn mục tiêu thời gian giao notification (`OPEN-03`) | Không; chốt sau E0/E1 | Chưa ghi |
| Phân công và checkpoint tiếp theo | Có | Chưa ghi |

**Tài liệu:** [Project spec — Câu hỏi còn mở](../spec.md), [PRD — mục 17](../product/prd.md), [Kế hoạch — cổng SDD](../plans/plan.md)

## 10. Lưu ý khi trình bày

- Không dùng số liệu hiệu năng trong tài liệu research như kết quả đã đạt.
- Dùng quyết định hiện hành: giữ hàng **5 phút**, session cookie lưu ở PostgreSQL, payment simulator và AKS là cloud target.
- Không đi sâu từng bảng hoặc từng endpoint nếu nội dung đó không cần quyết định trong buổi họp.
- Khi có ý kiến thay đổi, ghi rõ phần PRD, bất biến, API và test nào bị ảnh hưởng.

**Tài liệu:** [PRD — mục 17.4, hợp nhất research](../product/prd.md), [ADR-003 — Quyền tài liệu](../decisions/003-document-authority.md)

## Kết quả buổi họp cần có

- [ ] Phạm vi P0/P1 và ngoài phạm vi được xác nhận.
- [ ] Luồng checkout, payment và expiry được hiểu giống nhau.
- [ ] PRD/spec/API P0 được duyệt hoặc có danh sách điểm phải sửa và người phụ trách.
- [ ] `TECH-01` có quyết định hoặc có hạn chốt rõ ràng.
- [ ] Kế hoạch phase, phân công và checkpoint tiếp theo được thống nhất.
- [ ] Mọi quyết định mới được cập nhật vào PRD/spec/ADR tương ứng trước khi code.
