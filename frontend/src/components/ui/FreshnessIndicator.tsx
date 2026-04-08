"use client";

import { useMemo } from "react";
import { useI18n } from "@/context/I18nContext";

interface FreshnessIndicatorProps {
  lastUpdated: Date | null;
  isStale?: boolean;
  error?: string | null;
}

function getFreshnessColor(lastUpdated: Date | null): {
  color: string;
  labelKey: string;
} {
  if (!lastUpdated) return { color: "bg-zinc-500", labelKey: "common.no_data" };
  const ageMs = Date.now() - lastUpdated.getTime();
  const ONE_MIN = 60_000;
  const FIVE_MIN = 5 * 60_000;
  if (ageMs < ONE_MIN) return { color: "bg-green-500", labelKey: "freshness.live" };
  if (ageMs < FIVE_MIN) return { color: "bg-yellow-500", labelKey: "freshness.delayed" };
  return { color: "bg-red-500", labelKey: "freshness.stale" };
}

export function FreshnessIndicator({
  lastUpdated,
  isStale,
  error,
}: FreshnessIndicatorProps) {
  const { t } = useI18n();
  const { color, labelKey } = useMemo(
    () => getFreshnessColor(lastUpdated),
    // re-evaluate every render since time changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [lastUpdated, Date.now()],
  );

  const label = t(labelKey);
  const displayColor = error ? "bg-red-500" : isStale ? "bg-red-500" : color;
  const tooltip = error
    ? `Error — last update: ${lastUpdated?.toLocaleTimeString() ?? "never"}`
    : lastUpdated
      ? `${label} — ${lastUpdated.toLocaleTimeString()}`
      : t("common.no_data");

  return (
    <span className="relative group inline-flex items-center" title={tooltip}>
      <span
        className={`inline-block w-2 h-2 rounded-full ${displayColor}`}
        aria-label={tooltip}
      />
      <span className="pointer-events-none absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block whitespace-nowrap rounded bg-zinc-800 px-2 py-1 text-[10px] text-zinc-300 shadow-lg border border-zinc-700 z-50">
        {tooltip}
      </span>
    </span>
  );
}
