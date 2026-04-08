"use client";

import { useCallback } from "react";
import { ChevronUp, ChevronDown, ChevronLeft, ChevronRight } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

// --- Types ---

export interface ScreenerRow {
  symbol: string;
  name: string;
  price: number;
  changePct: number;
  sector: string;
  sectorName: string;
  pe: number;
  pb: number;
  roe: number;
  liquidityTier: number;
  tradabilityScore: number;
  exchange: string;
}

interface ResultsTableProps {
  rows: ScreenerRow[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
  sortBy: string;
  sortOrder: string;
  loading: boolean;
  onSort: (field: string) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (size: number) => void;
  onRowClick?: (symbol: string) => void;
}

const PAGE_SIZES = [10, 20, 50];

export function ResultsTable({
  rows,
  total,
  page,
  pageSize,
  totalPages,
  sortBy,
  sortOrder,
  loading,
  onSort,
  onPageChange,
  onPageSizeChange,
  onRowClick,
}: ResultsTableProps) {
  const { t, formatNumber, formatPercent } = useI18n();

  const changePctColor = (v: number) => (v >= 0 ? "text-green-400" : "text-red-400");

  const tierLabel = (tier: number) => {
    if (tier === 1) return <span className="text-green-400 font-semibold">T1</span>;
    if (tier === 2) return <span className="text-yellow-400 font-semibold">T2</span>;
    return <span className="text-red-400 font-semibold">T3</span>;
  };

  return (
    <div className="flex-1 flex flex-col min-w-0 bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between">
        <span className="text-sm text-zinc-400">
          {t("screener.results")}: <span className="text-white font-bold">{total}</span>
        </span>
        <div className="flex items-center gap-2 text-xs text-zinc-500">
          <select
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
            className="bg-zinc-800 border border-zinc-700 rounded px-1.5 py-0.5 text-zinc-300 focus:outline-none"
          >
            {PAGE_SIZES.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Table */}
      <div className="flex-1 overflow-auto">
        <table className="w-full text-xs text-left whitespace-nowrap">
          <thead className="bg-zinc-800/60 text-zinc-400 font-semibold sticky top-0 z-10">
            <tr>
              <th className="px-3 py-3 text-left">{t("table.symbol")}</th>
              <th className="px-3 py-3 text-left">{t("table.name")}</th>
              <SortHeader label={t("table.price")} field="price" current={sortBy} order={sortOrder} onSort={onSort} />
              <SortHeader label={t("table.change_pct")} field="changePct" current={sortBy} order={sortOrder} onSort={onSort} />
              <th className="px-3 py-3 text-left">{t("table.sector")}</th>
              <SortHeader label={t("finance.pe")} field="pe" current={sortBy} order={sortOrder} onSort={onSort} />
              <SortHeader label={t("finance.pb")} field="pb" current={sortBy} order={sortOrder} onSort={onSort} />
              <SortHeader label={t("finance.roe")} field="roe" current={sortBy} order={sortOrder} onSort={onSort} />
              <th className="px-3 py-3 text-center">{t("screener.liquidity_tier")}</th>
              <SortHeader label="Score" field="tradabilityScore" current={sortBy} order={sortOrder} onSort={onSort} />
            </tr>
          </thead>
          <tbody className="divide-y divide-zinc-800/50">
            {loading ? (
              <tr>
                <td colSpan={10} className="text-center py-16 text-zinc-500">{t("common.loading")}</td>
              </tr>
            ) : rows.length === 0 ? (
              <tr>
                <td colSpan={10} className="text-center py-16 text-zinc-500">{t("screener.no_results")}</td>
              </tr>
            ) : (
              rows.map((row) => (
                <tr
                  key={row.symbol}
                  onClick={() => onRowClick?.(row.symbol)}
                  className="hover:bg-zinc-800/40 transition cursor-pointer"
                >
                  <td className="px-3 py-2.5 font-bold text-white">{row.symbol}</td>
                  <td className="px-3 py-2.5 text-zinc-400 max-w-[120px] truncate">{row.name}</td>
                  <td className="px-3 py-2.5 text-right text-zinc-300">{formatNumber(row.price, 0)}</td>
                  <td className={`px-3 py-2.5 text-right font-medium ${changePctColor(row.changePct)}`}>
                    {row.changePct >= 0 ? "+" : ""}{formatPercent(row.changePct)}
                  </td>
                  <td className="px-3 py-2.5 text-zinc-400">{row.sectorName || row.sector}</td>
                  <td className="px-3 py-2.5 text-right text-zinc-300">{row.pe > 0 ? row.pe.toFixed(1) : "—"}</td>
                  <td className="px-3 py-2.5 text-right text-zinc-300">{row.pb > 0 ? row.pb.toFixed(1) : "—"}</td>
                  <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.roe)}</td>
                  <td className="px-3 py-2.5 text-center">{tierLabel(row.liquidityTier)}</td>
                  <td className="px-3 py-2.5 text-right text-zinc-300">{row.tradabilityScore}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 0 && (
        <div className="px-4 py-3 border-t border-zinc-800 flex items-center justify-between text-xs text-zinc-400">
          <span>
            {page} / {totalPages} ({total})
          </span>
          <div className="flex items-center gap-1">
            <button onClick={() => onPageChange(page - 1)} disabled={page <= 1} className="p-1.5 rounded hover:bg-zinc-800 disabled:opacity-30 transition">
              <ChevronLeft size={14} />
            </button>
            {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
              const start = Math.max(1, Math.min(page - 2, totalPages - 4));
              const p = start + i;
              if (p > totalPages) return null;
              return (
                <button
                  key={p}
                  onClick={() => onPageChange(p)}
                  className={`w-7 h-7 rounded text-xs font-medium transition ${
                    p === page ? "bg-blue-600 text-white" : "hover:bg-zinc-800 text-zinc-400"
                  }`}
                >
                  {p}
                </button>
              );
            })}
            <button onClick={() => onPageChange(page + 1)} disabled={page >= totalPages} className="p-1.5 rounded hover:bg-zinc-800 disabled:opacity-30 transition">
              <ChevronRight size={14} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

function SortHeader({
  label,
  field,
  current,
  order,
  onSort,
}: {
  label: string;
  field: string;
  current: string;
  order: string;
  onSort: (field: string) => void;
}) {
  const active = current === field;
  return (
    <th
      className="px-3 py-3 text-right cursor-pointer select-none hover:text-blue-400 transition whitespace-nowrap"
      onClick={() => onSort(field)}
    >
      <span className="inline-flex items-center gap-1">
        {label}
        {active && (order === "asc" ? <ChevronUp size={12} /> : <ChevronDown size={12} />)}
      </span>
    </th>
  );
}
