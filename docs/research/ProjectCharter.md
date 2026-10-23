
# 1. Document Control

## 1\. 1 Document Information

## 1.2 Document History

| **Version** | **Issue Date** | **Changes**                                                              |
| ----------------- | -------------------- | ------------------------------------------------------------------------------ |
| 1.0               | 23/09/2026           | Initial Version. Nộp bù cho cột mốc gốc ngày 05/09/2026 theo Master Plan |
| ---               | ---                  | ---                                                                            |

## 1.3 Document Approvals

| **Role**  | **Name**          | **Signature** | **Date** |
| --------------- | ----------------------- | ------------------- | -------------- |
| Project Sponsor | ThS. Nguyễn Công Hoan |                     |                |
| ---             | ---                     | ---                 | ---            |
| Project Manager | Nguyễn Khả An         | ✔                  | 23/09/2026     |
| ---             | ---                     | ---                 | ---            |
| Team Member     | Trần Nguyễn Lân      | ✔                  | 23/09/2026     |
| ---             | ---                     | ---                 | ---            |

## 1.4 Acronyms List

| **Acronym** | **Full Form**                            | **Description**                                                             |
| ----------------- | ---------------------------------------------- | --------------------------------------------------------------------------------- |
| CSKH              | Chăm sóc khách hàng                        | Nghiệp vụ được chọn làm bối cảnh ứng dụng cho AI Agent trong đồ án. |
| ---               | ---                                            | ---                                                                               |
| TMĐT             | Thương mại điện tử                       | Miền nghiệp vụ mô phỏng: các shop bán hàng trên sàn TMĐT.              |
| ---               | ---                                            | ---                                                                               |
| PoC               | Proof of Concept                               | Sản phẩm thử nghiệm chứng minh tính khả thi của kiến trúc.              |
| ---               | ---                                            | ---                                                                               |
| API               | Application Programming Interface              | Giao diện lập trình giữa các thành phần hệ thống.                        |
| ---               | ---                                            | ---                                                                               |
| JWT               | JSON Web Token                                 | Cơ chế xác thực/ủy quyền không trạng thái cho API.                       |
| ---               | ---                                            | ---                                                                               |
| RPM               | Requests Per Minute                            | Giới hạn số request mỗi phút của nhà cung cấp LLM.                        |
| ---               | ---                                            | ---                                                                               |
| p95               | 95th Percentile Latency                        | Độ trễ mà 95% request không vượt quá — chỉ số đo hiệu năng.         |
| ---               | ---                                            | ---                                                                               |
| CI/CD             | Continuous Integration / Continuous Deployment | Tự động build, test qua GitHub Actions.                                        |
| ---               | ---                                            | ---                                                                               |
| GVHD              | Giảng viên hướng dẫn                      | Người hướng dẫn, đánh giá tiến độ và nghiệm thu đồ án.            |
| ---               | ---                                            | ---                                                                               |
| PM                | Project Manager                                | Người quản lý phạm vi, tiến độ và chất lượng dự án.                 |
| ---               | ---                                            | ---                                                                               |
| BA                | Business Analyst                               | Vai trò phân tích nghiệp vụ/yêu cầu — PM kiêm nhiệm.                    |
| ---               | ---                                            | ---                                                                               |
| QA                | Quality Assurance                              | Đảm bảo chất lượng — kiểm thử và tự rà soát.                         |
| ---               | ---                                            | ---                                                                               |
| CNPM              | Công nghệ Phần mềm                         | Khoa quản lý và chấm đồ án (Khoa CNPM - UIT).                              |
| ---               | ---                                            | ---                                                                               |
| SE                | Software Engineering                           | Mã ngành dùng trong Document Id (UIT.SE.68).                                   |
| ---               | ---                                            | ---                                                                               |

# 2\. Project Summary, Purpose, Goals, and Success Criteria

## 2.1 Summary

