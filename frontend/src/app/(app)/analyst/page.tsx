"use client";

import { useState, useEffect, useCallback } from "react";
import {
  TrendingUp, TrendingDown, Minus, ArrowUpRight, Calendar, Search, SlidersHorizontal,
} from "lucide-react";
import { apiFetch } from "@/lib/api";
import { useApp } from "@/context/AppContext";
import { useRouter } from "next/navigation";

interface AnalystReport {
  id: string;
  symbol: string;
  analystName: string;
  firm: string;
  prevTargetPrice?: number;
  newTargetPrice?: number;
  prevRating?: string;
  newRating?: string;
  ratingChange?: "upgrade" | "downgrade" | "maintain";
  date: string;
  title?: string;
}

// Rich mock data in MiQuant Analyst IQ style
const MOCK_REPORTS: AnalystReport[] = [
  { id: "1",  symbol: "FPT",  analystName: "Minh Hoang",    firm: "SSI Research",    prevTargetPrice: 108000, newTargetPrice: 142000, prevRating: "Mua",         newRating: "Mua mạnh",    ratingChange: "upgrade",   date: "2026-04-03" },
  { id: "2",  symbol: "VCB",  analystName: "Thu Ha",        firm: "Vndirect",        prevTargetPrice: 85000,  newTargetPrice: 95000,  prevRating: "Mua",         newRating: "Mua",         ratingChange: "maintain",  date: "2026-04-02" },
  { id: "3",  symbol: "MWG",  analystName: "Duc Anh",       firm: "VCSC",            prevTargetPrice: 60000,  newTargetPrice: 42000,  prevRating: "Mua",         newRating: "Phù hợp TT", ratingChange: "downgrade", date: "2026-04-02" },
  { id: "4",  symbol: "HPG",  analystName: "Lan Phuong",    firm: "ACBS",            prevTargetPrice: 32000,  newTargetPrice: 38000,  prevRating: "Phù hợp TT", newRating: "Mua",         ratingChange: "upgrade",   date: "2026-04-01" },
  { id: "5",  symbol: "VHM",  analystName: "Thanh Trung",   firm: "Maybank IB",      prevTargetPrice: 45000,  newTargetPrice: 52000,  prevRating: "Mua",         newRating: "Mua",         ratingChange: "maintain",  date: "2026-04-01" },
  { id: "6",  symbol: "GAS",  analystName: "Bao Nguyen",    firm: "BSC Research",    prevTargetPrice: 118000, newTargetPrice: 131000, prevRating: "Phù hợp TT", newRating: "Mua",         ratingChange: "upgrade",   date: "2026-03-31" },
  { id: "7",  symbol: "NVL",  analystName: "Khanh Van",     firm: "MB Securities",   prevTargetPrice: 12000,  newTargetPrice: 9500,   prevRating: "Mua",         newRating: "Bán",         ratingChange: "downgrade", date: "2026-03-30" },
  { id: "8",  symbol: "DBC",  analystName: "Quoc Hung",     firm: "KBS Securities",  prevTargetPrice: 38000,  newTargetPrice: 42500,  prevRating: "Phù hợp TT", newRating: "Mua",         ratingChange: "upgrade",   date: "2026-03-29" },
  { id: "9",  symbol: "ACB",  analystName: "Phuong Linh",   firm: "Rong Viet Sec.",  prevTargetPrice: 31000,  newTargetPrice: 31000,  prevRating: "Mua",         newRating: "Mua",         ratingChange: "maintain",  date: "2026-03-28" },
  { id: "10", symbol: "TCB",  analystName: "Hai Tran",      firm: "VPBank Sec.",     prevTargetPrice: 24000,  newTargetPrice: 28000,  prevRating: "Phù hợp TT", newRating: "Mua",         ratingChange: "upgrade",   date: "2026-03-27" },
];

const RATING_COLORS: Record<string, string> = {
  "Mua mạnh": "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
  "Mua":       "text-green-400 bg-green-500/10 border-green-500/20",
  "Phù hợp TT":"text-yellow-400 bg-yellow-500/10 border-yellow-500/20",
  "Bán":       "text-red-400 bg-red-500/10 border-red-500/20",
  "Bán mạnh":  "text-rose-400 bg-rose-500/10 border-rose-500/20",
};

