# Thiết kế triển khai Phase 04 — Checkout và tích hợp Phase 05

**Trạng thái:** Sẵn sàng để AI/nhóm triển khai  
**Cập nhật:** 25/11/2026  
**Phạm vi:** Hoàn tất Phase 04 để mở khóa luồng P0 từ giỏ hàng đến thanh toán mô phỏng của Phase 05.

## 1. Mục đích và nguồn quyết định

Tài liệu này là handoff kỹ thuật cho người hoặc AI triển khai. Nó biến phần còn thiếu của Phase 04 thành các yêu cầu thực hiện được, đồng thời xác định chính xác hợp đồng đầu vào mà Phase 05 đang chờ. Không thay thế nguồn quyết định nghiệp vụ:

- [PRD](../product/prd.md) là nguồn sự thật cho phạm vi, BR, FR, INV và TEST theo [ADR-003](../decisions/003-document-authority.md).
- [OpenAPI](../technical/api/openapi.yaml) là hợp đồng HTTP.
- [Phase 04](../plans/phases/phase-04-checkout.md), [thiết kế checkout hiện có](../superpowers/specs/2026-11-15-phase-04-checkout-design.md), [Phase 05](../plans/phases/phase-05-payment-lifecycle.md) và [data model](../technical/data-model.md) là các tài liệu chi tiết cần đọc kèm.

Mục tiêu cuối của lát cắt này là: khách chọn các dòng giỏ, xác nhận báo giá, đặt một đơn giữ hàng nguyên tử, rồi đi vào các trang đơn/thanh toán mô phỏng hiện có. Không có thanh toán tiền thật hoặc tích hợp provider bên thứ ba; chỉ dùng payment simulator nội bộ.

## 2. Hiện trạng và kết quả cần đạt

Phase 05 Step 1–2 đã có trong nhánh `dev`: payment session, callback simulator có chữ ký, `order.Finish`, hủy đơn, expiry worker, danh sách/chi tiết đơn và trang thanh toán. Các phần đó hiện được kiểm thử bằng đơn pending tạo trực tiếp trong SQL.

Phase 04 vẫn chưa có đường tạo đơn thật. Migration `00009_create_orders.sql` và các model nền đã tồn tại; migration `00010_create_payment_lifecycle.sql` đã thêm cột cần cho Phase 05. Chưa có `POST /orders`, lookup idempotency, transaction checkout, hoặc UI đặt đơn từ giỏ.

Sau khi hoàn thành, hành trình bắt buộc là:

```text
Cart đã chọn
  -> POST /checkout/preview, khách xác nhận báo giá
  -> POST /orders (Idempotency-Key)
  -> Order pending_payment + OrderItems snapshot + Reservations active
  -> /orders/:id (xem đơn) -> /orders/:id/payment
  -> Phase 05 tạo/lấy payment session và xử lý callback/hủy/hết hạn
  -> order.Finish kết thúc reservation đúng một lần
```

Không gọi payment simulator, callback, cache hoặc broker trong transaction checkout.

## 3. Mục tiêu, ngoài phạm vi và điều kiện hoàn thành

### Mục tiêu Phase 04

1. Tạo đơn nhiều SKU theo nguyên tắc **giữ đủ tất cả hoặc không tạo gì**.
2. Server tính lại toàn bộ giá, tồn kho, quota và giới hạn mua; không tin giá/tổng do client gửi.
3. Cùng một ý định checkout chỉ có một kết quả bền vững, kể cả khi client timeout hoặc gửi lại request.
4. Trả trạng thái giao diện trung thực: kết quả chưa rõ không được hiển thị như đặt đơn thành công.
5. Tạo dữ liệu hoàn toàn tương thích với Phase 05 để callback, cancel và expiry worker có thể xử lý ngay.

### Ngoài phạm vi

- Không tạo hay gọi cổng thanh toán thật/provider sandbox.
- Không xác nhận thanh toán, hủy đơn, expiry worker, notification hay reconciliation mới; đó là trách nhiệm Phase 05/P1 đã có.
- Không xóa dòng giỏ khi checkout thành công. Chỉ `order.Finish(..., confirmed)` của Phase 05 được xóa dòng giỏ chưa thay đổi.
- Không dùng Redis làm nguồn quyết định checkout và không đưa checkout vào queue.

