"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";

// --- Types ---

export type DisplayCurrency = "VND" | "USD";

interface CurrencyContextType {
  currency: DisplayCurrency;
  toggleCurrency: () => void;
  fxRate: number;
  convertAndFormat: (vndAmount: number) => string;
}

// --- Constants ---

const STORAGE_KEY = "myfi-currency";
const FALLBACK_RATE = 25400;

// --- Context ---

const CurrencyContext = createContext<CurrencyContextType | undefined>(undefined);

// --- Provider ---

export function CurrencyProvider({ children }: { children: ReactNode }) {
  const [currency, setCurrency] = useState<DisplayCurrency>("VND");
  const [fxRate, setFxRate] = useState<number>(FALLBACK_RATE);
  const [mounted, setMounted] = useState(false);

  // Read persisted preference on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === "VND" || stored === "USD") {
        setCurrency(stored);
      }
    } catch {
      // localStorage unavailable
    }
    setMounted(true);
  }, []);

  // Persist whenever currency changes
  useEffect(() => {
    if (!mounted) return;
    try {
      localStorage.setItem(STORAGE_KEY, currency);
    } catch {
      // localStorage unavailable
    }
  }, [currency, mounted]);

  // Fetch FX rate on mount
  useEffect(() => {
    const controller = new AbortController();
    fetch("http://localhost:8080/api/prices/fx", { signal: controller.signal })
      .then((res) => res.json())
      .then((json) => {
        const rate = json?.data?.rate ?? json?.rate;
        if (typeof rate === "number" && rate > 0) {
          setFxRate(rate);
        }
      })
      .catch(() => {
        // keep fallback rate
      });
    return () => controller.abort();
  }, []);

  const toggleCurrency = useCallback(() => {
    setCurrency((prev) => (prev === "VND" ? "USD" : "VND"));
  }, []);

  const convertAndFormat = useCallback(
    (vndAmount: number): string => {
      if (currency === "VND") {
        const formatted = Math.round(vndAmount)
          .toString()
          .replace(/\B(?=(\d{3})+(?!\d))/g, ".");
        return `${formatted}₫`;
      }
      const usdAmount = vndAmount / fxRate;
      return `$${usdAmount.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
    },
    [currency, fxRate],
  );

  return (
    <CurrencyContext.Provider value={{ currency, toggleCurrency, fxRate, convertAndFormat }}>
      {children}
    </CurrencyContext.Provider>
  );
}

export function useCurrency() {
  const context = useContext(CurrencyContext);
  if (context === undefined) {
    throw new Error("useCurrency must be used within a CurrencyProvider");
  }
  return context;
}
