package service

import (
	"context"
	"log"
	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
	"sort"
	"sync"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

// VN100 constituents — always included regardless of ADTV.
// Source: HOSE VN100 index basket (top 100 by market cap & liquidity).
var vn100Tickers = map[string]bool{
	// VN30 (top 30)
	"ACB": true, "BCM": true, "BID": true, "BVH": true, "CTG": true,
	"FPT": true, "GAS": true, "GVR": true, "HDB": true, "HPG": true,
	"MBB": true, "MSN": true, "MWG": true, "PLX": true, "POW": true,
	"SAB": true, "SHB": true, "SSB": true, "SSI": true, "STB": true,
	"TCB": true, "TPB": true, "VCB": true, "VHM": true, "VIB": true,
	"VIC": true, "VJC": true, "VNM": true, "VPB": true, "VRE": true,
	// VN100 (next 70)
	"ANV": true, "ASM": true, "BAF": true, "BSI": true, "BWE": true,
	"CII": true, "CMG": true, "CRE": true, "CTR": true, "DCM": true,
	"DGC": true, "DGW": true, "DIG": true, "DPM": true, "DXG": true,
	"DXS": true, "ELC": true, "EVF": true, "FRT": true, "GEX": true,
	"GMD": true, "HAH": true, "HCM": true, "HDC": true, "HDG": true,
	"HHV": true, "HSG": true, "HT1": true, "IMP": true, "KBC": true,
	"KDC": true, "KDH": true, "KOS": true, "LPB": true, "MSB": true,
	"NLG": true, "NT2": true, "NVL": true, "OCB": true, "PAN": true,
	"PC1": true, "PDR": true, "PET": true, "PHR": true, "PNJ": true,
	"PPC": true, "PVD": true, "PVT": true, "REE": true, "SBT": true,
	"SCS": true, "SIP": true, "SJS": true, "SZC": true, "TCH": true,
	"TLG": true, "TNH": true, "VCA": true, "VCG": true, "VCI": true,
	"VDS": true, "VGC": true, "VHC": true, "VIX": true, "VND": true,
	"VOS": true, "VPI": true, "VTP": true, "VTO": true,
}

// LiquidityFilter computes and maintains a dynamic whitelist of tradable stocks.
// It filters the ~1600 VN tickers down to ~150-300 "real player" stocks based on
// ADTV (Average Daily Trading Value) and minimum price thresholds.
// VN100 constituents are always included regardless of filter results.
type LiquidityFilter struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
	config model.LiquidityConfig

	mu        sync.RWMutex
	whitelist map[string]model.WhitelistEntry
	updatedAt time.Time
	stopCh    chan struct{}
}

// NewLiquidityFilter creates a new filter with default config.
func NewLiquidityFilter(router *infra.DataSourceRouter, cache *infra.Cache) *LiquidityFilter {
	return &LiquidityFilter{
		router:    router,
		cache:     cache,
		config:    model.DefaultLiquidityConfig(),
		whitelist: make(map[string]model.WhitelistEntry),
		stopCh:    make(chan struct{}),
	}
}

// NewLiquidityFilterWithConfig creates a filter with custom thresholds.
func NewLiquidityFilterWithConfig(router *infra.DataSourceRouter, cache *infra.Cache, cfg model.LiquidityConfig) *LiquidityFilter {
	f := NewLiquidityFilter(router, cache)
	f.config = cfg
	return f
}

// Start runs the initial filter computation and schedules daily refreshes.
// Call this once at server startup.
func (f *LiquidityFilter) Start(ctx context.Context) {
	// Run initial computation
	if err := f.Refresh(ctx); err != nil {
		log.Printf("[LiquidityFilter] Initial refresh failed: %v (whitelist will contain VN100 only)", err)
		f.seedVN100Only()
	}

	// Schedule daily refresh at 17:00 ICT (UTC+7) = 10:00 UTC
	go f.dailyRefreshLoop()
}

