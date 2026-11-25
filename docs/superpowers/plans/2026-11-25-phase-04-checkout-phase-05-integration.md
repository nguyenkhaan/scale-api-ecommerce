# Phase 04 Checkout + tích hợp Phase 05 — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Khách đặt được một đơn nhiều SKU giữ hàng nguyên tử từ giỏ hàng đã xem trước, với idempotency bền vững, rồi đi tiếp vào các trang đơn/thanh toán mô phỏng Phase 05 đã có.

**Architecture:** Module `order` sở hữu transaction checkout (`CheckoutService`) và hai endpoint `POST /orders`, `GET /idempotency/orders/{key}`. Chính sách tính giá/điều kiện mua được tái sử dụng từ `cart` qua một hàm mới `cart.LockAndQuote` chạy dưới khóa theo đúng thứ tự data-model, nên preview và checkout chỉ có một implementation của quy tắc giá. Frontend thêm feature `checkout` (intent key trong `sessionStorage`, form người nhận, trạng thái kết quả chưa rõ) gắn vào panel xem trước của giỏ hàng; thành công thì chuyển tới `/orders/:orderId` có sẵn.

**Tech Stack:** Go 1.27, Gin 1.12, GORM + pgx, PostgreSQL (Neon), Goose; React 19, React Router 7, TanStack Query 5, Bun test, TypeScript 6.

**Spec:** `docs/designs/phase-04-checkout-phase-05-integration.md` (đọc kèm `docs/technical/api/openapi.yaml` các operation `placeOrder`, `getOrderAttempt`; `docs/technical/data-model.md` mục "Thứ tự khóa và transaction boundary"; mockup `docs/ui-design/pages/Preview*.html`, `OrderResult*.html`).

## Global Constraints

- Không sửa migration đã áp (`00001`–`00010`). Schema hiện có đủ cho checkout; không cần migration mới.
- Toàn bộ idempotency/order/items/reservations/counter chạy trong **một** PostgreSQL transaction; không gọi HTTP, payment simulator, Redis hay broker bên trong.
- Thời gian lấy từ database: `expires_at = now() + interval '5 minutes'` trong cùng transaction; retry/replay không bao giờ gia hạn.
- Thứ tự khóa bắt buộc: idempotency row → cart lines theo `id` tăng → `inventories` theo `sku_id` tăng → `flash_sale_items` theo `id` tăng → `flash_sale_customer_usage` theo `flash_sale_item_id` tăng. Không khóa `orders`/`payment_sessions` trong checkout.
- Server tính lại mọi giá; không tin `unit_price`/`subtotal`/`expected_total` của client ngoài việc so sánh.
- Checkout **không** xóa dòng giỏ; chỉ `order.Finish(..., OutcomeConfirmed)` của Phase 05 xóa dòng còn đúng `cart_item_version`.
- Không đổi behavior Phase 05 đã merge (`order.Finish`, payment, expiry). Nếu phát hiện contract không khớp, dừng và cập nhật PRD/OpenAPI/traceability trước (ADR-003).
- Mã lỗi theo OpenAPI: `IDEMPOTENCY_CONFLICT`, `IDEMPOTENCY_IN_PROGRESS`, `PRICE_CHANGED`, `CART_ITEM_CHANGED`, các mã line của BR-24, `VALIDATION_ERROR`.
- `Idempotency-Key`: 8–255 ký tự. Recipient sau khi trim: `name` 1–120, `phone` 8–20, `address` 1–500 ký tự (đếm rune, khớp `char_length` của PostgreSQL).
- Log và metric không chứa user ID, order ID, idempotency key, recipient hay request body.
- Test PostgreSQL dùng `TEST_DATABASE_URL` riêng (khác `DATABASE_URL`); test bị skip không được tính là pass.
- Copy giao diện tiếng Việt lấy đúng từ mockup; tiền hiển thị bằng `formatVND`.

## Review Focus

1. **Hai request cùng key đến gần như đồng thời qua hai instance:** request sau phải chờ rồi replay đúng response của request trước, hoặc nhận `409 IDEMPOTENCY_IN_PROGRESS` + `Retry-After`; không bao giờ tạo đơn thứ hai. Test: Task 4 `TestPlaceOrderSameKeyWhileFirstHoldsLockReturnsInProgress`, Task 5 `TestPlaceOrderSameKeyAcrossTwoRoutersCreatesOneOrder`.
2. **Từ chối nghiệp vụ không để lại dấu vết:** sau khi bị từ chối, không còn row `flash_sale_customer_usage` rỗng do checkout tạo ra, không còn counter delta, không còn khóa bị giữ. Test: Task 3 `TestPlaceOrderRejectsWholeOrderWhenLaterSKUIsShort` (kiểm usage count = 0).
3. **Khách sửa giỏ trong lúc đơn cũ đang chờ thanh toán, rồi đặt tiếp từ cùng dòng giỏ:** checkout và `Finish(confirmed)` cùng khóa dòng giỏ trước inventory, nên không deadlock; thanh toán đơn cũ không xóa dòng đã sửa. Test: Task 6 `TestCheckoutRacesPaymentConfirmationOnSameCartLine`, `TestPaymentSuccessKeepsCartLineEditedAfterCheckout`.
4. **Mất phản hồi rồi reload trang:** frontend đọc lại intent từ `sessionStorage` và tra cứu/retry bằng **cùng** key; không tự tạo key mới; không hiện "Đặt đơn thành công" khi chưa có response 201 hoặc lookup 201. Test: Task 7 `resolveIntent reuses the stored key for the same fingerprint`, Task 8 `unknown outcome never renders a success heading`.
5. **Client gửi giá lạ (sửa tay payload):** `subtotal ≠ quantity × unit_price` hoặc `expected_total ≠ Σ subtotal` bị `422` trước khi khóa; giá hợp lệ về số học nhưng khác giá server thì `409 PRICE_CHANGED` kèm `current_unit_price`. Test: Task 1 bảng validation, Task 3 `TestPlaceOrderPriceChangedReturnsCurrentPriceWithoutReserving`.

---

## File map

