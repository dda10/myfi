"use client";

import { useState, useEffect } from "react";
import { Users, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface AnalystReport {
  analystName: string;
  brokerage: string;
  recommendation: "buy" | "hold" | "sell";
  targetPrice: number;
  date: string;
  accuracyScore: number;
}

interface ConsensusData {
  recommendation: string;
  targetPrice: number;
  numAnalysts: number;
  buyCount: number;
  holdCount: number;
  sellCount: number;
}

const REC_STYLE: Record<string, string> = {
  buy: "bg-green-500/20 text-green-400",
  hold: "bg-yellow-500/20 text-yellow-400",
  sell: "bg-red-500/20 text-red-400",
};

const REC_ICON: Record<string, typeof TrendingUp> = {
  buy: TrendingUp,
  hold: Minus,
  sell: TrendingDown,
};

export function AnalystConsensus({ symbol }: { symbol: string }) {
  const { t, formatNumber, formatDate } = useI18n();
  const [consensus, setConsensus] = useState<ConsensusData | null>(null);
  const [reports, setReports] = useState<AnalystReport[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      apiFetch<ConsensusData>(`/api/analyst/consensus/${symbol}`),
      apiFetch<AnalystReport[]>(`/api/analyst/reports/${symbol}`),
    ])
      .then(([c, r]) => {
        if (c) setConsensus(c);
        if (r) setReports(r);
      })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-20 bg-surface rounded" />
        {[1, 2, 3].map((i) => <div key={i} className="h-12 bg-surface rounded" />)}
      </div>
    );
  }

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Users size={20} className="text-indigo-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.analyst_consensus")}</h3>
      </div>

      {/* Consensus Summary */}
      {consensus && (
        <div className="bg-surface rounded-lg p-4 mb-4">
          <div className="flex items-center justify-between mb-3">
            <span className={`text-sm font-bold px-3 py-1 rounded-full ${REC_STYLE[consensus.recommendation.toLowerCase()] ?? REC_STYLE.hold}`}>
              {consensus.recommendation.toUpperCase()}
            </span>
            <span className="text-sm text-text-muted">
              {consensus.numAnalysts} {t("stock.analysts")}
            </span>
          </div>
          <div className="text-2xl font-bold text-foreground mb-2">
            {t("ideas.target_price")}: {formatNumber(consensus.targetPrice)}₫
          </div>
          <div className="flex gap-4 text-xs">
            <span className="text-green-400">{t("stock.buy_count")}: {consensus.buyCount}</span>
            <span className="text-yellow-400">{t("stock.hold_count")}: {consensus.holdCount}</span>
            <span className="text-red-400">{t("stock.sell_count")}: {consensus.sellCount}</span>
          </div>
        </div>
      )}

      {/* Individual Reports */}
      {reports.length === 0 ? (
        <p className="text-sm text-text-muted">{t("common.no_data")}</p>
      ) : (
        <div className="space-y-2">
          {reports.map((r, idx) => {
            const Icon = REC_ICON[r.recommendation] ?? Minus;
            return (
              <div key={idx} className="flex items-center gap-3 bg-surface rounded-lg p-3">
                <Icon size={16} className={REC_STYLE[r.recommendation]?.split(" ")[1] ?? "text-zinc-400"} />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-foreground truncate">{r.analystName}</p>
                  <p className="text-xs text-text-muted">{r.brokerage} • {formatDate(r.date)}</p>
                </div>
                <div className="text-right">
                  <p className="text-sm font-medium text-foreground">{formatNumber(r.targetPrice)}₫</p>
                  <p className="text-xs text-text-muted">{t("stock.accuracy")}: {r.accuracyScore}%</p>
                </div>
                <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${REC_STYLE[r.recommendation] ?? REC_STYLE.hold}`}>
                  {r.recommendation.toUpperCase()}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </section>
  );
}
