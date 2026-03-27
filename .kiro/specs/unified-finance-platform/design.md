# Design Document

## Overview

MyFi is a unified personal finance platform that consolidates multiple asset types (VN stocks, gold, crypto, savings, bonds) into a single dashboard with near real-time price tracking, advanced charting with 21 technical indicators, and a multi-agent AI advisory system. The platform is built with a Go backend (Gin framework) using vnstock-go for Vietnamese stock market data, and a Next.js 16 frontend with TypeScript, Tailwind CSS, and lightweight-charts.

The system architecture follows a modular design with clear separation between data acquisition, business logic, and presentation layers. The backend provides RESTful APIs for all platform features, while the frontend implements a responsive single-page application with real-time updates and persistent state management.

Key architectural principles:
- Dual-source data routing with intelligent failover between VCI and KBS data sources
- Multi-agent AI system using langchaingo with specialized agents for price, analysis, news, monitoring, and supervision
- Autonomous pattern detection with knowledge base learning
- Comprehensive caching strategy with source-specific TTLs
- Rate limiting and circuit breaker patterns for external API resilience
- JWT-based authentication with 2FA and biometric support
- Multi-currency display with FX rate tracking
- Dark theme and multi-language support
- Real-time WebSocket price streaming with fallback to HTTP polling
- Progressive Web App (PWA) with offline capabilities
- Push notifications via Web Push API
- Email and SMS alert delivery
- Social features with community sentiment tracking
- Broker integration and CSV transaction import
- Portfolio rebalancing with tax-loss harvesting
- Voice input and natural language query processing
- Advanced tax optimization and reporting
- Performance optimization with Redis caching and CDN
- Compliance audit trail with immutable logging
- Advanced charting with pattern recognition
- Portfolio stress testing and Monte Carlo simulation
- Public API with webhook support
- External service integrations (Google Sheets, tax software)
- Customizable dashboard layouts with drag-and-drop
- Vietnamese news sentiment analysis
- Performance attribution analysis

## Architecture

### System Components

The platform consists of three primary layers:

**1. Data Layer**
- Data_Source_Router: Intelligent routing between VCI and KBS sources with failover logic
- Price_Service: Multi-asset price fetching with caching (stocks, crypto, gold)
- Market_Data_Service: Unified access to all vnstock-go data categories
- Sector_Service: ICB sector classification and performance tracking
- Gold_Service: Doji/SJC gold price integration
- FX_Service: USD/VND exchange rate tracking
- Commodity_Service: Commodity market data
- Macro_Service: Macroeconomic indicators
- Fund_Service: Open fund data

**2. Business Logic Layer**
- Portfolio_Engine: Asset holdings, transactions, cost basis, P&L calculations
- Performance_Engine: TWR, MWRR/XIRR, equity curve, benchmark comparison
- Risk_Service: Sharpe ratio, max drawdown, beta, volatility, VaR
- Screener_Service: Advanced stock filtering with fundamental and sector criteria
- Comparison_Engine: Multi-stock valuation, performance, and correlation analysis
- Watchlist_Service: Named watchlists with price alerts
- Savings_Tracker: Interest calculation for savings accounts
- Corporate_Action_Service: Dividend and stock split tracking
- Goal_Planner: Financial goal tracking and progress calculation
- Backtest_Engine: Strategy simulation against historical data
- Export_Service: CSV and PDF report generation

**3. AI Agent Layer**
- Multi_Agent_System: Orchestration framework using langchaingo
- Price_Agent: Asset price fetching for AI analysis
- Analysis_Agent: Technical analysis with 21 indicators, fundamental analysis, sector-relative analysis
- News_Agent: Financial news aggregation, summarization, and Vietnamese sentiment analysis
- Monitor_Agent: Autonomous pattern detection (accumulation, distribution, breakout)
- Supervisor_Agent: Recommendation synthesis with NAV-based advice and natural language query processing
- Knowledge_Base: Pattern observation storage and outcome tracking
- Alert_Service: Proactive notification delivery via in-app, push, email, and SMS
- Recommendation_Audit_Log: AI recommendation tracking and accuracy measurement
- Sentiment_Analyzer: Vietnamese NLP for news and community sentiment analysis

**4. Presentation Layer**
- Chart_Engine: TradingView-style charting with 21 technical indicators, drawing tools, pattern recognition, and chart templates
- Dashboard: NAV overview, asset allocation, alerts, quick metrics, customizable widget layouts
- Portfolio View: Holdings, transactions, P&L, performance analytics, rebalancing tools
- Screener View: Advanced stock filtering with sector context
- Comparison View: Multi-stock valuation, performance, correlation analysis
- Sector Trend Dashboard: Heatmap and performance visualization
- Chat Widget: Multi-agent AI advisory interface with voice input support
- Theme_Service: Light/dark mode management with auto-switching
- I18n_Service: Vietnamese/English localization
- Community_Dashboard: Leaderboards, public watchlists, sentiment indicators
- Tax_Dashboard: Tax optimization suggestions, tax reports, scenario modeling
- Educational_Content: Tutorials, glossary, strategy templates, onboarding flows
- Market_Calendar: Events, earnings dates, IPOs, economic releases
- Stress_Test_View: Historical stress testing, Monte Carlo simulation, scenario analysis

**5. Communication Layer**
- WebSocket_Server: Real-time price streaming with subscription management
- Push_Notification_Service: Web Push API integration for browser/mobile notifications
- Email_Service: SMTP/API integration for email alerts (SendGrid, AWS SES)
- SMS_Service: SMS gateway integration for critical alerts (Twilio, AWS SNS)
- Webhook_Service: Outbound webhook delivery with retry logic

**6. Integration Layer**
- CSV_Import_Service: Transaction import with broker format detection
- Broker_API_Connector: Direct integration with broker APIs (SSI, VPS, HSC, VCBS)
- Google_Sheets_Integration: OAuth-based export to Google Sheets
- Calendar_Integration: Sync corporate events to Google Calendar/Outlook
- Zapier_Integration: Custom automation workflows
- Public_API: REST API with API key authentication and rate limiting

**7. Security and Compliance Layer**
- Auth_Service: JWT authentication, 2FA (TOTP), biometric login (WebAuthn)
- Session_Manager: Multi-device session tracking and management
- Audit_Logger: Immutable audit trail for all user actions
- Suspicious_Activity_Detector: Anomaly detection for security events
- Data_Retention_Service: GDPR-compliant data export and deletion

**8. Performance and Reliability Layer**
- Redis_Cache: Distributed caching for frequently accessed data
- CDN_Service: Static asset delivery with edge caching
- Rate_Limiter: Per-user and per-IP rate limiting
- Circuit_Breaker: Fault tolerance for external API calls
- Load_Balancer: Request distribution across backend instances
- Database_Read_Replica: Read-heavy query distribution

### Project Directory Layout

