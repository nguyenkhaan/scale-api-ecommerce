## 10. User stories và các tình huống nghiệm thu chính

US-01 đến US-18 mô tả hành vi backend theo từng mục đích sử dụng; US-19 và US-20 nghiệm thu giao diện tích hợp các hành vi đó. Các story vẫn cần được tách thành công việc triển khai nhỏ nếu việc thiết kế và kiểm thử vượt quá một phiên làm việc.

### US-01 — Sử dụng tài khoản khách

**Câu chuyện:** Là khách hàng, tôi muốn đăng ký và đăng nhập để quản lý giỏ và đơn của mình.

**Liên quan:** FR-001 đến FR-003.

- [ ] Email đăng ký trùng bị từ chối; thông tin đăng nhập sai không cấp quyền.
- [ ] Phiên hết hạn/đăng xuất không tiếp tục truy cập API riêng tư.
- [ ] Khách không truy cập dữ liệu của khách khác hoặc chức năng quản trị.

### US-02 — Quản lý hàng hóa bán được

**Câu chuyện:** Là quản trị, tôi muốn tạo danh mục, sản phẩm và SKU để đưa đúng biến thể lên cửa hàng.

**Liên quan:** FR-005 đến FR-009.

- [ ] SKU không trùng; dữ liệu bắt buộc và giá được kiểm tra.
- [ ] Hàng ngừng bán không được đặt mới.
- [ ] Thay đổi tên/giá hiện tại không thay đổi dòng đơn đã chụp.

### US-03 — Điều chỉnh tồn kho có thể kiểm tra

**Câu chuyện:** Là quản trị, tôi muốn nhập hoặc điều chỉnh hàng thực có để số lượng bán được đúng với kho.

**Liên quan:** FR-010, FR-011.

- [ ] Có số trước/sau, lý do, tác nhân và thời điểm điều chỉnh.
- [ ] Điều chỉnh không làm tồn âm hoặc nhỏ hơn hàng đang giữ.
- [ ] Đối chiếu được số hiện còn, đang giữ và khả dụng.

### US-04 — Chuẩn bị ưu đãi flash sale

**Câu chuyện:** Là quản trị, tôi muốn chọn SKU, giá, quota và giới hạn mua để kiểm soát lượng hàng ưu đãi.

**Liên quan:** FR-014.

- [ ] Dữ liệu thời gian, giá, quota và giới hạn không hợp lệ bị từ chối.
- [ ] Cùng SKU không xuất hiện hai lần trong chương trình.
- [ ] Cấu hình nháp có thể được sửa trước khi lên lịch.

### US-05 — Mở bán theo lịch và dừng khi cần

**Câu chuyện:** Là quản trị, tôi muốn lên lịch hoặc dừng nhận đơn để chương trình chạy đúng thời gian và điều kiện.

**Liên quan:** FR-015, FR-016, FR-018.

- [ ] Chương trình không chồng thời gian của cùng SKU; cấu hình được khóa khi lên lịch.
- [ ] Không nhận đơn ưu đãi trước giờ mở hoặc từ giờ kết thúc.
- [ ] Dừng chương trình không phá đơn đã giữ hàng còn hạn.

### US-06 — Khám phá sản phẩm và ưu đãi

**Câu chuyện:** Là khách hàng, tôi muốn xem hàng theo danh mục và SKU để chọn đúng mặt hàng có thể mua.

**Liên quan:** FR-008, FR-016.

- [ ] Danh sách có phân trang, lọc danh mục và tìm theo tên đơn giản.
- [ ] Chi tiết thể hiện đúng biến thể và giá/nguồn giá.
- [ ] Thông tin xem sản phẩm chưa tạo quyền giữ hàng.

### US-07 — Chuẩn bị giỏ nhiều mặt hàng

**Câu chuyện:** Là khách hàng, tôi muốn thêm và sửa nhiều SKU trong giỏ trước khi thanh toán chung.

**Liên quan:** FR-019, FR-023.

