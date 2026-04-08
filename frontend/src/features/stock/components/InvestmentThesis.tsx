"use client";

import { useState, useEffect } from "react";
import { Brain, ShieldAlert, TrendingUp, Loader2 } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface ThesisData {
  thesis: string;
  confidence: number;
  riskLevel: string;
  action: string;
}

export function InvestmentThesis({ symbol }: { symbol: string }) {
  const { t } = useI18n();
  const [data, setData] = useState<ThesisData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    setLoading(true);
    setError(false);
    apiFetch<ThesisData>(`/api/analyze/${symbol}`)
      .then((res) => {
        if (res) setData(res);
        else setError(true);
      })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 space-y-3 animate-pulse">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-4 w-full bg-surface rounded" />
        <div className="h-4 w-5/6 bg-surface rounded" />
        <div className="h-4 w-2/3 bg-surface rounded" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 text-text-muted text-sm">
        {t("error.ai_unavailable")}
      </div>
    );
  }

  const riskColor =
    data.riskLevel === "high" ? "text-red-500" : data.riskLevel === "medium" ? "text-yellow-500" : "text-green-500";

  const actionColor =
    data.action === "buy" ? "bg-green-500/20 text-green-400" : data.action === "sell" ? "bg-red-500/20 text-red-400" : "bg-yellow-500/20 text-yellow-400";

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Brain size={20} className="text-blue-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.ai_thesis")}</h3>
      </div>

      <p className="text-sm text-text-muted leading-relaxed mb-4 whitespace-pre-line">{data.thesis}</p>

      <div className="flex flex-wrap gap-3 mb-4">
        <div className="flex items-center gap-1.5 text-sm">
          <TrendingUp size={14} className="text-blue-400" />
          <span className="text-text-muted">{t("ideas.confidence")}:</span>
          <span className="font-semibold text-foreground">{data.confidence}%</span>
        </div>
        <div className="flex items-center gap-1.5 text-sm">
          <ShieldAlert size={14} className={riskColor} />
          <span className="text-text-muted">{t("portfolio.risk")}:</span>
          <span className={`font-semibold capitalize ${riskColor}`}>{data.riskLevel}</span>
        </div>
        <span className={`text-xs font-medium px-2.5 py-1 rounded-full capitalize ${actionColor}`}>
          {data.action}
        </span>
      </div>

      <p className="text-xs text-text-muted italic border-t border-border-theme pt-3">
        {t("chat.disclaimer")}
      </p>
    </section>
  );
}
