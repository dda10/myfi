"use client";

import { useEffect, useState, useCallback } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
} from "recharts";
import {
  Briefcase,
  ArrowUpRight,
  ArrowDownRight,
  History,
  TrendingUp,
  ShieldAlert,
  Download,
  FileText,
  ChevronDown,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types ---

interface HoldingDetail {
  asset: {
    id: number;
    assetType: string;
    symbol: string;
    quantity: number;
    averageCost: number;
  };
  currentPrice: number;
  marketValue: number;
  unrealizedPL: number;
  unrealizedPLPct: number;
}

interface PortfolioSummary {
  nav: number;
  navChange24h: number;
  navChangePercent: number;
  allocationByType: Record<string, number>;
  allocationPercent: Record<string, number>;
  holdings: HoldingDetail[];
}

interface Transaction {
  id: number;
  assetType: string;
  symbol: string;
  quantity: number;
  unitPrice: number;
  totalValue: number;
  transactionDate: string;
  transactionType: string;
}

interface NAVSnapshot {
  date: string;
  nav: number;
}

interface PerformanceMetrics {
  twr: number;
  mwrr: number;
  equityCurve: NAVSnapshot[];
}

interface RiskMetrics {
  sharpeRatio: number;
  maxDrawdown: number;
  beta: number;
  volatility: number;
  var95: number;
}

type SubTab = "holdings" | "transactions" | "performance" | "risk";
type TimePeriod = "1W" | "1M" | "3M" | "6M" | "1Y" | "YTD" | "ALL";
type TxFilter = "all" | "buy" | "sell" | "dividend" | "deposit" | "withdrawal" | "interest";

const SUB_TABS: { key: SubTab; label: string; icon: React.ReactNode }[] = [
  { key: "holdings", label: "Holdings", icon: <Briefcase size={16} /> },
  { key: "transactions", label: "Transactions", icon: <History size={16} /> },
  { key: "performance", label: "Performance", icon: <TrendingUp size={16} /> },
  { key: "risk", label: "Risk", icon: <ShieldAlert size={16} /> },
];

const TIME_PERIODS: TimePeriod[] = ["1W", "1M", "3M", "6M", "1Y", "YTD", "ALL"];
const TX_FILTERS: TxFilter[] = ["all", "buy", "sell", "dividend", "deposit", "withdrawal", "interest"];

// --- Fetch helper ---

async function apiFetch<T>(path: string): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, { credentials: "include" });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

