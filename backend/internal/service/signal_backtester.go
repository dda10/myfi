package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
)

// SignalBacktester runs historical simulations of the SignalEngine's factor-based strategy.
type SignalBacktester struct {
	router          *infra.DataSourceRouter
	liquidityFilter *LiquidityFilter
	sectorService   *SectorService
	screenerService *ScreenerService
	config          model.SignalConfig
}

// SignalBacktestRequest defines the parameters for a signal backtest.
type SignalBacktestRequest struct {
	StartDate time.Time          `json:"startDate"`
	EndDate   time.Time          `json:"endDate"`
	Symbols   []string           `json:"symbols"`  // Empty = use liquidity filter whitelist
	Config    model.SignalConfig `json:"config"`   // Factor weights to test
	MaxHold   int                `json:"maxHold"`  // Max holding days (0 = until SL/TP)
	TopN      int                `json:"topN"`     // Signals per scan day
	ScanFreq  int                `json:"scanFreq"` // Days between scans (1 = daily)
}

// SignalBacktestTrade represents a simulated trade.
type SignalBacktestTrade struct {
	Symbol         string             `json:"symbol"`
	SignalDate     time.Time          `json:"signalDate"`
	EntryDate      time.Time          `json:"entryDate"`
	ExitDate       time.Time          `json:"exitDate"`
	EntryPrice     float64            `json:"entryPrice"`
	ExitPrice      float64            `json:"exitPrice"`
	StopLoss       float64            `json:"stopLoss"`
	TakeProfit     float64            `json:"takeProfit"`
	ReturnPct      float64            `json:"returnPct"`
	ExitReason     string             `json:"exitReason"` // "stop_loss", "take_profit", "max_hold", "end_of_data"
	HoldingDays    int                `json:"holdingDays"`
	CompositeScore float64            `json:"compositeScore"`
	Factors        model.FactorScores `json:"factors"`
}

// SignalBacktestResult contains the full backtest output.
type SignalBacktestResult struct {
	// Performance metrics
	TotalReturn    float64 `json:"totalReturn"`
	WinRate        float64 `json:"winRate"`
	AvgWin         float64 `json:"avgWin"`
	AvgLoss        float64 `json:"avgLoss"`
	ProfitFactor   float64 `json:"profitFactor"` // Gross profit / Gross loss
	MaxDrawdown    float64 `json:"maxDrawdown"`
	SharpeRatio    float64 `json:"sharpeRatio"`
	NumTrades      int     `json:"numTrades"`
	AvgHoldingDays float64 `json:"avgHoldingDays"`

	// Factor analysis
	AvgCompositeScore float64            `json:"avgCompositeScore"`
	AvgFactorScores   model.FactorScores `json:"avgFactorScores"`

	// Breakdown
	WinsByExitReason map[string]int     `json:"winsByExitReason"`
	LossByExitReason map[string]int     `json:"lossByExitReason"`
	ReturnByStrength map[string]float64 `json:"returnByStrength"`

	// Details
	Trades      []SignalBacktestTrade `json:"trades"`
	EquityCurve []model.EquityPoint   `json:"equityCurve"`
	Config      model.SignalConfig    `json:"config"`
	Request     SignalBacktestRequest `json:"request"`
}

// NewSignalBacktester creates a new backtester instance.
func NewSignalBacktester(
	router *infra.DataSourceRouter,
	liquidityFilter *LiquidityFilter,
	sectorService *SectorService,
	screenerService *ScreenerService,
) *SignalBacktester {
	return &SignalBacktester{
		router:          router,
		liquidityFilter: liquidityFilter,
		sectorService:   sectorService,
		screenerService: screenerService,
		config:          model.DefaultSignalConfig(),
	}
}

