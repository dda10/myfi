"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Search,
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  Save,
  Trash2,
  RotateCcw,
  SlidersHorizontal,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { useApp } from "@/context/AppContext";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types ---

interface ScreenerFilters {
  minPE?: number | null;
  maxPE?: number | null;
  minPB?: number | null;
  maxPB?: number | null;
  minMarketCap?: number | null;
  minEVEBITDA?: number | null;
  maxEVEBITDA?: number | null;
  minROE?: number | null;
  maxROE?: number | null;
  minROA?: number | null;
  maxROA?: number | null;
  minRevenueGrowth?: number | null;
  maxRevenueGrowth?: number | null;
  minProfitGrowth?: number | null;
  maxProfitGrowth?: number | null;
  minDivYield?: number | null;
  maxDivYield?: number | null;
  minDebtToEquity?: number | null;
  maxDebtToEquity?: number | null;
  sectors?: string[];
  exchanges?: string[];
  sectorTrends?: string[];
  sortBy: string;
  sortOrder: string;
  page: number;
  pageSize: number;
}

interface ScreenerResult {
  symbol: string;
  exchange: string;
  sector: string;
  sectorName: string;
  marketCap: number;
  pe: number;
  pb: number;
  evEbitda: number;
  roe: number;
  roa: number;
  revenueGrowth: number;
  profitGrowth: number;
  divYield: number;
  debtToEquity: number;
  sectorTrend: string;
}

interface ScreenerResponse {
  data: ScreenerResult[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

interface FilterPreset {
  id: number;
  name: string;
  filters: ScreenerFilters;
}

// --- Constants ---

const ICB_SECTORS = [
  { code: "VNIT", name: "Technology" },
  { code: "VNIND", name: "Industrial" },
  { code: "VNCONS", name: "Consumer Staples" },
  { code: "VNCOND", name: "Consumer Disc." },
  { code: "VNHEAL", name: "Healthcare" },
  { code: "VNENE", name: "Energy" },
  { code: "VNUTI", name: "Utilities" },
  { code: "VNREAL", name: "Real Estate" },
  { code: "VNFIN", name: "Finance" },
  { code: "VNMAT", name: "Materials" },
];

const EXCHANGES = ["HOSE", "HNX", "UPCOM"];
const SECTOR_TRENDS = ["uptrend", "downtrend", "sideways"];
const PAGE_SIZES = [10, 20, 50];

const EMPTY_FILTERS: ScreenerFilters = {
  sortBy: "marketCap",
  sortOrder: "desc",
  page: 1,
  pageSize: 20,
};

type BuiltinPreset = { name: string; filters: Partial<ScreenerFilters> };

const BUILTIN_PRESETS: BuiltinPreset[] = [
  {
    name: "Value Investing",
    filters: { maxPE: 12, maxPB: 1.5 },
  },
  {
    name: "High Growth",
    filters: { minRevenueGrowth: 20, minProfitGrowth: 20 },
  },
  {
    name: "High Dividend",
    filters: { minDivYield: 5 },
  },
  {
    name: "Low Debt",
    filters: { maxDebtToEquity: 0.5 },
  },
];

// --- Helpers ---

function cleanFilters(f: ScreenerFilters): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(f)) {
    if (v === null || v === undefined || v === "") continue;
    if (Array.isArray(v) && v.length === 0) continue;
    out[k] = v;
  }
  return out;
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T | null> {
  try {
    const res = await fetch(`${API_URL}${path}`, { credentials: "include", ...init });
    if (!res.ok) return null;
    return (await res.json()) as T;
  } catch {
    return null;
  }
}

// --- Range Input Component ---

function RangeInput({
  label,
  minVal,
  maxVal,
  onMinChange,
  onMaxChange,
}: {
  label: string;
  minVal: number | null | undefined;
  maxVal: number | null | undefined;
  onMinChange: (v: number | null) => void;
  onMaxChange: (v: number | null) => void;
}) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-zinc-400">{label}</label>
      <div className="flex gap-1.5">
        <input
          type="number"
          placeholder="Min"
          value={minVal ?? ""}
          onChange={(e) => onMinChange(e.target.value === "" ? null : Number(e.target.value))}
          className="w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:border-blue-500 focus:outline-none"
        />
        <input
          type="number"
          placeholder="Max"
          value={maxVal ?? ""}
          onChange={(e) => onMaxChange(e.target.value === "" ? null : Number(e.target.value))}
          className="w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:border-blue-500 focus:outline-none"
        />
      </div>
    </div>
  );
}

