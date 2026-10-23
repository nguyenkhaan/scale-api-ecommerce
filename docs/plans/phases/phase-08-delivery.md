# Phase 08 — Triển khai, báo cáo và bàn giao

**Phụ thuộc:** 06–07. **Mục tiêu:** `FR-044–046`, `NFR-11–14`, nghiệm thu PRD mục 15.

## Step 1 — Môi trường tái lập

- [ ] BE: hoàn chỉnh Compose cho API/worker/PostgreSQL/Redis/RabbitMQ/metrics, migration và seed/reset có guard; cấu hình AKS theo ADR-006.
- [ ] FE: build static app và kết nối đúng gateway ở Compose/AKS; kiểm cookie/CSRF/HTTPS trên môi trường mục tiêu.
- [ ] Kiểm: dựng lại từ hướng dẫn trên môi trường sạch, một và nhiều API pod; không tuyên bố HA ngoài phạm vi.

## Step 2 — Nghiệm thu sản phẩm

- [ ] BE: chạy TEST-01–29, oracle INV-01–12, fault/restart và lưu kết quả cùng phiên bản.
- [ ] FE: kiểm bằng trình duyệt hành trình khách và quản trị, thông báo, lỗi dòng, trạng thái chưa rõ, màn hình nhỏ/bàn phím.
- [ ] Kiểm: demo đủ các kịch bản PRD; phần P1 không đạt được ghi rõ thay đổi phạm vi đã duyệt.

## Step 3 — Báo cáo và bàn giao

- [ ] BE: lưu script/raw result E0–E4, cấu hình, trung vị/biến thiên, phân tích khóa/pool/cache/queue và giới hạn kết luận.
- [ ] FE: chụp/ghi luồng demo, hướng dẫn tài khoản thử nghiệm và các bước thao tác.
- [ ] Checkpoint: README, API, ADR, PRD, traceability và runbook khớp bản giao; người khác tái lập test và benchmark được.