// RunBacktest executes a historical simulation of the signal strategy.
func (b *SignalBacktester) RunBacktest(ctx context.Context, req SignalBacktestRequest) (SignalBacktestResult, error) {
	result := SignalBacktestResult{
		WinsByExitReason: make(map[string]int),
		LossByExitReason: make(map[string]int),
		ReturnByStrength: make(map[string]float64),
		Config:           req.Config,
		Request:          req,
	}

	// Apply config
	if req.Config.MomentumWeight > 0 {
		b.config = req.Config
	} else {
		b.config = model.DefaultSignalConfig()
	}

	// Set defaults
	if req.TopN == 0 {
		req.TopN = 5
	}
	if req.ScanFreq == 0 {
		req.ScanFreq = 5 // Weekly scans
	}
	if req.MaxHold == 0 {
		req.MaxHold = 20 // 20 trading days max hold
	}

	// Get universe
	symbols := req.Symbols
	if len(symbols) == 0 {
		snapshot := b.liquidityFilter.GetWhitelist()
		for _, entry := range snapshot.Entries {
			symbols = append(symbols, entry.Symbol)
		}
	}
	if len(symbols) == 0 {
		return result, fmt.Errorf("no symbols to backtest")
	}

	// Pre-fetch all historical data
	priceCache := make(map[string][]model.OHLCVBar)
	lookbackStart := req.StartDate.AddDate(0, 0, -100) // Extra data for indicators
	for _, sym := range symbols {
		bars, err := b.fetchOHLCV(ctx, sym, lookbackStart, req.EndDate)
		if err == nil && len(bars) > 30 {
			priceCache[sym] = bars
		}
	}

	// Get sector trends (static for simplicity)
	sectorTrendMap := make(map[model.ICBSector]model.SectorTrend)
	sectorChangeMap := make(map[model.ICBSector]float64)
	for _, sector := range model.AllICBSectors {
		sectorTrendMap[sector] = model.Sideways
		sectorChangeMap[sector] = 0
	}

	// Simulate day by day
	var trades []SignalBacktestTrade
	openPositions := make(map[string]*SignalBacktestTrade)
	equity := 100000.0
	equityCurve := []model.EquityPoint{{Date: req.StartDate, Value: equity}}
	peakEquity := equity

	currentDate := req.StartDate
	dayCount := 0

	for !currentDate.After(req.EndDate) {
		dayCount++

		// Check open positions for exit
		for sym, pos := range openPositions {
			bars := priceCache[sym]
			barIdx := findBarIndex(bars, currentDate)
			if barIdx < 0 {
				continue
			}
			bar := bars[barIdx]
			holdingDays := int(currentDate.Sub(pos.EntryDate).Hours() / 24)

			var exitPrice float64
			var exitReason string

			// Check stop loss (using low)
			if bar.Low <= pos.StopLoss {
				exitPrice = pos.StopLoss
				exitReason = "stop_loss"
			} else if bar.High >= pos.TakeProfit {
				// Check take profit (using high)
				exitPrice = pos.TakeProfit
				exitReason = "take_profit"
			} else if holdingDays >= req.MaxHold {
				// Max hold reached
				exitPrice = bar.Close
				exitReason = "max_hold"
			}

			if exitReason != "" {
				pos.ExitDate = currentDate
				pos.ExitPrice = exitPrice
				pos.HoldingDays = holdingDays
				pos.ReturnPct = (exitPrice - pos.EntryPrice) / pos.EntryPrice * 100
				pos.ExitReason = exitReason
				trades = append(trades, *pos)
				delete(openPositions, sym)

				// Update equity
				equity *= (1 + pos.ReturnPct/100)
			}
		}

		// Generate new signals on scan days
		if dayCount%req.ScanFreq == 1 || req.ScanFreq == 1 {
			signals := b.generateSignalsAtDate(ctx, priceCache, currentDate, sectorTrendMap, sectorChangeMap)

			// Sort by composite score
			sort.Slice(signals, func(i, j int) bool {
				return signals[i].CompositeScore > signals[j].CompositeScore
			})

			// Take top N that aren't already in position
			taken := 0
			for _, sig := range signals {
				if taken >= req.TopN {
					break
				}
				if _, exists := openPositions[sig.Symbol]; exists {
					continue
				}
				if sig.CompositeScore < b.config.MinCompositeScore {
					continue
				}

				// Open position
				pos := &SignalBacktestTrade{
					Symbol:         sig.Symbol,
					SignalDate:     currentDate,
					EntryDate:      currentDate,
					EntryPrice:     sig.CurrentPrice,
					StopLoss:       sig.StopLoss,
					TakeProfit:     sig.TakeProfit,
					CompositeScore: sig.CompositeScore,
					Factors:        sig.Factors,
				}
				openPositions[sig.Symbol] = pos
				taken++
			}
		}

		// Update equity curve
		equityCurve = append(equityCurve, model.EquityPoint{Date: currentDate, Value: equity})
		if equity > peakEquity {
			peakEquity = equity
		}

		// Next trading day
		currentDate = currentDate.AddDate(0, 0, 1)
		// Skip weekends
		for currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday {
			currentDate = currentDate.AddDate(0, 0, 1)
		}
	}

	// Close remaining positions at end
	for sym, pos := range openPositions {
		bars := priceCache[sym]
		if len(bars) > 0 {
			lastBar := bars[len(bars)-1]
			pos.ExitDate = lastBar.Time
			pos.ExitPrice = lastBar.Close
			pos.HoldingDays = int(pos.ExitDate.Sub(pos.EntryDate).Hours() / 24)
			pos.ReturnPct = (pos.ExitPrice - pos.EntryPrice) / pos.EntryPrice * 100
			pos.ExitReason = "end_of_data"
			trades = append(trades, *pos)
		}
	}

	// Calculate metrics
	result.Trades = trades
	result.EquityCurve = equityCurve
	result.NumTrades = len(trades)

	if len(trades) == 0 {
		return result, nil
	}

	// Win/loss stats
	var wins, losses int
	var totalWin, totalLoss float64
	var totalReturn float64
	var totalHoldingDays int
	var totalComposite float64
	var factorSums model.FactorScores

	for _, t := range trades {
		totalReturn += t.ReturnPct
		totalHoldingDays += t.HoldingDays
		totalComposite += t.CompositeScore
		factorSums.Momentum += t.Factors.Momentum
		factorSums.Trend += t.Factors.Trend
		factorSums.Volume += t.Factors.Volume
		factorSums.Fundamental += t.Factors.Fundamental
		factorSums.Sector += t.Factors.Sector

		strength := strengthFromScore(t.CompositeScore)
		result.ReturnByStrength[strength] += t.ReturnPct

		if t.ReturnPct > 0 {
			wins++
			totalWin += t.ReturnPct
			result.WinsByExitReason[t.ExitReason]++
		} else {
			losses++
			totalLoss += math.Abs(t.ReturnPct)
			result.LossByExitReason[t.ExitReason]++
		}
	}

	n := float64(len(trades))
	result.TotalReturn = (equity - 100000) / 100000 * 100
	result.WinRate = float64(wins) / n * 100
	result.AvgHoldingDays = float64(totalHoldingDays) / n
	result.AvgCompositeScore = totalComposite / n

	if wins > 0 {
		result.AvgWin = totalWin / float64(wins)
	}
	if losses > 0 {
		result.AvgLoss = totalLoss / float64(losses)
	}
	if totalLoss > 0 {
		result.ProfitFactor = totalWin / totalLoss
	}

	result.AvgFactorScores = model.FactorScores{
		Momentum:    factorSums.Momentum / n,
		Trend:       factorSums.Trend / n,
		Volume:      factorSums.Volume / n,
		Fundamental: factorSums.Fundamental / n,
		Sector:      factorSums.Sector / n,
	}

	// Max drawdown
	result.MaxDrawdown = calculateMaxDrawdown(equityCurve)

	// Sharpe ratio (simplified: assume daily returns, 252 trading days)
	result.SharpeRatio = calculateSharpeFromTrades(trades)

	return result, nil
}

