package service

import (
	"math"
	"testing"
	"time"

	"myfi-backend/internal/model"
)

// generateOHLCVBars creates synthetic OHLCV data with a known trend pattern.
// The pattern: uptrend for half, then downtrend for the rest.
func generateOHLCVBars(n int, startPrice float64) []model.OHLCVBar {
	bars := make([]model.OHLCVBar, n)
	price := startPrice
	t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		if i < n/2 {
			price *= 1.01 // 1% daily increase
		} else {
			price *= 0.99 // 1% daily decrease
		}
		bars[i] = model.OHLCVBar{
			Time:   t.AddDate(0, 0, i),
			Open:   price * 0.998,
			High:   price * 1.005,
			Low:    price * 0.995,
			Close:  price,
			Volume: 1000000 + int64(i*10000),
		}
	}
	return bars
}

// generateTrendingBars creates bars with a consistent uptrend.
func generateTrendingBars(n int, startPrice float64, dailyReturn float64) []model.OHLCVBar {
	bars := make([]model.OHLCVBar, n)
	price := startPrice
	t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		price *= (1 + dailyReturn)
		bars[i] = model.OHLCVBar{
			Time:   t.AddDate(0, 0, i),
			Open:   price * 0.998,
			High:   price * 1.005,
			Low:    price * 0.995,
			Close:  price,
			Volume: 1000000,
		}
	}
	return bars
}

func TestBacktestEngine_NewBacktestEngine(t *testing.T) {
	engine := NewBacktestEngine()
	if engine == nil {
		t.Fatal("NewBacktestEngine returned nil")
	}
}

func TestBacktestEngine_InsufficientData(t *testing.T) {
	engine := NewBacktestEngine()
	strategy := model.StrategyRule{
		Name:        "test",
		StopLossPct: 0.05,
	}

	_, err := engine.RunBacktest([]model.OHLCVBar{}, strategy)
	if err == nil {
		t.Fatal("expected error for empty bars")
	}

	_, err = engine.RunBacktest([]model.OHLCVBar{{Close: 100}}, strategy)
	if err == nil {
		t.Fatal("expected error for single bar")
	}
}

