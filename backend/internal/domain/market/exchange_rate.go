package market

import (
	"context"
	"log/slog"
	"time"

	"myfi-backend/internal/infra"

	"github.com/dda10/vnstock-go"
)

// ExchangeRateService fetches Vietcombank official exchange rates.
type ExchangeRateService struct {
	cache *infra.Cache
}

func NewExchangeRateService(cache *infra.Cache) *ExchangeRateService {
	return &ExchangeRateService{cache: cache}
}

const fxCacheTTL = 1 * time.Hour

// ExchangeRate represents a single VCB exchange rate entry.
type ExchangeRate struct {
	CurrencyCode string  `json:"currencyCode"`
	CurrencyName string  `json:"currencyName"`
	BuyCash      float64 `json:"buyCash"`
	BuyTransfer  float64 `json:"buyTransfer"`
	Sell         float64 `json:"sell"`
}

// GetExchangeRates returns Vietcombank official exchange rates.
func (s *ExchangeRateService) GetExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	cacheKey := "fx:vcb:rates"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]ExchangeRate); ok {
			return data, nil
		}
	}

	vcbRates, err := vnstock.VCBExchangeRate(ctx, "") // empty = today
	if err != nil {
		slog.Error("failed to fetch VCB exchange rates", "error", err)
		return []ExchangeRate{}, nil
	}

	results := make([]ExchangeRate, 0, len(vcbRates))
	for _, r := range vcbRates {
		results = append(results, ExchangeRate{
			CurrencyCode: r.CurrencyCode,
			CurrencyName: r.CurrencyName,
			BuyCash:      r.BuyCash,
			BuyTransfer:  r.BuyTransfer,
			Sell:         r.Sell,
		})
	}

	s.cache.Set(cacheKey, results, fxCacheTTL)
	return results, nil
}
