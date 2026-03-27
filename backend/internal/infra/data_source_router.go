package infra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"myfi-backend/internal/model"

	"github.com/dda10/vnstock-go"
)

// DataSourceRouter manages intelligent routing between VCI and KBS data sources
type DataSourceRouter struct {
	preferences     map[model.DataCategory]model.SourcePreference
	vciClient       *vnstock.Client
	kbsClient       *vnstock.Client
	rateLimiter     *RateLimiter
	circuitBreakers map[string]*CircuitBreaker
	mu              sync.RWMutex
}

// NewDataSourceRouter creates a new router with default source preferences
func NewDataSourceRouter() (*DataSourceRouter, error) {
	// Initialize VCI client (primary source - 77 columns)
	vciCfg := vnstock.Config{
		Connector: "VCI",
		Timeout:   10 * time.Second,
	}
	vciClient, err := vnstock.New(vciCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize VCI client: %w", err)
	}

	// Initialize KBS client (fallback source)
	kbsCfg := vnstock.Config{
		Connector: "KBS",
		Timeout:   10 * time.Second,
	}
	kbsClient, err := vnstock.New(kbsCfg)
	if err != nil {
		log.Printf("[DataSourceRouter] Warning: failed to initialize KBS client: %v (will use VCI only)", err)
		kbsClient = nil
	} else {
		log.Printf("[DataSourceRouter] KBS client initialized successfully as fallback")
	}

	// Define source preferences: VCI primary, KBS fallback
	preferences := map[model.DataCategory]model.SourcePreference{
		model.PriceQuotes: {
			Category: model.PriceQuotes,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.OHLCVHistory: {
			Category: model.OHLCVHistory,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.IntradayData: {
			Category: model.IntradayData,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.OrderBook: {
			Category: model.OrderBook,
			Primary:  "VCI",
			Fallback: "VCI", // KBS doesn't support order book
		},
		model.CompanyOverview: {
			Category: model.CompanyOverview,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.Shareholders: {
			Category: model.Shareholders,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.Officers: {
			Category: model.Officers,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.News: {
			Category: model.News,
			Primary:  "VCI",
			Fallback: "VCI", // KBS doesn't support news
		},
		model.IncomeStatement: {
			Category: model.IncomeStatement,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.BalanceSheet: {
			Category: model.BalanceSheet,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.CashFlow: {
			Category: model.CashFlow,
			Primary:  "VCI",
			Fallback: "KBS",
		},
		model.FinancialRatios: {
			Category: model.FinancialRatios,
			Primary:  "VCI",
			Fallback: "KBS",
		},
	}

	return &DataSourceRouter{
		preferences: preferences,
		vciClient:   vciClient,
		kbsClient:   kbsClient,
		rateLimiter: NewRateLimiter(),
		circuitBreakers: map[string]*CircuitBreaker{
			"VCI": NewCircuitBreaker(3, 60*time.Second),
			"KBS": NewCircuitBreaker(3, 60*time.Second),
		},
	}, nil
}

// RateLimiter returns the router's rate limiter instance
func (r *DataSourceRouter) RateLimiter() *RateLimiter {
	return r.rateLimiter
}

// VCIClient returns the VCI client for direct API calls (e.g., IndexCurrent, CompanyProfile)
func (r *DataSourceRouter) VCIClient() *vnstock.Client {
	return r.vciClient
}

// selectSource returns the appropriate client based on category preference
func (r *DataSourceRouter) selectSource(category model.DataCategory) (*vnstock.Client, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pref, exists := r.preferences[category]
	if !exists {
		// Default to VCI if category not found
		log.Printf("[DataSourceRouter] Category %s not found in preferences, defaulting to VCI", category)
		return r.vciClient, "VCI"
	}

	if pref.Primary == "VCI" {
		return r.vciClient, "VCI"
	}
	return r.kbsClient, "KBS"
}

// getFallbackSource returns the fallback client for a category
func (r *DataSourceRouter) getFallbackSource(category model.DataCategory) (*vnstock.Client, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pref, exists := r.preferences[category]
	if !exists {
		log.Printf("[DataSourceRouter] Category %s not found in preferences, defaulting to VCI fallback", category)
		return r.vciClient, "VCI"
	}

	// Return KBS if configured as fallback and available
	if pref.Fallback == "KBS" && r.kbsClient != nil {
		return r.kbsClient, "KBS"
	}
	return r.vciClient, "VCI"
}

// FetchRealTimeQuotes fetches real-time quotes with automatic failover
func (r *DataSourceRouter) FetchRealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, string, error) {
	category := model.PriceQuotes
	startTime := time.Now()

	// Try primary source
	primaryClient, primarySource := r.selectSource(category)
	log.Printf("[DataSourceRouter] Fetching %s from primary source: %s", category, primarySource)

	// Check rate limit
	if err := r.rateLimiter.Allow(primarySource); err != nil {
		log.Printf("[DataSourceRouter] Rate limit exceeded for %s: %v", primarySource, err)
		// Try fallback immediately
		fallbackClient, fallbackSource := r.getFallbackSource(category)
		if err := r.rateLimiter.Allow(fallbackSource); err != nil {
			return nil, "", fmt.Errorf("rate limit exceeded for both sources")
		}
		primaryClient = fallbackClient
		primarySource = fallbackSource
	}

	// Use circuit breaker
	var quotes []vnstock.Quote
	var fetchErr error

	cb := r.circuitBreakers[primarySource]
	err := cb.Call(func() error {
		quotes, fetchErr = r.fetchQuotesWithTimeout(ctx, primaryClient, symbols, 10*time.Second)
		return fetchErr
	})

	if err == nil && !r.isEmptyData(quotes) {
		elapsed := time.Since(startTime)
		log.Printf("[DataSourceRouter] Success: %s from %s (response time: %v)", category, primarySource, elapsed)
		return quotes, primarySource, nil
	}

	// Log primary failure reason
	if err != nil {
		log.Printf("[DataSourceRouter] Primary source %s failed for %s: %v", primarySource, category, err)
	} else {
		log.Printf("[DataSourceRouter] Primary source %s returned empty data for %s", primarySource, category)
	}

	// Try fallback source
	fallbackClient, fallbackSource := r.getFallbackSource(category)
	log.Printf("[DataSourceRouter] Attempting fallback source: %s", fallbackSource)

	// Check rate limit for fallback
	if err := r.rateLimiter.Allow(fallbackSource); err != nil {
		return nil, "", fmt.Errorf("rate limit exceeded for fallback source %s", fallbackSource)
	}

	// Use circuit breaker for fallback
	fallbackCB := r.circuitBreakers[fallbackSource]
	err = fallbackCB.Call(func() error {
		quotes, fetchErr = r.fetchQuotesWithTimeout(ctx, fallbackClient, symbols, 10*time.Second)
		return fetchErr
	})

	if err == nil && !r.isEmptyData(quotes) {
		elapsed := time.Since(startTime)
		log.Printf("[DataSourceRouter] Success: %s from fallback %s (response time: %v)", category, fallbackSource, elapsed)
		return quotes, fallbackSource, nil
	}

	// Both sources failed
	if err != nil {
		log.Printf("[DataSourceRouter] Fallback source %s also failed for %s: %v", fallbackSource, category, err)
		return nil, "", fmt.Errorf("both sources failed: primary=%s, fallback=%s: %w", primarySource, fallbackSource, err)
	}

	log.Printf("[DataSourceRouter] Fallback source %s also returned empty data for %s", fallbackSource, category)
	return nil, "", errors.New("both sources returned empty data")
}

// FetchQuoteHistory fetches historical OHLCV data with automatic failover
func (r *DataSourceRouter) FetchQuoteHistory(ctx context.Context, req vnstock.QuoteHistoryRequest) ([]vnstock.Quote, string, error) {
	category := model.OHLCVHistory
	startTime := time.Now()

	// Try primary source
	primaryClient, primarySource := r.selectSource(category)
	log.Printf("[DataSourceRouter] Fetching %s for %s from primary source: %s", category, req.Symbol, primarySource)

	history, err := r.fetchHistoryWithTimeout(ctx, primaryClient, req, 10*time.Second)
	if err == nil && len(history) > 0 {
		elapsed := time.Since(startTime)
		log.Printf("[DataSourceRouter] Success: %s for %s from %s (response time: %v, records: %d)", category, req.Symbol, primarySource, elapsed, len(history))
		return history, primarySource, nil
	}

	// Log primary failure reason
	if err != nil {
		log.Printf("[DataSourceRouter] Primary source %s failed for %s: %v", primarySource, category, err)
	} else {
		log.Printf("[DataSourceRouter] Primary source %s returned empty data for %s", primarySource, category)
	}

	// Try fallback source
	fallbackClient, fallbackSource := r.getFallbackSource(category)
	log.Printf("[DataSourceRouter] Attempting fallback source: %s", fallbackSource)

	history, err = r.fetchHistoryWithTimeout(ctx, fallbackClient, req, 10*time.Second)
	if err == nil && len(history) > 0 {
		elapsed := time.Since(startTime)
		log.Printf("[DataSourceRouter] Success: %s for %s from fallback %s (response time: %v, records: %d)", category, req.Symbol, fallbackSource, elapsed, len(history))
		return history, fallbackSource, nil
	}

	// Both sources failed
	if err != nil {
		log.Printf("[DataSourceRouter] Fallback source %s also failed for %s: %v", fallbackSource, category, err)
		return nil, "", fmt.Errorf("both sources failed: primary=%s, fallback=%s: %w", primarySource, fallbackSource, err)
	}

	log.Printf("[DataSourceRouter] Fallback source %s also returned empty data for %s", fallbackSource, category)
	return nil, "", errors.New("both sources returned empty data")
}

// fetchQuotesWithTimeout fetches quotes with a timeout
func (r *DataSourceRouter) fetchQuotesWithTimeout(ctx context.Context, client *vnstock.Client, symbols []string, timeout time.Duration) ([]vnstock.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan struct {
		quotes []vnstock.Quote
		err    error
	}, 1)

	go func() {
		quotes, err := client.RealTimeQuotes(ctx, symbols)
		resultChan <- struct {
			quotes []vnstock.Quote
			err    error
		}{quotes, err}
	}()

	select {
	case result := <-resultChan:
		return result.quotes, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("request timeout after %v", timeout)
	}
}

// fetchHistoryWithTimeout fetches historical data with a timeout
func (r *DataSourceRouter) fetchHistoryWithTimeout(ctx context.Context, client *vnstock.Client, req vnstock.QuoteHistoryRequest, timeout time.Duration) ([]vnstock.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultChan := make(chan struct {
		history []vnstock.Quote
		err     error
	}, 1)

	go func() {
		history, err := client.QuoteHistory(ctx, req)
		resultChan <- struct {
			history []vnstock.Quote
			err     error
		}{history, err}
	}()

	select {
	case result := <-resultChan:
		return result.history, result.err
	case <-ctx.Done():
		return nil, fmt.Errorf("request timeout after %v", timeout)
	}
}

// isEmptyData checks if quote data is empty or incomplete
func (r *DataSourceRouter) isEmptyData(quotes []vnstock.Quote) bool {
	if len(quotes) == 0 {
		return true
	}

	// Check if all quotes have zero or missing key fields
	allEmpty := true
	for _, q := range quotes {
		// A quote is considered non-empty if it has a valid close price
		if q.Close > 0 {
			allEmpty = false
			break
		}
	}

	return allEmpty
}

// FetchIndexHistory fetches historical index data with automatic failover
func (r *DataSourceRouter) FetchIndexHistory(ctx context.Context, req vnstock.IndexHistoryRequest) ([]vnstock.IndexRecord, string, error) {
	primaryClient, primarySource := r.selectSource(model.OHLCVHistory)
	log.Printf("[DataSourceRouter] Fetching index history for %s from %s", req.Name, primarySource)

	records, err := primaryClient.IndexHistory(ctx, req)
	if err == nil && len(records) > 0 {
		return records, primarySource, nil
	}
	if err != nil {
		log.Printf("[DataSourceRouter] Primary source %s failed for index history %s: %v", primarySource, req.Name, err)
	}

	fallbackClient, fallbackSource := r.getFallbackSource(model.OHLCVHistory)
	records, err = fallbackClient.IndexHistory(ctx, req)
	if err == nil && len(records) > 0 {
		return records, fallbackSource, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("both sources failed for index history %s: %w", req.Name, err)
	}
	return nil, "", fmt.Errorf("both sources returned empty index history for %s", req.Name)
}

// FetchListing fetches stock listing data with automatic failover
func (r *DataSourceRouter) FetchListing(ctx context.Context, exchange string) ([]vnstock.ListingRecord, string, error) {
	primaryClient, primarySource := r.selectSource(model.CompanyOverview)
	log.Printf("[DataSourceRouter] Fetching listing from %s", primarySource)

	records, err := primaryClient.Listing(ctx, exchange)
	if err == nil && len(records) > 0 {
		return records, primarySource, nil
	}
	if err != nil {
		log.Printf("[DataSourceRouter] Primary source %s failed for listing: %v", primarySource, err)
	}

	fallbackClient, fallbackSource := r.getFallbackSource(model.CompanyOverview)
	records, err = fallbackClient.Listing(ctx, exchange)
	if err == nil && len(records) > 0 {
		return records, fallbackSource, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("both sources failed for listing: %w", err)
	}
	return nil, "", errors.New("both sources returned empty listing data")
}

// FetchFinancialStatement fetches financial statement data with automatic failover
func (r *DataSourceRouter) FetchFinancialStatement(ctx context.Context, req vnstock.FinancialRequest) ([]vnstock.FinancialPeriod, string, error) {
	primaryClient, primarySource := r.selectSource(model.FinancialRatios)
	log.Printf("[DataSourceRouter] Fetching financial data for %s from %s", req.Symbol, primarySource)

	periods, err := primaryClient.FinancialStatement(ctx, req)
	if err == nil && len(periods) > 0 {
		return periods, primarySource, nil
	}
	if err != nil {
		log.Printf("[DataSourceRouter] Primary source %s failed for financial data %s: %v", primarySource, req.Symbol, err)
	}

	fallbackClient, fallbackSource := r.getFallbackSource(model.FinancialRatios)
	periods, err = fallbackClient.FinancialStatement(ctx, req)
	if err == nil && len(periods) > 0 {
		return periods, fallbackSource, nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("both sources failed for financial data %s: %w", req.Symbol, err)
	}
	return nil, "", fmt.Errorf("both sources returned empty financial data for %s", req.Symbol)
}

// GetSourcePreferences returns the current source preference mapping
func (r *DataSourceRouter) GetSourcePreferences() map[model.DataCategory]model.SourcePreference {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	prefs := make(map[model.DataCategory]model.SourcePreference)
	for k, v := range r.preferences {
		prefs[k] = v
	}
	return prefs
}