The backend follows the official Go "Server project" layout from [go.dev/doc/modules/layout](https://go.dev/doc/modules/layout). All Go packages reside under `internal/` (private to the module) with the server entry point in `cmd/server/`.

```
backend/
├── go.mod                          # Module: myfi-backend
├── go.sum
├── cmd/
│   └── server/
│       └── main.go                 # Entry point: Gin router setup, CORS, route registration, dependency init
├── internal/
│   ├── handler/                    # HTTP handlers (accept *gin.Context, delegate to services)
│   │   ├── market.go              # Market data handlers (quotes, charts, screener, company, finance, trading, statistics, valuation, funds, commodities, macro)
│   │   ├── crypto.go              # Cryptocurrency handlers
│   │   ├── news.go                # News aggregation handlers
│   │   └── agent.go               # AI/LLM chat handlers
│   ├── service/                    # Business logic and domain engines
│   │   ├── price_service.go       # Multi-asset price fetching with caching
│   │   ├── crypto_service.go      # CoinGecko crypto data
│   │   ├── gold_service.go        # Doji/SJC gold prices
│   │   ├── fx_service.go          # USD/VND exchange rate
│   │   ├── savings_tracker.go     # Savings/term deposit interest calculation
│   │   ├── watchlist_service.go   # Named watchlists with price alerts
│   │   ├── screener_service.go    # Advanced stock filtering
│   │   ├── sector_service.go      # ICB sector classification and performance
│   │   ├── commodity_service.go   # Commodity market data
│   │   ├── fund_service.go        # Open fund data
│   │   ├── macro_service.go       # Macroeconomic indicators
│   │   ├── market_data_service.go # Unified vnstock-go data layer
│   │   ├── portfolio_engine.go    # Holdings, transactions, cost basis, P&L
│   │   ├── performance_engine.go  # TWR, MWRR/XIRR, equity curve, benchmarks
│   │   ├── comparison_engine.go   # Multi-stock valuation and correlation
│   │   ├── transaction_ledger.go  # Transaction recording and querying
│   │   └── asset_registry.go      # Asset type registry and metadata
│   ├── infra/                      # Infrastructure and cross-cutting concerns
│   │   ├── cache.go               # In-memory cache with TTL
│   │   ├── database.go            # PostgreSQL connection and migrations
│   │   ├── circuit_breaker.go     # Circuit breaker pattern for external APIs
│   │   ├── rate_limiter.go        # Per-source API rate limiting
│   │   └── data_source_router.go  # VCI/KBS intelligent source selection and failover
│   └── model/                      # Shared types, structs, constants, enums
│       └── types.go               # AssetType, TransactionType, ICBSector, DataCategory, request/response structs
├── testdata/                       # Test fixtures
└── myfi-backend                    # Compiled binary
```

**Package dependency direction** (acyclic):
- `cmd/server` → `handler`, `service`, `infra`, `model`
- `handler` → `service`, `model`
- `service` → `infra`, `model`
- `infra` → `model`
- `model` → (no internal imports)

**Key conventions:**
- Entry point (`cmd/server/main.go`) contains only wiring: router setup, CORS middleware, route registration, dependency initialization via constructor injection
- Handlers accept service dependencies via struct fields (no package-level globals)
- All internal packages use import paths like `myfi-backend/internal/handler`
- Run with `go run ./cmd/server` from the `backend/` directory
- Build with `go build ./cmd/server` from the `backend/` directory

### Data Flow

**Price Update Flow:**
1. Frontend polls Price_Service at configured intervals (15s stocks, 60s crypto, 300s gold)
2. Price_Service checks cache TTL (15min stocks, 5min crypto, 1hr gold)
3. On cache miss, Data_Source_Router selects optimal source (VCI primary, KBS fallback)
4. Rate_Limiter enforces per-source request limits
5. Circuit breaker prevents repeated failures
6. Price_Service updates cache and returns data
7. Frontend updates NAV, allocation, and holdings without page reload

**AI Advisory Flow:**
1. User sends query through chat widget
2. Supervisor_Agent parses query and determines required sub-agents
3. Sub-agents execute in parallel:
   - Price_Agent fetches current/historical prices via Data_Source_Router
   - Analysis_Agent computes 21 indicators, fundamental metrics, sector context
   - News_Agent fetches CafeF RSS and Google search results
4. Supervisor_Agent synthesizes outputs with portfolio context from Portfolio_Engine
5. Supervisor_Agent queries Knowledge_Base for relevant historical patterns
6. Supervisor_Agent produces structured recommendation with NAV-based advice
7. Recommendation_Audit_Log persists full inputs/outputs for outcome tracking
8. Frontend renders structured response with sections for data, analysis, news, advice

**Autonomous Monitoring Flow:**
1. Monitor_Agent runs on schedule (30min during trading hours, 2hr outside)
2. Monitor_Agent fetches OHLCV data for all watchlist symbols via Data_Source_Router
3. Pattern_Detector identifies accumulation, distribution, breakout patterns
4. Monitor_Agent generates observations with confidence scores
5. Knowledge_Base stores observations with supporting data snapshots
6. Alert_Service delivers notifications for high-confidence patterns (≥60)
7. Knowledge_Base tracks outcomes at 1d, 7d, 14d, 30d intervals
8. Historical accuracy metrics inform future pattern detection

### Technology Stack

**Backend:**
- Language: Go 1.26+
- Framework: Gin (HTTP router)
- Data Sources: vnstock-go (VCI/KBS), CoinGecko, Doji, SJC
- AI: langchaingo (OpenAI, Anthropic, Google, Qwen, Bedrock)
- Database: SQLite (dev), PostgreSQL (prod)
- Authentication: JWT, bcrypt

**Frontend:**
- Framework: Next.js 16 with App Router
- Language: TypeScript
- Styling: Tailwind CSS
- Charting: lightweight-charts
- State Management: React Context API
- HTTP Client: fetch with AbortSignal timeout

**External APIs:**
- VCI: https://trading.vietcap.com.vn/api (77 columns, primary source)
- KBS: Alternative stock data source (28 columns)
- CoinGecko: https://api.coingecko.com/api/v3 (crypto, FX rates)
- Doji: https://giavang.doji.vn/api/giavang/ (gold prices)
- SJC: https://sjc.com.vn/GoldPrice/Services/PriceService.ashx (gold fallback)

## Components and Interfaces

### Backend Modules

**Data_Source_Router**
```go
type DataCategory string
const (
    PriceQuotes DataCategory = "price_quotes"
    OHLCVHistory DataCategory = "ohlcv_history"
    IntradayData DataCategory = "intraday_data"
    OrderBook DataCategory = "order_book"
    CompanyOverview DataCategory = "company_overview"
    Shareholders DataCategory = "shareholders"
    Officers DataCategory = "officers"
    News DataCategory = "news"
    IncomeStatement DataCategory = "income_statement"
    BalanceSheet DataCategory = "balance_sheet"
    CashFlow DataCategory = "cash_flow"
    FinancialRatios DataCategory = "financial_ratios"
)

type SourcePreference struct {
    Category DataCategory
    Primary  string // "VCI" or "KBS"
    Fallback string
}

type DataSourceRouter struct {
    preferences map[DataCategory]SourcePreference
    vciClient   *vnstock.Client
    kbsClient   *vnstock.Client
    rateLimiter *RateLimiter
    circuitBreakers map[string]*CircuitBreaker
    cache       *Cache
}

func (r *DataSourceRouter) FetchData(ctx context.Context, category DataCategory, params map[string]interface{}) (interface{}, error)
func (r *DataSourceRouter) selectSource(category DataCategory) string
func (r *DataSourceRouter) fetchWithFallback(ctx context.Context, category DataCategory, params map[string]interface{}) (interface{}, error)
```

**Price_Service**
```go
type AssetType string
const (
    VNStock AssetType = "vn_stock"
    Crypto  AssetType = "crypto"
    Gold    AssetType = "gold"
)

type PriceQuote struct {
    Symbol        string    `json:"symbol"`
    AssetType     AssetType `json:"assetType"`
    Price         float64   `json:"price"`
    Change        float64   `json:"change"`
    ChangePercent float64   `json:"changePercent"`
    Volume        int64     `json:"volume"`
    Timestamp     time.Time `json:"timestamp"`
    Source        string    `json:"source"`
    IsStale       bool      `json:"isStale"`
}

type PriceService struct {
    router      *DataSourceRouter
    goldService *GoldService
    cache       *Cache
}

func (s *PriceService) GetQuotes(ctx context.Context, symbols []string, assetType AssetType) ([]PriceQuote, error)
func (s *PriceService) GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, interval string) ([]OHLCVBar, error)
func (s *PriceService) batchFetchStocks(ctx context.Context, symbols []string) ([]PriceQuote, error)
func (s *PriceService) fetchCrypto(ctx context.Context, symbols []string) ([]PriceQuote, error)
```

**Portfolio_Engine**
```go
type Asset struct {
    ID              int64     `json:"id"`
    UserID          int64     `json:"userId"`
    AssetType       AssetType `json:"assetType"`
    Symbol          string    `json:"symbol"`
    Quantity        float64   `json:"quantity"`
    AverageCost     float64   `json:"averageCost"`
    AcquisitionDate time.Time `json:"acquisitionDate"`
    Account         string    `json:"account"`
}

type Transaction struct {
    ID              int64           `json:"id"`
    UserID          int64           `json:"userId"`
    AssetType       AssetType       `json:"assetType"`
    Symbol          string          `json:"symbol"`
    Quantity        float64         `json:"quantity"`
    UnitPrice       float64         `json:"unitPrice"`
    TotalValue      float64         `json:"totalValue"`
    TransactionDate time.Time       `json:"transactionDate"`
    TransactionType TransactionType `json:"transactionType"`
}

type TransactionType string
const (
    Buy        TransactionType = "buy"
    Sell       TransactionType = "sell"
    Deposit    TransactionType = "deposit"
    Withdrawal TransactionType = "withdrawal"
    Interest   TransactionType = "interest"
    Dividend   TransactionType = "dividend"
)

type PortfolioSummary struct {
    NAV               float64                    `json:"nav"`
    NAVChange24h      float64                    `json:"navChange24h"`
    NAVChangePercent  float64                    `json:"navChangePercent"`
    AllocationByType  map[AssetType]float64      `json:"allocationByType"`
    AllocationPercent map[AssetType]float64      `json:"allocationPercent"`
    Holdings          []HoldingDetail            `json:"holdings"`
}

type HoldingDetail struct {
    Asset           Asset   `json:"asset"`
    CurrentPrice    float64 `json:"currentPrice"`
    MarketValue     float64 `json:"marketValue"`
    UnrealizedPL    float64 `json:"unrealizedPL"`
    UnrealizedPLPct float64 `json:"unrealizedPLPct"`
}

type PortfolioEngine struct {
    db           *sql.DB
    priceService *PriceService
}

func (e *PortfolioEngine) AddAsset(ctx context.Context, asset Asset) error
func (e *PortfolioEngine) RecordTransaction(ctx context.Context, tx Transaction) error
func (e *PortfolioEngine) GetPortfolioSummary(ctx context.Context, userID int64) (PortfolioSummary, error)
func (e *PortfolioEngine) ComputeNAV(ctx context.Context, userID int64) (float64, error)
func (e *PortfolioEngine) ComputeUnrealizedPL(ctx context.Context, holding Asset, currentPrice float64) (float64, error)
```

**Sector_Service**
```go
type ICBSector string
const (
    VNIT   ICBSector = "VNIT"   // Technology
    VNIND  ICBSector = "VNIND"  // Industrial
    VNCONS ICBSector = "VNCONS" // Consumer
    VNCOND ICBSector = "VNCOND" // Consumer Staples
    VNHEAL ICBSector = "VNHEAL" // Healthcare
    VNENE  ICBSector = "VNENE"  // Energy
    VNUTI  ICBSector = "VNUTI"  // Utilities
    VNREAL ICBSector = "VNREAL" // Real Estate
    VNFIN  ICBSector = "VNFIN"  // Finance
    VNMAT  ICBSector = "VNMAT"  // Materials
)

type SectorTrend string
const (
    Uptrend   SectorTrend = "uptrend"
    Downtrend SectorTrend = "downtrend"
    Sideways  SectorTrend = "sideways"
)

type SectorPerformance struct {
    Sector         ICBSector   `json:"sector"`
    Trend          SectorTrend `json:"trend"`
    TodayChange    float64     `json:"todayChange"`
    OneWeekChange  float64     `json:"oneWeekChange"`
    OneMonthChange float64     `json:"oneMonthChange"`
    ThreeMonthChange float64   `json:"threeMonthChange"`
    SixMonthChange float64     `json:"sixMonthChange"`
    OneYearChange  float64     `json:"oneYearChange"`
    CurrentPrice   float64     `json:"currentPrice"`
    SMA20          float64     `json:"sma20"`
    SMA50          float64     `json:"sma50"`
}

type SectorAverages struct {
    Sector         ICBSector `json:"sector"`
    MedianPE       float64   `json:"medianPE"`
    MedianPB       float64   `json:"medianPB"`
    MedianROE      float64   `json:"medianROE"`
    MedianROA      float64   `json:"medianROA"`
    MedianDivYield float64   `json:"medianDivYield"`
    MedianDebtToEquity float64 `json:"medianDebtToEquity"`
}

type SectorService struct {
    router          *DataSourceRouter
    cache           *Cache
    stockToSector   map[string]ICBSector
}

func (s *SectorService) GetSectorPerformance(ctx context.Context, sector ICBSector) (SectorPerformance, error)
func (s *SectorService) GetAllSectorPerformances(ctx context.Context) ([]SectorPerformance, error)
func (s *SectorService) GetStockSector(symbol string) (ICBSector, error)
func (s *SectorService) GetSectorAverages(ctx context.Context, sector ICBSector) (SectorAverages, error)
func (s *SectorService) computeTrend(currentPrice, sma20, sma50 float64) SectorTrend
func (s *SectorService) refreshStockToSectorMapping(ctx context.Context) error
```

**Multi_Agent_System**
```go
type AgentMessage struct {
    AgentName   string                 `json:"agentName"`
    PayloadType string                 `json:"payloadType"`
    Payload     map[string]interface{} `json:"payload"`
    Timestamp   time.Time              `json:"timestamp"`
}

type PriceAgentResponse struct {
    Symbol         string    `json:"symbol"`
    CurrentPrice   float64   `json:"currentPrice"`
    Change         float64   `json:"change"`
    ChangePercent  float64   `json:"changePercent"`
    Volume         int64     `json:"volume"`
    Source         string    `json:"source"`
    HistoricalData []OHLCVBar `json:"historicalData,omitempty"`
}

type AnalysisAgentResponse struct {
    Symbol              string                 `json:"symbol"`
    TrendAssessment     string                 `json:"trendAssessment"`
    IndicatorSignals    map[string]string      `json:"indicatorSignals"` // indicator -> bullish/bearish/neutral
    KeyPriceLevels      map[string]float64     `json:"keyPriceLevels"`   // support/resistance
    SectorContext       SectorContextData      `json:"sectorContext"`
    CompositeSignal     string                 `json:"compositeSignal"`  // strongly bullish/bullish/neutral/bearish/strongly bearish
    BullishCount        int                    `json:"bullishCount"`
    BearishCount        int                    `json:"bearishCount"`
    ConfidenceScore     int                    `json:"confidenceScore"`  // 0-100
}

type SectorContextData struct {
    SectorName           string      `json:"sectorName"`
    SectorTrend          SectorTrend `json:"sectorTrend"`
    StockVsSectorPerf    string      `json:"stockVsSectorPerf"` // outperforming/underperforming
    SectorRotationPhase  string      `json:"sectorRotationPhase"` // capital flowing in/out/stable
}

type NewsAgentResponse struct {
    Articles []NewsArticle `json:"articles"`
    Summary  string        `json:"summary"`
}

type NewsArticle struct {
    Title       string    `json:"title"`
    Source      string    `json:"source"`
    URL         string    `json:"url"`
    PublishedAt time.Time `json:"publishedAt"`
    Snippet     string    `json:"snippet"`
}

type SupervisorRecommendation struct {
    Summary                string                   `json:"summary"`
    AssetRecommendations   []AssetRecommendation    `json:"assetRecommendations"`
    PortfolioSuggestions   []string                 `json:"portfolioSuggestions"`
    IdentifiedOpportunities []string                `json:"identifiedOpportunities"`
    SectorContext          string                   `json:"sectorContext"`
    KnowledgeBaseInsights  []string                 `json:"knowledgeBaseInsights"`
}

type AssetRecommendation struct {
    Symbol         string  `json:"symbol"`
    Action         string  `json:"action"` // buy/sell/hold
    PositionSize   float64 `json:"positionSize"` // percentage of NAV
    RiskAssessment string  `json:"riskAssessment"` // low/medium/high
    Reasoning      string  `json:"reasoning"`
}

type MultiAgentSystem struct {
    llm             llms.Model
    priceAgent      *PriceAgent
    analysisAgent   *AnalysisAgent
    newsAgent       *NewsAgent
    monitorAgent    *MonitorAgent
    supervisorAgent *SupervisorAgent
}

func (m *MultiAgentSystem) ProcessQuery(ctx context.Context, userQuery string, userID int64) (SupervisorRecommendation, error)
```

**Monitor_Agent and Knowledge_Base**
```go
type PatternType string
const (
    Accumulation PatternType = "accumulation"
    Distribution PatternType = "distribution"
    Breakout     PatternType = "breakout"
)

type PatternObservation struct {
    ID              int64       `json:"id"`
    Symbol          string      `json:"symbol"`
    PatternType     PatternType `json:"patternType"`
    DetectionDate   time.Time   `json:"detectionDate"`
    ConfidenceScore int         `json:"confidenceScore"` // 0-100
    PriceAtDetection float64    `json:"priceAtDetection"`
    SupportingData  string      `json:"supportingData"` // JSON blob
    Outcome1Day     *float64    `json:"outcome1Day,omitempty"`
    Outcome7Day     *float64    `json:"outcome7Day,omitempty"`
    Outcome14Day    *float64    `json:"outcome14Day,omitempty"`
    Outcome30Day    *float64    `json:"outcome30Day,omitempty"`
}

type PatternAccuracy struct {
    PatternType     PatternType `json:"patternType"`
    TotalObservations int       `json:"totalObservations"`
    SuccessCount    int         `json:"successCount"`
    FailureCount    int         `json:"failureCount"`
    AvgPriceChange  float64     `json:"avgPriceChange"`
    AvgConfidence   float64     `json:"avgConfidence"`
}

type MonitorAgent struct {
    router          *DataSourceRouter
    knowledgeBase   *KnowledgeBase
    alertService    *AlertService
    watchlistService *WatchlistService
    patternDetector *PatternDetector
}

func (m *MonitorAgent) RunScanCycle(ctx context.Context) error
func (m *MonitorAgent) detectPatterns(ctx context.Context, symbol string, ohlcv []OHLCVBar) ([]PatternObservation, error)

type KnowledgeBase struct {
    db *sql.DB
}

func (k *KnowledgeBase) StoreObservation(ctx context.Context, obs PatternObservation) error
func (k *KnowledgeBase) QueryObservations(ctx context.Context, filters map[string]interface{}) ([]PatternObservation, error)
func (k *KnowledgeBase) UpdateOutcomes(ctx context.Context) error
func (k *KnowledgeBase) GetAccuracyMetrics(ctx context.Context, patternType PatternType) (PatternAccuracy, error)
```

**Screener_Service**
```go
type ScreenerFilters struct {
    MinPE           *float64    `json:"minPE,omitempty"`
    MaxPE           *float64    `json:"maxPE,omitempty"`
    MinPB           *float64    `json:"minPB,omitempty"`
    MaxPB           *float64    `json:"maxPB,omitempty"`
    MinMarketCap    *float64    `json:"minMarketCap,omitempty"`
    MinEVEBITDA     *float64    `json:"minEVEBITDA,omitempty"`
    MaxEVEBITDA     *float64    `json:"maxEVEBITDA,omitempty"`
    MinROE          *float64    `json:"minROE,omitempty"`
    MaxROE          *float64    `json:"maxROE,omitempty"`
    MinROA          *float64    `json:"minROA,omitempty"`
    MaxROA          *float64    `json:"maxROA,omitempty"`
    MinRevenueGrowth *float64   `json:"minRevenueGrowth,omitempty"`
    MaxRevenueGrowth *float64   `json:"maxRevenueGrowth,omitempty"`
    MinProfitGrowth  *float64   `json:"minProfitGrowth,omitempty"`
    MaxProfitGrowth  *float64   `json:"maxProfitGrowth,omitempty"`
    MinDivYield     *float64    `json:"minDivYield,omitempty"`
    MaxDivYield     *float64    `json:"maxDivYield,omitempty"`
    MinDebtToEquity *float64    `json:"minDebtToEquity,omitempty"`
    MaxDebtToEquity *float64    `json:"maxDebtToEquity,omitempty"`
    Sectors         []ICBSector `json:"sectors,omitempty"`
    Exchanges       []string    `json:"exchanges,omitempty"` // HOSE, HNX, UPCOM
    SectorTrends    []SectorTrend `json:"sectorTrends,omitempty"`
    SortBy          string      `json:"sortBy"`
    SortOrder       string      `json:"sortOrder"` // asc/desc
    Page            int         `json:"page"`
    PageSize        int         `json:"pageSize"`
}

type ScreenerResult struct {
    Symbol          string      `json:"symbol"`
    Exchange        string      `json:"exchange"`
    Sector          ICBSector   `json:"sector"`
    MarketCap       float64     `json:"marketCap"`
    PE              float64     `json:"pe"`
    PB              float64     `json:"pb"`
    EVEBITDA        float64     `json:"evEbitda"`
    ROE             float64     `json:"roe"`
    ROA             float64     `json:"roa"`
    RevenueGrowth   float64     `json:"revenueGrowth"`
    ProfitGrowth    float64     `json:"profitGrowth"`
    DivYield        float64     `json:"divYield"`
    DebtToEquity    float64     `json:"debtToEquity"`
    SectorTrend     SectorTrend `json:"sectorTrend"`
}

type FilterPreset struct {
    ID      int64            `json:"id"`
    UserID  int64            `json:"userId"`
    Name    string           `json:"name"`
    Filters ScreenerFilters  `json:"filters"`
}

type ScreenerService struct {
    router        *DataSourceRouter
    sectorService *SectorService
    db            *sql.DB
}

func (s *ScreenerService) Screen(ctx context.Context, filters ScreenerFilters) ([]ScreenerResult, int, error)
func (s *ScreenerService) SavePreset(ctx context.Context, preset FilterPreset) error
func (s *ScreenerService) GetPresets(ctx context.Context, userID int64) ([]FilterPreset, error)
```

**Performance_Engine and Risk_Service**
```go
type PerformanceMetrics struct {
    TWR              float64            `json:"twr"`              // Time-weighted return
    MWRR             float64            `json:"mwrr"`             // Money-weighted return (XIRR)
    EquityCurve      []NAVSnapshot      `json:"equityCurve"`
    BenchmarkComparison BenchmarkData   `json:"benchmarkComparison"`
    PerformanceByType map[AssetType]float64 `json:"performanceByType"`
}

type NAVSnapshot struct {
    Date  time.Time `json:"date"`
    NAV   float64   `json:"nav"`
}

type BenchmarkData struct {
    VNIndexReturn float64 `json:"vnIndexReturn"`
    VN30Return    float64 `json:"vn30Return"`
    PortfolioReturn float64 `json:"portfolioReturn"`
    Alpha         float64 `json:"alpha"`
}

type RiskMetrics struct {
    SharpeRatio      float64            `json:"sharpeRatio"`
    MaxDrawdown      float64            `json:"maxDrawdown"`
    Beta             float64            `json:"beta"`
    Volatility       float64            `json:"volatility"`
    VaR95            float64            `json:"var95"`
    RiskContribution map[string]float64 `json:"riskContribution"` // symbol -> contribution %
}

type PerformanceEngine struct {
    db           *sql.DB
    priceService *PriceService
    router       *DataSourceRouter
}

func (e *PerformanceEngine) ComputeTWR(ctx context.Context, userID int64, startDate, endDate time.Time) (float64, error)
func (e *PerformanceEngine) ComputeMWRR(ctx context.Context, userID int64, startDate, endDate time.Time) (float64, error)
func (e *PerformanceEngine) GetEquityCurve(ctx context.Context, userID int64, startDate, endDate time.Time) ([]NAVSnapshot, error)
func (e *PerformanceEngine) StoreNAVSnapshot(ctx context.Context, userID int64, nav float64) error

type RiskService struct {
    db              *sql.DB
    performanceEngine *PerformanceEngine
    router          *DataSourceRouter
}

func (r *RiskService) ComputeRiskMetrics(ctx context.Context, userID int64) (RiskMetrics, error)
func (r *RiskService) ComputeSharpeRatio(returns []float64, riskFreeRate float64) float64
func (r *RiskService) ComputeMaxDrawdown(navHistory []NAVSnapshot) float64
func (r *RiskService) ComputeBeta(portfolioReturns, benchmarkReturns []float64) float64
func (r *RiskService) ComputeVaR(returns []float64, confidenceLevel float64) float64
```

**Rate_Limiter and Circuit_Breaker**
```go
type RateLimiter struct {
    limits map[string]*RateLimit // source -> limit
    mu     sync.RWMutex
}

type RateLimit struct {
    MaxRequests int
    Window      time.Duration
    Queue       chan struct{}
    Counter     int
    WindowStart time.Time
}

func (r *RateLimiter) Allow(source string) error
func (r *RateLimiter) GetMetrics(source string) RateLimitMetrics

type RateLimitMetrics struct {
    Source       string `json:"source"`
    CurrentCount int    `json:"currentCount"`
    QueueDepth   int    `json:"queueDepth"`
    ThrottleEvents int  `json:"throttleEvents"`
}

type CircuitBreaker struct {
    maxFailures  int
    timeout      time.Duration
    state        CircuitState
    failures     int
    lastFailTime time.Time
    mu           sync.RWMutex
}

type CircuitState string
const (
    Closed   CircuitState = "closed"
    Open     CircuitState = "open"
    HalfOpen CircuitState = "half_open"
)

func (cb *CircuitBreaker) Call(fn func() error) error
func (cb *CircuitBreaker) RecordSuccess()
func (cb *CircuitBreaker) RecordFailure()
```

**WebSocket_Server**
```go
type PriceSubscription struct {
    UserID  int64
    Symbols []string
}

type WebSocketMessage struct {
    Type    string                 `json:"type"` // subscribe, unsubscribe, price_update, market_status
    Payload map[string]interface{} `json:"payload"`
}

type WebSocketServer struct {
    clients      map[*websocket.Conn]*PriceSubscription
    priceService *PriceService
    mu           sync.RWMutex
}

func (ws *WebSocketServer) HandleConnection(conn *websocket.Conn, userID int64) error
func (ws *WebSocketServer) Subscribe(conn *websocket.Conn, symbols []string) error
func (ws *WebSocketServer) Unsubscribe(conn *websocket.Conn, symbols []string) error
func (ws *WebSocketServer) BroadcastPriceUpdate(symbol string, quote PriceQuote)
func (ws *WebSocketServer) BroadcastMarketStatus(status string)
func (ws *WebSocketServer) Heartbeat(conn *websocket.Conn) error
```

**Push_Notification_Service**
```go
type PushSubscription struct {
    ID       int64  `json:"id"`
    UserID   int64  `json:"userId"`
    Endpoint string `json:"endpoint"`
    P256dh   string `json:"p256dh"`
    Auth     string `json:"auth"`
}

type PushNotification struct {
    Title   string            `json:"title"`
    Body    string            `json:"body"`
    Icon    string            `json:"icon"`
    Actions []NotificationAction `json:"actions"`
    Data    map[string]interface{} `json:"data"`
}

type NotificationAction struct {
    Action string `json:"action"`
    Title  string `json:"title"`
}

type PushNotificationService struct {
    db         *sql.DB
    vapidKeys  VAPIDKeys
}

func (p *PushNotificationService) Subscribe(userID int64, subscription PushSubscription) error
func (p *PushNotificationService) SendNotification(userID int64, notification PushNotification) error
func (p *PushNotificationService) CleanupExpiredSubscriptions() error
```

**Email_Service**
```go
type EmailAlert struct {
    To      string
    Subject string
    Body    string
    HTML    bool
}

type EmailService struct {
    smtpHost     string
    smtpPort     int
    smtpUsername string
    smtpPassword string
    fromAddress  string
}

func (e *EmailService) SendAlert(alert EmailAlert) error
func (e *EmailService) SendBulkAlerts(alerts []EmailAlert) error
```

**SMS_Service**
```go
type SMSAlert struct {
    To      string
    Message string
}

type SMSService struct {
    apiKey    string
    apiSecret string
    provider  string // twilio, aws_sns
}

func (s *SMSService) SendAlert(alert SMSAlert) error
```

**CSV_Import_Service**
```go
type BrokerFormat string
const (
    SSI_Format   BrokerFormat = "SSI"
    VPS_Format   BrokerFormat = "VPS"
    HSC_Format   BrokerFormat = "HSC"
    VCBS_Format  BrokerFormat = "VCBS"
    Generic_Format BrokerFormat = "Generic"
)

type ImportedTransaction struct {
    Symbol          string
    Quantity        float64
    UnitPrice       float64
    TransactionDate time.Time
    TransactionType TransactionType
    Broker          string
}

type ImportResult struct {
    Imported  int
    Skipped   int
    Errors    []string
    Conflicts []ImportedTransaction
}

type CSVImportService struct {
    db              *sql.DB
    portfolioEngine *PortfolioEngine
}

func (c *CSVImportService) DetectBrokerFormat(csvData []byte) (BrokerFormat, error)
func (c *CSVImportService) ParseCSV(csvData []byte, format BrokerFormat) ([]ImportedTransaction, error)
func (c *CSVImportService) ImportTransactions(userID int64, transactions []ImportedTransaction, resolveConflicts bool) (ImportResult, error)
func (c *CSVImportService) GetImportHistory(userID int64) ([]ImportLog, error)
```

**Rebalancing_Engine**
```go
type TargetAllocation struct {
    AssetType  AssetType `json:"assetType"`
    Percentage float64   `json:"percentage"`
}

type RebalancingStrategy string
const (
    Threshold_Based   RebalancingStrategy = "threshold"
    Calendar_Based    RebalancingStrategy = "calendar"
    Opportunistic     RebalancingStrategy = "opportunistic"
)

type RebalancingRecommendation struct {
    Symbol         string  `json:"symbol"`
    Action         string  `json:"action"` // buy, sell
    Quantity       float64 `json:"quantity"`
    EstimatedCost  float64 `json:"estimatedCost"`
    Reasoning      string  `json:"reasoning"`
}

type TaxLossHarvestingOpportunity struct {
    Symbol            string  `json:"symbol"`
    UnrealizedLoss    float64 `json:"unrealizedLoss"`
    EstimatedTaxSavings float64 `json:"estimatedTaxSavings"`
    ReplacementSuggestions []string `json:"replacementSuggestions"`
}

type RebalancingEngine struct {
    db              *sql.DB
    portfolioEngine *PortfolioEngine
    priceService    *PriceService
}

func (r *RebalancingEngine) SetTargetAllocation(userID int64, targets []TargetAllocation) error
func (r *RebalancingEngine) GetTargetAllocation(userID int64) ([]TargetAllocation, error)
func (r *RebalancingEngine) ComputeRebalancingRecommendations(userID int64, strategy RebalancingStrategy) ([]RebalancingRecommendation, error)
func (r *RebalancingEngine) SimulateRebalancing(userID int64, recommendations []RebalancingRecommendation) (PortfolioSummary, error)
func (r *RebalancingEngine) IdentifyTaxLossHarvesting(userID int64) ([]TaxLossHarvestingOpportunity, error)
func (r *RebalancingEngine) GetRebalancingHistory(userID int64) ([]RebalancingEvent, error)
```

**Voice_Input_Service**
```go
type VoiceQuery struct {
    Transcript string
    Language   string
    Confidence float64
}

type NaturalLanguageIntent struct {
    Intent     string                 `json:"intent"` // portfolio_query, price_check, set_alert, market_overview
    Entities   map[string]interface{} `json:"entities"` // symbols, time_periods, metrics, actions
    Confidence float64                `json:"confidence"`
}

type VoiceInputService struct {
    llm llms.Model
}

func (v *VoiceInputService) ParseIntent(query VoiceQuery) (NaturalLanguageIntent, error)
func (v *VoiceInputService) ExtractEntities(query string) (map[string]interface{}, error)
```

**Tax_Optimization_Service**
```go
type TaxReport struct {
    Year                int       `json:"year"`
    TotalSellValue      float64   `json:"totalSellValue"`
    TotalTaxPaid        float64   `json:"totalTaxPaid"`
    RealizedGains       float64   `json:"realizedGains"`
    RealizedLosses      float64   `json:"realizedLosses"`
    DividendIncome      float64   `json:"dividendIncome"`
    CapitalGainsByAsset map[AssetType]float64 `json:"capitalGainsByAsset"`
}

type TaxOptimizationSuggestion struct {
    Type        string  `json:"type"` // tax_loss_harvest, defer_gains, optimal_withdrawal
    Description string  `json:"description"`
    EstimatedSavings float64 `json:"estimatedSavings"`
    Actions     []string `json:"actions"`
}

type CostBasisMethod string
const (
    FIFO_Method     CostBasisMethod = "FIFO"
    LIFO_Method     CostBasisMethod = "LIFO"
    SpecificID_Method CostBasisMethod = "SpecificID"
)

type TaxOptimizationService struct {
    db              *sql.DB
    portfolioEngine *PortfolioEngine
    exportService   *Export_Service
}

func (t *TaxOptimizationService) ComputeTaxLiability(userID int64, year int) (float64, error)
func (t *TaxOptimizationService) GenerateTaxReport(userID int64, year int) (TaxReport, error)
func (t *TaxOptimizationService) GetOptimizationSuggestions(userID int64) ([]TaxOptimizationSuggestion, error)
func (t *TaxOptimizationService) SimulateTaxImpact(userID int64, plannedTrades []Transaction) (float64, error)
func (t *TaxOptimizationService) SetCostBasisMethod(userID int64, method CostBasisMethod) error
```

**Sentiment_Analyzer**
```go
type SentimentScore struct {
    Symbol     string    `json:"symbol"`
    Sentiment  string    `json:"sentiment"` // positive, negative, neutral
    Score      float64   `json:"score"` // -1.0 to 1.0
    Confidence float64   `json:"confidence"`
    Source     string    `json:"source"` // news, community
    Timestamp  time.Time `json:"timestamp"`
}

type CommunitySentiment struct {
    Symbol          string  `json:"symbol"`
    DiscussionVolume string `json:"discussionVolume"` // high, medium, low
    BullishPercent  float64 `json:"bullishPercent"`
    BearishPercent  float64 `json:"bearishPercent"`
    NeutralPercent  float64 `json:"neutralPercent"`
    Trending        bool    `json:"trending"`
    TopTopics       []string `json:"topTopics"`
}

type SentimentAnalyzer struct {
    llm         llms.Model
    newsAgent   *NewsAgent
    cache       *Cache
}

func (s *SentimentAnalyzer) AnalyzeNewsSentiment(symbol string, articles []NewsArticle) (SentimentScore, error)
func (s *SentimentAnalyzer) AnalyzeCommunitySentiment(symbol string) (CommunitySentiment, error)
func (s *SentimentAnalyzer) GetSentimentTimeline(symbol string, startDate, endDate time.Time) ([]SentimentScore, error)
func (s *SentimentAnalyzer) DetectSentimentShift(symbol string) (bool, float64, error)
```

**Stress_Test_Service**
```go
type HistoricalCrisis string
const (
    Crisis_2008_Global    HistoricalCrisis = "2008_global_financial"
    Crisis_2020_COVID     HistoricalCrisis = "2020_covid"
    Crisis_2011_VN_Banking HistoricalCrisis = "2011_vn_banking"
)

type StressTestResult struct {
    Scenario        string  `json:"scenario"`
    MaxDrawdown     float64 `json:"maxDrawdown"`
    RecoveryDays    int     `json:"recoveryDays"`
    FinalNAV        float64 `json:"finalNAV"`
    VNIndexDrawdown float64 `json:"vnIndexDrawdown"`
}

type CustomScenario struct {
    Name              string             `json:"name"`
    VNIndexChange     float64            `json:"vnIndexChange"`
    SectorShocks      map[ICBSector]float64 `json:"sectorShocks"`
    CurrencyChange    float64            `json:"currencyChange"`
    InterestRateChange float64           `json:"interestRateChange"`
}

type MonteCarloResult struct {
    Percentile10  float64   `json:"percentile10"`
    Percentile50  float64   `json:"percentile50"`
    Percentile90  float64   `json:"percentile90"`
    ProbabilityOfGoal float64 `json:"probabilityOfGoal"`
    Simulations   [][]float64 `json:"simulations"` // sample paths
}

type StressTestService struct {
    db              *sql.DB
    portfolioEngine *PortfolioEngine
    riskService     *RiskService
    router          *DataSourceRouter
}

func (s *StressTestService) RunHistoricalStressTest(userID int64, crisis HistoricalCrisis) (StressTestResult, error)
func (s *StressTestService) RunCustomScenario(userID int64, scenario CustomScenario) (StressTestResult, error)
func (s *StressTestService) RunMonteCarloSimulation(userID int64, years int, simulations int) (MonteCarloResult, error)
func (s *StressTestService) IdentifyConcentrationRisks(userID int64) ([]string, error)
```

**Webhook_Service**
```go
type WebhookEvent string
const (
    Event_PriceAlert      WebhookEvent = "price_alert"
    Event_NAVMilestone    WebhookEvent = "nav_milestone"
    Event_PatternDetection WebhookEvent = "pattern_detection"
)

type WebhookRegistration struct {
    ID        int64        `json:"id"`
    UserID    int64        `json:"userId"`
    URL       string       `json:"url"`
    Events    []WebhookEvent `json:"events"`
    Secret    string       `json:"secret"`
    Active    bool         `json:"active"`
}

type WebhookPayload struct {
    Event     WebhookEvent           `json:"event"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
    Signature string                 `json:"signature"`
}

