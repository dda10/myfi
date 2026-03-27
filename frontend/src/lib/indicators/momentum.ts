/**
 * Momentum indicator computation functions.
 * Pure math — no chart rendering logic.
 */

import type { Time } from "lightweight-charts";
import type { OHLCVBar, IndicatorPoint } from "./trend";

// ── Result types ───────────────────────────────────────────────────────────

export interface MACDResult {
  time: Time;
  macd: number;
  signal: number;
  histogram: number;
}

export interface StochasticResult {
  time: Time;
  k: number;
  d: number;
}

// ── RSI (Relative Strength Index) ──────────────────────────────────────────

/**
 * RSI using Wilder's smoothing method.
 * @param period Default 14
 */
export function computeRSI(bars: OHLCVBar[], period = 14): IndicatorPoint[] {
  if (bars.length < period + 1 || period < 1) return [];

  const result: IndicatorPoint[] = [];
  let gainSum = 0;
  let lossSum = 0;

  // Initial average gain/loss over first `period` changes
  for (let i = 1; i <= period; i++) {
    const change = bars[i].close - bars[i - 1].close;
    if (change > 0) gainSum += change;
    else lossSum += Math.abs(change);
  }

  let avgGain = gainSum / period;
  let avgLoss = lossSum / period;

  const rsi = avgLoss === 0 ? 100 : 100 - 100 / (1 + avgGain / avgLoss);
  result.push({ time: bars[period].time, value: rsi });

  // Wilder's smoothing for subsequent values
  for (let i = period + 1; i < bars.length; i++) {
    const change = bars[i].close - bars[i - 1].close;
    const gain = change > 0 ? change : 0;
    const loss = change < 0 ? Math.abs(change) : 0;
    avgGain = (avgGain * (period - 1) + gain) / period;
    avgLoss = (avgLoss * (period - 1) + loss) / period;
    const val = avgLoss === 0 ? 100 : 100 - 100 / (1 + avgGain / avgLoss);
    result.push({ time: bars[i].time, value: val });
  }

  return result;
}

// ── MACD (Moving Average Convergence Divergence) ───────────────────────────

/**
 * MACD with fast EMA, slow EMA, and signal EMA.
 * @param fastPeriod  Default 12
 * @param slowPeriod  Default 26
 * @param signalPeriod Default 9
 */
export function computeMACD(
  bars: OHLCVBar[],
  fastPeriod = 12,
  slowPeriod = 26,
  signalPeriod = 9,
): MACDResult[] {
  if (bars.length < slowPeriod + signalPeriod - 1 || fastPeriod < 1 || slowPeriod < 1 || signalPeriod < 1) return [];

  // Compute full EMA arrays over close prices
  const closes = bars.map((b) => b.close);

  const emaArr = (data: number[], period: number): number[] => {
    const k = 2 / (period + 1);
    let sum = 0;
    for (let i = 0; i < period; i++) sum += data[i];
    let ema = sum / period;
    const out: number[] = new Array(period - 1).fill(NaN);
    out.push(ema);
    for (let i = period; i < data.length; i++) {
      ema = data[i] * k + ema * (1 - k);
      out.push(ema);
    }
    return out;
  };

  const fastEMA = emaArr(closes, fastPeriod);
  const slowEMA = emaArr(closes, slowPeriod);

  // MACD line = fast EMA - slow EMA (valid from slowPeriod - 1 onward)
  const macdLine: number[] = [];
  const macdStartIdx = slowPeriod - 1;
  for (let i = macdStartIdx; i < bars.length; i++) {
    macdLine.push(fastEMA[i] - slowEMA[i]);
  }

  if (macdLine.length < signalPeriod) return [];

  // Signal line = EMA of MACD line
  const signalArr = emaArr(macdLine, signalPeriod);

  const result: MACDResult[] = [];
  const signalStartIdx = signalPeriod - 1;
  for (let i = signalStartIdx; i < macdLine.length; i++) {
    const barIdx = macdStartIdx + i;
    const m = macdLine[i];
    const s = signalArr[i];
    result.push({
      time: bars[barIdx].time,
      macd: m,
      signal: s,
      histogram: m - s,
    });
  }

  return result;
}

// ── Williams %R ────────────────────────────────────────────────────────────

/**
 * Williams %R: (highest high - close) / (highest high - lowest low) × -100.
 * Range: -100 to 0.
 * @param period Default 14
 */
