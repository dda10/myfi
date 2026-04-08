package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
)

// DefaultVNRiskFreeRate is the default VN government bond yield (4.5% per annum).
// Requirement 27.1: configurable default of 4.5%.
const DefaultVNRiskFreeRate = 0.045

// TradingDaysPerYear is the number of trading days used for annualization.
const TradingDaysPerYear = 252

// RiskService computes portfolio-level and per-holding risk metrics.
// Requirement 27: Sharpe ratio, max drawdown, beta, volatility, VaR, risk contribution.
type RiskService struct {
	db                *sql.DB
	performanceEngine *PerformanceEngine
	router            *infra.DataSourceRouter
}

// NewRiskService creates a new RiskService instance.
func NewRiskService(database *sql.DB, perfEngine *PerformanceEngine, router *infra.DataSourceRouter) *RiskService {
	return &RiskService{
		db:                database,
		performanceEngine: perfEngine,
		router:            router,
	}
}

// ComputeRiskMetrics returns the full set of risk analytics for a user's portfolio.
func (r *RiskService) ComputeRiskMetrics(ctx context.Context, userID string) (model.RiskMetrics, error) {
	metrics := model.RiskMetrics{
		RiskContribution: make(map[string]float64),
	}

	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0) // trailing 1-year

	// Get NAV equity curve for the trailing year
	navHistory, err := r.performanceEngine.GetEquityCurve(ctx, userID, startDate, endDate)
	if err != nil {
		return metrics, fmt.Errorf("failed to get equity curve: %w", err)
	}
	if len(navHistory) < 2 {
		log.Printf("[RiskService] Insufficient NAV history for user %s: %d points", userID, len(navHistory))
		return metrics, nil
	}

	// Compute daily portfolio returns from NAV history
	portfolioReturns := computeDailyReturns(navHistory)
	if len(portfolioReturns) == 0 {
		return metrics, nil
	}

	// Volatility: σ × √252
	volatility := computeAnnualizedVolatility(portfolioReturns)
	metrics.Volatility = volatility

	// Sharpe ratio: (annualized return - risk-free rate) / annualized volatility
	metrics.SharpeRatio = r.ComputeSharpeRatio(portfolioReturns, DefaultVNRiskFreeRate)

	// Max drawdown from NAV history
	metrics.MaxDrawdown = r.ComputeMaxDrawdown(navHistory)

	// VaR at 95% confidence (historical simulation)
	currentNAV := navHistory[len(navHistory)-1].NAV
	metrics.VaR95 = r.ComputeVaR(portfolioReturns, 0.95, currentNAV)

	// Beta against VN-Index
	benchmarkReturns, err := r.fetchBenchmarkReturns(ctx, startDate, endDate)
	if err != nil {
		log.Printf("[RiskService] Failed to fetch VN-Index returns for beta: %v", err)
	} else if len(benchmarkReturns) > 0 {
		metrics.Beta = r.ComputeBeta(portfolioReturns, benchmarkReturns)
	}

	// Per-holding risk contribution
	riskContrib, err := r.computeHoldingRiskContribution(ctx, userID, startDate, endDate, volatility)
	if err != nil {
		log.Printf("[RiskService] Failed to compute risk contribution: %v", err)
	} else {
		metrics.RiskContribution = riskContrib
	}

	return metrics, nil
}

// ComputeSharpeRatio computes the Sharpe ratio from daily returns.
// Sharpe = (annualized portfolio return - risk-free rate) / annualized volatility.
// Requirement 27.1.
func (r *RiskService) ComputeSharpeRatio(dailyReturns []float64, riskFreeRate float64) float64 {
	if len(dailyReturns) == 0 {
		return 0
	}

	vol := computeAnnualizedVolatility(dailyReturns)
	if vol == 0 {
		return 0
	}

	avgDailyReturn := mean(dailyReturns)
	annualizedReturn := avgDailyReturn * TradingDaysPerYear

	return (annualizedReturn - riskFreeRate) / vol
}

