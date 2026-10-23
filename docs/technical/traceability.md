# Ma trận truy vết SaleScale

**Trạng thái:** Draft
**Quy tắc:** PRD định nghĩa yêu cầu; bảng này chỉ chỉ đường tới hợp đồng kỹ thuật và bằng chứng.

## Quy ước tài liệu

| Ký hiệu | Tài liệu |
|---|---|
| PRD | `docs/product/prd.md` |
| DOMAIN | `docs/technical/domain-model.md` |
| DATA | `docs/technical/data-model.md` |
| API | `docs/technical/api/openapi.yaml` |
| TEST | `docs/technical/testing.md` |
| ARCH | `docs/technical/architecture.md` |

## Functional requirements

| Yêu cầu | Ưu tiên | Hợp đồng kỹ thuật chính | Endpoint/boundary | Bằng chứng |
|---|---|---|---|---|
| FR-001–FR-003 | P0 | DOMAIN Identity; DATA users/sessions; ADR-005 | `/v1/auth/*` | TEST-24 + auth contract/security tests |
| FR-004 | P1 | ARCH single-store deployment | Store configuration boundary chưa chi tiết | P1 acceptance |
| FR-005 | P1 | DATA categories | Admin category contract cần chi tiết trước P1 | Catalog acceptance |
| FR-006–FR-009 | P0 | DOMAIN Catalog; DATA products/skus/order snapshot | public products; admin products/skus | TEST-08, TEST-20, snapshot integration |
| FR-010–FR-013 | P0 | DATA inventories, adjustments, reservations | admin inventories/adjustments | TEST-01–04, TEST-11–15, TEST-21 |
| FR-014–FR-017 | P0 | DOMAIN FlashSale; DATA campaign/items/usage | admin flash-sales; preview/order | TEST-09, TEST-10, TEST-20 |
| FR-018 | P1 | DOMAIN FlashSale transition | admin flash-sales stop | TEST-20 + P1 statistics acceptance |
| FR-019–FR-022 | P0 | DOMAIN Cart; DATA carts/items | cart; checkout preview | TEST-08, TEST-25, HTTP contract |
| FR-023 | P1 | DATA cart item version + order snapshot | cart/order boundary | TEST-25 |
| FR-024–FR-029 | P0 | ARCH checkout; DOMAIN Idempotency; DATA order/reservation/idempotency | orders; idempotency lookup | TEST-01–10 |
| FR-030–FR-035 | P0 | DOMAIN Payment; DATA sessions/events/reconciliation | payment session, simulator, callback | TEST-11–18 |
| FR-036 | P1 | DATA reconciliation | admin payment-reconciliations | reconciliation acceptance |
| FR-037–FR-040 | P0 | DOMAIN Order state; ARCH expiry worker | orders get/list/cancel; worker boundary | TEST-13–19 |
| FR-041 | P1 | Order read model | Admin order contract cần chi tiết trước P1 | Admin acceptance |
| FR-042–FR-043 | P0 | ARCH Observability | Internal metrics/log boundary | Verification of logs/metrics + benchmark |
| FR-044–FR-045 | P1 | API toàn bộ public/admin | React + Vite | Browser acceptance |
| FR-046 | P0 | SPEC commands; TEST gates | Runbook cần hoàn thiện cùng môi trường | Rebuild/seed/test rehearsal |
| FR-047–FR-049 | P1 | ARCH outbox; DATA P1 boundary | notifications + AMQP event contract | TEST-26–29 |

Các contract P1 được đánh dấu “cần chi tiết” phải được đặc tả trước khi task P1 tương ứng được lập; không chặn PLAN cho lát cắt P0.

## Bất biến

| Bất biến | Thiết kế bảo vệ | Test bắt buộc |
|---|---|---|
| INV-01 | Inventory check constraint + row lock | TEST-03, TEST-21 |
| INV-02 | Reservation held ledger ↔ inventory.reserved | TEST-01–04, oracle SQL |
| INV-03 | Một transaction tạo order/items/reservations | TEST-01, TEST-02, TEST-04 |
| INV-04 | Rollback toàn bộ + stored business rejection | TEST-02, TEST-08, TEST-09, TEST-23 |
| INV-05 | Unique idempotency identity + hash/response | TEST-05–07 |
| INV-06 | Guarded terminal transition | TEST-11, TEST-12, TEST-14, TEST-15, TEST-19 |
| INV-07 | Verified/unique payment event | TEST-11, TEST-17, TEST-18 |
| INV-08 | Terminal order không còn HELD | TEST-12–16, TEST-19 |
| INV-09 | Flash sale quota check constraint/counter | TEST-09 |
| INV-10 | Locked customer usage counter | TEST-10 |
| INV-11 | Integer money + order snapshot + payment equality | TEST-01, TEST-08, TEST-17 |
| INV-12 | Inventory adjustment ledger và confirmed sold | TEST-03, TEST-11, oracle SQL |

## Non-functional requirements

| Yêu cầu | Thiết kế/bằng chứng |
|---|---|
| NFR-01–NFR-03 | PostgreSQL authoritative state, stateless API/worker, TEST concurrency/restart |
| NFR-04–NFR-06 | ADR-005, API security schemes, auth/ownership/callback tests |
| NFR-07–NFR-08 | `timestamptz`, database time, expiry worker và TEST-13/15/16/19 |
| NFR-09 | Pagination/payload limits, bounded pool/retry/batch; benchmark resource metrics |
| NFR-10 | ARCH Observability; metric cardinality rules |
| NFR-11 | Environment boundaries; Compose/AKS runbook ở phase tương ứng |
| NFR-12–NFR-13 | OpenAPI payload; browser acceptance; Go test gates |
| NFR-14 | ARCH outbox, DATA P1 boundary, TEST-26–29 |

## Hướng tối ưu

| PERF | Boundary không được phá | Bằng chứng |
|---|---|---|
| PERF-01 | Order/inventory lock order và transaction ngắn | TEST-01–04; lock/transaction metrics |
| PERF-02 | Index xuất phát từ query thật; không bỏ validation | Query plan, query count, integration tests |
| PERF-03 | Tổng pool/retry/timeout hữu hạn | Pool wait/connections/timeout metrics |
| PERF-04 | Stateless instances, shared idempotency/session | TEST-05, LOAD-05, 1-vs-many instance report |
| PERF-05 | Redis chỉ public read | TEST-22, LOAD-04 |
| PERF-06 | Rejection không tạo state | TEST-23, LOAD-09 |
| PERF-07 | Durable expiry scan, idempotent transition | TEST-19, expiry lag/backlog |
| PERF-08 | Chỉ tối ưu Go sau profiling | Profile gắn với commit/workload |
| PERF-09 | Outbox/RabbitMQ không quyết định order | TEST-26–29, checkout + delivery latency |

## Quy tắc kiểm tra thay đổi

Mỗi pull request triển khai sau này phải nêu:

- mã `FR/NFR/PERF` được thực hiện;
- `BR/INV` bị tác động;
- endpoint/schema migration liên quan;
- test ID hoặc test mới cung cấp bằng chứng;
- ADR nếu thay đổi một quyết định nền tảng.

Nếu không tìm được dòng truy vết phù hợp, thay đổi chưa có đặc tả và phải quay lại PRD/SPECIFY.

