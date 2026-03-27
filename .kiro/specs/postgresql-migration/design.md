# Design Document: PostgreSQL Migration

## Overview

This design covers the complete migration of the MyFi backend from SQLite to PostgreSQL as the sole database engine. The migration touches four layers:

1. **Driver layer** — swap `go-sqlite3` for `pgx/v5/stdlib`, change `sql.Open` call
2. **DDL layer** — convert 14 `CREATE TABLE` constants in `database.go` to PostgreSQL syntax
3. **DML layer** — convert all parameterized queries (`?` → `$N`), remove string-based date formatting/parsing, remove `boolToInt()`, replace `LastInsertId()` with `RETURNING id`
4. **Test layer** — replace in-memory SQLite with `testcontainers-go` PostgreSQL containers

The migration is a clean replacement — no dual-driver support, no SQLite fallback. The `database/sql` interface is retained throughout, so all higher-level code (PortfolioEngine, PriceService, etc.) remains unchanged.

### Key Design Decisions

- **pgx/v5/stdlib over lib/pq**: `pgx` is actively maintained, supports `RETURNING` natively, and handles `time.Time`/`bool` scanning without custom drivers. `lib/pq` is in maintenance mode.
- **testcontainers-go over embedded PostgreSQL**: Real PostgreSQL in Docker ensures test parity with production. testcontainers-go handles lifecycle automatically.
- **No migration tool (goose, migrate)**: The app already uses raw DDL constants executed in `runMigrations()`. We keep this pattern — just convert the SQL. A migration tool can be added later if needed.
- **Clean replacement**: No `DB_TYPE` switch. The `InitDB()` function reads `DATABASE_URL` and opens a `pgx` connection. Period.

## Architecture

### Current Architecture

```mermaid
graph LR
    A[main.go] --> B[InitDB]
    B -->|DB_TYPE=sqlite| C[go-sqlite3]
    B -->|DB_TYPE=postgres| D[lib/pq stub]
    C --> E[database/sql.DB]
    D --> E
    E --> F[asset_registry.go]
    E --> G[transaction_ledger.go]
    E --> H[savings_tracker.go]
    E --> I[portfolio_engine.go]
```

### Target Architecture

```mermaid
graph LR
    A[main.go] --> B[InitDB]
    B -->|DATABASE_URL| C[pgx/v5/stdlib]
    C --> D[database/sql.DB]
    D --> E[asset_registry.go]
    D --> F[transaction_ledger.go]
    D --> G[savings_tracker.go]
    D --> H[portfolio_engine.go]

    subgraph Tests
        I[testcontainers-go] --> J[PostgreSQL 16 container]
        J --> K[*sql.DB for tests]
    end
```

### Connection Management

`InitDB()` simplifies to:

```go
func InitDB() error {
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        return fmt.Errorf("DATABASE_URL environment variable is required")
    }
    var err error
    db, err = sql.Open("pgx", connStr)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    if err = db.Ping(); err != nil {
        return fmt.Errorf("failed to ping database: %w", err)
    }
    // Connection pool defaults from pgx are sensible;
    // optionally tune MaxOpenConns, MaxIdleConns, ConnMaxLifetime
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    return runMigrations()
}
```

Import changes:
```go
// Remove:  _ "github.com/mattn/go-sqlite3"
// Add:     _ "github.com/jackc/pgx/v5/stdlib"
```

## Components and Interfaces

### 1. database.go — DDL Conversion

All 14 `CREATE TABLE` constants are converted. The conversion rules:

| SQLite | PostgreSQL |
|--------|-----------|
| `INTEGER PRIMARY KEY AUTOINCREMENT` | `SERIAL PRIMARY KEY` |
| `REAL` | `DOUBLE PRECISION` |
| `DATETIME` | `TIMESTAMP WITH TIME ZONE` |
| `DATE` | `DATE` (unchanged) |
| `INTEGER DEFAULT 0` (booleans) | `BOOLEAN DEFAULT FALSE` |
| `INTEGER DEFAULT 1` (booleans) | `BOOLEAN DEFAULT TRUE` |
| `DEFAULT CURRENT_TIMESTAMP` | `DEFAULT NOW()` |
| `TEXT PRIMARY KEY` | `TEXT PRIMARY KEY` (unchanged) |

