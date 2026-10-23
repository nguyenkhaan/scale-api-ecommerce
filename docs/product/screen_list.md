# Màn hình FE cần có

Danh sách dẫn xuất từ [PRD](prd.md) và [OpenAPI](../technical/api/openapi.yaml). UI demo dùng trạng thái, giá và hạn từ backend.

| Màn hình | Nội dung chính | Phase |
|---|---|---|
| Đăng ký / đăng nhập | Tạo tài khoản khách, mở/đóng session, báo lỗi | 01 |
| Sản phẩm danh sách / chi tiết | Tìm, lọc, phân trang, chọn SKU | 02 |
| Flash sale danh sách / chi tiết | Xem giờ, SKU, giá, điều kiện và trạng thái | 03 |
| Giỏ hàng | Thêm/sửa/xóa SKU, chọn tập dòng sẽ mua | 03 |
| Preview / xác nhận | Người nhận, giá hiện tại, tổng tiền, cảnh báo đổi giá | 04 |
| Kết quả checkout | Mã đơn, hạn, retry/tra cứu cùng key khi kết quả chưa rõ | 04 |
| Thanh toán mô phỏng | Mở lại một phiên, xem kết quả chính thức từ backend | 05 |
| Đơn của tôi | Danh sách, chi tiết, snapshot, hủy khi còn điều kiện | 05 |
| Quản trị hàng / tồn | Form sản phẩm, SKU, điều chỉnh và lịch sử tồn | 02–03 |
| Quản trị flash sale | Nháp, lịch, dừng, thống kê | 03, 06 |
| Quản trị đơn / đối soát | Tra cứu, xem sự kiện bất thường, ghi chú | 06 |
| Hộp thông báo | Danh sách riêng tư, đánh dấu đã đọc | 06 |
| Thông tin cửa hàng | Hiển thị công khai, chỉnh sửa nhận diện | 06 |

Các màn hình có trạng thái tải/rỗng/lỗi, dùng bàn phím được và phù hợp màn hình nhỏ. Form quản trị chỉ cần đủ cho demo. Phần P1 phải chi tiết hóa contract trước khi làm.