| File | Trách nhiệm |
|---|---|
| `app/backend/internal/httpserver/problem/problem.go` | Thêm 2 mã idempotency; `LineErrorsProblem` dựng body để lưu |
| `app/backend/internal/module/order/checkout_dto.go` | `PlaceOrderRequest`, `PlaceOrderLine`, `RecipientRequest`, `IdempotencyResult` |
| `app/backend/internal/module/order/checkout_validate.go` | Validate + chuẩn hóa request và key |
| `app/backend/internal/module/order/checkout_hash.go` | Canonical request hash SHA-256 |
| `app/backend/internal/module/cart/checkout_quote.go` | `LockAndQuote`: khóa theo thứ tự + quote bằng `evaluateLine` |
| `app/backend/internal/module/cart/preview.go`, `lookup.go` | Tách `quoteLines` dùng chung; `skuInfo` có `attributes` |
| `app/backend/internal/module/order/checkout_service.go` | Transaction đặt đơn, so quote, ghi snapshot/reservation/counter, idempotency |
| `app/backend/internal/module/order/checkout_number.go` | Sinh `order_number` |
| `app/backend/internal/module/order/checkout_attempt.go` | `LookupAttempt` |
| `app/backend/internal/module/order/checkout_metrics.go` | `CheckoutRecorder`, `CheckoutCounters` |
| `app/backend/internal/module/order/checkout_controller.go` | HTTP mapping, `RegisterCheckoutRoutes` |
| `app/backend/internal/model/order.go`, `order_item.go` | Bổ sung cột đã có trong migration `00010` |
| `app/backend/internal/httpserver/router.go`, `middleware/cors.go`, `cmd/api/main.go` | Wiring + header CORS |
| `app/frontend/src/lib/json-http.ts` | Header tùy chỉnh, timeout, đọc response header |
| `app/frontend/src/features/checkout/*` | Type, service, logic, hook, components |
| `app/frontend/src/features/cart/components/CheckoutPreviewPanel.tsx`, `pages/CartPage.tsx` | Gắn luồng đặt đơn |

---

### Task 1: Hợp đồng request, validation và canonical hash

**Files:**
- Create: `app/backend/internal/module/order/checkout_dto.go`, `checkout_validate.go`, `checkout_hash.go`
- Modify: `app/backend/internal/httpserver/problem/problem.go`
- Test: `app/backend/test/order_checkout_validation_test.go` (không cần DB)

**Interfaces:**
- Produces:
  - `type PlaceOrderLine struct { CartItemID string \`json:"cart_item_id"\`; CartItemVersion int64 \`json:"cart_item_version"\`; SKUID string \`json:"sku_id"\`; Quantity int \`json:"quantity"\`; UnitPrice int64 \`json:"unit_price"\`; Subtotal int64 \`json:"subtotal"\`; PriceSource string \`json:"price_source"\`; FlashSaleItemID *string \`json:"flash_sale_item_id"\` }`
  - `type RecipientRequest struct { Name, Phone, Address string }` (json `name`, `phone`, `address`)
  - `type PlaceOrderRequest struct { Lines []PlaceOrderLine; Currency string; ExpectedTotal int64; Recipient RecipientRequest }` (json `lines`, `currency`, `expected_total`, `recipient`)
  - `type IdempotencyResult struct { IdempotencyKey string; ResponseStatus int; ResponseBody json.RawMessage; CompletedAt time.Time }` (json snake_case theo OpenAPI)
  - `func ValidatePlaceOrder(req *PlaceOrderRequest) []problem.Violation` — chuẩn hóa tại chỗ: UUID về chữ thường, recipient `strings.TrimSpace`.
  - `func ValidateIdempotencyKey(key string) []problem.Violation` — field `"Idempotency-Key"`.
  - `func CanonicalRequestHash(req PlaceOrderRequest) [32]byte` — gọi sau `ValidatePlaceOrder`.
  - `problem.CodeIdempotencyConflict = "IDEMPOTENCY_CONFLICT"`, `problem.CodeIdempotencyInProgress = "IDEMPOTENCY_IN_PROGRESS"`
  - `func problem.LineErrorsProblem(detail string, lines []LineProblem) Problem` (status 409, code = mã nghiêm trọng nhất, dùng lại `mostSevereLineCode`) và `func problem.CodeProblem(status int, code, detail string) Problem`; `func problem.WriteRaw(ctx *gin.Context, status int, contentType string, body []byte, headers map[string]string)`.

- [ ] **Step 1: Viết test bảng thất bại** `TestValidatePlaceOrderRejectsMalformedRequests` với các case và field mong đợi:
  - `lines` rỗng → `lines`; 21 dòng → `lines`
  - `cart_item_id` không phải UUID → `lines[0].cart_item_id`; trùng `cart_item_id` → `lines[1].cart_item_id`; trùng `sku_id` → `lines[1].sku_id`
  - `quantity: 0` → `lines[0].quantity`; `cart_item_version: 0` → `lines[0].cart_item_version`; `unit_price: 0` → `lines[0].unit_price`
  - `subtotal` ≠ `quantity*unit_price` → `lines[0].subtotal`
  - `price_source: "flash_sale"` thiếu `flash_sale_item_id` → `lines[0].flash_sale_item_id`; `"regular"` có `flash_sale_item_id` → cùng field; `price_source: "vip"` → `lines[0].price_source`
  - `currency: "USD"` → `currency`; `expected_total` ≠ Σ subtotal → `expected_total`
  - `recipient.name` chỉ có khoảng trắng → `recipient.name`; name 121 rune → `recipient.name`; phone `"1234567"` → `recipient.phone`; address 501 rune → `recipient.address`
  
  Và `TestValidatePlaceOrderNormalizesValidRequest`: UUID chữ hoa → chữ thường; `"  An  "` → `"An"`; không có violation.
  Và `TestValidateIdempotencyKeyLength`: 7 ký tự lỗi, 8 và 255 hợp lệ, 256 lỗi.
- [ ] **Step 2: Viết test hash thất bại** `TestCanonicalRequestHash`:
  - đảo thứ tự `lines` → hash bằng nhau
  - recipient `" An "` vs `"An"` (sau validate) → bằng nhau
  - đổi `phone`, `cart_item_version`, `unit_price`, `flash_sale_item_id` → mỗi thay đổi cho hash khác
