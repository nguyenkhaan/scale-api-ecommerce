# Catalog API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the Phase 02 public and admin catalog API without exposing inventory counters or hard-deleting catalog history.

**Architecture:** Keep the existing Gin controller → service → GORM structure used by auth. Controllers own strict HTTP decoding and stable problem responses; services own visibility rules, keyset queries, and product/SKU/inventory creation transactions. Public responses reuse catalog DTOs that never contain inventory fields.

**Tech Stack:** Go 1.27.1, Gin, GORM, PostgreSQL, opaque base64url JSON cursors.

**Spec:** `docs/plans/phases/phase-02-catalog-inventory.md` and `docs/technical/api/openapi.yaml`

## Global Constraints

- Public endpoints return only active products in visible categories (or with no category) and only active SKUs.
- Public DTOs never expose `on_hand`, `reserved`, `available`, adjustment rows, or actor IDs.
- SKU codes are trimmed and unique case-sensitively.
- Products and SKUs are hidden with active/inactive status; no DELETE route is added.
- Admin routes run `RequireSession` before `RequireRole(admin)`; cookie-authenticated mutations receive CSRF enforcement from `RequireSession`.
- Creating a product/SKU creates its inventory and positive initial-stock ledger row in the same PostgreSQL transaction.
- Do not modify inventory adjustment behavior or flash-sale code.
- Existing user changes are preserved; stage/commit only catalog-specific hunks.

## Review Focus

- Invalid UUID/filter/cursor input stops before service/DB access and returns one 422 response.
- Public list/detail cannot leak inactive products/SKUs, hidden categories, or inventory data.
- Cursor ordering remains stable under inserts and rejects reuse with different filters.
- Duplicate SKU or any failed child insert rolls back the complete product/SKU/inventory creation.
- Customer/anonymous callers cannot reach admin operations; mutations require valid CSRF.

---

### Task 1: Repair and test the existing public detail endpoint

**Files:**
- Modify: `app/backend/internal/module/catalog/controller.go`
- Modify: `app/backend/internal/module/catalog/dto.go`
- Modify: `app/backend/internal/module/catalog/service.go`
- Modify: `app/backend/internal/httpserver/problem/problem.go`
- Create: `app/backend/test/catalog_controller_test.go`
- Create: `app/backend/test/catalog_api_test.go`

**Interfaces:**
- Produces: `GetPublicProduct(ctx, productID) (ProductResponse, error)` and stable `RESOURCE_NOT_FOUND` response.

**Acceptance Criteria:**
- Invalid UUID returns 422 and stops without calling a service.
- Active detail contains only active SKUs; inactive/hidden/unbuyable products return 404.
- Response JSON contains no inventory counter names.

- [ ] Write a failing no-DB controller test for invalid UUID producing one 422 response.
- [ ] Write PostgreSQL integration cases for visibility and response privacy; leave them skipped when `TEST_DATABASE_URL` is absent.
- [ ] Add the missing early return, align the not-found code with OpenAPI, format the existing implementation, and remove tutorial comments that no longer explain non-obvious policy.
- [ ] Run the controller test RED→GREEN and compile the integration test.

### Task 2: Add public search/filter/keyset pagination

**Files:**
- Modify: `app/backend/internal/module/catalog/controller.go`
- Modify: `app/backend/internal/module/catalog/dto.go`
- Modify: `app/backend/internal/module/catalog/service.go`
- Create: `app/backend/internal/module/catalog/cursor.go`
- Modify: `app/backend/test/catalog_controller_test.go`
- Modify: `app/backend/test/catalog_api_test.go`

**Interfaces:**
- Produces: `ListPublicProducts(ctx, ListPublicProductsQuery) (ProductPageResponse, error)`; cursor encodes `(created_at, id, filter fingerprint)`.

**Acceptance Criteria:**
- `query` is trimmed, max 100 characters, case-insensitive literal substring search.
- `category_id` is UUID when present; limit defaults to 20 and stays within 1–100.
- Query uses `limit+1`, deterministic `(created_at DESC, id DESC)`, and `EXISTS` active SKU filtering.
- Products are fetched once and SKUs are fetched once for the page; no N+1 query.

