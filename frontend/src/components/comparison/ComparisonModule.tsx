"use client";

import React, { useState, useCallback, useEffect, useMemo } from "react";
import { X, Plus, Trash2 } from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { useI18n } from "@/context/I18nContext";

// --- Types ---

type TimePeriod = "3M" | "6M" | "1Y" | "3Y" | "5Y";
type ComparisonTab = "valuation" | "performance" | "correlation";

interface ValuationPoint {
  timestamp: string;
  pe: number;
  pb: number;
}

interface ValuationSeries {
  symbol: string;
  data: ValuationPoint[];
}

interface ValuationResult {
  series: ValuationSeries[];
  period: TimePeriod;
  warnings?: string[];
}

interface PerformancePoint {
  timestamp: string;
  returnPercent: number;
}

interface PerformanceSeries {
  symbol: string;
  data: PerformancePoint[];
}

interface PerformanceResult {
  series: PerformanceSeries[];
  period: TimePeriod;
  warnings?: string[];
}

interface CorrelationResult {
  symbols: string[];
  matrix: number[][];
  period: TimePeriod;
  warnings?: string[];
}

// --- Constants ---

const STOCK_COLORS = [
  "#3b82f6", "#ef4444", "#10b981", "#f59e0b", "#8b5cf6",
  "#ec4899", "#06b6d4", "#f97316", "#84cc16", "#6366f1",
];

const PERIODS: TimePeriod[] = ["3M", "6M", "1Y", "3Y", "5Y"];

const SECTOR_OPTIONS = [
  { code: "VNFIN", label: "Tài chính / Finance" },
  { code: "VNREAL", label: "Bất động sản / Real Estate" },
  { code: "VNIT", label: "Công nghệ / Technology" },
  { code: "VNIND", label: "Công nghiệp / Industrial" },
  { code: "VNCONS", label: "Tiêu dùng / Consumer" },
  { code: "VNCOND", label: "Tiêu dùng thiết yếu / Consumer Staples" },
  { code: "VNHEAL", label: "Y tế / Healthcare" },
  { code: "VNENE", label: "Năng lượng / Energy" },
  { code: "VNUTI", label: "Tiện ích / Utilities" },
  { code: "VNMAT", label: "Vật liệu / Materials" },
];

const MAX_STOCKS = 10;

// --- Component ---

