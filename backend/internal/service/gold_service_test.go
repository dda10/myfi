package service

import (
	"context"
	"testing"

	"myfi-backend/internal/infra"
)

func TestGoldService_GetGoldPrices(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()

	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create GoldService: %v", err)
	}

	ctx := context.Background()

	// First call - should fetch from API
	prices, err := goldService.GetGoldPrices(ctx)
	if err != nil {
		t.Logf("Gold API request failed (may be expected if API is down): %v", err)
		t.Skip("Skipping test - Gold API unavailable")
		return
	}

	if len(prices) == 0 {
		t.Fatal("Expected at least one gold price record")
	}

	t.Logf("Fetched %d gold prices", len(prices))

	// Validate first record
	first := prices[0]
	if first.TypeName == "" {
		t.Error("Expected TypeName to be non-empty")
	}
	if first.BuyPrice <= 0 {
		t.Error("Expected BuyPrice to be positive")
	}
	if first.SellPrice <= 0 {
		t.Error("Expected SellPrice to be positive")
	}
	if first.Source == "" {
		t.Error("Expected Source to be non-empty")
	}

	t.Logf("Sample record: %s - Buy: %.0f, Sell: %.0f (Source: %s)",
		first.TypeName, first.BuyPrice, first.SellPrice, first.Source)

	// Second call - should hit cache
	cachedPrices, err := goldService.GetGoldPrices(ctx)
	if err != nil {
		t.Fatalf("Failed to get cached gold prices: %v", err)
	}

	if len(cachedPrices) != len(prices) {
		t.Errorf("Expected %d cached prices, got %d", len(prices), len(cachedPrices))
	}

	t.Logf("Cache hit - returned %d prices", len(cachedPrices))
}

func TestGoldService_GetGoldPriceByType(t *testing.T) {
	cache := infra.NewCache()
	rateLimiter := infra.NewRateLimiter()

	goldService, err := NewGoldService(cache, rateLimiter)
	if err != nil {
		t.Fatalf("Failed to create GoldService: %v", err)
	}

	ctx := context.Background()

	// Get all prices first
	prices, err := goldService.GetGoldPrices(ctx)
	if err != nil {
		t.Logf("Gold API request failed: %v", err)
		t.Skip("Skipping test - Gold API unavailable")
		return
	}

	if len(prices) == 0 {
		t.Skip("No gold prices available to test")
	}

	// Test getting a specific type (use the first available type)
	firstType := prices[0].TypeName
	price, err := goldService.GetGoldPriceByType(ctx, firstType)
	if err != nil {
		t.Fatalf("Failed to get gold price by type '%s': %v", firstType, err)
	}

	if price.TypeName != firstType {
		t.Errorf("Expected type '%s', got '%s'", firstType, price.TypeName)
	}

	t.Logf("Found price for type '%s': Buy: %.0f, Sell: %.0f",
		price.TypeName, price.BuyPrice, price.SellPrice)

	// Test non-existent type
	_, err = goldService.GetGoldPriceByType(ctx, "NonExistentType")
	if err == nil {
		t.Error("Expected error for non-existent gold type")
	}
}
