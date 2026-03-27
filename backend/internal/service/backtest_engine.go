package service

import (
	"fmt"
	"math"
	"sort"

	"myfi-backend/internal/model"
)

// BacktestEngine runs indicator-based trading strategies against historical OHLCV data.
// Requirement 32: Backtesting and Strategy Simulation.
type BacktestEngine struct{}

// NewBacktestEngine creates a new BacktestEngine instance.
func NewBacktestEngine() *BacktestEngine {
	return &BacktestEngine{}
}

// RunBacktest executes a strategy against the provided OHLCV bars and returns results.
func (b *BacktestEngine) RunBacktest(bars []model.OHLCVBar, strategy model.StrategyRule) (model.BacktestResult, error) {
	if len(bars) < 2 {
		return model.BacktestResult{}, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(bars))
	}

	// Sort bars by time ascending.
	sorted := make([]model.OHLCVBar, len(bars))
	copy(sorted, bars)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	// Pre-compute all indicator values needed by the strategy.
	indicators := b.computeAllIndicators(sorted, strategy)

	// Simulate trades.
	var trades []model.BacktestTrade
	inPosition := false
	var entryPrice float64
	var entryIdx int

	initialCapital := 1000000.0 // 1M VND notional
	capital := initialCapital
	equityCurve := make([]model.EquityPoint, 0, len(sorted))

	for i := 0; i < len(sorted); i++ {
		bar := sorted[i]
		currentValue := capital
		if inPosition {
			currentValue = capital * (bar.Close / entryPrice)
		}
		equityCurve = append(equityCurve, model.EquityPoint{Date: bar.Time, Value: currentValue})

		if !inPosition {
			// Check entry conditions.
			if b.evaluateConditions(strategy.EntryConditions, sorted, indicators, i) {
				inPosition = true
				entryPrice = bar.Close
				entryIdx = i
			}
		} else {
			// Check stop-loss and take-profit first.
			returnPct := (bar.Close - entryPrice) / entryPrice
			exitReason := ""

			if strategy.StopLossPct > 0 && returnPct <= -strategy.StopLossPct {
				exitReason = "stop_loss"
			} else if strategy.TakeProfitPct > 0 && returnPct >= strategy.TakeProfitPct {
				exitReason = "take_profit"
			} else if b.evaluateConditions(strategy.ExitConditions, sorted, indicators, i) {
				exitReason = "signal"
			}

			if exitReason != "" {
				holdingDays := int(bar.Time.Sub(sorted[entryIdx].Time).Hours() / 24)
				if holdingDays < 1 {
					holdingDays = 1
				}
				trade := model.BacktestTrade{
					EntryDate:   sorted[entryIdx].Time,
					ExitDate:    bar.Time,
					EntryPrice:  entryPrice,
					ExitPrice:   bar.Close,
					ReturnPct:   returnPct,
					ExitReason:  exitReason,
					HoldingDays: holdingDays,
				}
				trades = append(trades, trade)
				capital = capital * (1 + returnPct)
				inPosition = false
			}
		}
	}

	// Close any open position at end of data.
	if inPosition {
		lastBar := sorted[len(sorted)-1]
		returnPct := (lastBar.Close - entryPrice) / entryPrice
		holdingDays := int(lastBar.Time.Sub(sorted[entryIdx].Time).Hours() / 24)
		if holdingDays < 1 {
			holdingDays = 1
		}
		trades = append(trades, model.BacktestTrade{
			EntryDate:   sorted[entryIdx].Time,
			ExitDate:    lastBar.Time,
			EntryPrice:  entryPrice,
			ExitPrice:   lastBar.Close,
			ReturnPct:   returnPct,
			ExitReason:  "end_of_data",
			HoldingDays: holdingDays,
		})
		capital = capital * (1 + returnPct)
	}

	// Update last equity point with final capital.
	if len(equityCurve) > 0 {
		equityCurve[len(equityCurve)-1].Value = capital
	}

	result := model.BacktestResult{
		TotalReturn:    (capital - initialCapital) / initialCapital,
		NumTrades:      len(trades),
		EquityCurve:    equityCurve,
		Trades:         trades,
		WinRate:        computeWinRate(trades),
		MaxDrawdown:    computeMaxDrawdownFromEquity(equityCurve),
		SharpeRatio:    computeSharpeFromEquity(equityCurve),
		AvgHoldingDays: computeAvgHoldingDays(trades),
	}
	return result, nil
}