**Đồ án 1 (SE121)**, nhóm 2 sinh viên, định hướng **Fullstack/PM**, hướng dẫn bởi **ThS. Nguyễn Công Hoan**. Đồ án tập trung vào việc tối ưu hóa được các API bên trong một hệ thống phân tán. Đồ án xây dựng và đánh giá một Backend Flash Sale dành cho doanh nghiệp, vận hành một cửa hàng thương mại điện tử độc lập. Sản phẩm cho phép quản trị các thành phần cần thiết của một mô hình thương mại điện tử: hàng hóa, SKU, tồn kho, chương trình flash sale và đơn hàng; Khách hàng có thể xem sản phẩm, tạo giỏ gồm nhiều mặt hàng, đặt đơn và thực hiện thanh toán mô phỏng.

Đồ án trọng tâm vào việc tối ưu tầng API đặt đơn dưới áp lực của việc tải đồng thời, trong khi vẫn đảm bảo nhất quán trong dữ liệu: giữ hàng theo nguyên tắc toàn bộ hoặc không giữ, và xử lý an toàn các tình huống retry, hết hạn đơn hoặc phản hồi thanh toán đến muộn.

Toàn bộ quá trình phát triển áp dụng theo quy trình Test-Drive Development với AI Coding Agent. Xây dựng baseline và thực hiện mô phỏng thông qua các dịch vụ test chịu tải (AWS K6).

## 2.2 Purpose and Vision

- **Purpose:** Hệ thống nghiên cứu cách thiết kế và tối ưu backend cho nghiệp vụ flash sale, nơi nhiều khách hàng đồng thời cạnh tranh số lượng hàng hóa có giới hạn và có thể đặt đơn chứa nhiều SKU. Đồ án hướng đến giải quyết đồng thời hai vấn đề: duy trì phản hồi ổn định khi tải tăng cao và đảm bảo tính đúng đắn của dữ liệu (đơn hàng, tồn kho, thanh toán…). Kết quả được đánh giá bằng kiểm thử các biến nghiệp vụ cùng số liệu thực nghiệm về độ trễ, throughput, lỗi hệ thống, tranh chấp khóa và hiệu quả mở rộng API.
- **Vision:** Hệ thống sau thực nghiệm có thể thực hiện quá trình mở rộng, xây dựng một hệ thống Backend bán hàng có thể triển khai độc lập cho từng doanh nghiệp. Hệ thống đủ vững chãi cho các đợt flash sale. PoC hiện tại ưu tiên vào kiến trúc và lý thuyết bên trong, thay vì mở rộng sớm. Các kết quả về transaction, idempotency, cache, giới hạn tải, worker sẽ là cơ sở để phát triển hệ thống trong các giai đoạn sau này.

## 2.3 Project Goals and Objectives

### 2.3.1 Goals

- **G1 - 13/10/2026:** Hoàn thành và chốt PRD, phạm vi MVP, các quy tắc nghiệp vụ, kiến trúc hệ thống và sheet kế hoạch trước **13/10/2026**.
- **G2 - 04/11/2026:** Hoàn thành các chức năng nền tảng của hệ thống: tài khoản, catalog, SKU, quản lý tồn kho, flash sale, giỏ hàng, hóa đơn. Có thể demo được luồng bán hàng đơn giản.
- **G3 - 21/11/2026**: Hoàn thành PoC đầu - cuối cho checkout nguyên tử, idempotency, thanh toán mô phỏng, hủy/hết hạn đơn và giao diện demo. Thiết lập baseline đo hiệu suất.
- **G4 - 01/12/2026:** Thực hiện kiểm tra bằng các công cụ kiểm tra chịu tải. Chỉnh sửa mã nguồn.
- **G5 - 29/11/2026:** hoàn thiện và thử nghiệm. Bàn giao mã nguồn, video demo và đồ án.

### 2.3.2 Objectives

