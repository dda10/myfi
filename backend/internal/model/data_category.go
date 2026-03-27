package model

// DataCategory represents the type of data being requested
type DataCategory string

const (
	PriceQuotes     DataCategory = "price_quotes"
	OHLCVHistory    DataCategory = "ohlcv_history"
	IntradayData    DataCategory = "intraday_data"
	OrderBook       DataCategory = "order_book"
	CompanyOverview DataCategory = "company_overview"
	Shareholders    DataCategory = "shareholders"
	Officers        DataCategory = "officers"
	News            DataCategory = "news"
	IncomeStatement DataCategory = "income_statement"
	BalanceSheet    DataCategory = "balance_sheet"
	CashFlow        DataCategory = "cash_flow"
	FinancialRatios DataCategory = "financial_ratios"
)

// SourcePreference defines the primary and fallback sources for a data category
type SourcePreference struct {
	Category DataCategory
	Primary  string // "VCI" or "KBS"
	Fallback string
}
