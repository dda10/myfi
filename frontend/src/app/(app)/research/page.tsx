"use client";

import { useState, useEffect } from "react";
import { Search, Newspaper, BookOpen, TrendingUp, BarChart2, Download, Tag, ArrowRight } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";

interface ResearchReport {
  id: string;
  title: string;
  type: "factor_snapshot" | "sector_deepdive" | "market_outlook" | "strategy";
  date: string;
  summary?: string;
  pdfUrl?: string;
  imageUrl?: string;
  author?: string;
}

const TABS = [
  { key: "all",      label: "Khám phá",   icon: BookOpen },
  { key: "news",     label: "Bản tin",    icon: Newspaper },
  { key: "strategy", label: "Chiến lược", icon: TrendingUp },
  { key: "analysis", label: "Phân tích",  icon: BarChart2 },
];

const TYPE_STYLE: Record<string, { label: string; color: string }> = {
  factor_snapshot: { label: "Factor Strategy",  color: "bg-blue-500/20 text-blue-400 border-blue-500/20" },
  sector_deepdive: { label: "Sector Deepdive",  color: "bg-purple-500/20 text-purple-400 border-purple-500/20" },
  market_outlook:  { label: "Daily Take",       color: "bg-amber-500/20 text-amber-400 border-amber-500/20" },
  strategy:        { label: "Chiến lược",       color: "bg-emerald-500/20 text-emerald-400 border-emerald-500/20" },
};

// Mock featured article
const FEATURED: ResearchReport = {
  id: "0",
  title: "Dòng tiền đang trở nên chọn lọc hơn",
  type: "factor_snapshot",
  date: "2026-04-03",
  summary: "Cấu trúc nội tại của thị trường lại cho thấy dấu hiệu phân hóa với mức tương quan trung bình giữa các cổ phiếu và chỉ số giảm về mức thấp nhất trong vòng 5 năm, trong khi độ phân hóa lợi suất tăng lên mức cao so với giai đoạn 2023-2024.",
  author: "Miquant Research Team",
};

