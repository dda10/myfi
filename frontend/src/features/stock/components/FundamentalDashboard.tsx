"use client";

import { useState, useEffect } from "react";
import { BarChart3 } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface RatioData {
  pe?: number;
  pb?: number;
  evEbitda?: number;
  roe?: number;
  roa?: number;
  revenueGrowth?: number;
  profitGrowth?: number;
  divYield?: number;
  debtEquity?: number;
  sectorPe?: number;
  sectorPb?: number;
  sectorEvEbitda?: number;
  sectorRoe?: number;
  sectorRoa?: number;
  sectorRevenueGrowth?: number;
  sectorProfitGrowth?: number;
  sectorDivYield?: number;
  sectorDebtEquity?: number;
}

const METRICS: { key: string; i18nKey: string; stockField: keyof RatioData; sectorField: keyof RatioData; pct?: boolean }[] = [
  { key: "pe", i18nKey: "finance.pe", stockField: "pe", sectorField: "sectorPe" },
  { key: "pb", i18nKey: "finance.pb", stockField: "pb", sectorField: "sectorPb" },
  { key: "evEbitda", i18nKey: "finance.ev_ebitda", stockField: "evEbitda", sectorField: "sectorEvEbitda" },
  { key: "roe", i18nKey: "finance.roe", stockField: "roe", sectorField: "sectorRoe", pct: true },
  { key: "roa", i18nKey: "finance.roa", stockField: "roa", sectorField: "sectorRoa", pct: true },
  { key: "revGrowth", i18nKey: "finance.revenue_growth", stockField: "revenueGrowth", sectorField: "sectorRevenueGrowth", pct: true },
  { key: "profitGrowth", i18nKey: "finance.profit_growth", stockField: "profitGrowth", sectorField: "sectorProfitGrowth", pct: true },
  { key: "divYield", i18nKey: "finance.div_yield", stockField: "divYield", sectorField: "sectorDivYield", pct: true },
  { key: "debtEquity", i18nKey: "finance.debt_equity", stockField: "debtEquity", sectorField: "sectorDebtEquity" },
];

export function FundamentalDashboard({ symbol }: { symbol: string }) {
  const { t, formatNumber, formatPercent } = useI18n();
  const [data, setData] = useState<RatioData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<RatioData>(`/api/market/ratios/${symbol}`)
      .then((res) => { if (res) setData(res); })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse">
        <div className="h-5 w-1/3 bg-surface rounded mb-4" />
        <div className="grid grid-cols-3 gap-3">
          {Array.from({ length: 9 }).map((_, i) => (
            <div key={i} className="h-16 bg-surface rounded" />
          ))}
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 text-text-muted text-sm">
        {t("common.no_data")}
      </div>
    );
  }

  function fmt(val: number | undefined, pct?: boolean): string {
    if (val === undefined || val === null) return "—";
    return pct ? formatPercent(val) : formatNumber(val, 2);
  }

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <BarChart3 size={20} className="text-orange-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.overview")}</h3>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
        {METRICS.map((m) => {
          const stockVal = data[m.stockField] as number | undefined;
          const sectorVal = data[m.sectorField] as number | undefined;
          return (
            <div key={m.key} className="bg-surface rounded-lg p-3">
              <p className="text-xs text-text-muted mb-1">{t(m.i18nKey)}</p>
              <p className="text-sm font-semibold text-foreground">{fmt(stockVal, m.pct)}</p>
              {sectorVal !== undefined && (
                <p className="text-xs text-text-muted">
                  {t("table.sector")}: {fmt(sectorVal, m.pct)}
                </p>
              )}
            </div>
          );
        })}
      </div>
    </section>
  );
}
