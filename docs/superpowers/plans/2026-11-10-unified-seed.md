# Unified Development Seed Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the separate admin and catalog seed commands with one guarded seed command that loads a complete, relationally valid development fixture from JSON.

**Architecture:** `cmd/seed/script.go` owns the command, fixture parsing, validation, password hashing, and one PostgreSQL transaction. `data.json` supplies two login users and four stable-ID records for every other table created by migrations 00001–00008. Inserts use stable primary keys and conflict updates, so rerunning the command restores the fixture instead of duplicating it.

**Tech Stack:** Go 1.27, GORM/pgx, PostgreSQL, bcrypt, `embed`, Goose-managed schema.

**Spec:** `docs/superpowers/specs/2026-11-10-unified-seed-design.md`

## Global Constraints

- The command is restricted to `ENV=dev`, `development`, or `test` and requires `ALLOW_SEED=true`.
- `data.json` contains plaintext fixture passwords only; `script.go` must bcrypt-hash them before persistence.
- Do not change the normal registration password policy to accommodate the demo password `cloudian`.
- Seed all current application tables except Goose's migration bookkeeping table: two `users`; four rows each for sessions, login throttles, categories, products, SKUs, inventories, inventory adjustments, flash-sale campaigns/items/customer usage, carts, and cart items.
- One seed transaction must either apply the full fixture or roll back.
- Use stable UUIDs and preserve all foreign keys, checks, counters, and the one-active-cart-per-user rule.

## Review Focus

- Fixture password must never appear in PostgreSQL or logs in plaintext; test bcrypt verification and absent plaintext persistence path.
- `cloudian` is eight characters, so seed validation must not reuse the 12-character production/admin-bootstrap validation.
- Four seeded sessions must be revoked and their token/CSRF hashes must be exactly 32 bytes, so they cannot become accidental authenticated browser sessions.
- The flash-sale price, quota/usage counters, inventory/adjustment ledger, and flash-sale cart items must all satisfy database checks and foreign keys.
- A failed row insert must roll back prior rows; rerunning the command must retain fixture row counts rather than adding duplicates.

---

### Task 1: Define and validate the embedded fixture

**Files:**
- Create: `app/backend/cmd/seed/data.json`
- Create: `app/backend/cmd/seed/script.go`
- Create: `app/backend/cmd/seed/script_test.go`

**Interfaces:**
- Produces: `func loadFixture() (fixture, error)` in `cmd/seed/script.go`.
- Produces: `func validateFixture(fixture) error` in `cmd/seed/script.go`.
- Consumes: the PostgreSQL tables from migrations `00001`–`00008`.

**Acceptance Criteria:**
- The fixture has exactly two users, `admin@gmail.com`/`admin` and `customer@gmail.com`/`customer`, each with password source `cloudian`.
- Every other application-table collection contains exactly four records with explicit stable UUIDs and valid cross-references.
- The fixture has one active cart for the customer; its other carts are archived.
- Every inventory row matches its adjustment ledger; each flash-sale item has a lower sale price than its SKU's regular price and valid counters.

- [ ] **Step 1: Write failing fixture-shape tests in `app/backend/cmd/seed/script_test.go`**

Assert two users, four records for each remaining table, the specified login identities, unique fixture IDs, and all referenced IDs existing in their parent fixture collections.

- [ ] **Step 2: Run the fixture-shape test to verify it fails**

Run: `cd app/backend && go test ./cmd/seed -run TestLoadFixture -count=1`

Expected: FAIL because `cmd/seed` and `loadFixture` do not exist.

- [ ] **Step 3: Add `data.json` and implement `loadFixture() (fixture, error)` in `app/backend/cmd/seed/script.go`**

Embed `data.json`, decode it into focused typed structs, and reject malformed JSON or missing required fields.

- [ ] **Step 4: Implement `validateFixture(f fixture) error` in `app/backend/cmd/seed/script.go`**

Validate record counts, explicit UUID format, emails/roles, FK references, inventory and quota counter relationships, cart price-source combinations, and one active cart per user before any database write.

- [ ] **Step 5: Run fixture tests to verify they pass**

Run: `cd app/backend && go test ./cmd/seed -run 'Test(LoadFixture|ValidateFixture)' -count=1`

Expected: PASS.

- [ ] **Step 6: Commit fixture foundation**

```bash
git add app/backend/cmd/seed/data.json app/backend/cmd/seed/script.go app/backend/cmd/seed/script_test.go
git commit -m "feat: add validated unified seed fixture"
```

### Task 2: Add guarded, transactional fixture persistence

**Files:**
- Modify: `app/backend/cmd/seed/script.go`
- Modify: `app/backend/cmd/seed/script_test.go`

**Interfaces:**
- Consumes: `loadFixture() (fixture, error)` and `validateFixture(fixture) error` from Task 1.
- Produces: `func canSeed(environment, confirmation string) bool` and `func seed(ctx context.Context, db *gorm.DB, f fixture) error`.

