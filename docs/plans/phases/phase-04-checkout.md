# Phase 04 — Checkout nhiều SKU nguyên tử (An)

**Phụ thuộc:** 03. **Mục tiêu:** `FR-021–029`, `INV-01–05/09–11`, `TEST-01–10/25`. Đây là lát cắt rủi ro cao nhất.

**Rà soát mã nguồn 26/11/2026 (nhánh `dev`):** `9/9`. Live browser checkpoint đã qua sau khi phát hiện DB thiếu `payment_sessions`, `payment_events` và `payment_reconciliations` dù goose ghi nhận `00010`; migration repair `00011`/`00012` đã khôi phục các bảng thiếu trên DB dev. Playwright đặt đơn nhận 201, mở trang đơn và hủy đơn thành công ở viewport 390×844.

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
- [x] Live checkpoint: `TEST-01–10/25`, oracle INV-01–05/09–11 và HTTP contract qua trước payment/benchmark.
  - Automated checkout/oracle gates đã qua ở lượt backend trước; lượt này chạy migration repair tests, kiểm tra đủ ba bảng payment trên DB dev, và Playwright Chrome 390×844 đặt đơn HTTP 201 rồi hủy đơn thành công. Không còn lỗi schema `payment_sessions`/payment lifecycle trong luồng checkout đã kiểm tra.
