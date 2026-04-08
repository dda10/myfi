"use client";

import { useState, useEffect } from "react";
import { Newspaper } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface NewsItem {
  title: string;
  source: string;
  date: string;
  sentiment: "positive" | "negative" | "neutral";
  url?: string;
}

const SENTIMENT_BADGE: Record<string, string> = {
  positive: "bg-green-500/20 text-green-400",
  negative: "bg-red-500/20 text-red-400",
  neutral: "bg-zinc-500/20 text-zinc-400",
};

export function NewsSentiment({ symbol }: { symbol: string }) {
  const { t, formatDate } = useI18n();
  const [news, setNews] = useState<NewsItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    apiFetch<NewsItem[]>(`/api/market/news/${symbol}`)
      .then((res) => { if (res) setNews(res); })
      .finally(() => setLoading(false));
  }, [symbol]);

  if (loading) {
    return (
      <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse space-y-3">
        <div className="h-5 w-1/4 bg-surface rounded" />
        {[1, 2, 3].map((i) => (
          <div key={i} className="h-12 bg-surface rounded" />
        ))}
      </div>
    );
  }

  return (
    <section className="bg-card-bg border border-border-theme rounded-xl p-6">
      <div className="flex items-center gap-2 mb-4">
        <Newspaper size={20} className="text-indigo-400" />
        <h3 className="text-lg font-semibold text-foreground">{t("stock.news")}</h3>
      </div>

      {news.length === 0 ? (
        <p className="text-sm text-text-muted">{t("common.no_data")}</p>
      ) : (
        <div className="space-y-3">
          {news.map((item, idx) => (
            <div key={idx} className="flex items-start gap-3 bg-surface rounded-lg p-3">
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-foreground truncate">
                  {item.url ? (
                    <a href={item.url} target="_blank" rel="noopener noreferrer" className="hover:underline">
                      {item.title}
                    </a>
                  ) : item.title}
                </p>
                <p className="text-xs text-text-muted mt-0.5">
                  {item.source} • {formatDate(item.date)}
                </p>
              </div>
              <span className={`text-xs font-medium px-2 py-0.5 rounded-full whitespace-nowrap ${SENTIMENT_BADGE[item.sentiment] ?? SENTIMENT_BADGE.neutral}`}>
                {item.sentiment}
              </span>
            </div>
          ))}
        </div>
      )}
    </section>
  );
}
