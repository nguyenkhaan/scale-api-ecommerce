# SaleScale Frontend UI Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Implement the approved SaleScale frontend UI foundation through the Phase 03 boundary: shared theme, customer/admin shells, reusable components, /api normalization, catalog search, and live FE–BE validation for flash sale, cart, and preview flows.

**Architecture:** Keep feature services, hooks, and business logic in their existing feature folders. Add CSS tokens and presentational primitives under app/frontend/src/components/common, add route-level layouts under app/frontend/src/components/layout, and compose them through React Router layout routes. Normalize one API root for Axios and fetch clients, then validate each UI slice against the real Phase 03 backend before moving on.

**Tech Stack:** React 19, Vite 8, TypeScript, Tailwind CSS 4, React Router 7, TanStack Query 5, Axios, Bun, Go/Gin backend, PostgreSQL test fixture.

**Spec:** docs/superpowers/specs/2026-11-20-frontend-ui-shell-design.md

## Global Constraints

- Treat Phase 03 as the implementation boundary: FR-014–020, INV-09/10, TEST-09/10/20.
- Use only /api routes; remove every frontend /api/v1 path.
- Keep React/Vite/Tailwind/React Router/TanStack Query; do not add a UI framework or icon dependency.
- Use Inter/system-ui typography, the approved SaleScale tokens, 40px controls, 44px touch/action targets, 2px control radius, and 4px card radius.
- Keep customer and admin shells separate; do not put admin navigation into customer pages.
- Keep GET /api/products public; add-to-cart still requires authentication and preserves the next path.
- Catalog search is limited to GET /api/products?query=...; do not build global search or admin/flash-sale search in this phase.
- Do not create order, payment, order-history, inventory, reconciliation, notification, or product-detail screens in this plan.
- Do not use client-side price, total, inventory, quota, role, ownership, or payment state as authoritative.
- Checkout preview must remain read-only: it must not reserve inventory, quota, or usage.
- Backend changes are allowed only when a live integration mismatch proves them necessary; prefer frontend fixes first and preserve business rules, transaction boundaries, invariants, and stable error codes.
- Every BE change must have a focused Go regression test and rerun the impacted FE integration check.
- Do not overwrite docs/ui-design/; it remains the visual source provided by the user.
- Every task ends with the relevant frontend checks and, when it crosses an HTTP boundary, a real backend integration check.

## Review Focus

- API root and preflight: direct-origin requests and Vite-proxy requests must both resolve to one /api prefix; PATCH and DELETE mutations must pass CORS preflight. Test in Task 1.
- Anonymous catalog vs authenticated mutation: an anonymous user can load/search products, while add-to-cart redirects to login without losing the destination. Test in Tasks 4 and 6.
- Search state and pagination: changing query resets the cursor; reload/back preserves the URL query and the displayed result set. Test in Task 6.
- Phase 03 cart/preview errors: price changes, sale inactive, quota/limit failures, and insufficient stock remain line-specific and do not mutate inventory/quota/usage. Test in Task 8 and the live checkpoint.
- Admin state transitions: draft/scheduled/active/stopped/ended actions, field errors, CSRF, and role checks remain visible and recoverable; no invalid action is exposed. Test in Task 9.

---

## Dependency and File Map

The implementation follows this dependency graph:

API root + CORS → theme → common primitives → shells/routes → catalog/search, flash sale, cart/preview, admin flash sale → live Phase 03 checkpoint.

The main files introduced or changed are:

- API/config: app/frontend/src/api/base-url.ts, app/frontend/src/api/client.ts, health.service.ts, flashsale.service.ts, .env.example, and, only if direct-origin integration requires it, app/backend/internal/httpserver/middleware/cors.go.
- Theme/layout: app/frontend/src/styles/theme.css, app/frontend/src/index.css, and the layout components under app/frontend/src/components/layout.
- Common UI: the components under app/frontend/src/components/common, plus focused component tests.
- Existing features: catalog, flash-sale, cart/preview, auth navigation, admin flash-sale pages/components, and their existing tests.
- Phase evidence: docs/plans/phases/phase-03-sale-cart.md.

Components created in a task must export their public prop types. Later tasks consume those exports rather than duplicating styles or status mappings.

## Tasks

### Task 1: Normalize the API root and make Phase 03 mutations browser-compatible

