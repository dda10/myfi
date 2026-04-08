package market

import (
	"context"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/infra"

	"github.com/dda10/vnstock-go"
)

// PriceService handles price fetching for Vietnamese stocks only.
// Uses the Data Source Router for VCI/KBS failover, batches RealTimeQuotes calls,
// caches with 15-minute TTL, and falls back to QuoteHistory (last 10 days) when
// real-time data is unavailable.
type PriceService struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
}

// NewPriceService creates a new price service.
func NewPriceService(router *infra.DataSourceRouter, cache *infra.Cache) *PriceService {
	return &PriceService{
		router: router,
		cache:  cache,
	}
}

// GetQuotes fetches real-time quotes for VN stock symbols via the Data Source Router.
// Batches all symbols into a single RealTimeQuotes call. Caches with 15-minute TTL.
// Falls back to QuoteHistory (last 10 days) if real-time fetch fails after retries.
// Returns stale cached data as last resort.
func (s *PriceService) GetQuotes(ctx context.Context, symbols []string) ([]PriceQuote, error) {
	var quotes []PriceQuote
	var uncachedSymbols []string

	// Check cache first (15-minute TTL)
	for _, symbol := range symbols {
		cacheKey := fmt.Sprintf("price:stock:%s", symbol)
		if cachedQuote, found := s.cache.Get(cacheKey); found {
			if quote, ok := cachedQuote.(PriceQuote); ok {
				quotes = append(quotes, quote)
				continue
			}
		}
		uncachedSymbols = append(uncachedSymbols, symbol)
	}

	if len(uncachedSymbols) == 0 {
		return quotes, nil
	}

	// Batch fetch uncached symbols with retry logic (Req 3.4, 3.5)
	fetchedQuotes, err := s.fetchStocksWithRetry(ctx, uncachedSymbols, 3)
	if err != nil {
		log.Printf("[PriceService] RealTimeQuotes failed after retries: %v, falling back to QuoteHistory", err)

		// Fallback: fetch from QuoteHistory using last 10 days (Req 3.2)
		fetchedQuotes, err = s.fallbackToQuoteHistory(ctx, uncachedSymbols)
		if err != nil {
			log.Printf("[PriceService] QuoteHistory fallback also failed: %v", err)
			// Last resort: return stale cached quotes
			return s.returnStaleQuotes(symbols), nil
		}
	}

	// Cache the fetched quotes with 15-minute TTL (Req 3.3)
	for _, quote := range fetchedQuotes {
		cacheKey := fmt.Sprintf("price:stock:%s", quote.Symbol)
		s.cache.Set(cacheKey, quote, 15*time.Minute)
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// fetchStocksWithRetry fetches stock quotes via RealTimeQuotes with exponential backoff retry.
// Retries up to maxRetries times (Req 3.5).
func (s *PriceService) fetchStocksWithRetry(ctx context.Context, symbols []string, maxRetries int) ([]PriceQuote, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[PriceService] Retry attempt %d/%d after %v", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}

		// Batch fetch via Data Source Router (Req 3.1, 3.4)
		vnstockQuotes, source, err := s.router.FetchRealTimeQuotes(ctx, symbols)
		if err != nil {
			lastErr = err
			log.Printf("[PriceService] Attempt %d failed: %v", attempt+1, err)
			continue
		}

		return s.convertQuotes(vnstockQuotes, source, false), nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// fallbackToQuoteHistory fetches the most recent close price from QuoteHistory
// (last 10 days) for each symbol when RealTimeQuotes is unavailable (Req 3.2).
func (s *PriceService) fallbackToQuoteHistory(ctx context.Context, symbols []string) ([]PriceQuote, error) {
	var quotes []PriceQuote
	end := time.Now()
	start := end.AddDate(0, 0, -10)

	for _, symbol := range symbols {
		req := vnstock.QuoteHistoryRequest{
			Symbol:   symbol,
			Start:    start,
			End:      end,
			Interval: "1D",
		}

		history, source, err := s.router.FetchQuoteHistory(ctx, req)
		if err != nil || len(history) == 0 {
			log.Printf("[PriceService] QuoteHistory fallback failed for %s: %v", symbol, err)
			continue
		}

		// Use the most recent bar as the price quote
		latest := history[len(history)-1]
		quote := PriceQuote{
			Symbol:    symbol,
			Price:     latest.Close,
			Volume:    latest.Volume,
			Timestamp: latest.Timestamp,
			Source:    source + " (history)",
			IsStale:   true,
		}
		quotes = append(quotes, quote)
	}

	if len(quotes) == 0 {
		return nil, fmt.Errorf("QuoteHistory fallback returned no data for any symbol")
	}

	return quotes, nil
}

// convertQuotes converts vnstock.Quote slices to PriceQuote slices.
func (s *PriceService) convertQuotes(vnstockQuotes []vnstock.Quote, source string, isStale bool) []PriceQuote {
	var quotes []PriceQuote
	for _, vq := range vnstockQuotes {
		quote := PriceQuote{
			Symbol:    vq.Symbol,
			Price:     vq.Close,
			Volume:    vq.Volume,
			Timestamp: vq.Timestamp,
			Source:    source,
			IsStale:   isStale,
		}
		quotes = append(quotes, quote)
	}
	return quotes
}

// GetHistoricalData fetches OHLCV historical data for a symbol.
// Caches with 15-minute TTL and retries with exponential backoff.
func (s *PriceService) GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, interval string) ([]OHLCVBar, error) {
	cacheKey := fmt.Sprintf("history:%s:%s:%s:%s", symbol, start.Format("2006-01-02"), end.Format("2006-01-02"), interval)
	if cachedData, found := s.cache.Get(cacheKey); found {
		if bars, ok := cachedData.([]OHLCVBar); ok {
			log.Printf("[PriceService] Cache hit for historical data: %s", symbol)
			return bars, nil
		}
	}

	bars, err := s.fetchHistoricalWithRetry(ctx, symbol, start, end, interval, 3)
	if err != nil {
		log.Printf("[PriceService] Failed to fetch historical data after retries: %v", err)
		if cachedData, found := s.cache.Get(cacheKey); found {
			if bars, ok := cachedData.([]OHLCVBar); ok {
				log.Printf("[PriceService] Returning stale cached historical data: %s", symbol)
				return bars, nil
			}
		}
		return nil, err
	}

	s.cache.Set(cacheKey, bars, 15*time.Minute)
	return bars, nil
}

// fetchHistoricalWithRetry fetches historical data with exponential backoff retry.
func (s *PriceService) fetchHistoricalWithRetry(ctx context.Context, symbol string, start, end time.Time, interval string, maxRetries int) ([]OHLCVBar, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[PriceService] Historical data retry attempt %d/%d after %v", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2
		}

		req := vnstock.QuoteHistoryRequest{
			Symbol:   symbol,
			Start:    start,
			End:      end,
			Interval: interval,
		}

		vnstockQuotes, source, err := s.router.FetchQuoteHistory(ctx, req)
		if err != nil {
			lastErr = err
			log.Printf("[PriceService] Historical data attempt %d failed: %v", attempt+1, err)
			continue
		}

		var bars []OHLCVBar
		for _, vq := range vnstockQuotes {
			bar := OHLCVBar{
				Time:   vq.Timestamp,
				Open:   vq.Open,
				High:   vq.High,
				Low:    vq.Low,
				Close:  vq.Close,
				Volume: vq.Volume,
			}
			bars = append(bars, bar)
		}

		log.Printf("[PriceService] Successfully fetched %d historical bars for %s from %s", len(bars), symbol, source)
		return bars, nil
	}

	return nil, fmt.Errorf("failed to fetch historical data after %d retries: %w", maxRetries, lastErr)
}

// returnStaleQuotes returns cached quotes with stale indicator as last resort.
func (s *PriceService) returnStaleQuotes(symbols []string) []PriceQuote {
	var quotes []PriceQuote
	for _, symbol := range symbols {
		cacheKey := fmt.Sprintf("price:stock:%s", symbol)
		if cachedQuote, found := s.cache.Get(cacheKey); found {
			if quote, ok := cachedQuote.(PriceQuote); ok {
				quote.IsStale = true
				quotes = append(quotes, quote)
			}
		}
	}
	return quotes
}