export function ComparisonModule() {
  const { t, formatNumber } = useI18n();

  const [symbols, setSymbols] = useState<string[]>([]);
  const [inputValue, setInputValue] = useState("");
  const [activeTab, setActiveTab] = useState<ComparisonTab>("valuation");
  const [period, setPeriod] = useState<TimePeriod>("1Y");
  const [hiddenSymbols, setHiddenSymbols] = useState<Set<string>>(new Set());

  const [valuationData, setValuationData] = useState<ValuationResult | null>(null);
  const [performanceData, setPerformanceData] = useState<PerformanceResult | null>(null);
  const [correlationData, setCorrelationData] = useState<CorrelationResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Add symbol
  const addSymbol = useCallback((sym: string) => {
    const upper = sym.trim().toUpperCase();
    if (!upper || symbols.includes(upper) || symbols.length >= MAX_STOCKS) return;
    setSymbols((prev) => [...prev, upper]);
  }, [symbols]);

  // Remove symbol
  const removeSymbol = useCallback((sym: string) => {
    setSymbols((prev) => prev.filter((s) => s !== sym));
    setHiddenSymbols((prev) => {
      const next = new Set(prev);
      next.delete(sym);
      return next;
    });
  }, []);

  // Clear all
  const clearAll = useCallback(() => {
    setSymbols([]);
    setHiddenSymbols(new Set());
    setValuationData(null);
    setPerformanceData(null);
    setCorrelationData(null);
  }, []);

  // Toggle visibility
  const toggleVisibility = useCallback((sym: string) => {
    setHiddenSymbols((prev) => {
      const next = new Set(prev);
      if (next.has(sym)) next.delete(sym);
      else next.add(sym);
      return next;
    });
  }, []);

  // Handle input submit
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter" && inputValue.trim()) {
        addSymbol(inputValue);
        setInputValue("");
      }
    },
    [inputValue, addSymbol],
  );

  // Sector auto-populate
  const handleSectorSelect = useCallback(
    async (sectorCode: string) => {
      if (!sectorCode) return;
      try {
        const res = await fetch(`http://localhost:8080/api/sectors/${sectorCode}/stocks`);
        if (!res.ok) return;
        const data = await res.json();
        const sectorSymbols: string[] = data.symbols || [];
        setSymbols((prev) => {
          const combined = [...new Set([...prev, ...sectorSymbols])];
          return combined.slice(0, MAX_STOCKS);
        });
      } catch {
        // silently fail
      }
    },
    [],
  );

  // Fetch data when symbols/period/tab change
  useEffect(() => {
    if (symbols.length < 2) return;

    const controller = new AbortController();
    setLoading(true);
    setError(null);

    const symbolsParam = symbols.join(",");
    const endpoint =
      activeTab === "valuation"
        ? "valuation"
        : activeTab === "performance"
          ? "performance"
          : "correlation";

    fetch(
      `http://localhost:8080/api/comparison/${endpoint}?symbols=${symbolsParam}&period=${period}`,
      { signal: controller.signal },
    )
      .then((res) => {
        if (!res.ok) throw new Error("API error");
        return res.json();
      })
      .then((data) => {
        if (activeTab === "valuation") setValuationData(data);
        else if (activeTab === "performance") setPerformanceData(data);
        else setCorrelationData(data);
      })
      .catch((err) => {
        if (err?.name !== "AbortError") setError(t("comparison.error"));
      })
      .finally(() => setLoading(false));

    return () => controller.abort();
  }, [symbols, period, activeTab, t]);

  const tabs: { key: ComparisonTab; label: string }[] = [
    { key: "valuation", label: t("comparison.valuation") },
    { key: "performance", label: t("comparison.performance") },
    { key: "correlation", label: t("comparison.correlation") },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-foreground">{t("comparison.title")}</h1>
        {symbols.length > 0 && (
          <button
            onClick={clearAll}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 transition-colors"
          >
            <Trash2 size={14} />
            {t("comparison.clear_all")}
          </button>
        )}
      </div>

      {/* Stock Selector */}
      <div className="bg-surface border border-border-theme rounded-xl p-4 space-y-3">
        <div className="flex flex-wrap gap-2 items-center">
          {symbols.map((sym, i) => (
            <span
              key={sym}
              className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium text-white"
              style={{ backgroundColor: STOCK_COLORS[i % STOCK_COLORS.length] + "cc" }}
            >
              {sym}
              <button onClick={() => removeSymbol(sym)} className="hover:opacity-70">
                <X size={14} />
              </button>
            </span>
          ))}
          {symbols.length < MAX_STOCKS && (
            <div className="relative">
              <input
                type="text"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value.toUpperCase())}
                onKeyDown={handleKeyDown}
                placeholder={t("comparison.add_symbol")}
                className="w-36 px-3 py-1.5 text-sm rounded-lg bg-surface-hover border border-border-theme text-foreground placeholder:text-text-muted focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
          )}
        </div>

        {/* Sector dropdown */}
        <div className="flex items-center gap-3">
          <select
            onChange={(e) => handleSectorSelect(e.target.value)}
            defaultValue=""
            className="px-3 py-1.5 text-sm rounded-lg bg-surface-hover border border-border-theme text-foreground focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="">{t("comparison.sector_dropdown")}</option>
            {SECTOR_OPTIONS.map((s) => (
              <option key={s.code} value={s.code}>
                {s.label}
              </option>
            ))}
          </select>
          {symbols.length >= MAX_STOCKS && (
            <span className="text-xs text-amber-400">{t("comparison.max_stocks")}</span>
          )}
        </div>
      </div>

      {/* Prompt when < 2 stocks */}
      {symbols.length < 2 ? (
        <div className="flex items-center justify-center h-64 bg-surface border border-border-theme rounded-xl">
          <p className="text-text-muted text-lg">{t("comparison.select_prompt")}</p>
        </div>
      ) : (
        <>
          {/* Tabs + Period selector */}
          <div className="flex items-center justify-between">
            <div className="flex gap-1 bg-surface border border-border-theme rounded-lg p-1">
              {tabs.map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key)}
                  className={`px-4 py-1.5 text-sm font-medium rounded-md transition-colors ${
                    activeTab === tab.key
                      ? "bg-blue-600 text-white"
                      : "text-text-muted hover:text-foreground"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            <div className="flex gap-1 bg-surface border border-border-theme rounded-lg p-1">
              {PERIODS.map((p) => (
                <button
                  key={p}
                  onClick={() => setPeriod(p)}
                  className={`px-3 py-1 text-sm font-medium rounded-md transition-colors ${
                    period === p
                      ? "bg-blue-600 text-white"
                      : "text-text-muted hover:text-foreground"
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          {/* Visibility toggles */}
          {activeTab !== "correlation" && (
            <div className="flex flex-wrap gap-2">
              {symbols.map((sym, i) => (
                <button
                  key={sym}
                  onClick={() => toggleVisibility(sym)}
                  className={`px-3 py-1 text-xs rounded-full border transition-colors ${
                    hiddenSymbols.has(sym)
                      ? "border-border-theme text-text-muted opacity-50"
                      : "text-white"
                  }`}
                  style={{
                    borderColor: hiddenSymbols.has(sym) ? undefined : STOCK_COLORS[i % STOCK_COLORS.length],
                    backgroundColor: hiddenSymbols.has(sym) ? "transparent" : STOCK_COLORS[i % STOCK_COLORS.length] + "22",
                  }}
                >
                  {sym}
                </button>
              ))}
            </div>
          )}

          {/* Chart area */}
          <div className="bg-surface border border-border-theme rounded-xl p-4 min-h-[400px]">
            {loading ? (
              <div className="flex items-center justify-center h-80">
                <p className="text-text-muted">{t("comparison.loading")}</p>
              </div>
            ) : error ? (
              <div className="flex items-center justify-center h-80">
                <p className="text-red-400">{error}</p>
              </div>
            ) : activeTab === "valuation" ? (
              <ValuationChart
                data={valuationData}
                symbols={symbols}
                hiddenSymbols={hiddenSymbols}
                formatNumber={formatNumber}
                t={t}
              />
            ) : activeTab === "performance" ? (
              <PerformanceChart
                data={performanceData}
                symbols={symbols}
                hiddenSymbols={hiddenSymbols}
                formatNumber={formatNumber}
                t={t}
              />
            ) : (
              <CorrelationMatrix
                data={correlationData}
                formatNumber={formatNumber}
              />
            )}
          </div>
        </>
      )}
    </div>
  );
}


