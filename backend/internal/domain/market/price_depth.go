package market

import (
	"context"

	"github.com/dda10/vnstock-go"
)

// GetPriceDepth fetches 3-level order book depth for a symbol.
func (s *PriceService) GetPriceDepth(ctx context.Context, symbol string) (vnstock.PriceDepth, error) {
	depth, _, err := s.router.FetchPriceDepth(ctx, symbol)
	if err != nil {
		return vnstock.PriceDepth{}, err
	}
	return depth, nil
}
