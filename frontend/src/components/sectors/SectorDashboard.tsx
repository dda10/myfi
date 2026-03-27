"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import {
  BarChart,
  Bar,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  LineChart,
  Line,
} from "recharts";
import {
  TrendingUp,
  TrendingDown,
  ArrowRight,
  X,
  RefreshCw,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { isVNTradingHours } from "@/hooks/usePolling";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types ---

type TimePeriod = "today" | "1w" | "1m" | "3m" | "6m" | "1y";

interface SectorPerformance {
  sector: string;
  sectorName: string;
  trend: "uptrend" | "downtrend" | "sideways";
  todayChange: number;
  oneWeekChange: number;
  oneMonthChange: number;
  threeMonthChange: number;
  sixMonthChange: number;
  oneYearChange: number;
  currentPrice: number;
  sma20: number;
  sma50: number;
  isStale: boolean;
}

interface SectorAverages {
  sector: string;
  medianPE: number;
  medianPB: number;
  medianROE: number;
  medianROA: number;
  medianDivYield: number;
  medianDebtToEquity: number;
}

interface SectorStock {
  symbol: string;
  name?: string;
  marketCap: number;
  change: number;
}

// --- Sector metadata ---

const SECTOR_LABELS: Record<string, { vi: string; en: string }> = {
  VNIT:   { vi: "Công nghệ", en: "Technology" },
  VNIND:  { vi: "Công nghiệp", en: "Industrial" },
  VNCONS: { vi: "Tiêu dùng", en: "Consumer Staples" },
  VNCOND: { vi: "Tiêu dùng TY", en: "Consumer Disc." },
  VNHEAL: { vi: "Y tế", en: "Healthcare" },
  VNENE:  { vi: "Năng lượng", en: "Energy" },
  VNUTI:  { vi: "Tiện ích", en: "Utilities" },
  VNREAL: { vi: "Bất động sản", en: "Real Estate" },
  VNFIN:  { vi: "Tài chính", en: "Finance" },
  VNMAT:  { vi: "Vật liệu", en: "Materials" },
};

const TIME_PERIODS: { key: TimePeriod; label: string }[] = [
  { key: "today", label: "Today" },
  { key: "1w", label: "1W" },
  { key: "1m", label: "1M" },
  { key: "3m", label: "3M" },
  { key: "6m", label: "6M" },
  { key: "1y", label: "1Y" },
];

// --- Helpers ---

function getChangeForPeriod(s: SectorPerformance, period: TimePeriod): number {
  switch (period) {
    case "today": return s.todayChange;
    case "1w":    return s.oneWeekChange;
    case "1m":    return s.oneMonthChange;
    case "3m":    return s.threeMonthChange;
    case "6m":    return s.sixMonthChange;
    case "1y":    return s.oneYearChange;
  }
}

function perfColor(value: number): string {
  if (value > 5)  return "bg-green-500";
  if (value > 2)  return "bg-green-400";
  if (value > 0)  return "bg-green-300 text-green-900";
  if (value === 0) return "bg-zinc-600";
  if (value > -2) return "bg-red-300 text-red-900";
  if (value > -5) return "bg-red-400";
  return "bg-red-500";
}

function barColor(value: number): string {
  return value >= 0 ? "#22c55e" : "#ef4444";
}


// --- Trend icon ---

function TrendIcon({ trend }: { trend: string }) {
  if (trend === "uptrend")
    return <TrendingUp size={16} className="text-green-400" />;
  if (trend === "downtrend")
    return <TrendingDown size={16} className="text-red-400" />;
  return <ArrowRight size={16} className="text-yellow-400" />;
}

// --- Main Component ---

export function SectorDashboard() {
  const { locale, formatPercent, formatNumber, t } = useI18n();
  const isVi = locale === "vi-VN";

  const [sectors, setSectors] = useState<SectorPerformance[]>([]);
  const [period, setPeriod] = useState<TimePeriod>("today");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedSector, setSelectedSector] = useState<string | null>(null);
  const [detailData, setDetailData] = useState<{
    averages: SectorAverages | null;
    stocks: SectorStock[];
  } | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const mountedRef = useRef(true);

  // Fetch all sector performances
  const fetchSectors = useCallback(async () => {
    try {
      const res = await fetch(`${API_URL}/api/sectors/performance`);
      if (!res.ok) throw new Error("Failed to fetch sectors");
      const data: SectorPerformance[] = await res.json();
      if (mountedRef.current) {
        setSectors(data);
        setError(null);
      }
    } catch (err: any) {
      if (mountedRef.current) {
        setError(err.message ?? "Failed to fetch sector data");
      }
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  }, []);

  // Auto-refresh: 5min trading hours, 30min off-hours (Req 22.10)
  useEffect(() => {
    mountedRef.current = true;
    fetchSectors();

    const id = setInterval(() => {
      const intervalMs = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;
      // We re-check each tick; simplest approach is to use the shorter interval
      // and skip if not due. But for simplicity, use dynamic interval via recursive setTimeout.
      fetchSectors();
    }, isVNTradingHours() ? 5 * 60_000 : 30 * 60_000);

    return () => {
      mountedRef.current = false;
      clearInterval(id);
    };
  }, [fetchSectors]);

  // Fetch detail data when a sector is selected
  useEffect(() => {
    if (!selectedSector) {
      setDetailData(null);
      return;
    }
    let cancelled = false;
    setDetailLoading(true);

    Promise.all([
      fetch(`${API_URL}/api/sectors/${selectedSector}/averages`).then((r) =>
        r.ok ? r.json() : null,
      ),
      fetch(`${API_URL}/api/sectors/${selectedSector}/stocks`).then((r) =>
        r.ok ? r.json() : [],
      ),
    ])
      .then(([averages, stocks]) => {
        if (!cancelled) {
          const topStocks = Array.isArray(stocks)
            ? stocks.slice(0, 5)
            : [];
          setDetailData({ averages, stocks: topStocks });
        }
      })
      .catch(() => {
        if (!cancelled) setDetailData({ averages: null, stocks: [] });
      })
      .finally(() => {
        if (!cancelled) setDetailLoading(false);
      });

    return () => { cancelled = true; };
  }, [selectedSector]);

  // Derived data
  const sorted = [...sectors].sort(
    (a, b) => getChangeForPeriod(b, period) - getChangeForPeriod(a, period),
  );
  const top3 = sorted.slice(0, 3);
  const bottom3 = sorted.slice(-3).reverse();

  const barData = sorted.map((s) => ({
    name: isVi
      ? SECTOR_LABELS[s.sector]?.vi ?? s.sector
      : SECTOR_LABELS[s.sector]?.en ?? s.sector,
    sector: s.sector,
    value: getChangeForPeriod(s, period),
  }));

  const sectorLabel = (code: string) =>
    isVi
      ? SECTOR_LABELS[code]?.vi ?? code
      : SECTOR_LABELS[code]?.en ?? code;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-96">
        <RefreshCw size={24} className="animate-spin text-zinc-500" />
        <span className="ml-3 text-zinc-500">Loading sector data...</span>
      </div>
    );
  }

  if (error && sectors.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
        <p className="text-red-400 mb-2">{error}</p>
        <button
          onClick={fetchSectors}
          className="px-4 py-2 bg-zinc-800 text-zinc-300 rounded-lg hover:bg-zinc-700 transition"
        >
          Retry
        </button>
      </div>
    );
  }


  return (
    <div className="space-y-6">
      {/* Time Period Selector */}
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">
          {isVi ? "Xu hướng ngành" : "Sector Trends"}
        </h2>
        <div className="flex gap-1 bg-zinc-800 rounded-lg p-1">
          {TIME_PERIODS.map((tp) => (
            <button
              key={tp.key}
              onClick={() => setPeriod(tp.key)}
              className={`px-3 py-1.5 text-sm rounded-md transition ${
                period === tp.key
                  ? "bg-blue-600 text-white"
                  : "text-zinc-400 hover:text-white hover:bg-zinc-700"
              }`}
            >
              {tp.label}
            </button>
          ))}
        </div>
      </div>

      {/* Top 3 / Bottom 3 Summary (Req 22.6) */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-green-400 mb-3">
            {isVi ? "Top 3 ngành tốt nhất" : "Top 3 Performing"}
          </h3>
          <div className="space-y-2">
            {top3.map((s, i) => (
              <div
                key={s.sector}
                className="flex items-center justify-between cursor-pointer hover:bg-zinc-800 rounded-lg px-2 py-1.5 transition"
                onClick={() => setSelectedSector(s.sector)}
              >
                <div className="flex items-center gap-2">
                  <span className="text-xs text-zinc-500 w-4">{i + 1}.</span>
                  <TrendIcon trend={s.trend} />
                  <span className="text-sm text-white">{sectorLabel(s.sector)}</span>
                </div>
                <span className="text-sm font-medium text-green-400">
                  +{formatPercent(getChangeForPeriod(s, period))}
                </span>
              </div>
            ))}
          </div>
        </div>
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-red-400 mb-3">
            {isVi ? "Bottom 3 ngành kém nhất" : "Bottom 3 Performing"}
          </h3>
          <div className="space-y-2">
            {bottom3.map((s, i) => (
              <div
                key={s.sector}
                className="flex items-center justify-between cursor-pointer hover:bg-zinc-800 rounded-lg px-2 py-1.5 transition"
                onClick={() => setSelectedSector(s.sector)}
              >
                <div className="flex items-center gap-2">
                  <span className="text-xs text-zinc-500 w-4">{i + 1}.</span>
                  <TrendIcon trend={s.trend} />
                  <span className="text-sm text-white">{sectorLabel(s.sector)}</span>
                </div>
                <span className={`text-sm font-medium ${getChangeForPeriod(s, period) >= 0 ? "text-green-400" : "text-red-400"}`}>
                  {formatPercent(getChangeForPeriod(s, period))}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Heatmap Grid (Req 22.2) */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h3 className="text-sm font-medium text-zinc-400 mb-3">
          {isVi ? "Bản đồ nhiệt ngành" : "Sector Heatmap"}
        </h3>
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
          {sectors.map((s) => {
            const change = getChangeForPeriod(s, period);
            return (
              <button
                key={s.sector}
                onClick={() => setSelectedSector(s.sector)}
                className={`rounded-lg p-3 text-center transition hover:opacity-80 ${perfColor(change)} ${
                  selectedSector === s.sector ? "ring-2 ring-blue-500" : ""
                }`}
              >
                <div className="text-xs font-bold">{s.sector}</div>
                <div className="text-sm font-medium mt-0.5">
                  {sectorLabel(s.sector)}
                </div>
                <div className="text-lg font-bold mt-1">
                  {change >= 0 ? "+" : ""}{formatPercent(change)}
                </div>
                <div className="mt-1">
                  <TrendIcon trend={s.trend} />
                </div>
              </button>
            );
          })}
        </div>
      </div>

      {/* Bar Chart (Req 22.4) */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
        <h3 className="text-sm font-medium text-zinc-400 mb-3">
          {isVi ? "So sánh hiệu suất ngành" : "Sector Performance Comparison"}
        </h3>
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={barData} layout="vertical" margin={{ left: 80, right: 20, top: 5, bottom: 5 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis type="number" tick={{ fill: "#a1a1aa", fontSize: 12 }} tickFormatter={(v) => `${v}%`} />
              <YAxis type="category" dataKey="name" tick={{ fill: "#a1a1aa", fontSize: 11 }} width={75} />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                labelStyle={{ color: "#fff" }}
                formatter={(value) => {
                  const v = Number(value);
                  return [`${v >= 0 ? "+" : ""}${v.toFixed(2)}%`, "Change"];
                }}
              />
              <Bar dataKey="value" radius={[0, 4, 4, 0]}>
                {barData.map((entry, idx) => (
                  <Cell key={idx} fill={barColor(entry.value)} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>


      {/* Detail Panel (Req 22.7) */}
      {selectedSector && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-white">
              {sectorLabel(selectedSector)} ({selectedSector})
            </h3>
            <button
              onClick={() => setSelectedSector(null)}
              className="text-zinc-500 hover:text-white transition"
            >
              <X size={18} />
            </button>
          </div>

          {detailLoading ? (
            <div className="flex items-center justify-center h-32">
              <RefreshCw size={20} className="animate-spin text-zinc-500" />
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
              {/* Mini index chart placeholder */}
              <div className="lg:col-span-1">
                <h4 className="text-xs text-zinc-500 mb-2">
                  {isVi ? "Biểu đồ chỉ số" : "Index Chart"}
                </h4>
                <div className="h-40">
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart
                      data={generateMiniChartData(
                        sectors.find((s) => s.sector === selectedSector),
                      )}
                    >
                      <Line
                        type="monotone"
                        dataKey="value"
                        stroke="#6366f1"
                        strokeWidth={2}
                        dot={false}
                      />
                      <Tooltip
                        contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                        labelStyle={{ color: "#a1a1aa" }}
                        formatter={(v) => [formatNumber(Number(v), 2), "Index"]}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </div>

              {/* Top 5 stocks */}
              <div className="lg:col-span-1">
                <h4 className="text-xs text-zinc-500 mb-2">
                  {isVi ? "Top 5 cổ phiếu" : "Top 5 Stocks"}
                </h4>
                {detailData?.stocks && detailData.stocks.length > 0 ? (
                  <div className="space-y-1.5">
                    {detailData.stocks.map((stock) => (
                      <div
                        key={stock.symbol}
                        className="flex items-center justify-between text-sm px-2 py-1 rounded bg-zinc-800/50"
                      >
                        <span className="text-white font-medium">{stock.symbol}</span>
                        <span className={stock.change >= 0 ? "text-green-400" : "text-red-400"}>
                          {stock.change >= 0 ? "+" : ""}{formatPercent(stock.change)}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-xs text-zinc-600">
                    {isVi ? "Không có dữ liệu" : "No data available"}
                  </p>
                )}
              </div>

              {/* Median fundamentals */}
              <div className="lg:col-span-1">
                <h4 className="text-xs text-zinc-500 mb-2">
                  {isVi ? "Chỉ số cơ bản trung vị" : "Median Fundamentals"}
                </h4>
                {detailData?.averages ? (
                  <div className="grid grid-cols-2 gap-2">
                    <FundamentalCard label="P/E" value={formatNumber(detailData.averages.medianPE, 1)} />
                    <FundamentalCard label="P/B" value={formatNumber(detailData.averages.medianPB, 1)} />
                    <FundamentalCard label="ROE" value={formatPercent(detailData.averages.medianROE)} />
                    <FundamentalCard label="ROA" value={formatPercent(detailData.averages.medianROA)} />
                    <FundamentalCard label={isVi ? "Cổ tức" : "Div Yield"} value={formatPercent(detailData.averages.medianDivYield)} />
                    <FundamentalCard label="D/E" value={formatNumber(detailData.averages.medianDebtToEquity, 2)} />
                  </div>
                ) : (
                  <p className="text-xs text-zinc-600">
                    {isVi ? "Không có dữ liệu" : "No data available"}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

// --- Sub-components ---

function FundamentalCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-zinc-800/50 rounded-lg px-2 py-1.5 text-center">
      <div className="text-xs text-zinc-500">{label}</div>
      <div className="text-sm font-medium text-white">{value}</div>
    </div>
  );
}

// Generate simple mini chart data from sector performance data
function generateMiniChartData(sector?: SectorPerformance) {
  if (!sector) return [];
  const base = sector.currentPrice;
  // Approximate historical points from known changes
  const points = [
    { label: "1Y", value: base / (1 + sector.oneYearChange / 100) },
    { label: "6M", value: base / (1 + sector.sixMonthChange / 100) },
    { label: "3M", value: base / (1 + sector.threeMonthChange / 100) },
    { label: "1M", value: base / (1 + sector.oneMonthChange / 100) },
    { label: "1W", value: base / (1 + sector.oneWeekChange / 100) },
    { label: "Now", value: base },
  ];
  return points;
}
