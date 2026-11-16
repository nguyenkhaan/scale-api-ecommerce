# Phase 04 Checkout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Atomically create pending-payment orders from confirmed cart previews, with durable idempotency and an honest frontend outcome flow.

**Architecture:** The order module owns the PostgreSQL transaction and exposes create/lookup endpoints. It reuses cart quote policy under deterministic locks, writes immutable order snapshots/reservations/counters together, and stores terminal idempotency responses. The frontend order feature owns recipient state, intent-key persistence, transport, and order/unknown-result UI; cart remains the preview owner.

**Tech Stack:** Go 1.27, Gin, GORM/PostgreSQL, Goose, React 19, TanStack Query, Axios, TypeScript, Bun.

**Spec:** `docs/superpowers/specs/2026-11-15-phase-04-checkout-design.md`

## Global Constraints

- All reservation, quota, customer-usage, order, and idempotency writes succeed or roll back as one transaction.
- Use PostgreSQL time and canonical ascending SKU lock order; no network operation occurs in a transaction.
- The server recalculates every price; changed confirmed quotes reject with `PRICE_CHANGED`.
- The same `(user, place_order, key, canonical payload)` has one durable terminal result; altered payload returns `IDEMPOTENCY_CONFLICT`.
- A timeout is an unknown frontend state, never a success claim.

## Review Focus

- Mixed regular/flash multi-SKU checkout never increments only a subset of counters.
- A stale cart version, price change, stock/quota/limit rejection, or invalid recipient creates no order/reservation.
- Concurrent same-key requests replay one result while different keys still obey inventory/quota locks.
- The five-minute expiry is persisted from one database clock and retries never extend it.
- Frontend edits reset only the checkout intent key; retry/lookup retains it.

### Task 1: Persist order and idempotency state

**Files:**
- Create: `app/backend/internal/migration/sql/00009_create_orders.sql`, `app/backend/internal/model/{order,order_item,inventory_reservation,idempotency_record}.go`, `app/backend/test/order_migration_test.go`.

**Interfaces:**
- Produces tables and model constants consumed by `order.Service`: pending-payment order, active reservation, and terminal idempotency record.

- [ ] Write migration constraint tests for order status, reservation ownership/quantity, and unique idempotency identity; run them to prove the migration is absent.
- [ ] Add the migration and models with foreign keys/checks/indexes for Phase 04 invariants.
- [ ] Run migration/integration tests and verify constraints pass.

### Task 2: Implement atomic order placement and HTTP API

**Files:**
- Create: `app/backend/internal/module/order/{dto,errors,service,controller,quote,locks}.go`, `app/backend/test/order_{api,concurrency}_test.go`.
- Modify: backend router/bootstrap and problem codes.

**Interfaces:**
- Produces `PlaceOrder(ctx, userID, key, request) (OrderResponse, replayed, error)` and `GetAttempt(ctx, userID, key) (IdempotencyResult, error)`.
- Consumes cart quote/line data and Phase 04 models.

- [ ] Write failing integration tests for a regular multi-SKU success and a rejected line rolling back all state.
- [ ] Implement canonical request validation, deterministic locks, authoritative quote comparison, snapshots/reservations/counter writes, and response mapping.
- [ ] Add failing tests for replay, altered-payload conflict, lookup ownership/missing behavior, and concurrent stock contention; implement durable idempotency handling.
- [ ] Run focused order tests plus the backend suite.

### Task 3: Build frontend checkout intent and result flow

**Files:**
- Create: `app/frontend/src/features/order/{order.type,order.service,order.hook}.ts`, `components/{RecipientForm,CheckoutOutcome}.tsx`, `pages/OrderResultPage.tsx`, focused tests.
- Modify: cart components/page, routes, generated contract if necessary.

**Interfaces:**
- Produces `usePlaceOrder`, persisted `CheckoutIntent`, recipient validation, unknown/retry/lookup UI, and `/orders/:orderId` result route.
- Consumes acknowledged `CheckoutPreview` and Cart selection.

- [ ] Write failing tests for stable intent keys and unknown results not rendering a success state.
- [ ] Implement contract-derived transport, CSRF/idempotency headers, local intent persistence, recipient form, retry/lookup state, and result page.
- [ ] Wire an acknowledged cart preview to order placement and preserve cart on business rejection.
- [ ] Run frontend tests, lint, typecheck, and production build.

### Task 4: Validate Phase 04 boundaries

**Files:**
- Modify: Phase 04 plan checklist and relevant traceability/testing documentation only where evidence requires it.

- [ ] Run the complete backend build/test/vet suite and frontend test/lint/typecheck/build suite.
- [ ] Run the Phase 04 integration tests against PostgreSQL; record any unavailable external-environment evidence explicitly.
- [ ] Check no success UI is reachable from an unresolved network outcome and no rejection leaves persisted reservations.