// GetPresetStrategies returns the built-in preset strategies.
// Requirement 32.6: RSI Oversold Bounce, MACD Crossover, Bollinger Band Squeeze.
func (b *BacktestEngine) GetPresetStrategies() []model.StrategyRule {
	return []model.StrategyRule{
		{
			Name: "RSI Oversold Bounce",
			EntryConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorRSI, Field: "value", Period: 14},
					Operator: model.OpLessThan,
					Right:    model.ConditionOperand{Type: "constant", Constant: 30},
				},
			},
			ExitConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorRSI, Field: "value", Period: 14},
					Operator: model.OpGreaterThan,
					Right:    model.ConditionOperand{Type: "constant", Constant: 70},
				},
			},
			StopLossPct:   0.05,
			TakeProfitPct: 0.10,
		},
		{
			Name: "MACD Crossover",
			EntryConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorMACD, Field: "macd", Period: 12, Param2: 26, Param3: 9},
					Operator: model.OpCrossesAbove,
					Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorMACD, Field: "signal", Period: 12, Param2: 26, Param3: 9},
				},
			},
			ExitConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorMACD, Field: "macd", Period: 12, Param2: 26, Param3: 9},
					Operator: model.OpCrossesBelow,
					Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorMACD, Field: "signal", Period: 12, Param2: 26, Param3: 9},
				},
			},
			StopLossPct:   0.05,
			TakeProfitPct: 0.15,
		},
		{
			Name: "Bollinger Band Squeeze",
			EntryConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "price", Field: "close"},
					Operator: model.OpLessThan,
					Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorBollingerBands, Field: "lower", Period: 20, ParamF: 2.0},
				},
			},
			ExitConditions: []model.StrategyCondition{
				{
					Left:     model.ConditionOperand{Type: "price", Field: "close"},
					Operator: model.OpGreaterThan,
					Right:    model.ConditionOperand{Type: "indicator", Indicator: model.IndicatorBollingerBands, Field: "upper", Period: 20, ParamF: 2.0},
				},
			},
			StopLossPct:   0.03,
			TakeProfitPct: 0.08,
		},
	}
}

// --- Indicator computation ---

// indicatorCache holds pre-computed indicator series keyed by a descriptor string.
type indicatorCache map[string][]float64

// computeAllIndicators pre-computes every indicator referenced in the strategy.
func (b *BacktestEngine) computeAllIndicators(bars []model.OHLCVBar, strategy model.StrategyRule) indicatorCache {
	cache := make(indicatorCache)
	allConds := append(strategy.EntryConditions, strategy.ExitConditions...)
	for _, cond := range allConds {
		for _, op := range []model.ConditionOperand{cond.Left, cond.Right} {
			if op.Type == "indicator" {
				key := indicatorKey(op)
				if _, ok := cache[key]; !ok {
					cache[key] = b.computeIndicator(bars, op)
				}
			}
		}
	}
	return cache
}

func indicatorKey(op model.ConditionOperand) string {
	return fmt.Sprintf("%s_%s_%d_%d_%d_%.2f", op.Indicator, op.Field, op.Period, op.Param2, op.Param3, op.ParamF)
}

