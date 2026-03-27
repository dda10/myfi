package service

import (
	"context"
	"log"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
)

// FundService provides open fund (mutual fund) data.
// TODO: Integrate with a dedicated fund data source when available in vnstock-go.
type FundService struct {
	cache *infra.Cache
}

// NewFundService creates a new FundService.
func NewFundService(cache *infra.Cache) *FundService {
	return &FundService{cache: cache}
}

// GetFundList returns the list of available open funds.
// TODO: Integrate with vnstock-go fund API or external fund data source.
func (s *FundService) GetFundList(ctx context.Context) ([]model.FundInfo, error) {
	cacheKey := "fund:list"
	if cached, found := s.cache.Get(cacheKey); found {
		if funds, ok := cached.([]model.FundInfo); ok {
			log.Println("[FundService] Cache hit for fund list")
			return funds, nil
		}
	}

	// Placeholder: vnstock-go does not yet provide fund data
	// Return empty list until data source is available
	funds := []model.FundInfo{}

	s.cache.Set(cacheKey, funds, 24*time.Hour)
	log.Println("[FundService] Fund list fetched (placeholder — no data source yet)")
	return funds, nil
}

// GetFundNAV returns the NAV for a specific fund.
// TODO: Integrate with vnstock-go fund API or external fund data source.
func (s *FundService) GetFundNAV(ctx context.Context, fundCode string) (*model.FundNAV, error) {
	cacheKey := "fund:nav:" + fundCode
	if cached, found := s.cache.Get(cacheKey); found {
		if nav, ok := cached.(*model.FundNAV); ok {
			log.Printf("[FundService] Cache hit for fund NAV: %s", fundCode)
			return nav, nil
		}
	}

	// Placeholder: no data source yet
	log.Printf("[FundService] Fund NAV not available for %s (no data source yet)", fundCode)
	return nil, nil
}

// GetFundPerformance returns performance metrics for a specific fund.
// TODO: Integrate with vnstock-go fund API or external fund data source.
func (s *FundService) GetFundPerformance(ctx context.Context, fundCode string) (*model.FundPerformance, error) {
	cacheKey := "fund:performance:" + fundCode
	if cached, found := s.cache.Get(cacheKey); found {
		if perf, ok := cached.(*model.FundPerformance); ok {
			log.Printf("[FundService] Cache hit for fund performance: %s", fundCode)
			return perf, nil
		}
	}

	// Placeholder: no data source yet
	log.Printf("[FundService] Fund performance not available for %s (no data source yet)", fundCode)
	return nil, nil
}
