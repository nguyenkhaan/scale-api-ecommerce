# Phase 06 — Phần sản phẩm P1 (An)

**Phụ thuộc:** 05. Trước mỗi task, chi tiết hóa OpenAPI/traceability cho contract P1 chưa đủ. **Mục tiêu:** `FR-004/005/018/023/036/041/047–049`, `TEST-26–29`.

## Step 1 — Admin và cấu hình còn thiếu

- [ ] BE: API category admin, cấu hình cửa hàng, thống kê sale, tìm đơn/đối soát và ghi chú; ghi rõ quyền, schema và test.
- [ ] FE: màn hình thông tin cửa hàng, category, thống kê sale, đơn và đối soát tối giản.
- [ ] Kiểm: admin/owner guard, ghi chú không đổi order/inventory; build và browser acceptance.

## Step 2 — Thông báo bất đồng bộ

- [ ] BE: migration outbox/notification, ghi event cùng transaction order; publisher confirm, consumer ghi unique event ID rồi ACK, retry hữu hạn/DLQ.
- [ ] FE: hộp thông báo riêng tư, phân trang, đánh dấu đọc; trạng thái đơn vẫn lấy từ order API.
- [ ] Kiểm: `TEST-26–29`, RabbitMQ dừng không làm mất order, redelivery không trùng notification.

## Step 3 — Cổng P1

- [ ] BE: metric tuổi outbox, queue backlog/retry/DLQ; kiểm quyền và trường hợp poison message.
- [ ] FE: hoàn thiện form admin và hộp thông báo ở màn hình nhỏ/bàn phím.
- [ ] Checkpoint: mọi P1 còn trong phạm vi có contract và test; phần bị cắt phải được duyệt trong PRD, không tự bỏ.
