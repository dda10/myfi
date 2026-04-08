package screener

import (
	"context"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"myfi-backend/internal/infra"

	vnstock "github.com/dda10/vnstock-go"
)

// refreshHourICT is the hour (in ICT = UTC+7) after which daily refresh runs.
// Vietnam market closes at 15:00 ICT; we refresh shortly after (Req 39.11).
const refreshHourICT = 15

// ictZone is the Vietnam timezone (Indochina Time, UTC+7).
var ictZone = time.FixedZone("ICT", 7*3600)

// ADTV scoring benchmarks calibrated for the Vietnamese market (VND/day).
const (
	adtvExcellent = 50_000_000_000 // 50B VND → score 100
	adtvGood      = 5_000_000_000  // 5B VND  → score 70
	adtvModerate  = 500_000_000    // 500M VND → score 40
)

// LiquidityFilter computes tradability scores (0-100) for Vietnamese stocks
// and classifies them into tiers. It maintains an in-memory score cache
// refreshed daily after 15:00 ICT (Vietnam market close).
//
// Scoring factors (Req 39.1):
//   - Average Daily Trading Value (ADTV) over 20 days
//   - Volume consistency (coefficient of variation)
//   - Bid-ask spread proxy (avg high-low range / close)
//   - Zero-volume trading days in last 20 sessions
//   - Free-float ratio
type LiquidityFilter struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
	config LiquidityConfig

	mu        sync.RWMutex
	scores    map[string]LiquidityScore
	updatedAt time.Time
	stopCh    chan struct{}
}

// NewLiquidityFilter creates a new filter with default config.
func NewLiquidityFilter(router *infra.DataSourceRouter, cache *infra.Cache) *LiquidityFilter {
	return &LiquidityFilter{
		router: router,
		cache:  cache,
		config: DefaultLiquidityConfig(),
		scores: make(map[string]LiquidityScore),
		stopCh: make(chan struct{}),
	}
}

// NewLiquidityFilterWithConfig creates a filter with custom thresholds.
func NewLiquidityFilterWithConfig(router *infra.DataSourceRouter, cache *infra.Cache, cfg LiquidityConfig) *LiquidityFilter {
	f := NewLiquidityFilter(router, cache)
	f.config = cfg
	return f
}

// Start runs the initial score computation and schedules daily refreshes
// after 15:00 ICT (Req 39.11). Call once at server startup.
func (f *LiquidityFilter) Start(ctx context.Context) {
	if err := f.RefreshAll(ctx); err != nil {
		log.Printf("[LiquidityFilter] Initial refresh failed: %v", err)
	}
	go f.dailyRefreshLoop()
}

// Stop terminates the daily refresh loop.
func (f *LiquidityFilter) Stop() {
	close(f.stopCh)
}

// ComputeScore computes the tradability score (0-100) for a single symbol (Req 39.1).
func (f *LiquidityFilter) ComputeScore(ctx context.Context, symbol string) (LiquidityScore, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -(f.config.LookbackDays + 10)) // buffer for weekends/holidays

	quotes, exchange, err := f.fetchHistory(ctx, symbol, start, end)
	if err != nil {
		return LiquidityScore{}, err
	}

	return f.computeScoreFromQuotes(symbol, exchange, quotes), nil
}

// ClassifyTier returns the tier for a given score (Req 39.2).
//
//	score >= 70 → Tier 1 (highly liquid)
//	score 40-69 → Tier 2 (moderate)
//	score < 40  → Tier 3 (illiquid)
func (f *LiquidityFilter) ClassifyTier(score int) LiquidityTier {
	switch {
	case score >= f.config.Tier1Threshold:
		return Tier1
	case score >= f.config.Tier2Threshold:
		return Tier2
	default:
		return Tier3
	}
}

// FilterUniverse removes stocks below minTier from a symbol list.
// For example, minTier=Tier2 keeps Tier 1 and Tier 2, removing Tier 3.
func (f *LiquidityFilter) FilterUniverse(ctx context.Context, symbols []string, minTier LiquidityTier) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var filtered []string
	for _, sym := range symbols {
		if ls, ok := f.scores[sym]; ok && ls.Tier <= minTier {
			filtered = append(filtered, sym)
		}
	}
	return filtered, nil
}

// GetScore returns the cached score for a symbol, if available.
func (f *LiquidityFilter) GetScore(symbol string) (LiquidityScore, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	ls, ok := f.scores[symbol]
	return ls, ok
}

// IsWhitelisted checks if a symbol passes the minimum liquidity threshold
// (Tier 1 or Tier 2). Backward-compatible with the old whitelist API.
func (f *LiquidityFilter) IsWhitelisted(symbol string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	ls, ok := f.scores[symbol]
	return ok && ls.Tier <= Tier2
}

