// Package portfolio provides portfolio management domain logic including
// buy/sell engine, transaction ledger, performance analytics, risk metrics,
// corporate action processing, and export services.
//
// # Migration Status
//
// This package is a placeholder for the portfolio domain. The files listed
// below currently live in service/, handler/, and model/ and will be moved
// here in a future refactoring pass once cross-package dependencies
// (HoldingStore, PriceService, TransactionLedger, Handlers struct) are
// resolved and backward-compatibility aliases are in place.
//
// # Planned File Moves
//
// From service/ (business logic):
//
//	service/portfolio_engine.go        → domain/portfolio/engine.go
//	service/transaction_ledger.go      → domain/portfolio/ledger.go
//	service/performance_engine.go      → domain/portfolio/performance.go
//	service/risk_service.go            → domain/portfolio/risk.go
//	service/corporate_action_service.go → domain/portfolio/corporate.go
//	service/export_service.go          → domain/portfolio/export.go
//
// From handler/ (HTTP handlers):
//
//	handler/portfolio.go               → domain/portfolio/handler.go
//	handler/export.go                  → domain/portfolio/export_handler.go
//
// From model/ (type definitions):
//
//	model/portfolio_types.go           → domain/portfolio/types.go
//	model/transaction_types.go         → domain/portfolio/transaction_types.go
//	model/corporate_action_types.go    → domain/portfolio/corporate_types.go
//
// # Blocking Dependencies
//
// The move is deferred because these files have deep cross-package
// dependencies that would create import cycles if moved today:
//
//   - portfolio_engine.go depends on service.TransactionLedger and
//     service.PriceService (aliased from domain/market).
//   - corporate_action_service.go depends on service.TransactionLedger
//     and infra.DataSourceRouter.
//   - performance_engine.go and risk_service.go depend on infra and
//     vnstock-go directly.
//   - handler/portfolio.go and handler/export.go depend on the
//     handler.Handlers struct for dependency injection.
//
// Once task 3.6 replaces handler.Handlers with a platform-level DI
// container and all service cross-references are resolved, these files
// can be moved here with backward-compatibility aliases in service/ and
// model/ (following the same pattern used for domain/market).
package portfolio