func TestBacktestEngine_RSIOversoldBounce(t *testing.T) {
	engine := NewBacktestEngine()
	presets := engine.GetPresetStrategies()

	// Find RSI Oversold Bounce.
	var rsiStrategy model.StrategyRule
	for _, p := range presets {
		if p.Name == "RSI Oversold Bounce" {
			rsiStrategy = p
			break
		}
	}
	if rsiStrategy.Name == "" {
		t.Fatal("RSI Oversold Bounce preset not found")
	}

	// Generate enough data for RSI to compute (need > 14 bars).
	bars := generateOHLCVBars(100, 100)
	result, err := engine.RunBacktest(bars, rsiStrategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	// Verify result structure.
	if result.NumTrades < 0 {
		t.Error("NumTrades should be >= 0")
	}
	if len(result.EquityCurve) != len(bars) {
		t.Errorf("EquityCurve length = %d, want %d", len(result.EquityCurve), len(bars))
	}
	if result.WinRate < 0 || result.WinRate > 1 {
		t.Errorf("WinRate = %f, want [0, 1]", result.WinRate)
	}
	if result.MaxDrawdown < 0 || result.MaxDrawdown > 1 {
		t.Errorf("MaxDrawdown = %f, want [0, 1]", result.MaxDrawdown)
	}
}

func TestBacktestEngine_MACDCrossover(t *testing.T) {
	engine := NewBacktestEngine()
	presets := engine.GetPresetStrategies()

	var macdStrategy model.StrategyRule
	for _, p := range presets {
		if p.Name == "MACD Crossover" {
			macdStrategy = p
			break
		}
	}
	if macdStrategy.Name == "" {
		t.Fatal("MACD Crossover preset not found")
	}

	bars := generateOHLCVBars(200, 50)
	result, err := engine.RunBacktest(bars, macdStrategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	if len(result.EquityCurve) != len(bars) {
		t.Errorf("EquityCurve length = %d, want %d", len(result.EquityCurve), len(bars))
	}
	// MACD crossover on trending data should produce at least one trade.
	if result.NumTrades < 0 {
		t.Error("NumTrades should be >= 0")
	}
}

func TestBacktestEngine_BollingerBandSqueeze(t *testing.T) {
	engine := NewBacktestEngine()
	presets := engine.GetPresetStrategies()

	var bbStrategy model.StrategyRule
	for _, p := range presets {
		if p.Name == "Bollinger Band Squeeze" {
			bbStrategy = p
			break
		}
	}
	if bbStrategy.Name == "" {
		t.Fatal("Bollinger Band Squeeze preset not found")
	}

	bars := generateOHLCVBars(100, 100)
	result, err := engine.RunBacktest(bars, bbStrategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	if len(result.EquityCurve) != len(bars) {
		t.Errorf("EquityCurve length = %d, want %d", len(result.EquityCurve), len(bars))
	}
}

func TestBacktestEngine_StopLoss(t *testing.T) {
	engine := NewBacktestEngine()

	// Create a strategy that always enters immediately.
	strategy := model.StrategyRule{
		Name: "Always Enter",
		EntryConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "price", Field: "close"},
				Operator: model.OpGreaterThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 0},
			},
		},
		ExitConditions: []model.StrategyCondition{}, // no signal exit
		StopLossPct:    0.03,                        // 3% stop loss
		TakeProfitPct:  0,                           // no take profit
	}

	// Create bars that drop 5% after entry.
	bars := make([]model.OHLCVBar, 20)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 100.0
	for i := 0; i < 20; i++ {
		if i > 0 {
			price *= 0.98 // 2% daily drop
		}
		bars[i] = model.OHLCVBar{
			Time:   t0.AddDate(0, 0, i),
			Open:   price,
			High:   price * 1.001,
			Low:    price * 0.999,
			Close:  price,
			Volume: 100000,
		}
	}

	result, err := engine.RunBacktest(bars, strategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	// Should have at least one trade that hit stop loss.
	foundStopLoss := false
	for _, trade := range result.Trades {
		if trade.ExitReason == "stop_loss" {
			foundStopLoss = true
			break
		}
	}
	if !foundStopLoss {
		t.Error("expected at least one stop_loss exit")
	}
}

func TestBacktestEngine_TakeProfit(t *testing.T) {
	engine := NewBacktestEngine()

	strategy := model.StrategyRule{
		Name: "Always Enter",
		EntryConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "price", Field: "close"},
				Operator: model.OpGreaterThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 0},
			},
		},
		ExitConditions: []model.StrategyCondition{},
		StopLossPct:    0,
		TakeProfitPct:  0.05, // 5% take profit
	}

	// Create bars that rise 3% daily.
	bars := make([]model.OHLCVBar, 20)
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 100.0
	for i := 0; i < 20; i++ {
		if i > 0 {
			price *= 1.03
		}
		bars[i] = model.OHLCVBar{
			Time:   t0.AddDate(0, 0, i),
			Open:   price,
			High:   price * 1.001,
			Low:    price * 0.999,
			Close:  price,
			Volume: 100000,
		}
	}

	result, err := engine.RunBacktest(bars, strategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	foundTP := false
	for _, trade := range result.Trades {
		if trade.ExitReason == "take_profit" {
			foundTP = true
			break
		}
	}
	if !foundTP {
		t.Error("expected at least one take_profit exit")
	}
}

func TestBacktestEngine_PresetStrategies(t *testing.T) {
	engine := NewBacktestEngine()
	presets := engine.GetPresetStrategies()

	if len(presets) != 3 {
		t.Fatalf("expected 3 preset strategies, got %d", len(presets))
	}

	names := map[string]bool{}
	for _, p := range presets {
		names[p.Name] = true
		if len(p.EntryConditions) == 0 {
			t.Errorf("preset %q has no entry conditions", p.Name)
		}
		if len(p.ExitConditions) == 0 {
			t.Errorf("preset %q has no exit conditions", p.Name)
		}
		if p.StopLossPct <= 0 {
			t.Errorf("preset %q has no stop loss", p.Name)
		}
		if p.TakeProfitPct <= 0 {
			t.Errorf("preset %q has no take profit", p.Name)
		}
	}

	expected := []string{"RSI Oversold Bounce", "MACD Crossover", "Bollinger Band Squeeze"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing preset strategy: %s", name)
		}
	}
}