export function computeWilliamsR(bars: OHLCVBar[], period = 14): IndicatorPoint[] {
  if (bars.length < period || period < 1) return [];

  const result: IndicatorPoint[] = [];
  for (let i = period - 1; i < bars.length; i++) {
    let hh = -Infinity;
    let ll = Infinity;
    for (let j = i - period + 1; j <= i; j++) {
      if (bars[j].high > hh) hh = bars[j].high;
      if (bars[j].low < ll) ll = bars[j].low;
    }
    const range = hh - ll;
    const value = range === 0 ? -50 : ((hh - bars[i].close) / range) * -100;
    result.push({ time: bars[i].time, value });
  }
  return result;
}


// ── CMO (Chande Momentum Oscillator) ───────────────────────────────────────

/**
 * CMO = (sum of gains - sum of losses) / (sum of gains + sum of losses) × 100.
 * Range: -100 to +100.
 * @param period Default 14
 */
export function computeCMO(bars: OHLCVBar[], period = 14): IndicatorPoint[] {
  if (bars.length < period + 1 || period < 1) return [];

  const result: IndicatorPoint[] = [];

  for (let i = period; i < bars.length; i++) {
    let gains = 0;
    let losses = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const change = bars[j].close - bars[j - 1].close;
      if (change > 0) gains += change;
      else losses += Math.abs(change);
    }
    const denom = gains + losses;
    const value = denom === 0 ? 0 : ((gains - losses) / denom) * 100;
    result.push({ time: bars[i].time, value });
  }

  return result;
}

// ── Stochastic Oscillator ──────────────────────────────────────────────────

/**
 * Stochastic Oscillator with %K and %D lines.
 * %K = (close - lowest low) / (highest high - lowest low) × 100
 * %D = SMA of %K over dPeriod
 * @param kPeriod  Lookback for %K (default 14)
 * @param dPeriod  SMA period for %D (default 3)
 * @param smooth   Smoothing period for %K (default 3)
 */
export function computeStochastic(
  bars: OHLCVBar[],
  kPeriod = 14,
  dPeriod = 3,
  smooth = 3,
): StochasticResult[] {
  if (bars.length < kPeriod || kPeriod < 1 || dPeriod < 1 || smooth < 1) return [];

  // Raw %K values
  const rawK: number[] = [];
  for (let i = kPeriod - 1; i < bars.length; i++) {
    let hh = -Infinity;
    let ll = Infinity;
    for (let j = i - kPeriod + 1; j <= i; j++) {
      if (bars[j].high > hh) hh = bars[j].high;
      if (bars[j].low < ll) ll = bars[j].low;
    }
    const range = hh - ll;
    rawK.push(range === 0 ? 50 : ((bars[i].close - ll) / range) * 100);
  }

  // Smooth %K with SMA
  const sma = (arr: number[], p: number): number[] => {
    if (arr.length < p) return [];
    const out: number[] = [];
    let sum = 0;
    for (let i = 0; i < p; i++) sum += arr[i];
    out.push(sum / p);
    for (let i = p; i < arr.length; i++) {
      sum += arr[i] - arr[i - p];
      out.push(sum / p);
    }
    return out;
  };

  const smoothedK = sma(rawK, smooth);
  if (smoothedK.length < dPeriod) return [];

  // %D = SMA of smoothed %K
  const dLine = sma(smoothedK, dPeriod);

  const result: StochasticResult[] = [];
  const kOffset = kPeriod - 1 + (smooth - 1);
  const dOffset = dPeriod - 1;

  for (let i = 0; i < dLine.length; i++) {
    const barIdx = kOffset + dOffset + i;
    if (barIdx < bars.length) {
      result.push({
        time: bars[barIdx].time,
        k: smoothedK[dOffset + i],
        d: dLine[i],
      });
    }
  }

  return result;
}

// ── ROC (Rate of Change) ───────────────────────────────────────────────────

/**
 * ROC = ((current close - close n periods ago) / close n periods ago) × 100.
 * @param period Default 12
 */
export function computeROC(bars: OHLCVBar[], period = 12): IndicatorPoint[] {
  if (bars.length <= period || period < 1) return [];

  const result: IndicatorPoint[] = [];
  for (let i = period; i < bars.length; i++) {
    const prev = bars[i - period].close;
    const value = prev === 0 ? 0 : ((bars[i].close - prev) / prev) * 100;
    result.push({ time: bars[i].time, value });
  }
  return result;
}

// ── Momentum ───────────────────────────────────────────────────────────────

/**
 * Momentum = current close - close n periods ago.
 * @param period Default 10
 */
export function computeMomentum(bars: OHLCVBar[], period = 10): IndicatorPoint[] {
  if (bars.length <= period || period < 1) return [];

  const result: IndicatorPoint[] = [];
  for (let i = period; i < bars.length; i++) {
    result.push({
      time: bars[i].time,
      value: bars[i].close - bars[i - period].close,
    });
  }
  return result;
}
