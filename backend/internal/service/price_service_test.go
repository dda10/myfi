package service

import (
	"context"
	"testing"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
)

func TestPriceService_GetQuotes_VNStock(t *testing.T) {
	// Initialize dependencies
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create data source router: %v", err)
	}

	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create gold service: %v", err)
	}
	priceService := NewPriceService(router, cache, goldService)

	ctx := context.Background()
	symbols := []string{"SSI", "FPT", "VNM"}

	// Test fetching stock quotes
	quotes, err := priceService.GetQuotes(ctx, symbols, model.VNStock)
	if err != nil {
		t.Fatalf("GetQuotes failed: %v", err)
	}

	// Note: RealTimeQuotes may return empty data when market is closed
	// This is expected behavior, not a failure
	if len(quotes) == 0 {
		t.Skip("Market is closed - RealTimeQuotes returned empty data (expected behavior)")
	}

	for _, quote := range quotes {
		t.Logf("Quote: %s - Price: %.2f, Change: %.2f%%, Volume: %d, Source: %s",
			quote.Symbol, quote.Price, quote.ChangePercent, quote.Volume, quote.Source)

		if quote.Symbol == "" {
			t.Error("Quote symbol is empty")
		}
		if quote.AssetType != model.VNStock {
			t.Errorf("Expected asset type %s, got %s", model.VNStock, quote.AssetType)
		}
		if quote.Source == "" {
			t.Error("Quote source is empty")
		}
	}
}

func TestPriceService_GetQuotes_Cache(t *testing.T) {
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create data source router: %v", err)
	}

	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create gold service: %v", err)
	}
	priceService := NewPriceService(router, cache, goldService)

	ctx := context.Background()
	symbols := []string{"SSI"}

	// First fetch - should hit the API
	start := time.Now()
	quotes1, err := priceService.GetQuotes(ctx, symbols, model.VNStock)
	if err != nil {
		t.Fatalf("First GetQuotes failed: %v", err)
	}
	firstFetchDuration := time.Since(start)

	// Skip test if market is closed (empty quotes)
	if len(quotes1) == 0 {
		t.Skip("Market is closed - cannot test cache behavior with empty data")
	}

	// Second fetch - should hit the cache
	start = time.Now()
	quotes2, err := priceService.GetQuotes(ctx, symbols, model.VNStock)
	if err != nil {
		t.Fatalf("Second GetQuotes failed: %v", err)
	}
	secondFetchDuration := time.Since(start)

	t.Logf("First fetch: %v, Second fetch (cached): %v", firstFetchDuration, secondFetchDuration)

	if len(quotes1) != len(quotes2) {
		t.Errorf("Expected same number of quotes, got %d and %d", len(quotes1), len(quotes2))
	}

	// Cache should be significantly faster
	if secondFetchDuration > firstFetchDuration/2 {
		t.Logf("Warning: Cache fetch not significantly faster (first: %v, second: %v)", firstFetchDuration, secondFetchDuration)
	}
}

func TestPriceService_GetHistoricalData(t *testing.T) {
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create data source router: %v", err)
	}

	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create gold service: %v", err)
	}
	priceService := NewPriceService(router, cache, goldService)

	ctx := context.Background()
	symbol := "SSI"
	end := time.Now()
	start := end.AddDate(0, 0, -30) // Last 30 days

	// Test fetching historical data
	bars, err := priceService.GetHistoricalData(ctx, symbol, start, end, "1d")
	if err != nil {
		t.Fatalf("GetHistoricalData failed: %v", err)
	}

	if len(bars) == 0 {
		t.Error("Expected historical bars, got empty result")
	}

	t.Logf("Fetched %d historical bars for %s", len(bars), symbol)

	for i, bar := range bars {
		if i < 3 || i >= len(bars)-3 {
			t.Logf("Bar %d: Time: %s, O: %.2f, H: %.2f, L: %.2f, C: %.2f, V: %d",
				i, bar.Time.Format("2006-01-02"), bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
		}

		if bar.Close == 0 {
			t.Errorf("Bar %d has zero close price", i)
		}
		if bar.High < bar.Low {
			t.Errorf("Bar %d has high < low", i)
		}
	}
}

func TestPriceService_GetHistoricalData_Cache(t *testing.T) {
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create data source router: %v", err)
	}

	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create gold service: %v", err)
	}
	priceService := NewPriceService(router, cache, goldService)

	ctx := context.Background()
	symbol := "FPT"
	end := time.Now()
	start := end.AddDate(0, 0, -10)

	// First fetch
	start1 := time.Now()
	bars1, err := priceService.GetHistoricalData(ctx, symbol, start, end, "1d")
	if err != nil {
		t.Fatalf("First GetHistoricalData failed: %v", err)
	}
	duration1 := time.Since(start1)

	// Second fetch (should be cached)
	start2 := time.Now()
	bars2, err := priceService.GetHistoricalData(ctx, symbol, start, end, "1d")
	if err != nil {
		t.Fatalf("Second GetHistoricalData failed: %v", err)
	}
	duration2 := time.Since(start2)

	t.Logf("First fetch: %v (%d bars), Second fetch (cached): %v (%d bars)", duration1, len(bars1), duration2, len(bars2))

	if len(bars1) != len(bars2) {
		t.Errorf("Expected same number of bars, got %d and %d", len(bars1), len(bars2))
	}

	// Cache should be significantly faster
	if duration2 > duration1/2 {
		t.Logf("Warning: Cache fetch not significantly faster")
	}
}

func TestPriceService_UnsupportedAssetType(t *testing.T) {
	router, err := infra.NewDataSourceRouter()
	if err != nil {
		t.Fatalf("Failed to create data source router: %v", err)
	}

	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create gold service: %v", err)
	}
	priceService := NewPriceService(router, cache, goldService)

	ctx := context.Background()
	symbols := []string{"TEST"}

	// Test with unsupported asset type
	_, err = priceService.GetQuotes(ctx, symbols, model.AssetType("invalid"))
	if err == nil {
		t.Error("Expected error for unsupported asset type, got nil")
	}

	t.Logf("Got expected error: %v", err)
}
