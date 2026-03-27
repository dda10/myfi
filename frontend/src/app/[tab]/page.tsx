"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import dynamic from "next/dynamic";
import { MarketChart, type OHLCVBar } from "@/components/dashboard/MarketChart";
import { IntervalSelector } from "@/components/dashboard/IntervalSelector";
import { fetchOHLCV, type TimeInterval } from "@/lib/chart-engine";
import { StockDetailHeader } from "@/components/dashboard/StockDetailHeader";
import { Watchlist } from "@/components/dashboard/Watchlist";
import { IndicatorPanel } from "@/components/dashboard/IndicatorPanel";
import { DrawingToolbar } from "@/components/dashboard/DrawingToolbar";
import { useApp } from "@/context/AppContext";
import { OverviewModule } from "@/components/dashboard/OverviewModule";
import { FilterModule } from "@/components/dashboard/FilterModule";
import { SettingsModule } from "@/components/dashboard/SettingsModule";
import { SignalsModule } from "@/components/dashboard/SignalsModule";
import { PortfolioModule } from "@/components/portfolio/PortfolioModule";
import { ChartSkeleton } from "@/components/common/Skeleton";
import type { IndicatorConfig } from "@/lib/indicator-renderer";
import type { Drawing, DrawingType } from "@/lib/drawing-tools";
import { loadDrawings, saveDrawings, clearDrawings } from "@/lib/drawing-tools";

// Lazy-load heavy components (Task 38.2)
const ComparisonModule = dynamic(
  () => import("@/components/comparison/ComparisonModule").then((m) => ({ default: m.ComparisonModule })),
  { loading: () => <ChartSkeleton /> },
);
const SectorDashboard = dynamic(
  () => import("@/components/sectors/SectorDashboard").then((m) => ({ default: m.SectorDashboard })),
  { loading: () => <ChartSkeleton /> },
);
const WatchlistManager = dynamic(
  () => import("@/components/watchlist/WatchlistManager").then((m) => ({ default: m.WatchlistManager })),
  { loading: () => <ChartSkeleton /> },
);

