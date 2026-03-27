# Requirements Document

## Introduction

Restructure the myfi-backend Go codebase from a flat single-package layout (all files in `backend/`, all `package main`) to the official Go "Server project" layout recommended at https://go.dev/doc/modules/layout. The restructuring introduces `cmd/server/main.go` as the entry point and organizes all source files into `internal/` sub-packages (`handler`, `service`, `infra`, `model`). The goal is zero behavior change — all existing API endpoints, business logic, and tests continue to work identically after the migration.

## Glossary

- **Backend**: The Go API server located in the `backend/` directory of the monorepo.
- **Flat_Layout**: The current project structure where all 30+ Go source files reside in a single `backend/` directory under `package main`.
- **Server_Layout**: The Go "Server project" layout pattern from https://go.dev/doc/modules/layout, using `cmd/` for entry points and `internal/` for private packages.
- **Entry_Point**: The `cmd/server/main.go` file that contains only server wiring, dependency initialization, and startup logic.
- **Handler_Package**: The `internal/handler/` package containing HTTP handler functions that accept `*gin.Context` and delegate to services.
- **Service_Package**: The `internal/service/` package containing business logic, domain engines, and data-fetching services.
- **Infra_Package**: The `internal/infra/` package containing infrastructure concerns: caching, database, circuit breaker, rate limiter, and data source routing.
- **Model_Package**: The `internal/model/` package containing shared types, structs, constants, and enums used across packages.
- **Import_Path**: A fully qualified Go import path using the module name (e.g., `myfi-backend/internal/handler`).
- **Steering_Docs**: The Kiro steering documents in `.kiro/steering/` that describe the project structure and conventions.

## Requirements

### Requirement 1: Create the cmd/server Entry Point

**User Story:** As a developer, I want a dedicated `cmd/server/main.go` entry point, so that the server startup logic is separated from business code following the Go Server project layout.

#### Acceptance Criteria

1. THE Entry_Point SHALL reside at `backend/cmd/server/main.go` with `package main`.
2. THE Entry_Point SHALL contain only server wiring: Gin router setup, CORS middleware, route registration, dependency initialization, and `r.Run(":8080")`.
3. THE Entry_Point SHALL import Handler_Package, Service_Package, Infra_Package, and Model_Package using fully qualified Import_Paths.
4. WHEN the Entry_Point is executed via `go run ./cmd/server`, THE Backend SHALL start and listen on port 8080 with all existing API routes registered.
5. THE Entry_Point SHALL initialize all dependencies (database, cache, data source router, services, handlers) and pass them to handlers via constructor injection or struct fields.
6. IF the database initialization fails, THEN THE Entry_Point SHALL log the error and terminate the process.

### Requirement 2: Create the internal/handler Package

**User Story:** As a developer, I want HTTP handlers grouped in an `internal/handler/` package, so that request/response logic is separated from business logic.

#### Acceptance Criteria

1. THE Handler_Package SHALL contain the following files migrated from the Flat_Layout: `market.go` (market data handlers), `crypto.go` (cryptocurrency handlers), `news.go` (news handlers), `agent.go` (AI chat handlers).
2. THE Handler_Package SHALL declare `package handler`.
3. WHEN a handler function is moved to Handler_Package, THE handler function name SHALL be exported (capitalized) so the Entry_Point can reference the handler.
4. THE Handler_Package SHALL import Service_Package and Model_Package for business logic and shared types.
5. THE Handler_Package SHALL accept service dependencies via a struct (e.g., `Handlers` struct) initialized by the Entry_Point, rather than relying on package-level global variables.
6. WHEN any existing API endpoint is called after migration, THE Handler_Package SHALL return the same HTTP status codes and JSON response bodies as the Flat_Layout.

### Requirement 3: Create the internal/service Package

**User Story:** As a developer, I want business logic and domain engines grouped in an `internal/service/` package, so that services are reusable and testable independently of HTTP concerns.

