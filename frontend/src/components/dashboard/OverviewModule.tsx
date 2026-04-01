"use client";

import { useEffect, useState, useCallback } from "react";
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip as RechartsTooltip,
  Legend,
} from "recharts";
import {
  Wallet,
  TrendingUp,
  TrendingDown,
  CreditCard,
  Plus,
  Filter as FilterIcon,
  Bell,
  ArrowUpRight,
} from "lucide-react";
import { useApp } from "@/context/AppContext";
import { useI18n } from "@/context/I18nContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { getFreshnessStatus } from "@/hooks/usePricePolling";
import { FreshnessIndicator } from "@/components/ui/FreshnessIndicator";
import { apiFetch } from "@/lib/api";

// Color palette for stock portfolio
const ASSET_COLORS: Record<string, string> = {
  vn_stock: "#6366f1",
  cash: "#64748b",
};

const ASSET_LABELS: Record<string, string> = {
  vn_stock: "Stocks (VN)",
  cash: "Cash",
};

// --- Types matching backend models ---

interface PortfolioSummary {
  nav: number;
  navChange24h: number;
  navChangePercent: number;
  allocationByType: Record<string, number>;
  allocationPercent: Record<string, number>;
  holdings: HoldingDetail[];
}

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

interface Alert {
  id: number;
  symbol: string;
  message: string;
  patternType: string;
  confidenceScore: number;
  createdAt: string;
  viewed: boolean;
}

// --- Skeleton components ---

function SkeletonBlock({ className = "" }: { className?: string }) {
  return (
    <div
      className={`animate-pulse bg-zinc-800 rounded-lg ${className}`}
    />
  );
}

function NavSkeleton() {
  return (
    <div className="bg-gradient-to-br from-indigo-900/40 to-black border border-indigo-500/20 rounded-2xl p-8">
      <SkeletonBlock className="h-4 w-40 mb-3" />
      <SkeletonBlock className="h-14 w-72 mb-3" />
      <SkeletonBlock className="h-6 w-48" />
    </div>
  );
}

// --- Main Component ---

