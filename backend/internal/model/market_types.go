package model

import (
	"time"

	"github.com/dda10/vnstock-go"
)

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
	Profile      vnstock.CompanyProfile `json:"profile"`
	Shareholders []vnstock.Shareholder  `json:"shareholders"`
	Officers     []vnstock.Officer      `json:"officers"`
	News         []CompanyNewsItem      `json:"news"`
	IsStale      bool                   `json:"isStale"`
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
