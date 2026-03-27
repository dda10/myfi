"use client";

import { useEffect, useRef, useState, useCallback } from "react";

export interface UsePollingResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  lastUpdated: Date | null;
  isStale: boolean;
}

const STALE_THRESHOLD_MS = 5 * 60 * 1000; // 5 minutes

/**
 * Returns true if current time is within VN stock trading hours:
 * Monday–Friday, 9:00–15:00 ICT (UTC+7).
 */
export function isVNTradingHours(now: Date = new Date()): boolean {
  // Convert to ICT (UTC+7)
  const utcMs = now.getTime() + now.getTimezoneOffset() * 60_000;
  const ictDate = new Date(utcMs + 7 * 3600_000);
  const day = ictDate.getDay(); // 0=Sun, 6=Sat
  if (day === 0 || day === 6) return false;
  const hour = ictDate.getHours();
  return hour >= 9 && hour < 15;
}

export function usePolling<T>(
  fetchFn: () => Promise<T | null>,
  intervalMs: number,
  enabled: boolean = true,
): UsePollingResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [isStale, setIsStale] = useState(false);
  const mountedRef = useRef(true);
  const fetchRef = useRef(fetchFn);
  fetchRef.current = fetchFn;

  const doFetch = useCallback(async () => {
    if (!mountedRef.current) return;
    setLoading(true);
    try {
      const result = await fetchRef.current();
      if (!mountedRef.current) return;
      if (result !== null) {
        setData(result);
        setLastUpdated(new Date());
        setError(null);
        setIsStale(false);
      } else {
        setError("Failed to fetch data");
        // retain last known data — don't clear `data`
      }
    } catch {
      if (mountedRef.current) {
        setError("Failed to fetch data");
      }
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  }, []);

  // Initial fetch + polling
  useEffect(() => {
    mountedRef.current = true;
    if (!enabled) return;
    doFetch();
    const id = setInterval(doFetch, intervalMs);
    return () => {
      mountedRef.current = false;
      clearInterval(id);
    };
  }, [doFetch, intervalMs, enabled]);

  // Staleness checker — runs every 30s
  useEffect(() => {
    const id = setInterval(() => {
      if (lastUpdated) {
        setIsStale(Date.now() - lastUpdated.getTime() > STALE_THRESHOLD_MS);
      }
    }, 30_000);
    return () => clearInterval(id);
  }, [lastUpdated]);

  return { data, loading, error, lastUpdated, isStale };
}
