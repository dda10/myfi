"use client";

import { useState, useEffect } from "react";
import { Landmark } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from "recharts";

type FlowClass = "strong_inflow" | "inflow" | "neutral" | "outflow" | "strong_outflow";

interface FlowData {
  foreignNetBuy: number;
  institutionalNetBuy: number;
  flowClassification: FlowClass;
  history: { date: string; foreign: number; institutional: number }[];
}

const FLOW_COLORS: Record<FlowClass, { color: string; label: string }> = {
  strong_inflow: { color: "text-green-400", label: "Strong Inflow" },
  inflow: { color: "text-green-500", label: "Inflow" },
  neutral: { color: "text-yellow-500", label: "Neutral" },
  outflow: { color: "text-red-500", label: "Outflow" },
  strong_outflow: { color: "text-red-400", label: "Strong Outflow" },
};

export function SmartMoneyFlow({ symbol }: { symbol: string }) {
  const { t, formatNumber } = useI18n();
  const [data, setData] = useState<FlowData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<FlowData>(`/api/market/moneyflow/${symbol}`)
      .then((res) => { if (res) setData(res); })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/3 bg-surface rounded" />
        <div className="h-40 bg-surface rounded" />
      </div>
    );
  }

  if (!data) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 text-text-muted text-sm">
        {t("common.no_data")}
      </div>
    );
  }

  const flowCfg = FLOW_COLORS[data.flowClassification] ?? FLOW_COLORS.neutral;

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          <Landmark size={20} className="text-amber-400" />
          <h3 className="text-lg font-semibold text-foreground">Smart Money Flow</h3>
        </div>
        <span className={`text-xs font-semibold px-2.5 py-1 rounded-full bg-surface ${flowCfg.color}`}>
          {flowCfg.label}
        </span>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="bg-surface rounded-lg p-3">
          <p className="text-xs text-text-muted mb-1">Foreign Net</p>
          <p className={`text-sm font-bold ${data.foreignNetBuy >= 0 ? "text-green-500" : "text-red-500"}`}>
            {data.foreignNetBuy >= 0 ? "+" : ""}{formatNumber(data.foreignNetBuy / 1e9, 2)}B
          </p>
        </div>
        <div className="bg-surface rounded-lg p-3">
          <p className="text-xs text-text-muted mb-1">Institutional Net</p>
          <p className={`text-sm font-bold ${data.institutionalNetBuy >= 0 ? "text-green-500" : "text-red-500"}`}>
            {data.institutionalNetBuy >= 0 ? "+" : ""}{formatNumber(data.institutionalNetBuy / 1e9, 2)}B
          </p>
        </div>
      </div>

      {data.history.length > 0 && (
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={data.history} margin={{ top: 5, right: 5, bottom: 5, left: 5 }}>
              <XAxis dataKey="date" tick={{ fontSize: 10, fill: "#9ca3af" }} tickLine={false} axisLine={false} />
              <YAxis tick={{ fontSize: 10, fill: "#9ca3af" }} tickLine={false} axisLine={false} tickFormatter={(v: number) => `${(v / 1e9).toFixed(0)}B`} />
              <Tooltip
                contentStyle={{ backgroundColor: "#1f2937", border: "1px solid #374151", borderRadius: 8, fontSize: 12 }}
                labelStyle={{ color: "#9ca3af" }}
              />
              <Bar dataKey="foreign" name="Foreign" radius={[2, 2, 0, 0]}>
                {data.history.map((entry, i) => (
                  <Cell key={i} fill={entry.foreign >= 0 ? "#22c55e" : "#ef4444"} />
                ))}
              </Bar>
              <Bar dataKey="institutional" name="Institutional" radius={[2, 2, 0, 0]}>
                {data.history.map((entry, i) => (
                  <Cell key={i} fill={entry.institutional >= 0 ? "#3b82f6" : "#f97316"} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>
      )}
    </section>
  );
}
