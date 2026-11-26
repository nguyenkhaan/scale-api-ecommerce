# Design: Bổ sung dữ liệu cho giao diện v2

**Ngày:** 26/11/2026
**Trạng thái:** Đã chốt 3 câu hỏi mở với nhóm (26/11/2026). Chưa sửa code backend.
**Liên quan:**
- Thiết kế giao diện: `docs/superpowers/specs/2026-11-26-ui-v2-redesign-design.md`
- Màn hình: `docs/ui-design/v2/pages`
- Model hiện tại: `app/backend/internal/model`
- Migration hiện tại: `app/backend/internal/migration/sql`, mới nhất là `00012`

## 1. Mục đích

Giao diện v2 hiển thị một số thông tin mà database hiện chưa có: thương hiệu, sao đánh giá, số đã bán, nhiều ảnh cho một sản phẩm, bảng thông số, icon danh mục, ảnh sản phẩm trong lịch sử đơn và tên hiển thị của người dùng.

Tài liệu này liệt kê:
- trường nào **cần thêm**;
- trường nào **chỉ cần tính từ dữ liệu sẵn có**;
- trường nào **không làm**.

Nguyên tắc chung là mở rộng ở mức tối thiểu, đúng quyết định "mở rộng backend tối thiểu để có giao diện giống Tiki".

## 2. Nguyên tắc

1. **Không đụng luồng nghiệp vụ lõi.** Giữ suất, giỏ hàng, đặt đơn, idempotency và thanh toán không đổi. Mọi trường mới chỉ phục vụ hiển thị.
2. **Không thêm cột cập nhật trong giao dịch đặt đơn hoặc thanh toán.** Nếu cập nhật một dòng `products` mỗi lần có đơn thành công, các sản phẩm hot sẽ tranh khóa đúng lúc flash sale cao điểm. Đây là điều đồ án cần tránh.
3. **Cái gì tính được thì không lưu.** Các thông tin sau là giá trị suy ra:
   - trạng thái "đang diễn ra" hay "đã kết thúc";
   - "giảm đến 45%";
   - "đã bán 78/130 suất";
   - "đang giữ";
   - số sản phẩm trong danh mục.
4. **Đơn hàng là bản chụp.** `order_items` đã lưu tên sản phẩm, mã SKU và thuộc tính tại thời điểm đặt đơn. Ảnh sản phẩm của đơn cũng nên được lưu theo cách đó.
5. **Thứ tự cập nhật theo ADR-003:** PRD và traceability trước, sau đó `data-model.md`, OpenAPI, migration, code và test.

## 3. Tổng hợp

| # | Giao diện cần | Bảng | Thay đổi | Ưu tiên |
|---|---|---|---|---|
| 1 | Thương hiệu: "BASICO", "Thương hiệu: Basico", ô nhập trong ProductForm | `products` | Thêm `brand` | Cao |
| 2 | Nhiều ảnh: gallery 5 ảnh, ảnh theo màu, quản lý ảnh trong ProductForm | **mới** `product_images` | Bảng mới | Cao |
| 3 | "4.0 ★★★★☆ (1.284 đánh giá)" | `products` | Thêm `rating_avg` (mặc định 4.0), `rating_count` (dữ liệu mô phỏng) | Trung bình |
| 4 | "Đã bán 5,3k" trên thẻ, cột "Đã bán" ở AdminProducts | `products` | Thêm `sold_count`, cập nhật bất đồng bộ | Trung bình |
| 5 | Bảng "Thông tin chi tiết" (Chất liệu, Kiểu dáng…) | `products` | Thêm `specs` (jsonb) | Thấp |
| 6 | Icon danh mục và thứ tự trên thanh bên | `categories` | Thêm `icon_url`, `sort_order` | Thấp |
| 7 | Ảnh nhỏ ở MyOrders và OrderDetail | `order_items` | Thêm `product_image_url` (bản chụp) | Trung bình |
| 8 | "Minh Anh" trên thanh đầu trang, "Nguyễn Minh Anh" ở trang tài khoản | `users` | Thêm `display_name` | Thấp |

Tất cả gộp vào **một migration**, đề xuất số `00013_ui_v2_display_fields.sql`. Mọi cột mới đều nullable hoặc có giá trị mặc định, nên dữ liệu cũ vẫn hợp lệ và có thể rollback.

## 4. Chi tiết

### 4.1. `products.brand`

- Kiểu `text NULL`, giới hạn 1–100 ký tự sau khi trim.
- Thẻ sản phẩm chỉ in thương hiệu khi trường có giá trị.
- Chỉ lưu tên chuỗi, chưa làm bảng `brands` riêng: chưa có màn hình nào lọc hay quản lý thương hiệu. Khi cần bộ lọc theo thương hiệu, tách bảng sau.

