# SaleScale UI v2 – HTML tĩnh (phong cách Tiki, màu xanh)

- Xem: mở `pages/index.html` – mục lục 76 màn hình, cùng tên file với bộ cũ ở `../pages`.
- Màn hình `*-mobile.html` hiển thị trang tương ứng trong khung điện thoại 390 × 844; mọi trang đều responsive.
- Ảnh sản phẩm là placeholder SVG; số liệu (đánh giá, đã bán, thương hiệu) là dữ liệu mẫu.

## Sửa và build lại

- Token và component chung: `assets/salescale.css`; CSS riêng từng nhóm: `assets/{catalog,auth,checkout,orders,admin}.css`.
- Helper chung: `src/lib.mjs`; mỗi nhóm màn hình là một module trong `src/pages/`.
- Build toàn bộ (xóa và sinh lại `pages/`): `node docs/ui-design/v2/src/build.mjs`
- Build riêng một nhóm: `node docs/ui-design/v2/src/render-one.mjs pages/30-checkout.mjs`

Quyết định thiết kế: [`docs/superpowers/specs/2026-11-26-ui-v2-redesign-design.md`](../../superpowers/specs/2026-11-26-ui-v2-redesign-design.md).
