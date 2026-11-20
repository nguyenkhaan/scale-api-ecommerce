# Phase 03 — Flash sale và giỏ hàng (Lân)

**Phụ thuộc:** 02. **Mục tiêu:** `FR-014–020`, `INV-09/10`, `TEST-09/10/20`.

**Trạng thái kiểm tra 21/11/2026:** FE/BE đã được kiểm tra lại sau đợt chuẩn hóa UI và accessibility review. Frontend có 97 test pass; backend quality gates pass, trong đó 55 test pass và 66 test PostgreSQL integration bị skip vì checkout hiện không cấu hình `TEST_DATABASE_URL`. Live API trên database development, browser smoke bằng Chrome headless và các oracle cart/preview đã đạt trong phạm vi Phase 03.

## Step 1 — Chương trình flash sale

- [x] BE: migration campaign/item/usage; tạo nháp, lên lịch, hủy lịch, dừng, chống chồng SKU và khóa cấu hình.
  - Bằng chứng: schedule chạy trong một transaction, khóa row campaign và `pg_advisory_xact_lock` theo từng SKU, kiểm lại giá/SKU rồi mới `draft → scheduled`; sửa nháp khóa row campaign; hủy lịch/dừng là câu `UPDATE` có điều kiện. Campaign `stopped` vẫn tính là chồng thời gian (PRD §7.1).
- [x] FE: trang sale công khai và form admin tạo/lên lịch/dừng theo contract.
  - Bằng chứng: chọn SKU từ danh sách sản phẩm thay vì nhập UUID; hành động theo trạng thái (nháp → lên lịch, sắp diễn ra → hủy lịch, đang chạy → dừng); lỗi hiển thị cạnh trường.
- [x] Kiểm: giờ database, sale hết hạn/dừng, schedule đồng thời không chồng; `TEST-20`.
  - Bằng chứng: `test/flash_sale_concurrency_test.go` (8 nháp cùng SKU lên lịch đồng thời → đúng 1), `flash_sale_lifecycle_test.go`; e2e: sale đã dừng → preview `SALE_NOT_ACTIVE`, không tự đổi sang giá thường.

## Step 2 — Giỏ hàng nhiều SKU

- [x] BE: migration cart/items; một cart active/user, tối đa 20 SKU, gộp SKU, version item; không giữ tồn.
  - Bằng chứng: khóa row cart trước khi đếm; `expected_version` cập nhật có điều kiện (`CART_ITEM_CHANGED`); thêm cùng SKU đồng thời dùng `ON CONFLICT`; SKU của sản phẩm/danh mục bị ẩn trả `ITEM_UNAVAILABLE` (BR-03); migration `00008` chỉnh index.
- [x] FE: chọn SKU và số lượng, thêm/sửa/xóa, chọn tập dòng thường + sale.
  - Bằng chứng: trang `/products` thêm hàng giá thường (bản tối giản, có thể thay bằng UI catalog của Phase 02), trang Flash Sale thêm hàng sale; giỏ hiển thị tên, mã SKU, nguồn giá, giá tham khảo; xóa dòng xử lý đúng `204`.
- [x] Kiểm: auth/owner, trùng SKU, giới hạn dòng, sửa version; giỏ không làm đổi inventory.
  - Bằng chứng: `test/cart_concurrency_test.go`, `cart_availability_test.go`, `cart_api_test.go`; e2e đối chiếu `inventories`, quota và usage trước/sau không đổi.

## Step 3 — Preview cơ bản

- [x] BE: tính giá/quota/giới hạn mua hiện thời cho tập dòng; trả nguồn giá và lỗi theo dòng, không reserve.
  - Bằng chứng: preview đọc trong transaction read-only với một `now()`, số query không phụ thuộc số dòng; mọi dòng lỗi trả trong `line_errors`, `code` cấp cao nhất là mã nghiêm trọng nhất (`test/cart_preview_test.go`, `cart_preview_sale_rules_test.go`; TEST-09/10).
- [x] FE: hiển thị tổng, giá/điều kiện mới và yêu cầu xác nhận lại.
  - Bằng chứng: lỗi hiện dưới đúng dòng; nút "Chuyển sang giá thường" chỉ khi khách chủ động bấm; giá đổi so với lần preview trước phải bấm "Tôi đã xem giá mới".
- [x] Live checkpoint: preview đúng cả dòng thường/sale; test quota/limit và build/browser đạt. Chi tiết đặt đơn sang phase 04.
  - Đạt: live API, build, frontend tests, browser smoke và oracle dữ liệu đã chạy lại sau đợt sửa UI. Các test PostgreSQL integration tự động vẫn cần `TEST_DATABASE_URL` riêng để chạy đầy đủ.
- [ ] Dedicated PostgreSQL integration checkpoint: chờ cấp `TEST_DATABASE_URL`; không coi các test bị skip là pass.

## Thay đổi hợp đồng

OpenAPI: thêm mã `SALE_SCHEDULE_OVERLAP`, `CART_ITEM_CHANGED`; `FlashSale.items` có `sku_code`, `product_name`, `regular_price`; `CartItem` có `product_name`, `sku_code`, `reference_unit_price`; `UpsertCartItemRequest` có `expected_version`, `price_source` mặc định `regular`; mô tả lỗi theo dòng của preview.

## Checkpoint FE–BE sau chuẩn hóa giao diện — 21/11/2026

### Phạm vi và thay đổi tối thiểu

