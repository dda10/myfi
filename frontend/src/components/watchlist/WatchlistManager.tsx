"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import {
  Star, Plus, Pencil, Trash2, Check, X, RefreshCw,
  ArrowUp, ArrowDown, Bell, TrendingUp, TrendingDown,
  ChevronRight, ChevronDown, Eye,
} from "lucide-react";
import { useApp } from "@/context/AppContext";
import { useI18n } from "@/context/I18nContext";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// --- Types ---

interface WatchlistSymbol {
  id: number;
  watchlistId: number;
  symbol: string;
  position: number;
  priceAlertAbove?: number | null;
  priceAlertBelow?: number | null;
}

interface WatchlistData {
  id: number;
  userId: number;
  name: string;
  symbols: WatchlistSymbol[];
}

interface QuoteData {
  symbol: string;
  price: string;
  change: string;
  isPositive: boolean;
  rawPrice: number;
}

// --- API helper ---

async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API}${path}`, {
    credentials: "include",
    ...opts,
    headers: { "Content-Type": "application/json", ...opts?.headers },
  });
  if (!res.ok) throw new Error(`API ${res.status}`);
  return res.json();
}

// --- Component ---

export function WatchlistManager() {
  const { setActiveSymbol, setActiveTab } = useApp();
  const { formatCurrency } = useI18n();

  // Data state
  const [watchlists, setWatchlists] = useState<WatchlistData[]>([]);
  const [quotes, setQuotes] = useState<Record<string, QuoteData>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Expanded watchlists (show symbols)
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set());

  // Create watchlist
  const [showCreate, setShowCreate] = useState(false);
  const [createName, setCreateName] = useState("");

  // Rename
  const [renamingId, setRenamingId] = useState<number | null>(null);
  const [renameValue, setRenameValue] = useState("");

  // Delete confirmation
  const [deletingId, setDeletingId] = useState<number | null>(null);

  // Add symbol per watchlist
  const [addingToId, setAddingToId] = useState<number | null>(null);
  const [addSymbolInput, setAddSymbolInput] = useState("");

  // Alert editor
  const [alertKey, setAlertKey] = useState<string | null>(null); // "wlId:symbol"
  const [alertAbove, setAlertAbove] = useState("");
  const [alertBelow, setAlertBelow] = useState("");

  // --- Data fetching ---

  const fetchWatchlists = useCallback(async () => {
    try {
      const data = await apiFetch<WatchlistData[]>("/api/watchlists");
      const lists = data ?? [];
      setWatchlists(lists);
      // Auto-expand all on first load
      if (lists.length > 0 && expandedIds.size === 0) {
        setExpandedIds(new Set(lists.map((w) => w.id)));
      }
      setError(null);
    } catch {
      setError("Failed to load watchlists");
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const fetchQuotes = useCallback(async () => {
    const allSymbols = Array.from(
      new Set(watchlists.flatMap((w) => w.symbols.map((s) => s.symbol)))
    );
    if (allSymbols.length === 0) return;
    try {
      const res = await fetch(
        `${API}/api/market/quote?symbols=${allSymbols.join(",")}`,
        { credentials: "include" }
      );
      if (!res.ok) return;
      const json = await res.json();
      if (json.data) {
        const map: Record<string, QuoteData> = {};
        json.data.forEach((q: any) => {
          const changeVal = q.change ?? 0;
          const pct = q.changePercent !== undefined ? q.changePercent.toFixed(2) : "0.00";
          const isPositive = changeVal >= 0;
          map[q.symbol] = {
            symbol: q.symbol,
            price: formatCurrency(q.close ?? 0),
            change: `${isPositive ? "+" : ""}${pct}%`,
            isPositive,
            rawPrice: q.close ?? 0,
          };
        });
        setQuotes(map);
      }
    } catch { /* ignore */ }
  }, [watchlists, formatCurrency]);

  useEffect(() => {
    setLoading(true);
    fetchWatchlists().finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    fetchQuotes();
    const id = setInterval(fetchQuotes, 15000);
    return () => clearInterval(id);
  }, [fetchQuotes]);

  // --- Toggle expand ---
  const toggleExpand = (id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  // --- Actions ---

  const createWatchlist = async () => {
    const name = createName.trim();
    if (!name) return;
    try {
      const wl = await apiFetch<WatchlistData>("/api/watchlists", {
        method: "POST",
        body: JSON.stringify({ name }),
      });
      setWatchlists((prev) => [...prev, wl]);
      setExpandedIds((prev) => new Set([...prev, wl.id]));
      setCreateName("");
      setShowCreate(false);
    } catch { /* ignore */ }
  };

  const renameWatchlist = async (id: number) => {
    const name = renameValue.trim();
    if (!name) return;
    try {
      await apiFetch(`/api/watchlists/${id}`, {
        method: "PUT",
        body: JSON.stringify({ name }),
      });
      setWatchlists((prev) => prev.map((w) => (w.id === id ? { ...w, name } : w)));
      setRenamingId(null);
      setRenameValue("");
    } catch { /* ignore */ }
  };

  const deleteWatchlist = async (id: number) => {
    try {
      await apiFetch(`/api/watchlists/${id}`, { method: "DELETE" });
      setWatchlists((prev) => prev.filter((w) => w.id !== id));
      setDeletingId(null);
    } catch { /* ignore */ }
  };

  const addSymbol = async (wlId: number) => {
    const sym = addSymbolInput.trim().toUpperCase();
    if (!sym) return;
    try {
      await apiFetch(`/api/watchlists/${wlId}/symbols`, {
        method: "POST",
        body: JSON.stringify({ symbol: sym }),
      });
      await fetchWatchlists();
      setAddSymbolInput("");
      setAddingToId(null);
    } catch { /* ignore */ }
  };

  const removeSymbol = async (wlId: number, symbol: string) => {
    try {
      await apiFetch(`/api/watchlists/${wlId}/symbols/${symbol}`, { method: "DELETE" });
      setWatchlists((prev) =>
        prev.map((w) =>
          w.id === wlId ? { ...w, symbols: w.symbols.filter((s) => s.symbol !== symbol) } : w
        )
      );
    } catch { /* ignore */ }
  };

  const moveSymbol = async (wl: WatchlistData, idx: number, dir: -1 | 1) => {
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= wl.symbols.length) return;
    const reordered = [...wl.symbols];
    [reordered[idx], reordered[newIdx]] = [reordered[newIdx], reordered[idx]];
    const newSymbols = reordered.map((s, i) => ({ ...s, position: i }));
    setWatchlists((prev) =>
      prev.map((w) => (w.id === wl.id ? { ...w, symbols: newSymbols } : w))
    );
    try {
      await apiFetch(`/api/watchlists/${wl.id}/reorder`, {
        method: "PUT",
        body: JSON.stringify({ symbols: newSymbols.map((s) => s.symbol) }),
      });
    } catch {
      await fetchWatchlists();
    }
  };

  const openAlertEditor = (wlId: number, ws: WatchlistSymbol) => {
    const key = `${wlId}:${ws.symbol}`;
    if (alertKey === key) {
      setAlertKey(null);
      return;
    }
    setAlertKey(key);
    setAlertAbove(ws.priceAlertAbove != null ? String(ws.priceAlertAbove) : "");
    setAlertBelow(ws.priceAlertBelow != null ? String(ws.priceAlertBelow) : "");
  };

  const saveAlert = async (wlId: number, symbol: string) => {
    const above = alertAbove ? parseFloat(alertAbove) : null;
    const below = alertBelow ? parseFloat(alertBelow) : null;
    try {
      await apiFetch(`/api/watchlists/${wlId}/symbols/${symbol}/alert`, {
        method: "PUT",
        body: JSON.stringify({
          alertAbove: above && !isNaN(above) ? above : null,
          alertBelow: below && !isNaN(below) ? below : null,
        }),
      });
      await fetchWatchlists();
      setAlertKey(null);
      setAlertAbove("");
      setAlertBelow("");
    } catch { /* ignore */ }
  };

  // --- Render ---

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Star size={22} className="text-yellow-400 fill-yellow-400" />
          <h1 className="text-2xl font-bold text-white">Watchlists</h1>
          <span className="text-sm text-zinc-500">{watchlists.length} lists</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowCreate(true)}
            className="flex items-center gap-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold rounded-lg transition"
          >
            <Plus size={15} /> New Watchlist
          </button>
          <button
            onClick={() => { fetchWatchlists(); fetchQuotes(); }}
            className="p-2 text-zinc-400 hover:text-white transition rounded-lg hover:bg-zinc-800"
            title="Refresh"
          >
            <RefreshCw size={16} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      {/* Create watchlist inline */}
      {showCreate && (
        <div className="bg-zinc-900 border border-zinc-800 rounded-xl p-4 flex items-center gap-3">
          <input
            value={createName}
            onChange={(e) => setCreateName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && createWatchlist()}
            placeholder="Watchlist name..."
            autoFocus
            className="flex-1 bg-black border border-zinc-700 rounded-lg px-4 py-2 text-sm text-white placeholder-zinc-600 outline-none focus:border-indigo-500"
          />
          <button onClick={createWatchlist} className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold rounded-lg transition">
            Create
          </button>
          <button onClick={() => { setShowCreate(false); setCreateName(""); }} className="p-2 text-zinc-500 hover:text-zinc-300 transition">
            <X size={16} />
          </button>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="px-4 py-3 text-sm text-red-400 bg-red-950/20 border border-red-900/30 rounded-xl">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading && watchlists.length === 0 && (
        <div className="flex flex-col items-center justify-center py-16 text-zinc-500">
          <RefreshCw size={24} className="animate-spin mb-3" />
          <p className="text-sm">Loading watchlists...</p>
        </div>
      )}

      {/* Empty state */}
      {!loading && watchlists.length === 0 && !error && (
        <div className="flex flex-col items-center justify-center py-16 text-zinc-500">
          <Star size={36} className="mb-3 text-zinc-700" />
          <p className="text-sm">No watchlists yet.</p>
          <p className="text-xs text-zinc-600 mt-1">Create one to start tracking stocks.</p>
        </div>
      )}

      {/* Watchlist cards */}
      {watchlists.map((wl) => {
        const isExpanded = expandedIds.has(wl.id);
        return (
          <div key={wl.id} className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg">
            {/* Watchlist header */}
            <div className="px-5 py-4 border-b border-zinc-800 flex items-center justify-between">
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <button onClick={() => toggleExpand(wl.id)} className="text-zinc-500 hover:text-white transition">
                  {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                </button>

                {renamingId === wl.id ? (
                  <div className="flex items-center gap-2 flex-1">
                    <input
                      value={renameValue}
                      onChange={(e) => setRenameValue(e.target.value)}
                      onKeyDown={(e) => e.key === "Enter" && renameWatchlist(wl.id)}
                      autoFocus
                      className="flex-1 bg-black border border-zinc-600 rounded-lg px-3 py-1 text-sm text-white outline-none focus:border-indigo-500"
                    />
                    <button onClick={() => renameWatchlist(wl.id)} className="p-1 text-green-400 hover:text-green-300"><Check size={15} /></button>
                    <button onClick={() => setRenamingId(null)} className="p-1 text-zinc-500 hover:text-zinc-300"><X size={15} /></button>
                  </div>
                ) : (
                  <>
                    <h2 className="text-base font-semibold text-white truncate">{wl.name}</h2>
                    <span className="text-xs bg-zinc-800 text-zinc-400 px-2 py-0.5 rounded-full font-mono">
                      {wl.symbols.length}
                    </span>
                  </>
                )}
              </div>

              {renamingId !== wl.id && (
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => { setAddingToId(addingToId === wl.id ? null : wl.id); setAddSymbolInput(""); }}
                    className="p-1.5 text-zinc-500 hover:text-indigo-400 transition rounded hover:bg-zinc-800"
                    title="Add symbol"
                  >
                    <Plus size={15} />
                  </button>
                  <button
                    onClick={() => { setRenamingId(wl.id); setRenameValue(wl.name); }}
                    className="p-1.5 text-zinc-500 hover:text-indigo-400 transition rounded hover:bg-zinc-800"
                    title="Rename"
                  >
                    <Pencil size={13} />
                  </button>
                  <button
                    onClick={() => setDeletingId(wl.id)}
                    className="p-1.5 text-zinc-500 hover:text-red-400 transition rounded hover:bg-zinc-800"
                    title="Delete"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              )}
            </div>

            {/* Delete confirmation */}
            {deletingId === wl.id && (
              <div className="px-5 py-3 border-b border-zinc-800 bg-red-950/30 flex items-center justify-between">
                <span className="text-sm text-red-300">Delete &quot;{wl.name}&quot;?</span>
                <div className="flex gap-2">
                  <button onClick={() => deleteWatchlist(wl.id)} className="px-3 py-1 bg-red-600 hover:bg-red-500 text-white text-xs font-semibold rounded transition">
                    Delete
                  </button>
                  <button onClick={() => setDeletingId(null)} className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs rounded transition">
                    Cancel
                  </button>
                </div>
              </div>
            )}

            {/* Add symbol input */}
            {addingToId === wl.id && (
              <div className="px-5 py-3 border-b border-zinc-800 flex gap-2">
                <input
                  type="text"
                  value={addSymbolInput}
                  onChange={(e) => setAddSymbolInput(e.target.value.toUpperCase())}
                  onKeyDown={(e) => e.key === "Enter" && addSymbol(wl.id)}
                  placeholder="Type symbol, e.g. VCB"
                  autoFocus
                  className="flex-1 bg-black border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white placeholder-zinc-600 outline-none focus:border-indigo-500"
                />
                <button onClick={() => addSymbol(wl.id)} className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold rounded-lg transition">
                  Add
                </button>
              </div>
            )}

            {/* Symbol list */}
            {isExpanded && (
              <div className="p-2">
                {wl.symbols.length === 0 ? (
                  <div className="flex flex-col items-center justify-center py-8 text-zinc-500">
                    <Eye size={22} className="mb-2 text-zinc-700" />
                    <p className="text-sm">No symbols in this watchlist.</p>
                  </div>
                ) : (
                  wl.symbols.map((ws, idx) => {
                    const q = quotes[ws.symbol];
                    const hasAlert = ws.priceAlertAbove != null || ws.priceAlertBelow != null;
                    const isAlertOpen = alertKey === `${wl.id}:${ws.symbol}`;

                    return (
                      <div key={ws.symbol}>
                        <div className="flex items-center justify-between p-3 rounded-lg transition group hover:bg-zinc-800/80">
                          {/* Reorder */}
                          <div className="flex flex-col mr-2 gap-0.5">
                            <button
                              onClick={() => moveSymbol(wl, idx, -1)}
                              disabled={idx === 0}
                              className="p-0.5 text-zinc-700 hover:text-zinc-400 disabled:opacity-20 transition"
                            >
                              <ArrowUp size={11} />
                            </button>
                            <button
                              onClick={() => moveSymbol(wl, idx, 1)}
                              disabled={idx === wl.symbols.length - 1}
                              className="p-0.5 text-zinc-700 hover:text-zinc-400 disabled:opacity-20 transition"
                            >
                              <ArrowDown size={11} />
                            </button>
                          </div>

                          {/* Symbol info */}
                          <div
                            className="flex items-center gap-3 flex-1 cursor-pointer"
                            onClick={() => { setActiveSymbol(ws.symbol); setActiveTab("Markets"); }}
                          >
                            <div className="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-600/40 to-zinc-800 flex items-center justify-center font-bold text-zinc-300 text-sm">
                              {ws.symbol[0]}
                            </div>
                            <div>
                              <div className="flex items-center gap-1.5">
                                <h4 className="font-bold text-white text-sm">{ws.symbol}</h4>
                                {hasAlert && (
                                  <span title={`Alert: ${ws.priceAlertAbove ? `above ${ws.priceAlertAbove}` : ""}${ws.priceAlertAbove && ws.priceAlertBelow ? ", " : ""}${ws.priceAlertBelow ? `below ${ws.priceAlertBelow}` : ""}`}>
                                    <Bell size={10} className="text-yellow-500" />
                                  </span>
                                )}
                              </div>
                              {q ? <p className="text-xs text-zinc-500">{q.price}</p> : <p className="text-xs text-zinc-600">—</p>}
                            </div>
                          </div>

                          {/* Change + actions */}
                          <div className="flex items-center gap-2">
                            {q && (
                              <span className={`flex items-center gap-1 text-xs font-semibold ${q.isPositive ? "text-green-400" : "text-red-400"}`}>
                                {q.isPositive ? <TrendingUp size={11} /> : <TrendingDown size={11} />}
                                {q.change}
                              </span>
                            )}
                            <button
                              onClick={() => openAlertEditor(wl.id, ws)}
                              className="opacity-0 group-hover:opacity-100 p-1 text-zinc-600 hover:text-yellow-400 transition rounded hover:bg-zinc-700"
                              title="Set price alert"
                            >
                              <Bell size={13} />
                            </button>
                            <button
                              onClick={() => removeSymbol(wl.id, ws.symbol)}
                              className="opacity-0 group-hover:opacity-100 p-1 text-zinc-600 hover:text-red-400 transition rounded hover:bg-zinc-700"
                              title="Remove"
                            >
                              <X size={13} />
                            </button>
                          </div>
                        </div>

                        {/* Alert editor */}
                        {isAlertOpen && (
                          <div className="mx-3 mb-2 p-3 bg-zinc-800/60 rounded-lg border border-zinc-700 flex items-center gap-3">
                            <div className="flex items-center gap-1.5">
                              <label className="text-[10px] text-zinc-500 uppercase">Above</label>
                              <input
                                type="number"
                                value={alertAbove}
                                onChange={(e) => setAlertAbove(e.target.value)}
                                placeholder="—"
                                className="w-28 bg-black border border-zinc-700 rounded px-2 py-1 text-xs text-white outline-none focus:border-yellow-500"
                              />
                            </div>
                            <div className="flex items-center gap-1.5">
                              <label className="text-[10px] text-zinc-500 uppercase">Below</label>
                              <input
                                type="number"
                                value={alertBelow}
                                onChange={(e) => setAlertBelow(e.target.value)}
                                placeholder="—"
                                className="w-28 bg-black border border-zinc-700 rounded px-2 py-1 text-xs text-white outline-none focus:border-yellow-500"
                              />
                            </div>
                            <button
                              onClick={() => saveAlert(wl.id, ws.symbol)}
                              className="px-2.5 py-1 bg-yellow-600 hover:bg-yellow-500 text-white text-xs font-semibold rounded transition"
                            >
                              Save
                            </button>
                            <button
                              onClick={() => setAlertKey(null)}
                              className="p-1 text-zinc-500 hover:text-zinc-300 transition"
                            >
                              <X size={12} />
                            </button>
                          </div>
                        )}
                      </div>
                    );
                  })
                )}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
