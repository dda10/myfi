# Requirements Document

## Introduction

Migrate the MyFi backend from SQLite to PostgreSQL as the sole database engine. This involves replacing the SQLite driver, converting all DDL and DML statements to PostgreSQL syntax, setting up Docker Compose for local development, and updating all tests to run against PostgreSQL. The migration also resolves existing date-parsing issues by leveraging PostgreSQL's native DATE/TIMESTAMP types.

## Glossary

- **Backend**: The MyFi Go application using the Gin framework, located in `myfi/backend/`.
- **Database_Layer**: The `database.go` file containing `InitDB`, `runMigrations`, and all `CREATE TABLE` DDL constants.
- **Migration_Statements**: The 14 `CREATE TABLE IF NOT EXISTS` constants in `database.go` that define the schema.
- **Query_Layer**: All Go source files (`asset_registry.go`, `transaction_ledger.go`, `savings_tracker.go`, `portfolio_engine.go`) that execute raw SQL via `database/sql`.
- **Test_Suite**: All `*_test.go` files that create in-memory SQLite databases for unit and property-based testing.
- **Docker_Compose_File**: The `docker-compose.yml` file that defines the local PostgreSQL service.
- **Env_File**: The `.env` file at `myfi/.env` containing configuration variables including `DATABASE_URL`.
- **Placeholder_Syntax**: The parameterized query format; SQLite uses `?` positional placeholders, PostgreSQL uses `$1`, `$2`, etc.
- **testcontainers-go**: A Go library that provisions throwaway Docker containers for integration testing.

## Requirements

### Requirement 1: Replace SQLite Driver with PostgreSQL Driver

**User Story:** As a developer, I want the backend to use a PostgreSQL driver instead of SQLite, so that the application connects to PostgreSQL exclusively.

#### Acceptance Criteria

1. THE Backend SHALL use `github.com/jackc/pgx/v5/stdlib` or `github.com/lib/pq` as the sole database driver.
2. THE Backend SHALL remove the `github.com/mattn/go-sqlite3` dependency from `go.mod`.
3. THE Backend SHALL remove all `_ "github.com/mattn/go-sqlite3"` import statements from all Go source and test files.
4. WHEN the Backend starts, THE Database_Layer SHALL open a connection using the `pgx` or `postgres` driver name registered with `database/sql`.
5. THE Backend SHALL retain the `database/sql` interface for all database operations.

### Requirement 2: Convert DDL to PostgreSQL Syntax

**User Story:** As a developer, I want all CREATE TABLE statements to use PostgreSQL-native syntax, so that the schema leverages PostgreSQL data types and features.

#### Acceptance Criteria

1. THE Migration_Statements SHALL use `SERIAL` or `BIGSERIAL` as the primary key type instead of `INTEGER PRIMARY KEY AUTOINCREMENT`.
2. THE Migration_Statements SHALL use `BOOLEAN` with `DEFAULT FALSE` instead of `INTEGER DEFAULT 0` for boolean columns (`is_matured`, `viewed`, `expired`).
3. THE Migration_Statements SHALL use `TIMESTAMP WITH TIME ZONE` instead of `DATETIME` for timestamp columns (`created_at`, `updated_at`, `detection_date`, `timestamp`, `last_login`, `account_locked_until`, `detection_timestamp`).
4. THE Migration_Statements SHALL use `DATE` for date-only columns (`start_date`, `maturity_date`, `target_date`, `snapshot_date`).
5. THE Migration_Statements SHALL use `DOUBLE PRECISION` instead of `REAL` for floating-point columns.
6. THE Migration_Statements SHALL use `DEFAULT NOW()` instead of `DEFAULT CURRENT_TIMESTAMP` for timestamp defaults.
7. THE Migration_Statements SHALL preserve all existing `CHECK` constraints, `FOREIGN KEY` constraints, `UNIQUE` constraints, and `INDEX` definitions.
8. THE Migration_Statements SHALL convert all 14 tables: `users`, `assets`, `transactions`, `savings_accounts`, `nav_snapshots`, `pattern_observations`, `alerts`, `watchlists`, `watchlist_symbols`, `filter_presets`, `recommendation_audit_log`, `financial_goals`, `stock_sector_mapping`, `cache_entries`.

### Requirement 3: Convert All SQL Queries to PostgreSQL Syntax

**User Story:** As a developer, I want all raw SQL queries to use PostgreSQL parameter syntax and functions, so that queries execute correctly against PostgreSQL.

#### Acceptance Criteria

1. THE Query_Layer SHALL use numbered placeholders (`$1`, `$2`, `$3`, ...) instead of `?` in all parameterized queries.
2. THE Query_Layer SHALL pass `time.Time` values directly to PostgreSQL instead of formatting dates as strings via `time.Format()`.
3. THE Query_Layer SHALL scan `time.Time` values directly from PostgreSQL result sets instead of scanning strings and parsing with `time.Parse()`.
4. THE Query_Layer SHALL pass `bool` values directly to PostgreSQL instead of converting via `boolToInt()`.
5. THE Query_Layer SHALL update queries in `asset_registry.go`, `transaction_ledger.go`, `savings_tracker.go`, and all other files containing raw SQL.
6. WHEN the Query_Layer retrieves a row with a `BOOLEAN` column, THE Query_Layer SHALL scan the value into a Go `bool` variable directly.
7. THE Query_Layer SHALL remove the `boolToInt()` helper function after all usages are replaced with direct `bool` parameters.

### Requirement 4: Configure PostgreSQL Connection

**User Story:** As a developer, I want the backend to read a PostgreSQL connection string from environment variables, so that the database connection is configurable per environment.

