"use client";

import { useState, useEffect, useCallback } from "react";
import {
  History,
  RefreshCw,
  Search,
  Filter,
  TrendingUp,
  TrendingDown,
  Minus,
  AlertTriangle,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";

// --- Types ---

type ActionType = "buy" | "sell" | "hold" | "";

interface RecommendationRecord {
  id: number;
  symbol: string;
  action: ActionType;
  positionSize: number;
  riskAssessment: string;
  confidenceScore: number;
  reasoning: string;
  priceAtSignal: number;
  createdAt: string;
  price1Day?: number;
  price7Day?: number;
  price14Day?: number;
  price30Day?: number;
  return1Day?: number;
  return7Day?: number;
  return14Day?: number;
  return30Day?: number;
}

interface Filters {
  symbol: string;
  action: ActionType;
  startDate: string;
  endDate: string;
}

// --- Helpers ---

function isCorrect(action: string, returnPct: number | undefined): boolean | null {
  if (returnPct == null) return null;
  if (action === "buy") return returnPct > 0;
  if (action === "sell") return returnPct < 0;
  return null; // hold — no correctness judgment
}

function outcomeColor(action: string, returnPct: number | undefined): string {
  const correct = isCorrect(action, returnPct);
  if (correct === null) return "text-zinc-400";
  return correct ? "text-emerald-600" : "text-red-500";
}

// --- Component ---

export function RecommendationHistory() {
  const { formatDate, formatNumber, formatPercent } = useI18n();

  const [records, setRecords] = useState<RecommendationRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<Filters>({
    symbol: "",
    action: "",
    startDate: "",
    endDate: "",
  });
  const [showFilters, setShowFilters] = useState(false);

  const fetchHistory = useCallback(async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (filters.symbol) params.set("symbol", filters.symbol.toUpperCase());
      if (filters.action) params.set("action", filters.action);
      params.set("limit", "200");

      const res = await fetch(
        `http://localhost:8080/api/recommendations?${params.toString()}`
      );
      if (!res.ok) throw new Error("fetch failed");
      const json = await res.json();
      let recs: RecommendationRecord[] = json.recommendations ?? [];

      // Client-side date filtering
      if (filters.startDate) {
        const start = new Date(filters.startDate).getTime();
        recs = recs.filter((r) => new Date(r.createdAt).getTime() >= start);
      }
      if (filters.endDate) {
        const end = new Date(filters.endDate).getTime() + 86400000; // include end day
        recs = recs.filter((r) => new Date(r.createdAt).getTime() < end);
      }

      setRecords(recs);
    } catch (e) {
      console.error("Failed to fetch recommendation history", e);
      setRecords([]);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  const actionIcon = (action: string) => {
    if (action === "buy") return <TrendingUp size={14} className="text-emerald-600" />;
    if (action === "sell") return <TrendingDown size={14} className="text-red-500" />;
    return <Minus size={14} className="text-zinc-400" />;
  };

  const actionBadge = (action: string) => {
    const colors: Record<string, string> = {
      buy: "bg-emerald-100 text-emerald-700",
      sell: "bg-red-100 text-red-700",
      hold: "bg-zinc-100 text-zinc-600",
    };
    return (
      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold ${colors[action] ?? "bg-zinc-100 text-zinc-600"}`}>
        {actionIcon(action)}
        {action.toUpperCase()}
      </span>
    );
  };

  const renderReturn = (action: string, ret: number | undefined) => {
    if (ret == null) return <span className="text-zinc-300">—</span>;
    const color = outcomeColor(action, ret);
    return (
      <span className={`font-medium ${color}`}>
        {ret >= 0 ? "+" : ""}
        {formatPercent(ret)}
      </span>
    );
  };

  return (
    <div className="flex flex-col gap-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
        <h3 className="text-lg font-bold text-zinc-800 flex items-center gap-2">
          <History size={20} className="text-blue-600" />
          Recommendation History
        </h3>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition ${
              showFilters
                ? "bg-blue-100 text-blue-700"
                : "bg-zinc-100 text-zinc-600 hover:bg-zinc-200"
            }`}
          >
            <Filter size={14} />
            Filters
          </button>
          <button
            onClick={fetchHistory}
            disabled={loading}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white rounded-lg text-sm font-medium transition"
          >
            <RefreshCw size={14} className={loading ? "animate-spin" : ""} />
            Refresh
          </button>
        </div>
      </div>

      {/* Filter Controls */}
      {showFilters && (
        <div className="bg-zinc-50 rounded-xl p-4 grid grid-cols-2 md:grid-cols-4 gap-3">
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Symbol</label>
            <div className="relative">
              <Search size={14} className="absolute left-2.5 top-2.5 text-zinc-400" />
              <input
                type="text"
                placeholder="e.g. FPT"
                value={filters.symbol}
                onChange={(e) =>
                  setFilters((f) => ({ ...f, symbol: e.target.value }))
                }
                className="w-full pl-8 pr-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Action</label>
            <select
              value={filters.action}
              onChange={(e) =>
                setFilters((f) => ({
                  ...f,
                  action: e.target.value as ActionType,
                }))
              }
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All</option>
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
              <option value="hold">Hold</option>
            </select>
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">From</label>
            <input
              type="date"
              value={filters.startDate}
              onChange={(e) =>
                setFilters((f) => ({ ...f, startDate: e.target.value }))
              }
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">To</label>
            <input
              type="date"
              value={filters.endDate}
              onChange={(e) =>
                setFilters((f) => ({ ...f, endDate: e.target.value }))
              }
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
      )}

      {/* Table */}
      {loading ? (
        <div className="flex items-center justify-center h-48 text-zinc-500">
          <RefreshCw className="animate-spin mr-2" size={18} /> Loading history...
        </div>
      ) : records.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-48 text-zinc-500">
          <AlertTriangle size={36} className="mb-3 text-zinc-300" />
          <p className="text-sm">No recommendation history found.</p>
        </div>
      ) : (
        <div className="overflow-x-auto rounded-xl border border-zinc-200">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-zinc-50 text-zinc-600 text-xs uppercase tracking-wider">
                <th className="px-4 py-3 text-left">Date</th>
                <th className="px-4 py-3 text-left">Symbol</th>
                <th className="px-4 py-3 text-left">Action</th>
                <th className="px-4 py-3 text-right">Confidence</th>
                <th className="px-4 py-3 text-right">Entry Price</th>
                <th className="px-4 py-3 text-right">1D</th>
                <th className="px-4 py-3 text-right">7D</th>
                <th className="px-4 py-3 text-right">14D</th>
                <th className="px-4 py-3 text-right">30D</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-100">
              {records.map((rec) => (
                <tr key={rec.id} className="hover:bg-zinc-50 transition">
                  <td className="px-4 py-3 text-zinc-700 whitespace-nowrap">
                    {formatDate(rec.createdAt)}
                  </td>
                  <td className="px-4 py-3 font-semibold text-zinc-900">
                    {rec.symbol}
                  </td>
                  <td className="px-4 py-3">{actionBadge(rec.action)}</td>
                  <td className="px-4 py-3 text-right">
                    <span
                      className={`font-medium ${
                        rec.confidenceScore >= 70
                          ? "text-emerald-600"
                          : rec.confidenceScore >= 40
                          ? "text-blue-600"
                          : "text-zinc-500"
                      }`}
                    >
                      {rec.confidenceScore}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-right text-zinc-700">
                    {formatNumber(rec.priceAtSignal, 0)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {renderReturn(rec.action, rec.return1Day)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {renderReturn(rec.action, rec.return7Day)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {renderReturn(rec.action, rec.return14Day)}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {renderReturn(rec.action, rec.return30Day)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Record count */}
      {!loading && records.length > 0 && (
        <div className="text-xs text-zinc-500 text-right">
          Showing {records.length} recommendation{records.length !== 1 ? "s" : ""}
        </div>
      )}
    </div>
  );
}