func (b *BacktestEngine) computeIndicator(bars []model.OHLCVBar, op model.ConditionOperand) []float64 {
	closes := extractCloses(bars)
	highs := extractHighs(bars)
	lows := extractLows(bars)
	volumes := extractVolumes(bars)
	n := len(bars)
	period := op.Period
	if period < 1 {
		period = 14
	}

	switch op.Indicator {
	case model.IndicatorSMA:
		return computeSMA(closes, period)
	case model.IndicatorEMA:
		return computeEMA(closes, period)
	case model.IndicatorRSI:
		return computeRSI(closes, period)
	case model.IndicatorMACD:
		fast, slow, sig := period, op.Param2, op.Param3
		if slow == 0 {
			slow = 26
		}
		if sig == 0 {
			sig = 9
		}
		macdLine, signalLine, _ := computeMACD(closes, fast, slow, sig)
		switch op.Field {
		case "signal":
			return signalLine
		default:
			return macdLine
		}
	case model.IndicatorBollingerBands:
		mult := op.ParamF
		if mult == 0 {
			mult = 2.0
		}
		upper, middle, lower, bw := computeBollingerBands(closes, period, mult)
		switch op.Field {
		case "upper":
			return upper
		case "lower":
			return lower
		case "middle":
			return middle
		case "bandwidth":
			return bw
		default:
			return middle
		}
	case model.IndicatorStochastic:
		k, d := computeStochastic(highs, lows, closes, period, op.Param2)
		if op.Field == "d" {
			return d
		}
		return k
	case model.IndicatorADX:
		return computeADX(highs, lows, closes, period)
	case model.IndicatorAroon:
		up, down := computeAroon(highs, lows, period)
		if op.Field == "down" {
			return down
		}
		return up
	case model.IndicatorParabolicSAR:
		return computeParabolicSAR(highs, lows, closes)
	case model.IndicatorSupertrend:
		mult := op.ParamF
		if mult == 0 {
			mult = 3.0
		}
		return computeSupertrend(highs, lows, closes, period, mult)
	case model.IndicatorVWAP:
		return computeVWAP(highs, lows, closes, volumes)
	case model.IndicatorVWMA:
		return computeVWMA(closes, volumes, period)
	case model.IndicatorWilliamsR:
		return computeWilliamsR(highs, lows, closes, period)
	case model.IndicatorCMO:
		return computeCMO(closes, period)
	case model.IndicatorROC:
		return computeROC(closes, period)
	case model.IndicatorMomentum:
		return computeMomentum(closes, period)
	case model.IndicatorKeltnerChannel:
		mult := op.ParamF
		if mult == 0 {
			mult = 2.0
		}
		upper, middle, lower := computeKeltnerChannel(highs, lows, closes, period, mult)
		switch op.Field {
		case "upper":
			return upper
		case "lower":
			return lower
		default:
			return middle
		}
	case model.IndicatorATR:
		return computeATR(highs, lows, closes, period)
	case model.IndicatorStdDev:
		return computeStdDev(closes, period)
	case model.IndicatorOBV:
		return computeOBV(closes, volumes)
	case model.IndicatorLinearReg:
		return computeLinearRegression(closes, period)
	}

	// Unknown indicator — return zeros.
	return make([]float64, n)
}

// --- Condition evaluation ---

func (b *BacktestEngine) evaluateConditions(conditions []model.StrategyCondition, bars []model.OHLCVBar, cache indicatorCache, idx int) bool {
	if len(conditions) == 0 {
		return false
	}
	for _, cond := range conditions {
		if !b.evaluateCondition(cond, bars, cache, idx) {
			return false
		}
	}
	return true
}

func (b *BacktestEngine) evaluateCondition(cond model.StrategyCondition, bars []model.OHLCVBar, cache indicatorCache, idx int) bool {
	leftVal := b.resolveOperand(cond.Left, bars, cache, idx)
	rightVal := b.resolveOperand(cond.Right, bars, cache, idx)

	if math.IsNaN(leftVal) || math.IsNaN(rightVal) {
		return false
	}

	switch cond.Operator {
	case model.OpLessThan:
		return leftVal < rightVal
	case model.OpGreaterThan:
		return leftVal > rightVal
	case model.OpLessEqual:
		return leftVal <= rightVal
	case model.OpGreaterEqual:
		return leftVal >= rightVal
	case model.OpCrossesAbove:
		if idx < 1 {
			return false
		}
		prevLeft := b.resolveOperand(cond.Left, bars, cache, idx-1)
		prevRight := b.resolveOperand(cond.Right, bars, cache, idx-1)
		if math.IsNaN(prevLeft) || math.IsNaN(prevRight) {
			return false
		}
		return prevLeft <= prevRight && leftVal > rightVal
	case model.OpCrossesBelow:
		if idx < 1 {
			return false
		}
		prevLeft := b.resolveOperand(cond.Left, bars, cache, idx-1)
		prevRight := b.resolveOperand(cond.Right, bars, cache, idx-1)
		if math.IsNaN(prevLeft) || math.IsNaN(prevRight) {
			return false
		}
		return prevLeft >= prevRight && leftVal < rightVal
	}
	return false
}

