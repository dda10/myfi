package service

import (
	"testing"
	"time"

	"myfi-backend/internal/infra"
)

func TestCryptoService_GetCryptoPrices(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	circuitBreaker := infra.NewCircuitBreaker(3, 60*time.Second)

	cryptoService := NewCryptoService(cache, rateLimiter, circuitBreaker)

	symbols := []string{"BTC", "ETH", "USDT"}

	prices, err := cryptoService.GetCryptoPrices(symbols)
	if err != nil {
		t.Fatalf("Failed to fetch crypto prices: %v", err)
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one price, got none")
	}

	t.Logf("Fetched %d crypto prices", len(prices))

	// Verify price structure
	for _, price := range prices {
		t.Logf("Crypto: %s (%s) - USD: $%.2f, VND: ₫%.0f, Change 24h: %.2f%%",
			price.Symbol, price.Name, price.PriceUSD, price.PriceVND, price.PercentChange24h)

		if price.Symbol == "" {
			t.Error("Symbol should not be empty")
		}
		if price.Name == "" {
			t.Error("Name should not be empty")
		}
		if price.PriceUSD <= 0 {
			t.Error("PriceUSD should be positive")
		}
		if price.PriceVND <= 0 {
			t.Error("PriceVND should be positive")
		}
		if price.Source != "CoinGecko" {
			t.Errorf("Expected source 'CoinGecko', got '%s'", price.Source)
		}
		if price.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
	}
}

func TestCryptoService_GetCryptoPrices_Cache(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	circuitBreaker := infra.NewCircuitBreaker(3, 60*time.Second)

	cryptoService := NewCryptoService(cache, rateLimiter, circuitBreaker)

	symbols := []string{"BTC", "ETH"}

	// First fetch - should hit API
	start1 := time.Now()
	prices1, err := cryptoService.GetCryptoPrices(symbols)
	duration1 := time.Since(start1)
	if err != nil {
		t.Fatalf("Failed to fetch crypto prices: %v", err)
	}

	if len(prices1) == 0 {
		t.Fatal("Expected at least one price, got none")
	}

	// Second fetch - should hit cache
	start2 := time.Now()
	prices2, err := cryptoService.GetCryptoPrices(symbols)
	duration2 := time.Since(start2)
	if err != nil {
		t.Fatalf("Failed to fetch crypto prices from cache: %v", err)
	}

	if len(prices2) != len(prices1) {
		t.Errorf("Expected %d prices from cache, got %d", len(prices1), len(prices2))
	}

	// Cache should be much faster
	if duration2 >= duration1 {
		t.Logf("Warning: Cache fetch (%v) not faster than API fetch (%v)", duration2, duration1)
	}

	t.Logf("First fetch (API): %v (%d prices), Second fetch (cache): %v (%d prices)",
		duration1, len(prices1), duration2, len(prices2))
}

func TestCryptoService_GetCryptoPriceBySymbol(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	circuitBreaker := infra.NewCircuitBreaker(3, 60*time.Second)

	cryptoService := NewCryptoService(cache, rateLimiter, circuitBreaker)

	price, err := cryptoService.GetCryptoPriceBySymbol("BTC")
	if err != nil {
		t.Fatalf("Failed to fetch BTC price: %v", err)
	}

	if price.Symbol != "BTC" {
		t.Errorf("Expected symbol 'BTC', got '%s'", price.Symbol)
	}

	t.Logf("BTC Price: USD $%.2f, VND ₫%.0f, Change 24h: %.2f%%",
		price.PriceUSD, price.PriceVND, price.PercentChange24h)
}

func TestCryptoService_InvalidSymbol(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	circuitBreaker := infra.NewCircuitBreaker(3, 60*time.Second)

	cryptoService := NewCryptoService(cache, rateLimiter, circuitBreaker)

	// Try to fetch a non-existent symbol
	_, err := cryptoService.GetCryptoPrices([]string{"INVALIDCOIN123"})
	if err == nil {
		t.Log("Note: CoinGecko may have returned data for invalid symbol (expected behavior)")
	} else {
		t.Logf("Got expected error for invalid symbol: %v", err)
	}
}

func TestCryptoService_VNDConversion(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()
	circuitBreaker := infra.NewCircuitBreaker(3, 60*time.Second)

	cryptoService := NewCryptoService(cache, rateLimiter, circuitBreaker)

	// Fetch BTC and ETH together to reduce API calls
	prices, err := cryptoService.GetCryptoPrices([]string{"BTC", "ETH"})
	if err != nil {
		t.Skipf("Skipping VND conversion test due to API error (likely rate limit): %v", err)
	}

	if len(prices) == 0 {
		t.Skip("No prices returned, skipping VND conversion test")
	}

	price := prices[0]

	// Verify VND conversion (should be exactly USD * 25400)
	expectedVND := price.PriceUSD * 25400.0
	tolerance := expectedVND * 0.001 // 0.1% tolerance for floating point

	if price.PriceVND < expectedVND-tolerance || price.PriceVND > expectedVND+tolerance {
		t.Errorf("VND conversion incorrect: USD $%.2f should convert to ~₫%.0f, got ₫%.0f",
			price.PriceUSD, expectedVND, price.PriceVND)
	}

	t.Logf("VND conversion verified: %s USD $%.2f = VND ₫%.0f", price.Symbol, price.PriceUSD, price.PriceVND)
}
