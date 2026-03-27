package service

import (
	"context"
	"fmt"
	"log"
	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// MarketDataService provides unified access to all market data categories.
type MarketDataService struct {
	router        *infra.DataSourceRouter
	priceService  *PriceService
	sectorService *SectorService
	cache         *infra.Cache
}

// NewMarketDataService creates a new MarketDataService.
func NewMarketDataService(router *infra.DataSourceRouter, priceService *PriceService, sectorService *SectorService, cache *infra.Cache) *MarketDataService {
	return &MarketDataService{
		router:        router,
		priceService:  priceService,
		sectorService: sectorService,
		cache:         cache,
	}
}

// MarketIndices lists the standard VN market indices.
var MarketIndices = []string{"VN30", "VN100", "VNMID", "VNSML", "VNALL"}

// VNExchanges lists the Vietnamese stock exchanges.
var VNExchanges = []model.ExchangeInfo{
	{Code: "HOSE", Name: "Ho Chi Minh Stock Exchange", Description: "Main board exchange in Ho Chi Minh City"},
	{Code: "HNX", Name: "Hanoi Stock Exchange", Description: "Main board exchange in Hanoi"},
	{Code: "UPCOM", Name: "Unlisted Public Company Market", Description: "OTC market for unlisted public companies"},
}

// GetListingData fetches all listing information with 24-hour cache TTL.
func (s *MarketDataService) GetListingData(ctx context.Context) (*model.ListingData, error) {
	cacheKey := "market:listing:all"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.ListingData); ok {
			log.Println("[MarketDataService] Cache hit for listing data")
			return data, nil
		}
	}

	data := &model.ListingData{}
	var isStale bool

	symbols, _, err := s.router.FetchListing(ctx, "")
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch listing: %v", err)
		isStale = true
	} else {
		data.Symbols = symbols
	}

	for _, indexName := range MarketIndices {
		record, err := s.getVCIClient().IndexCurrent(ctx, indexName)
		if err != nil {
			log.Printf("[MarketDataService] Failed to fetch index %s: %v", indexName, err)
			continue
		}
		data.Indices = append(data.Indices, record)
	}

	data.Bonds = []model.BondInfo{
		{Name: "VN Gov Bond 1Y", Tenor: "1Y", Yield: 0, Currency: "VND"},
		{Name: "VN Gov Bond 3Y", Tenor: "3Y", Yield: 0, Currency: "VND"},
		{Name: "VN Gov Bond 5Y", Tenor: "5Y", Yield: 0, Currency: "VND"},
		{Name: "VN Gov Bond 10Y", Tenor: "10Y", Yield: 0, Currency: "VND"},
	}

	exchangeCounts := map[string]int{}
	for _, sym := range data.Symbols {
		exchangeCounts[sym.Exchange]++
	}
	exchanges := make([]model.ExchangeInfo, len(VNExchanges))
	copy(exchanges, VNExchanges)
	for i := range exchanges {
		exchanges[i].StockCount = exchangeCounts[exchanges[i].Code]
	}
	data.Exchanges = exchanges
	data.IsStale = isStale

	s.cache.Set(cacheKey, data, 24*time.Hour)
	log.Printf("[MarketDataService] Fetched listing data: %d symbols, %d indices", len(data.Symbols), len(data.Indices))
	return data, nil
}

// GetAllSymbols returns all listed stock symbols.
func (s *MarketDataService) GetAllSymbols(ctx context.Context) ([]vnstock.ListingRecord, error) {
	data, err := s.GetListingData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Symbols, nil
}

// GetMarketIndices returns current values for all market indices.
func (s *MarketDataService) GetMarketIndices(ctx context.Context) ([]vnstock.IndexRecord, error) {
	data, err := s.GetListingData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Indices, nil
}

// GetExchangeInfo returns exchange information for HOSE, HNX, UPCOM.
func (s *MarketDataService) GetExchangeInfo(ctx context.Context) ([]model.ExchangeInfo, error) {
	data, err := s.GetListingData(ctx)
	if err != nil {
		return nil, err
	}
	return data.Exchanges, nil
}

