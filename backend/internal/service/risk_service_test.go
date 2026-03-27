package service

import (
	"math"
	"testing"
	"time"

	"myfi-backend/internal/model"
)

// --- ComputeSharpeRatio tests (Req 27.1) ---

func TestComputeSharpeRatio_PositiveReturns(t *testing.T) {
	rs := &RiskService{}
	// Daily returns averaging ~0.04% → annualized ~10.08%
	// With 4.5% risk-free rate, Sharpe should be positive
	dailyReturns := make([]float64, 252)
	for i := range dailyReturns {
		dailyReturns[i] = 0.0004 // 0.04% daily
	}

	sharpe := rs.ComputeSharpeRatio(dailyReturns, DefaultVNRiskFreeRate)
	if sharpe <= 0 {
		t.Errorf("expected positive Sharpe ratio for positive excess returns, got %f", sharpe)
	}
}

func TestComputeSharpeRatio_ZeroVolatility(t *testing.T) {
	rs := &RiskService{}
	// All returns identical → zero volatility → Sharpe = 0
	dailyReturns := []float64{0.001, 0.001, 0.001, 0.001, 0.001}
	sharpe := rs.ComputeSharpeRatio(dailyReturns, DefaultVNRiskFreeRate)
	if sharpe != 0 {
		t.Errorf("expected Sharpe 0 for zero volatility, got %f", sharpe)
	}
}

func TestComputeSharpeRatio_EmptyReturns(t *testing.T) {
	rs := &RiskService{}
	sharpe := rs.ComputeSharpeRatio(nil, DefaultVNRiskFreeRate)
	if sharpe != 0 {
		t.Errorf("expected Sharpe 0 for empty returns, got %f", sharpe)
	}
}

func TestComputeSharpeRatio_NegativeExcessReturn(t *testing.T) {
	rs := &RiskService{}
	// Daily returns averaging ~0.005% → annualized ~1.26%, below 4.5% risk-free
	dailyReturns := []float64{0.0001, -0.0001, 0.0002, -0.0002, 0.0001, -0.0001, 0.0002, -0.0002, 0.0001, 0.0001}
	sharpe := rs.ComputeSharpeRatio(dailyReturns, DefaultVNRiskFreeRate)
	if sharpe >= 0 {
		t.Errorf("expected negative Sharpe for returns below risk-free rate, got %f", sharpe)
	}
}

func TestComputeSharpeRatio_DefaultRiskFreeRate(t *testing.T) {
	// Verify the default VN risk-free rate is 4.5%
	if math.Abs(DefaultVNRiskFreeRate-0.045) > 1e-10 {
		t.Errorf("expected default risk-free rate 0.045, got %f", DefaultVNRiskFreeRate)
	}
}

// --- ComputeMaxDrawdown tests (Req 27.2) ---

func TestComputeMaxDrawdown_SimpleDecline(t *testing.T) {
	rs := &RiskService{}
	navHistory := []model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), NAV: 90},
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), NAV: 95},
	}

	dd := rs.ComputeMaxDrawdown(navHistory)
	// Peak 100, trough 90 → drawdown = 10%
	expected := 0.10
	if math.Abs(dd-expected) > 0.001 {
		t.Errorf("expected max drawdown %f, got %f", expected, dd)
	}
}

func TestComputeMaxDrawdown_MultiplePeaks(t *testing.T) {
	rs := &RiskService{}
	navHistory := []model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), NAV: 95},  // -5%
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), NAV: 110}, // new peak
		{Date: time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC), NAV: 88},  // -20% from 110
		{Date: time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC), NAV: 105},
	}

	dd := rs.ComputeMaxDrawdown(navHistory)
	// Peak 110, trough 88 → drawdown = 22/110 = 0.2
	expected := (110.0 - 88.0) / 110.0
	if math.Abs(dd-expected) > 0.001 {
		t.Errorf("expected max drawdown %f, got %f", expected, dd)
	}
}

func TestComputeMaxDrawdown_NoDrawdown(t *testing.T) {
	rs := &RiskService{}
	navHistory := []model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), NAV: 110},
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), NAV: 120},
	}

	dd := rs.ComputeMaxDrawdown(navHistory)
	if dd != 0 {
		t.Errorf("expected max drawdown 0 for monotonically increasing NAV, got %f", dd)
	}
}

func TestComputeMaxDrawdown_InsufficientData(t *testing.T) {
	rs := &RiskService{}
	dd := rs.ComputeMaxDrawdown([]model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
	})
	if dd != 0 {
		t.Errorf("expected max drawdown 0 for single data point, got %f", dd)
	}
}

