package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dda10/vnstock-go"
)

// AssetType represents the type of asset
type AssetType string

const (
	VNStock AssetType = "vn_stock"
	Crypto  AssetType = "crypto"
	Gold    AssetType = "gold"
	Savings AssetType = "savings"
	Bond    AssetType = "bond"
	Cash    AssetType = "cash"
)

// ValidAssetTypes contains all supported asset types for validation
var ValidAssetTypes = map[AssetType]bool{
	VNStock: true,
	Crypto:  true,
	Gold:    true,
	Savings: true,
	Bond:    true,
	Cash:    true,
}

// PriceQuote represents a price quote for an asset
type PriceQuote struct {
	Symbol        string    `json:"symbol"`
	AssetType     AssetType `json:"assetType"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"changePercent"`
	Volume        int64     `json:"volume"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"`
	IsStale       bool      `json:"isStale"`
}

// OHLCVBar represents a single OHLCV data point
type OHLCVBar struct {
	Time   time.Time `json:"time"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// PriceService handles price fetching for all asset types
type PriceService struct {
	router      *DataSourceRouter
	cache       *Cache
	goldService *GoldService
}

// NewPriceService creates a new price service
func NewPriceService(router *DataSourceRouter, cache *Cache, goldService *GoldService) *PriceService {
	return &PriceService{
		router:      router,
		cache:       cache,
		goldService: goldService,
	}
}

// GetQuotes fetches quotes for multiple symbols with retry logic and caching
func (s *PriceService) GetQuotes(ctx context.Context, symbols []string, assetType AssetType) ([]PriceQuote, error) {
	switch assetType {
	case VNStock:
		return s.batchFetchStocks(ctx, symbols)
	case Crypto:
		return s.fetchCrypto(ctx, symbols)
	case Gold:
		return s.fetchGold(ctx, symbols)
	default:
		return nil, fmt.Errorf("unsupported asset type: %s", assetType)
	}
}

// batchFetchStocks fetches stock quotes with caching and retry logic
func (s *PriceService) batchFetchStocks(ctx context.Context, symbols []string) ([]PriceQuote, error) {
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

	// If all symbols were cached, return immediately
	if len(uncachedSymbols) == 0 {
		return quotes, nil
	}

	// Fetch uncached symbols with retry logic
	fetchedQuotes, err := s.fetchStocksWithRetry(ctx, uncachedSymbols, 3)
	if err != nil {
		log.Printf("[PriceService] Failed to fetch stocks after retries: %v", err)
		// Return cached quotes with stale indicator
		return s.returnStaleQuotes(symbols), nil
	}

	// Cache the fetched quotes
	for _, quote := range fetchedQuotes {
		cacheKey := fmt.Sprintf("price:stock:%s", quote.Symbol)
		s.cache.Set(cacheKey, quote, 15*time.Minute)
		quotes = append(quotes, quote)
	}

	return quotes, nil
}

// fetchStocksWithRetry fetches stock quotes with exponential backoff retry
func (s *PriceService) fetchStocksWithRetry(ctx context.Context, symbols []string, maxRetries int) ([]PriceQuote, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[PriceService] Retry attempt %d/%d after %v", attempt+1, maxRetries, backoff)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}

		// Fetch via Data_Source_Router
		vnstockQuotes, source, err := s.router.FetchRealTimeQuotes(ctx, symbols)
		if err != nil {
			lastErr = err
			log.Printf("[PriceService] Attempt %d failed: %v", attempt+1, err)
			continue
		}

		// Convert vnstock.Quote to PriceQuote
		var quotes []PriceQuote
		for _, vq := range vnstockQuotes {
			// Calculate change and percent change (will be 0 if we don't have previous data)
			// In a real implementation, we would fetch previous close from historical data
			quote := PriceQuote{
				Symbol:        vq.Symbol,
				AssetType:     VNStock,
				Price:         vq.Close,
				Change:        0, // TODO: Calculate from previous close
				ChangePercent: 0, // TODO: Calculate from previous close
				Volume:        vq.Volume,
				Timestamp:     vq.Timestamp,
				Source:        source,
				IsStale:       false,
			}
			quotes = append(quotes, quote)
		}

		return quotes, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// GetHistoricalData fetches OHLCV historical data for a symbol
func (s *PriceService) GetHistoricalData(ctx context.Context, symbol string, start, end time.Time, interval string) ([]OHLCVBar, error) {
	// Check cache first (15-minute TTL)
	cacheKey := fmt.Sprintf("history:%s:%s:%s:%s", symbol, start.Format("2006-01-02"), end.Format("2006-01-02"), interval)
	if cachedData, found := s.cache.Get(cacheKey); found {
		if bars, ok := cachedData.([]OHLCVBar); ok {
			log.Printf("[PriceService] Cache hit for historical data: %s", symbol)
			return bars, nil
		}
	}

	// Fetch with retry logic
	bars, err := s.fetchHistoricalWithRetry(ctx, symbol, start, end, interval, 3)
	if err != nil {
		log.Printf("[PriceService] Failed to fetch historical data after retries: %v", err)
		// Try to return stale cached data
		if cachedData, found := s.cache.Get(cacheKey); found {
			if bars, ok := cachedData.([]OHLCVBar); ok {
				log.Printf("[PriceService] Returning stale cached historical data: %s", symbol)
				return bars, nil
			}
		}
		return nil, err
	}

	// Cache the result
	s.cache.Set(cacheKey, bars, 15*time.Minute)
	return bars, nil
}

// fetchHistoricalWithRetry fetches historical data with exponential backoff retry
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

		// Convert to OHLCVBar
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

// fetchCrypto fetches cryptocurrency prices (placeholder for future implementation)
func (s *PriceService) fetchCrypto(ctx context.Context, symbols []string) ([]PriceQuote, error) {
	// TODO: Implement CoinGecko integration with 5-minute cache TTL
	return nil, fmt.Errorf("crypto price fetching not yet implemented")
}

// fetchGold fetches gold prices via GoldService
func (s *PriceService) fetchGold(ctx context.Context, symbols []string) ([]PriceQuote, error) {
	// Fetch gold prices via GoldService (handles caching internally)
	goldPrices, err := s.goldService.GetGoldPrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gold prices: %w", err)
	}

	// Convert GoldPriceResponse to PriceQuote
	var quotes []PriceQuote
	for _, gp := range goldPrices {
		// Use sell price as the primary price (what you pay to buy gold)
		quote := PriceQuote{
			Symbol:        gp.TypeName,
			AssetType:     Gold,
			Price:         gp.SellPrice,
			Change:        0, // TODO: Calculate from previous day
			ChangePercent: 0, // TODO: Calculate from previous day
			Volume:        0, // Gold doesn't have volume
			Timestamp:     gp.Date,
			Source:        gp.Source,
			IsStale:       false,
		}
		quotes = append(quotes, quote)
	}

	if len(quotes) > 0 {
		log.Printf("[PriceService] Fetched %d gold prices from %s", len(quotes), quotes[0].Source)
	}
	return quotes, nil
}

// returnStaleQuotes returns cached quotes with stale indicator
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