type WebhookService struct {
    db *sql.DB
}

func (w *WebhookService) RegisterWebhook(registration WebhookRegistration) error
func (w *WebhookService) DeliverWebhook(userID int64, event WebhookEvent, data map[string]interface{}) error
func (w *WebhookService) RetryFailedDeliveries() error
func (w *WebhookService) GetDeliveryLogs(userID int64) ([]WebhookDeliveryLog, error)
```

**Performance_Attribution_Engine**
```go
type HoldingAttribution struct {
    Symbol            string  `json:"symbol"`
    ContributionVND   float64 `json:"contributionVND"`
    ContributionPct   float64 `json:"contributionPct"`
    Weight            float64 `json:"weight"`
    HoldingReturn     float64 `json:"holdingReturn"`
}

type SectorAttribution struct {
    Sector            ICBSector `json:"sector"`
    ContributionVND   float64   `json:"contributionVND"`
    ContributionPct   float64   `json:"contributionPct"`
    Weight            float64   `json:"weight"`
    SectorReturn      float64   `json:"sectorReturn"`
    SelectionEffect   float64   `json:"selectionEffect"`
    AllocationEffect  float64   `json:"allocationEffect"`
}

type TimingAttribution struct {
    Symbol          string  `json:"symbol"`
    ActualReturn    float64 `json:"actualReturn"`
    BuyHoldReturn   float64 `json:"buyHoldReturn"`
    TimingEffect    float64 `json:"timingEffect"`
}

type PerformanceAttributionEngine struct {
    db                *sql.DB
    portfolioEngine   *PortfolioEngine
    performanceEngine *PerformanceEngine
}

func (p *PerformanceAttributionEngine) ComputeHoldingAttribution(userID int64, startDate, endDate time.Time) ([]HoldingAttribution, error)
func (p *PerformanceAttributionEngine) ComputeSectorAttribution(userID int64, startDate, endDate time.Time) ([]SectorAttribution, error)
func (p *PerformanceAttributionEngine) ComputeTimingAttribution(userID int64, startDate, endDate time.Time) ([]TimingAttribution, error)
func (p *PerformanceAttributionEngine) GenerateAttributionReport(userID int64, startDate, endDate time.Time) (AttributionReport, error)
```

**Market_Calendar_Service**
```go
type MarketEvent struct {
    ID          int64     `json:"id"`
    EventType   string    `json:"eventType"` // holiday, earnings, dividend, ipo, economic_release, corporate_action
    Symbol      string    `json:"symbol,omitempty"`
    Date        time.Time `json:"date"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Data        map[string]interface{} `json:"data"`
}

type EarningsEvent struct {
    Symbol      string    `json:"symbol"`
    Date        time.Time `json:"date"`
    EstimatedEPS float64  `json:"estimatedEPS"`
    ActualEPS   *float64  `json:"actualEPS,omitempty"`
}

type IPOEvent struct {
    Symbol         string    `json:"symbol"`
    CompanyName    string    `json:"companyName"`
    IPODate        time.Time `json:"ipoDate"`
    PriceRange     string    `json:"priceRange"`
    LotSize        int       `json:"lotSize"`
    SubscriptionStart time.Time `json:"subscriptionStart"`
    SubscriptionEnd   time.Time `json:"subscriptionEnd"`
}

type MarketCalendarService struct {
    db     *sql.DB
    router *DataSourceRouter
}

func (m *MarketCalendarService) GetMarketHolidays(year int) ([]time.Time, error)
func (m *MarketCalendarService) GetEarningsCalendar(startDate, endDate time.Time) ([]EarningsEvent, error)
func (m *MarketCalendarService) GetIPOCalendar(startDate, endDate time.Time) ([]IPOEvent, error)
func (m *MarketCalendarService) GetEconomicReleases(startDate, endDate time.Time) ([]MarketEvent, error)
func (m *MarketCalendarService) GetEventsForSymbol(symbol string, startDate, endDate time.Time) ([]MarketEvent, error)
```

**Audit_Logger**
```go
type AuditEvent struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"userId"`
    EventType   string    `json:"eventType"`
    Timestamp   time.Time `json:"timestamp"`
    IPAddress   string    `json:"ipAddress"`
    UserAgent   string    `json:"userAgent"`
    ResourceType string   `json:"resourceType"`
    ResourceID  *int64    `json:"resourceId,omitempty"`
    OldValues   map[string]interface{} `json:"oldValues,omitempty"`
    NewValues   map[string]interface{} `json:"newValues,omitempty"`
    Result      string    `json:"result"` // success, failure
}

type AuditLogger struct {
    db *sql.DB
}

func (a *AuditLogger) LogEvent(event AuditEvent) error
func (a *AuditLogger) QueryAuditLog(userID int64, startDate, endDate time.Time, eventTypes []string) ([]AuditEvent, error)
func (a *AuditLogger) GenerateComplianceReport(startDate, endDate time.Time) (ComplianceReport, error)
```

**Drawing_Service**
```go
type DrawingType string
const (
    TrendLine         DrawingType = "trendline"
    HorizontalLine    DrawingType = "horizontal"
    ParallelChannel   DrawingType = "parallel_channel"
    RegressionChannel DrawingType = "regression_channel"
    FibonacciChannel  DrawingType = "fibonacci_channel"
    AndrewsPitchfork  DrawingType = "andrews_pitchfork"
    SchiffPitchfork   DrawingType = "schiff_pitchfork"
    GannFan           DrawingType = "gann_fan"
    GannGrid          DrawingType = "gann_grid"
    GannSquare        DrawingType = "gann_square"
    ElliottWave       DrawingType = "elliott_wave"
    Circle            DrawingType = "circle"
    Ellipse           DrawingType = "ellipse"
    Triangle          DrawingType = "triangle"
    Rectangle         DrawingType = "rectangle"
    Polygon           DrawingType = "polygon"
    Arc               DrawingType = "arc"
    TextLabel         DrawingType = "text_label"
    Callout           DrawingType = "callout"
    Arrow             DrawingType = "arrow"
    PriceNote         DrawingType = "price_note"
    XABCDPattern      DrawingType = "xabcd_pattern"
    HeadAndShoulders  DrawingType = "head_and_shoulders"
    PriceTarget       DrawingType = "price_target"
    CycleLines        DrawingType = "cycle_lines"
    TimeZones         DrawingType = "time_zones"
    PeriodicMarkers   DrawingType = "periodic_markers"
    PriceRange        DrawingType = "price_range"
    DateRange         DrawingType = "date_range"
    DistanceTool      DrawingType = "distance_tool"
    FibonacciRetracement DrawingType = "fibonacci_retracement"
)

type DrawingPoint struct {
    Time  time.Time `json:"time"`
    Price float64   `json:"price"`
}

type DrawingStyle struct {
    LineColor   string  `json:"lineColor"`
    LineWidth   int     `json:"lineWidth"`
    LineStyle   string  `json:"lineStyle"` // solid, dashed, dotted
    FillColor   string  `json:"fillColor,omitempty"`
    FillOpacity float64 `json:"fillOpacity,omitempty"`
    TextFont    string  `json:"textFont,omitempty"`
    TextSize    int     `json:"textSize,omitempty"`
    TextColor   string  `json:"textColor,omitempty"`
}

type Drawing struct {
    ID          int64         `json:"id"`
    UserID      int64         `json:"userId"`
    Symbol      string        `json:"symbol"`
    DrawingType DrawingType   `json:"drawingType"`
    Points      []DrawingPoint `json:"points"`
    Style       DrawingStyle  `json:"style"`
    Text        string        `json:"text,omitempty"`
    Visible     bool          `json:"visible"`
    ZIndex      int           `json:"zIndex"`
    GroupID     *int64        `json:"groupId,omitempty"`
    CreatedAt   time.Time     `json:"createdAt"`
    UpdatedAt   time.Time     `json:"updatedAt"`
}

type DrawingTemplate struct {
    ID          int64        `json:"id"`
    UserID      int64        `json:"userId"`
    Name        string       `json:"name"`
    Style       DrawingStyle `json:"style"`
    CreatedAt   time.Time    `json:"createdAt"`
}

type DrawingService struct {
    db *sql.DB
}

func (d *DrawingService) SaveDrawing(ctx context.Context, drawing Drawing) (int64, error)
func (d *DrawingService) GetDrawings(ctx context.Context, userID int64, symbol string) ([]Drawing, error)
func (d *DrawingService) UpdateDrawing(ctx context.Context, drawing Drawing) error
func (d *DrawingService) DeleteDrawing(ctx context.Context, userID int64, drawingID int64) error
func (d *DrawingService) GroupDrawings(ctx context.Context, userID int64, drawingIDs []int64) (int64, error)
func (d *DrawingService) UngroupDrawings(ctx context.Context, userID int64, groupID int64) error
func (d *DrawingService) CloneDrawings(ctx context.Context, userID int64, fromSymbol, toSymbol string) error
func (d *DrawingService) SaveTemplate(ctx context.Context, template DrawingTemplate) (int64, error)
func (d *DrawingService) GetTemplates(ctx context.Context, userID int64) ([]DrawingTemplate, error)
```

**Indicator_Template_Service**
```go
type IndicatorTemplate struct {
    ID          int64                    `json:"id"`
    UserID      int64                    `json:"userId"`
    Name        string                   `json:"name"`
    Description string                   `json:"description"`
    Category    string                   `json:"category"` // trend_following, mean_reversion, momentum, volatility, volume_analysis, custom
    Indicators  []IndicatorConfiguration `json:"indicators"`
    IsFavorite  bool                     `json:"isFavorite"`
    IsPublic    bool                     `json:"isPublic"`
    CreatedAt   time.Time                `json:"createdAt"`
    UpdatedAt   time.Time                `json:"updatedAt"`
}

type IndicatorConfiguration struct {
    Type       string                 `json:"type"` // SMA, EMA, RSI, MACD, etc.
    Parameters map[string]interface{} `json:"parameters"`
    Color      string                 `json:"color"`
    LineStyle  string                 `json:"lineStyle"`
    PaneIndex  int                    `json:"paneIndex"` // 0 for main chart, 1+ for separate panes
}

type IndicatorTemplateService struct {
    db *sql.DB
}

func (i *IndicatorTemplateService) SaveTemplate(ctx context.Context, template IndicatorTemplate) (int64, error)
func (i *IndicatorTemplateService) GetTemplates(ctx context.Context, userID int64) ([]IndicatorTemplate, error)
func (i *IndicatorTemplateService) GetTemplateByID(ctx context.Context, userID int64, templateID int64) (IndicatorTemplate, error)
func (i *IndicatorTemplateService) UpdateTemplate(ctx context.Context, template IndicatorTemplate) error
func (i *IndicatorTemplateService) DeleteTemplate(ctx context.Context, userID int64, templateID int64) error
func (i *IndicatorTemplateService) GetBuiltInTemplates(ctx context.Context) ([]IndicatorTemplate, error)
func (i *IndicatorTemplateService) PublishTemplate(ctx context.Context, userID int64, templateID int64) error
func (i *IndicatorTemplateService) ImportTemplate(ctx context.Context, userID int64, templateJSON string) (int64, error)
func (i *IndicatorTemplateService) ExportTemplate(ctx context.Context, userID int64, templateID int64) (string, error)
func (i *IndicatorTemplateService) ToggleFavorite(ctx context.Context, userID int64, templateID int64) error
```

**Event_Mark_Service**
```go
type EventMarkType string
const (
    EarningsEvent       EventMarkType = "earnings"
    DividendEvent       EventMarkType = "dividend"
    StockSplitEvent     EventMarkType = "stock_split"
    BonusSharesEvent    EventMarkType = "bonus_shares"
    RightsIssueEvent    EventMarkType = "rights_issue"
    MergerEvent         EventMarkType = "merger"
    CustomEvent         EventMarkType = "custom"
)

type EventMark struct {
    ID          int64         `json:"id"`
    UserID      *int64        `json:"userId,omitempty"` // null for system events
    Symbol      string        `json:"symbol"`
    EventType   EventMarkType `json:"eventType"`
    EventDate   time.Time     `json:"eventDate"`
    Title       string        `json:"title"`
    Description string        `json:"description"`
    Icon        string        `json:"icon"`
    Color       string        `json:"color"`
    Data        map[string]interface{} `json:"data"` // event-specific data (dividend amount, split ratio, etc.)
    Visible     bool          `json:"visible"`
    CreatedAt   time.Time     `json:"createdAt"`
}

type EventMarkAlert struct {
    ID          int64     `json:"id"`
    UserID      int64     `json:"userId"`
    EventMarkID int64     `json:"eventMarkId"`
    DaysBefore  int       `json:"daysBefore"`
    Triggered   bool      `json:"triggered"`
    CreatedAt   time.Time `json:"createdAt"`
}

type EventMarkService struct {
    db                    *sql.DB
    corporateActionService *Corporate_Action_Service
}

func (e *EventMarkService) CreateEventMark(ctx context.Context, eventMark EventMark) (int64, error)
func (e *EventMarkService) GetEventMarks(ctx context.Context, userID int64, symbol string, startDate, endDate time.Time) ([]EventMark, error)
func (e *EventMarkService) UpdateEventMark(ctx context.Context, eventMark EventMark) error
func (e *EventMarkService) DeleteEventMark(ctx context.Context, userID int64, eventMarkID int64) error
func (e *EventMarkService) SyncCorporateActions(ctx context.Context, symbol string) error
func (e *EventMarkService) GetUpcomingEvents(ctx context.Context, userID int64, daysAhead int) ([]EventMark, error)
func (e *EventMarkService) SetEventAlert(ctx context.Context, alert EventMarkAlert) error
func (e *EventMarkService) BulkImportEvents(ctx context.Context, userID int64, events []EventMark) error
func (e *EventMarkService) GetEventStatistics(ctx context.Context, symbol string) (map[string]interface{}, error)
```

**Chart_Layout_Service**
```go
type ChartLayoutGrid string
const (
    Grid_1x1 ChartLayoutGrid = "1x1"
    Grid_1x2 ChartLayoutGrid = "1x2"
    Grid_2x1 ChartLayoutGrid = "2x1"
    Grid_2x2 ChartLayoutGrid = "2x2"
    Grid_1x3 ChartLayoutGrid = "1x3"
    Grid_3x1 ChartLayoutGrid = "3x1"
    Grid_2x3 ChartLayoutGrid = "2x3"
    Grid_3x2 ChartLayoutGrid = "3x2"
)

type ChartConfig struct {
    Symbol      string                     `json:"symbol"`
    Interval    string                     `json:"interval"`
    ChartType   string                     `json:"chartType"` // candlestick, hollow_candles, heikin_ashi, etc.
    Indicators  []IndicatorConfiguration   `json:"indicators"`
    ScaleMode   string                     `json:"scaleMode"` // auto, percent, indexed, locked, inverted, logarithmic
    ScaleConfig map[string]interface{}     `json:"scaleConfig"`
}

type SyncSettings struct {
    CursorSync   bool `json:"cursorSync"`
    SymbolSync   bool `json:"symbolSync"`
    TimeSync     bool `json:"timeSync"`
    DrawingsSync bool `json:"drawingsSync"`
}

type ChartLayout struct {
    ID           int64           `json:"id"`
    UserID       int64           `json:"userId"`
    Name         string          `json:"name"`
    GridLayout   ChartLayoutGrid `json:"gridLayout"`
    Charts       []ChartConfig   `json:"charts"`
    SyncSettings SyncSettings    `json:"syncSettings"`
    IsActive     bool            `json:"isActive"`
    CreatedAt    time.Time       `json:"createdAt"`
    UpdatedAt    time.Time       `json:"updatedAt"`
}

type ChartLayoutService struct {
    db *sql.DB
}

func (c *ChartLayoutService) SaveLayout(ctx context.Context, layout ChartLayout) (int64, error)
func (c *ChartLayoutService) GetLayouts(ctx context.Context, userID int64) ([]ChartLayout, error)
func (c *ChartLayoutService) GetActiveLayout(ctx context.Context, userID int64) (*ChartLayout, error)
func (c *ChartLayoutService) UpdateLayout(ctx context.Context, layout ChartLayout) error
func (c *ChartLayoutService) DeleteLayout(ctx context.Context, userID int64, layoutID int64) error
func (c *ChartLayoutService) SetActiveLayout(ctx context.Context, userID int64, layoutID int64) error
func (c *ChartLayoutService) GetBuiltInLayouts(ctx context.Context) ([]ChartLayout, error)
```

### Frontend Components

**Chart_Engine (React/TypeScript)**
```typescript
interface ChartConfig {
  symbol: string;
  interval: '1m' | '5m' | '15m' | '1h' | '1d' | '1w' | '1M';
  indicators: IndicatorConfig[];
  drawings: Drawing[];
  theme: 'light' | 'dark';
}

interface IndicatorConfig {
  type: IndicatorType;
  params: Record<string, number>;
  visible: boolean;
  color?: string;
}

type IndicatorType = 
  // Trend
  | 'SMA' | 'EMA' | 'VWAP' | 'VWMA' | 'ADX' | 'Aroon' | 'ParabolicSAR' | 'Supertrend'
  // Momentum
  | 'RSI' | 'MACD' | 'WilliamsR' | 'CMO' | 'Stochastic' | 'ROC' | 'Momentum'
  // Volatility
  | 'BollingerBands' | 'KeltnerChannel' | 'ATR' | 'StdDev'
  // Volume
  | 'OBV'
  // Statistics
  | 'LinearRegression';

interface Drawing {
  type: 'trendline' | 'horizontal' | 'fibonacci' | 'rectangle';
  points: Point[];
  color: string;
  id: string;
}

class ChartEngine {
  private chart: IChartApi;
  private candlestickSeries: ISeriesApi<'Candlestick'>;
  private volumeSeries: ISeriesApi<'Histogram'>;
  private indicators: Map<string, ISeriesApi<any>>;
  
  constructor(container: HTMLElement, config: ChartConfig);
  
  loadData(ohlcv: OHLCVBar[]): void;
  addIndicator(config: IndicatorConfig): void;
  removeIndicator(id: string): void;
  addDrawing(drawing: Drawing): void;
  removeDrawing(id: string): void;
  setTheme(theme: 'light' | 'dark'): void;
  setInterval(interval: string): void;
  exportState(): ChartState;
  restoreState(state: ChartState): void;
}
```

**Dashboard Components**
```typescript
interface DashboardProps {
  userId: number;
}

// Main dashboard showing NAV, allocation, alerts
const Dashboard: React.FC<DashboardProps> = ({ userId }) => {
  const [nav, setNav] = useState<number>(0);
  const [allocation, setAllocation] = useState<AllocationData>({});
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [recentTransactions, setRecentTransactions] = useState<Transaction[]>([]);
  
  // Poll price updates every 15s during trading hours
  useEffect(() => {
    const interval = setInterval(fetchPriceUpdates, 15000);
    return () => clearInterval(interval);
  }, []);
  
  return (
    <div className="dashboard">
      <NAVCard nav={nav} change24h={navChange} />
      <AllocationChart data={allocation} />
      <AlertPanel alerts={alerts} />
      <RecentTransactions transactions={recentTransactions} />
      <QuickMetrics goldRate={goldRate} btcPrice={btcPrice} />
    </div>
  );
};

// Screener with advanced filters
const Screener: React.FC = () => {
  const [filters, setFilters] = useState<ScreenerFilters>({});
  const [results, setResults] = useState<ScreenerResult[]>([]);
  const [presets, setPresets] = useState<FilterPreset[]>([]);
  
  const applyFilters = async () => {
    const response = await fetch('/api/screener', {
      method: 'POST',
      body: JSON.stringify(filters),
    });
    const data = await response.json();
    setResults(data.results);
  };
  
  return (
    <div className="screener">
      <FilterPanel filters={filters} onChange={setFilters} presets={presets} />
      <ResultsTable results={results} sortable />
    </div>
  );
};

// Sector trend dashboard
const SectorTrendDashboard: React.FC = () => {
  const [sectors, setSectors] = useState<SectorPerformance[]>([]);
  const [timePeriod, setTimePeriod] = useState<'1d' | '1w' | '1m' | '3m' | '6m' | '1y'>('1m');
  
  return (
    <div className="sector-dashboard">
      <TimePeriodSelector value={timePeriod} onChange={setTimePeriod} />
      <SectorHeatmap sectors={sectors} period={timePeriod} />
      <SectorBarChart sectors={sectors} period={timePeriod} />
      <TopPerformers sectors={sectors} />
    </div>
  );
};

// Stock comparison tool
const StockComparison: React.FC = () => {
  const [selectedStocks, setSelectedStocks] = useState<string[]>([]);
  const [mode, setMode] = useState<'valuation' | 'performance' | 'correlation'>('performance');
  const [timePeriod, setTimePeriod] = useState<'3m' | '6m' | '1y' | '3y' | '5y'>('1y');
  
  return (
    <div className="comparison">
      <StockSelector stocks={selectedStocks} onChange={setSelectedStocks} />
      <TabBar tabs={['Định giá', 'Hiệu suất', 'Tương quan']} active={mode} onChange={setMode} />
      {mode === 'valuation' && <ValuationChart stocks={selectedStocks} period={timePeriod} />}
      {mode === 'performance' && <PerformanceChart stocks={selectedStocks} period={timePeriod} />}
      {mode === 'correlation' && <CorrelationMatrix stocks={selectedStocks} period={timePeriod} />}
    </div>
  );
};
```

## Data Models

### Database Schema

**users**
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    failed_login_attempts INT DEFAULT 0,
    account_locked_until TIMESTAMP,
    theme_preference VARCHAR(10) DEFAULT 'light',
    language_preference VARCHAR(10) DEFAULT 'vi-VN'
);
```

**assets**
```sql
CREATE TABLE assets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_type VARCHAR(50) NOT NULL, -- vn_stock, crypto, gold, savings, bond
    symbol VARCHAR(50) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    average_cost DECIMAL(20, 2) NOT NULL,
    acquisition_date TIMESTAMP NOT NULL,
    account VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_asset_type CHECK (asset_type IN ('vn_stock', 'crypto', 'gold', 'savings', 'bond', 'cash'))
);

CREATE INDEX idx_assets_user_id ON assets(user_id);
CREATE INDEX idx_assets_symbol ON assets(symbol);
```