### Cổng hoàn thành

- `TEST-01` đến `TEST-10` và `TEST-25` qua cùng oracle `INV-01` đến `INV-05`, `INV-09` đến `INV-11`.
- Hành trình browser/API từ cart tới `/orders/:id/payment` chạy được với một đơn tạo qua API, không qua seed SQL.
- Sau khi ghép, toàn bộ cổng P0 `TEST-01` đến `TEST-25` và `INV-01` đến `INV-12` được chạy lại trước baseline E0.

## 4. Hợp đồng đầu vào của checkout

### HTTP

Triển khai đúng hai operation đã có trong OpenAPI:

| Operation | Trách nhiệm |
| --- | --- |
| `POST /orders` | Nhận `PlaceOrderRequest` và header `Idempotency-Key`; tạo hoặc replay kết quả checkout. |
| `GET /idempotency/orders/{idempotencyKey}` | Chủ đơn tra cứu response checkout đã hoàn tất sau timeout/mất response. Không tiết lộ key của user khác. |

`POST /orders` chỉ chấp nhận các dòng đã được preview/acknowledge: `cart_item_id`, `cart_item_version`, SKU, quantity, price source, flash-sale item (nếu có), unit price, subtotal, `expected_total`, tiền tệ `VND`, và recipient. Header key bắt buộc, dài 8–255 ký tự. Xác thực session và CSRF giống các mutation hiện có.

### Validation trước lock

- Kiểm tra JSON, UUID, số lượng dòng (1–20), dòng trùng, VND, số tiền nguyên dương và recipient (`name` 1–120, `phone` 8–20, `address` 1–500).
- Chuẩn hóa dữ liệu cần hash; canonical request phải gồm tập dòng được sắp ổn định, version, quote đã xác nhận, total/currency và recipient sau chuẩn hóa.
- Sai cấu trúc trả `422 VALIDATION_ERROR`; không tạo idempotency outcome giả cho lỗi hạ tầng.

## 5. Backend Phase 04

### 5.1 Module và thay đổi dữ liệu

Tạo hoàn thiện module `app/backend/internal/module/order/` cho place-order và attempt lookup, nối route vào router/bootstrap hiện có. Tái sử dụng logic quote của cart làm chính sách chung, không sao chép quy tắc giá/flash-sale thành một implementation khác.

Schema hiện có đã có bảng `orders`, `order_items`, `inventory_reservations`, `idempotency_records`; không sửa lịch sử migration đã áp. Khi triển khai cần:

- Bổ sung vào model/DTO phần `order_items.cart_item_id` và `cart_item_version` mà migration `00010` đã thêm.
- Ghi đúng snapshot SKU/product/tên/attributes, quantity, price source, flash-sale item, unit price và subtotal vào `order_items`.
- Sinh `order_number` công khai duy nhất; dùng ID nội bộ cho khóa ngoại.
- Ghi `orders.status = pending_payment`, `currency = VND`, recipient snapshot và `expires_at = database_now + 5 phút` một lần duy nhất.
- Tạo đúng một `inventory_reservation` trạng thái `active` cho từng order item.

### 5.2 Transaction đặt đơn nguyên tử

Toàn bộ các bước dưới đây chạy trong **một PostgreSQL transaction**; bất kỳ rejection nào phải rollback mọi order, reservation và counter delta.

1. Điều phối record idempotency theo `(user_id, place_order, key)`, kiểm request hash và phân biệt request mới, replay, conflict hoặc đang xử lý.
2. Đọc thời gian từ database một lần để tính deadline cố định.
3. Khóa các cart line đã chọn theo ID ổn định; version thay đổi trả `CART_ITEM_CHANGED`.
4. Sắp xếp theo SKU và khóa theo thứ tự cố định: inventory theo `sku_id`, flash-sale item theo ID, rồi `flash_sale_customer_usage` của customer theo flash-sale item ID. Tạo/khóa usage row an toàn cho line flash-sale.
5. Tính lại quote từ dữ liệu đã khóa: product/SKU active, chiến dịch/window, inventory available, quota remaining, purchase limit, giá và tổng tiền.
6. Nếu quote khác client đã xác nhận, trả `PRICE_CHANGED` kèm quote mới; nếu một line không đủ điều kiện, trả line errors chuẩn. Cả hai loại business rejection là kết quả idempotent ổn định nhưng không tạo order/reservation/counter.
7. Chỉ khi mọi dòng hợp lệ, insert order/items/reservations; tăng `inventories.reserved`, `flash_sale_items.quota_reserved` và `flash_sale_customer_usage.reserved_quantity`; lưu `201` response vào idempotency record rồi commit.