// generateSignalsAtDate generates signals for all symbols at a specific historical date.
func (b *SignalBacktester) generateSignalsAtDate(
	ctx context.Context,
	priceCache map[string][]model.OHLCVBar,
	targetDate time.Time,
	sectorTrendMap map[model.ICBSector]model.SectorTrend,
	sectorChangeMap map[model.ICBSector]float64,
) []model.TradingSignal {
	var signals []model.TradingSignal

	for symbol, bars := range priceCache {
		// Find the bar index for target date
		barIdx := findBarIndex(bars, targetDate)
		if barIdx < 30 { // Need at least 30 bars of history
			continue
		}

		// Use only data up to target date
		historicalBars := bars[:barIdx+1]

		signal, err := b.scoreStockAtDate(ctx, symbol, historicalBars, sectorTrendMap, sectorChangeMap)
		if err != nil {
			continue
		}

		signals = append(signals, signal)
	}

	return signals
}

// scoreStockAtDate computes factor scores using historical data up to a specific date.
func (b *SignalBacktester) scoreStockAtDate(
	ctx context.Context,
	symbol string,
	bars []model.OHLCVBar,
	sectorTrendMap map[model.ICBSector]model.SectorTrend,
	sectorChangeMap map[model.ICBSector]float64,
) (model.TradingSignal, error) {
	signal := model.TradingSignal{
		Symbol:      symbol,
		GeneratedAt: bars[len(bars)-1].Time,
	}

	if len(bars) < 30 {
		return signal, fmt.Errorf("insufficient data")
	}

	// Extract price arrays
	closes := make([]float64, len(bars))
	highs := make([]float64, len(bars))
	lows := make([]float64, len(bars))
	volumes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
		highs[i] = b.High
		lows[i] = b.Low
		volumes[i] = float64(b.Volume)
	}

	lastIdx := len(bars) - 1
	signal.CurrentPrice = closes[lastIdx]
	signal.EntryPrice = closes[lastIdx]

	// Compute technical indicators
	rsi := computeRSI(closes, b.config.RSIPeriod)
	macdLine, macdSignal, _ := computeMACD(closes, b.config.MACDFast, b.config.MACDSlow, b.config.MACDSignal)
	adx := computeADX(highs, lows, closes, b.config.ADXPeriod)
	atr := computeATR(highs, lows, closes, b.config.ATRPeriod)
	obv := computeOBV(closes, volumes)
	obvSMA := computeSMA(obv, b.config.OBVSMAPeriod)
	priceSMA := computeSMA(closes, b.config.PriceSMAPeriod)

	// Get sector info
	sector, _ := b.sectorService.GetStockSector(symbol)
	signal.Sector = sector
	signal.SectorName = model.SectorNameMap[sector]

	// --- Factor 1: Momentum (0-100) ---
	var momentumScore float64
	rsiVal := safeGet(rsi, lastIdx)
	if rsiVal < 30 {
		momentumScore += 40
	} else if rsiVal < 50 {
		momentumScore += 25
	} else if rsiVal < 70 {
		momentumScore += 15
	}

	macdVal := safeGet(macdLine, lastIdx)
	macdSigVal := safeGet(macdSignal, lastIdx)
	macdPrev := safeGet(macdLine, lastIdx-1)
	macdSigPrev := safeGet(macdSignal, lastIdx-1)
	if macdVal > macdSigVal && macdPrev <= macdSigPrev {
		momentumScore += 40
	} else if macdVal > macdSigVal {
		momentumScore += 20
	}

	smaVal := safeGet(priceSMA, lastIdx)
	if smaVal > 0 && closes[lastIdx] > smaVal {
		momentumScore += 20
	}
	signal.Factors.Momentum = clamp(momentumScore, 0, 100)

	// --- Factor 2: Trend (0-100) ---
	var trendScore float64
	adxVal := safeGet(adx, lastIdx)
	if adxVal > 25 {
		trendScore += 50
	} else if adxVal > 20 {
		trendScore += 30
	}

	if lastIdx >= 10 {
		higherHighs := closes[lastIdx] > closes[lastIdx-5] && closes[lastIdx-5] > closes[lastIdx-10]
		if higherHighs {
			trendScore += 50
		}
	}
	signal.Factors.Trend = clamp(trendScore, 0, 100)

	// --- Factor 3: Volume (0-100) ---
	var volumeScore float64
	obvVal := safeGet(obv, lastIdx)
	obvSMAVal := safeGet(obvSMA, lastIdx)
	if obvSMAVal > 0 && obvVal > obvSMAVal {
		volumeScore += 50
	}

	if lastIdx >= b.config.VolumeLookback {
		avgVol := 0.0
		for i := lastIdx - b.config.VolumeLookback; i < lastIdx; i++ {
			avgVol += volumes[i]
		}
		avgVol /= float64(b.config.VolumeLookback)
		if avgVol > 0 && volumes[lastIdx] > avgVol*1.5 {
			volumeScore += 50
		}
	}
	signal.Factors.Volume = clamp(volumeScore, 0, 100)

	// --- Factor 4: Fundamental (0-100) ---
	// For backtest, use simplified scoring (no real-time fundamental data)
	signal.Factors.Fundamental = 50 // Neutral

	// --- Factor 5: Sector (0-100) ---
	var sectorScore float64
	sectorTrend := sectorTrendMap[sector]
	sectorChange := sectorChangeMap[sector]

	if sectorTrend == model.Uptrend {
		sectorScore += 50
	} else if sectorTrend == model.Sideways {
		sectorScore += 25
	}
	if sectorChange > 5 {
		sectorScore += 50
	} else if sectorChange > 0 {
		sectorScore += 25
	}
	signal.Factors.Sector = clamp(sectorScore, 0, 100)

	// --- Composite Score ---
	signal.CompositeScore = signal.Factors.Momentum*b.config.MomentumWeight +
		signal.Factors.Trend*b.config.TrendWeight +
		signal.Factors.Volume*b.config.VolumeWeight +
		signal.Factors.Fundamental*b.config.FundamentalWeight +
		signal.Factors.Sector*b.config.SectorWeight

	// --- Signal Direction and Strength ---
	if signal.CompositeScore >= 75 {
		signal.Direction = model.SignalLong
		signal.Strength = model.StrengthStrong
	} else if signal.CompositeScore >= 60 {
		signal.Direction = model.SignalLong
		signal.Strength = model.StrengthModerate
	} else {
		signal.Direction = model.SignalLong
		signal.Strength = model.StrengthWeak
	}

	// --- Stop Loss and Take Profit ---
	atrVal := safeGet(atr, lastIdx)
	if atrVal > 0 {
		signal.StopLoss = signal.EntryPrice - (atrVal * b.config.ATRMultiplierSL)
		slDistance := signal.EntryPrice - signal.StopLoss
		signal.TakeProfit = signal.EntryPrice + (slDistance * b.config.RiskRewardRatio)
		signal.RiskReward = b.config.RiskRewardRatio
	} else {
		signal.StopLoss = signal.EntryPrice * 0.95
		signal.TakeProfit = signal.EntryPrice * 1.10
		signal.RiskReward = 2.0
	}

	return signal, nil
}

