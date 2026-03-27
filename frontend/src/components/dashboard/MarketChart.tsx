"use client";

import { useEffect, useRef, useCallback } from "react";
import { ChartEngine, type OHLCVBar } from "@/lib/chart-engine";
import { useTheme } from "@/context/ThemeContext";
import {
  addIndicatorToChart,
  removeIndicatorFromChart,
  type IndicatorConfig,
  type ActiveIndicatorSeries,
} from "@/lib/indicator-renderer";
import type { ISeriesApi, Time } from "lightweight-charts";
import { LineSeries } from "lightweight-charts";
import type {
  Drawing,
  DrawingType,
  DrawingPoint,
} from "@/lib/drawing-tools";
import {
  createTrendLine,
  createHorizontalLine,
  createFibonacci,
  createRectangle,
} from "@/lib/drawing-tools";

interface ChartProps {
  data: OHLCVBar[];
  activeIndicators?: IndicatorConfig[];
  drawings?: Drawing[];
  activeTool?: DrawingType | null;
  onDrawingCreated?: (drawing: Drawing) => void;
}

export type { OHLCVBar };

export function MarketChart({
  data,
  activeIndicators = [],
  drawings = [],
  activeTool = null,
  onDrawingCreated,
}: ChartProps) {
  const chartContainerRef = useRef<HTMLDivElement>(null);
  const engineRef = useRef<ChartEngine | null>(null);
  const activeSeriesRef = useRef<ActiveIndicatorSeries[]>([]);
  const drawingSeriesRef = useRef<ISeriesApi<any>[]>([]);
  const clickCountRef = useRef(0);
  const firstClickRef = useRef<DrawingPoint | null>(null);
  const { theme } = useTheme();

  // Create chart engine once
  useEffect(() => {
    if (!chartContainerRef.current) return;

    const engine = new ChartEngine();
    engine.create(chartContainerRef.current, theme);
    engineRef.current = engine;

    return () => {
      activeSeriesRef.current = [];
      drawingSeriesRef.current = [];
      engine.destroy();
      engineRef.current = null;
    };
    // Only recreate on mount/unmount — theme changes handled via applyTheme
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Update data when it changes
  useEffect(() => {
    if (engineRef.current && data.length > 0) {
      engineRef.current.setData(data);
    }
  }, [data]);

  // React to theme changes
  useEffect(() => {
    engineRef.current?.applyTheme(theme);
  }, [theme]);

  // Render indicators when data or activeIndicators change
  useEffect(() => {
    const chart = engineRef.current?.getChart();
    if (!chart || data.length === 0) return;

    // Remove all existing indicator series
    for (const active of activeSeriesRef.current) {
      removeIndicatorFromChart(chart, active);
    }
    activeSeriesRef.current = [];

    // Add visible indicators
    for (const config of activeIndicators) {
      if (!config.visible) continue;
      const result = addIndicatorToChart(chart, config, data);
      if (result) {
        activeSeriesRef.current.push(result);
      }
    }
  }, [data, activeIndicators]);

  // Render drawings as line series on the chart
  useEffect(() => {
    const chart = engineRef.current?.getChart();
    if (!chart) return;

    // Remove previous drawing series
    for (const s of drawingSeriesRef.current) {
      try { chart.removeSeries(s); } catch { /* already removed */ }
    }
    drawingSeriesRef.current = [];

    if (data.length === 0) return;

    const firstTime = data[0].time;
    const lastTime = data[data.length - 1].time;

    for (const d of drawings) {
      switch (d.type) {
        case "trendline": {
          const s = chart.addSeries(LineSeries, {
            color: d.color,
            lineWidth: d.lineWidth as any,
            priceScaleId: "right",
            lastValueVisible: false,
            priceLineVisible: false,
          });
          s.setData([
            { time: d.start.time, value: d.start.price },
            { time: d.end.time, value: d.end.price },
          ] as any);
          drawingSeriesRef.current.push(s);
          break;
        }
        case "horizontal": {
          const s = chart.addSeries(LineSeries, {
            color: d.color,
            lineWidth: d.lineWidth as any,
            lineStyle: 2,
            priceScaleId: "right",
            lastValueVisible: true,
            priceLineVisible: false,
          });
          s.setData([
            { time: firstTime, value: d.price },
            { time: lastTime, value: d.price },
          ] as any);
          drawingSeriesRef.current.push(s);
          break;
        }
        case "fibonacci": {
          const highPrice = Math.max(d.start.price, d.end.price);
          const lowPrice = Math.min(d.start.price, d.end.price);
          const range = highPrice - lowPrice;
          for (const level of d.levels) {
            const price = highPrice - range * level;
            const s = chart.addSeries(LineSeries, {
              color: d.color,
              lineWidth: 1 as any,
              lineStyle: level === 0 || level === 1 ? 0 : 2,
              priceScaleId: "right",
              lastValueVisible: false,
              priceLineVisible: false,
            });
            s.setData([
              { time: d.start.time, value: price },
              { time: d.end.time, value: price },
            ] as any);
            drawingSeriesRef.current.push(s);
          }
          break;
        }
        case "rectangle": {
          // Render rectangle as top and bottom horizontal lines between start/end times
          const topPrice = Math.max(d.start.price, d.end.price);
          const bottomPrice = Math.min(d.start.price, d.end.price);
          for (const price of [topPrice, bottomPrice]) {
            const s = chart.addSeries(LineSeries, {
              color: d.color,
              lineWidth: d.lineWidth as any,
              priceScaleId: "right",
              lastValueVisible: false,
              priceLineVisible: false,
            });
            s.setData([
              { time: d.start.time, value: price },
              { time: d.end.time, value: price },
            ] as any);
            drawingSeriesRef.current.push(s);
          }
          break;
        }
      }
    }
  }, [drawings, data]);

  // Handle click-to-place drawing interactions
  const handleChartClick = useCallback(
    (e: React.MouseEvent<HTMLDivElement>) => {
      if (!activeTool || !onDrawingCreated || !engineRef.current) return;
      const chart = engineRef.current.getChart();
      if (!chart || data.length === 0) return;

      const rect = e.currentTarget.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      const timeCoord = chart.timeScale().coordinateToTime(x);
      const mainSeries = (chart as any).getSeries?.()[0];

      // Use a simpler approach: map y to price range from data
      const prices = data.map((b) => [b.high, b.low]).flat();
      const maxPrice = Math.max(...prices);
      const minPrice = Math.min(...prices);
      const chartHeight = rect.height;
      // Approximate price from y coordinate (top = max, bottom = min)
      const priceRange = maxPrice - minPrice;
      const padding = priceRange * 0.1;
      const effectiveMax = maxPrice + padding;
      const effectiveMin = minPrice - padding;
      const price = effectiveMax - (y / chartHeight) * (effectiveMax - effectiveMin);

      if (!timeCoord) return;

      const point: DrawingPoint = { time: timeCoord as Time, price };

      if (activeTool === "horizontal") {
        onDrawingCreated(createHorizontalLine(price));
        return;
      }

      // Two-click tools: trendline, fibonacci, rectangle
      if (clickCountRef.current === 0) {
        firstClickRef.current = point;
        clickCountRef.current = 1;
        return;
      }

      // Second click
      const start = firstClickRef.current!;
      clickCountRef.current = 0;
      firstClickRef.current = null;

      switch (activeTool) {
        case "trendline":
          onDrawingCreated(createTrendLine(start, point));
          break;
        case "fibonacci":
          onDrawingCreated(createFibonacci(start, point));
          break;
        case "rectangle":
          onDrawingCreated(createRectangle(start, point));
          break;
      }
    },
    [activeTool, onDrawingCreated, data],
  );

  // Reset click state when tool changes
  useEffect(() => {
    clickCountRef.current = 0;
    firstClickRef.current = null;
  }, [activeTool]);

  return (
    <div
      ref={chartContainerRef}
      className={`w-full h-full min-h-[480px] ${activeTool ? "cursor-crosshair" : ""}`}
      onClick={handleChartClick}
    />
  );
}
