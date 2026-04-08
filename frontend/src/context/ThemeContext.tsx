"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, useMemo, ReactNode } from "react";

export type Theme = "light" | "dark";

/**
 * Chart color tokens that adapt to the current theme.
 * Designed for use with lightweight-charts (TradingView).
 * Requirement 30.2: chart colors update immediately with theme toggle.
 */
export interface ChartColors {
  background: string;
  text: string;
  grid: string;
  crosshair: string;
  upColor: string;
  downColor: string;
  volumeUp: string;
  volumeDown: string;
  borderUp: string;
  borderDown: string;
  wickUp: string;
  wickDown: string;
}

const CHART_COLORS: Record<Theme, ChartColors> = {
  light: {
    background: "#ffffff",
    text: "#333333",
    grid: "#e0e0e0",
    crosshair: "#9b9b9b",
    upColor: "#22c55e",
    downColor: "#ef4444",
    volumeUp: "rgba(34,197,94,0.4)",
    volumeDown: "rgba(239,68,68,0.4)",
    borderUp: "#16a34a",
    borderDown: "#dc2626",
    wickUp: "#16a34a",
    wickDown: "#dc2626",
  },
  dark: {
    background: "#0f1729",
    text: "#d1d5db",
    grid: "#1e293b",
    crosshair: "#6b7280",
    upColor: "#22c55e",
    downColor: "#ef4444",
    volumeUp: "rgba(34,197,94,0.35)",
    volumeDown: "rgba(239,68,68,0.35)",
    borderUp: "#16a34a",
    borderDown: "#dc2626",
    wickUp: "#16a34a",
    wickDown: "#dc2626",
  },
};

interface ThemeContextType {
  theme: Theme;
  toggleTheme: () => void;
  setTheme: (theme: Theme) => void;
  chartColors: ChartColors;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

const STORAGE_KEY = "ezistock-theme";

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>("light");
  const [mounted, setMounted] = useState(false);

  // Read persisted preference on mount (Req 30.1)
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === "dark" || stored === "light") {
        setThemeState(stored);
      }
    } catch {
      // localStorage unavailable — keep default
    }
    setMounted(true);
  }, []);

  // Apply the dark class on <html> and persist whenever theme changes (Req 30.2)
  useEffect(() => {
    if (!mounted) return;
    const root = document.documentElement;
    if (theme === "dark") {
      root.classList.add("dark");
    } else {
      root.classList.remove("dark");
    }
    try {
      localStorage.setItem(STORAGE_KEY, theme);
    } catch {
      // localStorage unavailable
    }
  }, [theme, mounted]);

  const toggleTheme = useCallback(() => {
    setThemeState((prev) => (prev === "dark" ? "light" : "dark"));
  }, []);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
  }, []);

  // Chart colors adapt immediately when theme changes (Req 30.2)
  const chartColors = useMemo(() => CHART_COLORS[theme], [theme]);

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, setTheme, chartColors }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
}
