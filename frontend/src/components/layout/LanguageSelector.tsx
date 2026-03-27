"use client";

import { useState, useRef, useEffect } from "react";
import { Globe } from "lucide-react";
import { useI18n, Locale } from "@/context/I18nContext";

const LOCALE_OPTIONS: { value: Locale; label: string }[] = [
  { value: "vi-VN", label: "Tiếng Việt" },
  { value: "en-US", label: "English" },
];

export function LanguageSelector() {
  const { locale, setLocale } = useI18n();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const current = LOCALE_OPTIONS.find((o) => o.value === locale);

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen((prev) => !prev)}
        className="flex items-center gap-1.5 p-2 text-text-muted hover:text-foreground transition rounded-full hover:bg-surface"
        title={current?.label}
        aria-label="Select language"
      >
        <Globe size={20} />
        <span className="text-xs font-medium hidden sm:inline">
          {locale === "vi-VN" ? "VI" : "EN"}
        </span>
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-2 w-40 bg-card-bg border border-border-theme rounded-xl shadow-2xl py-1 z-50">
          {LOCALE_OPTIONS.map((option) => (
            <button
              key={option.value}
              onClick={() => {
                setLocale(option.value);
                setOpen(false);
              }}
              className={`w-full text-left px-4 py-2.5 text-sm transition hover:bg-surface-hover ${
                locale === option.value
                  ? "text-accent font-medium"
                  : "text-foreground"
              }`}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
