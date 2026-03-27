/**
 * Volatility indicator computation functions.
 * Pure math — no chart rendering logic.
 */

import type { Time } from "lightweight-charts";
import type { OHLCVBar, IndicatorPoint } from "./trend";

// ── Result types ───────────────────────────────────────────────────────────

export interface BollingerBandsResult {
  time: Time;
  upper: number;
  middle: number;
  lower: number;
}

export interface KeltnerChannelResult {
  time: Time;
  upper: number;
  middle: number;
  lower: number;
}

// ── Bollinger Bands ────────────────────────────────────────────────────────

/**
 * Bollinger Bands: SMA middle band, upper = SMA + k×stddev, lower = SMA - k×stddev.
 * @param period SMA period (default 20)
 * @param k      Standard deviation multiplier (default 2)
 */
export function computeBollingerBands(
  bars: OHLCVBar[],
  period = 20,
  k = 2,
): BollingerBandsResult[] {
  if (bars.length < period || period < 1) return [];

  const result: BollingerBandsResult[] = [];

  for (let i = period - 1; i < bars.length; i++) {
    let sum = 0;
    for (let j = i - period + 1; j <= i; j++) sum += bars[j].close;
    const sma = sum / period;

    let sqSum = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const diff = bars[j].close - sma;
      sqSum += diff * diff;
    }
    const stddev = Math.sqrt(sqSum / period);

    result.push({
      time: bars[i].time,
      upper: sma + k * stddev,
      middle: sma,
      lower: sma - k * stddev,
    });
  }

  return result;
}


// ── ATR (Average True Range) ───────────────────────────────────────────────

/**
 * ATR using Wilder's smoothing of True Range.
 * @param period Default 14
 */
export function computeATR(bars: OHLCVBar[], period = 14): IndicatorPoint[] {
  if (bars.length < period + 1 || period < 1) return [];

  // True Range starting at index 1
  const tr: number[] = [];
  for (let i = 1; i < bars.length; i++) {
    tr.push(
      Math.max(
        bars[i].high - bars[i].low,
        Math.abs(bars[i].high - bars[i - 1].close),
        Math.abs(bars[i].low - bars[i - 1].close),
      ),
    );
  }

  // First ATR = simple average of first `period` TRs
  let atrSum = 0;
  for (let i = 0; i < period; i++) atrSum += tr[i];
  let atr = atrSum / period;

  const result: IndicatorPoint[] = [{ time: bars[period].time, value: atr }];

  // Wilder's smoothing for subsequent values
  for (let i = period; i < tr.length; i++) {
    atr = (atr * (period - 1) + tr[i]) / period;
    result.push({ time: bars[i + 1].time, value: atr });
  }

  return result;
}

// ── Keltner Channel ────────────────────────────────────────────────────────

/**
 * Keltner Channel: EMA middle, upper = EMA + multiplier×ATR, lower = EMA - multiplier×ATR.
 * @param period     EMA and ATR period (default 20)
 * @param multiplier ATR multiplier (default 1.5)
 */
export function computeKeltnerChannel(
  bars: OHLCVBar[],
  period = 20,
  multiplier = 1.5,
): KeltnerChannelResult[] {
  if (bars.length < period + 1 || period < 1) return [];

  // Compute EMA of close prices
  const emaK = 2 / (period + 1);
  let emaSum = 0;
  for (let i = 0; i < period; i++) emaSum += bars[i].close;
  let ema = emaSum / period;
  const emaValues: { time: Time; value: number }[] = [
    { time: bars[period - 1].time, value: ema },
  ];
  for (let i = period; i < bars.length; i++) {
    ema = bars[i].close * emaK + ema * (1 - emaK);
    emaValues.push({ time: bars[i].time, value: ema });
  }

  // Compute ATR using Wilder's smoothing
  const tr: number[] = [];
  for (let i = 1; i < bars.length; i++) {
    tr.push(
      Math.max(
        bars[i].high - bars[i].low,
        Math.abs(bars[i].high - bars[i - 1].close),
        Math.abs(bars[i].low - bars[i - 1].close),
      ),
    );
  }

  let atrSum = 0;
  for (let i = 0; i < period; i++) atrSum += tr[i];
  let atr = atrSum / period;
  // atr[0] corresponds to bars[period] (TR starts at bar 1, first ATR after period TRs)
  const atrValues: { time: Time; value: number }[] = [
    { time: bars[period].time, value: atr },
  ];
  for (let i = period; i < tr.length; i++) {
    atr = (atr * (period - 1) + tr[i]) / period;
    atrValues.push({ time: bars[i + 1].time, value: atr });
  }

  // Align: EMA starts at bars[period-1], ATR starts at bars[period]
  // Both available from bars[period] onward
  const result: KeltnerChannelResult[] = [];
  for (let i = 0; i < atrValues.length; i++) {
    // emaValues index offset: ATR starts 1 bar after EMA
    const emaIdx = i + 1; // emaValues[0] = bars[period-1], emaValues[1] = bars[period]
    if (emaIdx < emaValues.length) {
      const e = emaValues[emaIdx].value;
      const a = atrValues[i].value;
      result.push({
        time: atrValues[i].time,
        upper: e + multiplier * a,
        middle: e,
        lower: e - multiplier * a,
      });
    }
  }

  return result;
}

// ── Standard Deviation ─────────────────────────────────────────────────────

/**
 * Rolling standard deviation of close prices.
 * @param period Default 20
 */
export function computeStdDev(bars: OHLCVBar[], period = 20): IndicatorPoint[] {
  if (bars.length < period || period < 1) return [];

  const result: IndicatorPoint[] = [];

  for (let i = period - 1; i < bars.length; i++) {
    let sum = 0;
    for (let j = i - period + 1; j <= i; j++) sum += bars[j].close;
    const mean = sum / period;

    let sqSum = 0;
    for (let j = i - period + 1; j <= i; j++) {
      const diff = bars[j].close - mean;
      sqSum += diff * diff;
    }

    result.push({ time: bars[i].time, value: Math.sqrt(sqSum / period) });
  }

  return result;
}
