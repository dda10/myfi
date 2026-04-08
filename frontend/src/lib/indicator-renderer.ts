/**
 * Indicator configuration registry and chart rendering utilities.
 * Maps indicator computation functions to lightweight-charts series.
 */

import type { IChartApi, ISeriesApi, Time } from "lightweight-charts";
import { LineSeries, HistogramSeries } from "lightweight-charts";
import type { OHLCVBar } from "./indicators/trend";

// ── Types ──────────────────────────────────────────────────────────────────

export type IndicatorPane = "overlay" | "oscillator";

export type IndicatorCategory = "Trend" | "Momentum" | "Volatility" | "Volume";

export interface IndicatorParamDef {
  name: string;
  label: string;
  default: number;
  min?: number;
  max?: number;
  step?: number;
}

export interface IndicatorConfig {
  /** Unique instance id (e.g. "sma-1", "rsi-2") */
  id: string;
  /** Registry key (e.g. "sma", "rsi") */
  indicatorKey: string;
  /** Display name */
  name: string;
  /** Overlay on price pane or oscillator in separate pane */
  pane: IndicatorPane;
  category: IndicatorCategory;
  /** Current parameter values keyed by param name */
  params: Record<string, number>;
  color: string;
  visible: boolean;
}

export interface IndicatorRegistryEntry {
  key: string;
  name: string;
  pane: IndicatorPane;
  category: IndicatorCategory;
  paramDefs: IndicatorParamDef[];
  color: string;
}

// ── Color palette ──────────────────────────────────────────────────────────

const COLORS = {
  blue: "#2962ff",
  orange: "#ff6d00",
  purple: "#aa00ff",
  teal: "#00bfa5",
  pink: "#ff4081",
  cyan: "#00e5ff",
  lime: "#c6ff00",
  amber: "#ffab00",
  indigo: "#304ffe",
  red: "#ff1744",
  green: "#00c853",
  deepPurple: "#6200ea",
  lightBlue: "#0091ea",
  yellow: "#ffd600",
  deepOrange: "#dd2c00",
  brown: "#795548",
  blueGrey: "#546e7a",
  magenta: "#e040fb",
  coral: "#ff6e40",
  mint: "#69f0ae",
  gold: "#ffd740",
};

// ── Registry ───────────────────────────────────────────────────────────────

