# Implementation Plan: PostgreSQL Migration

## Overview

Migrate the MyFi backend from SQLite to PostgreSQL as the sole database engine. The implementation proceeds in layers: dependencies and infrastructure first, then DDL conversion, then DML query conversion across all four query files, then test infrastructure, and finally property-based tests for migration correctness. Each step builds on the previous and ends with wiring things together.

## Tasks

- [x] 1. Update dependencies and infrastructure
  - [x] 1.1 Update `go.mod`: add `github.com/jackc/pgx/v5`, `github.com/testcontainers/testcontainers-go`, `github.com/testcontainers/testcontainers-go/modules/postgres`; remove `github.com/mattn/go-sqlite3`; run `go mod tidy`
    - _Requirements: 8.1, 8.2, 8.3_
  - [x] 1.2 Create `myfi/docker-compose.yml` with PostgreSQL 16 service, named volume `pgdata`, port 5432, health check using `pg_isready`, environment variables for db name/user/password
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_
  - [x] 1.3 Update `myfi/.env`: set `DATABASE_URL=postgres://myfi:myfi_dev@localhost:5432/myfi?sslmode=disable`, remove `DB_TYPE` and `DB_PATH` references
    - _Requirements: 4.5_

- [x] 2. Convert `database.go`: driver, InitDB, and all 14 DDL statements
  - [x] 2.1 Replace SQLite driver with pgx: change import from `_ "github.com/mattn/go-sqlite3"` to `_ "github.com/jackc/pgx/v5/stdlib"`, rewrite `InitDB()` to read `DATABASE_URL`, open with `"pgx"` driver, remove `DB_TYPE`/`DB_PATH` branching, add connection pool settings (`SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5*time.Minute)`), return error if `DATABASE_URL` is empty
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 4.1, 4.2, 4.3, 4.4_
  - [x] 2.2 Convert all 14 `CREATE TABLE` constants to PostgreSQL syntax: `INTEGER PRIMARY KEY AUTOINCREMENT` → `SERIAL PRIMARY KEY`, `REAL` → `DOUBLE PRECISION`, `DATETIME` → `TIMESTAMP WITH TIME ZONE`, `INTEGER DEFAULT 0` (booleans) → `BOOLEAN DEFAULT FALSE`, `INTEGER DEFAULT 1` → `BOOLEAN DEFAULT TRUE`, `DEFAULT CURRENT_TIMESTAMP` → `DEFAULT NOW()`. Preserve all CHECK, FOREIGN KEY, UNIQUE constraints and CREATE INDEX statements. Tables: `users`, `assets`, `transactions`, `savings_accounts`, `nav_snapshots`, `pattern_observations`, `alerts`, `watchlists`, `watchlist_symbols`, `filter_presets`, `recommendation_audit_log`, `financial_goals`, `stock_sector_mapping`, `cache_entries`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8_
  - [x] 2.3 Export migration list via `allMigrations()` function that returns the same slice used by `runMigrations()`, so tests can reuse production DDL
    - _Requirements: 6.6_

- [x] 3. Checkpoint - Verify database layer compiles
  - Ensure `go build ./...` passes in `myfi/backend/`. Ask the user if questions arise.

- [x] 4. Convert `asset_registry.go` queries to PostgreSQL
  - [x] 4.1 Convert all `?` placeholders to `$1`, `$2`, ... in `AddAsset`, `UpdateAsset`, `DeleteAsset`, `GetAsset`, `GetAssetsByUser`, `computeNAV`, and the cascade delete in `DeleteAsset`
    - _Requirements: 3.1, 3.5_
  - [x] 4.2 Remove all `time.Format(time.RFC3339)` calls in `AddAsset` and `UpdateAsset` — pass `time.Time` directly for `acquisition_date`, `created_at`, `updated_at`
    - _Requirements: 3.2, 7.1, 7.4_
  - [x] 4.3 Remove all `time.Parse(time.RFC3339, ...)` calls in `GetAsset` and `GetAssetsByUser` — scan `time.Time` directly for `acqDate`, `createdAt`, `updatedAt` (change scan variables from `string` to `time.Time`)
    - _Requirements: 3.3, 7.2, 7.3_
  - [x] 4.4 Replace `ExecContext` + `LastInsertId()` with `QueryRowContext` + `RETURNING id` + `Scan(&id)` in `AddAsset`
    - _Requirements: 9.6_

