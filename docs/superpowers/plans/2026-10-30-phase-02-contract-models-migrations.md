# Phase 02 Contract, Models, and Migrations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Finalize the Phase 02 catalog/inventory HTTP contract and add compile-safe GORM models plus forward-only Goose migrations without applying those migrations to a database.

**Architecture:** OpenAPI defines the public/admin boundary first. Catalog and inventory models mirror PostgreSQL but do not use AutoMigrate; two ordered SQL migrations create catalog before inventory and append the deferred flash-sale SKU foreign key. Unit and static migration tests verify mappings and SQL while `TEST_DATABASE_URL` is explicitly unset so Codex never applies migrations.

**Tech Stack:** OpenAPI 3.1 YAML, Go 1.27.1, GORM 1.31, PostgreSQL, Goose 3.29, `github.com/goccy/go-yaml` for the contract test.

**Spec:** `docs/superpowers/specs/2026-10-30-phase-02-catalog-inventory-design.md`

## Global Constraints

- SKU codes are trimmed at the API boundary and unique case-sensitively; `ABC-01` and `abc-01` may coexist.
- Public product/SKU responses must not expose inventory counters, adjustment rows, or actor IDs.
- Pagination is opaque keyset pagination; public products use `(created_at DESC, id DESC)` and adjustment history uses `(created_at DESC, id DESC)`.
- `0 <= reserved <= on_hand`; `available` is derived and is never stored.
- Do not modify migrations `00001`–`00004`, run `go run ./cmd/migrate up`, reset a database, or run tests with `TEST_DATABASE_URL` set.
- Preserve unrelated working-tree changes and stage only the exact files named by each task.
- Category administration, reservation/order tables, controllers, services, frontend screens, and seed commands are outside this plan.

## Review Focus

- A cursor reused with different `query`/`category_id` filters must produce `422 VALIDATION_ERROR`; Task 1 pins the documented filter fingerprint.
- Exact duplicate SKU codes must produce `409 SKU_CODE_ALREADY_EXISTS`, while differently cased codes remain valid; Tasks 1, 2, and 4 pin the contract/model/constraint.
- An inventory adjustment below `reserved`, including below zero, must produce `409 INVENTORY_BELOW_RESERVED`; Tasks 1 and 5 pin the contract and final constraint.
- Initial inventory must be auditable from zero without permitting a zero-delta adjustment; Tasks 1, 3, and 5 pin the result schema and ledger constraints.
- Adding the deferred flash-sale SKU foreign key must not leave test fixtures inserting orphan SKU IDs; Task 6 updates every affected fixture.

---

### Task 1: Finalize and test the Phase 02 OpenAPI contract

**Files:**
- Modify: `docs/technical/api/openapi.yaml`
- Create: `app/backend/test/openapi_phase02_contract_test.go`
- Include in commit: `docs/superpowers/specs/2026-10-30-phase-02-catalog-inventory-design.md`
- Include in commit: `docs/superpowers/plans/2026-10-30-phase-02-contract-models-migrations.md`

**Interfaces:**
- Consumes: the approved design document.
- Produces: OpenAPI operations `listProducts`, `createProduct`, `createSku`, `listInventories`, `listInventoryAdjustments`, and `adjustInventory`; schemas `CreateProductResult`, `CreateSkuResult`, `InventoryAdjustment`, and `InventoryAdjustmentPage`; problem codes `SKU_CODE_ALREADY_EXISTS` and `INVENTORY_BELOW_RESERVED`.

**Acceptance Criteria:**
- Public product search documents literal case-insensitive name matching, trimmed empty-query behavior, stable ordering, filter-bound opaque cursors, and active-only visibility.
- Admin adjustment history is cursor-paginated and returns the complete ledger shape without CSRF on GET.
- Create-product and create-SKU responses return inventory through admin-only wrapper schemas.
- Duplicate SKU and below-reserved conflicts have stable HTTP 409 codes; nonexistent `category_id` is a field-level 422.

- [ ] **Step 1: Write the failing contract test**

Create `TestPhase02OpenAPIContract` using `github.com/goccy/go-yaml` to parse `../../../docs/technical/api/openapi.yaml` into `map[string]any`. Assert:

