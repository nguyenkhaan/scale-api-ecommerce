# Phase 03 — Flash sale và giỏ hàng

**Phụ thuộc:** 02. **Mục tiêu:** `FR-014–020`, `INV-09/10`, `TEST-09/10/20`.

## Step 1 — Chương trình flash sale

- [ ] BE: migration campaign/item/usage; tạo nháp, lên lịch, hủy lịch, dừng, chống chồng SKU và khóa cấu hình.
- [ ] FE: trang sale công khai và form admin tạo/lên lịch/dừng theo contract.
- [ ] Kiểm: giờ database, sale hết hạn/dừng, schedule đồng thời không chồng; `TEST-20`.

## Step 2 — Giỏ hàng nhiều SKU

- [ ] BE: migration cart/items; một cart active/user, tối đa 20 SKU, gộp SKU, version item; không giữ tồn.
- [ ] FE: chọn SKU và số lượng, thêm/sửa/xóa, chọn tập dòng thường + sale.
- [ ] Kiểm: auth/owner, trùng SKU, giới hạn dòng, sửa version; giỏ không làm đổi inventory.

## Step 3 — Preview cơ bản

- [ ] BE: tính giá/quota/giới hạn mua hiện thời cho tập dòng; trả nguồn giá và lỗi theo dòng, không reserve.
- [ ] FE: hiển thị tổng, giá/điều kiện mới và yêu cầu xác nhận lại.
- [ ] Checkpoint: preview đúng cả dòng thường/sale; test quota/limit và build/browser đạt. Chi tiết đặt đơn sang phase 04.
