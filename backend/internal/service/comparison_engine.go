package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
)

// ComparisonEngine computes valuation, performance, and correlation comparisons
type ComparisonEngine struct {
	router        *infra.DataSourceRouter
	sectorService *SectorService
	cache         *infra.Cache
}

// NewComparisonEngine creates a new ComparisonEngine
func NewComparisonEngine(router *infra.DataSourceRouter, sectorService *SectorService, cache *infra.Cache) *ComparisonEngine {
	return &ComparisonEngine{
		router:        router,
		sectorService: sectorService,
		cache:         cache,
	}
}

// periodToStartDate converts a TimePeriod to a start date relative to now
func periodToStartDate(period model.TimePeriod) (time.Time, error) {
	now := time.Now()
	switch period {
	case model.Period3M:
		return now.AddDate(0, -3, 0), nil
	case model.Period6M:
		return now.AddDate(0, -6, 0), nil
	case model.Period1Y:
		return now.AddDate(-1, 0, 0), nil
	case model.Period3Y:
		return now.AddDate(-3, 0, 0), nil
	case model.Period5Y:
		return now.AddDate(-5, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported time period: %s", period)
	}
}

// validateSymbols checks that the symbol count is within bounds
func validateSymbols(symbols []string) error {
	if len(symbols) < 2 {
		return fmt.Errorf("at least 2 symbols required for comparison, got %d", len(symbols))
	}
	if len(symbols) > model.MaxComparisonStocks {
		return fmt.Errorf("maximum %d stocks allowed for comparison, got %d", model.MaxComparisonStocks, len(symbols))
	}
	return nil
}

// CompareValuation fetches financial data for multiple symbols and returns P/E, P/B time-series
func (ce *ComparisonEngine) CompareValuation(ctx context.Context, symbols []string, period model.TimePeriod) (model.ValuationResult, error) {
	if err := validateSymbols(symbols); err != nil {
		return model.ValuationResult{}, err
	}

	startDate, err := periodToStartDate(period)
	if err != nil {
		return model.ValuationResult{}, err
	}

	cacheKey := fmt.Sprintf("comparison:valuation:%v:%s", symbols, period)
	if cached, ok := ce.cache.Get(cacheKey); ok {
		return cached.(model.ValuationResult), nil
	}

	type symbolResult struct {
		symbol string
		series model.ValuationSeries
		err    error
	}

	results := make(chan symbolResult, len(symbols))
	var wg sync.WaitGroup

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			series, fetchErr := ce.fetchValuationSeries(ctx, symbol, startDate)
			results <- symbolResult{symbol: symbol, series: series, err: fetchErr}
		}(sym)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allSeries []model.ValuationSeries
	var warnings []string

	for r := range results {
		if r.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", r.symbol, r.err))
			log.Printf("[ComparisonEngine] Valuation fetch failed for %s: %v", r.symbol, r.err)
			continue
		}
		if len(r.series.Data) == 0 {
			warnings = append(warnings, fmt.Sprintf("%s: no valuation data available for the selected period", r.symbol))
			continue
		}
		allSeries = append(allSeries, r.series)
	}

	sort.Slice(allSeries, func(i, j int) bool {
		return allSeries[i].Symbol < allSeries[j].Symbol
	})

	result := model.ValuationResult{
		Series:   allSeries,
		Period:   period,
		Warnings: warnings,
	}

	ce.cache.Set(cacheKey, result, 15*time.Minute)
	return result, nil
}