- `/v1/products` GET has `query`, `category_id`, `cursor`, and `limit` parameters and a description containing `created_at DESC`, filter fingerprint, and active visibility.
- `/v1/admin/inventories/{skuId}/adjustments` has GET and POST, with GET returning `InventoryAdjustmentPage` and declaring 403/404/422.
- create-product 201 references `CreateProductResult`; create-SKU 201 references `CreateSkuResult`.
- the four new schemas exist and `Problem.code.enum` contains both new conflict codes.

- [ ] **Step 2: Run the focused test and confirm it fails without applying migrations**

Run: `cd app/backend && env -u TEST_DATABASE_URL go test ./test -run TestPhase02OpenAPIContract -count=1`

Expected: FAIL because the adjustment GET, wrapper schemas, and conflict codes do not exist yet.

- [ ] **Step 3: Update public catalog parameters and pagination semantics**

In `docs/technical/api/openapi.yaml`, keep the shared `Cursor`/`Limit`, describe opaque keyset behavior, and define `query` as trimmed, maximum 100 characters, case-insensitive literal substring matching. Document `(created_at DESC, id DESC)`, cursor/filter mismatch as 422, and active product/visible category/active SKU visibility.

- [ ] **Step 4: Add admin result and adjustment-history schemas**

Define:

- `CreateProductResult` with required `product` and `inventories`.
- `CreateSkuResult` with required `sku` and `inventory`.
- `InventoryAdjustment` with required `id`, `sku_id`, `actor_user_id`, `delta`, `on_hand_before`, `on_hand_after`, `reason`, `created_at`.
- `InventoryAdjustmentPage` with required `items` and nullable `next_cursor`.

Switch create endpoint 201 responses to the wrapper schemas and add the adjustment-history GET operation with cursor/limit and 403/404/422 responses.

- [ ] **Step 5: Add stable validation and conflict semantics**

Document case-sensitive trimmed SKU codes, `SKU_CODE_ALREADY_EXISTS`, `INVENTORY_BELOW_RESERVED`, nonexistent-category 422, zero/blank adjustment validation, and `(skus.code ASC, inventories.sku_id ASC)` inventory-list ordering. Add both codes to `Problem.code.enum`.

- [ ] **Step 6: Run the contract test**

Run: `cd app/backend && env -u TEST_DATABASE_URL go test ./test -run TestPhase02OpenAPIContract -count=1`

Expected: PASS; PostgreSQL integration tests report skipped because `TEST_DATABASE_URL` is absent.

- [ ] **Step 7: Commit the contract and approved planning documents**

```bash
git add docs/technical/api/openapi.yaml app/backend/test/openapi_phase02_contract_test.go docs/superpowers/specs/2026-10-30-phase-02-catalog-inventory-design.md docs/superpowers/plans/2026-10-30-phase-02-contract-models-migrations.md
git commit -m "docs: finalize phase 2 catalog inventory contract"
```

### Task 2: Add catalog GORM models

**Files:**
- Create: `app/backend/internal/model/category.go`
- Create: `app/backend/internal/model/product.go`
- Create: `app/backend/internal/model/sku.go`
- Create: `app/backend/internal/model/catalog_test.go`

**Interfaces:**
- Consumes: OpenAPI names and validation bounds from Task 1.
- Produces: `model.Category`, `model.Product`, `model.SKU`; constants `ProductStatusActive`, `ProductStatusInactive`, `SKUStatusActive`, `SKUStatusInactive`; exact table mappings `categories`, `products`, `skus`.

**Acceptance Criteria:**
- Models expose only persisted fields and use UUID/string IDs, integer VND price, JSONB attributes, nullable category/image fields, and explicit table names.
- SKU uniqueness remains case-sensitive; no lowercase hook or canonicalized shadow column exists.

- [ ] **Step 1: Write failing GORM schema tests**

In `catalog_test.go`, add `TestCatalogModelsMapToExpectedTablesAndColumns`. Use `gorm.io/gorm/schema.Parse` as in `flash_sale_test.go` to assert table names, UUID defaults, primary keys, nullability of `Product.CategoryID`/`ImageURL`, SKU JSONB mapping, and all expected DB columns.

- [ ] **Step 2: Run the focused tests and confirm compilation fails**

