package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// CryptoService handles cryptocurrency price fetching from CoinGecko
type CryptoService struct {
	cache          *Cache
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
	httpClient     *http.Client
}

// CryptoPriceResponse represents a cryptocurrency price with metadata
type CryptoPriceResponse struct {
	Symbol           string    `json:"symbol"`
	Name             string    `json:"name"`
	PriceVND         float64   `json:"price_vnd"`
	PriceUSD         float64   `json:"price_usd"`
	Change24h        float64   `json:"change_24h"`
	PercentChange24h float64   `json:"percent_change_24h"`
	Volume24h        float64   `json:"volume_24h"`
	MarketCapVND     float64   `json:"market_cap_vnd"`
	Source           string    `json:"source"`
	Timestamp        time.Time `json:"timestamp"`
}

// CoinGeckoResponse represents the API response from CoinGecko
type CoinGeckoResponse struct {
	ID                       string  `json:"id"`
	Symbol                   string  `json:"symbol"`
	Name                     string  `json:"name"`
	CurrentPrice             float64 `json:"current_price"`
	MarketCap                float64 `json:"market_cap"`
	TotalVolume              float64 `json:"total_volume"`
	PriceChange24h           float64 `json:"price_change_24h"`
	PriceChangePercentage24h float64 `json:"price_change_percentage_24h"`
}

// NewCryptoService creates a new crypto service instance
func NewCryptoService(cache *Cache, rateLimiter *RateLimiter, circuitBreaker *CircuitBreaker) *CryptoService {
	return &CryptoService{
		cache:          cache,
		rateLimiter:    rateLimiter,
		circuitBreaker: circuitBreaker,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetCryptoPrices fetches cryptocurrency prices from CoinGecko with VND conversion
func (s *CryptoService) GetCryptoPrices(symbols []string) ([]CryptoPriceResponse, error) {
	cacheKey := fmt.Sprintf("crypto_prices_%s", strings.Join(symbols, ","))

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if prices, ok := cached.([]CryptoPriceResponse); ok {
			log.Println("[CryptoService] Cache hit for crypto prices")
			return prices, nil
		}
	}

	// Check rate limiter
	if err := s.rateLimiter.Allow("CoinGecko"); err != nil {
		log.Printf("[CryptoService] Rate limit exceeded: %v", err)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Use circuit breaker to fetch from CoinGecko
	var prices []CryptoPriceResponse
	err := s.circuitBreaker.Call(func() error {
		var fetchErr error
		prices, fetchErr = s.fetchFromCoinGecko(symbols)
		return fetchErr
	})

	if err != nil {
		log.Printf("[CryptoService] Failed to fetch from CoinGecko: %v", err)

		// Try to return cached data even if stale
		if cached, found := s.cache.Get(cacheKey); found {
			if cachedPrices, ok := cached.([]CryptoPriceResponse); ok {
				log.Println("[CryptoService] Returning stale cached data due to fetch failure")
				return cachedPrices, nil
			}
		}
		return nil, err
	}

	// Cache the result with 5-minute TTL
	s.cache.Set(cacheKey, prices, 5*time.Minute)

	log.Printf("[CryptoService] Successfully fetched %d crypto prices", len(prices))
	return prices, nil
}

// fetchFromCoinGecko fetches prices from CoinGecko API
func (s *CryptoService) fetchFromCoinGecko(symbols []string) ([]CryptoPriceResponse, error) {
	// Map common symbols to CoinGecko IDs
	coinIDs := s.mapSymbolsToCoinGeckoIDs(symbols)
	if len(coinIDs) == 0 {
		return nil, fmt.Errorf("no valid coin IDs found for symbols: %v", symbols)
	}

	// Build API URL - fetch prices in both USD and VND
	idsParam := strings.Join(coinIDs, ",")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=100&page=1&sparkline=false&price_change_percentage=24h", idsParam)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from CoinGecko: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CoinGecko API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var coinGeckoData []CoinGeckoResponse
	if err := json.Unmarshal(body, &coinGeckoData); err != nil {
		return nil, fmt.Errorf("failed to parse CoinGecko response: %w", err)
	}

	if len(coinGeckoData) == 0 {
		return nil, fmt.Errorf("no data returned from CoinGecko for symbols: %v", symbols)
	}

	// Get USD/VND exchange rate (hardcoded fallback for now, will use FX_Service later)
	usdToVND := 25400.0

	// Convert to our response format
	prices := make([]CryptoPriceResponse, 0, len(coinGeckoData))
	for _, coin := range coinGeckoData {
		price := CryptoPriceResponse{
			Symbol:           strings.ToUpper(coin.Symbol),
			Name:             coin.Name,
			PriceUSD:         coin.CurrentPrice,
			PriceVND:         coin.CurrentPrice * usdToVND,
			Change24h:        coin.PriceChange24h,
			PercentChange24h: coin.PriceChangePercentage24h,
			Volume24h:        coin.TotalVolume,
			MarketCapVND:     coin.MarketCap * usdToVND,
			Source:           "CoinGecko",
			Timestamp:        time.Now(),
		}
		prices = append(prices, price)
	}

	return prices, nil
}

// mapSymbolsToCoinGeckoIDs maps common crypto symbols to CoinGecko coin IDs
func (s *CryptoService) mapSymbolsToCoinGeckoIDs(symbols []string) []string {
	// Common mappings
	symbolToID := map[string]string{
		"BTC":   "bitcoin",
		"ETH":   "ethereum",
		"USDT":  "tether",
		"BNB":   "binancecoin",
		"SOL":   "solana",
		"XRP":   "ripple",
		"ADA":   "cardano",
		"DOGE":  "dogecoin",
		"DOT":   "polkadot",
		"MATIC": "matic-network",
		"SHIB":  "shiba-inu",
		"AVAX":  "avalanche-2",
		"LINK":  "chainlink",
		"UNI":   "uniswap",
		"ATOM":  "cosmos",
		"LTC":   "litecoin",
		"BCH":   "bitcoin-cash",
		"XLM":   "stellar",
		"ALGO":  "algorand",
		"VET":   "vechain",
	}

	coinIDs := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		upperSymbol := strings.ToUpper(strings.TrimSpace(symbol))
		if coinID, found := symbolToID[upperSymbol]; found {
			coinIDs = append(coinIDs, coinID)
		} else {
			// Try lowercase symbol as ID (works for many coins)
			coinIDs = append(coinIDs, strings.ToLower(upperSymbol))
		}
	}

	return coinIDs
}

// GetCryptoPriceBySymbol fetches a single crypto price by symbol
func (s *CryptoService) GetCryptoPriceBySymbol(symbol string) (*CryptoPriceResponse, error) {
	prices, err := s.GetCryptoPrices([]string{symbol})
	if err != nil {
		return nil, err
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no price found for symbol: %s", symbol)
	}

	return &prices[0], nil
}