- **O1:** Xây dựng backend bằng Go và Gin cho một doanh nghiệp vận hành; bao gồm các chức năng cơ bản: xác thực, quản lý sản phẩm, danh mục, SKU, tồn kho, flash sale, giỏ hàng, thông tin đơn hàng.
- **O2**: Hiện thực luồng đặt đơn hàng nhiều mặt hàng. Quản lý việc tạo đơn nguyên tử: Nếu một dòng không hợp lệ thì lập tức từ chối toàn bộ, chỉ rõ nguyên nhân.
- **O3:** Xây dựng mô phỏng thanh toán bất đồng bộ. Quản lý đầy đủ vòng đời của đơn hàng: xác nhận thanh toán, thất bại, hủy, hết hạn và đối soát kết quả đến muộn. Reservation phải chỉ được chuyển thành hàng bán hoặc giải phóng đúng một lần, kể cả khi callback, worker và thao tác hủy diễn ra đồng thời.
- **O4:** Thực hiện kiểm thử nghiệp vụ. Đánh giá hiệu quả của API trên nhiều instance, với các cấu hình khác nhau (hệ thống Backend sẽ tạo ra nhiều instance, thực hiện phép đo trên từng cấu hình khác nhau của instance).
- **O5:** Phân tích kết quả và rút ra được kết luận cho hệ thống. Bản bàn giao sẽ bao gồm: mã nguồn, benchmark, kịch bản kiểm thử, tài liệu dự án liên quan (có hướng dẫn tái lập kiểm tra), UI demo, báo cáo kết quả thực nghiệm.

## 2.4 Success Criteria

- **Stakeholder Satisfaction:** GVHD nghiệm thu demo đủ 3 kịch bản, không có lỗi nghiêm trọng.
- **Deliverables:** Nộp đủ sản phẩm trên sheet.
- **Traceability:** Cập nhật đầy đủ tài liệu liên quan (specs, documents, prd, ard…) và tập prompt được lưu trong Git, truy vết được từ yêu cầu đến mã nguồn và test.

# 3\. Project Scope, Quality Management, and Timeline (Nghiên cứu và chỉnh sửa)

## 3.1 Project Scope

### 3.1.1 Inclusions

- Xây dựng Backend flash sale cho một doanh nghiệp vận hành cửa hàng độc lập; mỗi bản triển khai có dữ liệu và cấu hình cửa hàng riêng.
- Hiện thực xác thực tài khoản, đăng ký/đăng nhập/đăng xuất và phân quyền cơ bản giữa khách hàng với quản trị viên.
- Quản lý danh mục, sản phẩm vật lý, biến thể SKU và tồn kho; hỗ trợ hiển thị, tìm kiếm, lọc và phân trang catalog ở mức cần thiết cho demo.
- Quản lý chương trình flash sale, gồm cấu hình SKU tham gia, giá ưu đãi, quota, giới hạn mua, thời gian mở/đóng và trạng thái chương trình.
- Xây dựng giỏ hàng nhiều mặt hàng, chức năng xem trước đơn, xác nhận lại khi giá hoặc điều kiện mua thay đổi, và ghi nhận thông tin người nhận.
- Hiện thực checkout nguyên tử cho toàn bộ giỏ: chỉ giữ hàng khi mọi dòng hợp lệ; nếu thiếu bất kỳ SKU nào thì từ chối toàn bộ đơn và không để lại reservation đã commit.
- Bảo đảm idempotency khi đặt đơn, hỗ trợ retry hoặc tra cứu kết quả khi mất phản hồi; một khách hàng với cùng khóa idempotency chỉ tạo tối đa một đơn.
- Xây dựng mô phỏng thanh toán bất đồng bộ, xử lý trạng thái xác nhận, thất bại, hủy và hết hạn đơn; hỗ trợ phát hiện thanh toán đến muộn hoặc mâu thuẫn để quản trị đối soát.
- Triển khai worker xử lý đơn quá hạn, có khả năng chạy lại sau gián đoạn và không giải phóng reservation trùng lặp.
- Nghiên cứu và đánh giá các hướng tối ưu tầng API: transaction ngắn, truy vấn/index, connection pool, timeout/retry, nhiều API instance, cache đọc, kiểm soát tải và xử lý tác vụ nền.
- Thiết lập môi trường thực nghiệm với Go/Gin; PostgreSQL cho dữ liệu nghiệp vụ, Redis cho cache đọc, NGINX cho cân bằng tải và k6 để kiểm thử tải.
- Cung cấp metrics, log liên kết và dữ liệu quan sát cho endpoint, kết quả nghiệp vụ, lỗi, tài nguyên và các điểm nghẽn như chờ khóa hoặc chờ pool.
- Thực hiện bộ kiểm thử nghiệp vụ, đồng thời và chịu lỗi; bao gồm tranh chấp tồn kho, retry qua nhiều instance, callback thanh toán lặp/trễ, worker khởi động lại, Redis lỗi và giới hạn tải.
- Xây dựng giao diện demo tối thiểu cho luồng khách hàng và quản trị viên; chuẩn bị mã nguồn, hướng dẫn dựng môi trường, dữ liệu demo/benchmark có thể tái tạo, tài liệu API và báo cáo kết quả thực nghiệm.

