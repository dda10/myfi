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
 * High-level hook that polls stock price data at the correct interval:
 *  - Portfolio summary: 15 s during VN trading hours (Mon–Fri 9:00–15:00 ICT), 300 s off-hours
 *
 * Returns polling result with freshness status.
 */
export function usePricePolling<TSummary>(
  fetchSummary: () => Promise<TSummary | null>,
): {
  summary: PricePollingResult<TSummary>;
  isTradingHours: boolean;
} {
  const tradingHours = isVNTradingHours();
  const summaryInterval = tradingHours ? 15_000 : 300_000;

  const summaryPoll = usePolling<TSummary>(fetchSummary, summaryInterval);

  return {
    summary: wrapWithFreshness(summaryPoll),
    isTradingHours: tradingHours,
  };
}

export { isVNTradingHours };