export default function TabPage() {
  const params = useParams();
  const tab = (params?.tab as string) ?? "overview";
  const { activeSymbol } = useApp();

  const [chartData, setChartData] = useState<OHLCVBar[]>([]);
  const [quoteData, setQuoteData] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [marketView, setMarketView] = useState("Overview");
  const [interval, setInterval_] = useState<TimeInterval>("1d");
  const [activeIndicators, setActiveIndicators] = useState<IndicatorConfig[]>([]);
  const [drawings, setDrawings] = useState<Drawing[]>([]);
  const [activeTool, setActiveTool] = useState<DrawingType | null>(null);

  const handleAddIndicator = useCallback((config: IndicatorConfig) => {
    setActiveIndicators((prev) => [...prev, config]);
  }, []);

  const handleRemoveIndicator = useCallback((id: string) => {
    setActiveIndicators((prev) => prev.filter((c) => c.id !== id));
  }, []);

  const handleUpdateIndicator = useCallback((id: string, params: Record<string, number>) => {
    setActiveIndicators((prev) =>
      prev.map((c) => (c.id === id ? { ...c, params } : c)),
    );
  }, []);

  const handleToggleVisibility = useCallback((id: string) => {
    setActiveIndicators((prev) =>
      prev.map((c) => (c.id === id ? { ...c, visible: !c.visible } : c)),
    );
  }, []);

  // Load drawings from localStorage when symbol changes
  useEffect(() => {
    setDrawings(loadDrawings(activeSymbol));
  }, [activeSymbol]);

  const handleDrawingCreated = useCallback(
    (drawing: Drawing) => {
      setDrawings((prev) => {
        const next = [...prev, drawing];
        saveDrawings(activeSymbol, next);
        return next;
      });
    },
    [activeSymbol],
  );

  const handleClearDrawings = useCallback(() => {
    setDrawings([]);
    clearDrawings(activeSymbol);
    setActiveTool(null);
  }, [activeSymbol]);

  useEffect(() => {
    if (tab !== "markets") return;

    setLoading(true);
    const controller = new AbortController();

    const fetchChart = async () => {
      try {
        const [bars, quoteRes] = await Promise.all([
          fetchOHLCV(activeSymbol, interval, controller.signal),
          fetch(`http://localhost:8080/api/market/quote?symbols=${activeSymbol}`, { signal: controller.signal }),
        ]);

        setChartData(bars);

        if (bars.length >= 2) {
          const last = bars[bars.length - 1];
          const prev = bars[bars.length - 2];
          const change = last.close - prev.close;
          const changePercent = (change / prev.close) * 100;
          setQuoteData({ symbol: activeSymbol, close: last.close, change, changePercent });
        }
      } catch (err: any) {
        if (err?.name !== "AbortError") {
          console.error("Failed to fetch chart data", err);
        }
      } finally {
        setLoading(false);
      }
    };

    fetchChart();
    const pollId = setInterval(fetchChart, 15000);
    return () => {
      controller.abort();
      clearInterval(pollId);
    };
  }, [activeSymbol, tab, interval]);

  if (tab === "overview") return <OverviewModule />;
  if (tab === "portfolio") return <PortfolioModule />;
  if (tab === "comparison") return <ComparisonModule />;
  if (tab === "allocation") return <SectorDashboard />;
  if (tab === "filter")   return <FilterModule />;
  if (tab === "signals")  return <SignalsModule />;
  if (tab === "watchlist") return <WatchlistManager />;
  if (tab === "settings") return <SettingsModule />;

  if (tab === "markets") return (
    <>
      {quoteData && (
        <StockDetailHeader
          symbol={quoteData.symbol || activeSymbol}
          name={`${activeSymbol} Corporation`}
          price={quoteData.close}
          change={quoteData.change}
          changePercent={quoteData.changePercent}
          activeView={marketView}
          onViewChange={setMarketView}
        />
      )}

      {marketView === "Overview" ? (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          <div className="lg:col-span-3 bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg flex flex-col">
            <div className="px-6 py-4 border-b border-zinc-800 flex justify-between items-center">
              <h2 className="text-lg font-semibold text-white">Full Features Chart</h2>
              <div className="flex items-center gap-3">
                <DrawingToolbar
                  activeTool={activeTool}
                  onSelectTool={setActiveTool}
                  onClearAll={handleClearDrawings}
                />
                <IntervalSelector selected={interval} onChange={setInterval_} />
              </div>
            </div>
            <div className="flex-1 w-full bg-zinc-900 flex items-center justify-center">
              {loading ? (
                <span className="text-zinc-500 text-sm">Loading market chart...</span>
              ) : chartData.length > 0 ? (
                <MarketChart
                  data={chartData}
                  activeIndicators={activeIndicators}
                  drawings={drawings}
                  activeTool={activeTool}
                  onDrawingCreated={handleDrawingCreated}
                />
              ) : (
                <span className="text-zinc-500 text-sm">Failed to load or no data available.</span>
              )}
            </div>
          </div>
          <div className="lg:col-span-1 space-y-4">
            <IndicatorPanel
              activeIndicators={activeIndicators}
              onAdd={handleAddIndicator}
              onRemove={handleRemoveIndicator}
              onUpdate={handleUpdateIndicator}
              onToggleVisibility={handleToggleVisibility}
            />
            <Watchlist />
          </div>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
          <h2 className="text-2xl font-bold text-zinc-400 mb-2">{activeSymbol} {marketView}</h2>
          <p className="text-zinc-600">The detailed {marketView} module is currently under construction.</p>
        </div>
      )}
    </>
  );

  // Allocation, Portfolio, etc.
  return (
    <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
      <h2 className="text-2xl font-bold text-zinc-400 mb-2 capitalize">{tab} Module</h2>
      <p className="text-zinc-600">This section is currently under construction.</p>
    </div>
  );
}
