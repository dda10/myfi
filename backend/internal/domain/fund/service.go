package fund

import (
	"context"
	"log/slog"
	"time"

	"myfi-backend/internal/infra"
)

const fundCacheTTL = 10 * time.Minute

// NewFundService creates a new FundService.
func NewFundService(router *infra.DataSourceRouter, cache *infra.Cache) *FundService {
	return &FundService{router: router, cache: cache}
}

// ListFunds returns all available mutual funds.
func (s *FundService) ListFunds(ctx context.Context) ([]FundRecord, error) {
	cacheKey := "fund:listing"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]FundRecord); ok {
			return data, nil
		}
	}

	client := s.router.FMARKETClient()
	if client == nil {
		slog.Warn("FMARKET connector not available")
		return []FundRecord{}, nil
	}

	records, err := client.FundListing(ctx)
	if err != nil {
		slog.Error("failed to fetch fund listing", "error", err)
		return nil, err
	}

	result := make([]FundRecord, 0, len(records))
	for _, r := range records {
		result = append(result, FundRecord{
			FundCode:          r.FundCode,
			FundName:          r.FundName,
			ManagementCompany: r.ManagementCompany,
			FundType:          r.FundType,
			NAV:               r.NAV,
			InceptionDate:     r.InceptionDate,
		})
	}

	s.cache.Set(cacheKey, result, fundCacheTTL)
	return result, nil
}

// SearchFunds returns funds matching the search query.
func (s *FundService) SearchFunds(ctx context.Context, query string) ([]FundRecord, error) {
	if query == "" {
		return s.ListFunds(ctx)
	}

	client := s.router.FMARKETClient()
	if client == nil {
		slog.Warn("FMARKET connector not available")
		return []FundRecord{}, nil
	}

	records, err := client.FundFilter(ctx, query)
	if err != nil {
		slog.Error("failed to search funds", "error", err, "query", query)
		return nil, err
	}

	result := make([]FundRecord, 0, len(records))
	for _, r := range records {
		result = append(result, FundRecord{
			FundCode:          r.FundCode,
			FundName:          r.FundName,
			ManagementCompany: r.ManagementCompany,
			FundType:          r.FundType,
			NAV:               r.NAV,
			InceptionDate:     r.InceptionDate,
		})
	}

	return result, nil
}

// GetTopHoldings returns the top stock holdings for a fund.
func (s *FundService) GetTopHoldings(ctx context.Context, fundCode string) ([]FundHolding, error) {
	cacheKey := "fund:holdings:" + fundCode
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]FundHolding); ok {
			return data, nil
		}
	}

	client := s.router.FMARKETClient()
	if client == nil {
		slog.Warn("FMARKET connector not available")
		return []FundHolding{}, nil
	}

	holdings, err := client.FundTopHolding(ctx, fundCode)
	if err != nil {
		slog.Error("failed to fetch fund holdings", "error", err, "fundCode", fundCode)
		return nil, err
	}

	result := make([]FundHolding, 0, len(holdings))
	for _, h := range holdings {
		result = append(result, FundHolding{
			StockSymbol: h.StockSymbol,
			StockName:   h.StockName,
			Percentage:  h.Percentage,
			MarketValue: h.MarketValue,
		})
	}

	s.cache.Set(cacheKey, result, fundCacheTTL)
	return result, nil
}

// GetNAVHistory returns NAV history for a fund over a date range.
func (s *FundService) GetNAVHistory(ctx context.Context, fundCode string, start, end time.Time) ([]FundNAV, error) {
	client := s.router.FMARKETClient()
	if client == nil {
		slog.Warn("FMARKET connector not available")
		return []FundNAV{}, nil
	}

	navs, err := client.FundNAVReport(ctx, fundCode, start, end)
	if err != nil {
		slog.Error("failed to fetch fund NAV history", "error", err, "fundCode", fundCode)
		return nil, err
	}

	result := make([]FundNAV, 0, len(navs))
	for _, n := range navs {
		result = append(result, FundNAV{
			Date:       n.Date,
			NAVPerUnit: n.NAVPerUnit,
			TotalNAV:   n.TotalNAV,
		})
	}

	return result, nil
}

// GetAllocation returns industry and asset allocation for a fund.
func (s *FundService) GetAllocation(ctx context.Context, fundCode string) (*FundAllocation, error) {
	cacheKey := "fund:allocation:" + fundCode
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*FundAllocation); ok {
			return data, nil
		}
	}

	client := s.router.FMARKETClient()
	if client == nil {
		slog.Warn("FMARKET connector not available")
		return &FundAllocation{}, nil
	}

	industryAllocs, err := client.FundIndustryHolding(ctx, fundCode)
	if err != nil {
		slog.Error("failed to fetch fund industry allocation", "error", err, "fundCode", fundCode)
		return nil, err
	}

	assetAllocs, err := client.FundAssetHolding(ctx, fundCode)
	if err != nil {
		slog.Error("failed to fetch fund asset allocation", "error", err, "fundCode", fundCode)
		return nil, err
	}

	industry := make([]FundIndustryAlloc, 0, len(industryAllocs))
	for _, a := range industryAllocs {
		industry = append(industry, FundIndustryAlloc{
			IndustryName: a.IndustryName,
			Percentage:   a.Percentage,
		})
	}

	asset := make([]FundAssetAlloc, 0, len(assetAllocs))
	for _, a := range assetAllocs {
		asset = append(asset, FundAssetAlloc{
			AssetClass: a.AssetClass,
			Percentage: a.Percentage,
		})
	}

	result := &FundAllocation{Industry: industry, Asset: asset}
	s.cache.Set(cacheKey, result, fundCacheTTL)
	return result, nil
}
