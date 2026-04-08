package market

import (
	"context"
	"log/slog"
	"time"

	"myfi-backend/internal/infra"
)

// WorldMarketService fetches world indices (S&P 500, NASDAQ, Nikkei, etc.)
// via the MSN connector.
type WorldMarketService struct {
	router *infra.DataSourceRouter
	cache  *infra.Cache
}

func NewWorldMarketService(router *infra.DataSourceRouter, cache *infra.Cache) *WorldMarketService {
	return &WorldMarketService{router: router, cache: cache}
}

const worldCacheTTL = 5 * time.Minute

// WorldIndexQuote represents a single world market index.
type WorldIndexQuote struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Value         float64 `json:"value"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	Source        string  `json:"source"`
}

// Well-known world index symbols for the MSN connector.
var worldIndexSymbols = []string{
	".DJI",   // Dow Jones
	".SPX",   // S&P 500
	".IXIC",  // NASDAQ
	".N225",  // Nikkei 225
	".HSI",   // Hang Seng
	".FTSE",  // FTSE 100
	".GDAXI", // DAX
	".SSEC",  // Shanghai Composite
	".KS11",  // KOSPI
	".TWII",  // TAIEX
}

// GetWorldIndices returns quotes for major world market indices.
func (s *WorldMarketService) GetWorldIndices(ctx context.Context) ([]WorldIndexQuote, error) {
	cacheKey := "world:indices"
	if cached, found := s.cache.Get(cacheKey); found {
		if data, ok := cached.([]WorldIndexQuote); ok {
			return data, nil
		}
	}

	msnClient := s.router.MSNClient()
	if msnClient == nil {
		slog.Warn("MSN connector not available, returning empty world indices")
		return []WorldIndexQuote{}, nil
	}

	quotes, err := msnClient.RealTimeQuotes(ctx, worldIndexSymbols)
	if err != nil {
		slog.Error("failed to fetch world indices from MSN", "error", err)
		return []WorldIndexQuote{}, nil
	}

	results := make([]WorldIndexQuote, 0, len(quotes))
	for _, q := range quotes {
		// Quote doesn't have Change/ChangePercent — compute from Open vs Close
		change := 0.0
		changePct := 0.0
		if q.Open > 0 {
			change = q.Close - q.Open
			changePct = (change / q.Open) * 100
		}
		results = append(results, WorldIndexQuote{
			Symbol:        q.Symbol,
			Name:          nameForWorldIndex(q.Symbol),
			Value:         q.Close,
			Change:        change,
			ChangePercent: changePct,
			Source:        "MSN",
		})
	}

	s.cache.Set(cacheKey, results, worldCacheTTL)
	return results, nil
}

func nameForWorldIndex(symbol string) string {
	names := map[string]string{
		".DJI":   "Dow Jones",
		".SPX":   "S&P 500",
		".IXIC":  "NASDAQ",
		".N225":  "Nikkei 225",
		".HSI":   "Hang Seng",
		".FTSE":  "FTSE 100",
		".GDAXI": "DAX",
		".SSEC":  "Shanghai",
		".KS11":  "KOSPI",
		".TWII":  "TAIEX",
	}
	if n, ok := names[symbol]; ok {
		return n
	}
	return symbol
}