// ComputeMaxDrawdown computes the largest peak-to-trough percentage decline in NAV.
// Requirement 27.2.
func (r *RiskService) ComputeMaxDrawdown(navHistory []model.NAVSnapshot) float64 {
	if len(navHistory) < 2 {
		return 0
	}

	peak := navHistory[0].NAV
	maxDD := 0.0

	for _, snap := range navHistory {
		if snap.NAV > peak {
			peak = snap.NAV
		}
		if peak > 0 {
			drawdown := (peak - snap.NAV) / peak
			if drawdown > maxDD {
				maxDD = drawdown
			}
		}
	}

	return maxDD
}

// ComputeBeta computes portfolio beta against a benchmark using daily returns.
// Beta = Cov(portfolio, market) / Var(market).
// Requirement 27.3.
func (r *RiskService) ComputeBeta(portfolioReturns, benchmarkReturns []float64) float64 {
	// Align lengths: use the shorter of the two
	n := len(portfolioReturns)
	if len(benchmarkReturns) < n {
		n = len(benchmarkReturns)
	}
	if n < 2 {
		return 0
	}

	// Use the most recent n data points from each
	pReturns := portfolioReturns[len(portfolioReturns)-n:]
	bReturns := benchmarkReturns[len(benchmarkReturns)-n:]

	pMean := mean(pReturns)
	bMean := mean(bReturns)

	var cov, varB float64
	for i := 0; i < n; i++ {
		pDev := pReturns[i] - pMean
		bDev := bReturns[i] - bMean
		cov += pDev * bDev
		varB += bDev * bDev
	}

	if varB == 0 {
		return 0
	}

	return cov / varB
}

// ComputeVaR computes Value at Risk at the given confidence level using historical simulation.
// VaR = 5th percentile of daily returns × current NAV (for 95% confidence).
// Requirement 27.5.
func (r *RiskService) ComputeVaR(dailyReturns []float64, confidenceLevel float64, currentNAV float64) float64 {
	if len(dailyReturns) == 0 || currentNAV <= 0 {
		return 0
	}

	// Sort returns ascending
	sorted := make([]float64, len(dailyReturns))
	copy(sorted, dailyReturns)
	sort.Float64s(sorted)

	// Find the percentile index: for 95% confidence, we want the 5th percentile
	percentileIdx := int(math.Floor((1 - confidenceLevel) * float64(len(sorted))))
	if percentileIdx >= len(sorted) {
		percentileIdx = len(sorted) - 1
	}
	if percentileIdx < 0 {
		percentileIdx = 0
	}

	// VaR is the absolute loss at the percentile
	percentileReturn := sorted[percentileIdx]

	// Return as positive number representing potential loss
	return math.Abs(percentileReturn) * currentNAV
}

// fetchBenchmarkReturns fetches VN-Index daily returns for the given period.
func (r *RiskService) fetchBenchmarkReturns(ctx context.Context, startDate, endDate time.Time) ([]float64, error) {
	req := vnstock.IndexHistoryRequest{
		Name:     "VNINDEX",
		Start:    startDate,
		End:      endDate,
		Interval: "1D",
	}

	records, _, err := r.router.FetchIndexHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch VN-Index history: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("insufficient VN-Index data: got %d points", len(records))
	}

	// Sort by timestamp ascending
	sort.Slice(records, func(i, j int) bool {
		return records[i].Timestamp.Before(records[j].Timestamp)
	})

	// Compute daily returns from close prices
	var returns []float64
	for i := 1; i < len(records); i++ {
		if records[i-1].Close > 0 {
			ret := (records[i].Close - records[i-1].Close) / records[i-1].Close
			returns = append(returns, ret)
		}
	}

	return returns, nil
}

