/**
 * Trend indicator computation functions.
 * Pure math — no chart rendering logic.
 */

import type { Time } from "lightweight-charts";

// ── Input type ─────────────────────────────────────────────────────────────

export interface OHLCVBar {
  time: Time;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

// ── Result types ───────────────────────────────────────────────────────────

export interface IndicatorPoint {
  time: Time;
  value: number;
}

export interface ADXResult {
  time: Time;
  adx: number;
  diPlus: number;
  diMinus: number;
}

export interface AroonResult {
  time: Time;
  aroonUp: number;
  aroonDown: number;
}

export interface ParabolicSARResult {
  time: Time;
  sar: number;
  isUpTrend: boolean;
}

export interface SupertrendResult {
  time: Time;
  supertrend: number;
  isUpTrend: boolean;
}

// ── SMA ────────────────────────────────────────────────────────────────────

/** Simple Moving Average of close prices. */
export function computeSMA(bars: OHLCVBar[], period: number): IndicatorPoint[] {
  if (bars.length < period || period < 1) return [];
  const result: IndicatorPoint[] = [];
  let sum = 0;
  for (let i = 0; i < period; i++) sum += bars[i].close;
  result.push({ time: bars[period - 1].time, value: sum / period });
  for (let i = period; i < bars.length; i++) {
    sum += bars[i].close - bars[i - period].close;
    result.push({ time: bars[i].time, value: sum / period });
  }
  return result;
}

// ── EMA ────────────────────────────────────────────────────────────────────

/** Exponential Moving Average of close prices. */
export function computeEMA(bars: OHLCVBar[], period: number): IndicatorPoint[] {
  if (bars.length < period || period < 1) return [];
  const k = 2 / (period + 1);
  // Seed with SMA of first `period` bars
  let sum = 0;
  for (let i = 0; i < period; i++) sum += bars[i].close;
  let ema = sum / period;
  const result: IndicatorPoint[] = [{ time: bars[period - 1].time, value: ema }];
  for (let i = period; i < bars.length; i++) {
    ema = bars[i].close * k + ema * (1 - k);
    result.push({ time: bars[i].time, value: ema });
  }
  return result;
}

// ── VWAP ───────────────────────────────────────────────────────────────────

/** Volume Weighted Average Price (cumulative). */
export function computeVWAP(bars: OHLCVBar[]): IndicatorPoint[] {
  if (bars.length === 0) return [];
  const result: IndicatorPoint[] = [];
  let cumPV = 0;
  let cumVol = 0;
  for (const bar of bars) {
    const tp = (bar.high + bar.low + bar.close) / 3;
    cumPV += tp * bar.volume;
    cumVol += bar.volume;
    result.push({ time: bar.time, value: cumVol === 0 ? tp : cumPV / cumVol });
  }
  return result;
}

// ── VWMA ───────────────────────────────────────────────────────────────────

/** Volume Weighted Moving Average over `period` bars. */
export function computeVWMA(bars: OHLCVBar[], period: number): IndicatorPoint[] {
  if (bars.length < period || period < 1) return [];
  const result: IndicatorPoint[] = [];
  for (let i = period - 1; i < bars.length; i++) {
    let sumPV = 0;
    let sumV = 0;
    for (let j = i - period + 1; j <= i; j++) {
      sumPV += bars[j].close * bars[j].volume;
      sumV += bars[j].volume;
    }
    result.push({ time: bars[i].time, value: sumV === 0 ? bars[i].close : sumPV / sumV });
  }
  return result;
}

// ── ADX ────────────────────────────────────────────────────────────────────

/** Average Directional Movement Index (DI+, DI−, ADX). */
export function computeADX(bars: OHLCVBar[], period: number): ADXResult[] {
  if (bars.length < period * 2 || period < 1) return [];

  // Step 1: True Range, +DM, -DM for each bar (starting at index 1)
  const tr: number[] = [];
  const plusDM: number[] = [];
  const minusDM: number[] = [];
  for (let i = 1; i < bars.length; i++) {
    const high = bars[i].high;
    const low = bars[i].low;
    const prevClose = bars[i - 1].close;
    tr.push(Math.max(high - low, Math.abs(high - prevClose), Math.abs(low - prevClose)));
    const upMove = high - bars[i - 1].high;
    const downMove = bars[i - 1].low - low;
    plusDM.push(upMove > downMove && upMove > 0 ? upMove : 0);
    minusDM.push(downMove > upMove && downMove > 0 ? downMove : 0);
  }

  // Step 2: Smoothed TR, +DM, -DM using Wilder's smoothing
  const smooth = (arr: number[]): number[] => {
    const out: number[] = [];
    let sum = 0;
    for (let i = 0; i < period; i++) sum += arr[i];
    out.push(sum);
    for (let i = period; i < arr.length; i++) {
      out.push(out[out.length - 1] - out[out.length - 1] / period + arr[i]);
    }
    return out;
  };

  const sTR = smooth(tr);
  const sPlusDM = smooth(plusDM);
  const sMinusDM = smooth(minusDM);

  // Step 3: DI+, DI-, DX
  const dx: number[] = [];
  const diPlusArr: number[] = [];
  const diMinusArr: number[] = [];
  for (let i = 0; i < sTR.length; i++) {
    const diP = sTR[i] === 0 ? 0 : (sPlusDM[i] / sTR[i]) * 100;
    const diM = sTR[i] === 0 ? 0 : (sMinusDM[i] / sTR[i]) * 100;
    diPlusArr.push(diP);
    diMinusArr.push(diM);
    const dxVal = diP + diM === 0 ? 0 : (Math.abs(diP - diM) / (diP + diM)) * 100;
    dx.push(dxVal);
  }

  // Step 4: ADX = smoothed DX over `period`
  if (dx.length < period) return [];
  const result: ADXResult[] = [];
  let adxSum = 0;
  for (let i = 0; i < period; i++) adxSum += dx[i];
  let adx = adxSum / period;

  // The first ADX value corresponds to bar index: 1 (for TR offset) + period-1 (smooth offset) + period-1 (ADX smooth)
  const startIdx = 1 + (period - 1) + (period - 1);
  result.push({
    time: bars[startIdx].time,
    adx,
    diPlus: diPlusArr[period - 1],
    diMinus: diMinusArr[period - 1],
  });

  for (let i = period; i < dx.length; i++) {
    adx = (adx * (period - 1) + dx[i]) / period;
    const barIdx = 1 + (period - 1) + i;
    if (barIdx < bars.length) {
      result.push({
        time: bars[barIdx].time,
        adx,
        diPlus: diPlusArr[i],
        diMinus: diMinusArr[i],
      });
    }
  }
  return result;
}

// ── Aroon ──────────────────────────────────────────────────────────────────

/** Aroon Up and Aroon Down indicator. */
export function computeAroon(bars: OHLCVBar[], period: number): AroonResult[] {
  if (bars.length < period + 1 || period < 1) return [];
  const result: AroonResult[] = [];
  for (let i = period; i < bars.length; i++) {
    let highIdx = 0;
    let lowIdx = 0;
    let highVal = -Infinity;
    let lowVal = Infinity;
    for (let j = 0; j <= period; j++) {
      const bar = bars[i - period + j];
      if (bar.high >= highVal) { highVal = bar.high; highIdx = j; }
      if (bar.low <= lowVal) { lowVal = bar.low; lowIdx = j; }
    }
    result.push({
      time: bars[i].time,
      aroonUp: (highIdx / period) * 100,
      aroonDown: (lowIdx / period) * 100,
    });
  }
  return result;
}

// ── Parabolic SAR ──────────────────────────────────────────────────────────

/**
 * Parabolic SAR (Stop and Reverse).
 * @param afStep  Acceleration factor step (default 0.02)
 * @param afMax   Maximum acceleration factor (default 0.2)
 */
export function computeParabolicSAR(
  bars: OHLCVBar[],
  afStep = 0.02,
  afMax = 0.2,
): ParabolicSARResult[] {
  if (bars.length < 2) return [];

  const result: ParabolicSARResult[] = [];
  let isUpTrend = bars[1].close >= bars[0].close;
  let sar = isUpTrend ? bars[0].low : bars[0].high;
  let ep = isUpTrend ? bars[1].high : bars[1].low;
  let af = afStep;

  result.push({ time: bars[1].time, sar, isUpTrend });

  for (let i = 2; i < bars.length; i++) {
    const prevSar = sar;
    sar = prevSar + af * (ep - prevSar);

    if (isUpTrend) {
      // Clamp SAR to not exceed the two prior lows
      sar = Math.min(sar, bars[i - 1].low, bars[i - 2].low);
      if (bars[i].low < sar) {
        // Reverse to downtrend
        isUpTrend = false;
        sar = ep;
        ep = bars[i].low;
        af = afStep;
      } else {
        if (bars[i].high > ep) {
          ep = bars[i].high;
          af = Math.min(af + afStep, afMax);
        }
      }
    } else {
      // Clamp SAR to not be below the two prior highs
      sar = Math.max(sar, bars[i - 1].high, bars[i - 2].high);
      if (bars[i].high > sar) {
        // Reverse to uptrend
        isUpTrend = true;
        sar = ep;
        ep = bars[i].high;
        af = afStep;
      } else {
        if (bars[i].low < ep) {
          ep = bars[i].low;
          af = Math.min(af + afStep, afMax);
        }
      }
    }

    result.push({ time: bars[i].time, sar, isUpTrend });
  }
  return result;
}

// ── Supertrend ─────────────────────────────────────────────────────────────

/**
 * Supertrend indicator.
 * @param period  ATR period (default 10)
 * @param multiplier  ATR multiplier (default 3)
 */
export function computeSupertrend(
  bars: OHLCVBar[],
  period = 10,
  multiplier = 3,
): SupertrendResult[] {
  if (bars.length < period + 1 || period < 1) return [];

  // Step 1: Compute ATR using Wilder's smoothing
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

  const atr: number[] = [];
  let atrSum = 0;
  for (let i = 0; i < period; i++) atrSum += tr[i];
  atr.push(atrSum / period);
  for (let i = period; i < tr.length; i++) {
    atr.push((atr[atr.length - 1] * (period - 1) + tr[i]) / period);
  }

  // Step 2: Compute Supertrend
  // atr[j] corresponds to bars[j + period] (first ATR at index period)
  const result: SupertrendResult[] = [];
  let prevUpperBand = 0;
  let prevLowerBand = 0;
  let prevIsUp = true;

  for (let j = 0; j < atr.length; j++) {
    const barIdx = j + period; // offset: tr starts at bar 1, atr starts after period TRs
    const hl2 = (bars[barIdx].high + bars[barIdx].low) / 2;
    let upperBand = hl2 + multiplier * atr[j];
    let lowerBand = hl2 - multiplier * atr[j];

    // Adjust bands based on previous values
    if (j > 0) {
      lowerBand = lowerBand > prevLowerBand || bars[barIdx - 1].close < prevLowerBand
        ? lowerBand
        : prevLowerBand;
      upperBand = upperBand < prevUpperBand || bars[barIdx - 1].close > prevUpperBand
        ? upperBand
        : prevUpperBand;
    }

    let isUpTrend: boolean;
    let supertrend: number;

    if (j === 0) {
      isUpTrend = bars[barIdx].close > upperBand ? true : bars[barIdx].close < lowerBand ? false : true;
      supertrend = isUpTrend ? lowerBand : upperBand;
    } else {
      if (prevIsUp) {
        isUpTrend = bars[barIdx].close >= lowerBand;
      } else {
        isUpTrend = bars[barIdx].close > upperBand;
      }
      supertrend = isUpTrend ? lowerBand : upperBand;
    }

    result.push({ time: bars[barIdx].time, supertrend, isUpTrend });
    prevUpperBand = upperBand;
    prevLowerBand = lowerBand;
    prevIsUp = isUpTrend;
  }

  return result;
}
