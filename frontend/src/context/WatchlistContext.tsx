"use client";

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import { apiFetch } from "@/lib/api";

// --- Types ---

export interface WatchlistSymbolEntry {
  id: number;
  watchlistId: number;
  symbol: string;
  position: number;
  priceAlertAbove?: number | null;
  priceAlertBelow?: number | null;
}

export interface WatchlistData {
  id: number;
  userId: number;
  name: string;
  symbols: WatchlistSymbolEntry[];
}

interface WatchlistContextType {
  /** All named watchlists from backend */
  watchlists: WatchlistData[];
  /** Flat list of all watched symbols (for backward compat) */
  watchlist: string[];
  loading: boolean;
  error: string | null;
  isWatched: (symbol: string) => boolean;
  addToWatchlist: (symbol: string) => void;
  removeFromWatchlist: (symbol: string) => void;
  toggleWatchlist: (symbol: string) => void;
  reorderWatchlist: (newOrder: string[]) => void;
  /** Backend-synced operations */
  createWatchlist: (name: string) => Promise<WatchlistData | null>;
  renameWatchlist: (id: number, name: string) => Promise<void>;
  deleteWatchlist: (id: number) => Promise<void>;
  addSymbolToWatchlist: (wlId: number, symbol: string) => Promise<void>;
  removeSymbolFromWatchlist: (wlId: number, symbol: string) => Promise<void>;
  refreshWatchlists: () => Promise<void>;
}

const WatchlistContext = createContext<WatchlistContextType | undefined>(undefined);

const DEFAULT_WATCHLIST = ["VNM", "FPT", "SSI", "HPG", "MWG"];
const STORAGE_KEY = "myfi_watchlist";

export function WatchlistProvider({ children }: { children: ReactNode }) {
  const [watchlists, setWatchlists] = useState<WatchlistData[]>([]);
  const [localWatchlist, setLocalWatchlist] = useState<string[]>(DEFAULT_WATCHLIST);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [backendAvailable, setBackendAvailable] = useState(false);

  // Flat list: union of all watchlist symbols
  const watchlist = backendAvailable
    ? Array.from(new Set(watchlists.flatMap((w) => w.symbols.map((s) => s.symbol))))
    : localWatchlist;

  // Fetch watchlists from backend
  const refreshWatchlists = useCallback(async () => {
    try {
      const data = await apiFetch<WatchlistData[]>("/api/watchlists");
      if (data) {
        setWatchlists(data);
        setBackendAvailable(true);
        setError(null);
      }
    } catch {
      setError("Failed to load watchlists");
    }
  }, []);

  // Load on mount: try backend first, fall back to localStorage
  useEffect(() => {
    async function init() {
      setLoading(true);
      const data = await apiFetch<WatchlistData[]>("/api/watchlists");
      if (data) {
        setWatchlists(data);
        setBackendAvailable(true);
      } else {
        // Fallback to localStorage
        try {
          const stored = localStorage.getItem(STORAGE_KEY);
          if (stored) {
            const parsed = JSON.parse(stored);
            if (Array.isArray(parsed) && parsed.length > 0) {
              setLocalWatchlist(parsed);
            }
          }
        } catch {
          // keep defaults
        }
      }
      setLoading(false);
    }
    init();
  }, []);

  // Persist local watchlist to localStorage
  useEffect(() => {
    if (!backendAvailable) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(localWatchlist));
    }
  }, [localWatchlist, backendAvailable]);

  const isWatched = (symbol: string) => watchlist.includes(symbol.toUpperCase());

  const addToWatchlist = (symbol: string) => {
    const s = symbol.toUpperCase();
    if (backendAvailable && watchlists.length > 0) {
      // Add to first watchlist
      addSymbolToWatchlist(watchlists[0].id, s);
    } else {
      setLocalWatchlist((prev) => (prev.includes(s) ? prev : [...prev, s]));
    }
  };

  const removeFromWatchlist = (symbol: string) => {
    const s = symbol.toUpperCase();
    if (backendAvailable) {
      // Remove from all watchlists that contain it
      watchlists.forEach((wl) => {
        if (wl.symbols.some((ws) => ws.symbol === s)) {
          removeSymbolFromWatchlist(wl.id, s);
        }
      });
    } else {
      setLocalWatchlist((prev) => prev.filter((w) => w !== s));
    }
  };

  const toggleWatchlist = (symbol: string) => {
    isWatched(symbol) ? removeFromWatchlist(symbol) : addToWatchlist(symbol);
  };

  const reorderWatchlist = (newOrder: string[]) => {
    if (!backendAvailable) {
      setLocalWatchlist(newOrder);
    }
  };

  // Backend CRUD
  const createWatchlist = async (name: string): Promise<WatchlistData | null> => {
    const wl = await apiFetch<WatchlistData>("/api/watchlists", {
      method: "POST",
      body: JSON.stringify({ name }),
    });
    if (wl) {
      setWatchlists((prev) => [...prev, wl]);
    }
    return wl;
  };

  const renameWatchlist = async (id: number, name: string) => {
    await apiFetch(`/api/watchlists/${id}`, {
      method: "PUT",
      body: JSON.stringify({ name }),
    });
    setWatchlists((prev) => prev.map((w) => (w.id === id ? { ...w, name } : w)));
  };

  const deleteWatchlist = async (id: number) => {
    await apiFetch(`/api/watchlists/${id}`, { method: "DELETE" });
    setWatchlists((prev) => prev.filter((w) => w.id !== id));
  };

  const addSymbolToWatchlist = async (wlId: number, symbol: string) => {
    await apiFetch(`/api/watchlists/${wlId}/symbols`, {
      method: "POST",
      body: JSON.stringify({ symbol }),
    });
    await refreshWatchlists();
  };

  const removeSymbolFromWatchlist = async (wlId: number, symbol: string) => {
    await apiFetch(`/api/watchlists/${wlId}/symbols/${symbol}`, { method: "DELETE" });
    setWatchlists((prev) =>
      prev.map((w) =>
        w.id === wlId ? { ...w, symbols: w.symbols.filter((s) => s.symbol !== symbol) } : w
      )
    );
  };

  return (
    <WatchlistContext.Provider
      value={{
        watchlists,
        watchlist,
        loading,
        error,
        isWatched,
        addToWatchlist,
        removeFromWatchlist,
        toggleWatchlist,
        reorderWatchlist,
        createWatchlist,
        renameWatchlist,
        deleteWatchlist,
        addSymbolToWatchlist,
        removeSymbolFromWatchlist,
        refreshWatchlists,
      }}
    >
      {children}
    </WatchlistContext.Provider>
  );
}

export function useWatchlist() {
  const ctx = useContext(WatchlistContext);
  if (!ctx) throw new Error("useWatchlist must be used within WatchlistProvider");
  return ctx;
}
