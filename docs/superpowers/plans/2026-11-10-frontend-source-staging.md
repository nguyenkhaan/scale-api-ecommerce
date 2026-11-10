# Frontend Source Staging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Archive the existing frontend TypeScript source tree and create the requested empty module-first `src/` skeleton.

**Architecture:** Every existing `src/**/*.ts` and `src/**/*.tsx` file moves beneath `src/temp/` without content or relative-subpath changes. The new top-level and feature-module folders are empty except for `.gitkeep`; the application is intentionally left without a Vite entrypoint until routing is rebuilt.

**Tech Stack:** React, TypeScript, Vite, Bun.

**Spec:** `docs/superpowers/specs/2026-11-10-frontend-source-staging-design.md`

## Global Constraints

- Move all current `*.ts` and `*.tsx` files under `app/frontend/src/` to `app/frontend/src/temp/`, preserving their relative paths.
- Preserve non-TypeScript source files and all configuration files in their current locations.
- Create only `.gitkeep` marker files in the new skeleton; do not create application code, routes, TanStack Query setup, contracts, hooks, or services.
- The frontend is expected not to build because `src/main.tsx` moves to `src/temp/main.tsx`.

## Review Focus

- Every old TypeScript/TSX source file must be archived, including tests and the Vite entrypoint.
- Relative paths inside the archive must match the prior `src/` tree exactly.
- `src/index.css` must remain outside the archive.
- Every requested shared folder and every requested feature module must have a `.gitkeep`.
- No accidental TypeScript implementation file may be added to the new skeleton.

---

### Task 1: Archive current source and scaffold the new module tree

**Files:**
- Move: every current `app/frontend/src/**/*.ts` and `app/frontend/src/**/*.tsx` file to the equivalent `app/frontend/src/temp/**` path.
- Create: `.gitkeep` files in `api`, `assets`, `components/common`, `components/layout`, `context`, `contract`, `features`, `features/auth`, `features/product`, `features/flashsale`, `features/payment`, `features/bill`, `features/cart`, `hooks`, `lib`, `pages`, `routes`, `store`, and `store/slices`.
- Preserve: `app/frontend/src/index.css` and `app/frontend/src/.gitkeep`.

**Interfaces:**
- Consumes: the source-file inventory captured before this task.
- Produces: an archived `src/temp/` tree and an empty target skeleton for the next routing rebuild.

**Acceptance Criteria:**
- The TypeScript inventory is exactly reproduced below `src/temp/`.
- No `.ts` or `.tsx` file remains below `src/` outside `src/temp/`.
- All requested directories contain `.gitkeep`.
- `index.css` remains at `src/index.css`.

- [ ] **Step 1: Move the TypeScript/TSX tree preserving relative paths**

Use a path-preserving move for each file under `app/frontend/src/`. Do not alter file contents or import paths.

- [ ] **Step 2: Create the empty target structure**

Add `.gitkeep` files for every requested shared folder and feature module, including the parent `store` and `features` directories.

- [ ] **Step 3: Verify the staging boundary**

Run: `find app/frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) ! -path 'app/frontend/src/temp/*' -print`

Expected: no output. Verify `index.css` and every marker path exists.

- [ ] **Step 4: Record the intentional build boundary**

Run: `cd app/frontend && bun run build`

Expected: failure because `src/main.tsx` was intentionally archived; capture the error as evidence rather than treating it as a defect.

- [ ] **Step 5: Commit source staging**

```bash
git add app/frontend/src docs/superpowers/specs/2026-11-10-frontend-source-staging-design.md docs/superpowers/plans/2026-11-10-frontend-source-staging.md
git commit -m "refactor: stage frontend source tree"
```
