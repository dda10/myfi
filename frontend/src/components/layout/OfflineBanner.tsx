"use client";

import { WifiOff, Wifi, Clock, AlertCircle, CheckCircle } from "lucide-react";
import { useOnlineStatus, type SourceHealth } from "@/hooks/useOnlineStatus";
import { useI18n } from "@/context/I18nContext";

// --- Helpers ---

function statusColor(status: SourceHealth["status"]) {
  switch (status) {
    case "ok": return "bg-green-500";
    case "degraded": return "bg-yellow-500";
    case "unavailable": return "bg-red-500";
  }
}

function statusIcon(status: SourceHealth["status"]) {
  switch (status) {
    case "ok": return <CheckCircle size={10} className="text-green-400" />;
    case "degraded": return <AlertCircle size={10} className="text-yellow-400" />;
    case "unavailable": return <AlertCircle size={10} className="text-red-400" />;
  }
}

// --- Component ---

interface OfflineBannerProps {
  /** Called when connectivity is restored — triggers data refresh (Req 34.4) */
  onReconnect?: () => void;
}

export function OfflineBanner({ onReconnect }: OfflineBannerProps) {
  const { isOnline, sourceHealth, lastFetchedAt } = useOnlineStatus(onReconnect);
  const { t, formatTime } = useI18n();

  // Req 34.2: persistent "Offline Mode" banner when offline
  if (!isOnline) {
    return (
      <div className="w-full bg-red-900/80 border-b border-red-700 px-4 py-2 flex items-center justify-between" role="alert">
        <div className="flex items-center gap-2">
          <WifiOff size={16} className="text-red-300" />
          <span className="text-red-100 text-sm font-medium">{t("offline.banner")}</span>
          <span className="text-red-300 text-xs">— {t("offline.cached_data")}</span>
        </div>
        {/* Req 34.7: timestamp of last successful fetch */}
        {lastFetchedAt && (
          <div className="flex items-center gap-1 text-red-300 text-xs">
            <Clock size={12} />
            <span>Last updated: {formatTime(lastFetchedAt)}</span>
          </div>
        )}
      </div>
    );
  }

  // Req 34.6: per-source health indicators when some sources are degraded
  const hasDegraded = sourceHealth.some((s) => s.status !== "ok");
  if (!hasDegraded || sourceHealth.length === 0) return null;

  return (
    <div className="w-full bg-amber-900/50 border-b border-amber-700/50 px-4 py-1.5 flex items-center justify-between">
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1.5">
          <Wifi size={14} className="text-amber-300" />
          <span className="text-amber-200 text-xs font-medium">Degraded Mode</span>
        </div>
        {/* Per-source health dots */}
        <div className="flex items-center gap-2">
          {sourceHealth.map((s) => (
            <span key={s.name} className="flex items-center gap-1 text-xs text-zinc-300" title={`${s.name}: ${s.status}`}>
              {statusIcon(s.status)}
              <span className="hidden sm:inline">{s.name}</span>
              <span className={`inline-block w-1.5 h-1.5 rounded-full ${statusColor(s.status)} sm:hidden`} />
            </span>
          ))}
        </div>
      </div>
      {lastFetchedAt && (
        <div className="flex items-center gap-1 text-amber-300/70 text-xs">
          <Clock size={11} />
          <span>{formatTime(lastFetchedAt)}</span>
        </div>
      )}
    </div>
  );
}
