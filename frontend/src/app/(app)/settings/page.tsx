"use client";

import { useState, useEffect } from "react";
import { Settings, Sun, Moon, Globe, Bot, Bell, BarChart3, Droplets, Save, CheckCircle } from "lucide-react";
import { useI18n, Locale } from "@/context/I18nContext";
import { useTheme, Theme } from "@/context/ThemeContext";
import { apiFetch } from "@/lib/api";
import { SettingsModule } from "@/components/dashboard/SettingsModule";

interface UserSettings {
  language: Locale;
  theme: Theme;
  llmProvider: string;
  notifications: { email: boolean; inApp: boolean };
  liquidityThreshold: "strict" | "normal" | "relaxed";
  chartDefaults: { interval: string; indicators: string[] };
}

const INTERVALS = ["1m", "5m", "15m", "1h", "1d", "1w", "1M"];
const INDICATORS = ["SMA", "EMA", "RSI", "MACD", "Bollinger", "VWAP", "OBV"];

export default function SettingsPage() {
  const { t, locale, setLocale } = useI18n();
  const { theme, setTheme } = useTheme();
  const [emailNotif, setEmailNotif] = useState(true);
  const [inAppNotif, setInAppNotif] = useState(true);
  const [liquidityThreshold, setLiquidityThreshold] = useState<"strict" | "normal" | "relaxed">("normal");
  const [defaultInterval, setDefaultInterval] = useState("1d");
  const [defaultIndicators, setDefaultIndicators] = useState<string[]>(["SMA", "RSI"]);
  const [saved, setSaved] = useState(false);
  const [saving, setSaving] = useState(false);
  const [showAIConfig, setShowAIConfig] = useState(false);

  // Load settings from localStorage on mount
  useEffect(() => {
    try {
      const stored = localStorage.getItem("ezistock-settings");
      if (stored) {
        const s = JSON.parse(stored);
        if (s.notifications) {
          setEmailNotif(s.notifications.email ?? true);
          setInAppNotif(s.notifications.inApp ?? true);
        }
        if (s.liquidityThreshold) setLiquidityThreshold(s.liquidityThreshold);
        if (s.chartDefaults?.interval) setDefaultInterval(s.chartDefaults.interval);
        if (s.chartDefaults?.indicators) setDefaultIndicators(s.chartDefaults.indicators);
      }
    } catch { /* ignore */ }
  }, []);

  const handleSave = async () => {
    setSaving(true);
    const settings: UserSettings = {
      language: locale,
      theme,
      llmProvider: "",
      notifications: { email: emailNotif, inApp: inAppNotif },
      liquidityThreshold,
      chartDefaults: { interval: defaultInterval, indicators: defaultIndicators },
    };
    localStorage.setItem("ezistock-settings", JSON.stringify(settings));
    await apiFetch("/api/auth/settings", { method: "PUT", body: JSON.stringify(settings) });
    setSaving(false);
    setSaved(true);
    setTimeout(() => setSaved(false), 2500);
  };

  const toggleIndicator = (ind: string) => {
    setDefaultIndicators((prev) =>
      prev.includes(ind) ? prev.filter((i) => i !== ind) : [...prev, ind]
    );
  };

  if (showAIConfig) {
    return (
      <div className="space-y-4">
        <button onClick={() => setShowAIConfig(false)} className="text-sm text-indigo-400 hover:text-indigo-300 transition">
          ← {t("btn.back")} {t("settings.title")}
        </button>
        <SettingsModule />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <h1 className="text-2xl font-bold text-foreground flex items-center gap-2">
        <Settings size={24} className="text-indigo-400" />
        {t("settings.title")}
      </h1>

      {/* Theme */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          {theme === "dark" ? <Moon size={16} /> : <Sun size={16} />}
          {t("settings.theme")}
        </h2>
        <div className="flex gap-3">
          {(["light", "dark"] as Theme[]).map((th) => (
            <button
              key={th}
              onClick={() => setTheme(th)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm transition ${
                theme === th
                  ? "bg-indigo-600 text-white"
                  : "bg-surface text-text-muted hover:bg-surface/80"
              }`}
            >
              {th === "light" ? <Sun size={14} /> : <Moon size={14} />}
              {t(`settings.theme_${th}`)}
            </button>
          ))}
        </div>
      </section>

      {/* Language */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <Globe size={16} />
          {t("settings.language")}
        </h2>
        <div className="flex gap-3">
          {([{ id: "vi-VN", label: "language.vi" }, { id: "en-US", label: "language.en" }] as const).map((lang) => (
            <button
              key={lang.id}
              onClick={() => setLocale(lang.id)}
              className={`px-4 py-2 rounded-lg text-sm transition ${
                locale === lang.id
                  ? "bg-indigo-600 text-white"
                  : "bg-surface text-text-muted hover:bg-surface/80"
              }`}
            >
              {t(lang.label)}
            </button>
          ))}
        </div>
      </section>

      {/* LLM Provider */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <Bot size={16} />
          {t("settings.ai_config")}
        </h2>
        <button
          onClick={() => setShowAIConfig(true)}
          className="text-sm text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 hover:bg-indigo-500/20 px-4 py-2 rounded-lg transition"
        >
          {t("settings.ai_config")} →
        </button>
      </section>

      {/* Notifications */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <Bell size={16} />
          {t("settings.notifications_config")}
        </h2>
        <div className="space-y-2">
          <label className="flex items-center gap-3 text-sm text-foreground">
            <input type="checkbox" checked={emailNotif} onChange={(e) => setEmailNotif(e.target.checked)} className="rounded" />
            Email
          </label>
          <label className="flex items-center gap-3 text-sm text-foreground">
            <input type="checkbox" checked={inAppNotif} onChange={(e) => setInAppNotif(e.target.checked)} className="rounded" />
            In-app
          </label>
        </div>
      </section>

      {/* Liquidity Threshold */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <Droplets size={16} />
          {t("screener.liquidity_tier")}
        </h2>
        <div className="flex gap-3">
          {(["strict", "normal", "relaxed"] as const).map((tier) => (
            <button
              key={tier}
              onClick={() => setLiquidityThreshold(tier)}
              className={`px-4 py-2 rounded-lg text-sm capitalize transition ${
                liquidityThreshold === tier
                  ? "bg-indigo-600 text-white"
                  : "bg-surface text-text-muted hover:bg-surface/80"
              }`}
            >
              {tier}
            </button>
          ))}
        </div>
      </section>

      {/* Chart Settings */}
      <section className="bg-card-bg border border-border-theme rounded-xl p-5">
        <h2 className="text-sm font-semibold text-foreground mb-3 flex items-center gap-2">
          <BarChart3 size={16} />
          {t("chart.interval")} & {t("chart.indicators")}
        </h2>
        <div className="space-y-3">
          <div>
            <label className="text-xs text-text-muted font-medium block mb-1.5">{t("chart.interval")}</label>
            <div className="flex gap-2 flex-wrap">
              {INTERVALS.map((iv) => (
                <button
                  key={iv}
                  onClick={() => setDefaultInterval(iv)}
                  className={`text-xs px-3 py-1.5 rounded-lg transition ${
                    defaultInterval === iv
                      ? "bg-indigo-600 text-white"
                      : "bg-surface text-text-muted hover:bg-surface/80"
                  }`}
                >
                  {iv}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="text-xs text-text-muted font-medium block mb-1.5">{t("chart.indicators")}</label>
            <div className="flex gap-2 flex-wrap">
              {INDICATORS.map((ind) => (
                <button
                  key={ind}
                  onClick={() => toggleIndicator(ind)}
                  className={`text-xs px-3 py-1.5 rounded-lg transition ${
                    defaultIndicators.includes(ind)
                      ? "bg-indigo-600 text-white"
                      : "bg-surface text-text-muted hover:bg-surface/80"
                  }`}
                >
                  {ind}
                </button>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Save Button */}
      <button
        onClick={handleSave}
        disabled={saving}
        className={`flex items-center justify-center gap-2 w-full py-3 rounded-xl text-sm font-bold shadow-lg transition ${
          saved
            ? "bg-green-600 text-white"
            : "bg-indigo-600 hover:bg-indigo-500 text-white"
        }`}
      >
        {saved ? (
          <><CheckCircle size={18} /> {t("btn.save")}!</>
        ) : (
          <><Save size={18} /> {t("btn.save")}</>
        )}
      </button>
    </div>
  );
}
