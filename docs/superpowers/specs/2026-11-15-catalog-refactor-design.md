# Catalog Refactor Design

## Intent

Remove the obsolete temporary authentication implementation and migrate the customer catalog from `src/temp/` into the feature-first frontend architecture. The catalog must retain its existing product-list, active-SKU selection, pagination, and add-to-cart behavior while adopting the generated OpenAPI contract, Axios clients, and React Query conventions already established by `features/auth`.

## Goals

- Delete legacy auth code, tests, API client, and obsolete temporary bootstrap/pages that solely depend on it.
- Define catalog DTOs from `components['schemas']` in `src/contract/api-contract.ts`.
- Provide catalog service, React Query hooks, components, and page under `src/features/catalog/`.
- Expose the authenticated `/products` route and navigation in the current application shell.
- Preserve legacy catalog behavior and keep temporary cart/admin-flash-sale consumers working during their later migration.

## Non-goals

- Refactor cart, admin flash-sale, or their transport/error layers.
- Add catalog filtering, detail pages, or admin catalog UI.

## Chosen approach

Use the same split as auth: OpenAPI-derived aliases in `catalog.type.ts`, HTTP calls and normalized error translation in `catalog.service.ts`, React Query and pure SKU helpers in `catalog.hook.ts`, rendering in `components/` and `pages/`. The catalog page will temporarily consume the legacy cart hook/components because cart has not yet been migrated; this keeps the existing add-to-cart workflow intact.

Deleting `temp/auth/` also requires deleting the unused temporary app entrypoint, auth-only legacy pages/tests, and `temp/lib/auth-api*`, otherwise TypeScript's full `src` compilation retains dangling imports. The active entrypoint is already `src/main.tsx`.
