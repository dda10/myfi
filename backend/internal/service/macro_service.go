package service

import (
	"context"
	"log"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"
)

// MacroService provides macroeconomic indicators relevant to the VN market.
// TODO: Integrate with a dedicated macro data source when available in vnstock-go.
type MacroService struct {
	cache *infra.Cache
}

// NewMacroService creates a new MacroService.
func NewMacroService(cache *infra.Cache) *MacroService {
	return &MacroService{cache: cache}
}

// GetMacroIndicators fetches macroeconomic indicators for the VN market with 6-hour cache TTL.
// TODO: Integrate with vnstock-go macro API or external macro data source.
func (s *MacroService) GetMacroIndicators(ctx context.Context) (*model.MacroData, error) {
	cacheKey := "macro:indicators:vn"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*model.MacroData); ok {
			log.Println("[MacroService] Cache hit for macro indicators")
			return data, nil
		}
	}

	// Placeholder: vnstock-go does not yet provide macro data
	data := &model.MacroData{
		Indicators: []model.MacroIndicator{
			{Name: "GDP Growth Rate", Code: "VN_GDP_GROWTH", Value: 0, Unit: "%", Period: "quarterly", Country: "VN", Source: "placeholder"},
			{Name: "CPI Inflation", Code: "VN_CPI", Value: 0, Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "Interest Rate (Refinancing)", Code: "VN_INTEREST_RATE", Value: 0, Unit: "%", Period: "current", Country: "VN", Source: "placeholder"},
			{Name: "USD/VND Exchange Rate", Code: "VN_USDVND", Value: 0, Unit: "VND", Period: "current", Country: "VN", Source: "placeholder"},
			{Name: "Trade Balance", Code: "VN_TRADE_BALANCE", Value: 0, Unit: "USD million", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "FDI Inflow", Code: "VN_FDI", Value: 0, Unit: "USD million", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "Industrial Production Index", Code: "VN_IPI", Value: 0, Unit: "index", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "PMI Manufacturing", Code: "VN_PMI", Value: 0, Unit: "index", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "Credit Growth", Code: "VN_CREDIT_GROWTH", Value: 0, Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "M2 Money Supply Growth", Code: "VN_M2_GROWTH", Value: 0, Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
		},
	}

	s.cache.Set(cacheKey, data, 6*time.Hour)
	log.Println("[MacroService] Macro indicators fetched (placeholder — no data source yet)")
	return data, nil
}

// GetIndicatorByCode returns a specific macro indicator by its code.
// TODO: Integrate with real data source.
func (s *MacroService) GetIndicatorByCode(ctx context.Context, code string) (*model.MacroIndicator, error) {
	data, err := s.GetMacroIndicators(ctx)
	if err != nil {
		return nil, err
	}

	for _, ind := range data.Indicators {
		if ind.Code == code {
			return &ind, nil
		}
	}

	return nil, nil
}
