"use client";

import React, { createContext, useContext, useState, useEffect, ReactNode } from "react";

interface WatchlistContextType {
  watchlist: string[];
  isWatched: (symbol: string) => boolean;
  addToWatchlist: (symbol: string) => void;
  removeFromWatchlist: (symbol: string) => void;
  toggleWatchlist: (symbol: string) => void;
  reorderWatchlist: (newOrder: string[]) => void;
}

const WatchlistContext = createContext<WatchlistContextType | undefined>(undefined);

const DEFAULT_WATCHLIST = ["VNM", "FPT", "SSI", "HPG", "MWG"];
const STORAGE_KEY = "myfi_watchlist";

export function WatchlistProvider({ children }: { children: ReactNode }) {
  const [watchlist, setWatchlist] = useState<string[]>(DEFAULT_WATCHLIST);

  // Load from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        const parsed = JSON.parse(stored);
        if (Array.isArray(parsed) && parsed.length > 0) {
          setWatchlist(parsed);
        }
      }
    } catch (e) {
      console.warn("Failed to load watchlist from storage");
    }
  }, []);

  // Persist to localStorage on change
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(watchlist));
  }, [watchlist]);

  const isWatched = (symbol: string) => watchlist.includes(symbol.toUpperCase());

  const addToWatchlist = (symbol: string) => {
    const s = symbol.toUpperCase();
    setWatchlist(prev => prev.includes(s) ? prev : [...prev, s]);
  };

  const removeFromWatchlist = (symbol: string) => {
    const s = symbol.toUpperCase();
    setWatchlist(prev => prev.filter(w => w !== s));
  };

  const toggleWatchlist = (symbol: string) => {
    isWatched(symbol) ? removeFromWatchlist(symbol) : addToWatchlist(symbol);
  };

  const reorderWatchlist = (newOrder: string[]) => {
    setWatchlist(newOrder);
  };

  return (
    <WatchlistContext.Provider value={{
      watchlist, isWatched, addToWatchlist, removeFromWatchlist, toggleWatchlist, reorderWatchlist
    }}>
      {children}
    </WatchlistContext.Provider>
  );
}

export function useWatchlist() {
  const ctx = useContext(WatchlistContext);
  if (!ctx) throw new Error("useWatchlist must be used within WatchlistProvider");
  return ctx;
}