**Files:**
- Create: app/frontend/src/api/base-url.ts
- Test: app/frontend/src/test/api/base-url.test.ts
- Modify: app/frontend/src/api/client.ts
- Modify: app/frontend/src/features/health/health.service.ts
- Modify: app/frontend/src/features/flashsale/flashsale.service.ts
- Modify: app/frontend/src/test/flashsale/flashsale.service.test.ts
- Modify: app/frontend/src/test/health/health.service.test.ts
- Modify: app/frontend/.env.example
- Modify if required by direct-origin integration: app/backend/internal/httpserver/middleware/cors.go
- Test if BE CORS changes: app/backend/test/cors_test.go

**Interfaces:**
- Produce normalizeApiBaseUrl(raw: string | undefined): string, returning /api for an empty value, http://host/api for http://host, and preserving a single existing /api.
- Produce API_BASE_URL, consumed by Axios, health, and flash-sale clients.
- Keep createFlashSaleClient(options: FlashSaleClientOptions) accepting an already normalized API root; its internal paths begin with /flash-sales, never /api/v1.
- Keep createHealthClient(options: HealthClientOptions) using /health/ready relative to the normalized API root.

**Acceptance Criteria:**
- No frontend service or test URL contains /api/v1.
- VITE_API_BASE_URL=http://localhost:4000 and an empty value both resolve correctly.
- Axios, health fetch, and flash-sale fetch all use the same API root.
- Direct-origin OPTIONS requests for PATCH and DELETE succeed with credentials when the backend CORS path is exercised.
- No business behavior or response schema changes.

- [ ] Step 1: Add failing URL normalization and route assertions.
  - In base-url.test.ts, cover empty input, origin input, input ending in /api, and trailing slashes.
  - Update flash-sale service tests to expect http://api.test/api/flash-sales and http://api.test/api/admin/flash-sales.
  - Update health service tests to expect http://localhost:4000/api/health/ready.
  - Add a backend preflight test only if the direct-origin path is part of the chosen integration environment; assert allowed methods include PATCH, DELETE, and OPTIONS.

- [ ] Step 2: Run focused tests and verify they fail for the current /api/v1 and split-env behavior.
  - Run: cd app/frontend && bun test src/test/api/base-url.test.ts src/test/flashsale/flashsale.service.test.ts src/test/health/health.service.test.ts
  - Expected: FAIL on the old URL expectations or missing normalizer.

- [ ] Step 3: Implement normalizeApiBaseUrl and route all clients through it.
  - Use VITE_API_BASE_URL as the only frontend API environment variable.
  - Make api/client.ts use API_BASE_URL as its Axios baseURL.
  - Change health and flash-sale path constants to be relative to /api.
  - Update .env.example and keep Vite proxy target unchanged.
  - If browser direct-origin testing fails at CORS preflight, minimally add PATCH, DELETE, and OPTIONS to Access-Control-Allow-Methods in cors.go; do not broaden authorization or change route behavior.

- [ ] Step 4: Run focused tests and the impacted backend checks.
  - Run: cd app/frontend && bun test src/test/api/base-url.test.ts src/test/flashsale/flashsale.service.test.ts src/test/health/health.service.test.ts
  - Run if cors.go changed: cd app/backend && gofmt -w internal/httpserver/middleware/cors.go test/cors_test.go && go test ./test -run TestCORS -count=1
  - Expected: PASS; all emitted URLs contain exactly one /api.

- [ ] Step 5: Commit the API-root slice.
  - Run: git add app/frontend/src/api app/frontend/src/features/health app/frontend/src/features/flashsale app/frontend/src/test/api app/frontend/src/test/flashsale app/frontend/src/test/health app/frontend/.env.example app/backend/internal/httpserver/middleware/cors.go app/backend/test/cors_test.go
  - Commit: fix: normalize frontend api routes for phase three

### Task 2: Add the shared SaleScale theme tokens and base CSS

**Files:**
- Create: app/frontend/src/styles/theme.css
- Modify: app/frontend/src/index.css

**Interfaces:**
- Produce semantic CSS variables for the approved palette, typography, spacing, radii, control heights, focus ring, and state colors.
- Produce base rules for body, links, buttons, focus-visible, and reduced motion that later components can rely on.

