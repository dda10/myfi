"use client";

import { Globe, TrendingUp, TrendingDown, Minus } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { usePolling } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";

interface WorldIndex {
  symbol: string;
  name: string;
  value: number;
  change: number;
  changePercent: number;
  region: string;
}

interface RegionGroup {
  region: string;
  items: WorldIndex[];
}

function groupByRegion(indices: WorldIndex[]): RegionGroup[] {
  const order = ["US", "Europe", "Asia"];
  const map = new Map<string, WorldIndex[]>();
  for (const idx of indices) {
    const list = map.get(idx.region) ?? [];
    list.push(idx);
    map.set(idx.region, list);
  }
  return order
    .filter((r) => map.has(r))
    .map((r) => ({ region: r, items: map.get(r)! }));
}

function ChangeIndicator({ change, changePercent }: { change: number; changePercent: number }) {
  const { formatNumber, formatPercent } = useI18n();
  const isUp = change > 0;
  const isDown = change < 0;
  const colorClass = isUp ? "text-positive" : isDown ? "text-negative" : "text-text-muted";
  const Icon = isUp ? TrendingUp : isDown ? TrendingDown : Minus;

  return (
    <span className={`flex items-center gap-1 text-xs font-medium ${colorClass}`}>
      <Icon size={12} />
      {isUp ? "+" : ""}{formatNumber(change, 2)} ({isUp ? "+" : ""}{formatPercent(changePercent)})
    </span>
  );
}

/** Unwrap backend response — handles { data: [...] } or raw array */
function unwrapArray<T>(raw: unknown): T[] | null {
  if (Array.isArray(raw)) return raw as T[];
  if (raw && typeof raw === "object" && "data" in (raw as Record<string, unknown>)) {
    const inner = (raw as Record<string, unknown>).data;
    if (Array.isArray(inner)) return inner as T[];
  }
  return null;
}

const REGION_MAP: Record<string, string> = {
  ".SPX": "US", ".DJI": "US", ".IXIC": "US",
  ".FTSE": "Europe", ".GDAXI": "Europe",
  ".N225": "Asia", ".HSI": "Asia", ".SSEC": "Asia", ".KS11": "Asia", ".TWII": "Asia",
};

const POLL_INTERVAL = 5 * 60 * 1000;

export function GlobalMarkets() {
  const { t, formatNumber } = useI18n();

  const fetcher = async () => {
    const raw = await apiFetch<unknown>("/api/market/world-indices");
    const items = unwrapArray<WorldIndex>(raw);
    if (!items) return null;
    // Attach region if missing
    return items.map(idx => ({
      ...idx,
      region: idx.region || REGION_MAP[idx.symbol] || "Other",
    }));
  };

  const { data, loading } = usePolling<WorldIndex[]>(fetcher, POLL_INTERVAL);

  if (loading && !data) {
    return (
      <section>
        <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
          <Globe size={18} className="text-blue-400" />
          {t("dashboard.global_markets")}
        </h2>
        <div className="bg-card-bg border border-border-theme rounded-xl p-4 animate-pulse space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="space-y-2">
              <div className="h-3 w-12 bg-surface rounded" />
              <div className="h-3 w-full bg-surface rounded" />
            </div>
          ))}
        </div>
      </section>
    );
  }

  if (!data || data.length === 0) {
    return (
      <section>
        <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
          <Globe size={18} className="text-blue-400" />
          {t("dashboard.global_markets")}
        </h2>
        <div className="bg-card-bg border border-border-theme rounded-xl p-4">
          <p className="text-sm text-text-muted">{t("common.no_data")}</p>
        </div>
      </section>
    );
  }

  const groups = groupByRegion(data);

  return (
    <section>
      <h2 className="text-lg font-semibold text-foreground mb-3 flex items-center gap-2">
        <Globe size={18} className="text-blue-400" />
        {t("dashboard.global_markets")}
      </h2>
      <div className="bg-card-bg border border-border-theme rounded-xl p-4 space-y-4">
        {groups.map((group) => (
          <div key={group.region}>
            <p className="text-xs text-text-muted font-semibold uppercase mb-2">{group.region}</p>
            <div className="space-y-1.5">
              {group.items.map((idx) => (
                <div key={idx.symbol} className="flex items-center justify-between">
                  <span className="text-sm text-foreground">{idx.name}</span>
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-foreground tabular-nums">
                      {formatNumber(idx.value, 2)}
                    </span>
                    <ChangeIndicator change={idx.change} changePercent={idx.changePercent} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
