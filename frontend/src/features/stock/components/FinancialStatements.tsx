"use client";

import { useState, useEffect } from "react";
import { FileSpreadsheet } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

type StmtType = "income" | "balance" | "cashflow";
type Period = "quarter" | "year";

interface FinancialRow {
  label: string;
  values: number[];
}

interface FinancialData {
  periods: string[];
  rows: FinancialRow[];
}

const STMT_TABS: { key: StmtType; label: string }[] = [
  { key: "income", label: "Income Statement" },
  { key: "balance", label: "Balance Sheet" },
  { key: "cashflow", label: "Cash Flow" },
];

export function FinancialStatements({ symbol }: { symbol: string }) {
  const { t, formatNumber } = useI18n();
  const [stmtType, setStmtType] = useState<StmtType>("income");
  const [period, setPeriod] = useState<Period>("quarter");
  const [data, setData] = useState<FinancialData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<FinancialData>(`/api/market/finance/${symbol}?type=${stmtType}&period=${period}`)
      .then((res) => { if (res) setData(res); else setData(null); })
      .finally(() => setLoading(false));
  }, [symbol, stmtType, period]);

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <FileSpreadsheet size={20} className="text-teal-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.financials")}</h3>
      </div>

      <div className="flex flex-wrap gap-2 mb-4">
        {STMT_TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setStmtType(tab.key)}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg transition ${
              stmtType === tab.key
                ? "bg-blue-600 text-white"
                : "bg-surface text-text-muted hover:text-foreground"
            }`}
          >
            {tab.label}
          </button>
        ))}
        <div className="ml-auto flex gap-1">
          {(["quarter", "year"] as Period[]).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`px-3 py-1.5 text-xs font-medium rounded-lg transition ${
                period === p
                  ? "bg-blue-600 text-white"
                  : "bg-surface text-text-muted hover:text-foreground"
              }`}
            >
              {p === "quarter" ? "Q" : "Y"}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="animate-pulse space-y-2">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="h-8 bg-surface rounded" />
          ))}
        </div>
      ) : !data || data.rows.length === 0 ? (
        <p className="text-sm text-text-muted">{t("common.no_data")}</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border-theme">
                <th className="text-left py-2 pr-4 text-text-muted font-medium" />
                {data.periods.map((p) => (
                  <th key={p} className="text-right py-2 px-2 text-text-muted font-medium whitespace-nowrap">
                    {p}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.rows.map((row, idx) => (
                <tr key={idx} className="border-b border-border-theme/50 hover:bg-surface/50 transition">
                  <td className="py-2 pr-4 text-foreground font-medium whitespace-nowrap">{row.label}</td>
                  {row.values.map((val, vi) => (
                    <td key={vi} className="text-right py-2 px-2 text-foreground tabular-nums">
                      {formatNumber(val / 1e9, 1)}B
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}
