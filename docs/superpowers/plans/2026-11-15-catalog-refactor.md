# Catalog Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove legacy temporary auth and migrate the customer catalog into the current feature-first frontend architecture without changing user-visible behavior.

**Architecture:** Catalog types alias generated OpenAPI schemas. An Axios service owns transport and normalized API errors; React Query hooks own product data and pure SKU-selection helpers. The feature page and component keep rendering separate from service logic while temporarily reusing the not-yet-migrated cart UI.

**Tech Stack:** React 19, React Router, TanStack React Query, Axios, TypeScript, Bun.

**Spec:** `docs/superpowers/specs/2026-11-15-catalog-refactor-design.md`

## Global Constraints

- Preserve the existing public catalog list, active-SKU selection, pagination, and regular-price add-to-cart behavior.
- Derive catalog API DTOs from `src/contract/api-contract.ts`; do not duplicate generated schemas.
- Keep cart and admin flash-sale migration out of scope except for imports required to preserve the build.
- Use `@/` imports in migrated source.

## Review Focus

- A product with no active SKU must remain non-purchasable.
- Inactive products or inactive SKUs must not appear as selectable SKU options.
- A returned `next_cursor` must enable fetching and appending the next page.
- API errors must surface through the catalog page's existing retry state.
- Deleting legacy auth must leave no legacy auth import in compiled source.

---

### Task 1: Remove legacy authentication implementation

**Files:**
- Delete: `app/frontend/src/temp/auth/`, `app/frontend/src/temp/lib/auth-api.ts`, `app/frontend/src/temp/lib/auth-api.test.ts`, `app/frontend/src/temp/App.tsx`, `app/frontend/src/temp/main.tsx`, auth-only temporary pages and tests.

**Interfaces:**
- Produces: a source tree with the current `src/main.tsx`/`src/context/AuthProvider.tsx` as the sole auth integration.

**Acceptance Criteria:**
- No compiled source imports `temp/auth` or `temp/lib/auth-api`.
- No obsolete temporary app entrypoint remains.

- [x] **Step 1: Remove obsolete auth sources and their dependent temporary bootstrap/pages**
- [x] **Step 2: Search for legacy auth imports and run `bun run typecheck`**

### Task 2: Create the catalog feature and migrate its consumers

**Files:**
- Create: `app/frontend/src/features/catalog/catalog.type.ts`, `catalog.service.ts`, `catalog.hook.ts`, `components/ProductCard.tsx`, `pages/ProductsPage.tsx`.
- Create: `app/frontend/src/test/catalog/catalog.hook.test.ts`.
- Modify: `app/frontend/src/routes/AppRoute.tsx`, `app/frontend/src/App.tsx`, `app/frontend/src/temp/admin-flash-sale/use-admin-sku-options.ts`.
- Delete: `app/frontend/src/temp/catalog/`, `app/frontend/src/temp/lib/catalog-api.ts`, `app/frontend/src/temp/pages/ProductsPage.tsx`.

**Interfaces:**
- Consumes: OpenAPI `Product`, `Sku`, and `ProductPage` schemas; Axios clients; legacy cart add-to-cart UI.
- Produces: `listProducts`, `listAdminProducts`, `useProductsQuery`, `activeSkus`, `buildSkuOptions`, `mergeCampaignSkuOptions`, and `ProductsPage`.

**Acceptance Criteria:**
- `/products` is an authenticated route in the current app shell.
- Catalog list API calls use `/products` and OpenAPI-derived types.
- SKU helper behavior remains covered by migrated tests.

- [x] **Step 1: Move SKU-helper test to the new hook API and run it to establish the behavior**
- [x] **Step 2: Implement OpenAPI-derived types, Axios service, React Query hooks, component, and page**
- [x] **Step 3: Integrate the current route/navigation and update legacy admin consumer imports**
- [x] **Step 4: Delete legacy catalog files and run the frontend test, typecheck, lint, and build commands**