func (b *BacktestEngine) resolveOperand(op model.ConditionOperand, bars []model.OHLCVBar, cache indicatorCache, idx int) float64 {
	switch op.Type {
	case "constant":
		return op.Constant
	case "price":
		if idx < 0 || idx >= len(bars) {
			return math.NaN()
		}
		bar := bars[idx]
		switch op.Field {
		case "open":
			return bar.Open
		case "high":
			return bar.High
		case "low":
			return bar.Low
		case "volume":
			return float64(bar.Volume)
		default: // "close" or empty
			return bar.Close
		}
	case "indicator":
		key := indicatorKey(op)
		series, ok := cache[key]
		if !ok || idx < 0 || idx >= len(series) {
			return math.NaN()
		}
		return series[idx]
	}
	return math.NaN()
}

// --- Result computation helpers ---

func computeWinRate(trades []model.BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}
	wins := 0
	for _, t := range trades {
		if t.ReturnPct > 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(trades))
}

func computeMaxDrawdownFromEquity(curve []model.EquityPoint) float64 {
	if len(curve) == 0 {
		return 0
	}
	peak := curve[0].Value
	maxDD := 0.0
	for _, pt := range curve {
		if pt.Value > peak {
			peak = pt.Value
		}
		dd := (peak - pt.Value) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

func computeSharpeFromEquity(curve []model.EquityPoint) float64 {
	if len(curve) < 2 {
		return 0
	}
	returns := make([]float64, 0, len(curve)-1)
	for i := 1; i < len(curve); i++ {
		if curve[i-1].Value == 0 {
			continue
		}
		r := (curve[i].Value - curve[i-1].Value) / curve[i-1].Value
		returns = append(returns, r)
	}
	if len(returns) == 0 {
		return 0
	}
	avg := 0.0
	for _, r := range returns {
		avg += r
	}
	avg /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		variance += (r - avg) * (r - avg)
	}
	variance /= float64(len(returns))
	stddev := math.Sqrt(variance)
	if stddev == 0 {
		return 0
	}
	// Annualize: assume ~252 trading days.
	return (avg / stddev) * math.Sqrt(252)
}

func computeAvgHoldingDays(trades []model.BacktestTrade) float64 {
	if len(trades) == 0 {
		return 0
	}
	total := 0
	for _, t := range trades {
		total += t.HoldingDays
	}
	return float64(total) / float64(len(trades))
}

// --- Data extraction helpers ---

func extractCloses(bars []model.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.Close
	}
	return out
}

func extractHighs(bars []model.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.High
	}
	return out
}

func extractLows(bars []model.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.Low
	}
	return out
}

func extractVolumes(bars []model.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = float64(b.Volume)
	}
	return out
}

// --- Technical Indicator Implementations ---

// computeSMA computes Simple Moving Average.
func computeSMA(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period > n || period < 1 {
		return result
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)
	for i := period; i < n; i++ {
		sum += data[i] - data[i-period]
		result[i] = sum / float64(period)
	}
	return result
}

// computeEMA computes Exponential Moving Average.
func computeEMA(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period > n || period < 1 {
		return result
	}
	k := 2.0 / float64(period+1)
	// Seed with SMA.
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)
	for i := period; i < n; i++ {
		result[i] = data[i]*k + result[i-1]*(1-k)
	}
	return result
}

// computeRSI computes Relative Strength Index.
func computeRSI(closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period >= n || period < 1 {
		return result
	}

	gains := make([]float64, n)
	losses := make([]float64, n)
	for i := 1; i < n; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	avgGain := 0.0
	avgLoss := 0.0
	for i := 1; i <= period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - 100/(1+rs)
	}

	for i := period + 1; i < n; i++ {
		avgGain = (avgGain*float64(period-1) + gains[i]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i]) / float64(period)
		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - 100/(1+rs)
		}
	}
	return result
}

// computeMACD computes MACD line, signal line, and histogram.
func computeMACD(closes []float64, fast, slow, signal int) ([]float64, []float64, []float64) {
	n := len(closes)
	macdLine := make([]float64, n)
	signalLine := make([]float64, n)
	histogram := make([]float64, n)
	for i := range macdLine {
		macdLine[i] = math.NaN()
		signalLine[i] = math.NaN()
		histogram[i] = math.NaN()
	}

	fastEMA := computeEMA(closes, fast)
	slowEMA := computeEMA(closes, slow)

	// MACD line = fast EMA - slow EMA.
	startIdx := -1
	for i := 0; i < n; i++ {
		if !math.IsNaN(fastEMA[i]) && !math.IsNaN(slowEMA[i]) {
			macdLine[i] = fastEMA[i] - slowEMA[i]
			if startIdx == -1 {
				startIdx = i
			}
		}
	}

	if startIdx == -1 {
		return macdLine, signalLine, histogram
	}

	// Signal line = EMA of MACD line.
	macdValues := make([]float64, 0, n-startIdx)
	for i := startIdx; i < n; i++ {
		macdValues = append(macdValues, macdLine[i])
	}
	sigEMA := computeEMA(macdValues, signal)
	for i, v := range sigEMA {
		if !math.IsNaN(v) {
			signalLine[startIdx+i] = v
			histogram[startIdx+i] = macdLine[startIdx+i] - v
		}
	}
	return macdLine, signalLine, histogram
}

