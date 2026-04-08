package ranking

import (
	"fmt"
	"sort"

	"myfi-backend/internal/domain/market"
)

// BacktestEngine runs indicator-based trading strategies against historical OHLCV data.
// Requirement 32: Backtesting and Strategy Simulation.
type BacktestEngine struct{}

// NewBacktestEngine creates a new BacktestEngine instance.
func NewBacktestEngine() *BacktestEngine {
	return &BacktestEngine{}
}

// RunBacktest executes a strategy against the provided OHLCV bars and returns results.
func (b *BacktestEngine) RunBacktest(bars []market.OHLCVBar, strategy StrategyRule) (BacktestResult, error) {
	if len(bars) < 2 {
		return BacktestResult{}, fmt.Errorf("insufficient data: need at least 2 bars, got %d", len(bars))
	}

	sorted := make([]market.OHLCVBar, len(bars))
	copy(sorted, bars)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Time.Before(sorted[j].Time) })

	indicators := b.computeAllIndicators(sorted, strategy)

	var trades []BacktestTrade
	inPosition := false
	var entryPrice float64
	var entryIdx int

	initialCapital := 1000000.0
	capital := initialCapital
	equityCurve := make([]EquityPoint, 0, len(sorted))

	for i := 0; i < len(sorted); i++ {
		bar := sorted[i]
		currentValue := capital
		if inPosition {
			currentValue = capital * (bar.Close / entryPrice)
		}
		equityCurve = append(equityCurve, EquityPoint{Date: bar.Time, Value: currentValue})

		if !inPosition {
			if b.evaluateConditions(strategy.EntryConditions, sorted, indicators, i) {
				inPosition = true
				entryPrice = bar.Close
				entryIdx = i
			}
		} else {
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
				trade := BacktestTrade{
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

	if inPosition {
		lastBar := sorted[len(sorted)-1]
		returnPct := (lastBar.Close - entryPrice) / entryPrice
		holdingDays := int(lastBar.Time.Sub(sorted[entryIdx].Time).Hours() / 24)
		if holdingDays < 1 {
			holdingDays = 1
		}
		trades = append(trades, BacktestTrade{
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

	if len(equityCurve) > 0 {
		equityCurve[len(equityCurve)-1].Value = capital
	}

	result := BacktestResult{
		TotalReturn:    (capital - initialCapital) / initialCapital,
		NumTrades:      len(trades),
		EquityCurve:    equityCurve,
		Trades:         trades,
		WinRate:        ComputeWinRate(trades),
		MaxDrawdown:    ComputeMaxDrawdownFromEquity(equityCurve),
		SharpeRatio:    ComputeSharpeFromEquity(equityCurve),
		AvgHoldingDays: computeAvgHoldingDays(trades),
	}
	return result, nil
}

// GetPresetStrategies returns the built-in preset strategies.
func (b *BacktestEngine) GetPresetStrategies() []StrategyRule {
	return []StrategyRule{
		{
			Name: "RSI Oversold Bounce",
			EntryConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "indicator", Indicator: IndicatorRSI, Field: "value", Period: 14},
					Operator: OpLessThan,
					Right:    ConditionOperand{Type: "constant", Constant: 30},
				},
			},
			ExitConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "indicator", Indicator: IndicatorRSI, Field: "value", Period: 14},
					Operator: OpGreaterThan,
					Right:    ConditionOperand{Type: "constant", Constant: 70},
				},
			},
			StopLossPct:   0.05,
			TakeProfitPct: 0.10,
		},
		{
			Name: "MACD Crossover",
			EntryConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "indicator", Indicator: IndicatorMACD, Field: "macd", Period: 12, Param2: 26, Param3: 9},
					Operator: OpCrossesAbove,
					Right:    ConditionOperand{Type: "indicator", Indicator: IndicatorMACD, Field: "signal", Period: 12, Param2: 26, Param3: 9},
				},
			},
			ExitConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "indicator", Indicator: IndicatorMACD, Field: "macd", Period: 12, Param2: 26, Param3: 9},
					Operator: OpCrossesBelow,
					Right:    ConditionOperand{Type: "indicator", Indicator: IndicatorMACD, Field: "signal", Period: 12, Param2: 26, Param3: 9},
				},
			},
			StopLossPct:   0.05,
			TakeProfitPct: 0.15,
		},
		{
			Name: "Bollinger Band Squeeze",
			EntryConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "price", Field: "close"},
					Operator: OpLessThan,
					Right:    ConditionOperand{Type: "indicator", Indicator: IndicatorBollingerBands, Field: "lower", Period: 20, ParamF: 2.0},
				},
			},
			ExitConditions: []StrategyCondition{
				{
					Left:     ConditionOperand{Type: "price", Field: "close"},
					Operator: OpGreaterThan,
					Right:    ConditionOperand{Type: "indicator", Indicator: IndicatorBollingerBands, Field: "upper", Period: 20, ParamF: 2.0},
				},
			},
			StopLossPct:   0.03,
			TakeProfitPct: 0.08,
		},
	}
}

// --- Indicator computation ---

type indicatorCache map[string][]float64

func (b *BacktestEngine) computeAllIndicators(bars []market.OHLCVBar, strategy StrategyRule) indicatorCache {
	cache := make(indicatorCache)
	allConds := append(strategy.EntryConditions, strategy.ExitConditions...)
	for _, cond := range allConds {
		for _, op := range []ConditionOperand{cond.Left, cond.Right} {
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

func indicatorKey(op ConditionOperand) string {
	return fmt.Sprintf("%s_%s_%d_%d_%d_%.2f", op.Indicator, op.Field, op.Period, op.Param2, op.Param3, op.ParamF)
}

func (b *BacktestEngine) computeIndicator(bars []market.OHLCVBar, op ConditionOperand) []float64 {
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
	case IndicatorSMA:
		return computeSMA(closes, period)
	case IndicatorEMA:
		return computeEMA(closes, period)
	case IndicatorRSI:
		return computeRSI(closes, period)
	case IndicatorMACD:
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
	case IndicatorBollingerBands:
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
	case IndicatorStochastic:
		k, d := computeStochastic(highs, lows, closes, period, op.Param2)
		if op.Field == "d" {
			return d
		}
		return k
	case IndicatorADX:
		return computeADX(highs, lows, closes, period)
	case IndicatorAroon:
		up, down := computeAroon(highs, lows, period)
		if op.Field == "down" {
			return down
		}
		return up
	case IndicatorParabolicSAR:
		return computeParabolicSAR(highs, lows, closes)
	case IndicatorSupertrend:
		mult := op.ParamF
		if mult == 0 {
			mult = 3.0
		}
		return computeSupertrend(highs, lows, closes, period, mult)
	case IndicatorVWAP:
		return computeVWAP(highs, lows, closes, volumes)
	case IndicatorVWMA:
		return computeVWMA(closes, volumes, period)
	case IndicatorWilliamsR:
		return computeWilliamsR(highs, lows, closes, period)
	case IndicatorCMO:
		return computeCMO(closes, period)
	case IndicatorROC:
		return computeROC(closes, period)
	case IndicatorMomentum:
		return computeMomentum(closes, period)
	case IndicatorKeltnerChannel:
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
	case IndicatorATR:
		return computeATR(highs, lows, closes, period)
	case IndicatorStdDev:
		return computeStdDev(closes, period)
	case IndicatorOBV:
		return computeOBV(closes, volumes)
	case IndicatorLinearReg:
		return computeLinearRegression(closes, period)
	}

	return make([]float64, n)
}