// --- Multi-Select Chip Component ---

function ChipSelect({
  label,
  options,
  selected,
  onChange,
}: {
  label: string;
  options: { value: string; label: string }[];
  selected: string[];
  onChange: (v: string[]) => void;
}) {
  const toggle = (val: string) => {
    onChange(selected.includes(val) ? selected.filter((s) => s !== val) : [...selected, val]);
  };
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-zinc-400">{label}</label>
      <div className="flex flex-wrap gap-1">
        {options.map((o) => (
          <button
            key={o.value}
            type="button"
            onClick={() => toggle(o.value)}
            className={`px-2 py-0.5 rounded text-xs font-medium transition ${
              selected.includes(o.value)
                ? "bg-blue-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
            }`}
          >
            {o.label}
          </button>
        ))}
      </div>
    </div>
  );
}

// --- Sortable Header ---

function SortHeader({
  label,
  field,
  currentSort,
  currentOrder,
  onSort,
}: {
  label: string;
  field: string;
  currentSort: string;
  currentOrder: string;
  onSort: (field: string) => void;
}) {
  const active = currentSort === field;
  return (
    <th
      className="px-3 py-3 text-right cursor-pointer select-none hover:text-blue-400 transition whitespace-nowrap"
      onClick={() => onSort(field)}
    >
      <span className="inline-flex items-center gap-1">
        {label}
        {active &&
          (currentOrder === "asc" ? (
            <ChevronUp size={12} />
          ) : (
            <ChevronDown size={12} />
          ))}
      </span>
    </th>
  );
}

// --- Main Component ---