**transactions**
```sql
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_type VARCHAR(50) NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    total_value DECIMAL(20, 2) NOT NULL,
    transaction_date TIMESTAMP NOT NULL,
    transaction_type VARCHAR(50) NOT NULL, -- buy, sell, deposit, withdrawal, interest, dividend
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_transaction_type CHECK (transaction_type IN ('buy', 'sell', 'deposit', 'withdrawal', 'interest', 'dividend'))
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_symbol ON transactions(symbol);
```

**savings_accounts**
```sql
CREATE TABLE savings_accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_name VARCHAR(255) NOT NULL,
    principal DECIMAL(20, 2) NOT NULL,
    annual_rate DECIMAL(5, 4) NOT NULL, -- e.g., 0.0650 for 6.5%
    compounding_frequency VARCHAR(20) NOT NULL, -- monthly, quarterly, yearly
    start_date DATE NOT NULL,
    maturity_date DATE,
    is_matured BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_compounding CHECK (compounding_frequency IN ('monthly', 'quarterly', 'yearly'))
);

CREATE INDEX idx_savings_user_id ON savings_accounts(user_id);
```

**nav_snapshots**
```sql
CREATE TABLE nav_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    nav DECIMAL(20, 2) NOT NULL,
    snapshot_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, snapshot_date)
);

CREATE INDEX idx_nav_user_date ON nav_snapshots(user_id, snapshot_date);
```

**pattern_observations**
```sql
CREATE TABLE pattern_observations (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL, -- accumulation, distribution, breakout
    detection_date TIMESTAMP NOT NULL,
    confidence_score INT NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 100),
    price_at_detection DECIMAL(20, 2) NOT NULL,
    supporting_data JSONB, -- stores OHLCV snapshot, volume data, etc.
    outcome_1day DECIMAL(10, 4),
    outcome_7day DECIMAL(10, 4),
    outcome_14day DECIMAL(10, 4),
    outcome_30day DECIMAL(10, 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_pattern_type CHECK (pattern_type IN ('accumulation', 'distribution', 'breakout'))
);

CREATE INDEX idx_observations_symbol ON pattern_observations(symbol);
CREATE INDEX idx_observations_pattern ON pattern_observations(pattern_type);
CREATE INDEX idx_observations_date ON pattern_observations(detection_date);
```

**alerts**
```sql
CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    pattern_type VARCHAR(50) NOT NULL,
    confidence_score INT NOT NULL,
    explanation TEXT,
    detection_timestamp TIMESTAMP NOT NULL,
    viewed BOOLEAN DEFAULT FALSE,
    expired BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_alerts_user_id ON alerts(user_id);
CREATE INDEX idx_alerts_viewed ON alerts(viewed);
```

**watchlists**
```sql
CREATE TABLE watchlists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_watchlists_user_id ON watchlists(user_id);
```

**watchlist_symbols**
```sql
CREATE TABLE watchlist_symbols (
    id BIGSERIAL PRIMARY KEY,
    watchlist_id BIGINT NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    position INT NOT NULL, -- for ordering
    price_alert_above DECIMAL(20, 2),
    price_alert_below DECIMAL(20, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(watchlist_id, symbol)
);

CREATE INDEX idx_watchlist_symbols_watchlist_id ON watchlist_symbols(watchlist_id);
```

**filter_presets**
```sql
CREATE TABLE filter_presets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    filters JSONB NOT NULL, -- stores ScreenerFilters as JSON
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_presets_user_id ON filter_presets(user_id);
```

**recommendation_audit_log**
```sql
CREATE TABLE recommendation_audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    user_query TEXT NOT NULL,
    sub_agent_inputs JSONB NOT NULL, -- price data, analysis data, news data
    recommendation_output JSONB NOT NULL, -- SupervisorRecommendation
    symbols_involved TEXT[], -- array of symbols
    recommended_actions JSONB, -- array of {symbol, action}
    outcome_1day JSONB,
    outcome_7day JSONB,
    outcome_14day JSONB,
    outcome_30day JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_user_id ON recommendation_audit_log(user_id);
CREATE INDEX idx_audit_timestamp ON recommendation_audit_log(timestamp);
```

**financial_goals**
```sql
CREATE TABLE financial_goals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    target_amount DECIMAL(20, 2) NOT NULL,
    target_date DATE NOT NULL,
    associated_asset_types TEXT[], -- array of asset types
    category VARCHAR(50), -- retirement, emergency_fund, property, education, custom
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_goals_user_id ON financial_goals(user_id);
```

**stock_sector_mapping**
```sql
CREATE TABLE stock_sector_mapping (
    symbol VARCHAR(50) PRIMARY KEY,
    sector VARCHAR(50) NOT NULL, -- ICB sector code
    sector_name VARCHAR(255),
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sector_mapping_sector ON stock_sector_mapping(sector);
```

**cache_entries**
```sql
CREATE TABLE cache_entries (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cache_expires ON cache_entries(expires_at);
```

**push_subscriptions**
```sql
CREATE TABLE push_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used TIMESTAMP,
    UNIQUE(user_id, endpoint)
);

CREATE INDEX idx_push_user_id ON push_subscriptions(user_id);
```

**alert_preferences**
```sql
CREATE TABLE alert_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    in_app_enabled BOOLEAN DEFAULT TRUE,
    push_enabled BOOLEAN DEFAULT TRUE,
    email_enabled BOOLEAN DEFAULT FALSE,
    sms_enabled BOOLEAN DEFAULT FALSE,
    min_confidence INT DEFAULT 60,
    quiet_hours_start TIME,
    quiet_hours_end TIME,
    schedule VARCHAR(20) DEFAULT '24/7', -- trading_hours, business_hours, 24/7
    UNIQUE(user_id, alert_type)
);

CREATE INDEX idx_alert_prefs_user_id ON alert_preferences(user_id);
```

**community_profiles**
```sql
CREATE TABLE community_profiles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_sharing_enabled BOOLEAN DEFAULT FALSE,
    show_on_leaderboard BOOLEAN DEFAULT TRUE,
    anonymized_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_community_user_id ON community_profiles(user_id);
```

**public_watchlists**
```sql
CREATE TABLE public_watchlists (
    id BIGSERIAL PRIMARY KEY,
    watchlist_id BIGINT NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
    share_link VARCHAR(255) UNIQUE NOT NULL,
    follower_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(watchlist_id)
);

CREATE INDEX idx_public_watchlist_id ON public_watchlists(watchlist_id);
```

**watchlist_followers**
```sql
CREATE TABLE watchlist_followers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    public_watchlist_id BIGINT NOT NULL REFERENCES public_watchlists(id) ON DELETE CASCADE,
    followed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, public_watchlist_id)
);

CREATE INDEX idx_followers_user_id ON watchlist_followers(user_id);
CREATE INDEX idx_followers_watchlist_id ON watchlist_followers(public_watchlist_id);
```

**import_logs**
```sql
CREATE TABLE import_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source VARCHAR(255) NOT NULL, -- CSV filename or broker API name
    broker_format VARCHAR(50),
    transactions_imported INT NOT NULL,
    transactions_skipped INT NOT NULL,
    errors TEXT[],
    imported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_import_logs_user_id ON import_logs(user_id);
```

**target_allocations**
```sql
CREATE TABLE target_allocations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    asset_type VARCHAR(50) NOT NULL,
    target_percentage DECIMAL(5, 2) NOT NULL CHECK (target_percentage >= 0 AND target_percentage <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, asset_type)
);

CREATE INDEX idx_target_alloc_user_id ON target_allocations(user_id);
```

**rebalancing_events**
```sql
CREATE TABLE rebalancing_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    strategy VARCHAR(50) NOT NULL,
    executed_at TIMESTAMP NOT NULL,
    trades JSONB NOT NULL, -- array of executed trades
    nav_before DECIMAL(20, 2) NOT NULL,
    nav_after DECIMAL(20, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rebalancing_user_id ON rebalancing_events(user_id);
```

**tax_reports**
```sql
CREATE TABLE tax_reports (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year INT NOT NULL,
    total_sell_value DECIMAL(20, 2) NOT NULL,
    total_tax_paid DECIMAL(20, 2) NOT NULL,
    realized_gains DECIMAL(20, 2) NOT NULL,
    realized_losses DECIMAL(20, 2) NOT NULL,
    dividend_income DECIMAL(20, 2) NOT NULL,
    capital_gains_by_asset JSONB NOT NULL,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, year)
);

CREATE INDEX idx_tax_reports_user_id ON tax_reports(user_id);
```

**sentiment_scores**
```sql
CREATE TABLE sentiment_scores (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    sentiment VARCHAR(20) NOT NULL, -- positive, negative, neutral
    score DECIMAL(5, 4) NOT NULL CHECK (score >= -1 AND score <= 1),
    confidence DECIMAL(5, 4) NOT NULL,
    source VARCHAR(50) NOT NULL, -- news, community
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sentiment_symbol ON sentiment_scores(symbol);
CREATE INDEX idx_sentiment_timestamp ON sentiment_scores(timestamp);
```

**market_events**
```sql
CREATE TABLE market_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    symbol VARCHAR(50),
    event_date DATE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_date ON market_events(event_date);
CREATE INDEX idx_events_symbol ON market_events(symbol);
CREATE INDEX idx_events_type ON market_events(event_type);
```

**webhook_registrations**
```sql
CREATE TABLE webhook_registrations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    events TEXT[] NOT NULL,
    secret VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_webhooks_user_id ON webhook_registrations(user_id);
```

**webhook_delivery_logs**
```sql
CREATE TABLE webhook_delivery_logs (
    id BIGSERIAL PRIMARY KEY,
    webhook_id BIGINT NOT NULL REFERENCES webhook_registrations(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status_code INT,
    response_body TEXT,
    delivered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    retry_count INT DEFAULT 0
);

CREATE INDEX idx_webhook_logs_webhook_id ON webhook_delivery_logs(webhook_id);
```

**chart_templates**
```sql
CREATE TABLE chart_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    indicators JSONB NOT NULL,
    drawings JSONB NOT NULL,
    interval VARCHAR(10) NOT NULL,
    chart_style VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_chart_templates_user_id ON chart_templates(user_id);
```

**dashboard_layouts**
```sql
CREATE TABLE dashboard_layouts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    layout_name VARCHAR(255) NOT NULL,
    widgets JSONB NOT NULL, -- array of widget configurations
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, layout_name)
);

CREATE INDEX idx_dashboard_layouts_user_id ON dashboard_layouts(user_id);
```

**two_factor_auth**
```sql
CREATE TABLE two_factor_auth (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(255) NOT NULL,
    backup_codes TEXT[] NOT NULL,
    enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id)
);

CREATE INDEX idx_2fa_user_id ON two_factor_auth(user_id);
```

**user_sessions**
```sql
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    device_type VARCHAR(50),
    browser VARCHAR(100),
    ip_address VARCHAR(45),
    location VARCHAR(255),
    last_activity TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(session_token);
```

**audit_log**
```sql
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    resource_type VARCHAR(100),
    resource_id BIGINT,
    old_values JSONB,
    new_values JSONB,
    result VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp);
CREATE INDEX idx_audit_event_type ON audit_log(event_type);
```

## API Design

### REST Endpoints

**Authentication**
```
POST /api/auth/login
Request: { "username": string, "password": string }
Response: { "token": string, "expiresAt": timestamp, "user": User }

POST /api/auth/logout
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/auth/change-password
Headers: Authorization: Bearer <token>
Request: { "currentPassword": string, "newPassword": string }
Response: { "success": boolean }

GET /api/auth/me
Headers: Authorization: Bearer <token>
Response: User
```

**Price Service**
```
GET /api/prices/quotes?symbols=VNM,FPT,SSI&assetType=vn_stock
Headers: Authorization: Bearer <token>
Response: { "data": PriceQuote[] }

GET /api/prices/history?symbol=VNM&start=<timestamp>&end=<timestamp>&interval=1d
Headers: Authorization: Bearer <token>
Response: { "data": OHLCVBar[] }

GET /api/prices/gold
Headers: Authorization: Bearer <token>
Response: { "data": { "SJC": { "buy": number, "sell": number }, ... } }

GET /api/prices/crypto?ids=bitcoin,ethereum
Headers: Authorization: Bearer <token>
Response: { "data": { "bitcoin": { "usd": number, "vnd": number, "change24h": number }, ... } }

GET /api/prices/fx
Headers: Authorization: Bearer <token>
Response: { "usdVnd": number, "timestamp": timestamp, "source": string }
```

**Portfolio**
```
GET /api/portfolio/summary
Headers: Authorization: Bearer <token>
Response: PortfolioSummary

POST /api/portfolio/assets
Headers: Authorization: Bearer <token>
Request: Asset
Response: { "id": number, "asset": Asset }

PUT /api/portfolio/assets/:id
Headers: Authorization: Bearer <token>
Request: Asset
Response: { "asset": Asset }

DELETE /api/portfolio/assets/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

GET /api/portfolio/transactions?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": Transaction[], "total": number }

POST /api/portfolio/transactions
Headers: Authorization: Bearer <token>
Request: Transaction
Response: { "id": number, "transaction": Transaction }

GET /api/portfolio/performance?period=1y
Headers: Authorization: Bearer <token>
Response: PerformanceMetrics

GET /api/portfolio/risk
Headers: Authorization: Bearer <token>
Response: RiskMetrics
```

**Sector Service**
```
GET /api/sectors/performance
Headers: Authorization: Bearer <token>
Response: { "data": SectorPerformance[] }

GET /api/sectors/:sector/performance
Headers: Authorization: Bearer <token>
Response: SectorPerformance

GET /api/sectors/symbol/:symbol
Headers: Authorization: Bearer <token>
Response: { "sector": ICBSector, "sectorName": string }

GET /api/sectors/:sector/averages
Headers: Authorization: Bearer <token>
Response: SectorAverages

GET /api/sectors/:sector/stocks
Headers: Authorization: Bearer <token>
Response: { "data": string[] } // array of symbols
```

**Screener**
```
POST /api/screener
Headers: Authorization: Bearer <token>
Request: ScreenerFilters
Response: { "data": ScreenerResult[], "total": number, "page": number, "pageSize": number }

GET /api/screener/presets
Headers: Authorization: Bearer <token>
Response: { "data": FilterPreset[] }

POST /api/screener/presets
Headers: Authorization: Bearer <token>
Request: { "name": string, "filters": ScreenerFilters }
Response: { "id": number, "preset": FilterPreset }

DELETE /api/screener/presets/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }
```

**Comparison**
```
GET /api/comparison/valuation?symbols=VNM,FPT,SSI&period=1y
Headers: Authorization: Bearer <token>
Response: { "data": { "symbol": string, "timeSeries": { "date": string, "pe": number, "pb": number }[] }[] }

GET /api/comparison/performance?symbols=VNM,FPT,SSI&period=1y
Headers: Authorization: Bearer <token>
Response: { "data": { "symbol": string, "timeSeries": { "date": string, "return": number }[] }[] }

GET /api/comparison/correlation?symbols=VNM,FPT,SSI&period=1y
Headers: Authorization: Bearer <token>
Response: { "matrix": number[][], "symbols": string[] }
```

**Watchlist**
```
GET /api/watchlists
Headers: Authorization: Bearer <token>
Response: { "data": Watchlist[] }

POST /api/watchlists
Headers: Authorization: Bearer <token>
Request: { "name": string }
Response: { "id": number, "watchlist": Watchlist }

PUT /api/watchlists/:id
Headers: Authorization: Bearer <token>
Request: { "name": string }
Response: { "watchlist": Watchlist }

DELETE /api/watchlists/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/watchlists/:id/symbols
Headers: Authorization: Bearer <token>
Request: { "symbol": string, "priceAlertAbove": number?, "priceAlertBelow": number? }
Response: { "success": boolean }

DELETE /api/watchlists/:id/symbols/:symbol
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

PUT /api/watchlists/:id/reorder
Headers: Authorization: Bearer <token>
Request: { "symbols": string[] }
Response: { "success": boolean }
```

**AI Chat**
```
POST /api/chat
Headers: Authorization: Bearer <token>
Request: { 
  "message": string, 
  "provider": string, 
  "model": string, 
  "apiKey": string?,
  "conversationHistory": ChatMessage[]?
}
Response: SupervisorRecommendation

POST /api/models
Headers: Authorization: Bearer <token>
Request: { "provider": string, "apiKey": string? }
Response: { "models": string[] }
```

**Alerts**
```
GET /api/alerts?viewed=false
Headers: Authorization: Bearer <token>
Response: { "data": Alert[], "total": number }

PUT /api/alerts/:id/viewed
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

PUT /api/alerts/preferences
Headers: Authorization: Bearer <token>
Request: { "minConfidence": number, "patternTypes": PatternType[], "symbols": string[] }
Response: { "success": boolean }
```

**Knowledge Base**
```
GET /api/knowledge/observations?symbol=VNM&patternType=accumulation&startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": PatternObservation[], "total": number }

GET /api/knowledge/accuracy/:patternType
Headers: Authorization: Bearer <token>
Response: PatternAccuracy
```

**Market Data**
```
GET /api/market/listing
Headers: Authorization: Bearer <token>
Response: { "data": { "symbol": string, "name": string, "exchange": string }[] }

GET /api/market/company/:symbol
Headers: Authorization: Bearer <token>
Response: { "overview": object, "shareholders": object[], "officers": object[], "news": object[] }

GET /api/market/finance/:symbol?period=year
Headers: Authorization: Bearer <token>
Response: { "incomeStatement": object, "balanceSheet": object, "cashFlow": object, "ratios": object }

GET /api/market/statistics
Headers: Authorization: Bearer <token>
Response: { "indices": object, "marketBreadth": object, "foreignTrading": object }

GET /api/market/commodities
Headers: Authorization: Bearer <token>
Response: { "gold": object, "oil": object, "steel": object, ... }

GET /api/market/macro
Headers: Authorization: Bearer <token>
Response: { "indicators": object[] }
```

**Goals**
```
GET /api/goals
Headers: Authorization: Bearer <token>
Response: { "data": FinancialGoal[] }

POST /api/goals
Headers: Authorization: Bearer <token>
Request: { "name": string, "targetAmount": number, "targetDate": date, "assetTypes": string[], "category": string }
Response: { "id": number, "goal": FinancialGoal }

PUT /api/goals/:id
Headers: Authorization: Bearer <token>
Request: FinancialGoal
Response: { "goal": FinancialGoal }

DELETE /api/goals/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

GET /api/goals/:id/progress
Headers: Authorization: Bearer <token>
Response: { "currentValue": number, "targetAmount": number, "progress": number, "requiredMonthlyContribution": number }
```

**Backtest**
```
POST /api/backtest
Headers: Authorization: Bearer <token>
Request: { 
  "symbol": string, 
  "startDate": date, 
  "endDate": date,
  "strategy": {
    "entryCondition": string,
    "exitCondition": string,
    "stopLoss": number,
    "takeProfit": number
  }
}
Response: { 
  "totalReturn": number, 
  "winRate": number, 
  "maxDrawdown": number, 
  "sharpeRatio": number,
  "trades": number,
  "avgHoldingPeriod": number,
  "equityCurve": { "date": string, "value": number }[],
  "trades": { "entryDate": string, "exitDate": string, "return": number }[]
}
```

**Export**
```
GET /api/export/transactions?startDate=<date>&endDate=<date>&format=csv
Headers: Authorization: Bearer <token>
Response: CSV file download

GET /api/export/snapshot?format=csv
Headers: Authorization: Bearer <token>
Response: CSV file download

GET /api/export/report?format=pdf
Headers: Authorization: Bearer <token>
Response: PDF file download

GET /api/export/tax?year=2024&format=csv
Headers: Authorization: Bearer <token>
Response: CSV file download
```

**WebSocket**
```
WS /ws/prices
Headers: Authorization: Bearer <token>
Messages:
  Subscribe: { "type": "subscribe", "payload": { "symbols": ["VNM", "FPT"] } }
  Unsubscribe: { "type": "unsubscribe", "payload": { "symbols": ["VNM"] } }
  Price Update: { "type": "price_update", "payload": { "symbol": "VNM", "price": 85000, "change": 1000, ... } }
  Market Status: { "type": "market_status", "payload": { "status": "open" } }
```

**Push Notifications**
```
POST /api/notifications/subscribe
Headers: Authorization: Bearer <token>
Request: { "endpoint": string, "p256dh": string, "auth": string }
Response: { "success": boolean }

DELETE /api/notifications/unsubscribe
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

PUT /api/notifications/preferences
Headers: Authorization: Bearer <token>
Request: { "alertType": string, "push": boolean, "email": boolean, "sms": boolean, "quietHoursStart": string, "quietHoursEnd": string }
Response: { "success": boolean }
```

**Community**
```
GET /api/community/leaderboard?period=1y&metric=return
Headers: Authorization: Bearer <token>
Response: { "data": { "rank": number, "anonymizedName": string, "return": number, "sharpeRatio": number }[] }

POST /api/community/share-portfolio
Headers: Authorization: Bearer <token>
Request: { "enabled": boolean }
Response: { "success": boolean }

POST /api/community/watchlists/:id/publish
Headers: Authorization: Bearer <token>
Response: { "shareLink": string }

POST /api/community/watchlists/:shareLink/follow
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

GET /api/community/sentiment/:symbol
Headers: Authorization: Bearer <token>
Response: { "discussionVolume": string, "bullishPercent": number, "bearishPercent": number, "trending": boolean, "topTopics": string[] }
```

**Transaction Import**
```
POST /api/import/csv
Headers: Authorization: Bearer <token>
Request: multipart/form-data with CSV file
Response: { "format": string, "preview": ImportedTransaction[], "conflicts": ImportedTransaction[] }

POST /api/import/execute
Headers: Authorization: Bearer <token>
Request: { "transactions": ImportedTransaction[], "resolveConflicts": boolean }
Response: ImportResult

GET /api/import/history
Headers: Authorization: Bearer <token>
Response: { "data": ImportLog[] }

POST /api/import/broker-sync
Headers: Authorization: Bearer <token>
Request: { "broker": string, "apiKey": string }
Response: { "success": boolean, "transactionsImported": number }
```