Mọi key giống với hash khác trả `409 IDEMPOTENCY_CONFLICT`. Key giống và hash giống đã hoàn thành trả lại đúng HTTP status/body trước đó và header `Idempotency-Replayed: true`. Request đã commit nhưng client mất response phải có thể retry với cùng key hoặc dùng endpoint lookup. Lỗi 5xx/transaction rollback phải còn khả năng retry cùng key.

### 5.3 Ràng buộc tương tranh với Phase 05

`order.Finish` đã là đường duy nhất kết thúc reservation trong Phase 05. Checkout phải tạo dữ liệu theo đúng các giả định sau:

| Dữ liệu/tác động do checkout tạo | Cách Phase 05 dùng nó |
| --- | --- |
| Order `pending_payment`, `expires_at` cố định, `terminal_at = NULL` | Payment session chỉ được tạo cho đơn còn payable; cancel/expiry/callback chuyển trạng thái qua `Finish`. |
| Một reservation `active` cho mỗi order item | `Finish` khóa và đổi toàn bộ sang `finalized` khi confirmed hoặc `released` ở terminal state khác. |
| `inventory.reserved`, quota reserved, usage reserved đã tăng đầy đủ | `Finish` giảm reservation counter đúng một lần; confirmed cũng giảm `on_hand` và tăng sold. |
| `cart_item_id` và `cart_item_version` | Confirmation chỉ xóa cart line còn cùng version; dòng khách đã sửa phải được giữ lại. |
| Snapshot amount/currency ở order | Payment session và callback so khớp số tiền `VND`, không lấy giá live. |

Thứ tự lock là điều kiện không được phá: checkout khóa cart lines trước, rồi inventory theo SKU tăng dần, flash-sale item theo ID tăng dần và usage theo item ID tăng dần. `Finish` đã khóa order, cart lines (chỉ khi confirmed), reservations, rồi inventory → flash-sale item → usage. Không được thêm khóa payment session vào checkout hoặc đảo bất kỳ thứ tự tài nguyên chung nào. Phase 05 tự kiểm deadline bằng database `clock_timestamp()` sau khi giữ lock; checkout chỉ tạo deadline, không gia hạn nó khi retry/replay.

### 5.4 Metrics và log

Thêm metric checkout ở observability boundary hiện có, không dùng user ID, order ID hoặc idempotency key làm label. Tối thiểu phân biệt: order mới, replay, business rejection, in-progress/conflict và system error. Log phải có request ID/outcome code nhưng không log recipient đầy đủ, session secret hoặc raw idempotency payload.

## 6. Frontend Phase 04 và nối vào giao diện có sẵn

Cart hiện đã có selection, preview và acknowledge. Mở rộng feature theo các nguyên tắc sau:

1. Chỉ hiện form recipient/nút đặt đơn khi có preview hiện tại đã được acknowledge.
2. Tạo một `CheckoutIntent` gồm selection/quote/recipient canonical và một UUID idempotency key. Lưu key trong `sessionStorage` để reload, retry và lookup dùng lại cùng key.
3. Khi selection, quote acknowledged hoặc recipient thay đổi, hủy intent cũ và tạo key mới cho ý định mới.
4. Gọi `POST /orders` với credentials, CSRF và `Idempotency-Key`; response `201` chuyển tới `/orders/:orderId` (trang đã có), nơi khách chọn liên kết **Thanh toán** tới `/orders/:orderId/payment`.
5. Timeout/lỗi mạng là trạng thái **chưa rõ**: hiển thị `Thử lại` (cùng key) và `Kiểm tra kết quả` (GET idempotency). Không redirect/không báo success trước khi nhận hoặc tra cứu được response đã lưu.
6. Business rejection giữ nguyên giỏ. Hiển thị line error; `PRICE_CHANGED` phải đưa quote mới về luồng preview/acknowledge trước khi có thể đặt lại.

