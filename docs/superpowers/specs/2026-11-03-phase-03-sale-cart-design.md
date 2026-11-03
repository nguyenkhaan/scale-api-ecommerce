# Phase 03 Flash Sale and Cart Design

## Intent

Phase 03 implements flash sale campaign management, the public sale catalog, customer multi-SKU shopping carts, and pre-checkout pricing preview without holding or reserving inventory. This delivers requirements `FR-014` through `FR-020`, enforces invariants `INV-09` and `INV-10`, and establishes test coverage for `TEST-09`, `TEST-10`, and `TEST-20`.

## Goals

- Establish Goose migration `00007_create_carts.sql` for `carts` and `cart_items`, plus the FK from `flash_sale_items.sku_id` to `skus(id)`.
- Implement admin flash sale management: create draft, update draft, schedule with anti-overlapping SKU protection, cancel schedule, and stop.
- Derive effective campaign status (`draft`, `scheduled`, `active`, `stopped`, `ended`) from server/database time without requiring background cron mutations.
- Implement public flash sale listing and detail endpoints according to OpenAPI contract.
- Implement customer cart endpoints: retrieve active cart, upsert SKU item with quantity & price source, delete SKU item, enforce 20 distinct SKU limit, and handle SKU deduplication/merging.
- Maintain optimistic concurrency via `cart_items.version` incremented on every modification.
- Implement `POST /v1/checkout/preview`: calculate line items, subtotal, and total without creating inventory reservations. Detect price discrepancies, expired sales, and exhausted quotas.
- Build clean, accessible React/TypeScript frontend interfaces for public flash sales, customer cart with selection preview, and admin flash sale management.

## Non-goals

- Order creation and inventory reservation (`FR-024` to `FR-027`, Phase 04).
- Payment processing, webhook callbacks, or expiration worker (Phase 05).
- Real-time stock notifications or outbox events (Phase 06).

## Invariants & Business Rules

1. **INV-09:** `quota_reserved >= 0`, `quota_sold >= 0`, and `quota_reserved + quota_sold <= quota_total`.
2. **INV-10:** A user's held quota + purchased quota per flash sale item cannot exceed `purchase_limit`.
3. **BR-15:** No overlapping schedules for the same SKU across active or scheduled campaigns.
4. **BR-16:** A scheduled campaign is locked; it cannot be modified unless cancelled back to `draft` before its start time.
5. **BR-19:** Exactly one active cart per customer. Cart items cannot exceed 20 distinct SKUs.
6. **BR-20:** Adding to cart or previewing checkout never changes inventory or quota counters.

## Architecture

- **Backend:** Gin Controller -> Service -> GORM -> PostgreSQL.
  - Module `flashsale`: campaign and item management, anti-overlap validation, public & admin queries.
  - Module `cart`: customer cart state, upsert with merge, version tracking.
  - Module `checkout` (or cart preview): pre-checkout calculation and validation against authoritative DB state.
- **Frontend:**
  - `src/lib/flash-sale-api.ts`: API client for public & admin flash sales.
  - `src/lib/cart-api.ts`: API client for customer cart & checkout preview.
  - `src/pages/FlashSalesPage.tsx`: Public flash sale browsing.
  - `src/pages/CartPage.tsx`: Customer cart management & preview.
  - `src/pages/AdminFlashSalePage.tsx`: Admin campaign creation, scheduling, and stopping.
