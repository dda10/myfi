package screener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"myfi-backend/internal/domain/market"
	"myfi-backend/internal/infra"

	vnstock "github.com/dda10/vnstock-go"
)

// maxPresetsPerUser is the configurable limit on saved presets per user.
const maxPresetsPerUser = 10

// ScreenerService provides advanced stock filtering with real data from the
// DataSourceRouter, supporting fundamental, sector, and trend-based criteria.
type ScreenerService struct {
	router          *infra.DataSourceRouter
	sectorService   *market.SectorService
	cache           *infra.Cache
	db              *sql.DB
	liquidityFilter *LiquidityFilter
}

// NewScreenerService creates a new ScreenerService.
func NewScreenerService(router *infra.DataSourceRouter, sectorService *market.SectorService, cache *infra.Cache, database *sql.DB) *ScreenerService {
	return &ScreenerService{
		router:        router,
		sectorService: sectorService,
		cache:         cache,
		db:            database,
	}
}

// SetLiquidityFilter attaches the liquidity filter for pre-screening.
// When set, the screener only evaluates whitelisted (liquid) stocks.
func (s *ScreenerService) SetLiquidityFilter(lf *LiquidityFilter) {
	s.liquidityFilter = lf
}

// Screen fetches real stock data, applies filters, sorts, and paginates.
// Delegates initial fundamental filtering to vnstock-go's built-in screener
// (vnstock-go v2) and layers EziStock-specific filters on top.
func (s *ScreenerService) Screen(ctx context.Context, filters ScreenerFilters) ([]ScreenerResult, int, error) {
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}

	// --- Phase 1: Delegate to vnstock-go screener for fundamental pre-filtering ---
	criteria := vnstock.ScreenerCriteria{}
	if filters.MinPE != nil {
		criteria.PEMin = filters.MinPE
	}
	if filters.MaxPE != nil {
		criteria.PEMax = filters.MaxPE
	}
	if filters.MinPB != nil {
		criteria.PBMin = filters.MinPB
	}
	if filters.MaxPB != nil {
		criteria.PBMax = filters.MaxPB
	}
	if filters.MinROE != nil {
		criteria.ROEMin = filters.MinROE
	}
	if filters.MaxROE != nil {
		criteria.ROEMax = filters.MaxROE
	}
	if filters.MinMarketCap != nil {
		criteria.MarketCapMin = filters.MinMarketCap
	}
	if filters.MinRevenueGrowth != nil {
		criteria.RevenueGrowthMin = filters.MinRevenueGrowth
	}
	if filters.MaxRevenueGrowth != nil {
		criteria.RevenueGrowthMax = filters.MaxRevenueGrowth
	}
	if len(filters.Exchanges) == 1 {
		criteria.Exchange = filters.Exchanges[0]
	}

	vnSymbols, _, err := s.router.FetchScreen(ctx, criteria)
	if err != nil {
		log.Printf("[ScreenerService] vnstock screener failed, falling back to full scan: %v", err)
		return s.screenFallback(ctx, filters)
	}

	// --- Phase 2: Fetch listings to get exchange info for screened symbols ---
	listings, listErr := s.fetchAllListings(ctx)
	symbolExchange := make(map[string]string)
	if listErr == nil {
		for _, rec := range listings {
			symbolExchange[rec.Symbol] = rec.Exchange
		}
	}

	// Build listing records for the screened symbols.
	var candidates []vnstock.ListingRecord
	for _, sym := range vnSymbols {
		ex := symbolExchange[sym]
		if s.liquidityFilter != nil && !s.liquidityFilter.IsWhitelisted(sym) {
			continue
		}
		candidates = append(candidates, vnstock.ListingRecord{Symbol: sym, Exchange: ex})
	}

	// --- Phase 3: Layer EziStock-specific filters on top ---
	sectorSet := make(map[ICBSector]bool, len(filters.Sectors))
	for _, sec := range filters.Sectors {
		sectorSet[sec] = true
	}

	sectorTrendMap := make(map[ICBSector]SectorTrend)
	trendSet := make(map[SectorTrend]bool, len(filters.SectorTrends))
	for _, t := range filters.SectorTrends {
		trendSet[t] = true
	}
	if len(trendSet) > 0 || len(sectorSet) > 0 {
		perfs, err := s.sectorService.GetAllSectorPerformances(ctx)
		if err != nil {
			log.Printf("[ScreenerService] Failed to fetch sector performances: %v", err)
		} else {
			for _, p := range perfs {
				sectorTrendMap[p.Sector] = p.Trend
			}
		}
	}

	results := s.fetchAndFilter(ctx, candidates, filters, sectorSet, trendSet, sectorTrendMap)

	sortScreenerResults(results, filters.SortBy, filters.SortOrder)

	total := len(results)
	start := (filters.Page - 1) * filters.PageSize
	if start >= total {
		return []ScreenerResult{}, total, nil
	}
	end := start + filters.PageSize
	if end > total {
		end = total
	}

	return results[start:end], total, nil
}

