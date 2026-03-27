package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
	_ "github.com/dda10/vnstock-go/all"
)

// GoldService handles gold price data retrieval with caching and resilience.
type GoldService struct {
	client         *vnstock.Client
	cache          *infra.Cache
	rateLimiter    *infra.RateLimiter
	circuitBreaker *infra.CircuitBreaker
}

// NewGoldService creates a new GoldService instance.
func NewGoldService(cache *infra.Cache, rateLimiter *infra.RateLimiter) (*GoldService, error) {
	client, err := vnstock.New(vnstock.Config{
		Connector:  "GOLD",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create vnstock GOLD client: %w", err)
	}

	return &GoldService{
		client:         client,
		cache:          cache,
		rateLimiter:    rateLimiter,
		circuitBreaker: infra.NewCircuitBreaker(3, 60*time.Second),
	}, nil
}

// GetGoldPrices retrieves gold prices from all sources (SJC and BTMC).
func (s *GoldService) GetGoldPrices(ctx context.Context) ([]model.GoldPriceResponse, error) {
	cacheKey := "gold:prices:all"

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if prices, ok := cached.([]model.GoldPriceResponse); ok {
			log.Printf("[GoldService] Cache hit for gold prices")
			return prices, nil
		}
	}

	// Rate limiting
	if err := s.rateLimiter.Allow("GOLD"); err != nil {
		return nil, fmt.Errorf("rate limit exceeded for gold prices: %w", err)
	}

	// Fetch with circuit breaker + retry
	var prices []model.GoldPriceResponse
	err := s.circuitBreaker.Call(func() error {
		vnPrices, fetchErr := s.client.GoldPrice(ctx, vnstock.GoldPriceRequest{
			Date: time.Now(),
		})
		if fetchErr != nil {
			return fetchErr
		}
		if len(vnPrices) == 0 {
			return fmt.Errorf("no gold price data returned")
		}

		prices = make([]model.GoldPriceResponse, len(vnPrices))
		for i, p := range vnPrices {
			prices[i] = model.GoldPriceResponse{
				TypeName:  p.TypeName,
				Branch:    p.Branch,
				BuyPrice:  p.BuyPrice,
				SellPrice: p.SellPrice,
				Date:      p.Date,
				Source:    p.Source,
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("[GoldService] Failed to fetch gold prices: %v", err)
		return nil, fmt.Errorf("failed to fetch gold prices: %w", err)
	}

	// Cache for 1 hour
	s.cache.Set(cacheKey, prices, 1*time.Hour)
	log.Printf("[GoldService] Successfully fetched %d gold prices", len(prices))
	return prices, nil
}

// GetGoldPriceByType retrieves gold prices filtered by type name.
func (s *GoldService) GetGoldPriceByType(ctx context.Context, typeName string) (*model.GoldPriceResponse, error) {
	prices, err := s.GetGoldPrices(ctx)
	if err != nil {
		return nil, err
	}

	for _, price := range prices {
		if price.TypeName == typeName {
			return &price, nil
		}
	}

	return nil, fmt.Errorf("gold type '%s' not found", typeName)
}
