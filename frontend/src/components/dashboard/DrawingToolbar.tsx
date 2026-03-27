"use client";

import { TrendingUp, Minus, GitBranch, Square, Trash2 } from "lucide-react";
import type { DrawingType } from "@/lib/drawing-tools";

interface DrawingToolbarProps {
  activeTool: DrawingType | null;
  onSelectTool: (tool: DrawingType | null) => void;
  onClearAll: () => void;
}

const TOOLS: { type: DrawingType; label: string; Icon: React.ComponentType<{ className?: string }> }[] = [
  { type: "trendline", label: "Trend Line", Icon: TrendingUp },
  { type: "horizontal", label: "Horizontal", Icon: Minus },
  { type: "fibonacci", label: "Fibonacci", Icon: GitBranch },
  { type: "rectangle", label: "Rectangle", Icon: Square },
];

export function DrawingToolbar({ activeTool, onSelectTool, onClearAll }: DrawingToolbarProps) {
  return (
    <div className="flex items-center gap-1">
      {TOOLS.map(({ type, label, Icon }) => {
        const isActive = activeTool === type;
        return (
          <button
            key={type}
            title={label}
            aria-label={label}
            aria-pressed={isActive}
            onClick={() => onSelectTool(isActive ? null : type)}
            className={`p-1.5 rounded transition-colors ${
              isActive
                ? "bg-indigo-600 text-white"
                : "text-zinc-400 hover:text-white hover:bg-zinc-700"
            }`}
          >
            <Icon className="w-4 h-4" />
          </button>
        );
      })}
      <button
        title="Clear all drawings"
        aria-label="Clear all drawings"
        onClick={onClearAll}
        className="p-1.5 rounded text-zinc-400 hover:text-red-400 hover:bg-zinc-700 transition-colors ml-1"
      >
        <Trash2 className="w-4 h-4" />
      </button>
    </div>
  );
}
