# Phase 05 — Thanh toán mô phỏng và vòng đời đơn (Lân)

**Phụ thuộc:** 04; chốt TECH-02 và OPEN-02 trước nghiệm thu. **Mục tiêu:** `FR-030–040`, `INV-06–08/11/12`, `TEST-11–19/25`.

## Step 1 — Phiên và callback

- [ ] BE: một payment session/đơn; simulator tạo success/failure/delay/duplicate; callback xác thực nguồn, phiên, tiền, event ID và lưu bền vững.
- [ ] FE: mở lại cùng phiên, theo dõi trạng thái từ order API; điều khiển simulator chỉ hiện ở môi trường test đúng quyền.
- [ ] Kiểm: `TEST-11/12/17/18`, event lặp/sai/mâu thuẫn không trừ tồn thêm.

## Step 2 — Kết thúc reservation

- [ ] BE: một đường transition có guard cho confirmed/failed/cancelled/expired; worker quét batch từ PostgreSQL, chạy lại an toàn; case đối soát cho success muộn.
- [ ] FE: danh sách/chi tiết đơn, hủy hợp lệ, hiển thị expired dù worker chưa cập nhật và trạng thái cần đối soát.
- [ ] Kiểm: `TEST-13–16/19/25`, callback đối đầu hủy/hết hạn, restart và nhiều worker; oracle INV-06–08/12.

## Step 3 — Baseline đúng nghiệp vụ

- [ ] BE: hoàn thiện metrics order/payment/expiry, seed/reset và oracle SQL đầy đủ.
- [ ] FE: chạy hành trình khách từ xem hàng tới kết quả đơn; kiểm trạng thái tải/lỗi.
- [ ] Checkpoint: `TEST-01–25` qua trong lab, INV-01–12 không vi phạm; mới được đo E0.