### 3.1.2 Exclusions (Mai xem lại phần Outscope bên trong dự án, có một số cái kỳ lạ)

- Không xây dựng một trang web thương mại điện tử dùng chung cho nhiều doanh nghiệp, nhiều người bán, theo một cơ chế nghiệp vụ bắt buộc
- Không thanh toán tiền thật, chỉ dùng sandbox
- Không triển khai nghiệp vụ giao hàng, vận chuyển, logistic…
- Không tách **tập sản phẩm khách đã chọn trong một lần checkout** thành nhiều đơn thanh toán độc lập; nếu bất kỳ SKU nào trong tập đã chọn không đủ điều kiện, hệ thống từ chối toàn bộ lần đặt thay vì tự động mua một phần. Khách có thể điều chỉnh lựa chọn trong giỏ và thực hiện một lần đặt đơn mới.
- **Không phát triển voucher, mã giảm giá, điểm thưởng, thành viên** hoặc các cơ chế ưu đãi cộng dồn. Cơ chế giảm giá duy nhất trong phạm vi MVP là giá ưu đãi theo chương trình flash sale, được cấu hình theo SKU, thời gian, quota và giới hạn mua.
- Không xây dựng giao diện TMĐT hoàn chỉnh, UI chỉ phục vụ cho việc kiểm chứng và demo dự án.
- Không triển khai microservice, k8s
- Không tuyên bố hệ thống là production, sẵn sàng cho việc phục vụ.

**3.2 Quality Management**

**3.2.1 Standards**

- Các API đọc (catalog sản phẩm, chi tiết flash sale, giỏ hàng) đạt p95 < 150 ms ở mức 500 – 1.000 người dùng ảo đồng thời trên 1 API instance.
- API đặt đơn/checkout (nghiệp vụ ghi có tranh chấp dữ liệu, kiểm tra tồn kho và khóa reservation) đạt p95 < 500 ms dưới áp lực tải cao điểm.
- Thông lượng hệ thống (RPS) tăng ≥ 2,5 lần khi mở rộng quy mô từ 1 lên 3 API instance phía sau NGINX.
- 0% sai lệch tồn kho và quota flash sale (Zero Overselling); đảm bảo tính toàn vẹn dữ liệu khi có nhiều giao dịch cạnh tranh đồng thời.
- 100% đơn hàng áp dụng nguyên tắc nguyên tử (All-or-nothing); nếu bất kỳ SKU nào trong giỏ không đủ hàng thì lập tức từ chối toàn bộ đơn và không để lại reservation rác.
- 100% request trùng lặp (cùng Idempotency Key khi retry) chỉ sinh ra đúng 1 đơn hàng duy nhất, không trừ kho lặp.
- Worker xử lý đơn quá hạn hoàn trả 100% lượng hàng giữ về kho chính xác, không giải phóng trùng lặp hoặc sót hàng khi gặp gián đoạn.
- Tỷ lệ request lỗi hệ thống (5xx) dưới kịch bản tải định mức < 0,5%.
- 100% endpoint nhạy cảm (giỏ hàng, checkout, thông tin đơn hàng, API quản trị) bắt buộc xác thực qua JWT và phân quyền đúng vai trò (Customer vs Admin).
- Không lưu API key, mật khẩu cơ sở dữ liệu hay secret trong repository; 100% cấu hình qua biến môi trường (.env).
- Tối đa 2 lỗi nhỏ (minor) khi nghiệm thu; không có lỗi nghiêm trọng (crash service, bán âm tồn kho, deadlock cạn connection pool, sai lệch thanh toán) khi demo.