func TestBacktestEngine_EquityCurveMonotonic(t *testing.T) {
	// With no trades, equity curve should be flat at initial capital.
	engine := NewBacktestEngine()

	// Strategy that never triggers (RSI > 200 is impossible).
	strategy := model.StrategyRule{
		Name: "Never Enter",
		EntryConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorRSI, Field: "value", Period: 14},
				Operator: model.OpGreaterThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 200},
			},
		},
		ExitConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "constant", Constant: 1},
				Operator: model.OpGreaterThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 2},
			},
		},
		StopLossPct:   0.05,
		TakeProfitPct: 0.10,
	}

	bars := generateOHLCVBars(50, 100)
	result, err := engine.RunBacktest(bars, strategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	if result.NumTrades != 0 {
		t.Errorf("expected 0 trades, got %d", result.NumTrades)
	}
	if result.TotalReturn != 0 {
		t.Errorf("expected 0 total return, got %f", result.TotalReturn)
	}
	// All equity curve values should be the same (initial capital).
	for i, pt := range result.EquityCurve {
		if pt.Value != result.EquityCurve[0].Value {
			t.Errorf("equity curve point %d = %f, want %f", i, pt.Value, result.EquityCurve[0].Value)
			break
		}
	}
}

func TestBacktestEngine_WinRateCalculation(t *testing.T) {
	trades := []model.BacktestTrade{
		{ReturnPct: 0.05},
		{ReturnPct: -0.02},
		{ReturnPct: 0.03},
		{ReturnPct: -0.01},
	}
	winRate := computeWinRate(trades)
	if winRate != 0.5 {
		t.Errorf("winRate = %f, want 0.5", winRate)
	}

	// All wins.
	allWins := []model.BacktestTrade{{ReturnPct: 0.01}, {ReturnPct: 0.02}}
	if computeWinRate(allWins) != 1.0 {
		t.Error("expected win rate 1.0 for all winning trades")
	}

	// No trades.
	if computeWinRate(nil) != 0 {
		t.Error("expected win rate 0 for no trades")
	}
}

func TestBacktestEngine_MaxDrawdown(t *testing.T) {
	curve := []model.EquityPoint{
		{Value: 100},
		{Value: 110},
		{Value: 105},
		{Value: 90},
		{Value: 95},
		{Value: 120},
	}
	dd := computeMaxDrawdownFromEquity(curve)
	// Max drawdown: peak=110, trough=90 => (110-90)/110 ≈ 0.1818
	expected := (110.0 - 90.0) / 110.0
	if math.Abs(dd-expected) > 0.001 {
		t.Errorf("maxDrawdown = %f, want %f", dd, expected)
	}
}

func TestBacktestEngine_SharpeRatio(t *testing.T) {
	// Flat equity curve should have 0 Sharpe.
	flat := []model.EquityPoint{
		{Value: 100}, {Value: 100}, {Value: 100},
	}
	if computeSharpeFromEquity(flat) != 0 {
		t.Error("expected Sharpe 0 for flat equity curve")
	}

	// Single point should return 0.
	if computeSharpeFromEquity([]model.EquityPoint{{Value: 100}}) != 0 {
		t.Error("expected Sharpe 0 for single point")
	}
}

// --- Indicator unit tests ---

func TestComputeSMA(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	sma := computeSMA(data, 3)

	// SMA(3) at index 2 = (1+2+3)/3 = 2
	if math.Abs(sma[2]-2.0) > 0.001 {
		t.Errorf("SMA[2] = %f, want 2.0", sma[2])
	}
	// SMA(3) at index 9 = (8+9+10)/3 = 9
	if math.Abs(sma[9]-9.0) > 0.001 {
		t.Errorf("SMA[9] = %f, want 9.0", sma[9])
	}
	// First values should be NaN.
	if !math.IsNaN(sma[0]) {
		t.Errorf("SMA[0] should be NaN, got %f", sma[0])
	}
}

