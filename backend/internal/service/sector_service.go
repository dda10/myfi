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

// sectorKeywords maps listing sector strings to ICB sector codes
var sectorKeywords = map[string]model.ICBSector{
	// Banks / Financial Services
	"Ngân hàng":          model.VNFIN,
	"Tài chính":          model.VNFIN,
	"Dịch vụ tài chính":  model.VNFIN,
	"Bảo hiểm":           model.VNFIN,
	"Chứng khoán":        model.VNFIN,
	"Finance":            model.VNFIN,
	"Banks":              model.VNFIN,
	"Financial Services": model.VNFIN,
	"Insurance":          model.VNFIN,
	// Real Estate
	"Bất động sản": model.VNREAL,
	"Real Estate":  model.VNREAL,
	// Technology
	"Công nghệ thông tin": model.VNIT,
	"Công nghệ":           model.VNIT,
	"Technology":          model.VNIT,
	// Basic Resources / Materials
	"Tài nguyên cơ bản": model.VNMAT,
	"Nguyên vật liệu":   model.VNMAT,
	"Khoáng sản":        model.VNMAT,
	"Thép":              model.VNMAT,
	"Basic Resources":   model.VNMAT,
	"Materials":         model.VNMAT,
	// Food & Beverage / Consumer Staples
	"Thực phẩm & Đồ uống": model.VNCONS,
	"Thực phẩm":           model.VNCONS,
	"Đồ uống":             model.VNCONS,
	"Hàng tiêu dùng":      model.VNCONS,
	"Food & Beverage":     model.VNCONS,
	"Consumer Staples":    model.VNCONS,
	// Consumer Discretionary
	"Hàng cá nhân & Gia dụng": model.VNCOND,
	"Ô tô & phụ tùng":         model.VNCOND,
	"Du lịch & Giải trí":      model.VNCOND,
	"Truyền thông":            model.VNCOND,
	"Bán lẻ":                  model.VNCOND,
	"Consumer Discretionary":  model.VNCOND,
	"Retail":                  model.VNCOND,
	// Chemicals / Industrial
	"Hóa chất":                   model.VNIND,
	"Công nghiệp":                model.VNIND,
	"Xây dựng & Vật liệu":        model.VNIND,
	"Xây dựng":                   model.VNIND,
	"Vận tải":                    model.VNIND,
	"Hàng & Dịch vụ công nghiệp": model.VNIND,
	"Chemicals":                  model.VNIND,
	"Construction & Materials":   model.VNIND,
	"Industrial":                 model.VNIND,
	// Oil & Gas / Energy
	"Dầu khí":    model.VNENE,
	"Năng lượng": model.VNENE,
	"Oil & Gas":  model.VNENE,
	"Energy":     model.VNENE,
	// Healthcare
	"Y tế":              model.VNHEAL,
	"Chăm sóc sức khỏe": model.VNHEAL,
	"Dược phẩm":         model.VNHEAL,
	"Healthcare":        model.VNHEAL,
	// Utilities
	"Tiện ích":  model.VNUTI,
	"Điện":      model.VNUTI,
	"Nước":      model.VNUTI,
	"Utilities": model.VNUTI,
}

// SectorService provides ICB sector classification and performance tracking
type SectorService struct {
	router             *infra.DataSourceRouter
	cache              *infra.Cache
	stockToSector      map[string]model.ICBSector
	mu                 sync.RWMutex
	lastMappingRefresh time.Time
}

// NewSectorService creates a new SectorService
func NewSectorService(router *infra.DataSourceRouter, cache *infra.Cache) *SectorService {
	return &SectorService{
		router:        router,
		cache:         cache,
		stockToSector: make(map[string]model.ICBSector),
	}
}

// GetStockSector returns the ICB sector classification for a stock symbol
func (s *SectorService) GetStockSector(symbol string) (model.ICBSector, error) {
	s.mu.RLock()
	sector, ok := s.stockToSector[symbol]
	s.mu.RUnlock()

	if ok {
		return sector, nil
	}

	// If mapping is empty or stale, try refreshing
	if err := s.ensureMappingFresh(context.Background()); err != nil {
		log.Printf("[SectorService] Failed to refresh mapping: %v", err)
	}

	s.mu.RLock()
	sector, ok = s.stockToSector[symbol]
	s.mu.RUnlock()

	if ok {
		return sector, nil
	}

	return "", fmt.Errorf("sector not found for symbol: %s", symbol)
}