- [ ] **Step 3: Chạy** `cd app/backend && go test ./test -run 'TestValidatePlaceOrder|TestValidateIdempotencyKey|TestCanonicalRequestHash' -v` → FAIL (undefined).
- [ ] **Step 4: Implement** DTO, validate, hash và các bổ sung `problem`. Hash: sao chép request, sắp `Lines` theo `CartItemID`, `json.Marshal` một struct có thứ tự field cố định, `sha256.Sum256`. Đếm độ dài bằng `utf8.RuneCountInString`.
- [ ] **Step 5: Chạy lại lệnh Step 3** → PASS; `go vet ./...` sạch.
- [ ] **Step 6: Commit** `feat(order): add checkout request validation and canonical hash`

---

### Task 2: `cart.LockAndQuote` — quote dưới khóa, dùng chung chính sách preview

**Files:**
- Create: `app/backend/internal/module/cart/checkout_quote.go`
- Modify: `app/backend/internal/module/cart/preview.go` (tách vòng lặp đánh giá ra `quoteLines`), `lookup.go` (`skuInfo` thêm `Attributes json.RawMessage` từ `s.attributes`)
- Test: `app/backend/test/cart_lock_quote_test.go`

**Interfaces:**
- Consumes: `CheckoutSelection`, `evaluateLine`, `loadLineContexts`, `dbNow` có sẵn trong `cart`.
- Produces:
  - `type QuotedLine struct { CheckoutLineResponse; SKUCode string; ProductName string; SKUAttributes json.RawMessage }`
  - `type LockedQuote struct { Now time.Time; Lines []QuotedLine; TotalAmount int64 }`
  - `func LockAndQuote(tx *gorm.DB, userID string, selection CheckoutSelection) (LockedQuote, error)` — phải gọi trong transaction của caller. Lỗi: `*ValidationError` (id không thuộc giỏ), `ErrCartItemChanged`, `*LineErrors`.
  - `func quoteLines(lines []selectedLine, contexts map[string]lineContext, now time.Time) ([]QuotedLine, int64, []problem.LineProblem)` (nội bộ; `previewInSnapshot` dùng lại).

- [ ] **Step 1: Viết test thất bại**
  - `TestLockAndQuoteMatchesPreviewForMixedLines`: một dòng regular và một dòng flash sale; trong `database.Transaction`, `LockAndQuote` trả cùng `TotalAmount`, `UnitPrice`, `PriceSource` với `PreviewCheckout` và có `SKUCode`, `ProductName`.
  - `TestLockAndQuoteHoldsInventoryAndCartLocks`: trong lúc transaction của `LockAndQuote` còn mở, transaction thứ hai chạy `SELECT 1 FROM inventories WHERE sku_id = ? FOR UPDATE NOWAIT` và `SELECT 1 FROM cart_items WHERE id = ? FOR UPDATE NOWAIT` → cả hai lỗi SQLSTATE `55P03`.
  - `TestLockAndQuoteCreatesZeroUsageRowForFirstFlashSaleLine`: sau `LockAndQuote` (chưa commit) có row usage `reserved_quantity=0, sold_quantity=0`; dòng regular không tạo row.
  - `TestLockAndQuoteStaleVersionReturnsCartItemChanged`.
- [ ] **Step 2: Chạy** `go test ./test -run 'TestLockAndQuote' -v` (có `TEST_DATABASE_URL`) → FAIL.
- [ ] **Step 3: Implement.** Thứ tự câu lệnh: `dbNow`; `SELECT ... FROM cart_items ci JOIN carts c ... WHERE ci.id IN ? ORDER BY ci.id FOR UPDATE OF ci` (giữ kiểm tra chủ sở hữu, id lạ và version như `loadSelectedLines`); `SELECT sku_id FROM inventories WHERE sku_id IN ? ORDER BY sku_id FOR UPDATE`; `SELECT id FROM flash_sale_items WHERE id IN ? ORDER BY id FOR UPDATE`; `INSERT INTO flash_sale_customer_usage (flash_sale_item_id, user_id) SELECT unnest(?::uuid[]), ? ORDER BY 1 ON CONFLICT DO NOTHING`; `SELECT ... FROM flash_sale_customer_usage WHERE user_id = ? AND flash_sale_item_id IN ? ORDER BY flash_sale_item_id FOR UPDATE`; rồi `loadLineContexts` + `quoteLines`. Kết quả trả theo thứ tự selection.
- [ ] **Step 4: Chạy** `go test ./test -run 'TestLockAndQuote|TestCheckoutPreview|TestCart' -v` → PASS (preview cũ không đổi hành vi).
- [ ] **Step 5: Commit** `feat(cart): share checkout quote policy under deterministic locks`

---

### Task 3: Transaction đặt đơn nguyên tử

**Files:**
- Create: `app/backend/internal/module/order/checkout_service.go`, `checkout_number.go`, `checkout_metrics.go`
- Modify: `app/backend/internal/model/order_item.go` (`CartItemID *string`, `CartItemVersion *int64`), `model/order.go` (`UpdatedAt time.Time`, `TerminalAt *time.Time`)
- Test: `app/backend/test/order_checkout_support_test.go`, `app/backend/test/order_checkout_test.go`

