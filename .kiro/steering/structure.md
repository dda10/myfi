# Project Structure

## Root Layout

```
/
├── backend/          # Go API server
├── frontend/         # Next.js application
└── .kiro/           # Kiro configuration and specs
```

## Backend Structure (`backend/`)

```
backend/
├── go.mod                          # Go dependencies (module: myfi-backend)
├── go.sum
├── cmd/
│   └── server/
│       └── main.go                 # Entry point — wiring, DI, startup
├── internal/
│   ├── handler/                    # HTTP handlers (Gin)
│   │   ├── handlers.go            # Handlers struct (dependency injection)
│   │   ├── routes.go              # Route registration + CORS
│   │   ├── market.go              # Market data handlers
│   │   ├── crypto.go              # Cryptocurrency handlers
│   │   ├── news.go                # News handlers
│   │   └── agent.go               # AI chat handlers
│   ├── service/                    # Business logic
│   │   ├── price_service.go
│   │   ├── crypto_service.go
│   │   ├── gold_service.go
│   │   ├── fx_service.go
│   │   ├── portfolio_engine.go
│   │   ├── transaction_ledger.go
│   │   ├── asset_registry.go
│   │   ├── savings_tracker.go
│   │   ├── sector_service.go
│   │   ├── market_data_service.go
│   │   ├── screener_service.go
│   │   ├── watchlist_service.go
│   │   ├── commodity_service.go
│   │   ├── fund_service.go
│   │   ├── macro_service.go
│   │   ├── performance_engine.go
│   │   ├── comparison_engine.go
│   │   └── liquidity_filter.go
│   ├── infra/                      # Infrastructure (cache, DB, resilience)
│   │   ├── cache.go
│   │   ├── database.go
│   │   ├── circuit_breaker.go
│   │   ├── rate_limiter.go
│   │   └── data_source_router.go
│   ├── model/                      # Shared types, constants, enums
│   │   ├── asset_types.go
│   │   ├── transaction_types.go
│   │   ├── price_types.go
│   │   ├── market_types.go
│   │   ├── portfolio_types.go
│   │   ├── sector_types.go
│   │   ├── savings_types.go
│   │   ├── data_category.go
│   │   ├── handler_types.go
│   │   ├── rate_limit_types.go
│   │   └── liquidity_types.go
│   └── testutil/
│       └── testhelper.go           # Shared test helpers
```

## Frontend Structure (`frontend/`)

```
frontend/
├── src/
│   ├── app/
│   │   ├── [tab]/page.tsx      # Dynamic route for tabs
│   │   ├── layout.tsx           # Root layout with providers
│   │   ├── page.tsx             # Redirects to /overview
│   │   └── globals.css          # Global styles
│   ├── components/
│   │   ├── dashboard/           # Dashboard modules
│   │   │   ├── MarketChart.tsx  # Candlestick/volume chart
│   │   │   ├── Watchlist.tsx    # Stock watchlist
│   │   │   ├── Stats.tsx        # Market statistics
│   │   │   ├── FilterModule.tsx # Stock screener
│   │   │   └── ...
│   │   ├── chat/
│   │   │   └── ChatWidget.tsx   # AI chat interface
│   │   └── layout/              # Layout components
│   └── context/
│       ├── AppContext.tsx       # Global app state (active tab, symbol)
│       └── WatchlistContext.tsx # Watchlist state management
├── public/                      # Static assets
└── package.json
```

## Key Conventions

- Frontend uses App Router with dynamic `[tab]` routing
- Tab navigation managed via URL paths (`/overview`, `/markets`, etc.)
- Components are organized by feature (dashboard, chat, layout)
- Context providers wrap the app in `layout.tsx`
- Backend follows Go Server project layout (`cmd/server/` + `internal/`)
- Package dependency direction: `cmd → handler → service → infra → model` (acyclic)
- Handler dependencies injected via `Handlers` struct (no package-level globals)
- All API routes prefixed with `/api/`