function RatingBadge({ rating }: { rating?: string }) {
  if (!rating) return null;
  const cls = RATING_COLORS[rating] ?? "text-text-muted bg-surface border-border-theme";
  return (
    <span className={`text-[10px] font-semibold px-2 py-0.5 rounded border ${cls}`}>
      {rating}
    </span>
  );
}

function ChangeIcon({ change }: { change?: "upgrade" | "downgrade" | "maintain" }) {
  if (change === "upgrade")   return <TrendingUp  size={14} className="text-green-400" />;
  if (change === "downgrade") return <TrendingDown size={14} className="text-red-400" />;
  return <Minus size={14} className="text-text-muted" />;
}

export default function AnalystIQPage() {
  const router = useRouter();
  const { setActiveSymbol } = useApp();
  const [reports, setReports] = useState<AnalystReport[]>(MOCK_REPORTS);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState("");
  const [filterChange, setFilterChange] = useState<"all" | "upgrade" | "downgrade" | "maintain">("all");
  const [sortField, setSortField] = useState<keyof AnalystReport>("date");
  const [sortDir, setSortDir] = useState<"asc" | "desc">("desc");

  useEffect(() => {
    setLoading(true);
    apiFetch<AnalystReport[]>("/api/analyst/reports")
      .then(res => { if (res && Array.isArray(res) && res.length > 0) setReports(res); })
      .finally(() => setLoading(false));
  }, []);

  const handleSort = useCallback((field: keyof AnalystReport) => {
    if (sortField === field) setSortDir(d => d === "asc" ? "desc" : "asc");
    else { setSortField(field); setSortDir("desc"); }
  }, [sortField]);

  const filtered = reports.filter(r => {
    const matchSearch = !search || r.symbol.toLowerCase().includes(search.toLowerCase()) || (r.firm ?? "").toLowerCase().includes(search.toLowerCase());
    const matchChange = filterChange === "all" || r.ratingChange === filterChange;
    return matchSearch && matchChange;
  }).sort((a, b) => {
    const av = a[sortField] as string ?? "";
    const bv = b[sortField] as string ?? "";
    return sortDir === "desc" ? bv.localeCompare(av) : av.localeCompare(bv);
  });

  const SortTh = ({ label, field }: { label: string; field: keyof AnalystReport }) => (
    <th
      className="px-3 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-text-muted cursor-pointer hover:text-foreground transition whitespace-nowrap"
      onClick={() => handleSort(field)}
    >
      <span className="flex items-center gap-1">
        {label}
        {sortField === field && (sortDir === "desc" ? " ↓" : " ↑")}
      </span>
    </th>
  );

  return (
    <div className="animate-fade-in space-y-5">
      {/* Page header */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-xl font-bold text-foreground">Analyst IQ</h1>
          <p className="text-xs text-text-muted mt-0.5">Theo dõi thay đổi khuyến nghị và giá mục tiêu từ các công ty chứng khoán</p>
        </div>
        <div className="flex items-center gap-2">
          {/* Change filter */}
          <div className="flex gap-1">
            {(["all","upgrade","downgrade","maintain"] as const).map(f => (
              <button
                key={f}
                onClick={() => setFilterChange(f)}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium transition ${
                  filterChange === f
                    ? f === "upgrade"   ? "bg-green-500/20 text-green-400 border border-green-500/30"
                    : f === "downgrade" ? "bg-red-500/20 text-red-400 border border-red-500/30"
                    : "bg-surface-hover text-foreground border border-border-theme"
                    : "text-text-muted hover:text-foreground hover:bg-surface"
                }`}
              >
                {f === "all" ? "Tất cả" : f === "upgrade" ? "▲ Nâng" : f === "downgrade" ? "▼ Hạ" : "= Giữ"}
              </button>
            ))}
          </div>
          {/* Search */}
          <div className="relative">
            <Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
            <input
              type="text"
              placeholder="Tìm mã CP, CTCK..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="bg-surface border border-border-theme rounded-lg pl-8 pr-3 py-1.5 text-xs text-foreground placeholder-text-muted focus:outline-none focus:border-accent/50 transition w-44"
            />
          </div>
        </div>
      </div>

      {/* Table */}
      <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead className="bg-surface/60 border-b border-border-theme">
              <tr>
                <SortTh label="Mã CP"      field="symbol" />
                <SortTh label="Phân tích"  field="analystName" />
                <SortTh label="CTCK"        field="firm" />
                <th className="px-3 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-text-muted whitespace-nowrap">Khuyến nghị cũ</th>
                <th className="px-3 py-2.5 text-left text-[10px] font-semibold uppercase tracking-wider text-text-muted whitespace-nowrap">Khuyến nghị mới</th>
                <th className="px-3 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-text-muted whitespace-nowrap">TP cũ</th>
                <th className="px-3 py-2.5 text-right text-[10px] font-semibold uppercase tracking-wider text-text-muted whitespace-nowrap">TP mới</th>
                <th className="px-3 py-2.5 text-center text-[10px] font-semibold uppercase tracking-wider text-text-muted whitespace-nowrap">Thay đổi</th>
                <SortTh label="Ngày"        field="date" />
              </tr>
            </thead>
            <tbody className="divide-y divide-border-theme">
              {loading ? (
                Array.from({ length: 8 }).map((_, i) => (
                  <tr key={i}>
                    {Array.from({ length: 9 }).map((_, j) => (
                      <td key={j} className="px-3 py-3">
                        <div className="h-4 bg-surface rounded animate-pulse" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : filtered.length === 0 ? (
                <tr>
                  <td colSpan={9} className="text-center py-12 text-text-muted">Không có dữ liệu</td>
                </tr>
              ) : (
                filtered.map(report => {
                  const tpChange = report.newTargetPrice && report.prevTargetPrice
                    ? ((report.newTargetPrice - report.prevTargetPrice) / report.prevTargetPrice) * 100
                    : null;
                  return (
                    <tr
                      key={report.id}
                      className="hover:bg-surface transition cursor-pointer"
                      onClick={() => {
                        setActiveSymbol(report.symbol);
                        router.push(`/stock/${report.symbol}`);
                      }}
                    >
                      <td className="px-3 py-3">
                        <span className="font-bold text-foreground">{report.symbol}</span>
                      </td>
                      <td className="px-3 py-3 text-text-muted">{report.analystName}</td>
                      <td className="px-3 py-3">
                        <span className="text-text-muted bg-surface px-2 py-0.5 rounded text-[10px] font-medium">{report.firm}</span>
                      </td>
                      <td className="px-3 py-3">
                        <RatingBadge rating={report.prevRating} />
                      </td>
                      <td className="px-3 py-3">
                        <RatingBadge rating={report.newRating} />
                      </td>
                      <td className="px-3 py-3 text-right text-text-muted tabular-nums">
                        {report.prevTargetPrice ? report.prevTargetPrice.toLocaleString() : "—"}
                      </td>
                      <td className="px-3 py-3 text-right font-semibold text-foreground tabular-nums">
                        {report.newTargetPrice ? report.newTargetPrice.toLocaleString() : "—"}
                        {tpChange !== null && (
                          <span className={`ml-1.5 text-[10px] font-bold ${tpChange >= 0 ? "text-green-400" : "text-red-400"}`}>
                            {tpChange >= 0 ? "+" : ""}{tpChange.toFixed(1)}%
                          </span>
                        )}
                      </td>
                      <td className="px-3 py-3">
                        <div className="flex justify-center">
                          <ChangeIcon change={report.ratingChange} />
                        </div>
                      </td>
                      <td className="px-3 py-3 text-text-muted tabular-nums">{report.date}</td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
        <div className="px-4 py-3 border-t border-border-theme text-xs text-text-muted">
          {filtered.length} báo cáo phân tích
        </div>
      </div>
    </div>
  );
}
