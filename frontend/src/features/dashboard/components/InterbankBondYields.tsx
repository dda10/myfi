"use client";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { useI18n } from "@/context/I18nContext";
import { useTheme } from "@/context/ThemeContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";

interface MacroData {
  interbankRates: { tenor: string; rate: number }[];
  bondYields: { tenor: string; yield: number }[];
}

export function InterbankBondYields() {
  const { t, formatNumber } = useI18n();
  const { theme } = useTheme();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;
  const isDark = theme === "dark";

  const { data, loading } = usePolling<MacroData>(
    () => apiFetch<MacroData>("/api/market/macro"),
    interval,
  );

  const tooltipStyle = {
    backgroundColor: isDark ? "#18181b" : "#fff",
    border: `1px solid ${isDark ? "#3f3f46" : "#e5e7eb"}`,
    borderRadius: 8,
    color: isDark ? "#fff" : "#111",
  };
  const tickFill = isDark ? "#a1a1aa" : "#6b7280";
  const gridStroke = isDark ? "#27272a" : "#e5e7eb";

  return (
    <section>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Interbank Rates */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-3">
            {t("dashboard.interbank_rates")}
          </h2>
          <div className="bg-card-bg border border-border-theme rounded-xl p-4">
            {loading && !data ? (
              <div className="h-48 bg-surface rounded animate-pulse" />
            ) : data?.interbankRates && data.interbankRates.length > 0 ? (
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={data.interbankRates} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke={gridStroke} />
                    <XAxis dataKey="tenor" tick={{ fill: tickFill, fontSize: 11 }} />
                    <YAxis tick={{ fill: tickFill, fontSize: 11 }} tickFormatter={(v) => `${v}%`} />
                    <Tooltip contentStyle={tooltipStyle} formatter={(v) => [`${formatNumber(Number(v), 2)}%`, "Rate"]} />
                    <Line type="monotone" dataKey="rate" stroke="#6366f1" strokeWidth={2} dot={{ r: 3 }} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <p className="text-sm text-text-muted py-8 text-center">{t("common.no_data")}</p>
            )}
          </div>
        </div>

        {/* Bond Yields */}
        <div>
          <h2 className="text-lg font-semibold text-foreground mb-3">
            {t("dashboard.bond_yields")}
          </h2>
          <div className="bg-card-bg border border-border-theme rounded-xl p-4">
            {loading && !data ? (
              <div className="h-48 bg-surface rounded animate-pulse" />
            ) : data?.bondYields && data.bondYields.length > 0 ? (
              <div className="h-48">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={data.bondYields} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke={gridStroke} />
                    <XAxis dataKey="tenor" tick={{ fill: tickFill, fontSize: 11 }} />
                    <YAxis tick={{ fill: tickFill, fontSize: 11 }} tickFormatter={(v) => `${v}%`} />
                    <Tooltip contentStyle={tooltipStyle} formatter={(v) => [`${formatNumber(Number(v), 2)}%`, "Yield"]} />
                    <Line type="monotone" dataKey="yield" stroke="#f59e0b" strokeWidth={2} dot={{ r: 3 }} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            ) : (
              <p className="text-sm text-text-muted py-8 text-center">{t("common.no_data")}</p>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
