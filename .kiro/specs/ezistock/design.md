# Design Document — EziStock Platform

## Overview

EziStock is a Vietnamese stock quantitative analysis and AI-powered investment platform focused exclusively on stocks listed on HOSE, HNX, and UPCOM exchanges. The system is composed of three services:

1. **Go Backend** — Market data ingestion (via vnstock-go), portfolio management, screener, watchlists, missions, notifications, corporate actions, export, auth, and all REST API endpoints. Uses Gin framework, PostgreSQL, Redis, and S3-compatible storage.
2. **Python AI Service** — Multi-agent AI system (LangChain/LangGraph) with Technical Analyst, News Analyst, Investment Advisor, and Strategy Builder agents. Hosts the Alpha Mining Engine (Data/Model/Backtest/Deployment layers) and Feedback Loop Engine. Communicates with Go Backend via gRPC.
3. **Next.js Frontend** — App Router with TypeScript, Tailwind CSS, lightweight-charts for TradingView-style charting, Recharts for dashboards, Framer Motion for animations. Supports Vietnamese/English, dark/light themes, mobile responsive.

The platform's core differentiator is the multi-agent AI system where the quantitative engine underneath is the real value — agents are the storytelling layer that translates quantitative signals into human-readable investment theses. The Alpha Mining Engine moves beyond fixed factors to data-driven, regime-aware signal discovery with multi-strategy ensemble and consensus-based stock ranking.

All references to gold services (Doji), cryptocurrency (CoinGecko), savings tracker, FX service, commodity service, fund service, goal planner, bonds, and multi-asset types from the previous "MyFi" spec are removed.

### vnstock-go Expanded Capabilities (v2 Update)

The vnstock-go library now supports 12 connectors beyond the original VCI/KBS/VND:

| Connector | Coverage | EziStock Usage |
|-----------|----------|----------------|
| VCI | Full stock data (OHLCV, listings, indices, profiles, financials, ratios, price board, price depth, events, news, insider trades, subsidiaries) | Primary data source |
| KBS | Full stock data + shareholders, covered warrants, ETFs, bonds, industry groups | Primary fallback + KBS-specific data |
| VND | OHLCV via dchart API, listings | Tertiary fallback for OHLCV |
| VND_FINFO | Quotes, listings, profiles, financial ratios | Additional ratio source |
| DNSE | Quotes, profiles, officers | Supplementary quotes |
| ENTRADE | OHLCV chart data (1m to 1M intervals) | Intraday chart fallback |
| CAFEF | Historical price data | Historical data fallback |
| MSN | Forex quotes, crypto quotes, world market indices (S&P 500, NASDAQ, Nikkei, etc.) | Dashboard world markets widget |
| FMARKET | Mutual fund listing, search, holdings, NAV, industry/asset allocation | Fund analysis page |
| GOLD | SJC + BTMC gold prices | Dashboard gold price widget, macro page |
| FMP | Stub (not implemented) | — |
| Binance | Stub (not implemented) | — |

New API methods available:
- `PriceBoard(ctx, symbols)` — real-time bid/ask depth (VCI, KBS)
- `PriceDepth(ctx, symbol)` — 3-level order book (VCI, KBS)
- `Shareholders(ctx, symbol)` — shareholder data (KBS)
- `Subsidiaries(ctx, symbol)` — subsidiary companies (VCI, KBS)
- `Screen(ctx, criteria)` — built-in stock screener with financial criteria
- `GoldPrice(ctx, request)` — gold prices from SJC/BTMC
- `VCBExchangeRate(ctx, date)` — Vietcombank official exchange rates
- `TradingHours(market)` — market session status (HOSE, HNX, UPCOM, Futures)
- `FundListing/FundFilter/FundTopHolding/FundNAVReport/FundAssetHolding` — mutual fund data
- DNSE Trading API — `PlaceOrder`, `CancelOrder`, `AccountBalance`, conditional orders, margin