// fetchValuationSeries fetches quarterly financial data and builds P/E, P/B time-series for a symbol
func (ce *ComparisonEngine) fetchValuationSeries(ctx context.Context, symbol string, startDate time.Time) (model.ValuationSeries, error) {
	incomeReq := vnstock.FinancialRequest{
		Symbol: symbol,
		Type:   "income",
		Period: "quarterly",
	}
	incomePeriods, _, err := ce.router.FetchFinancialStatement(ctx, incomeReq)
	if err != nil {
		return model.ValuationSeries{}, fmt.Errorf("failed to fetch income statement: %w", err)
	}

	balanceReq := vnstock.FinancialRequest{
		Symbol: symbol,
		Type:   "balance",
		Period: "quarterly",
	}
	balancePeriods, _, err := ce.router.FetchFinancialStatement(ctx, balanceReq)
	if err != nil {
		return model.ValuationSeries{}, fmt.Errorf("failed to fetch balance sheet: %w", err)
	}

	type periodKey struct {
		year    int
		quarter int
	}
	balanceMap := make(map[periodKey]map[string]float64)
	for _, bp := range balancePeriods {
		balanceMap[periodKey{bp.Year, bp.Quarter}] = bp.Fields
	}

	var points []model.ValuationPoint
	for _, ip := range incomePeriods {
		month := (ip.Quarter-1)*3 + 1
		ts := time.Date(ip.Year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		if ts.Before(startDate) {
			continue
		}

		pe := extractField(ip.Fields, "pe", "PE", "priceToEarning")
		pb := 0.0
		if bf, ok := balanceMap[periodKey{ip.Year, ip.Quarter}]; ok {
			pb = extractField(bf, "pb", "PB", "priceToBook")
		}

		if pe != 0 || pb != 0 {
			points = append(points, model.ValuationPoint{
				Timestamp: ts,
				PE:        pe,
				PB:        pb,
			})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp.Before(points[j].Timestamp)
	})

	return model.ValuationSeries{Symbol: symbol, Data: points}, nil
}

// extractField tries multiple field name variants and returns the first non-zero value
func extractField(fields map[string]float64, names ...string) float64 {
	for _, name := range names {
		if v, ok := fields[name]; ok && v != 0 {
			return v
		}
	}
	return 0
}

// ComparePerformance fetches OHLCV history and normalizes to % returns from start date
func (ce *ComparisonEngine) ComparePerformance(ctx context.Context, symbols []string, period model.TimePeriod) (model.PerformanceResult, error) {
	if err := validateSymbols(symbols); err != nil {
		return model.PerformanceResult{}, err
	}

	startDate, err := periodToStartDate(period)
	if err != nil {
		return model.PerformanceResult{}, err
	}

	cacheKey := fmt.Sprintf("comparison:performance:%v:%s", symbols, period)
	if cached, ok := ce.cache.Get(cacheKey); ok {
		return cached.(model.PerformanceResult), nil
	}

	type symbolResult struct {
		symbol string
		series model.PerformanceSeries
		err    error
	}

	results := make(chan symbolResult, len(symbols))
	var wg sync.WaitGroup

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			series, fetchErr := ce.fetchPerformanceSeries(ctx, symbol, startDate)
			results <- symbolResult{symbol: symbol, series: series, err: fetchErr}
		}(sym)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var allSeries []model.PerformanceSeries
	var warnings []string

	for r := range results {
		if r.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", r.symbol, r.err))
			log.Printf("[ComparisonEngine] Performance fetch failed for %s: %v", r.symbol, r.err)
			continue
		}
		if len(r.series.Data) == 0 {
			warnings = append(warnings, fmt.Sprintf("%s: no OHLCV data available for the selected period", r.symbol))
			continue
		}
		allSeries = append(allSeries, r.series)
	}

	sort.Slice(allSeries, func(i, j int) bool {
		return allSeries[i].Symbol < allSeries[j].Symbol
	})

	result := model.PerformanceResult{
		Series:   allSeries,
		Period:   period,
		Warnings: warnings,
	}

	ce.cache.Set(cacheKey, result, 15*time.Minute)
	return result, nil
}

// fetchPerformanceSeries fetches OHLCV data and computes normalized % returns from start date
func (ce *ComparisonEngine) fetchPerformanceSeries(ctx context.Context, symbol string, startDate time.Time) (model.PerformanceSeries, error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    startDate,
		End:      time.Now(),
		Interval: "1D",
	}

	quotes, _, err := ce.router.FetchQuoteHistory(ctx, req)
	if err != nil {
		return model.PerformanceSeries{}, fmt.Errorf("failed to fetch OHLCV data: %w", err)
	}

	if len(quotes) == 0 {
		return model.PerformanceSeries{}, nil
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Timestamp.Before(quotes[j].Timestamp)
	})

	basePrice := quotes[0].Close
	if basePrice == 0 {
		return model.PerformanceSeries{}, fmt.Errorf("base price is zero, cannot normalize returns")
	}

	points := make([]model.PerformancePoint, len(quotes))
	for i, q := range quotes {
		points[i] = model.PerformancePoint{
			Timestamp:     q.Timestamp,
			ReturnPercent: ((q.Close - basePrice) / basePrice) * 100.0,
		}
	}

	return model.PerformanceSeries{Symbol: symbol, Data: points}, nil
}