// GetSectorPerformance computes performance metrics for a single ICB sector
func (s *SectorService) GetSectorPerformance(ctx context.Context, sector model.ICBSector) (model.SectorPerformance, error) {
	cacheKey := fmt.Sprintf("sector:perf:%s", sector)

	if cached, found := s.cache.Get(cacheKey); found {
		if perf, ok := cached.(model.SectorPerformance); ok {
			return perf, nil
		}
	}

	perf, err := s.computeSectorPerformance(ctx, sector)
	if err != nil {
		staleCacheKey := fmt.Sprintf("sector:perf:stale:%s", sector)
		if cached, found := s.cache.Get(staleCacheKey); found {
			if stalePerf, ok := cached.(model.SectorPerformance); ok {
				stalePerf.IsStale = true
				log.Printf("[SectorService] Returning stale data for sector %s", sector)
				return stalePerf, nil
			}
		}
		return model.SectorPerformance{}, fmt.Errorf("failed to get sector performance for %s: %w", sector, err)
	}

	ttl := s.getCacheTTL()
	s.cache.Set(cacheKey, perf, ttl)
	s.cache.Set(fmt.Sprintf("sector:perf:stale:%s", sector), perf, 24*time.Hour)

	return perf, nil
}

// GetAllSectorPerformances returns performance metrics for all 10 ICB sectors
func (s *SectorService) GetAllSectorPerformances(ctx context.Context) ([]model.SectorPerformance, error) {
	cacheKey := "sector:perf:all"

	if cached, found := s.cache.Get(cacheKey); found {
		if perfs, ok := cached.([]model.SectorPerformance); ok {
			return perfs, nil
		}
	}

	var (
		results []model.SectorPerformance
		mu      sync.Mutex
		wg      sync.WaitGroup
		errs    []error
	)

	for _, sector := range model.AllICBSectors {
		wg.Add(1)
		go func(sec model.ICBSector) {
			defer wg.Done()
			perf, err := s.GetSectorPerformance(ctx, sec)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", sec, err))
				return
			}
			results = append(results, perf)
		}(sector)
	}
	wg.Wait()

	if len(results) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("failed to fetch any sector data: %v", errs[0])
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Sector < results[j].Sector
	})

	ttl := s.getCacheTTL()
	s.cache.Set(cacheKey, results, ttl)

	if len(errs) > 0 {
		log.Printf("[SectorService] Partial failures fetching sectors: %d/%d succeeded", len(results), len(model.AllICBSectors))
	}

	return results, nil
}

// GetSectorAverages computes median fundamental metrics for stocks in a sector
func (s *SectorService) GetSectorAverages(ctx context.Context, sector model.ICBSector) (model.SectorAverages, error) {
	cacheKey := fmt.Sprintf("sector:avg:%s", sector)

	if cached, found := s.cache.Get(cacheKey); found {
		if avg, ok := cached.(model.SectorAverages); ok {
			return avg, nil
		}
	}

	symbols := s.getStocksInSector(sector)
	if len(symbols) == 0 {
		if err := s.ensureMappingFresh(ctx); err != nil {
			log.Printf("[SectorService] Failed to refresh mapping for averages: %v", err)
		}
		symbols = s.getStocksInSector(sector)
	}

	if len(symbols) == 0 {
		return model.SectorAverages{Sector: sector}, nil
	}

	if len(symbols) > 30 {
		symbols = symbols[:30]
	}

	var (
		peValues, pbValues, roeValues, roaValues []float64
		divYieldValues, debtEquityValues         []float64
		mu                                       sync.Mutex
		wg                                       sync.WaitGroup
	)

	for _, sym := range symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			ratios, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
				Symbol: symbol,
				Type:   "income",
				Period: "annual",
			})
			if err != nil || len(ratios) == 0 {
				return
			}

			latest := ratios[0]
			mu.Lock()
			defer mu.Unlock()

			if v, ok := latest.Fields["pe"]; ok && v > 0 && v < 1000 {
				peValues = append(peValues, v)
			}
			if v, ok := latest.Fields["pb"]; ok && v > 0 && v < 100 {
				pbValues = append(pbValues, v)
			}
			if v, ok := latest.Fields["roe"]; ok && !math.IsNaN(v) {
				roeValues = append(roeValues, v)
			}
			if v, ok := latest.Fields["roa"]; ok && !math.IsNaN(v) {
				roaValues = append(roaValues, v)
			}
			if v, ok := latest.Fields["dividend_yield"]; ok && v >= 0 {
				divYieldValues = append(divYieldValues, v)
			}
			if v, ok := latest.Fields["debt_to_equity"]; ok && v >= 0 {
				debtEquityValues = append(debtEquityValues, v)
			}
		}(sym)
	}
	wg.Wait()

	avg := model.SectorAverages{
		Sector:             sector,
		MedianPE:           median(peValues),
		MedianPB:           median(pbValues),
		MedianROE:          median(roeValues),
		MedianROA:          median(roaValues),
		MedianDivYield:     median(divYieldValues),
		MedianDebtToEquity: median(debtEquityValues),
	}

	s.cache.Set(cacheKey, avg, s.getCacheTTL())
	return avg, nil
}