// screenFallback is the original full-scan path used when the vnstock screener
// is unavailable. It fetches all listings and filters locally.
func (s *ScreenerService) screenFallback(ctx context.Context, filters ScreenerFilters) ([]ScreenerResult, int, error) {
	exchangeSet := make(map[string]bool, len(filters.Exchanges))
	for _, ex := range filters.Exchanges {
		exchangeSet[ex] = true
	}

	sectorSet := make(map[ICBSector]bool, len(filters.Sectors))
	for _, sec := range filters.Sectors {
		sectorSet[sec] = true
	}

	sectorTrendMap := make(map[ICBSector]SectorTrend)
	trendSet := make(map[SectorTrend]bool, len(filters.SectorTrends))
	for _, t := range filters.SectorTrends {
		trendSet[t] = true
	}
	perfs, err := s.sectorService.GetAllSectorPerformances(ctx)
	if err == nil {
		for _, p := range perfs {
			sectorTrendMap[p.Sector] = p.Trend
		}
	}

	listings, err := s.fetchAllListings(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch listings: %w", err)
	}

	var candidates []vnstock.ListingRecord
	for _, rec := range listings {
		if len(exchangeSet) > 0 && !exchangeSet[rec.Exchange] {
			continue
		}
		if s.liquidityFilter != nil && !s.liquidityFilter.IsWhitelisted(rec.Symbol) {
			continue
		}
		candidates = append(candidates, rec)
	}

	results := s.fetchAndFilter(ctx, candidates, filters, sectorSet, trendSet, sectorTrendMap)
	sortScreenerResults(results, filters.SortBy, filters.SortOrder)

	total := len(results)
	start := (filters.Page - 1) * filters.PageSize
	if start >= total {
		return []ScreenerResult{}, total, nil
	}
	end := start + filters.PageSize
	if end > total {
		end = total
	}

	return results[start:end], total, nil
}

