"use client";

import { Clock, Droplets, Target, TrendingUp, TrendingDown, AlertCircle, Star, BarChart2 } from "lucide-react";

interface Props {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  activeView: string;
  onViewChange: (view: string) => void;
  liquidityTier?: number;
  tradabilityScore?: number;
  // MiQuant AI metrics
  targetPrice?: number;
  upsidePercent?: number;
  riskLevel?: "Thấp" | "Trung bình" | "Cao";
  recommendation?: "Mua mạnh" | "Mua" | "Phù hợp TT" | "Bán" | "Bán mạnh";
  sentimentScore?: number;
}

const RECOMMENDATION_COLORS: Record<string, string> = {
  "Mua mạnh":     "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  "Mua":          "bg-green-500/20 text-green-400 border-green-500/30",
  "Phù hợp TT":  "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
  "Bán":          "bg-red-500/20 text-red-400 border-red-500/30",
  "Bán mạnh":     "bg-rose-500/20 text-rose-400 border-rose-500/30",
};

const RISK_COLORS: Record<string, string> = {
  "Thấp":       "text-green-400 bg-green-500/10 border-green-500/20",
  "Trung bình": "text-yellow-400 bg-yellow-500/10 border-yellow-500/20",
  "Cao":        "text-red-400 bg-red-500/10 border-red-500/20",
};

const TABS = ["Overview", "Financials", "News", "Technicals", "AI Thesis", "Smart Money", "Company"];

export function StockDetailHeader({
  symbol, name, price, change, changePercent,
  activeView, onViewChange,
  liquidityTier, tradabilityScore,
  targetPrice, upsidePercent, riskLevel, recommendation, sentimentScore,
}: Props) {
  const isPositive = change >= 0;

  const tierConfig: Record<number, { color: string; label: string }> = {
    1: { color: "bg-green-500/20 text-green-400 border-green-500/30", label: "Tier 1" },
    2: { color: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30", label: "Tier 2" },
    3: { color: "bg-red-500/20 text-red-400 border-red-500/30", label: "Tier 3" },
  };

  const upside = upsidePercent ?? (targetPrice && price > 0 ? ((targetPrice - price) / price) * 100 : undefined);

  return (
    <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden shadow-lg mb-6 animate-fade-in">
      {/* Top Header */}
      <div className="p-5">
        <div className="flex items-start gap-4 flex-wrap">
          {/* Left: symbol + price */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-2 flex-wrap">
              <div className="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-600 to-purple-600 flex items-center justify-center font-bold text-white text-base border border-indigo-500/30 flex-shrink-0">
                {symbol.substring(0, 1)}
              </div>
              <div>
                <h1 className="text-base font-bold text-foreground flex items-center gap-2 flex-wrap">
                  {symbol}
                  <span className="text-[10px] font-medium bg-badge-bg text-badge-text px-2 py-0.5 rounded">HOSE</span>
                  {liquidityTier && tierConfig[liquidityTier] && (
                    <span className={`text-[10px] font-medium px-2 py-0.5 rounded border ${tierConfig[liquidityTier].color}`}>
                      <Droplets size={9} className="inline mr-0.5" />
                      {tierConfig[liquidityTier].label}
                      {tradabilityScore !== undefined && <span className="ml-1 opacity-60">({tradabilityScore})</span>}
                    </span>
                  )}
                </h1>
                <p className="text-xs text-text-muted">{name}</p>
              </div>
            </div>

            <div className="flex items-baseline gap-3 mt-1">
              <span className="text-3xl font-black text-foreground tabular-nums">
                {price.toLocaleString()} <span className="text-base text-text-muted font-medium">VND</span>
              </span>
              <span className={`text-sm font-bold ${isPositive ? "text-positive" : "text-negative"} flex items-center gap-1`}>
                {isPositive ? <TrendingUp size={14} /> : <TrendingDown size={14} />}
                {isPositive ? "+" : ""}{change.toLocaleString()} ({isPositive ? "+" : ""}{changePercent.toFixed(2)}%)
              </span>
            </div>

            <div className="flex items-center gap-1.5 mt-1 text-[11px] text-text-muted">
              <Clock size={11} />
              <span>Đang giao dịch · Cập nhật {new Date().toLocaleTimeString("vi-VN", { hour: "2-digit", minute: "2-digit" })}</span>
            </div>
          </div>
        </div>

        {/* MiQuant AI metrics row */}
        <div className="mt-4 p-3 bg-surface rounded-xl border border-border-theme grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
          {/* Target Price */}
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] text-text-muted flex items-center gap-1"><Target size={10} />Giá mục tiêu AI</span>
            <span className="text-sm font-bold text-foreground tabular-nums">
              {targetPrice ? targetPrice.toLocaleString() : "—"}
            </span>
          </div>

          {/* Upside */}
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] text-text-muted">Upside</span>
            <span className={`text-sm font-black tabular-nums ${
              upside === undefined ? "text-text-muted" : upside >= 30 ? "text-emerald-400" : upside >= 0 ? "text-green-400" : "text-red-400"
            }`}>
              {upside !== undefined ? `${upside >= 0 ? "+" : ""}${upside.toFixed(2)}%` : "—"}
            </span>
          </div>

          {/* Risk Level */}
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] text-text-muted flex items-center gap-1"><AlertCircle size={10} />Mức độ rủi ro</span>
            {riskLevel ? (
              <span className={`text-[11px] font-semibold px-2 py-0.5 rounded border w-fit ${RISK_COLORS[riskLevel] ?? "text-text-muted"}`}>
                {riskLevel}
              </span>
            ) : (
              <span className="text-sm font-bold text-text-muted">—</span>
            )}
          </div>

          {/* Recommendation */}
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] text-text-muted flex items-center gap-1"><Star size={10}/>Khuyến nghị</span>
            {recommendation ? (
              <span className={`text-[11px] font-bold px-2 py-0.5 rounded border w-fit ${RECOMMENDATION_COLORS[recommendation] ?? "text-text-muted"}`}>
                {recommendation}
              </span>
            ) : (
              <span className="text-sm font-bold text-text-muted">—</span>
            )}
          </div>

          {/* Sentiment Score */}
          <div className="flex flex-col gap-0.5">
            <span className="text-[10px] text-text-muted flex items-center gap-1"><BarChart2 size={10} />Sentiment</span>
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-foreground">
                {sentimentScore !== undefined ? sentimentScore : "—"}
              </span>
              {sentimentScore !== undefined && (
                <div className="flex-1 h-1.5 bg-surface-hover rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full ${sentimentScore >= 70 ? "bg-green-500" : sentimentScore >= 40 ? "bg-yellow-500" : "bg-red-500"}`}
                    style={{ width: `${Math.min(sentimentScore, 100)}%` }}
                  />
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex px-5 border-t border-border-theme bg-surface/30 overflow-x-auto no-scrollbar">
        {TABS.map(tab => (
          <button
            key={tab}
            onClick={() => onViewChange(tab)}
            className={`py-3 px-1 mr-5 text-xs font-medium border-b-2 whitespace-nowrap outline-none transition-colors ${
              activeView === tab
                ? "text-foreground border-indigo-500"
                : "text-text-muted border-transparent hover:text-foreground hover:border-border-theme"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>
    </div>
  );
}
