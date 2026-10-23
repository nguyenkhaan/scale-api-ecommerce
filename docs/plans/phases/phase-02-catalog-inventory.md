# Phase 02 — Hàng hóa và tồn kho

**Phụ thuộc:** 01. **Mục tiêu:** `FR-006–013`, `INV-01/02/12`, `TEST-21`. Category admin P1 được để phase 06 nếu chưa cần UI.

## Step 1 — Catalog

- [ ] BE: migration categories/products/SKUs; admin tạo/sửa/ẩn, kiểm SKU duy nhất, không hard-delete lịch sử.
- [ ] FE: danh sách/chi tiết sản phẩm có SKU, tìm/lọc/phân trang; form quản trị sản phẩm/SKU tối giản.
- [ ] Kiểm: API contract và browser với sản phẩm active/inactive, dữ liệu phân trang.

## Step 2 — Tồn kho

- [ ] BE: migration inventory/adjustment ledger; nhập và điều chỉnh trong transaction, không giảm dưới reserved.
- [ ] FE: form điều chỉnh có lý do; hiển thị on_hand, reserved, available cho admin.
- [ ] Kiểm: `TEST-21`, kiểm constraint INV-01 và ledger INV-12 trên PostgreSQL thật.

## Step 3 — Tích hợp

- [ ] BE: seed có nhiều sản phẩm và biến thể, API đọc không lộ số liệu nội bộ sai quyền.
- [ ] FE: nối form/list với API thật, trạng thái lỗi/rỗng và cập nhật sau sửa.
- [ ] Checkpoint: admin tạo SKU + tồn; khách xem đúng hàng; build/test/contract đạt.
