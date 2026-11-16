# Cart Refactor Design

## Intent

Migrate the temporary cart implementation into the frontend's feature-first architecture without changing cart, checkout-preview, or error-handling behavior. The feature must use generated OpenAPI types, Axios clients, and React Query while preserving the cart's concurrency and customer-choice rules.

## Goals

- Put cart DTOs and API errors behind `features/cart/cart.type.ts` and `cart.service.ts`.
- Replace temporary fetch/state implementations with React Query hooks in `cart.hook.ts`.
- Keep pure cart policy separate in `cart.logic.ts`: absolute-quantity plans, one retry after `CART_ITEM_CHANGED`, preview-change acknowledgement, and Vietnamese error mapping.
- Move cart UI into `features/cart/components/` and `features/cart/pages/`.
- Move cart-related tests under `src/test/cart/`; split the unrelated Flash Sale component assertion into `src/test/flashsale/`.
- Update catalog, Flash Sale, admin Flash Sale, current routes, and navigation to consume the migrated feature.

## Non-goals

- Do not refactor Flash Sale or admin Flash Sale themselves.
- Do not implement the Phase 04 order-placement flow.
- Do not create common components: CartActionNotice, CartLineItem, price-source badge, and CheckoutPreviewPanel all encode cart-specific state or business copy. No reusable Logo/Avatar-style primitive is introduced.

## Chosen approach

The module follows auth's boundary: contract-derived aliases in `cart.type.ts`; HTTP and CSRF/error translation in `cart.service.ts`; query, mutation, and UI-state hooks in `cart.hook.ts`; domain-only rules in `cart.logic.ts`; rendering in `components/` and `pages/`. Existing cart behavior is preserved by moving its tests before replacing implementations, then adapting transport tests to Axios's adapter boundary.