// --- ComputeBeta tests (Req 27.3) ---

func TestComputeBeta_PerfectCorrelation(t *testing.T) {
	rs := &RiskService{}
	// Portfolio returns = 2× benchmark returns → beta should be ~2.0
	benchmark := []float64{0.01, -0.005, 0.02, -0.01, 0.015}
	portfolio := make([]float64, len(benchmark))
	for i, b := range benchmark {
		portfolio[i] = b * 2
	}

	beta := rs.ComputeBeta(portfolio, benchmark)
	if math.Abs(beta-2.0) > 0.01 {
		t.Errorf("expected beta ~2.0, got %f", beta)
	}
}

func TestComputeBeta_MarketNeutral(t *testing.T) {
	rs := &RiskService{}
	// Portfolio returns uncorrelated with benchmark → beta ~0
	benchmark := []float64{0.01, -0.01, 0.01, -0.01, 0.01, -0.01, 0.01, -0.01}
	portfolio := []float64{0.005, 0.005, 0.005, 0.005, 0.005, 0.005, 0.005, 0.005}

	beta := rs.ComputeBeta(portfolio, benchmark)
	if math.Abs(beta) > 0.01 {
		t.Errorf("expected beta ~0 for uncorrelated returns, got %f", beta)
	}
}

func TestComputeBeta_InsufficientData(t *testing.T) {
	rs := &RiskService{}
	beta := rs.ComputeBeta([]float64{0.01}, []float64{0.02})
	if beta != 0 {
		t.Errorf("expected beta 0 for insufficient data, got %f", beta)
	}
}

func TestComputeBeta_DifferentLengths(t *testing.T) {
	rs := &RiskService{}
	// Different lengths — should align to shorter
	benchmark := []float64{0.01, -0.005, 0.02}
	portfolio := []float64{0.02, -0.01, 0.04, 0.01, -0.005}

	beta := rs.ComputeBeta(portfolio, benchmark)
	// Should use last 3 from portfolio, all 3 from benchmark
	if math.IsNaN(beta) || math.IsInf(beta, 0) {
		t.Errorf("expected valid beta for different length inputs, got %f", beta)
	}
}

func TestComputeBeta_ZeroBenchmarkVariance(t *testing.T) {
	rs := &RiskService{}
	// All benchmark returns identical → zero variance → beta = 0
	benchmark := []float64{0.01, 0.01, 0.01, 0.01}
	portfolio := []float64{0.02, -0.01, 0.03, 0.005}

	beta := rs.ComputeBeta(portfolio, benchmark)
	if beta != 0 {
		t.Errorf("expected beta 0 for zero benchmark variance, got %f", beta)
	}
}

// --- Annualized Volatility tests (Req 27.4) ---

func TestComputeAnnualizedVolatility_KnownValues(t *testing.T) {
	// Daily returns with known std dev
	dailyReturns := []float64{0.01, -0.01, 0.01, -0.01}
	// Mean = 0, variance = 0.0001, stddev = 0.01
	// Annualized = 0.01 × √252 ≈ 0.1587

	vol := computeAnnualizedVolatility(dailyReturns)
	expected := 0.01 * math.Sqrt(252)
	if math.Abs(vol-expected) > 0.001 {
		t.Errorf("expected annualized volatility %f, got %f", expected, vol)
	}
}

func TestComputeAnnualizedVolatility_ZeroReturns(t *testing.T) {
	dailyReturns := []float64{0, 0, 0, 0}
	vol := computeAnnualizedVolatility(dailyReturns)
	if vol != 0 {
		t.Errorf("expected volatility 0 for zero returns, got %f", vol)
	}
}

func TestComputeAnnualizedVolatility_Empty(t *testing.T) {
	vol := computeAnnualizedVolatility(nil)
	if vol != 0 {
		t.Errorf("expected volatility 0 for empty returns, got %f", vol)
	}
}

func TestComputeAnnualizedVolatility_UsesSquareRoot252(t *testing.T) {
	if TradingDaysPerYear != 252 {
		t.Errorf("expected TradingDaysPerYear = 252, got %d", TradingDaysPerYear)
	}
}

// --- ComputeVaR tests (Req 27.5) ---

func TestComputeVaR_95Confidence(t *testing.T) {
	rs := &RiskService{}
	// 20 returns: sorted, 5th percentile index = floor(0.05 * 20) = 1
	dailyReturns := []float64{
		-0.05, -0.03, -0.02, -0.01, -0.005,
		0.001, 0.002, 0.003, 0.004, 0.005,
		0.006, 0.007, 0.008, 0.009, 0.01,
		0.011, 0.012, 0.013, 0.014, 0.015,
	}
	currentNAV := 100_000_000.0

	var95 := rs.ComputeVaR(dailyReturns, 0.95, currentNAV)
	// Index 1 (0-based) in sorted = -0.03, VaR = 0.03 × 100M = 3M
	expected := 0.03 * currentNAV
	if math.Abs(var95-expected) > 1 {
		t.Errorf("expected VaR %f, got %f", expected, var95)
	}
}

