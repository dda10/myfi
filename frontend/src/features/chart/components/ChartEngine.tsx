"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { MarketChart } from "@/components/dashboard/MarketChart";
import { IndicatorPanel } from "./IndicatorPanel";
import { DrawingToolbar } from "./DrawingToolbar";
import { IntervalSelector } from "./IntervalSelector";
import { fetchOHLCV, type TimeInterval, type OHLCVBar } from "@/lib/chart-engine";
import {
  loadDrawings,
  saveDrawings,
  clearDrawings,
  type Drawing,
  type DrawingType,
} from "@/lib/drawing-tools";
import type { IndicatorConfig } from "@/lib/indicator-renderer";
import { useTheme } from "@/context/ThemeContext";
import { useI18n } from "@/context/I18nContext";
import { isVNTradingHours } from "@/hooks/usePolling";
import { BarChart3, Loader2, AlertCircle } from "lucide-react";

interface ChartEngineProps {
  symbol: string;
}

/** Polling interval: 15s during trading hours, 5min outside. */
function getRefreshInterval(): number {
  return isVNTradingHours() ? 15_000 : 300_000;
}

export function ChartEngine({ symbol }: ChartEngineProps) {
  const { theme } = useTheme();
  const { t } = useI18n();
  const isDark = theme === "dark";

  // ── State ──────────────────────────────────────────────────────────────
  const [interval, setInterval_] = useState<TimeInterval>("1d");
  const [data, setData] = useState<OHLCVBar[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [indicators, setIndicators] = useState<IndicatorConfig[]>([]);
  const [drawings, setDrawings] = useState<Drawing[]>([]);
  const [activeTool, setActiveTool] = useState<DrawingType | null>(null);
  const [showIndicators, setShowIndicators] = useState(false);
  const abortRef = useRef<AbortController | null>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // ── Data fetching ──────────────────────────────────────────────────────
  const fetchData = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    try {
      setLoading((prev) => (data.length === 0 ? true : prev)); // only show spinner on first load
      const bars = await fetchOHLCV(symbol, interval, controller.signal);
      if (!controller.signal.aborted) {
        setData(bars);
        setError(null);
      }
    } catch (err: any) {
      if (err?.name !== "AbortError") {
        setError(t("chart.error"));
      }
    } finally {
      if (!controller.signal.aborted) {
        setLoading(false);
      }
    }
  }, [symbol, interval, t, data.length]);

  // Fetch on mount and when symbol/interval changes
  useEffect(() => {
    fetchData();
    return () => {
      abortRef.current?.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [symbol, interval]);

  // Auto-refresh polling
  useEffect(() => {
    function scheduleNext() {
      timerRef.current = setTimeout(async () => {
        await fetchData();
        scheduleNext();
      }, getRefreshInterval());
    }
    scheduleNext();
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, [fetchData]);

  // ── Drawings persistence ───────────────────────────────────────────────
  useEffect(() => {
    setDrawings(loadDrawings(symbol));
  }, [symbol]);

  useEffect(() => {
    saveDrawings(symbol, drawings);
  }, [symbol, drawings]);

  const handleDrawingCreated = useCallback((drawing: Drawing) => {
    setDrawings((prev) => [...prev, drawing]);
  }, []);

  const handleClearDrawings = useCallback(() => {
    setDrawings([]);
    clearDrawings(symbol);
  }, [symbol]);

  // ── Indicator management ───────────────────────────────────────────────
  const handleAddIndicator = useCallback((config: IndicatorConfig) => {
    setIndicators((prev) => [...prev, config]);
  }, []);

  const handleRemoveIndicator = useCallback((id: string) => {
    setIndicators((prev) => prev.filter((i) => i.id !== id));
  }, []);

  const handleUpdateIndicator = useCallback((id: string, params: Record<string, number>) => {
    setIndicators((prev) =>
      prev.map((i) => (i.id === id ? { ...i, params } : i)),
    );
  }, []);

  const handleToggleVisibility = useCallback((id: string) => {
    setIndicators((prev) =>
      prev.map((i) => (i.id === id ? { ...i, visible: !i.visible } : i)),
    );
  }, []);

  // ── Render ─────────────────────────────────────────────────────────────
  return (
    <div className={`flex flex-col rounded-xl border overflow-hidden ${
      isDark ? "bg-zinc-950 border-zinc-800" : "bg-white border-gray-200"
    }`}>
      {/* Toolbar row */}
      <div className={`flex items-center gap-3 px-3 py-2 border-b flex-wrap ${
        isDark ? "border-zinc-800 bg-zinc-900/50" : "border-gray-200 bg-gray-50"
      }`}>
        <IntervalSelector selected={interval} onChange={setInterval_} />
        <div className={`w-px h-5 ${isDark ? "bg-zinc-700" : "bg-gray-300"}`} />
        <DrawingToolbar
          activeTool={activeTool}
          onSelectTool={setActiveTool}
          onClearAll={handleClearDrawings}
        />
        <div className="ml-auto">
          <button
            onClick={() => setShowIndicators((v) => !v)}
            title={t("chart.indicators")}
            aria-label={t("chart.indicators")}
            aria-pressed={showIndicators}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded text-sm transition-colors ${
              showIndicators
                ? "bg-indigo-600 text-white"
                : isDark
                  ? "text-zinc-400 hover:text-white hover:bg-zinc-700"
                  : "text-gray-500 hover:text-gray-800 hover:bg-gray-200"
            }`}
          >
            <BarChart3 className="w-4 h-4" />
            <span className="hidden sm:inline">{t("chart.indicators")}</span>
            {indicators.length > 0 && (
              <span className={`text-xs rounded-full px-1.5 ${
                showIndicators ? "bg-indigo-500" : isDark ? "bg-zinc-700" : "bg-gray-300"
              }`}>
                {indicators.length}
              </span>
            )}
          </button>
        </div>
      </div>

      {/* Main content area */}
      <div className="flex flex-1 min-h-0">
        {/* Chart area */}
        <div className="flex-1 relative min-h-[480px]">
          {loading && data.length === 0 && (
            <div className="absolute inset-0 flex items-center justify-center z-10">
              <div className="flex items-center gap-2">
                <Loader2 className={`w-5 h-5 animate-spin ${isDark ? "text-zinc-400" : "text-gray-400"}`} />
                <span className={`text-sm ${isDark ? "text-zinc-400" : "text-gray-500"}`}>{t("chart.loading")}</span>
              </div>
            </div>
          )}
          {error && data.length === 0 && (
            <div className="absolute inset-0 flex items-center justify-center z-10">
              <div className="flex items-center gap-2">
                <AlertCircle className="w-5 h-5 text-red-400" />
                <span className={`text-sm ${isDark ? "text-zinc-400" : "text-gray-500"}`}>{error}</span>
              </div>
            </div>
          )}
          <MarketChart
            data={data}
            activeIndicators={indicators}
            drawings={drawings}
            activeTool={activeTool}
            onDrawingCreated={handleDrawingCreated}
          />
        </div>

        {/* Indicator panel sidebar */}
        {showIndicators && (
          <div className={`w-64 border-l overflow-y-auto flex-shrink-0 ${
            isDark ? "border-zinc-800" : "border-gray-200"
          }`}>
            <IndicatorPanel
              activeIndicators={indicators}
              onAdd={handleAddIndicator}
              onRemove={handleRemoveIndicator}
              onUpdate={handleUpdateIndicator}
              onToggleVisibility={handleToggleVisibility}
            />
          </div>
        )}
      </div>
    </div>
  );
}