export function OverviewModule() {
  const { setActiveTab, setActiveSymbol } = useApp();
  const { formatCurrency, formatPercent, formatDate, t } = useI18n();

  // Polled data source with trading-hours-aware interval
  const tradingHours = isVNTradingHours();
  const summaryInterval = tradingHours ? 15_000 : 300_000;
  const summaryPoll = usePolling<PortfolioSummary>(
    () => apiFetch<PortfolioSummary>("/api/portfolio/summary"),
    summaryInterval,
  );

  // Non-polled data (fetched once on mount)
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [initialLoading, setInitialLoading] = useState(true);

  const fetchOnce = useCallback(async () => {
    const [txData, alertData] = await Promise.all([
      apiFetch<Transaction[]>("/api/portfolio/transactions"),
      apiFetch<Alert[]>("/api/alerts"),
    ]);
    setTransactions(txData ?? []);
    setAlerts((alertData ?? []).filter((a) => !a.viewed));
    setInitialLoading(false);
  }, []);

  useEffect(() => {
    fetchOnce();
  }, [fetchOnce]);

  const summary = summaryPoll.data;
  const loading = initialLoading && summaryPoll.loading;
  const error = summaryPoll.error;
  const freshness = getFreshnessStatus(summaryPoll.lastUpdated);

  const handleAlertClick = (alert: Alert) => {
    setActiveSymbol(alert.symbol);
    setActiveTab("Markets");
  };

  if (loading) {
    return (
      <div className="flex flex-col gap-8 w-full">
        <NavSkeleton />
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <div className="lg:col-span-2">
            <SkeletonBlock className="h-80" />
          </div>
          <div className="space-y-4">
            <SkeletonBlock className="h-52" />
          </div>
        </div>
      </div>
    );
  }

  if (error && !summary) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
        <p className="text-zinc-400 mb-4">{error}</p>
        <button
          onClick={fetchOnce}
          className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition"
        >
          Retry
        </button>
      </div>
    );
  }

  // Prepare pie chart data from real allocation
  const allocationData = summary
    ? Object.entries(summary.allocationByType).map(([type, value]) => ({
        name: ASSET_LABELS[type] ?? type,
        value,
        color: ASSET_COLORS[type] ?? "#64748b",
        percent: summary.allocationPercent[type] ?? 0,
      }))
    : [];

  const navChange = summary?.navChange24h ?? 0;
  const navChangePct = summary?.navChangePercent ?? 0;
  const isPositive = navChange >= 0;

  // 5 most recent transactions
  const recentTx = [...transactions]
    .sort(
      (a, b) =>
        new Date(b.transactionDate).getTime() -
        new Date(a.transactionDate).getTime(),
    )
    .slice(0, 5);

  return (
    <div className="flex flex-col gap-8 w-full">
      {/* Master NAV Header */}
      <div className="bg-gradient-to-br from-indigo-900/40 to-black border border-indigo-500/20 rounded-2xl p-8 relative overflow-hidden shadow-2xl">
        <div className="absolute top-0 right-0 w-64 h-64 bg-indigo-500/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/3" />
        <div className="relative z-10 flex flex-col md:flex-row justify-between items-start md:items-end gap-6">
          <div>
            <p className="text-zinc-400 font-medium mb-1 flex items-center gap-2">
              <Wallet size={16} /> {t("nav.totalNav") || "Total Net Asset Value"}
              <FreshnessIndicator lastUpdated={summaryPoll.lastUpdated} isStale={summaryPoll.isStale} error={summaryPoll.error} />
            </p>
            <h1 className="text-4xl md:text-6xl font-black text-white tracking-tight drop-shadow-lg">
              {formatCurrency(summary?.nav ?? 0)}
            </h1>
            <div className="mt-3 flex items-center gap-2">
              <span
                className={`${
                  isPositive
                    ? "bg-green-500/20 text-green-400"
                    : "bg-red-500/20 text-red-400"
                } px-3 py-1 rounded-full text-sm font-semibold flex items-center gap-1`}
              >
                {isPositive ? (
                  <TrendingUp size={14} />
                ) : (
                  <TrendingDown size={14} />
                )}
                {isPositive ? "+" : ""}
                {formatCurrency(navChange)} ({formatPercent(navChangePct)})
              </span>
              <span className="text-zinc-500 text-sm">
                {t("nav.past24h") || "Past 24 hours"}
              </span>
            </div>
          </div>
          <div className="flex gap-3">
            <button className="flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-xl transition shadow-lg shadow-indigo-600/20">
              <Plus size={18} /> {t("nav.addAsset") || "Add Stock"}
            </button>
            <button
              onClick={() => setActiveTab("Filter")}
              className="flex items-center gap-2 px-5 py-2.5 bg-zinc-800 hover:bg-zinc-700 text-white font-medium rounded-xl transition border border-zinc-700"
            >
              <FilterIcon size={18} /> {t("nav.filterStocks") || "Filter Stocks"}
            </button>
          </div>
        </div>
      </div>

      {/* Stale data warning */}
      {freshness === "expired" && (
        <div className="flex items-center gap-2 px-4 py-2 bg-red-500/10 border border-red-500/30 rounded-lg text-red-400 text-sm">
          <span className="inline-block w-2 h-2 rounded-full bg-red-500" />
          {t("nav.staleWarning") || "Price data may be outdated. Displaying last known values."}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Asset Allocation Pie Chart */}
        <div className="lg:col-span-2 bg-zinc-900 border border-zinc-800 rounded-xl p-6 shadow-lg">
          <h2 className="text-xl font-bold text-white flex items-center gap-2 mb-6">
            <PieChartIcon size={20} className="text-indigo-400" />
            {t("nav.assetAllocation") || "Portfolio Allocation"}
          </h2>
          {allocationData.length > 0 ? (
            <div className="flex flex-col md:flex-row items-center justify-between">
              <div className="w-full h-64 md:h-80 md:w-1/2 relative">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={allocationData}
                      innerRadius={60}
                      outerRadius={90}
                      paddingAngle={5}
                      dataKey="value"
                      stroke="none"
                    >
                      {allocationData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <RechartsTooltip
                      formatter={(value) =>
                        formatCurrency(Number(value ?? 0))
                      }
                      contentStyle={{
                        backgroundColor: "#18181b",
                        borderColor: "#27272a",
                        borderRadius: "8px",
                        color: "#fff",
                      }}
                      itemStyle={{ color: "#fff" }}
                    />
                    <Legend verticalAlign="bottom" height={36} />
                  </PieChart>
                </ResponsiveContainer>
                <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none mt-[-20px]">
                  <span className="text-zinc-500 text-xs font-medium">
                    HOLDINGS
                  </span>
                  <span className="text-white font-bold text-lg">
                    {allocationData.length}
                  </span>
                </div>
              </div>

              <div className="w-full md:w-1/2 mt-6 md:mt-0 space-y-4 pl-0 md:pl-8 border-t md:border-t-0 md:border-l border-zinc-800 pt-6 md:pt-0">
                {allocationData.map((asset) => (
                  <div
                    key={asset.name}
                    className="flex justify-between items-center group"
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: asset.color }}
                      />
                      <div>
                        <p className="font-semibold text-zinc-200 group-hover:text-white transition">
                          {asset.name}
                        </p>
                        <p className="text-xs text-zinc-500">
                          {formatPercent(asset.percent)} of Portfolio
                        </p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-medium text-white">
                        {formatCurrency(asset.value)}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <div className="flex items-center justify-center h-48 text-zinc-500">
              No stocks in portfolio yet
            </div>
          )}
        </div>

        {/* Right column: Recent Activity + Alerts */}
        <div className="lg:col-span-1 space-y-6">
          {/* Recent Transactions */}
          <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 shadow-lg">
            <h2 className="text-lg font-bold text-white flex items-center gap-2 mb-4">
              <CreditCard size={18} className="text-purple-400" />
              {t("nav.recentActivity") || "Recent Activity"}
            </h2>
            {recentTx.length > 0 ? (
              <div className="space-y-3">
                {recentTx.map((tx) => {
                  const isBuyOrDeposit =
                    tx.transactionType === "buy" ||
                    tx.transactionType === "dividend";
                  return (
                    <div
                      key={tx.id}
                      className="flex justify-between items-center bg-black/40 p-3 rounded-lg border border-zinc-800/50"
                    >
                      <div className="flex flex-col">
                        <span className="font-semibold text-sm text-zinc-200 capitalize">
                          {tx.transactionType} {tx.symbol}
                        </span>
                        <span className="text-[10px] text-zinc-500">
                          {formatDate(tx.transactionDate)}
                        </span>
                      </div>
                      <div className="text-right flex flex-col">
                        <span
                          className={`font-bold text-sm ${
                            isBuyOrDeposit
                              ? "text-green-400"
                              : "text-red-400"
                          }`}
                        >
                          {isBuyOrDeposit ? "+" : "-"}
                          {formatCurrency(tx.totalValue)}
                        </span>
                        <span className="text-[10px] text-zinc-500">
                          {tx.quantity} × {formatCurrency(tx.unitPrice)}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p className="text-zinc-500 text-sm text-center py-4">
                No transactions yet
              </p>
            )}
          </div>

          {/* Notification Panel */}
          {alerts.length > 0 && (
            <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 shadow-lg">
              <h2 className="text-lg font-bold text-white flex items-center gap-2 mb-4">
                <Bell size={18} className="text-amber-400" />
                {t("nav.alerts") || "Alerts"}
                <span className="ml-auto bg-amber-500/20 text-amber-400 text-xs font-bold px-2 py-0.5 rounded-full">
                  {alerts.length}
                </span>
              </h2>
              <div className="space-y-2">
                {alerts.slice(0, 5).map((alert) => (
                  <button
                    key={alert.id}
                    onClick={() => handleAlertClick(alert)}
                    className="w-full text-left flex items-start gap-3 p-3 rounded-lg bg-black/40 border border-zinc-800/50 hover:border-amber-500/30 transition group"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-zinc-200 group-hover:text-white truncate">
                        {alert.symbol} — {alert.message}
                      </p>
                      <p className="text-[10px] text-zinc-500 mt-0.5">
                        {alert.patternType} · {formatDate(alert.createdAt)}
                      </p>
                    </div>
                    <ArrowUpRight
                      size={14}
                      className="text-zinc-600 group-hover:text-amber-400 mt-1 shrink-0"
                    />
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// --- Sub-components ---

function PieChartIcon(props: React.SVGProps<SVGSVGElement> & { size?: number }) {
  const { size = 24, ...rest } = props;
  return (
    <svg
      {...rest}
      xmlns="http://www.w3.org/2000/svg"
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M21.21 15.89A10 10 0 1 1 8 2.83" />
      <path d="M22 12A10 10 0 0 0 12 2v10z" />
    </svg>
  );
}