#### Acceptance Criteria

1. THE Service_Package SHALL contain the following files migrated from the Flat_Layout: `price_service.go`, `crypto_service.go`, `gold_service.go`, `fx_service.go`, `savings_tracker.go`, `watchlist_service.go`, `screener_service.go`, `sector_service.go`, `commodity_service.go`, `fund_service.go`, `macro_service.go`, `market_data_service.go`, `portfolio_engine.go`, `performance_engine.go`, `comparison_engine.go`, `transaction_ledger.go`, `asset_registry.go`.
2. THE Service_Package SHALL declare `package service`.
3. WHEN a type, function, or method is moved to Service_Package, THE exported name SHALL remain unchanged (already-exported symbols keep their names; previously unexported symbols used cross-package SHALL be exported).
4. THE Service_Package SHALL import Infra_Package for cache, database, circuit breaker, and rate limiter dependencies.
5. THE Service_Package SHALL import Model_Package for shared types and constants.

### Requirement 4: Create the internal/infra Package

**User Story:** As a developer, I want infrastructure code grouped in an `internal/infra/` package, so that cross-cutting concerns like caching, database access, and resilience patterns are centralized.

#### Acceptance Criteria

1. THE Infra_Package SHALL contain the following files migrated from the Flat_Layout: `cache.go`, `database.go`, `circuit_breaker.go`, `rate_limiter.go`, `data_source_router.go`.
2. THE Infra_Package SHALL declare `package infra`.
3. WHEN a type or function is moved to Infra_Package, THE exported name SHALL remain unchanged.
4. THE Infra_Package SHALL not import Handler_Package or Service_Package (no upward dependency).

### Requirement 5: Create the internal/model Package

**User Story:** As a developer, I want shared types, structs, and constants in an `internal/model/` package, so that all packages can reference common domain types without circular imports.

#### Acceptance Criteria

1. THE Model_Package SHALL contain shared types currently defined across multiple files in the Flat_Layout, including: `AssetType` and its constants, `TransactionType` and its constants, `CompoundingFrequency` and its constants, `ICBSector` and its constants, `DataCategory` and its constants, request/response structs used by both handlers and services (e.g., `ChatRequest`, `ModelsRequest`, `RSS`, `Channel`, `Item`).
2. THE Model_Package SHALL declare `package model`.
3. THE Model_Package SHALL not import Handler_Package, Service_Package, or Infra_Package (no upward dependency).
4. WHEN a type is moved to Model_Package, THE type name SHALL remain unchanged.

### Requirement 6: Update All Import Paths

**User Story:** As a developer, I want all import paths updated to reference the new package structure, so that the codebase compiles without errors after restructuring.

#### Acceptance Criteria

1. WHEN a file is moved from the Flat_Layout to a sub-package, THE file SHALL use Import_Paths of the form `myfi-backend/internal/<package>` to reference other internal packages.
2. THE Backend SHALL compile successfully with `go build ./...` from the `backend/` directory after all files are moved.
3. THE Backend SHALL pass `go vet ./...` with zero warnings after restructuring.
4. IF a circular import is detected during restructuring, THEN THE developer SHALL resolve the circular import by moving the shared type to Model_Package.

### Requirement 7: Migrate Test Files

**User Story:** As a developer, I want test files co-located with their source files in the new package structure, so that tests remain discoverable and runnable.

#### Acceptance Criteria

1. WHEN a source file is moved to a sub-package, THE corresponding test file SHALL be moved to the same sub-package directory.
2. THE test files SHALL update their `package` declaration to match the new package (e.g., `package service` or `package service_test` for black-box tests).
3. THE test files SHALL update import paths to reference types and functions from the correct internal packages.
4. WHEN `go test ./...` is run from the `backend/` directory, THE Backend SHALL execute all existing tests and all tests SHALL pass.
5. THE `testhelper_test.go` file SHALL be moved to the package where the test helpers are used, or to a shared `internal/testutil/` package if used by multiple packages.
6. THE `testdata/` directory SHALL be moved alongside the test files that reference the test data.

