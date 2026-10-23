# ADR-003: PRD là nguồn sự thật nghiệp vụ

## Trạng thái

Accepted

## Ngày

2026-10-18

## Bối cảnh

Dự án có PRD, functional requirements, user stories, research notes và nhiều tài liệu kỹ thuật. Một số research note còn chứa quyết định cũ như JWT, thời hạn giữ 10 phút, PayOS sandbox, mục tiêu RPS cố định hoặc không dùng Kubernetes. Nếu mọi file cùng có quyền quyết định, nhóm và coding agent có thể triển khai các hành vi khác nhau.

## Quyết định

- PRD là nguồn sự thật cho mục tiêu, phạm vi, `BR`, `FR`, `INV`, `NFR`, `PERF`, `TEST` và quyết định sản phẩm.
- `spec.md` điều phối quy trình, chuẩn kỹ thuật và liên kết; không lặp lại hay ghi đè nghiệp vụ.
- Research là đầu vào/lịch sử. Nội dung chỉ có hiệu lực sau khi được hợp nhất vào PRD hoặc ADR được chấp nhận.
- Khi hành vi đổi: cập nhật PRD và traceability trước, sau đó domain/data/API/test specification.
- Khi quyết định kiến trúc đổi: tạo ADR mới và đánh dấu ADR cũ superseded; không xóa lịch sử.

## Phương án đã cân nhắc

### Mỗi feature có spec độc lập là nguồn sự thật

Dễ chia việc nhưng nhanh tạo trùng lặp BR/INV và khó biết bản nào mới nhất. Không chọn cho nhóm hai người ở giai đoạn đầu.

### Research document có quyền ngang PRD

Giữ được lịch sử nhưng các bản nghiên cứu chứa giả thuyết và mâu thuẫn chưa được duyệt. Không chọn.

## Hệ quả

- Mọi tài liệu kỹ thuật phải liên kết bằng mã yêu cầu thay vì chép lại nội dung dài.
- PRD cần bảng reconciliation khi hợp nhất research.
- Coding agent phải dừng và quay lại SPECIFY nếu yêu cầu không có trong PRD/traceability.
- Một thay đổi nhỏ có thể cần sửa nhiều contract, nhưng thứ tự cập nhật rõ ràng và kiểm tra được.