- [x] 5. Convert `transaction_ledger.go` queries to PostgreSQL
  - [x] 5.1 Convert all `?` placeholders to `$1`, `$2`, ... in `RecordTransaction`, `GetTransactionsByUser`, `GetTransactionsBySymbol`
    - _Requirements: 3.1, 3.5_
  - [x] 5.2 Remove `tx.TransactionDate.Format(time.RFC3339)` in `RecordTransaction` — pass `time.Time` directly
    - _Requirements: 3.2, 7.1, 7.4_
  - [x] 5.3 In `scanTransactions`, change `txDate` and `createdAt` from `string` to `time.Time`, remove `time.Parse(time.RFC3339, ...)` calls — scan directly
    - _Requirements: 3.3, 7.2, 7.3_
  - [x] 5.4 Replace `ExecContext` + `LastInsertId()` with `QueryRowContext` + `RETURNING id` + `Scan(&id)` in `RecordTransaction`
    - _Requirements: 9.6_

- [x] 6. Convert `savings_tracker.go` queries to PostgreSQL
  - [x] 6.1 Convert all `?` placeholders to `$1`, `$2`, ... in `AddSavingsAccount`, `GetSavingsAccount`, `GetSavingsAccountsByUser`, `UpdateMaturityStatus`, `RefreshAllMaturityStatuses`, `DeleteSavingsAccount`
    - _Requirements: 3.1, 3.5_
  - [x] 6.2 Remove date string formatting: remove `account.StartDate.Format("2006-01-02")` and `maturityDate.Format("2006-01-02")` in `AddSavingsAccount` — pass `time.Time` directly. Remove `now.Format("2006-01-02")` in `RefreshAllMaturityStatuses` — pass `time.Time` directly
    - This is the core fix for the savings_tracker date bug that motivated the migration
    - _Requirements: 3.2, 7.1, 7.4, 7.5_
  - [x] 6.3 Remove date string parsing: in `GetSavingsAccount` and `GetSavingsAccountsByUser`, change `startDate`, `createdAt` from `string` to `time.Time`, change `maturityDate` from `sql.NullString` to `sql.NullTime`, remove all `time.Parse("2006-01-02", ...)` and `time.Parse(time.RFC3339, ...)` calls — scan directly
    - _Requirements: 3.3, 7.2, 7.3, 7.5_
  - [x] 6.4 Replace boolean handling: remove `boolToInt()` calls in `AddSavingsAccount` (use `false` directly) and `UpdateMaturityStatus` (use `true` directly). In `RefreshAllMaturityStatuses`, change `is_matured = 1` to `is_matured = TRUE` and `is_matured = 0` to `is_matured = FALSE`. Change scan variable `isMatured` from `int` to `bool`, remove `isMatured != 0` conversion
    - _Requirements: 3.4, 3.6, 3.7_
  - [x] 6.5 Replace `ExecContext` + `LastInsertId()` with `QueryRowContext` + `RETURNING id` + `Scan(&id)` in `AddSavingsAccount`
    - _Requirements: 9.6_
  - [x] 6.6 Delete the `boolToInt()` helper function from `savings_tracker.go`
    - _Requirements: 3.7_

- [x] 7. Verify no remaining SQLite patterns in query files
  - Grep all `.go` files in `myfi/backend/` for `?` placeholders in SQL strings, `boolToInt`, `time.Format`, `time.Parse` used for DB operations, `LastInsertId`, and `go-sqlite3` imports. Fix any remaining occurrences. `portfolio_engine.go` has no direct SQL queries (it delegates to registry and ledger), so no changes needed there.
  - _Requirements: 3.1, 3.5, 1.2, 1.3_

- [x] 8. Checkpoint - Verify all source files compile
  - Ensure `go build ./...` passes in `myfi/backend/`. Ask the user if questions arise.

