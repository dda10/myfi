"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import {
  TrendingUp, TrendingDown, RefreshCw, Star, X, Plus,
  ChevronDown, Pencil, Trash2, ArrowUp, ArrowDown, Bell, Check,
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
  whitelisted?: boolean;
}

// --- API helpers ---

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

export function Watchlist() {
  const { setActiveSymbol, setActiveTab } = useApp();
  const { formatCurrency } = useI18n();

  // State
  const [watchlists, setWatchlists] = useState<WatchlistData[]>([]);
  const [activeId, setActiveId] = useState<number | null>(null);
  const [quotes, setQuotes] = useState<Record<string, QuoteData>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // UI state
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [showAddSymbol, setShowAddSymbol] = useState(false);
  const [addSymbolInput, setAddSymbolInput] = useState("");
  const [showCreateWl, setShowCreateWl] = useState(false);
  const [createWlName, setCreateWlName] = useState("");
  const [renamingId, setRenamingId] = useState<number | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [alertSymbol, setAlertSymbol] = useState<string | null>(null);
  const [alertAbove, setAlertAbove] = useState("");
  const [alertBelow, setAlertBelow] = useState("");

  const dropdownRef = useRef<HTMLDivElement>(null);

  const active = watchlists.find((w) => w.id === activeId) ?? watchlists[0] ?? null;
  const symbols = active?.symbols ?? [];

  // --- Data fetching ---

  const fetchWatchlists = useCallback(async () => {
    try {
      const data = await apiFetch<WatchlistData[]>("/api/watchlists");
      const lists = data ?? [];
      setWatchlists(lists);
      if (lists.length > 0 && !lists.find((w) => w.id === activeId)) {
        setActiveId(lists[0].id);
      }
      setError(null);
    } catch {
      setError("Failed to load watchlists");
    }
  }, [activeId]);

  const fetchQuotes = useCallback(async () => {
    const syms = symbols.map((s) => s.symbol);
    if (syms.length === 0) return;
    try {
      const res = await fetch(
        `${API}/api/market/quote?symbols=${syms.join(",")}`,
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
            whitelisted: q.whitelisted,
          };
        });
        setQuotes(map);
      }
    } catch {
      /* ignore quote errors */
    }
  }, [symbols, formatCurrency]);

  // Initial load
  useEffect(() => {
    setLoading(true);
    fetchWatchlists().finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  // Refresh quotes when active watchlist changes
  useEffect(() => {
    fetchQuotes();
    const id = setInterval(fetchQuotes, 15000);
    return () => clearInterval(id);
  }, [fetchQuotes]);

  // Close dropdown on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  // --- Actions ---

  const createWatchlist = async () => {
    const name = createWlName.trim();
    if (!name) return;
    try {
      const wl = await apiFetch<WatchlistData>("/api/watchlists", {
        method: "POST",
        body: JSON.stringify({ name }),
      });
      setWatchlists((prev) => [...prev, wl]);
      setActiveId(wl.id);
      setCreateWlName("");
      setShowCreateWl(false);
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
      if (activeId === id) setActiveId(null);
      setDeletingId(null);
    } catch { /* ignore */ }
  };

  const addSymbol = async () => {
    if (!active) return;
    const sym = addSymbolInput.trim().toUpperCase();
    if (!sym) return;
    try {
      await apiFetch(`/api/watchlists/${active.id}/symbols`, {
        method: "POST",
        body: JSON.stringify({ symbol: sym }),
      });
      await fetchWatchlists();
      setAddSymbolInput("");
      setShowAddSymbol(false);
    } catch { /* ignore */ }
  };

  const removeSymbol = async (symbol: string) => {
    if (!active) return;
    try {
      await apiFetch(`/api/watchlists/${active.id}/symbols/${symbol}`, {
        method: "DELETE",
      });
      setWatchlists((prev) =>
        prev.map((w) =>
          w.id === active.id
            ? { ...w, symbols: w.symbols.filter((s) => s.symbol !== symbol) }
            : w
        )
      );
    } catch { /* ignore */ }
  };

  const moveSymbol = async (idx: number, dir: -1 | 1) => {
    if (!active) return;
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= symbols.length) return;
    const reordered = [...symbols];
    [reordered[idx], reordered[newIdx]] = [reordered[newIdx], reordered[idx]];
    const newSymbols = reordered.map((s, i) => ({ ...s, position: i }));
    // Optimistic update
    setWatchlists((prev) =>
      prev.map((w) => (w.id === active.id ? { ...w, symbols: newSymbols } : w))
    );
    try {
      await apiFetch(`/api/watchlists/${active.id}/reorder`, {
        method: "PUT",
        body: JSON.stringify({ symbols: newSymbols.map((s) => s.symbol) }),
      });
    } catch {
      await fetchWatchlists(); // rollback
    }
  };

  const saveAlert = async (symbol: string) => {
    if (!active) return;
    const above = alertAbove ? parseFloat(alertAbove) : null;
    const below = alertBelow ? parseFloat(alertBelow) : null;
    try {
      await apiFetch(`/api/watchlists/${active.id}/symbols/${symbol}/alert`, {
        method: "PUT",
        body: JSON.stringify({
          alertAbove: above && !isNaN(above) ? above : null,
          alertBelow: below && !isNaN(below) ? below : null,
        }),
      });
      await fetchWatchlists();
      setAlertSymbol(null);
      setAlertAbove("");
      setAlertBelow("");
    } catch { /* ignore */ }
  };

  const openAlertEditor = (ws: WatchlistSymbol) => {
    setAlertSymbol(ws.symbol);
    setAlertAbove(ws.priceAlertAbove != null ? String(ws.priceAlertAbove) : "");
    setAlertBelow(ws.priceAlertBelow != null ? String(ws.priceAlertBelow) : "");
  };

  // --- Render ---

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-xl overflow-hidden shadow-lg">
      {/* Header with watchlist selector */}
      <div className="px-5 py-4 border-b border-zinc-800 flex justify-between items-center">
        <div className="flex items-center gap-2 flex-1 min-w-0">
          <Star size={15} className="text-yellow-400 fill-yellow-400 shrink-0" />

          {/* Watchlist dropdown */}
          <div className="relative" ref={dropdownRef}>
            <button
              onClick={() => setDropdownOpen((v) => !v)}
              className="flex items-center gap-1.5 text-base font-semibold text-white hover:text-indigo-300 transition truncate max-w-[180px]"
            >
              <span className="truncate">{active?.name ?? "Watchlist"}</span>
              <ChevronDown size={14} className={`shrink-0 transition-transform ${dropdownOpen ? "rotate-180" : ""}`} />
            </button>

            {dropdownOpen && (
              <div className="absolute top-full left-0 mt-1 w-56 bg-zinc-800 border border-zinc-700 rounded-lg shadow-xl z-50 py-1">
                {watchlists.map((wl) => (
                  <div key={wl.id} className="flex items-center group">
                    {renamingId === wl.id ? (
                      <div className="flex items-center gap-1 px-3 py-1.5 w-full">
                        <input
                          value={renameValue}
                          onChange={(e) => setRenameValue(e.target.value)}
                          onKeyDown={(e) => e.key === "Enter" && renameWatchlist(wl.id)}
                          autoFocus
                          className="flex-1 bg-black border border-zinc-600 rounded px-2 py-0.5 text-sm text-white outline-none focus:border-indigo-500"
                        />
                        <button onClick={() => renameWatchlist(wl.id)} className="p-0.5 text-green-400 hover:text-green-300">
                          <Check size={13} />
                        </button>
                        <button onClick={() => setRenamingId(null)} className="p-0.5 text-zinc-500 hover:text-zinc-300">
                          <X size={13} />
                        </button>
                      </div>
                    ) : (
                      <>
                        <button
                          onClick={() => { setActiveId(wl.id); setDropdownOpen(false); }}
                          className={`flex-1 text-left px-3 py-1.5 text-sm truncate transition ${
                            wl.id === active?.id ? "text-indigo-400 bg-zinc-700/50" : "text-zinc-300 hover:bg-zinc-700/50"
                          }`}
                        >
                          {wl.name}
                          <span className="ml-1.5 text-[10px] text-zinc-500">{wl.symbols.length}</span>
                        </button>
                        <button
                          onClick={() => { setRenamingId(wl.id); setRenameValue(wl.name); }}
                          className="opacity-0 group-hover:opacity-100 p-1 text-zinc-500 hover:text-indigo-400 transition"
                          title="Rename"
                        >
                          <Pencil size={11} />
                        </button>
                        <button
                          onClick={() => setDeletingId(wl.id)}
                          className="opacity-0 group-hover:opacity-100 p-1 text-zinc-500 hover:text-red-400 transition mr-1"
                          title="Delete"
                        >
                          <Trash2 size={11} />
                        </button>
                      </>
                    )}
                  </div>
                ))}

                {/* Create new watchlist */}
                <div className="border-t border-zinc-700 mt-1 pt-1">
                  {showCreateWl ? (
                    <div className="flex items-center gap-1 px-3 py-1.5">
                      <input
                        value={createWlName}
                        onChange={(e) => setCreateWlName(e.target.value)}
                        onKeyDown={(e) => e.key === "Enter" && createWatchlist()}
                        placeholder="New watchlist name"
                        autoFocus
                        className="flex-1 bg-black border border-zinc-600 rounded px-2 py-0.5 text-sm text-white placeholder-zinc-600 outline-none focus:border-indigo-500"
                      />
                      <button onClick={createWatchlist} className="p-0.5 text-green-400 hover:text-green-300">
                        <Check size={13} />
                      </button>
                      <button onClick={() => { setShowCreateWl(false); setCreateWlName(""); }} className="p-0.5 text-zinc-500 hover:text-zinc-300">
                        <X size={13} />
                      </button>
                    </div>
                  ) : (
                    <button
                      onClick={() => setShowCreateWl(true)}
                      className="w-full text-left px-3 py-1.5 text-sm text-indigo-400 hover:bg-zinc-700/50 transition flex items-center gap-1.5"
                    >
                      <Plus size={12} /> New watchlist
                    </button>
                  )}
                </div>
              </div>
            )}
          </div>

          <span className="text-[10px] bg-zinc-800 text-zinc-400 px-1.5 py-0.5 rounded font-mono">
            {symbols.length}
          </span>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowAddSymbol((v) => !v)}
            className="p-1.5 text-zinc-400 hover:text-indigo-400 transition rounded hover:bg-zinc-800"
            title="Add symbol"
          >
            <Plus size={15} />
          </button>
          <button
            onClick={() => { fetchWatchlists(); fetchQuotes(); }}
            className="p-1.5 text-zinc-400 hover:text-white transition rounded hover:bg-zinc-800"
          >
            <RefreshCw size={13} className={loading ? "animate-spin" : ""} />
          </button>
        </div>
      </div>

      {/* Delete confirmation */}
      {deletingId !== null && (
        <div className="px-4 py-3 border-b border-zinc-800 bg-red-950/30 flex items-center justify-between">
          <span className="text-sm text-red-300">
            Delete &quot;{watchlists.find((w) => w.id === deletingId)?.name}&quot;?
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => deleteWatchlist(deletingId)}
              className="px-3 py-1 bg-red-600 hover:bg-red-500 text-white text-xs font-semibold rounded transition"
            >
              Delete
            </button>
            <button
              onClick={() => setDeletingId(null)}
              className="px-3 py-1 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-xs rounded transition"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Add symbol input */}
      {showAddSymbol && (
        <div className="px-4 py-3 border-b border-zinc-800 flex gap-2">
          <input
            type="text"
            value={addSymbolInput}
            onChange={(e) => setAddSymbolInput(e.target.value.toUpperCase())}
            onKeyDown={(e) => e.key === "Enter" && addSymbol()}
            placeholder="Type symbol, e.g. VCB"
            autoFocus
            className="flex-1 bg-black border border-zinc-700 rounded-lg px-3 py-1.5 text-sm text-white placeholder-zinc-600 outline-none focus:border-indigo-500"
          />
          <button
            onClick={addSymbol}
            className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white text-sm font-semibold rounded-lg transition"
          >
            Add
          </button>
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="px-4 py-3 text-sm text-red-400 bg-red-950/20 border-b border-zinc-800">
          {error}
        </div>
      )}

      {/* Symbol list */}
      <div className="p-2 min-h-[280px]">
        {loading && symbols.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 text-zinc-500">
            <RefreshCw size={22} className="animate-spin mb-3" />
            <p className="text-sm">Loading watchlists...</p>
          </div>
        ) : symbols.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 text-zinc-500">
            <Star size={28} className="mb-3 text-zinc-700" />
            <p className="text-sm">No stocks in this watchlist.</p>
            <p className="text-xs text-zinc-600 mt-1">Click <strong>+</strong> to add stocks.</p>
          </div>
        ) : (
          symbols.map((ws, idx) => {
            const q = quotes[ws.symbol];
            const hasAlert = ws.priceAlertAbove != null || ws.priceAlertBelow != null;
            return (
              <div key={ws.symbol}>
                <div className="flex items-center justify-between p-3 rounded-lg transition group hover:bg-zinc-800/80">
                  {/* Reorder buttons */}
                  <div className="flex flex-col mr-2 gap-0.5">
                    <button
                      onClick={() => moveSymbol(idx, -1)}
                      disabled={idx === 0}
                      className="p-0.5 text-zinc-700 hover:text-zinc-400 disabled:opacity-20 transition"
                    >
                      <ArrowUp size={11} />
                    </button>
                    <button
                      onClick={() => moveSymbol(idx, 1)}
                      disabled={idx === symbols.length - 1}
                      className="p-0.5 text-zinc-700 hover:text-zinc-400 disabled:opacity-20 transition"
                    >
                      <ArrowDown size={11} />
                    </button>
                  </div>

                  {/* Symbol info — click to navigate */}
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
                        {q?.whitelisted === true && (
                          <span className="text-[9px] bg-emerald-900/60 text-emerald-400 px-1 py-0.5 rounded font-bold leading-none" title="High liquidity">LIQ</span>
                        )}
                        {q?.whitelisted === false && (
                          <span className="text-[9px] bg-amber-900/40 text-amber-500 px-1 py-0.5 rounded font-bold leading-none" title="Low liquidity">LOW</span>
                        )}
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
                      onClick={() => alertSymbol === ws.symbol ? setAlertSymbol(null) : openAlertEditor(ws)}
                      className="opacity-0 group-hover:opacity-100 p-1 text-zinc-600 hover:text-yellow-400 transition rounded hover:bg-zinc-700"
                      title="Set price alert"
                    >
                      <Bell size={13} />
                    </button>
                    <button
                      onClick={() => removeSymbol(ws.symbol)}
                      className="opacity-0 group-hover:opacity-100 p-1 text-zinc-600 hover:text-red-400 transition rounded hover:bg-zinc-700"
                      title="Remove"
                    >
                      <X size={13} />
                    </button>
                  </div>
                </div>

                {/* Alert editor inline */}
                {alertSymbol === ws.symbol && (
                  <div className="mx-3 mb-2 p-3 bg-zinc-800/60 rounded-lg border border-zinc-700 flex items-center gap-3">
                    <div className="flex items-center gap-1.5">
                      <label className="text-[10px] text-zinc-500 uppercase">Above</label>
                      <input
                        type="number"
                        value={alertAbove}
                        onChange={(e) => setAlertAbove(e.target.value)}
                        placeholder="—"
                        className="w-24 bg-black border border-zinc-700 rounded px-2 py-1 text-xs text-white outline-none focus:border-yellow-500"
                      />
                    </div>
                    <div className="flex items-center gap-1.5">
                      <label className="text-[10px] text-zinc-500 uppercase">Below</label>
                      <input
                        type="number"
                        value={alertBelow}
                        onChange={(e) => setAlertBelow(e.target.value)}
                        placeholder="—"
                        className="w-24 bg-black border border-zinc-700 rounded px-2 py-1 text-xs text-white outline-none focus:border-yellow-500"
                      />
                    </div>
                    <button
                      onClick={() => saveAlert(ws.symbol)}
                      className="px-2.5 py-1 bg-yellow-600 hover:bg-yellow-500 text-white text-xs font-semibold rounded transition"
                    >
                      Save
                    </button>
                    <button
                      onClick={() => setAlertSymbol(null)}
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
    </div>
  );
}
