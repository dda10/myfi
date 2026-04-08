"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Briefcase,
  History,
  TrendingUp,
  ShieldAlert,
  Download,
  FileText,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import { Holdings, type HoldingRow } from "@/features/portfolio/components/Holdings";
import { Transactions, type TransactionRow } from "@/features/portfolio/components/Transactions";
import { PerformanceChart } from "@/features/portfolio/components/PerformanceChart";
import { RiskMetrics } from "@/features/portfolio/components/RiskMetrics";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type SubTab = "holdings" | "transactions" | "performance" | "risk";

const SUB_TABS: { key: SubTab; label: string; icon: React.ReactNode }[] = [
  { key: "holdings", label: "portfolio.holdings", icon: <Briefcase size={16} /> },
  { key: "transactions", label: "portfolio.transactions", icon: <History size={16} /> },
  { key: "performance", label: "portfolio.performance", icon: <TrendingUp size={16} /> },
  { key: "risk", label: "portfolio.risk", icon: <ShieldAlert size={16} /> },
];

interface PortfolioSummary {
  holdings: HoldingRow[];
}

function triggerDownload(path: string) {
  const link = document.createElement("a");
  link.href = `${API_URL}${path}`;
  link.download = "";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

export default function PortfolioPage() {
  const { t } = useI18n();
  const [activeTab, setActiveTab] = useState<SubTab>("holdings");
  const [holdings, setHoldings] = useState<HoldingRow[]>([]);
  const [transactions, setTransactions] = useState<TransactionRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [sumData, txData] = await Promise.all([
        apiFetch<PortfolioSummary>("/api/portfolio/summary"),
        apiFetch<TransactionRow[]>("/api/portfolio/transactions"),
      ]);
      setHoldings(sumData?.holdings ?? []);
      setTransactions(txData ?? []);
    } catch {
      setError(t("error.generic"));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="flex flex-col gap-6 w-full">
        <div className="animate-pulse bg-zinc-800 rounded-xl h-16" />
        <div className="animate-pulse bg-zinc-800 rounded-xl h-96" />
      </div>
    );
  }

  if (error && holdings.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-96 bg-zinc-900 border border-zinc-800 rounded-xl">
        <p className="text-zinc-400 mb-4">{error}</p>
        <button onClick={fetchData} className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition">
          {t("btn.refresh")}
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Header */}
      <h1 className="text-xl font-bold text-white">{t("portfolio.title")}</h1>

      {/* Tab navigation + Export */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div className="flex gap-1 bg-zinc-900 border border-zinc-800 rounded-xl p-1">
          {SUB_TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition ${
                activeTab === tab.key
                  ? "bg-indigo-600 text-white shadow"
                  : "text-zinc-400 hover:text-white hover:bg-zinc-800"
              }`}
            >
              {tab.icon}
              {t(tab.label)}
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
      {activeTab === "holdings" && <Holdings holdings={holdings} />}
      {activeTab === "transactions" && <Transactions transactions={transactions} />}
      {activeTab === "performance" && <PerformanceChart />}
      {activeTab === "risk" && <RiskMetrics />}
    </div>
  );
}