The `runMigrations()` function stays the same — it iterates the constant slice and executes each.

`CREATE INDEX IF NOT EXISTS` statements remain unchanged — PostgreSQL supports this syntax.

#### Example: savings_accounts table

Before (SQLite):
```sql
CREATE TABLE IF NOT EXISTS savings_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    account_name TEXT NOT NULL,
    principal REAL NOT NULL,
    annual_rate REAL NOT NULL,
    compounding_frequency TEXT NOT NULL CHECK (...),
    start_date DATE NOT NULL,
    maturity_date DATE,
    is_matured INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

After (PostgreSQL):
```sql
CREATE TABLE IF NOT EXISTS savings_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    account_name TEXT NOT NULL,
    principal DOUBLE PRECISION NOT NULL,
    annual_rate DOUBLE PRECISION NOT NULL,
    compounding_frequency TEXT NOT NULL CHECK (...),
    start_date DATE NOT NULL,
    maturity_date DATE,
    is_matured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### Full Table Conversion Summary

| # | Table | Key Changes |
|---|-------|-------------|
| 1 | `users` | `SERIAL PK`, `TIMESTAMPTZ` for `created_at`/`last_login`/`account_locked_until` |
| 2 | `assets` | `SERIAL PK`, `DOUBLE PRECISION` for `quantity`/`average_cost`, `TIMESTAMPTZ` for dates |
| 3 | `transactions` | `SERIAL PK`, `DOUBLE PRECISION` for monetary cols, `TIMESTAMPTZ` for `transaction_date`/`created_at` |
| 4 | `savings_accounts` | `SERIAL PK`, `DOUBLE PRECISION`, `BOOLEAN` for `is_matured`, `DATE` for `start_date`/`maturity_date` |
| 5 | `nav_snapshots` | `SERIAL PK`, `DOUBLE PRECISION` for `nav`, `DATE` for `snapshot_date` |
| 6 | `pattern_observations` | `SERIAL PK`, `DOUBLE PRECISION` for price/outcome cols, `TIMESTAMPTZ` for `detection_date` |
| 7 | `alerts` | `SERIAL PK`, `BOOLEAN` for `viewed`/`expired`, `TIMESTAMPTZ` for `detection_timestamp` |
| 8 | `watchlists` | `SERIAL PK`, `TIMESTAMPTZ` for `created_at` |
| 9 | `watchlist_symbols` | `SERIAL PK`, `DOUBLE PRECISION` for price alert cols |
| 10 | `filter_presets` | `SERIAL PK`, `TIMESTAMPTZ` for `created_at` |
| 11 | `recommendation_audit_log` | `SERIAL PK`, `TIMESTAMPTZ` for `timestamp`/`created_at` |
| 12 | `financial_goals` | `SERIAL PK`, `DOUBLE PRECISION` for `target_amount`, `DATE` for `target_date` |
| 13 | `stock_sector_mapping` | `TEXT PK` (unchanged), `TIMESTAMPTZ` for `last_updated` |
| 14 | `cache_entries` | `TEXT PK` (unchanged), `TIMESTAMPTZ` for `expires_at`/`created_at` |

### 2. Query Layer — DML Conversion

Four files contain parameterized SQL queries that need conversion:

#### 2a. Placeholder Conversion (`?` → `$N`)

Every `?` placeholder is replaced with sequential `$1`, `$2`, etc. This is a mechanical transformation.

Example from `asset_registry.go`:
```go
// Before
`INSERT INTO assets (...) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

// After
`INSERT INTO assets (...) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
```

#### 2b. Date/Time Handling

All `time.Format()` calls for INSERT parameters and all `time.Parse()` calls for Scan results are removed. `pgx` handles `time.Time` ↔ PostgreSQL `TIMESTAMPTZ`/`DATE` natively.

Files affected and specific changes:

**savings_tracker.go:**
- `AddSavingsAccount`: Remove `account.StartDate.Format("2006-01-02")` and `maturityDate.Format("2006-01-02")` — pass `time.Time` directly
- `GetSavingsAccount`: Remove `time.Parse("2006-01-02", startDate)` — scan into `time.Time` directly
- `GetSavingsAccountsByUser`: Same scan changes as `GetSavingsAccount`
- `RefreshAllMaturityStatuses`: Remove `now.Format("2006-01-02")` — pass `time.Time` directly

