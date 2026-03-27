package service

// ---------------------------------------------------------------------------
// Pattern_Detector — identifies market patterns (accumulation, distribution, breakout)
// ---------------------------------------------------------------------------
//
// Requirements satisfied:
//   - 12.3 Detect accumulation patterns (price consolidation 5% for 10+ days,
//          volume >1.5x avg, institutional buying)
//   - 12.4 Detect distribution patterns (price near highs, volume on down days,
//          institutional selling)
//   - 12.5 Detect breakout signals (price above resistance, volume >2x avg)
//   - 12.6 Generate confidence scores 0-100

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"myfi-backend/internal/model"
)

// PatternDetector identifies market patterns from OHLCV data.
type PatternDetector struct {
	// Configuration thresholds
	consolidationRangePercent float64 // Max price range for consolidation (default 5%)
	consolidationMinDays      int     // Min days for consolidation (default 10)
	volumeAccumThreshold      float64 // Volume ratio for accumulation (default 1.5x)
	volumeBreakoutThreshold   float64 // Volume ratio for breakout (default 2.0x)
	nearHighThreshold         float64 // % from high to be "near high" (default 5%)
	resistanceLookbackDays    int     // Days to look back for resistance (default 20)
}

// NewPatternDetector creates a PatternDetector with default configuration.
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{
		consolidationRangePercent: 5.0,
		consolidationMinDays:      10,
		volumeAccumThreshold:      1.5,
		volumeBreakoutThreshold:   2.0,
		nearHighThreshold:         5.0,
		resistanceLookbackDays:    20,
	}
}

// DetectPatterns analyzes OHLCV data and returns all detected patterns.
func (p *PatternDetector) DetectPatterns(symbol string, bars []model.OHLCVBar) []model.PatternObservation {
	if len(bars) < 20 {
		return nil // Need at least 20 days of data
	}

	var observations []model.PatternObservation

	// Sort bars by time ascending
	sortedBars := make([]model.OHLCVBar, len(bars))
	copy(sortedBars, bars)
	sort.Slice(sortedBars, func(i, j int) bool {
		return sortedBars[i].Time.Before(sortedBars[j].Time)
	})

	// Detect each pattern type
	if obs := p.DetectAccumulation(symbol, sortedBars); obs != nil {
		observations = append(observations, *obs)
	}
	if obs := p.DetectDistribution(symbol, sortedBars); obs != nil {
		observations = append(observations, *obs)
	}
	if obs := p.DetectBreakout(symbol, sortedBars); obs != nil {
		observations = append(observations, *obs)
	}

	return observations
}

// DetectAccumulation detects accumulation patterns (gom hàng).
// Criteria:
//   - Price consolidation within 5% range for 10+ days
//   - Volume >1.5x 20-day average
//   - Signs of institutional buying (large block trades)
//
// Requirement: 12.3
func (p *PatternDetector) DetectAccumulation(symbol string, bars []model.OHLCVBar) *model.PatternObservation {
	if len(bars) < p.consolidationMinDays+10 {
		return nil
	}

	// Use the most recent bars for analysis
	recentBars := bars[len(bars)-p.consolidationMinDays:]
	lookbackBars := bars[len(bars)-30:] // 30 days for volume average

	// Calculate price range over consolidation period
	var high, low float64 = recentBars[0].High, recentBars[0].Low
	for _, bar := range recentBars {
		if bar.High > high {
			high = bar.High
		}
		if bar.Low < low {
			low = bar.Low
		}
	}

	// Calculate price range percentage
	midPrice := (high + low) / 2
	if midPrice == 0 {
		return nil
	}
	priceRangePercent := ((high - low) / midPrice) * 100

	// Check if price is consolidating within threshold
	isConsolidating := priceRangePercent <= p.consolidationRangePercent

	// Calculate 20-day average volume
	avgVolume20Day := p.calculateAvgVolume(lookbackBars, 20)
	if avgVolume20Day == 0 {
		return nil
	}

	// Calculate recent average volume (last 5 days)
	recentAvgVolume := p.calculateAvgVolume(recentBars[len(recentBars)-5:], 5)
	volumeRatio := recentAvgVolume / avgVolume20Day

	// Check for institutional buying (simplified: large volume on up days)
	institutionalBuying := p.detectInstitutionalBuying(recentBars)

	// Calculate confidence score
	confidence := p.calculateAccumulationConfidence(
		isConsolidating,
		priceRangePercent,
		volumeRatio,
		institutionalBuying,
	)

	// Only return observation if confidence meets minimum threshold
	if confidence < 30 {
		return nil
	}

	// Build supporting data
	supportingData := model.AccumulationData{
		ConsolidationDays:   len(recentBars),
		PriceRangePercent:   math.Round(priceRangePercent*100) / 100,
		VolumeRatio:         math.Round(volumeRatio*100) / 100,
		AvgVolume20Day:      avgVolume20Day,
		CurrentVolume:       recentAvgVolume,
		InstitutionalBuying: institutionalBuying,
		PriceHigh:           high,
		PriceLow:            low,
	}

	supportingJSON, _ := json.Marshal(supportingData)
	currentPrice := bars[len(bars)-1].Close

	return &model.PatternObservation{
		Symbol:           symbol,
		PatternType:      model.PatternAccumulation,
		DetectionDate:    time.Now(),
		ConfidenceScore:  confidence,
		PriceAtDetection: currentPrice,
		SupportingData:   string(supportingJSON),
	}
}