// fetchAllListings retrieves listings from all three exchanges with caching.
func (s *ScreenerService) fetchAllListings(ctx context.Context) ([]vnstock.ListingRecord, error) {
	cacheKey := "screener:listings:all"
	if cached, found := s.cache.Get(cacheKey); found {
		if recs, ok := cached.([]vnstock.ListingRecord); ok {
			return recs, nil
		}
	}

	exchanges := []string{"HOSE", "HNX", "UPCOM"}
	var all []vnstock.ListingRecord
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for _, ex := range exchanges {
		wg.Add(1)
		go func(exchange string) {
			defer wg.Done()
			recs, _, err := s.router.FetchListing(ctx, exchange)
			if err != nil {
				log.Printf("[ScreenerService] Failed to fetch listing for %s: %v", exchange, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			mu.Lock()
			all = append(all, recs...)
			mu.Unlock()
		}(ex)
	}
	wg.Wait()

	if len(all) == 0 && firstErr != nil {
		return nil, firstErr
	}

	s.cache.Set(cacheKey, all, 30*time.Minute)
	return all, nil
}

// fetchAndFilter concurrently fetches financial data for candidates and applies all filters.
func (s *ScreenerService) fetchAndFilter(
	ctx context.Context,
	candidates []vnstock.ListingRecord,
	filters ScreenerFilters,
	sectorSet map[ICBSector]bool,
	trendSet map[SectorTrend]bool,
	sectorTrendMap map[ICBSector]SectorTrend,
) []ScreenerResult {
	type indexedResult struct {
		idx    int
		result *ScreenerResult
	}

	resultsCh := make(chan indexedResult, len(candidates))
	sem := make(chan struct{}, 2) // limit concurrency to avoid VCI 429 rate limits
	var wg sync.WaitGroup

	for i, rec := range candidates {
		wg.Add(1)
		go func(idx int, record vnstock.ListingRecord) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Stagger to avoid VCI WAF rate limiting
			if idx > 0 {
				time.Sleep(300 * time.Millisecond)
			}

			result := s.buildScreenerResult(ctx, record, sectorTrendMap)
			if result == nil {
				return
			}

			if len(sectorSet) > 0 && !sectorSet[result.Sector] {
				return
			}
			if len(trendSet) > 0 && !trendSet[result.SectorTrend] {
				return
			}
			if !matchesFundamentalFilters(result, filters) {
				return
			}

			resultsCh <- indexedResult{idx: idx, result: result}
		}(i, rec)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []ScreenerResult
	for ir := range resultsCh {
		results = append(results, *ir.result)
	}
	return results
}

// buildScreenerResult fetches financial data for a single listing and builds a ScreenerResult.
func (s *ScreenerService) buildScreenerResult(
	ctx context.Context,
	rec vnstock.ListingRecord,
	sectorTrendMap map[ICBSector]SectorTrend,
) *ScreenerResult {
	cacheKey := fmt.Sprintf("screener:stock:%s", rec.Symbol)
	if cached, found := s.cache.Get(cacheKey); found {
		if sr, ok := cached.(*ScreenerResult); ok {
			return sr
		}
	}

	periods, _, err := s.router.FetchFinancialStatement(ctx, vnstock.FinancialRequest{
		Symbol: rec.Symbol,
		Type:   "income",
		Period: "annual",
	})
	if err != nil || len(periods) == 0 {
		log.Printf("[ScreenerService] Missing financial data for %s, excluding from results: %v", rec.Symbol, err)
		return nil
	}

	latest := periods[0]
	fields := latest.Fields

	pe := fields["pe"]
	if pe == 0 {
		pe = fields["pe_ratio"]
	}
	marketCap := fields["market_cap"]
	if pe == 0 && marketCap == 0 {
		log.Printf("[ScreenerService] Insufficient financial data for %s, excluding", rec.Symbol)
		return nil
	}

	pb := fields["pb"]
	if pb == 0 {
		pb = fields["pb_ratio"]
	}
	evEbitda := fields["ev_ebitda"]
	roe := fields["roe"]
	roa := fields["roa"]
	revenueGrowth := fields["revenue_growth"]
	profitGrowth := fields["profit_growth"]
	if profitGrowth == 0 {
		profitGrowth = fields["net_income_growth"]
	}
	divYield := fields["dividend_yield"]
	debtToEquity := fields["debt_to_equity"]

	sector, _ := s.sectorService.GetStockSector(rec.Symbol)
	sectorName := market.SectorNameMap[sector]
	trend := sectorTrendMap[sector]

	result := &ScreenerResult{
		Symbol:        rec.Symbol,
		Exchange:      rec.Exchange,
		Sector:        sector,
		SectorName:    sectorName,
		MarketCap:     marketCap,
		PE:            pe,
		PB:            pb,
		EVEBITDA:      evEbitda,
		ROE:           roe,
		ROA:           roa,
		RevenueGrowth: revenueGrowth,
		ProfitGrowth:  profitGrowth,
		DivYield:      divYield,
		DebtToEquity:  debtToEquity,
		SectorTrend:   trend,
	}

	s.cache.Set(cacheKey, result, 30*time.Minute)
	return result
}

// matchesFundamentalFilters checks whether a result passes all fundamental range filters.
func matchesFundamentalFilters(r *ScreenerResult, f ScreenerFilters) bool {
	if !inRange(r.PE, f.MinPE, f.MaxPE) {
		return false
	}
	if !inRange(r.PB, f.MinPB, f.MaxPB) {
		return false
	}
	if f.MinMarketCap != nil && r.MarketCap < *f.MinMarketCap {
		return false
	}
	if !inRange(r.EVEBITDA, f.MinEVEBITDA, f.MaxEVEBITDA) {
		return false
	}
	if !inRange(r.ROE, f.MinROE, f.MaxROE) {
		return false
	}
	if !inRange(r.ROA, f.MinROA, f.MaxROA) {
		return false
	}
	if !inRange(r.RevenueGrowth, f.MinRevenueGrowth, f.MaxRevenueGrowth) {
		return false
	}
	if !inRange(r.ProfitGrowth, f.MinProfitGrowth, f.MaxProfitGrowth) {
		return false
	}
	if !inRange(r.DivYield, f.MinDivYield, f.MaxDivYield) {
		return false
	}
	if !inRange(r.DebtToEquity, f.MinDebtToEquity, f.MaxDebtToEquity) {
		return false
	}
	return true
}

// inRange checks if val is within [min, max]. nil bounds are treated as unbounded.
func inRange(val float64, min, max *float64) bool {
	if min != nil && val < *min {
		return false
	}
	if max != nil && val > *max {
		return false
	}
	return true
}

// sortScreenerResults sorts results by the given field and order.
func sortScreenerResults(results []ScreenerResult, sortBy, sortOrder string) {
	if sortBy == "" {
		sortBy = "marketCap"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	asc := sortOrder == "asc"

	sort.SliceStable(results, func(i, j int) bool {
		vi := getScreenerSortValue(results[i], sortBy)
		vj := getScreenerSortValue(results[j], sortBy)
		if asc {
			return vi < vj
		}
		return vi > vj
	})
}

// getScreenerSortValue extracts the numeric value for sorting.
func getScreenerSortValue(r ScreenerResult, field string) float64 {
	switch field {
	case "marketCap":
		return r.MarketCap
	case "pe":
		return r.PE
	case "pb":
		return r.PB
	case "evEbitda":
		return r.EVEBITDA
	case "roe":
		return r.ROE
	case "roa":
		return r.ROA
	case "revenueGrowth":
		return r.RevenueGrowth
	case "profitGrowth":
		return r.ProfitGrowth
	case "divYield":
		return r.DivYield
	case "debtToEquity":
		return r.DebtToEquity
	default:
		return r.MarketCap
	}
}

// SavePreset persists a new filter preset for a user.
func (s *ScreenerService) SavePreset(ctx context.Context, preset FilterPreset) (int64, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM filter_presets WHERE user_id = $1", preset.UserID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count presets: %w", err)
	}
	if count >= maxPresetsPerUser {
		return 0, fmt.Errorf("preset limit reached: maximum %d presets per user", maxPresetsPerUser)
	}

	filtersJSON, err := json.Marshal(preset.Filters)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal filters: %w", err)
	}

	var id int64
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO filter_presets (user_id, name, filters) VALUES ($1, $2, $3) RETURNING id",
		preset.UserID, preset.Name, string(filtersJSON),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to save preset: %w", err)
	}

	return id, nil
}

