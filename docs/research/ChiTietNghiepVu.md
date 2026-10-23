
**CHI TIẾT**

**TẦNG NGHIỆP VỤ TRONG HỆ THỐNG**

## Các thành phần cốt lõi trong Topo nâng cao

### 1.1. Tầng biên & điều phối cổng vào

- CDN (Cloudflare): đảm nhiệm nhiệm vụ chặn DDOS, đồng thời cache các nội dung tĩnh tại cá cedge server gần người dùng nhất. Thực hiện triển khai Frontend trên Cloudflare để tận dụng sẵn hệ thống CDN.
- Load Balancer (Azure Standard Load Balancer): Tiếp nhận gói tin TCP thô, cân bằng tải ở tầng mạng. Thực hiện tích hợp qua mạng sau khi đã có source code
- Reverse Proxy & Gateway (Nginx): Giải mã TLS/SSL, áp dụng thêm rate limit để bảo vệ backend và phân phối request (theo cơ chế round robin hoặc least connection) tới các Go API instance

### 1.2. Stateless Compute Layer

- Các prod/container Go/Gin backend được cấu hình hoàn toàn ở trạng thái Stateless (không lưu session, idempotency key hay state trong RAM). Các phiên làm việc được xử lý bằng Cookie hoặc JWT Token
- Các pod được triển khai bằng AKS - hệ thống K8S được hỗ trợ bởi Azure.

### 1.3. Tầng lưu trữ dữ liệu

- PostgresSQL: Có tốc độ request chậm hơn so với MySQL hay Mariadb, nhưng đảm bảo phù hợp về tính toàn vẹn dữ liệu → Phù hợp với các ứng dụng đòi hỏi tính chính xác cao như sàn thương mại điện tử.
- Redis: Chỉ phục vụ mô hình Cache-side cho các dữ liệu đọc. Hỗ trợ cơ chế Cache-Hit để không tốn thời gian query vào CSDL

### 1.4. Tầng bất đồng bộ

- RabbitMQ: Đóng vai trò vùng đệm (buffer) cho các tác vụ tốn thời gian như Notification, gửi email SMS, hoặc phân tích số liệu.
- Transaction Outbox Pattern & Background Worker: đảm bảo khi checkout thành công, sự kiện thông báo sẽ được ghi vào outbox trong cùng transaction với đơn hàng. RabbitMQ sẽ đóng vai trò như queue truyền tải dữ liệu hoặc thực hiện cơ chế pub/sub. Các worker chạy ngầm sẽ thực hiện các công việc còn lại sau quá trình đặt hàng (gửi thông báo,...) để người dùng không cảm giác tốc độ bị chậm.
- Reversation Expiry Worker: Quét và giải phóng các đơn hàng quá hạn thanh toán (10 phút), hoàn trả lại các thông tin tồn kho vật lý một cách an toàn.

### 1.5. Giám sát và kiểm thử

- Sử dụng Prometheus để thu thập metrics
- Grafana trực quan hóa và cảnh báo nút thắt cổ chai
- K6 để test cân bằng tải.

## Một số bài toán yêu cầu khi thực hiện Scale API

Khi mở rộng tầng API từ 1 lên nhiều instance, hệ thống sẽ đối diện với 5 bài toán

