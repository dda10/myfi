"use client";

import { useI18n } from "@/context/I18nContext";
import { TradingViewHeatmap } from "@/features/heatmap/components/TradingViewHeatmap";

export default function HeatmapPage() {
  const { t } = useI18n();

  return (
    <div className="flex flex-col gap-4 h-full">
      <h1 className="text-xl font-bold text-white">{t("heatmap.title")}</h1>
      <div className="flex-1 min-h-[600px] bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden">
        <TradingViewHeatmap />
      </div>
    </div>
  );
}
