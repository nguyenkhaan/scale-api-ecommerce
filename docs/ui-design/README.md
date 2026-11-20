# SaleScale UI – HTML tĩnh

- `index.html`: trang mục lục, xem và chuyển giữa 76 màn hình (dùng phím ← → hoặc menu bên trái).
- `pages/*.html`: mỗi màn hình là một file HTML độc lập, chỉ dùng HTML + CSS inline, không cần JavaScript hay thư viện nào.
- Link giữa các màn hình đã được nối sẵn (ví dụ `Main.html` → `Cart.html`).
- Font Inter tải từ Google Fonts; khi offline sẽ tự dùng font hệ thống.

Lưu ý: đây là bản chụp tĩnh của từng trạng thái. Các tương tác trong bản thiết kế gốc
(đồng hồ đếm ngược chạy, đổi tab, mở menu, tăng/giảm số lượng) không còn chạy; mỗi trạng thái
đã có màn hình riêng (ví dụ Cart-updating.html, Cart-empty.html).