- FE giữ nguyên phạm vi Phase 03: catalog public, flash sale public/admin, cart nhiều SKU và checkout preview; nút đặt đơn vẫn khóa để chờ Phase 04.
- API client đã chuẩn hóa về một root `/api`; Vite proxy và direct-origin `VITE_API_BASE_URL=http://localhost:4000` đều được kiểm tra.
- Backend chỉ có thay đổi tương thích CORS cho `PATCH`/`DELETE` và chuẩn hóa whitespace `router.go`; không đổi business logic sale/cart/preview.
- Shell responsive được chỉnh để navigation customer tự wrap trên màn hình hẹp; browser smoke trước đó phát hiện overflow 403px trên viewport 390px, sau sửa còn `overflow=false`.
- Accessibility follow-up đã bổ sung URL-backed search IDs riêng, focus trap/restore cho dialog, focus vào field lỗi đầu tiên và recovery action cho admin auth/CSRF failures.

### Automated gates

| Khu vực | Lệnh | Kết quả |
|---|---|---|
| Backend format/build/test/vet | `cd app/backend && gofmt -l . && go build ./... && go test ./... && go vet ./...` | Pass; `gofmt` sạch, build/vet sạch, package test pass |
| Backend verbose suite | `cd app/backend && go test -v ./test` | 55 pass, 66 skip; integration skip nêu rõ `set TEST_DATABASE_URL` |
| Frontend | `cd app/frontend && bun run typecheck && bun run lint && bun test && bun run build` | Pass; 97 tests, 281 assertions, lint/build pass |
| Legacy API route audit | `rg -n '/api/v1' app/frontend` | Không có kết quả |
| Migration/fixture | `go run ./cmd/migrate status`, `go run ./cmd/migrate up`, `ALLOW_SEED=true go run ./cmd/seed` | Migration `00009` đã áp dụng; fixture development nạp thành công |

### Live API scenario

API chạy tại `http://localhost:4000`; frontend Vite chạy tại `http://localhost:5173` và `/api` proxy trả dữ liệu từ cùng API.

- Health, public products, public flash sales, customer/admin login và CORS preflight cho `POST`/CSRF đều đạt.
- Admin update/schedule/stop campaign đạt; public list chỉ trả campaign đang active.
- Customer thêm một SKU thường và một SKU flash sale; preview trả `VND`, total `900000`, đúng `price_source` và line subtotal.
- Sau khi dừng sale, preview trả `409 SALE_NOT_ACTIVE` theo đúng dòng; chuyển dòng sang giá thường thành công và preview recover với total `1000000`.
- Thay đổi giá thường `500000 → 510000` giữa hai lần preview làm total đổi `1000000 → 1010000`; FE `detectPreviewChanges` yêu cầu acknowledgement trước khi cho phép bước tiếp theo.
- Campaign quota `1` với request `2` trả `409 SALE_QUOTA_EXCEEDED`, `available_quantity=1`; request vượt purchase limit bị backend từ chối `422`.

### Browser smoke

Chrome `150.0.7871.100` headless được chạy bằng DevTools Protocol với session sạch ở cả viewport `1280×900` và `390×844`.

- Anonymous: `/`, `/products`, catalog search, `/flash-sales`, `/login`, `/register`; heading, form controls, CTA và navigation render đúng, không horizontal overflow.
- Customer: login development account, `/cart`, chọn cart lines, click `Xem trước thanh toán`; preview lines/tổng tiền hiển thị và mobile overflow không còn.
- Admin: login development account, `/admin/flash-sales`; active navigation, campaign status/action, create dialog, input fields và invalid-submit field errors render đúng; dialog đưa focus vào nút đóng, form invalid đưa focus vào `campaign-name`.
- Kiểm tra link trong customer/admin shell không có dead route trong phạm vi Phase 03; focusable elements có focus ring class; cookie session/CSRF và direct-origin CORS đã được kiểm tra qua request thật.

### Data oracles và giới hạn còn lại

- Snapshot trước/sau preview có cùng digest `da8e91536ecbce078f2d8c9b169568c6`; `inventory.reserved=0`, `quota_reserved=0`, `quota_sold=0`, customer usage không đổi, `orders=0`, `inventory_reservations=0`.
- Từ chối `SALE_NOT_ACTIVE` và `SALE_QUOTA_EXCEEDED` không làm thay đổi cart hoặc reservation state; cart test data đã được dọn sau checkpoint.
- Repository hiện không có runner/browser collection được track để tái lập claim lịch sử `35/35` API e2e; checkpoint này ghi nhận live HTTP scenario tương đương và không nâng claim lịch sử thành một command mới.
- Muốn đóng hoàn toàn risk tự động của backend cần cấp một database test riêng qua `TEST_DATABASE_URL`, rồi chạy lại 66 test đang skip, đặc biệt các test `TEST-09`, `TEST-10`, `TEST-20`. Đặt đơn và reservation vẫn thuộc Phase 04.

### Traceability

| Requirement | Evidence checkpoint |
|---|---|
| `FR-014–020` | Admin/public campaign lifecycle, public sale card, cart add/update/remove, regular + flash preview, line recovery |
| `INV-09` | Live quota rejection and before/after quota oracle; backend quota tests remain listed for dedicated PostgreSQL suite |
| `INV-10` | Live customer usage oracle unchanged during preview; backend purchase-limit validation and preview rules |
| `TEST-09` | Mixed regular/flash preview, total/line source, read-only digest |
| `TEST-10` | Sale quota/purchase-limit rejection and line error recovery |
| `TEST-20` | Stop transition, `SALE_NOT_ACTIVE`, no fallback to regular price |
