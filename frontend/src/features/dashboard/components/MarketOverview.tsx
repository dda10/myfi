"use client";

import { TrendingUp, TrendingDown, Minus } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { usePolling, isVNTradingHours } from "@/hooks/usePolling";
import { apiFetch } from "@/lib/api";
import { LineChart, Line, ResponsiveContainer } from "recharts";

interface IndexData {
  name: string;
  value: number;
  change: number;
  changePercent: number;
  history?: { value: number }[];
}

const INDEX_NAMES = ["VNINDEX", "HNX", "UPCOM"] as const;

// Mock sparkline data
const MOCK_HISTORY_UP   = [10,11,10.5,12,11.5,13,12.5,14,13,14.5];
const MOCK_HISTORY_DOWN = [14,13.5,13,12,13,11.5,12,11,10.5,10];
const MOCK_HISTORY_FLAT = [10,10.5,10,11,10.5,10,11,10.5,11,10.5];

function IndexCard({ data, loading, mockHistory }: { data: IndexData | null; loading: boolean; mockHistory: number[] }) {
  const { formatNumber, formatPercent } = useI18n();

  if (loading || !data) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-4 animate-pulse">
        <div className="h-3 w-20 bg-surface rounded mb-3" />
        <div className="h-6 w-28 bg-surface rounded mb-1.5" />
        <div className="h-3 w-24 bg-surface rounded" />
      </div>
    );
  }

  const isUp = data.change > 0;
  const isDown = data.change < 0;
  const colorClass = isUp ? "text-positive" : isDown ? "text-negative" : "text-text-muted";
  const sparkColor = isUp ? "#22c55e" : isDown ? "#ef4444" : "#71717a";
  const historyData = (data.history && data.history.length > 0
    ? data.history
    : mockHistory.map(v => ({ value: v })
  ));

  return (
    <div className="bg-card-bg border border-border-theme rounded-xl p-4 hover:border-border-theme/80 transition">
      <div className="flex items-start justify-between mb-2">
        <div>
          <p className="text-xs text-text-muted font-medium">{data.name}</p>
          <div className={`flex items-center gap-1 mt-0.5 ${colorClass}`}>
            {isUp ? <TrendingUp size={11} /> : isDown ? <TrendingDown size={11} /> : <Minus size={11} />}
            <span className="text-xs font-semibold">
              {isUp ? "+" : ""}{formatPercent(data.changePercent)}
            </span>
            <span className="text-xs text-text-muted">
              {isUp ? "+" : ""}{formatNumber(data.change, 2)}
            </span>
          </div>
        </div>
        {/* Sparkline */}
        <div style={{ width: 80, height: 36 }}>
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={historyData}>
              <Line type="monotone" dataKey="value" stroke={sparkColor} strokeWidth={1.5} dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
      <p className="text-xl font-bold text-foreground tabular-nums">
        {formatNumber(data.value, 2)}
      </p>
    </div>
  );
}

export function MarketOverview() {
  const { t } = useI18n();
  const interval = isVNTradingHours() ? 5 * 60_000 : 30 * 60_000;

  const fetcher = async () => {
    const results = await Promise.all(
      INDEX_NAMES.map(name => apiFetch<IndexData>(`/api/market/index?name=${name}`)),
    );
    const valid = results.filter(Boolean) as IndexData[];
    return valid.length > 0 ? valid : null;
  };

  const { data, loading } = usePolling<IndexData[]>(fetcher, interval);
  const mockHistories = [MOCK_HISTORY_DOWN, MOCK_HISTORY_FLAT, MOCK_HISTORY_UP];

  return (
    <section>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        {INDEX_NAMES.map((name, i) => (
          <IndexCard
            key={name}
            data={data?.[i] ?? null}
            loading={loading && !data}
            mockHistory={mockHistories[i] ?? MOCK_HISTORY_FLAT}
          />
        ))}
      </div>
    </section>
  );
}
