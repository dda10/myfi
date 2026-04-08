package market

import (
	"context"
	"log/slog"
	"time"

	"myfi-backend/internal/infra"

	"github.com/dda10/vnstock-go"
)

// GoldPriceService fetches SJC + BTMC gold prices via the GOLD connector.
type GoldPriceService struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
}

func NewGoldPriceService(router *infra.DataSourceRouter, cache *infra.Cache) *GoldPriceService {
	return &GoldPriceService{router: router, cache: cache}
}

const goldCacheTTL = 15 * time.Minute

// GoldQuote represents a gold price entry.
type GoldQuote struct {
	TypeName  string  `json:"typeName"`
	BuyPrice  float64 `json:"buyPrice"`
	SellPrice float64 `json:"sellPrice"`
	Source    string  `json:"source"`
	UpdatedAt string  `json:"updatedAt,omitempty"`
}

// GetGoldPrices returns SJC and BTMC gold prices.
func (s *GoldPriceService) GetGoldPrices(ctx context.Context) ([]GoldQuote, error) {
	cacheKey := "gold:prices"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]GoldQuote); ok {
			return data, nil
		}
	}

	goldClient := s.router.GOLDClient()
	if goldClient == nil {
		slog.Warn("GOLD connector not available, returning empty gold prices")
		return []GoldQuote{}, nil
	}

	prices, err := goldClient.GoldPrice(ctx, vnstock.GoldPriceRequest{})
	if err != nil {
		slog.Error("failed to fetch gold prices", "error", err)
		return []GoldQuote{}, nil
	}

	results := make([]GoldQuote, 0, len(prices))
	for _, gp := range prices {
		results = append(results, GoldQuote{
			TypeName:  gp.TypeName,
			BuyPrice:  gp.BuyPrice,
			SellPrice: gp.SellPrice,
			Source:    gp.Source,
		})
	}

	s.cache.Set(cacheKey, results, goldCacheTTL)
	return results, nil
}