- [ ] Write failing unit tests for query parsing, wildcard escaping, cursor round-trip/malformed/filter-mismatch behavior.
- [ ] Write skipped PostgreSQL cases for active visibility, literal `%/_` search, category filter, and two-page traversal.
- [ ] Implement DTOs, cursor helper, service keyset query, batched SKU hydration, and `GET /v1/products`.
- [ ] Run unit tests RED→GREEN and compile integration tests.

### Task 3: Add admin catalog read endpoints and authorization

**Files:**
- Modify: `app/backend/internal/module/catalog/controller.go`
- Modify: `app/backend/internal/module/catalog/dto.go`
- Modify: `app/backend/internal/module/catalog/service.go`
- Modify: `app/backend/internal/httpserver/router.go`
- Modify: `app/backend/test/catalog_controller_test.go`
- Modify: `app/backend/test/catalog_api_test.go`

**Interfaces:**
- Produces: `ListAdminProducts(ctx, AdminProductListQuery)` and `GetAdminProduct(ctx, productID)`; admin results include inactive rows but no inventory counters.

**Acceptance Criteria:**
- Anonymous/customer access is rejected before controller execution.
- Admin list optionally filters active/inactive and uses stable keyset pagination.
- Admin detail returns inactive product and inactive SKU data.

- [ ] Write failing route/controller tests for 401/403, invalid status, and invalid cursor.
- [ ] Write skipped PostgreSQL cases for inactive admin visibility.
- [ ] Register `/v1/admin/products` GET and `/v1/admin/products/:productId` GET behind session/admin middleware.
- [ ] Implement service queries and run focused tests.

### Task 4: Add transactional admin product and SKU creation

**Files:**
- Modify: `app/backend/internal/module/catalog/controller.go`
- Modify: `app/backend/internal/module/catalog/dto.go`
- Modify: `app/backend/internal/module/catalog/service.go`
- Modify: `app/backend/internal/httpserver/problem/problem.go`
- Modify: `app/backend/test/catalog_controller_test.go`
- Modify: `app/backend/test/catalog_api_test.go`

**Interfaces:**
- Produces: `CreateProduct(ctx, actorUserID, CreateProductRequest)` and `CreateSKU(ctx, actorUserID, productID, CreateSKUInput)`.

**Acceptance Criteria:**
- JSON bodies are size-limited, object-only, strict about unknown fields, and validated before DB access.
- Product/SKUs/inventories/positive initial ledgers commit all-or-nothing.
- Duplicate exact SKU returns 409 `SKU_CODE_ALREADY_EXISTS`; differently cased codes remain valid.
- Missing category is a field-level 422; missing parent product is 404.

- [ ] Write failing no-DB validation tests for names, SKU codes, JSON object attributes, integer price/quantity/initial stock, duplicate codes in one request, and unknown fields.
- [ ] Write skipped PostgreSQL atomicity/duplicate/initial-ledger cases.
- [ ] Implement transaction methods and POST routes behind session/admin middleware; use identity UserID as ledger actor.
- [ ] Run validation tests RED→GREEN and compile integration tests.

### Task 5: Add admin product/SKU update and hiding

**Files:**
- Modify: `app/backend/internal/module/catalog/controller.go`
- Modify: `app/backend/internal/module/catalog/dto.go`
- Modify: `app/backend/internal/module/catalog/service.go`
- Modify: `app/backend/test/catalog_controller_test.go`
- Modify: `app/backend/test/catalog_api_test.go`

**Interfaces:**
- Produces: `UpdateProduct(ctx, productID, UpdateProductRequest)` and `UpdateSKU(ctx, skuID, UpdateSKURequest)`.

**Acceptance Criteria:**
- Patch distinguishes omitted fields from explicit null for nullable category/image fields.
- Empty patch and invalid values return 422.
- `active:false` changes status to inactive; no hard delete is performed.
- Updating current name/price does not mutate historical order data.

