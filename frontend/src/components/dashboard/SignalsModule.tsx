"use client";

import { useState, useEffect } from "react";
import { 
  TrendingUp, TrendingDown, Target, AlertTriangle, 
  Play, BarChart3, History, Settings2, RefreshCw,
  ChevronDown, ChevronUp, Calendar, Zap
} from "lucide-react";
import { useApp } from "@/context/AppContext";
import { RecommendationHistory } from "@/components/dashboard/RecommendationHistory";

// Types
interface FactorScores {
  momentum: number;
  trend: number;
  volume: number;
  fundamental: number;
  sector: number;
}

interface TradingSignal {
  symbol: string;
  direction: "long" | "short";
  strength: "strong" | "moderate" | "weak";
  compositeScore: number;
  factors: FactorScores;
  currentPrice: number;
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  riskReward: number;
  exchange: string;
  sectorName: string;
  reasoning: string[];
  generatedAt: string;
}

interface SignalScanResult {
  signals: TradingSignal[];
  totalScanned: number;
  passedFilter: number;
  generatedAt: string;
}

interface AccuracySummary {
  totalRecommendations: number;
  winRate1d: number;
  winRate7d: number;
  winRate14d: number;
  winRate30d: number;
  avgReturn1d: number;
  avgReturn7d: number;
  avgReturn14d: number;
  avgReturn30d: number;
}

interface BacktestResult {
  totalReturn: number;
  winRate: number;
  avgWin: number;
  avgLoss: number;
  profitFactor: number;
  maxDrawdown: number;
  sharpeRatio: number;
  numTrades: number;
  avgHoldingDays: number;
  avgCompositeScore: number;
}

type TabType = "signals" | "accuracy" | "backtest" | "history";

export function SignalsModule() {
  const { setActiveSymbol, setActiveTab } = useApp();
  const [activeView, setActiveView] = useState<TabType>("signals");
  
  // Signals state
  const [signals, setSignals] = useState<TradingSignal[]>([]);
  const [scanLoading, setScanLoading] = useState(false);
  const [scanMeta, setScanMeta] = useState({ totalScanned: 0, passedFilter: 0 });
  
  // Accuracy state
  const [accuracy, setAccuracy] = useState<AccuracySummary | null>(null);
  const [accuracyLoading, setAccuracyLoading] = useState(false);
  
  // Backtest state
  const [backtestResult, setBacktestResult] = useState<BacktestResult | null>(null);
  const [backtestLoading, setBacktestLoading] = useState(false);
  const [backtestConfig, setBacktestConfig] = useState({
    startDate: getDefaultStartDate(),
    endDate: getDefaultEndDate(),
    topN: 5,
    scanFreq: 5,
    maxHold: 20
  });

  // Load initial data
  useEffect(() => {
    if (activeView === "signals") fetchSignals();
    if (activeView === "accuracy") fetchAccuracy();
  }, [activeView]);

  const fetchSignals = async () => {
    setScanLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/signals/scan");
      const json = await res.json();
      if (json.signals) {
        setSignals(json.signals);
        setScanMeta({ totalScanned: json.totalScanned, passedFilter: json.passedFilter });
      }
    } catch (e) {
      console.error("Failed to fetch signals", e);
    } finally {
      setScanLoading(false);
    }
  };

  const fetchAccuracy = async () => {
    setAccuracyLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/recommendations/summary");
      const json = await res.json();
      setAccuracy(json);
    } catch (e) {
      console.error("Failed to fetch accuracy", e);
    } finally {
      setAccuracyLoading(false);
    }
  };

  const runBacktest = async () => {
    setBacktestLoading(true);
    try {
      const res = await fetch("http://localhost:8080/api/signals/backtest", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          startDate: new Date(backtestConfig.startDate).toISOString(),
          endDate: new Date(backtestConfig.endDate).toISOString(),
          topN: backtestConfig.topN,
          scanFreq: backtestConfig.scanFreq,
          maxHold: backtestConfig.maxHold
        })
      });
      const json = await res.json();
      setBacktestResult(json);
    } catch (e) {
      console.error("Failed to run backtest", e);
    } finally {
      setBacktestLoading(false);
    }
  };

  const handleSymbolClick = (symbol: string) => {
    setActiveSymbol(symbol);
    setActiveTab("Markets");
  };

  return (
    <div className="flex flex-col gap-6 w-full min-h-[800px] bg-white text-zinc-900 rounded-2xl overflow-hidden shadow-2xl p-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center border-b border-zinc-200 pb-4">
        <div>
          <h1 className="text-2xl font-bold text-blue-600 uppercase tracking-tight flex items-center gap-2">
            <Zap className="w-6 h-6" /> AI Signal Engine
          </h1>
          <p className="text-sm text-zinc-500 mt-1">
            Systematic factor-based stock recommendations with backtesting
          </p>
        </div>
        
        {/* Tab Switcher */}
        <div className="flex gap-2 mt-4 sm:mt-0">
          <TabButton 
            active={activeView === "signals"} 
            onClick={() => setActiveView("signals")}
            icon={<TrendingUp size={16} />}
            label="Live Signals"
          />
          <TabButton 
            active={activeView === "accuracy"} 
            onClick={() => setActiveView("accuracy")}
            icon={<BarChart3 size={16} />}
            label="Accuracy"
          />
          <TabButton 
            active={activeView === "backtest"} 
            onClick={() => setActiveView("backtest")}
            icon={<History size={16} />}
            label="Backtest"
          />
          <TabButton 
            active={activeView === "history"} 
            onClick={() => setActiveView("history")}
            icon={<Calendar size={16} />}
            label="History"
          />
        </div>
      </div>

      {/* Content */}
      {activeView === "signals" && (
        <SignalsView 
          signals={signals}
          loading={scanLoading}
          meta={scanMeta}
          onRefresh={fetchSignals}
          onSymbolClick={handleSymbolClick}
        />
      )}
      
      {activeView === "accuracy" && (
        <AccuracyView 
          accuracy={accuracy}
          loading={accuracyLoading}
          onRefresh={fetchAccuracy}
        />
      )}
      
      {activeView === "backtest" && (
        <BacktestView 
          result={backtestResult}
          loading={backtestLoading}
          config={backtestConfig}
          setConfig={setBacktestConfig}
          onRun={runBacktest}
        />
      )}

      {activeView === "history" && (
        <RecommendationHistory />
      )}
    </div>
  );
}


