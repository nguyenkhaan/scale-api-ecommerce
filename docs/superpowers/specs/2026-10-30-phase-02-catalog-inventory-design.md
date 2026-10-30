# Phase 02 Catalog and Inventory Design

## Intent

Phase 02 delivers the first complete merchandise slice: an administrator can define products and SKUs, establish and adjust stock with an auditable ledger, and a customer can browse the active catalog without seeing private inventory counters. This design first closes the HTTP contract, then adds the GORM models and forward-only Goose migrations. The user will run the migrations; this work must not apply them to any database.

## Goals

- Make the Phase 02 catalog and inventory OpenAPI contract implementable without guessing.
- Provide stable keyset pagination for public products and inventory-adjustment history.
- Add stable conflict codes for duplicate SKU codes and adjustments below reserved stock.
- Add models and schema for categories, products, SKUs, inventories, and inventory adjustments.
- Preserve product/SKU and stock history; no hard-delete API or cascading delete that erases business history.
- Keep public catalog responses free of `on_hand`, `reserved`, adjustment, and actor data.

## Non-goals

- Category administration UI/API (`FR-005`, P1).
- Reservation/order tables and the complete `INV-02` oracle; those arrive with checkout in Phase 04.
- Catalog/inventory controllers, services, frontend screens, seed data, or running migrations in this change.
- Full-text search or a trigram index before query-plan evidence justifies one.

## Anti-goals

- Offset/page-number pagination whose results shift when products are inserted.
- Treating Redis or frontend state as authoritative for catalog or inventory.
- Updating inventory without appending its ledger entry in the same transaction.
- Rewriting already-numbered migrations `00003` and `00004`.

## Constraints

- PostgreSQL is authoritative. Gin/GORM remain the application stack and Goose owns versioned SQL.
- VND prices and quantities are integers. SKU regular price and maximum order quantity are positive.
- `0 <= reserved <= on_hand`; `available` is derived as `on_hand - reserved` and is not stored.
- The working tree contains unrelated user changes; implementation must not overwrite them.
- Existing flash-sale migrations precede the catalog tables and currently cannot declare their intended SKU foreign key. A later forward migration may add that foreign key after `skus` exists, but must not edit an applied migration.

## Approaches considered

### A. Stable cursor plus readable adjustment history — selected

Use deterministic keyset order `(created_at DESC, id DESC)`, opaque cursors, and a paginated admin adjustment-history endpoint. This remains stable under inserts and lets the admin UI and audit checks use the same ledger.

### B. Stable cursor with internal-only ledger

This saves one read endpoint but cannot satisfy the planned inventory-history screen and forces audit work to query PostgreSQL directly.

### C. Offset/page pagination plus readable adjustment history

This is simpler for numbered-page UI but can duplicate or skip rows while the catalog changes and conflicts with the repository's existing cursor convention.

## HTTP contract

### Public catalog

`GET /v1/products` keeps `category_id`, `cursor`, and `limit`, and adds `query`:

- `query` is at most 100 characters. The server trims surrounding whitespace and performs a case-insensitive literal substring match against product name; SQL wildcard characters in user input are treated literally. An empty trimmed query is equivalent to omitting it.
- `limit` defaults to 20 and is constrained to 1–100.
- Results are ordered by immutable `(products.created_at DESC, products.id DESC)`.
- `cursor` is an opaque server-issued token representing the last returned ordering tuple and a fingerprint of the effective `query` and `category_id` filters. Clients must return it unchanged. A malformed cursor or reuse with different filters returns `422 VALIDATION_ERROR`.
- `next_cursor` is null when no later row exists.
- Public results contain only active products in visible categories (or with no category) that have at least one active SKU; only active SKUs are returned.
- `GET /v1/products/{productId}` applies the same visibility rule and otherwise returns `404 RESOURCE_NOT_FOUND`.
- Public product/SKU schemas never include inventory counters or adjustment history.

The token encoding and signing are left to the builder. It must round-trip the ordering tuple and filter fingerprint without exposing a client-editable pagination contract.

### Admin inventory list

`GET /v1/admin/inventories` retains cursor pagination with the shared default and maximum. Its deterministic ordering is `(skus.code ASC, inventories.sku_id ASC)` so stock rows are useful to operators and stable when stock changes. The opaque cursor represents that tuple.

### Adjustment history

Add `GET /v1/admin/inventories/{skuId}/adjustments` alongside the existing `POST` operation:

- Requires an authenticated admin session; it does not require CSRF because it is read-only.
- Accepts shared `cursor` and `limit` parameters.
- Returns `InventoryAdjustmentPage { items, next_cursor }` ordered by `(created_at DESC, id DESC)`.
- Each item exposes `id`, `sku_id`, `actor_user_id`, `delta`, `on_hand_before`, `on_hand_after`, `reason`, and `created_at`.
- Unknown SKU returns `404 RESOURCE_NOT_FOUND`; malformed cursor returns `422 VALIDATION_ERROR`; non-admin access returns `403 FORBIDDEN`.