#### Acceptance Criteria

1. THE Database_Layer SHALL read the `DATABASE_URL` environment variable for the PostgreSQL connection string.
2. THE Database_Layer SHALL remove the `DB_TYPE` and `DB_PATH` environment variable handling.
3. THE Database_Layer SHALL remove all SQLite-specific branching logic from `InitDB()`.
4. IF the `DATABASE_URL` environment variable is empty, THEN THE Database_Layer SHALL return a descriptive error message.
5. THE Env_File SHALL contain a `DATABASE_URL` value in the format `postgres://user:password@localhost:5432/myfi?sslmode=disable`.

### Requirement 5: Set Up Docker Compose for Local PostgreSQL

**User Story:** As a developer, I want a Docker Compose configuration for local PostgreSQL, so that I can run the database locally without manual installation.

#### Acceptance Criteria

1. THE Docker_Compose_File SHALL define a PostgreSQL 16 service with a named volume for data persistence.
2. THE Docker_Compose_File SHALL expose PostgreSQL on port `5432`.
3. THE Docker_Compose_File SHALL configure the database name, username, and password via environment variables.
4. THE Docker_Compose_File SHALL include a health check that verifies PostgreSQL readiness using `pg_isready`.
5. THE Docker_Compose_File SHALL be located at `myfi/docker-compose.yml`.

### Requirement 6: Migrate Test Infrastructure to PostgreSQL

**User Story:** As a developer, I want all tests to run against PostgreSQL instead of in-memory SQLite, so that tests validate behavior against the production database engine.

#### Acceptance Criteria

1. THE Test_Suite SHALL use `testcontainers-go` to provision a PostgreSQL container for each test run.
2. THE Test_Suite SHALL provide a shared test helper function that starts a PostgreSQL container, runs migrations, inserts seed data, and returns a `*sql.DB` connection.
3. THE Test_Suite SHALL remove all `sql.Open("sqlite3", ":memory:")` calls from test setup functions.
4. THE Test_Suite SHALL remove all `PRAGMA foreign_keys = ON` statements from test setup functions.
5. THE Test_Suite SHALL remove all SQLite-specific DDL (e.g., `INTEGER PRIMARY KEY AUTOINCREMENT`) from inline test schemas.
6. THE Test_Suite SHALL use the same PostgreSQL DDL as the production Migration_Statements for test schema creation.
7. THE Test_Suite SHALL update test files: `asset_registry_test.go`, `savings_tracker_test.go`, `transaction_ledger_test.go`, `portfolio_engine_test.go`, `portfolio_engine_property_test.go`, `debug_nav_test.go`.
8. WHEN a test completes, THE Test_Suite SHALL terminate the PostgreSQL container to free resources.

### Requirement 7: Resolve Date and Timestamp Handling

**User Story:** As a developer, I want dates and timestamps to be stored and retrieved using PostgreSQL native types, so that date-parsing issues caused by SQLite string storage are eliminated.

#### Acceptance Criteria

1. THE Query_Layer SHALL store `time.Time` values directly into `DATE` and `TIMESTAMP WITH TIME ZONE` columns without string formatting.
2. THE Query_Layer SHALL retrieve `time.Time` values directly from `DATE` and `TIMESTAMP WITH TIME ZONE` columns without string parsing.
3. THE Query_Layer SHALL remove all `time.Parse("2006-01-02", ...)` and `time.Parse(time.RFC3339, ...)` calls used for scanning database results.
4. THE Query_Layer SHALL remove all `.Format("2006-01-02")` and `.Format(time.RFC3339)` calls used for inserting database values.
5. WHEN the savings_tracker retrieves a savings account, THE Query_Layer SHALL return a `StartDate` with correct year, month, and day values matching the originally stored date.

### Requirement 8: Update Go Module Dependencies

**User Story:** As a developer, I want `go.mod` to reflect the new PostgreSQL dependencies, so that the project builds cleanly with only the required drivers.

#### Acceptance Criteria

1. THE Backend SHALL add `github.com/jackc/pgx/v5` (or `github.com/lib/pq`) as a direct dependency in `go.mod`.
2. THE Backend SHALL add `github.com/testcontainers/testcontainers-go` as a test dependency in `go.mod`.
3. THE Backend SHALL remove `github.com/mattn/go-sqlite3` from `go.mod` and `go.sum`.
4. WHEN `go build ./...` is executed in the `myfi/backend/` directory, THE Backend SHALL compile without errors.
5. WHEN `go vet ./...` is executed in the `myfi/backend/` directory, THE Backend SHALL report no issues related to the database migration.

### Requirement 9: Ensure All Existing Tests Pass

**User Story:** As a developer, I want all existing tests to pass against PostgreSQL, so that I have confidence the migration preserves existing behavior.

#### Acceptance Criteria

1. WHEN `go test ./...` is executed in the `myfi/backend/` directory, THE Test_Suite SHALL pass all tests that previously passed with SQLite.
2. THE Test_Suite SHALL pass all property-based tests in `portfolio_engine_property_test.go` against PostgreSQL.
3. THE Test_Suite SHALL pass all savings tracker tests including compound interest calculations and maturity checks.
4. THE Test_Suite SHALL pass all asset registry tests including CRUD operations and cascade deletes.
5. THE Test_Suite SHALL pass all transaction ledger tests including recording, retrieval, and validation.
6. IF a test uses `result.LastInsertId()`, THEN THE Query_Layer SHALL replace the call with a `RETURNING id` clause in the `INSERT` statement and use `QueryRowContext` with `Scan` instead of `ExecContext`.
