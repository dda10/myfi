package model

import "time"

// --- Transaction (from transaction_ledger.go) ---

// Transaction represents a single stock transaction in the ledger.
type Transaction struct {
	ID              int64           `json:"id"`
	UserID          int64           `json:"userId"`
	Symbol          string          `json:"symbol"`
	Quantity        float64         `json:"quantity"`
	UnitPrice       float64         `json:"unitPrice"`
	TotalValue      float64         `json:"totalValue"`
	TransactionDate time.Time       `json:"transactionDate"`
	TransactionType TransactionType `json:"transactionType"`
	Notes           string          `json:"notes,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
}

// --- Asset (from asset_registry.go) ---

// Holding represents a user's stock holding.
type Holding struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"userId"`
	Symbol          string    `json:"symbol"`
	Quantity        float64   `json:"quantity"`
	AverageCost     float64   `json:"averageCost"`
	AcquisitionDate time.Time `json:"acquisitionDate"`
	Account         string    `json:"account"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// --- Portfolio Engine (from portfolio_engine.go) ---

// PortfolioSummary contains the full portfolio overview for a user.
type PortfolioSummary struct {
	NAV              float64         `json:"nav"`
	NAVChange24h     float64         `json:"navChange24h"`
	NAVChangePercent float64         `json:"navChangePercent"`
	Holdings         []HoldingDetail `json:"holdings"`
}

// HoldingDetail contains per-holding valuation and P&L data.
type HoldingDetail struct {
	Holding         Holding `json:"holding"`
	CurrentPrice    float64 `json:"currentPrice"`
	MarketValue     float64 `json:"marketValue"`
	UnrealizedPL    float64 `json:"unrealizedPL"`
	UnrealizedPLPct float64 `json:"unrealizedPLPct"`
}

// SellResult contains the outcome of a sell transaction.
type SellResult struct {
	TransactionID int64   `json:"transactionId"`
	RealizedPL    float64 `json:"realizedPL"`
}

// --- Screener (from screener_service.go) ---

// ScreenerFilters defines all filter criteria for the stock screener.
type ScreenerFilters struct {
	MinPE            *float64      `json:"minPE,omitempty"`
	MaxPE            *float64      `json:"maxPE,omitempty"`
	MinPB            *float64      `json:"minPB,omitempty"`
	MaxPB            *float64      `json:"maxPB,omitempty"`
	MinMarketCap     *float64      `json:"minMarketCap,omitempty"`
	MinEVEBITDA      *float64      `json:"minEVEBITDA,omitempty"`
	MaxEVEBITDA      *float64      `json:"maxEVEBITDA,omitempty"`
	MinROE           *float64      `json:"minROE,omitempty"`
	MaxROE           *float64      `json:"maxROE,omitempty"`
	MinROA           *float64      `json:"minROA,omitempty"`
	MaxROA           *float64      `json:"maxROA,omitempty"`
	MinRevenueGrowth *float64      `json:"minRevenueGrowth,omitempty"`
	MaxRevenueGrowth *float64      `json:"maxRevenueGrowth,omitempty"`
	MinProfitGrowth  *float64      `json:"minProfitGrowth,omitempty"`
	MaxProfitGrowth  *float64      `json:"maxProfitGrowth,omitempty"`
	MinDivYield      *float64      `json:"minDivYield,omitempty"`
	MaxDivYield      *float64      `json:"maxDivYield,omitempty"`
	MinDebtToEquity  *float64      `json:"minDebtToEquity,omitempty"`
	MaxDebtToEquity  *float64      `json:"maxDebtToEquity,omitempty"`
	Sectors          []ICBSector   `json:"sectors,omitempty"`
	Exchanges        []string      `json:"exchanges,omitempty"`
	SectorTrends     []SectorTrend `json:"sectorTrends,omitempty"`
	SortBy           string        `json:"sortBy"`
	SortOrder        string        `json:"sortOrder"`
	Page             int           `json:"page"`
	PageSize         int           `json:"pageSize"`
}

// ScreenerResult represents a single stock in the screener output.
type ScreenerResult struct {
	Symbol        string      `json:"symbol"`
	Exchange      string      `json:"exchange"`
	Sector        ICBSector   `json:"sector"`
	SectorName    string      `json:"sectorName"`
	MarketCap     float64     `json:"marketCap"`
	PE            float64     `json:"pe"`
	PB            float64     `json:"pb"`
	EVEBITDA      float64     `json:"evEbitda"`
	ROE           float64     `json:"roe"`
	ROA           float64     `json:"roa"`
	RevenueGrowth float64     `json:"revenueGrowth"`
	ProfitGrowth  float64     `json:"profitGrowth"`
	DivYield      float64     `json:"divYield"`
	DebtToEquity  float64     `json:"debtToEquity"`
	SectorTrend   SectorTrend `json:"sectorTrend"`
}

// ScreenerResponse wraps paginated screener results.
type ScreenerResponse struct {
	Data       []ScreenerResult `json:"data"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// FilterPreset represents a saved set of screener filter criteria.
type FilterPreset struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"userId"`
	Name      string          `json:"name"`
	Filters   ScreenerFilters `json:"filters"`
	CreatedAt time.Time       `json:"createdAt"`
}

// --- Macro (from macro_service.go) ---

// MacroIndicator represents a single macroeconomic indicator.
type MacroIndicator struct {
	Name        string  `json:"name"`
	Code        string  `json:"code"`
	Value       float64 `json:"value"`
	Unit        string  `json:"unit"`
	Period      string  `json:"period"`
	Country     string  `json:"country"`
	Source      string  `json:"source"`
	Description string  `json:"description"`
}

// MacroData contains all macroeconomic indicators.
type MacroData struct {
	Indicators []MacroIndicator `json:"indicators"`
	IsStale    bool             `json:"isStale"`
}

// --- Watchlist (from watchlist_service.go) ---

// Watchlist represents a named watchlist belonging to a user.
type Watchlist struct {
	ID        int               `json:"id"`
	UserID    int               `json:"userId"`
	Name      string            `json:"name"`
	Symbols   []WatchlistSymbol `json:"symbols"`
	CreatedAt time.Time         `json:"createdAt"`
}

// WatchlistSymbol represents a symbol entry within a watchlist.
type WatchlistSymbol struct {
	ID              int       `json:"id"`
	WatchlistID     int       `json:"watchlistId"`
	Symbol          string    `json:"symbol"`
	Position        int       `json:"position"`
	PriceAlertAbove *float64  `json:"priceAlertAbove,omitempty"`
	PriceAlertBelow *float64  `json:"priceAlertBelow,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// --- Performance Engine (from performance_engine.go) ---

// PerformanceMetrics contains all portfolio performance analytics.
type PerformanceMetrics struct {
	TWR                 float64       `json:"twr"`
	MWRR                float64       `json:"mwrr"`
	EquityCurve         []NAVSnapshot `json:"equityCurve"`
	BenchmarkComparison BenchmarkData `json:"benchmarkComparison"`
}

// NAVSnapshot represents a daily NAV data point for the equity curve.
type NAVSnapshot struct {
	Date time.Time `json:"date"`
	NAV  float64   `json:"nav"`
}

// BenchmarkData contains benchmark comparison results.
type BenchmarkData struct {
	VNIndexReturn   float64 `json:"vnIndexReturn"`
	VN30Return      float64 `json:"vn30Return"`
	PortfolioReturn float64 `json:"portfolioReturn"`
	Alpha           float64 `json:"alpha"`
}

// CashFlowEvent represents an external cash flow for MWRR calculation.
type CashFlowEvent struct {
	Date   time.Time
	Amount float64 // positive = inflow, negative = outflow
}

// --- Comparison Engine (from comparison_engine.go) ---

// TimePeriod represents a comparison time period
type TimePeriod string

const (
	Period3M TimePeriod = "3M"
	Period6M TimePeriod = "6M"
	Period1Y TimePeriod = "1Y"
	Period3Y TimePeriod = "3Y"
	Period5Y TimePeriod = "5Y"
)

// MaxComparisonStocks is the maximum number of stocks that can be compared simultaneously
const MaxComparisonStocks = 10

// ValuationPoint represents a single valuation data point for a symbol
type ValuationPoint struct {
	Timestamp time.Time `json:"timestamp"`
	PE        float64   `json:"pe"`
	PB        float64   `json:"pb"`
}

// ValuationSeries holds the valuation time-series for one symbol
type ValuationSeries struct {
	Symbol string           `json:"symbol"`
	Data   []ValuationPoint `json:"data"`
}

// ValuationResult is the response for a valuation comparison
type ValuationResult struct {
	Series   []ValuationSeries `json:"series"`
	Period   TimePeriod        `json:"period"`
	Warnings []string          `json:"warnings,omitempty"`
}

// PerformancePoint represents a normalized return data point
type PerformancePoint struct {
	Timestamp     time.Time `json:"timestamp"`
	ReturnPercent float64   `json:"returnPercent"`
}

// PerformanceSeries holds the performance time-series for one symbol
type PerformanceSeries struct {
	Symbol string             `json:"symbol"`
	Data   []PerformancePoint `json:"data"`
}

// PerformanceResult is the response for a performance comparison
type PerformanceResult struct {
	Series   []PerformanceSeries `json:"series"`
	Period   TimePeriod          `json:"period"`
	Warnings []string            `json:"warnings,omitempty"`
}

// CorrelationPair holds the correlation between two symbols
type CorrelationPair struct {
	SymbolA     string  `json:"symbolA"`
	SymbolB     string  `json:"symbolB"`
	Correlation float64 `json:"correlation"`
}

// CorrelationResult is the response for a correlation comparison
type CorrelationResult struct {
	Symbols  []string          `json:"symbols"`
	Matrix   [][]float64       `json:"matrix"`
	Pairs    []CorrelationPair `json:"pairs"`
	Period   TimePeriod        `json:"period"`
	Warnings []string          `json:"warnings,omitempty"`
}

// SectorStocksResult holds stocks grouped by sector
type SectorStocksResult struct {
	Sector  ICBSector `json:"sector"`
	Symbols []string  `json:"symbols"`
}

// --- Risk Service (from risk_service.go) ---

// RiskMetrics contains portfolio-level and per-holding risk analytics.
// Requirement 27: Sharpe ratio, max drawdown, beta, volatility, VaR, risk contribution.
type RiskMetrics struct {
	SharpeRatio      float64            `json:"sharpeRatio"`
	MaxDrawdown      float64            `json:"maxDrawdown"`
	Beta             float64            `json:"beta"`
	Volatility       float64            `json:"volatility"`
	VaR95            float64            `json:"var95"`
	RiskContribution map[string]float64 `json:"riskContribution"` // symbol -> contribution %
}
