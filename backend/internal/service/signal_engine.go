package service

import (
	"context"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"myfi-backend/internal/infra"
	"myfi-backend/internal/model"

	vnstock "github.com/dda10/vnstock-go"
)

// SignalEngine generates systematic trading signals by scanning the market,
// computing composite scores from multiple factors, and outputting ranked signals.
type SignalEngine struct {
	router          *infra.DataSourceRouter
	liquidityFilter *LiquidityFilter
	sectorService   *SectorService
	screenerService *ScreenerService
	recTracker      *RecommendationTracker
	config          model.SignalConfig
}

// NewSignalEngine creates a new signal engine with default config.
func NewSignalEngine(
	router *infra.DataSourceRouter,
	liquidityFilter *LiquidityFilter,
	sectorService *SectorService,
	screenerService *ScreenerService,
) *SignalEngine {
	return &SignalEngine{
		router:          router,
		liquidityFilter: liquidityFilter,
		sectorService:   sectorService,
		screenerService: screenerService,
		config:          model.DefaultSignalConfig(),
	}
}

// SetConfig updates the signal generation configuration.
func (e *SignalEngine) SetConfig(cfg model.SignalConfig) {
	e.config = cfg
}

// SetRecommendationTracker attaches a tracker for logging signals.
func (e *SignalEngine) SetRecommendationTracker(rt *RecommendationTracker) {
	e.recTracker = rt
}