// computeSectorPerformance fetches index OHLCV data and computes metrics
func (s *SectorService) computeSectorPerformance(ctx context.Context, sector model.ICBSector) (model.SectorPerformance, error) {
	now := time.Now()
	start := now.AddDate(-1, -1, 0)

	req := vnstock.IndexHistoryRequest{
		Name:     string(sector),
		Start:    start,
		End:      now,
		Interval: "1d",
	}

	records, _, err := s.router.FetchIndexHistory(ctx, req)
	if err != nil {
		return model.SectorPerformance{}, fmt.Errorf("failed to fetch index history for %s: %w", sector, err)
	}

	if len(records) == 0 {
		return model.SectorPerformance{}, fmt.Errorf("no index data returned for %s", sector)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})

	latest := records[len(records)-1]
	currentPrice := latest.Close
	if currentPrice == 0 {
		currentPrice = latest.Value
	}

	todayChange := computeChangeFromRecords(records, 1)
	oneWeekChange := computeChangeFromRecords(records, 5)
	oneMonthChange := computeChangeFromRecords(records, 22)
	threeMonthChange := computeChangeFromRecords(records, 66)
	sixMonthChange := computeChangeFromRecords(records, 132)
	oneYearChange := computeChangeFromRecords(records, 252)

	sma20 := computeSMAFromRecords(records, 20)
	sma50 := computeSMAFromRecords(records, 50)

	trend := computeTrend(currentPrice, sma20, sma50)

	perf := model.SectorPerformance{
		Sector:           sector,
		SectorName:       model.SectorNameMap[sector],
		Trend:            trend,
		TodayChange:      todayChange,
		OneWeekChange:    oneWeekChange,
		OneMonthChange:   oneMonthChange,
		ThreeMonthChange: threeMonthChange,
		SixMonthChange:   sixMonthChange,
		OneYearChange:    oneYearChange,
		CurrentPrice:     currentPrice,
		SMA20:            sma20,
		SMA50:            sma50,
		IsStale:          false,
	}

	return perf, nil
}

// computeChangeFromRecords calculates percentage change over N trading days
func computeChangeFromRecords(records []vnstock.IndexRecord, daysBack int) float64 {
	if len(records) < 2 {
		return 0
	}

	latestIdx := len(records) - 1
	refIdx := latestIdx - daysBack
	if refIdx < 0 {
		refIdx = 0
	}

	latestClose := records[latestIdx].Close
	if latestClose == 0 {
		latestClose = records[latestIdx].Value
	}
	refClose := records[refIdx].Close
	if refClose == 0 {
		refClose = records[refIdx].Value
	}

	if refClose == 0 {
		return 0
	}

	return ((latestClose - refClose) / refClose) * 100
}

