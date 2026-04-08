"use client";

import { useState, useEffect, useCallback } from "react";
import { Save, Trash2 } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import type { ScreenerFilters } from "./FilterPanel";

// --- Types ---

interface FilterPreset {
  id: number;
  name: string;
  filters: ScreenerFilters;
}

export interface BuiltinPreset {
  name: string;
  filters: Partial<ScreenerFilters>;
}

const BUILTIN_PRESETS: BuiltinPreset[] = [
  { name: "Value Investing", filters: { maxPE: 12, maxPB: 1.5 } },
  { name: "High Growth", filters: { minRevenueGrowth: 20, minProfitGrowth: 20 } },
  { name: "High Dividend", filters: { minDivYield: 5 } },
  { name: "Low Debt", filters: { maxDebtToEquity: 0.5 } },
  { name: "Momentum", filters: { minRSI: 50, maCrossover: "golden_cross" } },
];

interface PresetSelectorProps {
  activePreset: string | null;
  currentFilters: ScreenerFilters;
  onApplyPreset: (name: string, filters: Partial<ScreenerFilters>) => void;
}

export function PresetSelector({ activePreset, currentFilters, onApplyPreset }: PresetSelectorProps) {
  const { t } = useI18n();
  const [savedPresets, setSavedPresets] = useState<FilterPreset[]>([]);
  const [presetName, setPresetName] = useState("");
  const [saving, setSaving] = useState(false);

  const fetchPresets = useCallback(async () => {
    const data = await apiFetch<FilterPreset[]>("/api/screener/presets");
    if (data) setSavedPresets(data);
  }, []);

  useEffect(() => {
    fetchPresets();
  }, [fetchPresets]);

  const savePreset = useCallback(async () => {
    if (!presetName.trim()) return;
    setSaving(true);
    await apiFetch<{ id: number }>("/api/screener/presets", {
      method: "POST",
      body: JSON.stringify({ name: presetName.trim(), filters: currentFilters }),
    });
    setPresetName("");
    setSaving(false);
    fetchPresets();
  }, [presetName, currentFilters, fetchPresets]);

  const deletePreset = useCallback(
    async (id: number) => {
      await apiFetch<{ ok: boolean }>(`/api/screener/presets/${id}`, { method: "DELETE" });
      fetchPresets();
    },
    [fetchPresets],
  );

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-xs text-zinc-500 font-medium">Presets:</span>

      {/* Built-in presets */}
      {BUILTIN_PRESETS.map((p) => (
        <button
          key={p.name}
          onClick={() => onApplyPreset(p.name, p.filters)}
          className={`px-3 py-1 rounded-full text-xs font-medium transition ${
            activePreset === p.name
              ? "bg-blue-600 text-white"
              : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
          }`}
        >
          {p.name}
        </button>
      ))}

      {/* Saved presets */}
      {savedPresets.map((p) => (
        <div key={p.id} className="flex items-center gap-1">
          <button
            onClick={() => onApplyPreset(p.name, p.filters)}
            className={`px-3 py-1 rounded-full text-xs font-medium transition ${
              activePreset === p.name
                ? "bg-emerald-600 text-white"
                : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
            }`}
          >
            {p.name}
          </button>
          <button
            onClick={() => deletePreset(p.id)}
            className="text-zinc-600 hover:text-red-400 transition"
            title={t("btn.delete")}
          >
            <Trash2 size={12} />
          </button>
        </div>
      ))}

      {/* Save new preset */}
      <div className="flex items-center gap-1 ml-2">
        <input
          type="text"
          placeholder={t("screener.save_preset")}
          value={presetName}
          onChange={(e) => setPresetName(e.target.value)}
          className="bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-xs text-zinc-200 focus:border-blue-500 focus:outline-none w-28"
        />
        <button
          onClick={savePreset}
          disabled={!presetName.trim() || saving}
          className="px-2 py-1 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-40 text-white rounded text-xs transition"
        >
          <Save size={12} />
        </button>
      </div>
    </div>
  );
}