**Acceptance Criteria:**
- The app background, text, surfaces, borders, primary orange, success, warning, and danger states match the design document.
- The base stylesheet does not introduce page-specific styling or feature logic.
- Focus indicators remain visible and motion is disabled or reduced under prefers-reduced-motion.

- [ ] Step 1: Write the token stylesheet with the exact values from the design.
  - Define #f5f5f5 app background, #ffffff surface, #222222 primary text, #757575 muted text, #e8e8e8 border, #bdbdbd input border, #ee4d2d primary, #d73211 primary-strong, #26aa99 success, #ffbf00 warning, #d0011b danger, control heights 40px/44px, radii 2px/4px, and the 4/8/12/16/24/32 spacing rhythm.
  - Keep Inter first in the font stack with system fallbacks.

- [ ] Step 2: Import the token stylesheet from index.css and add only global/base rules.
  - Keep the Tailwind import.
  - Set body margin, min-width, background, color, font, line-height, and text rendering.
  - Add a single visible focus-visible rule using the theme ring.
  - Add a reduced-motion media rule.

- [ ] Step 3: Run the frontend build and inspect the base shell at 390px and 1280px.
  - Run: cd app/frontend && bun run build
  - Expected: PASS; no layout-wide overflow introduced.

- [ ] Step 4: Commit the theme slice.
  - Run: git add app/frontend/src/styles/theme.css app/frontend/src/index.css
  - Commit: feat: add salescale frontend theme tokens

### Task 3: Build common structural, feedback, and dialog primitives

**Files:**
- Create: app/frontend/src/components/common/Button.tsx
- Create: app/frontend/src/components/common/IconButton.tsx
- Create: app/frontend/src/components/common/Surface.tsx
- Create: app/frontend/src/components/common/Alert.tsx
- Create: app/frontend/src/components/common/FeedbackState.tsx
- Create: app/frontend/src/components/common/Dialog.tsx
- Create: app/frontend/src/components/common/index.ts
- Test: app/frontend/src/test/components/common/feedback-components.test.tsx

**Interfaces:**
- ButtonProps: native button props plus variant primary, secondary, danger, or ghost and loading boolean.
- IconButtonProps: native button props plus required label string and optional pressed boolean.
- SurfaceProps: children, optional as, and optional className; no data fetching.
- AlertProps: tone info, success, warning, or danger; children; optional role.
- FeedbackStateProps: kind loading, empty, error, forbidden, notFound, or overload; title; message; optional action with label, onClick, and disabled.
- DialogProps: open state, title, onClose, children, and optional footer; it exposes role dialog, aria-modal, Escape close, and a visible close control.

**Acceptance Criteria:**
- All controls have 40–44px usable hit areas, disabled semantics, loading feedback, and visible focus.
- Feedback states use the correct role, state copy, retry/action path, and no color-only meaning.
- Dialog markup is keyboard dismissible and does not place focus behind the overlay.
- Tests cover primary/disabled/loading button, icon label, error alert, loading/empty/error feedback, and Escape dialog close.

- [ ] Step 1: Write failing component tests.
  - Assert Button variants and disabled/aria-busy.
  - Assert IconButton exposes the supplied accessible name.
  - Assert FeedbackState renders role status for loading and role alert for error.
  - Assert Dialog renders aria-modal true and the close action.

- [ ] Step 2: Run the focused component test.
  - Run: cd app/frontend && bun test src/test/components/common/feedback-components.test.tsx
  - Expected: FAIL because the common components do not exist.

- [ ] Step 3: Implement the primitives using theme tokens and native HTML controls.
  - Keep all components presentational.
  - Use inline SVG only for structural icons; decorative SVG receives aria-hidden true.
  - Do not add an icon package.

- [ ] Step 4: Run the focused test and full frontend type checks.
  - Run: cd app/frontend && bun test src/test/components/common/feedback-components.test.tsx && bun run typecheck
  - Expected: PASS.

- [ ] Step 5: Commit the common primitive slice.
  - Run: git add app/frontend/src/components/common app/frontend/src/test/components/common
  - Commit: feat: add shared frontend feedback primitives

### Task 4: Build domain/form common components used by Phase 03