// calculateAccumulationConfidence computes confidence score for accumulation pattern.
func (p *PatternDetector) calculateAccumulationConfidence(
	isConsolidating bool,
	priceRangePercent float64,
	volumeRatio float64,
	institutionalBuying bool,
) int {
	var score float64

	// Consolidation score (0-40 points)
	if isConsolidating {
		// Tighter consolidation = higher score
		consolidationScore := 40 * (1 - (priceRangePercent / p.consolidationRangePercent))
		score += math.Max(0, consolidationScore)
	}

	// Volume score (0-35 points)
	if volumeRatio >= p.volumeAccumThreshold {
		// Higher volume ratio = higher score, capped at 3x
		volumeScore := 35 * math.Min((volumeRatio-1)/(3-1), 1)
		score += volumeScore
	}

	// Institutional buying score (0-25 points)
	if institutionalBuying {
		score += 25
	}

	return int(math.Min(100, math.Max(0, score)))
}

// DetectDistribution detects distribution patterns.
// Criteria:
//   - Price near recent highs
//   - Higher volume on down days
//   - Signs of institutional selling
//
// Requirement: 12.4
func (p *PatternDetector) DetectDistribution(symbol string, bars []model.OHLCVBar) *model.PatternObservation {
	if len(bars) < 20 {
		return nil
	}

	// Use recent 20 days for analysis
	recentBars := bars[len(bars)-20:]
	currentBar := bars[len(bars)-1]

	// Find recent high (last 20 days)
	var recentHigh float64
	for _, bar := range recentBars {
		if bar.High > recentHigh {
			recentHigh = bar.High
		}
	}

	if recentHigh == 0 {
		return nil
	}

	// Calculate how close current price is to recent high
	currentPrice := currentBar.Close
	priceFromHigh := ((recentHigh - currentPrice) / recentHigh) * 100
	isNearHigh := priceFromHigh <= p.nearHighThreshold

	// Calculate volume on down days vs up days
	var downDayVolume, upDayVolume float64
	var downDays, upDays int

	for i := 1; i < len(recentBars); i++ {
		if recentBars[i].Close < recentBars[i-1].Close {
			downDayVolume += float64(recentBars[i].Volume)
			downDays++
		} else {
			upDayVolume += float64(recentBars[i].Volume)
			upDays++
		}
	}

	// Calculate average volume on down days vs up days
	var avgDownDayVolume, avgUpDayVolume float64
	if downDays > 0 {
		avgDownDayVolume = downDayVolume / float64(downDays)
	}
	if upDays > 0 {
		avgUpDayVolume = upDayVolume / float64(upDays)
	}

	// Check if volume is higher on down days
	var downDayVolumeRatio float64
	if avgUpDayVolume > 0 {
		downDayVolumeRatio = avgDownDayVolume / avgUpDayVolume
	}
	hasHigherDownDayVolume := downDayVolumeRatio > 1.2

	// Check for institutional selling (simplified: large volume on down days)
	institutionalSelling := p.detectInstitutionalSelling(recentBars)

	// Calculate confidence score
	confidence := p.calculateDistributionConfidence(
		isNearHigh,
		priceFromHigh,
		downDayVolumeRatio,
		hasHigherDownDayVolume,
		institutionalSelling,
	)

	// Only return observation if confidence meets minimum threshold
	if confidence < 30 {
		return nil
	}

	// Build supporting data
	supportingData := model.DistributionData{
		PriceNearHighPercent: math.Round((100-priceFromHigh)*100) / 100,
		RecentHigh:           recentHigh,
		CurrentPrice:         currentPrice,
		DownDayVolumeRatio:   math.Round(downDayVolumeRatio*100) / 100,
		DownDaysCount:        downDays,
		InstitutionalSelling: institutionalSelling,
	}

	supportingJSON, _ := json.Marshal(supportingData)

	return &model.PatternObservation{
		Symbol:           symbol,
		PatternType:      model.PatternDistribution,
		DetectionDate:    time.Now(),
		ConfidenceScore:  confidence,
		PriceAtDetection: currentPrice,
		SupportingData:   string(supportingJSON),
	}
}

