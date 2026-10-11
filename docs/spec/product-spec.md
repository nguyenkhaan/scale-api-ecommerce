# Đặc tả quyết định sản phẩm — SaleScale

**Trạng thái:** Đã xác nhận

**Ngày cập nhật:** 11/10/2026

Các quyết định dưới đây là cơ sở triển khai và nghiệm thu. Thay đổi phải cập nhật đồng thời PRD, kế hoạch, yêu cầu chức năng, user stories và test liên quan.

| Nội dung | Quyết định |
|---|---|
| Thời gian giữ hàng | 5 phút; lưu hạn cụ thể trên đơn; retry không gia hạn |
| Quota flash sale | Dùng chung tồn vật lý; quota chỉ là trần ưu đãi |
| Giới hạn mua | Theo tài khoản + chương trình + SKU; tính cả đang giữ và đã mua |
| Giỏ thường và flash sale | Cho phép cùng giỏ; mỗi dòng ghi rõ nguồn giá |
| Chọn dòng checkout | Cho chọn một tập dòng; không tự lấy toàn bộ giỏ |
| Giỏ sau khi đặt | Giữ nguyên khi chờ; chỉ xóa an toàn dòng chưa đổi sau khi xác nhận |
| Thanh toán | Chỉ dùng bộ mô phỏng bắt buộc; không dùng tiền thật hoặc provider sandbox |
| Phiên thanh toán | Một phiên cho một đơn; retry trả phiên cũ |
| Tiền | VND số nguyên; không phí vận chuyển, không tách thuế |
| Người nhận | Tên, số điện thoại và địa chỉ văn bản |
| Giới hạn giỏ | Tối đa 20 SKU; số lượng theo cấu hình SKU/chương trình |
| Xác thực | Session cookie an toàn; session lưu ở PostgreSQL để dùng được nhiều instance |
| Frontend | React + Vite |
| Mục tiêu hiệu năng | p95 checkout ≤ 500 ms, p99 ≤ 1.000 ms tại tải tham chiếu; lỗi hệ thống < 1% |
| Thông báo đơn | Feature P1 qua RabbitMQ và PostgreSQL outbox; broker lỗi có thể làm thông báo trễ nhưng không làm checkout lỗi hoặc mất trạng thái đơn |
| Chậm giải phóng | 95% reservation được giải phóng trong 30 giây sau hạn tại tải tham chiếu |

Tải và cấu hình tham chiếu được ghi sau baseline E0; các mục tiêu hiệu năng trên không tự thay đổi theo baseline.
