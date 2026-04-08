"use client";

import { useState, useCallback, useEffect } from "react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import {
  BarChart3,
  TrendingUp,
  Gem,
  Zap,
  Activity,
  Loader2,
  AlertCircle,
} from "lucide-react";
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

// --- Types ---

interface FactorGroup {
  key: string;
  i18nKey: string;
  icon: React.ReactNode;
  factors: string[];
  weight: number;
}

interface RankedStock {
  rank: number;
  symbol: string;
  name: string;
  consensusScore: number;
  qualityScore: number;
  valueScore: number;
  growthScore: number;
  momentumScore: number;
  volatilityScore: number;
}

interface BacktestPoint {
  date: string;
  strategy: number;
  benchmark: number;
}

interface PeriodPerformance {
  period: string;
  strategyReturn: number;
  benchmarkReturn: number;
  alpha: number;
}

interface TopHolding {
  period: string;
  holdings: { symbol: string; weight: number }[];
}

interface BacktestResult {
  cumulativeReturns: BacktestPoint[];
  monthlyPerformance: PeriodPerformance[];
  yearlyPerformance: PeriodPerformance[];
  topHoldings: TopHolding[];
}

type Universe = "VN30" | "VN100" | "HOSE" | "HNX" | "UPCOM" | "Custom";
type TabView = "ranking" | "backtest";

const UNIVERSES: Universe[] = ["VN30", "VN100", "HOSE", "HNX", "UPCOM", "Custom"];

// --- Component ---