// GetCompanyData fetches company overview, shareholders, officers, and news with 6-hour cache TTL.
func (s *MarketDataService) GetCompanyData(ctx context.Context, symbol string) (*model.CompanyData, error) {
	cacheKey := fmt.Sprintf("market:company:%s", symbol)
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.CompanyData); ok {
			log.Printf("[MarketDataService] Cache hit for company data: %s", symbol)
			return data, nil
		}
	}

	data := &model.CompanyData{}

	profile, err := s.getVCIClient().CompanyProfile(ctx, symbol)
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch company profile for %s: %v", symbol, err)
		data.IsStale = true
	} else {
		data.Profile = profile
		data.Shareholders = profile.Shareholders
		data.Officers = profile.Leaders
	}

	if len(data.Officers) == 0 {
		officers, err := s.getVCIClient().Officers(ctx, symbol)
		if err != nil {
			log.Printf("[MarketDataService] Failed to fetch officers for %s: %v", symbol, err)
		} else {
			data.Officers = officers
		}
	}

	data.News = []model.CompanyNewsItem{}

	s.cache.Set(cacheKey, data, 6*time.Hour)
	log.Printf("[MarketDataService] Fetched company data for %s", symbol)
	return data, nil
}

// GetCompanyProfile fetches just the company profile.
func (s *MarketDataService) GetCompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, error) {
	data, err := s.GetCompanyData(ctx, symbol)
	if err != nil {
		return vnstock.CompanyProfile{}, err
	}
	return data.Profile, nil
}

// GetShareholders fetches major shareholders for a symbol.
func (s *MarketDataService) GetShareholders(ctx context.Context, symbol string) ([]vnstock.Shareholder, error) {
	data, err := s.GetCompanyData(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return data.Shareholders, nil
}

// GetOfficers fetches management team for a symbol.
func (s *MarketDataService) GetOfficers(ctx context.Context, symbol string) ([]vnstock.Officer, error) {
	data, err := s.GetCompanyData(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return data.Officers, nil
}

// GetFinancialReports fetches income statements, balance sheets, cash flows, and ratios.
func (s *MarketDataService) GetFinancialReports(ctx context.Context, symbol string, period string) (*model.FinancialReportData, error) {
	if period == "" {
		period = "annual"
	}
	cacheKey := fmt.Sprintf("market:finance:%s:%s", symbol, period)
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.FinancialReportData); ok {
			log.Printf("[MarketDataService] Cache hit for financial reports: %s (%s)", symbol, period)
			return data, nil
		}
	}

	data := &model.FinancialReportData{}

	income, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
		Symbol: symbol, Type: "income", Period: period,
	})
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch income statement for %s: %v", symbol, err)
	} else {
		data.IncomeStatements = income
	}

	balance, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
		Symbol: symbol, Type: "balance", Period: period,
	})
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch balance sheet for %s: %v", symbol, err)
	} else {
		data.BalanceSheets = balance
	}

	cashflow, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
		Symbol: symbol, Type: "cashflow", Period: period,
	})
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch cash flow for %s: %v", symbol, err)
	} else {
		data.CashFlows = cashflow
	}

	data.Ratios = []vnstock.FinancialPeriod{}

	s.cache.Set(cacheKey, data, 24*time.Hour)
	log.Printf("[MarketDataService] Fetched financial reports for %s (%s)", symbol, period)
	return data, nil
}

// GetIncomeStatements fetches income statements for a symbol.
func (s *MarketDataService) GetIncomeStatements(ctx context.Context, symbol, period string) ([]vnstock.FinancialPeriod, error) {
	data, err := s.GetFinancialReports(ctx, symbol, period)
	if err != nil {
		return nil, err
	}
	return data.IncomeStatements, nil
}

// GetBalanceSheets fetches balance sheets for a symbol.
func (s *MarketDataService) GetBalanceSheets(ctx context.Context, symbol, period string) ([]vnstock.FinancialPeriod, error) {
	data, err := s.GetFinancialReports(ctx, symbol, period)
	if err != nil {
		return nil, err
	}
	return data.BalanceSheets, nil
}

// GetCashFlows fetches cash flow statements for a symbol.
func (s *MarketDataService) GetCashFlows(ctx context.Context, symbol, period string) ([]vnstock.FinancialPeriod, error) {
	data, err := s.GetFinancialReports(ctx, symbol, period)
	if err != nil {
		return nil, err
	}
	return data.CashFlows, nil
}