```sql
ALTER TABLE products
    ADD COLUMN brand text,
    ADD CONSTRAINT products_brand_check
        CHECK (brand IS NULL OR char_length(btrim(brand)) BETWEEN 1 AND 100);
```

### 4.2. Bảng `product_images`

Gallery ở ProductDetail có ảnh chính và ảnh phụ. Khi chọn màu, gallery đổi ảnh theo màu. ProductForm có "Ảnh chính", "Đặt làm ảnh chính", "Thêm ảnh" và "Thêm ảnh từ URL".

```sql
CREATE TABLE product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    sku_id uuid REFERENCES skus (id) ON DELETE SET NULL,  -- ảnh riêng của một biến thể (màu), có thể trống
    url text NOT NULL,
    alt_text text NOT NULL DEFAULT '',
    position integer NOT NULL,                            -- 0 = ảnh chính
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT product_images_url_check CHECK (char_length(btrim(url)) BETWEEN 1 AND 2000),
    CONSTRAINT product_images_position_check CHECK (position >= 0),
    CONSTRAINT product_images_position_unique UNIQUE (product_id, position) DEFERRABLE INITIALLY DEFERRED
);
CREATE INDEX product_images_by_product ON product_images (product_id, position);
```

- **Giữ cột `products.image_url` làm ảnh bìa.** Danh sách sản phẩm là đường đọc nóng, không nên join thêm bảng. Service đặt `image_url` bằng URL của ảnh có `position = 0` mỗi khi gallery thay đổi, trong cùng giao dịch admin.
- **Migration dữ liệu:** với sản phẩm đã có `image_url`, chèn một dòng `product_images` với `position = 0`.
- **`DEFERRABLE` trên ràng buộc vị trí** để thao tác "đặt làm ảnh chính", tức đổi chỗ hai ảnh, chạy được trong một giao dịch.
- **Giới hạn số ảnh** (đề xuất tối đa 9 ảnh mỗi sản phẩm) kiểm tra ở service, không làm bằng trigger.
- **Upload file thật** nằm ngoài phạm vi. Trước mắt chỉ nhận URL, khớp nút "Thêm ảnh từ URL".

### 4.3. `products.rating_avg`, `products.rating_count`

PRD chưa có tính năng đánh giá, nên **không** làm bảng `reviews`. Hai cột này là **dữ liệu mô phỏng**, chỉ để hiển thị. Nhóm đã chốt: mọi sản phẩm mặc định **4 sao**.

```sql
ALTER TABLE products
    ADD COLUMN rating_avg numeric(2,1) NOT NULL DEFAULT 4.0,
    ADD COLUMN rating_count integer NOT NULL DEFAULT 0,
    ADD CONSTRAINT products_rating_avg_check CHECK (rating_avg BETWEEN 1.0 AND 5.0),
    ADD CONSTRAINT products_rating_count_check CHECK (rating_count >= 0);
```

- **Sản phẩm mới:** khi tạo sản phẩm, kể cả qua API admin, `rating_avg` tự nhận 4.0. Không có API nào ghi vào hai cột này.
- **Seed:** để `rating_avg = 4.0` cho tất cả sản phẩm. `rating_count` có thể đặt số mô phỏng (ví dụ 100–2.000) để dòng "(1.284 đánh giá)" trông tự nhiên.
- **Hiển thị:** luôn hiện sao. Phần "(N đánh giá)" chỉ hiện khi `rating_count > 0`.
- **PRD:** ghi rõ đây là **dữ liệu minh họa**, để hội đồng không hiểu nhầm hệ thống có chức năng đánh giá thật.

### 4.4. `products.sold_count`

Thẻ sản phẩm có "Đã bán 5,3k". AdminProducts có cột "Đã bán".

```sql
ALTER TABLE products
    ADD COLUMN sold_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN sold_count_refreshed_at timestamptz,
    ADD CONSTRAINT products_sold_count_check CHECK (sold_count >= 0);
```

**Cách cập nhật.** Đây là điểm quan trọng nhất của tài liệu. Không tăng cột này trong giao dịch xác nhận thanh toán (nguyên tắc 2). Thay vào đó, một job định kỳ (ví dụ mỗi 1–5 phút, chạy trong worker sẵn có) tính lại:

```sql
UPDATE products p
SET sold_count = s.qty, sold_count_refreshed_at = now()
FROM (
    SELECT sk.product_id, SUM(oi.quantity) AS qty
    FROM order_items oi
    JOIN orders o ON o.id = oi.order_id AND o.status = 'confirmed'
    JOIN skus sk ON sk.id = oi.sku_id
    GROUP BY sk.product_id
) s
WHERE p.id = s.product_id AND p.sold_count <> s.qty;
```