**Interfaces:**
- Consumes: Task 1 DTO/hash, Task 2 `cart.LockAndQuote`, `getOwnedOrder(ctx, db, userID, orderID)` trong `order/query.go`.
- Produces:
  - `type PlaceOrderInput struct { UserID, IdempotencyKey, RequestID string; Request PlaceOrderRequest }` (Request đã qua `ValidatePlaceOrder`)
  - `type PlaceOrderResult struct { Status int; Body json.RawMessage; Replayed bool; Outcome CheckoutOutcome }` — `Status` là 201 hoặc 409 (từ chối nghiệp vụ đã lưu).
  - Trong `checkout_metrics.go`: `type CheckoutOutcome string` với `OutcomeCreated="created"`, `OutcomeReplayed="replayed"`, `OutcomeRejected="rejected"`, `OutcomeConflict="conflict"`, `OutcomeInProgress="in_progress"`, `OutcomeError="error"`; `type CheckoutRecorder interface { RecordCheckout(outcome CheckoutOutcome, code string) }`.
  - `type CheckoutService struct` ; `func NewCheckoutService(db *gorm.DB, recorder CheckoutRecorder) (*CheckoutService, error)` (`recorder` nil thì dùng no-op).
  - `func (s *CheckoutService) PlaceOrder(ctx context.Context, in PlaceOrderInput) (PlaceOrderResult, error)`
  - `func newOrderNumber(now time.Time) (string, error)` → `"SS-" + now.UTC().Format("20060102") + "-" + 8 ký tự crypto/rand từ bảng "0123456789ABCDEFGHJKMNPQRSTVWXYZ"`.
  - Test helper: `checkoutRequestFromPreview(preview cart.CheckoutPreviewResponse) order.PlaceOrderRequest` (recipient mặc định `Nguyễn Văn An` / `0901234567` / `12 Nguyễn Trãi, Quận 1`) và `checkoutNewService(t, db) *order.CheckoutService`.

- [ ] **Step 1: Viết test thất bại** (mọi test gọi `lifecycleRequireInvariants` ở cuối):
  - `TestPlaceOrderRegularMultiSKUCreatesPendingOrderWithReservations` (TEST-01): 2 SKU regular (on_hand 10, qty 2 và 1). Kỳ vọng `Status 201`; body có `status:"pending_payment"`, `payable:true`, `order_number` khớp `^SS-\d{8}-[0-9A-Z]{8}$`; `expires_at - created_at = 5 phút` (đọc bằng SQL); 2 `order_items` có `cart_item_id`/`cart_item_version` đúng dòng giỏ; 2 reservation `active`; `inventories.reserved` tăng đúng 2 và 1; dòng giỏ vẫn còn; `idempotency_records.response_status = 201`.
  - `TestPlaceOrderMixedRegularAndFlashSaleReservesQuotaAndUsage`: dòng sale qty 2 → `quota_reserved +2`, usage `reserved_quantity = 2`; `order_items.price_source='flash_sale'`, `flash_sale_item_id` đúng.
  - `TestPlaceOrderRejectsWholeOrderWhenLaterSKUIsShort` (TEST-02): SKU A đủ, SKU B (dòng flash sale, on_hand 1, qty 2). Kỳ vọng `Status 409`, body `code:"INSUFFICIENT_STOCK"` có `line_errors[0].sku_id = B`; 0 order của user; counters A/B không đổi; **0 row usage** của user cho sale đó; idempotency lưu 409.
  - `TestPlaceOrderPriceChangedReturnsCurrentPriceWithoutReserving` (TEST-08): preview, rồi `UPDATE skus SET regular_price = 510000`; kỳ vọng 409 `PRICE_CHANGED`, `line_errors[0].current_unit_price = 510000`; không có reservation.
  - `TestPlaceOrderSaleQuotaExhaustedRejectsWithoutRegularFallback` (TEST-09): quota còn 0, tồn còn 10 → 409 `SALE_QUOTA_EXCEEDED`, không order.
  - `TestPlaceOrderStaleCartVersionReturnsCartItemChanged`: sửa qty dòng giỏ sau preview → 409 `CART_ITEM_CHANGED`.
- [ ] **Step 2: Chạy** `go test ./test -run 'TestPlaceOrder' -v` → FAIL.
- [ ] **Step 3: Implement `PlaceOrder`.** Một `db.WithContext(ctx).Transaction` (READ COMMITTED):
  1. `SET LOCAL lock_timeout = '2s'`; `INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash) VALUES (?, 'place_order', ?, ?) ON CONFLICT (user_id, operation, idempotency_key) DO NOTHING RETURNING id`; `SET LOCAL lock_timeout = DEFAULT`. Không có row trả về → trả lỗi nội bộ `errIdempotencyExists` (Task 4 xử lý).
  2. `tx.SavePoint("checkout_quote")`; `cart.LockAndQuote` với selection từ `Lines`.
  3. So quote với request (hàm `compareQuote(req PlaceOrderRequest, quote cart.LockedQuote) (lineProblems []problem.LineProblem, cartChanged bool)`): khác `sku_id`/`quantity` → `cartChanged`; khác `unit_price`/`price_source`/`flash_sale_item_id` → `LineProblem{Code: PRICE_CHANGED, CurrentUnitPrice: giá server}`.
  4. Từ chối nghiệp vụ (`*cart.LineErrors`, `ErrCartItemChanged`, `PRICE_CHANGED`): `tx.RollbackTo("checkout_quote")` (nhả khóa và row usage rỗng), dựng body bằng `problem.LineErrorsProblem`/`problem.CodeProblem` với `RequestID` của input, lưu vào idempotency (`response_status=409, response_body, completed_at=now()`), commit.
  5. `*cart.ValidationError` hoặc lỗi hạ tầng: trả lỗi → rollback toàn bộ, không lưu outcome.
  6. Thành công: insert `orders` (`expires_at = now() + interval '5 minutes'`, recipient đã trim, `total_amount = quote.TotalAmount`), `order_items` (snapshot từ `QuotedLine` + `cart_item_id`/`cart_item_version`), một `inventory_reservations` `active` mỗi item; `UPDATE inventories SET reserved = reserved + ?, version = version + 1, updated_at = now()`; `UPDATE flash_sale_items SET quota_reserved = quota_reserved + ?, updated_at = now()`; `UPDATE flash_sale_customer_usage SET reserved_quantity = reserved_quantity + ?, updated_at = now()` — cập nhật theo đúng thứ tự khóa. Body = `json.Marshal(getOwnedOrder(ctx, tx, userID, orderID))`; lưu idempotency 201; commit.
- [ ] **Step 4: Chạy** `go test ./test -run 'TestPlaceOrder' -v` → PASS.
- [ ] **Step 5: Commit** `feat(order): place multi-SKU orders atomically with reservations`

---

### Task 4: Ngữ nghĩa idempotency và tra cứu kết quả

