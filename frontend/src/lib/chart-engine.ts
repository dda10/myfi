import {
  createChart,
  ColorType,
  IChartApi,
  CandlestickSeries,
  HistogramSeries,
  ISeriesApi,
  CandlestickData,
  HistogramData,
  Time,
  DeepPartial,
  ChartOptions,
} from "lightweight-charts";

// ── Types ──────────────────────────────────────────────────────────────────

export type TimeInterval = "1m" | "5m" | "15m" | "1h" | "1d" | "1w" | "1M";

export interface OHLCVBar {
  time: Time;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface ChartThemeColors {
  backgroundColor: string;
  textColor: string;
  gridColor: string;
  borderColor: string;
  upColor: string;
  downColor: string;
  crosshairColor: string;
}

const DARK_THEME: ChartThemeColors = {
  backgroundColor: "transparent",
  textColor: "#9ca3af",
  gridColor: "#27272a",
  borderColor: "#27272a",
  upColor: "#22c55e",
  downColor: "#ef4444",
  crosshairColor: "#4f46e5",
};

const LIGHT_THEME: ChartThemeColors = {
  backgroundColor: "transparent",
  textColor: "#374151",
  gridColor: "#e5e7eb",
  borderColor: "#d1d5db",
  upColor: "#16a34a",
  downColor: "#dc2626",
  crosshairColor: "#4f46e5",
};

export function getThemeColors(theme: "light" | "dark"): ChartThemeColors {
  return theme === "dark" ? DARK_THEME : LIGHT_THEME;
}

// ── Interval helpers ───────────────────────────────────────────────────────

/** Map frontend interval codes to backend-compatible interval strings */
const INTERVAL_MAP: Record<TimeInterval, string> = {
  "1m": "1m",
  "5m": "5m",
  "15m": "15m",
  "1h": "1h",
  "1d": "1d",
  "1w": "1w",
  "1M": "1M",
};

export const TIME_INTERVALS: { label: string; value: TimeInterval }[] = [
  { label: "1m", value: "1m" },
  { label: "5m", value: "5m" },
  { label: "15m", value: "15m" },
  { label: "1H", value: "1h" },
  { label: "1D", value: "1d" },
  { label: "1W", value: "1w" },
  { label: "1M", value: "1M" },
];

// ── Data fetching ──────────────────────────────────────────────────────────

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export async function fetchOHLCV(
  symbol: string,
  interval: TimeInterval,
  signal?: AbortSignal,
): Promise<OHLCVBar[]> {
  const backendInterval = INTERVAL_MAP[interval];
  const url = `${API_BASE}/api/market/chart?symbol=${encodeURIComponent(symbol)}&interval=${backendInterval}`;

  const res = await fetch(url, { signal });
  if (!res.ok) throw new Error(`Failed to fetch OHLCV: ${res.status}`);

  const json = await res.json();
  const items: any[] = json.data ?? [];

  return items
    .map((d) => ({
      time: (d.timestamp?.split("T")[0] ?? d.time) as Time,
      open: Number(d.open),
      high: Number(d.high),
      low: Number(d.low),
      close: Number(d.close),
      volume: Number(d.volume ?? d.value ?? 0),
    }))
    .sort((a, b) => (a.time < b.time ? -1 : a.time > b.time ? 1 : 0));
}

// ── ChartEngine class ──────────────────────────────────────────────────────

export class ChartEngine {
  private chart: IChartApi | null = null;
  private candleSeries: ISeriesApi<"Candlestick"> | null = null;
  private volumeSeries: ISeriesApi<"Histogram"> | null = null;
  private container: HTMLElement | null = null;
  private resizeObserver: ResizeObserver | null = null;
  private colors: ChartThemeColors = DARK_THEME;

  /** Create the chart inside the given container */
  create(container: HTMLElement, theme: "light" | "dark" = "dark"): void {
    this.container = container;
    this.colors = getThemeColors(theme);

    const chartOptions: DeepPartial<ChartOptions> = {
      layout: {
        background: { type: ColorType.Solid, color: this.colors.backgroundColor },
        textColor: this.colors.textColor,
      },
      width: container.clientWidth,
      height: container.clientHeight || 480,
      grid: {
        vertLines: { color: this.colors.gridColor },
        horzLines: { color: this.colors.gridColor },
      },
      timeScale: {
        borderColor: this.colors.borderColor,
        timeVisible: true,
      },
      crosshair: {
        mode: 1,
        vertLine: {
          color: this.colors.crosshairColor,
          width: 1,
          style: 1,
          labelBackgroundColor: this.colors.crosshairColor,
        },
        horzLine: {
          color: this.colors.crosshairColor,
          width: 1,
          style: 1,
          labelBackgroundColor: this.colors.crosshairColor,
        },
      },
    };

    this.chart = createChart(container, chartOptions);

    // Candlestick series
    this.candleSeries = this.chart.addSeries(CandlestickSeries, {
      upColor: this.colors.upColor,
      downColor: this.colors.downColor,
      borderVisible: false,
      wickUpColor: this.colors.upColor,
      wickDownColor: this.colors.downColor,
    });

    // Volume histogram as overlay
    this.volumeSeries = this.chart.addSeries(HistogramSeries, {
      color: "#26a69a",
      priceFormat: { type: "volume" },
      priceScaleId: "",
    });
    this.volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.8, bottom: 0 },
    });

    // Auto-resize
    this.resizeObserver = new ResizeObserver(() => {
      if (this.chart && this.container) {
        this.chart.applyOptions({
          width: this.container.clientWidth,
          height: this.container.clientHeight || 480,
        });
      }
    });
    this.resizeObserver.observe(container);
  }

  /** Set OHLCV data on both candlestick and volume series */
  setData(bars: OHLCVBar[]): void {
    if (!this.candleSeries || !this.volumeSeries) return;

    const candles: CandlestickData[] = bars.map((b) => ({
      time: b.time,
      open: b.open,
      high: b.high,
      low: b.low,
      close: b.close,
    }));
    this.candleSeries.setData(candles);

    const volumes: HistogramData[] = bars.map((b) => ({
      time: b.time,
      value: b.volume,
      color:
        b.close >= b.open
          ? "rgba(34, 197, 94, 0.4)"
          : "rgba(239, 68, 68, 0.4)",
    }));
    this.volumeSeries.setData(volumes);

    this.chart?.timeScale().fitContent();
  }

  /** Update theme colors without recreating the chart */
  applyTheme(theme: "light" | "dark"): void {
    if (!this.chart || !this.candleSeries) return;
    this.colors = getThemeColors(theme);

    this.chart.applyOptions({
      layout: {
        background: { type: ColorType.Solid, color: this.colors.backgroundColor },
        textColor: this.colors.textColor,
      },
      grid: {
        vertLines: { color: this.colors.gridColor },
        horzLines: { color: this.colors.gridColor },
      },
      timeScale: { borderColor: this.colors.borderColor },
    });

    this.candleSeries.applyOptions({
      upColor: this.colors.upColor,
      downColor: this.colors.downColor,
      wickUpColor: this.colors.upColor,
      wickDownColor: this.colors.downColor,
    });
  }

  /** Get the underlying chart API for advanced usage (e.g. adding indicator series) */
  getChart(): IChartApi | null {
    return this.chart;
  }

  /** Destroy the chart and clean up */
  destroy(): void {
    this.resizeObserver?.disconnect();
    this.resizeObserver = null;
    this.chart?.remove();
    this.chart = null;
    this.candleSeries = null;
    this.volumeSeries = null;
    this.container = null;
  }
}