**Rebalancing**
```
GET /api/rebalancing/target-allocation
Headers: Authorization: Bearer <token>
Response: { "data": TargetAllocation[] }

PUT /api/rebalancing/target-allocation
Headers: Authorization: Bearer <token>
Request: { "allocations": TargetAllocation[] }
Response: { "success": boolean }

POST /api/rebalancing/recommendations
Headers: Authorization: Bearer <token>
Request: { "strategy": string }
Response: { "recommendations": RebalancingRecommendation[], "estimatedCost": number }

POST /api/rebalancing/simulate
Headers: Authorization: Bearer <token>
Request: { "recommendations": RebalancingRecommendation[] }
Response: PortfolioSummary

GET /api/rebalancing/tax-loss-harvesting
Headers: Authorization: Bearer <token>
Response: { "opportunities": TaxLossHarvestingOpportunity[] }

GET /api/rebalancing/history
Headers: Authorization: Bearer <token>
Response: { "data": RebalancingEvent[] }
```

**Voice and Natural Language**
```
POST /api/voice/parse
Headers: Authorization: Bearer <token>
Request: { "transcript": string, "language": string }
Response: NaturalLanguageIntent

POST /api/voice/query
Headers: Authorization: Bearer <token>
Request: { "query": string, "language": string }
Response: { "answer": string, "data": object }
```

**Tax Optimization**
```
GET /api/tax/liability?year=2024
Headers: Authorization: Bearer <token>
Response: { "liability": number }

GET /api/tax/report?year=2024
Headers: Authorization: Bearer <token>
Response: TaxReport

GET /api/tax/suggestions
Headers: Authorization: Bearer <token>
Response: { "suggestions": TaxOptimizationSuggestion[] }

POST /api/tax/simulate
Headers: Authorization: Bearer <token>
Request: { "plannedTrades": Transaction[] }
Response: { "estimatedTaxImpact": number }

PUT /api/tax/cost-basis-method
Headers: Authorization: Bearer <token>
Request: { "method": string }
Response: { "success": boolean }
```

**Sentiment Analysis**
```
GET /api/sentiment/:symbol/news
Headers: Authorization: Bearer <token>
Response: SentimentScore

GET /api/sentiment/:symbol/community
Headers: Authorization: Bearer <token>
Response: CommunitySentiment

GET /api/sentiment/:symbol/timeline?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": SentimentScore[] }

GET /api/sentiment/trending
Headers: Authorization: Bearer <token>
Response: { "data": { "symbol": string, "sentiment": CommunitySentiment }[] }
```

**Stress Testing**
```
POST /api/stress-test/historical
Headers: Authorization: Bearer <token>
Request: { "crisis": string }
Response: StressTestResult

POST /api/stress-test/custom
Headers: Authorization: Bearer <token>
Request: CustomScenario
Response: StressTestResult

POST /api/stress-test/monte-carlo
Headers: Authorization: Bearer <token>
Request: { "years": number, "simulations": number }
Response: MonteCarloResult

GET /api/stress-test/concentration-risks
Headers: Authorization: Bearer <token>
Response: { "risks": string[] }
```

**Webhooks**
```
GET /api/webhooks
Headers: Authorization: Bearer <token>
Response: { "data": WebhookRegistration[] }

POST /api/webhooks
Headers: Authorization: Bearer <token>
Request: { "url": string, "events": string[] }
Response: { "id": number, "secret": string }

DELETE /api/webhooks/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/webhooks/:id/test
Headers: Authorization: Bearer <token>
Response: { "success": boolean, "statusCode": number }

GET /api/webhooks/:id/logs
Headers: Authorization: Bearer <token>
Response: { "data": WebhookDeliveryLog[] }
```

**Performance Attribution**
```
GET /api/attribution/holdings?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": HoldingAttribution[] }

GET /api/attribution/sectors?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": SectorAttribution[] }

GET /api/attribution/timing?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": TimingAttribution[] }

GET /api/attribution/report?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: AttributionReport
```

**Market Calendar**
```
GET /api/calendar/holidays?year=2024
Headers: Authorization: Bearer <token>
Response: { "data": time.Time[] }

GET /api/calendar/earnings?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": EarningsEvent[] }

GET /api/calendar/ipos?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": IPOEvent[] }

GET /api/calendar/economic?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": MarketEvent[] }

GET /api/calendar/symbol/:symbol?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": MarketEvent[] }

POST /api/calendar/export
Headers: Authorization: Bearer <token>
Request: { "format": string, "startDate": date, "endDate": date }
Response: iCal file download
```

**Chart Templates**
```
GET /api/charts/templates
Headers: Authorization: Bearer <token>
Response: { "data": ChartTemplate[] }

POST /api/charts/templates
Headers: Authorization: Bearer <token>
Request: { "name": string, "indicators": object, "drawings": object, "interval": string, "chartStyle": string }
Response: { "id": number }

DELETE /api/charts/templates/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }
```

**Dashboard Layouts**
```
GET /api/dashboard/layouts
Headers: Authorization: Bearer <token>
Response: { "data": DashboardLayout[] }

POST /api/dashboard/layouts
Headers: Authorization: Bearer <token>
Request: { "layoutName": string, "widgets": object[] }
Response: { "id": number }

PUT /api/dashboard/layouts/:id/activate
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

DELETE /api/dashboard/layouts/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }
```

**Two-Factor Authentication**
```
POST /api/auth/2fa/setup
Headers: Authorization: Bearer <token>
Response: { "secret": string, "qrCode": string, "backupCodes": string[] }

POST /api/auth/2fa/verify
Headers: Authorization: Bearer <token>
Request: { "code": string }
Response: { "success": boolean }

POST /api/auth/2fa/disable
Headers: Authorization: Bearer <token>
Request: { "code": string }
Response: { "success": boolean }
```

**Session Management**
```
GET /api/auth/sessions
Headers: Authorization: Bearer <token>
Response: { "data": UserSession[] }

DELETE /api/auth/sessions/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

DELETE /api/auth/sessions/all
Headers: Authorization: Bearer <token>
Response: { "success": boolean }
```

**Audit Log**
```
GET /api/audit?startDate=<date>&endDate=<date>&eventTypes=<types>
Headers: Authorization: Bearer <token>
Response: { "data": AuditEvent[], "total": number }

GET /api/audit/compliance-report?startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: ComplianceReport
```

**Integrations**
```
POST /api/integrations/google-sheets/export
Headers: Authorization: Bearer <token>
Request: { "sheetName": string }
Response: { "sheetUrl": string }

POST /api/integrations/google-sheets/auth
Headers: Authorization: Bearer <token>
Request: { "authCode": string }
Response: { "success": boolean }

POST /api/integrations/calendar/sync
Headers: Authorization: Bearer <token>
Request: { "provider": string, "authCode": string }
Response: { "success": boolean }
```

**Health and Metrics**
```
GET /api/health
Response: { 
  "status": "ok", 
  "sources": { 
    "VCI": "available", 
    "KBS": "available", 
    "CoinGecko": "available", 
    "Doji": "degraded" 
  },
  "timestamp": timestamp
}

GET /api/metrics/rate-limits
Headers: Authorization: Bearer <token>
Response: { "data": RateLimitMetrics[] }
```

## AI Agent Pipeline Design

### Agent Orchestration Flow

The Multi_Agent_System implements a coordinated pipeline where the Supervisor_Agent orchestrates specialized sub-agents:

**1. Query Analysis Phase**
- Supervisor_Agent parses user query using LLM
- Extracts symbols, asset types, and intent (analysis, news, recommendation)
- Determines which sub-agents to invoke

**2. Parallel Data Gathering Phase**
- Price_Agent: Fetches current and historical prices via Data_Source_Router
- Analysis_Agent: Computes 21 technical indicators, fundamental metrics, sector context
- News_Agent: Fetches CafeF RSS and Google search results
- All agents execute in parallel with 30-second timeout per agent

**3. Context Enrichment Phase**
- Supervisor_Agent retrieves portfolio context from Portfolio_Engine
- Queries Knowledge_Base for relevant historical patterns
- Fetches sector performance data from Sector_Service

**4. Synthesis Phase**
- Supervisor_Agent aggregates all sub-agent outputs
- Applies NAV-based position sizing logic
- Incorporates Knowledge_Base accuracy metrics
- Generates structured recommendation with reasoning

**5. Audit and Learning Phase**
- Recommendation_Audit_Log persists full inputs/outputs
- Background job tracks outcomes at 1d, 7d, 14d, 30d intervals
- Knowledge_Base ingests outcomes to improve future recommendations

### Agent Implementation Details

**Price_Agent**
- Uses Data_Source_Router for optimal source selection
- Supports VN stocks (VCI/KBS), crypto (CoinGecko), gold (Doji/SJC)
- Returns structured PriceAgentResponse with current price, change, volume, historical data
- Handles source failover transparently

**Analysis_Agent**
- Computes all 21 indicators using default parameters
- Identifies support/resistance levels using local minima/maxima
- Detects volume anomalies (>2x 20-day average)
- Retrieves ICB sector classification from Sector_Service
- Compares stock performance vs sector index over 1w, 1m, 3m, 1y
- Compares fundamental metrics vs sector medians
- Evaluates sector trend (uptrend/downtrend/sideways)
- Detects sector rotation by comparing relative performance changes
- Generates composite signal by aggregating indicator signals
- Returns AnalysisAgentResponse with confidence score 0-100

**News_Agent**
- Fetches CafeF RSS feed: https://cafef.vn/rss/...
- Filters articles by symbol mentions and keywords
- Falls back to Google search if RSS insufficient
- Limits to 10 most recent/relevant articles
- Uses LLM to summarize articles into brief insights
- Returns NewsAgentResponse with articles and summary

**Monitor_Agent (Autonomous)**
- Runs on schedule: 30min during trading hours, 2hr outside
- Fetches OHLCV data for all watchlist symbols
- Pattern_Detector identifies:
  - Accumulation: price consolidation within 5% for 10+ days, volume >1.5x 20-day avg, net institutional buying
  - Distribution: price near highs, increasing volume on down days, net institutional selling
  - Breakout: price above resistance with volume >2x 20-day avg
- Generates PatternObservation with confidence score
- Stores in Knowledge_Base with supporting data snapshot
- Alert_Service delivers notifications for confidence ≥60
- Deduplicates alerts within 48-hour window

**Supervisor_Agent**
- Orchestrates sub-agents based on query intent
- Incorporates portfolio context (NAV, allocation, holdings)
- Queries Knowledge_Base for relevant patterns
- Applies position sizing: suggests % of NAV for buy recommendations
- Checks diversification: flags if single asset >40% NAV
- Incorporates sector context from Analysis_Agent
- Formats structured SupervisorRecommendation
- Handles partial failures gracefully (proceeds with available data)

### Knowledge Base Learning Loop

**Observation Storage:**
- Pattern_Detector generates observations with supporting data
- Stores: symbol, pattern type, detection date, confidence, price, OHLCV snapshot

**Outcome Tracking:**
- Background job runs daily
- For each observation, fetches price at 1d, 7d, 14d, 30d after detection
- Computes price change percentage
- Updates observation record with outcomes

**Accuracy Metrics:**
- Aggregates observations by pattern type
- Computes success rate (% where price moved in predicted direction)
- Computes average price change after detection
- Computes average confidence score

**Recommendation Integration:**
- Supervisor_Agent queries historical observations for similar patterns
- Cites past accuracy in recommendations
- Adjusts confidence based on historical success rate
- Example: "This accumulation pattern has a 68% success rate over 30 days based on 25 past observations"

## Data Source Routing Logic

### Source Preference Mapping

The Data_Source_Router maintains a preference map for each data category:

| Data Category | Primary Source | Fallback Source | Reason |
|---------------|----------------|-----------------|--------|
| Price Quotes | VCI | KBS | VCI provides 77 columns vs KBS 28 columns |
| OHLCV History | VCI | KBS | Both equivalent, VCI preferred for consistency |
| Intraday Data | VCI | KBS | VCI supports minute-level granularity |
| Order Book | VCI | KBS | VCI provides 3-level bid/ask depth |
| Company Overview | VCI | KBS | VCI provides more complete company profiles |
| Shareholders | VCI | KBS | Both provide shareholder data |
| Officers/Directors | VCI | KBS | VCI provides more detailed officer information |
| News | VCI | KBS | Both provide company news |
| Income Statement | VCI | KBS | Both provide financial statements |
| Balance Sheet | VCI | KBS | Both provide financial statements |
| Cash Flow | VCI | KBS | Both provide financial statements |
| Financial Ratios | VCI | KBS | Both provide financial ratios |

### Failover Logic

**Timeout-Based Failover:**
1. Data_Source_Router sends request to primary source
2. If no response within 10 seconds, route to fallback source
3. Return whichever responds first

**Empty Data Failover:**
1. Primary source responds but data is empty or incomplete
2. Data_Source_Router checks for missing key fields
3. Fetch from fallback source
4. Return response with more populated fields

**Circuit Breaker Failover:**
1. Track consecutive failures per source
2. After 3 failures within 60 seconds, open circuit
3. Route all requests to fallback source
4. Attempt recovery after 60-second timeout
5. Close circuit on successful request

**Cache Fallback:**
1. Both primary and fallback sources fail
2. Return last cached result with stale indicator flag
3. Log failure for monitoring

### Rate Limiting Integration

The Data_Source_Router integrates Rate_Limiter before dispatching requests:

**Per-Source Limits:**
- VCI: 60 requests/minute
- KBS: 60 requests/minute
- CoinGecko: 10 requests/minute
- Doji: 30 requests/minute

**Queue Management:**
- When limit reached, queue pending requests
- Process queue when rate limit window resets
- Reject requests if queue depth exceeds 100

**Metrics Tracking:**
- Current request count per source within active window
- Queue depth per source
- Total throttle events count
- Exposed via GET /api/metrics/rate-limits

## Caching Strategy

### Cache TTL by Data Type

| Data Type | TTL | Reason |
|-----------|-----|--------|
| VN Stock Prices | 15 minutes | Prices update during trading hours 9:00-15:00 ICT |
| Crypto Prices | 5 minutes | 24/7 market with high volatility |
| Gold Prices | 1 hour | Gold prices update less frequently |
| USD/VND Rate | 1 hour | FX rates relatively stable intraday |
| OHLCV History | 24 hours | Historical data doesn't change |
| Company Info | 6 hours | Company profiles change infrequently |
| Financial Statements | 24 hours | Financial reports published quarterly/yearly |
| Sector Index Data | 30 minutes (trading hours), 6 hours (off-hours) | Sector indices update during trading |
| Sector Mapping | 24 hours | Stock sector classifications rarely change |
| Market Statistics | 30 minutes | Market breadth and foreign trading update frequently |
| Valuation Metrics | 1 hour | Computed from financial data |

### Cache Implementation

**In-Memory Cache (Go):**
```go
type CacheEntry struct {
    Value     interface{}
    ExpiresAt time.Time
}

type Cache struct {
    entries map[string]CacheEntry
    mu      sync.RWMutex
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    entry, exists := c.entries[key]
    if !exists || time.Now().After(entry.ExpiresAt) {
        return nil, false
    }
    return entry.Value, true
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.entries[key] = CacheEntry{
        Value:     value,
        ExpiresAt: time.Now().Add(ttl),
    }
}
```

**Database Cache (PostgreSQL):**
- Used for persistent caching across server restarts
- cache_entries table with key, value (JSONB), expires_at
- Background job cleans expired entries every hour

**Frontend Cache (localStorage):**
- User preferences (theme, language, chart settings)
- Watchlist state
- Drawing state for charts
- Last known prices for offline mode

### Cache Invalidation

**Time-Based:**
- Automatic expiration based on TTL
- Background cleanup job removes expired entries

**Event-Based:**
- Portfolio changes invalidate NAV cache
- Transaction recording invalidates portfolio summary cache
- Watchlist changes invalidate Monitor_Agent scan list

**Manual:**
- User-triggered refresh bypasses cache
- Admin endpoint to clear specific cache keys

## Security Design

### Authentication Flow

**Login Process:**
1. User submits username and password to POST /api/auth/login
2. Auth_Service retrieves user record from database
3. Check if account is locked (account_locked_until > now)
4. If locked, return error with lockout duration
5. Compare submitted password with bcrypt hash
6. If mismatch, increment failed_login_attempts
7. If failed_login_attempts ≥ 5 within 15 minutes, lock account for 30 minutes
8. If match, reset failed_login_attempts to 0
9. Generate JWT token with claims: user_id, username, issued_at, expires_at
10. Set JWT expiry to 24 hours (configurable)
11. Return token and user info to client
12. Client stores token in httpOnly cookie

**Request Authorization:**
1. Client sends request with Authorization: Bearer <token> header
2. Middleware extracts token from header
3. Verify JWT signature using secret key
4. Check token expiration
5. Extract user_id from claims
6. Attach user_id to request context
7. Proceed to route handler

**Session Management:**
1. Track last_activity timestamp per session
2. Auto-expire sessions after 4 hours of inactivity (configurable)
3. Refresh token on each request (sliding expiration)
4. Logout endpoint invalidates token (add to blacklist or use short-lived tokens)

**Password Security:**
- Hash passwords with bcrypt (cost factor 12)
- Enforce minimum password length (8 characters)
- Password change requires current password verification
- No password reset via email (out of scope for MVP)

### JWT Token Structure

```json
{
  "user_id": 123,
  "username": "investor1",
  "iat": 1704067200,
  "exp": 1704153600
}
```

Signed with HS256 algorithm using secret key from environment variable.

### API Security

**HTTPS Enforcement:**
- Production deployment requires HTTPS
- Redirect HTTP to HTTPS
- Set Secure flag on cookies

**CORS Configuration:**
- Allow requests from frontend origin only
- Credentials: true (for httpOnly cookies)
- Allowed methods: GET, POST, PUT, DELETE, OPTIONS
- Allowed headers: Content-Type, Authorization

**Rate Limiting (API Level):**
- Per-user rate limit: 100 requests/minute
- Per-IP rate limit: 200 requests/minute (for unauthenticated endpoints)
- Return 429 Too Many Requests when exceeded

**Input Validation:**
- Validate all request parameters and body fields
- Sanitize user inputs to prevent SQL injection
- Use parameterized queries for all database operations
- Validate data types, ranges, and formats

**Error Handling:**
- Never expose internal error details to client
- Log detailed errors server-side
- Return generic error messages to client
- Use appropriate HTTP status codes

### Data Protection

**Database Security:**
- Use connection pooling with max connections limit
- Store database credentials in environment variables
- Use least-privilege database user for application
- Enable SSL/TLS for database connections in production

**Sensitive Data:**
- Never log passwords or tokens
- Mask sensitive data in logs (e.g., last 4 digits of account numbers)
- Encrypt sensitive fields at rest (if required by regulations)

**API Key Management:**
- Store external API keys (OpenAI, Anthropic, etc.) in environment variables
- Allow user-provided API keys for AI chat (not persisted)
- Validate API keys before use

### Audit Trail

**Logged Events:**
- User login/logout
- Failed login attempts
- Password changes
- Portfolio transactions
- Asset additions/deletions
- AI recommendations (via Recommendation_Audit_Log)
- Pattern detections (via Knowledge_Base)
- Alert deliveries

**Log Format:**
```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "event_type": "user_login",
  "user_id": 123,
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "success": true,
  "details": {}
}
```

**Log Retention:**
- Application logs: 30 days
- Audit logs: 2 years
- Recommendation audit: 2 years
- Pattern observations: 2 years

## Error Handling

### Error Categories

**1. External API Errors**
- VCI/KBS timeout or unavailable
- CoinGecko rate limit exceeded
- Doji/SJC API failure
- LLM provider errors

**Handling:**
- Retry with exponential backoff (3 attempts)
- Failover to alternative source
- Return cached data with stale indicator
- Log error with source and timestamp

**2. Database Errors**
- Connection pool exhausted
- Query timeout
- Constraint violation
- Deadlock

**Handling:**
- Retry transient errors (connection, timeout)
- Return 500 Internal Server Error for persistent failures
- Log full error details server-side
- Return generic error message to client

**3. Validation Errors**
- Invalid input format
- Missing required fields
- Out-of-range values
- Business rule violations (e.g., insufficient holdings for sell)

**Handling:**
- Return 400 Bad Request with specific error message
- Include field name and validation rule in response
- No retry (client must fix input)

**4. Authentication/Authorization Errors**
- Invalid or expired token
- Account locked
- Insufficient permissions

**Handling:**
- Return 401 Unauthorized for invalid/expired token
- Return 403 Forbidden for insufficient permissions
- Include error message and lockout duration if applicable

**5. Rate Limiting Errors**
- Per-user or per-IP limit exceeded
- External API rate limit exceeded

**Handling:**
- Return 429 Too Many Requests
- Include Retry-After header with seconds to wait
- Queue requests if possible (for external APIs)

### Error Response Format

