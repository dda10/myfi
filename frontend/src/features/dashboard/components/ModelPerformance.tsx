"use client";

import { useEffect, useState } from "react";
import { Brain, AlertTriangle } from "lucide-react";
import {
  LineChart, Line, BarChart, Bar, ScatterChart, Scatter,
  XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend,
} from "recharts";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

// --- Types (Req 41.8) ---

interface AgentAccuracyPoint {
  date: string;
  technicalAnalyst: number;
  newsAnalyst: number;
  investmentAdvisor: number;
  strategyBuilder: number;
}

interface RegimeAccuracy {
  regime: string;
  accuracy: number;
  sampleSize: number;
}

interface ConfidenceBucket {
  bucket: string;
  hitRate: number;
  count: number;
}

interface PredictedVsActual {
  predicted: number;
  actual: number;
  symbol: string;
}

interface PerformanceData {
  agentAccuracy: AgentAccuracyPoint[];
  regimeAccuracy: RegimeAccuracy[];
  confidenceBuckets: ConfidenceBucket[];
  predictedVsActual: PredictedVsActual[];
}

const AGENT_COLORS: Record<string, string> = {
  technicalAnalyst: "#6366f1",
  newsAnalyst: "#f59e0b",
  investmentAdvisor: "#22c55e",
  strategyBuilder: "#ef4444",
};

const AGENT_LABELS: Record<string, string> = {
  technicalAnalyst: "Technical Analyst",
  newsAnalyst: "News Analyst",
  investmentAdvisor: "Investment Advisor",
  strategyBuilder: "Strategy Builder",
};

// --- Component (Req 41.8) ---

export function ModelPerformance() {
  const { t } = useI18n();
  const [data, setData] = useState<PerformanceData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      const res = await apiFetch<PerformanceData>("/api/feedback/performance");
      if (cancelled) return;
      if (res) {
        setData(res);
        setError(null);
      } else {
        setError("Failed to load performance data");
      }
      setLoading(false);
    }
    load();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 animate-pulse">
        <div className="h-6 bg-zinc-800 rounded w-56 mb-4" />
        <div className="grid grid-cols-2 gap-4">
          {[1, 2, 3, 4].map((i) => <div key={i} className="h-48 bg-zinc-800 rounded" />)}
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-6">
        <div className="flex items-center gap-2 text-red-400">
          <AlertTriangle size={16} />
          <span className="text-sm">{error ?? t("model_perf.no_data")}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center gap-2">
        <Brain size={18} className="text-purple-400" />
        <h2 className="text-lg font-semibold text-white">{t("model_perf.title")}</h2>
      </div>

      <div className="p-4 grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 1. Agent Accuracy Trends */}
        <div>
          <h3 className="text-sm font-medium text-zinc-300 mb-3">{t("model_perf.agent_accuracy")}</h3>
          <ResponsiveContainer width="100%" height={220}>
            <LineChart data={data.agentAccuracy}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="date" tick={{ fill: "#71717a", fontSize: 10 }} />
              <YAxis tick={{ fill: "#71717a", fontSize: 10 }} domain={[0, 100]} />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                labelStyle={{ color: "#a1a1aa" }}
              />
              <Legend wrapperStyle={{ fontSize: 11 }} />
              {Object.entries(AGENT_COLORS).map(([key, color]) => (
                <Line
                  key={key}
                  type="monotone"
                  dataKey={key}
                  name={AGENT_LABELS[key]}
                  stroke={color}
                  strokeWidth={2}
                  dot={false}
                />
              ))}
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* 2. Alpha Mining Accuracy by Regime */}
        <div>
          <h3 className="text-sm font-medium text-zinc-300 mb-3">{t("model_perf.regime")}</h3>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={data.regimeAccuracy}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="regime" tick={{ fill: "#71717a", fontSize: 10 }} />
              <YAxis tick={{ fill: "#71717a", fontSize: 10 }} domain={[0, 100]} />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                formatter={(value: unknown) => [`${Number(value).toFixed(1)}%`, "Accuracy"]}
              />
              <Bar dataKey="accuracy" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* 3. Hit Rate by Confidence Bucket */}
        <div>
          <h3 className="text-sm font-medium text-zinc-300 mb-3">{t("model_perf.hit_rate")}</h3>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart data={data.confidenceBuckets}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="bucket" tick={{ fill: "#71717a", fontSize: 10 }} />
              <YAxis tick={{ fill: "#71717a", fontSize: 10 }} domain={[0, 100]} />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                formatter={(value: unknown, name: unknown) => {
                  if (name === "hitRate") return [`${Number(value).toFixed(1)}%`, "Hit Rate"];
                  return [String(value), String(name)];
                }}
              />
              <Bar dataKey="hitRate" fill="#22c55e" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* 4. Predicted vs Actual Returns */}
        <div>
          <h3 className="text-sm font-medium text-zinc-300 mb-3">{t("model_perf.predicted_vs_actual")}</h3>
          <ResponsiveContainer width="100%" height={220}>
            <ScatterChart>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis
                type="number"
                dataKey="predicted"
                name="Predicted"
                tick={{ fill: "#71717a", fontSize: 10 }}
                label={{ value: "Predicted %", position: "insideBottom", offset: -5, fill: "#71717a", fontSize: 10 }}
              />
              <YAxis
                type="number"
                dataKey="actual"
                name="Actual"
                tick={{ fill: "#71717a", fontSize: 10 }}
                label={{ value: "Actual %", angle: -90, position: "insideLeft", fill: "#71717a", fontSize: 10 }}
              />
              <Tooltip
                contentStyle={{ backgroundColor: "#18181b", border: "1px solid #3f3f46", borderRadius: 8 }}
                formatter={(value: unknown, name: unknown) => [`${Number(value).toFixed(2)}%`, String(name)]}
              />
              <Scatter data={data.predictedVsActual} fill="#6366f1" />
            </ScatterChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}
