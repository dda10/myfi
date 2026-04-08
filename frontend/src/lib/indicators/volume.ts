/**
 * Volume and statistics indicator computation functions.
 * Pure math — no chart rendering logic.
 */

import type { OHLCVBar, IndicatorPoint } from "./trend";

// ── OBV (On-Balance Volume) ────────────────────────────────────────────────

/**
 * On-Balance Volume: cumulative volume indicator.
 * Adds volume on up days, subtracts on down days, unchanged on flat days.
 */
export function computeOBV(bars: OHLCVBar[]): IndicatorPoint[] {
  if (bars.length === 0) return [];
  const result: IndicatorPoint[] = [];
  let obv = 0;
  result.push({ time: bars[0].time, value: obv });
  for (let i = 1; i < bars.length; i++) {
    if (bars[i].close > bars[i - 1].close) {
      obv += bars[i].volume;
    } else if (bars[i].close < bars[i - 1].close) {
      obv -= bars[i].volume;
    }
    // flat day: obv unchanged
    result.push({ time: bars[i].time, value: obv });
  }
  return result;
}

// ── MFI (Money Flow Index) ──────────────────────────────────────────────────

/**
 * Money Flow Index: volume-weighted RSI.
 * MFI = 100 - 100 / (1 + positive money flow / negative money flow)
 * @param period Default 14
 */
export function computeMFI(bars: OHLCVBar[], period = 14): IndicatorPoint[] {
  if (bars.length < period + 1 || period < 1) return [];

  const result: IndicatorPoint[] = [];

  for (let i = period; i < bars.length; i++) {
    let posFlow = 0;
    let negFlow = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const tp = (bars[j].high + bars[j].low + bars[j].close) / 3;
      const prevTp = (bars[j - 1].high + bars[j - 1].low + bars[j - 1].close) / 3;
      const rawFlow = tp * bars[j].volume;
      if (tp > prevTp) posFlow += rawFlow;
      else if (tp < prevTp) negFlow += rawFlow;
    }
    const mfi = negFlow === 0 ? 100 : 100 - 100 / (1 + posFlow / negFlow);
    result.push({ time: bars[i].time, value: mfi });
  }

  return result;
}

// ── Linear Regression ──────────────────────────────────────────────────────

/**
 * Rolling linear regression of close prices over `period`.
 * Returns the regression line value (fitted y) at each point.
 * @param period Default 14
 */
export function computeLinearRegression(
  bars: OHLCVBar[],
  period = 14,
): IndicatorPoint[] {
  if (bars.length < period || period < 2) return [];
  const result: IndicatorPoint[] = [];
  for (let i = period - 1; i < bars.length; i++) {
    // x values: 0, 1, ..., period-1
    let sumX = 0;
    let sumY = 0;
    let sumXY = 0;
    let sumX2 = 0;
    for (let j = 0; j < period; j++) {
      const x = j;
      const y = bars[i - period + 1 + j].close;
      sumX += x;
      sumY += y;
      sumXY += x * y;
      sumX2 += x * x;
    }
    const n = period;
    const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
    const intercept = (sumY - slope * sumX) / n;
    // Regression value at the last point in the window (x = period - 1)
    const value = intercept + slope * (period - 1);
    result.push({ time: bars[i].time, value });
  }
  return result;
}