**asset_registry.go:**
- `AddAsset`: Remove `.Format(time.RFC3339)` for `acquisition_date`, `created_at`, `updated_at`
- `UpdateAsset`: Remove `.Format(time.RFC3339)` for `acquisition_date`, `updated_at`
- `GetAsset`: Remove `time.Parse(time.RFC3339, ...)` for `acqDate`, `createdAt`, `updatedAt` — scan `time.Time` directly
- `GetAssetsByUser`: Same scan changes

**transaction_ledger.go:**
- `RecordTransaction`: Remove `.Format(time.RFC3339)` for `transaction_date`
- `scanTransactions`: Remove `time.Parse(time.RFC3339, ...)` for `txDate`, `createdAt` — scan `time.Time` directly

#### 2c. Boolean Handling

Remove `boolToInt()` helper entirely. Pass Go `bool` values directly to PostgreSQL `BOOLEAN` columns.

Files affected:
- `savings_tracker.go`: `AddSavingsAccount` — replace `boolToInt(false)` with `false`
- `savings_tracker.go`: `UpdateMaturityStatus` — replace `boolToInt(true)` with `true`
- `savings_tracker.go`: `RefreshAllMaturityStatuses` — replace `is_matured = 1` with `is_matured = TRUE` and `is_matured = 0` with `is_matured = FALSE`
- Scan side: Replace `var isMatured int` + `isMatured != 0` with `var isMatured bool` directly

#### 2d. LastInsertId → RETURNING

PostgreSQL's `pgx` driver does not support `LastInsertId()`. All INSERT operations that need the generated ID must use `RETURNING id` with `QueryRowContext` + `Scan`.

Files affected:
- `asset_registry.go`: `AddAsset` — change `ExecContext` → `QueryRowContext(...).Scan(&id)`
- `transaction_ledger.go`: `RecordTransaction` — same pattern
- `savings_tracker.go`: `AddSavingsAccount` — same pattern

Example:
```go
// Before
result, err := r.db.ExecContext(ctx, `INSERT INTO assets (...) VALUES (?, ...)`, ...)
id, err := result.LastInsertId()

// After
var id int64
err := r.db.QueryRowContext(ctx,
    `INSERT INTO assets (...) VALUES ($1, ...) RETURNING id`, ...,
).Scan(&id)
```

### 3. Test Infrastructure

#### Shared Test Helper: `testhelper_test.go`

A new file provides a reusable function that:
1. Starts a PostgreSQL 16 container via `testcontainers-go`
2. Waits for readiness
3. Runs all 14 production DDL migrations
4. Inserts a seed user (`testuser`)
5. Returns `*sql.DB`
6. Registers `t.Cleanup` to terminate the container

```go
func setupPostgresTestDB(t *testing.T) *sql.DB {
    t.Helper()
    ctx := context.Background()
    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("myfi_test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).WithStartupTimeout(30*time.Second)),
    )
    if err != nil {
        t.Fatalf("failed to start postgres container: %v", err)
    }
    t.Cleanup(func() { container.Terminate(ctx) })

    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        t.Fatalf("failed to get connection string: %v", err)
    }
    db, err := sql.Open("pgx", connStr)
    if err != nil {
        t.Fatalf("failed to open test db: %v", err)
    }
    t.Cleanup(func() { db.Close() })

    // Run production migrations
    for _, m := range allMigrations() {
        if _, err := db.Exec(m); err != nil {
            t.Fatalf("migration failed: %v", err)
        }
    }
    // Seed test user
    if _, err := db.Exec(
        `INSERT INTO users (username, password_hash) VALUES ('testuser', 'hash')`); err != nil {
        t.Fatalf("failed to seed user: %v", err)
    }
    return db
}
```

The `allMigrations()` function returns the same slice used by `runMigrations()`, ensuring test and production schemas are always in sync.

#### Test File Changes