- [ ] SKU trùng được gộp; số lượng và giới hạn dòng được kiểm tra trên backend.
- [ ] Thao tác giỏ không thay đổi reservation, tồn hoặc quota.
- [ ] Giỏ vẫn còn khi đặt đơn bị từ chối.

### US-08 — Xác nhận giá và người nhận

**Câu chuyện:** Là khách hàng, tôi muốn xem trước các dòng, tổng tiền và người nhận để xác nhận đúng nội dung mua.

**Liên quan:** FR-020 đến FR-022.

- [ ] Tổng tiền bằng tổng các dòng đã chọn, có nguồn giá và tiền tệ rõ ràng.
- [ ] Thông tin người nhận bắt buộc được kiểm tra.
- [ ] Thay đổi giá/nguồn giá yêu cầu xác nhận mới, không tự đặt với giá khác.

### US-09 — Đặt một đơn giữ đủ hàng

**Câu chuyện:** Là khách hàng, tôi muốn đặt toàn bộ tập dòng đã xác nhận thành một đơn để thanh toán chung.

**Liên quan:** FR-024, FR-025, FR-029.

- [ ] Thành công tạo một đơn với đủ dòng và reservation.
- [ ] Có bản chụp giá, tổng tiền và hạn thanh toán.
- [ ] Đặt đồng thời vẫn tuân thủ tồn, quota và giới hạn mua.

### US-10 — Sửa giỏ khi một mặt hàng thiếu

**Câu chuyện:** Là khách hàng, tôi muốn biết mặt hàng nào thiếu để tự quyết định sửa giỏ và xác nhận lại.

**Liên quan:** FR-023, FR-026.

- [ ] Toàn bộ lần đặt bị từ chối, không giữ sót mặt hàng khác.
- [ ] Lỗi nêu SKU, số lượng yêu cầu và nguyên nhân có thể xác định.
- [ ] Không tự mua một phần; đặt lại sau sửa giỏ dùng xác nhận và khóa mới.

### US-11 — Gửi lại khi kết quả đặt đơn chưa rõ

**Câu chuyện:** Là khách hàng, tôi muốn kiểm tra lại yêu cầu bị mất phản hồi mà không mua trùng.

**Liên quan:** FR-027, FR-028.

- [ ] Hai request cùng khóa qua hai instance tạo tối đa một đơn.
- [ ] Timeout sau commit được tra cứu hoặc retry cùng khóa.
- [ ] Cùng khóa với nội dung khác bị báo xung đột.

### US-12 — Mở thanh toán cho đúng đơn

**Câu chuyện:** Là khách hàng, tôi muốn mở phiên thanh toán thử nghiệm cho đơn đã giữ hàng.

**Liên quan:** FR-030.

- [ ] Chỉ chủ đơn còn đủ điều kiện được mở phiên.
- [ ] Retry không tạo phiên thứ hai và không gia hạn giữ hàng.
- [ ] Phiên dùng đúng mã đơn, số tiền và tiền tệ đã chốt.

### US-13 — Nhận kết quả thanh toán đáng tin cậy

**Câu chuyện:** Là doanh nghiệp, tôi muốn chỉ xác nhận đơn khi có kết quả thanh toán hợp lệ để tránh trừ tồn sai.

**Liên quan:** FR-031 đến FR-035.

- [ ] Thành công đúng hạn xác nhận và chuyển hàng thành bán đúng một lần.
- [ ] Thất bại chắc chắn giải phóng; timeout không bị coi là thất bại chắc chắn.
- [ ] Sai nguồn/số tiền, sự kiện lặp và trái thứ tự không làm sai trạng thái/tồn.

### US-14 — Hủy đơn chưa thanh toán

**Câu chuyện:** Là khách hàng, tôi muốn hủy đơn đang chờ trước hạn khi không tiếp tục mua.

**Liên quan:** FR-038, FR-040.

