# Frontend source-staging design

## Intent

Reset the frontend source layout before routing is rebuilt. The current feature code and its import graph will be preserved intact in a temporary archive, while `src/` receives an empty target structure that makes future module boundaries explicit.

## Goals

- Move every current `*.ts` and `*.tsx` file below `app/frontend/src/` into `app/frontend/src/temp/`, preserving each file's path relative to `src/`.
- This includes the application entrypoint, `App.tsx`, all feature files, library files, pages, and tests.
- Create the requested empty folders using `.gitkeep`: `api`, `assets`, `components/common`, `components/layout`, `context`, `contract`, `features`, `hooks`, `lib`, `pages`, `routes`, and `store/slices`.
- Create empty feature-module folders using `.gitkeep`: `auth`, `product`, `flashsale`, `payment`, `bill`, and `cart`.
- Leave non-TypeScript source assets and configuration files in place, including `index.css`, `tsconfig.json`, `vite.config.ts`, and package manifests.

## Non-goals

- Do not repair import paths, routing, or the Vite entrypoint in this pass.
- Do not move or rewrite CSS, configuration, dependency manifests, or static assets.
- Do not add TanStack Query, Redux, route definitions, contracts, hooks, services, or feature implementations yet.
- Do not rename, delete, or merge the archived source files.

## Chosen approach

Archive the entire TypeScript source tree under `src/temp/` and build only the requested directory skeleton beside it. This deliberately leaves the frontend unable to build until the subsequent routing-and-feature reconstruction phase, but it preserves the old implementation without import rewriting or partial migration ambiguity.

## File mapping

`src/auth/AuthProvider.tsx` becomes `src/temp/auth/AuthProvider.tsx`; the same path-preserving rule applies to every TypeScript and TSX file. `src/main.tsx` becomes `src/temp/main.tsx` and `src/App.tsx` becomes `src/temp/App.tsx`.

The target skeleton contains only `.gitkeep` files. `contract/api-contract.ts` is intentionally deferred: creating it would be new TypeScript source rather than an empty directory marker, and its contract belongs to the next implementation phase.

## Constraints and acceptance criteria

- `src/temp/` contains every pre-existing `*.ts` and `*.tsx` source file with unchanged contents.
- No `*.ts` or `*.tsx` files remain outside `src/temp/`, except TypeScript configuration files outside `src/` such as `vite.config.ts`.
- Every requested target folder and feature module exists and contains `.gitkeep`.
- `index.css` stays at `src/index.css`.
- A failed frontend build after this change is expected and must be reported rather than treated as a regression.
