"use client";

import { Target } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

interface Props {
  symbol: string;
  targetPrice?: number;
  currentPrice?: number;
}

export function AIValuation({ symbol, targetPrice, currentPrice }: Props) {
  const { t, formatCurrency, formatPercent } = useI18n();

  if (!targetPrice || !currentPrice) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 text-text-muted text-sm">
        {t("common.no_data")}
      </div>
    );
  }

  const diff = targetPrice - currentPrice;
  const pct = (diff / currentPrice) * 100;
  const isUpside = diff >= 0;
  const barPct = Math.min(Math.max((currentPrice / (targetPrice * 1.2)) * 100, 5), 95);

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Target size={20} className="text-purple-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.valuation")}</h3>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-4">
        <div>
          <p className="text-xs text-text-muted mb-1">{t("table.price")}</p>
          <p className="text-lg font-bold text-foreground">{formatCurrency(currentPrice)}</p>
        </div>
        <div>
          <p className="text-xs text-text-muted mb-1">{t("ideas.target_price")}</p>
          <p className="text-lg font-bold text-blue-400">{formatCurrency(targetPrice)}</p>
        </div>
        <div>
          <p className="text-xs text-text-muted mb-1">{isUpside ? "Upside" : "Downside"}</p>
          <p className={`text-lg font-bold ${isUpside ? "text-green-500" : "text-red-500"}`}>
            {isUpside ? "+" : ""}{formatPercent(pct)}
          </p>
        </div>
      </div>

      <div className="relative h-4 bg-zinc-700/50 rounded-full overflow-hidden">
        <div
          className={`absolute inset-y-0 left-0 rounded-full transition-all ${isUpside ? "bg-green-500" : "bg-red-500"}`}
          style={{ width: `${barPct}%` }}
        />
        <div
          className="absolute top-0 w-0.5 h-full bg-white/80"
          style={{ left: `${barPct}%` }}
          title={formatCurrency(currentPrice)}
        />
      </div>
      <div className="flex justify-between text-xs text-text-muted mt-1">
        <span>0</span>
        <span>{formatCurrency(targetPrice)} ({t("ideas.target_price")})</span>
      </div>
    </section>
  );
}