// calculateDistributionConfidence computes confidence score for distribution pattern.
func (p *PatternDetector) calculateDistributionConfidence(
	isNearHigh bool,
	priceFromHigh float64,
	downDayVolumeRatio float64,
	hasHigherDownDayVolume bool,
	institutionalSelling bool,
) int {
	var score float64

	// Near high score (0-35 points)
	if isNearHigh {
		// Closer to high = higher score
		nearHighScore := 35 * (1 - (priceFromHigh / p.nearHighThreshold))
		score += math.Max(0, nearHighScore)
	}

	// Down day volume score (0-40 points)
	if hasHigherDownDayVolume {
		// Higher ratio = higher score, capped at 2x
		volumeScore := 40 * math.Min((downDayVolumeRatio-1)/(2-1), 1)
		score += volumeScore
	}

	// Institutional selling score (0-25 points)
	if institutionalSelling {
		score += 25
	}

	return int(math.Min(100, math.Max(0, score)))
}

// DetectBreakout detects breakout signals.
// Criteria:
//   - Price breaking above resistance levels
//   - Volume >2x 20-day average
//
// Requirement: 12.5
func (p *PatternDetector) DetectBreakout(symbol string, bars []model.OHLCVBar) *model.PatternObservation {
	if len(bars) < p.resistanceLookbackDays+5 {
		return nil
	}

	// Get lookback period for resistance calculation (excluding last 5 days)
	lookbackEnd := len(bars) - 5
	lookbackStart := lookbackEnd - p.resistanceLookbackDays
	if lookbackStart < 0 {
		lookbackStart = 0
	}
	lookbackBars := bars[lookbackStart:lookbackEnd]
	recentBars := bars[len(bars)-5:]
	currentBar := bars[len(bars)-1]

	// Calculate resistance level (highest high in lookback period)
	resistanceLevel := p.calculateResistanceLevel(lookbackBars)
	if resistanceLevel == 0 {
		return nil
	}

	// Check if current price is above resistance
	currentPrice := currentBar.Close
	breakoutPercent := ((currentPrice - resistanceLevel) / resistanceLevel) * 100
	isBreakout := breakoutPercent > 0

	// Calculate 20-day average volume
	volumeLookback := bars[len(bars)-25 : len(bars)-5]
	avgVolume20Day := p.calculateAvgVolume(volumeLookback, 20)
	if avgVolume20Day == 0 {
		return nil
	}

	// Calculate current volume (average of last 3 days for confirmation)
	currentVolume := p.calculateAvgVolume(recentBars[len(recentBars)-3:], 3)
	volumeRatio := currentVolume / avgVolume20Day

	// Check for elevated volume
	hasBreakoutVolume := volumeRatio >= p.volumeBreakoutThreshold

	// Check for prior consolidation (adds to breakout validity)
	priorConsolidDays := p.countConsolidationDays(lookbackBars)

	// Calculate confidence score
	confidence := p.calculateBreakoutConfidence(
		isBreakout,
		breakoutPercent,
		volumeRatio,
		hasBreakoutVolume,
		priorConsolidDays,
	)

	// Only return observation if confidence meets minimum threshold
	if confidence < 30 {
		return nil
	}

	// Build supporting data
	supportingData := model.BreakoutData{
		ResistanceLevel:   resistanceLevel,
		BreakoutPrice:     currentPrice,
		BreakoutPercent:   math.Round(breakoutPercent*100) / 100,
		VolumeRatio:       math.Round(volumeRatio*100) / 100,
		AvgVolume20Day:    avgVolume20Day,
		CurrentVolume:     currentVolume,
		PriorConsolidDays: priorConsolidDays,
	}

	supportingJSON, _ := json.Marshal(supportingData)

	return &model.PatternObservation{
		Symbol:           symbol,
		PatternType:      model.PatternBreakout,
		DetectionDate:    time.Now(),
		ConfidenceScore:  confidence,
		PriceAtDetection: currentPrice,
		SupportingData:   string(supportingJSON),
	}
}

