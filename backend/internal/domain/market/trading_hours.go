package market

import (
	"context"
	"log/slog"

	"github.com/dda10/vnstock-go"
)

// TradingHoursService wraps vnstock.TradingHours() for market session status.
type TradingHoursService struct{}

func NewTradingHoursService() *TradingHoursService {
	return &TradingHoursService{}
}

// MarketStatus represents the current trading session status.
type MarketStatus struct {
	Market        string `json:"market"`
	IsTradingHour bool   `json:"isTradingHour"`
	Session       string `json:"session"` // pre-open, continuous, ATC, closed, break
}

// GetStatus returns the current trading session status for a market.
func (s *TradingHoursService) GetStatus(_ context.Context, market string) MarketStatus {
	if market == "" {
		market = "HOSE"
	}

	hours, err := vnstock.TradingHours(market)
	if err != nil {
		slog.Warn("TradingHours failed, assuming closed", "market", market, "error", err)
		return MarketStatus{Market: market, Session: "closed"}
	}

	return MarketStatus{
		Market:        market,
		IsTradingHour: hours.IsTradingHour,
		Session:       hours.TradingSession,
	}
}

// GetAllStatuses returns trading status for HOSE, HNX, and UPCOM.
func (s *TradingHoursService) GetAllStatuses(_ context.Context) []MarketStatus {
	markets := []string{"HOSE", "HNX", "UPCOM"}
	results := make([]MarketStatus, 0, len(markets))
	for _, m := range markets {
		hours, err := vnstock.TradingHours(m)
		if err != nil {
			results = append(results, MarketStatus{Market: m, Session: "closed"})
			continue
		}
		results = append(results, MarketStatus{
			Market:        m,
			IsTradingHour: hours.IsTradingHour,
			Session:       hours.TradingSession,
		})
	}
	return results
}