**Files:**
- Create: app/frontend/src/components/common/StatusBadge.tsx
- Create: app/frontend/src/components/common/PriceBlock.tsx
- Create: app/frontend/src/components/common/QuantityStepper.tsx
- Create: app/frontend/src/components/common/FormField.tsx
- Create: app/frontend/src/components/common/SearchField.tsx
- Create: app/frontend/src/components/common/LoadMoreButton.tsx
- Modify: app/frontend/src/components/common/index.ts
- Test: app/frontend/src/test/components/common/domain-components.test.tsx

**Interfaces:**
- StatusBadgeProps: domain order, payment, flashSale, or priceSource; value; optional audience customer or admin.
- PriceBlockProps: regular price, effective price, source, optional end time, and optional compact mode; money uses formatVND.
- QuantityStepperProps: value, min, max, busy, label, and onChange(value: number).
- FormFieldProps: id, label, errorMessages, optional helper, and a render/children slot that receives aria-invalid and aria-describedby.
- SearchFieldProps: value, placeholder, onChange, onSubmit, and optional label.
- LoadMoreButtonProps: loading, disabled, and onClick.

**Acceptance Criteria:**
- Status labels and colors match the HTML source for flash sale and price source.
- Price blocks distinguish regular price, sale price, old price, and source without relying on color alone.
- Quantity controls expose an accessible group and disable the correct action at min/max/busy.
- Form fields connect labels, helpers, and inline errors.
- Search submits through a native form and supports keyboard use.

- [ ] Step 1: Write tests for status mapping and control semantics.
  - Cover customer/admin flash-sale labels, price-source labels, VND formatting, quantity min/max, aria-invalid, and search submit.
  - Assert StatusBadge renders text plus a non-color indicator/icon.

- [ ] Step 2: Run the focused test and confirm it fails.
  - Run: cd app/frontend && bun test src/test/components/common/domain-components.test.tsx
  - Expected: FAIL because the shared domain components do not exist.

- [ ] Step 3: Implement the components as pure presentational components.
  - Reuse formatVND and existing date formatting helpers.
  - Keep status mapping in one file; feature pages must not duplicate status color classes.
  - Defer Breadcrumbs, Tabs, and DataTable until a consuming page enters a later phase.

- [ ] Step 4: Run tests and typecheck.
  - Run: cd app/frontend && bun test src/test/components/common/domain-components.test.tsx && bun run typecheck
  - Expected: PASS.

- [ ] Step 5: Commit the domain component slice.
  - Run: git add app/frontend/src/components/common app/frontend/src/test/components/common/domain-components.test.tsx
  - Commit: feat: add shared salescale domain components

### Task 5: Introduce customer/admin shells and layout routes

**Files:**
- Create: app/frontend/src/components/layout/CustomerShell.tsx
- Create: app/frontend/src/components/layout/CustomerHeader.tsx
- Create: app/frontend/src/components/layout/AdminShell.tsx
- Create: app/frontend/src/components/layout/AdminSidebar.tsx
- Create: app/frontend/src/components/layout/AdminHeader.tsx
- Modify: app/frontend/src/App.tsx
- Modify: app/frontend/src/routes/AppRoute.tsx
- Modify: app/frontend/src/features/auth/components/AuthNav.tsx
- Test: app/frontend/src/test/components/layout/shells.test.tsx

**Interfaces:**
- CustomerShell and AdminShell render an Outlet; they do not fetch data.
- CustomerHeader consumes AuthNav, SearchField, and route links; search submits to /products?query=....
- AdminSidebar receives current location from React Router and renders only implemented Phase 03 admin routes, currently /admin/flash-sales.
- AdminHeader consumes the current authenticated user and logout callback.
- Route groups preserve RequireAuth for cart/account and RequireRole('admin') for admin routes.

**Acceptance Criteria:**
- App.tsx no longer owns a second global header.
- / renders the public catalog entry point; /products and /flash-sales are public.
- /cart, /account, and admin routes still enforce their existing guards.
- Customer shell matches the orange customer header and is usable at 390px.
- Admin shell has no dead Products/Tồn kho links before their routes exist; flash-sale navigation has an active state.
- Shell tests assert header/search/cart/account controls, aria-current, and admin navigation.

- [ ] Step 1: Write shell and route tests against MemoryRouter.
  - Render each shell with a test outlet and assert landmarks, links, search label, active state, and no dead admin links.
  - Add a route test asserting anonymous /products can render while anonymous /cart redirects to login.