- Con số có thể chậm vài phút. Điều này chấp nhận được với một con số tham khảo, đã làm tròn kiểu "5,3k".
- **Chỉ lấy từ đơn thật** (nhóm đã chốt): không seed số mô phỏng, không có cột `sold_count_seed`. Sản phẩm chưa có đơn `confirmed` thì `sold_count = 0`, và giao diện ẩn dòng "Đã bán" thay vì hiển thị "Đã bán 0".
- **Không nhầm với suất flash sale.** "Đã bán 32/50 suất" là `flash_sale_items.quota_sold / quota_total`, đã có sẵn và cập nhật đúng thời điểm.

### 4.5. `products.specs`

Bảng "Thông tin chi tiết" gồm Thương hiệu, Danh mục, Chất liệu, Kiểu dáng và Phân loại. Thương hiệu, danh mục và phân loại đã có nguồn khác. Chỉ "Chất liệu", "Kiểu dáng" và các dòng tự do tương tự là cần lưu thêm.

```sql
ALTER TABLE products
    ADD COLUMN specs jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD CONSTRAINT products_specs_array_check CHECK (jsonb_typeof(specs) = 'array');
-- Mỗi phần tử: {"label": "Chất liệu", "value": "95% cotton, 5% spandex"}
```

- Dùng mảng thay vì object để giữ thứ tự hiển thị.
- Service kiểm tra tối đa 20 dòng, `label` ≤ 50 ký tự, `value` ≤ 300 ký tự.
- ProductForm cần thêm một khối nhập "Thông số" gồm các dòng nhãn/giá trị. Màn hình v2 hiện chưa có khối này, sẽ bổ sung khi làm React.

### 4.6. `categories.icon_url`, `categories.sort_order`

Thanh bên trang chủ có icon cho từng danh mục và thứ tự cố định: Đang flash sale, Thời trang, Điện tử…

```sql
ALTER TABLE categories
    ADD COLUMN icon_url text,
    ADD COLUMN sort_order integer NOT NULL DEFAULT 0;
CREATE INDEX categories_visible_order ON categories (sort_order, name) WHERE is_visible;
```

- Mục "Đang flash sale" là một bộ lọc, **không phải danh mục thật**, nên không lưu vào bảng này.
- Số sản phẩm cạnh mỗi danh mục (38, 52…) tính bằng `COUNT` sản phẩm `active`. Danh mục ít nên không cần cột đếm.

### 4.7. `order_items.product_image_url`

MyOrders và OrderDetail hiển thị ảnh nhỏ cho từng dòng. OrderDetail ghi "Thông tin được lưu tại thời điểm đặt đơn". Ảnh cũng phải là bản chụp, vì nếu đọc `products.image_url` thì đơn cũ sẽ đổi ảnh khi admin thay ảnh.

```sql
ALTER TABLE order_items ADD COLUMN product_image_url text;
```

- Gán giá trị khi tạo đơn, cùng lúc với `product_name` và `sku_code`. Lấy ảnh của SKU nếu có, nếu không thì lấy ảnh bìa sản phẩm.
- Đơn cũ để `NULL`, giao diện hiển thị ô xám.
- **Chi phí:** thêm một giá trị chuỗi vào câu `INSERT` đã có. Không thêm truy vấn hay khóa nào, vì dữ liệu sản phẩm đã được đọc ở bước xem trước đơn.

### 4.8. `users.display_name`

Thanh đầu trang hiển thị "Minh Anh". Trang tài khoản và avatar hiển thị "Nguyễn Minh Anh" và chữ tắt "MA".

```sql
ALTER TABLE users
    ADD COLUMN display_name text,
    ADD CONSTRAINT users_display_name_check
        CHECK (display_name IS NULL OR char_length(btrim(display_name)) BETWEEN 1 AND 120);
```

- Khi trống, hiển thị phần trước dấu `@` của email.
- Form đăng ký **không** thêm ô họ tên, vì giao diện v2 cố ý chỉ hỏi email và mật khẩu. Người dùng sửa tên ở trang tài khoản.
- Thông tin người nhận ở bước Xem trước đơn được điền sẵn từ **đơn gần nhất** (`orders.recipient_*`), không cần sổ địa chỉ (xem mục 5).

## 5. Giá trị suy ra (không thêm cột)

