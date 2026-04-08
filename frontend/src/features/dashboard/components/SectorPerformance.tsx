"use client";

import { TrendingUp, TrendingDown, ArrowRight } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";

interface SectorData {
  sector: string;
  todayChange: number;
  trend: "uptrend" | "downtrend" | "sideways";
}

const ICB_SECTORS = [
  "VNIT", "VNIND", "VNCONS", "VNCOND", "VNHEAL",
  "VNENE", "VNUTI", "VNREAL", "VNFIN", "VNMAT", "VNCOM",
] as const;

function TrendIcon({ trend }: { trend: string }) {
  if (trend === "uptrend") return <TrendingUp size={14} className="text-green-500" />;
  if (trend === "downtrend") return <TrendingDown size={14} className="text-red-500" />;
  return <ArrowRight size={14} className="text-yellow-500" />;
}

export function SectorPerformance() {
  const { t, formatPercent } = useI18n();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;

  const { data, loading } = usePolling<SectorData[]>(
    () => apiFetch<SectorData[]>("/api/sectors/performance"),
    interval,
  );

  const sectorMap = new Map((Array.isArray(data) ? data : []).map((s) => [s.sector, s]));

  return (
    <section>
      <h2 className="text-lg font-semibold text-foreground mb-3">
        {t("dashboard.sector_performance")}
      </h2>
      <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
        <div className="grid grid-cols-[1fr_auto_auto] gap-x-4 px-4 py-2 border-b border-border-theme text-xs text-text-muted font-medium">
          <span>{t("table.sector")}</span>
          <span>{t("table.change_pct")}</span>
          <span className="w-6" />
        </div>
        {loading && !data ? (
          <div className="p-4 space-y-2">
            {[1, 2, 3, 4, 5].map((i) => (
              <div key={i} className="h-6 bg-surface rounded animate-pulse" />
            ))}
          </div>
        ) : (
          <div className="divide-y divide-border-theme">
            {ICB_SECTORS.map((code) => {
              const s = sectorMap.get(code);
              const change = s?.todayChange ?? 0;
              const isUp = change > 0;
              const isDown = change < 0;
              const colorClass = isUp ? "text-green-500" : isDown ? "text-red-500" : "text-text-muted";

              return (
                <div
                  key={code}
                  className="grid grid-cols-[1fr_auto_auto] gap-x-4 px-4 py-2.5 items-center hover:bg-surface transition text-sm"
                >
                  <span className="text-foreground font-medium">
                    {t(`sector.${code}`)} <span className="text-text-muted text-xs">({code})</span>
                  </span>
                  <span className={`font-medium tabular-nums ${colorClass}`}>
                    {isUp ? "+" : ""}{formatPercent(change)}
                  </span>
                  <TrendIcon trend={s?.trend ?? "sideways"} />
                </div>
              );
            })}
          </div>
        )}
      </div>
    </section>
  );
}
