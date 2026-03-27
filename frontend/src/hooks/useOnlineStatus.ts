"use client";

import { useState, useEffect, useCallback, useRef } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types ---

export interface SourceHealth {
  name: string;
  status: "ok" | "degraded" | "unavailable";
}

export interface OnlineStatus {
  /** Browser navigator.onLine — Req 34.1 */
  isOnline: boolean;
  /** Per-source health indicators — Req 34.6 */
  sourceHealth: SourceHealth[];
  /** Timestamp of last successful data fetch — Req 34.7 */
  lastFetchedAt: Date | null;
}

/**
 * useOnlineStatus — detects online/offline status using browser APIs (Req 34.1),
 * fetches per-source health from GET /api/health (Req 34.6),
 * and auto-refreshes on connectivity restore (Req 34.4).
 */
export function useOnlineStatus(onReconnect?: () => void): OnlineStatus {
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== "undefined" ? navigator.onLine : true,
  );
  const [sourceHealth, setSourceHealth] = useState<SourceHealth[]>([]);
  const [lastFetchedAt, setLastFetchedAt] = useState<Date | null>(null);
  const wasOfflineRef = useRef(false);
  const onReconnectRef = useRef(onReconnect);
  onReconnectRef.current = onReconnect;

  const fetchHealth = useCallback(async () => {
    try {
      const res = await fetch(`${API_URL}/api/health`, { signal: AbortSignal.timeout(5000) });
      if (!res.ok) return;
      const data = await res.json();
      setLastFetchedAt(new Date());

      // Parse health response — backend returns { status: "ok", sources?: {...} }
      if (data.sources && typeof data.sources === "object") {
        const health: SourceHealth[] = Object.entries(data.sources).map(
          ([name, status]) => ({
            name,
            status: status === "ok" ? "ok" : status === "degraded" ? "degraded" : "unavailable",
          } as SourceHealth),
        );
        setSourceHealth(health);
      } else {
        // Minimal health — just mark backend as ok
        setSourceHealth([{ name: "backend", status: "ok" }]);
      }
    } catch {
      // Health check failed — don't update sourceHealth, keep last known
    }
  }, []);

  // Req 34.1: Listen for online/offline events
  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      // Req 34.4: auto-refresh on connectivity restore
      if (wasOfflineRef.current) {
        wasOfflineRef.current = false;
        fetchHealth();
        onReconnectRef.current?.();
      }
    };
    const handleOffline = () => {
      setIsOnline(false);
      wasOfflineRef.current = true;
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    // Initial health check
    if (navigator.onLine) fetchHealth();

    // Poll health every 60s when online
    const id = setInterval(() => {
      if (navigator.onLine) fetchHealth();
    }, 60_000);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
      clearInterval(id);
    };
  }, [fetchHealth]);

  return { isOnline, sourceHealth, lastFetchedAt };
}
