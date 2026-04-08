"use client";

import { useState, useEffect } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

// --- Types ---

interface NAVSnapshot {
  date: string;
  nav: number;
}

interface PerformanceData {
  twr: number;
  xirr: number;
  equityCurve: NAVSnapshot[];
  benchmarkVNIndex?: NAVSnapshot[];
  benchmarkVN30?: NAVSnapshot[];
}

type TimePeriod = "1W" | "1M" | "3M" | "6M" | "1Y" | "YTD" | "ALL";
const TIME_PERIODS: TimePeriod[] = ["1W", "1M", "3M", "6M", "1Y", "YTD", "ALL"];

export function PerformanceChart() {
  const { t, formatPercent, formatCurrency, formatDate } = useI18n();
  const [period, setPeriod] = useState<TimePeriod>("1Y");
  const [data, setData] = useState<PerformanceData | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    apiFetch<PerformanceData>(`/api/portfolio/performance?period=${period}`).then((resp) => {
      if (!cancelled) {
        setData(resp);
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, [period]);

  // Merge equity curve with benchmarks for chart
  const chartData = data?.equityCurve?.map((point, i) => ({
    date: point.date,
    nav: point.nav,
    vnindex: data.benchmarkVNIndex?.[i]?.nav,
    vn30: data.benchmarkVN30?.[i]?.nav,
  })) ?? [];

  return (
    <div className="space-y-6">
      {/* Period selector */}
      <div className="flex items-center gap-2">
        {TIME_PERIODS.map((p) => (
          <button
            key={p}
            onClick={() => setPeriod(p)}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
              period === p
                ? "bg-indigo-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-700"
            }`}
          >
            {p}
          </button>
        ))}
      </div>

      {/* TWR / XIRR cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <MetricCard
          label="TWR"
          value={data ? formatPercent(data.twr) : "—"}
          positive={data ? data.twr >= 0 : null}
        />
        <MetricCard
          label="XIRR"
          value={data ? formatPercent(data.xirr) : "—"}
          positive={data ? data.xirr >= 0 : null}
        />
      </div>

      {/* Equity curve chart */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
        <h3 className="text-lg font-bold text-white mb-4">{t("portfolio.performance")}</h3>
        {loading ? (
          <div className="flex items-center justify-center h-64 text-zinc-500">{t("common.loading")}</div>
        ) : chartData.length > 0 ? (
          <ResponsiveContainer width="100%" height={320}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis
                dataKey="date"
                tick={{ fill: "#71717a", fontSize: 11 }}
                tickFormatter={(v: string) => { try { return formatDate(v); } catch { return v; } }}
                interval="preserveStartEnd"
              />
              <YAxis
                tick={{ fill: "#71717a", fontSize: 11 }}
                tickFormatter={(v: number) => {
                  if (v >= 1e9) return `${(v / 1e9).toFixed(1)}B`;
                  if (v >= 1e6) return `${(v / 1e6).toFixed(1)}M`;
                  if (v >= 1e3) return `${(v / 1e3).toFixed(0)}K`;
                  return String(v);
                }}
                width={60}
              />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", borderColor: "#27272a", borderRadius: 8, color: "#fff" }}
                labelFormatter={(label) => { try { return formatDate(String(label)); } catch { return String(label); } }}
                formatter={(value, name) => [formatCurrency(Number(value ?? 0)), name === "nav" ? "Portfolio" : name === "vnindex" ? "VN-Index" : "VN30"]}
              />
              <Legend />
              <Line type="monotone" dataKey="nav" stroke="#6366f1" strokeWidth={2} dot={false} name="Portfolio" />
              <Line type="monotone" dataKey="vnindex" stroke="#22c55e" strokeWidth={1.5} dot={false} name="VN-Index" strokeDasharray="4 4" />
              <Line type="monotone" dataKey="vn30" stroke="#f59e0b" strokeWidth={1.5} dot={false} name="VN30" strokeDasharray="4 4" />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex items-center justify-center h-64 text-zinc-500 text-sm">
            {t("common.no_data")}
          </div>
        )}
      </div>
    </div>
  );
}

function MetricCard({ label, value, positive }: { label: string; value: string; positive: boolean | null }) {
  const color = positive === null ? "text-zinc-400" : positive ? "text-green-400" : "text-red-400";
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
      <p className="text-xs text-zinc-500 font-medium mb-1">{label}</p>
      <p className={`text-3xl font-bold ${color}`}>
        {positive !== null && positive ? "+" : ""}{value}
      </p>
    </div>
  );
}
