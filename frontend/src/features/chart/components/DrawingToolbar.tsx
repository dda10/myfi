"use client";

import { TrendingUp, Minus, GitBranch, Square, Trash2 } from "lucide-react";
import type { DrawingType } from "@/lib/drawing-tools";
import { useTheme } from "@/context/ThemeContext";
import { useI18n } from "@/context/I18nContext";

interface DrawingToolbarProps {
  activeTool: DrawingType | null;
  onSelectTool: (tool: DrawingType | null) => void;
  onClearAll: () => void;
}

const TOOLS: { type: DrawingType; i18nKey: string; Icon: React.ComponentType<{ className?: string }> }[] = [
  { type: "trendline", i18nKey: "chart.drawing.trendline", Icon: TrendingUp },
  { type: "horizontal", i18nKey: "chart.drawing.horizontal", Icon: Minus },
  { type: "fibonacci", i18nKey: "chart.drawing.fibonacci", Icon: GitBranch },
  { type: "rectangle", i18nKey: "chart.drawing.rectangle", Icon: Square },
];

export function DrawingToolbar({ activeTool, onSelectTool, onClearAll }: DrawingToolbarProps) {
  const { theme } = useTheme();
  const { t } = useI18n();
  const isDark = theme === "dark";

  return (
    <div className="flex items-center gap-1">
      {TOOLS.map(({ type, i18nKey, Icon }) => {
        const isActive = activeTool === type;
        return (
          <button
            key={type}
            title={t(i18nKey)}
            aria-label={t(i18nKey)}
            aria-pressed={isActive}
            onClick={() => onSelectTool(isActive ? null : type)}
            className={`p-1.5 rounded transition-colors ${
              isActive
                ? "bg-indigo-600 text-white"
                : isDark
                  ? "text-zinc-400 hover:text-white hover:bg-zinc-700"
                  : "text-gray-500 hover:text-gray-800 hover:bg-gray-200"
            }`}
          >
            <Icon className="w-4 h-4" />
          </button>
        );
      })}
      <button
        title={t("chart.drawing.clear_all")}
        aria-label={t("chart.drawing.clear_all")}
        onClick={onClearAll}
        className={`p-1.5 rounded transition-colors ml-1 ${
          isDark
            ? "text-zinc-400 hover:text-red-400 hover:bg-zinc-700"
            : "text-gray-500 hover:text-red-500 hover:bg-gray-200"
        }`}
      >
        <Trash2 className="w-4 h-4" />
      </button>
    </div>
  );
}