// GetWhitelist returns a backward-compatible whitelist snapshot.
func (f *LiquidityFilter) GetWhitelist() any {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entries := make([]WhitelistEntry, 0, len(f.scores))
	for _, ls := range f.scores {
		if ls.Tier <= Tier2 {
			entries = append(entries, WhitelistEntry{
				Symbol:   ls.Symbol,
				Exchange: ls.Exchange,
				ADTV:     ls.ADTV,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ADTV > entries[j].ADTV
	})

	return WhitelistSnapshot{
		Entries:   entries,
		Total:     len(entries),
		Config:    f.config,
		UpdatedAt: f.updatedAt,
	}
}

// FilterSymbols returns only symbols that are in Tier 1 or Tier 2.
// Backward-compatible with the old whitelist API.
func (f *LiquidityFilter) FilterSymbols(symbols []string) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var filtered []string
	for _, s := range symbols {
		if ls, ok := f.scores[s]; ok && ls.Tier <= Tier2 {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// GetSnapshot returns the full liquidity snapshot with all scores.
func (f *LiquidityFilter) GetSnapshot() LiquiditySnapshot {
	f.mu.RLock()
	defer f.mu.RUnlock()

	all := make([]LiquidityScore, 0, len(f.scores))
	for _, ls := range f.scores {
		all = append(all, ls)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Score > all[j].Score
	})

	return LiquiditySnapshot{
		Scores:    all,
		Total:     len(all),
		Config:    f.config,
		UpdatedAt: f.updatedAt,
	}
}

// RefreshAll recomputes tradability scores for all listed stocks (Req 39.11).
// Called daily after 15:00 ICT and at startup.
func (f *LiquidityFilter) RefreshAll(ctx context.Context) error {
	log.Println("[LiquidityFilter] Starting tradability score refresh...")
	startTime := time.Now()

	// Fetch all HOSE + HNX + UPCoM listings.
	listings, _, err := f.router.FetchListing(ctx, "")
	if err != nil {
		return err
	}

	// Build symbol → exchange map.
	symbolExchange := make(map[string]string, len(listings))
	var symbols []string
	for _, rec := range listings {
		if rec.Exchange == "HOSE" || rec.Exchange == "HNX" || rec.Exchange == "UPCOM" {
			symbolExchange[rec.Symbol] = rec.Exchange
			symbols = append(symbols, rec.Symbol)
		}
	}
	log.Printf("[LiquidityFilter] Evaluating %d tickers", len(symbols))

	// Fetch 20-day OHLCV for all symbols concurrently.
	end := time.Now()
	lookbackStart := end.AddDate(0, 0, -(f.config.LookbackDays + 10))

	type scoreResult struct {
		symbol string
		score  LiquidityScore
		err    error
	}

	resultsCh := make(chan scoreResult, len(symbols))
	sem := make(chan struct{}, 20) // concurrency limit

	for _, sym := range symbols {
		sem <- struct{}{}
		go func(symbol string) {
			defer func() { <-sem }()

			quotes, _, fetchErr := f.fetchHistory(ctx, symbol, lookbackStart, end)
			if fetchErr != nil {
				resultsCh <- scoreResult{symbol: symbol, err: fetchErr}
				return
			}

			ls := f.computeScoreFromQuotes(symbol, symbolExchange[symbol], quotes)
			resultsCh <- scoreResult{symbol: symbol, score: ls}
		}(sym)
	}

	// Collect results.
	newScores := make(map[string]LiquidityScore, len(symbols))
	for i := 0; i < len(symbols); i++ {
		res := <-resultsCh
		if res.err != nil {
			continue // skip symbols with fetch errors
		}
		newScores[res.symbol] = res.score
	}

	// Atomically swap scores.
	f.mu.Lock()
	f.scores = newScores
	f.updatedAt = time.Now()
	f.mu.Unlock()

	elapsed := time.Since(startTime)
	log.Printf("[LiquidityFilter] Refresh complete: %d scores computed (took %v)", len(newScores), elapsed)

	// Cache the snapshot with 24h TTL (Req 39.11).
	snapshot := f.GetSnapshot()
	f.cache.Set("liquidity:scores", snapshot, 24*time.Hour)

	return nil
}

// computeScoreFromQuotes calculates the composite tradability score from OHLCV data.
// Uses PriceBoard for real bid-ask spread and CompanyProfile.FreeFloat when available
// (vnstock-go v2).
func (f *LiquidityFilter) computeScoreFromQuotes(symbol, exchange string, quotes []vnstock.Quote) LiquidityScore {
	n := f.config.LookbackDays
	if len(quotes) < n {
		n = len(quotes)
	}
	if n == 0 {
		return LiquidityScore{
			Symbol:    symbol,
			Exchange:  exchange,
			Score:     0,
			Tier:      Tier3,
			UpdatedAt: time.Now(),
		}
	}

	recent := quotes[len(quotes)-n:]

	// --- Factor 1: ADTV (Average Daily Trading Value) ---
	var totalValue float64
	var totalVolume float64
	volumes := make([]float64, 0, n)
	zeroDays := 0

	for _, q := range recent {
		tv := q.Close * float64(q.Volume)
		totalValue += tv
		totalVolume += float64(q.Volume)
		volumes = append(volumes, float64(q.Volume))
		if q.Volume == 0 {
			zeroDays++
		}
	}

	adtv := totalValue / float64(n)
	avgVolume := totalVolume / float64(n)
	adtvScore := scoreADTV(adtv)

	// --- Factor 2: Volume Consistency (inverse of coefficient of variation) ---
	volConsistencyScore := scoreVolumeConsistency(volumes, avgVolume)

	// --- Factor 3: Bid-Ask Spread ---
	// Try real bid-ask spread from PriceBoard (vnstock-go v2).
	// Fall back to OHLCV range proxy if PriceBoard unavailable.
	spreadScore := f.fetchRealSpreadScore(symbol)
	if spreadScore < 0 {
		spreadScore = scoreSpreadProxy(recent)
	}

	// --- Factor 4: Zero-Volume Days penalty ---
	zeroVolScore := scoreZeroVolumeDays(zeroDays, n)

	// --- Factor 5: Free-Float Ratio ---
	// Use CompanyProfile().FreeFloat from vnstock-go v2 instead of neutral default.
	freeFloatScore, freeFloat := f.fetchFreeFloatScore(symbol)

	// --- Composite Score (weighted average) ---
	cfg := f.config
	composite := (adtvScore*cfg.WeightADTV +
		volConsistencyScore*cfg.WeightVolumeConsistency +
		spreadScore*cfg.WeightSpreadProxy +
		zeroVolScore*cfg.WeightZeroVolumeDays +
		freeFloatScore*cfg.WeightFreeFloat) / 100

	// Clamp to [0, 100].
	if composite > 100 {
		composite = 100
	}
	if composite < 0 {
		composite = 0
	}

	tier := f.ClassifyTier(composite)

	return LiquidityScore{
		Symbol:                 symbol,
		Exchange:               exchange,
		Score:                  composite,
		Tier:                   tier,
		ADTV:                   adtv,
		AvgVolume:              avgVolume,
		ZeroDays:               zeroDays,
		FreeFloat:              freeFloat,
		UpdatedAt:              time.Now(),
		ADTVScore:              adtvScore,
		VolumeConsistencyScore: volConsistencyScore,
		SpreadProxyScore:       spreadScore,
		ZeroVolumeDaysScore:    zeroVolScore,
		FreeFloatScore:         freeFloatScore,
	}
}

// fetchRealSpreadScore attempts to get the real bid-ask spread from PriceBoard.
// Returns -1 if unavailable, otherwise a 0-100 score.
func (f *LiquidityFilter) fetchRealSpreadScore(symbol string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	board, _, err := f.router.FetchPriceBoard(ctx, []string{symbol})
	if err != nil || len(board) == 0 {
		return -1 // signal to use proxy
	}

	entry := board[0]
	if entry.AskPrice1 <= 0 || entry.BidPrice1 <= 0 {
		return -1
	}

	mid := (entry.AskPrice1 + entry.BidPrice1) / 2
	if mid <= 0 {
		return -1
	}
	spread := (entry.AskPrice1 - entry.BidPrice1) / mid

	// spread < 0.2% → very tight (100), > 2% → very wide (0)
	switch {
	case spread <= 0.002:
		return 100
	case spread >= 0.02:
		return 0
	default:
		return int(100.0 * (0.02 - spread) / 0.018)
	}
}

// fetchFreeFloatScore estimates free-float from CompanyProfile shareholders.
// Returns (score, freeFloat). Falls back to neutral 50/0.0 if unavailable.
func (f *LiquidityFilter) fetchFreeFloatScore(symbol string) (int, float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	profile, _, err := f.router.FetchCompanyProfile(ctx, symbol)
	if err != nil || profile.ListedVolume <= 0 {
		return 50, 0.0 // neutral default
	}

	// Estimate free-float: 1 - (sum of major shareholder ownership).
	// Shareholders with >5% are typically reported.
	var majorPct float64
	for _, sh := range profile.Shareholders {
		if sh.Percentage > 0 {
			majorPct += sh.Percentage
		}
	}
	// majorPct is in percentage points (e.g. 65.0 means 65%).
	ff := 1.0 - majorPct/100.0
	if ff < 0 {
		ff = 0.05
	}
	if ff > 1 {
		ff = 1.0
	}

	// > 0.5 → score 100, < 0.1 → score 10
	switch {
	case ff >= 0.5:
		return 100, ff
	case ff <= 0.1:
		return 10, ff
	default:
		return 10 + int(90.0*(ff-0.1)/0.4), ff
	}
}

// scoreADTV maps ADTV (VND/day) to a 0-100 score using a log-linear curve.
func scoreADTV(adtv float64) int {
	if adtv <= 0 {
		return 0
	}
	if adtv >= adtvExcellent {
		return 100
	}
	// Log-linear interpolation between moderate (40) and excellent (100).
	logVal := math.Log10(adtv)
	logMin := math.Log10(adtvModerate)
	logMax := math.Log10(adtvExcellent)

	if logVal <= logMin {
		// Below moderate: linear scale 0-40.
		if adtv <= 0 {
			return 0
		}
		return int(40.0 * logVal / logMin)
	}
	// Between moderate and excellent: scale 40-100.
	return 40 + int(60.0*(logVal-logMin)/(logMax-logMin))
}

// scoreVolumeConsistency scores how consistent daily volume is.
// Low coefficient of variation = high consistency = high score.
func scoreVolumeConsistency(volumes []float64, avgVolume float64) int {
	if avgVolume <= 0 || len(volumes) < 2 {
		return 0
	}

	// Compute standard deviation.
	var sumSqDiff float64
	for _, v := range volumes {
		diff := v - avgVolume
		sumSqDiff += diff * diff
	}
	stdDev := math.Sqrt(sumSqDiff / float64(len(volumes)))
	cv := stdDev / avgVolume // coefficient of variation

	// CV < 0.3 → very consistent (score 100)
	// CV > 1.5 → very inconsistent (score 0)
	switch {
	case cv <= 0.3:
		return 100
	case cv >= 1.5:
		return 0
	default:
		// Linear interpolation: 100 at cv=0.3, 0 at cv=1.5
		return int(100.0 * (1.5 - cv) / 1.2)
	}
}

// scoreSpreadProxy uses the average (high-low)/close ratio as a bid-ask spread proxy.
// Tighter spreads (lower ratio) = more liquid = higher score.
func scoreSpreadProxy(quotes []vnstock.Quote) int {
	if len(quotes) == 0 {
		return 0
	}

	var totalRatio float64
	validDays := 0
	for _, q := range quotes {
		if q.Close <= 0 || q.Volume == 0 {
			continue
		}
		spread := (q.High - q.Low) / q.Close
		totalRatio += spread
		validDays++
	}

	if validDays == 0 {
		return 0
	}

	avgSpread := totalRatio / float64(validDays)

	// avgSpread < 1% → very tight (score 100)
	// avgSpread > 5% → very wide (score 0)
	switch {
	case avgSpread <= 0.01:
		return 100
	case avgSpread >= 0.05:
		return 0
	default:
		return int(100.0 * (0.05 - avgSpread) / 0.04)
	}
}

// scoreZeroVolumeDays penalizes stocks with zero-volume trading days.
// 0 zero days → 100, maxAllowed → 40, all days zero → 0.
func scoreZeroVolumeDays(zeroDays, totalDays int) int {
	if totalDays == 0 {
		return 0
	}
	if zeroDays == 0 {
		return 100
	}
	ratio := float64(zeroDays) / float64(totalDays)
	// Linear: 100 at 0%, 0 at 100%.
	score := int(100.0 * (1.0 - ratio))
	if score < 0 {
		score = 0
	}
	return score
}

// fetchHistory fetches OHLCV history for a symbol via the DataSourceRouter.
func (f *LiquidityFilter) fetchHistory(ctx context.Context, symbol string, start, end time.Time) ([]vnstock.Quote, string, error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    start,
		End:      end,
		Interval: "1D",
	}
	return f.router.FetchQuoteHistory(ctx, req)
}

// dailyRefreshLoop runs RefreshAll every day after 15:00 ICT (Req 39.11).
func (f *LiquidityFilter) dailyRefreshLoop() {
	for {
		now := time.Now().In(ictZone)
		// Target: 15:30 ICT (30 min after market close for data settlement).
		next := time.Date(now.Year(), now.Month(), now.Day(), refreshHourICT, 30, 0, 0, ictZone)
		if now.After(next) {
			next = next.AddDate(0, 0, 1)
		}
		waitDuration := time.Until(next)

		log.Printf("[LiquidityFilter] Next refresh at %s (in %v)", next.Format("2006-01-02 15:04 MST"), waitDuration)

		select {
		case <-time.After(waitDuration):
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := f.RefreshAll(ctx); err != nil {
				log.Printf("[LiquidityFilter] Daily refresh failed: %v", err)
			}
			cancel()
		case <-f.stopCh:
			log.Println("[LiquidityFilter] Daily refresh loop stopped")
			return
		}
	}
}
