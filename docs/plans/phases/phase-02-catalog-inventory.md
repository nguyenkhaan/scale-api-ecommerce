# Phase 02 — Hàng hóa và tồn kho (An)

**Phụ thuộc:** 01. **Mục tiêu:** `FR-006–013`, `INV-01/02/12`, `TEST-21`. Category admin P1 được để phase 06 nếu chưa cần UI.

**Rà soát mã nguồn 25/11/2026 (nhánh `dev`):** `4/9`. Backend catalog/inventory và test đã xong; **frontend Phase 02 chưa triển khai**. Hiện chỉ có trang `/products` tối giản (tìm theo tên, tải thêm) làm trong Phase 03. Chưa có trang chi tiết sản phẩm, form quản trị sản phẩm/SKU hoặc màn hình tồn kho (`features/product/` đang rỗng).

## Step 1 — Catalog

- [x] BE: migration categories/products/SKUs; admin tạo/sửa/ẩn, kiểm SKU duy nhất, không hard-delete lịch sử.
- [ ] FE: danh sách/chi tiết sản phẩm có SKU, tìm/lọc/phân trang; form quản trị sản phẩm/SKU tối giản.
  - Đã có: danh sách `/products` với tìm theo tên và phân trang cursor. Còn thiếu: trang chi tiết, lọc theo danh mục, form admin sản phẩm/SKU (BE đã có `POST/PATCH /admin/products`, `/admin/skus`).
- [ ] Kiểm: API contract và browser với sản phẩm active/inactive, dữ liệu phân trang.
  - Đã có phần API: `test/catalog_api_test.go` (ẩn sản phẩm inactive, tìm kiếm, phân trang, CSRF admin). Còn thiếu: kiểm bằng trình duyệt cho sản phẩm active/inactive.

## Step 2 — Tồn kho

- [x] BE: migration inventory/adjustment ledger; nhập và điều chỉnh trong transaction, không giảm dưới reserved.
- [ ] FE: form điều chỉnh có lý do; hiển thị on_hand, reserved, available cho admin.
  - Chưa triển khai. BE đã có `GET /admin/inventory`, `GET/POST /admin/inventory/:skuId/adjustments`.
- [x] Kiểm: `TEST-21`, kiểm constraint INV-01 và ledger INV-12 trên PostgreSQL thật.
  - Bằng chứng: `TestInventoryServiceAdjustsAtomicallyAndPaginatesLedger` (giảm dưới reserved → `ErrBelowReserved`), `TestInventoryMigrationEnforcesCountersAndLedger`, `inventory_controller_test.go` (`409 INVENTORY_BELOW_RESERVED`). Test PostgreSQL cần `TEST_DATABASE_URL`.

## Step 3 — Tích hợp

- [x] BE: seed có nhiều sản phẩm và biến thể, API đọc không lộ số liệu nội bộ sai quyền.
- [ ] FE: nối form/list với API thật, trạng thái lỗi/rỗng và cập nhật sau sửa.
  - Chưa triển khai (phụ thuộc form admin sản phẩm/tồn kho ở trên).
- [ ] Checkpoint: admin tạo SKU + tồn; khách xem đúng hàng; build/test/contract đạt.
  - Chưa đạt: admin chỉ tạo SKU/tồn được qua API hoặc seed, chưa có UI.
