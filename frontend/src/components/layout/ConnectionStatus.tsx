"use client";

import { useOnlineStatus } from "@/hooks/useOnlineStatus";
import { usePolling } from "@/hooks/usePolling";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface MarketStatus {
  exchange: string;
  status: string;
  label: string;
  isTradingHour: boolean;
  nextSession: string;
}

const STATUS_CONFIG: Record<string, { color: string; pulse?: boolean }> = {
  pre_open: { color: "bg-yellow-400" },
  continuous: { color: "bg-green-500", pulse: true },
  atc: { color: "bg-orange-500" },
  break: { color: "bg-yellow-400" },
  closed: { color: "bg-gray-400" },
};

function getStatusI18nKey(status: string): string {
  switch (status) {
    case "pre_open":
      return "market.pre_open";
    case "continuous":
      return "market.continuous";
    case "atc":
      return "market.atc";
    case "break":
      return "market.break";
    case "closed":
      return "market.closed";
    default:
      return "market.closed";
  }
}

export function ConnectionStatus() {
  const { isOnline } = useOnlineStatus();
  const { t } = useI18n();

  const { data: rawStatus } = usePolling<MarketStatus[] | MarketStatus | Record<string, unknown>>(
    () => apiFetch("/api/market/status"),
    60_000,
    isOnline,
  );

  const connectionColor = isOnline ? "bg-green-500" : "bg-red-500";
  const connectionLabel = isOnline ? "Connected" : "Offline";

  // Normalize: API may return array, single object, or wrapped { data: [...] }
  let hose: MarketStatus | undefined;
  if (Array.isArray(rawStatus)) {
    hose = rawStatus.find((s: MarketStatus) => s.exchange === "HOSE") ?? rawStatus[0];
  } else if (rawStatus && typeof rawStatus === "object") {
    const obj = rawStatus as Record<string, unknown>;
    if ("data" in obj && Array.isArray(obj.data)) {
      hose = (obj.data as MarketStatus[]).find((s) => s.exchange === "HOSE");
    } else if ("exchange" in obj) {
      hose = rawStatus as MarketStatus;
    }
  }
  const statusKey = hose?.status ?? "closed";
  const config = STATUS_CONFIG[statusKey] ?? STATUS_CONFIG.closed;

  return (
    <span className="flex items-center gap-2">
      {/* Connection indicator */}
      <span className="flex items-center gap-1" title={connectionLabel}>
        <span className={`w-1.5 h-1.5 rounded-full ${connectionColor}`} />
      </span>

      {/* Market status badge */}
      {isOnline && (
        <span className="flex items-center gap-1" title={hose?.label ?? t("market.closed")}>
          <span
            className={`w-1.5 h-1.5 rounded-full ${config.color}${config.pulse ? " animate-pulse" : ""}`}
          />
          <span className="text-[10px] text-text-muted hidden sm:inline">
            {t(getStatusI18nKey(statusKey))}
          </span>
        </span>
      )}
    </span>
  );
}
