"use client";

import { useState, useEffect } from "react";
import { Users } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface Shareholder {
  name: string;
  shares: number;
  percentage: number;
  type: string;
}

const TYPE_COLORS: Record<string, string> = {
  Institutional: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  Individual: "bg-purple-500/20 text-purple-400 border-purple-500/30",
  State: "bg-amber-500/20 text-amber-400 border-amber-500/30",
  Foreign: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
};

export function Shareholders({ symbol }: { symbol: string }) {
  const { t, formatNumber, formatPercent } = useI18n();
  const [data, setData] = useState<Shareholder[] | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<Shareholder[]>(`/api/market/shareholders?symbol=${symbol}`)
      .then((res) => {
        if (res) {
          setData([...res].sort((a, b) => b.percentage - a.percentage));
        }
      })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-4 w-full bg-surface rounded" />
        <div className="h-4 w-full bg-surface rounded" />
        <div className="h-4 w-3/4 bg-surface rounded" />
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 text-text-muted text-sm">
        {t("common.no_data")}
      </div>
    );
  }

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Users size={20} className="text-blue-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.shareholders")}</h3>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border-theme text-text-muted text-xs">
              <th className="text-left py-2 pr-4 font-medium">{t("table.name")}</th>
              <th className="text-right py-2 px-4 font-medium">{t("stock.shares")}</th>
              <th className="text-right py-2 px-4 font-medium">{t("stock.ownership")}</th>
              <th className="text-right py-2 pl-4 font-medium">{t("stock.shareholder_type")}</th>
            </tr>
          </thead>
          <tbody>
            {data.map((sh, i) => (
              <tr key={i} className="border-b border-border-theme/50 hover:bg-surface/30 transition-colors">
                <td className="py-2.5 pr-4 text-foreground font-medium">{sh.name}</td>
                <td className="py-2.5 px-4 text-right text-foreground tabular-nums">
                  {formatNumber(sh.shares)}
                </td>
                <td className="py-2.5 px-4 text-right text-foreground tabular-nums">
                  {formatPercent(sh.percentage)}
                </td>
                <td className="py-2.5 pl-4 text-right">
                  <span className={`text-[11px] font-medium px-2 py-0.5 rounded border ${TYPE_COLORS[sh.type] ?? "bg-surface text-text-muted border-border-theme"}`}>
                    {sh.type}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