- [ ] Step 2: Run focused tests and verify the current layout/routing fails the new assertions.
  - Run: cd app/frontend && bun test src/test/components/layout/shells.test.tsx src/test/phase3-pages.test.tsx
  - Expected: FAIL on the old global header, guarded catalog route, or missing shell components.

- [ ] Step 3: Move the global header into CustomerShell and add the admin layout route.
  - Use React Router Outlet layout routes.
  - Keep HealthPage as a separate diagnostic route.
  - Change / to the public catalog route; preserve next in auth redirects.
  - Make mobile admin navigation a keyboard-accessible disclosure/drawer, not a hover-only menu.

- [ ] Step 4: Run shell tests, typecheck, and build.
  - Run: cd app/frontend && bun test src/test/components/layout/shells.test.tsx src/test/phase3-pages.test.tsx && bun run typecheck && bun run build
  - Expected: PASS with no horizontal overflow at 390px.

- [ ] Step 5: Commit the shell/routing slice.
  - Run: git add app/frontend/src/App.tsx app/frontend/src/routes/AppRoute.tsx app/frontend/src/components/layout app/frontend/src/features/auth/components/AuthNav.tsx app/frontend/src/test/components/layout app/frontend/src/test/phase3-pages.test.tsx
  - Commit: feat: add customer and admin layout shells

### Task 6: Refactor the public catalog and add URL-backed search

**Files:**
- Modify: app/frontend/src/features/catalog/catalog.hook.ts
- Modify: app/frontend/src/features/catalog/pages/ProductsPage.tsx
- Modify: app/frontend/src/features/catalog/components/ProductCard.tsx
- Modify if request mapping needs adjustment: app/frontend/src/features/catalog/catalog.service.ts
- Modify: app/frontend/src/test/catalog/catalog.hook.test.ts
- Create or modify: app/frontend/src/test/catalog/ProductsPage.test.tsx

**Interfaces:**
- Change catalogQueryKeys.products to accept limit and query.
- Change useProductsQuery(query: string) to include query in its key and pass query or undefined to listProducts.
- Keep listProducts(params: ProductListParams): Promise<ProductPage> and use the generated OpenAPI type for query.
- ProductsPage owns useSearchParams; changing the search term resets the infinite query cursor.

**Acceptance Criteria:**
- Anonymous users can load /products.
- Submitting Áo requests /api/products?query=...&limit=20 and updates the URL.
- Changing or clearing the query resets pagination; reload/back restores the query.
- Empty, loading, network, 429, and generic error states use common feedback components.
- Add-to-cart keeps the existing auth requirement and next=/products?query=....
- Product cards use the approved surface, price, SKU selector, quantity control, and primary action.

- [ ] Step 1: Add failing hook/page tests.
  - Assert query key separation for empty vs non-empty query.
  - Assert listProducts receives query and cursor reset behavior.
  - Assert anonymous page markup has search, product grid, empty state, and add-to-cart CTA.

- [ ] Step 2: Run focused catalog tests and verify failure.
  - Run: cd app/frontend && bun test src/test/catalog/catalog.hook.test.ts src/test/catalog/ProductsPage.test.tsx
  - Expected: FAIL on the missing query argument and old page markup.

- [ ] Step 3: Implement URL-backed query state and visual refactor.
  - Read query from useSearchParams, submit through SearchField, and update the URL without duplicating server state.
  - Include query in the TanStack Query key so a new term starts at the first page.
  - Remove the stale login-required assumption from generic catalog load errors.
  - Reuse PriceBlock, QuantityStepper, Button, Surface, FeedbackState, and LoadMoreButton.

- [ ] Step 4: Run focused tests and FE checks.
  - Run: cd app/frontend && bun test src/test/catalog/catalog.hook.test.ts src/test/catalog/ProductsPage.test.tsx src/test/phase3-pages.test.tsx && bun run typecheck && bun run lint
  - Expected: PASS.

- [ ] Step 5: Commit the catalog slice.
  - Run: git add app/frontend/src/features/catalog app/frontend/src/test/catalog app/frontend/src/test/phase3-pages.test.tsx
  - Commit: feat: align public catalog with salescale design

### Task 7: Refactor customer flash-sale browsing against the Phase 03 API

