"use client";

import { useEffect, useState, useMemo } from "react";
import { History, TrendingUp, TrendingDown, Minus, AlertTriangle, ChevronDown, ChevronUp } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

// --- Types ---

interface OutcomeMetric {
  interval: string; // "1d" | "7d" | "14d" | "30d"
  priceChange: number;
  hit: boolean;
}

interface Recommendation {
  id: string;
  symbol: string;
  action: "buy" | "sell" | "hold";
  targetPrice: number;
  confidence: number;
  createdAt: string;
  priceAtRecommendation: number;
  outcomes: OutcomeMetric[];
}

interface RecommendationHistoryResponse {
  recommendations: Recommendation[];
  overallAccuracy: number;
  totalCount: number;
}

// --- Helpers ---

function actionColor(action: string) {
  switch (action) {
    case "buy": return "text-green-400 bg-green-950/40";
    case "sell": return "text-red-400 bg-red-950/40";
    default: return "text-zinc-400 bg-zinc-800";
  }
}

function actionLabel(action: string) {
  switch (action) {
    case "buy": return "Buy";
    case "sell": return "Sell";
    default: return "Hold";
  }
}

// --- Component (Req 32.3) ---

export function RecommendationHistory() {
  const { t, formatCurrency, formatPercent, formatDate } = useI18n();
  const [data, setData] = useState<RecommendationHistoryResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedId, setExpandedId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      const res = await apiFetch<RecommendationHistoryResponse>("/api/recommendations/history");
      if (cancelled) return;
      if (res) {
        setData(res);
        setError(null);
      } else {
        setError("Failed to load recommendation history");
      }
      setLoading(false);
    }
    load();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 animate-pulse">
        <div className="h-6 bg-zinc-800 rounded w-56 mb-4" />
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <div key={i} className="h-20 bg-zinc-800 rounded" />)}
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
        <div className="flex items-center gap-2 text-red-400">
          <AlertTriangle size={16} />
          <span className="text-sm">{error ?? t("recommendations.no_data")}</span>
        </div>
      </div>
    );
  }

  const { recommendations, overallAccuracy, totalCount } = data;

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <History size={18} className="text-indigo-400" />
          <h2 className="text-lg font-semibold text-white">{t("recommendations.title")}</h2>
          <span className="text-xs text-zinc-500">{totalCount} total</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-xs text-zinc-400">{t("recommendations.accuracy")}:</span>
          <span className={`text-sm font-bold ${overallAccuracy >= 50 ? "text-green-400" : "text-red-400"}`}>
            {formatPercent(overallAccuracy)}
          </span>
        </div>
      </div>

      <div className="p-4">
        {recommendations.length === 0 ? (
          <p className="text-zinc-500 text-sm text-center py-8">{t("recommendations.no_data")}</p>
        ) : (
          <div className="space-y-2">
            {recommendations.map((rec) => {
              const isExpanded = expandedId === rec.id;
              return (
                <div key={rec.id} className="border border-zinc-800 rounded-lg bg-zinc-800/40">
                  <button
                    onClick={() => setExpandedId(isExpanded ? null : rec.id)}
                    className="w-full flex items-center justify-between p-3 text-left"
                  >
                    <div className="flex items-center gap-3">
                      <span className={`px-2 py-0.5 rounded text-xs font-bold uppercase ${actionColor(rec.action)}`}>
                        {actionLabel(rec.action)}
                      </span>
                      <span className="text-white font-medium text-sm">{rec.symbol}</span>
                      <span className="text-zinc-500 text-xs">{formatDate(rec.createdAt)}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-zinc-400 text-xs">
                        {t("recommendations.confidence")}: {rec.confidence}%
                      </span>
                      {/* Show best outcome */}
                      {rec.outcomes.length > 0 && (() => {
                        const best = rec.outcomes[rec.outcomes.length - 1];
                        const isHit = best.hit;
                        return (
                          <span className={`flex items-center gap-1 text-xs font-semibold ${isHit ? "text-green-400" : "text-red-400"}`}>
                            {isHit ? <TrendingUp size={12} /> : <TrendingDown size={12} />}
                            {isHit ? t("recommendations.hit") : t("recommendations.miss")}
                          </span>
                        );
                      })()}
                      {isExpanded ? <ChevronUp size={14} className="text-zinc-500" /> : <ChevronDown size={14} className="text-zinc-500" />}
                    </div>
                  </button>

                  {isExpanded && (
                    <div className="border-t border-zinc-700 p-3">
                      <div className="grid grid-cols-2 gap-3 mb-3 text-xs">
                        <div>
                          <span className="text-zinc-500">Price at recommendation</span>
                          <p className="text-zinc-300">{formatCurrency(rec.priceAtRecommendation)}</p>
                        </div>
                        <div>
                          <span className="text-zinc-500">{t("recommendations.target")}</span>
                          <p className="text-zinc-300">{formatCurrency(rec.targetPrice)}</p>
                        </div>
                      </div>
                      {/* Outcome table: 1d, 7d, 14d, 30d */}
                      <table className="w-full text-xs">
                        <thead>
                          <tr className="text-zinc-500">
                            <th className="text-left pb-2">Interval</th>
                            <th className="text-right pb-2">Change</th>
                            <th className="text-right pb-2">{t("recommendations.outcome")}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {rec.outcomes.map((o) => (
                            <tr key={o.interval} className="text-zinc-300 border-t border-zinc-800">
                              <td className="py-1.5 font-mono">{o.interval}</td>
                              <td className={`text-right ${o.priceChange >= 0 ? "text-green-400" : "text-red-400"}`}>
                                {o.priceChange >= 0 ? "+" : ""}{formatPercent(o.priceChange)}
                              </td>
                              <td className="text-right">
                                {o.hit ? (
                                  <span className="text-green-400">{t("recommendations.hit")}</span>
                                ) : (
                                  <span className="text-red-400">{t("recommendations.miss")}</span>
                                )}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
