package main

import (
	"testing"
	"time"
)

func TestFXService_GetUSDVNDRate(t *testing.T) {
	cache := NewCache()
	rateLimiter := NewRateLimiter()
	circuitBreaker := NewCircuitBreaker(3, 60*time.Second)

	fxService := NewFXService(cache, rateLimiter, circuitBreaker)

	rate, err := fxService.GetUSDVNDRate()
	if err != nil {
		t.Fatalf("Failed to get USD/VND rate: %v", err)
	}

	if rate.Rate <= 0 {
		t.Errorf("Expected positive rate, got %.2f", rate.Rate)
	}

	// USD/VND should be in a reasonable range (20,000 - 30,000)
	if rate.Rate < 20000 || rate.Rate > 30000 {
		t.Errorf("Rate %.0f outside reasonable range [20000, 30000]", rate.Rate)
	}

	if rate.Source == "" {
		t.Error("Source should not be empty")
	}

	if rate.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	t.Logf("USD/VND rate: %.0f (source: %s, stale: %v)", rate.Rate, rate.Source, rate.IsStale)
}

func TestFXService_Cache(t *testing.T) {
	cache := NewCache()
	rateLimiter := NewRateLimiter()
	circuitBreaker := NewCircuitBreaker(3, 60*time.Second)

	fxService := NewFXService(cache, rateLimiter, circuitBreaker)

	// First fetch
	rate1, err := fxService.GetUSDVNDRate()
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}

	// Second fetch should hit cache
	start := time.Now()
	rate2, err := fxService.GetUSDVNDRate()
	cacheDuration := time.Since(start)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}

	if rate1.Rate != rate2.Rate {
		t.Errorf("Cached rate (%.0f) differs from original (%.0f)", rate2.Rate, rate1.Rate)
	}

	// Cache lookup should be sub-millisecond
	if cacheDuration > 10*time.Millisecond {
		t.Logf("Warning: cache fetch took %v (expected < 10ms)", cacheDuration)
	}

	t.Logf("First: %.0f (%s), Cached: %.0f (cache fetch: %v)", rate1.Rate, rate1.Source, rate2.Rate, cacheDuration)
}

func TestFXService_FallbackRate(t *testing.T) {
	cache := NewCache()
	rateLimiter := NewRateLimiter()
	circuitBreaker := NewCircuitBreaker(3, 60*time.Second)

	fxService := NewFXService(cache, rateLimiter, circuitBreaker)

	fallback := fxService.fallbackRate()

	if fallback.Rate != FallbackUSDVND {
		t.Errorf("Expected fallback rate %.0f, got %.0f", FallbackUSDVND, fallback.Rate)
	}

	if fallback.Source != "fallback" {
		t.Errorf("Expected source 'fallback', got '%s'", fallback.Source)
	}

	if !fallback.IsStale {
		t.Error("Fallback rate should be marked as stale")
	}

	t.Logf("Fallback rate: %.0f VND/USD", fallback.Rate)
}

func TestFXService_ConvertVNDToUSD(t *testing.T) {
	cache := NewCache()
	rateLimiter := NewRateLimiter()
	circuitBreaker := NewCircuitBreaker(3, 60*time.Second)

	fxService := NewFXService(cache, rateLimiter, circuitBreaker)

	amountVND := 25400000.0 // 25.4 million VND
	usd, err := fxService.ConvertVNDToUSD(amountVND)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Should be roughly 1000 USD (depending on actual rate)
	if usd <= 0 {
		t.Errorf("Expected positive USD amount, got %.2f", usd)
	}

	t.Logf("%.0f VND = %.2f USD", amountVND, usd)
}

func TestFXService_ConvertUSDToVND(t *testing.T) {
	cache := NewCache()
	rateLimiter := NewRateLimiter()
	circuitBreaker := NewCircuitBreaker(3, 60*time.Second)

	fxService := NewFXService(cache, rateLimiter, circuitBreaker)

	amountUSD := 1000.0
	vnd, err := fxService.ConvertUSDToVND(amountUSD)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	// Should be in the millions of VND
	if vnd < 20000000 || vnd > 30000000 {
		t.Errorf("Converted amount %.0f VND outside reasonable range for $1000", vnd)
	}

	t.Logf("$%.0f USD = %.0f VND", amountUSD, vnd)
}
