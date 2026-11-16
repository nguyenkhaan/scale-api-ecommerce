# Legacy Feature Migration Design

## Intent

Move every remaining frontend source file under `app/frontend/src/temp/` into the feature-first source tree without changing user-visible behavior, request semantics, or existing Vietnamese copy.

## Goals

- Place public Flash Sale code in `features/flashsale/` with `flashsale.type.ts`, `flashsale.service.ts`, `flashsale.hook.ts`, `components/`, and `pages/`.
- Place administrator-only campaign management in `features/admin-flashsale/` with its own hook, components, and page while consuming the public Flash Sale contracts.
- Place readiness checking in `features/health/` with a service, hook, components, and page.
- Move cross-feature formatting, HTTP-problem parsing, CSRF, pagination, and violation-copy helpers to `src/lib/`.
- Update routes and tests so compiled frontend code has no `temp/` import, then delete `src/temp/`.

## Constraints

- Preserve request paths, cookie credentials, CSRF headers, pagination behavior, status transitions, field validation, and current UI copy.
- Do not redesign the UI or add dependencies.
- Keep the server authoritative for sale availability and campaign transitions.
- Existing dirty migration work for auth, catalog, and cart remains untouched except for their imports of moved shared utilities.

## Chosen approach

Perform a mechanical feature-boundary migration. The public Flash Sale service continues to own the existing fetch-based client and contracts; hooks own list state. Admin campaign code stays separate because it has an authorization and workflow boundary, but it imports Flash Sale request/response types. Health has a small independent boundary. Shared helpers move to `src/lib/` because catalog and cart already consume formatting and validation copy.
