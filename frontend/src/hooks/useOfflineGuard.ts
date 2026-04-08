"use client";

import { useOnlineStatus } from "./useOnlineStatus";

/**
 * Returns true when the app is offline — use to disable write operations (Req 45.2).
 * Usage: const isOffline = useOfflineGuard(); then disable buttons with `disabled={isOffline}`.
 */
export function useOfflineGuard(): boolean {
  const { isOnline } = useOnlineStatus();
  return !isOnline;
}
