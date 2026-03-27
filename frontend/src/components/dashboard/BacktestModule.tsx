"use client";

import { useState } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as RechartsTooltip,
  ResponsiveContainer,
  ReferenceDot,
} from "recharts";
import { FlaskConical, Play, AlertTriangle, TrendingUp, TrendingDown } from "lucide-react";
import { useI18n } from "@/context/I18nContext";

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types matching backend model/backtest_types.go ---

type ConditionOperator = "LT" | "GT" | "CROSSES_ABOVE" | "CROSSES_BELOW" | "LTE" | "GTE";

interface ConditionOperand {
  type: "indicator" | "price" | "constant";
  indicator?: string;
  field?: string;
  period?: number;
  param2?: number;
  param3?: number;
  paramF?: number;
  constant?: number;
}

interface StrategyCondition {
  left: ConditionOperand;
  operator: ConditionOperator;
  right: ConditionOperand;
}

interface StrategyRule {
  name: string;
  entryConditions: StrategyCondition[];
  exitConditions: StrategyCondition[];
  stopLossPct: number;
  takeProfitPct: number;
}

interface BacktestTrade {
  entryDate: string;
  exitDate: string;
  entryPrice: number;
  exitPrice: number;
  returnPct: number;
  exitReason: string;
  holdingDays: number;
}

interface EquityPoint {
  date: string;
  value: number;
}

interface BacktestResult {
  totalReturn: number;
  winRate: number;
  maxDrawdown: number;
  sharpeRatio: number;
  trades: number;
  avgHoldingPeriod: number;
  equityCurve: EquityPoint[];
  tradeList: BacktestTrade[];
}

// --- Preset strategies (Req 32.6) ---

const INDICATORS = [
  "SMA", "EMA", "RSI", "MACD", "BOLLINGER", "STOCHASTIC", "ADX", "AROON",
  "PARABOLIC_SAR", "SUPERTREND", "VWAP", "VWMA", "WILLIAMS_R", "CMO",
  "ROC", "MOMENTUM", "KELTNER", "ATR", "STDDEV", "OBV", "LINEAR_REG",
];

const OPERATORS: { value: ConditionOperator; label: string }[] = [
  { value: "LT", label: "< Less than" },
  { value: "GT", label: "> Greater than" },
  { value: "LTE", label: "≤ Less or equal" },
  { value: "GTE", label: "≥ Greater or equal" },
  { value: "CROSSES_ABOVE", label: "↑ Crosses above" },
  { value: "CROSSES_BELOW", label: "↓ Crosses below" },
];

const PRESET_STRATEGIES: { label: string; strategy: StrategyRule }[] = [
  {
    label: "RSI Oversold Bounce",
    strategy: {
      name: "RSI Oversold Bounce",
      entryConditions: [{ left: { type: "indicator", indicator: "RSI", field: "value", period: 14 }, operator: "LT", right: { type: "constant", constant: 30 } }],
      exitConditions: [{ left: { type: "indicator", indicator: "RSI", field: "value", period: 14 }, operator: "GT", right: { type: "constant", constant: 70 } }],
      stopLossPct: 0.05,
      takeProfitPct: 0.10,
    },
  },
  {
    label: "MACD Crossover",
    strategy: {
      name: "MACD Crossover",
      entryConditions: [{ left: { type: "indicator", indicator: "MACD", field: "value", period: 12, param2: 26, param3: 9 }, operator: "CROSSES_ABOVE", right: { type: "indicator", indicator: "MACD", field: "signal", period: 12, param2: 26, param3: 9 } }],
      exitConditions: [{ left: { type: "indicator", indicator: "MACD", field: "value", period: 12, param2: 26, param3: 9 }, operator: "CROSSES_BELOW", right: { type: "indicator", indicator: "MACD", field: "signal", period: 12, param2: 26, param3: 9 } }],
      stopLossPct: 0.05,
      takeProfitPct: 0.10,
    },
  },
  {
    label: "Bollinger Band Squeeze",
    strategy: {
      name: "Bollinger Band Squeeze",
      entryConditions: [{ left: { type: "price" }, operator: "LTE", right: { type: "indicator", indicator: "BOLLINGER", field: "lower", period: 20, paramF: 2 } }],
      exitConditions: [{ left: { type: "price" }, operator: "GTE", right: { type: "indicator", indicator: "BOLLINGER", field: "upper", period: 20, paramF: 2 } }],
      stopLossPct: 0.03,
      takeProfitPct: 0.08,
    },
  },
];

