package service

import (
	"context"
	"log"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
)

// CommodityService provides commodity market data including gold, oil, metals, and agricultural products.
type CommodityService struct {
	goldService *GoldService
	cache       *infra.Cache
}

// NewCommodityService creates a new CommodityService.
func NewCommodityService(goldService *GoldService, cache *infra.Cache) *CommodityService {
	return &CommodityService{
		goldService: goldService,
		cache:       cache,
	}
}

// GetAllCommodities fetches all commodity data with 1-hour cache TTL.
func (s *CommodityService) GetAllCommodities(ctx context.Context) (*model.CommodityData, error) {
	cacheKey := "commodity:all"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.CommodityData); ok {
			log.Println("[CommodityService] Cache hit for all commodities")
			return data, nil
		}
	}

	data := &model.CommodityData{}

	// VN gold prices via GoldService
	goldPrices, err := s.goldService.GetGoldPrices(ctx)
	if err != nil {
		log.Printf("[CommodityService] Failed to fetch VN gold prices: %v", err)
		data.IsStale = true
	} else {
		data.GoldVN = goldPrices
	}

	// Global gold OHLCV — vnstock-go GOLD connector does not provide global OHLCV
	// TODO: Integrate global gold OHLCV data source (e.g., via FMP connector)
	data.GoldGlobal = []model.OHLCVBar{}

	// Energy commodities — no direct vnstock-go support
	// TODO: Integrate energy commodity data source
	data.Energy = s.getEnergyPlaceholders()

	// Metal commodities — no direct vnstock-go support
	// TODO: Integrate metal commodity data source
	data.Metals = s.getMetalPlaceholders()

	// Agricultural commodities — no direct vnstock-go support
	// TODO: Integrate agricultural commodity data source
	data.Agricultural = s.getAgriculturalPlaceholders()

	// VN pork prices — no direct vnstock-go support
	// TODO: Integrate VN pork price data source
	data.VNPork = []model.CommodityPrice{
		{Name: "Heo hơi (VN)", Symbol: "VN_PORK", Category: "agricultural", Price: 0, Currency: "VND", Unit: "kg", Source: "placeholder"},
	}

	s.cache.Set(cacheKey, data, 1*time.Hour)
	log.Println("[CommodityService] Fetched all commodity data")
	return data, nil
}

// GetGoldVN returns VN gold prices via GoldService.
func (s *CommodityService) GetGoldVN(ctx context.Context) ([]model.GoldPriceResponse, error) {
	return s.goldService.GetGoldPrices(ctx)
}

// GetEnergyCommodities returns energy commodity prices.
// TODO: Integrate with real data source.
func (s *CommodityService) GetEnergyCommodities(ctx context.Context) ([]model.CommodityPrice, error) {
	return s.getEnergyPlaceholders(), nil
}

// GetMetalCommodities returns metal commodity prices.
// TODO: Integrate with real data source.
func (s *CommodityService) GetMetalCommodities(ctx context.Context) ([]model.CommodityPrice, error) {
	return s.getMetalPlaceholders(), nil
}

// GetAgriculturalCommodities returns agricultural commodity prices.
// TODO: Integrate with real data source.
func (s *CommodityService) GetAgriculturalCommodities(ctx context.Context) ([]model.CommodityPrice, error) {
	return s.getAgriculturalPlaceholders(), nil
}

func (s *CommodityService) getEnergyPlaceholders() []model.CommodityPrice {
	return []model.CommodityPrice{
		{Name: "Crude Oil (WTI)", Symbol: "CL", Category: "energy", Price: 0, Currency: "USD", Unit: "barrel", Source: "placeholder"},
		{Name: "Crude Oil (Brent)", Symbol: "BZ", Category: "energy", Price: 0, Currency: "USD", Unit: "barrel", Source: "placeholder"},
		{Name: "Natural Gas", Symbol: "NG", Category: "energy", Price: 0, Currency: "USD", Unit: "MMBtu", Source: "placeholder"},
	}
}

func (s *CommodityService) getMetalPlaceholders() []model.CommodityPrice {
	return []model.CommodityPrice{
		{Name: "Steel (HRC)", Symbol: "HRC", Category: "metals", Price: 0, Currency: "USD", Unit: "ton", Source: "placeholder"},
		{Name: "Iron Ore", Symbol: "IO", Category: "metals", Price: 0, Currency: "USD", Unit: "ton", Source: "placeholder"},
	}
}

func (s *CommodityService) getAgriculturalPlaceholders() []model.CommodityPrice {
	return []model.CommodityPrice{
		{Name: "Corn", Symbol: "ZC", Category: "agricultural", Price: 0, Currency: "USD", Unit: "bushel", Source: "placeholder"},
		{Name: "Soybean", Symbol: "ZS", Category: "agricultural", Price: 0, Currency: "USD", Unit: "bushel", Source: "placeholder"},
		{Name: "Sugar", Symbol: "SB", Category: "agricultural", Price: 0, Currency: "USD", Unit: "lb", Source: "placeholder"},
	}
}