| Test File | Changes |
|-----------|---------|
| `asset_registry_test.go` | Replace `setupTestDB` with `setupPostgresTestDB`, remove SQLite import, remove inline DDL, remove `PRAGMA` |
| `savings_tracker_test.go` | Replace `setupSavingsTestDB` with `setupPostgresTestDB`, remove SQLite import, remove inline DDL |
| `transaction_ledger_test.go` | Uses `setupTestDB` from asset_registry_test — inherits changes |
| `portfolio_engine_test.go` | Uses `setupTestDB` — inherits changes |
| `portfolio_engine_property_test.go` | Replace `setupPropDB` with `setupPostgresTestDB`, remove SQLite import, remove inline DDL |
| `debug_nav_test.go` | Uses `newTestSavingsTracker` — inherits changes |

### 4. Docker Compose

`myfi/docker-compose.yml`:

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myfi
      POSTGRES_USER: myfi
      POSTGRES_PASSWORD: myfi_dev
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myfi -d myfi"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

### 5. Environment Configuration

`myfi/.env` updated:

```env
DATABASE_URL=postgres://myfi:myfi_dev@localhost:5432/myfi?sslmode=disable
```

Remove `DB_TYPE` and `DB_PATH` references.

### 6. Go Module Changes

```
# Add
github.com/jackc/pgx/v5
github.com/testcontainers/testcontainers-go
github.com/testcontainers/testcontainers-go/modules/postgres

# Remove
github.com/mattn/go-sqlite3
```

## Data Models

No structural changes to Go structs. The data model layer (`Asset`, `Transaction`, `SavingsAccount`, etc.) already uses `time.Time` and `bool` — the migration removes the impedance mismatch between Go types and SQLite's string/integer storage.

Key type mappings that become direct:

| Go Type | SQLite Storage | PostgreSQL Storage |
|---------|---------------|-------------------|
| `time.Time` | `string` (formatted) | `TIMESTAMPTZ` / `DATE` (native) |
| `bool` | `int` (0/1 via `boolToInt`) | `BOOLEAN` (native) |
| `float64` | `REAL` (32-bit) | `DOUBLE PRECISION` (64-bit) |
| `int64` (ID) | `INTEGER AUTOINCREMENT` | `SERIAL` (sequence-backed) |

The `float64` → `DOUBLE PRECISION` upgrade is notable: SQLite's `REAL` is IEEE 754 64-bit in practice, but PostgreSQL's `DOUBLE PRECISION` makes this explicit and guaranteed.


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

The migration's core correctness concern is that data stored via PostgreSQL is retrieved identically to what was stored — particularly for types where SQLite required manual string conversion (dates, booleans). The properties below formalize this.

### Property 1: Date/Time Round-Trip Preservation

*For any* valid `time.Time` value (with year, month, day, hour, minute, second components), storing it into a `TIMESTAMP WITH TIME ZONE` or `DATE` column and then retrieving it should produce a `time.Time` value where the year, month, and day match the original. For `TIMESTAMPTZ` columns, hour, minute, and second should also match (within the same timezone).

This property subsumes the savings_tracker date bug: if the round-trip holds for all dates, it holds for savings account start dates specifically.

**Validates: Requirements 3.2, 3.3, 7.1, 7.2, 7.5**

### Property 2: Boolean Round-Trip Preservation

*For any* Go `bool` value (`true` or `false`), storing it into a `BOOLEAN` column and then retrieving it should produce the same `bool` value. Specifically, storing `true` and scanning back should yield `true`, and storing `false` and scanning back should yield `false` — without any integer conversion.

**Validates: Requirements 3.4, 3.6**

### Property 3: Constraint Preservation

*For any* value that violates a `CHECK`, `FOREIGN KEY`, `UNIQUE`, or `NOT NULL` constraint defined in the schema, attempting to insert or update that value should return a database error. The PostgreSQL schema must enforce the same constraints as the original SQLite schema.

**Validates: Requirements 2.7**

### Property 4: INSERT RETURNING ID Correctness

*For any* valid entity (asset, transaction, or savings account), inserting it via `INSERT ... RETURNING id` should return a positive, unique integer ID. Inserting N entities sequentially should produce N distinct IDs, all positive.

**Validates: Requirements 9.6**

## Error Handling

### Connection Errors

- `InitDB()` returns a descriptive error if `DATABASE_URL` is empty or malformed
- `db.Ping()` failure returns a wrapped error with context
- Connection pool exhaustion is mitigated by `SetMaxOpenConns(25)` — queries block rather than fail when the pool is full

