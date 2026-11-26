# Phase 02 — Hàng hóa và tồn kho (An)

**Phụ thuộc:** 01. **Mục tiêu:** `FR-006–013`, `INV-01/02/12`, `TEST-21`. Category admin P1 được để phase 06 nếu chưa cần UI.

**Rà soát mã nguồn 26/11/2026 (nhánh `dev`):** `8/9` checklist items; catalog checkpoint đã qua, UI tồn kho ở Step 2 còn lại. Catalog frontend có danh sách, tìm kiếm, lọc danh mục, cursor pagination, chi tiết sản phẩm và form quản trị Product/SKU. Admin có thể nhập tồn ban đầu khi tạo SKU; màn hình điều chỉnh tồn và ledger chưa triển khai.

## Step 1 — Catalog

- [x] BE: migration categories/products/SKUs; admin tạo/sửa/ẩn, kiểm SKU duy nhất, không hard-delete lịch sử.
- [x] FE: danh sách/chi tiết sản phẩm có SKU, tìm/lọc/phân trang; form quản trị sản phẩm/SKU tối giản.
  - `/products` có tìm kiếm, lọc danh mục và cursor “Tải thêm”; `/products/:productId` có chi tiết, SKU, giá và thêm giỏ. `/admin/products` và `/admin/products/:productId` hỗ trợ tạo/sửa sản phẩm, tạo SKU với tồn ban đầu và sửa giá/trạng thái SKU.
- [x] Kiểm: API contract và browser với sản phẩm active/inactive, dữ liệu phân trang.
  - API: `test/catalog_api_test.go` kiểm active/inactive, search, pagination và CSRF admin. Browser: `app/frontend/e2e/catalog-checkout.spec.ts` kiểm sản phẩm active, lọc danh mục, chi tiết và ẩn sản phẩm inactive; Playwright chạy Chrome ở viewport 390×844.

## Step 2 — Tồn kho

- [x] BE: migration inventory/adjustment ledger; nhập và điều chỉnh trong transaction, không giảm dưới reserved.
- [ ] FE: form điều chỉnh có lý do; hiển thị on_hand, reserved, available cho admin.
  - Chưa triển khai. BE đã có `GET /admin/inventory`, `GET/POST /admin/inventory/:skuId/adjustments`.
- [x] Kiểm: `TEST-21`, kiểm constraint INV-01 và ledger INV-12 trên PostgreSQL thật.
  - Bằng chứng: `TestInventoryServiceAdjustsAtomicallyAndPaginatesLedger` (giảm dưới reserved → `ErrBelowReserved`), `TestInventoryMigrationEnforcesCountersAndLedger`, `inventory_controller_test.go` (`409 INVENTORY_BELOW_RESERVED`). Test PostgreSQL cần `TEST_DATABASE_URL`.

## Step 3 — Tích hợp

- [x] BE: seed có nhiều sản phẩm và biến thể, API đọc không lộ số liệu nội bộ sai quyền.
- [x] FE: nối form/list với API thật, trạng thái lỗi/rỗng và cập nhật sau sửa.
  - Catalog và form Product/SKU dùng API thật, có loading/error/empty/not-found state; query được làm mới sau khi lưu sản phẩm hoặc SKU.
- [x] Checkpoint: admin tạo SKU + tồn; khách xem đúng hàng; build/test/contract đạt.
  - Form tạo sản phẩm/SKU nhận `initial_on_hand`; browser đã xác nhận khách xem active product và không truy cập được inactive product. `bun run build`, 199 frontend tests, lint, typecheck, backend catalog/migration tests và Playwright E2E đều pass. UI điều chỉnh tồn và ledger vẫn là việc còn lại ở Step 2.
