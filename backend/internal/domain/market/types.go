package market

import (
	"context"
	"fmt"
	"myfi-backend/internal/infra"
	"time"

	"github.com/dda10/vnstock-go"
)

// AssetType represents the type of a financial asset.
// EziStock is stock-only: only VNStock is supported.
type AssetType string

const (
	VNStock AssetType = "vnstock"
)

// ValidAssetTypes contains all supported asset types for validation.
var ValidAssetTypes = map[AssetType]bool{
	VNStock: true,
}

// ValidateAssetType checks if the given asset type is supported.
func ValidateAssetType(at AssetType) error {
	if ValidAssetTypes[at] {
		return nil
	}
	return fmt.Errorf("invalid asset type %q", at)
}

// --- Listing Data Types ---

// ListingData contains all listing-related information.
type ListingData struct {
	Symbols   []vnstock.ListingRecord `json:"symbols"`
	Indices   []vnstock.IndexRecord   `json:"indices"`
	Bonds     []BondInfo              `json:"bonds"`
	Exchanges []ExchangeInfo          `json:"exchanges"`
	IsStale   bool                    `json:"isStale"`
}

// BondInfo represents government bond information.
type BondInfo struct {
	Name     string  `json:"name"`
	Tenor    string  `json:"tenor"`
	Yield    float64 `json:"yield"`
	Currency string  `json:"currency"`
}

// ExchangeInfo represents a stock exchange.
type ExchangeInfo struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	StockCount  int    `json:"stockCount"`
}

// --- Company Data Types ---

// CompanyData contains all company-related information.
type CompanyData struct {
	Profile       vnstock.CompanyProfile `json:"profile"`
	Shareholders  []vnstock.Shareholder  `json:"shareholders"`
	Officers      []vnstock.Officer      `json:"officers"`
	Subsidiaries  []vnstock.Subsidiary   `json:"subsidiaries,omitempty"`
	Events        []vnstock.CompanyEvent `json:"events,omitempty"`
	News          []CompanyNewsItem      `json:"news"`
	InsiderTrades []vnstock.InsiderTrade `json:"insiderTrades,omitempty"`
	IsStale       bool                   `json:"isStale"`
}

// CompanyNewsItem represents a company news article.
type CompanyNewsItem struct {
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	URL       string    `json:"url"`
	Source    string    `json:"source"`
	Published time.Time `json:"published"`
}

// --- Financial Report Types ---

// FinancialReportData contains all financial report data for a symbol.
type FinancialReportData struct {
	IncomeStatements []vnstock.FinancialPeriod `json:"incomeStatements"`
	BalanceSheets    []vnstock.FinancialPeriod `json:"balanceSheets"`
	CashFlows        []vnstock.FinancialPeriod `json:"cashFlows"`
	Ratios           []vnstock.FinancialPeriod `json:"ratios"`
	IsStale          bool                      `json:"isStale"`
}

// --- Trading Statistics Types ---

// PriceQuote represents a price quote for a Vietnamese stock.
type PriceQuote struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
	Volume        int64     `json:"volume"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	IsStale       bool      `json:"isStale"`
}

// StockQuote is an alias for PriceQuote for backward compatibility.
type StockQuote = PriceQuote

// OHLCVBar represents a single OHLCV data point.
type OHLCVBar struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// TradingStatistics contains trading data for a symbol.
type TradingStatistics struct {
	Quote     *PriceQuote    `json:"quote,omitempty"`
	History   []OHLCVBar     `json:"history,omitempty"`
	Intraday  []OHLCVBar     `json:"intraday,omitempty"`
	OrderBook *OrderBookData `json:"orderBook,omitempty"`
	IsStale   bool           `json:"isStale"`
}

// OrderBookData represents the order book / price depth for a symbol.
type OrderBookData struct {
	Symbol string          `json:"symbol"`
	Bids   []OrderBookItem `json:"bids"`
	Asks   []OrderBookItem `json:"asks"`
}

// OrderBookItem represents a single bid or ask entry.
type OrderBookItem struct {
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

// --- Market Statistics Types ---

// MarketStatistics contains market-level statistics.
type MarketStatistics struct {
	Indices        []vnstock.IndexRecord `json:"indices"`
	SectorIndices  []SectorPerformance   `json:"sectorIndices"`
	MarketBreadth  *MarketBreadth        `json:"marketBreadth"`
	ForeignTrading []ForeignTradingData  `json:"foreignTrading"`
	IsStale        bool                  `json:"isStale"`
}

// MarketBreadth represents advancing vs declining stocks.
type MarketBreadth struct {
	Advancing  int     `json:"advancing"`
	Declining  int     `json:"declining"`
	Unchanged  int     `json:"unchanged"`
	AdvDecLine float64 `json:"advDecLine"`
}

// ForeignTradingData represents foreign investor trading data.
type ForeignTradingData struct {
	Symbol     string  `json:"symbol"`
	BuyVolume  int64   `json:"buyVolume"`
	SellVolume int64   `json:"sellVolume"`
	NetVolume  int64   `json:"netVolume"`
	BuyValue   float64 `json:"buyValue"`
	SellValue  float64 `json:"sellValue"`
	NetValue   float64 `json:"netValue"`
}

// --- Valuation Metrics Types ---

// ValuationMetrics contains valuation data at various levels.
type ValuationMetrics struct {
	Market  *MarketValuation  `json:"market,omitempty"`
	Sectors []SectorValuation `json:"sectors,omitempty"`
	Stock   *StockValuation   `json:"stock,omitempty"`
	IsStale bool              `json:"isStale"`
}

// MarketValuation represents market-level valuation metrics.
type MarketValuation struct {
	PE       float64 `json:"pe"`
	PB       float64 `json:"pb"`
	EVEBITDA float64 `json:"evEbitda"`
	DivYield float64 `json:"divYield"`
}

// SectorValuation represents sector-level valuation metrics.
type SectorValuation struct {
	Sector   ICBSector `json:"sector"`
	Name     string    `json:"name"`
	PE       float64   `json:"pe"`
	PB       float64   `json:"pb"`
	EVEBITDA float64   `json:"evEbitda"`
	DivYield float64   `json:"divYield"`
}

// StockValuation represents stock-level valuation metrics.
type StockValuation struct {
	Symbol    string  `json:"symbol"`
	PE        float64 `json:"pe"`
	PB        float64 `json:"pb"`
	EVEBITDA  float64 `json:"evEbitda"`
	DivYield  float64 `json:"divYield"`
	MarketCap float64 `json:"marketCap"`
}

// --- Macro Types ---

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

// --- Handlers ---

// LiquidityChecker is an interface for checking stock liquidity whitelist status.
// Uses any for GetWhitelist return to avoid importing model package.
type LiquidityChecker interface {
	IsWhitelisted(symbol string) bool
	GetWhitelist() any
	RefreshAll(ctx context.Context) error
}

// Handlers holds market-domain service dependencies for HTTP handler methods.
type Handlers struct {
	DataSourceRouter    *infra.DataSourceRouter
	PriceService        *PriceService
	SectorService       *SectorService
	MarketDataService   *MarketDataService
	MacroService        *MacroService
	SearchService       *SearchService
	WorldMarketService  *WorldMarketService
	GoldPriceService    *GoldPriceService
	ExchangeRateService *ExchangeRateService
	TradingHoursService *TradingHoursService
	LiquidityFilter     LiquidityChecker
	GRPCClient          *infra.GRPCClient
}