| Giao diện | Cách tính |
|---|---|
| Trạng thái chương trình: Nháp, Đã lên lịch, Đang diễn ra, Đã dừng, Đã kết thúc | `status` cùng `starts_at`, `ends_at` và giờ hệ thống. Chỉ lưu `draft`, `scheduled` và `stopped` như hiện tại. |
| "Giảm đến 45%" | `MAX(1 - sale_price / regular_price)` trên các `flash_sale_items` của chương trình |
| "Đã bán 78/130 suất", "Đang giữ 11 suất" | Tổng `quota_sold`, `quota_total` và `quota_reserved` của chương trình |
| "Sắp hết · còn 2 suất" | `quota_total - quota_sold - quota_reserved`, ngưỡng 85% xử lý ở giao diện |
| Phần trăm giảm trên thẻ ("-35%") | `sale_price` so với `regular_price` |
| "Tối đa 2/người" | `flash_sale_items.purchase_limit` |
| Tab "Cần đối soát" ở MyOrders | Đơn có `payment_reconciliations.status = 'open'` |
| "Còn 03:12 để thanh toán" | `orders.expires_at` trừ giờ hệ thống |
| "Giao đến: Q. 1, TP. Hồ Chí Minh" trên thanh đầu trang | Địa chỉ của đơn gần nhất. Khi chưa có đơn thì ẩn. |
| Tồn kho: Hiện có, Đang giữ, Khả dụng | `inventories.on_hand`, `reserved` và hiệu của hai cột |

## 6. Quyết định đã chốt (26/11/2026)

1. **Bỏ nút "Mua lại"** trên đơn thất bại, hết hạn hoặc đã hủy. Không thêm hành vi đặt lại đơn. Màn hình v2 đã gỡ nút này ở MyOrders. Ở Payment-expired, nút được thay bằng "Tiếp tục mua sắm".
2. **Số đã bán chỉ lấy từ đơn thật** (`orders.status = 'confirmed'`), không cộng số mô phỏng (mục 4.4).
3. **Sao đánh giá là dữ liệu mô phỏng, mặc định 4 sao** cho mọi sản phẩm (mục 4.3).

## 7. Không làm trong đợt này

- **Bảng `reviews` và chức năng viết đánh giá**: chưa có trong PRD.
- **Nút "Mua lại" và chức năng đặt lại đơn**: đã quyết định bỏ.
- **Bảng `brands` và bộ lọc theo thương hiệu.**
- **Danh mục nhiều cấp** (`parent_id`). Dòng "Thời trang › Áo thun" trong bảng thông số sẽ chỉ hiển thị "Thời trang".
- **Sổ địa chỉ, phí vận chuyển, mã giảm giá.** Màn hình hiện ghi "Phí vận chuyển: Chưa áp dụng".
- **Upload ảnh lên kho lưu trữ**: chỉ nhận URL.
- **Gợi ý tìm kiếm phổ biến** ("áo thun", "tai nghe bluetooth"…): để cố định ở frontend.
- **Banner hoặc bộ sưu tập do admin biên tập** ("Mới về · Chăm sóc da mùa nắng"): trước mắt lấy từ danh mục hoặc để cố định.

## 8. Ảnh hưởng và thứ tự triển khai

1. **PRD và `docs/technical/traceability.md`:** thêm yêu cầu hiển thị (thương hiệu, gallery, thông số, ảnh trong đơn, tên hiển thị). Ghi chú sao đánh giá và số đã bán là dữ liệu tham khảo. Ghi rõ job làm mới `sold_count` không nằm trong luồng đặt đơn.
2. **`docs/technical/data-model.md`:** thêm các cột và bảng ở mục 4.
3. **`docs/technical/api/openapi.yaml`:**
   - Product (danh sách và chi tiết) thêm `brand`, `rating_avg`, `rating_count`, `sold_count`, `images[]`, `specs[]`.
   - Category thêm `icon_url`.
   - Order item thêm `product_image_url`.
   - User (me) thêm `display_name`.
   - API admin cho sản phẩm thêm `brand`, `specs` và các thao tác gallery: thêm, xóa, đổi thứ tự, đặt làm ảnh chính.
4. **Migration `00013_ui_v2_display_fields.sql`** (có `-- +goose Down`) và backfill `product_images` từ `products.image_url`.
5. **Model GORM** (`product.go`, `category.go`, `order_item.go`, `user.go`) và model mới `product_image.go`.
6. **Service:**
   - Catalog: đọc và ghi các trường mới, quản lý gallery.
   - Checkout: chụp `product_image_url`.
   - Worker: job `sold_count`.
7. **Seed** (`docs/superpowers/specs/2026-11-10-unified-seed-design.md`): thêm thương hiệu, nhiều ảnh, thông số, `rating_avg = 4.0` cùng `rating_count` mô phỏng, và icon danh mục cho dữ liệu demo. Không seed `sold_count`.
8. **Test:**
   - Migration up/down.
   - Ràng buộc rating.
   - Đổi ảnh chính trong một giao dịch.
   - Bản chụp ảnh trong đơn không đổi khi admin thay ảnh.
   - Job `sold_count` chỉ đếm đơn `confirmed`.
   - Kiểm thử tải đặt đơn không chậm đi vì có thêm cột.
