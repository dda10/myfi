"use client";

import { usePolling, isVNTradingHours, type UsePollingResult } from "./usePolling";

export type FreshnessStatus = "fresh" | "stale" | "expired";

export interface PricePollingResult<T> extends UsePollingResult<T> {
  freshness: FreshnessStatus;
}

const ONE_MIN = 60_000;
const FIVE_MIN = 5 * 60_000;

/**
 * Compute freshness status from a last-updated timestamp.
 *  - fresh:   < 1 minute old
 *  - stale:   1–5 minutes old
 *  - expired: > 5 minutes old (or no data)
 */
export function getFreshnessStatus(lastUpdated: Date | null): FreshnessStatus {
  if (!lastUpdated) return "expired";
  const age = Date.now() - lastUpdated.getTime();
  if (age < ONE_MIN) return "fresh";
  if (age < FIVE_MIN) return "stale";
  return "expired";
}

function wrapWithFreshness<T>(poll: UsePollingResult<T>): PricePollingResult<T> {
  return { ...poll, freshness: getFreshnessStatus(poll.lastUpdated) };
}

/**
 * High-level hook that polls all price data sources at the correct intervals:
 *  - Portfolio summary: 15 s during VN trading hours (Mon–Fri 9:00–15:00 ICT), 300 s off-hours
 *  - Gold prices:       300 s always
 *  - Crypto prices:     60 s always
 *
 * Returns polling results with freshness status per data type.
 */
export function usePricePolling<TSummary, TGold, TCrypto>(
  fetchSummary: () => Promise<TSummary | null>,
  fetchGold: () => Promise<TGold | null>,
  fetchCrypto: () => Promise<TCrypto | null>,
): {
  summary: PricePollingResult<TSummary>;
  gold: PricePollingResult<TGold>;
  crypto: PricePollingResult<TCrypto>;
  isTradingHours: boolean;
} {
  const tradingHours = isVNTradingHours();
  const summaryInterval = tradingHours ? 15_000 : 300_000;

  const summaryPoll = usePolling<TSummary>(fetchSummary, summaryInterval);
  const goldPoll = usePolling<TGold>(fetchGold, 300_000);
  const cryptoPoll = usePolling<TCrypto>(fetchCrypto, 60_000);

  return {
    summary: wrapWithFreshness(summaryPoll),
    gold: wrapWithFreshness(goldPoll),
    crypto: wrapWithFreshness(cryptoPoll),
    isTradingHours: tradingHours,
  };
}

export { isVNTradingHours };