// fetchOHLCV retrieves historical price data for a symbol.
func (b *SignalBacktester) fetchOHLCV(ctx context.Context, symbol string, start, end time.Time) ([]model.OHLCVBar, error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    start,
		End:      end,
		Interval: "1D",
	}

	history, _, err := b.router.FetchQuoteHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	bars := make([]model.OHLCVBar, len(history))
	for i, h := range history {
		bars[i] = model.OHLCVBar{
			Time:   h.Timestamp,
			Open:   h.Open,
			High:   h.High,
			Low:    h.Low,
			Close:  h.Close,
			Volume: int64(h.Volume),
		}
	}
	return bars, nil
}

// findBarIndex finds the index of the bar closest to the target date.
func findBarIndex(bars []model.OHLCVBar, target time.Time) int {
	targetDay := target.Truncate(24 * time.Hour)
	for i := len(bars) - 1; i >= 0; i-- {
		barDay := bars[i].Time.Truncate(24 * time.Hour)
		if barDay.Equal(targetDay) || barDay.Before(targetDay) {
			return i
		}
	}
	return -1
}

// strengthFromScore converts composite score to strength label.
func strengthFromScore(score float64) string {
	if score >= 75 {
		return "strong"
	} else if score >= 60 {
		return "moderate"
	}
	return "weak"
}