// --- Valuation Chart ---

function ValuationChart({
  data,
  symbols,
  hiddenSymbols,
  formatNumber,
  t,
}: {
  data: ValuationResult | null;
  symbols: string[];
  hiddenSymbols: Set<string>;
  formatNumber: (v: number, d?: number) => string;
  t: (key: string) => string;
}) {
  const chartData = useMemo(() => {
    if (!data?.series?.length) return [];
    // Merge all series into unified date-keyed rows
    const dateMap = new Map<string, Record<string, number | string>>();
    for (const series of data.series) {
      if (hiddenSymbols.has(series.symbol)) continue;
      for (const pt of series.data) {
        const dateKey = pt.timestamp.slice(0, 10);
        if (!dateMap.has(dateKey)) dateMap.set(dateKey, { date: dateKey });
        const row = dateMap.get(dateKey)!;
        row[`${series.symbol}_pe`] = pt.pe;
        row[`${series.symbol}_pb`] = pt.pb;
      }
    }
    return Array.from(dateMap.values()).sort((a, b) =>
      (a.date as string).localeCompare(b.date as string),
    );
  }, [data, hiddenSymbols]);

  if (!chartData.length) {
    return <div className="flex items-center justify-center h-80 text-text-muted">No data</div>;
  }

  const visibleSymbols = symbols.filter((s) => !hiddenSymbols.has(s));

  return (
    <div className="space-y-6">
      {/* P/E Chart */}
      <div>
        <h3 className="text-sm font-medium text-text-muted mb-2">{t("comparison.pe_ratio")}</h3>
        <ResponsiveContainer width="100%" height={250}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: "#888" }} />
            <YAxis tick={{ fontSize: 11, fill: "#888" }} />
            <Tooltip
              contentStyle={{ backgroundColor: "#1a1a2e", border: "1px solid #333", borderRadius: 8 }}
              labelStyle={{ color: "#aaa" }}
              formatter={(value) => formatNumber(Number(value ?? 0), 2)}
            />
            <Legend />
            {visibleSymbols.map((sym, i) => (
              <Line
                key={sym}
                type="monotone"
                dataKey={`${sym}_pe`}
                name={`${sym} P/E`}
                stroke={STOCK_COLORS[symbols.indexOf(sym) % STOCK_COLORS.length]}
                dot={false}
                strokeWidth={2}
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* P/B Chart */}
      <div>
        <h3 className="text-sm font-medium text-text-muted mb-2">{t("comparison.pb_ratio")}</h3>
        <ResponsiveContainer width="100%" height={250}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: "#888" }} />
            <YAxis tick={{ fontSize: 11, fill: "#888" }} />
            <Tooltip
              contentStyle={{ backgroundColor: "#1a1a2e", border: "1px solid #333", borderRadius: 8 }}
              labelStyle={{ color: "#aaa" }}
              formatter={(value) => formatNumber(Number(value ?? 0), 2)}
            />
            <Legend />
            {visibleSymbols.map((sym, i) => (
              <Line
                key={sym}
                type="monotone"
                dataKey={`${sym}_pb`}
                name={`${sym} P/B`}
                stroke={STOCK_COLORS[symbols.indexOf(sym) % STOCK_COLORS.length]}
                dot={false}
                strokeWidth={2}
                strokeDasharray="5 5"
              />
            ))}
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// --- Performance Chart ---

