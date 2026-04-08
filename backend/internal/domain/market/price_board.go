package market

import (
	"context"
	"strings"
	"time"

	"github.com/dda10/vnstock-go"
)

const priceBoardCacheTTL = 30 * time.Second

// GetPriceBoard fetches real-time bid/ask depth for symbols via the Data Source Router.
// Cache TTL is 30 seconds (more real-time than OHLCV).
func (s *PriceService) GetPriceBoard(ctx context.Context, symbols []string) ([]vnstock.PriceBoard, error) {
	cacheKey := "priceboard:" + strings.Join(symbols, ",")
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]vnstock.PriceBoard); ok {
			return data, nil
		}
	}

	boards, _, err := s.router.FetchPriceBoard(ctx, symbols)
	if err != nil {
		return nil, err
	}

	s.cache.Set(cacheKey, boards, priceBoardCacheTTL)
	return boards, nil
}