// ComputeCorrelation computes a Pearson correlation matrix from daily returns
func (ce *ComparisonEngine) ComputeCorrelation(ctx context.Context, symbols []string, period model.TimePeriod) (model.CorrelationResult, error) {
	if err := validateSymbols(symbols); err != nil {
		return model.CorrelationResult{}, err
	}

	startDate, err := periodToStartDate(period)
	if err != nil {
		return model.CorrelationResult{}, err
	}

	cacheKey := fmt.Sprintf("comparison:correlation:%v:%s", symbols, period)
	if cached, ok := ce.cache.Get(cacheKey); ok {
		return cached.(model.CorrelationResult), nil
	}

	type symbolReturns struct {
		symbol  string
		returns []float64
		dates   []time.Time
		err     error
	}

	results := make(chan symbolReturns, len(symbols))
	var wg sync.WaitGroup

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			returns, dates, fetchErr := ce.fetchDailyReturns(ctx, symbol, startDate)
			results <- symbolReturns{symbol: symbol, returns: returns, dates: dates, err: fetchErr}
		}(sym)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	returnsMap := make(map[string]map[time.Time]float64)
	var validSymbols []string
	var warnings []string

	for r := range results {
		if r.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", r.symbol, r.err))
			log.Printf("[ComparisonEngine] Correlation fetch failed for %s: %v", r.symbol, r.err)
			continue
		}
		if len(r.returns) < 2 {
			warnings = append(warnings, fmt.Sprintf("%s: insufficient data for correlation (need at least 2 daily returns)", r.symbol))
			continue
		}

		dateMap := make(map[time.Time]float64)
		for i, d := range r.dates {
			dateMap[d] = r.returns[i]
		}
		returnsMap[r.symbol] = dateMap
		validSymbols = append(validSymbols, r.symbol)
	}

	sort.Strings(validSymbols)

	if len(validSymbols) < 2 {
		return model.CorrelationResult{
			Symbols:  validSymbols,
			Warnings: warnings,
			Period:   period,
		}, fmt.Errorf("need at least 2 symbols with valid data for correlation, got %d", len(validSymbols))
	}

	commonDates := findCommonDates(returnsMap, validSymbols)

	if len(commonDates) < 2 {
		return model.CorrelationResult{
			Symbols:  validSymbols,
			Warnings: warnings,
			Period:   period,
		}, fmt.Errorf("insufficient overlapping trading days for correlation computation")
	}

	alignedReturns := make(map[string][]float64)
	for _, sym := range validSymbols {
		aligned := make([]float64, len(commonDates))
		for i, d := range commonDates {
			aligned[i] = returnsMap[sym][d]
		}
		alignedReturns[sym] = aligned
	}

	n := len(validSymbols)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
	}

	var pairs []model.CorrelationPair
	for i := 0; i < n; i++ {
		matrix[i][i] = 1.0
		for j := i + 1; j < n; j++ {
			corr := pearsonCorrelation(alignedReturns[validSymbols[i]], alignedReturns[validSymbols[j]])
			matrix[i][j] = corr
			matrix[j][i] = corr
			pairs = append(pairs, model.CorrelationPair{
				SymbolA:     validSymbols[i],
				SymbolB:     validSymbols[j],
				Correlation: corr,
			})
		}
	}

	result := model.CorrelationResult{
		Symbols:  validSymbols,
		Matrix:   matrix,
		Pairs:    pairs,
		Period:   period,
		Warnings: warnings,
	}

	ce.cache.Set(cacheKey, result, 15*time.Minute)
	return result, nil
}

// fetchDailyReturns fetches OHLCV data and computes daily percentage returns
func (ce *ComparisonEngine) fetchDailyReturns(ctx context.Context, symbol string, startDate time.Time) ([]float64, []time.Time, error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    startDate,
		End:      time.Now(),
		Interval: "1D",
	}

	quotes, _, err := ce.router.FetchQuoteHistory(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch OHLCV data: %w", err)
	}

	if len(quotes) < 2 {
		return nil, nil, nil
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Timestamp.Before(quotes[j].Timestamp)
	})

	returns := make([]float64, 0, len(quotes)-1)
	dates := make([]time.Time, 0, len(quotes)-1)

	for i := 1; i < len(quotes); i++ {
		if quotes[i-1].Close == 0 {
			continue
		}
		dailyReturn := (quotes[i].Close - quotes[i-1].Close) / quotes[i-1].Close
		returns = append(returns, dailyReturn)
		d := quotes[i].Timestamp.Truncate(24 * time.Hour)
		dates = append(dates, d)
	}

	return returns, dates, nil
}

// findCommonDates returns sorted dates that exist in all symbols' return maps
func findCommonDates(returnsMap map[string]map[time.Time]float64, symbols []string) []time.Time {
	if len(symbols) == 0 {
		return nil
	}

	candidates := make(map[time.Time]bool)
	for d := range returnsMap[symbols[0]] {
		candidates[d] = true
	}

	for _, sym := range symbols[1:] {
		symDates := returnsMap[sym]
		for d := range candidates {
			if _, ok := symDates[d]; !ok {
				delete(candidates, d)
			}
		}
	}

	dates := make([]time.Time, 0, len(candidates))
	for d := range candidates {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates
}

// pearsonCorrelation computes the Pearson correlation coefficient between two slices
func pearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return 0
	}

	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)

	var cov, varX, varY float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varX += dx * dx
		varY += dy * dy
	}

	denom := math.Sqrt(varX * varY)
	if denom == 0 {
		return 0
	}

	return cov / denom
}

// GetSectorStocks returns stocks grouped by sector for easy comparison setup
func (ce *ComparisonEngine) GetSectorStocks(ctx context.Context, sector model.ICBSector) (model.SectorStocksResult, error) {
	cacheKey := fmt.Sprintf("comparison:sector_stocks:%s", sector)
	if cached, ok := ce.cache.Get(cacheKey); ok {
		return cached.(model.SectorStocksResult), nil
	}

	if err := ce.sectorService.ensureMappingFresh(ctx); err != nil {
		log.Printf("[ComparisonEngine] Failed to refresh sector mapping: %v", err)
	}

	symbols := ce.sectorService.getStocksInSector(sector)
	if len(symbols) == 0 {
		return model.SectorStocksResult{}, fmt.Errorf("no stocks found for sector %s", sector)
	}

	result := model.SectorStocksResult{
		Sector:  sector,
		Symbols: symbols,
	}

	ce.cache.Set(cacheKey, result, 1*time.Hour)
	return result, nil
}
