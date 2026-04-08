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
import { useI18n } from "@/context/I18nContext";
import { useTheme } from "@/context/ThemeContext";

interface IndicatorPanelProps {
  activeIndicators: IndicatorConfig[];
  onAdd: (config: IndicatorConfig) => void;
  onRemove: (id: string) => void;
  onUpdate: (id: string, params: Record<string, number>) => void;
  onToggleVisibility: (id: string) => void;
}

const CATEGORY_I18N: Record<IndicatorCategory, string> = {
  Trend: "chart.category.trend",
  Momentum: "chart.category.momentum",
  Volatility: "chart.category.volatility",
  Volume: "chart.category.volume",
};

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
  const { t } = useI18n();
  const { theme } = useTheme();
  const isDark = theme === "dark";

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
    <div className={`${isDark ? "bg-zinc-900 border-zinc-800" : "bg-white border-gray-200"} border rounded-xl overflow-hidden text-sm`}>
      <div className={`px-4 py-3 border-b ${isDark ? "border-zinc-800" : "border-gray-200"}`}>
        <h3 className={`font-semibold ${isDark ? "text-white" : "text-gray-900"}`}>{t("chart.indicators")}</h3>
      </div>

      {/* Active indicators */}
      {activeIndicators.length > 0 && (
        <div className={`border-b ${isDark ? "border-zinc-800" : "border-gray-200"}`}>
          <div className={`px-4 py-2 text-xs uppercase tracking-wider ${isDark ? "text-zinc-500" : "text-gray-400"}`}>Active</div>
          {activeIndicators.map((cfg) => {
            const entry = getRegistryEntry(cfg.indicatorKey);
            const isExpanded = expandedConfig === cfg.id;
            return (
              <div key={cfg.id} className={`border-t ${isDark ? "border-zinc-800/50" : "border-gray-100"}`}>
                <div className={`flex items-center gap-2 px-4 py-2 ${isDark ? "hover:bg-zinc-800/50" : "hover:bg-gray-50"}`}>
                  <span className="w-3 h-3 rounded-full flex-shrink-0" style={{ backgroundColor: cfg.color }} />
                  <button
                    className={`flex-1 text-left truncate ${isDark ? "text-zinc-300" : "text-gray-700"}`}
                    onClick={() => setExpandedConfig(isExpanded ? null : cfg.id)}
                  >
                    {cfg.name}
                    {entry && entry.paramDefs.length > 0 && (
                      <span className={`ml-1 ${isDark ? "text-zinc-500" : "text-gray-400"}`}>
                        ({Object.values(cfg.params).join(", ")})
                      </span>
                    )}
                  </button>
                  <button
                    onClick={() => onToggleVisibility(cfg.id)}
                    className={`p-1 ${isDark ? "text-zinc-500 hover:text-zinc-300" : "text-gray-400 hover:text-gray-600"}`}
                    aria-label={cfg.visible ? "Hide indicator" : "Show indicator"}
                  >
                    {cfg.visible ? <Eye size={14} /> : <EyeOff size={14} />}
                  </button>
                  <button
                    onClick={() => onRemove(cfg.id)}
                    className={`p-1 ${isDark ? "text-zinc-500 hover:text-red-400" : "text-gray-400 hover:text-red-500"}`}
                    aria-label="Remove indicator"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
                {isExpanded && entry && entry.paramDefs.length > 0 && (
                  <div className="px-4 pb-3 space-y-2">
                    {entry.paramDefs.map((pd) => (
                      <label key={pd.name} className="flex items-center gap-2">
                        <span className={`w-16 text-xs ${isDark ? "text-zinc-500" : "text-gray-400"}`}>{pd.label}</span>
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
                          className={`flex-1 border rounded px-2 py-1 text-xs w-20 ${
                            isDark
                              ? "bg-zinc-800 border-zinc-700 text-zinc-300"
                              : "bg-gray-50 border-gray-300 text-gray-700"
                          }`}
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
        <div className={`px-4 py-2 text-xs uppercase tracking-wider ${isDark ? "text-zinc-500" : "text-gray-400"}`}>
          Available
        </div>
        {categories.map((cat) => {
          const entries = grouped[cat];
          if (entries.length === 0) return null;
          const isOpen = expandedCategory === cat;
          return (
            <div key={cat}>
              <button
                className={`flex items-center gap-2 w-full px-4 py-2 ${
                  isDark
                    ? "text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200"
                    : "text-gray-500 hover:bg-gray-50 hover:text-gray-800"
                }`}
                onClick={() => setExpandedCategory(isOpen ? null : cat)}
              >
                {isOpen ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
                <span className="font-medium">{t(CATEGORY_I18N[cat])}</span>
                <span className={`text-xs ml-auto ${isDark ? "text-zinc-600" : "text-gray-400"}`}>{entries.length}</span>
              </button>
              {isOpen && (
                <div className="pb-1">
                  {entries.map((entry) => (
                    <button
                      key={entry.key}
                      className={`flex items-center gap-2 w-full px-6 py-1.5 ${
                        isDark
                          ? "text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200"
                          : "text-gray-500 hover:bg-gray-50 hover:text-gray-800"
                      }`}
                      onClick={() => handleAdd(entry)}
                    >
                      <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ backgroundColor: entry.color }} />
                      <span className="flex-1 text-left">{entry.name}</span>
                      <span className={isDark ? "text-zinc-600" : "text-gray-400"}>
                        {entry.pane === "overlay" ? "overlay" : "osc"}
                      </span>
                      <Plus size={12} className={isDark ? "text-zinc-600" : "text-gray-400"} />
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
