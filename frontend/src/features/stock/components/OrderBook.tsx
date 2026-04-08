"use client";

import { useCallback } from "react";
import { BookOpen } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";

interface PriceDepthLevel {
  price: number;
  volume: number;
}

interface PriceDepth {
  symbol: string;
  bids: PriceDepthLevel[];
  asks: PriceDepthLevel[];
  matchPrice: number;
  matchVolume: number;
  totalBidVolume: number;
  totalAskVolume: number;
}

const POLL_TRADING = 30_000;
const POLL_OFF = 5 * 60_000;

export function OrderBook({ symbol }: { symbol: string }) {
  const { t, formatNumber, formatCurrency } = useI18n();

  const fetchDepth = useCallback(
    () => apiFetch<PriceDepth>(`/api/market/price-depth?symbol=${symbol}`),
    [symbol],
  );

  const interval = isVNTradingHours() ? POLL_TRADING : POLL_OFF;
  const { data, loading } = usePolling(fetchDepth, interval);

  if (loading && !data) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-32 w-full bg-surface rounded" />
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

  const { bids, asks, matchPrice, matchVolume, totalBidVolume, totalAskVolume } = data;

  // Max volume across all levels for bar width calculation
  const maxVol = Math.max(
    ...bids.map((l) => l.volume),
    ...asks.map((l) => l.volume),
    1,
  );

  const totalVol = totalBidVolume + totalAskVolume;
  const bidPct = totalVol > 0 ? (totalBidVolume / totalVol) * 100 : 50;

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <BookOpen size={20} className="text-cyan-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.order_book")}</h3>
      </div>

      {/* Column headers */}
      <div className="grid grid-cols-2 gap-4 mb-2 text-xs text-text-muted font-medium">
        <div className="flex justify-between px-2">
          <span>{t("stock.bid")}</span>
          <span>{t("finance.volume")}</span>
        </div>
        <div className="flex justify-between px-2">
          <span>{t("finance.volume")}</span>
          <span>{t("stock.ask")}</span>
        </div>
      </div>

      {/* Bid / Ask rows — 3 levels */}
      <div className="space-y-1">
        {Array.from({ length: 3 }).map((_, i) => {
          const bid = bids[i];
          const ask = asks[i];
          const bidW = bid ? (bid.volume / maxVol) * 100 : 0;
          const askW = ask ? (ask.volume / maxVol) * 100 : 0;

          return (
            <div key={i} className="grid grid-cols-2 gap-4">
              {/* Bid side */}
              <div className="relative flex justify-between items-center px-2 py-1.5 rounded overflow-hidden">
                <div
                  className="absolute inset-y-0 right-0 bg-positive/10 rounded"
                  style={{ width: `${bidW}%` }}
                />
                <span className="relative text-sm font-semibold text-positive tabular-nums">
                  {bid ? formatCurrency(bid.price) : "—"}
                </span>
                <span className="relative text-xs text-text-muted tabular-nums">
                  {bid ? formatNumber(bid.volume) : "—"}
                </span>
              </div>

              {/* Ask side */}
              <div className="relative flex justify-between items-center px-2 py-1.5 rounded overflow-hidden">
                <div
                  className="absolute inset-y-0 left-0 bg-negative/10 rounded"
                  style={{ width: `${askW}%` }}
                />
                <span className="relative text-xs text-text-muted tabular-nums">
                  {ask ? formatNumber(ask.volume) : "—"}
                </span>
                <span className="relative text-sm font-semibold text-negative tabular-nums">
                  {ask ? formatCurrency(ask.price) : "—"}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Match price / volume center row */}
      <div className="mt-4 flex items-center justify-center gap-4 py-2 bg-surface rounded-lg">
        <div className="text-center">
          <p className="text-[10px] text-text-muted">{t("stock.match_price")}</p>
          <p className="text-sm font-bold text-foreground tabular-nums">{formatCurrency(matchPrice)}</p>
        </div>
        <div className="w-px h-6 bg-border-theme" />
        <div className="text-center">
          <p className="text-[10px] text-text-muted">{t("stock.match_volume")}</p>
          <p className="text-sm font-bold text-foreground tabular-nums">{formatNumber(matchVolume)}</p>
        </div>
      </div>

      {/* Bid vs Ask volume ratio bar */}
      <div className="mt-3">
        <div className="flex justify-between text-[10px] text-text-muted mb-1">
          <span>{t("stock.bid")} {formatNumber(totalBidVolume)}</span>
          <span>{formatNumber(totalAskVolume)} {t("stock.ask")}</span>
        </div>
        <div className="h-2 rounded-full overflow-hidden flex bg-surface">
          <div className="bg-positive/60 rounded-l-full" style={{ width: `${bidPct}%` }} />
          <div className="bg-negative/60 rounded-r-full flex-1" />
        </div>
      </div>
    </section>
  );
}