// calculateBreakoutConfidence computes confidence score for breakout pattern.
func (p *PatternDetector) calculateBreakoutConfidence(
	isBreakout bool,
	breakoutPercent float64,
	volumeRatio float64,
	hasBreakoutVolume bool,
	priorConsolidDays int,
) int {
	if !isBreakout {
		return 0
	}

	var score float64

	// Breakout magnitude score (0-30 points)
	// Higher breakout % = higher score, capped at 10%
	breakoutScore := 30 * math.Min(breakoutPercent/10, 1)
	score += breakoutScore

	// Volume score (0-45 points)
	if hasBreakoutVolume {
		// Higher volume ratio = higher score, capped at 4x
		volumeScore := 45 * math.Min((volumeRatio-1)/(4-1), 1)
		score += volumeScore
	}

	// Prior consolidation score (0-25 points)
	// Longer consolidation before breakout = more significant
	if priorConsolidDays >= 5 {
		consolidScore := 25 * math.Min(float64(priorConsolidDays)/15, 1)
		score += consolidScore
	}

	return int(math.Min(100, math.Max(0, score)))
}

// Helper functions

// calculateAvgVolume calculates the average volume over a period.
func (p *PatternDetector) calculateAvgVolume(bars []model.OHLCVBar, days int) float64 {
	if len(bars) == 0 {
		return 0
	}

	count := days
	if len(bars) < days {
		count = len(bars)
	}

	var totalVolume float64
	startIdx := len(bars) - count
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(bars); i++ {
		totalVolume += float64(bars[i].Volume)
	}

	return totalVolume / float64(count)
}

// calculateResistanceLevel finds the resistance level (highest high) in the lookback period.
func (p *PatternDetector) calculateResistanceLevel(bars []model.OHLCVBar) float64 {
	if len(bars) == 0 {
		return 0
	}

	var resistance float64
	for _, bar := range bars {
		if bar.High > resistance {
			resistance = bar.High
		}
	}
	return resistance
}

// countConsolidationDays counts consecutive days of price consolidation.
func (p *PatternDetector) countConsolidationDays(bars []model.OHLCVBar) int {
	if len(bars) < 5 {
		return 0
	}

	// Work backwards from the end to find consolidation period
	consolidDays := 0
	for windowSize := 5; windowSize <= len(bars); windowSize++ {
		window := bars[len(bars)-windowSize:]

		var high, low float64 = window[0].High, window[0].Low
		for _, bar := range window {
			if bar.High > high {
				high = bar.High
			}
			if bar.Low < low {
				low = bar.Low
			}
		}

		midPrice := (high + low) / 2
		if midPrice == 0 {
			break
		}

		rangePercent := ((high - low) / midPrice) * 100
		if rangePercent <= p.consolidationRangePercent {
			consolidDays = windowSize
		} else {
			break
		}
	}

	return consolidDays
}

// detectInstitutionalBuying detects signs of institutional buying.
// Simplified heuristic: large volume on up days with price closing near high.
func (p *PatternDetector) detectInstitutionalBuying(bars []model.OHLCVBar) bool {
	if len(bars) < 5 {
		return false
	}

	// Calculate average volume
	avgVolume := p.calculateAvgVolume(bars, len(bars))
	if avgVolume == 0 {
		return false
	}

	// Count days with institutional buying characteristics
	institutionalDays := 0
	for i := 1; i < len(bars); i++ {
		bar := bars[i]
		prevBar := bars[i-1]

		// Up day with high volume and close near high
		isUpDay := bar.Close > prevBar.Close
		hasHighVolume := float64(bar.Volume) > avgVolume*1.5
		closeNearHigh := (bar.High - bar.Close) < (bar.Close - bar.Low) // Close in upper half

		if isUpDay && hasHighVolume && closeNearHigh {
			institutionalDays++
		}
	}

	// Consider institutional buying if 30%+ of days show the pattern
	return float64(institutionalDays)/float64(len(bars)-1) >= 0.3
}

