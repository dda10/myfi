package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dda10/vnstock-go"
)

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
)

// SourcePreference defines the primary and fallback sources for a data category
type SourcePreference struct {
	Category DataCategory
	Primary  string // "VCI" or "KBS"
	Fallback string
}

// DataSourceRouter manages intelligent routing between VCI and KBS data sources
type DataSourceRouter struct {
	preferences     map[DataCategory]SourcePreference
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

	// Note: KBS connector is not yet available in vnstock-go library
	// Using VCI as both primary and fallback for now
	// TODO: Add KBS client when connector becomes available
	var kbsClient *vnstock.Client = nil

	// Define source preferences based on data richness
	// VCI provides 77 columns
	// Note: KBS fallback will be added when connector becomes available
	preferences := map[DataCategory]SourcePreference{
		PriceQuotes: {
			Category: PriceQuotes,
			Primary:  "VCI",
			Fallback: "VCI", // Using VCI as fallback until KBS is available
		},
		OHLCVHistory: {
			Category: OHLCVHistory,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		IntradayData: {
			Category: IntradayData,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		OrderBook: {
			Category: OrderBook,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		CompanyOverview: {
			Category: CompanyOverview,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		Shareholders: {
			Category: Shareholders,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		Officers: {
			Category: Officers,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		News: {
			Category: News,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		IncomeStatement: {
			Category: IncomeStatement,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		BalanceSheet: {
			Category: BalanceSheet,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		CashFlow: {
			Category: CashFlow,
			Primary:  "VCI",
			Fallback: "VCI",
		},
		FinancialRatios: {
			Category: FinancialRatios,
			Primary:  "VCI",
			Fallback: "VCI",
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

// selectSource returns the appropriate client based on category preference
func (r *DataSourceRouter) selectSource(category DataCategory) (*vnstock.Client, string) {
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
func (r *DataSourceRouter) getFallbackSource(category DataCategory) (*vnstock.Client, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	pref, exists := r.preferences[category]
	if !exists {
		// Default to VCI as fallback (KBS not yet available)
		log.Printf("[DataSourceRouter] Category %s not found in preferences, defaulting to VCI fallback", category)
		return r.vciClient, "VCI"
	}

	// For now, always return VCI since KBS is not available
	// TODO: Update when KBS connector is added
	if pref.Fallback == "VCI" || r.kbsClient == nil {
		return r.vciClient, "VCI"
	}
	return r.kbsClient, "KBS"
}

// FetchRealTimeQuotes fetches real-time quotes with automatic failover
func (r *DataSourceRouter) FetchRealTimeQuotes(ctx context.Context, symbols []string) ([]vnstock.Quote, string, error) {
	category := PriceQuotes
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
	category := OHLCVHistory
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

// GetSourcePreferences returns the current source preference mapping
func (r *DataSourceRouter) GetSourcePreferences() map[DataCategory]SourcePreference {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	prefs := make(map[DataCategory]SourcePreference)
	for k, v := range r.preferences {
		prefs[k] = v
	}
	return prefs
}