**3.2.2 Control Procedures**

- SDD: Mỗi tính năng có spec.md → plan.md → tasks.md trong thư mục specs/ trước khi code; bước Analyze kiểm tra tính nhất quán giữa ba tài liệu.
- AI Coding Agent theo vai trò: Architect/Reviewer dùng model suy luận mạnh để rà soát thiết kế và bảo mật; Coder dùng model nhanh/chi phí thấp; Coding Agent không được mở rộng phạm vi ngoài task.
- Unit Test: Dùng framework testing chuẩn của Go và thư viện testify cho tầng Domain/Application, coverage tối thiểu 60%.
- Integration Test: Dùng net/http/httptest và cơ sở dữ liệu test để kiểm thử toàn diện các route Gin API, transaction PostgreSQL và cache Redis.
- Load Test: Dùng k6 chạy theo kịch bản cố định (ramp-up, spike test, concurrency checkout), tự động lưu kết quả benchmark vào repository để so sánh định lượng giữa các lần chạy.
- CI: GitHub Actions tự động kiểm tra định dạng (golangci-lint), build và chạy test cho mỗi Pull Request trước khi merge vào main.
- Code review: Mọi Pull Request được review bởi thành viên còn lại và Reviewer Agent; PM duyệt merge.
- Quản lý tiến độ: Họp nhóm 2 lần/tuần; theo dõi task trên GitHub Projects; cập nhật link sản phẩm lên Master Plan tại mỗi mốc.
- UAT: GVHD là người đánh giá và nghiệm thu cuối cùng thông qua 3 kịch bản demo thực tế.

**3.2.3 Cam kết bảo mật (NDA)**

- Chỉ sử dụng dữ liệu mô phỏng; không sử dụng thông tin thật hoặc dữ liệu nhạy cảm trong quá trình phát triển và kiểm thử.
- Khóa bí mật JWT và thông tin tích hợp cổng thanh toán mô phỏng (PayOS Sandbox) được quản lý riêng qua biến môi trường, tuyệt đối không chia sẻ công khai.
- Không tiết lộ mã nguồn, nội dung đồ án chưa công bố cho bên thứ ba trước thời điểm bảo vệ chính thức.
- Tuân thủ quy định bảo mật thông tin và liêm chính học thuật của Khoa CNPM - UIT liên quan đến tài liệu và quy trình chấm đồ án.

## 3.3 Timeline and Milestones

### 3.3.1 Kế hoạch giai đoạn Thực hiện

| **Name**              | **Role**           | **Power** | **Interest** | **Contact Information**                          |
| --------------------------- | ------------------------ | --------------- | ------------------ | ------------------------------------------------------ |
| ThS. Nguyễn Công Hoan     | Project Sponsor, GVHD    | Cao             | Cao                | [hoannc@uit.edu.vn](mailto:hoannc@uit.edu.vn)           |
| ---                         | ---                      | ---             | ---                | ---                                                    |
| Hội đồng Khoa CNPM - UIT | Đánh giá, nghiệm thu | Cao             | Trung bình        |                                                        |
| ---                         | ---                      | ---             | ---                | ---                                                    |
| Nguyễn Khả An             | Project Manager          | Trung bình     | Cao                | [24520059@gm.uit.edu.vn](mailto:24520059@gm.uit.edu.vn) |
| ---                         | ---                      | ---             | ---                | ---                                                    |
| Trần Nguyễn Lân          | Team Member              | Trung bình     | Cao                | [24520945@gm.uit.edu.vn](mailto:24520945@gm.uit.edu.vn) |
| ---                         | ---                      | ---             | ---                | ---                                                    |
| Doanh nghiệp kinh doanh    | Người thuê hệ thống | Thấp           | Cao                |                                                        |
| ---                         | ---                      | ---             | ---                | ---                                                    |