Run: `cd app/backend && go test ./internal/model -run TestCatalogModelsMapToExpectedTablesAndColumns -count=1`

Expected: FAIL because `Category`, `Product`, and `SKU` are undefined.

- [ ] **Step 3: Implement `Category`**

Add fields `ID string`, `Name string`, `Slug string`, `IsVisible bool`, `CreatedAt time.Time`, `UpdatedAt time.Time` with explicit PostgreSQL/GORM tags and `TableName() string` returning `categories`.

- [ ] **Step 4: Implement `Product` and status constants**

Add fields `ID string`, `CategoryID *string`, `Name string`, `Description string`, `ImageURL *string`, `Status string`, `CreatedAt time.Time`, `UpdatedAt time.Time`; define active/inactive constants and `TableName() string` returning `products`.

- [ ] **Step 5: Implement `SKU` and status constants**

Use `json.RawMessage` for `Attributes` with `type:jsonb`; add `ID`, `ProductID`, `Code`, `RegularPrice int64`, `MaxQuantityPerOrder int`, `Status`, timestamps, UUID default, and `TableName() string` returning `skus`. Do not add code-normalization hooks.

- [ ] **Step 6: Format and run model tests**

Run: `cd app/backend && gofmt -w internal/model/category.go internal/model/product.go internal/model/sku.go internal/model/catalog_test.go && go test ./internal/model -count=1`

Expected: PASS.

- [ ] **Step 7: Commit catalog models**

```bash
git add app/backend/internal/model/category.go app/backend/internal/model/product.go app/backend/internal/model/sku.go app/backend/internal/model/catalog_test.go
git commit -m "feat: add catalog models"
```

### Task 3: Add inventory GORM models

**Files:**
- Create: `app/backend/internal/model/inventory.go`
- Create: `app/backend/internal/model/inventory_adjustment.go`
- Create: `app/backend/internal/model/inventory_test.go`

**Interfaces:**
- Consumes: `model.SKU` and existing `model.User` identifiers.
- Produces: `model.Inventory`, `model.InventoryAdjustment`, constant `InventoryAdjustmentReasonInitialStock = "INITIAL_STOCK"`; exact table mappings `inventories` and `inventory_adjustments`.

**Acceptance Criteria:**
- Inventory persists on-hand, reserved, version, and update time but never persists available.
- Adjustment persists actor, signed delta, before/after snapshots, reason, and creation time.

- [ ] **Step 1: Write failing GORM schema tests**

Add `TestInventoryModelsMapToExpectedTablesAndColumns`, asserting the inventory primary key is only `sku_id`, adjustment IDs default to `gen_random_uuid()`, every design field maps to its expected column, and no field maps to `available`.

- [ ] **Step 2: Run the focused tests and confirm compilation fails**

Run: `cd app/backend && go test ./internal/model -run TestInventoryModelsMapToExpectedTablesAndColumns -count=1`

Expected: FAIL because the inventory types are undefined.

- [ ] **Step 3: Implement `Inventory`**

Add `SKUID string`, `OnHand int64`, `Reserved int64`, `Version int64`, and `UpdatedAt time.Time` with explicit types/defaults and `TableName() string` returning `inventories`.

- [ ] **Step 4: Implement `InventoryAdjustment`**

Add `ID string`, `SKUID string`, `ActorUserID string`, `Delta int64`, `OnHandBefore int64`, `OnHandAfter int64`, `Reason string`, and `CreatedAt time.Time`; define the initial-stock reason constant and `TableName() string` returning `inventory_adjustments`.

- [ ] **Step 5: Format and run model tests**

Run: `cd app/backend && gofmt -w internal/model/inventory.go internal/model/inventory_adjustment.go internal/model/inventory_test.go && go test ./internal/model -count=1`

Expected: PASS.

- [ ] **Step 6: Commit inventory models**

```bash
git add app/backend/internal/model/inventory.go app/backend/internal/model/inventory_adjustment.go app/backend/internal/model/inventory_test.go
git commit -m "feat: add inventory ledger models"
```

### Task 4: Create the catalog migration and deferred flash-sale foreign key

**Files:**
- Create: `app/backend/internal/migration/sql/00005_create_catalog.sql`
- Create: `app/backend/internal/migration/catalog_inventory_migration_test.go`