**Files:**
- Modify: `app/backend/internal/module/order/checkout_service.go`
- Create: `app/backend/internal/module/order/checkout_attempt.go`
- Test: `app/backend/test/order_checkout_idempotency_test.go`

**Interfaces:**
- Produces:
  - `var ErrIdempotencyConflict, ErrIdempotencyInProgress, ErrAttemptNotFound error`
  - `PlaceOrder` khi key đã tồn tại: đọc row đã commit; hash khác → `ErrIdempotencyConflict`; hash bằng và `completed_at` có giá trị → `PlaceOrderResult{Status, Body: stored, Replayed: true}`; lock timeout (SQLSTATE `55P03`) ở bước insert → `ErrIdempotencyInProgress`.
  - `func (s *CheckoutService) LookupAttempt(ctx context.Context, userID, key string) (IdempotencyResult, error)` — chỉ row của chính user và đã hoàn tất; ngược lại `ErrAttemptNotFound`.

- [ ] **Step 1: Viết test thất bại**
  - `TestPlaceOrderSameKeySamePayloadReplaysWithoutSecondOrder`: gọi 2 lần → lần 2 `Replayed=true`, `Body` byte-bằng lần 1; 1 order; `expires_at` không đổi; counters không đổi.
  - `TestPlaceOrderSameKeyDifferentPayloadConflicts` (TEST-06): đổi `recipient.phone` → `ErrIdempotencyConflict`; vẫn 1 order.
  - `TestPlaceOrderReplaysStoredRejectionAfterStockReturns` (BR-36): lần 1 bị `INSUFFICIENT_STOCK`; tăng `on_hand`; gửi lại cùng key/payload → vẫn 409 đã lưu, `Replayed=true`, 0 order.
  - `TestPlaceOrderInfrastructureFailureLeavesKeyRetryable`: gọi với `ctx` đã cancel → error; không có idempotency row; gọi lại cùng key với ctx hợp lệ → 201.
  - `TestPlaceOrderSameKeyWhileFirstHoldsLockReturnsInProgress`: mở transaction thủ công insert cùng `(user, place_order, key)` và chưa commit; `PlaceOrder` trả `ErrIdempotencyInProgress` trong < 5 giây; rollback transaction giữ khóa; gọi lại → 201.
  - `TestLookupAttemptAfterLostResponse` (TEST-07): `PlaceOrder` rồi bỏ kết quả; `LookupAttempt` trả `ResponseStatus 201`, body có cùng `id` order; `PlaceOrder` lại → replay cùng `expires_at`.
  - `TestLookupAttemptIsOwnerOnly`: user B tra key của user A → `ErrAttemptNotFound`; key chưa dùng → `ErrAttemptNotFound`.
- [ ] **Step 2: Chạy** `go test ./test -run 'TestPlaceOrderSameKey|TestPlaceOrderReplays|TestPlaceOrderInfrastructure|TestLookupAttempt' -v` → FAIL.
- [ ] **Step 3: Implement** nhánh key-exists và `LookupAttempt` (so hash bằng `bytes.Equal` trên 32 byte; nhận diện `55P03` qua `*pgconn.PgError`).
- [ ] **Step 4: Chạy lại** cùng lệnh + `go test ./test -run 'TestPlaceOrder' -v` → PASS.
- [ ] **Step 5: Commit** `feat(order): replay, conflict and lookup for checkout idempotency`

---

### Task 5: HTTP, wiring, CORS và metrics/log

**Files:**
- Create: `app/backend/internal/module/order/checkout_controller.go`
- Modify: `order/checkout_metrics.go`, `internal/httpserver/router.go` (`Dependencies.CheckoutService *order.CheckoutService`), `internal/httpserver/middleware/cors.go`, `cmd/api/main.go`
- Test: `app/backend/test/order_checkout_api_test.go`, `app/backend/test/order_checkout_metrics_test.go`, bổ sung `app/backend/test/cors_test.go`

**Interfaces:**
- Produces:
  - `func NewCheckoutController(service *CheckoutService) *CheckoutController`
  - `func RegisterCheckoutRoutes(api *gin.RouterGroup, controller *CheckoutController, requireSession gin.HandlerFunc)` → `POST /orders`, `GET /idempotency/orders/:idempotencyKey`.
  - Consumes `CheckoutOutcome`, `CheckoutRecorder` từ Task 3.
  - `type CheckoutCounters struct` (atomic, an toàn đa luồng); `func NewCheckoutCounters() *CheckoutCounters`; `func (c *CheckoutCounters) Snapshot() map[CheckoutOutcome]int64`. `RecordCheckout` vừa tăng counter vừa `log.Printf("checkout outcome=%s code=%s", ...)`; controller thêm `request_id`.

- [ ] **Step 1: Viết test HTTP thất bại** (router dựng bằng `httpserver.NewRouter` với `AuthService`, `CartService`, `OrderQueries`, `CheckoutService`):
  - `TestPlaceOrderHTTPCreatesOrderAndReplays`: lần 1 `201`, header `Idempotency-Replayed: false`, body là `Order`; lần 2 cùng key `201` + `Idempotency-Replayed: true`, body bằng nhau.
  - `TestPlaceOrderHTTPRejectsInvalidInput`: thiếu header → 422 field `Idempotency-Key`; field lạ `"total":1` → 422 (`DisallowUnknownFields`); JSON hỏng → 422; không session → 401; thiếu CSRF → 403 (mã như các mutation hiện có).
  - `TestPlaceOrderHTTPStoredRejectionUsesProblemContentType`: 409 `application/problem+json` có `line_errors`.
  - `TestPlaceOrderHTTPConflictAndInProgress`: hash khác → 409 `IDEMPOTENCY_CONFLICT`; khóa giữ như Task 4 → 409 `IDEMPOTENCY_IN_PROGRESS` + `Retry-After: 1`.
  - `TestPlaceOrderSameKeyAcrossTwoRoutersCreatesOneOrder` (TEST-05): 2 router độc lập cùng DB, 2 goroutine cùng key → mỗi response là 201 (một trong hai replayed) hoặc 409 `IDEMPOTENCY_IN_PROGRESS`; đúng 1 order, 1 idempotency row.
  - `TestGetOrderAttemptHTTP`: chủ key → 200 đúng schema `IdempotencyResult`; user khác → 404 `RESOURCE_NOT_FOUND`.
  - `TestCORSAllowsIdempotencyHeaders`: preflight `OPTIONS /api/orders` có `Idempotency-Key` trong `Access-Control-Allow-Headers`; response thường có `Access-Control-Expose-Headers` gồm `Idempotency-Replayed, Retry-After, X-Request-ID`.
  - `TestCheckoutCountersAndLogsOmitSecrets`: sau một lần created, một lần replayed, một lần rejected → snapshot `{created:1, replayed:1, rejected:1}`; log đã bắt không chứa key, phone, address, user ID.
