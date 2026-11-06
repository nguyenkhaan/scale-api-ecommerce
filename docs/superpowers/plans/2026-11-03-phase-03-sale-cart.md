# Phase 03 Flash Sale and Cart Implementation Plan

**Goal:** Implement flash sale campaigns, public sale browsing, customer cart, and checkout preview without holding inventory.

**Spec:** `docs/plans/phases/phase-03-sale-cart.md`, `docs/technical/api/openapi.yaml`, and `docs/superpowers/specs/2026-11-03-phase-03-sale-cart-design.md`

## Task Breakdown

### Task 1: Migration and Models for Cart & Flash Sale SKU FK
- Create `00007_create_carts.sql` with `carts` and `cart_items` tables, plus FK constraint from `flash_sale_items.sku_id` to `skus(id)`.
- Define models `model.Cart` and `model.CartItem`.
- Update `internal/migration/migration.go` if needed.
- Write tests in `test/cart_model_test.go` and migration tests.

### Task 2: Flash Sale Module (Public & Admin)
- Implement `internal/module/flashsale/dto.go`.
- Implement `internal/module/flashsale/service.go`:
  - `CreateDraft`, `UpdateDraft`, `Schedule`, `CancelSchedule`, `Stop`.
  - Anti-overlapping SKU validation across active/scheduled campaigns.
  - Derived effective status (`draft`, `scheduled`, `active`, `stopped`, `ended`).
  - Public `ListPublicFlashSales` and `GetPublicFlashSale`.
  - Admin `ListAdminFlashSales` and `GetAdminFlashSale`.
- Implement `internal/module/flashsale/controller.go`.
- Register routes in `router.go`.
- Write unit tests in `test/flash_sale_controller_test.go` and `test/flash_sale_api_test.go`.

### Task 3: Customer Multi-SKU Cart Module
- Implement `internal/module/cart/dto.go`.
- Implement `internal/module/cart/service.go`:
  - `GetCart(ctx, userID)`: fetch or initialize active cart for customer.
  - `UpsertCartItem(ctx, userID, skuID, input)`: max 20 SKUs, deduplicate/merge by SKU ID, increment version, enforce price source rules.
  - `RemoveCartItem(ctx, userID, skuID)`: 204 No Content.
- Implement `internal/module/cart/controller.go`.
- Register routes in `router.go`.
- Write unit tests in `test/cart_controller_test.go` and `test/cart_api_test.go`.

### Task 4: Checkout Preview
- Implement `POST /v1/checkout/preview`.
- Validate selection items against customer's active cart.
- Check version match (409 Conflict on mismatch).
- Validate current prices:
  - Regular price from active SKU.
  - Flash sale price, quota availability, and customer purchase limit.
- Calculate line subtotals and overall total amount in VND.
- Return `CheckoutPreview` with current timestamp.
- Write unit tests in `test/checkout_preview_test.go`.

### Task 5: Frontend Implementation
- Implement API clients: `src/lib/flash-sale-api.ts`, `src/lib/cart-api.ts`.
- Implement UI components:
  - `src/pages/FlashSalesPage.tsx` (public flash sale browsing).
  - `src/pages/CartPage.tsx` (cart item list, quantity adjustment, remove, line selection, preview total).
  - `src/pages/AdminFlashSalePage.tsx` (campaign list, create modal/form, schedule, stop).
- Update navigation and routes in `App.tsx`.
- Write frontend unit tests for API clients, logic, and page components.

### Task 6: End-to-End Validation
- Run full backend test suite (`go test ./...`).
- Run full frontend test suite (`bun test`).
- Apply migration and seed test data to Neon.
- Verify through live API calls.
