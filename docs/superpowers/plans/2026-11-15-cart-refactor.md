# Cart Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate cart from `src/temp/` to an OpenAPI/Axios/React Query feature without changing cart or checkout-preview behavior.

**Architecture:** `cart.type.ts` aliases generated OpenAPI schemas. `cart.service.ts` owns authenticated HTTP, CSRF, and typed errors; `cart.logic.ts` preserves pure policy; `cart.hook.ts` coordinates React Query cache and view state. Components and page consume the feature boundary only.

**Tech Stack:** React 19, React Router, TanStack React Query, Axios, TypeScript, Bun.

**Spec:** `docs/superpowers/specs/2026-11-15-cart-refactor-design.md`

## Global Constraints

- Preserve absolute-quantity updates, one retry for `CART_ITEM_CHANGED`, selected-line semantics, and explicit-only fallback from Flash Sale to regular price.
- A checkout preview never reserves inventory or quota; a changed price/source must require acknowledgement.
- Use OpenAPI-derived DTOs and Axios clients; do not keep `temp/lib/cart-api.ts`.
- Place all migrated tests in `app/frontend/src/test/`.
- Do not add a common component unless it is independently reusable outside cart.

## Review Focus

- A stale line version must cause the cart cache to refresh, not silently overwrite a new value.
- A failed preview with line errors must leave the cart intact and annotate the matching SKU.
- A regular-price catalog add must not include `flash_sale_item_id`.
- The cart must continue to use CSRF proof on PUT, DELETE, and preview POST.
- Removing temporary cart files must leave no import path to `temp/cart` or `temp/lib/cart-api`.

---

### Task 1: Move and preserve cart tests

**Files:**
- Create: `app/frontend/src/test/cart/` tests and `app/frontend/src/test/flashsale/flash-sale-campaign-card.test.tsx`.
- Delete: cart tests under `app/frontend/src/temp/`.

**Interfaces:**
- Produces: centralized behavior tests that import the new cart public modules.

**Acceptance Criteria:**
- Existing pure cart and markup assertions are retained under `src/test/`.
- New service tests exercise Axios request path, method, body, and CSRF header.

- [x] **Step 1: Move pure and component tests, changing imports to the intended feature API**
- [x] **Step 2: Run moved tests to verify they fail because the new feature modules do not exist**

### Task 2: Implement cart feature boundary

**Files:**
- Create: `app/frontend/src/features/cart/cart.type.ts`, `cart.service.ts`, `cart.logic.ts`, `cart.hook.ts`, `components/CartActionNotice.tsx`, `components/CartLineItem.tsx`, `components/CheckoutPreviewPanel.tsx`, `pages/CartPage.tsx`.
- Test: `app/frontend/src/test/cart/`.

**Interfaces:**
- Consumes: generated Cart, CartItem, CheckoutSelection, CheckoutPreview, FieldViolation, and LineProblem schemas.
- Produces: `getCart`, `upsertCartItem`, `removeCartItem`, `previewCheckout`, `useAddToCart`, `useCart`, `useCheckoutPreview`, and cart components/page.

**Acceptance Criteria:**
- Transport paths are `/cart`, `/cart/items/{skuId}`, and `/checkout/preview` through the shared Axios client.
- UI behavior and messages match the legacy tests.

- [x] **Step 1: Implement contract-derived types and service until service tests pass**
- [x] **Step 2: Implement policy and React Query hooks until migrated pure tests pass**
- [x] **Step 3: Implement components and page until migrated markup tests pass**

### Task 3: Integrate consumers and remove legacy cart

**Files:**
- Modify: `app/frontend/src/App.tsx`, `src/routes/AppRoute.tsx`, catalog page, temporary Flash Sale page, and temporary admin Flash Sale action mapping.
- Delete: `app/frontend/src/temp/cart/`, `app/frontend/src/temp/lib/cart-api.ts`, `app/frontend/src/temp/pages/CartPage.tsx`.

**Interfaces:**
- Consumes: Task 2 public cart modules.
- Produces: current `/cart` route and no remaining source dependency on temporary cart.

**Acceptance Criteria:**
- `/cart` is authenticated and accessible from current navigation.
- Existing product and Flash Sale pages keep their add-to-cart notices.

- [x] **Step 1: Update consumers and route/navigation imports**
- [x] **Step 2: Delete legacy cart files and search for stale imports**
- [x] **Step 3: Run `bun test`, `bun run lint`, and `bun run build`**
