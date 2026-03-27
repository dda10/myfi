"use client";

import { useState, useCallback } from "react";
import {
  INDICATOR_REGISTRY,
  getIndicatorsByCategory,
  createIndicatorInstance,
  getRegistryEntry,
  type IndicatorConfig,
  type IndicatorCategory,
  type IndicatorRegistryEntry,
} from "@/lib/indicator-renderer";
import { ChevronDown, ChevronRight, Plus, Trash2, Eye, EyeOff } from "lucide-react";

interface IndicatorPanelProps {
  activeIndicators: IndicatorConfig[];
  onAdd: (config: IndicatorConfig) => void;
  onRemove: (id: string) => void;
  onUpdate: (id: string, params: Record<string, number>) => void;
  onToggleVisibility: (id: string) => void;
}

export function IndicatorPanel({
  activeIndicators,
  onAdd,
  onRemove,
  onUpdate,
  onToggleVisibility,
}: IndicatorPanelProps) {
  const [expandedCategory, setExpandedCategory] = useState<IndicatorCategory | null>(null);
  const [expandedConfig, setExpandedConfig] = useState<string | null>(null);
  const grouped = getIndicatorsByCategory();
  const categories: IndicatorCategory[] = ["Trend", "Momentum", "Volatility", "Volume"];

  const handleAdd = useCallback(
    (entry: IndicatorRegistryEntry) => {
      const instance = createIndicatorInstance(entry);
      onAdd(instance);
    },
    [onAdd],
  );

  const handleParamChange = useCallback(
    (id: string, paramName: string, value: number, currentParams: Record<string, number>) => {
      onUpdate(id, { ...currentParams, [paramName]: value });
    },
    [onUpdate],
  );

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden text-sm">
      <div className="px-4 py-3 border-b border-zinc-800">
        <h3 className="text-white font-semibold">Indicators</h3>
      </div>

      {/* Active indicators */}
      {activeIndicators.length > 0 && (
        <div className="border-b border-zinc-800">
          <div className="px-4 py-2 text-xs text-zinc-500 uppercase tracking-wider">Active</div>
          {activeIndicators.map((cfg) => {
            const entry = getRegistryEntry(cfg.indicatorKey);
            const isExpanded = expandedConfig === cfg.id;
            return (
              <div key={cfg.id} className="border-t border-zinc-800/50">
                <div className="flex items-center gap-2 px-4 py-2 hover:bg-zinc-800/50">
                  <span
                    className="w-3 h-3 rounded-full flex-shrink-0"
                    style={{ backgroundColor: cfg.color }}
                  />
                  <button
                    className="flex-1 text-left text-zinc-300 truncate"
                    onClick={() => setExpandedConfig(isExpanded ? null : cfg.id)}
                  >
                    {cfg.name}
                    {entry && entry.paramDefs.length > 0 && (
                      <span className="text-zinc-500 ml-1">
                        ({Object.values(cfg.params).join(", ")})
                      </span>
                    )}
                  </button>
                  <button
                    onClick={() => onToggleVisibility(cfg.id)}
                    className="p-1 text-zinc-500 hover:text-zinc-300"
                    aria-label={cfg.visible ? "Hide indicator" : "Show indicator"}
                  >
                    {cfg.visible ? <Eye size={14} /> : <EyeOff size={14} />}
                  </button>
                  <button
                    onClick={() => onRemove(cfg.id)}
                    className="p-1 text-zinc-500 hover:text-red-400"
                    aria-label="Remove indicator"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
                {/* Parameter editing */}
                {isExpanded && entry && entry.paramDefs.length > 0 && (
                  <div className="px-4 pb-3 space-y-2">
                    {entry.paramDefs.map((pd) => (
                      <label key={pd.name} className="flex items-center gap-2">
                        <span className="text-zinc-500 w-16 text-xs">{pd.label}</span>
                        <input
                          type="number"
                          value={cfg.params[pd.name] ?? pd.default}
                          min={pd.min}
                          max={pd.max}
                          step={pd.step ?? 1}
                          onChange={(e) => {
                            const v = parseFloat(e.target.value);
                            if (!isNaN(v)) handleParamChange(cfg.id, pd.name, v, cfg.params);
                          }}
                          className="flex-1 bg-zinc-800 border border-zinc-700 rounded px-2 py-1 text-zinc-300 text-xs w-20"
                        />
                      </label>
                    ))}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Available indicators grouped by category */}
      <div>
        <div className="px-4 py-2 text-xs text-zinc-500 uppercase tracking-wider">Available</div>
        {categories.map((cat) => {
          const entries = grouped[cat];
          if (entries.length === 0) return null;
          const isOpen = expandedCategory === cat;
          return (
            <div key={cat}>
              <button
                className="flex items-center gap-2 w-full px-4 py-2 text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200"
                onClick={() => setExpandedCategory(isOpen ? null : cat)}
              >
                {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                <span className="font-medium">{cat}</span>
                <span className="text-zinc-600 text-xs ml-auto">{entries.length}</span>
              </button>
              {isOpen && (
                <div className="pb-1">
                  {entries.map((entry) => (
                    <button
                      key={entry.key}
                      className="flex items-center gap-2 w-full px-6 py-1.5 text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200"
                      onClick={() => handleAdd(entry)}
                    >
                      <span
                        className="w-2 h-2 rounded-full flex-shrink-0"
                        style={{ backgroundColor: entry.color }}
                      />
                      <span className="flex-1 text-left">{entry.name}</span>
                      <span className="text-zinc-600">
                        {entry.pane === "overlay" ? "overlay" : "osc"}
                      </span>
                      <Plus size={12} className="text-zinc-600" />
                    </button>
                  ))}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
