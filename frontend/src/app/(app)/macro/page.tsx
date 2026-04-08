"use client";

import { useState, useEffect } from "react";
import { TrendingUp, TrendingDown, Minus, DollarSign, BarChart3, Activity, Gem } from "lucide-react";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from "recharts";
import { useI18n } from "@/context/I18nContext";
import { useTheme } from "@/context/ThemeContext";
import { apiFetch } from "@/lib/api";

interface MacroData {
  interbankRates: { term: string; rate: number }[];
  bondYields: { term: string; yield: number }[];
  fxRates: { pair: string; rate: number; change: number; history: { date: string; rate: number }[] }[];
  cpi: { period: string; value: number }[];
  gdpGrowth: { period: string; value: number }[];
}

interface ExchangeRate {
  currency: string;
  currencyName: string;
  buy: number;
  transfer: number;
  sell: number;
}

interface GoldPrice {
  brand: string;
  type: string;
  buy: number;
  sell: number;
  change: number;
  changePercent: number;
  updatedAt: string;
}

export default function MacroPage() {
  const { t, formatNumber } = useI18n();
  const { chartColors } = useTheme();
  const [data, setData] = useState<MacroData | null>(null);
  const [exchangeRates, setExchangeRates] = useState<ExchangeRate[] | null>(null);
  const [goldPrices, setGoldPrices] = useState<GoldPrice[] | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      apiFetch<MacroData>("/api/market/macro"),
      apiFetch<ExchangeRate[]>("/api/market/exchange-rates"),
      apiFetch<GoldPrice[]>("/api/market/gold"),
    ])
      .then(([macroRes, fxRes, goldRes]) => {
        if (macroRes) setData(macroRes);
        if (fxRes) setExchangeRates(fxRes);
        if (goldRes) setGoldPrices(goldRes);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-foreground">{t("macro.title")}</h1>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="bg-card-bg border border-border-theme rounded-xl p-6 h-72 animate-pulse">
              <div className="h-5 w-1/3 bg-surface rounded mb-4" />
              <div className="h-48 bg-surface rounded" />
            </div>
          ))}
        </div>
        {/* Loading skeletons for exchange rates and gold */}
        {[1, 2].map((i) => (
          <div key={`full-${i}`} className="bg-card-bg border border-border-theme rounded-xl p-6 h-48 animate-pulse">
            <div className="h-5 w-1/4 bg-surface rounded mb-4" />
            <div className="h-32 bg-surface rounded" />
          </div>
        ))}
      </div>
    );
  }

  const chartStyle = { fontSize: 11, fill: chartColors.text };

  return (
    <div className="space-y-6 animate-fade-in">
      <div>
        <h1 className="text-xl font-bold text-foreground">{t("macro.title")}</h1>
        <p className="text-xs text-text-muted mt-0.5">Theo dõi các chỉ tiêu kinh tế vĩ mô Việt Nam</p>
      </div>

      {/* Economic Cycle Prediction Widget — MiQuant style */}
      <div className="grid grid-cols-1 lg:grid-cols-[1fr_300px] gap-4">
        {/* Macro indicators color-coded table */}
        <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
          <div className="px-4 py-3 border-b border-border-theme">
            <h3 className="text-sm font-bold text-foreground">Chỉ tiêu kinh tế vĩ mô</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-surface/50 border-b border-border-theme">
                <tr>
                  {["Chỉ tiêu", "Giá trị hiện tại", "Tháng trước", "YoY", "Trạng thái"].map(h => (
                    <th key={h} className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border-theme">
                {[
                  { name: "GDP (tăng trưởng)", value: "7.09%", prev: "6.72%", yoy: "+0.37pp", status: "Tốt",    color: "text-green-400 bg-green-500/10" },
                  { name: "CPI (lạm phát)",     value: "3.21%", prev: "3.18%", yoy: "+0.03pp", status: "Ổn định", color: "text-yellow-400 bg-yellow-500/10" },
                  { name: "PMI sản xuất",       value: "52.4",  prev: "50.8",  yoy: "+1.6",    status: "Mở rộng", color: "text-green-400 bg-green-500/10" },
                  { name: "FDI (tỷ USD)",       value: "4.51",  prev: "4.20",  yoy: "+7.4%",   status: "Tốt",    color: "text-green-400 bg-green-500/10" },
                  { name: "Xuất khẩu (tỷ USD)", value: "36.2",  prev: "34.1",  yoy: "+6.2%",   status: "Tốt",    color: "text-green-400 bg-green-500/10" },
                  { name: "Tỷ lệ thất nghiệp",  value: "2.27%", prev: "2.31%", yoy: "-0.04pp", status: "Tốt",    color: "text-green-400 bg-green-500/10" },
                  { name: "Lãi suất OMO",        value: "4.50%", prev: "4.50%", yoy: "0%",      status: "Ổn định", color: "text-yellow-400 bg-yellow-500/10" },
                ].map(row => (
                  <tr key={row.name} className="hover:bg-surface transition">
                    <td className="px-4 py-2.5 font-medium text-foreground">{row.name}</td>
                    <td className="px-4 py-2.5 font-bold text-foreground tabular-nums">{row.value}</td>
                    <td className="px-4 py-2.5 text-text-muted tabular-nums">{row.prev}</td>
                    <td className="px-4 py-2.5 tabular-nums">
                      <span className={row.yoy.startsWith("+") || row.yoy.startsWith("-0.04") ? "text-green-400" : row.yoy.startsWith("-") ? "text-red-400" : "text-text-muted"}>
                        {row.yoy}
                      </span>
                    </td>
                    <td className="px-4 py-2.5">
                      <span className={`text-[10px] font-semibold px-2 py-0.5 rounded ${row.color}`}>{row.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Economic Cycle Prediction widget */}
        <div className="bg-card-bg border border-border-theme rounded-xl p-4">
          <h3 className="text-sm font-bold text-foreground mb-3">Chu kỳ Kinh tế</h3>
          <p className="text-[11px] text-text-muted mb-4">Dự báo xác suất pha chu kỳ kinh tế bởi AI</p>

          {/* Cycle phases */}
          {[
            { label: "Mở rộng",   prob: 62, color: "bg-emerald-500", textColor: "text-emerald-400" },
            { label: "Đỉnh",      prob: 18, color: "bg-yellow-500",  textColor: "text-yellow-400"  },
            { label: "Thu hẹp",   prob: 12, color: "bg-red-500",     textColor: "text-red-400"     },
            { label: "Phục hồi",  prob: 8,  color: "bg-blue-500",    textColor: "text-blue-400"    },
          ].map(phase => (
            <div key={phase.label} className="mb-3">
              <div className="flex items-center justify-between mb-1">
                <span className={`text-xs font-semibold ${phase.textColor}`}>{phase.label}</span>
                <span className="text-xs font-bold text-foreground">{phase.prob}%</span>
              </div>
              <div className="h-2 bg-surface-hover rounded-full overflow-hidden">
                <div
                  className={`h-full ${phase.color} rounded-full transition-all duration-1000`}
                  style={{ width: `${phase.prob}%` }}
                />
              </div>
            </div>
          ))}

          {/* Current phase badge */}
          <div className="mt-4 p-3 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
            <p className="text-[10px] text-text-muted mb-1">Pha hiện tại</p>
            <p className="text-sm font-black text-emerald-400">🟢 Mở rộng</p>
            <p className="text-[11px] text-text-muted mt-1">Kinh tế đang trong giai đoạn tăng trưởng với xác suất 62%</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Interbank Rates Yield Curve */}
        <div className="bg-card-bg border border-border-theme rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <Activity size={18} className="text-indigo-400" />
            <h3 className="font-semibold text-foreground">{t("macro.interbank_rate")}</h3>
          </div>
          {data?.interbankRates && data.interbankRates.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={data.interbankRates}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartColors.grid} />
                <XAxis dataKey="term" tick={chartStyle} />
                <YAxis tick={chartStyle} tickFormatter={(v) => `${v}%`} />
                <Tooltip contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.grid}`, color: chartColors.text }} />
                <Line type="monotone" dataKey="rate" stroke="#6366f1" strokeWidth={2} dot={{ r: 4 }} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-text-muted text-center py-12">{t("common.no_data")}</p>
          )}
        </div>

        {/* Bond Yields Yield Curve */}
        <div className="bg-card-bg border border-border-theme rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <BarChart3 size={18} className="text-green-400" />
            <h3 className="font-semibold text-foreground">{t("macro.bond_yield")}</h3>
          </div>
          {data?.bondYields && data.bondYields.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={data.bondYields}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartColors.grid} />
                <XAxis dataKey="term" tick={chartStyle} />
                <YAxis tick={chartStyle} tickFormatter={(v) => `${v}%`} />
                <Tooltip contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.grid}`, color: chartColors.text }} />
                <Line type="monotone" dataKey="yield" stroke="#22c55e" strokeWidth={2} dot={{ r: 4 }} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-text-muted text-center py-12">{t("common.no_data")}</p>
          )}
        </div>

        {/* FX Rate Trends */}
        <div className="bg-card-bg border border-border-theme rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <DollarSign size={18} className="text-yellow-400" />
            <h3 className="font-semibold text-foreground">{t("macro.fx_rate")}</h3>
          </div>
          {data?.fxRates && data.fxRates.length > 0 ? (
            <div className="space-y-3">
              {data.fxRates.map((fx, idx) => (
                <div key={idx} className="flex items-center justify-between bg-surface rounded-lg px-3 py-2">
                  <span className="text-sm font-medium text-foreground">{fx.pair}</span>
                  <div className="text-right">
                    <span className="text-sm font-bold text-foreground">{formatNumber(fx.rate, 2)}</span>
                    <span className={`text-xs ml-2 ${fx.change >= 0 ? "text-green-400" : "text-red-400"}`}>
                      {fx.change >= 0 ? "+" : ""}{formatNumber(fx.change, 2)}%
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-text-muted text-center py-12">{t("common.no_data")}</p>
          )}
        </div>

        {/* CPI & GDP Growth */}
        <div className="bg-card-bg border border-border-theme rounded-xl p-6">
          <div className="flex items-center gap-2 mb-4">
            <TrendingUp size={18} className="text-purple-400" />
            <h3 className="font-semibold text-foreground">{t("macro.cpi")} & {t("macro.gdp_growth")}</h3>
          </div>
          {data?.cpi && data.cpi.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <LineChart data={data.cpi.map((c, i) => ({
                period: c.period,
                cpi: c.value,
                gdp: data.gdpGrowth?.[i]?.value,
              }))}>
                <CartesianGrid strokeDasharray="3 3" stroke={chartColors.grid} />
                <XAxis dataKey="period" tick={chartStyle} />
                <YAxis tick={chartStyle} tickFormatter={(v) => `${v}%`} />
                <Tooltip contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.grid}`, color: chartColors.text }} />
                <Legend wrapperStyle={{ fontSize: 11 }} />
                <Line type="monotone" dataKey="cpi" name={t("macro.cpi")} stroke="#a855f7" strokeWidth={2} dot={{ r: 3 }} />
                <Line type="monotone" dataKey="gdp" name={t("macro.gdp_growth")} stroke="#f59e0b" strokeWidth={2} dot={{ r: 3 }} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-sm text-text-muted text-center py-12">{t("common.no_data")}</p>
          )}
        </div>
      </div>

      {/* VCB Exchange Rates — full width */}
      <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
        <div className="px-4 py-3 border-b border-border-theme flex items-center gap-2">
          <DollarSign size={18} className="text-blue-400" />
          <h3 className="text-sm font-bold text-foreground">{t("macro.exchange_rates")}</h3>
        </div>
        {exchangeRates && exchangeRates.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-surface/50 border-b border-border-theme">
                <tr>
                  {[t("macro.currency"), "", t("macro.buy_rate"), t("macro.transfer_rate"), t("macro.sell_rate")].map((h, idx) => (
                    <th key={idx} className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border-theme">
                {exchangeRates.map((rate) => (
                  <tr key={rate.currency} className="hover:bg-surface transition">
                    <td className="px-4 py-2.5 font-bold text-foreground">{rate.currency}</td>
                    <td className="px-4 py-2.5 text-text-muted">{rate.currencyName}</td>
                    <td className="px-4 py-2.5 tabular-nums text-foreground">{formatNumber(rate.buy, 0)}</td>
                    <td className="px-4 py-2.5 tabular-nums text-foreground">{formatNumber(rate.transfer, 0)}</td>
                    <td className="px-4 py-2.5 tabular-nums text-foreground">{formatNumber(rate.sell, 0)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-text-muted text-center py-8">{t("common.no_data")}</p>
        )}
      </div>

      {/* Gold Prices — full width */}
      <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
        <div className="px-4 py-3 border-b border-border-theme flex items-center gap-2">
          <Gem size={18} className="text-yellow-500" />
          <h3 className="text-sm font-bold text-foreground">{t("macro.gold")}</h3>
        </div>
        {goldPrices && goldPrices.length > 0 ? (
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-surface/50 border-b border-border-theme">
                <tr>
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider">Brand</th>
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider">Type</th>
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider">{t("macro.buy_rate")}</th>
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider">{t("macro.sell_rate")}</th>
                  <th className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider">+/-</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-theme">
                {goldPrices.map((g) => {
                  const isUp = g.change > 0;
                  const isDown = g.change < 0;
                  const ChangeIcon = isUp ? TrendingUp : isDown ? TrendingDown : Minus;
                  const changeColor = isUp ? "text-green-400" : isDown ? "text-red-400" : "text-text-muted";
                  return (
                    <tr key={`${g.brand}-${g.type}`} className="hover:bg-surface transition">
                      <td className="px-4 py-2.5 font-bold text-foreground">{g.brand}</td>
                      <td className="px-4 py-2.5 text-foreground">{g.type}</td>
                      <td className="px-4 py-2.5 tabular-nums text-foreground">{formatNumber(g.buy, 0)}</td>
                      <td className="px-4 py-2.5 tabular-nums text-foreground">{formatNumber(g.sell, 0)}</td>
                      <td className={`px-4 py-2.5 tabular-nums font-medium ${changeColor}`}>
                        <span className="flex items-center gap-1">
                          <ChangeIcon size={12} />
                          {isUp ? "+" : ""}{formatNumber(g.change, 0)}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        ) : (
          <p className="text-sm text-text-muted text-center py-8">{t("common.no_data")}</p>
        )}
      </div>
    </div>
  );
}