export const INDICATOR_REGISTRY: IndicatorRegistryEntry[] = [
  // ── Trend (overlay) ──
  {
    key: "sma",
    name: "SMA",
    pane: "overlay",
    category: "Trend",
    paramDefs: [{ name: "period", label: "Period", default: 20, min: 1, max: 500 }],
    color: COLORS.blue,
  },
  {
    key: "ema",
    name: "EMA",
    pane: "overlay",
    category: "Trend",
    paramDefs: [{ name: "period", label: "Period", default: 12, min: 1, max: 500 }],
    color: COLORS.orange,
  },
  {
    key: "vwap",
    name: "VWAP",
    pane: "overlay",
    category: "Trend",
    paramDefs: [],
    color: COLORS.purple,
  },
  {
    key: "vwma",
    name: "VWMA",
    pane: "overlay",
    category: "Trend",
    paramDefs: [{ name: "period", label: "Period", default: 20, min: 1, max: 500 }],
    color: COLORS.teal,
  },
  {
    key: "bollinger",
    name: "Bollinger Bands",
    pane: "overlay",
    category: "Volatility",
    paramDefs: [
      { name: "period", label: "Period", default: 20, min: 1, max: 500 },
      { name: "stdDev", label: "Std Dev", default: 2, min: 0.1, max: 5, step: 0.1 },
    ],
    color: COLORS.cyan,
  },
  {
    key: "keltner",
    name: "Keltner Channel",
    pane: "overlay",
    category: "Volatility",
    paramDefs: [
      { name: "period", label: "Period", default: 20, min: 1, max: 500 },
      { name: "multiplier", label: "Multiplier", default: 1.5, min: 0.1, max: 5, step: 0.1 },
    ],
    color: COLORS.pink,
  },
  {
    key: "parabolicSar",
    name: "Parabolic SAR",
    pane: "overlay",
    category: "Trend",
    paramDefs: [
      { name: "afStep", label: "AF Step", default: 0.02, min: 0.001, max: 0.1, step: 0.001 },
      { name: "afMax", label: "AF Max", default: 0.2, min: 0.01, max: 1, step: 0.01 },
    ],
    color: COLORS.lime,
  },
  {
    key: "supertrend",
    name: "Supertrend",
    pane: "overlay",
    category: "Trend",
    paramDefs: [
      { name: "period", label: "Period", default: 10, min: 1, max: 100 },
      { name: "multiplier", label: "Multiplier", default: 3, min: 0.5, max: 10, step: 0.5 },
    ],
    color: COLORS.amber,
  },
  {
    key: "linearRegression",
    name: "Linear Regression",
    pane: "overlay",
    category: "Volume",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 2, max: 500 }],
    color: COLORS.indigo,
  },
  // ── Oscillators (separate pane) ──
  {
    key: "rsi",
    name: "RSI",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.purple,
  },
  {
    key: "macd",
    name: "MACD",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [
      { name: "fastPeriod", label: "Fast", default: 12, min: 1, max: 100 },
      { name: "slowPeriod", label: "Slow", default: 26, min: 1, max: 200 },
      { name: "signalPeriod", label: "Signal", default: 9, min: 1, max: 50 },
    ],
    color: COLORS.blue,
  },
  {
    key: "williamsR",
    name: "Williams %R",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.red,
  },
  {
    key: "cmo",
    name: "CMO",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.green,
  },
  {
    key: "stochastic",
    name: "Stochastic",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [
      { name: "kPeriod", label: "%K", default: 14, min: 1, max: 100 },
      { name: "dPeriod", label: "%D", default: 3, min: 1, max: 50 },
      { name: "smooth", label: "Smooth", default: 3, min: 1, max: 50 },
    ],
    color: COLORS.deepPurple,
  },
  {
    key: "roc",
    name: "ROC",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [{ name: "period", label: "Period", default: 12, min: 1, max: 200 }],
    color: COLORS.lightBlue,
  },
  {
    key: "momentum",
    name: "Momentum",
    pane: "oscillator",
    category: "Momentum",
    paramDefs: [{ name: "period", label: "Period", default: 10, min: 1, max: 200 }],
    color: COLORS.coral,
  },
  {
    key: "adx",
    name: "ADX",
    pane: "oscillator",
    category: "Trend",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.yellow,
  },
  {
    key: "aroon",
    name: "Aroon",
    pane: "oscillator",
    category: "Trend",
    paramDefs: [{ name: "period", label: "Period", default: 25, min: 1, max: 200 }],
    color: COLORS.magenta,
  },
  {
    key: "atr",
    name: "ATR",
    pane: "oscillator",
    category: "Volatility",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.deepOrange,
  },
  {
    key: "stdDev",
    name: "Std Dev",
    pane: "oscillator",
    category: "Volatility",
    paramDefs: [{ name: "period", label: "Period", default: 20, min: 1, max: 500 }],
    color: COLORS.brown,
  },
  {
    key: "obv",
    name: "OBV",
    pane: "oscillator",
    category: "Volume",
    paramDefs: [],
    color: COLORS.mint,
  },
  {
    key: "mfi",
    name: "MFI",
    pane: "oscillator",
    category: "Volume",
    paramDefs: [{ name: "period", label: "Period", default: 14, min: 1, max: 100 }],
    color: COLORS.gold,
  },
];

// ── Helpers ────────────────────────────────────────────────────────────────

export function getRegistryEntry(key: string): IndicatorRegistryEntry | undefined {
  return INDICATOR_REGISTRY.find((e) => e.key === key);
}

/** Group registry entries by category. */
export function getIndicatorsByCategory(): Record<IndicatorCategory, IndicatorRegistryEntry[]> {
  const grouped: Record<IndicatorCategory, IndicatorRegistryEntry[]> = {
    Trend: [],
    Momentum: [],
    Volatility: [],
    Volume: [],
  };
  for (const entry of INDICATOR_REGISTRY) {
    grouped[entry.category].push(entry);
  }
  return grouped;
}

let instanceCounter = 0;

/** Create a new IndicatorConfig instance from a registry entry. */
export function createIndicatorInstance(entry: IndicatorRegistryEntry): IndicatorConfig {
  instanceCounter++;
  const params: Record<string, number> = {};
  for (const p of entry.paramDefs) {
    params[p.name] = p.default;
  }
  return {
    id: `${entry.key}-${instanceCounter}`,
    indicatorKey: entry.key,
    name: entry.name,
    pane: entry.pane,
    category: entry.category,
    params,
    color: entry.color,
    visible: true,
  };
}

