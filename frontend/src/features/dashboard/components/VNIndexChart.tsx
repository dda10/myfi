"use client";

import { useEffect, useState, useCallback } from "react";
import { MarketChart } from "@/components/dashboard/MarketChart";
import { fetchOHLCV, type OHLCVBar, type TimeInterval } from "@/lib/chart-engine";

const TIMEFRAMES: { label: string; value: TimeInterval }[] = [
  { label: "1D", value: "1d" },
  { label: "1W", value: "1w" },
  { label: "1M", value: "1M" },
];

export function VNIndexChart() {
  const [interval, setInterval] = useState<TimeInterval>("1d");
  const [data, setData] = useState<OHLCVBar[]>([]);
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const bars = await fetchOHLCV("VNINDEX", interval);
      setData(bars);
    } catch {
      setData([]);
    } finally {
      setLoading(false);
    }
  }, [interval]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-2 flex-shrink-0">
        <div className="flex items-center gap-1.5">
          <div className="w-1.5 h-1.5 rounded-full bg-green-400 animate-pulse" />
          <span className="text-xs font-semibold text-foreground">VN-Index · HOSE</span>
        </div>
        <div className="flex gap-0.5">
          {TIMEFRAMES.map(tf => (
            <button
              key={tf.value}
              onClick={() => setInterval(tf.value)}
              className={`px-2.5 py-1 rounded text-xs font-medium transition ${
                interval === tf.value
                  ? "bg-indigo-600 text-white"
                  : "text-text-muted hover:text-foreground hover:bg-surface"
              }`}
            >
              {tf.label}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1" style={{ minHeight: 380 }}>
        {loading && data.length === 0 ? (
          <div className="w-full h-full bg-surface rounded-xl animate-pulse" />
        ) : data.length === 0 ? (
          <div className="w-full h-full flex items-center justify-center text-text-muted text-sm">
            Không có dữ liệu biểu đồ
          </div>
        ) : (
          <MarketChart data={data} />
        )}
      </div>
    </div>
  );
}