- [ ] **Step 2: Chạy** `go test ./test -run 'TestPlaceOrderHTTP|TestPlaceOrderSameKeyAcross|TestGetOrderAttemptHTTP|TestCORSAllowsIdempotency|TestCheckoutCounters' -v` → FAIL.
- [ ] **Step 3: Implement controller** (bind với `json.Decoder.DisallowUnknownFields`, rồi `ValidateIdempotencyKey`, `ValidatePlaceOrder`; map lỗi như mô tả Interfaces; ghi body đã lưu bằng `problem.WriteRaw` với content type `application/json` cho 201, `application/problem+json` cho 409), wiring router/main (`order.NewCheckoutService(orm, order.NewCheckoutCounters())`), CORS.
- [ ] **Step 4: Chạy lại** → PASS; `go build ./... && go vet ./...` sạch.
- [ ] **Step 5: Commit** `feat(order): expose checkout and attempt lookup endpoints`

---

### Task 6: Cổng tương tranh và tích hợp Phase 05

**Files:**
- Test: `app/backend/test/order_checkout_concurrency_test.go`, `app/backend/test/order_checkout_payment_test.go`
- Modify: `app/backend/test/payment_support_test.go` (thêm `newCheckoutPaymentEnv(t) paymentEnv` đăng ký thêm cart + checkout routes trên cùng engine)

**Interfaces:**
- Consumes: `cartRunParallel`, `lifecycleRequireInvariants`, `order.Finish`, `order.ExpireDue(ctx, db, batchSize)`, payment simulator qua `paymentRouterSender`.

- [ ] **Step 1: Viết test**
  - `TestCheckoutLastUnitsContention` (TEST-03): `on_hand 5`, 12 khách mỗi người qty 1, key khác nhau, song song → đúng 5 đơn 201, 7 response 409 `INSUFFICIENT_STOCK`, không có 500; `reserved = 5`.
  - `TestCheckoutOverlappingCartsInReverseOrderNeverPartial` (TEST-04): SKU A, B mỗi SKU tồn 1; khách 1 thêm A rồi B, khách 2 thêm B rồi A; đặt song song; lặp 20 vòng → mỗi vòng đúng một đơn đủ 2 dòng, đơn kia 409; không có 500/deadlock.
  - `TestCheckoutPurchaseLimitAcrossKeys` (TEST-10): limit 2, cùng một dòng giỏ qty 1, 4 key song song → đúng 2 đơn; usage `reserved_quantity = 2`.
  - `TestCheckoutRacesPaymentConfirmationOnSameCartLine`: đơn O1 tạo qua API từ dòng L; song song (a) simulator success cho O1, (b) đặt O2 từ cùng dòng L → không lỗi 500/deadlock; O1 `confirmed`.
  - `TestCheckoutRacesExpiryWorker`: đơn sắp hết hạn (`UPDATE orders SET expires_at = now() + interval '1 second'` rồi chờ) + `ExpireDue` song song với checkout mới cùng SKU → bất biến giữ nguyên.
  - `TestCheckoutOrderFlowsIntoPaymentSuccess`: tạo đơn qua `POST /api/orders` → `POST /api/orders/{id}/payment-session` → admin gửi event success → order `confirmed`, reservation `finalized`, `on_hand` giảm đúng một lần, dòng giỏ đã đặt bị xóa.
  - `TestPaymentSuccessKeepsCartLineEditedAfterCheckout` (TEST-25): tạo đơn, sửa qty dòng giỏ (version tăng), thanh toán thành công → dòng giỏ còn, `order_items` giữ snapshot cũ.
  - `TestCheckoutOrderCancelReleasesEverything`: tạo đơn qua API rồi `POST /orders/{id}/cancel` → counters về như trước.
  
  Mọi test kết thúc bằng `lifecycleRequireInvariants`.
- [ ] **Step 2: Chạy** `go test ./test -run 'TestCheckout' -race -count=1 -v` → test nào lộ lỗi thì quay lại sửa trong Task 2–5 (không sửa `order.Finish`); kết thúc PASS.
- [ ] **Step 3: Chạy toàn bộ backend** `gofmt -l . && go build ./... && go vet ./... && go test ./... && go test ./test -race -count=1` → PASS, ghi số test pass/skip.
- [ ] **Step 4: Commit** `test(order): prove checkout contention and payment lifecycle handoff`

---

### Task 7: Frontend — transport, intent key và phân loại kết quả

**Files:**
- Modify: `app/frontend/src/lib/json-http.ts`
- Create: `app/frontend/src/features/checkout/checkout.type.ts`, `checkout.service.ts`, `checkout.logic.ts`
- Test: `app/frontend/src/test/lib/json-http.test.ts`, `src/test/checkout/checkout.service.test.ts`, `src/test/checkout/checkout.logic.test.ts`

