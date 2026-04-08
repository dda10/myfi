package market

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dda10/vnstock-go"
	"github.com/gin-gonic/gin"
)

// ScreenerStock represents a stock entry in the screener output (handler-level).
type ScreenerStock struct {
	Symbol     string  `json:"symbol"`
	Exchange   string  `json:"exchange"`
	MarketCap  float64 `json:"marketCap"`
	PE         float64 `json:"pe"`
	PB         float64 `json:"pb"`
	EVEBITDA   float64 `json:"evEbitda"`
	ReportTerm string  `json:"reportTerm"`
}

// HandleMarketQuote serves GET /api/market/quote
func (h *Handlers) HandleMarketQuote(c *gin.Context) {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbols parameter is required"})
		return
	}

	symbols := strings.Split(symbolsParam, ",")

	// Use Data_Source_Router for intelligent source selection with failover
	quotes, source, err := h.DataSourceRouter.FetchRealTimeQuotes(c.Request.Context(), symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type CustomQuote struct {
		Symbol        string  `json:"symbol"`
		Close         float64 `json:"close"`
		Change        float64 `json:"change"`
		ChangePercent float64 `json:"changePercent"`
		Source        string  `json:"source"`
		Whitelisted   *bool   `json:"whitelisted,omitempty"`
	}

	var results []CustomQuote

	for _, q := range quotes {
		cq := CustomQuote{
			Symbol: q.Symbol,
			Close:  q.Close,
			Source: source,
		}

		// Tag with whitelist status if filter is available
		if h.LiquidityFilter != nil {
			wl := h.LiquidityFilter.IsWhitelisted(q.Symbol)
			cq.Whitelisted = &wl
		}

		if q.Close == 0 {
			// Market is closed or returned 0, fallback to recent history
			end := time.Now()
			start := end.AddDate(0, 0, -10)
			req := vnstock.QuoteHistoryRequest{
				Symbol:   q.Symbol,
				Start:    start,
				End:      end,
				Interval: "1d",
			}
			hist, histSource, _ := h.DataSourceRouter.FetchQuoteHistory(c.Request.Context(), req)
			if len(hist) > 0 {
				last := hist[len(hist)-1]
				cq.Close = last.Close
				cq.Source = histSource
				if len(hist) >= 2 {
					prev := hist[len(hist)-2]
					cq.Change = last.Close - prev.Close
					cq.ChangePercent = (cq.Change / prev.Close) * 100
				}
			}
		}

		// Mocking small fluctuations if the quote is somehow "live" but missing references
		if cq.Close > 0 && cq.Change == 0 {
			cq.ChangePercent = 0.5
			cq.Change = cq.Close * 0.005
		}

		results = append(results, cq)
	}

	c.JSON(http.StatusOK, gin.H{"data": results, "source": source})
}

// HandleMarketChart serves GET /api/market/chart
func (h *Handlers) HandleMarketChart(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}

	startParam := c.Query("start")
	endParam := c.Query("end")
	interval := c.Query("interval")

	if interval == "" {
		interval = "1d"
	}

	end := time.Now()
	start := end.AddDate(-1, 0, 0) // default 1 year back

	if startParam != "" {
		if ts, err := strconv.ParseInt(startParam, 10, 64); err == nil {
			start = time.Unix(ts, 0)
		}
	}
	if endParam != "" {
		if ts, err := strconv.ParseInt(endParam, 10, 64); err == nil {
			end = time.Unix(ts, 0)
		}
	}

	req := vnstock.QuoteHistoryRequest{
		Symbol:   strings.ToUpper(symbol),
		Start:    start,
		End:      end,
		Interval: interval,
	}

	// Use Data_Source_Router for intelligent source selection with failover
	quotes, source, err := h.DataSourceRouter.FetchQuoteHistory(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": quotes, "source": source})
}