**Acceptance Criteria:**
- `canSeed` permits only development/test environments with exact `ALLOW_SEED=true`.
- `seed` bcrypt-hashes the two fixture passwords using `auth.DefaultPasswordCost` and persists only hashes.
- `seed` wraps all writes in one transaction and executes parent tables before children.
- Re-running the seed preserves two users and four rows for each other fixture collection, while refreshing fixture-owned fields.
- Seeded sessions have 32-byte token/CSRF hashes and a non-null `revoked_at`; no plaintext token or password is inserted.

- [ ] **Step 1: Write failing unit tests for guard and password hashing**

Cover rejected production/empty-confirmation environments, accepted `dev`/`development`/`test`, bcrypt `CompareHashAndPassword` for `cloudian`, and a negative assertion that fixture password text is not a stored field value.

- [ ] **Step 2: Run the unit tests to verify they fail**

Run: `cd app/backend && go test ./cmd/seed -run 'Test(CanSeed|SeedPassword)' -count=1`

Expected: FAIL because the guard and persistence helpers do not exist.

- [ ] **Step 3: Implement `canSeed` and `seed` in `app/backend/cmd/seed/script.go`**

Use GORM's transaction API. Insert/upsert users, categories, products, SKUs, inventories, adjustments, campaigns/items/usage, carts/items, revoked sessions, and throttles in dependency order. Use fixture UUIDs as conflict keys; bcrypt each password before the user upsert.

- [ ] **Step 4: Add PostgreSQL integration coverage in `app/backend/cmd/seed/script_test.go`**

When `TEST_DATABASE_URL` is set to a dedicated database, migrate it, seed twice, and assert the intended row counts, bcrypt verification, every FK/counter relationship, revoked sessions, and no duplicate fixture rows. Skip clearly when the variable is absent.

- [ ] **Step 5: Run command and integration test gates**

Run: `cd app/backend && go test ./cmd/seed -count=1 && go test ./... -count=1`

Expected: PASS; the dedicated PostgreSQL case is skipped only without `TEST_DATABASE_URL`.

- [ ] **Step 6: Commit guarded persistence**

```bash
git add app/backend/cmd/seed
git commit -m "feat: seed complete development fixture"
```

### Task 3: Replace legacy commands and document the workflow

**Files:**
- Delete: `app/backend/cmd/seed-admin/main.go`
- Delete: `app/backend/cmd/seed-catalog/main.go`
- Delete: `app/backend/internal/bootstrap/admin_seed.go`
- Delete: `app/backend/internal/bootstrap/catalog_seed.go`
- Delete: `app/backend/test/admin_seed_test.go`
- Delete: `app/backend/test/catalog_seed_test.go`
- Modify: `README.md`
- Modify: `docs/spec.md`

**Interfaces:**
- Consumes: `cmd/seed` from Tasks 1–2.
- Produces: documented command `ALLOW_SEED=true go run ./cmd/seed`.

**Acceptance Criteria:**
- No legacy seed command or bootstrap implementation remains.
- Runbooks state that `DATABASE_URL` must point to a disposable development/test database, migrations run before seed, and the command is idempotent but intentionally refreshes fixture values.
- Runbooks list the two demo accounts without putting their plaintext password in a production configuration example.

- [ ] **Step 1: Write/update a failing static command-path test**

In `app/backend/cmd/seed/script_test.go`, assert fixture code is the only seed command path used by documentation/code references, excluding historical documents.

- [ ] **Step 2: Delete legacy seed files and add the unified command entrypoint in `app/backend/cmd/seed/script.go`**

`main()` loads config and fixture, enforces `ALLOW_SEED`, opens the configured database, calls `seed`, and emits no password or connection-string data.

- [ ] **Step 3: Update `README.md` and `docs/spec.md`**

Replace `seed-admin`/`seed-catalog` usage with migration-first unified seed instructions and the two demo account emails.

- [ ] **Step 4: Run full verification**

Run: `cd app/backend && gofmt -w cmd/seed && go test ./... && go vet ./... && go build ./...`

Run: `cd app/frontend && bun run typecheck && bun run lint && bun run test && bun run build`

Expected: all commands exit 0; PostgreSQL-only tests may report a documented skip when `TEST_DATABASE_URL` is unset.

- [ ] **Step 5: Commit command replacement and documentation**

```bash
git add README.md docs/spec.md app/backend
git commit -m "refactor: unify development database seed"
```

## Self-review

- Fixture data/counts, bcrypt handling, environment guard, transaction, idempotency, operational-row behavior, legacy removal, and runbook changes are each covered by Tasks 1–3.
- Task 1 produces the fixture contract consumed by Task 2; Task 3 only removes callers after the replacement command exists.
- Tests cover malformed fixture input, unsafe environment, bcrypt verification, relational/counter constraints, repeat seeding, and missing PostgreSQL test configuration.
- No migration schema or production registration policy changes are planned.