// ── Computation dispatcher ─────────────────────────────────────────────────

import {
  computeSMA,
  computeEMA,
  computeVWAP,
  computeVWMA,
  computeADX,
  computeAroon,
  computeParabolicSAR,
  computeSupertrend,
  computeRSI,
  computeMACD,
  computeWilliamsR,
  computeCMO,
  computeStochastic,
  computeROC,
  computeMomentum,
  computeBollingerBands,
  computeKeltnerChannel,
  computeATR,
  computeStdDev,
  computeOBV,
  computeMFI,
  computeLinearRegression,
} from "./indicators";

import type {
  IndicatorPoint,
  MACDResult,
  StochasticResult,
  BollingerBandsResult,
  KeltnerChannelResult,
  ADXResult,
  AroonResult,
} from "./indicators";

export type IndicatorSeriesData =
  | { type: "line"; data: IndicatorPoint[] }
  | { type: "bands"; upper: IndicatorPoint[]; middle: IndicatorPoint[]; lower: IndicatorPoint[] }
  | { type: "macd"; macd: IndicatorPoint[]; signal: IndicatorPoint[]; histogram: IndicatorPoint[] }
  | { type: "dual"; lineA: IndicatorPoint[]; lineB: IndicatorPoint[]; labelA: string; labelB: string }
  | { type: "scatter"; data: IndicatorPoint[] };

/** Compute indicator data from OHLCV bars and config params. */
export function computeIndicator(
  bars: OHLCVBar[],
  config: IndicatorConfig,
): IndicatorSeriesData | null {
  const p = config.params;
  switch (config.indicatorKey) {
    case "sma":
      return { type: "line", data: computeSMA(bars, p.period) };
    case "ema":
      return { type: "line", data: computeEMA(bars, p.period) };
    case "vwap":
      return { type: "line", data: computeVWAP(bars) };
    case "vwma":
      return { type: "line", data: computeVWMA(bars, p.period) };
    case "linearRegression":
      return { type: "line", data: computeLinearRegression(bars, p.period) };
    case "rsi":
      return { type: "line", data: computeRSI(bars, p.period) };
    case "williamsR":
      return { type: "line", data: computeWilliamsR(bars, p.period) };
    case "cmo":
      return { type: "line", data: computeCMO(bars, p.period) };
    case "roc":
      return { type: "line", data: computeROC(bars, p.period) };
    case "momentum":
      return { type: "line", data: computeMomentum(bars, p.period) };
    case "atr":
      return { type: "line", data: computeATR(bars, p.period) };
    case "stdDev":
      return { type: "line", data: computeStdDev(bars, p.period) };
    case "obv":
      return { type: "line", data: computeOBV(bars) };
    case "mfi":
      return { type: "line", data: computeMFI(bars, p.period) };
    case "bollinger": {
      const res: BollingerBandsResult[] = computeBollingerBands(bars, p.period, p.stdDev);
      return {
        type: "bands",
        upper: res.map((r) => ({ time: r.time, value: r.upper })),
        middle: res.map((r) => ({ time: r.time, value: r.middle })),
        lower: res.map((r) => ({ time: r.time, value: r.lower })),
      };
    }
    case "keltner": {
      const res: KeltnerChannelResult[] = computeKeltnerChannel(bars, p.period, p.multiplier);
      return {
        type: "bands",
        upper: res.map((r) => ({ time: r.time, value: r.upper })),
        middle: res.map((r) => ({ time: r.time, value: r.middle })),
        lower: res.map((r) => ({ time: r.time, value: r.lower })),
      };
    }
    case "parabolicSar": {
      const res = computeParabolicSAR(bars, p.afStep, p.afMax);
      return { type: "scatter", data: res.map((r) => ({ time: r.time, value: r.sar })) };
    }
    case "supertrend": {
      const res = computeSupertrend(bars, p.period, p.multiplier);
      return { type: "line", data: res.map((r) => ({ time: r.time, value: r.supertrend })) };
    }
    case "macd": {
      const res: MACDResult[] = computeMACD(bars, p.fastPeriod, p.slowPeriod, p.signalPeriod);
      return {
        type: "macd",
        macd: res.map((r) => ({ time: r.time, value: r.macd })),
        signal: res.map((r) => ({ time: r.time, value: r.signal })),
        histogram: res.map((r) => ({ time: r.time, value: r.histogram })),
      };
    }
    case "stochastic": {
      const res: StochasticResult[] = computeStochastic(bars, p.kPeriod, p.dPeriod, p.smooth);
      return {
        type: "dual",
        lineA: res.map((r) => ({ time: r.time, value: r.k })),
        lineB: res.map((r) => ({ time: r.time, value: r.d })),
        labelA: "%K",
        labelB: "%D",
      };
    }
    case "adx": {
      const res: ADXResult[] = computeADX(bars, p.period);
      return { type: "line", data: res.map((r) => ({ time: r.time, value: r.adx })) };
    }
    case "aroon": {
      const res: AroonResult[] = computeAroon(bars, p.period);
      return {
        type: "dual",
        lineA: res.map((r) => ({ time: r.time, value: r.aroonUp })),
        lineB: res.map((r) => ({ time: r.time, value: r.aroonDown })),
        labelA: "Up",
        labelB: "Down",
      };
    }
    default:
      return null;
  }
}

