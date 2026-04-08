"use client";

import { AlertTriangle, WifiOff } from "lucide-react";

interface StaleIndicatorProps {
  stale?: boolean;
  aiUnavailable?: boolean;
}

/**
 * Displays a small banner when data is stale or the AI service is unavailable.
 * Requirements: 1.6 (graceful degradation notice), 28.3 (stale data display)
 */
export function StaleIndicator({ stale, aiUnavailable }: StaleIndicatorProps) {
  if (!stale && !aiUnavailable) return null;

  return (
    <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-yellow-500/10 border border-yellow-500/20 text-yellow-600 dark:text-yellow-400 text-xs">
      {aiUnavailable ? (
        <>
          <WifiOff size={12} />
          <span>AI service unavailable — showing cached data</span>
        </>
      ) : (
        <>
          <AlertTriangle size={12} />
          <span>Showing cached data — live data temporarily unavailable</span>
        </>
      )}
    </div>
  );
}
