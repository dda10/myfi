package infra

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	vnstock "github.com/dda10/vnstock-go"
)

type DataSourceRouter struct {
	preferences     map[DataCategory]SourcePreference
	clients         map[string]*vnstock.Client
	rateLimiter     *RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

type connectorDef struct {
	name     string
	required bool
}

var allConnectors = []connectorDef{
	{"VCI", true}, {"KBS", false}, {"VND", false}, {"ENTRADE", false},
	{"CAFEF", false}, {"VND_FINFO", false}, {"MSN", false}, {"GOLD", false}, {"FMARKET", false},
}

func NewDataSourceRouter() (*DataSourceRouter, error) {
	clients := make(map[string]*vnstock.Client)
	cbs := make(map[string]*CircuitBreaker)
	for _, cd := range allConnectors {
		c, err := vnstock.New(vnstock.Config{
			Connector:       cd.name,
			Timeout:         10 * time.Second,
			CacheMaxEntries: 256,             // LRU cache for listings, profiles, financials
			CacheTTL:        5 * time.Minute, // Reduce Redis load for slow-changing data
		})
		if err != nil {
			if cd.required {
				return nil, fmt.Errorf("init %s: %w", cd.name, err)
			}
			log.Printf("[DSR] %s init failed: %v (skipping)", cd.name, err)
			continue
		}
		clients[cd.name] = c
		cbs[cd.name] = NewCircuitBreaker(3, 60*time.Second)
		log.Printf("[DSR] %s initialized", cd.name)
	}
	prefs := map[DataCategory]SourcePreference{
		PriceQuotes:     {Category: PriceQuotes, Primary: "VCI", Fallbacks: []string{"KBS", "VND", "ENTRADE"}},
		OHLCVHistory:    {Category: OHLCVHistory, Primary: "VCI", Fallbacks: []string{"KBS", "VND", "ENTRADE"}},
		IntradayData:    {Category: IntradayData, Primary: "VCI", Fallbacks: []string{"KBS", "VND"}},
		OrderBook:       {Category: OrderBook, Primary: "VCI", Fallbacks: []string{"VND"}},
		CompanyOverview: {Category: CompanyOverview, Primary: "VCI", Fallbacks: []string{"KBS", "VND", "CAFEF"}},
		Shareholders:    {Category: Shareholders, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO"}},
		Officers:        {Category: Officers, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO"}},
		News:            {Category: News, Primary: "VCI", Fallbacks: []string{"CAFEF", "VND"}},
		IncomeStatement: {Category: IncomeStatement, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO", "CAFEF"}},
		BalanceSheet:    {Category: BalanceSheet, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO", "CAFEF"}},
		CashFlow:        {Category: CashFlow, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO", "CAFEF"}},
		FinancialRatios: {Category: FinancialRatios, Primary: "VCI", Fallbacks: []string{"KBS", "VND_FINFO"}},
		PriceBoard:      {Category: PriceBoard, Primary: "VCI", Fallbacks: []string{"VND"}},
		PriceDepth:      {Category: PriceDepth, Primary: "VCI", Fallbacks: []string{"VND"}},
		Screener:        {Category: Screener, Primary: "VCI", Fallbacks: []string{"VND_FINFO"}},
		GoldPrice:       {Category: GoldPrice, Primary: "GOLD", Fallbacks: []string{"MSN"}},
		FXRate:          {Category: FXRate, Primary: "VCI", Fallbacks: []string{"MSN"}},
		WorldIndices:    {Category: WorldIndices, Primary: "MSN", Fallbacks: []string{"CAFEF"}},
	}
	return &DataSourceRouter{preferences: prefs, clients: clients, rateLimiter: NewRateLimiter(), circuitBreakers: cbs}, nil
}

func (r *DataSourceRouter) RateLimiter() *RateLimiter          { return r.rateLimiter }
func (r *DataSourceRouter) VCIClient() *vnstock.Client         { return r.clients["VCI"] }
func (r *DataSourceRouter) KBSClient() *vnstock.Client         { return r.clients["KBS"] }
func (r *DataSourceRouter) MSNClient() *vnstock.Client         { return r.clients["MSN"] }
func (r *DataSourceRouter) GOLDClient() *vnstock.Client        { return r.clients["GOLD"] }
func (r *DataSourceRouter) FMARKETClient() *vnstock.Client     { return r.clients["FMARKET"] }
func (r *DataSourceRouter) Client(name string) *vnstock.Client { return r.clients[name] }

func (r *DataSourceRouter) selectSource(cat DataCategory) (*vnstock.Client, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.preferences[cat]; ok {
		if c := r.clients[p.Primary]; c != nil {
			return c, p.Primary
		}
	}
	return r.clients["VCI"], "VCI"
}

type fallbackEntry struct {
	client *vnstock.Client
	name   string
}

func (r *DataSourceRouter) getFallbackSources(cat DataCategory) []fallbackEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []fallbackEntry
	for _, fb := range r.preferences[cat].Fallbacks {
		if c := r.clients[fb]; c != nil {
			out = append(out, fallbackEntry{c, fb})
		}
	}
	return out
}

func isTripWorthy(err error) bool {
	if err == nil {
		return false
	}
	var ve *vnstock.Error
	if vnstock.AsError(err, &ve) {
		switch ve.Code {
		case vnstock.NetworkError, vnstock.RateLimited:
			return true
		case vnstock.NotSupported, vnstock.NoData:
			return false
		}
	}
	return true
}

func (r *DataSourceRouter) callWithBreaker(src string, fn func() error) error {
	cb, ok := r.circuitBreakers[src]
	if !ok {
		return fn()
	}
	var outerErr error
	cbErr := cb.Call(func() error {
		outerErr = fn()
		if outerErr != nil && !isTripWorthy(outerErr) {
			return nil
		}
		return outerErr
	})
	if cbErr != nil {
		return cbErr
	}
	return outerErr
}

func (r *DataSourceRouter) FetchRealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, string, error) {
	pc, ps := r.selectSource(PriceQuotes)
	if err := r.rateLimiter.Allow(ps); err == nil {
		var q []vnstock.Quote
		cbErr := r.callWithBreaker(ps, func() error {
			var e error
			q, e = r.fetchQuotesWithTimeout(ctx, pc, symbols, 10*time.Second)
			return e
		})
		if cbErr == nil && !r.isEmptyData(q) {
			return q, ps, nil
		}
	}
	for _, fb := range r.getFallbackSources(PriceQuotes) {
		if r.rateLimiter.Allow(fb.name) != nil {
			continue
		}
		var q []vnstock.Quote
		cbErr := r.callWithBreaker(fb.name, func() error {
			var e error
			q, e = r.fetchQuotesWithTimeout(ctx, fb.client, symbols, 10*time.Second)
			return e
		})
		if cbErr == nil && !r.isEmptyData(q) {
			return q, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for quotes")
}

func (r *DataSourceRouter) FetchQuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, string, error) {
	pc, ps := r.selectSource(OHLCVHistory)
	if h, err := r.fetchHistoryWithTimeout(ctx, pc, req, 10*time.Second); err == nil && len(h) > 0 {
		return h, ps, nil
	}
	for _, fb := range r.getFallbackSources(OHLCVHistory) {
		if h, err := r.fetchHistoryWithTimeout(ctx, fb.client, req, 10*time.Second); err == nil && len(h) > 0 {
			return h, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for OHLCV %s", req.Symbol)
}

func (r *DataSourceRouter) FetchIndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, string, error) {
	pc, ps := r.selectSource(OHLCVHistory)
	if recs, err := pc.IndexHistory(ctx, req); err == nil && len(recs) > 0 {
		return recs, ps, nil
	}
	for _, fb := range r.getFallbackSources(OHLCVHistory) {
		if recs, err := fb.client.IndexHistory(ctx, req); err == nil && len(recs) > 0 {
			return recs, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for index %s", req.Name)
}

func (r *DataSourceRouter) FetchListing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, string, error) {
	pc, ps := r.selectSource(CompanyOverview)
	if recs, err := pc.Listing(ctx, exchange); err == nil && len(recs) > 0 {
		return recs, ps, nil
	}
	for _, fb := range r.getFallbackSources(CompanyOverview) {
		if recs, err := fb.client.Listing(ctx, exchange); err == nil && len(recs) > 0 {
			return recs, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for listing")
}

func (r *DataSourceRouter) FetchFinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, string, error) {
	pc, ps := r.selectSource(FinancialRatios)
	if p, err := pc.FinancialStatement(ctx, req); err == nil && len(p) > 0 {
		return p, ps, nil
	}
	for _, fb := range r.getFallbackSources(FinancialRatios) {
		if p, err := fb.client.FinancialStatement(ctx, req); err == nil && len(p) > 0 {
			return p, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for financial %s", req.Symbol)
}

func (r *DataSourceRouter) FetchCompanyProfile(ctx context.Context, symbol string) (vnstock.CompanyProfile, string, error) {
	pc, ps := r.selectSource(CompanyOverview)
	if prof, err := pc.CompanyProfile(ctx, symbol); err == nil {
		return prof, ps, nil
	}
	for _, fb := range r.getFallbackSources(CompanyOverview) {
		if prof, err := fb.client.CompanyProfile(ctx, symbol); err == nil {
			return prof, fb.name, nil
		}
	}
	return vnstock.CompanyProfile{}, "", fmt.Errorf("all sources failed for profile %s", symbol)
}

func (r *DataSourceRouter) FetchSubsidiaries(ctx context.Context, symbol string) ([]vnstock.Subsidiary, string, error) {
	pc, ps := r.selectSource(CompanyOverview)
	if s, err := pc.Subsidiaries(ctx, symbol); err == nil {
		return s, ps, nil
	}
	for _, fb := range r.getFallbackSources(CompanyOverview) {
		if s, err := fb.client.Subsidiaries(ctx, symbol); err == nil {
			return s, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for subsidiaries %s", symbol)
}

func (r *DataSourceRouter) FetchCompanyEvents(ctx context.Context, symbol string) ([]vnstock.CompanyEvent, string, error) {
	pc, ps := r.selectSource(CompanyOverview)
	if ev, err := pc.CompanyEvents(ctx, symbol); err == nil {
		return ev, ps, nil
	}
	for _, fb := range r.getFallbackSources(CompanyOverview) {
		if ev, err := fb.client.CompanyEvents(ctx, symbol); err == nil {
			return ev, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for events %s", symbol)
}

func (r *DataSourceRouter) FetchCompanyNews(ctx context.Context, symbol string) ([]vnstock.CompanyNews, string, error) {
	pc, ps := r.selectSource(News)
	if n, err := pc.CompanyNews(ctx, symbol); err == nil {
		return n, ps, nil
	}
	for _, fb := range r.getFallbackSources(News) {
		if n, err := fb.client.CompanyNews(ctx, symbol); err == nil {
			return n, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for news %s", symbol)
}

func (r *DataSourceRouter) FetchInsiderTrades(ctx context.Context, symbol string) ([]vnstock.InsiderTrade, string, error) {
	pc, ps := r.selectSource(CompanyOverview)
	if t, err := pc.InsiderTrades(ctx, symbol); err == nil {
		return t, ps, nil
	}
	for _, fb := range r.getFallbackSources(CompanyOverview) {
		if t, err := fb.client.InsiderTrades(ctx, symbol); err == nil {
			return t, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for insider trades %s", symbol)
}

func (r *DataSourceRouter) FetchPriceBoard(ctx context.Context, symbols []string) ([]vnstock.PriceBoard, string, error) {
	pc, ps := r.selectSource(PriceBoard)
	if b, err := pc.PriceBoard(ctx, symbols); err == nil && len(b) > 0 {
		return b, ps, nil
	}
	for _, fb := range r.getFallbackSources(PriceBoard) {
		if b, err := fb.client.PriceBoard(ctx, symbols); err == nil && len(b) > 0 {
			return b, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for price board")
}

func (r *DataSourceRouter) FetchPriceDepth(ctx context.Context, symbol string) (vnstock.PriceDepth, string, error) {
	pc, ps := r.selectSource(PriceDepth)
	if d, err := pc.PriceDepth(ctx, symbol); err == nil {
		return d, ps, nil
	}
	for _, fb := range r.getFallbackSources(PriceDepth) {
		if d, err := fb.client.PriceDepth(ctx, symbol); err == nil {
			return d, fb.name, nil
		}
	}
	return vnstock.PriceDepth{}, "", fmt.Errorf("all sources failed for depth %s", symbol)
}

func (r *DataSourceRouter) FetchScreen(ctx context.Context, criteria vnstock.ScreenerCriteria) ([]string, string, error) {
	pc, ps := r.selectSource(Screener)
	if res, err := pc.Screen(ctx, criteria); err == nil {
		return res, ps, nil
	}
	for _, fb := range r.getFallbackSources(Screener) {
		if res, err := fb.client.Screen(ctx, criteria); err == nil {
			return res, fb.name, nil
		}
	}
	return nil, "", fmt.Errorf("all sources failed for screener")
}

func (r *DataSourceRouter) fetchQuotesWithTimeout(ctx context.Context, client *vnstock.Client, symbols []string, timeout time.Duration) ([]vnstock.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	type qr struct {
		q []vnstock.Quote
		e error
	}
	ch := make(chan qr, 1)
	go func() {
		q, e := client.RealTimeQuotes(ctx, symbols)
		ch <- qr{q, e}
	}()
	select {
	case v := <-ch:
		return v.q, v.e
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout after %v", timeout)
	}
}

func (r *DataSourceRouter) fetchHistoryWithTimeout(ctx context.Context, client *vnstock.Client, req vnstock.QuoteHistoryRequest, timeout time.Duration) ([]vnstock.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	type hr struct {
		h []vnstock.Quote
		e error
	}
	ch := make(chan hr, 1)
	go func() {
		h, e := client.QuoteHistory(ctx, req)
		ch <- hr{h, e}
	}()
	select {
	case v := <-ch:
		return v.h, v.e
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout after %v", timeout)
	}
}

func (r *DataSourceRouter) isEmptyData(quotes []vnstock.Quote) bool {
	for _, q := range quotes {
		if q.Close > 0 {
			return false
		}
	}
	return true
}

func (r *DataSourceRouter) GetSourcePreferences() map[DataCategory]SourcePreference {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[DataCategory]SourcePreference, len(r.preferences))
	for k, v := range r.preferences {
		out[k] = v
	}
	return out
}