**Files:**
- Modify: app/frontend/src/features/flashsale/pages/FlashSalesPage.tsx
- Modify: app/frontend/src/features/flashsale/components/FlashSaleCampaignCard.tsx
- Modify: app/frontend/src/features/flashsale/components/Countdown.tsx
- Modify: app/frontend/src/features/flashsale/flashsale.logic.ts
- Modify: app/frontend/src/test/flashsale/flashsale.service.test.ts
- Create or modify: app/frontend/src/test/flashsale/FlashSalesPage.test.tsx

**Interfaces:**
- Keep useFlashSales and flashSaleClient.listPublicFlashSales as the data boundary.
- FlashSaleCampaignCard continues to emit onAdd(skuId, flashSaleItemId, productName).
- Countdown remains display-only; backend status and API result remain authoritative.
- saleStatusLabel supports customer labels separately from admin labels where the sample distinguishes Sắp diễn ra and Đã lên lịch.

**Acceptance Criteria:**
- Public list works with /api/flash-sales, query pagination, and public status values.
- Active/scheduled/ended cards match the sample hierarchy: status, campaign time, price block, limit, countdown, and CTA.
- Scheduled and ended campaigns do not expose add-to-cart actions.
- Loading, empty, network, and API error states are common and retryable.
- Existing add-to-cart mutation behavior and line price source remain unchanged.

- [ ] Step 1: Update service URL assertions and add card/page tests.
  - Assert public and admin service paths use /api.
  - Cover active, scheduled, ended, empty, and failed list states.
  - Assert countdown has role timer but no per-second screen-reader announcement.

- [ ] Step 2: Run focused tests and verify old URL/style assertions fail.
  - Run: cd app/frontend && bun test src/test/flashsale/flashsale.service.test.ts src/test/flashsale/FlashSalesPage.test.tsx
  - Expected: FAIL until paths and UI are updated.

- [ ] Step 3: Implement the visual refactor using common components.
  - Use StatusBadge, PriceBlock, Countdown, Surface, Button, FeedbackState, and LoadMoreButton.
  - Keep backend status as the source of action availability; do not infer sale eligibility from the client clock alone.

- [ ] Step 4: Run tests, typecheck, lint, and build.
  - Run: cd app/frontend && bun test src/test/flashsale/flashsale.service.test.ts src/test/flashsale/FlashSalesPage.test.tsx && bun run typecheck && bun run lint && bun run build
  - Expected: PASS.

- [ ] Step 5: Commit the customer flash-sale slice.
  - Run: git add app/frontend/src/features/flashsale app/frontend/src/test/flashsale
  - Commit: feat: align customer flash sale screens

### Task 8: Refactor cart and checkout preview without introducing Phase 04 ordering

**Files:**
- Modify: app/frontend/src/features/cart/pages/CartPage.tsx
- Modify: app/frontend/src/features/cart/components/CartLineItem.tsx
- Modify: app/frontend/src/features/cart/components/CheckoutPreviewPanel.tsx
- Modify if needed: app/frontend/src/features/cart/components/CartActionNotice.tsx
- Modify: app/frontend/src/test/cart/cart-components.test.tsx
- Modify: app/frontend/src/test/cart/cart.service.test.ts
- Modify: app/frontend/src/test/cart/cart.logic.preview-changes.test.ts

**Interfaces:**
- Keep useCart and useCheckoutPreview(cart) as the feature state boundaries.
- Keep previewCheckout(selection: CheckoutSelection): Promise<CheckoutPreview> read-only.
- Keep CartLineItem callbacks for toggle, quantity, remove, and switch-to-regular.
- CheckoutPreviewPanel continues to require a preview and acknowledgement before the Phase 04 order action; it must not call an order service in this plan.

**Acceptance Criteria:**
- Cart layout matches the sample at desktop and mobile.
- Line errors appear below the correct SKU and preserve the cart.
- Price changes require explicit acknowledgement before the order CTA becomes enabled.
- Empty, loading, network, CART_ITEM_CHANGED, SALE_NOT_ACTIVE, SALE_QUOTA_EXCEEDED, PURCHASE_LIMIT_EXCEEDED, and INSUFFICIENT_STOCK states are recoverable.
- Preview tests prove no frontend mutation is introduced; live verification proves inventory/quota/usage do not change.

