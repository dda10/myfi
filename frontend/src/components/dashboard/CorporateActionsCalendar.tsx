"use client";

import { useEffect, useState, useMemo } from "react";
import { Calendar, Bell, DollarSign, Users, ChevronDown, ChevronUp, AlertTriangle } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types matching backend model/corporate_action_types.go ---

interface CorporateAction {
  id: number;
  symbol: string;
  actionType: "dividend" | "stock_split" | "bonus_share";
  exDate: string;
  recordDate: string;
  paymentDate: string;
  dividendPerShare?: number;
  splitRatioFrom?: number;
  splitRatioTo?: number;
  description?: string;
}

interface DividendRecord {
  id: number;
  symbol: string;
  exDate: string;
  paymentDate: string;
  dividendPerShare: number;
  sharesHeld: number;
  totalAmount: number;
}

interface DividendHistory {
  symbol: string;
  records: DividendRecord[];
  totalDividends: number;
  yieldOnCost: number;
}

// --- Helpers ---

async function apiFetch<T>(path: string): Promise<T | null> {
  try {
    const token = typeof window !== "undefined" ? localStorage.getItem("ezistock-token") : null;
    const res = await fetch(`${API_URL}${path}`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

function daysUntil(dateStr: string): number {
  const now = new Date();
  now.setHours(0, 0, 0, 0);
  const target = new Date(dateStr);
  target.setHours(0, 0, 0, 0);
  return Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
}

function actionIcon(type: CorporateAction["actionType"]) {
  switch (type) {
    case "dividend": return <DollarSign size={14} className="text-green-400" />;
    case "stock_split": return <ChevronUp size={14} className="text-blue-400" />;
    case "bonus_share": return <Users size={14} className="text-purple-400" />;
  }
}

function actionLabel(type: CorporateAction["actionType"]) {
  switch (type) {
    case "dividend": return "Dividend";
    case "stock_split": return "Stock Split";
    case "bonus_share": return "Bonus Share";
  }
}

// --- Component ---

export function CorporateActionsCalendar() {
  const { formatCurrency, formatDate, formatPercent, t } = useI18n();
  const [actions, setActions] = useState<CorporateAction[]>([]);
  const [dividendHistories, setDividendHistories] = useState<DividendHistory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedSymbol, setExpandedSymbol] = useState<string | null>(null);
  const [tab, setTab] = useState<"upcoming" | "history">("upcoming");

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      const [actionsRes, historyRes] = await Promise.all([
        apiFetch<CorporateAction[]>("/api/market/corporate-actions"),
        apiFetch<DividendHistory[]>("/api/portfolio/dividend-history"),
      ]);
      if (cancelled) return;
      if (!actionsRes && !historyRes) {
        setError("Failed to load corporate actions data");
      }
      setActions(actionsRes ?? []);
      setDividendHistories(historyRes ?? []);
      setLoading(false);
    }
    load();
    return () => { cancelled = true; };
  }, []);

  // Req 30.5: 3-day advance notifications for ex-dividend dates
  const upcomingWithAlerts = useMemo(() => {
    return actions
      .filter((a) => daysUntil(a.exDate) >= 0)
      .sort((a, b) => new Date(a.exDate).getTime() - new Date(b.exDate).getTime())
      .map((a) => ({ ...a, daysUntilEx: daysUntil(a.exDate), isAlert: daysUntil(a.exDate) <= 3 && daysUntil(a.exDate) >= 0 }));
  }, [actions]);

  if (loading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 animate-pulse">
        <div className="h-6 bg-zinc-800 rounded w-48 mb-4" />
        <div className="space-y-3">
          {[1, 2, 3].map((i) => <div key={i} className="h-16 bg-zinc-800 rounded" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
        <div className="flex items-center gap-2 text-red-400">
          <AlertTriangle size={16} />
          <span className="text-sm">{error}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Calendar size={18} className="text-emerald-400" />
          <h2 className="text-lg font-semibold text-white">{t("corporate.title")}</h2>
        </div>
        <div className="flex gap-1 bg-zinc-800 rounded-lg p-0.5">
          {(["upcoming", "history"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-3 py-1 text-xs rounded-md capitalize transition-colors ${
                tab === t ? "bg-zinc-700 text-white" : "text-zinc-400 hover:text-zinc-200"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      <div className="p-4">
        {tab === "upcoming" ? (
          upcomingWithAlerts.length === 0 ? (
            <p className="text-zinc-500 text-sm text-center py-8">{t("corporate.no_upcoming")}</p>
          ) : (
            <div className="space-y-2">
              {upcomingWithAlerts.map((action) => (
                <div
                  key={`${action.symbol}-${action.exDate}-${action.actionType}`}
                  className={`rounded-lg border p-3 ${
                    action.isAlert
                      ? "border-amber-600/50 bg-amber-950/20"
                      : "border-zinc-800 bg-zinc-800/40"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {actionIcon(action.actionType)}
                      <span className="text-white font-medium text-sm">{action.symbol}</span>
                      <span className="text-zinc-500 text-xs">{actionLabel(action.actionType)}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      {action.isAlert && (
                        <span className="flex items-center gap-1 text-amber-400 text-xs">
                          <Bell size={12} />
                          {action.daysUntilEx === 0 ? t("corporate.today") : `${action.daysUntilEx}d`}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="mt-2 grid grid-cols-3 gap-2 text-xs">
                    <div>
                      <span className="text-zinc-500">{t("corporate.ex_date")}</span>
                      <p className="text-zinc-300">{formatDate(action.exDate)}</p>
                    </div>
                    <div>
                      <span className="text-zinc-500">{t("corporate.payment")}</span>
                      <p className="text-zinc-300">{formatDate(action.paymentDate)}</p>
                    </div>
                    <div>
                      <span className="text-zinc-500">{t("corporate.amount")}</span>
                      <p className="text-zinc-300">
                        {action.actionType === "dividend" && action.dividendPerShare
                          ? formatCurrency(action.dividendPerShare) + t("corporate.per_share")
                          : action.actionType === "stock_split"
                            ? `${action.splitRatioFrom}:${action.splitRatioTo}`
                            : "—"}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )
        ) : (
          /* Dividend History Tab — Req 30.7 */
          dividendHistories.length === 0 ? (
            <p className="text-zinc-500 text-sm text-center py-8">{t("corporate.no_history")}</p>
          ) : (
            <div className="space-y-2">
              {dividendHistories.map((h) => (
                <div key={h.symbol} className="border border-zinc-800 rounded-lg bg-zinc-800/40">
                  <button
                    onClick={() => setExpandedSymbol(expandedSymbol === h.symbol ? null : h.symbol)}
                    className="w-full flex items-center justify-between p-3 text-left"
                  >
                    <div className="flex items-center gap-2">
                      <DollarSign size={14} className="text-green-400" />
                      <span className="text-white font-medium text-sm">{h.symbol}</span>
                      <span className="text-zinc-500 text-xs">
                        {h.records.length} payment{h.records.length !== 1 ? "s" : ""}
                      </span>
                    </div>
                    <div className="flex items-center gap-3">
                      <span className="text-emerald-400 text-xs font-medium">
                        {formatCurrency(h.totalDividends)}
                      </span>
                      <span className="text-zinc-500 text-xs">
                        YoC: {formatPercent(h.yieldOnCost)}
                      </span>
                      {expandedSymbol === h.symbol ? (
                        <ChevronUp size={14} className="text-zinc-500" />
                      ) : (
                        <ChevronDown size={14} className="text-zinc-500" />
                      )}
                    </div>
                  </button>
                  {expandedSymbol === h.symbol && (
                    <div className="border-t border-zinc-700 p-3">
                      <table className="w-full text-xs">
                        <thead>
                          <tr className="text-zinc-500">
                            <th className="text-left pb-2">Ex-Date</th>
                            <th className="text-right pb-2">Per Share</th>
                            <th className="text-right pb-2">Shares</th>
                            <th className="text-right pb-2">Total</th>
                          </tr>
                        </thead>
                        <tbody>
                          {h.records.map((r) => (
                            <tr key={r.id} className="text-zinc-300 border-t border-zinc-800">
                              <td className="py-1.5">{formatDate(r.exDate)}</td>
                              <td className="text-right">{formatCurrency(r.dividendPerShare)}</td>
                              <td className="text-right">{r.sharesHeld.toLocaleString()}</td>
                              <td className="text-right text-emerald-400">{formatCurrency(r.totalAmount)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )
        )}
      </div>
    </div>
  );
}
