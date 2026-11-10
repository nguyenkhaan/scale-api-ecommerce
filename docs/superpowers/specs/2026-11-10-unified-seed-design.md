# Unified development seed design

## Intent

Replace the separate admin and catalog seed commands with one development-only seed command. It must load a versioned, human-readable fixture from `cmd/seed/data.json`, create the two demo accounts, and populate every current application table with a small, internally consistent data set for manual testing.

## Goals

- Provide exactly two users: `admin@gmail.com` (`admin`) and `customer@gmail.com` (`customer`).
- Source the demo password `cloudian` from fixture data, hash it with the application's existing bcrypt configuration before inserting it, and never write plaintext passwords to PostgreSQL.
- Insert three or four records for each remaining application table in migrations `00001` through `00008`: sessions, login throttles, categories, products, SKUs, inventories, inventory adjustments, flash-sale campaigns/items/customer usage, carts, and cart items.
- Preserve foreign keys, checks, counter relationships, and the single-active-cart constraint.
- Be repeatable: the command upserts stable fixture IDs and restores fixture values rather than adding duplicates.
- Replace `cmd/seed-admin`, `cmd/seed-catalog`, and their dedicated bootstrap implementations with one `cmd/seed` entrypoint.

## Non-goals

- Seed Goose's internal migration-version table.
- Create usable pre-authenticated browser sessions. Seeded sessions will be revoked; normal login creates a real session.
- Populate later order, reservation, payment, outbox, or notification tables before their migrations exist.
- Change the production registration password policy; the eight-character password is limited to these development fixtures.

## Constraints

- The command runs only with `ENV` set to `dev`, `development`, or `test`, and requires an explicit `ALLOW_SEED=true` confirmation.
- All writes occur in one PostgreSQL transaction. An error rolls back the complete fixture.
- Fixture records use stable UUIDs. The seed order follows foreign keys: users/categories/products/SKUs/inventory, flash-sale data, carts/items, then operational rows where applicable.
- Seeded inventory counters, adjustment history, flash-sale quota counters, and customer usage counters must agree with the fixture values.

## Approaches considered

1. **Strict all-table fixture — selected.** Seed 3–4 rows for every application table, with two users as the stated exception. Sessions are revoked and login throttle rows use fixture-only emails, so they do not interfere with demo login.
2. Business-table fixture only. Cleaner operational state but does not meet the requested every-table coverage.
3. Separate seed commands by domain. Easier to run selectively, but retains the fragmented workflow being replaced.

## Implementation outline

- Add `app/backend/cmd/seed/data.json` and `app/backend/cmd/seed/script.go` with a `main.go` entrypoint.
- Parse and validate the fixture before opening the write transaction.
- Bcrypt-hash each fixture password in `script.go`; persist only `password_hash`.
- Use explicit `INSERT ... ON CONFLICT ... DO UPDATE` behavior keyed by fixture primary/unique keys, never destructive resets.
- Remove old seed commands and bootstrap seed code/tests, replacing them with focused tests for fixture validation, password hashing, idempotency, row counts, and relational invariants.
- Document `go run ./cmd/seed` and its environment guard in the repository README.