- [ ] Step 1: Extend component tests for design states.
  - Assert common status/price components are used in markup.
  - Assert price-change acknowledgement, disabled order action, line error placement, empty cart, and retry action.
  - Assert mutation requests preserve expected_version, price_source, and flash_sale_item_id.

- [ ] Step 2: Run focused cart tests and verify the old markup fails the new assertions.
  - Run: cd app/frontend && bun test src/test/cart/cart-components.test.tsx src/test/cart/cart.service.test.ts src/test/cart/cart.logic.preview-changes.test.ts
  - Expected: FAIL on the old layout or missing state assertions.

- [ ] Step 3: Implement the visual refactor with shared components.
  - Use Surface, Alert, StatusBadge, PriceBlock, QuantityStepper, Button, and FeedbackState.
  - Keep Đặt hàng (Phase 04) visibly deferred and disabled until a future order implementation; do not fake success or navigate to payment.
  - Keep cart data and preview state in existing hooks/services.

- [ ] Step 4: Run focused tests and FE checks.
  - Run: cd app/frontend && bun test src/test/cart/cart-components.test.tsx src/test/cart/cart.service.test.ts src/test/cart/cart.logic.preview-changes.test.ts && bun run typecheck && bun run lint && bun run build
  - Expected: PASS.

- [ ] Step 5: Commit the cart/preview slice.
  - Run: git add app/frontend/src/features/cart app/frontend/src/test/cart
  - Commit: feat: align cart and preview screens

### Task 9: Refactor the admin flash-sale workflow inside the admin shell

**Files:**
- Modify: app/frontend/src/features/admin-flashsale/pages/AdminPage.tsx
- Modify: app/frontend/src/features/admin-flashsale/pages/AdminFlashSalePage.tsx
- Modify: app/frontend/src/features/admin-flashsale/components/AdminCampaignCard.tsx
- Modify: app/frontend/src/features/admin-flashsale/components/CampaignFormModal.tsx
- Modify: app/frontend/src/features/admin-flashsale/components/CampaignItemRow.tsx
- Modify: app/frontend/src/features/admin-flashsale/components/FieldErrors.tsx
- Modify: app/frontend/src/test/admin-flashsale/admin-flashsale.logic.test.ts
- Create or modify: app/frontend/src/test/admin-flashsale/AdminFlashSalePage.test.tsx

**Interfaces:**
- Keep useAdminFlashSales as the mutation/list boundary.
- Keep adminActionsFor(status) as the single action-availability decision.
- CampaignFormModal consumes FormField, Dialog, Button, Alert, and StatusBadge; it continues to emit CreateFlashSaleRequest.
- AdminCampaignCard consumes admin status labels and emits AdminAction.
- Admin shell renders only the implemented flash-sale route.

**Acceptance Criteria:**
- Admin list/form match the sample’s sidebar/content/table hierarchy while remaining usable on mobile.
- Draft, scheduled, active, stopped, and ended campaigns expose only valid actions.
- Form errors map to fields and summary; submitting state disables duplicate actions.
- CSRF and role/authorization failures remain visible with retry/reload paths.
- No emoji is used as a structural icon.
- Create, update draft, schedule, cancel schedule, and stop continue using the existing backend services.

- [ ] Step 1: Add page/action/error tests.
  - Cover action availability per status, invalid form fields, duplicate submit prevention, close/escape behavior, and API conflict reload.
  - Assert no Product/Inventory dead links appear in current admin navigation.

- [ ] Step 2: Run focused admin tests and verify failure.
  - Run: cd app/frontend && bun test src/test/admin-flashsale/admin-flashsale.logic.test.ts src/test/admin-flashsale/AdminFlashSalePage.test.tsx
  - Expected: FAIL on the old card/modal structure or missing shared-component semantics.

- [ ] Step 3: Implement the visual/admin refactor.
  - Replace rounded/shadow-heavy cards with flat bordered surfaces, compact status badges, table-like rows, and clear primary/secondary actions.
  - Move modal semantics to Dialog; preserve field error IDs and focus on failed submission.
  - Keep the backend action contract unchanged; any backend mismatch is handled under Task 1’s minimal-change policy.

