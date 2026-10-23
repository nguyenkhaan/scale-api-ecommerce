# Chiến lược kiểm thử SaleScale

**Trạng thái:** Draft để duyệt trong phase SPECIFY
**Nguồn nghiệm thu:** [PRD](../product/prd.md) mục 12, 14 và 15

## Mục tiêu

Kiểm thử phải chứng minh hai điều độc lập:

1. Hệ thống giữ đúng các bất biến về tiền, tồn, quota, idempotency và trạng thái dưới cạnh tranh/lỗi.
2. Phiên bản đã đúng nghiệp vụ đạt số đo hiệu năng công bố trong điều kiện tái lập được.

Một HTTP status đúng chưa đủ để kết luận test tiền/tồn qua. Test bắt buộc truy vấn trạng thái bền vững sau khi tất cả goroutine/worker/event liên quan đã kết thúc.

## Các cấp kiểm thử

| Cấp | Phạm vi | Phụ thuộc | Dùng cho |
|---|---|---|---|
| Unit | Hàm thuần, validation, canonical hash, tính giá/tổng, state decision | Không I/O | Nhánh quy tắc và mã lỗi |
| Repository integration | SQL, constraint, index/query và transaction | PostgreSQL thật cùng major version mục tiêu | Inventory, idempotency, order/payment transition |
| Service concurrency | Nhiều goroutine/process gọi application service | PostgreSQL; clock kiểm soát được | Race checkout/callback/cancel/expiry |
| HTTP contract | Router, middleware, cookie, ownership, payload/error | API + PostgreSQL test | OpenAPI và phân quyền |
| Component fault | Restart/disable dependency và worker | Docker Compose lab | Redis/RabbitMQ/worker/API recovery |
| Browser acceptance | Hành trình khách và quản trị | Full stack | FR-044, FR-045 |
| Load/benchmark | Workload k6 cố định | Lab được pin cấu hình | LOAD-01 đến LOAD-10, E0–E4 |

Unit test không mock lại toàn bộ SQL transaction. Các bất biến dựa trên row lock/unique/check constraint phải được chứng minh bằng PostgreSQL thật.

## Công cụ và vị trí

| Loại | Công cụ/đường dẫn dự kiến |
|---|---|
| Go unit/integration | `go test`; test cạnh package hoặc `tests/integration/` |
| HTTP | `net/http/httptest` hoặc server test thật khi cần cookie/concurrency |
| PostgreSQL | Container database riêng cho test; migration giống application |
| Contract | Parse/validate `docs/technical/api/openapi.yaml`, kiểm response/error mẫu |
| Frontend | Framework theo Vite setup; browser acceptance ở `tests/e2e/` |
| Load | k6 scripts ở `tests/load/` |
| Oracle | SQL ở `tests/verification/` hoặc helper integration có version |

Không chốt thư viện assertion/migration/container trước TECH-01. Dependency được chọn phải hỗ trợ chạy local và CI bằng một lệnh rõ ràng.

## Kiểm soát thời gian và tính quyết định

- Application service nhận `Clock` cho unit test, nhưng quyết định hết hạn trong transaction dùng database time.
- Test không phụ thuộc `sleep` dài để tạo race. Dùng barrier/channel để đồng bộ điểm bắt đầu; delay simulator là dữ liệu test có giới hạn.
- UUID/idempotency key/seed được tạo xác định và ghi trong output khi test lỗi.
- Mỗi test dùng database/schema riêng hoặc transaction cleanup có guard; test song song không dùng chung SKU trừ khi đó là mục tiêu.
- Không đổi đồng hồ hệ thống của máy chạy test.

## Oracle bất biến

Sau test concurrency, chạy tối thiểu các kiểm tra:

