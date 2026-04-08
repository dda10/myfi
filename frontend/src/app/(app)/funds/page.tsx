"use client";

import { useState, useEffect, useCallback } from "react";
import { Search, X, TrendingUp, PieChart, BarChart3, ArrowLeft } from "lucide-react";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, BarChart, Bar, Cell,
} from "recharts";
import { useTheme } from "@/context/ThemeContext";
import { apiFetch } from "@/lib/api";

// --- Types ---

interface FundRecord {
  fundCode: string;
  fundName: string;
  managementCompany: string;
  fundType: string;
  nav: number;
  inceptionDate: string;
}

interface FundHolding {
  stockSymbol: string;
  stockName: string;
  percentage: number;
  marketValue: number;
}

interface FundIndustryAlloc {
  industryName: string;
  percentage: number;
}

interface FundAssetAlloc {
  assetClass: string;
  percentage: number;
}

interface FundDetail {
  fundCode: string;
  holdings: FundHolding[];
  allocation: {
    industry: FundIndustryAlloc[];
    asset: FundAssetAlloc[];
  };
}

interface FundNAV {
  date: string;
  navPerUnit: number;
  totalNav: number;
}

// --- Helpers ---

const COLORS = ["#6366f1", "#22c55e", "#f59e0b", "#ef4444", "#8b5cf6", "#06b6d4", "#ec4899", "#14b8a6", "#f97316", "#84cc16"];

function formatVND(v: number): string {
  if (v >= 1e9) return `${(v / 1e9).toFixed(1)} tỷ`;
  if (v >= 1e6) return `${(v / 1e6).toFixed(1)} tr`;
  return v.toLocaleString("vi-VN");
}

// --- Main Component ---