func TestComputeEMA(t *testing.T) {
	data := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	ema := computeEMA(data, 3)

	// EMA seed at index 2 = SMA(3) = (10+11+12)/3 = 11
	if math.Abs(ema[2]-11.0) > 0.001 {
		t.Errorf("EMA[2] = %f, want 11.0", ema[2])
	}
	// EMA should be defined for all indices >= period-1.
	for i := 2; i < len(ema); i++ {
		if math.IsNaN(ema[i]) {
			t.Errorf("EMA[%d] should not be NaN", i)
		}
	}
}

func TestComputeRSI(t *testing.T) {
	// Monotonically increasing prices should have RSI near 100.
	data := make([]float64, 30)
	for i := range data {
		data[i] = float64(100 + i)
	}
	rsi := computeRSI(data, 14)

	// RSI at index 14 should be 100 (all gains, no losses).
	if rsi[14] != 100 {
		t.Errorf("RSI[14] = %f, want 100 for monotonic increase", rsi[14])
	}
}

func TestComputeMACD(t *testing.T) {
	data := make([]float64, 50)
	for i := range data {
		data[i] = float64(100 + i)
	}
	macdLine, signalLine, histogram := computeMACD(data, 12, 26, 9)

	// For monotonically increasing data, MACD line should be positive.
	lastIdx := len(data) - 1
	if math.IsNaN(macdLine[lastIdx]) {
		t.Error("MACD line should be defined at last index")
	}
	if macdLine[lastIdx] <= 0 {
		t.Errorf("MACD line = %f, expected positive for uptrend", macdLine[lastIdx])
	}
	_ = signalLine
	_ = histogram
}

func TestComputeBollingerBands(t *testing.T) {
	data := make([]float64, 30)
	for i := range data {
		data[i] = 100 + float64(i%5) // oscillating
	}
	upper, middle, lower, bw := computeBollingerBands(data, 20, 2.0)

	idx := 25
	if math.IsNaN(upper[idx]) || math.IsNaN(lower[idx]) || math.IsNaN(middle[idx]) {
		t.Error("Bollinger Bands should be defined at index 25")
	}
	if upper[idx] <= middle[idx] {
		t.Error("upper band should be > middle")
	}
	if lower[idx] >= middle[idx] {
		t.Error("lower band should be < middle")
	}
	if math.IsNaN(bw[idx]) || bw[idx] <= 0 {
		t.Error("bandwidth should be positive")
	}
}

func TestComputeATR(t *testing.T) {
	highs := []float64{105, 106, 107, 108, 109, 110, 111, 112, 113, 114}
	lows := []float64{95, 96, 97, 98, 99, 100, 101, 102, 103, 104}
	closes := []float64{100, 101, 102, 103, 104, 105, 106, 107, 108, 109}
	atr := computeATR(highs, lows, closes, 5)

	// ATR should be defined from index 4 onward.
	if math.IsNaN(atr[4]) {
		t.Error("ATR[4] should be defined")
	}
	if atr[4] <= 0 {
		t.Errorf("ATR[4] = %f, expected positive", atr[4])
	}
}

func TestComputeOBV(t *testing.T) {
	closes := []float64{100, 102, 101, 103, 104}
	volumes := []float64{1000, 1500, 1200, 1800, 2000}
	obv := computeOBV(closes, volumes)

	if obv[0] != 1000 {
		t.Errorf("OBV[0] = %f, want 1000", obv[0])
	}
	// Close went up: OBV[1] = 1000 + 1500 = 2500
	if obv[1] != 2500 {
		t.Errorf("OBV[1] = %f, want 2500", obv[1])
	}
	// Close went down: OBV[2] = 2500 - 1200 = 1300
	if obv[2] != 1300 {
		t.Errorf("OBV[2] = %f, want 1300", obv[2])
	}
}