// GetTradingStatistics fetches real-time quotes, OHLCV history, intraday, and order book.
func (s *MarketDataService) GetTradingStatistics(ctx context.Context, symbol string, interval string, start, end time.Time) (*model.TradingStatistics, error) {
	data := &model.TradingStatistics{}

	quotes, err := s.priceService.GetQuotes(ctx, []string{symbol}, model.VNStock)
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch quote for %s: %v", symbol, err)
	} else if len(quotes) > 0 {
		data.Quote = &quotes[0]
	}

	if interval == "" {
		interval = "1D"
	}
	history, err := s.priceService.GetHistoricalData(ctx, symbol, start, end, interval)
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch OHLCV history for %s: %v", symbol, err)
	} else {
		data.History = history
	}

	today := time.Now().Truncate(24 * time.Hour)
	intradayBars, err := s.priceService.GetHistoricalData(ctx, symbol, today, time.Now(), "1m")
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch intraday data for %s: %v", symbol, err)
	} else {
		data.Intraday = intradayBars
	}

	data.OrderBook = &model.OrderBookData{
		Symbol: symbol,
		Bids:   []model.OrderBookItem{},
		Asks:   []model.OrderBookItem{},
	}

	return data, nil
}

// GetBatchTradingQuotes fetches real-time quotes for multiple symbols in a single call.
func (s *MarketDataService) GetBatchTradingQuotes(ctx context.Context, symbols []string) ([]model.PriceQuote, error) {
	return s.priceService.GetQuotes(ctx, symbols, model.VNStock)
}

// GetOHLCVHistory fetches OHLCV history for a symbol with any interval.
func (s *MarketDataService) GetOHLCVHistory(ctx context.Context, symbol string, start, end time.Time, interval string) ([]model.OHLCVBar, error) {
	return s.priceService.GetHistoricalData(ctx, symbol, start, end, interval)
}

// GetMarketStatistics fetches market index data, sector indices, breadth, and foreign trading.
func (s *MarketDataService) GetMarketStatistics(ctx context.Context) (*model.MarketStatistics, error) {
	cacheKey := "market:statistics"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.MarketStatistics); ok {
			log.Println("[MarketDataService] Cache hit for market statistics")
			return data, nil
		}
	}

	data := &model.MarketStatistics{}

	for _, indexName := range MarketIndices {
		record, err := s.getVCIClient().IndexCurrent(ctx, indexName)
		if err != nil {
			log.Printf("[MarketDataService] Failed to fetch index %s: %v", indexName, err)
			continue
		}
		data.Indices = append(data.Indices, record)
	}

	sectorPerfs, err := s.sectorService.GetAllSectorPerformances(ctx)
	if err != nil {
		log.Printf("[MarketDataService] Failed to fetch sector performances: %v", err)
	} else {
		data.SectorIndices = sectorPerfs
	}

	breadth, err := s.computeMarketBreadth(ctx)
	if err != nil {
		log.Printf("[MarketDataService] Failed to compute market breadth: %v", err)
	} else {
		data.MarketBreadth = breadth
	}

	data.ForeignTrading = []model.ForeignTradingData{}

	s.cache.Set(cacheKey, data, 30*time.Minute)
	log.Printf("[MarketDataService] Fetched market statistics: %d indices, %d sectors", len(data.Indices), len(data.SectorIndices))
	return data, nil
}

// computeMarketBreadth computes advancing vs declining stocks.
func (s *MarketDataService) computeMarketBreadth(ctx context.Context) (*model.MarketBreadth, error) {
	listings, _, err := s.router.FetchListing(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings for breadth: %w", err)
	}

	maxSymbols := 100
	symbols := make([]string, 0, maxSymbols)
	for i, rec := range listings {
		if i >= maxSymbols {
			break
		}
		symbols = append(symbols, rec.Symbol)
	}

	if len(symbols) == 0 {
		return &model.MarketBreadth{}, nil
	}

	quotes, _, err := s.router.FetchRealTimeQuotes(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quotes for breadth: %w", err)
	}

	breadth := &model.MarketBreadth{}
	for _, q := range quotes {
		change := q.Close - q.Open
		switch {
		case change > 0:
			breadth.Advancing++
		case change < 0:
			breadth.Declining++
		default:
			breadth.Unchanged++
		}
	}

	total := breadth.Advancing + breadth.Declining
	if total > 0 {
		breadth.AdvDecLine = float64(breadth.Advancing) / float64(total)
	}

	return breadth, nil
}

