# Legacy Feature Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the remaining `src/temp/` implementation by preserving it behind feature-first frontend modules.

**Architecture:** Shared formatting, CSRF, problem parsing, pagination, and violation translation live in `src/lib`. Flash Sale owns its public API, types, hooks, components and page; Admin Flash Sale owns the admin workflow and consumes Flash Sale types; Health is an isolated feature with service, hook, component and page.

**Tech Stack:** React 19, React Router, TypeScript, Bun, browser Fetch.

**Spec:** `docs/superpowers/specs/2026-11-15-legacy-feature-migration-design.md`

## Global Constraints

- Preserve all current request URLs, credentials, CSRF behavior, Vietnamese copy, and sale/admin flows.
- No new dependencies or visual redesign.
- Do not alter existing auth, catalog, or cart behavior; update only their shared-helper imports.
- End with no source import to `temp/` and no remaining `src/temp/` files.

## Review Focus

- Non-GET Flash Sale requests retain the CSRF header and cookie credentials.
- Public and admin cursor pagination retains query parameters and append/reload behavior.
- A campaign status transition conflict reloads the admin list rather than presenting stale success.
- Flash Sale add-to-cart remains unavailable for non-active campaigns.
- Health retry remains accessible and readiness errors do not expose response bodies.

### Task 1: Establish shared and public Flash Sale boundaries

**Files:**
- Create: `app/frontend/src/lib/{api-error,csrf,format,json-http,load-error,use-paged-list,violation-messages}.ts`, `app/frontend/src/features/flashsale/{flashsale.type.ts,flashsale.service.ts,flashsale.hook.ts}`, `components/*`, `pages/FlashSalesPage.tsx`.
- Modify: catalog/cart imports and Flash Sale tests.
- Delete: corresponding `src/temp/lib/` and `src/temp/flash-sale/` files plus `temp/pages/FlashSalesPage.tsx`.

**Interfaces:**
- Produces: `flashSaleClient`, `FlashSaleCampaign`, `useFlashSales`, and public page/component imports under `features/flashsale`.

- [x] Write a migration test importing the public Flash Sale card and page from `features/flashsale`; run it and observe the missing-module failure.
- [x] Move shared helpers and public Flash Sale code, updating relative imports and preserving client request behavior.
- [x] Move existing Flash Sale tests to `src/test/flashsale/`, update imports, and run the focused Flash Sale tests.

### Task 2: Establish Admin Flash Sale boundary

**Files:**
- Create: `app/frontend/src/features/admin-flashsale/admin-flashsale.hook.ts`, `admin-flashsale.logic.ts`, `components/*`, `pages/AdminFlashSalePage.tsx`.
- Modify: phase-three tests.
- Delete: `app/frontend/src/temp/admin-flash-sale/` and `temp/pages/AdminFlashSalePage.tsx`.

**Interfaces:**
- Consumes: Flash Sale service/types and shared helpers.
- Produces: `useAdminFlashSales`, `useAdminSkuOptions`, form/action logic, and `AdminFlashSalePage` under the admin feature.

- [x] Change the phase-three page test to import the feature page; run it and observe the missing-module failure.
- [x] Move the admin components, hooks, and pure logic; update imports to shared/public feature boundaries.
- [x] Move its logic test to `src/test/admin-flashsale/` and run focused admin tests.

### Task 3: Establish Health feature and integrate routes

**Files:**
- Create: `app/frontend/src/features/health/health.service.ts`, `health.hook.ts`, `components/HealthStatusView.tsx`, `pages/HealthPage.tsx`.
- Modify: `app/frontend/src/App.tsx`, `src/routes/AppRoute.tsx`, page/card tests.
- Delete: `app/frontend/src/temp/lib/health-api*`, `src/temp/pages/{HealthPage,HealthPage.test}.tsx`, `src/temp/pages/AdminPage.tsx`, and the empty `src/temp/` tree.

**Interfaces:**
- Produces: feature imports for the remaining pages and routes to public Flash Sale, admin Flash Sale, and Health.

- [x] Move Health tests to `src/test/health/`, update imports, and run them to establish the feature boundary.
- [x] Move Health service/hook/component/page and add feature routes guarded by the existing auth/role components.
- [x] Search for stale temporary imports, then run the full frontend test, lint, typecheck, and build commands.
