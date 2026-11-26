# Design: SaleScale UI v2 — thiết kế lại bộ màn hình tĩnh theo phong cách Tiki

**Trạng thái:** Đã chốt hướng qua bản thử và áp dụng đủ 76 màn hình tại `docs/ui-design/v2` (26/11/2026). Chờ nhóm duyệt bản xem trước.

## 1. Ý định

Bộ HTML mẫu tại `docs/ui-design/pages` (76 màn hình) đúng nghiệp vụ nhưng trông đơn giản, chưa ra dáng một trang thương mại điện tử hiện đại. UI v2 thiết kế lại toàn bộ bộ màn hình theo phong cách **Tiki: dày thông tin**. Mục tiêu là để nhóm xem trước rồi mới chuyển vào frontend React.

## 2. Mục tiêu, ngoài phạm vi, chống mục tiêu

**Mục tiêu**
- Tái hiện đủ 76 màn hình và trạng thái của bộ cũ, cùng tên file, tại `docs/ui-design/v2/pages`.
- Giao diện khách hàng gần Tiki: header trắng với ô tìm kiếm lớn, khối flash sale "Giá sốc hôm nay", card sản phẩm dày thông tin, trang chi tiết 3 cột, giỏ hàng chia nhóm.
- Giao diện quản trị dạng back-office gọn: thanh menu trái, chỉ số dạng chữ, bảng chỉ dùng đường kẻ.
- Mỗi màn hình chạy được độc lập bằng `file://`, không cần JavaScript.

**Ngoài phạm vi**
- Chưa sửa frontend React hay backend trong đợt này.
- Không làm giao hàng, người bán, voucher, hỏi đáp hay hệ thống review.

**Chống mục tiêu**
- Không làm mờ các trạng thái nghiệp vụ khó cho đẹp: giỏ chưa giữ hàng, giờ hệ thống, giá cuối cùng xác nhận ở bước xem trước, kết quả chưa rõ, quá tải, giá thay đổi, quá hạn, đối soát. Mọi câu nhắc này phải còn nguyên ý.
- Không hiển thị kết quả chưa rõ như thể đã thành công.

## 3. Các hướng đã cân nhắc

- **A. Làm thử 4 trang, chốt hướng rồi mới làm hết (đã chọn).** Nếu hướng sai thì chỉ phải sửa 4 trang.
- B. Làm thẳng cả 76 trang. Bị loại vì đổi hướng sẽ phải sửa lại toàn bộ.
- C. Làm bộ component trước. Bị loại vì khó đánh giá "có giống Tiki không".

## 4. Quyết định đã chốt

- **Màu:** xanh biển `--brand: #0b6cf0` (đã so với bản cam và chọn xanh). Giá và flash sale dùng đỏ `--price/--sale: #ff424e`. Nền `#f5f5fa`, card trắng bo góc 8px, không đổ bóng (chỉ có bóng khi hover card).
- **Font:** Be Vietnam Pro (400–800, có in nghiêng 800 cho tiêu đề flash sale).
- **Điểm nhận diện:** thanh suất flash sale gồm thanh 6px và nhãn dưới ("Đã bán 32/50 suất", "Sắp hết · còn 2 suất", "Đã hết suất"). Dùng ở trang chủ, trang chi tiết và quản trị.
- **Ảnh sản phẩm:** placeholder SVG phẳng trong `v2/assets/img`.
- **Màn hình `-mobile`:** hiển thị trang tương ứng trong khung điện thoại 390×844 (`mobileFrame`). Bản thân mọi trang đều responsive.
- **Dữ liệu card kiểu Tiki** (thương hiệu, sao, số đã bán, nhiều ảnh) cần mở rộng backend tối thiểu ở bước sau:
  - `products.brand`;
  - `products.rating_avg` / `rating_count` (dữ liệu seed mô phỏng);
  - bảng `product_images`;
  - số đã bán tính từ `order_items` của đơn `confirmed`.

## 5. Cấu trúc kỹ thuật

- `v2/assets/salescale.css`: token và component dùng chung.
- `v2/assets/<nhóm>.css`: CSS riêng của từng nhóm màn hình.
- `v2/src/lib.mjs`: helper dùng chung (`head`, `header`, `footer`, `customerPage`, `adminShell`, `mobileFrame`, `productCard`, `flashItem`, `quotaBar`, `countdown`, `steps`, dữ liệu mẫu `P`, icon `ic`).
- `v2/src/pages/NN-<nhóm>.mjs`: mỗi nhóm màn hình là một module.
- Build bằng `node docs/ui-design/v2/src/build.mjs`. Script sinh `v2/pages/*.html` và `index.html` (mục lục).

## 6. Phần để lại cho người triển khai

- Chuyển token sang `app/frontend/src/styles/theme.css` và tách component React theo các lớp CSS v2.
- Mở rộng backend theo mục 4, kèm cập nhật PRD, OpenAPI và traceability (ADR-003).