### Requirement 8: Preserve All API Endpoints

**User Story:** As a developer, I want all existing API endpoints to remain functional after restructuring, so that the frontend and any external clients experience zero breaking changes.

#### Acceptance Criteria

1. THE Backend SHALL register all existing API routes with identical paths after restructuring: `/api/health`, `/api/metrics/rate-limits`, `/api/market/quote`, `/api/market/chart`, `/api/market/listing`, `/api/market/screener`, `/api/market/company/:symbol`, `/api/market/finance/:symbol`, `/api/market/trading/:symbol`, `/api/market/statistics`, `/api/market/valuation`, `/api/market/funds`, `/api/market/commodities`, `/api/market/macro`, `/api/market/trading/batch`, `/api/crypto/quote`, `/api/prices/fx`, `/api/news`, `/api/chat`, `/api/models`.
2. THE Backend SHALL return identical JSON response structures for all endpoints.
3. THE Backend SHALL listen on port 8080 as before.
4. THE CORS middleware SHALL remain configured identically (allow all origins, same headers and methods).

### Requirement 9: Update go.mod Module Path

**User Story:** As a developer, I want the go.mod module path to remain valid and consistent with the new directory structure, so that all internal imports resolve correctly.

#### Acceptance Criteria

1. THE `go.mod` file SHALL retain the module name `myfi-backend`.
2. WHEN `go mod tidy` is run from the `backend/` directory, THE command SHALL complete without errors.
3. THE `go.sum` file SHALL be updated to reflect any changes from `go mod tidy`.

### Requirement 10: Update Steering Documentation

**User Story:** As a developer, I want the Kiro steering documents updated to reflect the new project layout, so that future development follows the correct structure conventions.

#### Acceptance Criteria

1. THE Steering_Docs file `.kiro/steering/structure.md` SHALL be updated to show the new backend directory tree with `cmd/server/`, `internal/handler/`, `internal/service/`, `internal/infra/`, and `internal/model/`.
2. THE Steering_Docs file `.kiro/steering/tech.md` SHALL be updated to reflect the new build and run commands: `go run ./cmd/server` instead of `go run .`.
3. THE Steering_Docs SHALL document the package dependency direction: `cmd → handler → service → infra/model`, with Model_Package and Infra_Package having no upward dependencies.

### Requirement 11: Maintain Package Dependency Direction

**User Story:** As a developer, I want a clear, acyclic package dependency graph, so that the codebase remains maintainable and free of circular imports.

#### Acceptance Criteria

1. THE package dependency direction SHALL follow: Entry_Point imports Handler_Package, Service_Package, Infra_Package, and Model_Package; Handler_Package imports Service_Package and Model_Package; Service_Package imports Infra_Package and Model_Package; Infra_Package imports Model_Package; Model_Package imports no internal packages.
2. THE Backend SHALL compile with zero circular import errors.
3. IF a function or type creates a circular dependency when moved, THEN THE developer SHALL extract the shared dependency into Model_Package or introduce an interface to break the cycle.

### Requirement 12: Remove Package-Level Global Variables

**User Story:** As a developer, I want service instances passed via dependency injection rather than package-level globals, so that the code is testable and the dependency graph is explicit.

#### Acceptance Criteria

1. WHEN services are initialized in the Entry_Point, THE Entry_Point SHALL pass service instances to handlers via constructor functions or struct initialization.
2. THE Handler_Package SHALL not declare package-level `var` statements for service instances (e.g., the current `var vnstockClient`, `var dataSourceRouter`, `var fxService` pattern in `market.go` SHALL be replaced with struct fields).
3. THE `init()` function currently in `market.go` SHALL be replaced with explicit initialization in the Entry_Point.