// computeBollingerBands computes upper, middle (SMA), lower bands and bandwidth.
func computeBollingerBands(closes []float64, period int, mult float64) (upper, middle, lower, bandwidth []float64) {
	n := len(closes)
	upper = make([]float64, n)
	middle = computeSMA(closes, period)
	lower = make([]float64, n)
	bandwidth = make([]float64, n)
	for i := range upper {
		upper[i] = math.NaN()
		lower[i] = math.NaN()
		bandwidth[i] = math.NaN()
	}

	sd := computeStdDev(closes, period)
	for i := 0; i < n; i++ {
		if !math.IsNaN(middle[i]) && !math.IsNaN(sd[i]) {
			upper[i] = middle[i] + mult*sd[i]
			lower[i] = middle[i] - mult*sd[i]
			if middle[i] != 0 {
				bandwidth[i] = (upper[i] - lower[i]) / middle[i]
			}
		}
	}
	return
}

// computeStochastic computes %K and %D.
func computeStochastic(highs, lows, closes []float64, kPeriod, dPeriod int) ([]float64, []float64) {
	n := len(closes)
	k := make([]float64, n)
	for i := range k {
		k[i] = math.NaN()
	}
	if dPeriod < 1 {
		dPeriod = 3
	}

	for i := kPeriod - 1; i < n; i++ {
		hh := highs[i]
		ll := lows[i]
		for j := i - kPeriod + 1; j <= i; j++ {
			if highs[j] > hh {
				hh = highs[j]
			}
			if lows[j] < ll {
				ll = lows[j]
			}
		}
		if hh-ll != 0 {
			k[i] = ((closes[i] - ll) / (hh - ll)) * 100
		} else {
			k[i] = 50
		}
	}

	d := computeSMA(k, dPeriod)
	return k, d
}

// computeADX computes Average Directional Index.
func computeADX(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if n < period+1 {
		return result
	}

	atr := computeATR(highs, lows, closes, period)
	plusDM := make([]float64, n)
	minusDM := make([]float64, n)
	for i := 1; i < n; i++ {
		upMove := highs[i] - highs[i-1]
		downMove := lows[i-1] - lows[i]
		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	smoothPlusDM := computeEMA(plusDM, period)
	smoothMinusDM := computeEMA(minusDM, period)

	dx := make([]float64, n)
	for i := range dx {
		dx[i] = math.NaN()
	}
	for i := 0; i < n; i++ {
		if math.IsNaN(atr[i]) || atr[i] == 0 || math.IsNaN(smoothPlusDM[i]) || math.IsNaN(smoothMinusDM[i]) {
			continue
		}
		plusDI := (smoothPlusDM[i] / atr[i]) * 100
		minusDI := (smoothMinusDM[i] / atr[i]) * 100
		sum := plusDI + minusDI
		if sum != 0 {
			dx[i] = math.Abs(plusDI-minusDI) / sum * 100
		} else {
			dx[i] = 0
		}
	}

	return computeEMA(dx, period)
}

// computeAroon computes Aroon Up and Aroon Down.
func computeAroon(highs, lows []float64, period int) ([]float64, []float64) {
	n := len(highs)
	up := make([]float64, n)
	down := make([]float64, n)
	for i := range up {
		up[i] = math.NaN()
		down[i] = math.NaN()
	}

	for i := period; i < n; i++ {
		highIdx := 0
		lowIdx := 0
		hh := highs[i-period]
		ll := lows[i-period]
		for j := 1; j <= period; j++ {
			if highs[i-period+j] >= hh {
				hh = highs[i-period+j]
				highIdx = j
			}
			if lows[i-period+j] <= ll {
				ll = lows[i-period+j]
				lowIdx = j
			}
		}
		up[i] = float64(highIdx) / float64(period) * 100
		down[i] = float64(lowIdx) / float64(period) * 100
	}
	return up, down
}

// computeParabolicSAR computes Parabolic SAR with default AF=0.02, maxAF=0.2.
func computeParabolicSAR(highs, lows, closes []float64) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if n < 2 {
		return result
	}

	af := 0.02
	maxAF := 0.2
	isLong := closes[1] > closes[0]
	sar := lows[0]
	ep := highs[0]
	if !isLong {
		sar = highs[0]
		ep = lows[0]
	}

	result[0] = sar
	for i := 1; i < n; i++ {
		prevSAR := sar
		sar = prevSAR + af*(ep-prevSAR)

		if isLong {
			if i >= 2 {
				sar = math.Min(sar, math.Min(lows[i-1], lows[i-2]))
			}
			if lows[i] < sar {
				isLong = false
				sar = ep
				ep = lows[i]
				af = 0.02
			} else {
				if highs[i] > ep {
					ep = highs[i]
					af = math.Min(af+0.02, maxAF)
				}
			}
		} else {
			if i >= 2 {
				sar = math.Max(sar, math.Max(highs[i-1], highs[i-2]))
			}
			if highs[i] > sar {
				isLong = true
				sar = ep
				ep = highs[i]
				af = 0.02
			} else {
				if lows[i] < ep {
					ep = lows[i]
					af = math.Min(af+0.02, maxAF)
				}
			}
		}
		result[i] = sar
	}
	return result
}