**Interfaces:**
- Produces:
  - `json-http`: `RequestSpec` thêm `headers?: Record<string, string>` và `timeoutMs?: number` (hết giờ → `networkError()`); method mới `withHeaders<T>(path, spec): Promise<{ data: T; headers: Headers }>`.
  - Types (từ `@/contract/api-contract`): `PlaceOrderRequest`, `IdempotencyResult`, `Recipient`; `type RecipientInput = { name: string; phone: string; address: string }`; `type RecipientErrors = Partial<Record<keyof RecipientInput, string>>`; `type CheckoutIntent = { key: string; fingerprint: string }`.
  - `type CheckoutOutcome = { kind: 'created'; order: Order } | { kind: 'unknown'; requestID?: string } | { kind: 'processing'; retryAfterSeconds: number } | { kind: 'overload'; retryAfterSeconds: number } | { kind: 'rejected'; code: string; lineErrors: LineError[] } | { kind: 'price_changed'; lineErrors: LineError[] } | { kind: 'cart_changed' } | { kind: 'intent_conflict' } | { kind: 'invalid'; violations: FieldViolation[] } | { kind: 'unauthenticated' }`
  - `createCheckoutClient(options: JsonHttpOptions)` → `{ placeOrder(key: string, request: PlaceOrderRequest): Promise<{ order: Order; replayed: boolean }>; getAttempt(key: string): Promise<IdempotencyResult> }`; `placeOrder` dùng `timeoutMs: 15000`. `checkoutClient` mặc định với `API_BASE_URL`.
  - `CHECKOUT_INTENT_STORAGE_KEY = 'salescale.checkout-intent'`
  - `buildPlaceOrderRequest(preview: CheckoutPreview, recipient: RecipientInput): PlaceOrderRequest`
  - `intentFingerprint(request: PlaceOrderRequest): string`
  - `resolveIntent(storage: Pick<Storage, 'getItem' | 'setItem'>, request: PlaceOrderRequest, newKey?: () => string): CheckoutIntent` (mặc định `crypto.randomUUID`)
  - `clearIntent(storage: Pick<Storage, 'removeItem'>): void`
  - `validateRecipient(input: RecipientInput): RecipientErrors`
  - `classifyPlaceOrderFailure(error: unknown): CheckoutOutcome`
  - `interpretAttempt(result: IdempotencyResult): CheckoutOutcome`

- [ ] **Step 1: Viết test thất bại**
  - json-http: gửi header tùy chỉnh; `withHeaders` trả `Idempotency-Replayed`; fetcher treo quá `timeoutMs` → `ApiError` `NETWORK_ERROR`.
  - service: `placeOrder` gửi `POST /api/orders` có `Idempotency-Key`, `X-CSRF-Token`, body đúng; header `Idempotency-Replayed: true` → `replayed: true`; `getAttempt('key-12345')` gọi `GET /api/idempotency/orders/key-12345` không kèm CSRF.
  - logic:
    - `buildPlaceOrderRequest` copy đủ field từng dòng preview, `expected_total = preview.total_amount`, `currency 'VND'`, recipient đã trim.
    - `resolveIntent reuses the stored key for the same fingerprint`; fingerprint khác (đổi phone hoặc đổi dòng) → key mới và ghi đè storage; storage chứa JSON hỏng → key mới.
    - `validateRecipient`: rỗng → 3 lỗi tiếng Việt; phone 7 ký tự lỗi; name 121 ký tự lỗi; hợp lệ → `{}`.
    - `classifyPlaceOrderFailure`: status 0 → `unknown`; 500/502/503/504 → `unknown` kèm `requestID`; `INVALID_RESPONSE` → `unknown`; 409 `IDEMPOTENCY_IN_PROGRESS` → `processing` (mặc định 1s); 429 → `overload` với `retryAfterSeconds`; 409 `PRICE_CHANGED` → `price_changed`; 409 `CART_ITEM_CHANGED` → `cart_changed`; 409 `INSUFFICIENT_STOCK` → `rejected`; 409 `IDEMPOTENCY_CONFLICT` → `intent_conflict`; 422 → `invalid`; 401 → `unauthenticated`.
    - `interpretAttempt`: `response_status 201` → `created`; 409 body `SALE_QUOTA_EXCEEDED` → `rejected`.
- [ ] **Step 2: Chạy** `cd app/frontend && bun test src/test/lib/json-http.test.ts src/test/checkout` → FAIL.
- [ ] **Step 3: Implement.** Thông điệp lỗi recipient: `"Vui lòng nhập họ tên."`, `"Họ tên tối đa 120 ký tự."`, `"Số điện thoại cần 8–20 ký tự."`, `"Vui lòng nhập địa chỉ nhận hàng."`, `"Địa chỉ tối đa 500 ký tự."`. Fingerprint: `JSON.stringify` request với lines sắp theo `cart_item_id`.
- [ ] **Step 4: Chạy** lệnh Step 2 rồi `bun run typecheck && bun run lint` → PASS.
- [ ] **Step 5: Commit** `feat(frontend): add checkout transport, intent key and outcome rules`

---

### Task 8: Frontend — form người nhận, đặt đơn và trạng thái kết quả trên giỏ hàng

**Files:**
- Create: `app/frontend/src/features/checkout/checkout.hook.ts`, `components/RecipientForm.tsx`, `components/PlaceOrderSection.tsx`, `components/CheckoutOutcomeView.tsx`
- Modify: `app/frontend/src/features/cart/components/CheckoutPreviewPanel.tsx` (bỏ nút "Đặt hàng (Phase 04)" và ghi chú tạm), `app/frontend/src/features/cart/pages/CartPage.tsx`
- Test: `app/frontend/src/test/checkout/checkout-components.test.tsx`; cập nhật `src/test/cart/cart-components.test.tsx`

**Interfaces:**
- Consumes: Task 7 toàn bộ; `useCheckoutPreview` (`preview`, `acknowledged`, `run`); `cartQueryKeys`; `useNavigate`.
- Produces:
  - `usePlaceOrder(args: { preview: CheckoutPreview | null; acknowledged: boolean; onPriceChanged: () => void; onCartChanged: () => void })` → `{ recipient, setRecipient, errors, submitting, outcome: CheckoutOutcome | null, submit(): Promise<void>, retry(): Promise<void>, checkResult(): Promise<void> }`. `submit` validate → `resolveIntent(sessionStorage, …)` → `placeOrder`; `created` → `clearIntent`, invalidate query giỏ và đơn, `navigate('/orders/' + order.id)`; `rejected`/`price_changed`/`cart_changed` → `clearIntent` (kết quả đã lưu là cuối cùng cho key đó), gọi callback tương ứng; `unknown` → tự `checkResult` sau 5 giây; `checkResult` gọi `getAttempt` (404 → giữ `unknown` và cho `retry`); `retry` gửi lại **cùng** key; `intent_conflict` → `clearIntent` và yêu cầu xem trước lại.
  - `RecipientForm({ value, errors, disabled, onChange })`, `PlaceOrderSection({ state })`, `CheckoutOutcomeView({ outcome, onRetry, onCheck, retryCountdown })`.

