package portfolio

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"myfi-backend/internal/infra"
)

// DefaultVNRiskFreeRate is the default VN government bond yield (4.5% per annum).
const DefaultVNRiskFreeRate = 0.045

// TradingDaysPerYear is the number of trading days used for annualization.
const TradingDaysPerYear = 252

// RiskService computes portfolio-level risk metrics:
// Sharpe ratio, max drawdown, beta, volatility, and VaR.
type RiskService struct {
	db     *sql.DB
	perf   *PerformanceEngine
	router *infra.DataSourceRouter
}

// NewRiskService creates a new RiskService instance.
func NewRiskService(db *sql.DB, perf *PerformanceEngine, router *infra.DataSourceRouter) *RiskService {
	return &RiskService{
		db:     db,
		perf:   perf,
		router: router,
	}
}

// ComputeSharpe computes the Sharpe ratio for a user's portfolio.
// Sharpe = (annualized return - risk-free rate) / annualized volatility.
func (s *RiskService) ComputeSharpe(ctx context.Context, userID string, riskFreeRate float64) (float64, error) {
	returns, err := s.getPortfolioReturns(ctx, userID)
	if err != nil || len(returns) == 0 {
		return 0, err
	}

	vol := annualizedVolatility(returns)
	if vol == 0 {
		return 0, nil
	}

	avgDaily := meanf(returns)
	annualReturn := avgDaily * TradingDaysPerYear

	return (annualReturn - riskFreeRate) / vol, nil
}

// ComputeMaxDrawdown computes the largest peak-to-trough percentage decline in NAV.
func (s *RiskService) ComputeMaxDrawdown(ctx context.Context, userID string) (float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0)

	navHistory, err := s.perf.GetEquityCurve(ctx, userID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get equity curve: %w", err)
	}
	if len(navHistory) < 2 {
		return 0, nil
	}

	peak := navHistory[0].NAV
	maxDD := 0.0
	for _, snap := range navHistory {
		if snap.NAV > peak {
			peak = snap.NAV
		}
		if peak > 0 {
			dd := (peak - snap.NAV) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD, nil
}

// ComputeBeta computes the portfolio beta relative to a benchmark (e.g., VN-Index).
func (s *RiskService) ComputeBeta(ctx context.Context, userID string, benchmark string) (float64, error) {
	// TODO: implement benchmark correlation
	return 1.0, nil
}

// ComputeVolatility computes the annualized volatility of the portfolio.
func (s *RiskService) ComputeVolatility(ctx context.Context, userID string) (float64, error) {
	returns, err := s.getPortfolioReturns(ctx, userID)
	if err != nil || len(returns) == 0 {
		return 0, err
	}
	return annualizedVolatility(returns), nil
}

// ComputeVaR computes the Value at Risk at the given confidence level (e.g., 0.95).
func (s *RiskService) ComputeVaR(ctx context.Context, userID string, confidence float64) (float64, error) {
	returns, err := s.getPortfolioReturns(ctx, userID)
	if err != nil || len(returns) < 2 {
		return 0, err
	}

	// Historical VaR: sort returns, pick the percentile
	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	sortFloat64s(sorted)

	idx := int(float64(len(sorted)) * (1 - confidence))
	if idx < 0 {
		idx = 0
	}
	return -sorted[idx], nil // return as positive loss
}

// --- helpers ---

// getPortfolioReturns computes daily returns from the NAV equity curve.
func (s *RiskService) getPortfolioReturns(ctx context.Context, userID string) ([]float64, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0)

	navHistory, err := s.perf.GetEquityCurve(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get equity curve: %w", err)
	}
	if len(navHistory) < 2 {
		return nil, nil
	}

	returns := make([]float64, 0, len(navHistory)-1)
	for i := 1; i < len(navHistory); i++ {
		if navHistory[i-1].NAV > 0 {
			r := (navHistory[i].NAV - navHistory[i-1].NAV) / navHistory[i-1].NAV
			returns = append(returns, r)
		}
	}
	return returns, nil
}

// meanf computes the arithmetic mean of a float64 slice.
func meanf(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// annualizedVolatility computes annualized standard deviation of daily returns.
func annualizedVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	avg := meanf(returns)
	sumSq := 0.0
	for _, r := range returns {
		d := r - avg
		sumSq += d * d
	}
	variance := sumSq / float64(len(returns)-1)
	return sqrtf(variance) * sqrtf(TradingDaysPerYear)
}

// sqrtf is a simple square root using math.Sqrt.
func sqrtf(x float64) float64 {
	if x <= 0 {
		return 0
	}
	return math.Sqrt(x)
}

// sortFloat64s sorts a float64 slice in ascending order.
func sortFloat64s(data []float64) {
	sort.Float64s(data)
}
