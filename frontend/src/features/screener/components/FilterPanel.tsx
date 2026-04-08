"use client";

import { useState, useCallback } from "react";
import {
  Search,
  RotateCcw,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { useI18n } from "@/context/I18nContext";

// --- Types ---

export interface ScreenerFilters {
  // Fundamental
  minPE?: number | null;
  maxPE?: number | null;
  minPB?: number | null;
  maxPB?: number | null;
  minMarketCap?: number | null;
  maxMarketCap?: number | null;
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
  // Technical
  minRSI?: number | null;
  maxRSI?: number | null;
  macdSignal?: string | null; // "bullish" | "bearish"
  maCrossover?: string | null; // "golden_cross" | "death_cross"
  volumeAnomaly?: boolean | null;
  // Classification
  sectors?: string[];
  exchanges?: string[];
  liquidityTier?: string | null; // "1" | "2" | "3"
}

export const EMPTY_FILTERS: ScreenerFilters = {};

const ICB_SECTORS = [
  { code: "VNIT", key: "sector.VNIT" },
  { code: "VNIND", key: "sector.VNIND" },
  { code: "VNCONS", key: "sector.VNCONS" },
  { code: "VNCOND", key: "sector.VNCOND" },
  { code: "VNHEAL", key: "sector.VNHEAL" },
  { code: "VNENE", key: "sector.VNENE" },
  { code: "VNUTI", key: "sector.VNUTI" },
  { code: "VNREAL", key: "sector.VNREAL" },
  { code: "VNFIN", key: "sector.VNFIN" },
  { code: "VNMAT", key: "sector.VNMAT" },
];

const EXCHANGES = ["HOSE", "HNX", "UPCOM"];
const LIQUIDITY_TIERS = [
  { value: "1", label: "Tier 1" },
  { value: "2", label: "Tier 2" },
  { value: "3", label: "Tier 3" },
];

interface FilterPanelProps {
  filters: ScreenerFilters;
  onChange: (filters: ScreenerFilters) => void;
  onApply: () => void;
  onReset: () => void;
}

export function FilterPanel({ filters, onChange, onApply, onReset }: FilterPanelProps) {
  const { t } = useI18n();
  const [showTechnical, setShowTechnical] = useState(false);

  const setField = useCallback(
    <K extends keyof ScreenerFilters>(key: K, val: ScreenerFilters[K]) => {
      onChange({ ...filters, [key]: val });
    },
    [filters, onChange],
  );

  const toggleArrayItem = useCallback(
    (key: "sectors" | "exchanges", value: string) => {
      const current = filters[key] ?? [];
      const next = current.includes(value)
        ? current.filter((v) => v !== value)
        : [...current, value];
      onChange({ ...filters, [key]: next });
    },
    [filters, onChange],
  );

  return (
    <div className="w-full lg:w-72 flex-shrink-0 bg-zinc-900 dark:bg-zinc-900 border border-zinc-800 rounded-xl p-4 overflow-y-auto space-y-3">
      <div className="flex items-center justify-between mb-1">
        <h2 className="text-sm font-bold text-zinc-300">{t("screener.filters")}</h2>
        <button onClick={onReset} className="text-zinc-500 hover:text-zinc-300 transition" title={t("btn.reset")}>
          <RotateCcw size={14} />
        </button>
      </div>

      {/* Fundamental Filters */}
      <p className="text-[10px] uppercase tracking-wider text-zinc-500 font-semibold pt-1">
        {t("screener.filters")} — Fundamental
      </p>
      <RangeInput label={t("finance.pe")} minVal={filters.minPE} maxVal={filters.maxPE} onMinChange={(v) => setField("minPE", v)} onMaxChange={(v) => setField("maxPE", v)} />
      <RangeInput label={t("finance.pb")} minVal={filters.minPB} maxVal={filters.maxPB} onMinChange={(v) => setField("minPB", v)} onMaxChange={(v) => setField("maxPB", v)} />
      <RangeInput label={t("finance.market_cap")} minVal={filters.minMarketCap} maxVal={filters.maxMarketCap} onMinChange={(v) => setField("minMarketCap", v)} onMaxChange={(v) => setField("maxMarketCap", v)} />
      <RangeInput label={t("finance.ev_ebitda")} minVal={filters.minEVEBITDA} maxVal={filters.maxEVEBITDA} onMinChange={(v) => setField("minEVEBITDA", v)} onMaxChange={(v) => setField("maxEVEBITDA", v)} />
      <RangeInput label={t("finance.roe") + " (%)"} minVal={filters.minROE} maxVal={filters.maxROE} onMinChange={(v) => setField("minROE", v)} onMaxChange={(v) => setField("maxROE", v)} />
      <RangeInput label={t("finance.roa") + " (%)"} minVal={filters.minROA} maxVal={filters.maxROA} onMinChange={(v) => setField("minROA", v)} onMaxChange={(v) => setField("maxROA", v)} />
      <RangeInput label={t("finance.revenue_growth") + " (%)"} minVal={filters.minRevenueGrowth} maxVal={filters.maxRevenueGrowth} onMinChange={(v) => setField("minRevenueGrowth", v)} onMaxChange={(v) => setField("maxRevenueGrowth", v)} />
      <RangeInput label={t("finance.profit_growth") + " (%)"} minVal={filters.minProfitGrowth} maxVal={filters.maxProfitGrowth} onMinChange={(v) => setField("minProfitGrowth", v)} onMaxChange={(v) => setField("maxProfitGrowth", v)} />
      <RangeInput label={t("finance.div_yield") + " (%)"} minVal={filters.minDivYield} maxVal={filters.maxDivYield} onMinChange={(v) => setField("minDivYield", v)} onMaxChange={(v) => setField("maxDivYield", v)} />
      <RangeInput label={t("finance.debt_equity")} minVal={filters.minDebtToEquity} maxVal={filters.maxDebtToEquity} onMinChange={(v) => setField("minDebtToEquity", v)} onMaxChange={(v) => setField("maxDebtToEquity", v)} />

      {/* Technical Filters (collapsible) */}
      <button
        onClick={() => setShowTechnical((v) => !v)}
        className="flex items-center justify-between w-full pt-2 text-[10px] uppercase tracking-wider text-zinc-500 font-semibold"
      >
        <span>Technical</span>
        {showTechnical ? <ChevronUp size={12} /> : <ChevronDown size={12} />}
      </button>

      {showTechnical && (
        <div className="space-y-3">
          <RangeInput label="RSI" minVal={filters.minRSI} maxVal={filters.maxRSI} onMinChange={(v) => setField("minRSI", v)} onMaxChange={(v) => setField("maxRSI", v)} />
          <SelectInput
            label="MACD"
            value={filters.macdSignal ?? ""}
            options={[
              { value: "", label: t("common.all") },
              { value: "bullish", label: "Bullish" },
              { value: "bearish", label: "Bearish" },
            ]}
            onChange={(v) => setField("macdSignal", v || null)}
          />
          <SelectInput
            label="MA Crossover"
            value={filters.maCrossover ?? ""}
            options={[
              { value: "", label: t("common.all") },
              { value: "golden_cross", label: "Golden Cross" },
              { value: "death_cross", label: "Death Cross" },
            ]}
            onChange={(v) => setField("maCrossover", v || null)}
          />
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={filters.volumeAnomaly ?? false}
              onChange={(e) => setField("volumeAnomaly", e.target.checked || null)}
              className="rounded border-zinc-700 bg-zinc-800 text-blue-600"
            />
            <label className="text-xs text-zinc-400">Volume Anomaly</label>
          </div>
        </div>
      )}

      {/* Sector Filter */}
      <ChipSelect
        label={t("table.sector")}
        options={ICB_SECTORS.map((s) => ({ value: s.code, label: t(s.key) }))}
        selected={filters.sectors ?? []}
        onToggle={(v) => toggleArrayItem("sectors", v)}
      />

      {/* Exchange Filter */}
      <ChipSelect
        label={t("table.exchange")}
        options={EXCHANGES.map((e) => ({ value: e, label: e }))}
        selected={filters.exchanges ?? []}
        onToggle={(v) => toggleArrayItem("exchanges", v)}
      />

      {/* Liquidity Tier */}
      <SelectInput
        label={t("screener.liquidity_tier")}
        value={filters.liquidityTier ?? ""}
        options={[{ value: "", label: t("common.all") }, ...LIQUIDITY_TIERS]}
        onChange={(v) => setField("liquidityTier", v || null)}
      />

      {/* Apply Button */}
      <button
        onClick={onApply}
        className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2 rounded-lg text-sm transition mt-2"
      >
        <Search size={14} className="inline mr-1.5 -mt-0.5" />
        {t("btn.apply")}
      </button>
    </div>
  );
}

// --- Sub-components ---

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

function SelectInput({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-zinc-400">{label}</label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:border-blue-500 focus:outline-none"
      >
        {options.map((o) => (
          <option key={o.value} value={o.value}>{o.label}</option>
        ))}
      </select>
    </div>
  );
}

function ChipSelect({
  label,
  options,
  selected,
  onToggle,
}: {
  label: string;
  options: { value: string; label: string }[];
  selected: string[];
  onToggle: (v: string) => void;
}) {
  return (
    <div className="flex flex-col gap-1">
      <label className="text-xs font-medium text-zinc-400">{label}</label>
      <div className="flex flex-wrap gap-1">
        {options.map((o) => (
          <button
            key={o.value}
            type="button"
            onClick={() => onToggle(o.value)}
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