```json
{
  "error": {
    "code": "INSUFFICIENT_HOLDINGS",
    "message": "Cannot sell 100 shares of VNM. Current holdings: 50 shares.",
    "field": "quantity",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### Circuit Breaker Pattern

**Implementation:**
- Track consecutive failures per external source
- Threshold: 3 failures within 60 seconds
- Open circuit: route all requests to fallback
- Half-open after 60 seconds: allow 1 test request
- Close circuit on successful test request

**States:**
- Closed: Normal operation, requests pass through
- Open: All requests fail fast, return cached data
- Half-Open: Allow limited requests to test recovery

## Testing Strategy

The testing strategy employs both unit tests and property-based tests to ensure comprehensive coverage.

**Unit Testing:**
- Test specific examples and edge cases
- Test integration points between components
- Test error conditions and failure scenarios
- Mock external dependencies (VCI, KBS, CoinGecko, LLM providers)
- Use table-driven tests for multiple input scenarios

**Property-Based Testing:**
- Verify universal properties across all inputs
- Use randomized input generation
- Minimum 100 iterations per property test
- Tag each test with feature name and property number
- Focus on invariants, round-trip properties, and metamorphic properties

**Testing Tools:**
- Go: testing package, testify for assertions, gomock for mocking
- Property-based testing: gopter or rapid library
- Frontend: Jest, React Testing Library
- E2E: Playwright or Cypress

**Test Coverage Goals:**
- Unit test coverage: >80% for business logic
- Property test coverage: All correctness properties from design
- Integration test coverage: All API endpoints
- E2E test coverage: Critical user flows (login, portfolio, chat)

**Continuous Integration:**
- Run all tests on every commit
- Block merge if tests fail
- Generate coverage reports
- Run linters and formatters

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Asset Persistence Round-Trip

*For any* valid asset with type, quantity, acquisition cost, acquisition date, and account, adding the asset to the Asset_Registry and then retrieving it should return an asset with all fields matching the original values.

**Validates: Requirements 1.2**

### Property 2: Asset Update NAV Consistency

*For any* existing asset, when the asset is updated with new quantity or cost, the computed NAV should reflect the change within the same request cycle without requiring a separate NAV recalculation call.

**Validates: Requirements 1.3**

### Property 3: Asset Deletion Cascade

*For any* asset with associated transactions in the Transaction_Ledger, deleting the asset should result in zero transactions remaining for that asset's symbol and user.

**Validates: Requirements 1.4**

### Property 4: Invalid Asset Type Rejection

*For any* asset type string not in the set {vn_stock, crypto, gold, savings, bond, cash}, attempting to add an asset with that type should return an error response.

**Validates: Requirements 1.5**

### Property 5: VND Currency Invariant

*For any* monetary value stored in the Asset_Registry, Portfolio_Engine, or Transaction_Ledger, the value should be denominated in VND regardless of the input currency.

**Validates: Requirements 1.6**

### Property 6: Data Source Timeout Failover

*For any* data category request where the primary source times out after 10 seconds, the Data_Source_Router should automatically route the request to the alternative source.

**Validates: Requirements 2.3**

### Property 7: Empty Data Failover

*For any* symbol where the primary source returns empty or incomplete data (zero populated fields or missing key fields), the Data_Source_Router should fetch from the alternative source and return whichever response contains more populated fields.

**Validates: Requirements 2.4**

### Property 8: Cache Fallback on Total Failure

*For any* data category where both VCI_Source and KBS_Source fail, the Data_Source_Router should return the last cached result with a stale indicator flag set to true.

**Validates: Requirements 2.8**

### Property 9: Source Selection Logging

*For any* data request processed by the Data_Source_Router, a log entry should be generated containing the chosen source, reason for selection, and response time.

**Validates: Requirements 2.9**

### Property 10: Circuit Breaker State Transition

*For any* data source that experiences 3 consecutive failures within 60 seconds, the circuit breaker should transition to open state and route all subsequent requests to the alternative source until recovery.

**Validates: Requirements 2.10**

### Property 11: Cache TTL Enforcement

*For any* cached price where the asset type is vn_stock, crypto, or gold, the cache entry should expire after 15 minutes, 5 minutes, or 1 hour respectively, and subsequent requests should fetch fresh data.

**Validates: Requirements 3.6**

### Property 12: Price Request Batching

*For any* set of multiple stock symbols requested simultaneously, the Price_Service should batch them into a single API call to the data source when the source supports batching.

**Validates: Requirements 3.7**

### Property 13: Retry Exhaustion Cache Fallback

*For any* price request that fails after 3 retries with exponential backoff, the Price_Service should return the last cached price with a stale indicator flag.

**Validates: Requirements 3.8**

### Property 14: Buy Transaction Double-Entry

*For any* buy transaction with quantity Q and unit price P, the Portfolio_Engine should decrease the cash account by Q×P and increase the asset holding by Q shares with cost basis P.

**Validates: Requirements 4.1**

### Property 15: Sell Transaction P&L Computation

*For any* sell transaction with quantity Q at price P, where the weighted average cost is C, the Portfolio_Engine should compute realized P&L as Q×(P-C) and update the cash account accordingly.

**Validates: Requirements 4.2**

### Property 16: Unrealized P&L Calculation

*For any* holding with quantity Q, weighted average cost C, and current market price P, the unrealized P&L should equal Q×(P-C).

**Validates: Requirements 4.3**

### Property 17: NAV Aggregation

*For any* portfolio with holdings H₁, H₂, ..., Hₙ where each holding has market value Vᵢ, the total NAV should equal ΣVᵢ.

**Validates: Requirements 4.4**

### Property 18: Insufficient Holdings Rejection

*For any* sell transaction where the quantity exceeds the current holding quantity, the Portfolio_Engine should reject the transaction and return an error indicating insufficient holdings.

**Validates: Requirements 4.7**

### Property 19: Gold Price Conversion

*For any* gold price P fetched from the Doji API (in thousands VND), the Gold_Service should return P×1000 as the price in full VND.

**Validates: Requirements 5.2**

### Property 20: Gold Source Failover

*For any* gold price request where the Doji API times out or returns an error within 10 seconds, the Gold_Service should attempt to fetch from the SJC fallback endpoint.

**Validates: Requirements 5.3**

### Property 21: Gold Cache Fallback

*For any* gold price request where both Doji and SJC APIs fail, the Gold_Service should return the last cached gold prices with a stale indicator.

**Validates: Requirements 5.4**

### Property 22: Savings Interest Calculation

*For any* savings account with principal P, annual rate r, compounding frequency n, and elapsed time t years, the accrued interest should equal P×[(1 + r/n)^(n×t) - 1].

**Validates: Requirements 6.2**

### Property 23: Term Deposit Maturity Detection

*For any* term deposit where the current date is greater than or equal to the maturity date, the Savings_Tracker should flag the deposit as matured.

**Validates: Requirements 6.3**

### Property 24: Chart Drawing Persistence

*For any* chart drawing state saved to local storage, reloading the page should restore the drawing state with all drawings intact.

**Validates: Requirements 7.11**

### Property 25: Indicator Recalculation on Interval Change

*For any* chart with active indicators, switching the time interval should trigger fetching new OHLCV data and recalculating all indicator values based on the new data.

**Validates: Requirements 7.13**

### Property 26: Agent Partial Failure Tolerance

*For any* multi-agent query where one or more sub-agents fail or timeout within 30 seconds, the Supervisor_Agent should proceed with outputs from the remaining agents and note the missing data sources in the final response.

**Validates: Requirements 8.5**

### Property 27: Price Agent VND Formatting

*For any* price returned by the Price_Agent, the price should be denominated in VND regardless of the asset type or source currency.

**Validates: Requirements 9.5**

### Property 28: Indicator Omission on Insufficient Data

*For any* technical indicator requiring N data points, if fewer than N OHLCV bars are available, the Analysis_Agent should omit that indicator from the summary and include a note explaining the omission.

**Validates: Requirements 10.6**

### Property 29: Accumulation Pattern Detection

*For any* OHLCV data where price consolidates within a 5% range for 10+ trading days, daily volume exceeds 1.5× the 20-day average, and net institutional buying is present, the Monitor_Agent should detect an accumulation pattern.

**Validates: Requirements 12.3**

### Property 30: Pattern Observation Completeness

*For any* detected pattern, the Monitor_Agent should generate an observation containing symbol, pattern type, confidence score (0-100), supporting data points, and detection timestamp.

**Validates: Requirements 12.6**

### Property 31: Monitor Scan Timeout Handling

*For any* Monitor_Agent scan cycle that exceeds 5 minutes, the agent should log the failure, skip incomplete symbols, and retry on the next scheduled cycle without crashing.

**Validates: Requirements 12.9**

### Property 32: Knowledge Base Outcome Tracking

*For any* pattern observation, the Knowledge_Base should record price changes at 1-day, 7-day, 14-day, and 30-day intervals after the detection date.

**Validates: Requirements 13.2**

### Property 33: Pattern Accuracy Metric Computation

*For any* pattern type with N observations, the Knowledge_Base should compute aggregate metrics including total observations, success count, failure count, average price change, and average confidence score.

**Validates: Requirements 13.7**

### Property 34: Alert Deduplication

*For any* symbol and pattern type, if an alert was sent within the past 48 hours, the Alert_Service should not send a new alert unless the confidence score increases by 10 or more points.

**Validates: Requirements 14.6**

### Property 35: Diversification Warning

*For any* portfolio where a single asset type represents more than 40% of total NAV, the Supervisor_Agent should flag this in the recommendation output.

**Validates: Requirements 15.8**

### Property 36: Price Freshness Indicator

*For any* displayed price with age A minutes, the frontend should show a green indicator if A < 1, yellow if 1 ≤ A ≤ 5, and red if A > 5.

**Validates: Requirements 16.4**

### Property 37: Screener Graceful Degradation

*For any* symbol where financial data cannot be retrieved, the Screener_Service should exclude that symbol from filtered results and log the data gap without failing the entire query.

**Validates: Requirements 18.15**

### Property 38: Database Transaction Atomicity

*For any* database write operation that fails, the system should rollback all changes within that transaction, leaving no partially written data.

**Validates: Requirements 20.5**

### Property 39: Sector Trend Classification

*For any* sector index with current price P, SMA(20) = S₂₀, and SMA(50) = S₅₀, the Sector_Service should classify the trend as uptrend if P > S₂₀ AND P > S₅₀, downtrend if P < S₂₀ AND P < S₅₀, and sideways otherwise.

**Validates: Requirements 21.4**

### Property 40: Correlation Matrix Symmetry

*For any* correlation matrix computed by the Comparison_Engine, the matrix should be symmetric (correlation(A,B) = correlation(B,A)) and have 1.0 on the diagonal (correlation(A,A) = 1.0).

**Validates: Requirements 24.4**

### Property 41: Price Alert Triggering

*For any* watched symbol with a configured alert threshold T and current price P, if P crosses T (P > T for above-threshold or P < T for below-threshold), the Alert_Service should deliver a notification.

**Validates: Requirements 25.4**

### Property 42: TWR Calculation Correctness

*For any* portfolio with cash flows at dates D₁, D₂, ..., Dₙ, the Performance_Engine should compute TWR by chain-linking sub-period returns: TWR = ∏(1 + Rᵢ) - 1, where Rᵢ is the return in sub-period i.

**Validates: Requirements 26.1**

### Property 43: MWRR Calculation Correctness

*For any* portfolio with cash flows CF₁, CF₂, ..., CFₙ at dates D₁, D₂, ..., Dₙ and final value V, the Performance_Engine should compute MWRR as the rate r satisfying: V = Σ[CFᵢ / (1+r)^tᵢ], where tᵢ is time from Dᵢ to present.

**Validates: Requirements 26.2**

### Property 44: Sharpe Ratio Calculation

*For any* portfolio with average return R, risk-free rate Rf, and standard deviation σ, the Risk_Service should compute Sharpe ratio as (R - Rf) / σ.

**Validates: Requirements 27.1**

### Property 45: Maximum Drawdown Calculation

*For any* NAV history with values V₁, V₂, ..., Vₙ, the Risk_Service should compute max drawdown as max[(peak - trough) / peak] over all peak-to-trough sequences.

**Validates: Requirements 27.2**

### Property 46: Portfolio Beta Calculation

*For any* portfolio with daily returns P₁, P₂, ..., Pₙ and benchmark returns B₁, B₂, ..., Bₙ, the Risk_Service should compute beta as Cov(P,B) / Var(B).

**Validates: Requirements 27.3**

### Property 47: VaR Calculation

*For any* portfolio with historical daily returns sorted in ascending order, the Risk_Service should compute VaR at 95% confidence as the 5th percentile of the return distribution.

**Validates: Requirements 27.5**

### Property 48: Transaction Export Round-Trip

*For any* set of transactions exported to CSV, parsing the CSV should reconstruct all transaction records with all fields intact.

**Validates: Requirements 28.1**

### Property 49: Currency Conversion Consistency

*For any* VND-denominated value V and USD/VND exchange rate R, the displayed USD value should equal V / R.

**Validates: Requirements 29.4**

### Property 50: Dividend Transaction Auto-Recording

*For any* holding with an ex-dividend date that has passed, the Corporate_Action_Service should automatically create a dividend transaction in the Transaction_Ledger.

**Validates: Requirements 30.2**

### Property 51: Stock Split Cost Basis Adjustment

*For any* holding with quantity Q and average cost C undergoing a stock split with ratio N:M, the adjusted quantity should be Q×(N/M) and adjusted cost should be C×(M/N), maintaining total value Q×C.

**Validates: Requirements 30.3**

### Property 52: Goal Progress Calculation

*For any* financial goal with target amount T and current NAV of associated assets N, the progress percentage should equal (N / T) × 100.

**Validates: Requirements 31.2**

### Property 53: Required Contribution Calculation

*For any* goal with remaining shortfall S and M months until target date, the required monthly contribution should account for S / M (simplified) or use a more sophisticated formula considering expected returns.

**Validates: Requirements 31.3**

### Property 54: Rate Limit Enforcement

*For any* data source with request limit L per minute, the Rate_Limiter should queue or reject requests that would exceed L within the current minute window.

**Validates: Requirements 33.1**

### Property 55: Queue Overflow Rejection

*For any* data source where the request queue depth exceeds the configured maximum (default 100), the Rate_Limiter should reject new requests with a "rate limited" error.

**Validates: Requirements 33.7**

### Property 56: Offline Mode Cache Serving

*For any* frontend request while navigator.onLine is false, the frontend should serve data from local cache or last known API responses without attempting network requests.

**Validates: Requirements 34.3**

### Property 57: Recommendation Outcome Tracking

*For any* recommendation logged in the Recommendation_Audit_Log, the system should record price changes of involved symbols at 1-day, 7-day, 14-day, and 30-day intervals after the recommendation timestamp.

**Validates: Requirements 35.2**

### Property 58: JWT Token Issuance

*For any* successful authentication, the Auth_Service should issue a JWT token with user_id, username, issued_at, and expires_at claims, where expires_at is 24 hours (configurable) after issued_at.

**Validates: Requirements 36.2**

### Property 59: Account Lockout on Failed Attempts

*For any* account with 5 failed login attempts within 15 minutes, the Auth_Service should lock the account for 30 minutes and return an error indicating the lockout duration.

**Validates: Requirements 36.7**

### Property 60: Theme Change Chart State Preservation

*For any* chart with active indicators, drawings, and zoom level, changing the theme should re-render the chart with theme-appropriate colors while preserving all chart state (indicators, drawings, zoom).

**Validates: Requirements 37.9**

### Property 61: Translation String Interpolation

*For any* translation string containing variable placeholders (e.g., "{amount}"), the I18n_Service should replace placeholders with the provided variable values formatted according to the current locale.

**Validates: Requirements 38.13**

### Property 62: Mobile Touch Target Size

*For any* interactive element on mobile viewport (< 768px), the touch target size should be at least 44x44 pixels to ensure accessibility.

**Validates: Requirements 39.8**

### Property 63: PWA Offline Asset Caching

*For any* critical asset (HTML, CSS, JS, fonts) when the PWA is installed, the service worker should cache the asset for offline access.

**Validates: Requirements 39.7**

### Property 64: WebSocket Reconnection Backoff

*For any* WebSocket connection that is lost, the frontend should attempt to reconnect with exponential backoff (1s, 2s, 4s, 8s, max 30s) before falling back to HTTP polling.

**Validates: Requirements 40.8**

### Property 65: WebSocket Subscription Limit

*For any* WebSocket connection, the backend should reject subscription requests that would exceed 100 symbols per connection.

**Validates: Requirements 40.10**

### Property 66: Push Notification Quiet Hours

*For any* push notification scheduled during user-configured quiet hours (default 22:00-07:00), the Alert_Service should suppress the notification until quiet hours end.

**Validates: Requirements 41.8**

### Property 67: Push Notification Grouping

*For any* user receiving more than 5 push notifications within 1 hour, the Alert_Service should queue additional alerts and send them as a summary notification.

**Validates: Requirements 41.9**

### Property 68: Email Alert Rate Limiting

*For any* user, the Alert_Service should send a maximum of 20 email alerts per day to prevent notification fatigue.

**Validates: Requirements 42.8**

### Property 69: SMS Alert Priority Filtering

*For any* alert with priority level below "critical", the Alert_Service should not send SMS notifications.

**Validates: Requirements 42.6**

### Property 70: Community Performance Anonymization

*For any* user with public sharing enabled, the published performance metrics should not reveal absolute NAV values or specific holdings.

**Validates: Requirements 43.2**

### Property 71: CSV Import Duplicate Detection

*For any* transaction being imported that matches an existing transaction by symbol, date, quantity, price, and type, the CSV_Import_Service should detect it as a duplicate and skip it.

**Validates: Requirements 44.12**

### Property 72: Broker API Daily Sync

*For any* user with broker API integration enabled, the system should sync transactions daily at the configured time (default 16:00 ICT).

**Validates: Requirements 44.8**

### Property 73: Rebalancing Cost Effectiveness

*For any* rebalancing recommendation, the Portfolio_Engine should factor in transaction costs (broker fees, taxes) to ensure rebalancing is cost-effective.

**Validates: Requirements 45.7**

### Property 74: Tax-Loss Harvesting Identification

*For any* holding with unrealized losses, the Portfolio_Engine should identify it as a tax-loss harvesting opportunity and compute estimated tax savings.

**Validates: Requirements 45.8**

### Property 75: Natural Language Intent Extraction

*For any* natural language query, the Supervisor_Agent should extract intent and entities including time periods, asset types, metrics, and actions.

**Validates: Requirements 46.6**

### Property 76: Voice Command Execution

*For any* voice command matching predefined shortcuts ("Show my portfolio", "Check VNM price"), the system should execute the corresponding action.

**Validates: Requirements 46.10**

### Property 77: VN Tax Liability Calculation

*For any* securities trading activity in a given year, the Tax_Optimization_Service should compute VN personal income tax liability based on 0.1% on sell value for stocks.

**Validates: Requirements 47.1**

### Property 78: Tax Report Generation

*For any* year, the Export_Service should generate an annual tax report containing total sell value, total tax paid, realized capital gains/losses by asset type, and dividend income.

**Validates: Requirements 47.2**

### Property 79: Cost Basis Corporate Action Adjustment

*For any* stock split with ratio N:M, the Portfolio_Engine should adjust quantity to Q×(N/M) and cost to C×(M/N), maintaining total value Q×C.

**Validates: Requirements 47.8**

### Property 80: Interactive Tutorial Completion Tracking

*For any* user completing tutorial steps, the system should track progress and display completion badges.

**Validates: Requirements 48.10**

### Property 81: Data Anomaly Detection

*For any* price data with sudden jumps > 20% without corresponding volume increase, the Data_Source_Router should flag it as an anomaly and attempt alternative source.

**Validates: Requirements 49.1**

### Property 82: OHLCV Logical Consistency Validation

*For any* OHLCV bar, the Price_Service should validate that high >= low, high >= open, high >= close, low <= open, low <= close, and volume >= 0.

**Validates: Requirements 49.4**

### Property 83: Cross-Source Data Validation

*For any* critical data point fetched from multiple sources, the Price_Service should flag discrepancies > 5% between sources.

**Validates: Requirements 49.12**

### Property 84: Code Splitting and Lazy Loading

*For any* route in the frontend, only the JavaScript required for that route should be loaded initially, with other routes lazy-loaded on demand.

**Validates: Requirements 50.1**

### Property 85: Optimistic UI Updates

*For any* user action that triggers an API call, the frontend should show expected results immediately while the API call completes in the background.

**Validates: Requirements 50.3**

### Property 86: Database Query Indexing

*For any* frequently queried column (user_id, symbol, transaction_date, asset_type), the database should have an index to optimize query performance.

**Validates: Requirements 50.5**

### Property 87: Audit Log Immutability

*For any* audit log entry, the system should prevent modification or deletion by users or administrators.

**Validates: Requirements 51.3**

### Property 88: Audit Log Retention

*For any* audit log entry, the system should retain it for a minimum of 7 years to meet financial record-keeping requirements.

**Validates: Requirements 51.4**

### Property 89: GDPR Data Export

*For any* user requesting data export, the system should provide all their data in machine-readable JSON format.

**Validates: Requirements 51.7**

### Property 90: Chart Template Persistence

*For any* saved chart template, loading the template should restore all indicators with parameters, drawings, time interval, and chart style.

**Validates: Requirements 52.2**

### Property 91: Chart Pattern Recognition

*For any* detected chart pattern (head and shoulders, double top, triangle, etc.), the Chart_Engine should draw the pattern on the chart with labels and provide pattern analysis.

**Validates: Requirements 52.7**

### Property 92: Chart Replay Mode Future Data Hiding

*For any* chart in replay mode, the Chart_Engine should hide all price data after the current replay timestamp.

**Validates: Requirements 52.5**

### Property 93: Historical Stress Test Accuracy

*For any* historical crisis period, the Stress_Test_Service should compute portfolio drawdown by replaying actual historical price movements during that period.

**Validates: Requirements 53.1**

### Property 94: Monte Carlo Simulation Count

*For any* Monte Carlo simulation request, the Risk_Service should run at least 10,000 simulations with randomized returns based on historical volatility.

**Validates: Requirements 53.6**

### Property 95: Concentration Risk Identification

*For any* portfolio where a single holding exceeds 20% of NAV, a sector exceeds 40%, or an asset type exceeds 60%, the Risk_Service should flag it as a concentration risk.

**Validates: Requirements 53.9**

### Property 96: Webhook Retry Logic

*For any* webhook delivery that fails, the Webhook_Service should retry with exponential backoff (1s, 5s, 15s) for up to 3 attempts.

**Validates: Requirements 54.6**

### Property 97: Webhook HMAC Signature

*For any* webhook payload, the system should include an HMAC-SHA256 signature for authenticity verification.

**Validates: Requirements 54.7**

### Property 98: API Rate Limiting

*For any* authenticated user, the public API should enforce a rate limit of 100 requests per minute.

**Validates: Requirements 54.9**

### Property 99: Google Sheets OAuth Export

*For any* user with Google account OAuth integration, the Export_Service should create a Google Sheet and populate it with portfolio data.

**Validates: Requirements 55.2**

### Property 100: Calendar Event Export

*For any* set of market events, the Market_Calendar_Service should generate an iCal file compatible with Google Calendar and Outlook.

**Validates: Requirements 55.9**

### Property 101: Command Palette Fuzzy Search

*For any* command palette query (Cmd+K), the frontend should perform fuzzy search across all features, navigation, and actions.

**Validates: Requirements 56.1**

### Property 102: Undo/Redo Transaction Operations

*For any* transaction operation, the frontend should maintain an undo history of the last 20 actions per session with Cmd+Z / Cmd+Shift+Z support.

**Validates: Requirements 56.8**

### Property 103: Dashboard Widget Drag-and-Drop

*For any* dashboard widget, users should be able to drag, drop, resize, and reorder widgets with the layout persisted to the backend.

**Validates: Requirements 56.5**

### Property 104: 2FA TOTP Verification

*For any* login attempt with 2FA enabled, the Auth_Service should require both password and valid TOTP code.

**Validates: Requirements 57.3**

### Property 105: Biometric Authentication Support

*For any* device supporting WebAuthn, the frontend should offer biometric authentication (fingerprint, Face ID) as a login option.

**Validates: Requirements 57.5**

### Property 106: Session Device Tracking

*For any* active session, the Auth_Service should track device type, browser, IP address, location, and last activity timestamp.

**Validates: Requirements 57.6**

### Property 107: Suspicious Activity Detection

*For any* login from a new device, unusual location, or after multiple failed attempts, the Auth_Service should flag it as suspicious and require additional verification.

**Validates: Requirements 57.8**

### Property 108: Password Breach Detection

*For any* password, the Auth_Service should check it against known breach databases (Have I Been Pwned API) and force password change if compromised.

**Validates: Requirements 57.14**

### Property 109: Market Holiday Calendar

*For any* year, the Market_Calendar_Service should maintain a complete list of VN market holidays and trading schedule.

**Validates: Requirements 58.1**

### Property 110: Earnings Reminder Delivery

*For any* holding with an upcoming earnings announcement, the Alert_Service should send reminders 1 week before, 1 day before, and on the day of announcement.

**Validates: Requirements 58.5**

### Property 111: Vietnamese News Sentiment Classification

*For any* Vietnamese financial news article, the Sentiment_Analyzer should classify it as positive, negative, or neutral using Vietnamese NLP models.

**Validates: Requirements 59.2**

### Property 112: Sentiment Timeline Tracking

*For any* symbol, the News_Agent should aggregate sentiment scores over rolling time windows (24 hours, 7 days, 30 days).

**Validates: Requirements 59.4**

### Property 113: Breaking News Sentiment Shift Detection

*For any* symbol with sentiment change > 30 points in 24 hours, the News_Agent should trigger an alert for holdings.

**Validates: Requirements 59.11**

### Property 114: Alert Schedule Enforcement

*For any* alert configured for "trading hours only", the Alert_Service should only deliver alerts between 9:00-15:00 ICT.

**Validates: Requirements 60.1**

### Property 115: Alert Grouping Within Time Window

*For any* user receiving multiple alerts of the same type within 1 hour, the Alert_Service should group them into a single summary notification.

**Validates: Requirements 60.4**

### Property 116: Complex Alert Rule Evaluation

*For any* alert rule with complex conditions (e.g., "price > 100,000 AND volume > 2M AND RSI < 30"), the Alert_Service should evaluate all conditions before triggering.

**Validates: Requirements 60.12**

### Property 117: Community Trending Stock Detection

*For any* stock with 3x or more discussion volume compared to 7-day average, the News_Agent should flag it as trending.

**Validates: Requirements 61.4**

### Property 118: Community Sentiment Spam Filtering

*For any* forum post or comment, the News_Agent should implement spam and bot detection to filter out low-quality posts.

**Validates: Requirements 61.12**

### Property 119: Holding-Level Attribution Calculation

*For any* holding with weight W and return R in a portfolio with total return P, the contribution should equal W × R.

**Validates: Requirements 62.2**

### Property 120: Sector Attribution Decomposition

*For any* sector, the Performance_Attribution_Engine should decompose return contribution into selection effect (stock picking within sector) and allocation effect (sector weight vs benchmark).

**Validates: Requirements 62.4**

### Property 121: Timing Attribution Analysis

*For any* holding, the Performance_Attribution_Engine should compare actual returns against buy-and-hold returns to compute timing effect.

**Validates: Requirements 62.6**

### Property 122: Best/Worst Decision Identification

*For any* time period, the Performance_Attribution_Engine should identify best and worst investment decisions based on realized returns and opportunity cost.

**Validates: Requirements 62.9**


### Property 123: Hollow Candles Rendering

*For any* OHLCV bar where close > open (bullish), the Chart_Engine should render it as a hollow candle; where close < open (bearish), render as filled candle.

**Validates: Requirements 63.1**

### Property 124: Heikin Ashi Calculation

*For any* sequence of OHLCV bars, the Chart_Engine should compute Heikin Ashi values using: HA_Close = (O+H+L+C)/4, HA_Open = (prev_HA_Open + prev_HA_Close)/2, HA_High = max(H, HA_Open, HA_Close), HA_Low = min(L, HA_Open, HA_Close).

**Validates: Requirements 63.2**

### Property 125: Chart Type Persistence

*For any* symbol and chart type selection, the Chart_Engine should persist the preference in local storage and restore it when the user views the same symbol again.

**Validates: Requirements 63.9**

### Property 126: Chart Type Indicator Preservation

*For any* chart with active indicators and drawings, switching chart types should preserve all indicators, drawings, and zoom level while re-rendering price data in the new style.

**Validates: Requirements 63.8**

### Property 127: Drawing Persistence Across Sessions

*For any* logged-in user, all drawings created on a symbol's chart should be persisted to the backend database and synchronized across devices and sessions.

**Validates: Requirements 64.20**

### Property 128: Drawing Group Operations

*For any* group of selected drawings, grouping them should allow moving, copying, or deleting all drawings in the group as a single unit.

**Validates: Requirements 64.16**

### Property 129: Drawing Snapping Precision

*For any* drawing tool with snapping enabled, the drawing should snap to the nearest OHLC data point, indicator line, or other drawing when placed within 5 pixels.

**Validates: Requirements 64.18**

### Property 130: Drawing Clone Scaling

*For any* set of drawings cloned from one symbol to another, the Chart_Engine should automatically scale the drawings to fit the target symbol's price range while maintaining relative proportions.

**Validates: Requirements 64.24**

### Property 131: Multiple Price Scales Independence

*For any* chart with multiple price scales (up to 8), each scale should operate independently with its own Y-axis range and scaling mode.

**Validates: Requirements 65.1**

### Property 132: Percent Scale Calculation

*For any* price in Percent scale mode, the displayed value should equal ((current_price - reference_price) / reference_price) × 100.

**Validates: Requirements 65.7**

### Property 133: Indexed Scale Normalization

*For any* price in Indexed to 100 mode, the displayed value should equal (current_price / reference_price) × 100.

**Validates: Requirements 65.8**

### Property 134: Locked Scale Range Enforcement

*For any* chart in Locked scale mode, zooming or panning should not change the Y-axis range, maintaining the configured min and max prices.

**Validates: Requirements 65.4**

### Property 135: Logarithmic Scale Spacing

*For any* chart in logarithmic scale mode, the Y-axis should use logarithmic spacing where equal percentage changes occupy equal vertical distances.

**Validates: Requirements 65.11**

### Property 136: Second-Based Timeframe Data Fetching

*For any* second-based timeframe (1s, 5s, 15s, 30s) selected during trading hours, the Chart_Engine should fetch second-level OHLCV data from the Data_Source_Router.

**Validates: Requirements 66.2**

### Property 137: Bar Countdown Accuracy

*For any* active timeframe, the bar countdown timer should display the time remaining until the next candle closes, updating every second.

**Validates: Requirements 66.5**

### Property 138: Go To Date Navigation

*For any* date selected via the "Go to Date" feature, the chart should immediately jump to display that date at the center of the visible range.

**Validates: Requirements 66.4**

### Property 139: Date Range Preset Application

*For any* date range preset (Last 7 Days, Last 30 Days, etc.), selecting it should immediately update the chart to display exactly that time period.

**Validates: Requirements 66.8**

### Property 140: Trading Visualization Buy/Sell Markers

*For any* executed buy transaction, the Chart_Engine should display a green upward-pointing triangle at the entry price level; for sell transactions, a red downward-pointing triangle at the exit price level.

**Validates: Requirements 67.1, 67.2**

### Property 141: Cost Basis Line Display

*For any* active holding, the Chart_Engine should draw a horizontal line at the weighted average purchase price with a label showing the cost basis value.

**Validates: Requirements 67.3**

### Property 142: P/L Zone Shading

*For any* active holding, the Chart_Engine should shade the area between cost basis and current price: green for profit zones (price > cost), red for loss zones (price < cost).

**Validates: Requirements 67.5**

### Property 143: Round-Trip Trade Visualization

*For any* completed buy-sell pair, the Chart_Engine should connect the markers with a line color-coded green for profitable trades, red for losing trades, with P/L percentage displayed.

**Validates: Requirements 67.6**

### Property 144: Trade Statistics Calculation

*For any* visible chart period, the trade statistics overlay should compute: total trades, win rate, average profit/loss per trade, largest win/loss, and total P&L.

**Validates: Requirements 67.7**

### Property 145: Quick Order Entry Pre-Fill

*For any* Buy or Sell button click, the quick order entry modal should pre-fill with the current symbol, current market price, and focus the quantity input field.

**Validates: Requirements 68.2, 68.3**

### Property 146: Bracket Order Risk/Reward Calculation

*For any* bracket order with entry, stop-loss, and take-profit prices defined, the Chart_Engine should calculate and display the risk/reward ratio.

**Validates: Requirements 68.6**

### Property 147: Position Sizing Calculator

*For any* risk amount (VND) and stop-loss percentage input, the position sizing calculator should suggest the appropriate quantity to risk only the specified amount.

**Validates: Requirements 68.11**

### Property 148: One-Click Watchlist Trading

*For any* symbol in the watchlist with Buy or Sell icon clicked, the quick order entry modal should open pre-filled with that symbol and current price.

**Validates: Requirements 68.7**

### Property 149: Watchlist Multi-Column Sorting

*For any* watchlist column header clicked, the watchlist should sort by that column; Shift+click should add secondary sort criteria.

**Validates: Requirements 69.1, 69.3**

### Property 150: Watchlist Drag-and-Drop Reordering

*For any* symbol dragged to a new position in the watchlist, the custom order should be persisted to the backend and maintained across sessions.

**Validates: Requirements 69.5, 69.7**

### Property 151: Watchlist Bulk Actions

*For any* multiple symbols selected via checkboxes, bulk actions (Add Alert, Remove, Move, Export) should apply to all selected symbols.

**Validates: Requirements 69.8**

### Property 152: Watchlist Column Visibility Persistence

*For any* column visibility configuration, the settings should be persisted per user and restored across sessions.

**Validates: Requirements 69.10**

### Property 153: Indicator Template Save and Load

*For any* chart with active indicators, saving as a template should store all indicators with parameters, colors, line styles, and pane assignments; loading should restore exactly that configuration.

**Validates: Requirements 70.1, 70.5**

### Property 154: Indicator Template Categorization

*For any* indicator template, it should be assigned to a category (Trend Following, Mean Reversion, Momentum, Volatility, Volume Analysis, Custom) for organization.

**Validates: Requirements 70.6**

### Property 155: Built-In Template Library

*For any* user, the Chart_Engine should provide built-in templates including: "Trend Trader", "Momentum Scalper", "Volatility Breakout", "Swing Trader", and "Multi-Timeframe".

**Validates: Requirements 70.7**

### Property 156: Template Import/Export Round-Trip

*For any* indicator template exported as JSON, importing that JSON should recreate the template with all indicators and settings intact.

**Validates: Requirements 70.9, 70.10**

### Property 157: Automatic Corporate Action Event Marks

*For any* symbol, the Chart_Engine should automatically display earnings dates, ex-dividend dates, and corporate actions (splits, bonus shares) fetched from Corporate_Action_Service.

**Validates: Requirements 71.1, 71.2, 71.3**

### Property 158: Custom Event Mark Creation

*For any* date on the chart, right-clicking and selecting "Add Event Mark" should allow the user to create a custom event with name, description, and icon.

**Validates: Requirements 71.4**

### Property 159: Event Mark Visibility Filtering

*For any* event mark category (earnings, dividends, corporate actions, custom), users should be able to toggle visibility without deleting the marks.

**Validates: Requirements 71.8**

### Property 160: Event Mark Alert Scheduling

*For any* event mark, users should be able to set alerts to be notified X days before the event date.

**Validates: Requirements 71.11**

### Property 161: Multi-Chart Cursor Synchronization

*For any* multi-chart layout with cursor sync enabled, moving the crosshair on one chart should highlight the same date/time position on all other charts with synchronized vertical lines.

**Validates: Requirements 72.3**

### Property 162: Multi-Chart Symbol Synchronization

*For any* multi-chart layout with symbol sync enabled, changing the symbol on one chart should automatically change the symbol on all other charts to the same symbol.

**Validates: Requirements 72.4**

### Property 163: Multi-Chart Time Synchronization

*For any* multi-chart layout with time sync enabled, zooming or panning on one chart should apply the same time range and zoom level to all other charts.

**Validates: Requirements 72.5**

### Property 164: Multi-Chart Drawings Synchronization

*For any* multi-chart layout with drawings sync enabled, creating, modifying, or deleting a drawing on one chart should replicate the action on all other charts with automatic price scaling.

**Validates: Requirements 72.6**

### Property 165: Multi-Chart Layout Persistence

*For any* multi-chart configuration including grid arrangement, symbols, timeframes, indicators, and sync settings, the layout should be persisted in local storage and restored when returning to the chart view.

**Validates: Requirements 72.10**

### Property 166: Multi-Chart Independent Configuration

*For any* chart in a multi-chart layout with sync modes disabled, it should support different symbols, timeframes, chart types, and indicator sets independently.

**Validates: Requirements 72.7**

### Property 167: Chart Layout Template Save and Load

*For any* multi-chart layout, users should be able to save it as a template with a name, and loading the template should restore the complete configuration.

**Validates: Requirements 72.9**

### Property 168: Multi-Chart Performance Optimization

*For any* multi-chart layout, off-screen charts should be lazy-rendered, updates should be throttled during rapid panning/zooming, and data should be cached across charts displaying the same symbol.

**Validates: Requirements 72.18**

## Error Handling

### Advanced Charting Error Handling

**Drawing Service Errors:**
- Invalid drawing type: Return error with list of supported drawing types
- Drawing not found: Return 404 with drawing ID
- Insufficient points for drawing type: Return error specifying required point count
- Drawing ownership violation: Return 403 when user attempts to modify another user's drawing
- Drawing group conflict: Return error when attempting to group drawings from different symbols

**Indicator Template Errors:**
- Template name conflict: Return error when saving template with duplicate name
- Invalid indicator type: Return error with list of supported indicators
- Invalid indicator parameters: Return error specifying valid parameter ranges
- Template not found: Return 404 with template ID
- Template import parse error: Return error with JSON validation details

**Event Mark Errors:**
- Invalid event type: Return error with list of supported event types
- Event date in future beyond 5 years: Return error with maximum allowed date
- Corporate action sync failure: Log error, return cached data with stale indicator
- Event mark not found: Return 404 with event mark ID
- Bulk import validation failure: Return partial success with list of failed imports and reasons

**Chart Layout Errors:**
- Invalid grid layout: Return error with list of supported grid layouts (1x1, 1x2, 2x1, 2x2, 1x3, 3x1, 2x3, 3x2)
- Chart count mismatch: Return error when number of charts doesn't match grid layout
- Layout not found: Return 404 with layout ID
- Active layout conflict: Return error when attempting to set multiple layouts as active

**Scale Configuration Errors:**
- Invalid scale mode: Return error with list of supported modes
- Locked scale missing min/max: Return error requiring both min and max for locked mode
- Reference point missing: Return error requiring reference point for percent/indexed modes
- Too many price scales: Return error when attempting to add more than 8 scales

**Timeframe Errors:**
- Second-based data unavailable: Return error with message indicating second data only available during trading hours (9:00-15:00 ICT)
- Invalid timeframe: Return error with list of supported timeframes
- Date range too large: Return error when requested range exceeds maximum (10 years for daily, 1 year for intraday)

**Trading Visualization Errors:**
- No transactions found: Return empty array with message
- Cost basis calculation failure: Log error, return null cost basis
- Trade pairing mismatch: Log warning when sell quantity exceeds buy quantity

**Quick Trading Errors:**
- Insufficient holdings for sell: Return error with available quantity
- Invalid quantity: Return error specifying minimum lot size and maximum quantity
- Market closed: Return error with next trading session time
- Order validation failure: Return error with specific validation failure reason

## Testing Strategy

The testing strategy employs both unit tests and property-based tests to ensure comprehensive coverage.

**Unit Testing:**
- Test specific examples and edge cases
- Test integration points between components
- Test error conditions and failure scenarios
- Mock external dependencies (VCI, KBS, CoinGecko, LLM providers)
- Use table-driven tests for multiple input scenarios

**Property-Based Testing:**
- Verify universal properties across all inputs
- Use randomized input generation
- Minimum 100 iterations per property test
- Tag each test with feature name and property number
- Focus on invariants, round-trip properties, and metamorphic properties

**Advanced Charting Testing:**
- Test all 11 chart types render correctly with sample OHLCV data
- Test Heikin Ashi calculation accuracy against known test vectors
- Test all 110+ drawing tools can be created, edited, and deleted
- Test drawing persistence round-trip (save to DB, retrieve, verify all fields match)
- Test drawing group operations (group, ungroup, move group, delete group)
- Test drawing clone scaling accuracy across different price ranges
- Test all 6 scale modes (auto, percent, indexed, locked, inverted, logarithmic) with various price data
- Test multi-chart synchronization (cursor, symbol, time, drawings) across all grid layouts
- Test indicator template save/load preserves all configuration
- Test event mark automatic sync from Corporate_Action_Service
- Test quick trading modal pre-fill accuracy
- Test watchlist drag-and-drop reordering persistence
- Test bar countdown timer accuracy for all timeframes

**Testing Tools:**
- Go: testing package, testify for assertions, gomock for mocking
- Property-based testing: gopter or rapid library
- Frontend: Jest, React Testing Library, @testing-library/user-event for drag-and-drop
- E2E: Playwright for multi-chart layout testing
- Visual regression: Percy or Chromatic for chart rendering consistency

**Test Coverage Goals:**
- Unit test coverage: >80% for business logic
- Property test coverage: All correctness properties from design (Properties 123-168)
- Integration test coverage: All API endpoints including new drawing, template, event mark, and layout endpoints
- E2E test coverage: Critical user flows (chart type switching, drawing creation, multi-chart setup, quick trading)

**Continuous Integration:**
- Run all tests on every commit
- Block merge if tests fail
- Generate coverage reports
- Run linters and formatters

## Additional Database Tables for Requirements 63-72

**drawings**
```sql
CREATE TABLE drawings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    drawing_type VARCHAR(50) NOT NULL,
    points JSONB NOT NULL, -- array of {time, price}
    style JSONB NOT NULL, -- {lineColor, lineWidth, lineStyle, fillColor, fillOpacity, textFont, textSize, textColor}
    text TEXT,
    visible BOOLEAN DEFAULT TRUE,
    z_index INT DEFAULT 0,
    group_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_drawings_user_symbol ON drawings(user_id, symbol);