// detectInstitutionalSelling detects signs of institutional selling.
// Simplified heuristic: large volume on down days with price closing near low.
func (p *PatternDetector) detectInstitutionalSelling(bars []model.OHLCVBar) bool {
	if len(bars) < 5 {
		return false
	}

	// Calculate average volume
	avgVolume := p.calculateAvgVolume(bars, len(bars))
	if avgVolume == 0 {
		return false
	}

	// Count days with institutional selling characteristics
	institutionalDays := 0
	for i := 1; i < len(bars); i++ {
		bar := bars[i]
		prevBar := bars[i-1]

		// Down day with high volume and close near low
		isDownDay := bar.Close < prevBar.Close
		hasHighVolume := float64(bar.Volume) > avgVolume*1.5
		closeNearLow := (bar.Close - bar.Low) < (bar.High - bar.Close) // Close in lower half

		if isDownDay && hasHighVolume && closeNearLow {
			institutionalDays++
		}
	}

	// Consider institutional selling if 30%+ of days show the pattern
	return float64(institutionalDays)/float64(len(bars)-1) >= 0.3
}

// SetConsolidationRangePercent sets the max price range for consolidation detection.
func (p *PatternDetector) SetConsolidationRangePercent(percent float64) {
	p.consolidationRangePercent = percent
}

// SetConsolidationMinDays sets the minimum days for consolidation detection.
func (p *PatternDetector) SetConsolidationMinDays(days int) {
	p.consolidationMinDays = days
}

// SetVolumeAccumThreshold sets the volume ratio threshold for accumulation.
func (p *PatternDetector) SetVolumeAccumThreshold(ratio float64) {
	p.volumeAccumThreshold = ratio
}

// SetVolumeBreakoutThreshold sets the volume ratio threshold for breakout.
func (p *PatternDetector) SetVolumeBreakoutThreshold(ratio float64) {
	p.volumeBreakoutThreshold = ratio
}

// SetNearHighThreshold sets the % from high to be considered "near high".
func (p *PatternDetector) SetNearHighThreshold(percent float64) {
	p.nearHighThreshold = percent
}

// SetResistanceLookbackDays sets the days to look back for resistance calculation.
func (p *PatternDetector) SetResistanceLookbackDays(days int) {
	p.resistanceLookbackDays = days
}

// ---------------------------------------------------------------------------
// Technical Indicators (ported from vnquant)
// ---------------------------------------------------------------------------

// TechnicalIndicators holds calculated technical indicators for a symbol.
type TechnicalIndicators struct {
	Symbol    string
	Timestamp time.Time
	MACD      MACDResult
	RSI       float64
	EMA12     float64
	EMA26     float64
}

// MACDResult holds MACD indicator values.
type MACDResult struct {
	MACD      float64 // MACD line (12-day EMA - 26-day EMA)
	Signal    float64 // Signal line (9-day EMA of MACD)
	Histogram float64 // MACD - Signal
}

// CalculateMACD computes MACD indicator from OHLCV bars.
// Uses standard 12/26/9 periods as in vnquant.
// Requires at least 35 bars for meaningful results.
func (p *PatternDetector) CalculateMACD(bars []model.OHLCVBar) MACDResult {
	if len(bars) < 35 {
		return MACDResult{}
	}

	// Extract close prices
	closes := make([]float64, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
	}

	// Calculate 12-day EMA
	ema12 := p.calculateEMA(closes, 12)
	// Calculate 26-day EMA
	ema26 := p.calculateEMA(closes, 26)

	// MACD line = 12-day EMA - 26-day EMA
	macdLine := make([]float64, len(closes))
	for i := range closes {
		macdLine[i] = ema12[i] - ema26[i]
	}

	// Signal line = 9-day EMA of MACD
	signal := p.calculateEMA(macdLine, 9)

	// Get latest values
	lastIdx := len(closes) - 1
	macd := macdLine[lastIdx]
	sig := signal[lastIdx]

	return MACDResult{
		MACD:      math.Round(macd*1000) / 1000,
		Signal:    math.Round(sig*1000) / 1000,
		Histogram: math.Round((macd-sig)*1000) / 1000,
	}
}