| Bất biến | Oracle dữ liệu |
|---|---|
| INV-01 | Mọi inventory có `0 <= reserved <= on_hand`; available không âm |
| INV-02 | `inventory.reserved = SUM(reservation.held)` theo SKU |
| INV-03 | Mọi order pending có đủ một reservation held cho mỗi dòng |
| INV-04 | Request bị từ chối không tạo order/reservation/counter delta ngoài idempotency result |
| INV-05 | Unique `(user, operation, key)` và tối đa một order |
| INV-06 | Reservation chỉ có một terminal state và một `finalized_at` |
| INV-07 | Order confirmed có verified success event; counter sold chỉ tăng một lần |
| INV-08 | Order failed/cancelled/expired không còn reservation held |
| INV-09 | `quota_reserved + quota_sold <= quota_total` và các counter không âm |
| INV-10 | Usage reserved + sold không vượt purchase limit |
| INV-11 | Order total bằng tổng snapshot; payment amount/currency bằng order |
| INV-12 | Initial + adjustments − confirmed sold bằng `on_hand` hiện tại |

Oracle chạy trong trạng thái hệ thống đã quiesce. Với worker/outbox, ghi rõ timeout chờ và báo lỗi nếu backlog chưa về điều kiện mong đợi.

## Ma trận TEST bắt buộc

| Test | Cấp chính | Thiết lập/cạnh tranh | Oracle quyết định |
|---|---|---|---|
| TEST-01 | Integration | Giỏ nhiều SKU đủ tồn | Một order; đủ reservation; INV-03, INV-11 |
| TEST-02 | Integration | SKU sau thiếu | Không order pending/reservation/counter delta; INV-04 |
| TEST-03 | Concurrency | N request tranh lượng cuối | Tổng held/sold không vượt on_hand; INV-01, INV-12 |
| TEST-04 | Concurrency | Giỏ giao nhau, thứ tự input đảo | Không partial; retry deadlock hữu hạn; INV-03, INV-04 |
| TEST-05 | HTTP concurrency | Cùng key qua hai API instance | Một idempotency row/order; replay cùng response |
| TEST-06 | HTTP | Cùng key, hash khác | `IDEMPOTENCY_CONFLICT`; không order thứ hai |
| TEST-07 | HTTP/integration | Commit rồi làm mất response | Retry/lookup trả order cũ; expiry không gia hạn |
| TEST-08 | HTTP | Đổi giá sau preview | `PRICE_CHANGED`; không reservation; total mới rõ ràng |
| TEST-09 | Integration | Quota hết, inventory còn | Toàn request bị từ chối; INV-04, INV-09 |
| TEST-10 | Concurrency | Một user, nhiều key, cùng sale item | Usage không vượt limit; INV-10 |
| TEST-11 | Concurrency | Success event lặp/cùng giao dịch | Một transition confirmed; sold/inventory đổi một lần |
| TEST-12 | Concurrency | Failure event lặp | Release đúng một lần; INV-06, INV-08 |
| TEST-13 | Integration | Không callback/unknown | Order không failed sớm; expiry xử lý đúng |
| TEST-14 | HTTP concurrency | Hủy lặp | Một transition; counter chỉ release một lần |
| TEST-15 | Concurrency | Success đối đầu cancel/expiry | Đúng một terminal result; không vừa sold vừa released |
| TEST-16 | Integration | Success sau expired/cancelled | Event lưu + reconciliation; order/inventory không hồi sinh |
| TEST-17 | HTTP/integration | Sai signature/session/amount/currency | Event rejected bền vững; order không confirmed |
| TEST-18 | Integration | Failure sau confirmed | Event conflict; confirmed giữ nguyên |
| TEST-19 | Component concurrency | Nhiều expiry worker/restart | Không sót tới hạn; không release lặp; backlog phục hồi |
| TEST-20 | Integration | Sale end/stop khi order còn hạn | Không nhận order mới; order cũ vẫn payable trước hạn |
| TEST-21 | HTTP/integration | Admin giảm on_hand dưới reserved | Conflict; ledger/counter không sai |
| TEST-22 | Component fault | Redis down/stale/simultaneous miss | Checkout không đọc cache; fallback có giới hạn |
| TEST-23 | Component/load | Rate limit + retry | Request rejected không giữ hàng; same-key retry an toàn |
| TEST-24 | HTTP | User A đọc B/admin route | 403/404 theo contract; không lộ/thay đổi dữ liệu |
| TEST-25 | Integration | Sửa cart sau khi tạo order | Snapshot giữ nguyên; success không xóa dòng version mới |
| TEST-26 | P1 integration | Order event → inbox | Chỉ owner đọc; unique event ID |
| TEST-27 | P1 component fault | RabbitMQ down sau order commit | Order/outbox còn; phục hồi gửi bù |
| TEST-28 | P1 component fault | Consumer commit rồi chết trước ACK | Redelivery nhưng một notification |
| TEST-29 | P1 component fault | Poison message | Retry hữu hạn; vào DLQ; tra cứu được |