// Tab Button Component
function TabButton({ active, onClick, icon, label }: { 
  active: boolean; 
  onClick: () => void; 
  icon: React.ReactNode; 
  label: string;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition ${
        active 
          ? "bg-blue-100 text-blue-700" 
          : "bg-zinc-100 text-zinc-600 hover:bg-zinc-200"
      }`}
    >
      {icon}
      {label}
    </button>
  );
}

// Signals View Component
function SignalsView({ signals, loading, meta, onRefresh, onSymbolClick }: {
  signals: TradingSignal[];
  loading: boolean;
  meta: { totalScanned: number; passedFilter: number };
  onRefresh: () => void;
  onSymbolClick: (symbol: string) => void;
}) {
  const [expandedSignal, setExpandedSignal] = useState<string | null>(null);

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-4">
          <span className="text-sm text-zinc-600">
            Scanned <span className="font-bold text-zinc-900">{meta.totalScanned}</span> stocks, 
            <span className="font-bold text-emerald-600 ml-1">{meta.passedFilter}</span> passed filter
          </span>
        </div>
        <button
          onClick={onRefresh}
          disabled={loading}
          className="flex items-center gap-2 bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white px-4 py-2 rounded-lg text-sm font-medium transition"
        >
          <RefreshCw size={16} className={loading ? "animate-spin" : ""} />
          {loading ? "Scanning..." : "Refresh Scan"}
        </button>
      </div>

      {/* Signals Grid */}
      {loading ? (
        <div className="flex items-center justify-center h-64 text-zinc-500">
          <RefreshCw className="animate-spin mr-2" /> Scanning market...
        </div>
      ) : signals.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-64 text-zinc-500">
          <AlertTriangle size={48} className="mb-4 text-zinc-300" />
          <p>No signals found. Try refreshing or adjusting filters.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {signals.map((signal) => (
            <SignalCard 
              key={signal.symbol}
              signal={signal}
              expanded={expandedSignal === signal.symbol}
              onToggle={() => setExpandedSignal(
                expandedSignal === signal.symbol ? null : signal.symbol
              )}
              onClick={() => onSymbolClick(signal.symbol)}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// Signal Card Component
function SignalCard({ signal, expanded, onToggle, onClick }: {
  signal: TradingSignal;
  expanded: boolean;
  onToggle: () => void;
  onClick: () => void;
}) {
  const strengthColors = {
    strong: "bg-emerald-100 text-emerald-700 border-emerald-300",
    moderate: "bg-blue-100 text-blue-700 border-blue-300",
    weak: "bg-zinc-100 text-zinc-600 border-zinc-300"
  };

  const riskPct = ((signal.entryPrice - signal.stopLoss) / signal.entryPrice * 100).toFixed(1);
  const rewardPct = ((signal.takeProfit - signal.entryPrice) / signal.entryPrice * 100).toFixed(1);

  return (
    <div className="bg-white border border-zinc-200 rounded-xl shadow-sm hover:shadow-md transition overflow-hidden">
      {/* Header */}
      <div 
        className="p-4 cursor-pointer hover:bg-zinc-50 transition"
        onClick={onClick}
      >
        <div className="flex justify-between items-start mb-3">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-zinc-100 flex items-center justify-center text-sm font-bold text-zinc-700 border border-zinc-300">
              {signal.symbol.slice(0, 2)}
            </div>
            <div>
              <div className="font-bold text-zinc-900">{signal.symbol}</div>
              <div className="text-xs text-zinc-500">{signal.exchange} • {signal.sectorName}</div>
            </div>
          </div>
          <div className={`px-2 py-1 rounded-md text-xs font-bold border ${strengthColors[signal.strength]}`}>
            {signal.strength.toUpperCase()}
          </div>
        </div>

        {/* Score Bar */}
        <div className="mb-3">
          <div className="flex justify-between text-xs mb-1">
            <span className="text-zinc-500">Composite Score</span>
            <span className="font-bold text-zinc-900">{signal.compositeScore.toFixed(0)}/100</span>
          </div>
          <div className="h-2 bg-zinc-200 rounded-full overflow-hidden">
            <div 
              className={`h-full rounded-full transition-all ${
                signal.compositeScore >= 75 ? "bg-emerald-500" :
                signal.compositeScore >= 60 ? "bg-blue-500" : "bg-zinc-400"
              }`}
              style={{ width: `${signal.compositeScore}%` }}
            />
          </div>
        </div>

        {/* Price Levels */}
        <div className="grid grid-cols-3 gap-2 text-center">
          <div className="bg-red-50 rounded-lg p-2">
            <div className="text-[10px] text-red-600 font-medium">STOP LOSS</div>
            <div className="text-sm font-bold text-red-700">{signal.stopLoss.toLocaleString()}</div>
            <div className="text-[10px] text-red-500">-{riskPct}%</div>
          </div>
          <div className="bg-zinc-50 rounded-lg p-2">
            <div className="text-[10px] text-zinc-600 font-medium">ENTRY</div>
            <div className="text-sm font-bold text-zinc-900">{signal.entryPrice.toLocaleString()}</div>
            <div className="text-[10px] text-zinc-500">Current</div>
          </div>
          <div className="bg-emerald-50 rounded-lg p-2">
            <div className="text-[10px] text-emerald-600 font-medium">TARGET</div>
            <div className="text-sm font-bold text-emerald-700">{signal.takeProfit.toLocaleString()}</div>
            <div className="text-[10px] text-emerald-500">+{rewardPct}%</div>
          </div>
        </div>
      </div>

      {/* Expand Toggle */}
      <button
        onClick={(e) => { e.stopPropagation(); onToggle(); }}
        className="w-full py-2 bg-zinc-50 hover:bg-zinc-100 transition flex items-center justify-center gap-1 text-xs text-zinc-600"
      >
        {expanded ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        {expanded ? "Hide Details" : "Show Factors"}
      </button>

      {/* Expanded Details */}
      {expanded && (
        <div className="p-4 border-t border-zinc-200 bg-zinc-50">
          <div className="text-xs font-medium text-zinc-700 mb-2">Factor Breakdown</div>
          <div className="space-y-2">
            <FactorBar label="Momentum" value={signal.factors.momentum} />
            <FactorBar label="Trend" value={signal.factors.trend} />
            <FactorBar label="Volume" value={signal.factors.volume} />
            <FactorBar label="Fundamental" value={signal.factors.fundamental} />
            <FactorBar label="Sector" value={signal.factors.sector} />
          </div>
          
          {signal.reasoning && signal.reasoning.length > 0 && (
            <div className="mt-3 pt-3 border-t border-zinc-200">
              <div className="text-xs font-medium text-zinc-700 mb-1">Key Reasons</div>
              <ul className="text-xs text-zinc-600 space-y-1">
                {signal.reasoning.slice(0, 4).map((r, i) => (
                  <li key={i} className="flex items-start gap-1">
                    <span className="text-emerald-500">•</span> {r}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function FactorBar({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-xs text-zinc-600 w-24">{label}</span>
      <div className="flex-1 h-1.5 bg-zinc-200 rounded-full overflow-hidden">
        <div 
          className="h-full bg-blue-500 rounded-full"
          style={{ width: `${value}%` }}
        />
      </div>
      <span className="text-xs font-medium text-zinc-700 w-8 text-right">{value.toFixed(0)}</span>
    </div>
  );
}


// Accuracy View Component
function AccuracyView({ accuracy, loading, onRefresh }: {
  accuracy: AccuracySummary | null;
  loading: boolean;
  onRefresh: () => void;
}) {
  if (loading) {
    return (
      <div className="flex items-center justify-center h-64 text-zinc-500">
        <RefreshCw className="animate-spin mr-2" /> Loading accuracy data...
      </div>
    );
  }

  if (!accuracy || accuracy.totalRecommendations === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-zinc-500">
        <BarChart3 size={48} className="mb-4 text-zinc-300" />
        <p>No recommendation history yet.</p>
        <p className="text-sm mt-2">Start using signals to track accuracy over time.</p>
      </div>
    );
  }

  const periods = [
    { label: "1 Day", winRate: accuracy.winRate1d, avgReturn: accuracy.avgReturn1d },
    { label: "7 Days", winRate: accuracy.winRate7d, avgReturn: accuracy.avgReturn7d },
    { label: "14 Days", winRate: accuracy.winRate14d, avgReturn: accuracy.avgReturn14d },
    { label: "30 Days", winRate: accuracy.winRate30d, avgReturn: accuracy.avgReturn30d },
  ];

  return (
    <div className="flex flex-col gap-6">
      {/* Summary Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatCard 
          label="Total Recommendations" 
          value={accuracy.totalRecommendations.toString()} 
          icon={<Target size={20} />}
        />
        <StatCard 
          label="Best Win Rate" 
          value={`${Math.max(accuracy.winRate1d, accuracy.winRate7d, accuracy.winRate14d, accuracy.winRate30d).toFixed(1)}%`}
          icon={<TrendingUp size={20} />}
          positive
        />
        <StatCard 
          label="Best Avg Return" 
          value={`${Math.max(accuracy.avgReturn1d, accuracy.avgReturn7d, accuracy.avgReturn14d, accuracy.avgReturn30d).toFixed(2)}%`}
          icon={<BarChart3 size={20} />}
          positive
        />
        <button
          onClick={onRefresh}
          className="flex flex-col items-center justify-center bg-blue-50 hover:bg-blue-100 rounded-xl p-4 transition"
        >
          <RefreshCw size={20} className="text-blue-600 mb-1" />
          <span className="text-sm font-medium text-blue-700">Refresh</span>
        </button>
      </div>

      {/* Period Breakdown */}
      <div className="bg-zinc-50 rounded-xl p-6">
        <h3 className="font-bold text-zinc-800 mb-4">Performance by Holding Period</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          {periods.map((period) => (
            <div key={period.label} className="bg-white rounded-lg p-4 border border-zinc-200">
              <div className="text-sm text-zinc-500 mb-2">{period.label}</div>
              <div className="flex justify-between items-end">
                <div>
                  <div className="text-2xl font-bold text-zinc-900">
                    {period.winRate.toFixed(1)}%
                  </div>
                  <div className="text-xs text-zinc-500">Win Rate</div>
                </div>
                <div className="text-right">
                  <div className={`text-lg font-bold ${period.avgReturn >= 0 ? "text-emerald-600" : "text-red-600"}`}>
                    {period.avgReturn >= 0 ? "+" : ""}{period.avgReturn.toFixed(2)}%
                  </div>
                  <div className="text-xs text-zinc-500">Avg Return</div>
                </div>
              </div>
              {/* Win Rate Bar */}
              <div className="mt-3 h-2 bg-zinc-200 rounded-full overflow-hidden">
                <div 
                  className={`h-full rounded-full ${period.winRate >= 50 ? "bg-emerald-500" : "bg-red-400"}`}
                  style={{ width: `${Math.min(period.winRate, 100)}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, icon, positive }: {
  label: string;
  value: string;
  icon: React.ReactNode;
  positive?: boolean;
}) {
  return (
    <div className="bg-zinc-50 rounded-xl p-4 flex items-center gap-3">
      <div className={`p-2 rounded-lg ${positive ? "bg-emerald-100 text-emerald-600" : "bg-zinc-200 text-zinc-600"}`}>
        {icon}
      </div>
      <div>
        <div className="text-xs text-zinc-500">{label}</div>
        <div className="text-xl font-bold text-zinc-900">{value}</div>
      </div>
    </div>
  );
}


