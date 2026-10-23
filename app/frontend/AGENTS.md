# Frontend agent guide

## Scope and source of truth

You work in `app/frontend/` on the React + Vite demo for customers and administrators. Start at [the SDD index](../../docs/CONTENT.md), then read the [PRD](../../docs/product/prd.md), [project spec](../../docs/spec.md), [current phase](../../docs/plans/plan.md), [screen list](../../docs/product/screen_list.md), and [OpenAPI contract](../../docs/technical/api/openapi.yaml) as relevant to the task. The PRD owns behavior; OpenAPI owns request and response shapes.

## Before implementation

1. Check the SDD gate and current phase. Do not treat a draft plan or unchecked box as approval to implement. Split a phase Step into a focused task with affected `FR`/`TEST` IDs, acceptance criteria, dependencies, likely files, and a verification step; aim for no more than about five files per task.
2. Agree on the API operation and error codes with BE before connecting a screen. Work in parallel from the OpenAPI schema; fixtures may unblock the UI temporarily, but replace them with the real API at the phase checkpoint.
3. If a user-facing rule changes, update the PRD and [traceability](../../docs/technical/traceability.md) before updating the screen list and API contract. Resolve missing P1 contracts before implementing a P1 screen.

## Implementation rules

- The backend decides prices, totals, stock, sale eligibility, payment deadlines, and order status. A preview does not reserve stock. When a price or condition changes, show the new information and require explicit confirmation.
- Select only the cart lines the customer chose. Show SKU-level business errors and preserve the cart after a rejected checkout. Never silently convert a flash-sale line to regular price or place a partial order.
- Generate one idempotency key per checkout intent. After a timeout or lost response, look up or retry with the same key and unchanged request. Generate a new key only after the customer changes the intent. Do not present an unknown outcome as a definite failure.
- Send session cookies and CSRF proof as specified by OpenAPI. Never trust client-side route guards for authorization, set a role or owner from the browser, or store secrets in frontend code. Render errors by stable `code` and structured line details, not message text.
- Keep customer and admin screens usable with keyboard navigation, labeled controls, clear loading/empty/error states, and narrow layouts. Keep admin UI limited to the PRD demo scope.

## Verification and handoff

- After phase 00 creates the frontend manifest, run its actual build, test, and lint commands. At each integration checkpoint, verify the complete browser flow against the real BE API, including changed prices, timeouts, 401/403, 429, and order status updates where relevant.
- Before marking a task done, report its requirement/test IDs, screen behavior, commands and results, remaining integration gaps, and the phase checkpoint it satisfies. Tell BE about any contract mismatch; do not invent a client-only substitute for missing server behavior.
