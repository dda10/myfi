package model

import (
	"time"

	"myfi-backend/internal/domain/screener"
	"myfi-backend/internal/domain/watchlist"
)

// --- Transaction (from transaction_ledger.go) ---

// Asset represents a user's financial asset holding with cost basis tracking.
type Asset struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"userId"`
	AssetType       AssetType `json:"assetType"`
	Symbol          string    `json:"symbol"`
	Quantity        float64   `json:"quantity"`
	AverageCost     float64   `json:"averageCost"`
	AcquisitionDate time.Time `json:"acquisitionDate"`
	Account         string    `json:"account"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Transaction represents a single stock transaction in the ledger.
type Transaction struct {
	ID              int64           `json:"id"`
	UserID          string          `json:"userId"`
	AssetType       AssetType       `json:"assetType"`
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
	UserID          string    `json:"userId"`
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

// --- Screener (canonical definition in domain/screener) ---

type ScreenerFilters = screener.ScreenerFilters
type ScreenerResult = screener.ScreenerResult
type ScreenerResponse = screener.ScreenerResponse
type FilterPreset = screener.FilterPreset

// --- Watchlist (canonical definition in domain/watchlist) ---

type Watchlist = watchlist.Watchlist
type WatchlistSymbol = watchlist.WatchlistSymbol

// --- Performance Engine (from performance_engine.go) ---

// PerformanceMetrics contains all portfolio performance analytics.
type PerformanceMetrics struct {
	TWR                 float64               `json:"twr"`
	MWRR                float64               `json:"mwrr"`
	EquityCurve         []NAVSnapshot         `json:"equityCurve"`
	BenchmarkComparison BenchmarkData         `json:"benchmarkComparison"`
	PerformanceByType   map[AssetType]float64 `json:"performanceByType,omitempty"`
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
type RiskMetrics struct {
	SharpeRatio      float64            `json:"sharpeRatio"`
	MaxDrawdown      float64            `json:"maxDrawdown"`
	Beta             float64            `json:"beta"`
	Volatility       float64            `json:"volatility"`
	VaR95            float64            `json:"var95"`
	RiskContribution map[string]float64 `json:"riskContribution"`
}