- [ ] Hủy hợp lệ giải phóng toàn bộ hàng đúng một lần.
- [ ] Không hủy được đơn đã xác nhận theo luồng này.
- [ ] Hủy cạnh tranh với callback không vừa bán vừa trả cùng reservation.

### US-15 — Trả lại hàng của đơn hết hạn

**Câu chuyện:** Là doanh nghiệp, tôi muốn hàng không được thanh toán đúng hạn trở lại khả dụng cho khách khác.

**Liên quan:** FR-039, FR-040.

- [ ] Đơn quá hạn không tiếp tục được thanh toán dù worker cập nhật chậm.
- [ ] Nhiều worker và worker khởi động lại không giải phóng lặp hoặc bỏ sót việc.
- [ ] Callback muộn không lấy lại hàng đã giải phóng.

### US-16 — Rà soát kết quả thanh toán bất thường

**Câu chuyện:** Là quản trị, tôi muốn xem bằng chứng kết quả muộn hoặc mâu thuẫn để biết trường hợp nào cần xử lý ngoài luồng tự động.

**Liên quan:** FR-035, FR-036.

- [ ] Thấy loại bất thường, đơn/phiên và sự kiện liên quan.
- [ ] Ghi chú/đánh dấu rà soát không đổi trạng thái đơn hoặc tồn.
- [ ] Không có thao tác hoàn tiền thật hoặc hồi sinh đơn trong MVP.

### US-17 — Tra cứu đơn và kết quả chính thức

**Câu chuyện:** Là khách hàng hoặc quản trị được phép, tôi muốn tra cứu đơn để biết chính xác hàng, số tiền và trạng thái backend đã ghi nhận.

**Liên quan:** FR-037, FR-041.

- [ ] Khách chỉ thấy đơn của mình; quản trị tra cứu theo mã/trạng thái/thời gian.
- [ ] Có bản chụp dòng hàng, tổng tiền và thời hạn.
- [ ] Trang redirect không được dùng thay trạng thái chính thức của backend.

### US-18 — Kiểm chứng khả năng chịu tải

**Câu chuyện:** Là doanh nghiệp/người vận hành, tôi muốn biết năng lực xử lý và điểm nghẽn để chọn cấu hình triển khai dựa trên bằng chứng.

**Liên quan:** FR-042, FR-043, PERF-01 đến PERF-08.

- [ ] Có baseline, điều kiện tài nguyên, tải mục tiêu và tải thực phát.
- [ ] Có kiểm tra bất biến sau test và dữ liệu gốc để tái lập.
- [ ] Không dùng tăng số request từ chối để tuyên bố tăng số đơn thành công.

### US-19 — Demo luồng khách bằng trình duyệt

**Câu chuyện:** Là người đánh giá, tôi muốn mua thử qua giao diện để kiểm tra các nghiệp vụ phía khách đã được tích hợp.

**Liên quan:** FR-044.

- [ ] Có các màn hình chọn SKU, giỏ, xem trước, đặt, thanh toán và trạng thái đơn.
- [ ] Lỗi thiếu hàng/giá thay đổi có thông tin để khách sửa và xác nhận lại.
- [ ] Kiểm chứng trực tiếp trong trình duyệt, dùng skill `dev-browser` nếu có; nếu không, dùng công cụ trình duyệt hiện có và lưu bằng chứng tương đương.

### US-20 — Demo luồng quản trị bằng trình duyệt

**Câu chuyện:** Là người đánh giá, tôi muốn chuẩn bị chương trình và tra cứu đơn qua giao diện để kiểm tra nghiệp vụ doanh nghiệp.

**Liên quan:** FR-004, FR-045.

- [ ] Có quản lý hàng/tồn/chương trình và xem đơn/đối soát đủ cho hành trình 8.1.
- [ ] Không cung cấp nút sửa giá đơn, ép trạng thái thành công hoặc sửa lượng giữ.
- [ ] Kiểm chứng trực tiếp trong trình duyệt, dùng skill `dev-browser` nếu có; nếu không, dùng công cụ trình duyệt hiện có và lưu bằng chứng tương đương.

