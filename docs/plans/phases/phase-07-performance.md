# Phase 07 — Đo, tối ưu và mở rộng (Lân)

**Phụ thuộc:** 05 để chạy E0; 06 để chạy E1. **Mục tiêu:** `PERF-01–09`, `LOAD-01–10`, `DEC-14/15`.

## Step 1 — E0 và tối ưu BE

- [ ] BE: viết k6 workloads, seed/reset/oracle; đo E0 chỉ sau TEST-01–25, lưu commit/tài nguyên/pool/latency theo outcome.
- [ ] FE: kiểm hành trình demo trên bản benchmark, đảm bảo UI không retry bằng key mới khi kết quả chưa rõ.
- [ ] Kiểm: ít nhất 3 lần/cấu hình, tách đủ hàng và khan hiếm; chốt OPEN-01 từ E0.

## Step 2 — E1 và E2

- [ ] BE: so E1 sau outbox/RabbitMQ và tối ưu transaction/query/pool từng thay đổi có đo; E2 với 1 và 2+ API instance.
- [ ] FE: kiểm session và checkout khi request qua các instance; đo cảm nhận luồng mua trên tải vừa.
- [ ] Kiểm: oracle INV sau mỗi run, lock/pool wait và thông báo trễ; không gán lợi ích cho nhiều thay đổi gộp.

## Step 3 — E3/E4

- [ ] BE: Redis cache chỉ cho catalog công khai, fallback hữu hạn; rate limit theo nhóm route; đo E3 đọc hỗn hợp và E4 quá tải.
- [ ] FE: xử lý 429/retry guidance, cache cũ được backend sửa, không tự coi lỗi hệ thống là hết hàng.
- [ ] Checkpoint: `TEST-22/23`, LOAD tương ứng, cache outage, p95/p99 và lỗi hệ thống báo theo tải tham chiếu; chốt OPEN-03 sau E0/E1.
