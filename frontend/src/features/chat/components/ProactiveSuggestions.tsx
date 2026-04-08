"use client";

import { useState, useEffect } from "react";
import { TrendingUp, Flame, HelpCircle } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface HotTopic {
  title: string;
  symbol?: string;
}

interface ProactiveSuggestionsProps {
  onSelect: (text: string) => void;
}

const STATIC_SUGGESTIONS_VI = [
  "Phân tích kỹ thuật VNM",
  "Top cổ phiếu tăng trưởng hôm nay",
  "So sánh FPT và VNM",
  "Xu hướng ngành ngân hàng",
];

const STATIC_SUGGESTIONS_EN = [
  "Technical analysis for VNM",
  "Top growth stocks today",
  "Compare FPT and VNM",
  "Banking sector trends",
];

export function ProactiveSuggestions({ onSelect }: ProactiveSuggestionsProps) {
  const { t, locale } = useI18n();
  const [hotTopics, setHotTopics] = useState<HotTopic[]>([]);

  useEffect(() => {
    apiFetch<HotTopic[]>("/api/market/hot-topics").then((res) => {
      if (res) setHotTopics(res.slice(0, 4));
    });
  }, []);

  const suggestions = locale === "vi-VN" ? STATIC_SUGGESTIONS_VI : STATIC_SUGGESTIONS_EN;

  return (
    <div className="space-y-3 px-1">
      {/* Hot Topics */}
      {hotTopics.length > 0 && (
        <div>
          <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium mb-1.5">
            <Flame size={12} className="text-orange-400" />
            {t("dashboard.hot_topics")}
          </div>
          <div className="flex flex-wrap gap-1.5">
            {hotTopics.map((topic, i) => (
              <button
                key={i}
                onClick={() => onSelect(topic.symbol ? `${t("stock.analysis")} ${topic.symbol}` : topic.title)}
                className="text-xs bg-orange-500/10 text-orange-300 hover:bg-orange-500/20 px-2.5 py-1 rounded-full transition"
              >
                {topic.symbol && <TrendingUp size={10} className="inline mr-1" />}
                {topic.title}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Suggested Questions */}
      <div>
        <div className="flex items-center gap-1.5 text-xs text-zinc-400 font-medium mb-1.5">
          <HelpCircle size={12} className="text-indigo-400" />
          {t("chat.suggestions")}
        </div>
        <div className="flex flex-col gap-1">
          {suggestions.map((q, i) => (
            <button
              key={i}
              onClick={() => onSelect(q)}
              className="text-xs text-left text-zinc-300 hover:text-white bg-zinc-800/60 hover:bg-zinc-700/60 px-3 py-1.5 rounded-lg transition"
            >
              {q}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