// ScanMarket scans all whitelisted stocks and generates ranked signals.
func (e *SignalEngine) ScanMarket(ctx context.Context) (model.SignalScanResult, error) {
	start := time.Now()

	// Step 1: Get whitelist
	whitelist := e.liquidityFilter.GetWhitelist()
	symbols := make([]string, 0, len(whitelist.Entries))
	symbolExchange := make(map[string]string)
	for _, entry := range whitelist.Entries {
		symbols = append(symbols, entry.Symbol)
		symbolExchange[entry.Symbol] = entry.Exchange
	}

	log.Printf("[SignalEngine] Scanning %d whitelisted stocks", len(symbols))

	// Step 2: Fetch sector performances for context
	sectorPerfs, _ := e.sectorService.GetAllSectorPerformances(ctx)
	sectorTrendMap := make(map[model.ICBSector]model.SectorTrend)
	sectorChangeMap := make(map[model.ICBSector]float64)
	for _, p := range sectorPerfs {
		sectorTrendMap[p.Sector] = p.Trend
		sectorChangeMap[p.Sector] = p.OneMonthChange
	}

	// Step 3: Score each stock concurrently
	type scoredStock struct {
		signal model.TradingSignal
		err    error
	}

	resultsCh := make(chan scoredStock, len(symbols))
	sem := make(chan struct{}, 3) // limit concurrency to avoid VCI rate limits
	var wg sync.WaitGroup

	for i, sym := range symbols {
		wg.Add(1)
		go func(symbol string, idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Stagger requests to avoid rate limiting
			if idx > 0 {
				time.Sleep(time.Duration(idx%3) * 100 * time.Millisecond)
			}

			signal, err := e.scoreStock(ctx, symbol, symbolExchange[symbol], sectorTrendMap, sectorChangeMap)
			resultsCh <- scoredStock{signal: signal, err: err}
		}(sym, i)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Step 4: Collect and filter results
	var allSignals []model.TradingSignal
	passedFilter := 0

	for res := range resultsCh {
		if res.err != nil {
			continue
		}
		if res.signal.CompositeScore >= e.config.MinCompositeScore {
			allSignals = append(allSignals, res.signal)
			passedFilter++
		}
	}

	// Step 5: Sort by composite score and take top N
	sort.Slice(allSignals, func(i, j int) bool {
		return allSignals[i].CompositeScore > allSignals[j].CompositeScore
	})

	topN := e.config.TopN
	if topN > len(allSignals) {
		topN = len(allSignals)
	}
	signals := allSignals[:topN]

	// Log signals to recommendation tracker for accuracy tracking
	if e.recTracker != nil {
		for _, sig := range signals {
			rec := model.AssetRecommendation{
				Symbol:         sig.Symbol,
				Action:         string(sig.Direction), // "long" -> treat as "buy"
				PositionSize:   5.0,                   // default 5% position
				RiskAssessment: string(sig.Strength),
				Reasoning:      joinReasons(sig.Reasoning),
			}
			confidence := int(sig.CompositeScore)
			if _, err := e.recTracker.LogRecommendation(ctx, rec, confidence); err != nil {
				log.Printf("[SignalEngine] Failed to log recommendation for %s: %v", sig.Symbol, err)
			}
		}
	}

	log.Printf("[SignalEngine] Scan complete: %d scanned, %d passed filter, %d signals generated (took %v)",
		len(symbols), passedFilter, len(signals), time.Since(start))

	return model.SignalScanResult{
		Signals:      signals,
		TotalScanned: len(symbols),
		PassedFilter: passedFilter,
		GeneratedAt:  time.Now(),
		Config:       e.config,
	}, nil
}

// scoreStock computes the composite score and generates a signal for one stock.
func (e *SignalEngine) scoreStock(
	ctx context.Context,
	symbol string,
	exchange string,
	sectorTrendMap map[model.ICBSector]model.SectorTrend,
	sectorChangeMap map[model.ICBSector]float64,
) (model.TradingSignal, error) {
	signal := model.TradingSignal{
		Symbol:      symbol,
		Exchange:    exchange,
		GeneratedAt: time.Now(),
	}

	// Fetch OHLCV data (60 days for indicator computation)
	end := time.Now()
	start := end.AddDate(0, 0, -90)
	bars, err := e.fetchOHLCV(ctx, symbol, start, end)
	if err != nil || len(bars) < 30 {
		return signal, err
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
	rsi := computeRSI(closes, e.config.RSIPeriod)
	macdLine, macdSignal, _ := computeMACD(closes, e.config.MACDFast, e.config.MACDSlow, e.config.MACDSignal)
	adx := computeADX(highs, lows, closes, e.config.ADXPeriod)
	atr := computeATR(highs, lows, closes, e.config.ATRPeriod)
	obv := computeOBV(closes, volumes)
	obvSMA := computeSMA(obv, e.config.OBVSMAPeriod)
	priceSMA := computeSMA(closes, e.config.PriceSMAPeriod)

	// Get sector info
	sector, _ := e.sectorService.GetStockSector(symbol)
	signal.Sector = sector
	signal.SectorName = model.SectorNameMap[sector]

	// --- Factor 1: Momentum (0-100) ---
	var momentumScore float64
	var momentumReasons []string

	// RSI component (oversold = bullish, overbought = bearish for mean reversion)
	rsiVal := safeGet(rsi, lastIdx)
	if rsiVal < 30 {
		momentumScore += 40
		momentumReasons = append(momentumReasons, "RSI oversold")
	} else if rsiVal < 50 {
		momentumScore += 25
	} else if rsiVal < 70 {
		momentumScore += 15
	}

	// MACD component (bullish crossover)
	macdVal := safeGet(macdLine, lastIdx)
	macdSigVal := safeGet(macdSignal, lastIdx)
	macdPrev := safeGet(macdLine, lastIdx-1)
	macdSigPrev := safeGet(macdSignal, lastIdx-1)
	if macdVal > macdSigVal && macdPrev <= macdSigPrev {
		momentumScore += 40
		momentumReasons = append(momentumReasons, "MACD bullish crossover")
	} else if macdVal > macdSigVal {
		momentumScore += 20
	}

	// Price vs SMA
	smaVal := safeGet(priceSMA, lastIdx)
	if smaVal > 0 && closes[lastIdx] > smaVal {
		momentumScore += 20
		momentumReasons = append(momentumReasons, "Price above SMA20")
	}

	signal.Factors.Momentum = clamp(momentumScore, 0, 100)

	// --- Factor 2: Trend (0-100) ---
	var trendScore float64
	var trendReasons []string

	adxVal := safeGet(adx, lastIdx)
	if adxVal > 25 {
		trendScore += 50
		trendReasons = append(trendReasons, "Strong trend (ADX>25)")
	} else if adxVal > 20 {
		trendScore += 30
	}

	// Price trend (higher highs, higher lows over last 10 bars)
	if lastIdx >= 10 {
		higherHighs := closes[lastIdx] > closes[lastIdx-5] && closes[lastIdx-5] > closes[lastIdx-10]
		if higherHighs {
			trendScore += 50
			trendReasons = append(trendReasons, "Uptrend pattern")
		}
	}

	signal.Factors.Trend = clamp(trendScore, 0, 100)

	// --- Factor 3: Volume (0-100) ---
	var volumeScore float64
	var volumeReasons []string

	// OBV trend (rising OBV = accumulation)
	obvVal := safeGet(obv, lastIdx)
	obvSMAVal := safeGet(obvSMA, lastIdx)
	if obvSMAVal > 0 && obvVal > obvSMAVal {
		volumeScore += 50
		volumeReasons = append(volumeReasons, "OBV above average (accumulation)")
	}

	// Volume spike
	if lastIdx >= e.config.VolumeLookback {
		avgVol := 0.0
		for i := lastIdx - e.config.VolumeLookback; i < lastIdx; i++ {
			avgVol += volumes[i]
		}
		avgVol /= float64(e.config.VolumeLookback)
		if avgVol > 0 && volumes[lastIdx] > avgVol*1.5 {
			volumeScore += 50
			volumeReasons = append(volumeReasons, "Volume spike (>1.5x avg)")
		}
	}

	signal.Factors.Volume = clamp(volumeScore, 0, 100)

	// --- Factor 4: Fundamental (0-100) ---
	var fundamentalScore float64
	var fundamentalReasons []string

	// Fetch fundamental data from screener cache
	fundData := e.getFundamentalData(ctx, symbol)
	if fundData != nil {
		// P/E score (lower is better, but not negative)
		if fundData.PE > 0 && fundData.PE < 15 {
			fundamentalScore += 40
			fundamentalReasons = append(fundamentalReasons, "Low P/E (<15)")
		} else if fundData.PE > 0 && fundData.PE < 25 {
			fundamentalScore += 20
		}

		// ROE score
		if fundData.ROE > 15 {
			fundamentalScore += 30
			fundamentalReasons = append(fundamentalReasons, "High ROE (>15%)")
		} else if fundData.ROE > 10 {
			fundamentalScore += 15
		}

		// Profit growth
		if fundData.ProfitGrowth > 10 {
			fundamentalScore += 30
			fundamentalReasons = append(fundamentalReasons, "Profit growth >10%")
		} else if fundData.ProfitGrowth > 0 {
			fundamentalScore += 15
		}
	}

	signal.Factors.Fundamental = clamp(fundamentalScore, 0, 100)

	// --- Factor 5: Sector (0-100) ---
	var sectorScore float64
	var sectorReasons []string

	sectorTrend := sectorTrendMap[sector]
	sectorChange := sectorChangeMap[sector]

	if sectorTrend == model.Uptrend {
		sectorScore += 50
		sectorReasons = append(sectorReasons, "Sector in uptrend")
	} else if sectorTrend == model.Sideways {
		sectorScore += 25
	}

	if sectorChange > 5 {
		sectorScore += 50
		sectorReasons = append(sectorReasons, "Sector momentum positive")
	} else if sectorChange > 0 {
		sectorScore += 25
	}

	signal.Factors.Sector = clamp(sectorScore, 0, 100)

	// --- Composite Score ---
	signal.CompositeScore = signal.Factors.Momentum*e.config.MomentumWeight +
		signal.Factors.Trend*e.config.TrendWeight +
		signal.Factors.Volume*e.config.VolumeWeight +
		signal.Factors.Fundamental*e.config.FundamentalWeight +
		signal.Factors.Sector*e.config.SectorWeight

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
		signal.StopLoss = signal.EntryPrice - (atrVal * e.config.ATRMultiplierSL)
		slDistance := signal.EntryPrice - signal.StopLoss
		signal.TakeProfit = signal.EntryPrice + (slDistance * e.config.RiskRewardRatio)
		signal.RiskReward = e.config.RiskRewardRatio
	} else {
		// Fallback: 5% stop loss, 10% take profit
		signal.StopLoss = signal.EntryPrice * 0.95
		signal.TakeProfit = signal.EntryPrice * 1.10
		signal.RiskReward = 2.0
	}

	// --- Reasoning ---
	signal.Reasoning = append(signal.Reasoning, momentumReasons...)
	signal.Reasoning = append(signal.Reasoning, trendReasons...)
	signal.Reasoning = append(signal.Reasoning, volumeReasons...)
	signal.Reasoning = append(signal.Reasoning, fundamentalReasons...)
	signal.Reasoning = append(signal.Reasoning, sectorReasons...)

	return signal, nil
}

// fetchOHLCV retrieves historical price data for a symbol.
func (e *SignalEngine) fetchOHLCV(ctx context.Context, symbol string, start, end time.Time) ([]model.OHLCVBar, error) {
	req := vnstock.QuoteHistoryRequest{
		Symbol:   symbol,
		Start:    start,
		End:      end,
		Interval: "1D",
	}

	quotes, _, err := e.router.FetchQuoteHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	bars := make([]model.OHLCVBar, len(quotes))
	for i, q := range quotes {
		bars[i] = model.OHLCVBar{
			Time:   q.Timestamp,
			Open:   q.Open,
			High:   q.High,
			Low:    q.Low,
			Close:  q.Close,
			Volume: q.Volume,
		}
	}
	return bars, nil
}

// getFundamentalData retrieves cached fundamental data for a symbol.
func (e *SignalEngine) getFundamentalData(ctx context.Context, symbol string) *model.ScreenerResult {
	if e.screenerService == nil {
		return nil
	}

	// Use screener to get fundamental data for single stock
	filters := model.ScreenerFilters{
		Page:     1,
		PageSize: 1,
	}

	// This is inefficient for single stock lookup, but works
	// In production, you'd have a dedicated fundamental data cache
	results, _, err := e.screenerService.Screen(ctx, filters)
	if err != nil {
		return nil
	}

	for _, r := range results {
		if r.Symbol == symbol {
			return &r
		}
	}
	return nil
}

// --- Helper functions ---

func safeGet(arr []float64, idx int) float64 {
	if idx < 0 || idx >= len(arr) {
		return math.NaN()
	}
	return arr[idx]
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func joinReasons(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	result := reasons[0]
	for i := 1; i < len(reasons); i++ {
		result += "; " + reasons[i]
	}
	return result
}
