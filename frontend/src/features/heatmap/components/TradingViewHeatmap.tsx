"use client";

import { useEffect, useRef, useMemo, memo } from "react";
import { useTheme } from "@/context/ThemeContext";
import { useI18n } from "@/context/I18nContext";

interface TradingViewHeatmapProps {
  dataSource?: string;
}

function TradingViewHeatmapInner({ dataSource = "SPX500" }: TradingViewHeatmapProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { theme } = useTheme();
  const { locale } = useI18n();

  const config = useMemo(
    () => ({
      dataSource,
      blockSize: "market_cap_basic",
      blockColor: "change",
      grouping: "sector",
      locale: locale === "vi-VN" ? "vi_VN" : "en",
      colorTheme: theme === "dark" ? "dark" : "light",
      exchanges: [],
      hasTopBar: false,
      isDataSetEnabled: false,
      isZoomEnabled: true,
      hasSymbolTooltip: true,
      isMonoSize: false,
      width: "100%",
      height: "100%",
    }),
    [dataSource, theme, locale],
  );

  useEffect(() => {
    if (!containerRef.current) return;

    // Clear previous widget
    containerRef.current.innerHTML = "";

    const widgetContainer = document.createElement("div");
    widgetContainer.className = "tradingview-widget-container__widget";
    widgetContainer.style.height = "calc(100% - 32px)";
    widgetContainer.style.width = "100%";
    containerRef.current.appendChild(widgetContainer);

    const script = document.createElement("script");
    script.src = "https://s3.tradingview.com/external-embedding/embed-widget-stock-heatmap.js";
    script.async = true;
    script.innerHTML = JSON.stringify(config);
    containerRef.current.appendChild(script);

    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, [config]);

  return (
    <div
      ref={containerRef}
      className="tradingview-widget-container"
      style={{ height: "100%", width: "100%" }}
    />
  );
}

export const TradingViewHeatmap = memo(TradingViewHeatmapInner);
