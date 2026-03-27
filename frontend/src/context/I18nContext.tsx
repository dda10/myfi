"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";
import viVN from "@/i18n/locales/vi-VN";
import enUS from "@/i18n/locales/en-US";

// --- Types ---

export type Locale = "vi-VN" | "en-US";

interface I18nContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string, vars?: Record<string, string>) => string;
  formatNumber: (value: number, decimals?: number) => string;
  formatCurrency: (value: number, currency?: string) => string;
  formatDate: (date: Date | string | number) => string;
  formatTime: (date: Date | string | number) => string;
  formatPercent: (value: number, decimals?: number) => string;
}

// --- Constants ---

const STORAGE_KEY = "myfi-locale";
const DEFAULT_LOCALE: Locale = "vi-VN";

const translations: Record<Locale, Record<string, string>> = {
  "vi-VN": viVN,
  "en-US": enUS,
};

// --- Context ---

const I18nContext = createContext<I18nContextType | undefined>(undefined);

// --- Formatting helpers ---

function formatNumberForLocale(value: number, locale: Locale, decimals?: number): string {
  const fractionDigits = decimals ?? 0;
  const abs = Math.abs(value);
  const fixed = abs.toFixed(fractionDigits);
  const [intPart, decPart] = fixed.split(".");

  // Add thousands separators
  const thousandsSep = locale === "vi-VN" ? "." : ",";
  const decimalSep = locale === "vi-VN" ? "," : ".";

  const withSep = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, thousandsSep);
  const sign = value < 0 ? "-" : "";

  if (decPart && fractionDigits > 0) {
    return `${sign}${withSep}${decimalSep}${decPart}`;
  }
  return `${sign}${withSep}`;
}

function formatCurrencyForLocale(value: number, locale: Locale, currency = "VND"): string {
  const formatted = formatNumberForLocale(value, locale, currency === "VND" ? 0 : 2);
  if (currency === "VND") {
    return locale === "vi-VN" ? `${formatted}₫` : `VND ${formatted}`;
  }
  // USD or other
  return locale === "vi-VN" ? `${formatted} ${currency}` : `${currency} ${formatted}`;
}

function formatDateForLocale(date: Date | string | number, locale: Locale): string {
  const d = date instanceof Date ? date : new Date(date);
  if (isNaN(d.getTime())) return String(date);

  const day = String(d.getDate()).padStart(2, "0");
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const year = d.getFullYear();

  // VN: dd/MM/yyyy, EN: MM/dd/yyyy
  return locale === "vi-VN" ? `${day}/${month}/${year}` : `${month}/${day}/${year}`;
}

function formatTimeForLocale(date: Date | string | number, locale: Locale): string {
  const d = date instanceof Date ? date : new Date(date);
  if (isNaN(d.getTime())) return String(date);

  const hours = d.getHours();
  const minutes = String(d.getMinutes()).padStart(2, "0");

  if (locale === "vi-VN") {
    // 24-hour format
    return `${String(hours).padStart(2, "0")}:${minutes}`;
  }
  // 12-hour format with AM/PM
  const period = hours >= 12 ? "PM" : "AM";
  const h12 = hours % 12 || 12;
  return `${h12}:${minutes} ${period}`;
}

function interpolate(template: string, vars: Record<string, string>): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => vars[key] ?? `{{${key}}}`);
}

// --- Provider ---

export function I18nProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(DEFAULT_LOCALE);
  const [mounted, setMounted] = useState(false);

  // Read persisted preference on mount (Req 38.4)
  useEffect(() => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored === "vi-VN" || stored === "en-US") {
        setLocaleState(stored);
      }
      // Req 38.5: default to vi-VN when no saved preference
    } catch {
      // localStorage unavailable — keep default
    }
    setMounted(true);
  }, []);

  // Persist whenever locale changes (Req 38.4)
  useEffect(() => {
    if (!mounted) return;
    try {
      localStorage.setItem(STORAGE_KEY, locale);
    } catch {
      // localStorage unavailable
    }
  }, [locale, mounted]);

  const setLocale = useCallback((l: Locale) => {
    setLocaleState(l);
  }, []);

  // Translation with interpolation (Req 38.6, 38.13)
  const t = useCallback(
    (key: string, vars?: Record<string, string>): string => {
      const str = translations[locale]?.[key] ?? key;
      return vars ? interpolate(str, vars) : str;
    },
    [locale],
  );

  // Locale-aware number formatting (Req 38.7)
  const formatNumber = useCallback(
    (value: number, decimals?: number): string => formatNumberForLocale(value, locale, decimals),
    [locale],
  );

  // Locale-aware currency formatting (Req 38.9)
  const formatCurrency = useCallback(
    (value: number, currency?: string): string => formatCurrencyForLocale(value, locale, currency),
    [locale],
  );

  // Locale-aware date formatting (Req 38.8)
  const formatDate = useCallback(
    (date: Date | string | number): string => formatDateForLocale(date, locale),
    [locale],
  );

  // Locale-aware time formatting (Req 38.10)
  const formatTime = useCallback(
    (date: Date | string | number): string => formatTimeForLocale(date, locale),
    [locale],
  );

  // Percent formatting helper
  const formatPercent = useCallback(
    (value: number, decimals = 2): string => {
      const formatted = formatNumberForLocale(value, locale, decimals);
      return `${formatted}%`;
    },
    [locale],
  );

  return (
    <I18nContext.Provider
      value={{ locale, setLocale, t, formatNumber, formatCurrency, formatDate, formatTime, formatPercent }}
    >
      {children}
    </I18nContext.Provider>
  );
}

export function useI18n() {
  const context = useContext(I18nContext);
  if (context === undefined) {
    throw new Error("useI18n must be used within an I18nProvider");
  }
  return context;
}
