"use client";

import { Gem, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { usePolling } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";

interface GoldPriceData {
  brand: string;
  type: string;
  buy: number;
  sell: number;
  change: number;
  changePercent: number;
  updatedAt: string;
}

function LoadingSkeleton() {
  return (
    <div className="bg-card-bg border border-border-theme rounded-xl p-4 animate-pulse space-y-3">
      <div className="h-4 w-24 bg-surface rounded" />
      {[1, 2].map((i) => (
        <div key={i} className="space-y-2">
          <div className="h-3 w-32 bg-surface rounded" />
          <div className="h-3 w-full bg-surface rounded" />
        </div>
      ))}
    </div>
  );
}

function PriceRow({ item }: { item: GoldPriceData }) {
  const { t, formatNumber } = useI18n();
  const isUp = item.change > 0;
  const isDown = item.change < 0;
  const colorClass = isUp ? "text-positive" : isDown ? "text-negative" : "text-text-muted";
  const Icon = isUp ? TrendingUp : isDown ? TrendingDown : Minus;

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-foreground">{item.type}</span>
        <span className={`flex items-center gap-1 text-xs font-medium ${colorClass}`}>
          <Icon size={12} />
          {isUp ? "+" : ""}{formatNumber(item.change, 0)}
        </span>
      </div>
      <div className="flex items-center justify-between text-xs">
        <span className="text-text-muted">{t("dashboard.gold_buy")}</span>
        <span className="text-foreground tabular-nums font-medium">{formatNumber(item.buy, 0)}</span>
      </div>
      <div className="flex items-center justify-between text-xs">
        <span className="text-text-muted">{t("dashboard.gold_sell")}</span>
        <span className="text-foreground tabular-nums font-medium">{formatNumber(item.sell, 0)}</span>
      </div>
    </div>
  );
}

const POLL_INTERVAL = 5 * 60 * 1000; // 5 minutes

export function GoldPrice() {
  const { t } = useI18n();

  const fetcher = async () => {
    const result = await apiFetch<GoldPriceData[] | Record<string, unknown>>("/api/market/gold");
    if (Array.isArray(result)) return result as GoldPriceData[];
    if (result && typeof result === "object" && "data" in result && Array.isArray(result.data)) {
      return result.data as GoldPriceData[];
    }
    return null;
  };

  const { data, loading, error } = usePolling<GoldPriceData[]>(fetcher, POLL_INTERVAL);

  // Group by brand
  const sjcItems = data?.filter((d) => d.brand === "SJC") ?? [];
  const btmcItems = data?.filter((d) => d.brand === "BTMC") ?? [];

  return (
    <section>
      <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
        <Gem size={18} className="text-yellow-500" />
        {t("dashboard.gold_price")}
      </h2>

      {loading && !data ? (
        <LoadingSkeleton />
      ) : !data || data.length === 0 ? (
        <div className="bg-card-bg border border-border-theme rounded-xl p-4">
          <p className="text-sm text-text-muted">{t("common.no_data")}</p>
        </div>
      ) : (
        <div className="bg-card-bg border border-border-theme rounded-xl p-4 space-y-4">
          {sjcItems.length > 0 && (
            <div>
              <p className="text-xs text-text-muted font-semibold uppercase mb-2">SJC</p>
              <div className="space-y-2">
                {sjcItems.map((item) => (
                  <PriceRow key={`${item.brand}-${item.type}`} item={item} />
                ))}
              </div>
            </div>
          )}
          {btmcItems.length > 0 && (
            <div>
              <p className="text-xs text-text-muted font-semibold uppercase mb-2">BTMC</p>
              <div className="space-y-2">
                {btmcItems.map((item) => (
                  <PriceRow key={`${item.brand}-${item.type}`} item={item} />
                ))}
              </div>
            </div>
          )}
          {error && (
            <p className="text-xs text-yellow-500 mt-1">{t("freshness.stale")}</p>
          )}
        </div>
      )}
    </section>
  );
}
