"use client";

import { useState, useEffect, useMemo } from "react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import {
  TrendingUp,
  TrendingDown,
  Loader2,
  AlertCircle,
  Filter,
} from "lucide-react";

// --- Types ---

interface InvestmentIdea {
  id: string;
  symbol: string;
  direction: "buy" | "sell";
  entryPrice: number;
  stopLoss: number;
  takeProfit: number;
  confidence: number;
  reasoning: string;
  createdAt: string;
  accuracy: {
    "1d": number | null;
    "7d": number | null;
    "14d": number | null;
    "30d": number | null;
  };
}

type DirectionFilter = "all" | "buy" | "sell";
type RecencyFilter = "all" | "today" | "week" | "month";

// --- Helpers ---

function isWithinRecency(dateStr: string, recency: RecencyFilter): boolean {
  if (recency === "all") return true;
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const ONE_DAY = 86_400_000;
  if (recency === "today") return diffMs < ONE_DAY;
  if (recency === "week") return diffMs < 7 * ONE_DAY;
  return diffMs < 30 * ONE_DAY;
}

// --- Component ---

export default function IdeasPage() {
  const { t, formatNumber, formatCurrency } = useI18n();

  const [ideas, setIdeas] = useState<InvestmentIdea[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [directionFilter, setDirectionFilter] = useState<DirectionFilter>("all");
  const [recencyFilter, setRecencyFilter] = useState<RecencyFilter>("all");

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      const data = await apiFetch<{ ideas: InvestmentIdea[] }>("/api/ideas");
      if (data?.ideas) {
        setIdeas(data.ideas);
      } else {
        setError(t("error.generic"));
      }
      setLoading(false);
    }
    load();
  }, [t]);

  const filtered = useMemo(() => {
    return ideas
      .filter((idea) => directionFilter === "all" || idea.direction === directionFilter)
      .filter((idea) => isWithinRecency(idea.createdAt, recencyFilter))
      .sort((a, b) => b.confidence - a.confidence);
  }, [ideas, directionFilter, recencyFilter]);

  return (
    <div className="space-y-6 pb-8">
      {/* Header */}
      <h1 className="text-2xl font-bold text-foreground">{t("ideas.title")}</h1>

      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center gap-2 text-sm text-text-muted">
          <Filter className="w-4 h-4" />
        </div>
        {/* Direction filter */}
        <div className="flex rounded-lg border border-border-theme overflow-hidden">
          {(["all", "buy", "sell"] as DirectionFilter[]).map((d) => (
            <button
              key={d}
              onClick={() => setDirectionFilter(d)}
              className={`px-3 py-1.5 text-xs font-medium transition-colors ${
                directionFilter === d
                  ? "bg-accent text-white"
                  : "bg-surface text-text-muted hover:text-foreground"
              }`}
            >
              {d === "all"
                ? t("common.all")
                : d === "buy"
                  ? t("ideas.buy_signal")
                  : t("ideas.sell_signal")}
            </button>
          ))}
        </div>
        {/* Recency filter */}
        <select
          value={recencyFilter}
          onChange={(e) => setRecencyFilter(e.target.value as RecencyFilter)}
          className="rounded-lg bg-background border border-border-theme px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-accent"
        >
          <option value="all">{t("common.all")}</option>
          <option value="today">Today</option>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
        </select>
      </div>

      {/* Loading */}
      {loading && (
        <div className="flex items-center justify-center py-12 text-text-muted">
          <Loader2 className="w-5 h-5 animate-spin mr-2" />
          {t("common.loading")}
        </div>
      )}

      {/* Error */}
      {error && !loading && (
        <div className="flex items-center gap-2 py-8 justify-center text-red-500 text-sm">
          <AlertCircle className="w-4 h-4" />
          {error}
        </div>
      )}

      {/* Ideas Cards */}
      {!loading && !error && (
        <>
          {filtered.length === 0 ? (
            <div className="p-8 rounded-xl bg-surface border border-border-theme text-center text-text-muted">
              {t("common.no_data")}
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
              {filtered.map((idea) => (
                <IdeaCard key={idea.id} idea={idea} />
              ))}
            </div>
          )}
        </>
      )}

      {/* Disclaimer */}
      <p className="text-xs text-text-muted text-center pt-4 border-t border-border-theme">
        {t("chat.disclaimer")}
      </p>
    </div>
  );
}