**Interfaces:**
- Consumes: catalog models from Task 2 and existing tables `flash_sale_items`/`users`.
- Produces: tables `categories`, `products`, `skus`; constraint `flash_sale_items_sku_fk`; indexes supporting public catalog keysets and category/status filtering.

**Acceptance Criteria:**
- SQL creates catalog tables in dependency order and drops them in reverse order.
- SKU code uses a normal PostgreSQL case-sensitive UNIQUE constraint, is trimmed/nonempty, and has no `lower(code)` index.
- Existing orphan flash-sale SKU values cause the foreign-key addition to fail visibly.

- [ ] **Step 1: Write the failing embedded-SQL test for `00005`**

Use package `migration` so the test can call `migrationFiles.ReadFile`. Add `TestCatalogMigrationDefinesRequiredConstraints` asserting the file contains the three tables, UUID defaults, status/price/quantity checks, case-sensitive `UNIQUE (code)`, catalog ordering indexes, restrictive foreign keys, and `ALTER TABLE flash_sale_items ... flash_sale_items_sku_fk`. Assert it does not contain `lower(code)` or destructive data cleanup.

- [ ] **Step 2: Run the focused test and confirm it fails**

Run: `cd app/backend && go test ./internal/migration -run TestCatalogMigrationDefinesRequiredConstraints -count=1`

Expected: FAIL because `sql/00005_create_catalog.sql` does not exist.

- [ ] **Step 3: Implement `00005_create_catalog.sql`**

Create:

- `categories`: UUID default, name 1–200, slug 1–100 unique, visible default true, timestamps.
- `products`: nullable category FK with `ON DELETE RESTRICT`, name 1–200, non-null description default empty, nullable image URL, active/inactive status, timestamps.
- `skus`: product FK with `ON DELETE RESTRICT`, trimmed code 1–100 with case-sensitive unique constraint, JSON object attributes default `{}`, positive regular price/max quantity, active/inactive status, timestamps.
- Keyset/filter indexes using status/category and `(created_at DESC, id DESC)`.
- The named restrictive FK from `flash_sale_items.sku_id` to `skus.id`.

The Goose Down section drops the deferred FK first, then SKUs, products, and categories.

- [ ] **Step 4: Run the static migration test**

Run: `cd app/backend && go test ./internal/migration -run TestCatalogMigrationDefinesRequiredConstraints -count=1`

Expected: PASS without opening PostgreSQL.

- [ ] **Step 5: Commit the catalog migration**

```bash
git add app/backend/internal/migration/sql/00005_create_catalog.sql app/backend/internal/migration/catalog_inventory_migration_test.go
git commit -m "feat: add catalog migration"
```

### Task 5: Create the inventory and adjustment-ledger migration

**Files:**
- Create: `app/backend/internal/migration/sql/00006_create_inventory.sql`
- Modify: `app/backend/internal/migration/catalog_inventory_migration_test.go`

**Interfaces:**
- Consumes: `skus` from Task 4 and `users` from Phase 01.
- Produces: tables `inventories`, `inventory_adjustments`; adjustment-history index `(sku_id, created_at DESC, id DESC)`.

**Acceptance Criteria:**
- Database checks enforce nonnegative counters, `reserved <= on_hand`, nonnegative version, nonzero deltas, exact before-plus-delta equality, valid before/after values, and nonblank reason up to 500 characters.
- Foreign keys use restrictive deletion so audit history cannot disappear through cascades.

- [ ] **Step 1: Extend the embedded-SQL test and confirm it fails**

Add `TestInventoryMigrationDefinesRequiredConstraints`, reading `sql/00006_create_inventory.sql` and asserting both tables, all counter/ledger checks, restrictive SKU/actor foreign keys, UUID default, and adjustment-history index.

Run: `cd app/backend && go test ./internal/migration -run TestInventoryMigrationDefinesRequiredConstraints -count=1`

Expected: FAIL because `00006_create_inventory.sql` does not exist.

- [ ] **Step 2: Implement `00006_create_inventory.sql`**