Không tạo màn hình payment hoặc simulator mới: các route `/orders`, `/orders/:orderId`, `/orders/:orderId/payment`, `/admin/payment-simulator` của Phase 05 được tái sử dụng. Checkout không được tự giả định trạng thái thanh toán; nguồn thật là order API mà các trang Phase 05 đang poll.

## 7. Trình tự triển khai bắt buộc

### Slice A — Contract và order reservation

- Viết test integration thất bại cho regular multi-SKU success, mixed regular/flash success và một line rejection rollback tất cả.
- Hoàn thiện DTO, controller, service, quote adapter và deterministic locking.
- Ghi snapshot/reservation/counter; cập nhật models cho cart provenance.
- Chạy `TEST-01` đến `TEST-04`, `TEST-08` đến `TEST-10` cùng SQL oracle.

### Slice B — Idempotency và recovery

- Viết test cho replay cùng key, cùng key nhưng payload khác, hai API instance tranh key, và commit thành công nhưng response bị mất.
- Hoàn thiện canonical hash, record persistence, in-progress outcome, lookup ownership/missing behavior.
- Chạy `TEST-05` đến `TEST-07`; xác nhận retry không tạo order thứ hai và không gia hạn deadline.

### Slice C — UI checkout và handoff payment

- Viết test frontend cho recipient validation, lifecycle của intent key, unknown outcome và không render success giả.
- Nối cart preview acknowledge → place-order → `/orders/:id`; kiểm business rejection không mất cart.
- Dùng đơn tạo từ UI/API để kiểm route existing `/orders/:id/payment`, session và simulator callback thực sự hoạt động.

### Slice D — Integration gate

- Chạy backend test/vet/race phù hợp và frontend test/lint/typecheck/build.
- Chạy các race checkout đối đầu với payment-success, cancel và expiry worker; xác nhận lock order không deadlock và oracle không vi phạm.
- Chạy trọn `TEST-01` đến `TEST-25`; chỉ sau đó Phase 05 Step 3/baseline E0 mới được đóng.

## 8. Ma trận xác nhận nhanh

| Tình huống | Kết quả bắt buộc |
| --- | --- |
| Hai SKU, đủ hàng | Một order pending, đủ reservation active, các counter reserved đúng. |
| Một SKU thiếu/quota/limit không hợp lệ | Không order, không reservation và không có partial counter delta. |
| Giá thay đổi sau preview | `PRICE_CHANGED`, trả quote hiện tại, không giữ hàng. |
| Hai request cùng idempotency key | Một kết quả bền vững; request sau replay hoặc nhận in-progress retryable. |
| Client mất response sau commit | Lookup/retry cùng key trả order cũ, không gia hạn 5 phút. |
| Payment success sau khi checkout | `Finish` finalize reservation, trừ tồn một lần và chỉ xóa cart line cùng version. |
| Callback success đối đầu cancel/expiry | Chỉ một terminal state thắng; bên còn lại lưu evidence cần thiết, không đảo counter lần nữa. |
| Callback success muộn | Không hồi sinh order; Phase 05 tạo reconciliation theo chính sách hiện có. |

## 9. Checklist trước khi bắt đầu code

- [ ] Đọc PRD requirement `FR-021`–`FR-040`, `INV-01`–`INV-12`, `TEST-01`–`TEST-25` và OpenAPI operation liên quan.
- [ ] Không sửa behavior Phase 05 đã merge nếu chưa chứng minh contract không khớp; ưu tiên làm checkout thỏa dữ liệu đầu vào của nó.
- [ ] Không dùng tài liệu research về PayOS sandbox làm yêu cầu hiện hành.
- [ ] Viết test thất bại cho từng slice trước implementation; mọi thay đổi tới money/inventory/state phải có test concurrency và oracle phù hợp.
- [ ] Khi phát hiện cần đổi BR/FR/INV/OpenAPI, dừng implementation và cập nhật nguồn quyết định/traceability trước theo ADR-003.