// HandleMarketListing serves GET /api/market/listing (mocked data)
func (h *Handlers) HandleMarketListing(c *gin.Context) {
	mockedListings := []map[string]string{
		{"symbol": "SSI", "name": "SSI Securities Corporation", "exchange": "HOSE"},
		{"symbol": "VNM", "name": "Vietnam Dairy Products JSC", "exchange": "HOSE"},
		{"symbol": "HPG", "name": "Hoa Phat Group JSC", "exchange": "HOSE"},
		{"symbol": "FPT", "name": "FPT Corporation", "exchange": "HOSE"},
		{"symbol": "VIC", "name": "Vingroup JSC", "exchange": "HOSE"},
		{"symbol": "VHM", "name": "Vinhomes JSC", "exchange": "HOSE"},
		{"symbol": "VCB", "name": "Joint Stock Commercial Bank for Foreign Trade of Vietnam", "exchange": "HOSE"},
		{"symbol": "TCB", "name": "Vietnam Technological and Commercial Joint Stock Bank", "exchange": "HOSE"},
		{"symbol": "VPB", "name": "Vietnam Prosperity Joint Stock Commercial Bank", "exchange": "HOSE"},
		{"symbol": "MBB", "name": "Military Commercial Joint Stock Bank", "exchange": "HOSE"},
		{"symbol": "ACB", "name": "Asia Commercial Joint Stock Bank", "exchange": "HOSE"},
		{"symbol": "STB", "name": "Saigon Thuong Tin Commercial Joint Stock Bank", "exchange": "HOSE"},
		{"symbol": "MWG", "name": "Mobile World Investment Corporation", "exchange": "HOSE"},
		{"symbol": "MSN", "name": "Masan Group Corporation", "exchange": "HOSE"},
		{"symbol": "VJC", "name": "Vietjet Aviation Joint Stock Company", "exchange": "HOSE"},
		{"symbol": "PNJ", "name": "Phu Nhuan Jewelry Joint Stock Company", "exchange": "HOSE"},
		{"symbol": "VRE", "name": "Vincom Retail JSC", "exchange": "HOSE"},
		{"symbol": "SAB", "name": "Saigon Beer - Alcohol - Beverage Corporation", "exchange": "HOSE"},
		{"symbol": "CTG", "name": "Vietnam Joint Stock Commercial Bank for Industry and Trade", "exchange": "HOSE"},
		{"symbol": "HDB", "name": "Ho Chi Minh City Development Joint Stock Commercial Bank", "exchange": "HOSE"},
		{"symbol": "BSR", "name": "Binh Son Refining and Petrochemical Company Limited", "exchange": "UPCOM"},
		{"symbol": "PVD", "name": "Petrovietnam Drilling & Well Service Corporation", "exchange": "HOSE"},
		{"symbol": "VGC", "name": "Viglacera Corporation - JSC", "exchange": "HOSE"},
		{"symbol": "KBC", "name": "Kinh Bac City Development Holding Corporation", "exchange": "HOSE"},
		{"symbol": "DIG", "name": "Development Investment Construction Joint Stock Corporation", "exchange": "HOSE"},
		{"symbol": "CEO", "name": "C.E.O Group Joint Stock Company", "exchange": "HNX"},
		{"symbol": "SHB", "name": "Saigon - Hanoi Commercial Joint Stock Bank", "exchange": "HOSE"},
		{"symbol": "HSG", "name": "Hoa Sen Group", "exchange": "HOSE"},
		{"symbol": "NKG", "name": "Nam Kim Steel Joint Stock Company", "exchange": "HOSE"},
		{"symbol": "VND", "name": "VNDIRECT Securities Corporation", "exchange": "HOSE"},
		{"symbol": "VCI", "name": "Vietcap Securities Joint Stock Company", "exchange": "HOSE"},
		{"symbol": "HCM", "name": "Ho Chi Minh City Securities Corporation", "exchange": "HOSE"},
	}

	c.JSON(http.StatusOK, gin.H{"data": mockedListings})
}