// Stop terminates the daily refresh loop.
func (f *LiquidityFilter) Stop() {
	close(f.stopCh)
}

// IsWhitelisted checks if a ticker is in the active whitelist. O(1) lookup.
func (f *LiquidityFilter) IsWhitelisted(symbol string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.whitelist[symbol]
	return ok
}

// GetWhitelist returns the current whitelist snapshot.
func (f *LiquidityFilter) GetWhitelist() model.WhitelistSnapshot {
	f.mu.RLock()
	defer f.mu.RUnlock()

	entries := make([]model.WhitelistEntry, 0, len(f.whitelist))
	for _, e := range f.whitelist {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ADTV > entries[j].ADTV // highest liquidity first
	})

	return model.WhitelistSnapshot{
		Entries:   entries,
		Total:     len(entries),
		Config:    f.config,
		UpdatedAt: f.updatedAt,
	}
}

// FilterSymbols takes a list of symbols and returns only those in the whitelist.
func (f *LiquidityFilter) FilterSymbols(symbols []string) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var filtered []string
	for _, s := range symbols {
		if _, ok := f.whitelist[s]; ok {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

// Refresh recomputes the whitelist from scratch.
// Fetches all listings, then computes 20-day ADTV for each ticker.
func (f *LiquidityFilter) Refresh(ctx context.Context) error {
	log.Println("[LiquidityFilter] Starting whitelist refresh...")
	start := time.Now()

	// Step 1: Fetch all HOSE + HNX + UPCoM listings
	listings, _, err := f.router.FetchListing(ctx, "")
	if err != nil {
		return err
	}

	// Build lookup: symbol -> exchange
	symbolExchange := make(map[string]string, len(listings))
	var symbols []string
	for _, rec := range listings {
		// Include HOSE, HNX, and UPCoM
		if rec.Exchange == "HOSE" || rec.Exchange == "HNX" || rec.Exchange == "UPCOM" {
			symbolExchange[rec.Symbol] = rec.Exchange
			symbols = append(symbols, rec.Symbol)
		}
	}
	log.Printf("[LiquidityFilter] Found %d tickers to evaluate (HOSE/HNX/UPCoM)", len(symbols))

	// Step 2: Compute ADTV for each symbol in batches
	newWhitelist := make(map[string]model.WhitelistEntry)

	// First, always add VN100 (we'll update their ADTV data if available)
	for ticker := range vn100Tickers {
		exchange := symbolExchange[ticker]
		if exchange == "" {
			exchange = "HOSE" // VN100 are predominantly HOSE
		}
		newWhitelist[ticker] = model.WhitelistEntry{
			Symbol:   ticker,
			Exchange: exchange,
			IsVN100:  true,
		}
	}

	// Step 3: Fetch 20-day OHLCV for all symbols in concurrent batches
	end := time.Now()
	lookbackStart := end.AddDate(0, 0, -(f.config.LookbackDays + 10)) // extra buffer for weekends/holidays

	type adtvResult struct {
		symbol string
		adtv   float64
		price  float64
		err    error
	}

	// Process in batches of 20 to avoid overwhelming the API
	batchSize := 20
	resultsCh := make(chan adtvResult, len(symbols))
	sem := make(chan struct{}, batchSize)

	for _, sym := range symbols {
		sem <- struct{}{}
		go func(symbol string) {
			defer func() { <-sem }()

			adtv, price, err := f.computeADTV(ctx, symbol, lookbackStart, end)
			resultsCh <- adtvResult{symbol: symbol, adtv: adtv, price: price, err: err}
		}(sym)
	}

	// Collect results
	for i := 0; i < len(symbols); i++ {
		res := <-resultsCh
		if res.err != nil {
			// If it's a VN100 ticker, keep it in whitelist even without data
			if vn100Tickers[res.symbol] {
				continue
			}
			// Skip non-VN100 tickers that fail
			continue
		}

		// Apply filters — UPCoM uses stricter ADTV threshold
		minADTV := f.config.MinADTV
		if symbolExchange[res.symbol] == "UPCOM" {
			minADTV = f.config.MinADTVUpcom
		}
		passesADTV := res.adtv >= minADTV
		passesPrice := res.price >= f.config.MinPrice
		isVN100 := vn100Tickers[res.symbol]

		if passesADTV && passesPrice || isVN100 {
			exchange := symbolExchange[res.symbol]
			newWhitelist[res.symbol] = model.WhitelistEntry{
				Symbol:   res.symbol,
				Exchange: exchange,
				ADTV:     res.adtv,
				Price:    res.price,
				IsVN100:  isVN100,
			}
		}
	}

	// Step 4: Atomically swap the whitelist
	f.mu.Lock()
	f.whitelist = newWhitelist
	f.updatedAt = time.Now()
	f.mu.Unlock()

	elapsed := time.Since(start)
	log.Printf("[LiquidityFilter] Whitelist refreshed: %d tickers passed (took %v)", len(newWhitelist), elapsed)

	// Cache the snapshot
	snapshot := f.GetWhitelist()
	f.cache.Set("liquidity:whitelist", snapshot, 24*time.Hour)

	return nil
}

// computeADTV fetches 20-day OHLCV and computes Average Daily Trading Value.
// ADTV = mean(close_i × volume_i) over the lookback period.
func (f *LiquidityFilter) computeADTV(ctx context.Context, symbol string, start, end time.Time) (adtv float64, lastPrice float64, err error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    start,
		End:      end,
		Interval: "1D",
	}

	quotes, _, fetchErr := f.router.FetchQuoteHistory(ctx, req)
	if fetchErr != nil {
		return 0, 0, fetchErr
	}

	if len(quotes) == 0 {
		return 0, 0, nil
	}

	// Take the last N trading days (up to LookbackDays)
	n := f.config.LookbackDays
	if len(quotes) < n {
		n = len(quotes)
	}
	recent := quotes[len(quotes)-n:]

	var totalValue float64
	for _, q := range recent {
		// Trading value = close price × volume
		totalValue += q.Close * float64(q.Volume)
	}

	adtv = totalValue / float64(n)
	lastPrice = quotes[len(quotes)-1].Close

	return adtv, lastPrice, nil
}

// seedVN100Only populates the whitelist with just VN100 tickers (fallback).
func (f *LiquidityFilter) seedVN100Only() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.whitelist = make(map[string]model.WhitelistEntry, len(vn100Tickers))
	for ticker := range vn100Tickers {
		f.whitelist[ticker] = model.WhitelistEntry{
			Symbol:   ticker,
			Exchange: "HOSE",
			IsVN100:  true,
		}
	}
	f.updatedAt = time.Now()
	log.Printf("[LiquidityFilter] Seeded VN100-only whitelist: %d tickers", len(f.whitelist))
}

// dailyRefreshLoop runs Refresh every day at ~17:00 ICT (after market close).
func (f *LiquidityFilter) dailyRefreshLoop() {
	for {
		now := time.Now()
		// Target 17:00 ICT = 10:00 UTC
		ict := time.FixedZone("ICT", 7*3600)
		next := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, ict)
		if now.In(ict).After(next.In(ict)) {
			next = next.AddDate(0, 0, 1)
		}
		waitDuration := time.Until(next)

		log.Printf("[LiquidityFilter] Next refresh scheduled in %v", waitDuration)

		select {
		case <-time.After(waitDuration):
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			if err := f.Refresh(ctx); err != nil {
				log.Printf("[LiquidityFilter] Daily refresh failed: %v", err)
			}
			cancel()
		case <-f.stopCh:
			log.Println("[LiquidityFilter] Daily refresh loop stopped")
			return
		}
	}
}