// computeHoldingRiskContribution computes each holding's percentage contribution
// to total portfolio volatility.
// Requirement 27.8.
func (r *RiskService) computeHoldingRiskContribution(ctx context.Context, userID string, startDate, endDate time.Time, portfolioVol float64) (map[string]float64, error) {
	contribution := make(map[string]float64)

	if portfolioVol == 0 {
		return contribution, nil
	}

	// Get current holdings
	rows, err := r.db.QueryContext(ctx,
		`SELECT symbol, asset_type, quantity, average_cost FROM assets WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query holdings: %w", err)
	}
	defer rows.Close()

	type holding struct {
		symbol    string
		assetType string
		quantity  float64
		avgCost   float64
	}

	var holdings []holding
	var totalValue float64
	for rows.Next() {
		var h holding
		if err := rows.Scan(&h.symbol, &h.assetType, &h.quantity, &h.avgCost); err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, h)
		totalValue += h.quantity * h.avgCost
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if totalValue == 0 || len(holdings) == 0 {
		return contribution, nil
	}

	// For each holding, compute its weight × individual volatility as a proxy
	// for marginal risk contribution. This is a simplified approach using
	// weight × asset volatility / portfolio volatility.
	navHistory, err := r.performanceEngine.GetEquityCurve(ctx, userID, startDate, endDate)
	if err != nil || len(navHistory) < 2 {
		// Fallback: use equal weight-based contribution
		for _, h := range holdings {
			weight := (h.quantity * h.avgCost) / totalValue
			contribution[h.symbol] = weight * 100
		}
		return contribution, nil
	}

	portfolioReturns := computeDailyReturns(navHistory)
	if len(portfolioReturns) == 0 {
		for _, h := range holdings {
			weight := (h.quantity * h.avgCost) / totalValue
			contribution[h.symbol] = weight * 100
		}
		return contribution, nil
	}

	// Compute portfolio variance for decomposition
	portfolioVar := variance(portfolioReturns)
	if portfolioVar == 0 {
		for _, h := range holdings {
			weight := (h.quantity * h.avgCost) / totalValue
			contribution[h.symbol] = weight * 100
		}
		return contribution, nil
	}

	// Use weight-squared × variance as a simplified risk contribution proxy.
	// In a full implementation, we'd compute covariance between each holding
	// and the portfolio, but that requires per-holding return series which
	// needs individual price histories.
	var totalRiskContrib float64
	riskContribs := make(map[string]float64)
	for _, h := range holdings {
		weight := (h.quantity * h.avgCost) / totalValue
		// Marginal contribution ≈ weight² (simplified; assumes equal correlation)
		rc := weight * weight
		riskContribs[h.symbol] = rc
		totalRiskContrib += rc
	}

	// Normalize to percentages
	if totalRiskContrib > 0 {
		for sym, rc := range riskContribs {
			contribution[sym] = (rc / totalRiskContrib) * 100
		}
	}

	return contribution, nil
}

// --- Helper functions ---

// computeDailyReturns computes daily percentage returns from NAV snapshots.
func computeDailyReturns(navHistory []model.NAVSnapshot) []float64 {
	if len(navHistory) < 2 {
		return nil
	}

	var returns []float64
	for i := 1; i < len(navHistory); i++ {
		if navHistory[i-1].NAV > 0 {
			ret := (navHistory[i].NAV - navHistory[i-1].NAV) / navHistory[i-1].NAV
			returns = append(returns, ret)
		}
	}
	return returns
}

// computeAnnualizedVolatility computes annualized volatility from daily returns.
// Volatility = σ_daily × √252.
// Requirement 27.4.
func computeAnnualizedVolatility(dailyReturns []float64) float64 {
	if len(dailyReturns) == 0 {
		return 0
	}
	stdDev := math.Sqrt(variance(dailyReturns))
	return stdDev * math.Sqrt(TradingDaysPerYear)
}

// mean computes the arithmetic mean of a slice.
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// variance computes the population variance of a slice.
func variance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	m := mean(values)
	var sumSq float64
	for _, v := range values {
		diff := v - m
		sumSq += diff * diff
	}
	return sumSq / float64(len(values))
}