// GetPresets returns all filter presets for a user.
func (s *ScreenerService) GetPresets(ctx context.Context, userID string) ([]FilterPreset, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, user_id, name, filters, created_at FROM filter_presets WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query presets: %w", err)
	}
	defer rows.Close()

	var presets []FilterPreset
	for rows.Next() {
		var p FilterPreset
		var filtersStr string
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &filtersStr, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan preset: %w", err)
		}
		if err := json.Unmarshal([]byte(filtersStr), &p.Filters); err != nil {
			log.Printf("[ScreenerService] Failed to unmarshal filters for preset %d: %v", p.ID, err)
			continue
		}
		presets = append(presets, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating presets: %w", err)
	}

	return presets, nil
}

// UpdatePreset updates an existing filter preset's name and filters.
func (s *ScreenerService) UpdatePreset(ctx context.Context, preset FilterPreset) error {
	filtersJSON, err := json.Marshal(preset.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE filter_presets SET name = $1, filters = $2 WHERE id = $3 AND user_id = $4",
		preset.Name, string(filtersJSON), preset.ID, preset.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update preset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("preset not found or not owned by user")
	}

	return nil
}

// DeletePreset removes a filter preset by ID for a given user.
func (s *ScreenerService) DeletePreset(ctx context.Context, presetID int64, userID string) error {
	result, err := s.db.ExecContext(ctx,
		"DELETE FROM filter_presets WHERE id = $1 AND user_id = $2",
		presetID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete preset: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("preset not found or not owned by user")
	}

	return nil
}

// Ensure math import is used
var _ = math.IsNaN