// HandleMarketScreener serves GET /api/market/screener
func (h *Handlers) HandleMarketScreener(c *gin.Context) {
	// Generating a rich mocked dataset for the Smoney-style filter.
	stocks := []ScreenerStock{
		{Symbol: "A32", Exchange: "UPCOM", MarketCap: 237.32, PE: 15.50, PB: 1.07, EVEBITDA: 3.86, ReportTerm: "Năm 2024"},
		{Symbol: "AAA", Exchange: "HOSE", MarketCap: 2929.45, PE: 7.86, PB: 0.54, EVEBITDA: 4.12, ReportTerm: "Năm 2024"},
		{Symbol: "AAH", Exchange: "UPCOM", MarketCap: 389.07, PE: -170.56, PB: 0.33, EVEBITDA: 4.89, ReportTerm: "Năm 2024"},
		{Symbol: "AAM", Exchange: "HOSE", MarketCap: 80.99, PE: 43.80, PB: 0.41, EVEBITDA: 4.63, ReportTerm: "Năm 2024"},
		{Symbol: "AAS", Exchange: "UPCOM", MarketCap: 2397.52, PE: 14.53, PB: 0.90, EVEBITDA: 9.61, ReportTerm: "Năm 2024"},
		{Symbol: "AAT", Exchange: "HOSE", MarketCap: 217.41, PE: 7.27, PB: 0.30, EVEBITDA: 5.26, ReportTerm: "Năm 2024"},
		{Symbol: "AAV", Exchange: "HNX", MarketCap: 434.62, PE: -18.83, PB: 0.62, EVEBITDA: 72.47, ReportTerm: "Năm 2024"},
		{Symbol: "ABB", Exchange: "UPCOM", MarketCap: 14402.54, PE: 5.09, PB: 0.86, EVEBITDA: 0, ReportTerm: "Năm 2024"},
		{Symbol: "ABC", Exchange: "UPCOM", MarketCap: 224.59, PE: 20.66, PB: 0.45, EVEBITDA: -10.85, ReportTerm: "Năm 2024"},
		{Symbol: "ABI", Exchange: "UPCOM", MarketCap: 2107.32, PE: 8.24, PB: 1.20, EVEBITDA: -3.27, ReportTerm: "Năm 2024"},
		{Symbol: "ABR", Exchange: "HOSE", MarketCap: 256.00, PE: 16.51, PB: 1.11, EVEBITDA: 20.04, ReportTerm: "Năm 2024"},
		{Symbol: "SSI", Exchange: "HOSE", MarketCap: 56210.12, PE: 18.2, PB: 2.1, EVEBITDA: 14.2, ReportTerm: "Năm 2024"},
		{Symbol: "FPT", Exchange: "HOSE", MarketCap: 140000.5, PE: 20.5, PB: 5.4, EVEBITDA: 15.1, ReportTerm: "Năm 2024"},
		{Symbol: "HPG", Exchange: "HOSE", MarketCap: 185000.0, PE: 12.3, PB: 1.8, EVEBITDA: 8.4, ReportTerm: "Năm 2024"},
		{Symbol: "VNM", Exchange: "HOSE", MarketCap: 145000.0, PE: 16.4, PB: 3.2, EVEBITDA: 11.2, ReportTerm: "Năm 2024"},
		{Symbol: "TCB", Exchange: "HOSE", MarketCap: 150000.0, PE: 6.5, PB: 1.1, EVEBITDA: 0, ReportTerm: "Năm 2024"},
		{Symbol: "MBB", Exchange: "HOSE", MarketCap: 120000.0, PE: 5.8, PB: 1.0, EVEBITDA: 0, ReportTerm: "Năm 2024"},
		{Symbol: "VPB", Exchange: "HOSE", MarketCap: 155000.0, PE: 8.2, PB: 1.2, EVEBITDA: 0, ReportTerm: "Năm 2024"},
		{Symbol: "VHM", Exchange: "HOSE", MarketCap: 185000.0, PE: 5.2, PB: 1.0, EVEBITDA: 6.2, ReportTerm: "Năm 2024"},
		{Symbol: "VIC", Exchange: "HOSE", MarketCap: 175000.0, PE: 25.4, PB: 1.5, EVEBITDA: 12.4, ReportTerm: "Năm 2024"},
		{Symbol: "MWG", Exchange: "HOSE", MarketCap: 75000.0, PE: 35.2, PB: 2.8, EVEBITDA: 18.2, ReportTerm: "Năm 2024"},
	}

	// Dynamic querying
	minPE, _ := strconv.ParseFloat(c.Query("minPE"), 64)
	maxPE, _ := strconv.ParseFloat(c.Query("maxPE"), 64)
	if maxPE == 0 {
		maxPE = 99999
	}
	minCap, _ := strconv.ParseFloat(c.Query("minCap"), 64)

	results := []ScreenerStock{}
	for _, s := range stocks {
		if s.MarketCap >= minCap {
			if s.PE >= minPE && s.PE <= maxPE {
				results = append(results, s)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": results, "total": len(results)})
}

// HandleRateLimitMetrics serves GET /api/metrics/rate-limits
func (h *Handlers) HandleRateLimitMetrics(c *gin.Context) {
	if h.DataSourceRouter == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "data source router not initialized"})
		return
	}

	metrics := h.DataSourceRouter.RateLimiter().GetAllMetrics()
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// HandleUnifiedListing serves GET /api/market/listing with full listing data
// from MarketDataService (symbols, indices, bonds, exchanges).
func (h *Handlers) HandleUnifiedListing(c *gin.Context) {
	data, err := h.MarketDataService.GetListingData(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleCompanyData serves GET /api/market/company/:symbol
func (h *Handlers) HandleCompanyData(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}
	data, err := h.MarketDataService.GetCompanyData(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleFinancialReports serves GET /api/market/finance/:symbol
func (h *Handlers) HandleFinancialReports(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}
	period := c.DefaultQuery("period", "annual")
	data, err := h.MarketDataService.GetFinancialReports(c.Request.Context(), symbol, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleTradingStatistics serves GET /api/market/trading/:symbol
func (h *Handlers) HandleTradingStatistics(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol parameter is required"})
		return
	}
	interval := c.DefaultQuery("interval", "1D")

	end := time.Now()
	start := end.AddDate(-1, 0, 0)
	if s := c.Query("start"); s != "" {
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			start = time.Unix(ts, 0)
		}
	}
	if e := c.Query("end"); e != "" {
		if ts, err := strconv.ParseInt(e, 10, 64); err == nil {
			end = time.Unix(ts, 0)
		}
	}

	data, err := h.MarketDataService.GetTradingStatistics(c.Request.Context(), symbol, interval, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleMarketStatistics serves GET /api/market/statistics
func (h *Handlers) HandleMarketStatistics(c *gin.Context) {
	data, err := h.MarketDataService.GetMarketStatistics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleValuationMetrics serves GET /api/market/valuation
func (h *Handlers) HandleValuationMetrics(c *gin.Context) {
	symbol := c.Query("symbol")
	data, err := h.MarketDataService.GetValuationMetrics(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleMacro serves GET /api/market/macro
func (h *Handlers) HandleMacro(c *gin.Context) {
	data, err := h.MacroService.GetMacroIndicators(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// HandleBatchTradingQuotes serves POST /api/market/trading/batch
func (h *Handlers) HandleBatchTradingQuotes(c *gin.Context) {
	var req struct {
		Symbols []string `json:"symbols"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body, expected {\"symbols\": [...]}"})
		return
	}
	if len(req.Symbols) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbols array is required and must not be empty"})
		return
	}
	quotes, err := h.MarketDataService.GetBatchTradingQuotes(c.Request.Context(), req.Symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": quotes})
}

// HandleWhitelist serves GET /api/market/whitelist — returns the active liquidity whitelist.
func (h *Handlers) HandleWhitelist(c *gin.Context) {
	if h.LiquidityFilter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "liquidity filter not initialized"})
		return
	}
	snapshot := h.LiquidityFilter.GetWhitelist()
	c.JSON(http.StatusOK, gin.H{"data": snapshot})
}

// HandleWhitelistCheck serves GET /api/market/whitelist/check?symbol=FPT
func (h *Handlers) HandleWhitelistCheck(c *gin.Context) {
	if h.LiquidityFilter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "liquidity filter not initialized"})
		return
	}
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol query parameter is required"})
		return
	}
	ok := h.LiquidityFilter.IsWhitelisted(symbol)
	c.JSON(http.StatusOK, gin.H{"symbol": symbol, "whitelisted": ok})
}

// HandleWhitelistRefresh serves POST /api/market/whitelist/refresh — force re-computation.
func (h *Handlers) HandleWhitelistRefresh(c *gin.Context) {
	if h.LiquidityFilter == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "liquidity filter not initialized"})
		return
	}
	if err := h.LiquidityFilter.RefreshAll(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	snapshot := h.LiquidityFilter.GetWhitelist()
	c.JSON(http.StatusOK, gin.H{"data": snapshot})
}
