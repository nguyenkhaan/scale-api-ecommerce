# Phase 03 — Theo dõi backend và đo E0

**Thời gian:** Tuần 6–7  
**Phụ thuộc:** Phase 02  
**Kết quả cần đạt:** Có baseline đúng nghiệp vụ cho API, PostgreSQL và hộp thông báo trước khi thêm RabbitMQ hoặc nhiều API instance.

## Nghiên cứu cần làm

- [ ] **R03.1 — Học cách phát tải bằng k6.** Đọc [constant arrival rate](https://grafana.com/docs/k6/latest/using-k6/scenarios/executors/constant-arrival-rate/) để k6 giữ đều số lượt bắt đầu mỗi giây. Nhờ đó tải không tự giảm khi API chậm và kết quả phản ánh khả năng nhận request.
- [ ] **R03.2 — Định nghĩa chỉ số bán hàng.** Dùng k6, Prometheus và truy vấn PostgreSQL để tách request, đơn mới, đơn xác nhận, p95/p99, hết hàng, rate limit và lỗi hệ thống. Việc tách này tránh coi nhiều request bị từ chối là nhiều giao dịch thành công.
- [ ] **R03.3 — Đặt quy tắc so sánh.** Cố định seed PostgreSQL, workload k6, thời gian warm-up, CPU/RAM Docker Compose và số lần chạy. Như vậy E0 có thể làm mốc để so sánh RabbitMQ, pool và nhiều API mà không nhầm thay đổi khác.

## Việc triển khai và đo

- [ ] **C03.1 — Gắn request ID.** Dùng middleware Go/Gin và header NGINX để theo một request từ gateway tới endpoint. ID này giúp nối log lỗi với luồng đặt đơn mà không ghi thông tin cá nhân.
- [ ] **C03.2 — Đo thời gian API/DB.** Dùng Prometheus Go client để thu thời gian handler, truy vấn/transaction PostgreSQL và kết quả nghiệp vụ. Không dùng user ID hoặc order ID làm metric label vì chúng tạo vô hạn chuỗi số liệu.
- [ ] **C03.3 — Theo dõi tài nguyên bằng Prometheus/Grafana.** Vẽ CPU/RAM API, DB connections, lock wait, request latency và worker expiry. Dashboard cho biết checkout chậm do API, DB hay tiến trình trả hàng.
- [ ] **C03.4 — Tạo lại dữ liệu test.** Dùng SQL hoặc Go seed tool để reset chỉ database thử nghiệm về cùng sản phẩm, tồn, tài khoản và sale. Điều này giữ phép đo công bằng, không đụng dữ liệu thật.
- [ ] **C03.5 — Viết workload k6.** Tạo LOAD-01..10 gồm SKU phân tán/nóng, thiếu hàng, catalog cùng checkout, retry, callback lặp/trễ, worker cạnh tranh, service restart và RabbitMQ chậm. Các kịch bản này liên hệ trực tiếp tới nghiệp vụ phải scale.
- [ ] **C03.6 — Đo E0.** Chạy một Go/Gin API sau NGINX, tắt Redis cache và chưa chạy RabbitMQ. Dùng notification inbox ghi trong PostgreSQL như Phase 02 để đo baseline gồm chi phí checkout và chi phí notification đồng bộ.
- [ ] **C03.7 — Chốt tải tham chiếu.** Dùng kết quả k6 cùng dashboard Prometheus/Grafana chọn mức tải máy tạo phát đủ mà hệ thống vẫn hoạt động ổn định. Ghi CPU/RAM, PostgreSQL, dữ liệu, độ trễ checkout và notification để E1 chạy lại cùng điều kiện.

## Tiêu chí hoàn thành

- [ ] TEST-01..29 đã có kịch bản; các test hoàn chỉnh sẽ được chạy ở phase nghiệp vụ/queue tương ứng.
- [ ] E0 có ít nhất ba lần chạy với lệnh, commit, cấu hình và kết quả gốc.
- [ ] Báo riêng đơn thành công, lỗi nghiệp vụ, 429 và lỗi hệ thống.
- [ ] Có dashboard E0 và danh sách điểm nghẽn thấy được trong số liệu.
- [ ] Nếu k6 không phát đủ tải, đã kiểm tra CPU/RAM của máy phát tải trước khi kết luận về API.

**Đưa vào báo cáo:** Workload, điều kiện đo, số liệu E0 và chi phí ghi notification trực tiếp qua PostgreSQL.
