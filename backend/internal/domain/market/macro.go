package market

import (
	"context"
	"log"
	"time"

	"myfi-backend/internal/infra"

	"github.com/dda10/vnstock-go"
)

// MacroService provides macroeconomic indicators relevant to the VN market.
type MacroService struct {
	cache  *infra.Cache
	router *infra.DataSourceRouter
}

func NewMacroService(cache *infra.Cache, router *infra.DataSourceRouter) *MacroService {
	return &MacroService{cache: cache, router: router}
}

const macroCacheTTL = 1 * time.Hour

func (s *MacroService) GetMacroIndicators(ctx context.Context) (*MacroIndicators, error) {
	cacheKey := "macro:indicators:vn"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*MacroIndicators); ok {
			return data, nil
		}
	}
	data := s.buildMacroSnapshot(ctx)
	s.cache.Set(cacheKey, data, macroCacheTTL)
	return data, nil
}

func (s *MacroService) GetInterbankRates(ctx context.Context) ([]InterbankRate, error) {
	data, err := s.GetMacroIndicators(ctx)
	if err != nil {
		return nil, err
	}
	return data.InterbankRates, nil
}

func (s *MacroService) GetBondYields(ctx context.Context) ([]BondYield, error) {
	data, err := s.GetMacroIndicators(ctx)
	if err != nil {
		return nil, err
	}
	return data.BondYields, nil
}

func (s *MacroService) GetFXRates(ctx context.Context) ([]MacroFXRate, error) {
	data, err := s.GetMacroIndicators(ctx)
	if err != nil {
		return nil, err
	}
	return data.FXRates, nil
}

func (s *MacroService) GetIndicatorByCode(ctx context.Context, code string) (*MacroIndicator, error) {
	legacyData, err := s.getLegacyMacroData(ctx)
	if err != nil {
		return nil, err
	}
	for _, ind := range legacyData.Indicators {
		if ind.Code == code {
			return &ind, nil
		}
	}
	return nil, nil
}

func (s *MacroService) buildMacroSnapshot(ctx context.Context) *MacroIndicators {
	indicators := &MacroIndicators{
		InterbankRates: []InterbankRate{
			{Tenor: "overnight"}, {Tenor: "1w"}, {Tenor: "2w"},
			{Tenor: "1m"}, {Tenor: "3m"}, {Tenor: "6m"}, {Tenor: "12m"},
		},
		BondYields: []BondYield{
			{Tenor: "1y"}, {Tenor: "2y"}, {Tenor: "3y"},
			{Tenor: "5y"}, {Tenor: "10y"}, {Tenor: "15y"},
		},
		UpdatedAt: time.Now(),
	}

	// --- VCB Exchange Rates (vnstock-go v2) ---
	vcbRates, err := vnstock.VCBExchangeRate(ctx, "") // empty = today
	if err != nil {
		log.Printf("[MacroService] VCB exchange rates failed: %v", err)
		indicators.FXRates = []MacroFXRate{
			{Pair: "USD/VND", Source: "placeholder"},
			{Pair: "EUR/VND", Source: "placeholder"},
			{Pair: "JPY/VND", Source: "placeholder"},
			{Pair: "CNY/VND", Source: "placeholder"},
		}
	} else {
		fxRates := make([]MacroFXRate, 0, len(vcbRates))
		for _, r := range vcbRates {
			fxRates = append(fxRates, MacroFXRate{
				Pair:   r.CurrencyCode + "/VND",
				Rate:   r.Sell,
				Source: "VCB",
			})
		}
		indicators.FXRates = fxRates
	}

	// --- Gold Prices (vnstock-go v2) ---
	if goldClient := s.router.GOLDClient(); goldClient != nil {
		goldPrices, err := goldClient.GoldPrice(ctx, vnstock.GoldPriceRequest{})
		if err != nil {
			log.Printf("[MacroService] Gold prices failed: %v", err)
		} else {
			gps := make([]MacroGoldPrice, 0, len(goldPrices))
			for _, gp := range goldPrices {
				gps = append(gps, MacroGoldPrice{
					TypeName:  gp.TypeName,
					BuyPrice:  gp.BuyPrice,
					SellPrice: gp.SellPrice,
					Source:    gp.Source,
				})
			}
			indicators.GoldPrices = gps
		}
	}

	// World indices: MSN connector doesn't expose a WorldIndices method.
	// Placeholder until a dedicated world-index API is available.
	indicators.WorldIndices = []WorldIndex{}

	return indicators
}

func (s *MacroService) getLegacyMacroData(ctx context.Context) (*MacroData, error) {
	cacheKey := "macro:legacy:vn"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.(*MacroData); ok {
			return data, nil
		}
	}
	data := &MacroData{
		Indicators: []MacroIndicator{
			{Name: "GDP Growth Rate", Code: "VN_GDP_GROWTH", Unit: "%", Period: "quarterly", Country: "VN", Source: "placeholder"},
			{Name: "CPI Inflation", Code: "VN_CPI", Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "Interest Rate", Code: "VN_INTEREST_RATE", Unit: "%", Period: "current", Country: "VN", Source: "placeholder"},
			{Name: "USD/VND", Code: "VN_USDVND", Unit: "VND", Period: "current", Country: "VN", Source: "placeholder"},
			{Name: "Trade Balance", Code: "VN_TRADE_BALANCE", Unit: "USD million", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "FDI Inflow", Code: "VN_FDI", Unit: "USD million", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "IPI", Code: "VN_IPI", Unit: "index", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "PMI", Code: "VN_PMI", Unit: "index", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "Credit Growth", Code: "VN_CREDIT_GROWTH", Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
			{Name: "M2 Growth", Code: "VN_M2_GROWTH", Unit: "%", Period: "monthly", Country: "VN", Source: "placeholder"},
		},
	}
	s.cache.Set(cacheKey, data, macroCacheTTL)
	return data, nil
}