- [x] 9. Create shared test helper and update all test files
  - [x] 9.1 Create `myfi/backend/testhelper_test.go` with `setupPostgresTestDB(t *testing.T) *sql.DB` function: start PostgreSQL 16 container via `testcontainers-go`, wait for readiness, run all production DDL via `allMigrations()`, seed test user, register `t.Cleanup` for container termination and db close. Add required imports: `_ "github.com/jackc/pgx/v5/stdlib"`, `testcontainers-go/modules/postgres`
    - _Requirements: 6.1, 6.2, 6.3, 6.6, 6.8_
  - [x] 9.2 Update `asset_registry_test.go`: replace `setupTestDB` with `setupPostgresTestDB`, remove `_ "github.com/mattn/go-sqlite3"` import, remove inline SQLite DDL, remove `PRAGMA foreign_keys = ON`. Keep `contains()` and `searchSubstring()` helpers. Keep `TestMain`
    - _Requirements: 6.3, 6.4, 6.5, 6.7_
  - [x] 9.3 Update `savings_tracker_test.go`: replace `setupSavingsTestDB` with `setupPostgresTestDB`, remove `_ "github.com/mattn/go-sqlite3"` import, remove inline SQLite DDL, remove `PRAGMA foreign_keys = ON`. Update `newTestSavingsTracker` to use `setupPostgresTestDB`
    - _Requirements: 6.3, 6.4, 6.5, 6.7_
  - [x] 9.4 Update `transaction_ledger_test.go`: it uses `setupTestDB` from `asset_registry_test.go` — verify it works with the new shared helper. Update any inline SQL that uses `?` placeholders or `time.Format` to match PostgreSQL syntax (e.g., the cascade delete test inserts transactions directly)
    - _Requirements: 6.7, 9.5_
  - [x] 9.5 Update `portfolio_engine_test.go`: it uses `setupTestDB` via `newTestPortfolioEngine` — verify it works with the new shared helper. No inline SQL to change
    - _Requirements: 6.7, 9.1_
  - [x] 9.6 Update `portfolio_engine_property_test.go`: replace `setupPropDB` with `setupPostgresTestDB`, remove `_ "github.com/mattn/go-sqlite3"` import, remove inline SQLite DDL and `PRAGMA`. Update `newPropEngine` to use `setupPostgresTestDB(t)` (note: `t` is `*rapid.T` which embeds `*testing.T`)
    - _Requirements: 6.5, 6.7, 9.2_
  - [x] 9.7 Delete `myfi/backend/debug_nav_test.go` — temporary debug file that used raw SQLite queries with `?` placeholders; the date bug it was debugging is resolved by this migration
    - _Requirements: 6.7_

- [x] 10. Checkpoint - Verify all tests compile
  - Ensure `go vet ./...` passes in `myfi/backend/`. Ask the user if questions arise.

- [ ] 11. Write property-based tests for migration correctness
  - [ ]* 11.1 Write property test for date/time round-trip preservation
    - **Property 1: Date/Time Round-Trip Preservation**
    - Generate random `time.Time` values, insert a savings account with that start date, retrieve it, assert year/month/day match. Also test `TIMESTAMPTZ` columns via assets (`acquisition_date`)
    - Use `pgregory.net/rapid` generators, `setupPostgresTestDB`, tag with `// Feature: postgresql-migration, Property 1`
    - **Validates: Requirements 3.2, 3.3, 7.1, 7.2, 7.5**

  - [ ]* 11.2 Write property test for boolean round-trip preservation
    - **Property 2: Boolean Round-Trip Preservation**
    - Generate random `bool` values, insert a savings account with `is_matured` set to that value, retrieve it, assert the boolean matches without integer conversion
    - Use `pgregory.net/rapid` generators, `setupPostgresTestDB`, tag with `// Feature: postgresql-migration, Property 2`
    - **Validates: Requirements 3.4, 3.6**

  - [ ]* 11.3 Write property test for constraint preservation
    - **Property 3: Constraint Preservation**
    - Generate invalid inputs (e.g., `asset_type` not in CHECK list, NULL for NOT NULL columns, invalid FK references), attempt insert, assert error is returned
    - Use `pgregory.net/rapid` generators, `setupPostgresTestDB`, tag with `// Feature: postgresql-migration, Property 3`
    - **Validates: Requirements 2.7**

  - [ ]* 11.4 Write property test for INSERT RETURNING ID correctness
    - **Property 4: INSERT RETURNING ID Correctness**
    - Generate N random assets (1-20), insert them sequentially via `AddAsset`, collect all returned IDs, assert all are positive and all are distinct
    - Use `pgregory.net/rapid` generators, `setupPostgresTestDB`, tag with `// Feature: postgresql-migration, Property 4`
    - **Validates: Requirements 9.6**

- [x] 12. Final checkpoint - Ensure all tests pass
  - Run `go test ./... -v -count=1` in `myfi/backend/` (requires Docker running for testcontainers). Ensure all existing unit tests and property tests pass. Ask the user if questions arise.
  - _Requirements: 8.4, 8.5, 9.1, 9.2, 9.3, 9.4, 9.5_

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties from the design document
- Tests require Docker to be running (for testcontainers-go)
- `portfolio_engine.go` has no direct SQL — it delegates to `AssetRegistry` and `TransactionLedger`, so no query conversion needed there