### Migration Errors

- Each DDL statement is executed sequentially; failure on any statement returns the migration index and wrapped error
- `CREATE TABLE IF NOT EXISTS` is idempotent — safe to re-run on existing schemas
- `CREATE INDEX IF NOT EXISTS` is similarly idempotent

### Query Errors

- All `QueryRowContext` / `QueryContext` / `ExecContext` errors are wrapped with `fmt.Errorf("context: %w", err)` — this pattern is already in place and unchanged
- `RETURNING id` scan errors are handled the same way as the previous `LastInsertId()` errors
- PostgreSQL constraint violations (CHECK, FK, UNIQUE) return `pq`-style error codes that can be inspected if needed, but the current code treats all DB errors uniformly via `error` interface

### Test Container Errors

- `testcontainers-go` container startup failure calls `t.Fatalf` — test fails immediately with a clear message
- Container cleanup is registered via `t.Cleanup` — runs even if the test panics
- Connection string retrieval failure is also fatal

## Testing Strategy

### Dual Testing Approach

The migration uses both unit tests and property-based tests:

- **Unit tests**: Verify specific examples (add a savings account with known values, delete an asset and check cascade, etc.). These are the existing tests in `*_test.go` files, updated to run against PostgreSQL.
- **Property-based tests**: Verify universal properties across randomly generated inputs. The existing property tests in `portfolio_engine_property_test.go` (Properties 14-18) continue to run. New property tests validate the migration-specific concerns (date round-trip, boolean round-trip, constraint preservation, RETURNING id).

### Property-Based Testing Configuration

- **Library**: `pgregory.net/rapid` (already in use for portfolio engine property tests)
- **Minimum iterations**: 100 per property test
- **Database**: Each property test uses `setupPostgresTestDB(t)` to get a fresh PostgreSQL container
- **Tag format**: Each test includes a comment referencing its design property

### New Property Tests

Each correctness property maps to a single property-based test:

1. **Property 1 (Date/Time Round-Trip)**: Generate random `time.Time` values, insert a savings account with that start date, retrieve it, assert year/month/day match. Also test with `TIMESTAMPTZ` columns via assets (acquisition_date).
   - Tag: `// Feature: postgresql-migration, Property 1: Date/Time Round-Trip Preservation`

2. **Property 2 (Boolean Round-Trip)**: Generate random `bool` values, insert a savings account with `is_matured` set to that value, retrieve it, assert the boolean matches.
   - Tag: `// Feature: postgresql-migration, Property 2: Boolean Round-Trip Preservation`

3. **Property 3 (Constraint Preservation)**: Generate invalid inputs (e.g., asset_type not in CHECK list, NULL for NOT NULL columns, duplicate UNIQUE keys, invalid FK references), attempt insert, assert error is returned.
   - Tag: `// Feature: postgresql-migration, Property 3: Constraint Preservation`

4. **Property 4 (INSERT RETURNING ID)**: Generate N random assets (1-20), insert them sequentially, collect all returned IDs, assert all are positive and all are distinct.
   - Tag: `// Feature: postgresql-migration, Property 4: INSERT RETURNING ID Correctness`

### Existing Tests (Updated, Not Rewritten)

The following existing test files are updated to use `setupPostgresTestDB` but their test logic remains unchanged:

| Test File | Test Count | Key Coverage |
|-----------|-----------|--------------|
| `asset_registry_test.go` | 10 tests | CRUD, validation, cascade delete, NAV computation |
| `transaction_ledger_test.go` | 5 tests | Record, retrieve, validate, auto-compute total |
| `savings_tracker_test.go` | 13 tests | Add, get, delete, maturity, compound interest, NAV |
| `portfolio_engine_property_test.go` | 4 property tests | Buy double-entry, sell P&L, unrealized P&L, NAV aggregation, insufficient holdings |
| `debug_nav_test.go` | 1 test | Date parsing debug (should pass cleanly after migration) |

### Test Execution

Tests require Docker to be running (for testcontainers-go). Run with:

```bash
cd myfi/backend && go test ./... -v -count=1
```

The `-count=1` flag disables test caching, which is important since testcontainers creates fresh containers each run.