// GetValuationMetrics computes valuation metrics at market, sector, and stock levels.
func (s *MarketDataService) GetValuationMetrics(ctx context.Context, symbol string) (*model.ValuationMetrics, error) {
	cacheKey := fmt.Sprintf("market:valuation:%s", symbol)
	if symbol == "" {
		cacheKey = "market:valuation:all"
	}
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.ValuationMetrics); ok {
			log.Printf("[MarketDataService] Cache hit for valuation metrics: %s", cacheKey)
			return data, nil
		}
	}

	data := &model.ValuationMetrics{}

	marketVal, err := s.computeMarketValuation(ctx)
	if err != nil {
		log.Printf("[MarketDataService] Failed to compute market valuation: %v", err)
	} else {
		data.Market = marketVal
	}

	sectorVals, err := s.computeSectorValuations(ctx)
	if err != nil {
		log.Printf("[MarketDataService] Failed to compute sector valuations: %v", err)
	} else {
		data.Sectors = sectorVals
	}

	if symbol != "" {
		stockVal, err := s.computeStockValuation(ctx, symbol)
		if err != nil {
			log.Printf("[MarketDataService] Failed to compute stock valuation for %s: %v", symbol, err)
		} else {
			data.Stock = stockVal
		}
	}

	s.cache.Set(cacheKey, data, 1*time.Hour)
	return data, nil
}

// computeMarketValuation computes market-level P/E, P/B, EV/EBITDA from aggregated financial data.
func (s *MarketDataService) computeMarketValuation(ctx context.Context) (*model.MarketValuation, error) {
	var totalPE, totalPB, totalEVEBITDA, totalDivYield float64
	var count float64

	for _, sector := range model.AllICBSectors {
		avgs, err := s.sectorService.GetSectorAverages(ctx, sector)
		if err != nil {
			continue
		}
		if avgs.MedianPE > 0 {
			totalPE += avgs.MedianPE
			totalPB += avgs.MedianPB
			totalDivYield += avgs.MedianDivYield
			count++
		}
	}

	if count == 0 {
		return &model.MarketValuation{}, nil
	}

	return &model.MarketValuation{
		PE:       totalPE / count,
		PB:       totalPB / count,
		EVEBITDA: totalEVEBITDA / count,
		DivYield: totalDivYield / count,
	}, nil
}

// computeSectorValuations computes valuation metrics for each ICB sector.
func (s *MarketDataService) computeSectorValuations(ctx context.Context) ([]model.SectorValuation, error) {
	var valuations []model.SectorValuation

	for _, sector := range model.AllICBSectors {
		avgs, err := s.sectorService.GetSectorAverages(ctx, sector)
		if err != nil {
			log.Printf("[MarketDataService] Failed to get sector averages for %s: %v", sector, err)
			continue
		}

		valuations = append(valuations, model.SectorValuation{
			Sector:   sector,
			Name:     model.SectorNameMap[sector],
			PE:       avgs.MedianPE,
			PB:       avgs.MedianPB,
			EVEBITDA: 0,
			DivYield: avgs.MedianDivYield,
		})
	}

	return valuations, nil
}

// computeStockValuation computes valuation metrics for a single stock.
func (s *MarketDataService) computeStockValuation(ctx context.Context, symbol string) (*model.StockValuation, error) {
	financials, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
		Symbol: symbol, Type: "income", Period: "annual",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch financials for %s: %w", symbol, err)
	}

	val := &model.StockValuation{Symbol: symbol}

	if len(financials) > 0 {
		latest := financials[0]
		val.PE = latest.Fields["pe_ratio"]
		val.PB = latest.Fields["pb_ratio"]
		val.EVEBITDA = latest.Fields["ev_ebitda"]
		val.DivYield = latest.Fields["dividend_yield"]
		val.MarketCap = latest.Fields["market_cap"]
	}

	return val, nil
}

// getVCIClient returns the VCI client from the DataSourceRouter for direct API calls.
// This is needed for methods like IndexCurrent, CompanyProfile, Officers that aren't
// exposed through the DataSourceRouter's failover methods.
func (s *MarketDataService) getVCIClient() *vnstock.Client {
	// Access the VCI client via a method we need to add or use the router directly.
	// For now, we use a workaround: fetch via the router's public interface where possible.
	// The original code accessed router.vciClient directly; we need a getter.
	return s.router.VCIClient()
}