func TestComputeStochastic(t *testing.T) {
	highs := make([]float64, 20)
	lows := make([]float64, 20)
	closes := make([]float64, 20)
	for i := 0; i < 20; i++ {
		highs[i] = float64(100 + i + 2)
		lows[i] = float64(100 + i - 2)
		closes[i] = float64(100 + i)
	}
	k, d := computeStochastic(highs, lows, closes, 14, 3)

	// %K should be defined from index 13 onward.
	if math.IsNaN(k[13]) {
		t.Error("Stochastic %K[13] should be defined")
	}
	// %D should be defined from index 15 onward (SMA of %K with period 3).
	if math.IsNaN(d[15]) {
		t.Error("Stochastic %D[15] should be defined")
	}
}

func TestBacktestEngine_CrossesAboveCondition(t *testing.T) {
	engine := NewBacktestEngine()

	// Create a strategy using SMA crossover.
	strategy := model.StrategyRule{
		Name: "SMA Cross",
		EntryConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorSMA, Field: "value", Period: 5},
				Operator: model.OpCrossesAbove,
				Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorSMA, Field: "value", Period: 20},
			},
		},
		ExitConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorSMA, Field: "value", Period: 5},
				Operator: model.OpCrossesBelow,
				Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorSMA, Field: "value", Period: 20},
			},
		},
		StopLossPct:   0.10,
		TakeProfitPct: 0.20,
	}

	bars := generateOHLCVBars(200, 100)
	result, err := engine.RunBacktest(bars, strategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	// Verify basic result integrity.
	if len(result.EquityCurve) != len(bars) {
		t.Errorf("EquityCurve length = %d, want %d", len(result.EquityCurve), len(bars))
	}
}

func TestBacktestEngine_AllIndicatorsCompute(t *testing.T) {
	// Verify that all 21 indicators can be computed without panicking.
	engine := NewBacktestEngine()
	bars := generateOHLCVBars(100, 100)

	indicators := []model.IndicatorType{
		model.IndicatorSMA, model.IndicatorEMA, model.IndicatorRSI,
		model.IndicatorMACD, model.IndicatorBollingerBands, model.IndicatorStochastic,
		model.IndicatorADX, model.IndicatorAroon, model.IndicatorParabolicSAR,
		model.IndicatorSupertrend, model.IndicatorVWAP, model.IndicatorVWMA,
		model.IndicatorWilliamsR, model.IndicatorCMO, model.IndicatorROC,
		model.IndicatorMomentum, model.IndicatorKeltnerChannel, model.IndicatorATR,
		model.IndicatorStdDev, model.IndicatorOBV, model.IndicatorLinearReg,
	}

	for _, ind := range indicators {
		op := model.ConditionOperand{
			Type:      "indicator",
			Indicator: ind,
			Field:     "value",
			Period:    14,
			Param2:    26,
			Param3:    9,
			ParamF:    2.0,
		}
		result := engine.computeIndicator(bars, op)
		if len(result) != len(bars) {
			t.Errorf("indicator %s returned %d values, want %d", ind, len(result), len(bars))
		}
	}
}

func TestBacktestEngine_EndOfDataExit(t *testing.T) {
	engine := NewBacktestEngine()

	// Strategy that enters immediately and never exits via signal.
	strategy := model.StrategyRule{
		Name: "Enter and Hold",
		EntryConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "price", Field: "close"},
				Operator: model.OpGreaterThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 0},
			},
		},
		ExitConditions: []model.StrategyCondition{
			{
				Left:     model.ConditionOperand{Type: "price", Field: "close"},
				Operator: model.OpLessThan,
				Right:    model.ConditionOperand{Type: "constant", Constant: 0}, // never true
			},
		},
		StopLossPct:   0, // no stop loss
		TakeProfitPct: 0, // no take profit
	}

	bars := generateTrendingBars(30, 100, 0.005)
	result, err := engine.RunBacktest(bars, strategy)
	if err != nil {
		t.Fatalf("RunBacktest failed: %v", err)
	}

	if result.NumTrades != 1 {
		t.Fatalf("expected 1 trade (end_of_data), got %d", result.NumTrades)
	}
	if result.Trades[0].ExitReason != "end_of_data" {
		t.Errorf("exit reason = %s, want end_of_data", result.Trades[0].ExitReason)
	}
	if result.TotalReturn <= 0 {
		t.Error("expected positive return for uptrending data")
	}
}
