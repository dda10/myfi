"use client";

import { use, useState, useEffect } from "react";
import { StockDetailHeader } from "@/components/dashboard/StockDetailHeader";
import { MarketChart } from "@/components/dashboard/MarketChart";
import { fetchOHLCV, type OHLCVBar, type TimeInterval } from "@/lib/chart-engine";
import { apiFetch } from "@/lib/api";
import { InvestmentThesis } from "@/features/stock/components/InvestmentThesis";
import { AIValuation } from "@/features/stock/components/AIValuation";
import { TechnicalSummary } from "@/features/stock/components/TechnicalSummary";
import { FundamentalDashboard } from "@/features/stock/components/FundamentalDashboard";
import { NewsSentiment } from "@/features/stock/components/NewsSentiment";
import { FinancialStatements } from "@/features/stock/components/FinancialStatements";
import { SmartMoneyFlow } from "@/features/stock/components/SmartMoneyFlow";
import { OrderBook } from "@/features/stock/components/OrderBook";
import { Shareholders } from "@/features/stock/components/Shareholders";
import { Subsidiaries } from "@/features/stock/components/Subsidiaries";

interface StockQuote {
  symbol: string;
  name: string;
  price: number;
  change: number;
  changePercent: number;
  targetPrice?: number;
  liquidityTier?: number;
  tradabilityScore?: number;
}

export default function StockDetailPage({ params }: { params: Promise<{ symbol: string }> }) {
  const { symbol } = use(params);
  const [activeView, setActiveView] = useState("Overview");
  const [quote, setQuote] = useState<StockQuote | null>(null);
  const [chartData, setChartData] = useState<OHLCVBar[]>([]);
  const [interval, setInterval] = useState<TimeInterval>("1d");

  useEffect(() => {
    apiFetch<StockQuote>(`/api/market/quote?symbols=${symbol}`).then((q) => {
      if (q) setQuote(q);
    });
  }, [symbol]);

  useEffect(() => {
    fetchOHLCV(symbol, interval).then(setChartData).catch(() => setChartData([]));
  }, [symbol, interval]);

  return (
    <div className="space-y-6">
      <StockDetailHeader
        symbol={symbol}
        name={quote?.name ?? symbol}
        price={quote?.price ?? 0}
        change={quote?.change ?? 0}
        changePercent={quote?.changePercent ?? 0}
        activeView={activeView}
        onViewChange={setActiveView}
        liquidityTier={quote?.liquidityTier}
        tradabilityScore={quote?.tradabilityScore}
      />

      {/* Chart — always visible */}
      <div className="h-[480px]">
        <MarketChart data={chartData} />
      </div>

      {/* Tab content sections */}
      {activeView === "Overview" && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <InvestmentThesis symbol={symbol} />
            <AIValuation
              symbol={symbol}
              targetPrice={quote?.targetPrice}
              currentPrice={quote?.price}
            />
          </div>
          <TechnicalSummary symbol={symbol} />
          <OrderBook symbol={symbol} />
          <FundamentalDashboard symbol={symbol} />
        </div>
      )}

      {activeView === "Financials" && (
        <FinancialStatements symbol={symbol} />
      )}

      {activeView === "News" && (
        <NewsSentiment symbol={symbol} />
      )}

      {activeView === "Technicals" && (
        <div className="space-y-6">
          <TechnicalSummary symbol={symbol} />
        </div>
      )}

      {activeView === "AI Thesis" && (
        <div className="space-y-6">
          <InvestmentThesis symbol={symbol} />
          <AIValuation
            symbol={symbol}
            targetPrice={quote?.targetPrice}
            currentPrice={quote?.price}
          />
        </div>
      )}

      {activeView === "Smart Money" && (
        <SmartMoneyFlow symbol={symbol} />
      )}

      {activeView === "Company" && (
        <div className="space-y-6">
          <Shareholders symbol={symbol} />
          <Subsidiaries symbol={symbol} />
        </div>
      )}
    </div>
  );
}