// computeSupertrend computes Supertrend indicator.
func computeSupertrend(highs, lows, closes []float64, period int, mult float64) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	atr := computeATR(highs, lows, closes, period)

	upperBand := make([]float64, n)
	lowerBand := make([]float64, n)
	isUpTrend := true

	for i := 0; i < n; i++ {
		if math.IsNaN(atr[i]) {
			continue
		}
		mid := (highs[i] + lows[i]) / 2
		upperBand[i] = mid + mult*atr[i]
		lowerBand[i] = mid - mult*atr[i]

		if i > 0 && !math.IsNaN(result[i-1]) {
			if lowerBand[i] < lowerBand[i-1] && isUpTrend {
				lowerBand[i] = lowerBand[i-1]
			}
			if upperBand[i] > upperBand[i-1] && !isUpTrend {
				upperBand[i] = upperBand[i-1]
			}
		}

		if isUpTrend {
			if closes[i] < lowerBand[i] {
				isUpTrend = false
				result[i] = upperBand[i]
			} else {
				result[i] = lowerBand[i]
			}
		} else {
			if closes[i] > upperBand[i] {
				isUpTrend = true
				result[i] = lowerBand[i]
			} else {
				result[i] = upperBand[i]
			}
		}
	}
	return result
}

// computeVWAP computes Volume Weighted Average Price (cumulative).
func computeVWAP(highs, lows, closes, volumes []float64) []float64 {
	n := len(closes)
	result := make([]float64, n)
	cumTPV := 0.0
	cumVol := 0.0
	for i := 0; i < n; i++ {
		tp := (highs[i] + lows[i] + closes[i]) / 3
		cumTPV += tp * volumes[i]
		cumVol += volumes[i]
		if cumVol != 0 {
			result[i] = cumTPV / cumVol
		} else {
			result[i] = closes[i]
		}
	}
	return result
}

// computeVWMA computes Volume Weighted Moving Average.
func computeVWMA(closes, volumes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period > n {
		return result
	}
	for i := period - 1; i < n; i++ {
		sumPV := 0.0
		sumV := 0.0
		for j := i - period + 1; j <= i; j++ {
			sumPV += closes[j] * volumes[j]
			sumV += volumes[j]
		}
		if sumV != 0 {
			result[i] = sumPV / sumV
		} else {
			result[i] = closes[i]
		}
	}
	return result
}

// computeWilliamsR computes Williams %R.
func computeWilliamsR(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	for i := period - 1; i < n; i++ {
		hh := highs[i]
		ll := lows[i]
		for j := i - period + 1; j <= i; j++ {
			if highs[j] > hh {
				hh = highs[j]
			}
			if lows[j] < ll {
				ll = lows[j]
			}
		}
		if hh-ll != 0 {
			result[i] = ((hh - closes[i]) / (hh - ll)) * -100
		} else {
			result[i] = -50
		}
	}
	return result
}