// calculateMaxDrawdown computes the maximum drawdown from equity curve.
func calculateMaxDrawdown(curve []model.EquityPoint) float64 {
	if len(curve) == 0 {
		return 0
	}

	peak := curve[0].Value
	maxDD := 0.0

	for _, pt := range curve {
		if pt.Value > peak {
			peak = pt.Value
		}
		dd := (peak - pt.Value) / peak * 100
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// calculateSharpeFromTrades computes a simplified Sharpe ratio from trade returns.
func calculateSharpeFromTrades(trades []SignalBacktestTrade) float64 {
	if len(trades) < 2 {
		return 0
	}

	// Calculate mean return
	var sum float64
	for _, t := range trades {
		sum += t.ReturnPct
	}
	mean := sum / float64(len(trades))

	// Calculate standard deviation
	var variance float64
	for _, t := range trades {
		diff := t.ReturnPct - mean
		variance += diff * diff
	}
	variance /= float64(len(trades) - 1)
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0
	}

	// Annualize (assume ~50 trades per year)
	annualizedReturn := mean * 50
	annualizedStdDev := stdDev * math.Sqrt(50)

	return annualizedReturn / annualizedStdDev
}

// OptimizeWeights runs a grid search over factor weights to find optimal configuration.
func (b *SignalBacktester) OptimizeWeights(ctx context.Context, req SignalBacktestRequest) ([]SignalBacktestResult, error) {
	var results []SignalBacktestResult

	// Grid search over weight combinations (must sum to 1.0)
	weightOptions := []float64{0.1, 0.2, 0.3, 0.4}

	for _, mw := range weightOptions {
		for _, tw := range weightOptions {
			for _, vw := range weightOptions {
				for _, fw := range weightOptions {
					sw := 1.0 - mw - tw - vw - fw
					if sw < 0.05 || sw > 0.5 {
						continue
					}

					testConfig := model.SignalConfig{
						MomentumWeight:    mw,
						TrendWeight:       tw,
						VolumeWeight:      vw,
						FundamentalWeight: fw,
						SectorWeight:      sw,
						MinCompositeScore: 60,
						TopN:              req.TopN,
						ATRMultiplierSL:   2.0,
						RiskRewardRatio:   2.0,
						RSIPeriod:         14,
						MACDFast:          12,
						MACDSlow:          26,
						MACDSignal:        9,
						ADXPeriod:         14,
						ATRPeriod:         14,
						OBVSMAPeriod:      20,
						PriceSMAPeriod:    20,
						VolumeLookback:    20,
					}

					testReq := req
					testReq.Config = testConfig

					result, err := b.RunBacktest(ctx, testReq)
					if err != nil {
						continue
					}

					results = append(results, result)
				}
			}
		}
	}

	// Sort by Sharpe ratio (or total return)
	sort.Slice(results, func(i, j int) bool {
		return results[i].SharpeRatio > results[j].SharpeRatio
	})

	// Return top 10 configurations
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}
