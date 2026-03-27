"use client";

import React, { createContext, useContext, useState, ReactNode, useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";

// Maps URL segments to tab names and vice versa
const PATH_TO_TAB: Record<string, string> = {
  "":           "Overview",
  "overview":   "Overview",
  "portfolio":  "Portfolio",
  "markets":    "Markets",
  "filter":     "Filter",
  "allocation": "Allocation",
  "settings":   "Settings",
};

const TAB_TO_PATH: Record<string, string> = {
  "Overview":   "/overview",
  "Portfolio":  "/portfolio",
  "Markets":    "/markets",
  "Filter":     "/filter",
  "Allocation": "/allocation",
  "Settings":   "/settings",
};

interface AppContextType {
  activeSymbol: string;
  setActiveSymbol: (val: string) => void;
  activeTab: string;
  setActiveTab: (val: string) => void;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export function AppProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  const [activeSymbol, setActiveSymbol] = useState("SSI");

  // Derive active tab from URL path
  const segment = pathname.split("/").filter(Boolean)[0] ?? "";
  const activeTab = PATH_TO_TAB[segment] ?? "Overview";

  const setActiveTab = (tab: string) => {
    const path = TAB_TO_PATH[tab] ?? "/overview";
    router.push(path);
  };

  return (
    <AppContext.Provider value={{
      activeSymbol,
      setActiveSymbol,
      activeTab,
      setActiveTab
    }}>
      {children}
    </AppContext.Provider>
  );
}

export function useApp() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error("useApp must be used within an AppProvider");
  }
  return context;
}
