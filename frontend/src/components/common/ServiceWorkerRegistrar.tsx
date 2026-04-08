"use client";

import { useEffect } from "react";

/** Registers the service worker on mount (Req 45.1) */
export function ServiceWorkerRegistrar() {
  useEffect(() => {
    if (typeof window !== "undefined" && "serviceWorker" in navigator) {
      navigator.serviceWorker.register("/sw.js").catch(() => {
        // SW registration failed — offline caching unavailable
      });
    }
  }, []);

  return null;
}