The existing `POST` operation keeps `delta` and `reason`. `delta = 0`, blank reason, values outside signed 64-bit range, and malformed input return `422 VALIDATION_ERROR`. The service will lock the inventory row, calculate the result, update inventory, and append the adjustment in one transaction.

### Stable conflict codes

Extend `Problem.code` with:

- `SKU_CODE_ALREADY_EXISTS`: HTTP 409 when product/SKU creation attempts a SKU code already present, including duplicates within one create-product request.
- `INVENTORY_BELOW_RESERVED`: HTTP 409 when an adjustment would make `on_hand < reserved`; this also covers a result below zero because `reserved` is non-negative.

The database constraints remain the final safety net. Services must translate the expected unique/check conflict to these stable codes rather than exposing PostgreSQL messages.

SKU codes are trimmed at the API boundary, stored with their original letter casing, and compared case-sensitively. The database uses a normal case-sensitive unique constraint on `skus.code`; therefore `ABC-01` and `abc-01` may coexist, while another exact `ABC-01` conflicts. A code that is empty after trimming returns `422 VALIDATION_ERROR`.

### Creation results and related-resource validation

- Product creation returns `CreateProductResult { product, inventories }`, where `product` contains the created SKUs and `inventories` contains one initial inventory row for each created SKU.
- SKU creation returns `CreateSkuResult { sku, inventory }`.
- Public `Product` and `Sku` schemas remain free of inventory data; the result wrappers are admin-only contract shapes.
- A syntactically valid but nonexistent `category_id` in create/update product returns `422 VALIDATION_ERROR` with a violation for `category_id`. The product path itself still returns `404 RESOURCE_NOT_FOUND` when the target product does not exist.

## Models

Add five GORM models under `app/backend/internal/model/`:

- `Category`: UUID, name, unique slug, visibility, timestamps.
- `Product`: UUID, nullable category ID, name, description, nullable image URL, active/inactive status, timestamps.
- `SKU`: UUID, product ID, globally unique code, JSONB attributes, regular price, maximum quantity per order, active/inactive status, timestamps.
- `Inventory`: SKU ID primary/foreign key, on-hand, reserved, version, updated timestamp. `available` is a DTO calculation, not a model column.
- `InventoryAdjustment`: UUID, SKU ID, actor user ID, delta, before/after values, reason, creation timestamp.

Use explicit table names, status constants, PostgreSQL types, and GORM tags consistent with existing models. Associations may be included only where they serve actual Phase 02 queries; schema enforcement belongs in SQL migrations rather than AutoMigrate.

Creating a SKU always creates its inventory row in the same transaction. `initial_on_hand > 0` also appends an initial adjustment from zero to that value, attributed to the creating administrator and marked with the service-owned reason `INITIAL_STOCK`; `initial_on_hand = 0` creates no zero-delta ledger row. Product creation performs this atomically for every supplied SKU, so it either creates the complete product/SKU/inventory/initial-ledger set or creates nothing. This makes the inventory ledger a zero-based audit source for `INV-12`.

## Migrations

Create forward-only migrations without running them:

1. `00005_create_catalog.sql`: categories, products, and SKUs; UUID defaults; status/price/quantity checks; unique SKU code; foreign keys and catalog/keyset indexes. After `skus` exists, add the intended foreign key from `flash_sale_items.sku_id` to `skus.id`. The migration should fail visibly if a shared database already contains orphan flash-sale SKU IDs rather than silently deleting or inventing catalog data.
2. `00006_create_inventory.sql`: inventories and inventory adjustments; UUID default for adjustments; inventory and ledger checks; actor/SKU foreign keys; indexes for inventory listing and adjustment history.

Down migrations remove objects in reverse dependency order. Product, SKU, and category relationships should default to restrictive deletion. Normal application behavior only toggles status/visibility; it does not hard-delete business entities.

Because `00005` adds the deferred flash-sale-to-SKU foreign key, existing flash-sale integration fixtures that insert arbitrary SKU UUIDs must be updated to create valid category/product/SKU rows first. This compatibility edit is part of the implementation and must preserve the intent of those tests.

## Verification boundary

- Format and compile the added Go model files.
- Validate OpenAPI structure using the repository's available contract tooling if it does not install new dependencies.
- Inspect migration SQL and add migration-focused tests where they can be written without applying migrations during this task.
- Do not run `go run ./cmd/migrate up`, test reset, or any integration command that applies the new migration. Hand the exact migration filenames and manual commands to the user.

## Builder decisions

- Choose the internal opaque-cursor serialization/signing helper while preserving the documented behavior.
- Choose sensible maximum lengths for category name/slug and adjustment reason at the database layer, aligned with OpenAPI limits.
- Keep model associations minimal; do not add repository abstractions until a service use case needs them.