- [ ] Write failing patch decoder/validation tests for absent/null/value and empty payload.
- [ ] Write skipped PostgreSQL cases for inactive visibility and persistence.
- [ ] Implement PATCH routes and update maps containing only present fields.
- [ ] Run focused tests RED→GREEN and compile integration tests.

### Task 6: Verification and contract handoff

**Files:**
- Modify only if required: `app/backend/cmd/api/main.go`
- Modify only if required: `docs/technical/api/openapi.yaml`

**Acceptance Criteria:**
- Router/main inject the catalog service exactly once and middleware order is session → admin → controller.
- Go formatting, build, vet, and no-DB tests pass; PostgreSQL cases are explicitly reported as skipped until a dedicated URL is supplied.
- OpenAPI and runtime stable codes/routes agree.

- [ ] Run focused catalog tests, `go build ./...`, `go vet ./...`, and `git diff --check`.
- [ ] Run `env -u TEST_DATABASE_URL go test ./...` and report every unrelated failure without editing inventory/flash-sale.
- [ ] Provide the exact dedicated-PostgreSQL command for the user to run.

### Task 7: Repair and test inventory adjustment routing

**Files:**
- Modify: `app/backend/internal/module/inventory/controller.go`
- Modify: `app/backend/internal/module/inventory/dto.go`
- Modify: `app/backend/internal/module/inventory/service.go`
- Create: `app/backend/test/inventory_controller_test.go`

**Acceptance Criteria:**
- Adjustment route is exactly `POST /api/v1/admin/inventories/:skuId/adjustments`, never under public products.
- Middleware order remains session → admin → controller, so POST receives CSRF enforcement.
- Invalid UUID/body stops before the service; stable 404/409/422 codes match OpenAPI.
- Existing row-lock/update/ledger transaction remains atomic.

- [ ] Write failing no-DB route/controller tests for the path, invalid UUID, and stable error mapping.
- [ ] Correct the route and make the controller service testable without PostgreSQL.
- [ ] Format the existing implementation and run controller tests RED→GREEN.

### Task 8: Add admin inventory list pagination

**Files:**
- Modify: `app/backend/internal/module/inventory/controller.go`
- Modify: `app/backend/internal/module/inventory/dto.go`
- Modify: `app/backend/internal/module/inventory/service.go`
- Create: `app/backend/internal/module/inventory/cursor.go`
- Modify: `app/backend/test/inventory_controller_test.go`
- Create: `app/backend/test/inventory_api_test.go`

**Acceptance Criteria:**
- Admin-only `GET /api/v1/admin/inventories` returns on-hand, reserved, derived available, and version.
- Limit defaults to 20 and stays within 1–100.
- Cursor keyset is stable on `(skus.code ASC, inventories.sku_id ASC)` and malformed cursors return 422.
- PostgreSQL integration tests cover pagination and derived available when a dedicated DB is supplied.

- [ ] Write failing cursor/query parsing tests and skipped PostgreSQL page traversal cases.
- [ ] Implement DTO, cursor, service join query, controller, and route.
- [ ] Run no-DB tests RED→GREEN and compile integration tests.

### Task 9: Add inventory adjustment history

**Files:**
- Modify: `app/backend/internal/module/inventory/controller.go`
- Modify: `app/backend/internal/module/inventory/dto.go`
- Modify: `app/backend/internal/module/inventory/service.go`
- Modify: `app/backend/test/inventory_controller_test.go`
- Modify: `app/backend/test/inventory_api_test.go`
- Modify: `docs/technical/api/openapi.yaml`

**Acceptance Criteria:**
- Admin-only `GET /api/v1/admin/inventories/:skuId/adjustments` returns actor, delta, before/after, reason, and timestamp in descending `(created_at, id)` order.
- Unknown SKU returns 404; malformed UUID/cursor returns 422.
- History is cursor-paginated and read-only, so no CSRF token is required for GET.
- OpenAPI declares the GET operation and adjustment page schemas.

- [ ] Write failing no-DB route/cursor tests and skipped PostgreSQL history cases.
- [ ] Implement DTO, service, controller, route, and contract updates.
- [ ] Run focused tests RED→GREEN and include inventory in final verification.
