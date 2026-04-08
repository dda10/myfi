"use client";

import { ArrowUpRight, ArrowDownRight } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
} from "recharts";

// --- Types ---

export interface HoldingRow {
  symbol: string;
  quantity: number;
  averageCost: number;
  currentPrice: number;
  marketValue: number;
  unrealizedPL: number;
  unrealizedPLPct: number;
  sector: string;
}

interface HoldingsProps {
  holdings: HoldingRow[];
}

const SECTOR_COLORS = [
  "#6366f1", "#22c55e", "#f59e0b", "#ef4444", "#8b5cf6",
  "#06b6d4", "#ec4899", "#14b8a6", "#f97316", "#84cc16",
];

export function Holdings({ holdings }: HoldingsProps) {
  const { t, formatCurrency, formatPercent } = useI18n();

  if (holdings.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 bg-zinc-900 border border-zinc-800 rounded-xl text-zinc-500">
        {t("common.no_data")}
      </div>
    );
  }

  // Sector allocation for pie chart
  const sectorMap = new Map<string, number>();
  for (const h of holdings) {
    const key = h.sector || "Other";
    sectorMap.set(key, (sectorMap.get(key) ?? 0) + h.marketValue);
  }
  const pieData = Array.from(sectorMap.entries()).map(([name, value]) => ({ name, value }));

  return (
    <div className="space-y-6">
      {/* Holdings Table */}
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 text-zinc-400 text-left">
                <th className="px-4 py-3 font-medium">{t("table.symbol")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.quantity")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.avg_price")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("table.price")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.total_value")}</th>
                <th className="px-4 py-3 font-medium text-right">{t("portfolio.unrealized_pl")}</th>
                <th className="px-4 py-3 font-medium text-right">P&L %</th>
              </tr>
            </thead>
            <tbody>
              {holdings.map((h) => {
                const positive = h.unrealizedPL >= 0;
                const plColor = positive ? "text-green-400" : "text-red-400";
                return (
                  <tr key={h.symbol} className="border-b border-zinc-800/50 hover:bg-zinc-800/30 transition">
                    <td className="px-4 py-3 font-semibold text-white">{h.symbol}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{h.quantity.toLocaleString()}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(h.averageCost)}</td>
                    <td className="px-4 py-3 text-right text-zinc-300">{formatCurrency(h.currentPrice)}</td>
                    <td className="px-4 py-3 text-right text-white font-medium">{formatCurrency(h.marketValue)}</td>
                    <td className={`px-4 py-3 text-right font-medium ${plColor}`}>
                      {positive ? "+" : ""}{formatCurrency(h.unrealizedPL)}
                    </td>
                    <td className={`px-4 py-3 text-right font-medium ${plColor}`}>
                      <span className="inline-flex items-center gap-0.5">
                        {positive ? <ArrowUpRight size={12} /> : <ArrowDownRight size={12} />}
                        {formatPercent(h.unrealizedPLPct)}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* Sector Allocation Pie */}
      {pieData.length > 0 && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4">
          <h3 className="text-sm font-medium text-zinc-400 mb-3">{t("portfolio.sector_allocation")}</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  dataKey="value"
                  nameKey="name"
                  label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                  labelLine={false}
                >
                  {pieData.map((_, i) => (
                    <Cell key={i} fill={SECTOR_COLORS[i % SECTOR_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                  formatter={(value) => [formatCurrency(Number(value ?? 0)), ""]}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}
    </div>
  );
}