function PerformanceChart({
  data,
  symbols,
  hiddenSymbols,
  formatNumber,
  t,
}: {
  data: PerformanceResult | null;
  symbols: string[];
  hiddenSymbols: Set<string>;
  formatNumber: (v: number, d?: number) => string;
  t: (key: string) => string;
}) {
  const chartData = useMemo(() => {
    if (!data?.series?.length) return [];
    const dateMap = new Map<string, Record<string, number | string>>();
    for (const series of data.series) {
      if (hiddenSymbols.has(series.symbol)) continue;
      for (const pt of series.data) {
        const dateKey = pt.timestamp.slice(0, 10);
        if (!dateMap.has(dateKey)) dateMap.set(dateKey, { date: dateKey });
        dateMap.get(dateKey)![series.symbol] = pt.returnPercent;
      }
    }
    return Array.from(dateMap.values()).sort((a, b) =>
      (a.date as string).localeCompare(b.date as string),
    );
  }, [data, hiddenSymbols]);

  if (!chartData.length) {
    return <div className="flex items-center justify-center h-80 text-text-muted">No data</div>;
  }

  const visibleSymbols = symbols.filter((s) => !hiddenSymbols.has(s));

  return (
    <div>
      <h3 className="text-sm font-medium text-text-muted mb-2">{t("comparison.return_pct")}</h3>
      <ResponsiveContainer width="100%" height={400}>
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#333" />
          <XAxis dataKey="date" tick={{ fontSize: 11, fill: "#888" }} />
          <YAxis
            tick={{ fontSize: 11, fill: "#888" }}
            tickFormatter={(v) => `${v}%`}
          />
          <Tooltip
            contentStyle={{ backgroundColor: "#1a1a2e", border: "1px solid #333", borderRadius: 8 }}
            labelStyle={{ color: "#aaa" }}
            formatter={(value) => `${formatNumber(Number(value ?? 0), 2)}%`}
          />
          <Legend />
          {visibleSymbols.map((sym) => (
            <Line
              key={sym}
              type="monotone"
              dataKey={sym}
              name={sym}
              stroke={STOCK_COLORS[symbols.indexOf(sym) % STOCK_COLORS.length]}
              dot={false}
              strokeWidth={2}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

// --- Correlation Matrix ---

function CorrelationMatrix({
  data,
  formatNumber,
}: {
  data: CorrelationResult | null;
  formatNumber: (v: number, d?: number) => string;
}) {
  if (!data?.matrix?.length || !data?.symbols?.length) {
    return <div className="flex items-center justify-center h-80 text-text-muted">No data</div>;
  }

  const { symbols: syms, matrix } = data;

  const getColor = (val: number): string => {
    if (val >= 0.7) return "bg-green-600/80";
    if (val >= 0.3) return "bg-green-600/40";
    if (val > -0.3) return "bg-zinc-700/60";
    if (val > -0.7) return "bg-red-600/40";
    return "bg-red-600/80";
  };

  return (
    <div className="overflow-x-auto">
      <table className="w-auto mx-auto border-collapse">
        <thead>
          <tr>
            <th className="p-2" />
            {syms.map((s) => (
              <th key={s} className="p-2 text-xs font-medium text-text-muted text-center">
                {s}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {syms.map((rowSym, ri) => (
            <tr key={rowSym}>
              <td className="p-2 text-xs font-medium text-text-muted text-right pr-3">
                {rowSym}
              </td>
              {syms.map((colSym, ci) => {
                const val = matrix[ri]?.[ci] ?? 0;
                return (
                  <td
                    key={colSym}
                    className={`p-2 text-center text-xs font-mono rounded ${getColor(val)}`}
                    style={{ minWidth: 56 }}
                    title={`${rowSym} vs ${colSym}`}
                  >
                    <span className="text-white">{formatNumber(val, 2)}</span>
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