## 3.6 Risk Management

| **Risk**                                      | **Likelihood** | **Impact** | **Mitigation**                                                                                                                                                                                        |
| --------------------------------------------------- | -------------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R1: Trễ ba mốc đầu làm dồn tiến độ         | Cao                  | Cao              | Nộp bù theo thứ tự. Cố gắng hoàn thiện các deadline phía sau trước thời hạn, tối ưu thời gian.                                                                                             |
| ---                                                 | ---                  | ---              | ---                                                                                                                                                                                                         |
| R2: Mất nhiệt trong lúc làm                     | Trung bình          | Cao              | Cố gắng hoàn thành deadline trước thời hạn, dồn hạn chót của công việc lên để có tinh thần chạy deadline. Yêu cầu mỗi 2- 3 ngày phải có công việc / task mới để thực hiện. |
| ---                                                 | ---                  | ---              | ---                                                                                                                                                                                                         |
| R3: Phình phạm vi do muốn nâng cấp tính năng | Trung bình          | Cao              | Ưu tiên MVP; tính năng ngoài phạm vi ghi vào backlog Đồ án 2; mọi thay đổi ghi vào Document History.                                                                                          |
| ---                                                 | ---                  | ---              | ---                                                                                                                                                                                                         |
| R4: Hết hạn mức Coding Agent Codex               | Trung bình          | Trung bình      | Sử dụng các Model Codex mạnh cho việc lên kế hoạch và tài liệu. Các model yếu - trung bình cho việc coding.                                                                                  |
| ---                                                 | ---                  | ---              | ---                                                                                                                                                                                                         |
| R5: GVHD phản hồi chậm                           | Cao                  | Trung bình      | Ghi các quyết định vào bên trong docs/specs. Trao đổi với nhau qua messenger.                                                                                                                      |
| ---                                                 | ---                  | ---              | ---                                                                                                                                                                                                         |

## 3.7 Assumptions and Constraints

### 3.7.1 Assumptions

- GVHD chấp nhận nộp bù các mốc đã trễ.
- Máy cá nhân của thành viên đủ cấu hình chạy Docker với toàn bộ hạ tầng cục bộ.
- Tài khoản cá nhân của từng thành viên được hỗ trợ AWS để kiểm thử.

### 3.7.2 Constraints

- 2 thành viên, khoảng 12 tuần
- **Công nghệ:** Backend bắt buộc sử dụng Go và Gin. PostgreSQL, Redis, NGINX và k6 là các công nghệ được định hướng cho lưu trữ dữ liệu nghiệp vụ, cache đọc, cân bằng tải và kiểm thử hiệu năng. Kiến trúc MVP là ứng dụng backend theo module; không đặt mục tiêu tách microservices, Kubernetes, sharding hoặc high availability cho database.
- Sử dụng tài khoản Codex Agent phục vụ cho việc coding và hoàn thành dự án.

# 4\. Additional Information

## 4.1 Technical Stack