function defaultDates() {
  const end = new Date();
  const start = new Date();
  start.setFullYear(start.getFullYear() - 1);
  return { start: start.toISOString().slice(0, 10), end: end.toISOString().slice(0, 10) };
}

// --- Component ---

export function BacktestModule() {
  const { formatPercent, formatCurrency, formatDate } = useI18n();
  const dates = defaultDates();

  const [symbol, setSymbol] = useState("SSI");
  const [startDate, setStartDate] = useState(dates.start);
  const [endDate, setEndDate] = useState(dates.end);
  const [selectedPreset, setSelectedPreset] = useState(0);
  const [customMode, setCustomMode] = useState(false);

  // Custom rule builder state
  const [entryIndicator, setEntryIndicator] = useState("RSI");
  const [entryField, setEntryField] = useState("value");
  const [entryPeriod, setEntryPeriod] = useState(14);
  const [entryOp, setEntryOp] = useState<ConditionOperator>("LT");
  const [entryValue, setEntryValue] = useState(30);
  const [exitIndicator, setExitIndicator] = useState("RSI");
  const [exitField, setExitField] = useState("value");
  const [exitPeriod, setExitPeriod] = useState(14);
  const [exitOp, setExitOp] = useState<ConditionOperator>("GT");
  const [exitValue, setExitValue] = useState(70);
  const [stopLoss, setStopLoss] = useState(5);
  const [takeProfit, setTakeProfit] = useState(10);

  const [result, setResult] = useState<BacktestResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const buildStrategy = (): StrategyRule => {
    if (!customMode) return PRESET_STRATEGIES[selectedPreset].strategy;
    return {
      name: "Custom Strategy",
      entryConditions: [{
        left: { type: "indicator", indicator: entryIndicator, field: entryField, period: entryPeriod },
        operator: entryOp,
        right: { type: "constant", constant: entryValue },
      }],
      exitConditions: [{
        left: { type: "indicator", indicator: exitIndicator, field: exitField, period: exitPeriod },
        operator: exitOp,
        right: { type: "constant", constant: exitValue },
      }],
      stopLossPct: stopLoss / 100,
      takeProfitPct: takeProfit / 100,
    };
  };

  const runBacktest = async () => {
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const token = typeof window !== "undefined" ? localStorage.getItem("myfi-token") : null;
      const res = await fetch(`${API_URL}/api/backtest`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({
          symbol,
          startDate: new Date(startDate).toISOString(),
          endDate: new Date(endDate).toISOString(),
          strategy: buildStrategy(),
        }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => null);
        throw new Error(err?.error ?? `Request failed (${res.status})`);
      }
      setResult(await res.json());
    } catch (e: any) {
      setError(e.message ?? "Backtest failed");
    } finally {
      setLoading(false);
    }
  };

  const equityData = result?.equityCurve?.map((p) => ({
    date: new Date(p.date).toLocaleDateString(),
    value: p.value,
  })) ?? [];

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-zinc-800 flex items-center gap-2">
        <FlaskConical size={18} className="text-cyan-400" />
        <h2 className="text-lg font-semibold text-white">Backtesting</h2>
      </div>

      {/* Config Panel */}
      <div className="p-4 border-b border-zinc-800 bg-zinc-800/20 space-y-3">
        {/* Symbol + Dates */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div>
            <label className="text-zinc-500 text-xs mb-1 block">Symbol</label>
            <input
              value={symbol}
              onChange={(e) => setSymbol(e.target.value.toUpperCase())}
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-cyan-500"
            />
          </div>
          <div>
            <label className="text-zinc-500 text-xs mb-1 block">Start Date</label>
            <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)}
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-cyan-500" />
          </div>
          <div>
            <label className="text-zinc-500 text-xs mb-1 block">End Date</label>
            <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)}
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-cyan-500" />
          </div>
        </div>

        {/* Strategy Selection */}
        <div className="flex items-center gap-2">
          <div className="flex gap-1 bg-zinc-800 rounded-lg p-0.5">
            <button onClick={() => setCustomMode(false)}
              className={`px-3 py-1 text-xs rounded-md transition-colors ${!customMode ? "bg-zinc-700 text-white" : "text-zinc-400 hover:text-zinc-200"}`}>
              Presets
            </button>
            <button onClick={() => setCustomMode(true)}
              className={`px-3 py-1 text-xs rounded-md transition-colors ${customMode ? "bg-zinc-700 text-white" : "text-zinc-400 hover:text-zinc-200"}`}>
              Custom
            </button>
          </div>
        </div>

        {!customMode ? (
          <div className="flex flex-wrap gap-2">
            {PRESET_STRATEGIES.map((p, i) => (
              <button key={p.label} onClick={() => setSelectedPreset(i)}
                className={`px-3 py-1.5 text-xs rounded-lg border transition-colors ${
                  selectedPreset === i ? "border-cyan-500 bg-cyan-950/30 text-cyan-300" : "border-zinc-700 text-zinc-400 hover:text-zinc-200"
                }`}>
                {p.label}
              </button>
            ))}
          </div>
        ) : (
          /* Custom Rule Builder — Req 32.8 */
          <div className="space-y-3">
            <div className="text-xs text-zinc-400 font-medium">Entry Condition</div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <select value={entryIndicator} onChange={(e) => setEntryIndicator(e.target.value)}
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white">
                {INDICATORS.map((ind) => <option key={ind} value={ind}>{ind}</option>)}
              </select>
              <input value={entryField} onChange={(e) => setEntryField(e.target.value)} placeholder="field"
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              <input type="number" value={entryPeriod} onChange={(e) => setEntryPeriod(+e.target.value)} placeholder="period"
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              <select value={entryOp} onChange={(e) => setEntryOp(e.target.value as ConditionOperator)}
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white">
                {OPERATORS.map((op) => <option key={op.value} value={op.value}>{op.label}</option>)}
              </select>
            </div>
            <input type="number" value={entryValue} onChange={(e) => setEntryValue(+e.target.value)} placeholder="Threshold"
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white w-32" />

            <div className="text-xs text-zinc-400 font-medium">Exit Condition</div>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2">
              <select value={exitIndicator} onChange={(e) => setExitIndicator(e.target.value)}
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white">
                {INDICATORS.map((ind) => <option key={ind} value={ind}>{ind}</option>)}
              </select>
              <input value={exitField} onChange={(e) => setExitField(e.target.value)} placeholder="field"
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              <input type="number" value={exitPeriod} onChange={(e) => setExitPeriod(+e.target.value)} placeholder="period"
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              <select value={exitOp} onChange={(e) => setExitOp(e.target.value as ConditionOperator)}
                className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white">
                {OPERATORS.map((op) => <option key={op.value} value={op.value}>{op.label}</option>)}
              </select>
            </div>
            <input type="number" value={exitValue} onChange={(e) => setExitValue(+e.target.value)} placeholder="Threshold"
              className="bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white w-32" />

            <div className="grid grid-cols-2 gap-2">
              <div>
                <label className="text-zinc-500 text-xs">Stop Loss %</label>
                <input type="number" value={stopLoss} onChange={(e) => setStopLoss(+e.target.value)}
                  className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              </div>
              <div>
                <label className="text-zinc-500 text-xs">Take Profit %</label>
                <input type="number" value={takeProfit} onChange={(e) => setTakeProfit(+e.target.value)}
                  className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-2 py-1.5 text-xs text-white" />
              </div>
            </div>
          </div>
        )}

        <button onClick={runBacktest} disabled={loading || !symbol}
          className="flex items-center gap-2 px-4 py-2 bg-cyan-600 hover:bg-cyan-500 disabled:opacity-40 text-white text-sm rounded-lg transition-colors">
          <Play size={14} />
          {loading ? "Running..." : "Run Backtest"}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="px-4 py-3 flex items-center gap-2 text-red-400 text-sm border-b border-zinc-800">
          <AlertTriangle size={14} /> {error}
        </div>
      )}

      {/* Results — Req 32.4, 32.8 */}
      {result && (
        <div className="p-4 space-y-4">
          {/* Metrics */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            <MetricCard label="Total Return" value={formatPercent(result.totalReturn * 100)} positive={result.totalReturn >= 0} />
            <MetricCard label="Win Rate" value={formatPercent(result.winRate * 100)} positive={result.winRate >= 0.5} />
            <MetricCard label="Max Drawdown" value={formatPercent(result.maxDrawdown * 100)} positive={false} />
            <MetricCard label="Sharpe Ratio" value={result.sharpeRatio.toFixed(2)} positive={result.sharpeRatio > 0} />
            <MetricCard label="Trades" value={String(result.trades)} />
            <MetricCard label="Avg Hold" value={`${result.avgHoldingPeriod.toFixed(0)}d`} />
          </div>

          {/* Equity Curve */}
          {equityData.length > 0 && (
            <div className="bg-zinc-800/40 border border-zinc-800 rounded-lg p-4">
              <h3 className="text-sm font-medium text-zinc-300 mb-3">Equity Curve</h3>
              <ResponsiveContainer width="100%" height={250}>
                <LineChart data={equityData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#333" />
                  <XAxis dataKey="date" tick={{ fontSize: 10, fill: "#888" }} />
                  <YAxis tick={{ fontSize: 10, fill: "#888" }} />
                  <RechartsTooltip
                    contentStyle={{ backgroundColor: "#1a1a2e", border: "1px solid #333", borderRadius: 8, fontSize: 12 }}
                    labelStyle={{ color: "#aaa" }}
                  />
                  <Line type="monotone" dataKey="value" stroke="#06b6d4" strokeWidth={1.5} dot={false} />
                  {/* Trade entry markers */}
                  {result.tradeList?.map((t, i) => {
                    const entryPt = equityData.find((p) => p.date === new Date(t.entryDate).toLocaleDateString());
                    return entryPt ? (
                      <ReferenceDot key={`e-${i}`} x={entryPt.date} y={entryPt.value} r={3}
                        fill={t.returnPct >= 0 ? "#22c55e" : "#ef4444"} stroke="none" />
                    ) : null;
                  })}
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}

          {/* Trade List */}
          {result.tradeList && result.tradeList.length > 0 && (
            <div className="bg-zinc-800/40 border border-zinc-800 rounded-lg overflow-hidden">
              <h3 className="text-sm font-medium text-zinc-300 px-4 py-3 border-b border-zinc-700">Trade History</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="text-zinc-500 border-b border-zinc-700">
                      <th className="text-left px-4 py-2">Entry</th>
                      <th className="text-left px-4 py-2">Exit</th>
                      <th className="text-right px-4 py-2">Entry Price</th>
                      <th className="text-right px-4 py-2">Exit Price</th>
                      <th className="text-right px-4 py-2">Return</th>
                      <th className="text-left px-4 py-2">Reason</th>
                      <th className="text-right px-4 py-2">Days</th>
                    </tr>
                  </thead>
                  <tbody>
                    {result.tradeList.slice(0, 20).map((t, i) => (
                      <tr key={i} className="text-zinc-300 border-t border-zinc-800 hover:bg-zinc-800/60">
                        <td className="px-4 py-1.5">{formatDate(t.entryDate)}</td>
                        <td className="px-4 py-1.5">{formatDate(t.exitDate)}</td>
                        <td className="text-right px-4">{formatCurrency(t.entryPrice)}</td>
                        <td className="text-right px-4">{formatCurrency(t.exitPrice)}</td>
                        <td className={`text-right px-4 ${t.returnPct >= 0 ? "text-green-400" : "text-red-400"}`}>
                          {formatPercent(t.returnPct * 100)}
                        </td>
                        <td className="px-4 capitalize">{t.exitReason.replace("_", " ")}</td>
                        <td className="text-right px-4">{t.holdingDays}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function MetricCard({ label, value, positive }: { label: string; value: string; positive?: boolean }) {
  return (
    <div className="bg-zinc-800/60 border border-zinc-700 rounded-lg p-3">
      <span className="text-zinc-500 text-xs">{label}</span>
      <p className={`text-sm font-semibold mt-0.5 ${
        positive === undefined ? "text-white" : positive ? "text-green-400" : "text-red-400"
      }`}>
        {value}
      </p>
    </div>
  );
}