Create `inventories` with SKU primary/FK, `on_hand`, `reserved`, version default zero, update time, and `CHECK (on_hand >= 0)` plus `CHECK (reserved >= 0 AND reserved <= on_hand)` and version check. Create `inventory_adjustments` with SKU/actor restrictive FKs, nonzero delta, before/after checks including `on_hand_after = on_hand_before + delta`, trimmed reason length 1–500, timestamp, and the history index. Drop adjustments before inventories in Down.

- [ ] **Step 3: Run all static migration tests**

Run: `cd app/backend && go test ./internal/migration -count=1`

Expected: PASS without opening PostgreSQL.

- [ ] **Step 4: Commit the inventory migration**

```bash
git add app/backend/internal/migration/sql/00006_create_inventory.sql app/backend/internal/migration/catalog_inventory_migration_test.go
git commit -m "feat: add inventory ledger migration"
```

### Task 6: Make existing PostgreSQL fixtures compatible and verify without migrating

**Files:**
- Modify: `app/backend/test/auth_test_support_test.go`
- Modify: `app/backend/test/flash_sale_migration_test.go`

**Interfaces:**
- Consumes: catalog/inventory migrations from Tasks 4–5.
- Produces: `insertTestSKU(t *testing.T, database *database.Database, skuID string)` test helper and FK-valid flash-sale migration fixtures.

**Acceptance Criteria:**
- Every flash-sale item fixture references a real product/SKU.
- Test cleanup truncates new tables in a single dependency-safe statement.
- Codex verification compiles and runs non-database tests only; migration execution is left to the user.

- [ ] **Step 1: Expand integration-test cleanup for the new schema**

Update the single `TRUNCATE` in `openIntegrationDatabase` to include `inventory_adjustments`, `inventories`, `skus`, `products`, and `categories` together with existing tables in dependency order.

- [ ] **Step 2: Add the FK-valid SKU fixture helper**

Implement `insertTestSKU(t *testing.T, testDatabase *database.Database, skuID string)` in `auth_test_support_test.go`. It inserts a product with a generated UUID and an active SKU using the exact caller-provided ID, a unique code derived from that ID, `{}` attributes, regular price `100_000`, and maximum quantity `10`. It does not create inventory because the flash-sale FK only requires the SKU.

- [ ] **Step 3: Update every flash-sale fixture**

Before each valid or constraint-targeted `flash_sale_items` insert, call `insertTestSKU` for that SKU ID. For table-driven invalid price/quota/limit cases, seed each distinct SKU first so the intended check—not the FK—is what rejects the row.

- [ ] **Step 4: Format and compile the integration test package with database execution disabled**

Run: `cd app/backend && gofmt -w test/auth_test_support_test.go test/flash_sale_migration_test.go && env -u TEST_DATABASE_URL go test ./test -count=1`

Expected: PASS with PostgreSQL integration cases skipped; no migrations are applied.

- [ ] **Step 5: Run complete non-migrating verification**

Run:

```bash
cd app/backend
env -u TEST_DATABASE_URL go test ./...
go build ./...
go vet ./...
git diff --check -- ../../docs/technical/api/openapi.yaml ../../docs/superpowers/specs/2026-10-30-phase-02-catalog-inventory-design.md ../../docs/superpowers/plans/2026-10-30-phase-02-contract-models-migrations.md internal/model internal/migration test/auth_test_support_test.go test/flash_sale_migration_test.go
```

Expected: all commands exit 0. `go test` runs model, migration-SQL, contract, auth unit tests and skips integration cases requiring `TEST_DATABASE_URL`.

- [ ] **Step 6: Commit fixture compatibility**

```bash
git add app/backend/test/auth_test_support_test.go app/backend/test/flash_sale_migration_test.go
git commit -m "test: keep flash sale fixtures valid with catalog foreign key"
```

### User-run PostgreSQL verification after handoff

These commands are intentionally not executed by Codex:

```bash
cd app/backend
go run ./cmd/migrate status
go run ./cmd/migrate up
TEST_DATABASE_URL='<dedicated-test-database>' go test ./test -count=1
```

Expected database evidence: migrations `00005` and `00006` apply once; exact duplicate SKU is rejected while differently cased SKU codes coexist; inventory constraints reject `reserved > on_hand`; flash-sale migration tests pass with FK-valid fixtures.
