package market

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
	PriceBoard      DataCategory = "price_board"
	PriceDepth      DataCategory = "price_depth"
	Screener        DataCategory = "screener"
	GoldPrice       DataCategory = "gold_price"
	FXRate          DataCategory = "fx_rate"
	WorldIndices    DataCategory = "world_indices"
)

// SourcePreference defines the primary and fallback sources for a data category.
// Fallbacks is an ordered list of fallback sources tried in sequence.
type SourcePreference struct {
	Category  DataCategory
	Primary   string   // "VCI", "KBS", "VND", "ENTRADE", "CAFEF", "VND_FINFO", "MSN", "GOLD", "FMARKET"
	Fallbacks []string // ordered fallback chain
}

// Fallback returns the first fallback source for backward compatibility.
func (sp SourcePreference) Fallback() string {
	if len(sp.Fallbacks) > 0 {
		return sp.Fallbacks[0]
	}
	return ""
}