// computeCMO computes Chande Momentum Oscillator.
func computeCMO(closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period >= n {
		return result
	}
	for i := period; i < n; i++ {
		sumUp := 0.0
		sumDown := 0.0
		for j := i - period + 1; j <= i; j++ {
			diff := closes[j] - closes[j-1]
			if diff > 0 {
				sumUp += diff
			} else {
				sumDown += -diff
			}
		}
		if sumUp+sumDown != 0 {
			result[i] = ((sumUp - sumDown) / (sumUp + sumDown)) * 100
		} else {
			result[i] = 0
		}
	}
	return result
}

// computeROC computes Rate of Change.
func computeROC(closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	for i := period; i < n; i++ {
		if closes[i-period] != 0 {
			result[i] = ((closes[i] - closes[i-period]) / closes[i-period]) * 100
		}
	}
	return result
}

// computeMomentum computes Momentum indicator.
func computeMomentum(closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	for i := period; i < n; i++ {
		result[i] = closes[i] - closes[i-period]
	}
	return result
}

// computeKeltnerChannel computes Keltner Channel (upper, middle, lower).
func computeKeltnerChannel(highs, lows, closes []float64, period int, mult float64) ([]float64, []float64, []float64) {
	n := len(closes)
	middle := computeEMA(closes, period)
	atr := computeATR(highs, lows, closes, period)
	upper := make([]float64, n)
	lower := make([]float64, n)
	for i := range upper {
		upper[i] = math.NaN()
		lower[i] = math.NaN()
	}
	for i := 0; i < n; i++ {
		if !math.IsNaN(middle[i]) && !math.IsNaN(atr[i]) {
			upper[i] = middle[i] + mult*atr[i]
			lower[i] = middle[i] - mult*atr[i]
		}
	}
	return upper, middle, lower
}

// computeATR computes Average True Range.
func computeATR(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if n < 2 || period < 1 {
		return result
	}

	tr := make([]float64, n)
	tr[0] = highs[0] - lows[0]
	for i := 1; i < n; i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	if period > n {
		return result
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	result[period-1] = sum / float64(period)
	for i := period; i < n; i++ {
		result[i] = (result[i-1]*float64(period-1) + tr[i]) / float64(period)
	}
	return result
}

// computeStdDev computes rolling standard deviation.
func computeStdDev(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period > n {
		return result
	}
	for i := period - 1; i < n; i++ {
		sum := 0.0
		for j := i - period + 1; j <= i; j++ {
			sum += data[j]
		}
		avg := sum / float64(period)
		variance := 0.0
		for j := i - period + 1; j <= i; j++ {
			variance += (data[j] - avg) * (data[j] - avg)
		}
		result[i] = math.Sqrt(variance / float64(period))
	}
	return result
}

// computeOBV computes On-Balance Volume.
func computeOBV(closes, volumes []float64) []float64 {
	n := len(closes)
	result := make([]float64, n)
	if n == 0 {
		return result
	}
	result[0] = volumes[0]
	for i := 1; i < n; i++ {
		if closes[i] > closes[i-1] {
			result[i] = result[i-1] + volumes[i]
		} else if closes[i] < closes[i-1] {
			result[i] = result[i-1] - volumes[i]
		} else {
			result[i] = result[i-1]
		}
	}
	return result
}

// computeLinearRegression computes rolling linear regression value.
func computeLinearRegression(data []float64, period int) []float64 {
	n := len(data)
	result := make([]float64, n)
	for i := range result {
		result[i] = math.NaN()
	}
	if period > n {
		return result
	}
	for i := period - 1; i < n; i++ {
		sumX := 0.0
		sumY := 0.0
		sumXY := 0.0
		sumX2 := 0.0
		for j := 0; j < period; j++ {
			x := float64(j)
			y := data[i-period+1+j]
			sumX += x
			sumY += y
			sumXY += x * y
			sumX2 += x * x
		}
		pf := float64(period)
		denom := pf*sumX2 - sumX*sumX
		if denom == 0 {
			result[i] = data[i]
			continue
		}
		slope := (pf*sumXY - sumX*sumY) / denom
		intercept := (sumY - slope*sumX) / pf
		result[i] = intercept + slope*float64(period-1)
	}
	return result
}