export function FilterModule() {
  const { t, formatNumber, formatPercent } = useI18n();
  const { setActiveSymbol, setActiveTab } = useApp();

  const [filters, setFilters] = useState<ScreenerFilters>({ ...EMPTY_FILTERS });
  const [results, setResults] = useState<ScreenerResult[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(false);
  const [showFilters, setShowFilters] = useState(true);

  // Presets
  const [savedPresets, setSavedPresets] = useState<FilterPreset[]>([]);
  const [presetName, setPresetName] = useState("");
  const [savingPreset, setSavingPreset] = useState(false);
  const [activePresetLabel, setActivePresetLabel] = useState<string | null>(null);

  // Fetch screener results
  const fetchResults = useCallback(async (f: ScreenerFilters) => {
    setLoading(true);
    const body = cleanFilters(f);
    const resp = await apiFetch<ScreenerResponse>("/api/screener", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (resp) {
      setResults(resp.data ?? []);
      setTotal(resp.total);
      setTotalPages(resp.totalPages);
    } else {
      setResults([]);
      setTotal(0);
      setTotalPages(0);
    }
    setLoading(false);
  }, []);

  // Fetch saved presets
  const fetchPresets = useCallback(async () => {
    const data = await apiFetch<FilterPreset[]>("/api/screener/presets");
    if (data) setSavedPresets(data);
  }, []);

  // Initial load
  useEffect(() => {
    fetchResults(filters);
    fetchPresets();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Sort handler
  const handleSort = useCallback(
    (field: string) => {
      const newOrder = filters.sortBy === field && filters.sortOrder === "desc" ? "asc" : "desc";
      const next = { ...filters, sortBy: field, sortOrder: newOrder, page: 1 };
      setFilters(next);
      fetchResults(next);
    },
    [filters, fetchResults],
  );

  // Page change
  const goToPage = useCallback(
    (p: number) => {
      if (p < 1 || p > totalPages) return;
      const next = { ...filters, page: p };
      setFilters(next);
      fetchResults(next);
    },
    [filters, totalPages, fetchResults],
  );

  // Page size change
  const changePageSize = useCallback(
    (size: number) => {
      const next = { ...filters, pageSize: size, page: 1 };
      setFilters(next);
      fetchResults(next);
    },
    [filters, fetchResults],
  );

  // Apply filters
  const applyFilters = useCallback(() => {
    const next = { ...filters, page: 1 };
    setFilters(next);
    fetchResults(next);
    setActivePresetLabel(null);
  }, [filters, fetchResults]);

  // Reset filters
  const resetFilters = useCallback(() => {
    const next = { ...EMPTY_FILTERS };
    setFilters(next);
    fetchResults(next);
    setActivePresetLabel(null);
  }, [fetchResults]);

  // Apply built-in preset
  const applyBuiltinPreset = useCallback(
    (preset: BuiltinPreset) => {
      const next: ScreenerFilters = { ...EMPTY_FILTERS, ...preset.filters };
      setFilters(next);
      fetchResults(next);
      setActivePresetLabel(preset.name);
    },
    [fetchResults],
  );

  // Apply saved preset
  const applySavedPreset = useCallback(
    (preset: FilterPreset) => {
      const next: ScreenerFilters = { ...EMPTY_FILTERS, ...preset.filters, page: 1 };
      setFilters(next);
      fetchResults(next);
      setActivePresetLabel(preset.name);
    },
    [fetchResults],
  );

  // Save preset
  const savePreset = useCallback(async () => {
    if (!presetName.trim()) return;
    setSavingPreset(true);
    await apiFetch<{ id: number }>("/api/screener/presets", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: presetName.trim(), filters }),
    });
    setPresetName("");
    setSavingPreset(false);
    fetchPresets();
  }, [presetName, filters, fetchPresets]);

  // Delete preset
  const deletePreset = useCallback(
    async (id: number) => {
      await apiFetch<{ ok: boolean }>(`/api/screener/presets/${id}`, { method: "DELETE" });
      fetchPresets();
    },
    [fetchPresets],
  );

  // Row click
  const handleRowClick = useCallback(
    (symbol: string) => {
      setActiveSymbol(symbol);
      setActiveTab("Markets");
    },
    [setActiveSymbol, setActiveTab],
  );

  // Filter updater helper
  const setField = useCallback(
    <K extends keyof ScreenerFilters>(key: K, val: ScreenerFilters[K]) => {
      setFilters((prev) => ({ ...prev, [key]: val }));
    },
    [],
  );

  const trendIcon = (trend: string) => {
    if (trend === "uptrend") return "↑";
    if (trend === "downtrend") return "↓";
    return "→";
  };

  const trendColor = (trend: string) => {
    if (trend === "uptrend") return "text-green-400";
    if (trend === "downtrend") return "text-red-400";
    return "text-yellow-400";
  };

  return (
    <div className="flex flex-col gap-4 w-full h-full text-zinc-200">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-white">{t("screener.title")}</h1>
        <button
          onClick={() => setShowFilters((v) => !v)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-sm transition"
        >
          <SlidersHorizontal size={14} />
          {showFilters ? "Hide Filters" : "Show Filters"}
        </button>
      </div>

      {/* Presets Bar */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-xs text-zinc-500 font-medium">Presets:</span>
        {BUILTIN_PRESETS.map((p) => (
          <button
            key={p.name}
            onClick={() => applyBuiltinPreset(p)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition ${
              activePresetLabel === p.name
                ? "bg-blue-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
            }`}
          >
            {p.name}
          </button>
        ))}
        {savedPresets.map((p) => (
          <div key={p.id} className="flex items-center gap-1">
            <button
              onClick={() => applySavedPreset(p)}
              className={`px-3 py-1 rounded-full text-xs font-medium transition ${
                activePresetLabel === p.name
                  ? "bg-emerald-600 text-white"
                  : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
              }`}
            >
              {p.name}
            </button>
            <button
              onClick={() => deletePreset(p.id)}
              className="text-zinc-600 hover:text-red-400 transition"
              title="Delete preset"
            >
              <Trash2 size={12} />
            </button>
          </div>
        ))}
      </div>

      <div className="flex flex-col lg:flex-row gap-4 flex-1 min-h-0">
        {/* Filter Panel */}
        {showFilters && (
          <div className="w-full lg:w-72 flex-shrink-0 bg-zinc-900 border border-zinc-800 rounded-xl p-4 overflow-y-auto space-y-3">
            <div className="flex items-center justify-between mb-1">
              <h2 className="text-sm font-bold text-zinc-300">{t("screener.filters")}</h2>
              <button onClick={resetFilters} className="text-zinc-500 hover:text-zinc-300 transition" title="Reset">
                <RotateCcw size={14} />
              </button>
            </div>

            {/* Fundamental Ranges */}
            <RangeInput label="P/E" minVal={filters.minPE} maxVal={filters.maxPE} onMinChange={(v) => setField("minPE", v)} onMaxChange={(v) => setField("maxPE", v)} />
            <RangeInput label="P/B" minVal={filters.minPB} maxVal={filters.maxPB} onMinChange={(v) => setField("minPB", v)} onMaxChange={(v) => setField("maxPB", v)} />
            <RangeInput label="Market Cap (B)" minVal={filters.minMarketCap} maxVal={null} onMinChange={(v) => setField("minMarketCap", v)} onMaxChange={() => {}} />
            <RangeInput label="EV/EBITDA" minVal={filters.minEVEBITDA} maxVal={filters.maxEVEBITDA} onMinChange={(v) => setField("minEVEBITDA", v)} onMaxChange={(v) => setField("maxEVEBITDA", v)} />
            <RangeInput label="ROE (%)" minVal={filters.minROE} maxVal={filters.maxROE} onMinChange={(v) => setField("minROE", v)} onMaxChange={(v) => setField("maxROE", v)} />
            <RangeInput label="ROA (%)" minVal={filters.minROA} maxVal={filters.maxROA} onMinChange={(v) => setField("minROA", v)} onMaxChange={(v) => setField("maxROA", v)} />
            <RangeInput label="Revenue Growth (%)" minVal={filters.minRevenueGrowth} maxVal={filters.maxRevenueGrowth} onMinChange={(v) => setField("minRevenueGrowth", v)} onMaxChange={(v) => setField("maxRevenueGrowth", v)} />
            <RangeInput label="Profit Growth (%)" minVal={filters.minProfitGrowth} maxVal={filters.maxProfitGrowth} onMinChange={(v) => setField("minProfitGrowth", v)} onMaxChange={(v) => setField("maxProfitGrowth", v)} />
            <RangeInput label="Div Yield (%)" minVal={filters.minDivYield} maxVal={filters.maxDivYield} onMinChange={(v) => setField("minDivYield", v)} onMaxChange={(v) => setField("maxDivYield", v)} />
            <RangeInput label="Debt/Equity" minVal={filters.minDebtToEquity} maxVal={filters.maxDebtToEquity} onMinChange={(v) => setField("minDebtToEquity", v)} onMaxChange={(v) => setField("maxDebtToEquity", v)} />

            {/* Sectors */}
            <ChipSelect
              label="Sectors"
              options={ICB_SECTORS.map((s) => ({ value: s.code, label: s.name }))}
              selected={filters.sectors ?? []}
              onChange={(v) => setField("sectors", v)}
            />

            {/* Exchanges */}
            <ChipSelect
              label="Exchanges"
              options={EXCHANGES.map((e) => ({ value: e, label: e }))}
              selected={filters.exchanges ?? []}
              onChange={(v) => setField("exchanges", v)}
            />

            {/* Sector Trends */}
            <ChipSelect
              label="Sector Trend"
              options={SECTOR_TRENDS.map((t) => ({ value: t, label: t.charAt(0).toUpperCase() + t.slice(1) }))}
              selected={filters.sectorTrends ?? []}
              onChange={(v) => setField("sectorTrends", v)}
            />

            {/* Apply Button */}
            <button
              onClick={applyFilters}
              className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2 rounded-lg text-sm transition mt-2"
            >
              <Search size={14} className="inline mr-1.5 -mt-0.5" />
              Apply Filters
            </button>

            {/* Save Preset */}
            <div className="border-t border-zinc-800 pt-3 mt-2">
              <label className="text-xs font-medium text-zinc-400 mb-1 block">{t("screener.save_preset")}</label>
              <div className="flex gap-1.5">
                <input
                  type="text"
                  placeholder="Preset name"
                  value={presetName}
                  onChange={(e) => setPresetName(e.target.value)}
                  className="flex-1 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:border-blue-500 focus:outline-none"
                />
                <button
                  onClick={savePreset}
                  disabled={!presetName.trim() || savingPreset}
                  className="px-2 py-1 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-40 text-white rounded text-xs transition"
                >
                  <Save size={12} />
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Results Area */}
        <div className="flex-1 flex flex-col min-w-0 bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
          {/* Results Header */}
          <div className="px-4 py-3 border-b border-zinc-800 flex items-center justify-between">
            <span className="text-sm text-zinc-400">
              {t("screener.results")}: <span className="text-white font-bold">{total}</span> stocks
            </span>
            <div className="flex items-center gap-2 text-xs text-zinc-500">
              <span>Page size:</span>
              <select
                value={filters.pageSize}
                onChange={(e) => changePageSize(Number(e.target.value))}
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
                  <th className="px-3 py-3 text-left">Symbol</th>
                  <th className="px-3 py-3 text-left">Exchange</th>
                  <th className="px-3 py-3 text-left">Sector</th>
                  <SortHeader label="Mkt Cap" field="marketCap" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="P/E" field="pe" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="P/B" field="pb" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="EV/EBITDA" field="evEbitda" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="ROE" field="roe" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="ROA" field="roa" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="Rev Grw" field="revenueGrowth" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="Pft Grw" field="profitGrowth" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="Div Yld" field="divYield" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <SortHeader label="D/E" field="debtToEquity" currentSort={filters.sortBy} currentOrder={filters.sortOrder} onSort={handleSort} />
                  <th className="px-3 py-3 text-center">Trend</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-zinc-800/50">
                {loading ? (
                  <tr>
                    <td colSpan={14} className="text-center py-16 text-zinc-500">Loading...</td>
                  </tr>
                ) : results.length === 0 ? (
                  <tr>
                    <td colSpan={14} className="text-center py-16 text-zinc-500">{t("screener.no_results")}</td>
                  </tr>
                ) : (
                  results.map((row) => (
                    <tr
                      key={row.symbol}
                      onClick={() => handleRowClick(row.symbol)}
                      className="hover:bg-zinc-800/40 transition cursor-pointer"
                    >
                      <td className="px-3 py-2.5 font-bold text-white">{row.symbol}</td>
                      <td className="px-3 py-2.5 text-zinc-400">{row.exchange}</td>
                      <td className="px-3 py-2.5 text-zinc-400">{row.sectorName || row.sector}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatNumber(row.marketCap, 0)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{row.pe.toFixed(2)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{row.pb.toFixed(2)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{row.evEbitda === 0 ? "-" : row.evEbitda.toFixed(2)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.roe)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.roa)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.revenueGrowth)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.profitGrowth)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{formatPercent(row.divYield)}</td>
                      <td className="px-3 py-2.5 text-right text-zinc-300">{row.debtToEquity.toFixed(2)}</td>
                      <td className={`px-3 py-2.5 text-center font-bold ${trendColor(row.sectorTrend)}`}>
                        {trendIcon(row.sectorTrend)}
                      </td>
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
                Page {filters.page} of {totalPages} ({total} total)
              </span>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => goToPage(filters.page - 1)}
                  disabled={filters.page <= 1}
                  className="p-1.5 rounded hover:bg-zinc-800 disabled:opacity-30 transition"
                >
                  <ChevronLeft size={14} />
                </button>
                {Array.from({ length: Math.min(totalPages, 5) }, (_, i) => {
                  const start = Math.max(1, Math.min(filters.page - 2, totalPages - 4));
                  const p = start + i;
                  if (p > totalPages) return null;
                  return (
                    <button
                      key={p}
                      onClick={() => goToPage(p)}
                      className={`w-7 h-7 rounded text-xs font-medium transition ${
                        p === filters.page ? "bg-blue-600 text-white" : "hover:bg-zinc-800 text-zinc-400"
                      }`}
                    >
                      {p}
                    </button>
                  );
                })}
                <button
                  onClick={() => goToPage(filters.page + 1)}
                  disabled={filters.page >= totalPages}
                  className="p-1.5 rounded hover:bg-zinc-800 disabled:opacity-30 transition"
                >
                  <ChevronRight size={14} />
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