func TestComputeVaR_EmptyReturns(t *testing.T) {
	rs := &RiskService{}
	var95 := rs.ComputeVaR(nil, 0.95, 100_000_000)
	if var95 != 0 {
		t.Errorf("expected VaR 0 for empty returns, got %f", var95)
	}
}

func TestComputeVaR_ZeroNAV(t *testing.T) {
	rs := &RiskService{}
	var95 := rs.ComputeVaR([]float64{-0.01, 0.01}, 0.95, 0)
	if var95 != 0 {
		t.Errorf("expected VaR 0 for zero NAV, got %f", var95)
	}
}

func TestComputeVaR_AllPositiveReturns(t *testing.T) {
	rs := &RiskService{}
	// All positive returns — VaR should still be computed from the lowest
	dailyReturns := []float64{0.01, 0.02, 0.03, 0.04, 0.05}
	currentNAV := 100_000_000.0

	var95 := rs.ComputeVaR(dailyReturns, 0.95, currentNAV)
	// 5th percentile of all positive returns → smallest positive return
	// VaR = |0.01| × 100M = 1M
	expected := 0.01 * currentNAV
	if math.Abs(var95-expected) > 1 {
		t.Errorf("expected VaR %f, got %f", expected, var95)
	}
}

// --- Helper function tests ---

func TestComputeDailyReturns_Basic(t *testing.T) {
	navHistory := []model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), NAV: 110},
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), NAV: 105},
	}

	returns := computeDailyReturns(navHistory)
	if len(returns) != 2 {
		t.Fatalf("expected 2 returns, got %d", len(returns))
	}

	// First return: (110-100)/100 = 0.10
	if math.Abs(returns[0]-0.10) > 0.001 {
		t.Errorf("expected first return 0.10, got %f", returns[0])
	}
	// Second return: (105-110)/110 ≈ -0.04545
	expected := (105.0 - 110.0) / 110.0
	if math.Abs(returns[1]-expected) > 0.001 {
		t.Errorf("expected second return %f, got %f", expected, returns[1])
	}
}

func TestComputeDailyReturns_InsufficientData(t *testing.T) {
	returns := computeDailyReturns([]model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 100},
	})
	if returns != nil {
		t.Errorf("expected nil returns for single data point, got %v", returns)
	}
}

func TestComputeDailyReturns_SkipsZeroNAV(t *testing.T) {
	navHistory := []model.NAVSnapshot{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), NAV: 0},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), NAV: 100},
		{Date: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC), NAV: 110},
	}

	returns := computeDailyReturns(navHistory)
	// First pair (0→100) should be skipped, only 100→110 computed
	if len(returns) != 1 {
		t.Fatalf("expected 1 return (skipping zero NAV), got %d", len(returns))
	}
	if math.Abs(returns[0]-0.10) > 0.001 {
		t.Errorf("expected return 0.10, got %f", returns[0])
	}
}

func TestMean_Basic(t *testing.T) {
	result := mean([]float64{1, 2, 3, 4, 5})
	if math.Abs(result-3.0) > 1e-10 {
		t.Errorf("expected mean 3.0, got %f", result)
	}
}

func TestMean_Empty(t *testing.T) {
	result := mean(nil)
	if result != 0 {
		t.Errorf("expected mean 0 for empty slice, got %f", result)
	}
}

func TestVariance_Basic(t *testing.T) {
	// Values: 1, 2, 3, 4, 5. Mean = 3.
	// Variance = ((1-3)² + (2-3)² + (3-3)² + (4-3)² + (5-3)²) / 5 = 10/5 = 2
	result := variance([]float64{1, 2, 3, 4, 5})
	if math.Abs(result-2.0) > 1e-10 {
		t.Errorf("expected variance 2.0, got %f", result)
	}
}

func TestVariance_Empty(t *testing.T) {
	result := variance(nil)
	if result != 0 {
		t.Errorf("expected variance 0 for empty slice, got %f", result)
	}
}

// --- NewRiskService constructor test ---

func TestNewRiskService_ReturnsValidInstance(t *testing.T) {
	rs := NewRiskService(nil, nil, nil)
	if rs == nil {
		t.Fatal("expected non-nil RiskService")
	}
}
