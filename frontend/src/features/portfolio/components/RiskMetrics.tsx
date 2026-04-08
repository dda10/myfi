"use client";

import { useEffect, useState } from "react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface RiskData {
  sharpeRatio: number;
  maxDrawdown: number;
  beta: number;
  volatility: number;
  var95: number;
}

export function RiskMetrics() {
  const { t, formatPercent } = useI18n();
  const [risk, setRisk] = useState<RiskData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    apiFetch<RiskData>("/api/portfolio/risk").then((data) => {
      if (!cancelled) {
        setRisk(data);
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="animate-pulse bg-zinc-800 rounded-xl h-28" />
        ))}
      </div>
    );
  }

  const metrics = [
    { label: t("portfolio.sharpe"), value: risk ? risk.sharpeRatio.toFixed(2) : "—", desc: "Risk-adjusted return" },
    { label: t("portfolio.max_drawdown"), value: risk ? formatPercent(risk.maxDrawdown) : "—", desc: "Largest peak-to-trough decline" },
    { label: t("portfolio.beta"), value: risk ? risk.beta.toFixed(2) : "—", desc: "Sensitivity to VN-Index" },
    { label: t("portfolio.volatility"), value: risk ? formatPercent(risk.volatility) : "—", desc: "Annualized std deviation" },
    { label: "VaR (95%)", value: risk ? formatPercent(risk.var95) : "—", desc: "Max expected daily loss at 95% confidence" },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {metrics.map((m) => (
        <div key={m.label} className="bg-zinc-900 border border-zinc-800 rounded-xl p-5">
          <p className="text-xs text-zinc-500 font-medium mb-1">{m.label}</p>
          <p className="text-2xl font-bold text-white mb-1">{m.value}</p>
          <p className="text-xs text-zinc-500">{m.desc}</p>
        </div>
      ))}
    </div>
  );
}
