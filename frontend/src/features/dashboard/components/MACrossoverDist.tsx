"use client";

import { useState } from "react";
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
} from "recharts";
import { useI18n } from "@/context/I18nContext";
import { useTheme } from "@/context/ThemeContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";
import { useApp } from "@/context/AppContext";
import { TrendingUp, TrendingDown } from "lucide-react";

interface MACrossoverData {
  ma: string;
  above: number;
  below: number;
}

interface MAStock {
  symbol: string;
  currentPrice: number;
  maValue: number;
  pctFromMA: number;
}

interface MADetailResponse {
  above: MAStock[];
  below: MAStock[];
}

const MA_OPTIONS = ["MA50", "MA100", "MA200"];
const EXCHANGE_OPTIONS = ["VN30", "VN100", "HOSE", "HNX", "UPCOM"];

export function MACrossoverDist() {
  const { t } = useI18n();
  const { theme } = useTheme();
  const { setActiveSymbol } = useApp();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;
  const isDark = theme === "dark";

  const [activeMa, setActiveMa] = useState("MA200");
  const [activeExchange, setActiveExchange] = useState("VN100");

  const { data, loading } = usePolling<MACrossoverData[]>(
    () => apiFetch<MACrossoverData[]>("/api/market/statistics"),
    interval,
  );

  const chartData = Array.isArray(data) ? data : [
    { ma: "MA50", above: 0, below: 0 },
    { ma: "MA100", above: 0, below: 0 },
    { ma: "MA200", above: 0, below: 0 },
  ];

  // Mock top stocks data — replace with real API when available
  const mockAbove: MAStock[] = [
    { symbol: "BSR", currentPrice: 26350, maValue: 18339, pctFromMA: 43.69 },
    { symbol: "BVH", currentPrice: 85000, maValue: 60055, pctFromMA: 41.54 },
    { symbol: "VIC", currentPrice: 141000, maValue: 106641, pctFromMA: 32.22 },
    { symbol: "PVD", currentPrice: 33600, maValue: 26007, pctFromMA: 29.20 },
    { symbol: "GEE", currentPrice: 194300, maValue: 152887, pctFromMA: 27.09 },
  ];
  const mockBelow: MAStock[] = [
    { symbol: "DGC", currentPrice: 56500, maValue: 83996, pctFromMA: -32.73 },
    { symbol: "HDC", currentPrice: 18500, maValue: 26090, pctFromMA: -29.09 },
    { symbol: "VIX", currentPrice: 16200, maValue: 21538, pctFromMA: -24.78 },
    { symbol: "DXS", currentPrice: 7470, maValue: 9808, pctFromMA: -23.84 },
    { symbol: "DIG", currentPrice: 14150, maValue: 18524, pctFromMA: -23.61 },
  ];

  const activeBarData = chartData.find(d => d.ma === activeMa) ?? { ma: activeMa, above: 0, below: 0 };

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-4 pt-3 pb-2 border-b border-border-theme">
        <h2 className="text-xs font-bold text-foreground mb-2">Top Cổ phiếu vượt MA</h2>
        {/* MA tabs */}
        <div className="flex gap-1 mb-2">
          {MA_OPTIONS.map(ma => (
            <button
              key={ma}
              onClick={() => setActiveMa(ma)}
              className={`px-2.5 py-0.5 rounded text-[11px] font-semibold transition ${
                activeMa === ma
                  ? "bg-indigo-600 text-white"
                  : "bg-surface text-text-muted hover:text-foreground hover:bg-surface-hover"
              }`}
            >
              {ma}
            </button>
          ))}
        </div>
        {/* Exchange filter */}
        <div className="flex gap-1 flex-wrap">
          {EXCHANGE_OPTIONS.map(ex => (
            <button
              key={ex}
              onClick={() => setActiveExchange(ex)}
              className={`px-2 py-0.5 rounded text-[10px] font-medium transition ${
                activeExchange === ex
                  ? "bg-surface-hover text-foreground border border-border-theme"
                  : "text-text-muted hover:text-foreground"
              }`}
            >
              {ex}
            </button>
          ))}
        </div>
      </div>

      <div className="p-3 space-y-3">
        {/* Bar chart */}
        {loading && !data ? (
          <div className="h-36 bg-surface rounded animate-pulse" />
        ) : (
          <div className="h-36">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData} margin={{ top: 2, right: 4, left: -20, bottom: 2 }} barSize={14}>
                <CartesianGrid strokeDasharray="3 3" stroke={isDark ? "#1f1f1f" : "#e5e7eb"} />
                <XAxis dataKey="ma" tick={{ fill: isDark ? "#71717a" : "#6b7280", fontSize: 10 }} />
                <YAxis tick={{ fill: isDark ? "#71717a" : "#6b7280", fontSize: 10 }} />
                <Tooltip
                  contentStyle={{ backgroundColor: isDark ? "#111" : "#fff", border: `1px solid ${isDark ? "#1f1f1f" : "#e5e7eb"}`, borderRadius: 6, color: isDark ? "#ededed" : "#111", fontSize: 11 }}
                />
                <Bar dataKey="above" name="Vượt MA" fill="#22c55e" radius={[3, 3, 0, 0]} />
                <Bar dataKey="below" name="Dưới MA" fill="#ef4444" radius={[3, 3, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}

        {/* Distribution label */}
        <div className="flex items-center gap-3 text-[10px] text-text-muted">
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-red-500 inline-block"/>Dưới: {activeBarData.below}</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-green-500 inline-block"/>Vượt: {activeBarData.above}</span>
        </div>

        {/* Top 5 Above */}
        <div>
          <p className="text-[10px] font-bold text-green-400 mb-1">Top 5 vượt {activeMa}</p>
          <div className="space-y-0.5">
            {mockAbove.map(s => (
              <div
                key={s.symbol}
                className="flex items-center justify-between px-2 py-1 rounded hover:bg-surface transition cursor-pointer text-[11px]"
                onClick={() => setActiveSymbol(s.symbol)}
              >
                <span className="font-semibold text-foreground w-12">{s.symbol}</span>
                <span className="text-text-muted">{s.currentPrice.toLocaleString()}</span>
                <span className="text-text-muted">{s.maValue.toLocaleString()}</span>
                <span className="text-green-400 font-bold flex items-center gap-0.5">
                  <TrendingUp size={10}/>{s.pctFromMA.toFixed(2)}%
                </span>
              </div>
            ))}
          </div>
        </div>

        {/* Top 5 Below */}
        <div>
          <p className="text-[10px] font-bold text-red-400 mb-1">Top 5 dưới {activeMa}</p>
          <div className="space-y-0.5">
            {mockBelow.map(s => (
              <div
                key={s.symbol}
                className="flex items-center justify-between px-2 py-1 rounded hover:bg-surface transition cursor-pointer text-[11px]"
                onClick={() => setActiveSymbol(s.symbol)}
              >
                <span className="font-semibold text-foreground w-12">{s.symbol}</span>
                <span className="text-text-muted">{s.currentPrice.toLocaleString()}</span>
                <span className="text-text-muted">{s.maValue.toLocaleString()}</span>
                <span className="text-red-400 font-bold flex items-center gap-0.5">
                  <TrendingDown size={10}/>{s.pctFromMA.toFixed(2)}%
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
