# Project Structure

## Root Layout

```
/
├── backend/          # Go API server
├── frontend/         # Next.js application
└── .kiro/           # Kiro configuration and specs
```

## Backend Structure (`backend/`)

```text
backend/
├── go.mod                          # Go dependencies (module: myfi-backend)
├── go.sum
├── cmd/
│   └── server/
│       └── main.go                 # Entry point — wiring, DI, startup
├── internal/
│   ├── domain/                     # Domain packages (each self-contained)
│   │   ├── market/                 # Market data, pricing, sectors, macro, search
│   │   │   ├── service.go          # PriceService
│   │   │   ├── market_data.go      # MarketDataService (unified data layer)
│   │   │   ├── sector.go           # SectorService
│   │   │   ├── macro.go            # MacroService (VCB FX, gold, interbank)
│   │   │   ├── search.go           # SearchService
│   │   │   ├── handler.go          # HTTP handlers
│   │   │   ├── types.go            # Domain types
│   │   │   ├── macro_types.go      # MacroFXRate, MacroGoldPrice, WorldIndex
│   │   │   └── data_category.go    # DataCategory enum, SourcePreference
│   │   ├── portfolio/              # Holdings, transactions, performance, risk, export
│   │   ├── screener/               # Screener (delegates to vnstock Screen), liquidity filter
│   │   ├── watchlist/              # Named watchlists with alerts
│   │   ├── agent/                  # AI chat proxy (gRPC to Python)
│   │   ├── ranking/                # AI ranking, backtest, recommendation tracker
│   │   ├── mission/                # User missions (price alerts, scheduled tasks)
│   │   ├── notification/           # Notifications with rate limiting
│   │   ├── knowledge/              # Knowledge base observations
│   │   ├── analyst/                # Analyst IQ, consensus
│   │   ├── auth/                   # JWT auth, account lockout
│   │   ├── consensus/              # LLM sentiment consensus
│   │   ├── sentiment/              # Sentiment analysis
│   │   └── fund/                   # Mutual fund data (via FMARKET connector)
│   ├── platform/                   # Cross-cutting glue
│   │   ├── router.go              # Gin router, route registration
│   │   ├── middleware.go          # Logging, recovery, auth middleware
│   │   ├── docs.go               # OpenAPI spec + Swagger UI (GET /api/docs)
│   │   ├── config.go             # App config from env vars
│   │   └── server.go             # HTTP server lifecycle
│   ├── infra/                      # Infrastructure
│   │   ├── cache.go               # Redis cache wrapper
│   │   ├── database.go            # PostgreSQL connection + migrations
│   │   ├── data_source_router.go  # Multi-connector failover (VCI/KBS/VND/ENTRADE/CAFEF/MSN/GOLD/FMARKET)
│   │   ├── data_category.go       # DataCategory enum (infra copy)
│   │   ├── circuit_breaker.go     # gobreaker with typed error classification
│   │   ├── rate_limiter.go        # Per-source rate limits
│   │   ├── scheduler.go           # Trading-hours-aware scheduler (vnstock TradingHours)
│   │   ├── grpc_client.go         # gRPC client to Python AI Service
│   │   ├── storage.go             # S3/MinIO storage abstraction
│   │   ├── email.go               # SMTP/SES email sender
│   │   ├── telemetry.go           # OpenTelemetry tracing
│   │   ├── metrics.go             # Prometheus metrics (GET /metrics)
│   │   └── logging.go            # Centralized JSON structured logging
│   ├── model/                      # Shared types (legacy, being migrated to domain/)
│   └── service/                    # Compat wrappers (legacy, being migrated to domain/)
```

## Frontend Structure (`frontend/`)

```text
frontend/
├── src/
│   ├── app/
│   │   ├── (app)/               # Authenticated routes
│   │   │   ├── dashboard/       # Market overview dashboard
│   │   │   ├── stock/[symbol]/  # Stock detail page
│   │   │   ├── screener/        # Stock screener
│   │   │   ├── portfolio/       # Portfolio management
│   │   │   ├── ranking/         # AI ranking
│   │   │   ├── ideas/           # Investment ideas
│   │   │   ├── macro/           # Macro indicators
│   │   │   ├── heatmap/         # Market heatmap
│   │   │   ├── research/        # Research reports
│   │   │   ├── funds/           # Mutual fund analysis
│   │   │   └── layout.tsx       # App shell (sidebar, header)
│   │   ├── (auth)/              # Auth routes (login)
│   │   ├── layout.tsx           # Root layout with providers
│   │   └── globals.css
│   ├── features/                # Feature modules
│   │   ├── stock/components/    # Stock detail components
│   │   ├── screener/components/ # Screener UI
│   │   ├── portfolio/components/# Portfolio UI
│   │   ├── dashboard/components/# Dashboard widgets
│   │   ├── chart/components/    # Chart engine + indicators
│   │   └── chat/components/     # AI chat widget
│   ├── components/
│   │   ├── layout/              # Header, Sidebar, MobileNav
│   │   └── dashboard/           # Shared dashboard components
│   ├── context/                 # React Context providers
│   ├── i18n/                    # Internationalization (vi-VN, en-US)
│   └── lib/                     # Utilities, indicator calculations
├── public/
└── package.json
```

## Key Conventions

- Frontend uses App Router with `(app)` and `(auth)` route groups
- Feature-based component organization (`features/stock/`, `features/screener/`, etc.)
- Context providers wrap the app in `layout.tsx`
- Backend follows domain-driven layout: `cmd → platform → domain → infra` (acyclic)
- Each domain package is self-contained: handler + service + types in one directory
- Domains never import each other; shared types live in `model/` (legacy) or are duplicated
- Infrastructure layer provides DataSourceRouter with multi-level failover across 9 vnstock-go connectors
- Circuit breaker uses typed error classification: only `NetworkError` and `RateLimited` trip the breaker
- Scheduler uses `vnstock.TradingHours("HOSE")` for dynamic trading session detection
- All API routes prefixed with `/api/`