CREATE INDEX idx_drawings_group ON drawings(group_id);
```

**drawing_templates**
```sql
CREATE TABLE drawing_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    style JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_drawing_templates_user ON drawing_templates(user_id);
```

**indicator_templates**
```sql
CREATE TABLE indicator_templates (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    indicators JSONB NOT NULL, -- array of {type, parameters, color, lineStyle, paneIndex}
    is_favorite BOOLEAN DEFAULT FALSE,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_indicator_templates_user ON indicator_templates(user_id);
CREATE INDEX idx_indicator_templates_category ON indicator_templates(category);
CREATE INDEX idx_indicator_templates_public ON indicator_templates(is_public) WHERE is_public = TRUE;
```

**event_marks**
```sql
CREATE TABLE event_marks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE, -- NULL for system events
    symbol VARCHAR(50) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_date DATE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(20),
    data JSONB, -- event-specific data
    visible BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_marks_symbol_date ON event_marks(symbol, event_date);
CREATE INDEX idx_event_marks_user ON event_marks(user_id);
CREATE INDEX idx_event_marks_type ON event_marks(event_type);
```

**event_mark_alerts**
```sql
CREATE TABLE event_mark_alerts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_mark_id BIGINT NOT NULL REFERENCES event_marks(id) ON DELETE CASCADE,
    days_before INT NOT NULL,
    triggered BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_event_alerts_user ON event_mark_alerts(user_id);
CREATE INDEX idx_event_alerts_event ON event_mark_alerts(event_mark_id);
```

**chart_layouts**
```sql
CREATE TABLE chart_layouts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    grid_layout VARCHAR(10) NOT NULL, -- 1x1, 1x2, 2x1, 2x2, 1x3, 3x1, 2x3, 3x2
    charts JSONB NOT NULL, -- array of {symbol, interval, chartType, indicators, scaleMode, scaleConfig}
    sync_settings JSONB NOT NULL, -- {cursorSync, symbolSync, timeSync, drawingsSync}
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name)
);

CREATE INDEX idx_chart_layouts_user ON chart_layouts(user_id);
CREATE INDEX idx_chart_layouts_active ON chart_layouts(user_id, is_active) WHERE is_active = TRUE;
```

**chart_preferences**
```sql
CREATE TABLE chart_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol VARCHAR(50) NOT NULL,
    chart_type VARCHAR(50) DEFAULT 'candlestick',
    scale_mode VARCHAR(50) DEFAULT 'auto',
    scale_config JSONB,
    timeframe VARCHAR(10) DEFAULT '1d',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, symbol)
);

CREATE INDEX idx_chart_prefs_user_symbol ON chart_preferences(user_id, symbol);
```

## Additional API Endpoints for Requirements 63-72

**Drawing Management**
```
GET /api/charts/drawings?symbol=VNM
Headers: Authorization: Bearer <token>
Response: { "data": Drawing[] }

POST /api/charts/drawings
Headers: Authorization: Bearer <token>
Request: Drawing
Response: { "id": number, "drawing": Drawing }

PUT /api/charts/drawings/:id
Headers: Authorization: Bearer <token>
Request: Drawing
Response: { "drawing": Drawing }

DELETE /api/charts/drawings/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/drawings/group
Headers: Authorization: Bearer <token>
Request: { "drawingIds": number[] }
Response: { "groupId": number }

DELETE /api/charts/drawings/group/:groupId
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/drawings/clone
Headers: Authorization: Bearer <token>
Request: { "fromSymbol": string, "toSymbol": string }
Response: { "clonedCount": number }

