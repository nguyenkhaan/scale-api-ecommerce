# Backend agent guide

## Scope and source of truth

You work in `app/backend/` on the Go/Gin API and its workers. Start at [the SDD index](../../docs/CONTENT.md), then read the [PRD](../../docs/product/prd.md), [project spec](../../docs/spec.md), [current phase](../../docs/plans/plan.md), [architecture](../../docs/technical/architecture.md), [data model](../../docs/technical/data-model.md), [OpenAPI contract](../../docs/technical/api/openapi.yaml), and [test strategy](../../docs/technical/testing.md) as relevant to the task. The PRD owns business rules; research notes are historical input.

## Before implementation

1. Check the SDD gate and the current phase. Do not treat a draft plan or unchecked box as approval to implement. Resolve `TECH-01` before choosing the migration/data-access stack and `TECH-02` before building the payment simulator.
2. Identify the affected `FR`/`NFR`, `BR`/`INV`, API operations, and `TEST` IDs in [traceability](../../docs/technical/traceability.md). Break a phase Step into a focused task with acceptance criteria, dependencies, likely files, and a verification command; aim for no more than about five files per task.
3. If behavior changes, update the PRD first, then traceability, domain/data model, OpenAPI, and tests. Record an expensive-to-reverse architecture decision in a new ADR; preserve earlier ADRs. Resolve missing P1 contracts before implementing that P1 task.

## Implementation rules

- PostgreSQL is authoritative for sessions, prices, inventory, sale quotas, purchase limits, idempotency, orders, payments, and outbox state. Redis is only a public-read cache; RabbitMQ transports notifications.
- Multi-SKU checkout commits one complete order with all reservations or a stable business rejection. Lock shared rows in a consistent order, keep transactions short, and make no network calls while a checkout transaction is open.
- Payment success, verified failure, customer cancellation, and expiry must compete through guarded order transitions. Finalize each reservation once. A late success creates a reconciliation case; it cannot revive a terminal order.
- Use database time to decide expiry. Keep the five-minute deadline fixed across retries. Use integer VND amounts and stored order snapshots; never accept client price, total, role, ownership, or payment status as authoritative.
- Use opaque session cookies backed by PostgreSQL. Hash stored tokens, protect cookie-based mutations against CSRF, and check role and ownership on the server. Do not log secrets or sensitive recipient data.
- Keep HTTP handlers focused on transport. Put business decisions and transactions in application/domain code. Pass `context.Context` to I/O and use stable OpenAPI error codes.

## Verification and handoff

- Format changed Go files. From `app/backend/`, run `go build ./...`, `go test ./...`, and `go vet ./...`; run `go test ./... -race` for concurrency work once the relevant test environment exists. Run PostgreSQL integration tests and invariant oracles for any change to money, stock, quota, idempotency, or terminal order state. Use the exact additional commands established during bootstrap; do not claim placeholder Compose or frontend commands work yet.
- Before marking a task done, report its requirement/test IDs, changed behavior, commands and results, any remaining risk, and the phase checkpoint it satisfies. Keep FE informed of contract changes before integration.