- [ ] **Step 1: Viết test render thất bại** (cùng cách `renderToStaticMarkup` + `MemoryRouter` như `src/test/cart/cart-components.test.tsx`):
  - Panel chưa có preview hoặc chưa acknowledge → không render form người nhận và không có nút "Đặt hàng".
  - Preview đã acknowledge → có `Họ tên`, `Số điện thoại`, `Địa chỉ`, nút "Đặt hàng", dòng "Giữ hàng sau khi đặt" "5 phút".
  - `submitting` → nút "Đang đặt hàng…" bị disabled và dòng "Đang gửi yêu cầu. Vui lòng không đóng trang.".
  - `RecipientForm` hiển thị lỗi cạnh field, `aria-invalid="true"`.
  - `unknown outcome never renders a success heading`: có "Đang kiểm tra kết quả đặt đơn…", "Kết nối bị gián đoạn nên chưa nhận được kết quả. Hệ thống đang tự kiểm tra lại, bạn không cần đặt lại đơn.", nút "Kiểm tra lại" và "Thử lại"; **không** có "Đặt đơn thành công".
  - `processing` → "Yêu cầu trước đó đang được xử lý, vui lòng đợi."; `unknown` kèm `requestID` từ 5xx → "Hệ thống tạm thời không xử lý được" và "Mã yêu cầu: req_1".
  - `overload` (`retryAfterSeconds 15`) → "Hệ thống đang quá tải", "Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên.", nút "Thử lại sau 00:15" disabled.
  - `rejected` → "Chưa đặt được đơn hàng" + "Chưa có đơn nào được tạo. Giỏ hàng của bạn vẫn giữ nguyên."; `price_changed` → "Giá đã thay đổi" + nút "Tôi đồng ý với giá mới".
- [ ] **Step 2: Chạy** `bun test src/test/checkout src/test/cart` → FAIL.
- [ ] **Step 3: Implement** hook và component; trong `CartPage`, `onPriceChanged` = `checkout.run()` (preview mới đi qua `detectPreviewChanges` nên khách phải xác nhận lại), `onCartChanged` = `cart.reload()`; line errors của `rejected` đưa vào `cart.applyLineErrors` để hiện dưới đúng dòng.
- [ ] **Step 4: Chạy** `bun run typecheck && bun run lint && bun test && bun run build` → PASS.
- [ ] **Step 5: Commit** `feat(frontend): place orders from the cart preview with honest outcomes`

---

### Task 9: Live checkpoint và cập nhật tài liệu phase

**Files:**
- Modify: `docs/plans/phases/phase-04-checkout.md`, `docs/plans/phases/phase-05-payment-lifecycle.md`, `docs/technical/traceability.md` (chỉ dòng TEST-01–10/25 nếu có cột bằng chứng)

- [ ] **Step 1: Chạy cổng tự động**: backend `gofmt -l . && go build ./... && go vet ./... && go test ./... && go test ./test -race -count=1` (có `TEST_DATABASE_URL`); frontend `bun run typecheck && bun run lint && bun test && bun run build`. Ghi số pass/skip; nếu còn test skip thì ghi rõ, không đánh dấu đạt.
- [ ] **Step 2: Hành trình trình duyệt** trên DB development (`go run ./cmd/migrate up`, `ALLOW_SEED=true go run ./cmd/seed`, `air`, `bun run dev`): customer thêm 1 dòng regular + 1 dòng flash sale → Xem trước → nhập người nhận → Đặt hàng → tới `/orders/:id` → "Thanh toán" → admin `/admin/payment-simulator` gửi success → đơn `confirmed`, dòng giỏ đã đặt biến mất. Lặp ở viewport 390×844. Thêm một lần tắt backend lúc bấm "Đặt hàng" → thấy trạng thái chưa rõ, bật lại → "Kiểm tra lại" ra đúng đơn, không có đơn thứ hai.
- [ ] **Step 3: Oracle SQL** sau hành trình: chạy các truy vấn INV-02/03/08/09/10/11 trong `lifecycleRequireInvariants` trên DB development → 0 dòng vi phạm.
- [ ] **Step 4: Cập nhật checklist**: đánh dấu các mục Phase 04 có bằng chứng (tên test + lệnh), cập nhật Phase 05 Step 3 mục FE "hành trình khách từ xem hàng tới kết quả đơn". Mục metrics Prometheus của Phase 05 vẫn để trống: checkout chỉ có `CheckoutCounters` + log, chưa expose endpoint.
- [ ] **Step 5: Commit** `docs: record phase 04 checkout checkpoint`

---

## Ghi chú cho người triển khai (các quyết định trong plan)

- Dùng `ROLLBACK TO SAVEPOINT` để vừa lưu outcome từ chối vừa xóa sạch tác động của bước khóa/quote, tránh để lại row usage rỗng.
- `IDEMPOTENCY_IN_PROGRESS` chỉ xảy ra khi chờ khóa idempotency quá 2 giây; thông thường request trùng key sẽ chờ rồi replay.
- `422` và lỗi 5xx không được lưu làm outcome; `409` nghiệp vụ (`PRICE_CHANGED`, `CART_ITEM_CHANGED`, mã line) được lưu và replay.
- Metrics là bộ đếm trong tiến trình + log; việc expose Prometheus để lại cho Phase 05 Step 3/Phase 07 vì dependency chưa được chốt.
- Luồng đặt đơn nằm trong panel xem trước của trang giỏ (theo mục 6 của design), copy và các trạng thái lấy từ `Preview*.html` và `OrderResult*.html`; thành công chuyển tới `/orders/:orderId` có sẵn chứ không làm trang kết quả riêng.
