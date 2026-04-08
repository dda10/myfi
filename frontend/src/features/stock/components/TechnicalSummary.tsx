"use client";

import { useState, useEffect } from "react";
import { Activity } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

type Signal = "strongly_bullish" | "bullish" | "neutral" | "bearish" | "strongly_bearish";

interface Indicator {
  name: string;
  signal: Signal;
  value: number;
}

interface TechnicalData {
  composite: Signal;
  indicators: Indicator[];
}

const SIGNAL_CONFIG: Record<Signal, { color: string; bg: string }> = {
  strongly_bullish: { color: "text-green-400", bg: "bg-green-500/20" },
  bullish: { color: "text-green-500", bg: "bg-green-500/10" },
  neutral: { color: "text-yellow-500", bg: "bg-yellow-500/10" },
  bearish: { color: "text-red-500", bg: "bg-red-500/10" },
  strongly_bearish: { color: "text-red-400", bg: "bg-red-500/20" },
};

export function TechnicalSummary({ symbol }: { symbol: string }) {
  const { t } = useI18n();
  const [data, setData] = useState<TechnicalData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<TechnicalData | Record<string, unknown>>(`/api/market/technical/${symbol}`)
      .then((res) => {
        if (!res) return;
        // Handle wrapped response { data: { composite, indicators } }
        if ("composite" in res) {
          setData(res as TechnicalData);
        } else if ("data" in res && typeof res.data === "object" && res.data !== null) {
          setData(res.data as TechnicalData);
        }
      })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-12 w-full bg-surface rounded" />
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

  const cfg = SIGNAL_CONFIG[data.composite] ?? SIGNAL_CONFIG.neutral;

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Activity size={20} className="text-cyan-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.technical")}</h3>
      </div>

      <div className={`inline-block px-4 py-2 rounded-lg text-sm font-semibold mb-4 ${cfg.bg} ${cfg.color}`}>
        {t(`stock.${data.composite}`)}
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
        {(data.indicators ?? []).map((ind) => {
          const ic = SIGNAL_CONFIG[ind.signal] ?? SIGNAL_CONFIG.neutral;
          return (
            <div key={ind.name} className="bg-surface rounded-lg p-3">
              <p className="text-xs text-text-muted mb-1">{ind.name}</p>
              <p className="text-sm font-semibold text-foreground">{ind.value.toFixed(2)}</p>
              <p className={`text-xs font-medium ${ic.color}`}>{t(`stock.${ind.signal}`)}</p>
            </div>
          );
        })}
      </div>
    </section>
  );
}
