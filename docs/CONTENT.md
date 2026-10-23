# Mục lục SDD — SaleScale

Đọc theo thứ tự: [PRD](product/prd.md) → [project spec](spec.md) → [kiến trúc](technical/architecture.md) → [domain](technical/domain-model.md) → [dữ liệu](technical/data-model.md) → [OpenAPI](technical/api/openapi.yaml) → [kiểm thử](technical/testing.md) → [truy vết](technical/traceability.md) → [kế hoạch](plans/plan.md).

| Thư mục | Nội dung | Vai trò |
|---|---|---|
| `product/` | PRD, user stories, functional requirements, [màn hình](product/screen_list.md), [chức năng BE](product/feature.md) | PRD có thẩm quyền; các file còn lại là góc nhìn dẫn xuất |
| `technical/` | Architecture, domain/data model, API, testing, traceability | Hợp đồng kỹ thuật và bằng chứng |
| `plans/` | Kế hoạch tổng và từng phase | Thứ tự thực hiện, checklist và cổng duyệt |
| `decisions/` | ADR-003 đến ADR-006 | Lý do quyết định; giữ lịch sử khi thay đổi |
| `research/` | Project Charter, nghiệp vụ gốc | Đầu vào nghiên cứu, xem đối chiếu tại PRD mục 17.4 |

Quy trình: `SPECIFY → PLAN → TASKS → IMPLEMENT`. Đặc tả hiện là draft, kế hoạch cần hai thành viên duyệt trước khi chuyển sang task/triển khai. Khi đổi nghiệp vụ, cập nhật PRD trước, rồi traceability, domain/data/API/test. Xem [ADR-003](decisions/003-document-authority.md).
