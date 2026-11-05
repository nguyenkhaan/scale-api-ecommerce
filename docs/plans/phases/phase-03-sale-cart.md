# Phase 03 — Flash sale và giỏ hàng (Lân)

**Phụ thuộc:** 02. **Mục tiêu:** `FR-014–020`, `INV-09/10`, `TEST-09/10/20`.

**Trạng thái kiểm tra 05/11/2026:** BE và FE triển khai xong trên nhánh `feat/phase-03-sale-cart`; backend 123 test (có PostgreSQL), frontend 110 test, kịch bản e2e qua API 35/35 bước đạt. Còn chạy lại luồng trên trình duyệt sau đợt sửa lỗi.

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
- [ ] Checkpoint: preview đúng cả dòng thường/sale; test quota/limit và build/browser đạt. Chi tiết đặt đơn sang phase 04.
  - Đạt: build, test hai phía và e2e qua API. Còn lại: chạy lại hành trình trên trình duyệt sau đợt sửa lỗi.

## Thay đổi hợp đồng

OpenAPI: thêm mã `SALE_SCHEDULE_OVERLAP`, `CART_ITEM_CHANGED`; `FlashSale.items` có `sku_code`, `product_name`, `regular_price`; `CartItem` có `product_name`, `sku_code`, `reference_unit_price`; `UpsertCartItemRequest` có `expected_version`, `price_source` mặc định `regular`; mô tả lỗi theo dòng của preview.