// ── Chart series management ────────────────────────────────────────────────

/** Tracks series added to the chart so they can be removed. */
export interface ActiveIndicatorSeries {
  configId: string;
  series: ISeriesApi<any>[];
}

/**
 * Add indicator series to a lightweight-charts IChartApi.
 * Overlay indicators attach to the main price scale.
 * Oscillator indicators use a separate price scale.
 */
export function addIndicatorToChart(
  chart: IChartApi,
  config: IndicatorConfig,
  bars: OHLCVBar[],
): ActiveIndicatorSeries | null {
  const computed = computeIndicator(bars, config);
  if (!computed) return null;

  const series: ISeriesApi<any>[] = [];
  const priceScaleId = config.pane === "overlay" ? "right" : `indicator-${config.id}`;

  const addLine = (data: IndicatorPoint[], color: string, width = 2, dashed = false): ISeriesApi<any> => {
    const s = chart.addSeries(LineSeries, {
      color,
      lineWidth: width as any,
      lineStyle: dashed ? 2 : 0,
      priceScaleId,
      lastValueVisible: false,
      priceLineVisible: false,
    });
    s.setData(data as any);
    if (config.pane === "oscillator") {
      s.priceScale().applyOptions({ scaleMargins: { top: 0.1, bottom: 0.1 } });
    }
    return s;
  };

  switch (computed.type) {
    case "line":
      series.push(addLine(computed.data, config.color));
      break;
    case "scatter":
      // Render scatter as a line with small width (dots effect)
      series.push(addLine(computed.data, config.color, 1));
      break;
    case "bands":
      series.push(addLine(computed.upper, config.color, 1));
      series.push(addLine(computed.middle, config.color, 2));
      series.push(addLine(computed.lower, config.color, 1));
      break;
    case "macd": {
      series.push(addLine(computed.macd, config.color, 2));
      series.push(addLine(computed.signal, "#ff6d00", 1, true));
      // Histogram
      const h = chart.addSeries(HistogramSeries, {
        priceScaleId,
        lastValueVisible: false,
        priceLineVisible: false,
      });
      h.setData(
        computed.histogram.map((pt) => ({
          time: pt.time as Time,
          value: pt.value,
          color: pt.value >= 0 ? "rgba(0, 200, 83, 0.5)" : "rgba(255, 23, 68, 0.5)",
        })),
      );
      if (config.pane === "oscillator") {
        h.priceScale().applyOptions({ scaleMargins: { top: 0.1, bottom: 0.1 } });
      }
      series.push(h);
      break;
    }
    case "dual":
      series.push(addLine(computed.lineA, config.color, 2));
      series.push(addLine(computed.lineB, "#ff6d00", 1, true));
      break;
  }

  return series.length > 0 ? { configId: config.id, series } : null;
}

/** Remove indicator series from the chart. */
export function removeIndicatorFromChart(
  chart: IChartApi,
  active: ActiveIndicatorSeries,
): void {
  for (const s of active.series) {
    try {
      chart.removeSeries(s);
    } catch {
      // Series may already be removed
    }
  }
}
