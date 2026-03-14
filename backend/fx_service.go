package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	// FallbackUSDVND is the hardcoded fallback rate when CoinGecko is unavailable.
	FallbackUSDVND = 25400.0

	// fxCacheTTL is the cache duration for the USD/VND rate.
	fxCacheTTL = 1 * time.Hour

	// fxCacheKey is the cache key for the USD/VND exchange rate.
	fxCacheKey = "fx:usd_vnd"
)

// FXRate represents a USD/VND exchange rate with metadata.
type FXRate struct {
	Rate      float64   `json:"rate"`
	Source    string    `json:"source"`
	IsStale   bool      `json:"is_stale"`
	Timestamp time.Time `json:"timestamp"`
}

// FXService fetches and caches the USD/VND exchange rate.
type FXService struct {
	cache          *Cache
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
	httpClient     *http.Client
}

// NewFXService creates a new FXService instance.
func NewFXService(cache *Cache, rateLimiter *RateLimiter, circuitBreaker *CircuitBreaker) *FXService {
	return &FXService{
		cache:          cache,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetUSDVNDRate returns the current USD/VND exchange rate.
// It checks cache first, then fetches from CoinGecko (USDT/VND pair),
// and falls back to the hardcoded rate of 25,400 if the API is unavailable.
func (s *FXService) GetUSDVNDRate() (*FXRate, error) {
	// Check cache first
	if cached, found := s.cache.Get(fxCacheKey); found {
		if rate, ok := cached.(*FXRate); ok {
			log.Println("[FXService] Cache hit for USD/VND rate")
			return rate, nil
		}
	}

	// Check rate limiter
	if err := s.rateLimiter.Allow("CoinGecko"); err != nil {
		log.Printf("[FXService] Rate limit exceeded, using fallback: %v", err)
		return s.fallbackRate(), nil
	}

	// Fetch from CoinGecko with circuit breaker
	var rate *FXRate
	err := s.circuitBreaker.Call(func() error {
		var fetchErr error
		rate, fetchErr = s.fetchFromCoinGecko()
		return fetchErr
	})

	if err != nil {
		log.Printf("[FXService] CoinGecko fetch failed, using fallback: %v", err)

		// Return stale cached data if available
		if cached, found := s.cache.Get(fxCacheKey); found {
			if stale, ok := cached.(*FXRate); ok {
				stale.IsStale = true
				return stale, nil
			}
		}

		return s.fallbackRate(), nil
	}

	// Cache with 1-hour TTL
	s.cache.Set(fxCacheKey, rate, fxCacheTTL)
	log.Printf("[FXService] USD/VND rate fetched: %.0f (source: %s)", rate.Rate, rate.Source)
	return rate, nil
}

// fetchFromCoinGecko fetches the USDT/VND rate from CoinGecko simple price API.
func (s *FXService) fetchFromCoinGecko() (*FXRate, error) {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=vnd"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CoinGecko request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Response format: {"tether":{"vnd":25400}}
	var data map[string]map[string]float64
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	tether, ok := data["tether"]
	if !ok {
		return nil, fmt.Errorf("missing 'tether' key in CoinGecko response")
	}

	vndRate, ok := tether["vnd"]
	if !ok || vndRate <= 0 {
		return nil, fmt.Errorf("missing or invalid 'vnd' rate in CoinGecko response")
	}

	return &FXRate{
		Rate:      vndRate,
		Source:    "CoinGecko",
		IsStale:   false,
		Timestamp: time.Now(),
	}, nil
}

// fallbackRate returns the hardcoded fallback USD/VND rate.
func (s *FXService) fallbackRate() *FXRate {
	return &FXRate{
		Rate:      FallbackUSDVND,
		Source:    "fallback",
		IsStale:   true,
		Timestamp: time.Now(),
	}
}

// ConvertVNDToUSD converts a VND amount to USD using the current rate.
func (s *FXService) ConvertVNDToUSD(amountVND float64) (float64, error) {
	rate, err := s.GetUSDVNDRate()
	if err != nil {
		return 0, err
	}
	if rate.Rate <= 0 {
		return 0, fmt.Errorf("invalid FX rate: %.2f", rate.Rate)
	}
	return amountVND / rate.Rate, nil
}

// ConvertUSDToVND converts a USD amount to VND using the current rate.
func (s *FXService) ConvertUSDToVND(amountUSD float64) (float64, error) {
	rate, err := s.GetUSDVNDRate()
	if err != nil {
		return 0, err
	}
	return amountUSD * rate.Rate, nil
}