// --- Idea Card ---

function IdeaCard({ idea }: { idea: InvestmentIdea }) {
  const { t, formatCurrency, formatNumber } = useI18n();
  const isBuy = idea.direction === "buy";

  const riskReward =
    idea.direction === "buy"
      ? (idea.takeProfit - idea.entryPrice) / (idea.entryPrice - idea.stopLoss || 1)
      : (idea.entryPrice - idea.takeProfit) / (idea.stopLoss - idea.entryPrice || 1);

  return (
    <div className="rounded-xl bg-surface border border-border-theme p-4 space-y-3 hover:border-accent/40 transition-colors">
      {/* Header: Symbol + Direction Badge */}
      <div className="flex items-center justify-between">
        <span className="text-lg font-bold text-accent">{idea.symbol}</span>
        <span
          className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold ${
            isBuy
              ? "bg-green-500/15 text-green-500"
              : "bg-red-500/15 text-red-500"
          }`}
        >
          {isBuy ? <TrendingUp className="w-3 h-3" /> : <TrendingDown className="w-3 h-3" />}
          {isBuy ? t("ideas.buy_signal") : t("ideas.sell_signal")}
        </span>
      </div>

      {/* Price Levels */}
      <div className="grid grid-cols-3 gap-2 text-xs">
        <div>
          <div className="text-text-muted">{t("ideas.entry_price")}</div>
          <div className="font-semibold text-foreground">{formatCurrency(idea.entryPrice)}</div>
        </div>
        <div>
          <div className="text-text-muted">{t("ideas.stop_loss")}</div>
          <div className="font-semibold text-red-500">{formatCurrency(idea.stopLoss)}</div>
        </div>
        <div>
          <div className="text-text-muted">{t("ideas.target_price")}</div>
          <div className="font-semibold text-green-500">{formatCurrency(idea.takeProfit)}</div>
        </div>
      </div>

      {/* Confidence Bar */}
      <div className="space-y-1">
        <div className="flex justify-between text-xs">
          <span className="text-text-muted">{t("ideas.confidence")}</span>
          <span className="font-semibold text-foreground">{formatNumber(idea.confidence)}/100</span>
        </div>
        <div className="w-full h-2 rounded-full bg-background overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${
              idea.confidence >= 70
                ? "bg-green-500"
                : idea.confidence >= 40
                  ? "bg-yellow-500"
                  : "bg-red-500"
            }`}
            style={{ width: `${Math.min(idea.confidence, 100)}%` }}
          />
        </div>
      </div>

      {/* Risk/Reward */}
      <div className="flex justify-between text-xs">
        <span className="text-text-muted">{t("ideas.risk_reward")}</span>
        <span className="font-semibold text-foreground">{formatNumber(Math.abs(riskReward), 2)}:1</span>
      </div>

      {/* Reasoning */}
      {idea.reasoning && (
        <p className="text-xs text-text-muted leading-relaxed line-clamp-3">{idea.reasoning}</p>
      )}

      {/* Historical Accuracy */}
      <div className="pt-2 border-t border-border-theme">
        <div className="text-xs text-text-muted mb-1.5">Historical Accuracy</div>
        <div className="grid grid-cols-4 gap-2 text-center text-xs">
          {(["1d", "7d", "14d", "30d"] as const).map((period) => {
            const val = idea.accuracy[period];
            return (
              <div key={period}>
                <div className="text-text-muted">{period}</div>
                <div
                  className={`font-semibold ${
                    val === null
                      ? "text-text-muted"
                      : val >= 0
                        ? "text-green-500"
                        : "text-red-500"
                  }`}
                >
                  {val !== null ? `${val >= 0 ? "+" : ""}${formatNumber(val, 1)}%` : "—"}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
