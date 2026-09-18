## 9. Yêu cầu chức năng và tiêu chí nghiệm thu

**Mức ưu tiên:** P0 là bắt buộc để nghiệm thu luồng nghiệp vụ và nghiên cứu; P1 là hỗ trợ tối thiểu nằm trong kế hoạch MVP. Giảm P1 phải được ghi nhận, không được ngầm bỏ chức năng đã xác nhận.

### 9.1. Tài khoản và cửa hàng

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-001 | P0 | Khách đăng ký bằng email và mật khẩu | Email được chuẩn hóa và duy nhất; dữ liệu không hợp lệ bị từ chối; khách không tự gán vai trò quản trị |
| FR-002 | P0 | Đăng nhập, đăng xuất và xác thực request riêng tư | Sai thông tin không cấp quyền; thông tin xác thực hết hạn không truy cập API riêng tư; logout vô hiệu hóa phiên theo cơ chế được chọn |
| FR-003 | P0 | Phân biệt quyền khách và quản trị | Khách không sửa hàng hóa/flash sale; khách A không đọc hoặc đổi giỏ, đơn và thanh toán của B |
| FR-004 | P1 | Cấu hình nhận diện cửa hàng | Quản trị sửa tên, mô tả ngắn, URL logo và liên hệ; khách xem thông tin công khai; không có thao tác chọn tenant |

### 9.2. Danh mục, sản phẩm và SKU

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-005 | P1 | Quản lý danh mục một cấp | Tạo/sửa/ngừng hiển thị; không cho danh mục không hợp lệ; lịch sử sản phẩm được giữ |
| FR-006 | P0 | Quản lý sản phẩm vật lý | Có tên, mô tả, danh mục, ảnh URL và trạng thái; sản phẩm ngừng bán không được đặt mới |
| FR-007 | P0 | Quản lý SKU và thuộc tính biến thể | Mã SKU duy nhất; màu/kích cỡ có thể khác nhau; sản phẩm đơn giản có SKU mặc định; giá/số lượng sai bị từ chối |
| FR-008 | P0 | Xem danh sách và chi tiết hàng công khai | Có phân trang; lọc theo danh mục; tìm theo tên đơn giản; hiển thị đúng SKU, giá và trạng thái bán |
| FR-009 | P0 | Không làm mất lịch sử hàng đã bán | Đổi tên/giá hiện tại không thay đổi bản chụp dòng đơn; không xóa SKU đang được reservation/đơn tham chiếu |

### 9.3. Tồn kho

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-010 | P0 | Nhập và điều chỉnh tồn theo SKU | Số lượng nguyên; điều chỉnh có lý do/tác nhân/thời điểm; không tạo tồn âm hoặc ít hơn hàng đang giữ |
| FR-011 | P0 | Quản trị xem tồn hiện còn, đang giữ và khả dụng | Số liệu nhất quán theo mục 6.7; số lượng nội bộ không bị cache riêng tư công khai nhầm |
| FR-012 | P0 | Ghi nhận reservation của từng dòng đơn | Mỗi dòng có số lượng giữ tương ứng; truy nguyên được đơn và hạn giữ; trạng thái đã commit không giữ một phần giỏ |
| FR-013 | P0 | Giải phóng hoặc chuyển reservation đúng một lần | Retry và chạy đồng thời không làm âm `reserved`, tăng tồn sai hoặc trừ hàng hai lần |

### 9.4. Flash sale

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-014 | P0 | Tạo/sửa chương trình nháp | Kiểm tra thời gian, giá ưu đãi, quota, giới hạn mua và SKU hoạt động; trả lỗi cấu hình cụ thể |
| FR-015 | P0 | Lên lịch và khóa cấu hình | Không cho cùng SKU chồng khoảng nhận đơn; chương trình lên lịch giữ nguyên các thông số bán; hủy lịch trước mở cho phép sửa |
| FR-016 | P0 | Hiển thị chương trình và giá ưu đãi | Khách xem giờ mở/đóng và điều kiện; backend kiểm tra thời gian chính thức, không tin đồng hồ client |
| FR-017 | P0 | Áp dụng đồng thời tồn kho, quota và giới hạn mua | Cùng một request và các request đồng thời đều không vượt INV-01, INV-09, INV-10 |
| FR-018 | P1 | Dừng nhận đơn mới và xem thống kê chương trình | Dừng không phá đơn giữ trước đó; thống kê tách lượng giữ, lượng đã bán và đơn bị từ chối |

