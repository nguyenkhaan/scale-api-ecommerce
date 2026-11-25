# Phase 04 — Checkout nhiều SKU nguyên tử (An)

**Phụ thuộc:** 03. **Mục tiêu:** `FR-021–029`, `INV-01–05/09–11`, `TEST-01–10/25`. Đây là lát cắt rủi ro cao nhất.

**Rà soát mã nguồn 25/11/2026 (nhánh `dev`):** `8/9` — code checkout và các cổng tự động đã hoàn tất; live browser checkpoint còn mở. Backend có checkout nhiều SKU nguyên tử, idempotency lookup/replay, reservation/payment handoff và test contention; frontend đã chuyển UI `Preview*.html`/`OrderResult*.html` thành React trong cart flow. Live `DATABASE_URL` hiện có schema drift: goose báo `00010` applied nhưng thiếu bảng `payment_sessions`, nên `POST /api/orders` chỉ đạt tới order read-model rồi trả 500; không tự ý reset hoặc sửa migration trong checkpoint này.

## Step 1 — Order và reservation

- [x] BE: migration order/items/reservations/idempotency; khóa SKU theo thứ tự ổn định; giữ đủ hoặc rollback toàn bộ.
  - Bằng chứng: `internal/module/order/checkout_service.go`, `cart/checkout_quote.go`; `test/order_checkout_concurrency_test.go` và full backend race gate.
- [x] FE: form người nhận, xác nhận tổng/nguồn giá; hiển thị lỗi theo SKU và giữ giỏ khi bị từ chối.
  - Bằng chứng: `features/checkout/components/RecipientForm.tsx`, `PlaceOrderSection.tsx`, `CheckoutOutcomeView.tsx`; `src/test/checkout/checkout-components.test.tsx`.
- [x] Kiểm tự động: `TEST-01–04`, `08–10`, oracle INV sau cạnh tranh, không partial order.
  - Bằng chứng: `go test ./test -race -count=1 -timeout 30m` → pass (`906.658s`); full normal → pass (`862.806s`); exact `go test ./test -run 'TestCheckout' -race -count=1 -v` → 22 pass (`199.781s`).

## Step 2 — Retry an toàn

- [x] BE: canonical hash, unique user/operation/key, lưu business rejection và response ổn định, endpoint tra cứu kết quả.
  - Bằng chứng: `checkout_hash.go`, `checkout_attempt.go`, `POST /api/orders`, `GET /api/idempotency/orders/{idempotencyKey}`.
- [x] FE: giữ key cho một ý định qua timeout/reload; tra cứu hoặc retry cùng key; tạo key mới sau khi sửa lựa chọn.
  - Bằng chứng: `checkout.logic.ts`, `checkout.hook.ts`; 12 transport/service/logic tests pass.
- [x] Kiểm: `TEST-05–07`, 2 API instance, mất response sau commit, cùng key khác payload.
  - Bằng chứng: idempotency/replay/conflict tests trong `app/backend/test`, full race gate pass.

## Step 3 — Cổng checkout

- [x] BE: snapshot giá/người nhận, hạn 5 phút cố định, counter new/replay/rejection/error.
  - Bằng chứng: order snapshot/reservation code và `CheckoutCounters`; Prometheus endpoint chưa được chốt nên vẫn để Phase 05 Step 3.
- [x] FE: trạng thái đơn chính thức hiện mã đơn/hạn; trạng thái unknown không hiển thị thành công giả.
  - Bằng chứng: điều hướng tới `/orders/:id` sau 201 và `CheckoutOutcomeView`/tests cho unknown, processing, overload, rejected, price-changed.
- [ ] Live checkpoint: `TEST-01–10/25`, oracle INV-01–05/09–11 và HTTP contract qua trước payment/benchmark.
  - Automated gates đã qua; live preview qua `DATABASE_URL`, nhưng live place-order bị chặn bởi schema drift (`payment_sessions` thiếu dù migration 00010 đã applied). Browser viewport 390×844 chưa chạy vì môi trường không có Chrome/Chromium/Playwright runner.