Built-in infrastructure:
- LRU response caching (configurable capacity + TTL) — reduces Redis dependency for read-heavy data
- Anti-rate-limiting (UA rotation, browser profiles, source-specific headers, proxy rotation)
- Typed error codes (`NetworkError`, `RateLimited`, `NotSupported`, etc.)
- CSV/JSON/Excel export
- Static reference data (indices, sectors, exchanges — no network calls)

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph Frontend["Next.js Frontend (Port 3000)"]
        UI[App Router Pages]
        Charts[Chart Engine<br/>lightweight-charts]
        SW[Service Worker<br/>Offline Cache]
    end

    subgraph GoBackend["Go Backend (Port 8080)"]
        GinRouter[Gin Router + Middleware]
        Handlers[HTTP Handlers]
        Services[Service Layer]
        Infra[Infrastructure Layer]
        DSR[Data Source Router]
    end

    subgraph PythonAI["Python AI Service (Port 50051 gRPC / 8000 REST)"]
        MAS[Multi-Agent System<br/>LangChain/LangGraph]
        AME[Alpha Mining Engine]
        FLE[Feedback Loop Engine]
        Research[Research Report Generator]
    end

    subgraph DataStores["Data Stores"]
        PG[(PostgreSQL)]
        Redis[(Redis)]
        S3[(S3 / MinIO<br/>Parquet + PDFs)]
    end

    subgraph External["External Data Sources"]
        VCI[VCI Connector]
        KBS[KBS Connector]
        VND[VND Connector]
        MSN_C[MSN Connector<br/>Forex/Crypto/World Indices]
        GOLD_C[GOLD Connector<br/>SJC + BTMC]
        FMARKET_C[FMarket Connector<br/>Mutual Funds]
        LLM[LLM Providers<br/>OpenAI/Anthropic/Google]
    end

    UI -->|REST /api/*| GinRouter
    GinRouter --> Handlers --> Services --> Infra
    Infra --> DSR
    DSR --> VCI & KBS & VND
    Services --> MSN_C & GOLD_C & FMARKET_C
    Services -->|gRPC| MAS
    MAS --> AME
    MAS --> FLE
    MAS --> LLM
    AME -->|PyArrow/DuckDB| S3
    Services --> PG & Redis
    FLE --> PG
    Research --> S3
```

### Service Communication

| From | To | Protocol | Purpose |
|------|----|----------|---------|
| Frontend | Go Backend | REST (HTTP/JSON) | All user-facing API calls |
| Go Backend | Python AI Service | gRPC (protobuf) | Agent invocations, alpha mining, feedback |
| Go Backend | Python AI Service | REST (fallback) | When gRPC unavailable |
| Python AI Service | LLM Providers | HTTPS | LLM inference calls |
| Python AI Service | S3/MinIO | S3 API | Read Parquet, write reports/models |
| Go Backend | PostgreSQL | TCP (pgx) | Transactional data |
| Go Backend | Redis | TCP | Cache layer |
| Go Backend | S3/MinIO | S3 API | Parquet archival, PDF storage |

### Data Flow: Price Update

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant GB as Go Backend
    participant DSR as Data Source Router
    participant VCI as VCI API
    participant KBS as KBS API
    participant Redis as Redis Cache

    FE->>GB: GET /api/market/quote?symbols=VNM,FPT
    GB->>Redis: Check cache (15min TTL)
    alt Cache hit
        Redis-->>GB: Cached quotes
    else Cache miss
        GB->>DSR: Route to primary source
        DSR->>VCI: RealTimeQuotes(["VNM","FPT"])
        alt VCI success
            VCI-->>DSR: Quote data
        else VCI fails
            DSR->>KBS: RealTimeQuotes(["VNM","FPT"])
            KBS-->>DSR: Quote data
        end
        DSR-->>GB: Quotes
        GB->>Redis: Cache with 15min TTL
    end
    GB-->>FE: JSON response with quotes + freshness indicator
```

### Data Flow: AI Advisory

```mermaid
sequenceDiagram
    participant FE as Frontend
    participant GB as Go Backend
    participant PY as Python AI Service
    participant TA as Technical Analyst
    participant NA as News Analyst
    participant IA as Investment Advisor
    participant SB as Strategy Builder
    participant LLM as LLM Provider

    FE->>GB: POST /api/chat {symbol: "VNM", message: "Phân tích VNM"}
    GB->>GB: Fetch market data (OHLCV, financials, news)
    GB->>PY: gRPC AnalyzeStock(symbol, market_data)
    par Parallel Agent Execution
        PY->>TA: Compute indicators + signals
        TA-->>PY: TechnicalAnalysis
        PY->>NA: Analyze news + sentiment
        NA->>LLM: Summarize news
        LLM-->>NA: Summary
        NA-->>PY: NewsAnalysis
    end
    PY->>IA: Synthesize(TechnicalAnalysis, NewsAnalysis, portfolio, sector)
    IA->>LLM: Generate recommendation
    LLM-->>IA: Recommendation
    IA-->>PY: InvestmentRecommendation
    PY->>SB: BuildStrategy(recommendation, regime)
    SB-->>PY: TradingStrategy
    PY-->>GB: AgentResponse with citations
    GB-->>FE: JSON response with structured analysis
```

### Data Flow: Alpha Mining Pipeline

```mermaid
sequenceDiagram
    participant Scheduler as Cron Scheduler
    participant DL as Data Layer
    participant ML as Model Layer
    participant BL as Backtest Layer
    participant Deploy as Deployment Layer
    participant S3 as S3 Parquet
    participant KB as Knowledge Base

    Scheduler->>DL: Daily refresh (after 15:00 ICT)
    DL->>S3: Read historical signals
    DL->>DL: Compute signal space (price, volume, fundamental, macro)
    DL->>DL: Z-score normalize signals
    DL->>S3: Write updated signal space (Parquet)

    Scheduler->>ML: Weekly retrain
    ML->>S3: Read signal space
    ML->>ML: Train gradient boosting + random forest
    ML->>ML: Regime-aware feature selection
    ML->>S3: Save model artifacts (versioned)

    Scheduler->>BL: Walk-forward backtest
    BL->>S3: Read signals + model predictions
    BL->>BL: Rolling backtest with transaction costs
    BL->>BL: Regime-aware validation
    BL->>S3: Write backtest results

    Deploy->>Deploy: Strategy ensemble (consensus voting)
    Deploy->>Deploy: Alpha decay monitoring
    Deploy->>KB: Store signal outcomes
```

## Components and Interfaces

### Monorepo Root Layout

Following best practices for a microservices monorepo:

```
ezistock/
├── proto/                             # Shared protobuf definitions (single source of truth)
│   └── ezistock/
│       ├── agent.proto
│       ├── alpha.proto
│       └── feedback.proto
├── backend/                           # Go Backend service
├── ai-service/                        # Python AI service
├── frontend/                          # Next.js frontend
├── infra/                             # Infrastructure-as-code
│   ├── terraform/                     # AWS provisioning
│   ├── k8s/                           # Kubernetes manifests / Helm charts
│   └── docker/                        # Shared Docker configs
├── docker-compose.yml                 # Local development
├── docker-compose.test.yml            # E2E test environment
├── Makefile                           # Top-level commands
├── .github/workflows/                 # CI/CD pipelines
└── README.md
```

### Go Backend Project Layout

Domain-driven organization following the official Go [Server project layout](https://go.dev/doc/modules/layout). Each domain is self-contained for clean boundaries and future extraction.

```
backend/
├── cmd/server/
│   └── main.go                        # Entry point: DI wiring, graceful shutdown
├── internal/
│   ├── domain/                        # Domain packages (extraction-ready)
│   │   ├── market/                    # Market data domain
│   │   │   ├── handler.go            # /api/market/* handlers
│   │   │   ├── service.go            # PriceService, MarketDataService
│   │   │   ├── sector.go             # SectorService
│   │   │   ├── macro.go              # MacroService
│   │   │   ├── search.go             # SearchService (⌘K)
│   │   │   ├── types.go
│   │   │   └── service_test.go
│   │   ├── portfolio/                 # Portfolio domain
│   │   │   ├── handler.go
│   │   │   ├── engine.go             # buy/sell/NAV
│   │   │   ├── ledger.go             # TransactionLedger
│   │   │   ├── performance.go        # TWR, XIRR
│   │   │   ├── risk.go               # Sharpe, VaR, beta
│   │   │   ├── corporate.go          # CorporateActionService
│   │   │   ├── export.go             # CSV/PDF
│   │   │   ├── types.go
│   │   │   └── engine_test.go
│   │   ├── screener/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── liquidity.go          # LiquidityFilter
│   │   │   └── types.go
│   │   ├── watchlist/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── agent/                     # AI proxy (delegates to Python)
│   │   │   ├── handler.go            # /api/chat, /api/analyze/*
│   │   │   ├── proxy.go              # gRPC proxy
│   │   │   └── types.go
│   │   ├── ranking/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── mission/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── notification/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── knowledge/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── research/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   ├── analyst/
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   └── types.go
│   │   └── auth/
│   │       ├── handler.go
│   │       ├── service.go
│   │       ├── middleware.go          # JWT, CORS, security headers
│   │       └── types.go
│   ├── platform/                      # Cross-cutting glue
│   │   ├── router.go                 # Gin router, imports all domain handlers
│   │   ├── middleware.go             # Logging, tracing, recovery
│   │   ├── config.go                 # App config from env vars
│   │   └── server.go                 # HTTP server lifecycle
│   ├── infra/                         # Infrastructure adapters (shared)
│   │   ├── postgres.go
│   │   ├── redis.go
│   │   ├── s3.go
│   │   ├── datasource.go            # DataSourceRouter (VCI/KBS/VND)
│   │   ├── circuit_breaker.go
│   │   ├── rate_limiter.go
│   │   ├── grpc.go                   # gRPC client to Python
│   │   ├── email.go
│   │   ├── scheduler.go
│   │   └── telemetry.go
│   └── generated/proto/              # Auto-generated protobuf stubs
├── migrations/
│   ├── 001_users.up.sql
│   ├── 001_users.down.sql
│   └── ...
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

**Dependency direction:** `cmd → platform → domain → infra` (acyclic, domains never import each other)

**Future extraction:** Move domain package → add `cmd/` entry point → replace calls with gRPC → deploy separately.

### Python AI Service Project Layout

Python `src` layout with domain separation:

```
ai-service/
├── src/ezistock_ai/
│   ├── __init__.py
│   ├── main.py                        # gRPC server + FastAPI fallback
│   ├── config.py                      # Pydantic Settings
│   ├── agents/
│   │   ├── orchestrator.py            # LangGraph workflow
│   │   ├── technical_analyst.py
│   │   ├── news_analyst.py
│   │   ├── investment_advisor.py
│   │   ├── strategy_builder.py
│   │   ├── base.py                    # BaseAgent abstract class
│   │   └── prompts/                   # Separated for easy tuning
│   │       ├── technical.py
│   │       ├── news.py
│   │       ├── advisor.py
│   │       └── strategy.py
│   ├── alpha/                         # Alpha Mining Engine
│   │   ├── data_layer.py
│   │   ├── model_layer.py
│   │   ├── backtest_layer.py
│   │   ├── deployment_layer.py
│   │   ├── regime_detector.py
│   │   └── decay_monitor.py
│   ├── feedback/
│   │   ├── engine.py
│   │   ├── accuracy.py
│   │   └── bias.py
│   ├── research/
│   │   ├── generator.py
│   │   ├── factor_snapshot.py
│   │   ├── sector_deepdive.py
│   │   └── pdf.py
│   ├── llm/
│   │   ├── router.py
│   │   ├── cost.py
│   │   └── cache.py
│   ├── infra/
│   │   ├── parquet.py
│   │   ├── s3.py
│   │   ├── postgres.py
│   │   └── telemetry.py
│   ├── grpc_server/
│   │   ├── agent_servicer.py
│   │   ├── alpha_servicer.py
│   │   └── feedback_servicer.py
│   └── generated/proto/
├── tests/
│   ├── conftest.py
│   ├── agents/
│   ├── alpha/
│   ├── feedback/
│   └── integration/
├── pyproject.toml
├── Dockerfile
└── Makefile
```

### Frontend Project Layout

Next.js App Router with feature-based organization:

```
frontend/src/
├── app/                               # Pages only (minimal logic)
│   ├── layout.tsx
│   ├── page.tsx                       # → /dashboard redirect
│   ├── globals.css
│   ├── (auth)/                        # No sidebar
│   │   ├── login/page.tsx
│   │   └── layout.tsx
│   ├── (app)/                         # With sidebar
│   │   ├── layout.tsx                 # AppShell
│   │   ├── dashboard/page.tsx
│   │   ├── stock/[symbol]/
│   │   │   ├── page.tsx
│   │   │   └── loading.tsx
│   │   ├── portfolio/page.tsx
│   │   ├── screener/page.tsx
│   │   ├── ranking/page.tsx
│   │   ├── ideas/page.tsx
│   │   ├── heatmap/page.tsx
│   │   ├── missions/page.tsx
│   │   ├── research/page.tsx
│   │   ├── macro/page.tsx
│   │   └── settings/page.tsx
│   └── terms/page.tsx
├── features/                          # Feature modules (co-located)
│   ├── dashboard/
│   │   ├── components/
│   │   └── hooks/
│   ├── stock/
│   │   ├── components/
│   │   └── hooks/
│   ├── chart/
│   │   ├── components/
│   │   └── lib/indicators/
│   ├── portfolio/
│   ├── screener/
│   ├── chat/
│   ├── heatmap/
│   └── watchlist/
├── shared/                            # Shared across features
│   ├── components/
│   │   ├── layout/                    # AppShell, Sidebar, Header, GlobalSearch
│   │   └── ui/                        # ErrorBoundary, Skeleton, FreshnessIndicator
│   ├── context/                       # App, Auth, Theme, I18n, Watchlist
│   ├── hooks/                         # usePolling, usePricePolling, useOnlineStatus
│   ├── lib/                           # api.ts, search-index.ts
│   └── i18n/                          # vi-VN.ts, en-US.ts
└── public/sw.js
```

**Import direction:** `app → features → shared` (features never import each other)

### Scaling Strategy

| Phase | Trigger | Action |
|---|---|---|
| Now | MVP | 3 services (Go + Python + Frontend) |
| Phase 2 | Mission scheduler contends with API | Extract `domain/mission/` → Mission Worker |
| Phase 3 | Market data rate limits | Extract `domain/market/` → Market Data Service |
| Phase 4 | 5+ developers | Split by team ownership |


### Go Backend — Key Interfaces and Structs

#### Data Source Router (infra/data_source_router.go)

```go
// SourcePreference defines primary and fallback sources for a data category.
type SourcePreference struct {
    Category  model.DataCategory
    Primary   string   // "VCI", "KBS", "VND", "ENTRADE", "CAFEF", "VND_FINFO"
    Fallbacks []string
}

// DataSourceRouter selects the optimal vnstock-go connector per data category
// with automatic failover, circuit breaking, and rate limiting.
// Uses vnstock-go typed error codes for smarter circuit breaker decisions.
type DataSourceRouter struct {
    clients       map[string]*vnstock.Client // keyed by connector name
    preferences   map[model.DataCategory]SourcePreference
    breakers      map[string]*gobreaker.CircuitBreaker[any]
    rateLimiters  map[string]*rate.Limiter
    cache         *infra.Cache
    logger        *slog.Logger
}

func NewDataSourceRouter(cache *infra.Cache, logger *slog.Logger) (*DataSourceRouter, error)

// Route selects the best source for a category, executing failover on error.
// Returns data from primary, fallback, or cache (with stale flag).
// Uses vnstock.RateLimited and vnstock.NetworkError to trigger circuit breaker;
// vnstock.NotSupported and vnstock.NoData do NOT trigger circuit breaker.
func (r *DataSourceRouter) Route(ctx context.Context, category model.DataCategory, fn func(*vnstock.Client) (any, error)) (any, bool, error)

// KBSClient returns the KBS connector for type-assertion to access KBS-specific methods.
func (r *DataSourceRouter) KBSClient() *vnstock.Client

// MSNClient returns the MSN connector for world indices, forex, crypto.
func (r *DataSourceRouter) MSNClient() *vnstock.Client

// FMarketClient returns the FMarket connector for mutual fund data.
func (r *DataSourceRouter) FMarketClient() *vnstock.Client
```

#### Price Service (service/price_service.go)

```go
type PriceService struct {
    router *infra.DataSourceRouter
    cache  *infra.Cache
}

func NewPriceService(router *infra.DataSourceRouter, cache *infra.Cache) *PriceService

// GetQuotes fetches real-time quotes for symbols via Data Source Router.
// Batches into single RealTimeQuotes call. Falls back to QuoteHistory (last 10 days).
// Caches with 15-minute TTL. Returns stale flag if serving cached data.
func (s *PriceService) GetQuotes(ctx context.Context, symbols []string) ([]model.StockQuote, bool, error)

// GetHistory fetches OHLCV history for a symbol with interval support.
func (s *PriceService) GetHistory(ctx context.Context, symbol string, start, end time.Time, interval string) ([]vnstock.Quote, error)
```

#### Market Data Service (service/market_data_service.go)

```go
type MarketDataService struct {
    router  *infra.DataSourceRouter
    cache   *infra.Cache
    storage *infra.Storage
}

func NewMarketDataService(router *infra.DataSourceRouter, cache *infra.Cache, storage *infra.Storage) *MarketDataService

func (s *MarketDataService) GetListing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, error)
func (s *MarketDataService) GetCompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error)
func (s *MarketDataService) GetOfficers(ctx context.Context, symbol string) ([]vnstock.Officer, error)
func (s *MarketDataService) GetFinancials(ctx context.Context, symbol, stmtType, period string) ([]vnstock.FinancialPeriod, error)
func (s *MarketDataService) GetFinancialRatios(ctx context.Context, symbol string) (vnstock.FinancialRatio, error)
func (s *MarketDataService) GetIndexCurrent(ctx context.Context, name string) (vnstock.IndexRecord, error)
func (s *MarketDataService) GetIndexHistory(ctx context.Context, name string, start, end time.Time, interval string) ([]vnstock.IndexRecord, error)

// KBS-specific data (uses type assertion on KBS connector)
func (s *MarketDataService) GetCompanyEvents(ctx context.Context, symbol string) ([]vnstock.CompanyEvent, error)
func (s *MarketDataService) GetCompanyNews(ctx context.Context, symbol string) ([]vnstock.CompanyNews, error)
func (s *MarketDataService) GetInsiderTrading(ctx context.Context, symbol string) ([]vnstock.InsiderTrade, error)
func (s *MarketDataService) GetSymbolsByGroup(ctx context.Context, groupCode string) (*vnstock.SymbolGroup, error)
func (s *MarketDataService) GetSymbolsByIndustry(ctx context.Context, industryCode string) (*vnstock.IndustryInfo, error)
```

#### Portfolio Engine (service/portfolio_engine.go)

```go
type PortfolioEngine struct {
    db     *infra.Database
    prices *PriceService
    ledger *TransactionLedger
}

func NewPortfolioEngine(db *infra.Database, prices *PriceService, ledger *TransactionLedger) *PortfolioEngine

func (e *PortfolioEngine) RecordBuy(ctx context.Context, userID string, tx model.Transaction) error
func (e *PortfolioEngine) RecordSell(ctx context.Context, userID string, tx model.Transaction) error
func (e *PortfolioEngine) GetHoldings(ctx context.Context, userID string) ([]model.Holding, error)
func (e *PortfolioEngine) ComputeNAV(ctx context.Context, userID string) (model.NAVResult, error)
func (e *PortfolioEngine) GetSectorAllocation(ctx context.Context, userID string) ([]model.SectorAllocation, error)
```

#### Performance Engine (service/performance_engine.go)

```go
type PerformanceEngine struct {
    db     *infra.Database
    prices *PriceService
}

func NewPerformanceEngine(db *infra.Database, prices *PriceService) *PerformanceEngine

func (e *PerformanceEngine) ComputeTWR(ctx context.Context, userID string, start, end time.Time) (float64, error)
func (e *PerformanceEngine) ComputeXIRR(ctx context.Context, userID string) (float64, error)
func (e *PerformanceEngine) GetEquityCurve(ctx context.Context, userID string, start, end time.Time) ([]model.NAVSnapshot, error)
func (e *PerformanceEngine) ComputeBenchmarkComparison(ctx context.Context, userID string, benchmark string, start, end time.Time) (model.BenchmarkComparison, error)
```

#### Risk Service (service/risk_service.go)

```go
type RiskService struct {
    db     *infra.Database
    prices *PriceService
}

func NewRiskService(db *infra.Database, prices *PriceService) *RiskService

func (s *RiskService) ComputeSharpe(ctx context.Context, userID string, riskFreeRate float64) (float64, error)
func (s *RiskService) ComputeMaxDrawdown(ctx context.Context, userID string) (float64, error)
func (s *RiskService) ComputeBeta(ctx context.Context, userID string, benchmark string) (float64, error)
func (s *RiskService) ComputeVolatility(ctx context.Context, userID string) (float64, error)
func (s *RiskService) ComputeVaR(ctx context.Context, userID string, confidence float64) (float64, error)
```

#### Screener Service (service/screener_service.go)

```go
type ScreenerService struct {
    marketData *MarketDataService
    liquidity  *LiquidityFilter
    db         *infra.Database
}

func NewScreenerService(md *MarketDataService, lf *LiquidityFilter, db *infra.Database) *ScreenerService

func (s *ScreenerService) Screen(ctx context.Context, filter model.ScreenerFilter) (model.ScreenerResult, error)
func (s *ScreenerService) SavePreset(ctx context.Context, userID string, preset model.FilterPreset) error
func (s *ScreenerService) ListPresets(ctx context.Context, userID string) ([]model.FilterPreset, error)
func (s *ScreenerService) DeletePreset(ctx context.Context, userID, presetID string) error
```

#### Sector Service (service/sector_service.go)

```go
type SectorService struct {
    router *infra.DataSourceRouter
    cache  *infra.Cache
}

func NewSectorService(router *infra.DataSourceRouter, cache *infra.Cache) *SectorService

func (s *SectorService) GetSectorPerformance(ctx context.Context, period string) ([]model.SectorPerformance, error)
func (s *SectorService) GetSectorTrend(ctx context.Context, sectorCode string) (model.SectorTrend, error)
func (s *SectorService) GetSectorMedianFundamentals(ctx context.Context, sectorCode string) (model.SectorFundamentals, error)
func (s *SectorService) GetStockSectorMapping(ctx context.Context) (map[string]string, error)
func (s *SectorService) GetSectorStocks(ctx context.Context, sectorCode string) ([]model.SectorStock, error)
```

#### Liquidity Filter (service/liquidity_filter.go)

```go
type LiquidityFilter struct {
    router *infra.DataSourceRouter
    cache  *infra.Cache
    config model.LiquidityConfig
}

func NewLiquidityFilter(router *infra.DataSourceRouter, cache *infra.Cache, config model.LiquidityConfig) *LiquidityFilter

// ComputeScore computes tradability score (0-100) for a symbol.
func (f *LiquidityFilter) ComputeScore(ctx context.Context, symbol string) (model.LiquidityScore, error)

// ClassifyTier returns Tier 1/2/3 based on score thresholds.
func (f *LiquidityFilter) ClassifyTier(score int) model.LiquidityTier

// FilterUniverse removes Tier 3 stocks from a symbol list.
func (f *LiquidityFilter) FilterUniverse(ctx context.Context, symbols []string, minTier model.LiquidityTier) ([]string, error)

// RefreshAll recomputes scores for all listed stocks (daily after 15:00 ICT).
func (f *LiquidityFilter) RefreshAll(ctx context.Context) error
```

#### Watchlist Service (service/watchlist_service.go)

```go
type WatchlistService struct {
    db    *infra.Database
    cache *infra.Cache
}

func NewWatchlistService(db *infra.Database, cache *infra.Cache) *WatchlistService

func (s *WatchlistService) Create(ctx context.Context, userID string, wl model.Watchlist) error
func (s *WatchlistService) List(ctx context.Context, userID string) ([]model.Watchlist, error)
func (s *WatchlistService) AddSymbol(ctx context.Context, userID, watchlistID string, entry model.WatchlistEntry) error
func (s *WatchlistService) RemoveSymbol(ctx context.Context, userID, watchlistID, symbol string) error
func (s *WatchlistService) SetAlert(ctx context.Context, userID, watchlistID, symbol string, alert model.PriceAlert) error
```

#### Mission Service (service/mission_service.go)

```go
type MissionService struct {
    db           *infra.Database
    notifier     *NotificationService
    prices       *PriceService
    screener     *ScreenerService
    grpcClient   *infra.GRPCClient
}

func NewMissionService(db *infra.Database, notifier *NotificationService, prices *PriceService, screener *ScreenerService, grpc *infra.GRPCClient) *MissionService

func (s *MissionService) Create(ctx context.Context, userID string, mission model.Mission) error
func (s *MissionService) List(ctx context.Context, userID string) ([]model.Mission, error)
func (s *MissionService) Pause(ctx context.Context, userID, missionID string) error
func (s *MissionService) Resume(ctx context.Context, userID, missionID string) error
func (s *MissionService) Delete(ctx context.Context, userID, missionID string) error

// EvaluateTriggers checks all active missions and fires notifications.
// Called by scheduler every 5 min during trading hours, 30 min outside.
func (s *MissionService) EvaluateTriggers(ctx context.Context) error
```

#### Notification Service (service/notification_service.go)

```go
type NotificationService struct {
    db    *infra.Database
    email *infra.EmailSender
}

func NewNotificationService(db *infra.Database, email *infra.EmailSender) *NotificationService

func (s *NotificationService) Send(ctx context.Context, userID string, notif model.Notification) error
func (s *NotificationService) List(ctx context.Context, userID string, limit int) ([]model.Notification, error)
func (s *NotificationService) MarkRead(ctx context.Context, userID, notifID string) error
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID string) (int, error)
```

#### Corporate Action Service (service/corporate_action_service.go)

```go
type CorporateActionService struct {
    router *infra.DataSourceRouter
    ledger *TransactionLedger
    cache  *infra.Cache
}

func NewCorporateActionService(router *infra.DataSourceRouter, ledger *TransactionLedger, cache *infra.Cache) *CorporateActionService

func (s *CorporateActionService) GetUpcomingEvents(ctx context.Context, symbols []string) ([]vnstock.CompanyEvent, error)
func (s *CorporateActionService) ProcessSplit(ctx context.Context, userID string, event vnstock.CompanyEvent) error
func (s *CorporateActionService) ProcessDividend(ctx context.Context, userID string, event vnstock.CompanyEvent) error
```

#### Knowledge Base (service/knowledge_base.go)

```go
type KnowledgeBase struct {
    db      *infra.Database
    storage *infra.Storage
}

func NewKnowledgeBase(db *infra.Database, storage *infra.Storage) *KnowledgeBase

func (kb *KnowledgeBase) RecordObservation(ctx context.Context, obs model.Observation) error
func (kb *KnowledgeBase) TrackOutcome(ctx context.Context, obsID string, interval string, priceChange float64) error
func (kb *KnowledgeBase) QuerySimilarPatterns(ctx context.Context, patternType string, minConfidence float64) ([]model.Observation, error)
func (kb *KnowledgeBase) GetAccuracyByPattern(ctx context.Context, patternType string) (model.PatternAccuracy, error)
func (kb *KnowledgeBase) ArchiveOldObservations(ctx context.Context, olderThan time.Time) error
```

#### Search Service (service/search_service.go)

```go
type SearchService struct {
    marketData *MarketDataService
    cache      *infra.Cache
    index      []model.SearchEntry // in-memory index
}

func NewSearchService(md *MarketDataService, cache *infra.Cache) *SearchService

// BuildIndex loads all VN stock symbols + company names into memory.
func (s *SearchService) BuildIndex(ctx context.Context) error

// Search performs fuzzy matching against the in-memory index. Must return within 200ms.
func (s *SearchService) Search(ctx context.Context, query string, limit int) ([]model.SearchResult, error)
```

#### Auth Service (service/auth_service.go)

```go
type AuthService struct {
    db *infra.Database
}

func NewAuthService(db *infra.Database) *AuthService

func (s *AuthService) Register(ctx context.Context, username, password string) error
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) // returns JWT
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*model.UserClaims, error)
func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error)
```

#### Storage Abstraction (infra/storage.go)

```go
// Storage abstracts S3/MinIO for Parquet files, PDFs, and model artifacts.
type Storage struct {
    bucket string
    client *s3.Client // AWS SDK v2
}

func NewStorage(bucket, endpoint string, usePathStyle bool) (*Storage, error)

func (s *Storage) PutParquet(ctx context.Context, key string, data []byte) error
func (s *Storage) GetParquet(ctx context.Context, key string) ([]byte, error)
func (s *Storage) PutPDF(ctx context.Context, key string, data []byte) error
func (s *Storage) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
func (s *Storage) PutModel(ctx context.Context, key string, data []byte) error
func (s *Storage) GetModel(ctx context.Context, key string) ([]byte, error)
```

### Python AI Service — Key Interfaces

#### Agent Orchestrator (app/agents/orchestrator.py)

```python
class AgentOrchestrator:
    """LangGraph workflow that orchestrates the multi-agent pipeline."""

    def __init__(self, config: Config):
        self.technical = TechnicalAnalystAgent(config)
        self.news = NewsAnalystAgent(config)
        self.advisor = InvestmentAdvisorAgent(config)
        self.strategy = StrategyBuilderAgent(config)
        self.llm_router = LLMRouter(config)
        self.cost_tracker = CostTracker(config)
        self.graph = self._build_graph()

    async def analyze_stock(
        self,
        symbol: str,
        market_data: MarketData,
        portfolio: Optional[Portfolio] = None,
        sector_context: Optional[SectorContext] = None,
        knowledge_history: Optional[list[Observation]] = None,
    ) -> AgentResponse:
        """Run full multi-agent pipeline for a stock. Timeout: 45s per agent."""
        ...

    async def generate_investment_ideas(
        self, universe: list[str], market_data: dict[str, MarketData]
    ) -> list[InvestmentIdea]:
        """Proactive scan of VN stock universe for opportunities."""
        ...

    async def chat(
        self, message: str, context: ChatContext
    ) -> ChatResponse:
        """Handle user chat with proactive suggestions and citations."""
        ...
```

#### Technical Analyst Agent (app/agents/technical_analyst.py)

```python
class TechnicalAnalystAgent:
    """Computes technical indicators and generates technical analysis."""

    def analyze(self, ohlcv: list[OHLCV]) -> TechnicalAnalysis:
        """Compute all indicators, detect patterns, classify signals."""
        ...

    def compute_indicators(self, ohlcv: list[OHLCV]) -> dict[str, list[float]]:
        """RSI, MACD, BB, SMA, EMA, ADX, Stochastic, ATR, OBV, MFI, etc."""
        ...

    def detect_patterns(self, ohlcv: list[OHLCV]) -> list[CandlestickPattern]:
        """Hammer, engulfing, doji, morning star, evening star, etc."""
        ...

    def classify_signal(self, indicators: dict) -> CompositeSignal:
        """Aggregate indicator signals into strongly_bullish..strongly_bearish."""
        ...

    def compute_smart_money_flow(self, foreign_flow: list, institutional_flow: list) -> SmartMoneyFlow:
        """Classify as strong_inflow..strong_outflow."""
        ...
```

#### Alpha Mining Engine (app/alpha_mining/)

```python
class DataLayer:
    """Constructs and maintains the multi-dimensional signal space."""

    def build_signal_space(self, symbols: list[str]) -> pd.DataFrame:
        """Build signal space: price, volume, fundamental, money flow, technical, macro."""
        ...

    def normalize_signals(self, df: pd.DataFrame) -> pd.DataFrame:
        """Z-score normalization within each signal category."""
        ...

class ModelLayer:
    """ML models for adaptive signal discovery."""

    def train(self, signal_space: pd.DataFrame, regime: MarketRegime) -> TrainedModel:
        """Train gradient boosting + random forest with regime-aware features."""
        ...

    def predict(self, model: TrainedModel, signals: pd.DataFrame) -> pd.DataFrame:
        """Generate signal importance scores and stock predictions."""
        ...

class BacktestLayer:
    """Walk-forward backtesting with regime-aware validation."""

    def run_backtest(self, signals: pd.DataFrame, config: BacktestConfig) -> BacktestResult:
        """Rolling backtest with transaction costs (0.15-0.25% commission)."""
        ...

    def detect_alpha_decay(self, signal_history: pd.DataFrame, threshold: float) -> list[DecayedSignal]:
        """Monitor signal predictive power degradation."""
        ...

class DeploymentLayer:
    """Strategy ensemble with consensus voting."""

    def ensemble_rank(self, strategies: list[StrategyResult], min_agreement: int = 2) -> list[RankedStock]:
        """Consensus-based stock ranking requiring min_agreement strategies."""
        ...

class RegimeDetector:
    """Classifies Vietnamese market regime."""

    def detect(self, vn_index: list[float], breadth: float, foreign_flow: float, volatility: float) -> MarketRegime:
        """Returns: bull, bear, sideways, risk_on, risk_off."""
        ...
```

#### Feedback Loop Engine (app/feedback_loop/engine.py)

```python
class FeedbackLoopEngine:
    """Closes the loop: recommendations → outcomes → model improvement."""

    def compute_agent_accuracy(self, agent_name: str, window_days: int = 30) -> AgentAccuracy:
        """Rolling accuracy score for an agent based on outcome tracking."""
        ...

    def detect_biases(self, agent_name: str) -> list[BiasReport]:
        """Detect systematic biases (e.g., sector overestimation)."""
        ...

    def generate_context_injection(self, agent_name: str) -> str:
        """Generate accuracy context string for agent prompt injection."""
        ...

    def should_reduce_weight(self, agent_name: str) -> bool:
        """True if 30-day accuracy < 40%. Restore when > 50%."""
        ...
```

### gRPC Proto Definitions

```protobuf
// proto/ezistock/agent.proto
syntax = "proto3";
package ezistock;

service AgentService {
  rpc AnalyzeStock(AnalyzeStockRequest) returns (AnalyzeStockResponse);
  rpc GenerateInvestmentIdeas(IdeaRequest) returns (IdeaResponse);
  rpc Chat(ChatRequest) returns (ChatResponse);
  rpc GetHotTopics(HotTopicsRequest) returns (HotTopicsResponse);
}

message AnalyzeStockRequest {
  string symbol = 1;
  MarketData market_data = 2;
  Portfolio portfolio = 3;
  SectorContext sector_context = 4;
  repeated Observation knowledge_history = 5;
}

message AnalyzeStockResponse {
  TechnicalAnalysis technical = 1;
  NewsAnalysis news = 2;
  InvestmentRecommendation recommendation = 3;
  TradingStrategy strategy = 4;
  repeated Citation citations = 5;
  string disclaimer = 6;
}

message ChatRequest {
  string user_id = 1;
  string message = 2;
  repeated ChatMessage history = 3; // last 10 messages
  Portfolio portfolio = 4;
}

message ChatResponse {
  string response = 1;
  repeated Citation citations = 2;
  repeated ProactiveSuggestion suggestions = 3;
  string disclaimer = 4;
}

// proto/ezistock/alpha.proto
service AlphaMiningService {
  rpc GetRanking(RankingRequest) returns (RankingResponse);
  rpc RunBacktest(BacktestRequest) returns (BacktestResponse);
  rpc GetRegime(RegimeRequest) returns (RegimeResponse);
}

// proto/ezistock/feedback.proto
service FeedbackService {
  rpc GetAgentAccuracy(AccuracyRequest) returns (AccuracyResponse);
  rpc GetModelPerformance(ModelPerfRequest) returns (ModelPerfResponse);
}
```

### REST API Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| POST | /api/auth/login | auth | Login, returns JWT |
| POST | /api/auth/register | auth | Register new user |
| GET | /api/health | health | Health check (Go + Python) |
| GET | /api/healthz | health | Liveness probe |
| GET | /api/readyz | health | Readiness probe |
| GET | /api/docs | swagger | Swagger UI |
| GET | /api/docs/swagger.json | swagger | OpenAPI spec |
| **Market Data** | | | |
| GET | /api/market/quote | market | Real-time quotes |
| GET | /api/market/chart | market | OHLCV history |
| GET | /api/market/listing | market | Stock listings |
| GET | /api/market/index | market | Index current/history |
| GET | /api/market/company/:symbol | market | Company profile |
| GET | /api/market/finance/:symbol | market | Financial statements |
| GET | /api/market/ratios/:symbol | market | Financial ratios |
| GET | /api/market/events/:symbol | market | Corporate events (KBS) |
| GET | /api/market/news/:symbol | market | Company news (KBS) |
| GET | /api/market/insider/:symbol | market | Insider trading (KBS) |
| GET | /api/market/statistics | market | Market statistics |
| GET | /api/market/macro | macro | Macro indicators |
| GET | /api/market/search | search | Global search ⌘K |
| **Sectors** | | | |
| GET | /api/sectors | sectors | All sector performance |
| GET | /api/sectors/:code | sectors | Sector detail + stocks |
| GET | /api/sectors/heatmap | sectors | Heatmap data |
| **Portfolio** | | | |
| GET | /api/portfolio | portfolio | Holdings + NAV |
| POST | /api/portfolio/buy | portfolio | Record buy |
| POST | /api/portfolio/sell | portfolio | Record sell |
| GET | /api/portfolio/transactions | portfolio | Transaction history |
| GET | /api/portfolio/performance | portfolio | TWR, XIRR, equity curve |
| GET | /api/portfolio/risk | portfolio | Risk metrics |
| GET | /api/portfolio/allocation | portfolio | Sector allocation |
| **Screener** | | | |
| POST | /api/screener | screener | Run screen with filters |
| GET | /api/screener/presets | screener | List saved presets |
| POST | /api/screener/presets | screener | Save preset |
| DELETE | /api/screener/presets/:id | screener | Delete preset |
| **Watchlists** | | | |
| GET | /api/watchlists | watchlists | List watchlists |
| POST | /api/watchlists | watchlists | Create watchlist |
| PUT | /api/watchlists/:id | watchlists | Update watchlist |
| DELETE | /api/watchlists/:id | watchlists | Delete watchlist |
| POST | /api/watchlists/:id/alerts | watchlists | Set price alert |
| **AI & Ranking** | | | |
| POST | /api/chat | agent | AI chat |
| POST | /api/analyze/:symbol | agent | Deep stock analysis |
| GET | /api/ideas | ranking | Investment ideas |
| POST | /api/ranking | ranking | Run AI ranking |
| POST | /api/ranking/backtest | ranking | Backtest ranking config |
| GET | /api/ranking/factors | ranking | Available factor groups |
| **Missions** | | | |
| GET | /api/missions | missions | List missions |
| POST | /api/missions | missions | Create mission |
| PUT | /api/missions/:id | missions | Update mission |
| DELETE | /api/missions/:id | missions | Delete mission |
| POST | /api/missions/:id/pause | missions | Pause mission |
| POST | /api/missions/:id/resume | missions | Resume mission |
| **Notifications** | | | |
| GET | /api/notifications | notifications | List notifications |
| POST | /api/notifications/read/:id | notifications | Mark as read |
| POST | /api/notifications/read-all | notifications | Mark all read |
| GET | /api/notifications/unread-count | notifications | Unread count |
| **Knowledge & Research** | | | |
| GET | /api/knowledge | knowledge | Query knowledge base |
| GET | /api/knowledge/accuracy | knowledge | Pattern accuracy |
| GET | /api/research | research | List research reports |
| GET | /api/research/:id | research | Get report detail |
| GET | /api/research/:id/pdf | research | Download PDF |
| **Analyst IQ** | | | |
| GET | /api/analyst/:symbol | analyst | Analyst consensus |
| GET | /api/analyst/reports | analyst | Analyst reports |
| **Corporate Actions** | | | |
| GET | /api/corporate-actions | corporate_actions | Upcoming events |
| GET | /api/corporate-actions/:symbol | corporate_actions | Events for symbol |
| **Export** | | | |
| GET | /api/export/transactions | export | CSV transaction export |
| GET | /api/export/portfolio | export | CSV/PDF portfolio report |
| GET | /api/export/pnl | export | P&L summary |
| **Settings** | | | |
| GET | /api/settings | settings | User settings |
| PUT | /api/settings | settings | Update settings |
| **Model Performance** | | | |
| GET | /api/model-performance | knowledge | Agent accuracy dashboard |


## Data Models

### PostgreSQL Schema

```sql
-- Users and Auth
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    disclaimer_acknowledged BOOLEAN DEFAULT FALSE,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Portfolio Holdings
CREATE TABLE holdings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    quantity DECIMAL(18,4) NOT NULL,
    avg_cost DECIMAL(18,4) NOT NULL,
    total_dividends DECIMAL(18,4) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, symbol)
);

-- Transaction Ledger
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    tx_type VARCHAR(20) NOT NULL, -- 'buy', 'sell', 'dividend', 'split', 'bonus'
    quantity DECIMAL(18,4) NOT NULL,
    unit_price DECIMAL(18,4) NOT NULL,
    total_value DECIMAL(18,4) NOT NULL,
    realized_pnl DECIMAL(18,4),
    tx_date TIMESTAMPTZ NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_transactions_user_date ON transactions(user_id, tx_date DESC);

-- Watchlists
CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE watchlist_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    watchlist_id UUID REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    sort_order INT DEFAULT 0,
    alert_upper DECIMAL(18,4),
    alert_lower DECIMAL(18,4),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(watchlist_id, symbol)
);

-- Screener Filter Presets
CREATE TABLE filter_presets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    filters JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Knowledge Base (hot data — last 90 days)
CREATE TABLE observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    pattern_type VARCHAR(100) NOT NULL,
    detection_date TIMESTAMPTZ NOT NULL,
    confidence DECIMAL(5,2) NOT NULL,
    data_snapshot JSONB NOT NULL,
    price_at_detection DECIMAL(18,4) NOT NULL,
    outcome_1d DECIMAL(8,4),
    outcome_7d DECIMAL(8,4),
    outcome_14d DECIMAL(8,4),
    outcome_30d DECIMAL(8,4),
    agent_name VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_observations_pattern ON observations(pattern_type, detection_date DESC);
CREATE INDEX idx_observations_symbol ON observations(symbol, detection_date DESC);

-- AI Recommendations (audit trail)
CREATE TABLE recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    direction VARCHAR(10) NOT NULL, -- 'buy', 'sell', 'hold'
    entry_price DECIMAL(18,4),
    stop_loss DECIMAL(18,4),
    take_profit DECIMAL(18,4),
    confidence DECIMAL(5,2) NOT NULL,
    reasoning JSONB NOT NULL,
    input_snapshot JSONB NOT NULL,
    outcome_1d DECIMAL(8,4),
    outcome_7d DECIMAL(8,4),
    outcome_14d DECIMAL(8,4),
    outcome_30d DECIMAL(8,4),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_recommendations_symbol ON recommendations(symbol, created_at DESC);

-- Missions
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    trigger_type VARCHAR(50) NOT NULL, -- 'price_threshold', 'schedule', 'event', 'news'
    trigger_config JSONB NOT NULL,
    target_symbols TEXT[] NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- 'alert', 'report', 'agent_analysis'
    notification_pref VARCHAR(50) DEFAULT 'in_app', -- 'in_app', 'email', 'push'
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'paused', 'triggered'
    last_triggered_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_missions_user ON missions(user_id, status);

-- Notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- 'price_alert', 'mission', 'idea', 'research', 'system'
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    link VARCHAR(500),
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read, created_at DESC);

-- Corporate Actions Cache
CREATE TABLE corporate_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    ex_date DATE,
    record_date DATE,
    payment_date DATE,
    value DECIMAL(18,4),
    content TEXT,
    fetched_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_corp_actions_symbol ON corporate_actions(symbol, ex_date DESC);

-- Analyst Reports
CREATE TABLE analyst_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    analyst_name VARCHAR(100),
    brokerage VARCHAR(100),
    recommendation VARCHAR(20), -- 'strong_buy', 'buy', 'hold', 'sell', 'strong_sell'
    target_price DECIMAL(18,4),
    report_date DATE,
    accuracy_1m DECIMAL(5,2),
    accuracy_3m DECIMAL(5,2),
    accuracy_6m DECIMAL(5,2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Research Reports
CREATE TABLE research_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(300) NOT NULL,
    report_type VARCHAR(50) NOT NULL, -- 'factor_snapshot', 'sector_deepdive', 'market_outlook'
    content JSONB NOT NULL,
    pdf_s3_key VARCHAR(500),
    published_at TIMESTAMPTZ DEFAULT NOW()
);

-- Chat History
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL, -- 'user', 'assistant'
    content TEXT NOT NULL,
    citations JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_chat_user ON chat_messages(user_id, created_at DESC);

-- NAV Snapshots (daily)
CREATE TABLE nav_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    nav DECIMAL(18,4) NOT NULL,
    snapshot_date DATE NOT NULL,
    holdings_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, snapshot_date)
);

-- Liquidity Scores (daily cache)
CREATE TABLE liquidity_scores (
    symbol VARCHAR(20) PRIMARY KEY,
    score INT NOT NULL, -- 0-100
    tier VARCHAR(10) NOT NULL, -- 'tier1', 'tier2', 'tier3'
    avg_volume DECIMAL(18,2),
    avg_value DECIMAL(18,2),
    zero_volume_days INT,
    computed_at TIMESTAMPTZ DEFAULT NOW()
);

-- LLM Cost Tracking
CREATE TABLE llm_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    agent_name VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    input_tokens INT NOT NULL,
    output_tokens INT NOT NULL,
    estimated_cost_usd DECIMAL(10,6) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_llm_usage_user ON llm_usage(user_id, created_at DESC);

-- Notification Preferences
CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    in_app_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT FALSE,
    email_enabled BOOLEAN DEFAULT FALSE,
    quiet_start TIME, -- e.g., 22:00
    quiet_end TIME,   -- e.g., 07:00
    price_alerts BOOLEAN DEFAULT TRUE,
    mission_alerts BOOLEAN DEFAULT TRUE,
    idea_alerts BOOLEAN DEFAULT TRUE,
    research_alerts BOOLEAN DEFAULT TRUE
);
```

### S3 Storage Layout

```
ezistock-data/
├── ohlcv/
│   ├── daily/
│   │   ├── HOSE/
│   │   │   ├── VNM/2024.parquet
│   │   │   ├── FPT/2024.parquet
│   │   │   └── ...
│   │   ├── HNX/
│   │   └── UPCOM/
│   └── intraday/
│       └── {symbol}/{date}.parquet
├── signals/
│   ├── signal_space/
│   │   └── {date}.parquet          # Daily signal space snapshot
│   └── backtest/
│       └── {strategy_id}/{date}.parquet
├── models/
│   ├── alpha/
│   │   └── v{version}/model.pkl    # Versioned ML models
│   └── regime/
│       └── v{version}/model.pkl
├── knowledge/
│   └── archived/
│       └── {year}/{month}.parquet   # Archived observations (>90 days)
├── research/
│   └── reports/
│       └── {report_id}.pdf
├── exports/
│   └── {user_id}/
│       └── {export_id}.{csv|pdf}
└── audit/
    └── {year}/{month}/logs.parquet
```

### Docker Compose (Development)

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ezistock
      POSTGRES_USER: ezistock
      POSTGRES_PASSWORD: ezistock_dev
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ezistock"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - miniodata:/data

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://ezistock:ezistock_dev@postgres:5432/ezistock?sslmode=disable
      REDIS_URL: redis://redis:6379
      S3_ENDPOINT: http://minio:9000
      S3_BUCKET: ezistock-data
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
      S3_USE_PATH_STYLE: "true"
      PYTHON_AI_GRPC: ai-service:50051
      VNSTOCK_CONNECTOR: VCI
      JWT_SECRET: dev-secret-change-in-prod
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      minio:
        condition: service_started

  ai-service:
    build:
      context: ./ai-service
      dockerfile: Dockerfile
    ports:
      - "50051:50051"
      - "8000:8000"
    environment:
      S3_ENDPOINT: http://minio:9000
      S3_BUCKET: ezistock-data
      S3_ACCESS_KEY: minioadmin
      S3_SECRET_KEY: minioadmin
      DATABASE_URL: postgres://ezistock:ezistock_dev@postgres:5432/ezistock?sslmode=disable
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
      LLM_PROVIDER: openai
      LLM_MODEL_ANALYSIS: gpt-4o
      LLM_MODEL_EXTRACTION: gpt-4o-mini
    depends_on:
      postgres:
        condition: service_healthy
      minio:
        condition: service_started

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      NEXT_PUBLIC_API_URL: http://localhost:8080
    depends_on:
      - backend

volumes:
  pgdata:
  redisdata:
  miniodata:
```

### Kubernetes Deployment Overview

```mermaid
graph TB
    subgraph K8s["Kubernetes Cluster (EKS)"]
        subgraph Ingress["Ingress (ALB)"]
            TLS[TLS Termination]
        end

        subgraph Frontend["Frontend Deployment"]
            FE1[Next.js Pod 1]
            FE2[Next.js Pod 2]
        end

        subgraph Backend["Go Backend Deployment (HPA: 2-10)"]
            BE1[Go Pod 1]
            BE2[Go Pod 2]
            BEN[Go Pod N]
        end

        subgraph AI["Python AI Deployment (HPA: 1-5)"]
            AI1[Python Pod 1]
            AIN[Python Pod N]
        end

        subgraph Data["Managed Services"]
            RDS[(RDS PostgreSQL<br/>db.t3.small)]
            ElastiCache[(ElastiCache Redis)]
            S3[(S3 Bucket<br/>Intelligent-Tiering)]
        end

        TLS -->|/| FE1 & FE2
        TLS -->|/api/*| BE1 & BE2 & BEN
        BE1 & BE2 & BEN -->|gRPC| AI1 & AIN
        BE1 & BE2 & BEN --> RDS & ElastiCache & S3
        AI1 & AIN --> S3 & RDS
    end

    subgraph External
        LLM[LLM APIs]
        VCI[VCI API]
        KBS[KBS API]
    end

    BE1 --> VCI & KBS
    AI1 --> LLM
```

Key Kubernetes resources:
- Go Backend: Deployment + HPA (2-10 replicas, CPU target 70%) + Service (ClusterIP)
- Python AI: Deployment + HPA (1-5 replicas, CPU target 60%) + Service (ClusterIP, port 50051)
- Frontend: Deployment (2 replicas) + Service (ClusterIP)
- Ingress: ALB Ingress Controller with TLS, path-based routing
- ConfigMaps: per-service configuration
- Secrets: DB credentials, API keys, JWT secret
- CronJob: nightly archival (PostgreSQL → S3 Parquet), daily liquidity refresh


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system — essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

<!-- Acceptance Criteria Testing Prework (summarized):

Req 1 (Architecture): 1.1-1.5 are structural/deployment — not testable. 1.6 (graceful degradation when Python down) — yes, property. 1.7 (health check) — yes, example. 1.8 (Docker packaging) — no.

Req 2 (Data Routing): 2.3 (failover within 10s) — yes, property. 2.4 (fallback on incomplete data) — yes, property. 2.5 (return cached on all-fail) — yes, property. 2.6 (circuit breaker after 3 failures) — yes, property. 2.1-2.2, 2.7-2.8 — structural/config, no.

Req 3 (Price Service): 3.2 (fallback chain) — yes, property. 3.3 (cache 15min TTL) — yes, property. 3.4 (batch requests) — yes, property. 3.5 (retry + stale) — yes, property. Combined with Req 2 failover into single property.

Req 4 (Multi-Agent): 4.6 (partial failure tolerance) — yes, property. 4.8 (citation mechanism) — yes, property. 4.1-4.5, 4.7 — structural/config, no.

Req 5 (Technical Analyst): 5.1 (compute indicators) — yes, property (invariants on indicator computation). 5.4 (composite signal aggregation) — yes, property. 5.7 (smart money flow classification) — yes, property. 5.2-5.3, 5.5-5.6 — specific detection logic, examples/edge cases.

Req 9 (Alpha Data Layer): 9.2 (z-score normalization) — yes, property (invariant: mean≈0, std≈1). 9.3 (missing data handling) — yes, property.

Req 10 (Alpha Model Layer): 10.4 (regime detection) — yes, property. 10.6 (signal importance scores) — yes, property.

Req 11 (Alpha Backtest): 11.4 (transaction cost simulation) — yes, property. 11.5 (backtest metrics computation) — yes, property. 11.6 (alpha decay detection) — yes, property.

Req 12 (Deployment Layer): 12.2 (consensus threshold) — yes, property. 12.3 (consensus ranking) — yes, property. 12.4 (alpha decay monitor) — yes, property.

Req 13 (AI Ranking): 13.5 (factor score computation) — yes, property.

Req 14 (Investment Ideas): 14.6 (deduplication) — yes, property.

Req 19 (Portfolio): 19.1 (buy credits holding) — yes, property. 19.2 (sell debits + P&L) — yes, property. 19.3 (unrealized P&L) — yes, property. 19.4 (NAV = sum of holdings) — yes, property. 19.7 (reject oversell) — yes, property.

Req 25 (Watchlists): 25.1-25.4 — CRUD, examples. 25.3 (alert trigger) — yes, property.

Req 26 (Knowledge Base): 26.3 (query similar patterns) — yes, property. 26.6 (aggregate accuracy) — yes, property.

Req 27 (Corporate Actions): 27.2 (auto-adjust cost basis on split) — yes, property.

Req 29 (Auth): 29.5 (account lockout after 5 failures) — yes, property.

Req 34 (Rate Limiting): 34.1-34.3 (rate limit + queue bound) — yes, property.

Req 37 (Global Search): 37.4 (sub-200ms) — yes, property. 37.3 (fuzzy matching) — yes, property.

Req 39 (Liquidity Filter): 39.1 (tradability score 0-100) — yes, property. 39.2 (tier classification) — yes, property. 39.4 (exclude Tier 3 from ideas) — yes, property. 39.8 (reduced position for Tier 2) — yes, property.

Req 40 (Storage): 40.5 (archival moves old data) — yes, property (round-trip: archive then read from S3 = original).

Req 41 (Feedback Loop): 41.2 (rolling accuracy) — yes, property. 41.7 (bias detection) — yes, property. 41.10 (weight reduction below 40%) — yes, property.

Req 42 (Notifications): 42.8 (rate limit: max 1/symbol/channel/hour) — yes, property.

Req 47 (Security): 47.2 (parameterized queries) — structural, no. 47.3 (sanitize HTML) — yes, property.

Req 48 (LLM Cost): 48.2 (token budget enforcement) — yes, property.

Property Reflection — Consolidation:
- Req 2.3/2.4/2.5 + Req 3.2/3.5 → Combined into Property 1 (Data Source Failover)
- Req 19.1/19.2/19.3/19.4 → Combined into Property 2 (Portfolio NAV Consistency) 
- Req 19.7 is edge case of Property 2
- Req 4.6 → Property 3 (Agent Partial Failure)
- Req 12.2/12.3 → Combined into Property 5 (Strategy Ensemble Consensus)
- Req 11.6/12.4 → Combined into Property 4 (Alpha Decay Detection)
- Req 14.6 → Property 6 (Idea Deduplication)
- Req 34.1/34.3 → Property 7 (Rate Limit Queue Bound)
- Req 37.3/37.4 → Property 9 (Search Latency + Correctness)
- Req 39.1/39.2/39.4/39.8 → Combined into Property 11 (Liquidity Gate)
- Req 41.2/41.10 → Combined into Property 12 (Feedback Loop Convergence)
- Req 9.2 → Property 13 (Signal Normalization Invariant)
- Req 27.2 → Property 14 (Corporate Action Cost Basis)
- Req 29.5 → Property 15 (Auth Lockout)
- Req 48.2 → Property 16 (Token Budget)
- Req 42.8 → Property 17 (Notification Rate Limit)
- Req 47.3 → Property 18 (Input Sanitization)
- Req 4.8 → Property 10 (Citation Integrity)
- Req 5.4 → Property 8 (Composite Signal Aggregation)
-->

### Property 1: Data Source Failover Guarantees

*For any* data category request and any combination of source failures, if the primary source fails, the Data Source Router must return data from the next available fallback source; if all sources fail, it must return the last cached result with a stale indicator — never an empty response when cached data exists.

**Validates: Requirements 2.3, 2.4, 2.5, 3.2, 3.5**

### Property 2: Portfolio NAV Consistency

*For any* sequence of buy and sell transactions on a portfolio, the NAV must always equal the sum of (quantity × current price) for all holdings. A buy transaction must increase the holding quantity by the transaction quantity. A sell transaction must decrease the holding quantity and compute realized P&L using weighted average cost. A sell transaction whose quantity exceeds the current holding must be rejected.

**Validates: Requirements 19.1, 19.2, 19.3, 19.4, 19.7**

### Property 3: Agent Partial Failure Tolerance

*For any* invocation of the Multi-Agent System where at least one sub-agent (Technical Analyst, News Analyst, Investment Advisor, Strategy Builder) succeeds, the system must produce a valid response using available agent outputs — never a total failure when at least one agent succeeds. The response must note which agents were unavailable.

**Validates: Requirements 4.6**

### Property 4: Alpha Decay Detection Timeliness

*For any* deployed signal whose predictive power (measured by rolling hit rate) degrades below the configured threshold over the monitoring window (default 20 trading days), the Alpha Decay Monitor must flag it within one monitoring cycle and trigger rebalancing.

**Validates: Requirements 11.6, 12.4**

### Property 5: Strategy Ensemble Consensus Threshold

*For any* set of strategy results and any stock, the stock must only appear in the top consensus ranking if it meets the minimum strategy agreement threshold (default: 2 out of 3 strategies agree). The composite score must reflect the agreement level.

**Validates: Requirements 12.2, 12.3**

### Property 6: Investment Idea Deduplication

*For any* symbol and signal direction, no duplicate Investment Idea shall be generated within a 48-hour window unless the new confidence score exceeds the previous one by at least 10 points.

**Validates: Requirements 14.6**

### Property 7: Rate Limit Queue Bound

*For any* data source, when the request rate exceeds the configured limit, excess requests must be queued. The queue depth must never exceed the configured maximum (default 100). Requests arriving when the queue is full must be rejected immediately with a rate limit error.

**Validates: Requirements 34.1, 34.3**

### Property 8: Composite Technical Signal Aggregation

*For any* set of computed technical indicators for a stock, the composite signal classification (strongly bullish, bullish, neutral, bearish, strongly bearish) must be consistent with the count of bullish vs bearish individual indicator signals. If bullish count > bearish count, the composite must not be bearish or strongly bearish, and vice versa.

**Validates: Requirements 5.4**

### Property 9: Global Search Latency and Correctness

*For any* search query against the full VN stock universe (1500+ symbols), results must be returned within 200ms. For any query that is a substring of a stock symbol or company name, the matching stock must appear in the results.

**Validates: Requirements 37.3, 37.4**

### Property 10: Citation Integrity

*For any* agent response that references a data point (price, indicator value, financial metric, news article), the citation must link to a valid, verifiable source on the platform. No claim referencing specific data may exist without an accompanying citation.

**Validates: Requirements 4.8**

### Property 11: Liquidity Gate

*For any* Investment Idea, AI Ranking result, or Strategy Builder recommendation, no Tier 3 (illiquid, score < 40) stock shall ever be included. Tier 2 stocks (score 40-69) must always carry a liquidity warning. The Strategy Builder must reduce maximum suggested position size for Tier 2 stocks by 50% compared to Tier 1.

**Validates: Requirements 39.4, 39.7, 39.8, 39.9**

### Property 12: Feedback Loop Convergence

*For any* agent whose rolling 30-day accuracy drops below 40%, the Feedback Loop Engine must reduce that agent's synthesis weight within one feedback cycle. The weight must not be restored until accuracy recovers above 50%.

**Validates: Requirements 41.2, 41.10**

### Property 13: Signal Space Normalization Invariant

*For any* signal category in the Signal Space after z-score normalization, the mean of all values within that category must be approximately 0 (within ε = 0.01) and the standard deviation must be approximately 1 (within ε = 0.05).

**Validates: Requirements 9.2**

### Property 14: Corporate Action Cost Basis Adjustment

*For any* stock split event with ratio R applied to a holding with quantity Q and average cost C, the resulting holding must have quantity Q×R and average cost C/R, preserving the total cost basis (Q×C = Q×R × C/R).

**Validates: Requirements 27.2**

### Property 15: Authentication Lockout

*For any* user account, if 5 failed login attempts occur within 15 minutes, the account must be locked for 30 minutes. During lockout, valid credentials must still be rejected.

**Validates: Requirements 29.5**

### Property 16: LLM Token Budget Enforcement

*For any* user, the cumulative token usage (input + output) across all LLM calls within a 24-hour window must not exceed the configured daily budget. When the budget is reached, subsequent LLM requests must be rejected with a budget exceeded error.

**Validates: Requirements 48.2**

### Property 17: Notification Rate Limiting

*For any* symbol and notification channel, at most 1 price alert notification may be sent per hour. For any user, at most 10 email notifications may be sent per day.

**Validates: Requirements 42.8**

### Property 18: Input Sanitization

*For any* user-provided text field (watchlist names, mission names, chat messages), after sanitization, the stored value must contain no HTML tags or script elements. The sanitized text must preserve the non-HTML content of the original input.

**Validates: Requirements 47.3**

### Property 19: Circuit Breaker State Transitions

*For any* data source, after 3 consecutive failures within 60 seconds, the circuit breaker must open and skip that source for subsequent requests. The circuit breaker must eventually transition to half-open to test recovery.

**Validates: Requirements 2.6**

### Property 20: Data Archival Round-Trip

*For any* set of OHLCV records or knowledge base observations archived from PostgreSQL to S3 Parquet, reading the archived data back must produce records equivalent to the originals.

**Validates: Requirements 40.5**


## Error Handling

### Go Backend Error Strategy

The Go Backend uses a layered error handling approach:

1. **vnstock-go errors**: All vnstock-go calls return `*vnstock.Error` with typed codes (`NetworkError`, `HTTPError`, `NotFound`, `NotSupported`, `InvalidInput`, `NoData`, `SerialiseError`, `ConfigError`). The Data Source Router inspects these codes to decide failover vs reject.

2. **Service-layer errors**: Each service returns domain-specific errors wrapped with context. Error types:
   - `ErrNotFound` — resource not found (404)
   - `ErrValidation` — invalid input (400)
   - `ErrConflict` — duplicate or constraint violation (409)
   - `ErrRateLimit` — rate limit exceeded (429)
   - `ErrBudgetExceeded` — LLM token budget exceeded (429)
   - `ErrServiceUnavailable` — Python AI Service down (503)
   - `ErrStaleData` — serving cached data, source unavailable (200 with stale flag)

3. **Handler-layer mapping**: Gin middleware maps service errors to HTTP status codes with structured JSON error responses:
   ```json
   {
     "error": {
       "code": "RATE_LIMIT_EXCEEDED",
       "message": "Request rate limit exceeded for VCI source",
       "retry_after": 5
     }
   }
   ```

4. **Circuit breaker integration**: The `gobreaker/v2` circuit breaker wraps each data source. States: Closed → Open (after 3 failures in 60s) → Half-Open (after 30s cooldown). Open state returns `ErrServiceUnavailable` immediately.

5. **Graceful degradation**: When the Python AI Service is unreachable:
   - Market data, portfolio, screener, watchlist endpoints work normally
   - AI-dependent endpoints (chat, analyze, ideas, ranking) return `503` with `"ai_service_unavailable": true`
   - Frontend shows "AI service temporarily unavailable" placeholder

### Python AI Service Error Strategy

1. **Agent timeout**: Each agent has a 30-second timeout. On timeout, the orchestrator proceeds with available outputs and marks the timed-out agent as `"status": "timeout"` in the response.

2. **LLM errors**: LLM API failures (rate limit, server error) trigger retry with exponential backoff (max 3 retries). If all retries fail, the agent returns a structured error with `"llm_error": true`.

3. **Cost control**: Token budget exceeded returns gRPC status `RESOURCE_EXHAUSTED` with remaining budget info.

4. **Data errors**: Missing or corrupt Parquet files in S3 return gRPC status `NOT_FOUND` or `DATA_LOSS`.

### Frontend Error Strategy

1. **Service worker cache**: On network failure, serve cached responses with "Offline — showing cached data" banner.
2. **Optimistic updates**: Write operations (add to watchlist, create mission) update UI immediately, retry on failure, rollback on permanent failure.
3. **Connection status**: Green (all services healthy), Yellow (partial degradation), Red (backend unreachable).
4. **Stale data indicators**: Price freshness badge (green < 1min, yellow 1-5min, red > 5min).

## Testing Strategy

### Dual Testing Approach

The platform uses both unit tests and property-based tests for comprehensive coverage:

- **Unit tests**: Verify specific examples, edge cases, integration points, and error conditions
- **Property-based tests**: Verify universal properties across randomly generated inputs (minimum 100 iterations per property)

Both are complementary — unit tests catch concrete bugs at specific values, property tests verify general correctness across the input space.

### Go Backend Testing

**Framework**: Standard `go test` + `pgregory.net/rapid` for property-based testing

**Property-based tests** (each maps to a design property):

```go
// Tag format: Feature: unified-finance-platform, Property {N}: {title}

// Feature: unified-finance-platform, Property 1: Data Source Failover Guarantees
func TestProperty_DataSourceFailover(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random data category and source failure combinations
        // Assert: never empty response when cache exists
    })
}

// Feature: unified-finance-platform, Property 2: Portfolio NAV Consistency
func TestProperty_PortfolioNAVConsistency(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random sequence of buy/sell transactions
        // Assert: NAV = sum(quantity * price) for all holdings
        // Assert: sell > holding quantity is rejected
    })
}

// Feature: unified-finance-platform, Property 5: Strategy Ensemble Consensus
func TestProperty_StrategyEnsembleConsensus(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random strategy results for random stocks
        // Assert: top ranking only includes stocks with >= min_agreement
    })
}

// Feature: unified-finance-platform, Property 6: Investment Idea Deduplication
func TestProperty_IdeaDeduplication(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random ideas with same symbol/direction within 48h
        // Assert: duplicates rejected unless confidence delta >= 10
    })
}

// Feature: unified-finance-platform, Property 7: Rate Limit Queue Bound
func TestProperty_RateLimitQueueBound(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random request bursts exceeding rate limit
        // Assert: queue depth never exceeds max (100)
    })
}

// Feature: unified-finance-platform, Property 9: Global Search Latency and Correctness
func TestProperty_GlobalSearchLatency(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random search queries (substrings of known symbols)
        // Assert: results returned within 200ms
        // Assert: matching symbol appears in results
    })
}

// Feature: unified-finance-platform, Property 11: Liquidity Gate
func TestProperty_LiquidityGate(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random stocks with random liquidity scores
        // Assert: Tier 3 never in recommendations
        // Assert: Tier 2 always has warning + 50% position reduction
    })
}

// Feature: unified-finance-platform, Property 13: Signal Normalization Invariant
func TestProperty_SignalNormalization(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random signal values
        // After z-score: assert mean ≈ 0, std ≈ 1
    })
}

// Feature: unified-finance-platform, Property 14: Corporate Action Cost Basis
func TestProperty_CorporateActionCostBasis(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random holdings and split ratios
        // Assert: total cost basis preserved (Q*C = Q*R * C/R)
    })
}

// Feature: unified-finance-platform, Property 15: Authentication Lockout
func TestProperty_AuthLockout(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random sequences of failed login attempts
        // Assert: locked after 5 failures in 15 min
        // Assert: valid creds rejected during lockout
    })
}

// Feature: unified-finance-platform, Property 17: Notification Rate Limiting
func TestProperty_NotificationRateLimit(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random notification bursts per symbol/channel
        // Assert: max 1 per symbol per channel per hour
        // Assert: max 10 emails per user per day
    })
}

// Feature: unified-finance-platform, Property 18: Input Sanitization
func TestProperty_InputSanitization(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random strings with embedded HTML/script tags
        // Assert: sanitized output contains no HTML tags
        // Assert: non-HTML content preserved
    })
}

// Feature: unified-finance-platform, Property 19: Circuit Breaker State Transitions
func TestProperty_CircuitBreakerTransitions(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random sequences of success/failure calls
        // Assert: opens after 3 consecutive failures in 60s
        // Assert: transitions to half-open after cooldown
    })
}

// Feature: unified-finance-platform, Property 20: Data Archival Round-Trip
func TestProperty_DataArchivalRoundTrip(t *testing.T) {
    rapid.Check(t, func(t *rapid.T) {
        // Generate random OHLCV records
        // Write to Parquet, read back
        // Assert: records are equivalent
    })
}
```

**Unit tests** (specific examples and edge cases):
- Data Source Router: specific failover scenarios, circuit breaker edge cases
- Portfolio Engine: specific buy/sell sequences, dividend recording, zero-quantity edge case
- Screener Service: default presets, max 10 presets per user, missing data exclusion
- Auth Service: JWT token generation/validation, password hashing, session expiry
- Export Service: CSV format correctness, PDF generation, tax calculation (0.1% on sell)
- Mission Service: max 20 missions per user, trigger evaluation timing

**Integration tests**: Using `testcontainers-go` with PostgreSQL for database-dependent services.

### Python AI Service Testing

**Framework**: `pytest` + `hypothesis` for property-based testing

**Property-based tests**:

```python
# Feature: unified-finance-platform, Property 3: Agent Partial Failure Tolerance
@given(agent_failures=st.sets(st.sampled_from(["technical", "news", "advisor", "strategy"])))
def test_property_agent_partial_failure(agent_failures):
    # At least one agent must succeed
    assume(len(agent_failures) < 4)
    # Assert: response produced with available agents
    # Assert: failed agents noted in response

# Feature: unified-finance-platform, Property 4: Alpha Decay Detection
@given(signal_history=st.lists(st.floats(0, 1), min_size=20))
def test_property_alpha_decay_detection(signal_history):
    # If rolling hit rate < threshold, must be flagged
    # Assert: flagged within one monitoring cycle

# Feature: unified-finance-platform, Property 8: Composite Signal Aggregation
@given(indicators=st.dictionaries(
    st.sampled_from(INDICATOR_NAMES),
    st.sampled_from(["bullish", "bearish", "neutral"])
))
def test_property_composite_signal(indicators):
    # Assert: composite consistent with bullish/bearish count

# Feature: unified-finance-platform, Property 12: Feedback Loop Convergence
@given(accuracy_history=st.lists(st.floats(0, 1), min_size=30))
def test_property_feedback_convergence(accuracy_history):
    # If 30-day rolling avg < 0.4, weight must be reduced
    # Weight restored only when > 0.5

# Feature: unified-finance-platform, Property 16: LLM Token Budget
@given(token_counts=st.lists(st.integers(100, 10000), min_size=1))
def test_property_token_budget(token_counts):
    # Assert: cumulative never exceeds daily budget
    # Assert: rejected when budget reached
```

**Unit tests**:
- Agent orchestrator: specific stock analysis scenarios, timeout handling
- Alpha Mining: specific signal space construction, model training with known data
- Research generator: report structure validation, PDF output
- LLM router: model selection for different task types
- Cost tracker: cost calculation accuracy for different models

### Frontend Testing

**Framework**: Vitest + React Testing Library + Playwright (E2E)

**Unit/Component tests**:
- Chart Engine: indicator computation, drawing tool state
- Global Search: fuzzy matching accuracy, result ranking
- Heatmap: color mapping, sector grouping
- Notification Bell: unread count, mark-as-read
- Service Worker: cache hit/miss behavior

**E2E tests** (Playwright):
- Login → Dashboard → Search stock → View analysis → Add to watchlist
- Create mission → Verify trigger → Check notification
- Portfolio: Record buy → Record sell → Verify P&L
- Screener: Apply filters → Save preset → Load preset
- Chat: Send message → Verify structured response with citations

### Test Environment

All tests run against Docker Compose with:
- PostgreSQL (test database, reset between test suites)
- Redis (flushed between tests)
- MinIO (test bucket)
- Python AI Service in test mode (mocked LLM responses for deterministic E2E)

CI pipeline (GitHub Actions):
1. Go: `go test -race -count=1 ./...` (includes property tests with rapid)
2. Python: `pytest --hypothesis-seed=0` (deterministic property tests)
3. Frontend: `vitest --run` (unit) + `playwright test` (E2E)
4. All tests must pass before merge to main
5. E2E suite target: < 10 minutes

