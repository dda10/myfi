"use client";

import { MarketOverview } from "@/features/dashboard/components/MarketOverview";
import { HotTopics } from "@/features/dashboard/components/HotTopics";
import { SectorPerformance } from "@/features/dashboard/components/SectorPerformance";
import { MACrossoverDist } from "@/features/dashboard/components/MACrossoverDist";
import { AIValuationRanking } from "@/features/dashboard/components/AIValuationRanking";
import { GlobalMarkets } from "@/features/dashboard/components/GlobalMarkets";
import { InterbankBondYields } from "@/features/dashboard/components/InterbankBondYields";
import { GoldPrice } from "@/features/dashboard/components/GoldPrice";
import { VNIndexChart } from "@/features/dashboard/components/VNIndexChart";

export default function DashboardPage() {
  return (
    <div className="animate-fade-in">
      {/*
        MiQuant 3-column layout:
        LEFT   | CENTER      | RIGHT
        240px  | flex-1      | 280px
      */}
      <div className="grid grid-cols-1 xl:grid-cols-[240px_1fr_280px] gap-4 items-start">

        {/* ── LEFT COLUMN ── */}
        <div className="flex flex-col gap-4 xl:min-h-0">
          {/* MA Crossover */}
          <MACrossoverDist />
          {/* Sector Performance */}
          <SectorPerformance />
        </div>

        {/* ── CENTER COLUMN ── */}
        <div className="flex flex-col gap-4">
          {/* VN-Index TradingView chart */}
          <div className="bg-card-bg border border-border-theme rounded-xl p-3" style={{ minHeight: 440 }}>
            <VNIndexChart />
          </div>

          {/* Hot Topics below chart */}
          <HotTopics />

          {/* Interbank Rates & Bond Yield */}
          <InterbankBondYields />
        </div>

        {/* ── RIGHT COLUMN ── */}
        <div className="flex flex-col gap-4">
          {/* AI Valuation Ranking */}
          <AIValuationRanking />
          {/* Global Markets */}
          <GlobalMarkets />
          {/* Gold Price */}
          <GoldPrice />
        </div>
      </div>

      {/* Market overview strip (VN-Index, HNX, UPCOM cards) */}
      <div className="mt-4">
        <MarketOverview />
      </div>
    </div>
  );
}