TEST-01 đến TEST-25 là cổng P0. TEST-26 đến TEST-29 là cổng notification P1 và không chặn việc hoàn thành đặc tả/triển khai lát cắt P0.

## Test API contract

Mỗi operation P0 trong OpenAPI cần kiểm:

- request hợp lệ và response schema;
- unknown field/invalid format/limit;
- thiếu session, sai role và sai owner;
- error `code`, `status`, `request_id`, `line_errors` khi liên quan;
- cookie flags ở môi trường HTTPS;
- `Idempotency-Key` bắt buộc cho `POST /v1/orders`;
- callback chỉ trả accepted sau khi event đã commit.

Không snapshot toàn bộ câu chữ `detail`; chỉ mã và cấu trúc là contract ổn định.

## Test bảo mật tối thiểu

- Password không xuất hiện trong log/response và chỉ lưu hash.
- Session token rõ không xuất hiện trong database; logout/revoke/expiry có hiệu lực qua instance khác.
- CSRF áp dụng cho mutation dùng cookie; callback/operator endpoint dùng cơ chế auth riêng.
- User không điều khiển role, user ID sở hữu order, unit price, total hoặc payment status.
- Pagination/filter không tạo SQL injection; payload/body/line count bị giới hạn.
- Metrics không dùng user/order/idempotency ID làm label.

## Benchmark

Chỉ benchmark commit/version đã qua cổng P0. Mỗi run lưu:

- commit, migration version, image/tool version;
- CPU/RAM, pool, instance count, gateway/DB/Redis/RabbitMQ config;
- seed, workload, duration, warm-up và cache state;
- tải mục tiêu/thực phát/dropped iterations;
- p95/p99 tách theo outcome; đơn mới/replay/confirmed/rejection/error;
- oracle INV sau run và raw result.

Chạy tối thiểu ba lần mỗi cấu hình và báo trung vị cùng độ biến thiên. Không dùng request hết hàng nhanh để suy ra checkout thành công nhanh hơn.

## Cổng chất lượng

### Trước khi merge thay đổi P0

- Formatter/static checks sạch.
- Unit + integration + HTTP tests liên quan qua.
- Nếu chạm tiền/tồn/state: chạy test concurrency và oracle INV liên quan.
- OpenAPI và traceability được cập nhật nếu contract đổi.

### Trước baseline E0

- TEST-01 đến TEST-25 qua trong lab.
- Không có vi phạm INV-01 đến INV-12.
- Seed/reset và oracle chạy lại được.
- Metrics phân biệt business rejection, rate limit, timeout và system error.

### Trước nghiệm thu cuối

- TEST-01 đến TEST-29 qua hoặc phần P1 bị cắt phải được ghi thành thay đổi phạm vi đã duyệt.
- Browser acceptance cho hành trình khách/quản trị qua.
- Benchmark và cấu hình tham chiếu được lưu; kết luận giới hạn trong điều kiện đã đo.

## Lệnh mục tiêu

```bash
cd app/backend && go test ./... -race -cover
docker compose run --rm integration-test
k6 run tests/load/<scenario>.js
```

Đây là hợp đồng lệnh cần được hiện thực trong phase triển khai; repository hiện chưa có test harness/Docker Compose hoàn chỉnh.