function triggerDownload(path: string) {
  const link = document.createElement("a");
  link.href = `${API_URL}${path}`;
  link.download = "";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

// --- Main Component ---

export function PortfolioModule() {
  const { formatCurrency, formatPercent, formatDate, t } = useI18n();

  const [activeSubTab, setActiveSubTab] = useState<SubTab>("holdings");
  const [period, setPeriod] = useState<TimePeriod>("1Y");
  const [txFilter, setTxFilter] = useState<TxFilter>("all");

  const [summary, setSummary] = useState<PortfolioSummary | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [performance, setPerformance] = useState<PerformanceMetrics | null>(null);
  const [risk, setRisk] = useState<RiskMetrics | null>(null);

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [sumData, txData] = await Promise.all([
        apiFetch<PortfolioSummary>("/api/portfolio/summary"),
        apiFetch<Transaction[]>("/api/portfolio/transactions"),
      ]);
      setSummary(sumData);
      setTransactions(txData ?? []);
    } catch {
      setError("Failed to load portfolio data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Fetch performance when tab or period changes
  useEffect(() => {
    if (activeSubTab !== "performance") return;
    let cancelled = false;
    (async () => {
      const data = await apiFetch<PerformanceMetrics>(`/api/portfolio/performance?period=${period}`);
      if (!cancelled) setPerformance(data);
    })();
    return () => { cancelled = true; };
  }, [activeSubTab, period]);

  // Fetch risk when tab changes
  useEffect(() => {
    if (activeSubTab !== "risk") return;
    let cancelled = false;
    (async () => {
      const data = await apiFetch<RiskMetrics>("/api/portfolio/risk");
      if (!cancelled) setRisk(data);
    })();
    return () => { cancelled = true; };
  }, [activeSubTab]);

  if (loading) {
    return (
      <div className="flex flex-col gap-6 w-full">
        <div className="animate-pulse bg-zinc-800 rounded-xl h-16" />
        <div className="animate-pulse bg-zinc-800 rounded-xl h-96" />
      </div>
    );
  }

  if (error && !summary) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
        <p className="text-zinc-400 mb-4">{error}</p>
        <button onClick={fetchData} className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition">
          Retry
        </button>
      </div>
    );
  }

  const filteredTx = txFilter === "all"
    ? transactions
    : transactions.filter((tx) => tx.transactionType === txFilter);

  const sortedTx = [...filteredTx].sort(
    (a, b) => new Date(b.transactionDate).getTime() - new Date(a.transactionDate).getTime(),
  );

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Sub-tab navigation + Export buttons */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div className="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-xl p-1">
          {SUB_TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveSubTab(tab.key)}
              className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition ${
                activeSubTab === tab.key
                  ? "bg-indigo-600 text-white shadow"
                  : "text-zinc-400 hover:text-white hover:bg-zinc-800"
              }`}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => triggerDownload("/api/export/transactions?format=csv")}
            className="flex items-center gap-1.5 px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm rounded-lg border border-zinc-700 transition"
          >
            <Download size={14} /> CSV
          </button>
          <button
            onClick={() => triggerDownload("/api/export/report?format=pdf")}
            className="flex items-center gap-1.5 px-3 py-2 bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-sm rounded-lg border border-zinc-700 transition"
          >
            <FileText size={14} /> PDF
          </button>
        </div>
      </div>

      {/* Content */}
      {activeSubTab === "holdings" && <HoldingsSection holdings={summary?.holdings ?? []} formatCurrency={formatCurrency} formatPercent={formatPercent} />}
      {activeSubTab === "transactions" && (
        <TransactionsSection
          transactions={sortedTx}
          txFilter={txFilter}
          setTxFilter={setTxFilter}
          formatCurrency={formatCurrency}
          formatDate={formatDate}
        />
      )}
      {activeSubTab === "performance" && (
        <PerformanceSection
          performance={performance}
          period={period}
          setPeriod={setPeriod}
          formatPercent={formatPercent}
          formatCurrency={formatCurrency}
          formatDate={formatDate}
        />
      )}
      {activeSubTab === "risk" && <RiskSection risk={risk} formatPercent={formatPercent} />}
    </div>
  );
}

// --- Holdings Section ---

function HoldingsSection({
  holdings,
  formatCurrency,
  formatPercent,
}: {
  holdings: HoldingDetail[];
  formatCurrency: (v: number, c?: string) => string;
  formatPercent: (v: number, d?: number) => string;
}) {
  if (holdings.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-zinc-900 border border-zinc-800 rounded-xl text-zinc-500">
        No holdings in portfolio
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg">
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-left">
              <th className="px-4 py-3 font-medium">Symbol</th>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium text-right">Qty</th>
              <th className="px-4 py-3 font-medium text-right">Avg Cost</th>
              <th className="px-4 py-3 font-medium text-right">Price</th>
              <th className="px-4 py-3 font-medium text-right">Market Value</th>
              <th className="px-4 py-3 font-medium text-right">P&L</th>
              <th className="px-4 py-3 font-medium text-right">P&L %</th>
            </tr>
          </thead>
          <tbody>
            {holdings.map((h) => {
              const isPositive = h.unrealizedPL >= 0;
              const plColor = isPositive ? "text-green-400" : "text-red-400";
              return (
                <tr key={h.asset.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition">
                  <td className="px-4 py-3 font-semibold text-white">{h.asset.symbol}</td>
                  <td className="px-4 py-3 text-zinc-400 capitalize">{h.asset.assetType.replace("_", " ")}</td>
                  <td className="px-4 py-3 text-right text-zinc-300">{h.asset.quantity.toLocaleString()}</td>
                  <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(h.asset.averageCost)}</td>
                  <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(h.currentPrice)}</td>
                  <td className="px-4 py-3 text-right text-white font-medium">{formatCurrency(h.marketValue)}</td>
                  <td className={`px-4 py-3 text-right font-medium ${plColor}`}>
                    {isPositive ? "+" : ""}{formatCurrency(h.unrealizedPL)}
                  </td>
                  <td className={`px-4 py-3 text-right font-medium ${plColor}`}>
                    <span className="inline-flex items-center gap-0.5">
                      {isPositive ? <ArrowUpRight size={12} /> : <ArrowDownRight size={12} />}
                      {formatPercent(h.unrealizedPLPct)}
                    </span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// --- Transactions Section ---

function TransactionsSection({
  transactions,
  txFilter,
  setTxFilter,
  formatCurrency,
  formatDate,
}: {
  transactions: Transaction[];
  txFilter: TxFilter;
  setTxFilter: (f: TxFilter) => void;
  formatCurrency: (v: number, c?: string) => string;
  formatDate: (d: Date | string | number) => string;
}) {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg">
      {/* Filter bar */}
      <div className="px-4 py-3 border-b border-zinc-800 flex items-center gap-2 flex-wrap">
        <span className="text-zinc-400 text-sm mr-1">Filter:</span>
        {TX_FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setTxFilter(f)}
            className={`px-3 py-1 rounded-md text-xs font-medium capitalize transition ${
              txFilter === f
                ? "bg-indigo-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:text-white hover:bg-zinc-700"
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      {transactions.length === 0 ? (
        <div className="flex items-center justify-center h-48 text-zinc-500 text-sm">
          No transactions found
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-400 text-left">
                <th className="px-4 py-3 font-medium">Date</th>
                <th className="px-4 py-3 font-medium">Type</th>
                <th className="px-4 py-3 font-medium">Symbol</th>
                <th className="px-4 py-3 font-medium text-right">Qty</th>
                <th className="px-4 py-3 font-medium text-right">Price</th>
                <th className="px-4 py-3 font-medium text-right">Total</th>
              </tr>
            </thead>
            <tbody>
              {transactions.map((tx) => {
                const isBuy = tx.transactionType === "buy" || tx.transactionType === "deposit" || tx.transactionType === "interest" || tx.transactionType === "dividend";
                return (
                  <tr key={tx.id} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition">
                    <td className="px-4 py-3 text-zinc-300">{formatDate(tx.transactionDate)}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-block px-2 py-0.5 rounded text-xs font-semibold capitalize ${
                        isBuy ? "bg-green-500/10 text-green-400" : "bg-red-500/10 text-red-400"
                      }`}>
                        {tx.transactionType}
                      </span>
                    </td>
                    <td className="px-4 py-3 font-semibold text-white">{tx.symbol}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{tx.quantity.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(tx.unitPrice)}</td>
                    <td className="px-4 py-3 text-right text-white font-medium">{formatCurrency(tx.totalValue)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

// --- Performance Section ---

function PerformanceSection({
  performance,
  period,
  setPeriod,
  formatPercent,
  formatCurrency,
  formatDate,
}: {
  performance: PerformanceMetrics | null;
  period: TimePeriod;
  setPeriod: (p: TimePeriod) => void;
  formatPercent: (v: number, d?: number) => string;
  formatCurrency: (v: number, c?: string) => string;
  formatDate: (d: Date | string | number) => string;
}) {
  return (
    <div className="flex flex-col gap-6">
      {/* Period selector */}
      <div className="flex items-center gap-2">
        <span className="text-zinc-400 text-sm mr-1">Period:</span>
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

      {/* TWR / MWRR cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <MetricCard
          label="Time-Weighted Return (TWR)"
          value={performance ? formatPercent(performance.twr) : "—"}
          positive={performance ? performance.twr >= 0 : null}
        />
        <MetricCard
          label="Money-Weighted Return (MWRR)"
          value={performance ? formatPercent(performance.mwrr) : "—"}
          positive={performance ? performance.mwrr >= 0 : null}
        />
      </div>

      {/* Equity curve chart */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 shadow-lg">
        <h3 className="text-lg font-bold text-white mb-4">Equity Curve</h3>
        {performance && performance.equityCurve && performance.equityCurve.length > 0 ? (
          <ResponsiveContainer width="100%" height={320}>
            <LineChart data={performance.equityCurve}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis
                dataKey="date"
                tick={{ fill: "#71717a", fontSize: 11 }}
                tickFormatter={(v: string) => {
                  try { return formatDate(v); } catch { return v; }
                }}
                interval="preserveStartEnd"
              />
              <YAxis
                tick={{ fill: "#71717a", fontSize: 11 }}
                tickFormatter={(v: number) => {
                  if (v >= 1_000_000_000) return `${(v / 1_000_000_000).toFixed(1)}B`;
                  if (v >= 1_000_000) return `${(v / 1_000_000).toFixed(1)}M`;
                  if (v >= 1_000) return `${(v / 1_000).toFixed(0)}K`;
                  return String(v);
                }}
                width={60}
              />
              <RechartsTooltip
                contentStyle={{ backgroundColor: "#18181b", borderColor: "#27272a", borderRadius: "8px", color: "#fff" }}
                labelFormatter={(label) => { try { return formatDate(String(label)); } catch { return String(label); } }}
                formatter={(value) => [formatCurrency(Number(value ?? 0)), "NAV"]}
              />
              <Line type="monotone" dataKey="nav" stroke="#6366f1" strokeWidth={2} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        ) : (
          <div className="flex items-center justify-center h-64 text-zinc-500 text-sm">
            No equity curve data available
          </div>
        )}
      </div>
    </div>
  );
}

// --- Risk Section ---

function RiskSection({
  risk,
  formatPercent,
}: {
  risk: RiskMetrics | null;
  formatPercent: (v: number, d?: number) => string;
}) {
  const metrics = [
    { label: "Sharpe Ratio", value: risk ? risk.sharpeRatio.toFixed(2) : "—", desc: "Risk-adjusted return" },
    { label: "Max Drawdown", value: risk ? formatPercent(risk.maxDrawdown) : "—", desc: "Largest peak-to-trough decline" },
    { label: "Beta", value: risk ? risk.beta.toFixed(2) : "—", desc: "Sensitivity to VN-Index" },
    { label: "Volatility", value: risk ? formatPercent(risk.volatility) : "—", desc: "Annualized std deviation" },
    { label: "VaR (95%)", value: risk ? formatPercent(risk.var95) : "—", desc: "Max expected daily loss at 95% confidence" },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {metrics.map((m) => (
        <div key={m.label} className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 shadow-lg">
          <p className="text-xs text-zinc-500 font-medium mb-1">{m.label}</p>
          <p className="text-2xl font-bold text-white mb-1">{m.value}</p>
          <p className="text-xs text-zinc-500">{m.desc}</p>
        </div>
      ))}
    </div>
  );
}

// --- Shared MetricCard ---

function MetricCard({
  label,
  value,
  positive,
}: {
  label: string;
  value: string;
  positive: boolean | null;
}) {
  const color =
    positive === null
      ? "text-zinc-400"
      : positive
        ? "text-green-400"
        : "text-red-400";

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-5 shadow-lg">
      <p className="text-xs text-zinc-500 font-medium mb-1">{label}</p>
      <p className={`text-3xl font-bold ${color}`}>
        {positive !== null && positive ? "+" : ""}
        {value}
      </p>
    </div>
  );
}
