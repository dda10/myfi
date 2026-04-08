package ranking

import (
	"math"

	"myfi-backend/internal/domain/market"
)

// --- Condition evaluation ---

func (b *BacktestEngine) evaluateConditions(conditions []StrategyCondition, bars []market.OHLCVBar, cache indicatorCache, idx int) bool {
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

func (b *BacktestEngine) evaluateCondition(cond StrategyCondition, bars []market.OHLCVBar, cache indicatorCache, idx int) bool {
	leftVal := b.resolveOperand(cond.Left, bars, cache, idx)
	rightVal := b.resolveOperand(cond.Right, bars, cache, idx)

	if math.IsNaN(leftVal) || math.IsNaN(rightVal) {
		return false
	}

	switch cond.Operator {
	case OpLessThan:
		return leftVal < rightVal
	case OpGreaterThan:
		return leftVal > rightVal
	case OpLessEqual:
		return leftVal <= rightVal
	case OpGreaterEqual:
		return leftVal >= rightVal
	case OpCrossesAbove:
		if idx < 1 {
			return false
		}
		prevLeft := b.resolveOperand(cond.Left, bars, cache, idx-1)
		prevRight := b.resolveOperand(cond.Right, bars, cache, idx-1)
		if math.IsNaN(prevLeft) || math.IsNaN(prevRight) {
			return false
		}
		return prevLeft <= prevRight && leftVal > rightVal
	case OpCrossesBelow:
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

func (b *BacktestEngine) resolveOperand(op ConditionOperand, bars []market.OHLCVBar, cache indicatorCache, idx int) float64 {
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
		default:
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

// ComputeWinRate computes the win rate from a list of trades.
func ComputeWinRate(trades []BacktestTrade) float64 {
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

// ComputeMaxDrawdownFromEquity computes the maximum drawdown from an equity curve.
func ComputeMaxDrawdownFromEquity(curve []EquityPoint) float64 {
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

// ComputeSharpeFromEquity computes the Sharpe ratio from an equity curve.
func ComputeSharpeFromEquity(curve []EquityPoint) float64 {
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
	return (avg / stddev) * math.Sqrt(252)
}

func computeAvgHoldingDays(trades []BacktestTrade) float64 {
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

func extractCloses(bars []market.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.Close
	}
	return out
}

func extractHighs(bars []market.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.High
	}
	return out
}

func extractLows(bars []market.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = b.Low
	}
	return out
}

func extractVolumes(bars []market.OHLCVBar) []float64 {
	out := make([]float64, len(bars))
	for i, b := range bars {
		out[i] = float64(b.Volume)
	}
	return out
}

// --- Technical Indicator Implementations ---

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