GET /api/charts/drawing-templates
Headers: Authorization: Bearer <token>
Response: { "data": DrawingTemplate[] }

POST /api/charts/drawing-templates
Headers: Authorization: Bearer <token>
Request: DrawingTemplate
Response: { "id": number }
```

**Indicator Templates**
```
GET /api/charts/indicator-templates
Headers: Authorization: Bearer <token>
Response: { "data": IndicatorTemplate[] }

GET /api/charts/indicator-templates/built-in
Headers: Authorization: Bearer <token>
Response: { "data": IndicatorTemplate[] }

POST /api/charts/indicator-templates
Headers: Authorization: Bearer <token>
Request: IndicatorTemplate
Response: { "id": number }

PUT /api/charts/indicator-templates/:id
Headers: Authorization: Bearer <token>
Request: IndicatorTemplate
Response: { "success": boolean }

DELETE /api/charts/indicator-templates/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/indicator-templates/:id/favorite
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/indicator-templates/:id/publish
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/indicator-templates/import
Headers: Authorization: Bearer <token>
Request: { "templateJSON": string }
Response: { "id": number }

GET /api/charts/indicator-templates/:id/export
Headers: Authorization: Bearer <token>
Response: { "templateJSON": string }
```

**Event Marks**
```
GET /api/charts/event-marks?symbol=VNM&startDate=<date>&endDate=<date>
Headers: Authorization: Bearer <token>
Response: { "data": EventMark[] }

POST /api/charts/event-marks
Headers: Authorization: Bearer <token>
Request: EventMark
Response: { "id": number }

PUT /api/charts/event-marks/:id
Headers: Authorization: Bearer <token>
Request: EventMark
Response: { "success": boolean }

DELETE /api/charts/event-marks/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/event-marks/sync/:symbol
Headers: Authorization: Bearer <token>
Response: { "syncedCount": number }

GET /api/charts/event-marks/upcoming?daysAhead=30
Headers: Authorization: Bearer <token>
Response: { "data": EventMark[] }

POST /api/charts/event-marks/:id/alert
Headers: Authorization: Bearer <token>
Request: { "daysBefore": number }
Response: { "id": number }

POST /api/charts/event-marks/bulk-import
Headers: Authorization: Bearer <token>
Request: { "events": EventMark[] }
Response: { "imported": number, "failed": number, "errors": string[] }

GET /api/charts/event-marks/:symbol/statistics
Headers: Authorization: Bearer <token>
Response: { "earningsBeats": number, "earningsMisses": number, "avgPriceMovementOnEarnings": number, "avgPriceMovementOnDividend": number }
```

**Chart Layouts**
```
GET /api/charts/layouts
Headers: Authorization: Bearer <token>
Response: { "data": ChartLayout[] }

GET /api/charts/layouts/active
Headers: Authorization: Bearer <token>
Response: ChartLayout | null

POST /api/charts/layouts
Headers: Authorization: Bearer <token>
Request: ChartLayout
Response: { "id": number }

PUT /api/charts/layouts/:id
Headers: Authorization: Bearer <token>
Request: ChartLayout
Response: { "success": boolean }

DELETE /api/charts/layouts/:id
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

POST /api/charts/layouts/:id/activate
Headers: Authorization: Bearer <token>
Response: { "success": boolean }

GET /api/charts/layouts/built-in
Headers: Authorization: Bearer <token>
Response: { "data": ChartLayout[] }
```

**Chart Preferences**
```
GET /api/charts/preferences/:symbol
Headers: Authorization: Bearer <token>
Response: { "chartType": string, "scaleMode": string, "scaleConfig": object, "timeframe": string }

PUT /api/charts/preferences/:symbol
Headers: Authorization: Bearer <token>
Request: { "chartType": string, "scaleMode": string, "scaleConfig": object, "timeframe": string }
Response: { "success": boolean }
```

**Quick Trading**
```
POST /api/trading/quick-order
Headers: Authorization: Bearer <token>
Request: { "symbol": string, "type": "buy" | "sell", "quantity": number, "price": number }
Response: { "transactionId": number, "success": boolean }

POST /api/trading/bracket-order
Headers: Authorization: Bearer <token>
Request: { "symbol": string, "entryPrice": number, "stopLoss": number, "takeProfit": number, "quantity": number }
Response: { "riskRewardRatio": number, "estimatedRisk": number, "estimatedReward": number }

POST /api/trading/position-size
Headers: Authorization: Bearer <token>
Request: { "riskAmount": number, "stopLossPercent": number, "currentPrice": number }
Response: { "suggestedQuantity": number }
```

**Watchlist Enhancements**
```
PUT /api/watchlists/:id/reorder
Headers: Authorization: Bearer <token>
Request: { "symbols": string[] }
Response: { "success": boolean }

POST /api/watchlists/:id/bulk-action
Headers: Authorization: Bearer <token>
Request: { "symbols": string[], "action": "add_alert" | "remove" | "move" | "export", "targetWatchlistId": number? }
Response: { "success": boolean, "affectedCount": number }

GET /api/watchlists/:id/export
Headers: Authorization: Bearer <token>
Response: CSV file download
```

**Advanced Time Features**
```
GET /api/market/ohlcv-seconds?symbol=VNM&interval=1s&startTime=<timestamp>&endTime=<timestamp>
Headers: Authorization: Bearer <token>
Response: { "data": OHLCVBar[], "available": boolean }

GET /api/market/bar-countdown?symbol=VNM&interval=1m
Headers: Authorization: Bearer <token>
Response: { "secondsRemaining": number, "nextCandleTime": timestamp }
```


---

## Requirements 83–95: Advanced Trading and Intelligence Features

### Architecture Additions

#### New Business Logic Layer Services

The following services extend the Business Logic Layer to support requirements 83–95:

- **SwingMax_Engine**: Swing trading signal generation, staged exit guidance, portfolio management, FOL warning, configurable stop-loss
- **AI_Stock_Picker**: Daily pre-market intraday pick list generation with accuracy tracking
- **Whale_Tracker**: Unusual volume detection, block trade feed, net foreign flow ranking
- **Daytrading_Center**: Intraday signal filtering by momentum/volume scores, session phase management
- **QuantAI_Alpha**: Four-factor quantitative model (momentum, value, quality, growth) with adjustable weights
- **Market_Movers_Service**: Real-time top-20 gainers, losers, most-active ranking with sector heatmap
- **Daily_Brief_Service**: AI-generated daily market summary via Multi_Agent_System, persisted per trading day

#### New Frontend Views

- **SwingMax_View**: Signal cards with staged exit guidance, portfolio tracker, header stats (win rate, annualized return, today count)
- **AI_Stock_Picker_View**: Daily pick cards with AI reasoning, accuracy metrics (win rate, avg return over trailing 30 days)
- **Whale_Tracker_View**: Foreign flow ranking table, unusual volume list, block trade feed
- **Daytrading_Center_View**: Intraday signal list, session phase indicator, session summary panel
- **Pattern_Detection_View**: Pattern cards with confidence scores, "New" badge, historical accuracy from KnowledgeBase
- **AI_Screener_View**: Enhanced screener with AI_Conviction_Score column, Top AI Picks section (top 10)
- **Stock_Monitor_View**: Unified alert feed (price/pattern/whale/signal), create/acknowledge actions
- **QuantAI_Alpha_View**: Weekly pick list with factor weight sliders, historical performance table
- **Backtest_Playground_View**: Visual strategy builder, equity curve with VN-Index overlay, trade log, stop-loss sensitivity panel
- **Market_Movers_View**: Gainers/losers/most-active tabs with sector mini-heatmap
- **Economic_Calendar_View**: Monthly grid + chronological list view with event type filters and impact badges
- **Daily_Brief_View**: AI-generated brief with language toggle (VI/EN) and rate-limited regenerate button

#### New Data Flow: SwingMax Signal Generation

1. SwingMax_Engine runs daily at 16:00 ICT after market close
2. Fetches full listing from Data_Source_Router, filters by ADTV ≥ 5B VND via LiquidityFilter
3. Calls SignalEngine.ScanMarket for composite signal scores
4. Selects top candidates with ConvictionScore ≥ threshold
5. Computes entry price, ATR-based stop-loss (default −15%), T1 and T2 targets
6. Checks FOL (foreign ownership limit) data via Data_Source_Router; excludes symbols at FOL limit, flags symbols within 5% of limit
7. Persists up to 5 opening signals in swingmax_signals table
8. SwingMaxPortfolio auto-selects top 5 by ConvictionScore passing liquidity + fundamental checks
9. Price monitoring job runs every 5 minutes during trading hours (09:00–15:00 ICT) to update signal statuses (partial_exit, closed) and enforce 30-day timeout

#### New Data Flow: Daily Brief Generation

1. Daily_Brief_Service triggers at 08:45 ICT on each trading day
2. Fetches VN-Index overnight context from Data_Source_Router
3. Queries SwingMax_Engine for current opening signals (top 3 by ConvictionScore)
4. Queries Whale_Tracker for prior session's top foreign flow events
5. Queries MarketCalendarService for today's events
6. Passes all context to Supervisor_Agent for synthesis (max 500 words, Vietnamese default)
7. Generates English translation via I18n_Service
8. Persists to daily_briefs table (unique per brief_date); subsequent requests serve cached version
9. If generation fails within 60 seconds, serves most recently persisted brief with a warning banner


### New Backend Modules (Requirements 83–95)

**SwingMax_Engine**
```go
type SignalStatus string
const (
    SignalOpening     SignalStatus = "opening"
    SignalPartialExit SignalStatus = "partial_exit"
    SignalClosed      SignalStatus = "closed"
)

type SwingMaxSignal struct {
    ID               int64       `json:"id"`
    Symbol           string      `json:"symbol"`
    Exchange         string      `json:"exchange"`
    EntryPrice       float64     `json:"entryPrice"`
    StopLossPrice    float64     `json:"stopLossPrice"`
    StopLossPercent  float64     `json:"stopLossPercent"` // configurable -5% to -25%
    SellTarget1      float64     `json:"sellTarget1"`
    SellTarget2      float64     `json:"sellTarget2"`
    UnrealizedReturn float64     `json:"unrealizedReturn"`
    PotentialGainT2  float64     `json:"potentialGainT2"`
    ConvictionScore  int         `json:"convictionScore"`
    SignalStatus     SignalStatus `json:"signalStatus"`
    FOLWarning       bool        `json:"folWarning"`
    GeneratedAt      time.Time   `json:"generatedAt"`
    T1HitAt          *time.Time  `json:"t1HitAt,omitempty"`
    ClosedAt         *time.Time  `json:"closedAt,omitempty"`
    ExitPrice        *float64    `json:"exitPrice,omitempty"`
    ExitReason       string      `json:"exitReason,omitempty"` // stop_loss, t2_hit, timeout
}

type SwingMaxPortfolio struct {
    Positions             []SwingMaxSignal `json:"positions"`
    TotalUnrealizedReturn float64          `json:"totalUnrealizedReturn"`
    TotalRealisedReturn   float64          `json:"totalRealisedReturn"`
    WinRate               float64          `json:"winRate"`
    AvgHoldingDays        float64          `json:"avgHoldingDays"`
}

type SwingMaxEngine struct {
    signalEngine    *SignalEngine
    liquidityFilter *LiquidityFilter
    router          *DataSourceRouter
    db              *sql.DB
}

func (s *SwingMaxEngine) GenerateSignals(ctx context.Context) ([]SwingMaxSignal, error)
func (s *SwingMaxEngine) GetOpeningSignals(ctx context.Context) ([]SwingMaxSignal, error)
func (s *SwingMaxEngine) GetClosedSignals(ctx context.Context) ([]SwingMaxSignal, error)
func (s *SwingMaxEngine) UpdateSignalStatuses(ctx context.Context) error
func (s *SwingMaxEngine) GetPortfolio(ctx context.Context) (SwingMaxPortfolio, error)
func (s *SwingMaxEngine) GetHeaderStats(ctx context.Context) (map[string]interface{}, error)
func (s *SwingMaxEngine) SetStopLossPercent(ctx context.Context, percent float64) error
```

**AI_Stock_Picker**
```go
type DailyPick struct {
    ID               int64      `json:"id"`
    Symbol           string     `json:"symbol"`
    Exchange         string     `json:"exchange"`
    EntryLow         float64    `json:"entryLow"`
    EntryHigh        float64    `json:"entryHigh"`
    IntradayTarget   float64    `json:"intradayTarget"`
    IntradayStopLoss float64    `json:"intradayStopLoss"`
    ExpectedReturn   float64    `json:"expectedReturn"`
    Reasoning        string     `json:"reasoning"` // max 100 words
    ConvictionScore  int        `json:"convictionScore"`
    GeneratedAt      time.Time  `json:"generatedAt"`
    ExitPrice        *float64   `json:"exitPrice,omitempty"`
    RealisedReturn   *float64   `json:"realisedReturn,omitempty"`
    IsProfitable     *bool      `json:"isProfitable,omitempty"`
    ClosedAt         *time.Time `json:"closedAt,omitempty"`
}

type AIStockPicker struct {
    signalEngine    *SignalEngine
    liquidityFilter *LiquidityFilter
    router          *DataSourceRouter
    db              *sql.DB
}

func (a *AIStockPicker) GenerateDailyPicks(ctx context.Context) ([]DailyPick, error)
func (a *AIStockPicker) GetTodayPicks(ctx context.Context) ([]DailyPick, error)
func (a *AIStockPicker) CloseSessionPicks(ctx context.Context) error
func (a *AIStockPicker) GetAccuracyMetrics(ctx context.Context, trailingDays int) (map[string]interface{}, error)
```

**Whale_Tracker**
```go
type WhaleEvent struct {
    ID        int64     `json:"id"`
    Symbol    string    `json:"symbol"`
    Exchange  string    `json:"exchange"`
    EventType string    `json:"eventType"` // unusual_volume, block_trade
    EventValue float64  `json:"eventValue"` // volume ratio or trade value in VND
    Price     float64   `json:"price"`
    Timestamp time.Time `json:"timestamp"`
    BuyerSide *bool     `json:"buyerSide,omitempty"`
}

type ForeignFlowData struct {
    Symbol        string  `json:"symbol"`
    Exchange      string  `json:"exchange"`
    NetBuyValue   float64 `json:"netBuyValue"`
    NetSellValue  float64 `json:"netSellValue"`
    NetFlow       float64 `json:"netFlow"`
    FlowDirection string  `json:"flowDirection"` // net_buy, net_sell, neutral
}

type WhaleTracker struct {
    router *DataSourceRouter
    db     *sql.DB
}

func (w *WhaleTracker) DetectUnusualVolume(ctx context.Context) ([]WhaleEvent, error)
func (w *WhaleTracker) DetectBlockTrades(ctx context.Context) ([]WhaleEvent, error)
func (w *WhaleTracker) GetForeignFlows(ctx context.Context, exchange string) ([]ForeignFlowData, error)
func (w *WhaleTracker) GetRecentBlockTrades(ctx context.Context, limit int) ([]WhaleEvent, error)
func (w *WhaleTracker) GetUnusualVolumeRanking(ctx context.Context) ([]WhaleEvent, error)
```

**QuantAI_Alpha**
```go
type FactorWeights struct {
    Momentum float64 `json:"momentum"` // default 0.30
    Value    float64 `json:"value"`    // default 0.25
    Quality  float64 `json:"quality"`  // default 0.25
    Growth   float64 `json:"growth"`   // default 0.20
}

type AlphaPick struct {
    Symbol        string    `json:"symbol"`
    Exchange      string    `json:"exchange"`
    Sector        string    `json:"sector"`
    AlphaScore    float64   `json:"alphaScore"`
    MomentumScore float64   `json:"momentumScore"`
    ValueScore    float64   `json:"valueScore"`
    QualityScore  float64   `json:"qualityScore"`
    GrowthScore   float64   `json:"growthScore"`
    CurrentPrice  float64   `json:"currentPrice"`
    Rationale     string    `json:"rationale"`
    GeneratedAt   time.Time `json:"generatedAt"`
}

type QuantAIAlpha struct {
    router          *DataSourceRouter
    liquidityFilter *LiquidityFilter
    db              *sql.DB
}

func (q *QuantAIAlpha) GenerateWeeklyPicks(ctx context.Context, weights FactorWeights) ([]AlphaPick, error)
func (q *QuantAIAlpha) GetCurrentPicks(ctx context.Context) ([]AlphaPick, error)
func (q *QuantAIAlpha) ComputeAlphaScore(ctx context.Context, symbol string, weights FactorWeights) (AlphaPick, error)
func (q *QuantAIAlpha) GetHistoricalPerformance(ctx context.Context) ([]map[string]interface{}, error)
```

**Market_Movers_Service**
```go
type MarketMover struct {
    Symbol        string  `json:"symbol"`
    Exchange      string  `json:"exchange"`
    CurrentPrice  float64 `json:"currentPrice"`
    PriceChange   float64 `json:"priceChange"`
    ChangePercent float64 `json:"changePercent"`
    Volume        int64   `json:"volume"`
    SessionValue  float64 `json:"sessionValue"`
}

type MarketMoversService struct {
    router *DataSourceRouter
    cache  *Cache
}

func (m *MarketMoversService) GetTopGainers(ctx context.Context, exchange string, sector string, limit int) ([]MarketMover, error)
func (m *MarketMoversService) GetTopLosers(ctx context.Context, exchange string, sector string, limit int) ([]MarketMover, error)
func (m *MarketMoversService) GetMostActive(ctx context.Context, exchange string, sector string, limit int) ([]MarketMover, error)
func (m *MarketMoversService) GetSectorMiniHeatmap(ctx context.Context) ([]map[string]interface{}, error)
```

**Daily_Brief_Service**
```go
type DailyBrief struct {
    ID          int64     `json:"id"`
    Content     string    `json:"content"`   // Vietnamese, max 500 words
    ContentEN   string    `json:"contentEN"` // English version
    GeneratedAt time.Time `json:"generatedAt"`
    BriefDate   time.Time `json:"briefDate"`
}

type DailyBriefService struct {
    multiAgentSystem *MultiAgentSystem
    swingMaxEngine   *SwingMaxEngine
    whaleTracker     *WhaleTracker
    calendarService  *MarketCalendarService
    db               *sql.DB
}

func (d *DailyBriefService) GenerateBrief(ctx context.Context) (DailyBrief, error)
func (d *DailyBriefService) GetTodayBrief(ctx context.Context) (DailyBrief, error)
func (d *DailyBriefService) RegenerateBrief(ctx context.Context, userID int64) (DailyBrief, error)
```

### New Database Tables (Requirements 83–95)

**swingmax_signals**
```sql
CREATE TABLE swingmax_signals (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(10) NOT NULL,
    entry_price DECIMAL(20, 2) NOT NULL,
    stop_loss_price DECIMAL(20, 2) NOT NULL,
    stop_loss_percent DECIMAL(5, 2) NOT NULL DEFAULT -15.0,
    sell_target_1 DECIMAL(20, 2) NOT NULL,
    sell_target_2 DECIMAL(20, 2) NOT NULL,
    conviction_score INT NOT NULL CHECK (conviction_score >= 0 AND conviction_score <= 100),
    signal_status VARCHAR(20) NOT NULL DEFAULT 'opening',
    fol_warning BOOLEAN DEFAULT FALSE,
    generated_at TIMESTAMP NOT NULL,
    t1_hit_at TIMESTAMP,
    closed_at TIMESTAMP,
    exit_price DECIMAL(20, 2),
    exit_reason VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_signal_status CHECK (signal_status IN ('opening', 'partial_exit', 'closed'))
);
CREATE INDEX idx_swingmax_symbol ON swingmax_signals(symbol);
CREATE INDEX idx_swingmax_status ON swingmax_signals(signal_status);
CREATE INDEX idx_swingmax_generated ON swingmax_signals(generated_at);
```

**daily_picks**
```sql
CREATE TABLE daily_picks (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(10) NOT NULL,
    entry_low DECIMAL(20, 2) NOT NULL,
    entry_high DECIMAL(20, 2) NOT NULL,
    intraday_target DECIMAL(20, 2) NOT NULL,
    intraday_stop_loss DECIMAL(20, 2) NOT NULL,
    expected_return DECIMAL(5, 2) NOT NULL,
    reasoning TEXT NOT NULL,
    conviction_score INT NOT NULL,
    generated_at TIMESTAMP NOT NULL,
    pick_date DATE NOT NULL,
    exit_price DECIMAL(20, 2),
    realised_return DECIMAL(10, 4),
    is_profitable BOOLEAN,
    closed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_daily_picks_date ON daily_picks(pick_date);
CREATE INDEX idx_daily_picks_symbol ON daily_picks(symbol);
```

**whale_events**
```sql
CREATE TABLE whale_events (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(10) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_value DECIMAL(20, 2) NOT NULL,
    price DECIMAL(20, 2) NOT NULL,
    buyer_side BOOLEAN,
    event_timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_event_type CHECK (event_type IN ('unusual_volume', 'block_trade'))
);
CREATE INDEX idx_whale_events_symbol ON whale_events(symbol);
CREATE INDEX idx_whale_events_timestamp ON whale_events(event_timestamp);
CREATE INDEX idx_whale_events_type ON whale_events(event_type);
```

**alpha_picks**
```sql
CREATE TABLE alpha_picks (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(10) NOT NULL,
    sector VARCHAR(50),
    alpha_score DECIMAL(5, 2) NOT NULL,
    momentum_score DECIMAL(5, 2) NOT NULL,
    value_score DECIMAL(5, 2) NOT NULL,
    quality_score DECIMAL(5, 2) NOT NULL,
    growth_score DECIMAL(5, 2) NOT NULL,
    current_price DECIMAL(20, 2) NOT NULL,
    rationale TEXT,
    generated_at TIMESTAMP NOT NULL,
    pick_week DATE NOT NULL, -- Monday of the pick week
    return_1w DECIMAL(10, 4),
    return_1m DECIMAL(10, 4),
    return_3m DECIMAL(10, 4),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_alpha_picks_week ON alpha_picks(pick_week);
CREATE INDEX idx_alpha_picks_symbol ON alpha_picks(symbol);
```

**daily_briefs**
```sql
CREATE TABLE daily_briefs (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    content_en TEXT NOT NULL,
    generated_at TIMESTAMP NOT NULL,
    brief_date DATE NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_daily_briefs_date ON daily_briefs(brief_date);
```
