# Phase 04 Checkout Design

## Intent

Implement the complete Phase 04 checkout boundary: atomically create a pending-payment order for a selected cart subset, reserve inventory and flash-sale quota, persist idempotent outcomes, and present reliable checkout results to the customer.

## Goals

- Add PostgreSQL migrations and models for orders, immutable order items, reservations, and idempotency records.
- Implement `POST /orders` and `GET /idempotency/orders/{idempotencyKey}` in `internal/module/order`.
- Reserve every selected SKU or reject the full checkout; never partially persist a pending order.
- Lock SKU-dependent rows in canonical ascending SKU order and make all time decisions from PostgreSQL `now()`.
- Use the preview quote as an explicit client confirmation: a changed line/total returns `PRICE_CHANGED`, never silently substitutes a regular price.
- Persist a canonical request hash and replay both successful and business-rejection responses for the same user, operation, and idempotency key.
- Add a checkout form, a persistent client intent key, outcome/unknown-result states, and an order result page.
- Record checkout new/replay/rejection/error metrics through the existing observability boundary when it is available.

## Non-goals

- Payment confirmation, cancellation, expiry worker implementation, notifications, and order-history UI belong to later phases. This phase stores the five-minute deadline and exposes the pending-payment result only.
- Do not mutate or remove cart lines when an order is created; payment-success cart reconciliation belongs to the payment phase.
- No external payment calls, Redis coordination, or client-computed prices.

## Architecture

`order.Service` owns the database transaction. It first coordinates a per-user idempotency record, then reads a single database clock, locks selected cart lines and each SKU's inventory and flash-sale rows in sorted SKU order, re-evaluates the authoritative quote, compares it with the client-confirmed lines/total, and writes the order, snapshots, reservations, counters, and idempotency response in one transaction.

The frontend's order feature owns its transport, idempotency-key persistence, recipient form, and result page. The cart feature supplies an acknowledged preview but does not create orders. A timeout is rendered as unknown until lookup/retry with the same key proves the persisted outcome.

## Data model

- `orders`: owner, public order number, pending-payment status, VND total, recipient snapshot, creation/deadline timestamps.
- `order_items`: immutable SKU/product/name/attribute/price/source snapshots and quantities.
- `inventory_reservations`: one reservation per order item with active/released state for exactly-once later release/finalization.
- Existing `inventories.reserved`, `flash_sale_items.quota_reserved`, and `flash_sale_customer_usage.reserved_quantity` increase only in the successful checkout transaction.
- `idempotency_records`: `(user_id, operation, idempotency_key)` unique identity, canonical request hash, terminal response status/body, and completion timestamp. Same key + different hash returns `IDEMPOTENCY_CONFLICT`; an in-flight record returns `REQUEST_IN_PROGRESS` with retry guidance.

## Transaction rules

1. Validate the request shape and UUIDs before locking.
2. Lock the customer's active cart and selected lines; stale versions return `CART_ITEM_CHANGED`.
3. Lock inventory, flash-sale item, and customer-usage rows by ascending SKU ID, creating/locking usage rows safely for flash-sale lines.
4. Re-evaluate availability, quota, purchase limit, campaign state, and price at database time. Collect safe line failures and roll back before creating any order.
5. Compare the confirmed request to the authoritative quote. A mismatch returns a stable `PRICE_CHANGED` problem containing current line data; never write an order.
6. Insert the order, item snapshots, reservations, and counter increments; set `expires_at = database_now + 5 minutes` exactly once; persist the success response against the idempotency record; commit.

## Contract decisions

- `POST /orders` accepts the existing OpenAPI `PlaceOrderRequest`: the acknowledged `CheckoutLine` array (cart line ID/version, SKU, quantity, unit price, subtotal, price source, and flash-sale item), `currency: VND`, `expected_total`, and a recipient with `name`, `phone`, and `address`. Recipient bounds are those already in OpenAPI: 1–120, 8–20, and 1–500 characters.
- A success is `201` with the OpenAPI `Order` snapshot. A business rejection is the existing RFC-9457 `Problem` shape, using line problems where appropriate. `PRICE_CHANGED` includes a current quote extension that the frontend can feed back into the existing preview acknowledgement flow.
- Idempotency hashes a canonical JSON form of all request fields, including recipient values and client-confirmed quote fields. It stores the complete success or business-rejection response. The operation is `place_order`.
- A concurrent same-key request uses the database uniqueness/row lock. Once the first transaction completes it replays the stored terminal response; only an independently observable non-terminal record returns `409 REQUEST_IN_PROGRESS` with a retryable problem. Same key with another user's identity, or a missing key, returns `404` from lookup so ownership is not disclosed.
- Flash eligibility exactly reuses the current cart quote policy, but evaluates only after locking. Campaign state/window, flash item ownership, SKU/product availability, inventory `on_hand - reserved`, quota `quota_total - quota_reserved - quota_sold`, and customer `reserved_quantity + sold_quantity` are all authoritative. Regular lines carry no flash-sale ID.

## Frontend behavior

- The customer may place an order only after a current preview has been acknowledged and recipient fields are valid.
- One canonical selected-lines/recipient intent owns one generated key; retain it in session storage across reload/retry. Editing the selection, confirmed quote, or recipient creates a new key.
- On network timeout/unknown response, show an unknown state with explicit Retry and Check-result actions, both using the original key. Do not show an order confirmation until the API returns or lookup proves it.
- Business rejections retain the cart and attach returned line problems; price changes return to preview acknowledgement.

## Verification

- Integration tests cover multi-SKU success, any-line rejection rollback, stock/quota/purchase-limit conflicts, canonical lock ordering under concurrency, stable five-minute deadline, replay, same-key/different-payload conflict, and lost-response lookup.
- SQL invariant tests prove `reserved <= on_hand`, flash quota limits, one order per successful key, and no reservations/counter changes after rejection.
- Frontend tests cover recipient validation, stable key reuse/new-key reset, unknown-result UI, and no-success rendering before a proved order response.

## Chosen approach

Implement Steps 1–3 in dependency order as vertical backend/frontend slices. This is the only safe fit for the existing schema: idempotency cannot safely wrap a partial reservation path, and the frontend cannot truthfully resolve a timeout before the lookup endpoint exists.
