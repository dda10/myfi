"use client";

import { useState, useEffect } from "react";
import { Building2 } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface Subsidiary {
  name: string;
  ownership: number;
  industry: string;
  status: string;
}

const STATUS_STYLES: Record<string, string> = {
  Active: "text-green-400 bg-green-500/10 border-green-500/20",
  Inactive: "text-gray-400 bg-gray-500/10 border-gray-500/20",
};

export function Subsidiaries({ symbol }: { symbol: string }) {
  const { t, formatPercent } = useI18n();
  const [data, setData] = useState<Subsidiary[] | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<Subsidiary[]>(`/api/market/subsidiaries?symbol=${symbol}`)
      .then((res) => { if (res) setData(res); })
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
        <Building2 size={20} className="text-purple-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.subsidiaries")}</h3>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border-theme text-text-muted text-xs">
              <th className="text-left py-2 pr-4 font-medium">{t("table.name")}</th>
              <th className="text-right py-2 px-4 font-medium">{t("stock.ownership")}</th>
              <th className="text-left py-2 px-4 font-medium">{t("stock.industry")}</th>
              <th className="text-right py-2 pl-4 font-medium">{t("stock.status")}</th>
            </tr>
          </thead>
          <tbody>
            {data.map((sub, i) => (
              <tr key={i} className="border-b border-border-theme/50 hover:bg-surface/30 transition-colors">
                <td className="py-2.5 pr-4 text-foreground font-medium">{sub.name}</td>
                <td className="py-2.5 px-4 text-right text-foreground tabular-nums">
                  {formatPercent(sub.ownership)}
                </td>
                <td className="py-2.5 px-4 text-text-muted">{sub.industry}</td>
                <td className="py-2.5 pl-4 text-right">
                  <span className={`text-[11px] font-medium px-2 py-0.5 rounded border ${STATUS_STYLES[sub.status] ?? "text-text-muted bg-surface border-border-theme"}`}>
                    {sub.status}
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