export default function FundsPage() {
  const { chartColors } = useTheme();
  const [funds, setFunds] = useState<FundRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState("");
  const [selectedFund, setSelectedFund] = useState<string | null>(null);
  const [detail, setDetail] = useState<FundDetail | null>(null);
  const [navData, setNavData] = useState<FundNAV[]>([]);
  const [detailLoading, setDetailLoading] = useState(false);

  // Fetch fund listing
  const fetchFunds = useCallback(async (q: string) => {
    setLoading(true);
    const path = q ? `/api/funds?q=${encodeURIComponent(q)}` : "/api/funds";
    const res = await apiFetch<{ data: FundRecord[] }>(path);
    if (res?.data) setFunds(res.data);
    setLoading(false);
  }, []);

  useEffect(() => { fetchFunds(""); }, [fetchFunds]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => fetchFunds(query), 400);
    return () => clearTimeout(timer);
  }, [query, fetchFunds]);

  // Fetch fund detail + NAV
  const openFund = useCallback(async (code: string) => {
    setSelectedFund(code);
    setDetailLoading(true);
    const end = new Date().toISOString().slice(0, 10);
    const start = new Date(Date.now() - 365 * 86400000).toISOString().slice(0, 10);
    const [detailRes, navRes] = await Promise.all([
      apiFetch<{ data: FundDetail }>(`/api/funds/${code}`),
      apiFetch<{ data: FundNAV[] }>(`/api/funds/${code}/nav?start=${start}&end=${end}`),
    ]);
    if (detailRes?.data) setDetail(detailRes.data);
    if (navRes?.data) setNavData(navRes.data);
    setDetailLoading(false);
  }, []);

  const chartStyle = { fontSize: 11, fill: chartColors.text };

  // --- Detail View ---
  if (selectedFund) {
    const fund = funds.find(f => f.fundCode === selectedFund);
    return (
      <div className="space-y-6 animate-fade-in">
        <button
          onClick={() => { setSelectedFund(null); setDetail(null); setNavData([]); }}
          className="flex items-center gap-1.5 text-xs text-text-muted hover:text-foreground transition"
        >
          <ArrowLeft size={14} /> Quay lại danh sách
        </button>

        <div>
          <h1 className="text-xl font-bold text-foreground">{fund?.fundName || selectedFund}</h1>
          <p className="text-xs text-text-muted mt-0.5">{fund?.managementCompany} · {fund?.fundType}</p>
        </div>

        {detailLoading ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="bg-card-bg border border-border-theme rounded-xl p-6 h-72 animate-pulse">
                <div className="h-5 w-1/3 bg-surface rounded mb-4" />
                <div className="h-48 bg-surface rounded" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* NAV Chart */}
            <div className="bg-card-bg border border-border-theme rounded-xl p-6 lg:col-span-2">
              <div className="flex items-center gap-2 mb-4">
                <TrendingUp size={18} className="text-indigo-400" />
                <h3 className="font-semibold text-foreground">Biến động NAV (1 năm)</h3>
              </div>
              {navData.length > 0 ? (
                <ResponsiveContainer width="100%" height={280}>
                  <LineChart data={navData.map(n => ({ date: n.date.slice(0, 10), nav: n.navPerUnit }))}>
                    <CartesianGrid strokeDasharray="3 3" stroke={chartColors.grid} />
                    <XAxis dataKey="date" tick={chartStyle} interval="preserveStartEnd" />
                    <YAxis tick={chartStyle} domain={["auto", "auto"]} />
                    <Tooltip
                      contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.grid}`, color: chartColors.text, fontSize: 12 }}
                      labelFormatter={l => `Ngày: ${l}`}
                      formatter={(v: unknown) => [Number(v).toLocaleString("vi-VN", { maximumFractionDigits: 0 }), "NAV/CCQ"]}
                    />
                    <Line type="monotone" dataKey="nav" stroke="#6366f1" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-sm text-text-muted text-center py-12">Không có dữ liệu NAV</p>
              )}
            </div>

            {/* Top Holdings */}
            <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
              <div className="px-4 py-3 border-b border-border-theme flex items-center gap-2">
                <BarChart3 size={16} className="text-green-400" />
                <h3 className="text-sm font-bold text-foreground">Top cổ phiếu nắm giữ</h3>
              </div>
              {detail?.holdings && detail.holdings.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="w-full text-xs">
                    <thead className="bg-surface/50 border-b border-border-theme">
                      <tr>
                        {["Mã CP", "Tên", "Tỷ trọng", "Giá trị"].map(h => (
                          <th key={h} className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap">{h}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border-theme">
                      {detail.holdings.slice(0, 10).map(h => (
                        <tr key={h.stockSymbol} className="hover:bg-surface transition">
                          <td className="px-4 py-2 font-bold text-foreground">{h.stockSymbol}</td>
                          <td className="px-4 py-2 text-text-muted truncate max-w-[140px]">{h.stockName}</td>
                          <td className="px-4 py-2 tabular-nums text-foreground">{h.percentage.toFixed(1)}%</td>
                          <td className="px-4 py-2 tabular-nums text-text-muted">{formatVND(h.marketValue)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <p className="text-sm text-text-muted text-center py-8">Không có dữ liệu</p>
              )}
            </div>

            {/* Industry Allocation */}
            <div className="bg-card-bg border border-border-theme rounded-xl p-6">
              <div className="flex items-center gap-2 mb-4">
                <PieChart size={16} className="text-purple-400" />
                <h3 className="text-sm font-bold text-foreground">Phân bổ ngành</h3>
              </div>
              {detail?.allocation?.industry && detail.allocation.industry.length > 0 ? (
                <ResponsiveContainer width="100%" height={220}>
                  <BarChart data={detail.allocation.industry.slice(0, 8)} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" stroke={chartColors.grid} />
                    <XAxis type="number" tick={chartStyle} tickFormatter={v => `${v}%`} />
                    <YAxis type="category" dataKey="industryName" tick={chartStyle} width={100} />
                    <Tooltip
                      contentStyle={{ backgroundColor: chartColors.background, border: `1px solid ${chartColors.grid}`, color: chartColors.text, fontSize: 12 }}
                      formatter={(v: unknown) => [`${Number(v).toFixed(1)}%`, "Tỷ trọng"]}
                    />
                    <Bar dataKey="percentage" radius={[0, 4, 4, 0]}>
                      {detail.allocation.industry.slice(0, 8).map((_, i) => (
                        <Cell key={i} fill={COLORS[i % COLORS.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              ) : (
                <p className="text-sm text-text-muted text-center py-12">Không có dữ liệu</p>
              )}
            </div>

            {/* Asset Allocation */}
            {detail?.allocation?.asset && detail.allocation.asset.length > 0 && (
              <div className="bg-card-bg border border-border-theme rounded-xl p-6 lg:col-span-2">
                <h3 className="text-sm font-bold text-foreground mb-4">Phân bổ tài sản</h3>
                <div className="flex flex-wrap gap-3">
                  {detail.allocation.asset.map((a, i) => (
                    <div key={a.assetClass} className="flex items-center gap-2 bg-surface rounded-lg px-3 py-2">
                      <div className="w-3 h-3 rounded-full" style={{ backgroundColor: COLORS[i % COLORS.length] }} />
                      <span className="text-xs text-foreground font-medium">{a.assetClass}</span>
                      <span className="text-xs text-text-muted tabular-nums">{a.percentage.toFixed(1)}%</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    );
  }

  // --- List View ---
  return (
    <div className="space-y-6 animate-fade-in">
      <div>
        <h1 className="text-xl font-bold text-foreground">Quỹ đầu tư</h1>
        <p className="text-xs text-text-muted mt-0.5">Phân tích quỹ mở, quỹ ETF tại Việt Nam</p>
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
        <input
          type="text"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="Tìm quỹ theo tên hoặc mã..."
          className="w-full pl-9 pr-8 py-2 text-sm bg-surface border border-border-theme rounded-lg text-foreground placeholder:text-text-muted focus:outline-none focus:ring-1 focus:ring-indigo-500"
        />
        {query && (
          <button onClick={() => setQuery("")} className="absolute right-2.5 top-1/2 -translate-y-1/2 text-text-muted hover:text-foreground">
            <X size={14} />
          </button>
        )}
      </div>

      {/* Fund Table */}
      {loading ? (
        <div className="bg-card-bg border border-border-theme rounded-xl p-6 animate-pulse">
          <div className="space-y-3">
            {[1, 2, 3, 4, 5].map(i => (
              <div key={i} className="h-10 bg-surface rounded" />
            ))}
          </div>
        </div>
      ) : funds.length === 0 ? (
        <div className="bg-card-bg border border-border-theme rounded-xl p-12 text-center">
          <p className="text-sm text-text-muted">Không tìm thấy quỹ nào</p>
        </div>
      ) : (
        <div className="bg-card-bg border border-border-theme rounded-xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-xs">
              <thead className="bg-surface/50 border-b border-border-theme">
                <tr>
                  {["Mã quỹ", "Tên quỹ", "Công ty quản lý", "Loại quỹ", "NAV/CCQ"].map(h => (
                    <th key={h} className="px-4 py-2.5 text-left text-[10px] font-semibold text-text-muted uppercase tracking-wider whitespace-nowrap">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-border-theme">
                {funds.map(f => (
                  <tr
                    key={f.fundCode}
                    onClick={() => openFund(f.fundCode)}
                    className="hover:bg-surface transition cursor-pointer"
                  >
                    <td className="px-4 py-2.5 font-bold text-indigo-400">{f.fundCode}</td>
                    <td className="px-4 py-2.5 font-medium text-foreground max-w-[240px] truncate">{f.fundName}</td>
                    <td className="px-4 py-2.5 text-text-muted truncate max-w-[180px]">{f.managementCompany}</td>
                    <td className="px-4 py-2.5">
                      <span className="text-[10px] font-semibold px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-400">{f.fundType}</span>
                    </td>
                    <td className="px-4 py-2.5 tabular-nums font-bold text-foreground">
                      {f.nav ? f.nav.toLocaleString("vi-VN", { maximumFractionDigits: 0 }) : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