// Backtest View Component
function BacktestView({ result, loading, config, setConfig, onRun }: {
  result: BacktestResult | null;
  loading: boolean;
  config: {
    startDate: string;
    endDate: string;
    topN: number;
    scanFreq: number;
    maxHold: number;
  };
  setConfig: (config: any) => void;
  onRun: () => void;
}) {
  return (
    <div className="flex flex-col gap-6">
      {/* Config Panel */}
      <div className="bg-zinc-50 rounded-xl p-6">
        <h3 className="font-bold text-zinc-800 mb-4 flex items-center gap-2">
          <Settings2 size={18} /> Backtest Configuration
        </h3>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Start Date</label>
            <input
              type="date"
              value={config.startDate}
              onChange={(e) => setConfig({ ...config, startDate: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">End Date</label>
            <input
              type="date"
              value={config.endDate}
              onChange={(e) => setConfig({ ...config, endDate: e.target.value })}
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Top N Signals</label>
            <input
              type="number"
              value={config.topN}
              onChange={(e) => setConfig({ ...config, topN: parseInt(e.target.value) || 5 })}
              min={1}
              max={20}
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Scan Frequency (days)</label>
            <input
              type="number"
              value={config.scanFreq}
              onChange={(e) => setConfig({ ...config, scanFreq: parseInt(e.target.value) || 5 })}
              min={1}
              max={20}
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="text-xs text-zinc-600 block mb-1">Max Hold (days)</label>
            <input
              type="number"
              value={config.maxHold}
              onChange={(e) => setConfig({ ...config, maxHold: parseInt(e.target.value) || 20 })}
              min={1}
              max={60}
              className="w-full px-3 py-2 border border-zinc-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>
        <button
          onClick={onRun}
          disabled={loading}
          className="mt-4 flex items-center gap-2 bg-blue-500 hover:bg-blue-600 disabled:bg-blue-300 text-white px-6 py-2 rounded-lg text-sm font-medium transition"
        >
          <Play size={16} />
          {loading ? "Running Backtest..." : "Run Backtest"}
        </button>
      </div>

      {/* Results */}
      {loading ? (
        <div className="flex items-center justify-center h-64 text-zinc-500">
          <RefreshCw className="animate-spin mr-2" /> Running historical simulation...
        </div>
      ) : result ? (
        <div className="space-y-6">
          {/* Key Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <MetricCard 
              label="Total Return" 
              value={`${result.totalReturn >= 0 ? "+" : ""}${result.totalReturn.toFixed(2)}%`}
              positive={result.totalReturn >= 0}
              large
            />
            <MetricCard 
              label="Win Rate" 
              value={`${result.winRate.toFixed(1)}%`}
              positive={result.winRate >= 50}
              large
            />
            <MetricCard 
              label="Sharpe Ratio" 
              value={result.sharpeRatio.toFixed(2)}
              positive={result.sharpeRatio >= 1}
              large
            />
            <MetricCard 
              label="Max Drawdown" 
              value={`-${result.maxDrawdown.toFixed(2)}%`}
              positive={result.maxDrawdown < 20}
              large
            />
          </div>

          {/* Secondary Metrics */}
          <div className="bg-zinc-50 rounded-xl p-6">
            <h3 className="font-bold text-zinc-800 mb-4">Detailed Statistics</h3>
            <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
              <div>
                <div className="text-xs text-zinc-500">Total Trades</div>
                <div className="text-lg font-bold text-zinc-900">{result.numTrades}</div>
              </div>
              <div>
                <div className="text-xs text-zinc-500">Avg Win</div>
                <div className="text-lg font-bold text-emerald-600">+{result.avgWin.toFixed(2)}%</div>
              </div>
              <div>
                <div className="text-xs text-zinc-500">Avg Loss</div>
                <div className="text-lg font-bold text-red-600">-{result.avgLoss.toFixed(2)}%</div>
              </div>
              <div>
                <div className="text-xs text-zinc-500">Profit Factor</div>
                <div className={`text-lg font-bold ${result.profitFactor >= 1.5 ? "text-emerald-600" : "text-zinc-900"}`}>
                  {result.profitFactor.toFixed(2)}
                </div>
              </div>
              <div>
                <div className="text-xs text-zinc-500">Avg Holding</div>
                <div className="text-lg font-bold text-zinc-900">{result.avgHoldingDays.toFixed(1)} days</div>
              </div>
            </div>
          </div>

          {/* Interpretation */}
          <div className="bg-blue-50 rounded-xl p-4 border border-blue-200">
            <h4 className="font-medium text-blue-800 mb-2">Strategy Assessment</h4>
            <p className="text-sm text-blue-700">
              {getStrategyAssessment(result)}
            </p>
          </div>
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center h-64 text-zinc-500">
          <History size={48} className="mb-4 text-zinc-300" />
          <p>Configure parameters and run a backtest to see results.</p>
        </div>
      )}
    </div>
  );
}

function MetricCard({ label, value, positive, large }: {
  label: string;
  value: string;
  positive?: boolean;
  large?: boolean;
}) {
  return (
    <div className={`bg-white rounded-xl p-4 border ${positive ? "border-emerald-200" : "border-red-200"}`}>
      <div className="text-xs text-zinc-500 mb-1">{label}</div>
      <div className={`font-bold ${large ? "text-2xl" : "text-lg"} ${positive ? "text-emerald-600" : "text-red-600"}`}>
        {value}
      </div>
    </div>
  );
}

function getStrategyAssessment(result: BacktestResult): string {
  const parts: string[] = [];
  
  if (result.totalReturn > 20) {
    parts.push("Strong positive returns.");
  } else if (result.totalReturn > 0) {
    parts.push("Modest positive returns.");
  } else {
    parts.push("Negative returns - consider adjusting parameters.");
  }

  if (result.winRate >= 55) {
    parts.push("Good win rate above 55%.");
  } else if (result.winRate >= 45) {
    parts.push("Win rate is acceptable but could improve.");
  } else {
    parts.push("Low win rate - signals may need refinement.");
  }

  if (result.sharpeRatio >= 1.5) {
    parts.push("Excellent risk-adjusted returns (Sharpe > 1.5).");
  } else if (result.sharpeRatio >= 1) {
    parts.push("Good risk-adjusted returns.");
  } else {
    parts.push("Risk-adjusted returns need improvement.");
  }

  if (result.maxDrawdown < 15) {
    parts.push("Drawdown well controlled.");
  } else if (result.maxDrawdown < 25) {
    parts.push("Moderate drawdown risk.");
  } else {
    parts.push("High drawdown - consider tighter stops.");
  }

  return parts.join(" ");
}

// Helper functions
function getDefaultStartDate(): string {
  const date = new Date();
  date.setFullYear(date.getFullYear() - 1);
  return date.toISOString().split("T")[0];
}

function getDefaultEndDate(): string {
  return new Date().toISOString().split("T")[0];
}