### 9.5. Giỏ hàng và xem trước đơn

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-019 | P0 | Thêm, đổi số lượng và xóa dòng giỏ | Giỏ thuộc khách; SKU trùng gộp; số lượng không hợp lệ bị từ chối; không tạo reservation |
| FR-020 | P0 | Xem trước tập dòng sẽ mua | Trả SKU, số lượng, giá từng dòng, nguồn giá và tổng; thông tin chỉ là xem trước, chưa giữ hàng |
| FR-021 | P0 | Nhận thông tin người nhận cho hàng vật lý | Đề xuất có tên, số điện thoại, địa chỉ dạng văn bản; lưu bản chụp trên đơn; không yêu cầu tích hợp bản đồ/vận chuyển |
| FR-022 | P0 | Bắt buộc xác nhận lại khi giá/điều kiện đổi | Đặt đơn dựa trên giá cũ trả giá mới và lỗi phù hợp; không tự nhận tổng tiền cao hơn khách đã xác nhận |
| FR-023 | P1 | Giữ giỏ khi đặt thất bại và cập nhật an toàn sau mua | Không xóa giỏ khi thiếu hàng; không xóa dòng khách vừa sửa hoặc mới thêm trong lúc đơn đang thanh toán |

### 9.6. Đặt đơn

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-024 | P0 | Kiểm tra toàn bộ tập dòng khi đặt | Giỏ không rỗng; SKU/giá/chương trình/số lượng/thông tin người nhận hợp lệ; kiểm tra từ dữ liệu có thẩm quyền |
| FR-025 | P0 | Tạo một đơn và giữ đủ hàng nguyên tử | Thành công có mã đơn và đầy đủ dòng/reservation; thất bại không tạo đơn chờ hoặc tác động giữ hàng đã commit |
| FR-026 | P0 | Báo mặt hàng không đủ điều kiện | Có mã lỗi, SKU, số lượng yêu cầu và nguyên nhân; không tự mua một phần hoặc đổi sang giá thường |
| FR-027 | P0 | Chống tạo đơn trùng bằng idempotency | Cùng khóa/nội dung trả cùng kết quả; khác nội dung báo xung đột; hoạt động qua hai instance |
| FR-028 | P0 | Cho tra cứu kết quả request chưa rõ | Khách tra cứu theo khóa thuộc mình hoặc retry cùng khóa; không lộ kết quả khách khác |
| FR-029 | P0 | Trả hạn thanh toán và bản chụp giá | Retry không gia hạn; đổi giá sản phẩm sau đặt không sửa tổng đơn |

### 9.7. Thanh toán thử nghiệm/mô phỏng

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-030 | P0 | Khởi tạo hoặc lấy lại phiên thanh toán | Đúng chủ đơn, đơn còn đủ điều kiện; phiên có đúng số tiền/tiền tệ; retry không tạo phiên thứ hai |
| FR-031 | P0 | Bộ mô phỏng phát sự kiện bất đồng bộ | Chủ động tạo thành công, thất bại, chậm, mất phản hồi, lặp và trái thứ tự; có định danh và xác thực nguồn |
| FR-032 | P0 | Xác minh sự kiện và ghi nhận bền vững | Sai nguồn/phiên/số tiền không xác nhận; sự kiện được truy nguyên; chỉ xác nhận tiếp nhận sau khi lưu |
| FR-033 | P0 | Thanh toán hợp lệ xác nhận đơn đúng một lần | Chuyển toàn bộ reservation thành bán; cập nhật trạng thái và quota nhất quán; callback lặp không trừ thêm |
| FR-034 | P0 | Phân biệt thất bại chắc chắn với kết quả chưa rõ | Thất bại xác minh giải phóng; timeout không tự làm đơn thất bại; không mở lại trạng thái kết thúc |
| FR-035 | P0 | Xử lý thành công đến muộn và mâu thuẫn | Đơn không được hồi sinh; có bản ghi đối soát; không tự trừ tồn hoặc hoàn tiền thật |
| FR-036 | P1 | Quản trị xem và ghi chú đối soát | Lọc đơn/sự kiện bất thường; xem bằng chứng tối thiểu; ghi chú không sửa trạng thái đơn/tồn |

