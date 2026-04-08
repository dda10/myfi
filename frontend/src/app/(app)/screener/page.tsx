"use client";

import { useState, useCallback, useEffect } from "react";
import { SlidersHorizontal } from "lucide-react";
import { useI18n } from "@/context/I18nContext";
import { apiFetch } from "@/lib/api";
import { FilterPanel, EMPTY_FILTERS, type ScreenerFilters } from "@/features/screener/components/FilterPanel";
import { ResultsTable, type ScreenerRow } from "@/features/screener/components/ResultsTable";
import { PresetSelector } from "@/features/screener/components/PresetSelector";

interface ScreenerResponse {
  data: ScreenerRow[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

function cleanFilters(f: ScreenerFilters): Record<string, unknown> {
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(f)) {
    if (v === null || v === undefined || v === "") continue;
    if (Array.isArray(v) && v.length === 0) continue;
    out[k] = v;
  }
  return out;
}

export default function ScreenerPage() {
  const { t } = useI18n();

  const [filters, setFilters] = useState<ScreenerFilters>({ ...EMPTY_FILTERS });
  const [sortBy, setSortBy] = useState("marketCap");
  const [sortOrder, setSortOrder] = useState("desc");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);

  const [rows, setRows] = useState<ScreenerRow[]>([]);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(false);
  const [showFilters, setShowFilters] = useState(true);
  const [activePreset, setActivePreset] = useState<string | null>(null);

  const fetchResults = useCallback(
    async (f: ScreenerFilters, sb: string, so: string, p: number, ps: number) => {
      setLoading(true);
      const body = { ...cleanFilters(f), sortBy: sb, sortOrder: so, page: p, pageSize: ps };
      const resp = await apiFetch<ScreenerResponse>("/api/screener", {
        method: "POST",
        body: JSON.stringify(body),
      });
      if (resp) {
        setRows(resp.data ?? []);
        setTotal(resp.total);
        setTotalPages(resp.totalPages);
      } else {
        setRows([]);
        setTotal(0);
        setTotalPages(0);
      }
      setLoading(false);
    },
    [],
  );

  useEffect(() => {
    fetchResults(filters, sortBy, sortOrder, page, pageSize);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleApply = useCallback(() => {
    setPage(1);
    setActivePreset(null);
    fetchResults(filters, sortBy, sortOrder, 1, pageSize);
  }, [filters, sortBy, sortOrder, pageSize, fetchResults]);

  const handleReset = useCallback(() => {
    const next = { ...EMPTY_FILTERS };
    setFilters(next);
    setPage(1);
    setActivePreset(null);
    fetchResults(next, "marketCap", "desc", 1, pageSize);
  }, [pageSize, fetchResults]);

  const handleSort = useCallback(
    (field: string) => {
      const newOrder = sortBy === field && sortOrder === "desc" ? "asc" : "desc";
      setSortBy(field);
      setSortOrder(newOrder);
      setPage(1);
      fetchResults(filters, field, newOrder, 1, pageSize);
    },
    [filters, sortBy, sortOrder, pageSize, fetchResults],
  );

  const handlePageChange = useCallback(
    (p: number) => {
      if (p < 1 || p > totalPages) return;
      setPage(p);
      fetchResults(filters, sortBy, sortOrder, p, pageSize);
    },
    [filters, sortBy, sortOrder, pageSize, totalPages, fetchResults],
  );

  const handlePageSizeChange = useCallback(
    (size: number) => {
      setPageSize(size);
      setPage(1);
      fetchResults(filters, sortBy, sortOrder, 1, size);
    },
    [filters, sortBy, sortOrder, fetchResults],
  );

  const handleApplyPreset = useCallback(
    (name: string, presetFilters: Partial<ScreenerFilters>) => {
      const next: ScreenerFilters = { ...EMPTY_FILTERS, ...presetFilters };
      setFilters(next);
      setPage(1);
      setActivePreset(name);
      fetchResults(next, sortBy, sortOrder, 1, pageSize);
    },
    [sortBy, sortOrder, pageSize, fetchResults],
  );

  return (
    <div className="flex flex-col gap-4 w-full h-full text-zinc-200">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-white">{t("screener.title")}</h1>
        <button
          onClick={() => setShowFilters((v) => !v)}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-sm transition"
        >
          <SlidersHorizontal size={14} />
          {showFilters ? t("screener.filters") : t("screener.filters")}
        </button>
      </div>

      {/* Presets */}
      <PresetSelector
        activePreset={activePreset}
        currentFilters={filters}
        onApplyPreset={handleApplyPreset}
      />

      {/* Main content */}
      <div className="flex flex-col lg:flex-row gap-4 flex-1 min-h-0">
        {showFilters && (
          <FilterPanel
            filters={filters}
            onChange={setFilters}
            onApply={handleApply}
            onReset={handleReset}
          />
        )}

        <ResultsTable
          rows={rows}
          total={total}
          page={page}
          pageSize={pageSize}
          totalPages={totalPages}
          sortBy={sortBy}
          sortOrder={sortOrder}
          loading={loading}
          onSort={handleSort}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
        />
      </div>
    </div>
  );
}