| **STT** | **Bài toán khi Scale**                             | **Rủi ro nếu không xử lý đúng**                                                                                                                                              | **Giải pháp kỹ thuật bắt buộc**                                                                                                                                                                                |
| ------------- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1             | Bán vượt tồn kho                                       | Nhiều API instance cùng đọc tồn kho cùng 1 lúc                                                                                                                                     | Sử dụng PostgresSQL Transaction với câu lệnh cập nhật nguyên tử hoặc thực hiện khóa mutex lock. Sắp xếp khóa SKU theo thứ tự ID tăng dần để triệt tiêu nguy cơ Deadlock.                          |
| ---           | ---                                                        | ---                                                                                                                                                                                       | ---                                                                                                                                                                                                                        |
| 2             | Xử lý tặp & Tính bất biến                            | Khách bấm đặt hàng nhiều lần hoặc mạng chập trơn, retry, 2 request rơi vào                                                                                                   | Distributed Idempotency Key: Bắt buộc client gửi idempotency key (uuid). Lưu khóa vào database (redis/postgresSQL) để chặn đứng request thứ 2 ngay tại cửa, trả về đúng kết quả của request ban đầu |
| ---           | ---                                                        | ---                                                                                                                                                                                       | ---                                                                                                                                                                                                                        |
| 3             | Cạn kiệt tài nguyên Downstream (Thundering Herb to DB) | Khi mở rộng 5 - 10 instance, mỗi instance mở 50 kết nối → 500 kết nối cùng dập vào database cùng 1 lúc → DB cạn RAM/CPU do chi phí context-switching giữa các progress | Định cỡ Connection Pool trần: Tối ưu truy vấn SQL, giảm thời gian transaction giữ kết nối xuống < 10msTriển khai thêm nhiều database, hạn chế tình trạng bottle threck cho database                    |
| ---           | ---                                                        | ---                                                                                                                                                                                       | ---                                                                                                                                                                                                                        |
| 4             | Nghẽn cổ chai & Hiệu ứng sâu Domino                   | Cổng thanh toán bên thứ 3 bị chậm. Hàng nghìn request bị treo socket ở Go API. Cạn goroutine và worker pool                                                                   | Cấu hình HTTP Connection Pool tối ưu; thiết lập Timeout nhiều lớp; tích hợp Circuit Breaker (ngắt mạch khi lỗi vượt ngưỡng).                                                                              |
| ---           | ---                                                        | ---                                                                                                                                                                                       | ---                                                                                                                                                                                                                        |
| 5             | Cache Stampede / Dog-pilling                               | Khi một SKU flash sale cực hot hết hạn cache (TTL expired), hàng nghìn request từ các instance cùng miss cache và ùa vào query database                                       | Áp dụng kỹ thuật Singleflight trong Go để gộp các request đọc trùng lặp thành một truy vấn DB duy nhất; áp dụng Mutex hoặc Probabilistic Early Expiration                                               |
| ---           | ---                                                        | ---                                                                                                                                                                                       | ---                                                                                                                                                                                                                        |

## Tối ưu hóa các khớp nối & tinh chỉnh thông số HTTP Pool

Trong hệ thống phân tán, các thành phần không đứng độc lập mà liên kết với nhau qua các khớp nối (gọi là joints). Độ trễ và thông lượng của toàn hệ thống bị giới hạn bởi các điểm nghẽn tại các joints này.

Làm sao để tối ưu hóa thông số HTTP Pool?

HTTP Pool quản lý các kết nối TCP ra ngoài. Nếu không tối ưu, hệ thống sẽ liên tục mở và đóng socket TCP → cạn kiệt port mạng cục bộ và độ trễ sẽ tăng vọt do quá trình bắt tay 3 bước ở mỗi request.

Một số phương pháp tối ưu trong Go

- Tuyệt đối không dùng http.DefaultClient: mặc định Go có MaxIdleConsPerHost = 2. Nếu có 200 request đồng thời gọi tới cổng thanh toán, chỉ có 2 kết nối được giữ lại, 198 kết nối còn lại bị đóng ngay lập tức và mở mới liên tục.
- Khởi tạo Singleton http

## Phân loại chi tiết các khớp nối & cấu hình

### 4.1. Khớp nối Deep API (HTTP ngoại vi/Payment Simulator)

- Bản chất: Lời gọi mạng từ backend sang dịch vụ đối tác
- Rủi ro: Bị động về độ trễ
- Cách tinh chỉnh: Cấu hình http.Transport như phân tích ở Mục 3, kết hợp Timeout retry và cơ chế ngắt dịch vụ thay vì chờ vô hạn.

### 4.2. Khớp nối database

- Bản chất: Thực thi câu lệnh SQL, ACID transaction và cơ chế khóa hàng tồn

### 4.3. Khớp nối Cache

- Bản chất: Giảm tải đọc cho PostgesSQL đối với dữ liệu sản phẩm, catalog công khai
- Lưu ý: Cache chỉ mang tính chất tăng tốc đọc, phục vụ các Read Heavy Task. Hệ thống phải Curcuit Break sang đọc trực tiếp trong database khi gặp Cache Miss.

### 4.4. Khớp NGINIX &lt;-&gt; Go API

- Bản chất: Điểm tiếp nhận request từ client và chuyển tiếp vào mạng nội bộ backend

### 4.5. Khớp nối Messge Broker

- Bản chất: Đưa các xử lý nặng hoặc không khẩn cấp (gửi email, tạo notification, thống kê) ra khỏi luồng chính.

## 6\. Kết luận

Giữ nguyên 100% phần cứng hạ tầng, chỉ bằng cách chia nhỏ Topo Tầng Compute, kết hợp với tối ưu thông số các khớp nối và Event Drive. Hệ thống đặt mục tiêu:

- Tăng thông lượng lên 3.5 lần (320 RPS -> 1150 RPS)
- Giảm độ trỡ p95 tới 8 lần
- Triệt tiêu lỗi cạn kiệt tài nguyên
- Bảo toàn 100% các bất biến: zero overselling, all-or-nothing và Idempotency.