### 9.8. Đơn hàng và hết hạn

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-037 | P0 | Khách xem danh sách/chi tiết đơn của mình | Có phân trang, trạng thái, hàng, số tiền và hạn; dùng dữ liệu backend làm trạng thái chính thức |
| FR-038 | P0 | Khách hủy đơn còn chờ trước hạn | Giải phóng toàn bộ một lần; đơn xác nhận không bị hủy theo luồng này; hủy lặp không tăng tồn |
| FR-039 | P0 | Tác vụ nền hết hạn đơn | Chạy lại sau gián đoạn; không bỏ sót đơn tới hạn; nhiều worker không giải phóng hai lần |
| FR-040 | P0 | Thanh toán/hủy/hết hạn cạnh tranh an toàn | Chỉ một chuyển trạng thái có hiệu lực; trả trạng thái hiện tại khi mất quyền chuyển; giữ INV-06 đến INV-08 |
| FR-041 | P1 | Quản trị tra cứu đơn | Lọc theo mã, trạng thái và thời gian; xem bản chụp hàng, lịch sử và kết quả thanh toán; không có nút sửa giá đơn đã tạo |

### 9.9. Quan sát và giao diện demo

| Mã | Ưu tiên | Yêu cầu | Tiêu chí nghiệm thu |
|---|---|---|---|
| FR-042 | P0 | Đo và liên kết request, đơn, sự kiện | Theo dõi được thời gian xử lý, kết quả, instance và định danh; không ghi mật khẩu/token/bí mật vào log |
| FR-043 | P0 | Xuất chỉ số nghiệp vụ và hạ tầng | Phân biệt đơn tạo, đơn xác nhận, thiếu hàng, quota/giới hạn mua, rate limit, timeout và lỗi hệ thống |
| FR-044 | P1 | Giao diện demo của khách | Có danh sách/chi tiết, chọn SKU, giỏ, xem trước, đặt, thanh toán thử nghiệm, trạng thái đơn và lỗi theo dòng |
| FR-045 | P1 | Giao diện demo của quản trị | Quản lý hàng/tồn/chương trình và xem đơn/đối soát đủ cho hành trình 8.1; không cần UI thương mại đầy đủ |
| FR-046 | P0 | Hướng dẫn kiểm chứng và bàn giao | Có cách dựng môi trường, tạo dữ liệu, chạy các test P0, benchmark và đọc kết quả |

### 9.10. Các nhóm thao tác API cần hỗ trợ

| Nhóm | Đầu vào nghiệp vụ chính | Kết quả nghiệp vụ chính |
|---|---|---|
| Tài khoản | Email/mật khẩu; thông tin xác thực | Danh tính, quyền và trạng thái phiên |
| Catalog | Bộ lọc, phân trang, mã sản phẩm/SKU | Hàng hóa công khai và giá hiển thị |
| Quản trị hàng/tồn | Thuộc tính hàng; lượng điều chỉnh và lý do | Hàng hóa được quản lý, tồn và lịch sử thay đổi |
| Chương trình | Thời gian; SKU; giá; quota; giới hạn | Cấu hình, trạng thái nhận đơn và thống kê |
| Giỏ/xem trước | SKU, số lượng, tập dòng, phiên bản giỏ | Giỏ hiện hành, giá xem trước và cảnh báo |
| Đặt đơn | Tập dòng đã xác nhận, giá/tổng, người nhận, khóa idempotency | Đơn giữ đủ hàng hoặc lỗi toàn bộ lần đặt |
| Tra cứu/hủy đơn | Mã đơn hoặc khóa đặt; thông tin xác thực | Trạng thái chính thức, chi tiết hoặc kết quả hủy |
| Thanh toán | Mã đơn/phiên; sự kiện đã xác thực | Phiên thử nghiệm, kết quả ghi nhận và trạng thái đơn |
| Quan sát | Quyền vận hành; khoảng thời gian | Chỉ số, nhật ký liên quan và bằng chứng kiểm thử |

Đường dẫn, HTTP method, cấu trúc request/response và cơ chế xác thực cụ thể được chốt trong đặc tả API. Không đặt thông tin nội bộ database vào luồng người dùng.