// computeSMAFromRecords calculates Simple Moving Average from the last N records
func computeSMAFromRecords(records []vnstock.IndexRecord, period int) float64 {
	if len(records) < period {
		if len(records) == 0 {
			return 0
		}
		period = len(records)
	}

	sum := 0.0
	startIdx := len(records) - period
	for i := startIdx; i < len(records); i++ {
		price := records[i].Close
		if price == 0 {
			price = records[i].Value
		}
		sum += price
	}

	return sum / float64(period)
}

// computeTrend determines sector trend based on price vs SMA(20) and SMA(50)
func computeTrend(currentPrice, sma20, sma50 float64) model.SectorTrend {
	if sma20 == 0 || sma50 == 0 {
		return model.Sideways
	}
	if currentPrice > sma20 && currentPrice > sma50 {
		return model.Uptrend
	}
	if currentPrice < sma20 && currentPrice < sma50 {
		return model.Downtrend
	}
	return model.Sideways
}

// refreshStockToSectorMapping fetches listing data and builds the stock-to-sector map
func (s *SectorService) refreshStockToSectorMapping(ctx context.Context) error {
	records, _, err := s.router.FetchListing(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to fetch listing for sector mapping: %w", err)
	}

	newMapping := make(map[string]model.ICBSector, len(records))
	for _, rec := range records {
		if rec.Sector == "" {
			continue
		}
		if sector, ok := sectorKeywords[rec.Sector]; ok {
			newMapping[rec.Symbol] = sector
		} else {
			matched := false
			for keyword, sec := range sectorKeywords {
				if containsIgnoreCase(rec.Sector, keyword) {
					newMapping[rec.Symbol] = sec
					matched = true
					break
				}
			}
			if !matched {
				log.Printf("[SectorService] Unknown sector '%s' for symbol %s", rec.Sector, rec.Symbol)
			}
		}
	}

	s.mu.Lock()
	s.stockToSector = newMapping
	s.lastMappingRefresh = time.Now()
	s.mu.Unlock()

	log.Printf("[SectorService] Refreshed stock-to-sector mapping: %d stocks mapped", len(newMapping))
	return nil
}

// ensureMappingFresh refreshes the mapping if it's stale (older than 24h or empty)
func (s *SectorService) ensureMappingFresh(ctx context.Context) error {
	s.mu.RLock()
	isEmpty := len(s.stockToSector) == 0
	isStale := time.Since(s.lastMappingRefresh) > 24*time.Hour
	s.mu.RUnlock()

	if isEmpty || isStale {
		return s.refreshStockToSectorMapping(ctx)
	}
	return nil
}

// getStocksInSector returns all stock symbols belonging to a sector
func (s *SectorService) getStocksInSector(sector model.ICBSector) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var symbols []string
	for sym, sec := range s.stockToSector {
		if sec == sector {
			symbols = append(symbols, sym)
		}
	}
	sort.Strings(symbols)
	return symbols
}

// getCacheTTL returns the appropriate cache TTL based on trading hours
func (s *SectorService) getCacheTTL() time.Duration {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.FixedZone("ICT", 7*60*60)
	}

	now := time.Now().In(loc)
	hour := now.Hour()
	weekday := now.Weekday()

	if weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour < 15 {
		return 30 * time.Minute
	}
	return 6 * time.Hour
}

// StartDailyMappingRefresh starts a goroutine that refreshes the stock-to-sector mapping daily at 9:00 ICT
func (s *SectorService) StartDailyMappingRefresh(ctx context.Context) {
	go func() {
		for {
			nextRefresh := s.nextRefreshTime()
			timer := time.NewTimer(time.Until(nextRefresh))

			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				if err := s.refreshStockToSectorMapping(ctx); err != nil {
					log.Printf("[SectorService] Daily mapping refresh failed: %v", err)
				}
			}
		}
	}()
}

// nextRefreshTime calculates the next 9:00 ICT
func (s *SectorService) nextRefreshTime() time.Time {
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		loc = time.FixedZone("ICT", 7*60*60)
	}

	now := time.Now().In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, loc)

	if now.After(next) {
		next = next.AddDate(0, 0, 1)
	}

	for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
		next = next.AddDate(0, 0, 1)
	}

	return next
}

// median computes the median of a float64 slice
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(sLower) >= len(substrLower) && contains(sLower, substrLower)
}

// toLower is a simple ASCII+UTF8 lowercase helper
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range len(s) {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