// CalculateRSI computes RSI indicator from OHLCV bars.
// Uses 14-period RSI with EMA smoothing as in vnquant.
// Returns RSI value between 0-100.
func (p *PatternDetector) CalculateRSI(bars []model.OHLCVBar, period int) float64 {
	if period <= 0 {
		period = 14 // Default RSI period
	}
	if len(bars) < period+1 {
		return 50 // Neutral if insufficient data
	}

	// Extract close prices
	closes := make([]float64, len(bars))
	for i, bar := range bars {
		closes[i] = bar.Close
	}

	// Calculate price changes (delta)
	deltas := make([]float64, len(closes))
	for i := 1; i < len(closes); i++ {
		deltas[i] = closes[i] - closes[i-1]
	}

	// Separate gains and losses
	gains := make([]float64, len(deltas))
	losses := make([]float64, len(deltas))
	for i, d := range deltas {
		if d > 0 {
			gains[i] = d
		} else {
			losses[i] = -d // Make losses positive
		}
	}

	// Calculate EMA of gains and losses (com=period-1 for pandas compatibility)
	emaGains := p.calculateEMAWithCom(gains, period-1)
	emaLosses := p.calculateEMAWithCom(losses, period-1)

	// Calculate RS and RSI
	lastIdx := len(closes) - 1
	avgGain := emaGains[lastIdx]
	avgLoss := emaLosses[lastIdx]

	if avgLoss == 0 {
		return 100 // All gains, no losses
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return math.Round(rsi*100) / 100
}

// CalculateTechnicalIndicators computes all technical indicators for a symbol.
func (p *PatternDetector) CalculateTechnicalIndicators(symbol string, bars []model.OHLCVBar) TechnicalIndicators {
	if len(bars) == 0 {
		return TechnicalIndicators{Symbol: symbol}
	}

	// Sort bars by time ascending
	sortedBars := make([]model.OHLCVBar, len(bars))
	copy(sortedBars, bars)
	sort.Slice(sortedBars, func(i, j int) bool {
		return sortedBars[i].Time.Before(sortedBars[j].Time)
	})

	return TechnicalIndicators{
		Symbol:    symbol,
		Timestamp: sortedBars[len(sortedBars)-1].Time,
		MACD:      p.CalculateMACD(sortedBars),
		RSI:       p.CalculateRSI(sortedBars, 14),
	}
}

// calculateEMA computes Exponential Moving Average with given span.
// Uses adjust=False for pandas compatibility.
func (p *PatternDetector) calculateEMA(data []float64, span int) []float64 {
	if len(data) == 0 || span <= 0 {
		return data
	}

	alpha := 2.0 / float64(span+1)
	ema := make([]float64, len(data))
	ema[0] = data[0]

	for i := 1; i < len(data); i++ {
		ema[i] = alpha*data[i] + (1-alpha)*ema[i-1]
	}

	return ema
}

// calculateEMAWithCom computes EMA with center of mass parameter (pandas ewm com).
// alpha = 1 / (1 + com)
func (p *PatternDetector) calculateEMAWithCom(data []float64, com int) []float64 {
	if len(data) == 0 || com < 0 {
		return data
	}

	alpha := 1.0 / float64(1+com)
	ema := make([]float64, len(data))
	ema[0] = data[0]

	for i := 1; i < len(data); i++ {
		ema[i] = alpha*data[i] + (1-alpha)*ema[i-1]
	}

	return ema
}

// DetectMACDCrossover detects MACD crossover signals.
// Returns "bullish" if MACD crosses above signal, "bearish" if below, "" if no crossover.
func (p *PatternDetector) DetectMACDCrossover(bars []model.OHLCVBar) string {
	if len(bars) < 36 {
		return ""
	}

	// Calculate MACD for last 2 periods
	prevBars := bars[:len(bars)-1]
	currBars := bars

	prevMACD := p.CalculateMACD(prevBars)
	currMACD := p.CalculateMACD(currBars)

	// Detect crossover
	prevDiff := prevMACD.MACD - prevMACD.Signal
	currDiff := currMACD.MACD - currMACD.Signal

	if prevDiff <= 0 && currDiff > 0 {
		return "bullish" // MACD crossed above signal
	}
	if prevDiff >= 0 && currDiff < 0 {
		return "bearish" // MACD crossed below signal
	}

	return ""
}

// DetectRSISignal detects RSI overbought/oversold conditions.
// Returns "oversold" if RSI < 30, "overbought" if RSI > 70, "" otherwise.
func (p *PatternDetector) DetectRSISignal(bars []model.OHLCVBar) string {
	rsi := p.CalculateRSI(bars, 14)

	if rsi < 30 {
		return "oversold"
	}
	if rsi > 70 {
		return "overbought"
	}

	return ""
}
