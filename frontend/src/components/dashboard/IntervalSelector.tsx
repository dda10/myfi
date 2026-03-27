"use client";

import { TIME_INTERVALS, type TimeInterval } from "@/lib/chart-engine";

interface IntervalSelectorProps {
  selected: TimeInterval;
  onChange: (interval: TimeInterval) => void;
}

export function IntervalSelector({ selected, onChange }: IntervalSelectorProps) {
  return (
    <div className="flex items-center gap-1">
      {TIME_INTERVALS.map(({ label, value }) => (
        <button
          key={value}
          onClick={() => onChange(value)}
          className={`px-3 py-1 rounded text-sm transition ${
            selected === value
              ? "bg-indigo-500/20 text-indigo-400 font-medium"
              : "bg-zinc-800 hover:bg-zinc-700 text-zinc-300 dark:bg-zinc-800 dark:hover:bg-zinc-700 dark:text-zinc-300"
          }`}
        >
          {label}
        </button>
      ))}
    </div>
  );
}
