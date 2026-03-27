"use client";

import { useCurrency } from "@/context/CurrencyContext";

export function CurrencyToggle() {
  const { currency, toggleCurrency, fxRate } = useCurrency();

  const formattedRate = fxRate.toLocaleString("vi-VN");

  return (
    <button
      onClick={toggleCurrency}
      className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-text-muted hover:text-foreground transition rounded-full hover:bg-surface border border-border-theme"
      title={`1 USD = ${formattedRate} VND`}
      aria-label={`Display currency: ${currency}. Click to switch.`}
    >
      <span>{currency}</span>
      <span className="text-[10px] opacity-60">⇄</span>
    </button>
  );
}
