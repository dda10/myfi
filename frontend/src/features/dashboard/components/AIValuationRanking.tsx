"use client";

import { useState } from "react";
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell,
} from "recharts";
import { useI18n } from "@/context/I18nContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";
import { useApp } from "@/context/AppContext";
import { useRouter } from "next/navigation";

interface RankedStock {
  symbol: string;
  name: string;
  currentPrice: number;
  targetPrice: number;
  upsidePercent: number;
}

const EXCHANGE_TABS = ["VN30", "VN100", "HOSE", "HNX", "UPCOM"];

function upsideColor(pct: number): string {
  if (pct >= 50) return "#22c55e";
  if (pct >= 30) return "#86efac";
  if (pct >= 10) return "#a3e635";
  if (pct >= 0)  return "#71717a";
  return "#ef4444";
}

function upsideClass(pct: number): string {
  if (pct >= 50) return "upside-high";
  if (pct >= 20) return "upside-mid";
  if (pct >= 0)  return "upside-zero";
  return "upside-neg";
}

export function AIValuationRanking() {
  const { formatCurrency, formatPercent } = useI18n();
  const { setActiveSymbol } = useApp();
  const router = useRouter();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;
  const [activeExchange, setActiveExchange] = useState("VN100");

  const { data, loading } = usePolling<RankedStock[]>(
    () => apiFetch<RankedStock[]>("/api/ranking"),
    interval,
  );

  const stocks = Array.isArray(data) ? data.slice(0, 10) : [];

  // Distribution data for mini bar chart
  const downside = stocks.filter(s => s.upsidePercent < 0).length;
  const sideways = stocks.filter(s => s.upsidePercent >= 0 && s.upsidePercent < 20).length;
  const upside   = stocks.filter(s => s.upsidePercent >= 20).length;

  const distData = [
    { label: "Downside", value: downside, fill: "#ef4444" },
    { label: "Sideways", value: sideways, fill: "#eab308" },
    { label: "Upside",   value: upside,   fill: "#22c55e" },
  ];

  const handleRowClick = (symbol: string) => {
    setActiveSymbol(symbol);
    router.push(`/stock/${symbol}`);
  };

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-4 pt-3 pb-2 border-b border-border-theme">
        <h2 className="text-xs font-bold text-foreground mb-2">Định giá theo Miquant AI</h2>
        {/* Exchange tabs */}
        <div className="flex gap-1 flex-wrap">
          {EXCHANGE_TABS.map(tab => (
            <button
              key={tab}
              onClick={() => setActiveExchange(tab)}
              className={`px-2.5 py-0.5 rounded text-[11px] font-medium transition ${
                activeExchange === tab
                  ? "bg-indigo-600 text-white"
                  : "text-text-muted hover:text-foreground hover:bg-surface"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      {/* Column headers */}
      <div className="grid grid-cols-[1fr_auto_auto_auto] gap-x-3 px-4 py-2 border-b border-border-theme text-[10px] text-text-muted font-semibold uppercase tracking-wider">
        <span>Mã CP</span>
        <span className="text-right">Giá hiện tại</span>
        <span className="text-right">Giá mục tiêu</span>
        <span className="text-right">Upside</span>
      </div>

      {/* Rows */}
      {loading && !data ? (
        <div className="p-3 space-y-1.5">
          {[1,2,3,4,5].map(i => <div key={i} className="h-7 bg-surface rounded animate-pulse" />)}
        </div>
      ) : stocks.length > 0 ? (
        <div className="divide-y divide-border-theme">
          {stocks.map(stock => (
            <div
              key={stock.symbol}
              onClick={() => handleRowClick(stock.symbol)}
              className="grid grid-cols-[1fr_auto_auto_auto] gap-x-3 px-4 py-2 items-center hover:bg-surface transition cursor-pointer text-[11px]"
            >
              <div>
                <span className="font-bold text-foreground">{stock.symbol}</span>
              </div>
              <span className="text-foreground tabular-nums font-medium">{stock.currentPrice.toLocaleString()}</span>
              <span className="text-text-muted tabular-nums">{stock.targetPrice.toLocaleString()}</span>
              <span
                className="font-bold tabular-nums text-right"
                style={{ color: upsideColor(stock.upsidePercent) }}
              >
                {stock.upsidePercent > 0 ? "+" : ""}{stock.upsidePercent.toFixed(2)}%
              </span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-xs text-text-muted p-4 text-center">Không có dữ liệu</p>
      )}

      {/* Distribution mini chart */}
      {stocks.length > 0 && (
        <div className="px-4 py-3 border-t border-border-theme">
          <p className="text-[10px] text-text-muted mb-2 font-medium">Phân phối Định giá</p>
          <div className="flex items-center gap-3 text-[10px] text-text-muted mb-2">
            <span className="flex items-center gap-1"><span className="w-2 h-2 bg-red-500 rounded-sm inline-block"/>Downside: {downside}%</span>
            <span className="flex items-center gap-1"><span className="w-2 h-2 bg-yellow-500 rounded-sm inline-block"/>Sideways: {sideways}%</span>
            <span className="flex items-center gap-1"><span className="w-2 h-2 bg-green-500 rounded-sm inline-block"/>Upside: {upside}%</span>
          </div>
          <div style={{ height: 60 }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={distData} margin={{ top: 0, right: 0, left: -30, bottom: 0 }} barSize={24}>
                <YAxis tick={false} axisLine={false} tickLine={false} />
                <XAxis dataKey="label" tick={{ fontSize: 9, fill: "#71717a" }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{ backgroundColor: "#111", border: "1px solid #1f1f1f", borderRadius: 6, fontSize: 11 }}
                  cursor={{ fill: "rgba(255,255,255,0.03)" }}
                />
                <Bar dataKey="value" radius={[3, 3, 0, 0]}>
                  {distData.map((entry, i) => (
                    <Cell key={i} fill={entry.fill} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </section>
  );
}