export default function RankingPage() {
  const { t, formatNumber, formatPercent } = useI18n();

  const [factorGroups, setFactorGroups] = useState<FactorGroup[]>([
    {
      key: "quality",
      i18nKey: "ranking.quality",
      icon: <Gem className="w-4 h-4" />,
      factors: ["ROE", "ROA", "Profit Margin"],
      weight: 20,
    },
    {
      key: "value",
      i18nKey: "ranking.value",
      icon: <BarChart3 className="w-4 h-4" />,
      factors: ["P/E", "P/B", "EV/EBITDA"],
      weight: 20,
    },
    {
      key: "growth",
      i18nKey: "ranking.growth",
      icon: <TrendingUp className="w-4 h-4" />,
      factors: ["Revenue Growth", "Profit Growth"],
      weight: 20,
    },
    {
      key: "momentum",
      i18nKey: "ranking.momentum",
      icon: <Zap className="w-4 h-4" />,
      factors: ["1M", "3M", "6M", "12M"],
      weight: 20,
    },
    {
      key: "volatility",
      i18nKey: "ranking.volatility_factor",
      icon: <Activity className="w-4 h-4" />,
      factors: ["Std Dev", "Beta"],
      weight: 20,
    },
  ]);

  const [universe, setUniverse] = useState<Universe>("VN30");
  const [activeTab, setActiveTab] = useState<TabView>("ranking");
  const [rankedStocks, setRankedStocks] = useState<RankedStock[]>([]);
  const [backtestResult, setBacktestResult] = useState<BacktestResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const updateWeight = useCallback((key: string, value: number) => {
    setFactorGroups((prev) =>
      prev.map((g) => (g.key === key ? { ...g, weight: value } : g)),
    );
  }, []);

  const buildConfig = useCallback(() => {
    const weights: Record<string, number> = {};
    factorGroups.forEach((g) => {
      weights[g.key] = g.weight;
    });
    return { universe, weights };
  }, [factorGroups, universe]);

  const fetchRanking = useCallback(async () => {
    setLoading(true);
    setError(null);
    const data = await apiFetch<{ stocks: RankedStock[] }>("/api/ranking", {
      method: "POST",
      body: JSON.stringify(buildConfig()),
    });
    if (data?.stocks) {
      setRankedStocks(data.stocks);
    } else {
      setError(t("error.generic"));
    }
    setLoading(false);
  }, [buildConfig, t]);

  const fetchBacktest = useCallback(async () => {
    setLoading(true);
    setError(null);
    const data = await apiFetch<BacktestResult>("/api/ranking/backtest", {
      method: "POST",
      body: JSON.stringify(buildConfig()),
    });
    if (data) {
      setBacktestResult(data);
    } else {
      setError(t("error.generic"));
    }
    setLoading(false);
  }, [buildConfig, t]);

  // Fetch on mount and when config changes
  useEffect(() => {
    if (activeTab === "ranking") {
      fetchRanking();
    } else {
      fetchBacktest();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab]);

  const handleApply = () => {
    if (activeTab === "ranking") fetchRanking();
    else fetchBacktest();
  };

  return (
    <div className="space-y-6 pb-8">
      {/* Header */}
      <h1 className="text-2xl font-bold text-foreground">{t("ranking.title")}</h1>

      {/* Factor Groups Configuration */}
      <div className="rounded-xl bg-surface border border-border-theme p-5 space-y-4">
        <h2 className="text-sm font-semibold text-foreground">{t("ranking.factor_groups")}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          {factorGroups.map((group) => (
            <div key={group.key} className="space-y-2">
              <div className="flex items-center gap-2 text-sm text-foreground">
                {group.icon}
                <span className="font-medium">{t(group.i18nKey)}</span>
              </div>
              <div className="text-xs text-text-muted">{group.factors.join(", ")}</div>
              <div className="flex items-center gap-2">
                <input
                  type="range"
                  min={0}
                  max={100}
                  value={group.weight}
                  onChange={(e) => updateWeight(group.key, Number(e.target.value))}
                  className="flex-1 h-1.5 accent-accent cursor-pointer"
                  aria-label={`${t(group.i18nKey)} weight`}
                />
                <span className="text-xs font-mono text-text-muted w-8 text-right">
                  {group.weight}
                </span>
              </div>
            </div>
          ))}
        </div>

        {/* Universe + Apply */}
        <div className="flex flex-wrap items-center gap-4 pt-2 border-t border-border-theme">
          <div className="flex items-center gap-2">
            <label className="text-sm text-text-muted">{t("ranking.universe")}:</label>
            <select
              value={universe}
              onChange={(e) => setUniverse(e.target.value as Universe)}
              className="rounded-lg bg-background border border-border-theme px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-accent"
            >
              {UNIVERSES.map((u) => (
                <option key={u} value={u}>{u}</option>
              ))}
            </select>
          </div>
          <button
            onClick={handleApply}
            disabled={loading}
            className="px-4 py-1.5 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors disabled:opacity-50"
          >
            {t("btn.apply")}
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border-theme">
        {(["ranking", "backtest"] as TabView[]).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-px ${
              activeTab === tab
                ? "border-accent text-accent"
                : "border-transparent text-text-muted hover:text-foreground"
            }`}
          >
            {tab === "ranking" ? t("ranking.title") : t("ranking.backtest")}
          </button>
        ))}
      </div>

      {/* Loading / Error */}
      {loading && (
        <div className="flex items-center justify-center py-12 text-text-muted">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          {t("common.loading")}
        </div>
      )}

      {error && !loading && (
        <div className="flex items-center gap-2 py-8 justify-center text-red-500 text-sm">
          <AlertCircle className="w-4 h-4" />
          {error}
        </div>
      )}

      {/* Ranking Tab */}
      {!loading && !error && activeTab === "ranking" && (
        <div className="rounded-xl bg-surface border border-border-theme overflow-x-auto">
          {rankedStocks.length === 0 ? (
            <div className="p-8 text-center text-text-muted">{t("common.no_data")}</div>
          ) : (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border-theme text-left text-text-muted">
                  <th className="px-4 py-3 font-medium">#</th>
                  <th className="px-4 py-3 font-medium">{t("table.symbol")}</th>
                  <th className="px-4 py-3 font-medium">{t("table.name")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.consensus_score")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.quality")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.value")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.growth")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.momentum")}</th>
                  <th className="px-4 py-3 font-medium text-right">{t("ranking.volatility_factor")}</th>
                </tr>
              </thead>
              <tbody>
                {rankedStocks.map((stock) => (
                  <tr
                    key={stock.symbol}
                    className="border-b border-border-theme last:border-b-0 hover:bg-surface-hover transition-colors"
                  >
                    <td className="px-4 py-3 text-text-muted">{stock.rank}</td>
                    <td className="px-4 py-3 font-semibold text-accent">{stock.symbol}</td>
                    <td className="px-4 py-3 text-foreground">{stock.name}</td>
                    <td className="px-4 py-3 text-right font-semibold text-foreground">
                      {formatNumber(stock.consensusScore, 1)}
                    </td>
                    <td className="px-4 py-3 text-right text-text-muted">{formatNumber(stock.qualityScore, 1)}</td>
                    <td className="px-4 py-3 text-right text-text-muted">{formatNumber(stock.valueScore, 1)}</td>
                    <td className="px-4 py-3 text-right text-text-muted">{formatNumber(stock.growthScore, 1)}</td>
                    <td className="px-4 py-3 text-right text-text-muted">{formatNumber(stock.momentumScore, 1)}</td>
                    <td className="px-4 py-3 text-right text-text-muted">{formatNumber(stock.volatilityScore, 1)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Backtest Tab */}
      {!loading && !error && activeTab === "backtest" && backtestResult && (
        <div className="space-y-6">
          {/* Cumulative Returns Chart */}
          <div className="rounded-xl bg-surface border border-border-theme p-5">
            <h3 className="text-sm font-semibold text-foreground mb-4">
              {t("ranking.backtest")} — Cumulative Returns
            </h3>
            <ResponsiveContainer width="100%" height={320}>
              <LineChart data={backtestResult.cumulativeReturns}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="date" tick={{ fontSize: 11, fill: "var(--text-muted)" }} />
                <YAxis
                  tickFormatter={(v: number) => formatPercent(v)}
                  tick={{ fontSize: 11, fill: "var(--text-muted)" }}
                />
                <Tooltip
                  contentStyle={{
                    backgroundColor: "var(--surface)",
                    border: "1px solid var(--border)",
                    borderRadius: 8,
                    fontSize: 12,
                  }}
                  formatter={(value) => formatPercent(Number(value))}
                />
                <Legend wrapperStyle={{ fontSize: 12 }} />
                <Line
                  type="monotone"
                  dataKey="strategy"
                  name="Strategy"
                  stroke="var(--accent)"
                  strokeWidth={2}
                  dot={false}
                />
                <Line
                  type="monotone"
                  dataKey="benchmark"
                  name="VN-Index"
                  stroke="var(--text-muted)"
                  strokeWidth={1.5}
                  strokeDasharray="4 4"
                  dot={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          {/* Performance Tables */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Monthly Performance */}
            <div className="rounded-xl bg-surface border border-border-theme overflow-x-auto">
              <div className="px-4 py-3 border-b border-border-theme">
                <h3 className="text-sm font-semibold text-foreground">Monthly Performance</h3>
              </div>
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border-theme text-left text-text-muted">
                    <th className="px-4 py-2 font-medium">Period</th>
                    <th className="px-4 py-2 font-medium text-right">Strategy</th>
                    <th className="px-4 py-2 font-medium text-right">Benchmark</th>
                    <th className="px-4 py-2 font-medium text-right">Alpha</th>
                  </tr>
                </thead>
                <tbody>
                  {backtestResult.monthlyPerformance.map((row) => (
                    <tr key={row.period} className="border-b border-border-theme last:border-b-0">
                      <td className="px-4 py-2 text-foreground">{row.period}</td>
                      <td className={`px-4 py-2 text-right ${row.strategyReturn >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.strategyReturn)}
                      </td>
                      <td className={`px-4 py-2 text-right ${row.benchmarkReturn >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.benchmarkReturn)}
                      </td>
                      <td className={`px-4 py-2 text-right font-medium ${row.alpha >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.alpha)}
                      </td>
                    </tr>
                  ))}
                  {backtestResult.monthlyPerformance.length === 0 && (
                    <tr><td colSpan={4} className="px-4 py-4 text-center text-text-muted">{t("common.no_data")}</td></tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Yearly Performance */}
            <div className="rounded-xl bg-surface border border-border-theme overflow-x-auto">
              <div className="px-4 py-3 border-b border-border-theme">
                <h3 className="text-sm font-semibold text-foreground">Yearly Performance</h3>
              </div>
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border-theme text-left text-text-muted">
                    <th className="px-4 py-2 font-medium">Period</th>
                    <th className="px-4 py-2 font-medium text-right">Strategy</th>
                    <th className="px-4 py-2 font-medium text-right">Benchmark</th>
                    <th className="px-4 py-2 font-medium text-right">Alpha</th>
                  </tr>
                </thead>
                <tbody>
                  {backtestResult.yearlyPerformance.map((row) => (
                    <tr key={row.period} className="border-b border-border-theme last:border-b-0">
                      <td className="px-4 py-2 text-foreground">{row.period}</td>
                      <td className={`px-4 py-2 text-right ${row.strategyReturn >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.strategyReturn)}
                      </td>
                      <td className={`px-4 py-2 text-right ${row.benchmarkReturn >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.benchmarkReturn)}
                      </td>
                      <td className={`px-4 py-2 text-right font-medium ${row.alpha >= 0 ? "text-green-500" : "text-red-500"}`}>
                        {formatPercent(row.alpha)}
                      </td>
                    </tr>
                  ))}
                  {backtestResult.yearlyPerformance.length === 0 && (
                    <tr><td colSpan={4} className="px-4 py-4 text-center text-text-muted">{t("common.no_data")}</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {/* Top Holdings per Rebalancing Period */}
          {backtestResult.topHoldings.length > 0 && (
            <div className="rounded-xl bg-surface border border-border-theme p-5 space-y-3">
              <h3 className="text-sm font-semibold text-foreground">Top Holdings per Rebalancing Period</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {backtestResult.topHoldings.map((period) => (
                  <div key={period.period} className="rounded-lg bg-background border border-border-theme p-3">
                    <div className="text-xs font-semibold text-text-muted mb-2">{period.period}</div>
                    <div className="space-y-1">
                      {period.holdings.map((h) => (
                        <div key={h.symbol} className="flex justify-between text-sm">
                          <span className="text-accent font-medium">{h.symbol}</span>
                          <span className="text-text-muted">{formatPercent(h.weight)}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {!loading && !error && activeTab === "backtest" && !backtestResult && (
        <div className="p-8 text-center text-text-muted">{t("common.no_data")}</div>
      )}
    </div>
  );
}