- [ ] Step 4: Run focused tests and FE checks.
  - Run: cd app/frontend && bun test src/test/admin-flashsale/admin-flashsale.logic.test.ts src/test/admin-flashsale/AdminFlashSalePage.test.tsx && bun run typecheck && bun run lint && bun run build
  - Expected: PASS.

- [ ] Step 5: Commit the admin slice.
  - Run: git add app/frontend/src/features/admin-flashsale app/frontend/src/test/admin-flashsale
  - Commit: feat: align admin flash sale workflow

### Task 10: Run the live Phase 03 FE–BE checkpoint and update phase evidence

**Files:**
- Modify after successful verification: docs/plans/phases/phase-03-sale-cart.md
- No new browser framework or UI dependency unless an existing project command requires it.

**Interfaces:**
- Consume the Phase 03 frontend routes and services from Tasks 1–9.
- Consume the existing backend API, migrations, seed fixture, OpenAPI contract, and Phase 03 test suite.
- Produce recorded evidence containing commands, viewport coverage, requirement/test IDs, browser behavior, backend data oracle results, and remaining gaps.

**Acceptance Criteria:**
- FE and BE are tested against the same running API, not only mocked service responses.
- Browser smoke passes at desktop and mobile widths for the current scope.
- Phase 03 sale/cart/preview behavior remains correct after visual refactors.
- Any backend patch is minimal, tested, and called out in phase evidence.
- The phase document does not claim browser completion before the browser run actually passes.

- [ ] Step 1: Prepare the test environment.
  - From app/backend, apply migrations and seed only a disposable development/test database following repository guards.
  - Start the API on http://localhost:4000.
  - From app/frontend, start Vite with VITE_API_BASE_URL= to exercise the proxy path; repeat with VITE_API_BASE_URL=http://localhost:4000 if direct-origin CORS coverage is required.

- [ ] Step 2: Run automated gates.
  - Run: cd app/backend && gofmt -l . && go build ./... && go test ./... && go vet ./...
  - Run: cd app/frontend && bun run typecheck && bun run lint && bun test && bun run build
  - Run the established Phase 03 API e2e scenario and preserve the 35/35 result or updated equivalent output.
  - Expected: all commands pass; skipped database tests are reported explicitly with their environment requirement.

- [ ] Step 3: Run browser smoke at 1280px and 390px.
  - Customer: anonymous / and /products, catalog search, login/register, public flash-sale list, add regular/sale SKU, cart quantity/remove/select, preview, price-change acknowledgement, and line-error recovery.
  - Admin: login as admin, open /admin/flash-sales, create draft, edit draft, schedule/cancel/stop where fixture state permits, submit invalid form, and recover from API error.
  - Check keyboard navigation, focus visibility, mobile overflow, shell active state, cookie/CSRF requests, and no dead navigation links.

- [ ] Step 4: Verify backend data oracles after cart/preview flows.
  - Compare inventory on_hand, reserved, quota reserved/sold, and customer usage before/after preview.
  - Confirm rejected preview leaves cart intact and does not create reservation/order data.
  - If an integration mismatch appears, patch only the smallest BE surface, add its Go regression test, rerun Task 1 checks and the full impacted scenario.

- [ ] Step 5: Record evidence and close the Phase 03 checkpoint.
  - Update docs/plans/phases/phase-03-sale-cart.md with actual commands/results, browser viewport results, requirement/test IDs, remaining risk, and any minimal BE change.
  - Do not mark unrelated Phase 04/05 screens complete.
  - Commit: docs: record phase three frontend integration checkpoint

## Final Verification Checklist

Before claiming the plan is implemented:

- [ ] rg -n /api/v1 app/frontend returns no matches.
- [ ] cd app/frontend && bun run typecheck
- [ ] cd app/frontend && bun run lint
- [ ] cd app/frontend && bun test
- [ ] cd app/frontend && bun run build
- [ ] cd app/backend && go build ./...
- [ ] cd app/backend && go test ./...
- [ ] cd app/backend && go vet ./...
- [ ] Phase 03 API e2e passes against the real backend.
- [ ] Browser smoke passes at 1280px and 390px.
- [ ] Cart preview leaves inventory/quota/usage unchanged.
- [ ] All changed behavior is mapped to FR-014–020, INV-09/10, TEST-09/10/20.
- [ ] Phase 03 evidence is updated only after the live checkpoint passes.