| **Nhóm**             | **Công nghệ** | **Vai trò**                                                                                                                                                                                                                 | **Trạng thái**                   |
| --------------------------- | --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| Backend                     | Go + Gin              | Xây dựng API và các module nghiệp vụ                                                                                                                                                                                         | Đã xác nhận                          |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Kiến trúc ứng dụng      | Modular Monolithic    | Giữ phạm vi phù hợp đồ án; xử lý tác vụ hết hạn ngoài luồng checkout                                                                                                                                                | Đã xác nhận                          |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| CSDL                        | PostgresSQL           | Nguồn dữ liệu có thẩm quyền cho đơn hàng, tồn kho, reservation, quota và idempotency                                                                                                                                    | Đã xác nhận                          |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Cache                       | Redis                 | Cache-aside cho dữ liệu đọc như catalog/sản phẩm; không quyết định tồn kho, giá hoặc quyền tạo đơn                                                                                                               | Đề xuất cho thực nghiệm             |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Gateway và cân bằng tải | Nginix                | Điểm vào hệ thống, phân phối request tới một hoặc nhiều API instance; hỗ trợ kiểm soát tải theo cấu hình                                                                                                         | Đề xuất cho thực nghiệm             |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Thanh toán                 | PayOS + Sandbox       | Mô phỏng chức năng thanh toán trong bán hàng. Có thể chủ động nghiên cứu để tạo các tình huống thành công, thất bại, mất phản hồi, sự kiện lặp, chậm hoặc sai thứ tự để kiểm thử nghiệp vụ | Đề xuất cho thực nghiệm             |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Kiểm thử tải             | K6                    | Phát tải, đo độ trễ, throughput, lỗi, khả năng phục hồi và so sánh cấu hình một/nhiều API instance                                                                                                                | Đề xuất cho thực nghiệm             |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Bảo mật                   | JWT + Validation      | Bảo vệ API riêng tư, dữ liệu giỏ/đơn/thanh toán và bí mật hệ thống                                                                                                                                                  | Bắt buộc                               |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |
| Benchmark                   |                       | Bộ test nghiệp vụ, phục vụ cho việc tái lập quá trình kiểm thử, tạo tính thuyết phục cho dự án.                                                                                                                  | Chưa có công nghệ cụ thểBắt buộc |
| ---                         | ---                   | ---                                                                                                                                                                                                                                | ---                                      |

## 4.2 Kiến trúc tổng quan (dự kiến)

Kiến trúc tổng quan được trình bày bên trong tài liệu [architecture.md](../technical/architecture.md) của Repository dự án. Đảm bảo thành viên có thể đồng bộ được sự thay đổi.

## 4.3 Lộ trình mở rộng

| **Giai đoạn** | **Định hướng**  | **Phạm vi dự kiến**                                                                                                                                                                                                                                                                                                                               |
| --------------------- | ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Đồ án 1            | PoC                       | Xây dựng backend flash sale theo kiến trúc module cho một doanh nghiệp; hiện thực catalog, tồn kho, flash sale, giỏ nhiều SKU, checkout nguyên tử, idempotency, thanh toán mô phỏng và worker hết hạn. Đo baseline và hiệu quả tối ưu khi chạy một/nhiều API instance.                                                         |
| ---                   | ---                       | ---                                                                                                                                                                                                                                                                                                                                                        |
| Đồ án 2            | Phát triển hệ thống   | Phát triển các năng lực được chứng minh cần thiết sau PoC: tự động hóa triển khai, giám sát vận hành, mở rộng tích hợp và tách các thành phần khi dữ liệu đo cho thấy modular monolith không còn đáp ứng yêu cầu. Có thể xem xét tích hợp sandbox thanh toán hoặc các nghiệp vụ thương mại mở rộng. |
| ---                   | ---                       | ---                                                                                                                                                                                                                                                                                                                                                        |
| Khóa luận           | Nghiên cứu chuyên sâu | Nghiên cứu cơ chế duy trì tính đúng đắn và khả năng phục hồi cho nghiệp vụ đặt đơn phân tán: kiểm soát tranh chấp SKU nóng, điều phối tác vụ nền, chống xử lý lặp, giới hạn tải và đánh giá các đánh đổi giữa độ nhất quán, độ trễ và khả năng mở rộng.                                      |
| ---                   | ---                       | ---                                                                                                                                                                                                                                                                                                                                                        |

Tài liệu này cung cấp một cái nhìn tổng quan ở cấp cao về dự án "Scaling API trong hệ thống phân tán", đảm bảo sự phù hợp với các mục tiêu tổ chức và kỳ vọng của các bên liên quan.