export default function ResearchPage() {
  const { t, formatDate } = useI18n();
  const [reports, setReports] = useState<ResearchReport[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("all");
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    setLoading(true);
    apiFetch<ResearchReport[]>("/api/research/reports")
      .then(res => { if (res) setReports(res); })
      .finally(() => setLoading(false));
  }, []);

  const filtered = reports.filter(r => {
    const matchesTab = activeTab === "all" || r.type === activeTab || r.type === activeTab;
    const matchesSearch = !searchQuery || r.title.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesTab && matchesSearch;
  });

  return (
    <div className="animate-fade-in space-y-6">
      {/* Page header with tabs and search */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div className="flex items-center gap-1">
          {TABS.map(tab => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition ${
                activeTab === tab.key
                  ? "bg-surface-hover text-foreground"
                  : "text-text-muted hover:text-foreground hover:bg-surface"
              }`}
            >
              <tab.icon size={14} />
              {tab.label}
            </button>
          ))}
        </div>
        <div className="relative">
          <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="Tìm kiếm bài viết..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            className="bg-surface border border-border-theme rounded-lg pl-9 pr-4 py-2 text-sm text-foreground placeholder-text-muted focus:outline-none focus:border-accent/50 transition w-56"
          />
        </div>
      </div>

      {/* Two-column layout */}
      <div className="grid grid-cols-1 xl:grid-cols-[1fr_300px] gap-6">
        {/* Left: Featured + article grid */}
        <div className="space-y-6">
          {/* Featured hero */}
          <div className="relative rounded-2xl overflow-hidden border border-border-theme cursor-pointer group" style={{ minHeight: 280 }}>
            <div className="absolute inset-0 bg-gradient-to-br from-indigo-900/60 via-black/60 to-black/80 z-10" />
            {/* Background placeholder */}
            <div className="absolute inset-0 bg-gradient-to-br from-indigo-900/40 to-zinc-900" />
            {/* Content */}
            <div className="relative z-20 p-8 flex flex-col justify-end h-full" style={{ minHeight: 280 }}>
              <div className="flex items-center gap-2 mb-3">
                <span className={`text-xs font-medium px-2 py-0.5 rounded border ${TYPE_STYLE[FEATURED.type].color}`}>
                  <Tag size={10} className="inline mr-1" />
                  {TYPE_STYLE[FEATURED.type].label}
                </span>
              </div>
              <h2 className="text-2xl md:text-3xl font-black text-white mb-2 group-hover:text-indigo-200 transition">
                {FEATURED.title}
              </h2>
              <p className="text-sm text-zinc-300 line-clamp-2 mb-4">{FEATURED.summary}</p>
              <div className="flex items-center gap-3">
                <span className="text-xs text-zinc-400">{FEATURED.author} · {formatDate(FEATURED.date)}</span>
                <button className="flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300 transition">
                  Đọc thêm <ArrowRight size={12} />
                </button>
              </div>
            </div>
          </div>

          {/* "Mới cập nhật" heading */}
          <div>
            <h3 className="text-sm font-bold text-foreground mb-3">Mới cập nhật</h3>
            {loading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {[1,2,3,4].map(i => (
                  <div key={i} className="bg-card-bg border border-border-theme rounded-xl p-4 h-32 animate-pulse">
                    <div className="h-3 w-2/3 bg-surface rounded mb-2" />
                    <div className="h-2.5 w-full bg-surface rounded" />
                  </div>
                ))}
              </div>
            ) : filtered.length === 0 ? (
              <div className="p-8 rounded-xl bg-surface border border-border-theme text-center text-text-muted text-sm">
                Không có bài viết nào
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {filtered.map(report => {
                  const typeInfo = TYPE_STYLE[report.type] ?? TYPE_STYLE.factor_snapshot;
                  return (
                    <div
                      key={report.id}
                      className="bg-card-bg border border-border-theme rounded-xl p-4 hover:border-indigo-500/30 transition cursor-pointer group"
                    >
                      {/* Mini gradient header */}
                      <div className="w-full h-20 rounded-lg mb-3 bg-gradient-to-br from-indigo-900/30 to-zinc-800 flex items-center justify-center">
                        <span className="text-2xl font-black text-indigo-400/30">miquant</span>
                      </div>
                      <div className="flex items-center gap-2 mb-1.5">
                        <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded border ${typeInfo.color}`}>
                          {typeInfo.label}
                        </span>
                        <span className="text-[10px] text-text-muted">{formatDate(report.date)}</span>
                      </div>
                      <h4 className="text-xs font-semibold text-foreground group-hover:text-indigo-300 transition line-clamp-2">{report.title}</h4>
                      {report.summary && (
                        <p className="text-[11px] text-text-muted mt-1 line-clamp-2">{report.summary}</p>
                      )}
                      {report.pdfUrl && (
                        <a href={report.pdfUrl} target="_blank" rel="noopener noreferrer"
                          className="mt-2 inline-flex items-center gap-1 text-[11px] text-indigo-400 hover:text-indigo-300">
                          <Download size={11} /> PDF
                        </a>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Right sidebar: recent/related */}
        <div className="space-y-4">
          <h3 className="text-sm font-bold text-foreground">Bài viết mới nhất</h3>
          {loading ? (
            <div className="space-y-3">
              {[1,2,3].map(i => <div key={i} className="h-24 bg-surface rounded-xl animate-pulse" />)}
            </div>
          ) : (
            <div className="space-y-3">
              {(filtered.length > 0 ? filtered : reports).slice(0, 6).map(report => {
                const typeInfo = TYPE_STYLE[report.type] ?? TYPE_STYLE.factor_snapshot;
                return (
                  <div key={report.id} className="bg-card-bg border border-border-theme rounded-xl overflow-hidden hover:border-indigo-500/30 transition cursor-pointer group">
                    <div className="h-16 bg-gradient-to-br from-indigo-900/20 to-zinc-900 flex items-center justify-center">
                      <span className="text-lg font-black text-indigo-800/40">miquant</span>
                    </div>
                    <div className="p-3">
                      <div className="flex items-center gap-1.5 mb-1">
                        <span className={`text-[9px] font-semibold px-1.5 py-0.5 rounded border ${typeInfo.color}`}>{typeInfo.label}</span>
                        <span className="text-[9px] text-text-muted">{formatDate(report.date)}</span>
                      </div>
                      <p className="text-[11px] font-semibold text-foreground group-hover:text-indigo-300 transition line-clamp-2">{report.title}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
